# Architecture Research

**Domain:** GPU cloud aggregation platform (multi-provider GPU provisioning with privacy layer)
**Researched:** 2026-02-24
**Confidence:** HIGH

## Standard Architecture

### System Overview

GPU.ai is a **single Go binary monolith** (`gpuctl`) that orchestrates GPU provisioning across upstream cloud providers, hides provider identity behind a WireGuard privacy layer, and exposes a unified REST API consumed by a Next.js dashboard. This is the correct architecture for Phase 1 -- a modular monolith in Go is the right call for a team shipping a first product. The binary contains all subsystems as internal packages communicating via function calls, with background goroutines for polling and health monitoring.

```
                         ┌──────────────────────────────┐
                         │    Next.js Dashboard          │
                         │    (Static export + API       │
                         │     calls to gpuctl)          │
                         └──────────────┬───────────────┘
                                        │ HTTPS (REST JSON)
                                        v
┌─────────────────────────────────────────────────────────────────────┐
│                       gpuctl (single Go binary)                      │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐    │
│  │                    HTTP Layer (stdlib net/http)                │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐  │    │
│  │  │ Auth MW  │  │ CORS MW  │  │ Log MW   │  │ Recovery MW │  │    │
│  │  └──────────┘  └──────────┘  └──────────┘  └─────────────┘  │    │
│  │  ┌──────────────────────────────────────────────────────────┐│    │
│  │  │  Route Handlers (instances, gpu, ssh-keys, billing)     ││    │
│  │  └──────────────────────────────────────────────────────────┘│    │
│  └──────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │                    Business Logic Layer                          │  │
│  │                                                                  │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │  │
│  │  │ Provisioning │  │ Availability │  │   Health Monitor     │  │  │
│  │  │   Engine     │  │   Poller     │  │   (60s goroutine)    │  │  │
│  │  │              │  │ (30s ticker) │  │                      │  │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │  │
│  │         │                 │                      │              │  │
│  │  ┌──────┴──────┐   ┌─────┴──────┐   ┌──────────┴───────────┐  │  │
│  │  │  Provider   │   │   Redis    │   │   WireGuard          │  │  │
│  │  │  Registry   │   │   Cache    │   │   Manager (wgctrl)   │  │  │
│  │  │  + Adapters │   │            │   │                      │  │  │
│  │  └──────┬──────┘   └────────────┘   └──────────────────────┘  │  │
│  │         │                                                      │  │
│  │  ┌──────┴──────────────────────┐                               │  │
│  │  │  Provider Adapters          │                               │  │
│  │  │  ┌─────────┐  ┌──────────┐ │                               │  │
│  │  │  │ RunPod  │  │  E2E     │ │                               │  │
│  │  │  │ (GQL)   │  │  (REST)  │ │                               │  │
│  │  │  └─────────┘  └──────────┘ │                               │  │
│  │  └─────────────────────────────┘                               │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │                    Data Layer                                    │  │
│  │  ┌──────────────────┐    ┌──────────────────────────────────┐  │  │
│  │  │  PostgreSQL      │    │  Redis                            │  │  │
│  │  │  (pgx pool)      │    │  (go-redis)                       │  │  │
│  │  │  instances, orgs │    │  GPU offerings cache (60s TTL)    │  │  │
│  │  │  users, billing  │    │                                    │  │  │
│  │  └──────────────────┘    └──────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │                    External Integrations                         │  │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────────────────────┐    │  │
│  │  │ Clerk    │  │ Stripe   │  │ Cloud-Init Template       │    │  │
│  │  │ (JWT)    │  │ (Billing)│  │ (text/template rendered)  │    │  │
│  │  └──────────┘  └──────────┘  └───────────────────────────┘    │  │
│  └────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
         │                           │
         │ WireGuard (wgctrl)        │ Upstream APIs (HTTPS)
         v                           v
┌─────────────────────┐    ┌─────────────────────────┐
│  GPU.ai Proxy       │    │  RunPod / E2E / Lambda  │
│  Server (wg0)       │    │  upstream APIs          │
│  10.0.0.1/24        │    └─────────────────────────┘
└────────┬────────────┘
         │ WireGuard tunnels (encrypted)
         v
┌──────────────────────────────────────┐
│  Upstream GPU Instances              │
│  (cloud-init bootstrapped:           │
│   WireGuard tunnel + SSH + firewall) │
│  10.0.0.x/32                         │
└──────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **cmd/gpuctl/main.go** | Dependency wiring, signal handling, server + goroutine lifecycle | Single entrypoint: init config, create deps, start HTTP server + background goroutines, block on signal, graceful shutdown |
| **internal/api/** | HTTP routing, middleware chain, request/response serialization | Go 1.22+ `net/http.ServeMux` with `METHOD /path/{param}` patterns. Middleware: auth, CORS, logging, recovery. Handlers call business logic, strip upstream fields from responses |
| **internal/provider/** | Provider interface + adapter registry | `Provider` interface with 5 methods. `Registry` holds `map[string]Provider`. Each adapter (RunPod, E2E) implements the interface, translating between GPU.ai types and upstream API formats |
| **internal/provider/runpod/** | RunPod-specific API client | GraphQL client using `podFindAndDeployOnDemand` mutation, `gpuTypes` query for availability, status mapping from RunPod states to GPU.ai states |
| **internal/provision/** | Provisioning orchestration (the "brain") | Receives create request, queries availability cache, selects provider, generates WireGuard keypair, renders cloud-init template, calls adapter, writes DB record, registers WG peer |
| **internal/availability/** | Real-time GPU inventory caching | 30s ticker goroutine polls all adapters concurrently via `ListAvailable()`, writes each offering to Redis with 60s TTL. API handler reads from Redis, filters, strips provider field |
| **internal/wireguard/** | WireGuard peer lifecycle on proxy server | Uses `wgctrl-go` library for programmatic peer add/remove (no shell exec). Manages IP allocation from 10.0.0.0/24 subnet. Persists config for reboot survival |
| **internal/billing/** | Usage metering and payment processing | Stripe SDK: start/stop metering on instance create/terminate, per-second usage records, webhook handler for payment events |
| **internal/auth/** | JWT verification and user context | Clerk JWKS-based JWT verification middleware. Extracts user_id, org_id, role into request context. Internal routes use bearer token instead |
| **internal/db/** | PostgreSQL connection pool and queries | pgx pool (min 5, max 20 conns). Raw SQL queries (no ORM). Org-scoped access patterns on all queries |
| **internal/health/** | Instance liveness monitoring | 60s ticker goroutine checks WireGuard peer handshake timestamps. Marks instances unhealthy if last handshake > 5 minutes ago |
| **internal/config/** | Environment variable loading | Typed struct populated from env vars. Validates required fields at startup. Fails fast on missing config |
| **infra/cloud-init/** | Instance bootstrap template | Go `text/template` rendered with WG keys, SSH keys, instance ID. Installs WireGuard, configures tunnel, sets hostname/MOTD, locks firewall to WG-only traffic |
| **frontend/** | Customer-facing dashboard | Next.js app calling gpuctl REST API. Pages: landing, auth, instance list/create, GPU availability, SSH keys, billing |

## Recommended Project Structure

The existing project structure in `docs/ARCHITECTURE.md` is well-designed. Here is the structure with rationale annotations.

```
gpu-ai/
├── cmd/gpuctl/
│   └── main.go                  # Single entrypoint. Wires ALL deps, starts server + goroutines
│
├── internal/                    # Go internal packages (not importable by external projects)
│   ├── api/
│   │   ├── server.go            # ServeMux setup, middleware chain, Handler() method
│   │   ├── router.go            # Route registration (separates route defs from handler logic)
│   │   ├── middleware.go        # Request ID, structured logging, CORS, panic recovery
│   │   └── handlers/
│   │       ├── instances.go     # CRUD handlers for /api/v1/instances
│   │       ├── gpu.go           # GET /api/v1/gpu/available (reads Redis, strips provider)
│   │       ├── ssh_keys.go      # CRUD handlers for /api/v1/ssh-keys
│   │       ├── billing.go       # GET usage/invoices, POST webhook
│   │       └── health.go        # GET /health (no auth)
│   │
│   ├── provider/
│   │   ├── provider.go          # Provider interface (5 methods)
│   │   ├── types.go             # Shared types: GPUOffering, ProvisionRequest, InstanceStatus
│   │   ├── registry.go          # map[string]Provider with Register/Get/All
│   │   ├── gpumap.go            # GPU type normalization across providers
│   │   └── runpod/
│   │       ├── adapter.go       # Provider interface implementation for RunPod
│   │       ├── client.go        # RunPod GraphQL HTTP client
│   │       ├── types.go         # RunPod-specific request/response types
│   │       └── adapter_test.go  # Tests with mocked HTTP responses
│   │
│   ├── provision/
│   │   ├── engine.go            # Orchestration: select provider -> WG keys -> cloud-init -> provision -> DB
│   │   ├── cloudinit.go         # Cloud-init template rendering
│   │   └── ipam.go              # WireGuard IP address allocation from subnet
│   │
│   ├── availability/
│   │   ├── poller.go            # 30s ticker goroutine, concurrent per-provider polling
│   │   └── cache.go             # Redis key scheme: gpu:{provider}:{type}:{tier}:{region}
│   │
│   ├── wireguard/
│   │   ├── manager.go           # wgctrl-go based peer add/remove/list
│   │   └── keygen.go            # WireGuard keypair generation (crypto/rand + curve25519)
│   │
│   ├── billing/
│   │   ├── stripe.go            # Stripe SDK: customers, usage records, invoices
│   │   └── metering.go          # Start/stop billing, per-second cost calculation
│   │
│   ├── auth/
│   │   ├── clerk.go             # Clerk JWKS fetch + JWT signature verification
│   │   └── middleware.go        # HTTP middleware: extract Bearer token, verify, inject Claims
│   │
│   ├── db/
│   │   ├── pool.go              # pgx pool creation with health checks
│   │   ├── instances.go         # Instance SQL queries (always org-scoped)
│   │   ├── organizations.go     # Org/user SQL queries
│   │   └── ssh_keys.go          # SSH key SQL queries (always user-scoped)
│   │
│   ├── health/
│   │   └── monitor.go           # 60s goroutine: check WG handshakes, mark unhealthy
│   │
│   └── config/
│       └── config.go            # Env var loading, required field validation, fail-fast
│
├── database/
│   └── migrations/              # SQL migration files (sequential numbering)
│       └── 001_init.sql
│
├── infra/
│   ├── cloud-init/
│   │   └── bootstrap.sh         # Go text/template: WG, SSH, hostname, MOTD, firewall
│   └── wireguard/
│       └── proxy-setup.sh       # One-time proxy server WireGuard configuration
│
├── tools/                       # Python offline tooling
│   ├── requirements.txt
│   ├── migrate.py               # DB migration runner
│   └── seed.py                  # Dev data seeding
│
├── frontend/                    # Next.js dashboard (separate build)
│   ├── src/
│   │   ├── app/                 # Next.js App Router pages
│   │   ├── components/          # React components
│   │   ├── lib/                 # API client, auth helpers
│   │   └── hooks/               # Custom React hooks (polling, etc.)
│   └── package.json
│
├── go.mod
├── go.sum
└── Makefile
```

### Structure Rationale

- **cmd/gpuctl/**: Single binary entrypoint. All dependency construction happens here, then injected downward. No business logic in main.go -- it is pure wiring.
- **internal/**: Go's built-in encapsulation. External projects cannot import these packages. This enforces the boundary that gpuctl is the only consumer.
- **internal/api/**: Thin HTTP layer. Handlers decode requests, call business logic, encode responses. No business logic in handlers.
- **internal/provider/**: Interface + registry pattern. Adding a new provider means implementing 5 methods and calling `registry.Register()`. Zero changes to existing code.
- **internal/provision/**: The orchestration "brain". This is the most complex package because it coordinates across providers, WireGuard, cloud-init, and the database. Keep it focused on orchestration, not implementation details.
- **internal/availability/**: Isolated polling loop. Runs as a background goroutine. Only dependency is the provider registry and Redis cache. No coupling to API handlers.
- **frontend/**: Completely separate build artifact. Next.js app communicates with gpuctl only via REST API. Can be deployed to Vercel/CDN independently.

## Architectural Patterns

### Pattern 1: Provider Adapter with Registry

**What:** Define a `Provider` interface with 5 methods (`Name`, `ListAvailable`, `Provision`, `GetStatus`, `Terminate`). Each upstream cloud (RunPod, E2E) implements this interface. A `Registry` holds all registered adapters in a map.

**When to use:** Always. This is the core extensibility mechanism. Adding a new provider requires zero changes to existing code -- just implement the interface and register.

**Trade-offs:** Simple and effective for 2-5 providers. If provider count grows beyond 10 or providers have wildly different capabilities, the interface may need capability flags. For Phase 1 with 1-2 providers, this is the right abstraction level.

**Example:**
```go
// internal/provider/provider.go
type Provider interface {
    Name() string
    ListAvailable(ctx context.Context) ([]GPUOffering, error)
    Provision(ctx context.Context, req ProvisionRequest) (*ProvisionResult, error)
    GetStatus(ctx context.Context, upstreamID string) (*InstanceStatus, error)
    Terminate(ctx context.Context, upstreamID string) error
}

