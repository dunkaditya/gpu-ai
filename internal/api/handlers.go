package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gpuai/gpuctl/internal/auth"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/provision"
	"github.com/gpuai/gpuctl/internal/wireguard"
)

// CreateInstanceRequest is the JSON body for POST /api/v1/instances.
type CreateInstanceRequest struct {
	GPUType         string   `json:"gpu_type"`                    // e.g. "a100_80gb"
	GPUCount        int      `json:"gpu_count"`                   // 1-8
	Region          string   `json:"region,omitempty"`            // optional
	Tier            string   `json:"tier"`                        // "spot" or "on_demand"
	SSHKeyIDs       []string `json:"ssh_key_ids,omitempty"`        // optional; auto-includes user's keys if empty
	Name            *string  `json:"name,omitempty"`              // optional display label
	MaxPricePerHour *float64 `json:"max_price_per_hour,omitempty"` // optional cap
}

// Validate checks that the CreateInstanceRequest fields are valid.
func (req *CreateInstanceRequest) Validate() error {
	if req.GPUType == "" {
		return errors.New("gpu_type is required")
	}
	if req.GPUCount < 1 || req.GPUCount > 8 {
		return errors.New("gpu_count must be between 1 and 8")
	}
	if req.Tier != "spot" && req.Tier != "on_demand" {
		return errors.New("tier must be 'spot' or 'on_demand'")
	}
	// SSHKeyIDs is now optional -- the provisioning engine handles the fallback
	// by auto-including the user's keys when none are specified.
	return nil
}

// InstanceResponse is the customer-facing JSON representation of an instance.
// It structurally excludes all upstream provider details (defense by omission).
type InstanceResponse struct {
	ID           string          `json:"id"`
	Name         *string         `json:"name,omitempty"`
	Status       string          `json:"status"`        // external state: starting, running, stopping, terminated, error
	GPUType      string          `json:"gpu_type"`
	GPUCount     int             `json:"gpu_count"`
	Tier         string          `json:"tier"`
	Region       string          `json:"region"`
	PricePerHour float64         `json:"price_per_hour"`
	Connection   *ConnectionInfo `json:"connection"`
	ErrorReason  *string         `json:"error_reason,omitempty"`
	CreatedAt    string          `json:"created_at"`               // RFC 3339
	ReadyAt      *string         `json:"ready_at,omitempty"`       // RFC 3339
	TerminatedAt *string         `json:"terminated_at,omitempty"` // RFC 3339
}

// ConnectionInfo contains SSH connection details for a running instance.
type ConnectionInfo struct {
	Hostname   string `json:"hostname"`
	Port       int    `json:"port"`
	SSHCommand string `json:"ssh_command"`
}

// instanceToResponse maps an internal db.Instance to a customer-facing InstanceResponse.
// Uses provision.ExternalState to collapse internal states to external.
// Excludes all upstream provider fields by structural omission.
// When WireGuard is configured and the instance has a tunnel address, SSH connection
// info points to the WG proxy (port derived from tunnel IP) instead of the raw hostname.
func (s *Server) instanceToResponse(inst *db.Instance) InstanceResponse {
	resp := InstanceResponse{
		ID:           inst.InstanceID,
		Name:         inst.Name,
		Status:       provision.ExternalState(inst.Status),
		GPUType:      inst.GPUType,
		GPUCount:     inst.GPUCount,
		Tier:         inst.Tier,
		Region:       inst.Region,
		PricePerHour: inst.PricePerHour,
		ErrorReason:  inst.ErrorReason,
		CreatedAt:    inst.CreatedAt.Format(time.RFC3339),
	}

	// Build connection info. Prefer WireGuard proxy coordinates when available.
	if inst.WGAddress != nil && s.config.WGProxyEndpoint != "" {
		// Parse tunnel IP from CIDR (e.g. "10.0.0.5/16" -> 10.0.0.5).
		ipStr := strings.SplitN(*inst.WGAddress, "/", 2)[0]
		tunnelIP := net.ParseIP(ipStr)

		// Extract proxy host from WGProxyEndpoint (e.g. "203.0.113.1:51820" -> 203.0.113.1).
		proxyHost, _, err := net.SplitHostPort(s.config.WGProxyEndpoint)

		if tunnelIP != nil && err == nil {
			port := wireguard.PortFromTunnelIP(tunnelIP)
			resp.Connection = &ConnectionInfo{
				Hostname:   proxyHost,
				Port:       port,
				SSHCommand: fmt.Sprintf("ssh -p %d root@%s", port, proxyHost),
			}
		}
	}

	// Fallback: direct hostname:22 when WG is not configured or parsing failed.
	if resp.Connection == nil {
		resp.Connection = &ConnectionInfo{
			Hostname:   inst.Hostname,
			Port:       22,
			SSHCommand: "ssh root@" + inst.Hostname,
		}
	}

	// Format optional timestamps.
	if inst.ReadyAt != nil {
		ts := inst.ReadyAt.Format(time.RFC3339)
		resp.ReadyAt = &ts
	}
	if inst.TerminatedAt != nil {
		ts := inst.TerminatedAt.Format(time.RFC3339)
		resp.TerminatedAt = &ts
	}

	return resp
}

