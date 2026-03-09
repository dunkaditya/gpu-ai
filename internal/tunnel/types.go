// Package tunnel implements FRP-based SSH tunneling for GPU instances.
// It replaces the internal/wireguard package with a userspace TCP reverse
// proxy that works in unprivileged containers (e.g., RunPod pods).
package tunnel

// BootstrapData holds all parameters for rendering the instance bootstrap script.
// The bootstrap script downloads frpc, configures SSH, and establishes a
// reverse tunnel back to the FRP server on the proxy.
type BootstrapData struct {
	InstanceID        string // e.g., "gpu-4a7f"
	ProxyHost         string // e.g., "134.199.214.138" (proxy server public IP)
	FRPServerPort     int    // e.g., 7000
	FRPToken          string // per-instance auth token
	RemotePort        int    // e.g., 10002 (unique port for this instance)
	SSHAuthorizedKeys string // newline-separated SSH public keys
	InternalToken     string // per-instance callback auth token
	Hostname          string // e.g., "gpu-4a7f.gpu.ai"
	CallbackURL       string // full URL: "https://api.gpu.ai/internal/instances/{id}/ready"
}

// Port range constants for FRP remote port allocation.
const (
	MinPort = 10000
	MaxPort = 10255
)

// portLockID is a constant used for the PostgreSQL advisory lock
// to serialize port allocations. Represents "FRPPT" as a numeric constant.
const portLockID int64 = 0x4650525054