// internal/provider/registry.go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *Registry) All() []Provider {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Provider, 0, len(r.providers))
    for _, p := range r.providers {
        out = append(out, p)
    }
    return out
}
```

**Build order dependency:** This must be built first (after config), because the provisioning engine, availability poller, and API handlers all depend on it.

### Pattern 2: Background Goroutine with Context Cancellation

**What:** Long-running background tasks (availability polling, health monitoring) run as goroutines started from main.go, controlled by a shared context.Context that cancels on SIGINT/SIGTERM.

**When to use:** For any periodic background work. The poller (30s) and health monitor (60s) both use this pattern.

**Trade-offs:** Simple, no external job scheduler needed. Goroutines are cheap. The risk is goroutine leaks if context cancellation is not properly propagated. Always use `select` on `ctx.Done()` alongside ticker channels.

**Example:**
```go
// internal/availability/poller.go
func (p *Poller) Run(ctx context.Context) {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()

    // Poll immediately on startup, don't wait for first tick
    p.poll(ctx)

    for {
        select {
        case <-ctx.Done():
            p.logger.Info("poller stopped")
            return
        case <-ticker.C:
            p.poll(ctx)
        }
    }
}

func (p *Poller) poll(ctx context.Context) {
    var wg sync.WaitGroup
    for _, prov := range p.registry.All() {
        wg.Add(1)
        go func(prov provider.Provider) {
            defer wg.Done()
            offerings, err := prov.ListAvailable(ctx)
            if err != nil {
                p.logger.Error("poll failed", "provider", prov.Name(), "err", err)
                return // don't fail the whole poll
            }
            p.cache.SetOfferings(ctx, prov.Name(), offerings)
        }(prov)
    }
    wg.Wait()
}
```

### Pattern 3: Provisioning as Orchestrated State Machine

**What:** Instance provisioning follows a deterministic state machine: `creating -> provisioning -> booting -> running -> stopping -> terminated`. The provisioning engine orchestrates multiple subsystems (provider adapter, WireGuard, database, billing) in a defined sequence.

**When to use:** For the instance lifecycle. Every state transition is explicit and recorded in the database. This makes debugging, retry, and cleanup predictable.

**Trade-offs:** More code than a simple "call API, return result" approach. But GPU provisioning involves real money (billing starts), network configuration (WireGuard), and external APIs (upstream provider) -- all of which can fail independently. The state machine makes partial failure recovery possible.

**State transitions:**
```
creating     -- API request received, DB record created
provisioning -- upstream provider API called
booting      -- upstream reports running, cloud-init executing
running      -- cloud-init callback received, WG tunnel up, billing started
stopping     -- terminate requested, upstream API called
terminated   -- upstream confirmed terminated, WG peer removed, billing stopped
error        -- any step failed, requires investigation
```

**Example:**
```go
// internal/provision/engine.go
func (e *Engine) Provision(ctx context.Context, req CreateRequest) (*Instance, error) {
    // 1. Generate instance ID
    instanceID := generateShortID() // e.g., "4a7f"

    // 2. Allocate WireGuard IP from subnet
    wgAddr, err := e.ipam.Allocate(ctx)
    if err != nil {
        return nil, fmt.Errorf("IPAM exhausted: %w", err)
    }

    // 3. Generate WireGuard keypair
    privKey, pubKey, err := wireguard.GenerateKeyPair()
    if err != nil {
        e.ipam.Release(ctx, wgAddr) // cleanup on failure
        return nil, fmt.Errorf("keygen failed: %w", err)
    }

    // 4. Select provider from availability cache
    provider, offering, err := e.selectProvider(ctx, req)
    if err != nil {
        e.ipam.Release(ctx, wgAddr)
        return nil, fmt.Errorf("no availability: %w", err)
    }

    // 5. Render cloud-init template
    cloudInit, err := e.renderCloudInit(instanceID, privKey, wgAddr, req)
    if err != nil {
        e.ipam.Release(ctx, wgAddr)
        return nil, fmt.Errorf("cloud-init render: %w", err)
    }

    // 6. Write DB record (status: creating)
    instance := &Instance{
        ID: instanceID, Status: "creating",
        // ... all fields
    }
    if err := e.db.CreateInstance(ctx, instance); err != nil {
        e.ipam.Release(ctx, wgAddr)
        return nil, fmt.Errorf("db write: %w", err)
    }

    // 7. Call upstream provider
    result, err := provider.Provision(ctx, ProvisionRequest{
        InstanceID:      instanceID,
        CloudInitScript: cloudInit,
        // ...
    })
    if err != nil {
        e.db.UpdateInstanceStatus(ctx, instanceID, "error")
        e.ipam.Release(ctx, wgAddr)
        return nil, fmt.Errorf("upstream provision: %w", err)
    }

    // 8. Update DB with upstream info (status: provisioning)
    e.db.UpdateInstanceUpstream(ctx, instanceID, result)

    // 9. Add WireGuard peer on proxy (peer becomes active when instance boots)
    if err := e.wg.AddPeer(pubKey, wgAddr); err != nil {
        // Non-fatal: instance will work, WG peer can be retried
        e.logger.Error("wg peer add failed", "instance", instanceID, "err", err)
    }

    return instance, nil
}
```

### Pattern 4: Privacy-First Response Filtering

**What:** Every API response that returns instance or availability data must strip upstream provider details. The `provider` field, upstream IP, upstream ID, and any provider-specific metadata are never exposed to customers.

**When to use:** On every public API handler. This is not optional -- it is a core product requirement.

**Trade-offs:** Requires discipline. Every new field added to internal types must be reviewed for leakage. Use separate internal and external response types to make this structural rather than relying on `json:"-"` tags.

**Example:**
```go
// Internal type (stored in DB, used by engine)
type Instance struct {
    ID               string  // "4a7f"
    UpstreamProvider string  // "runpod" -- NEVER expose
    UpstreamID       string  // "pod-abc123" -- NEVER expose
    UpstreamIP       string  // "192.168.1.5" -- NEVER expose
    Hostname         string  // "gpu-4a7f.gpu.ai"
    WGAddress        string  // "10.0.0.5"
    GPUType          string
    PricePerHour     float64
    Status           string
}