// handleCreateInstance handles POST /api/v1/instances.
// Creates a new GPU instance via the provisioning engine.
func (s *Server) handleCreateInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract Clerk claims.
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	// 2. Auto-provision org + user from Clerk IDs.
	// EnsureOrgAndUser returns both the internal org UUID and user UUID.
	// The user UUID (not the Clerk user ID) must be used for instances.user_id FK.
	orgID, userID, err := s.db.EnsureOrgAndUser(ctx, claims.OrgID, claims.UserID, "")
	if err != nil {
		slog.Error("failed to ensure org and user",
			slog.String("clerk_org_id", claims.OrgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 3. Decode and validate request body.
	var req CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid-request", "Invalid JSON request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeProblem(w, http.StatusBadRequest, "validation-error", err.Error())
		return
	}

	// 4. Build engine-level provision request.
	provReq := provision.ProvisionRequest{
		OrgID:           orgID,
		UserID:          userID,
		GPUType:         provider.GPUType(req.GPUType),
		GPUCount:        req.GPUCount,
		Tier:            provider.InstanceTier(req.Tier),
		Region:          req.Region,
		Name:            req.Name,
		SSHKeyIDs:       req.SSHKeyIDs,
		MaxPricePerHour: req.MaxPricePerHour,
	}

	// 5. Call provisioning engine.
	provResp, err := s.engine.Provision(ctx, provReq)
	if err != nil {
		// Check for specific known errors.
		if errors.Is(err, provision.ErrPriceExceeded) {
			writeProblem(w, http.StatusConflict, "price-exceeded",
				"Current price exceeds your maximum price per hour")
			return
		}
		if errors.Is(err, provision.ErrNoProvider) {
			writeProblem(w, http.StatusConflict, "no-availability",
				"No GPU availability matching your request")
			return
		}
		if errors.Is(err, provision.ErrSSHKeysNotFound) {
			writeProblem(w, http.StatusBadRequest, "ssh-keys-not-found",
				"None of the provided SSH key IDs were found")
			return
		}
		if errors.Is(err, provision.ErrSpendingLimitReached) {
			writeProblem(w, http.StatusPaymentRequired, "spending_limit_reached",
				"Instance creation blocked: organization spending limit reached. Remove limit or contact support.")
			return
		}
		// Generic error: log internally, return generic message to customer (API-09).
		slog.Error("provisioning failed",
			slog.String("org_id", orgID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "provisioning-error",
			"Failed to provision instance. Please try again.")
		return
	}

	// 6. Fetch the created instance for full response.
	inst, err := s.db.GetInstance(ctx, provResp.InstanceID)
	if err != nil {
		slog.Error("failed to fetch newly created instance",
			slog.String("instance_id", provResp.InstanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Instance created but failed to retrieve details")
		return
	}

	writeJSON(w, http.StatusCreated, s.instanceToResponse(inst))
}

// handleListInstances handles GET /api/v1/instances.
// Returns paginated list of instances for the authenticated organization.
func (s *Server) handleListInstances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract org ID from claims.
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	// Look up internal org ID from Clerk org ID.
	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// Org not provisioned yet -- return empty list.
			writeJSON(w, http.StatusOK, PageResult[InstanceResponse]{
				Data:    []InstanceResponse{},
				HasMore: false,
			})
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 2. Parse pagination params.
	params := ParsePageParams(r)

	// 3. Decode cursor if present.
	var cursorTime *time.Time
	var cursorID string
	if params.Cursor != "" {
		ct, cid, err := DecodeCursor(params.Cursor)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-cursor", "Invalid pagination cursor")
			return
		}
		cursorTime = &ct
		cursorID = cid
	}

	// 4. Query instances (limit+1 for has_more detection).
	instances, err := s.db.ListInstances(ctx, orgID, cursorTime, cursorID, params.Limit)
	if err != nil {
		slog.Error("failed to list instances", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to list instances")
		return
	}

	// 5. Determine has_more by checking if we got limit+1 results.
	hasMore := len(instances) > params.Limit
	if hasMore {
		instances = instances[:params.Limit]
	}

	// 6. Map to response type.
	data := make([]InstanceResponse, 0, len(instances))
	for i := range instances {
		data = append(data, s.instanceToResponse(&instances[i]))
	}

	// 7. Encode cursor from last item.
	var cursor string
	if hasMore && len(instances) > 0 {
		last := instances[len(instances)-1]
		cursor = EncodeCursor(last.CreatedAt, last.InstanceID)
	}

	writeJSON(w, http.StatusOK, PageResult[InstanceResponse]{
		Data:    data,
		Cursor:  cursor,
		HasMore: hasMore,
	})
}

// handleGetInstance handles GET /api/v1/instances/{id}.
// Returns instance details scoped to the authenticated organization.
func (s *Server) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract instance ID from path.
	instanceID := r.PathValue("id")
	if instanceID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "Instance ID is required")
		return
	}

	// 2. Extract org ID from claims.
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 3. Fetch instance scoped to org.
	inst, err := s.db.GetInstanceForOrg(ctx, instanceID, orgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to get instance", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to retrieve instance")
		return
	}

	writeJSON(w, http.StatusOK, s.instanceToResponse(inst))
}

