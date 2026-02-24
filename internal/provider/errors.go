package provider

import "errors"

// Sentinel errors for provider failure modes.
// Callers use errors.Is() to distinguish failure types.
var (
	// ErrNoCapacity indicates the provider has no GPU capacity available.
	// The provisioning engine uses this to try the next provider.
	ErrNoCapacity = errors.New("no GPU capacity available")

	// ErrProviderUnavailable indicates the provider API is unreachable or returning errors.
	ErrProviderUnavailable = errors.New("provider API unavailable")

	// ErrInvalidGPUType indicates the requested GPU type is not supported by the provider.
	ErrInvalidGPUType = errors.New("unsupported GPU type")
)
