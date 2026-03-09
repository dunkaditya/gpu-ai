# Phase 9: Replace WireGuard with FRP Tunneling - Research

**Researched:** 2026-03-09
**Domain:** Network tunneling / reverse proxy (WireGuard -> FRP migration)
**Confidence:** HIGH

## Summary

This phase replaces the existing WireGuard-based privacy layer with FRP (Fast Reverse Proxy) tunneling. The current WireGuard implementation requires kernel module loading or wireguard-go userspace, NET_ADMIN capabilities for iptables, and wgctrl-go for peer management -- all of which are problematic in Docker container environments like RunPod pods. FRP operates purely in userspace as a TCP reverse proxy, requiring no kernel modules, no NET_ADMIN, and no iptables rules. This makes it fundamentally more compatible with container-based GPU providers.

The architecture shifts from "proxy server manages WireGuard peers and iptables DNAT rules" to "proxy server runs frps, each GPU instance runs frpc which establishes an outbound TCP connection back to frps, exposing local SSH (port 22) on a unique remote port." The key advantage: frpc initiates the connection outward (NAT-friendly, no kernel privileges needed), while the current WireGuard approach requires the instance to set up a kernel-level tunnel interface and the proxy to manage iptables forwarding rules.

**Primary recommendation:** Run frps embedded in the gpuctl binary (the Go library supports this), and download the pre-built frpc binary in the bootstrap script on GPU instances. Use token-based authentication with per-instance unique tokens, and the frps server plugin (NewProxy hook) to validate proxy registrations against the database.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/fatedier/frp (server pkg) | v0.67.0 | Embed frps in gpuctl binary | Official Go library, Apache-2.0, can be embedded as a Go dependency |
| frpc binary (pre-built) | v0.67.0 | Downloaded in bootstrap script on GPU instances | Eliminates need for kernel modules; single static binary ~15MB |

### Replacing
| Current | Replaced By | Reason |
|---------|-------------|--------|
| golang.zx2c4.com/wireguard/wgctrl | github.com/fatedier/frp/server | wgctrl requires kernel WireGuard module or wireguard-go; frps is pure userspace TCP |
| iptables DNAT/FORWARD rules | frps built-in TCP port forwarding | No root/iptables required on proxy server for port mapping |
| wireguard-go (downloaded in bootstrap) | frpc binary (downloaded in bootstrap) | frpc is a single static binary, no TUN device needed, no CAP_NET_ADMIN |
| WireGuard key pairs (Curve25519) | FRP auth token (shared secret per instance) | Simpler auth model, no key generation/encryption overhead |
| IPAM (PostgreSQL advisory locks for /16 subnet) | Port allocation (assign unique remotePort per instance) | Ports map directly to SSH access; no virtual IP subnet needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Embedding frps in gpuctl | Running frps as a separate binary/service | Separate process adds deployment complexity; embedding keeps single-binary architecture |
| Downloading frpc in bootstrap | FRP SSH Tunnel Gateway (ssh -R, no frpc needed) | SSH gateway is simpler but less controllable; frpc gives us config-based proxy naming and metadata |
| Token auth | OIDC auth | OIDC is overkill for machine-to-machine; token is simpler and sufficient |
| FRP VirtualNet (TUN-based) | Standard TCP proxy | VirtualNet is Alpha/unstable, requires root/TUN, defeats the purpose of moving away from kernel-level networking |

### Dependencies to Add
```bash
go get github.com/fatedier/frp@v0.67.0
```

### Dependencies to Remove
```bash
# After migration is complete:
# - golang.zx2c4.com/wireguard/wgctrl (direct)
# - golang.zx2c4.com/wireguard (indirect)
# - github.com/mdlayher/genetlink (indirect)
# - github.com/mdlayher/netlink (indirect)
# - github.com/mdlayher/socket (indirect)
# - github.com/josharian/native (indirect)
```

## Architecture Patterns