// External type (returned in API responses)
type InstanceResponse struct {
    InstanceID   string  `json:"instance_id"`
    Hostname     string  `json:"hostname"`
    SSHCommand   string  `json:"ssh_command"`
    GPUType      string  `json:"gpu_type"`
    GPUCount     int     `json:"gpu_count"`
    PricePerHour float64 `json:"price_per_hour"`
    Status       string  `json:"status"`
    Region       string  `json:"region"`
}

func toInstanceResponse(i *Instance) InstanceResponse {
    return InstanceResponse{
        InstanceID:   i.ID,
        Hostname:     i.Hostname,
        SSHCommand:   fmt.Sprintf("ssh user@%s", i.Hostname),
        GPUType:      i.GPUType,
        PricePerHour: i.PricePerHour,
        Status:       i.Status,
        // Note: no upstream fields included
    }
}
```

### Pattern 5: WireGuard Peer Management via wgctrl-go

**What:** Use the official `wgctrl-go` library (`golang.zx2c4.com/wireguard/wgctrl`) to programmatically manage WireGuard peers instead of shelling out to `wg set` commands. This provides type-safe, cross-platform peer management without subprocess overhead.

**When to use:** For all WireGuard operations on the proxy server. The proxy server is where gpuctl runs, and it manages the wg0 interface that connects to all upstream instances.

**Trade-offs:** The library handles device configuration but not device creation or IP address assignment. Those still require system-level setup (done once during proxy server provisioning). The library works on Linux, FreeBSD, and with wireguard-go userspace implementations.

**Example:**
```go
// internal/wireguard/manager.go
import (
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Manager struct {
    client        *wgctrl.Client
    interfaceName string
    logger        *slog.Logger
}

func (m *Manager) AddPeer(publicKey string, allowedIP string) error {
    key, err := wgtypes.ParseKey(publicKey)
    if err != nil {
        return fmt.Errorf("invalid public key: %w", err)
    }

    _, allowedNet, err := net.ParseCIDR(allowedIP + "/32")
    if err != nil {
        return fmt.Errorf("invalid IP: %w", err)
    }

    keepalive := 25 * time.Second
    cfg := wgtypes.Config{
        Peers: []wgtypes.PeerConfig{{
            PublicKey:                   key,
            AllowedIPs:                  []net.IPNet{*allowedNet},
            PersistentKeepaliveInterval: &keepalive,
        }},
    }

    return m.client.ConfigureDevice(m.interfaceName, cfg)
}

func (m *Manager) RemovePeer(publicKey string) error {
    key, err := wgtypes.ParseKey(publicKey)
    if err != nil {
        return fmt.Errorf("invalid public key: %w", err)
    }

    cfg := wgtypes.Config{
        Peers: []wgtypes.PeerConfig{{
            PublicKey: key,
            Remove:    true,
        }},
    }

    return m.client.ConfigureDevice(m.interfaceName, cfg)
}
```

### Pattern 6: IPAM (IP Address Management) for WireGuard Subnet

**What:** Track allocated WireGuard tunnel IPs from the 10.0.0.0/24 subnet. The proxy server is 10.0.0.1, and each instance gets the next available address (10.0.0.2, 10.0.0.3, etc.). Store allocations in PostgreSQL to survive restarts.

**When to use:** Every time an instance is provisioned or terminated.

**Trade-offs:** A /24 subnet gives 253 usable addresses (excluding .0 network and .1 proxy). For Phase 1 with a closed beta of 10-20 users, this is more than sufficient. When scaling beyond ~200 concurrent instances, expand to /16 (65,534 addresses). IPAM must be the single source of truth -- never allocate IPs from two places.

**Example:**
```go
// internal/provision/ipam.go
type IPAM struct {
    subnet  net.IPNet      // 10.0.0.0/24
    gateway net.IP         // 10.0.0.1 (proxy server, reserved)
    db      *db.Pool       // Persistent allocation tracking
    mu      sync.Mutex     // Serialize allocations
}

func (m *IPAM) Allocate(ctx context.Context) (string, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Query DB for all allocated addresses (instances not terminated)
    allocated, err := m.db.ListAllocatedWGAddresses(ctx)
    if err != nil {
        return "", err
    }

    // Find next available in subnet
    allocSet := make(map[string]bool, len(allocated))
    for _, a := range allocated {
        allocSet[a] = true
    }

    for ip := nextIP(m.subnet.IP); m.subnet.Contains(ip); ip = nextIP(ip) {
        addr := ip.String()
        if addr == m.gateway.String() {
            continue // skip proxy server address
        }
        if !allocSet[addr] {
            return addr, nil
        }
    }

    return "", fmt.Errorf("subnet exhausted: all %d addresses allocated", len(allocated))
}

func (m *IPAM) Release(ctx context.Context, addr string) {
    // IP is released when instance status -> terminated
    // No explicit release needed -- IPAM queries active instances
}
```

### Pattern 7: Dependency Injection via Constructor

**What:** All components receive their dependencies as constructor arguments. No global variables, no init() functions, no service locators.

**When to use:** Every internal package. main.go is the composition root.

**Trade-offs:** More boilerplate in main.go (all wiring is explicit). But explicit wiring means no hidden dependencies, full testability with mocks, and the dependency graph is visible in one place.

**Example:**
```go
// cmd/gpuctl/main.go (composition root)
func main() {
    cfg := config.MustLoad()

    pool := db.MustNewPool(ctx, cfg.DatabaseURL)
    defer pool.Close()

    redisClient := redis.NewClient(cfg.RedisURL)
    cache := availability.NewCache(redisClient)

    wgManager := wireguard.MustNewManager("wg0")
    defer wgManager.Close()

    registry := provider.NewRegistry()
    registry.Register(runpod.NewAdapter(cfg.RunPodAPIKey))

    ipam := provision.NewIPAM(pool, cfg.WGSubnet, cfg.WGGateway)
    engine := provision.NewEngine(registry, wgManager, pool, cache, ipam, cfg)

    billing := billing.NewService(cfg.StripeSecretKey, cfg.StripeWebhookSecret, pool)
    auth := auth.NewVerifier(cfg.ClerkSecretKey)

    server := api.NewServer(engine, cache, billing, auth, pool, cfg)
    poller := availability.NewPoller(registry, cache, 30*time.Second)
    monitor := health.NewMonitor(pool, wgManager, 60*time.Second)

    // Start background goroutines...
    // Start HTTP server...
    // Wait for shutdown...
}
```

## Data Flow

### Request Flow: Instance Provisioning (Create)

```
Customer Dashboard (Next.js)
    |
    |  POST /api/v1/instances
    |  { gpu_type: "h100_sxm", gpu_count: 8, region: "us-west", ssh_public_keys: [...] }
    |
    v
Auth Middleware (Clerk JWT)
    |  Verify token -> extract user_id, org_id
    v
Create Instance Handler
    |  Validate request fields
    |  Check billing status (Stripe: has payment method? no unpaid invoices?)
    v
Provisioning Engine
    |
    |-- 1. Query Redis cache for matching offerings
    |       Key pattern: gpu:*:h100_sxm:on_demand:us-west
    |       Returns: best-priced offering with available_count > 0
    |
    |-- 2. Allocate WireGuard IP from IPAM (10.0.0.x)
    |
    |-- 3. Generate WireGuard keypair (curve25519)
    |
    |-- 4. Render cloud-init template
    |       Inject: WG keys, SSH keys, instance ID, proxy endpoint
    |
    |-- 5. Write DB record (status: "creating")
    |       Table: instances
    |       Fields: id, org_id, upstream_provider, wg_keys, gpu_config, price
    |
    |-- 6. Call RunPod adapter
    |       GraphQL mutation: podFindAndDeployOnDemand
    |       Passes: GPU type, cloud-init script, Docker image
    |       Returns: upstream_id, upstream_ip
    |
    |-- 7. Update DB (status: "provisioning", upstream_id, upstream_ip)
    |
    +-- 8. Register WireGuard peer on proxy
            wgctrl: ConfigureDevice("wg0", add peer with public key + allowed IP)

    ... (~15 seconds pass, instance boots) ...

Upstream Instance (cloud-init executes)
    |
    |-- Install WireGuard, configure tunnel to proxy
    |-- Set SSH keys, hostname, MOTD, firewall
    +-- POST /internal/instances/{id}/ready (callback to gpuctl)
            |
            v
        Internal Handler
            |-- Update DB: status -> "running", billing_start = now
            |-- Start Stripe usage metering
            +-- Return 200 OK

Customer receives (via polling or webhook):
    {
      "instance_id": "4a7f",
      "hostname": "gpu-4a7f.gpu.ai",
      "ssh_command": "ssh user@gpu-4a7f.gpu.ai",
      "status": "running",
      "gpu_type": "h100_sxm",
      "price_per_hour": 2.12
    }
```

### Request Flow: GPU Availability Query

```
Customer Dashboard
    |
    |  GET /api/v1/gpu/available?type=h100_sxm&tier=on_demand
    v
Auth Middleware
    v
Availability Handler
    |
    |-- Read from Redis
    |   SCAN keys matching gpu:*:h100_sxm:on_demand:*
    |   Deserialize GPUOffering structs
    |
    |-- Filter by query params (type, tier, region)
    |
    |-- Strip provider field (CRITICAL: customer never sees "runpod" or "e2e")
    |
    |-- Aggregate identical offerings across providers
    |   (same GPU type + tier + region = sum available_count, show best price)
    |
    +-- Return JSON array sorted by price
```

### Background Flow: Availability Polling

```
Poller Goroutine (started from main.go, runs until ctx cancelled)
    |
    |  Every 30 seconds:
    v
    +--- For each registered provider (concurrent goroutines) ---+
    |                                                              |
    |  RunPod Adapter                  E2E Adapter                |
    |  |                               |                          |
    |  | GraphQL: gpuTypes query       | REST: GET /gpu/inventory |
    |  | Parse: GPU type, price,       | Parse: GPU type, price,  |
    |  |   availability, region        |   availability, region   |
    |  |                               |                          |
    |  +---- Write to Redis -----------+                          |
    |        Key: gpu:{provider}:{type}:{tier}:{region}           |
    |        Value: JSON GPUOffering                              |
    |        TTL: 60 seconds                                      |
    +-------------------------------------------------------------+
```

### Background Flow: Health Monitoring

```
Health Monitor Goroutine (started from main.go)
    |
    |  Every 60 seconds:
    v
    Query DB: SELECT * FROM instances WHERE status = 'running'
    |
    For each running instance:
    |
    |-- Check WireGuard peer status via wgctrl
    |   device.Peers[publicKey].LastHandshakeTime
    |
    |-- If last handshake > 5 minutes ago:
    |   +-- Mark unhealthy in DB, log alert
    |
    +-- If last handshake recent:
        +-- Instance is healthy (no action)
```

### Key Data Flows

1. **Provisioning flow:** API handler -> Provisioning Engine -> Provider Adapter -> Upstream API -> (async) Cloud-init callback -> DB status update. This is the critical path. Every step must handle failure and rollback.

2. **Availability flow:** Background poller -> Provider adapters -> Redis cache -> API handler reads cache. Decoupled from the request path. If Redis is stale, customer sees slightly old data (acceptable for 30-60s staleness).

3. **Privacy flow:** All customer-facing responses are constructed from separate response types that structurally exclude upstream fields. The provider field in Redis is used internally for routing but stripped before API response serialization.

4. **Billing flow:** Instance creation -> start metering (billing_start in DB + Stripe usage record) -> Instance termination -> stop metering (billing_end, compute total_cost, report to Stripe). Must be atomic with instance lifecycle -- never bill without a running instance, never run without billing.

5. **Network flow:** Customer SSH -> DNS (gpu-4a7f.gpu.ai -> proxy IP) -> Proxy server -> WireGuard tunnel -> Upstream instance. The customer's packets never touch the upstream provider's public IP.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-50 instances (closed beta) | Single gpuctl binary on one server. Single WireGuard proxy. /24 subnet (253 IPs). Single PostgreSQL + Redis instance. This is Phase 1. |
| 50-500 instances | Same architecture. Expand WireGuard subnet to /16. Add connection pooling tuning (pgx max conns). Consider Redis Cluster if availability cache reads become a bottleneck. Add a second proxy server for geographic redundancy. |
| 500-5,000 instances | Multiple proxy servers (regional). Load balance gpuctl behind a reverse proxy. Read replicas for PostgreSQL. The single binary is still fine -- Go handles thousands of concurrent connections per process. Split WireGuard management to a separate service if peer operations become a bottleneck. |
| 5,000+ instances | Extract WireGuard management and availability polling into separate services. Consider event-driven provisioning (message queue between API and provisioning engine). Multiple regional proxy servers with independent subnets. This is Phase 2+ territory. |

### Scaling Priorities

1. **First bottleneck: WireGuard proxy throughput.** A single Linux server can handle ~1,000 WireGuard peers with negligible CPU overhead (WireGuard is kernel-level). The bottleneck is network bandwidth, not peer count. Mitigation: regional proxy servers.

2. **Second bottleneck: PostgreSQL write contention.** Instance creation and status updates are write-heavy. With connection pooling (pgx, max 20 conns), a single Postgres instance handles thousands of writes/second. Mitigation: read replicas, then partitioning by org.

3. **Third bottleneck: Redis availability cache.** Each poll writes ~50-100 keys (one per GPU offering per provider). At 30s intervals, this is ~200 writes/minute -- negligible for Redis. The read path (API handler scanning keys) is more concerning at scale. Mitigation: pre-aggregate offerings into a single key per region instead of key-per-offering.

## Anti-Patterns

### Anti-Pattern 1: Shell Exec for WireGuard

**What people do:** Call `exec.Command("wg", "set", ...)` to manage WireGuard peers.
**Why it's wrong:** Subprocess spawning is slow, error-prone (parsing stdout), not cross-platform, and creates race conditions when multiple goroutines add/remove peers simultaneously. The `wg` CLI is designed for human use, not programmatic control.
**Do this instead:** Use `wgctrl-go` (`golang.zx2c4.com/wireguard/wgctrl`). It provides a native Go API that talks directly to the kernel WireGuard module via netlink. Type-safe, concurrent-safe, no string parsing.

### Anti-Pattern 2: Business Logic in HTTP Handlers

**What people do:** Put provisioning orchestration, billing checks, database queries, and response formatting all in the handler function.
**Why it's wrong:** Handlers become 200+ line functions that are untestable (require full HTTP request/response cycle), impossible to reuse (CLI or background job can't call provisioning), and mix concerns (HTTP serialization + business rules + data access).
**Do this instead:** Handlers should do three things: decode request, call a service method, encode response. The provisioning engine, billing service, and database layer are separate packages with their own interfaces, testable independently.

### Anti-Pattern 3: Leaking Upstream Provider Details

**What people do:** Return the `Instance` database struct directly as JSON, or include provider-specific error messages in API responses.
**Why it's wrong:** The entire product value depends on the customer not knowing they are using RunPod or E2E. A single leaked `"provider": "runpod"` field or error message containing `"pod-abc123"` destroys the privacy layer.
**Do this instead:** Use separate internal and external response types (see Pattern 4 above). Review every new field. Log upstream details with slog for debugging but never return them in HTTP responses. Error messages should be generic: "provisioning failed" not "RunPod API returned 503".

### Anti-Pattern 4: Synchronous Provisioning in the HTTP Request

**What people do:** Block the HTTP response until the upstream instance is fully booted and WireGuard tunnel is established (~15-30 seconds).
**Why it's wrong:** HTTP timeouts, client disconnects, and poor user experience. If the client disconnects, the provisioning still happened but the client never got the instance ID.
**Do this instead:** Return immediately with status "creating" and the instance ID. The customer polls `GET /api/v1/instances/{id}` or the dashboard auto-refreshes. The cloud-init callback (`POST /internal/instances/{id}/ready`) transitions the status to "running" asynchronously.

### Anti-Pattern 5: Global Mutable State for IPAM

**What people do:** Keep a simple in-memory counter for WireGuard IP allocation (`nextIP++`).
**Why it's wrong:** Crashes lose allocation state, leading to IP conflicts. Two concurrent provisions can get the same IP (race condition). Restarting gpuctl requires re-scanning the WireGuard interface to reconstruct state.
**Do this instead:** Store WireGuard IP allocations in PostgreSQL (the `wg_address` column on the `instances` table). Query for allocated IPs, find the next gap. Use a mutex to serialize allocation within a single process. The database is the single source of truth.

### Anti-Pattern 6: Using an ORM for Financial Data

**What people do:** Use GORM or ent for billing and usage record queries.
**Why it's wrong:** ORMs generate SQL you cannot audit. Billing must be exact -- a wrong JOIN or lazy-load can cause incorrect charges. N+1 queries are common with ORMs and devastating for both correctness and performance in billing contexts.
**Do this instead:** Hand-written SQL in `internal/db/`. Every query is visible, auditable, and optimizable. Use pgx's native type support for NUMERIC, TIMESTAMPTZ, UUID. Billing queries deserve the most scrutiny.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| **RunPod** | GraphQL API via HTTP POST to `api.runpod.io/graphql`. Bearer token auth. | Use `podFindAndDeployOnDemand` mutation for on-demand, `podRentInterruptable` for spot. Query `gpuTypes` for availability with `stockStatus` and `lowestPrice`. No native cloud-init -- inject startup logic via Docker args or RunPod's startup command field. |
| **E2E Networks** | REST API at `api.e2enetworks.com`. Token auth. Rate limit: 5,000 req/hr. | Deferred to future milestone. Similar adapter pattern. |
| **Clerk** | JWKS endpoint for JWT verification. No SDK needed -- verify JWT signature against JWKS, parse standard claims. | Fetch JWKS at startup, cache with periodic refresh (every 6 hours). Standard `crypto/rsa` + `encoding/json` for verification. |
| **Stripe** | Official Go SDK (`github.com/stripe/stripe-go`). Usage-based billing with metered subscriptions. | Create Stripe customer per org. Report per-second usage via `UsageRecord.New()`. Handle webhooks for payment events. |
| **WireGuard Proxy** | Local wg0 interface on the same server running gpuctl. Managed via `wgctrl-go`. | Proxy server setup is a one-time infra task (install WireGuard, create wg0, assign 10.0.0.1/24, configure iptables forwarding). gpuctl manages peers dynamically. |
| **PostgreSQL** | `pgx/v5` connection pool. Raw SQL queries. | Min 5, max 20 connections. All queries parameterized (no SQL injection). Migrations via Python `tools/migrate.py`. |
| **Redis** | `go-redis/v9` client. Key-value cache for GPU offerings. | Keys: `gpu:{provider}:{type}:{tier}:{region}`. Values: JSON-serialized `GPUOffering`. TTL: 60 seconds. No persistence needed -- cache is rebuilt every 30s by poller. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| API handlers <-> Provisioning Engine | Direct function call (same process) | Handler calls `engine.Provision(ctx, req)`. Returns `(*Instance, error)`. No serialization overhead. |
| API handlers <-> Availability Cache | Direct function call to Redis wrapper | Handler calls `cache.GetOfferings(ctx, filters)`. Returns `[]GPUOffering`. |
| Provisioning Engine <-> Provider Adapters | Interface method call | Engine calls `provider.Provision(ctx, req)` on the selected adapter. Adapter handles all upstream API communication. |
| Provisioning Engine <-> WireGuard Manager | Direct function call | Engine calls `wg.AddPeer(pubKey, ip)` after successful upstream provisioning. Non-fatal if this fails (retry on health check). |
| Provisioning Engine <-> Database | Direct function call to db package | Engine calls `db.CreateInstance()`, `db.UpdateInstanceStatus()`. All writes are serialized per-instance via instance ID. |
| Poller <-> Provider Adapters | Interface method call (concurrent) | Poller calls `provider.ListAvailable(ctx)` on all adapters concurrently. Each adapter runs in its own goroutine. |
| Poller <-> Redis Cache | Direct function call | Poller calls `cache.SetOfferings(ctx, provider, offerings)`. Writes are per-provider, no contention. |
| Cloud-init callback <-> gpuctl | HTTP POST to /internal/instances/{id}/ready | Upstream instance calls back to gpuctl after boot. Authenticated via internal bearer token. This transitions instance from "booting" to "running" and starts billing. |
| Frontend <-> gpuctl | REST API over HTTPS | Complete separation. Frontend knows nothing about Go internals. All communication via JSON API. Can be developed and deployed independently. |

## Build Order (Dependency Graph)

Components should be built in this order based on their dependency relationships. Components at the same level can be built in parallel.

```
Level 1 (Foundation - no deps on other internal packages):
    config/config.go          -- loads env vars, used by everything
    provider/types.go         -- shared type definitions
    provider/provider.go      -- Provider interface
    db/pool.go                -- pgx connection pool

Level 2 (Core infrastructure - depends on Level 1):
    provider/registry.go      -- depends on Provider interface
    provider/runpod/           -- depends on Provider interface + types
    wireguard/keygen.go        -- standalone crypto
    wireguard/manager.go       -- depends on wgctrl-go (external), no internal deps
    auth/clerk.go              -- depends on config (Clerk keys)
    availability/cache.go      -- depends on types + Redis client

Level 3 (Business logic - depends on Level 2):
    provision/ipam.go          -- depends on db (allocation tracking)
    provision/cloudinit.go     -- depends on config (proxy endpoint/key)
    availability/poller.go     -- depends on registry + cache
    billing/stripe.go          -- depends on db + config (Stripe keys)
    health/monitor.go          -- depends on db + wireguard manager

Level 4 (Orchestration - depends on Level 3):
    provision/engine.go        -- depends on registry, wireguard, IPAM, cloud-init, db

Level 5 (HTTP layer - depends on Level 4):
    api/middleware.go          -- depends on auth
    api/handlers/              -- depends on engine, cache, billing, db
    api/server.go              -- depends on handlers, middleware
    api/router.go              -- depends on server

Level 6 (Entrypoint - depends on everything):
    cmd/gpuctl/main.go         -- wires all deps, starts server + goroutines

Level 7 (Frontend - independent build, depends on API contract):
    frontend/                  -- Next.js app, communicates via REST API only

Level 8 (Tooling - independent):
    database/migrations/       -- SQL files, run by tools/migrate.py
    tools/                     -- Python scripts, run manually or in CI
```

**Critical path for MVP:** config -> provider interface + RunPod adapter -> database + instance CRUD -> provisioning engine (without WireGuard initially) -> API handlers -> test end-to-end provisioning with RunPod. Then layer on WireGuard, auth, billing, availability polling, health monitoring.

## Key Architecture Decisions

| Decision | Rationale | Trade-off |
|----------|-----------|-----------|
| Single binary | Operational simplicity. One process to deploy, monitor, restart. No service discovery, no RPC overhead. | All components scale together. Cannot scale availability polling independently from API serving. Acceptable for Phase 1 scale. |
| Stdlib net/http | No framework lock-in. Go 1.22+ routing with METHOD /path/{param} patterns is sufficient. Middleware pattern is well-understood. | Less DX sugar than chi/Echo (no automatic route grouping, no built-in middleware chain helper). Acceptable trade-off for a focused API surface. |
| PostgreSQL for everything | Single data store simplifies operations. ACID for billing. pgcrypto for UUIDs. Good enough for instance metadata, user data, and usage records. | No dedicated time-series DB for usage records. Acceptable until volume exceeds 100K+ usage records/day. |
| Redis for availability only | Ephemeral data with TTL. Perfectly suited for "what is available right now." Fast reads for API responses. | If Redis goes down, availability API returns stale/empty data. Running instances are unaffected. Acceptable failure mode. |
| Async provisioning | Return 201 immediately, poll for status. Avoids long HTTP connections. Matches how all cloud providers work (AWS, GCP, Azure all use this pattern). | Client must implement polling. Slightly more complex frontend. But this is the industry standard pattern. |
| Cloud-init for instance setup | Single script handles all instance configuration. Portable across providers. No agent to maintain. | Cannot modify instance after boot without SSH. If cloud-init fails, instance is stuck. Must have reliable error reporting from cloud-init callback. |
| wgctrl-go over shell exec | Type-safe, concurrent-safe, no subprocess overhead. Official WireGuard project. | Requires CGO on some platforms (Linux netlink). Binary portability slightly reduced. Acceptable since proxy server is always Linux. |
| Raw SQL over ORM | Auditable, optimizable, exact control for billing queries. pgx provides excellent Go type mapping. | More boilerplate for simple CRUD. Worth it for correctness in financial operations. |

## Sources

- [RunPod Pod Management API](https://docs.runpod.io/sdks/graphql/manage-pods) -- Official GraphQL API documentation for pod provisioning, termination, and availability queries (HIGH confidence)
- [RunPod REST API announcement](https://www.runpod.io/blog/runpod-rest-api-gpu-management) -- REST API as alternative to GraphQL (HIGH confidence)
- [wgctrl-go package](https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl) -- Official WireGuard Go control library (HIGH confidence)
- [wgctrl-go GitHub](https://github.com/WireGuard/wgctrl-go) -- Source and examples for programmatic WireGuard management (HIGH confidence)
- [Cloud Computing Patterns: Provider Adapter](https://www.cloudcomputingpatterns.org/provider_adapter/) -- Canonical description of the provider adapter pattern for multi-cloud abstraction (HIGH confidence)
- [Provider Pattern in Go](https://medium.com/swlh/provider-model-in-go-and-why-you-should-use-it-clean-architecture-1d84cfe1b097) -- Go-specific implementation patterns (MEDIUM confidence)
- [Adapter Pattern in Go (Bitfield Consulting)](https://bitfieldconsulting.com/posts/adapter) -- Detailed Go adapter pattern discussion (MEDIUM confidence)
- [NetBird WireGuard Architecture](https://dasroot.net/posts/2026/02/netbird-architecture-wireguard-go-zero-trust/) -- WireGuard + Go at scale reference architecture (MEDIUM confidence)
- [Tailscale: How it works](https://tailscale.com/blog/how-tailscale-works) -- WireGuard overlay network architecture reference (HIGH confidence)
- [Google Compute Engine Instance Lifecycle](https://cloud.google.com/compute/docs/instances/instance-lifecycle) -- State machine pattern for cloud instance provisioning (HIGH confidence)
- [AWS: Making Retries Safe with Idempotent APIs](https://aws.amazon.com/builders-library/making-retries-safe-with-idempotent-APIs/) -- Retry and idempotency patterns for provisioning (HIGH confidence)
- [Redis Real-Time Inventory](https://redis.io/solutions/real-time-inventory/) -- Redis patterns for real-time availability caching (HIGH confidence)
- [wgipam Go IPAM](https://github.com/mdlayher/wgipam) -- WireGuard IP address management implementation reference (MEDIUM confidence)
- [Multi-Cloud GPU Orchestration Guide](https://introl.com/blog/multi-cloud-gpu-orchestration-aws-azure-gcp) -- Multi-cloud patterns and networking strategies (MEDIUM confidence)

---
*Architecture research for: GPU cloud aggregation platform*
*Researched: 2026-02-24*
