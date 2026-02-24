# Architecture

**Analysis Date:** 2026-02-24

## Pattern Overview

**Overall:** Layered monolith with provider adapter pattern

**Key Characteristics:**
- Single Go binary (`gpuctl`) serving both public and internal APIs
- Provider adapter abstraction for GPU cloud aggregation
- Background workers for availability polling and health monitoring
- Redis-based caching for real-time GPU inventory
- PostgreSQL for persistent state (instances, users, billing)
- WireGuard tunneling for privacy — instances connect via proxy bastion

## Layers

**Presentation (HTTP API):**
- Purpose: Handle customer requests and internal callbacks
- Location: `internal/api/`
- Contains: HTTP server setup, route definitions, HTTP handlers (not yet implemented)
- Depends on: auth, provision, billing, db, availability (cache), provider
- Used by: Frontend dashboard, cloud-init bootstrap callbacks

**Authentication & Authorization:**
- Purpose: Verify user identity and permissions
- Location: `internal/auth/`
- Contains: Clerk JWT verification middleware, claims extraction
- Depends on: None
- Used by: api handlers (public routes)

**Provisioning Orchestration:**
- Purpose: Orchestrate instance creation across providers
- Location: `internal/provision/`
- Contains: Provision engine that coordinates provider calls, WireGuard setup, cloud-init injection
- Depends on: provider, wireguard, config
- Used by: api handlers

**Provider Adapters:**
- Purpose: Abstract upstream GPU providers behind a unified interface
- Location: `internal/provider/` (interface), `internal/provider/runpod/` (implementations)
- Contains: Provider interface definition, type definitions (GPUOffering, ProvisionRequest), RunPod adapter
- Depends on: None (adapters are the leaves)
- Used by: provision engine, availability poller

**Availability Management:**
- Purpose: Poll providers for GPU inventory and cache results
- Location: `internal/availability/`
- Contains: Poller worker (runs every 30s), Redis cache helpers
- Depends on: provider, redis client
- Used by: api handlers (for listing available GPUs)

**Billing:**
- Purpose: Integration with Stripe for usage metering and invoices
- Location: `internal/billing/`
- Contains: Stripe service, webhook handling, usage record lifecycle
- Depends on: db
- Used by: api handlers, provision workflow

**WireGuard Tunneling:**
- Purpose: Manage VPN tunnel configuration for privacy layer
- Location: `internal/wireguard/`
- Contains: Peer management (add/remove), key generation, WireGuard command execution
- Depends on: None
- Used by: provision engine (during setup), health monitor (during checks)

**Health Monitoring:**
- Purpose: Periodic checks of instance health via WireGuard peer status
- Location: `internal/health/`
- Contains: Monitor worker (runs every 60s), checks last handshake times
- Depends on: db, wireguard
- Used by: Background goroutine in main

**Data Persistence:**
- Purpose: PostgreSQL connection management and query helpers
- Location: `internal/db/`
- Contains: pgx connection pool, queries for instances, users, organizations, SSH keys
- Depends on: None (queries may reference billing or provider types as DTO)
- Used by: api handlers, billing service, health monitor

**Configuration:**
- Purpose: Load and validate environment variables
- Location: `internal/config/`
- Contains: Config struct, Load() function, env var parsing
- Depends on: None
- Used by: main entrypoint

## Data Flow

**Instance Provisioning Flow:**

1. Customer calls `POST /api/v1/instances` with GPU specs
2. API handler extracts auth claims (org_id, user_id)
3. Handler calls `billing.CheckBillingStatus()` — validates payment on file
4. Handler calls `provision.Engine.Provision()`:
   - Selects provider adapter based on region/tier preferences
   - Generates WireGuard key pair via `wireguard.GenerateKeyPair()`
   - Builds cloud-init script from template:
     - Injects WireGuard private key and tunnel address
     - Injects SSH public keys from database
     - Injects instance hostname (gpu-xxxx)
     - Optionally injects Docker image URL
   - Calls `provider.Provision()` (RunPod GraphQL API, etc.)
   - Registers WireGuard peer on proxy via `wireguard.AddPeer()`
   - Stores instance record in database with upstream_id, upstream_ip
5. Handler returns `ProvisionResult` with hostname and SSH command
6. Instance boots, cloud-init runs, WireGuard tunnel established
7. Instance calls `POST /internal/instances/{id}/ready` with internal token
8. API handler updates instance status to "running", sets billing_start
9. Availability poller and health monitor begin tracking the instance

**Availability Polling Flow:**

1. Background worker `availability.Poller` runs on 30s interval
2. For each provider adapter concurrently:
   - Calls `provider.ListAvailable(ctx)` (e.g., RunPod GraphQL query)
   - Receives `[]GPUOffering` with prices, regions, tier info
   - For each offering, writes to Redis: `gpu:{provider}:{type}:{tier}:{region}`
   - Logs errors per provider but continues polling
3. TTL on cache entries: 60s (auto-purge stale data)
4. API handler queries cache on `GET /api/v1/gpu/available` — filters and returns (strips provider field from response)

**Health Monitoring Flow:**