### Current WireGuard Architecture (Being Replaced)
```
Customer SSH -> Proxy:10002 -> iptables DNAT -> wg0 tunnel -> Instance WG peer:22
                 (iptables)     (kernel/root)   (kernel module)  (kernel module)
```
- Proxy runs: wgctrl (Go library) + iptables commands (shell-out)
- Instance runs: wireguard-go or kernel WG + wg-quick (needs CAP_NET_ADMIN)
- Port mapping: iptables DNAT rule per instance
- Address space: 10.0.0.0/16 virtual subnet, IPAM via PostgreSQL

### New FRP Architecture
```
Customer SSH -> Proxy:10002 -> frps (embedded) -> frpc (on instance) -> localhost:22
                 (TCP listen)  (pure userspace)   (outbound TCP)       (local SSH)
```
- Proxy runs: frps embedded in gpuctl (Go library, pure userspace, no root for port mapping)
- Instance runs: frpc binary (downloaded in bootstrap, pure userspace, no privileges)
- Port mapping: frps assigns remotePort per proxy registration (built-in, no iptables)
- Address space: port range (10000-10255 or similar), stored in DB

### Recommended Project Structure Changes
```
internal/
├── wireguard/          # REMOVE entirely (or keep keygen.go temporarily for migration)
│   ├── manager.go      # DELETE - replaced by FRP manager
│   ├── ipam.go         # DELETE - replaced by port allocation
│   ├── keygen.go       # DELETE - no WG keys needed
│   ├── template.go     # DELETE - bootstrap template moves to tunnel/
│   ├── types.go        # DELETE
│   └── templates/
│       └── bootstrap.sh.tmpl  # DELETE - replaced by new template
├── tunnel/             # NEW package (replaces wireguard/)
│   ├── manager.go      # FRP server lifecycle + proxy validation
│   ├── ports.go        # Port allocation (replaces IPAM)
│   ├── template.go     # Bootstrap template renderer
│   ├── types.go        # FRP-related types
│   └── templates/
│       └── bootstrap.sh.tmpl  # New bootstrap script with frpc
```

### Pattern 1: Embedded frps Server
**What:** Run frps as part of the gpuctl process using the Go library
**When to use:** When you want single-binary deployment (matching project convention)
**Example:**
```go
// Source: pkg.go.dev/github.com/fatedier/frp/server
import (
    frpserver "github.com/fatedier/frp/server"
    v1 "github.com/fatedier/frp/pkg/config/v1"
)

func startFRPServer(ctx context.Context, bindPort int, token string) (*frpserver.Service, error) {
    cfg := &v1.ServerConfig{}
    cfg.BindPort = bindPort
    cfg.Auth.Method = v1.AuthMethodToken
    cfg.Auth.Token = token
    // Configure allowed ports
    cfg.AllowPorts, _ = types.NewPortsRangeSliceFromString("10000-10255")

    svc, err := frpserver.NewService(cfg)
    if err != nil {
        return nil, err
    }
    go svc.Run(ctx)
    return svc, nil
}
```

### Pattern 2: Bootstrap Script frpc Configuration
**What:** Bootstrap script downloads frpc and creates a TOML config
**When to use:** On every GPU instance during provisioning
**Example:**
```bash
# Download frpc binary
FRP_VERSION="0.67.0"
curl -sL "https://github.com/fatedier/frp/releases/download/v${FRP_VERSION}/frp_${FRP_VERSION}_linux_amd64.tar.gz" \
    | tar xzf - --strip-components=1 -C /usr/local/bin/ "frp_${FRP_VERSION}_linux_amd64/frpc"

# Write frpc config
cat > /etc/frpc.toml << 'FRPEOF'
serverAddr = "{{.ProxyHost}}"
serverPort = {{.FRPServerPort}}
auth.method = "token"
auth.token = "{{.FRPToken}}"

[[proxies]]
name = "{{.InstanceID}}-ssh"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = {{.RemotePort}}
FRPEOF

# Start frpc in background
frpc -c /etc/frpc.toml &
```