// handleDeleteInstance handles DELETE /api/v1/instances/{id}.
// Terminates an instance idempotently (returns 200 if already terminated).
func (s *Server) handleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Extract instance ID and org ID.
	instanceID := r.PathValue("id")
	if instanceID == "" {
		writeProblem(w, http.StatusBadRequest, "missing-id", "Instance ID is required")
		return
	}

	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		writeProblem(w, http.StatusUnauthorized, "unauthenticated", "Valid authentication required")
		return
	}

	orgID, err := s.db.GetOrgIDByClerkOrgID(ctx, claims.OrgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to look up org", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 2. Verify ownership.
	inst, err := s.db.GetInstanceForOrg(ctx, instanceID, orgID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "not-found", "Instance not found")
			return
		}
		slog.Error("failed to get instance", slog.String("error", err.Error()))
		writeProblem(w, http.StatusInternalServerError, "internal-error", "Failed to process request")
		return
	}

	// 3. If already terminated, return 200 with current state (idempotent per INST-06).
	if inst.Status == provision.StateTerminated {
		writeJSON(w, http.StatusOK, s.instanceToResponse(inst))
		return
	}

	// 4. Terminate via engine.
	if err := s.engine.Terminate(ctx, instanceID); err != nil {
		slog.Error("termination failed",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "termination-error",
			"Failed to terminate instance. Please try again.")
		return
	}

	// 5. Re-fetch instance from DB for updated state.
	inst, err = s.db.GetInstanceForOrg(ctx, instanceID, orgID)
	if err != nil {
		slog.Error("failed to re-fetch instance after termination",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		writeProblem(w, http.StatusInternalServerError, "internal-error",
			"Instance terminated but failed to retrieve updated details")
		return
	}

	writeJSON(w, http.StatusOK, s.instanceToResponse(inst))
}
