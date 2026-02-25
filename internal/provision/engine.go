// Package provision orchestrates the instance provisioning flow.
package provision

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/wireguard"
)

// providerCandidate pairs a provider with a matching offering for price-sorted selection.
type providerCandidate struct {
	prov     provider.Provider
	offering provider.GPUOffering
}

// ProvisionRequest contains the parameters for provisioning a new instance.
// This is the engine-level request, distinct from provider.ProvisionRequest.
type ProvisionRequest struct {
	OrgID           string
	UserID          string // internal user UUID from users.user_id (not Clerk user ID)
	GPUType         provider.GPUType
	GPUCount        int
	Tier            provider.InstanceTier
	Region          string   // optional
	Name            *string  // optional user display label
	SSHKeyIDs       []string // UUIDs of ssh_keys to resolve
	MaxPricePerHour *float64 // optional price cap
}

// ProvisionResponse is returned after a successful provisioning call.
type ProvisionResponse struct {
	InstanceID   string
	Hostname     string
	PricePerHour float64
	Status       string
}

// EngineDeps holds the dependencies for the provisioning engine.
// Follows the project's ServerDeps constructor injection pattern.
type EngineDeps struct {
	Registry *provider.Registry
	DB       *db.Pool
	Config   *config.Config
	Logger   *slog.Logger
	// Optional: set to nil if WG proxy not configured.
	WGManager *wireguard.Manager
	IPAM      *wireguard.IPAM
	// OnStatusChange is called when progressStatus changes an instance's state.
	// Used to publish SSE events. Nil means no notification.
	OnStatusChange func(instanceID, status string)
}

// Engine orchestrates instance provisioning and termination.
// It coordinates provider adapters, WireGuard, cloud-init, and DB persistence.
type Engine struct {
	registry       *provider.Registry
	db             *db.Pool
	config         *config.Config
	logger         *slog.Logger
	wgMgr          *wireguard.Manager
	ipam           *wireguard.IPAM
	onStatusChange func(instanceID, status string)
}

// NewEngine creates a new provisioning engine with the given dependencies.
func NewEngine(deps EngineDeps) *Engine {
	return &Engine{
		registry:       deps.Registry,
		db:             deps.DB,
		config:         deps.Config,
		logger:         deps.Logger,
		wgMgr:          deps.WGManager,
		ipam:           deps.IPAM,
		onStatusChange: deps.OnStatusChange,
	}
}

// SetOnStatusChange sets the callback invoked when progressStatus changes
// an instance's state. Allows post-construction wiring when Engine is created
// before the API server (which owns the SSE broker).
func (e *Engine) SetOnStatusChange(fn func(instanceID, status string)) {
	e.onStatusChange = fn
}

// ErrNoProvider is returned when no suitable provider is found for the request.
var ErrNoProvider = errors.New("no suitable provider available")

// ErrPriceExceeded is returned when the current price exceeds the requested max.
var ErrPriceExceeded = errors.New("current price exceeds maximum price per hour")

// ErrSSHKeysNotFound is returned when none of the requested SSH keys exist.
var ErrSSHKeysNotFound = errors.New("ssh keys not found")

// ErrSpendingLimitReached is returned when an org has reached their spending limit.
var ErrSpendingLimitReached = errors.New("spending limit reached: new instance creation blocked")