### Pattern 3: Server Plugin for Proxy Validation
**What:** Use frps httpPlugin to validate proxy registrations against DB
**When to use:** To ensure only authorized instances can register proxies
**Example:**
```go
// HTTP handler in gpuctl that frps calls on NewProxy
func (s *Server) handleFRPNewProxy(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Content struct {
            ProxyName  string            `json:"proxy_name"`
            RemotePort int               `json:"remote_port"`
            User       struct {
                Metas map[string]string `json:"metas"`
            } `json:"user"`
        } `json:"content"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Validate instance exists and port matches
    instanceID := req.Content.User.Metas["instance_id"]
    inst, err := s.db.GetInstance(r.Context(), instanceID)
    if err != nil || inst.Status == "terminated" {
        json.NewEncoder(w).Encode(map[string]any{
            "reject": true,
            "reject_reason": "unknown instance",
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]any{
        "reject": false,
        "unchange": true,
    })
}
```

### Anti-Patterns to Avoid
- **Using FRP VirtualNet:** Alpha feature, requires TUN/root, defeats the purpose of removing kernel-level networking
- **Running frps as a separate process:** Breaks single-binary deployment convention; complicates lifecycle management
- **Per-instance frps tokens as the ONLY auth:** Use server plugin validation too, so rogue frpc connections with leaked tokens are rejected if the instance is terminated
- **Keeping WireGuard as fallback:** Creates two parallel tunnel systems; complexity multiplies. Clean cut-over is better.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| TCP port forwarding | Custom TCP proxy with goroutine-per-conn | frps built-in TCP proxy | Handles connection pooling, keepalive, reconnection, backpressure |
| Reverse tunnel establishment | Custom protocol for instance-to-proxy tunnel | frpc -> frps connection | Handles NAT traversal, reconnection, multiplexing |
| Port allocation | Complex IPAM with advisory locks | Simple sequential port counter in DB | Only need ~255 ports, not 65k IPs; a simple `SELECT MAX(frp_port)` suffices |
| Proxy validation | Custom auth protocol | frps server plugin (NewProxy HTTP hook) | Built-in, well-tested, supports reject/modify responses |
| Binary distribution to instances | Build frpc from source on each instance | Download pre-built release binary | Faster, deterministic, no Go toolchain needed on instances |

**Key insight:** FRP already solves the exact problem (expose local service through a proxy) with a mature, battle-tested implementation. The migration is about replacing a lower-level (L3/kernel) solution with an application-level (L4/userspace) one that works in containers.

## Common Pitfalls

### Pitfall 1: frps Port Conflicts with gpuctl HTTP Server
**What goes wrong:** frps wants to bind its own port (e.g., 7000) for client connections, which may conflict with gpuctl's port or other services.
**Why it happens:** frps and gpuctl are in the same process, both binding TCP ports.
**How to avoid:** Use a dedicated port for frps (e.g., 7000) separate from gpuctl HTTP (9090). Configure in .env as FRP_BIND_PORT.
**Warning signs:** "address already in use" errors on startup.

### Pitfall 2: Token Leakage in Bootstrap Script
**What goes wrong:** The FRP auth token is embedded in the bootstrap script which is passed as dockerArgs (base64-encoded). If the token is the global frps token, a compromised instance could register arbitrary proxies.
**Why it happens:** Reusing a single shared token for all instances.
**How to avoid:** Use per-instance tokens. Generate a unique token for each instance, store it in the DB, and validate via the server plugin. Alternatively, use the existing `internal_token` field as the FRP auth mechanism (the instance already has this for the ready callback).
**Warning signs:** One instance's frpc connecting with another instance's proxy name.

### Pitfall 3: frpc Download Fails in Bootstrap
**What goes wrong:** GitHub release download fails due to network restrictions, rate limits, or DNS issues in the container environment.
**Why it happens:** RunPod containers may have restricted egress or GitHub CDN issues.
**How to avoid:** Host the frpc binary on your own infrastructure (e.g., the gpuctl public URL) as a fallback. The bootstrap script should try GitHub first, fall back to self-hosted.
**Warning signs:** "curl: (6) Could not resolve host" or 403 errors in bootstrap logs.

### Pitfall 4: Port Exhaustion / Collision
**What goes wrong:** Two instances get assigned the same remote port, or the port range is exhausted.
**Why it happens:** Concurrent provisioning without proper locking.
**How to avoid:** Use a DB-backed port allocator with a unique constraint on the port column. Reclaim ports from terminated instances. Use frps `allowPorts` to whitelist the valid range.
**Warning signs:** frps rejects proxy registration with "port already in use" error.

### Pitfall 5: Stale frpc Connections After Instance Termination
**What goes wrong:** An frpc client stays connected to frps after the instance is logically terminated, holding the port.
**Why it happens:** Termination only calls the provider's Terminate API; frpc may linger briefly.
**How to avoid:** The server plugin's CloseProxy hook can clean up DB state. Also, frps will naturally disconnect when the instance is destroyed by the provider. Port should be freed in DB during Terminate(), not on frpc disconnect.
**Warning signs:** Port shows as allocated in DB but no active frpc connection.

### Pitfall 6: Database Migration Breaks Running Instances
**What goes wrong:** Adding FRP columns and removing WG columns in one migration breaks any running instances that still use WG.
**Why it happens:** Column removal while instances reference those columns.
**How to avoid:** Two-phase migration: (1) ADD frp columns alongside WG columns, (2) only DROP WG columns after verifying no running instances use them. For this dev project, a single migration is fine since there are no production instances.
**Warning signs:** NOT NULL constraint violations on new columns for old instances.

## Code Examples

### Example 1: New Config Fields
```go
// internal/config/config.go additions
type Config struct {
    // ... existing fields ...

    // FRP tunneling (replaces WG_* fields)
    FRPBindPort    int    // Port for frps client connections. Default: 7000
    FRPToken       string // Shared auth token for frps<->frpc. Required when FRP enabled.
    FRPAllowPorts  string // Allowed remote port range. Default: "10000-10255"
}
```

### Example 2: New Bootstrap Template Data
```go
// internal/tunnel/types.go
type BootstrapData struct {
    InstanceID        string // e.g., "gpu-4a7f"
    ProxyHost         string // e.g., "134.199.214.138" (proxy server public IP)
    FRPServerPort     int    // e.g., 7000
    FRPToken          string // per-instance auth token
    RemotePort        int    // e.g., 10002 (unique port for this instance)
    SSHAuthorizedKeys string // newline-separated SSH public keys
    InternalToken     string // per-instance callback auth token
    Hostname          string // e.g., "gpu-4a7f.gpu.ai"
    CallbackURL       string // full URL for ready callback
}
```

### Example 3: Port Allocator (Replaces IPAM)
```go
// internal/tunnel/ports.go
const (
    MinPort = 10000
    MaxPort = 10255
)