1. Background worker `health.Monitor` runs on 60s interval
2. Queries database for all instances with status='running'
3. For each instance:
   - Calls `wireguard.ListPeers()` to get peer status
   - Checks last handshake time (from `wg show` output)
   - If handshake > 5 minutes ago, marks instance unhealthy
   - Updates instance status in database
   - Logs unhealthy instances for alerting
4. No automatic termination — humans decide

**State Management:**

- **Instances:** PostgreSQL `instances` table — source of truth for provision state
- **GPU Inventory:** Redis cache — ephemeral, sourced from provider polling
- **Billing:** PostgreSQL `usage_records` — computed from billing_start/billing_end
- **WireGuard Peers:** Linux kernel (via `wg show`) — synced from `instances` table on restart
- **User Identity:** External (Clerk) — resolved on each request via JWT

## Key Abstractions

**Provider Interface:**
- Purpose: Unify multiple upstream GPU cloud APIs
- Examples: `internal/provider/provider.go` (interface), `internal/provider/runpod/adapter.go` (implementation)
- Pattern: Each provider implements `Name()`, `ListAvailable()`, `Provision()`, `GetStatus()`, `Terminate()`
- New provider: Implement these 5 methods, register in provision engine's provider map

**GPUOffering / ProvisionRequest:**
- Purpose: Standardized types for inventory and request routing
- Examples: `internal/provider/types.go`
- Pattern: ProvisionRequest includes SSH keys, WireGuard details, cloud-init template directives; GPUOffering includes pricing tiers and availability

**Cloud-init Template:**
- Purpose: Inject configuration into upstream instances on first boot
- Location: Not yet visible in codebase (referenced in `infra/cloud-init/`)
- Pattern: Template receives variables for WireGuard, SSH, instance ID, optional Docker image
- Result: Each instance auto-joins VPN, accepts SSH keys, reports ready

## Entry Points

**HTTP Server:**
- Location: `cmd/gpuctl/main.go`
- Triggers: `go run ./cmd/gpuctl` or `make build && ./gpuctl`
- Responsibilities:
  - Parse port from environment (default :9090)
  - Wire all dependencies (db pool, redis, auth verifier, providers, etc.)
  - Create HTTP mux and register routes
  - Start three goroutines: HTTP server, availability poller, health monitor
  - Handle graceful shutdown on SIGINT/SIGTERM

**Public API Routes:**
- Prefix: `/api/v1/` — Clerk JWT auth required
- Instances:
  - `GET /api/v1/instances` — List user's instances
  - `POST /api/v1/instances` — Create new instance
  - `GET /api/v1/instances/{id}` — Get instance details
  - `DELETE /api/v1/instances/{id}` — Terminate instance
  - `GET /api/v1/instances/{id}/status` — Get current status
- GPU Availability:
  - `GET /api/v1/gpu/available` — List available GPUs (filtered, no provider field visible)
- SSH Keys:
  - `GET /api/v1/ssh-keys` — List user's SSH keys
  - `POST /api/v1/ssh-keys` — Add new SSH key
  - `DELETE /api/v1/ssh-keys/{id}` — Delete SSH key
- Billing:
  - `GET /api/v1/billing/usage` — Get current usage metrics
  - `GET /api/v1/billing/invoices` — List invoices from Stripe
  - `POST /api/v1/billing/webhook` — Stripe webhook (signature verified, no auth required)

**Internal Routes:**
- Prefix: `/internal/` — GPUCTL_INTERNAL_TOKEN auth required
- Used by cloud-init bootstrap:
  - `POST /internal/instances/{id}/ready` — Instance reports boot success, billing_start recorded
  - `POST /internal/instances/{id}/health` — Instance sends health ping

**Health Check:**
- `GET /health` — No auth, returns `{"status":"ok"}`

## Error Handling

**Strategy:** Context-aware error propagation with structured logging

**Patterns:**
- All functions that do I/O (database, HTTP, Redis) accept `context.Context` as first param
- Errors are logged via `slog` (structured logging) with request context (request ID, user ID)
- HTTP handlers return appropriate status codes (400 for validation, 401 for auth, 500 for server errors)
- Provider adapter errors logged per-provider — poller continues even if one provider fails
- Health monitor errors logged but don't stop monitoring loop
- Graceful shutdown waits up to 10s for in-flight requests

## Cross-Cutting Concerns

**Logging:**
- Via Go 1.21+ `log/slog` structured logger
- Each layer injects logger from dependencies
- Request ID middleware (not yet implemented) will enhance log correlation

**Validation:**
- API handlers validate request bodies before processing
- Database queries assume valid data (validation happens at API boundary)
- Provider adapters validate responses from upstream APIs

**Authentication:**
- Public API: Clerk JWT middleware extracts user/org claims
- Internal API: GPUCTL_INTERNAL_TOKEN (env var) simple token comparison
- Health check: No authentication
- Billing webhook: Stripe signature verification (HMAC)

**Context Propagation:**
- All I/O operations pass context down the stack
- Enables timeout enforcement, cancellation, and request-scoped values (user ID, org ID)

---

*Architecture analysis: 2026-02-24*