// Provision creates a new GPU instance.
//
// Steps:
//  1. Generate instance ID and internal token
//  2. Resolve SSH keys from DB
//  3. Select provider and verify pricing
//  4. Build callback URL (uses GPUCTL_PUBLIC_URL if configured)
//  5. Optionally generate WireGuard keys and render cloud-init
//  6. Call provider to create upstream instance
//  7. Persist instance record to DB
//  8. Kick off async status progression (creating -> provisioning -> booting)
func (e *Engine) Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResponse, error) {
	// 1. Generate instance ID: "gpu-" + 4 random hex bytes.
	instanceID, err := generateInstanceID()
	if err != nil {
		return nil, fmt.Errorf("provision: generate instance ID: %w", err)
	}

	// Generate internal token: 16 random hex bytes for callback auth.
	internalToken, err := generateHexToken(16)
	if err != nil {
		return nil, fmt.Errorf("provision: generate internal token: %w", err)
	}

	// Generate branded hostname.
	hostname := instanceID + ".gpu.ai"

	e.logger.Info("provisioning instance",
		slog.String("instance_id", instanceID),
		slog.String("gpu_type", string(req.GPUType)),
		slog.Int("gpu_count", req.GPUCount),
		slog.String("tier", string(req.Tier)),
	)

	// 2. Look up SSH keys from DB.
	// Smart default: if no explicit key IDs provided, auto-include all of the user's keys.
	var sshKeys []db.SSHKey
	if len(req.SSHKeyIDs) > 0 {
		// Explicit key IDs provided -- look them up.
		sshKeys, err = e.db.GetSSHKeysByIDs(ctx, req.SSHKeyIDs)
		if err != nil {
			return nil, fmt.Errorf("provision: look up ssh keys: %w", err)
		}
	} else {
		// No key IDs provided -- auto-include all of user's keys.
		sshKeys, err = e.db.GetSSHKeysByUserID(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("provision: look up user ssh keys: %w", err)
		}
	}
	if len(sshKeys) == 0 {
		return nil, ErrSSHKeysNotFound // "at least one SSH key required"
	}

	// Collect public keys.
	sshPubKeys := make([]string, 0, len(sshKeys))
	for _, k := range sshKeys {
		sshPubKeys = append(sshPubKeys, k.PublicKey)
	}

	// 3. Check spending limit: block new instance creation if org is at limit.
	if err := e.checkSpendingLimit(ctx, req.OrgID); err != nil {
		return nil, err
	}

	// 4. Select provider candidates (sorted by price ascending).
	candidates, err := e.selectProviderCandidates(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check price cap against cheapest candidate.
	if req.MaxPricePerHour != nil && candidates[0].offering.PricePerHour > *req.MaxPricePerHour {
		return nil, ErrPriceExceeded
	}

	// 5. Build callback URL using public URL if configured, falling back to branded hostname.
	callbackURL := buildCallbackURL(e.config.GpuctlPublicURL, hostname, instanceID)

	// 6. Optionally generate WireGuard keys and cloud-init.
	// WG setup happens once before the provider retry loop.
	var wgPubKey, wgPrivKeyEnc, wgAddress *string
	var startupScript string

	if e.wgMgr != nil && e.config.WGEncryptionKeyBytes != nil {
		kp, err := wireguard.GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("provision: generate wireguard keys: %w", err)
		}

		// Encrypt private key for DB storage.
		encrypted, err := wireguard.EncryptPrivateKey(kp.PrivateKey, e.config.WGEncryptionKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("provision: encrypt wireguard key: %w", err)
		}

		wgPubKey = &kp.PublicKey
		wgPrivKeyEnc = &encrypted

		// Allocate tunnel IP if IPAM is configured.
		if e.ipam != nil {
			tx, err := e.db.PgxPool().Begin(ctx)
			if err != nil {
				return nil, fmt.Errorf("provision: begin tx for IPAM: %w", err)
			}
			tunnelIP, err := e.ipam.AllocateAddress(ctx, tx)
			if err != nil {
				_ = tx.Rollback(ctx)
				return nil, fmt.Errorf("provision: allocate tunnel IP: %w", err)
			}
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("provision: commit IPAM tx: %w", err)
			}
			addr := tunnelIP.String() + "/16"
			wgAddress = &addr
		}

		// Add WireGuard peer to proxy BEFORE provider call.
		// The peer must be registered so the instance can establish the tunnel on boot.
		if wgAddress != nil {
			addrStr, _, _ := strings.Cut(*wgAddress, "/")
			tunnelIP := net.ParseIP(addrStr)
			if tunnelIP == nil {
				return nil, fmt.Errorf("provision: invalid tunnel IP: %s", *wgAddress)
			}
			externalPort := wireguard.PortFromTunnelIP(tunnelIP)
			if err := e.wgMgr.AddPeer(ctx, kp.PublicKey, tunnelIP, externalPort); err != nil {
				return nil, fmt.Errorf("provision: add wireguard peer: %w", err)
			}
			e.logger.Info("added WireGuard peer for instance",
				slog.String("instance_id", instanceID),
				slog.String("tunnel_ip", tunnelIP.String()),
				slog.Int("external_port", externalPort),
			)
		}

		// Render cloud-init bootstrap script.
		bootstrapData := wireguard.BootstrapData{
			InstanceID:         instanceID,
			ProxyEndpoint:      e.config.WGProxyEndpoint,
			ProxyPublicKey:     e.config.WGProxyPublicKey,
			InstancePrivateKey: kp.PrivateKey,
			InstanceAddress:    *wgAddress,
			AllowedIPs:         "10.0.0.0/16",
			SSHAuthorizedKeys:  strings.Join(sshPubKeys, "\n"),
			InternalToken:      internalToken,
			Hostname:           hostname,
			CallbackURL:        callbackURL,
		}
		startupScript, err = wireguard.RenderBootstrap(bootstrapData)
		if err != nil {
			return nil, fmt.Errorf("provision: render cloud-init: %w", err)
		}
	}

	// 7. Build provider-level request and try provisioning with fallback retry.
	// Only the provider.Provision call is retried with different providers (max 3 attempts).
	// WG setup (key gen, IPAM, AddPeer) happened once above.
	provReq := provider.ProvisionRequest{
		InstanceID:    instanceID,
		GPUType:       req.GPUType,
		GPUCount:      req.GPUCount,
		Tier:          req.Tier,
		Region:        req.Region,
		SSHPublicKeys: sshPubKeys,
		InternalToken: internalToken,
		CallbackURL:   callbackURL,
		StartupScript: startupScript,
	}
	if wgAddress != nil {
		provReq.WireGuardAddress = *wgAddress
	}

	const maxProvisionAttempts = 3
	var prov provider.Provider
	var offering *provider.GPUOffering
	var provResult *provider.ProvisionResult
	var lastErr error

	for attempt := 0; attempt < maxProvisionAttempts && attempt < len(candidates); attempt++ {
		c := candidates[attempt]
		prov = c.prov
		offering = &c.offering

		e.logger.Info("attempting provision with provider",
			slog.String("instance_id", instanceID),
			slog.String("provider", prov.Name()),
			slog.Float64("price_per_hour", offering.PricePerHour),
			slog.Int("attempt", attempt+1),
		)

		provResult, lastErr = prov.Provision(ctx, provReq)
		if lastErr == nil {
			break // Success
		}

		e.logger.Warn("provider provision failed, trying next candidate",
			slog.String("instance_id", instanceID),
			slog.String("provider", prov.Name()),
			slog.String("error", lastErr.Error()),
			slog.Int("attempt", attempt+1),
		)
	}

	if lastErr != nil {
		// All attempts failed. Best-effort cleanup: remove WG peer if we added one.
		if e.wgMgr != nil && wgPubKey != nil && wgAddress != nil {
			addrStr, _, _ := strings.Cut(*wgAddress, "/")
			if tunnelIP := net.ParseIP(addrStr); tunnelIP != nil {
				externalPort := wireguard.PortFromTunnelIP(tunnelIP)
				if rmErr := e.wgMgr.RemovePeer(ctx, *wgPubKey, tunnelIP, externalPort); rmErr != nil {
					e.logger.Error("failed to clean up WireGuard peer after provision failure",
						slog.String("instance_id", instanceID),
						slog.String("error", rmErr.Error()),
					)
				}
			}
		}
		return nil, fmt.Errorf("provision: all providers failed (last: %s): %w", prov.Name(), lastErr)
	}

	// 8. Build and persist instance record.
	upstreamIP := &provResult.UpstreamIP
	if provResult.UpstreamIP == "" {
		upstreamIP = nil
	}

	inst := db.Instance{
		InstanceID:           instanceID,
		OrgID:                req.OrgID,
		UserID:               req.UserID,
		UpstreamProvider:     prov.Name(),
		UpstreamID:           provResult.UpstreamID,
		UpstreamIP:           upstreamIP,
		Hostname:             hostname,
		WGPublicKey:          wgPubKey,
		WGPrivateKeyEnc:      wgPrivKeyEnc,
		WGAddress:            wgAddress,
		Name:                 req.Name,
		GPUType:              string(req.GPUType),
		GPUCount:             req.GPUCount,
		Tier:                 string(req.Tier),
		Region:               offering.Region,
		PricePerHour:         offering.PricePerHour,
		UpstreamPricePerHour: provResult.CostPerHour,
		Status:               StateCreating,
		InternalToken:        &internalToken,
	}
	if err := e.db.CreateInstance(ctx, &inst); err != nil {
		return nil, fmt.Errorf("provision: persist instance: %w", err)
	}

	// 9. Kick off async status progression: creating -> provisioning -> (poll) -> booting.
	go e.progressStatus(instanceID)

	e.logger.Info("instance provisioned",
		slog.String("instance_id", instanceID),
		slog.String("upstream_id", provResult.UpstreamID),
		slog.String("provider", prov.Name()),
	)

	return &ProvisionResponse{
		InstanceID:   instanceID,
		Hostname:     hostname,
		PricePerHour: offering.PricePerHour,
		Status:       StateCreating,
	}, nil
}