func AllocatePort(ctx context.Context, tx pgx.Tx) (int, error) {
    // Advisory lock to serialize allocations (same pattern as IPAM)
    _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", portLockID)
    if err != nil {
        return 0, fmt.Errorf("tunnel: acquire port lock: %w", err)
    }

    var maxPort int
    err = tx.QueryRow(ctx,
        "SELECT COALESCE(MAX(frp_remote_port), $1) FROM instances WHERE frp_remote_port IS NOT NULL AND status != 'terminated'",
        MinPort-1,
    ).Scan(&maxPort)
    if err != nil {
        return 0, err
    }

    next := maxPort + 1
    if next > MaxPort {
        return 0, fmt.Errorf("tunnel: port range exhausted (%d-%d)", MinPort, MaxPort)
    }
    return next, nil
}
```

### Example 4: FRP Manager (Replaces WireGuard Manager)
```go
// internal/tunnel/manager.go
type Manager struct {
    frpSvc *frpserver.Service
    logger *slog.Logger
}

func NewManager(bindPort int, token string, allowPorts string, logger *slog.Logger) (*Manager, error) {
    cfg := &v1.ServerConfig{}
    cfg.BindPort = bindPort
    cfg.Auth.Method = v1.AuthMethodToken
    cfg.Auth.Token = token
    // Set allowed ports range
    // Configure server plugin for proxy validation

    svc, err := frpserver.NewService(cfg)
    if err != nil {
        return nil, fmt.Errorf("tunnel: create frp service: %w", err)
    }
    return &Manager{frpSvc: svc, logger: logger}, nil
}

