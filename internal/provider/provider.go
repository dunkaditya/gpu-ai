// Package provider defines the interface all upstream GPU providers must implement.
package provider

import "context"

// Provider is the interface for upstream GPU cloud adapters.
// Each provider (RunPod, E2E Networks, Lambda, etc.) implements this interface.
type Provider interface {
	// Name returns the provider identifier (e.g., "runpod", "e2e_networks").
	Name() string

	// ListAvailable polls the provider for current GPU inventory and pricing.
	ListAvailable(ctx context.Context) ([]GPUOffering, error)

	// Provision creates a new GPU instance with the given configuration.
	// Returns immediately with an upstream ID; status polling happens separately
	// via GetStatus. The adapter is responsible for configuring the instance with
	// the appropriate startup mechanism (Docker image, startup scripts, environment variables).
	Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResult, error)

	// GetStatus returns the current status of an upstream instance.
	GetStatus(ctx context.Context, upstreamID string) (*InstanceStatus, error)

	// Terminate destroys an upstream instance. Returns nil on success.
	Terminate(ctx context.Context, upstreamID string) error
}
