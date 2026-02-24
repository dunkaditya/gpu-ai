// Package provision orchestrates the instance provisioning flow.
package provision

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gpuai/gpuctl/internal/config"
	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
	"github.com/gpuai/gpuctl/internal/wireguard"
)

// ProvisionRequest contains the parameters for provisioning a new instance.
// This is the engine-level request, distinct from provider.ProvisionRequest.
type ProvisionRequest struct {
	OrgID           string
	UserID          string
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
}

// Engine orchestrates instance provisioning and termination.
// It coordinates provider adapters, WireGuard, cloud-init, and DB persistence.
type Engine struct {
	registry *provider.Registry
	db       *db.Pool
	config   *config.Config
	logger   *slog.Logger
	wgMgr    *wireguard.Manager
	ipam     *wireguard.IPAM
}

// NewEngine creates a new provisioning engine with the given dependencies.
func NewEngine(deps EngineDeps) *Engine {
	return &Engine{
		registry: deps.Registry,
		db:       deps.DB,
		config:   deps.Config,
		logger:   deps.Logger,
		wgMgr:    deps.WGManager,
		ipam:     deps.IPAM,
	}
}

// ErrNoProvider is returned when no suitable provider is found for the request.
var ErrNoProvider = errors.New("no suitable provider available")

// ErrPriceExceeded is returned when the current price exceeds the requested max.
var ErrPriceExceeded = errors.New("current price exceeds maximum price per hour")

// ErrSSHKeysNotFound is returned when none of the requested SSH keys exist.
var ErrSSHKeysNotFound = errors.New("ssh keys not found")

// Provision creates a new GPU instance.
//
// Steps:
//  1. Generate instance ID and internal token
//  2. Resolve SSH keys from DB
//  3. Select provider and verify pricing
//  4. Optionally generate WireGuard keys and render cloud-init
//  5. Call provider to create upstream instance
//  6. Persist instance record to DB
//  7. Kick off async status progression
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

	// 2. Look up SSH keys from DB by IDs.
	sshKeys, err := e.db.GetSSHKeysByIDs(ctx, req.SSHKeyIDs)
	if err != nil {
		return nil, fmt.Errorf("provision: look up ssh keys: %w", err)
	}
	if len(sshKeys) == 0 {
		return nil, ErrSSHKeysNotFound
	}

	// Collect public keys.
	sshPubKeys := make([]string, 0, len(sshKeys))
	for _, k := range sshKeys {
		sshPubKeys = append(sshPubKeys, k.PublicKey)
	}

	// 3. Select provider from registry.
	prov, offering, err := e.selectProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Check price cap.
	if req.MaxPricePerHour != nil && offering.PricePerHour > *req.MaxPricePerHour {
		return nil, ErrPriceExceeded
	}

	// 5. Optionally generate WireGuard keys and cloud-init.
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

		// Build callback URL.
		callbackURL := fmt.Sprintf("https://%s/internal/instances/%s/ready", hostname, instanceID)

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

	// 6. Build provider-level request and call provider.
	provReq := provider.ProvisionRequest{
		InstanceID:       instanceID,
		GPUType:          req.GPUType,
		GPUCount:         req.GPUCount,
		Tier:             req.Tier,
		Region:           req.Region,
		SSHPublicKeys:    sshPubKeys,
		InternalToken:    internalToken,
		CallbackURL:      fmt.Sprintf("https://%s/internal/instances/%s/ready", hostname, instanceID),
		StartupScript:    startupScript,
	}
	if wgAddress != nil {
		provReq.WireGuardAddress = *wgAddress
	}

	provResult, err := prov.Provision(ctx, provReq)
	if err != nil {
		return nil, fmt.Errorf("provision: provider %s: %w", prov.Name(), err)
	}

	// 7. Build and persist instance record.
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

	// 8. Kick off async status progression: creating -> provisioning.
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

	// 6. Clean up WireGuard peer if configured.
	if e.wgMgr != nil && inst.WGPublicKey != nil && inst.WGAddress != nil {
		e.logger.Info("WireGuard cleanup would happen here",
			slog.String("instance_id", instanceID),
		)
		// Full WG cleanup (RemovePeer) requires parsing tunnel IP and computing port.
		// Deferred to handler integration in Plan 03 where we have full context.
	}

	e.logger.Info("instance terminated",
		slog.String("instance_id", instanceID),
	)

	return nil
}

// selectProvider finds the best available provider for the request.
// For Phase 4, it uses the first provider that has availability. Best-price
// selection across providers is deferred to Phase 6.
func (e *Engine) selectProvider(ctx context.Context, req ProvisionRequest) (provider.Provider, *provider.GPUOffering, error) {
	providers := e.registry.All()
	if len(providers) == 0 {
		return nil, nil, ErrNoProvider
	}

	for _, prov := range providers {
		offerings, err := prov.ListAvailable(ctx)
		if err != nil {
			e.logger.Warn("provider availability check failed",
				slog.String("provider", prov.Name()),
				slog.String("error", err.Error()),
			)
			continue
		}

		for i := range offerings {
			o := &offerings[i]
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
			return prov, o, nil
		}
	}

	return nil, nil, ErrNoProvider
}

// progressStatus runs in a goroutine to transition instance from creating to provisioning.
// The ready callback from the instance will handle further transitions.
func (e *Engine) progressStatus(instanceID string) {
	ctx := context.Background()

	updated, err := e.db.UpdateInstanceStatus(ctx, instanceID, StateCreating, StateProvisioning)
	if err != nil {
		e.logger.Error("failed to progress status to provisioning",
			slog.String("instance_id", instanceID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !updated {
		e.logger.Warn("status already changed from creating",
			slog.String("instance_id", instanceID),
		)
	}
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