func (m *Manager) Start(ctx context.Context) {
    m.frpSvc.Run(ctx)
}

func (m *Manager) Close() error {
    return m.frpSvc.Close()
}
```

### Example 5: Provisioning Engine Changes
```go
// In engine.Provision(), replace WireGuard block with:
if e.tunnelMgr != nil {
    // Allocate a unique remote port for this instance
    tx, err := e.db.PgxPool().Begin(ctx)
    remotePort, err := tunnel.AllocatePort(ctx, tx)
    tx.Commit(ctx)

    frpPort := &remotePort

    // Render bootstrap script with FRP config
    bootstrapData := tunnel.BootstrapData{
        InstanceID:        instanceID,
        ProxyHost:         e.config.FRPProxyHost,
        FRPServerPort:     e.config.FRPBindPort,
        FRPToken:          internalToken, // reuse instance token
        RemotePort:        remotePort,
        SSHAuthorizedKeys: strings.Join(sshPubKeys, "\n"),
        InternalToken:     internalToken,
        Hostname:          hostname,
        CallbackURL:       callbackURL,
    }
    startupScript, err = tunnel.RenderBootstrap(bootstrapData)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| WireGuard kernel module | FRP TCP proxy (userspace) | This phase | No kernel modules needed; works in unprivileged containers |
| iptables DNAT rules | frps built-in port mapping | This phase | No root/iptables on proxy server for SSH routing |
| wgctrl-go library | frp Go server library | This phase | Removes 6 indirect deps (genetlink, netlink, socket, native, wireguard) |
| IPAM /16 subnet | Simple port allocator | This phase | 256 ports is plenty for dev; simpler than IP allocation |
| WG key pair generation + AES encryption | FRP token auth | This phase | No crypto key management needed |

**Deprecated/outdated after this phase:**
- `WG_ENCRYPTION_KEY`, `WG_PROXY_ENDPOINT`, `WG_PROXY_PUBLIC_KEY`, `WG_INTERFACE_NAME` env vars
- `wg_public_key`, `wg_private_key_enc`, `wg_address` DB columns (nullable, can be left for historical records)
- `internal/wireguard/` package (entirely removed)
- WireGuard installation in bootstrap script

## DB Schema Changes

### New Columns on `instances` Table
```sql
ALTER TABLE instances ADD COLUMN frp_remote_port INTEGER;
ALTER TABLE instances ADD CONSTRAINT instances_frp_remote_port_unique
    UNIQUE (frp_remote_port)
    WHERE frp_remote_port IS NOT NULL AND status NOT IN ('terminated', 'error');
```

### Columns That Become Unused
```sql
-- These columns are no longer populated for new instances:
-- wg_public_key, wg_private_key_enc, wg_address
-- Keep them nullable (already nullable) for backward compat with old records.
-- Do NOT drop them in this migration.
```

## Scope of Changes

### Files to Create
| File | Purpose |
|------|---------|
| `internal/tunnel/manager.go` | FRP server lifecycle, startup, shutdown |
| `internal/tunnel/ports.go` | Port allocator (replaces IPAM) |
| `internal/tunnel/template.go` | Bootstrap template renderer |
| `internal/tunnel/types.go` | BootstrapData, config types |
| `internal/tunnel/templates/bootstrap.sh.tmpl` | New bootstrap script with frpc |
| `database/migrations/20260309_v7_frp_tunneling.sql` | Add frp_remote_port column |

### Files to Modify
| File | Changes |
|------|---------|
| `internal/config/config.go` | Add FRP_* env vars, remove WG_* validation (keep loading for backward compat) |
| `internal/provision/engine.go` | Replace WireGuard block with FRP tunnel setup |
| `internal/api/handlers.go` | Update `instanceToResponse` to derive SSH command from frp_remote_port |
| `internal/db/instances.go` | Add frp_remote_port to Instance struct and scan list |
| `cmd/gpuctl/deps.go` | Replace WG manager/IPAM init with FRP manager init |
| `cmd/gpuctl/serve.go` | Start/stop FRP server alongside HTTP server |
| `go.mod` | Add frp dependency, remove wgctrl |
| `provider/types.go` | Remove WireGuardAddress from ProvisionRequest (or leave unused) |

### Files to Delete
| File | Reason |
|------|--------|
| `internal/wireguard/manager.go` | Replaced by tunnel/manager.go |
| `internal/wireguard/manager_test.go` | Tests for removed code |
| `internal/wireguard/ipam.go` | Replaced by tunnel/ports.go |
| `internal/wireguard/ipam_test.go` | Tests for removed code |
| `internal/wireguard/keygen.go` | No WG key generation needed |
| `internal/wireguard/keygen_test.go` | Tests for removed code |
| `internal/wireguard/template.go` | Replaced by tunnel/template.go |
| `internal/wireguard/template_test.go` | Tests for removed code |
| `internal/wireguard/types.go` | KeyPair type no longer needed |
| `internal/wireguard/templates/bootstrap.sh.tmpl` | Replaced by tunnel/templates/bootstrap.sh.tmpl |

## Open Questions

1. **frps Embedding Stability**
   - What we know: The Go library API is documented and v0.67.0 is latest. The Service struct has Run/Close methods.
   - What's unclear: Whether embedding frps in a long-running Go process is well-tested by the frp community (most users run it standalone).
   - Recommendation: Proceed with embedding; fallback plan is to run frps as a sidecar process if embedding causes issues.

2. **Port Range Size**
   - What we know: Current IPAM supports /16 (65k addresses). The port range 10000-10255 gives 256 ports.
   - What's unclear: Whether 256 ports is enough for future growth.
   - Recommendation: Start with 10000-10255 (matching current formula), can expand to 10000-65535 later. For v1/dev milestone, 256 is more than sufficient.

3. **frpc Binary Size and Download Speed**
   - What we know: frpc is a Go binary, probably ~15-20MB compressed. GitHub releases are CDN-backed.
   - What's unclear: Exact download time in RunPod container environment.
   - Recommendation: Measure in practice. If slow, host frpc binary on the proxy server itself and download from GpuctlPublicURL.

4. **Server Plugin vs. Embedded Validation**
   - What we know: frps supports HTTP plugins for NewProxy validation. Alternative: since frps is embedded, we might be able to use the Go API directly.
   - What's unclear: Whether the embedded server library exposes hooks for proxy validation without the HTTP plugin mechanism.
   - Recommendation: Start with HTTP plugin (well-documented, proven pattern). The plugin endpoint runs inside gpuctl itself at a loopback address.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None -- Go convention |
| Quick run command | `go test ./internal/tunnel/...` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FRP-01 | FRP manager starts and accepts connections | unit | `go test ./internal/tunnel/ -run TestManagerStart -x` | Wave 0 |
| FRP-02 | Port allocator assigns unique ports | unit | `go test ./internal/tunnel/ -run TestAllocatePort -x` | Wave 0 |
| FRP-03 | Bootstrap template renders with frpc config | unit | `go test ./internal/tunnel/ -run TestRenderBootstrap -x` | Wave 0 |
| FRP-04 | Provisioning engine uses FRP instead of WG | unit | `go test ./internal/provision/ -run TestProvision -x` | Wave 0 |
| FRP-05 | SSH command uses proxy host + frp remote port | unit | `go test ./internal/api/ -run TestInstanceToResponse -x` | Wave 0 |
| FRP-06 | Config loads FRP env vars correctly | unit | `go test ./internal/config/ -run TestLoad -x` | Wave 0 |
| FRP-07 | WireGuard package removed, build succeeds | build | `go build ./...` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./internal/tunnel/... ./internal/provision/... ./internal/api/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green + `go build ./...` succeeds without wireguard imports

### Wave 0 Gaps
- [ ] `internal/tunnel/manager_test.go` -- covers FRP-01
- [ ] `internal/tunnel/ports_test.go` -- covers FRP-02
- [ ] `internal/tunnel/template_test.go` -- covers FRP-03

## Sources

### Primary (HIGH confidence)
- [pkg.go.dev/github.com/fatedier/frp/client](https://pkg.go.dev/github.com/fatedier/frp/client) - Go client API, ServiceOptions, dynamic proxy management
- [pkg.go.dev/github.com/fatedier/frp/server](https://pkg.go.dev/github.com/fatedier/frp/server) - Go server API, NewService, Run, HandleListener
- [gofrp.org/en/docs/examples/ssh/](https://gofrp.org/en/docs/examples/ssh/) - SSH tunneling TCP proxy example with frpc/frps
- [gofrp.org/en/docs/features/common/authentication/](https://gofrp.org/en/docs/features/common/authentication/) - Token and OIDC auth documentation
- [gofrp.org/en/docs/features/common/server-plugin/](https://gofrp.org/en/docs/features/common/server-plugin/) - Server plugin NewProxy hook for proxy validation
- [gofrp.org/en/docs/features/common/server-manage/](https://gofrp.org/en/docs/features/common/server-manage/) - allowPorts, bandwidth limiting
- [gofrp.org/en/docs/features/common/range/](https://gofrp.org/en/docs/features/common/range/) - Port range mapping templates

### Secondary (MEDIUM confidence)
- [github.com/fatedier/frp releases](https://github.com/fatedier/frp/releases) - v0.67.0 is latest (2026-01-31)
- [gofrp.org/en/docs/features/common/ssh/](https://gofrp.org/en/docs/features/common/ssh/) - SSH Tunnel Gateway (alternative to frpc)
- [gofrp.org/en/docs/features/common/virtualnet/](https://gofrp.org/en/docs/features/common/virtualnet/) - VirtualNet Alpha (confirmed unsuitable)

### Tertiary (LOW confidence)
- frpc binary size estimates (~15-20MB) - needs empirical verification

### Codebase Analysis (HIGH confidence)
- `internal/wireguard/` -- Full review of all 5 Go files + bootstrap template
- `internal/provision/engine.go` -- Provisioning flow with WG integration points
- `internal/api/handlers.go` -- SSH command generation from WG address
- `internal/config/config.go` -- WG environment variable loading
- `internal/db/instances.go` -- Instance struct with WG columns
- `cmd/gpuctl/deps.go` -- WG manager/IPAM initialization
- `database/migrations/` -- Schema for wg_address, wg_public_key columns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - frp is mature (88k+ GitHub stars), well-documented, Go-native
- Architecture: HIGH - TCP reverse proxy is a well-understood pattern; frp's Go library API is documented
- Pitfalls: HIGH - Identified from direct codebase analysis and understanding of container limitations
- Migration scope: HIGH - Complete codebase audit done; all WG touchpoints identified

**Research date:** 2026-03-09
**Valid until:** 2026-04-09 (frp is stable, not fast-moving)