// Terminate destroys a running instance.
//
// Steps:
//  1. Get instance from DB
//  2. Check state machine (idempotent if already terminated)
//  3. Update status to stopping
//  4. Call provider to terminate upstream instance
//  5. Mark as terminated in DB
//  6. Clean up WireGuard if configured
func (e *Engine) Terminate(ctx context.Context, instanceID string) error {
	// 1. Get instance from DB.
	inst, err := e.db.GetInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("terminate: get instance: %w", err)
	}

	// 2. If already in terminal state, return nil (idempotent).
	if inst.Status == StateTerminated {
		return nil
	}

	// 3. Transition to stopping using optimistic locking.
	if CanTransition(inst.Status, StateStopping) {
		updated, err := e.db.UpdateInstanceStatus(ctx, instanceID, inst.Status, StateStopping)
		if err != nil {
			return fmt.Errorf("terminate: update status to stopping: %w", err)
		}
		if !updated {
			// Status changed concurrently. Re-read and check.
			inst, err = e.db.GetInstance(ctx, instanceID)
			if err != nil {
				return fmt.Errorf("terminate: re-read instance: %w", err)
			}
			if inst.Status == StateTerminated {
				return nil
			}
		}
	}

	// 4. Look up provider and terminate upstream.
	prov, ok := e.registry.Get(inst.UpstreamProvider)
	if !ok {
		e.logger.Warn("provider not found for termination, marking as terminated",
			slog.String("instance_id", instanceID),
			slog.String("provider", inst.UpstreamProvider),
		)
	} else {
		if err := prov.Terminate(ctx, inst.UpstreamID); err != nil {
			// Set instance to error state with reason.
			reason := fmt.Sprintf("provider terminate failed: %v", err)
			if setErr := e.db.SetInstanceError(ctx, instanceID, reason); setErr != nil {
				e.logger.Error("failed to set error state after terminate failure",
					slog.String("instance_id", instanceID),
					slog.String("error", setErr.Error()),
				)
			}
			return fmt.Errorf("terminate: provider %s: %w", inst.UpstreamProvider, err)
		}
	}

	// 5. Mark as terminated in DB (sets terminated_at, billing_end).
	if _, err := e.db.TerminateInstance(ctx, instanceID); err != nil {
		return fmt.Errorf("terminate: update DB: %w", err)
	}

	// Billing stops at DELETE request time.
	if err := e.db.CloseBillingSession(ctx, instanceID, time.Now().UTC()); err != nil {
		e.logger.Error("failed to close billing session on termination",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		// Non-fatal: instance is terminated, log and continue.
	}

	// Log terminated event to instance_events table.
	metadata, _ := json.Marshal(map[string]string{
		"gpu_type": inst.GPUType,
		"region":   inst.Region,
		"tier":     inst.Tier,
	})
	terminatedEvent := &db.InstanceEvent{
		InstanceID: instanceID,
		OrgID:      inst.OrgID,
		EventType:  "terminated",
		Metadata:   metadata,
	}
	if err := e.db.CreateInstanceEvent(ctx, terminatedEvent); err != nil {
		e.logger.Error("failed to log terminated event",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		// Non-fatal: instance is terminated, event logging failure doesn't block.
	}

	// 6. Clean up WireGuard peer if configured.
	if e.wgMgr != nil && inst.WGPublicKey != nil && inst.WGAddress != nil {
		// WGAddress may contain a CIDR suffix (e.g., "10.0.0.2/16") from the INET column.
		// Strip the prefix length before parsing.
		addrStr, _, _ := strings.Cut(*inst.WGAddress, "/")
		tunnelIP := net.ParseIP(addrStr)
		if tunnelIP == nil {
			e.logger.Error("failed to parse WG address for cleanup",
				slog.String("instance_id", instanceID),
				slog.String("wg_address", *inst.WGAddress),
			)
		} else {
			externalPort := wireguard.PortFromTunnelIP(tunnelIP)
			if err := e.wgMgr.RemovePeer(ctx, *inst.WGPublicKey, tunnelIP, externalPort); err != nil {
				// Log but don't fail termination -- WG cleanup is best-effort.
				// The instance is already terminated in the DB and provider.
				e.logger.Error("WireGuard peer cleanup failed (best-effort)",
					slog.String("instance_id", instanceID),
					slog.String("error", err.Error()),
				)
			} else {
				e.logger.Info("WireGuard peer removed",
					slog.String("instance_id", instanceID),
					slog.String("tunnel_ip", tunnelIP.String()),
					slog.Int("external_port", externalPort),
				)
			}
		}
	}

	e.logger.Info("instance terminated",
		slog.String("instance_id", instanceID),
	)

	return nil
}

// StopInstancesForOrg transitions all running instances for an org to stopped state.
// "Stopped" preserves local storage but suspends billing. Does NOT call provider.Terminate.
// Used by the billing ticker when an org reaches their spending limit.
func (e *Engine) StopInstancesForOrg(ctx context.Context, orgID string) error {
	instances, err := e.db.ListRunningInstancesByOrg(ctx, orgID)
	if err != nil {
		return fmt.Errorf("list running instances for org %s: %w", orgID, err)
	}

	for _, inst := range instances {
		updated, err := e.db.UpdateInstanceStatus(ctx, inst.InstanceID, StateRunning, StateStopped)
		if err != nil {
			e.logger.Error("failed to stop instance for spending limit",
				slog.String("instance_id", inst.InstanceID),
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if updated && e.onStatusChange != nil {
			e.onStatusChange(inst.InstanceID, StateStopped)
		}
	}
	return nil
}

// TerminateStoppedInstancesForOrg terminates all stopped instances for an org.
// Used by the billing ticker 72 hours after a spending limit is reached.
func (e *Engine) TerminateStoppedInstancesForOrg(ctx context.Context, orgID string) error {
	instances, err := e.db.ListStoppedInstancesByOrg(ctx, orgID)
	if err != nil {
		return fmt.Errorf("list stopped instances for org %s: %w", orgID, err)
	}

	for _, inst := range instances {
		if err := e.Terminate(ctx, inst.InstanceID); err != nil {
			e.logger.Error("failed to terminate stopped instance for spending limit",
				slog.String("instance_id", inst.InstanceID),
				slog.String("org_id", orgID),
				slog.String("error", err.Error()),
			)
			// Continue terminating other instances.
		}
	}
	return nil
}

// checkSpendingLimit verifies an org is below their spending limit before creating
// a new instance. Returns nil if no limit is set or spend is under limit.
// Returns ErrSpendingLimitReached if the org is at or over their limit.
func (e *Engine) checkSpendingLimit(ctx context.Context, orgID string) error {
	limit, err := e.db.GetSpendingLimit(ctx, orgID)
	if errors.Is(err, db.ErrNotFound) {
		return nil // No limit set, allow.
	}
	if err != nil {
		return fmt.Errorf("check spending limit: %w", err)
	}
	if limit.LimitReachedAt != nil {
		return ErrSpendingLimitReached
	}
	// Also check live spend in case ticker hasn't run yet.
	spendCents, err := e.db.GetOrgMonthSpendCents(ctx, orgID, limit.BillingCycleStart)
	if err != nil {
		return fmt.Errorf("check spending limit: %w", err)
	}
	if spendCents >= limit.MonthlyLimitCents {
		return ErrSpendingLimitReached
	}
	return nil
}

// selectProvider finds the best-price provider for the request.
// Collects all matching offerings across providers, sorts by price ascending,
// and returns the cheapest match. Tiebreaker: registry iteration order provides
// the implicit tiebreak via sort.SliceStable (earlier-registered provider preferred).
func (e *Engine) selectProvider(ctx context.Context, req ProvisionRequest) (provider.Provider, *provider.GPUOffering, error) {
	candidates, err := e.selectProviderCandidates(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	return candidates[0].prov, &candidates[0].offering, nil
}

// selectProviderCandidates returns all matching providers sorted by price ascending.
// Used by selectProvider (returns first) and by Provision (iterates for fallback retry).
func (e *Engine) selectProviderCandidates(ctx context.Context, req ProvisionRequest) ([]providerCandidate, error) {
	providers := e.registry.All()
	if len(providers) == 0 {
		return nil, ErrNoProvider
	}

	var candidates []providerCandidate

	for _, prov := range providers {
		offerings, err := prov.ListAvailable(ctx)
		if err != nil {
			e.logger.Warn("provider availability check failed",
				slog.String("provider", prov.Name()),
				slog.String("error", err.Error()),
			)
			continue
		}

		for _, o := range offerings {
			if o.GPUType != req.GPUType {
				continue
			}
			if o.GPUCount < req.GPUCount {
				continue
			}
			if o.Tier != req.Tier {
				continue
			}
			if req.Region != "" && o.Region != req.Region {
				continue
			}
			if o.AvailableCount <= 0 {
				continue
			}
			candidates = append(candidates, providerCandidate{prov: prov, offering: o})
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoProvider
	}

	// Sort by price ascending. On equal price, registry iteration order provides
	// the implicit tiebreak (earlier-registered provider = higher priority).
	// Per CONTEXT.md: "higher-margin provider wins (existing priority list handles
	// this -- no new mechanism)". sort.SliceStable preserves registry insertion
	// order as the tiebreaker.
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].offering.PricePerHour < candidates[j].offering.PricePerHour
	})

	return candidates, nil
}

// progressStatus runs in a goroutine to drive the full provisioning lifecycle:
// creating -> provisioning -> (poll provider) -> booting.
// The ready callback from cloud-init handles the final booting -> running transition.
// Has a 10-minute timeout to prevent goroutine leaks.
func (e *Engine) progressStatus(instanceID string) {
	ctx := context.Background()

	// Step 1: Transition creating -> provisioning (immediate, as before).
	updated, err := e.db.UpdateInstanceStatus(ctx, instanceID, StateCreating, StateProvisioning)
	if err != nil {
		e.logger.Error("failed to progress to provisioning",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !updated {
		e.logger.Warn("status already changed from creating",
			slog.String("instance_id", instanceID),
		)
		return
	}

	// Notify SSE subscribers of provisioning state.
	if e.onStatusChange != nil {
		e.onStatusChange(instanceID, StateProvisioning)
	}

	// Step 2: Look up instance to get upstream_provider and upstream_id.
	inst, err := e.db.GetInstance(ctx, instanceID)
	if err != nil {
		e.logger.Error("failed to get instance for status polling",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return
	}

	// Step 3: Look up provider adapter from registry.
	prov, ok := e.registry.Get(inst.UpstreamProvider)
	if !ok {
		e.logger.Error("provider not found for status polling",
			slog.String("instance_id", instanceID),
			slog.String("provider", inst.UpstreamProvider),
		)
		return
	}

	// Step 4: Poll provider until running, terminal, or timeout.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-timeout:
			e.logger.Error("provisioning timeout",
				slog.String("instance_id", instanceID),
			)
			_ = e.db.SetInstanceError(ctx, instanceID,
				"provisioning timeout: instance did not start within 10 minutes")
			e.createZeroBillingSession(ctx, instanceID)
			return
		case <-ticker.C:
			// Check if instance has been concurrently terminated.
			current, err := e.db.GetInstance(ctx, instanceID)
			if err != nil {
				e.logger.Warn("failed to check instance status during polling",
					slog.String("instance_id", instanceID),
					slog.String("error", err.Error()),
				)
				continue
			}
			if current.Status != StateProvisioning {
				e.logger.Info("instance status changed during polling, exiting",
					slog.String("instance_id", instanceID),
					slog.String("status", current.Status),
				)
				return
			}

			// Poll upstream provider.
			status, err := prov.GetStatus(ctx, inst.UpstreamID)
			if err != nil {
				e.logger.Warn("status poll failed, will retry",
					slog.String("instance_id", instanceID),
					slog.String("error", err.Error()),
				)
				continue
			}

			e.logger.Debug("polled provider status",
				slog.String("instance_id", instanceID),
				slog.String("upstream_status", status.Status),
			)

			switch status.Status {
			case "running":
				// Provider reports pod is running -> transition to booting.
				// Cloud-init is now executing on the instance.
				updated, err := e.db.UpdateInstanceStatus(ctx, instanceID, StateProvisioning, StateBooting)
				if err != nil {
					e.logger.Error("failed to transition to booting",
						slog.String("instance_id", instanceID),
						slog.String("error", err.Error()),
					)
					return
				}
				if updated {
					e.logger.Info("instance transitioned to booting",
						slog.String("instance_id", instanceID),
					)
					if e.onStatusChange != nil {
						e.onStatusChange(instanceID, StateBooting)
					}

					// Billing starts at booting: provider has confirmed pod is allocated.
					session := &db.BillingSession{
						InstanceID:   instanceID,
						OrgID:        inst.OrgID,
						GPUType:      inst.GPUType,
						GPUCount:     inst.GPUCount,
						PricePerHour: inst.PricePerHour,
						StartedAt:    time.Now().UTC(),
					}
					if err := e.db.CreateBillingSession(ctx, session); err != nil {
						e.logger.Error("failed to create billing session at booting",
							slog.String("instance_id", instanceID),
							slog.String("error", err.Error()),
						)
						// Non-fatal: instance still transitions to booting. Billing gap is acceptable
						// vs preventing instance from reaching running state.
					}
				}
				return // Polling complete; ready callback handles booting->running.

			case "terminated", "error":
				e.logger.Error("upstream instance failed to start",
					slog.String("instance_id", instanceID),
					slog.String("upstream_status", status.Status),
				)
				_ = e.db.SetInstanceError(ctx, instanceID,
					fmt.Sprintf("upstream instance failed: %s", status.Status))
				e.createZeroBillingSession(ctx, instanceID)
				return
			}
			// For "creating" or any other status, continue polling.
		}
	}
}

// createZeroBillingSession creates a $0 audit billing session for a failed provision.
// Called when an instance never reached booting state (timeout or upstream failure).
// The session is created and immediately closed with the same timestamp, resulting in
// zero duration and zero cost.
func (e *Engine) createZeroBillingSession(ctx context.Context, instanceID string) {
	inst, err := e.db.GetInstance(ctx, instanceID)
	if err != nil {
		e.logger.Error("failed to get instance for zero billing session",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return
	}

	now := time.Now().UTC()
	session := &db.BillingSession{
		InstanceID:   instanceID,
		OrgID:        inst.OrgID,
		GPUType:      inst.GPUType,
		GPUCount:     inst.GPUCount,
		PricePerHour: inst.PricePerHour,
		StartedAt:    now,
	}
	if err := e.db.CreateBillingSession(ctx, session); err != nil {
		e.logger.Error("failed to create zero billing session",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return
	}

	if err := e.db.CloseBillingSession(ctx, instanceID, now); err != nil {
		e.logger.Error("failed to close zero billing session",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
	}
}

// buildCallbackURL constructs the ready callback URL for an instance.
// Uses GpuctlPublicURL when configured, falling back to branded hostname.
func buildCallbackURL(publicURL, hostname, instanceID string) string {
	if publicURL != "" {
		return strings.TrimRight(publicURL, "/") + "/internal/instances/" + instanceID + "/ready"
	}
	return "https://" + hostname + "/internal/instances/" + instanceID + "/ready"
}

// generateInstanceID produces a branded instance ID: "gpu-" + 4 random hex bytes.
// Result is 12 characters like "gpu-a1b2c3d4".
func generateInstanceID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "gpu-" + hex.EncodeToString(b), nil
}

// generateHexToken produces a random hex-encoded token of the specified byte length.
func generateHexToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
