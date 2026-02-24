# Codebase Structure

**Analysis Date:** 2026-02-24

## Directory Layout

```
gpu-ai/
├── cmd/                                    # Entrypoint binaries
│   └── gpuctl/
│       └── main.go                         # Single Go binary entrypoint
├── internal/                               # All business logic (Go packages)
│   ├── api/
│   │   ├── server.go                       # HTTP server setup (TODO: implement)
│   │   └── handlers.go                     # Request handlers (TODO: implement)
│   ├── auth/
│   │   └── clerk.go                        # Clerk JWT verification middleware
│   ├── availability/
│   │   ├── cache.go                        # Redis cache helpers
│   │   └── poller.go                       # Background worker: poll providers every 30s
│   ├── billing/
│   │   └── stripe.go                       # Stripe integration: metering, webhooks
│   ├── config/
│   │   └── config.go                       # Environment variable loading
│   ├── db/
│   │   ├── pool.go                         # pgx connection pool wrapper
│   │   ├── instances.go                    # Instance queries (TODO: implement)
│   │   ├── organizations.go                # Org/user queries (TODO: implement)
│   │   └── ssh_keys.go                     # SSH key queries (TODO: implement)
│   ├── health/
│   │   └── monitor.go                      # Background worker: instance health checks every 60s
│   ├── provider/
│   │   ├── provider.go                     # Provider interface definition
│   │   ├── types.go                        # GPUOffering, ProvisionRequest, InstanceStatus types
│   │   └── runpod/
│   │       ├── adapter.go                  # RunPod provider implementation (TODO: implement)
│   │       └── adapter_test.go             # RunPod adapter tests
│   ├── provision/
│   │   └── engine.go                       # Provisioning orchestration (TODO: implement)
│   └── wireguard/
│       ├── manager.go                      # WireGuard peer management
│       └── keygen.go                       # WireGuard key generation helpers
├── database/
│   └── migrations/
│       └── 20250224_v0.sql                 # Initial schema: orgs, users, instances, billing
├── infra/
│   ├── cloud-init/                         # Cloud-init bootstrap template
│   └── wireguard/                          # WireGuard configuration examples
├── docs/
│   └── ARCHITECTURE.md                     # High-level system design (product vision)
├── frontend/                               # Next.js dashboard (separate concern)
├── tools/                                  # Python tooling
│   ├── requirements.txt
│   ├── migrate.py                          # Run SQL migrations
│   ├── seed.py                             # Seed dev data
│   └── reports/
├── go.mod                                  # Go module definition
├── go.sum                                  # Go dependency lock
└── .env.example                            # Environment variable template
```

## Directory Purposes

**cmd/**
- Purpose: Executable entrypoints
- Contains: Only `gpuctl` binary main function
- Key files: `cmd/gpuctl/main.go`

**internal/api/**
- Purpose: HTTP API server and request routing
- Contains: Server struct, route registration, HTTP handlers
- Key files: `server.go` (mux setup), `handlers.go` (endpoint logic)

**internal/auth/**
- Purpose: Authentication and authorization
- Contains: Clerk JWT verification middleware, claims structs
- Key files: `clerk.go` (implement JWT verification)

**internal/availability/**
- Purpose: Real-time GPU inventory polling and caching
- Contains: Background poller worker, Redis cache operations
- Key files: `poller.go` (30s ticker), `cache.go` (Redis helpers)

**internal/billing/**
- Purpose: Stripe integration for payments and usage metering
- Contains: Usage record lifecycle, invoice retrieval, webhook handling
- Key files: `stripe.go` (implement Stripe service)

**internal/config/**
- Purpose: Configuration management
- Contains: Config struct, environment variable parsing
- Key files: `config.go` (Load() function, required env vars)

**internal/db/**
- Purpose: Database abstraction layer
- Contains: pgx connection pool, SQL query helpers
- Key files: `pool.go` (connection management), `instances.go`/`organizations.go`/`ssh_keys.go` (queries)

**internal/health/**
- Purpose: Instance health monitoring
- Contains: Background monitor worker
- Key files: `monitor.go` (60s ticker, WireGuard peer checks)

**internal/provider/**
- Purpose: GPU provider adapter abstraction
- Contains: Interface definition, shared types, provider implementations
- Key files: `provider.go` (interface), `types.go` (GPUOffering, ProvisionRequest)

**internal/provider/runpod/**
- Purpose: RunPod-specific adapter implementation
- Contains: RunPod API calls, response mapping
- Key files: `adapter.go` (implement Provider interface), `adapter_test.go` (tests)

**internal/provision/**
- Purpose: Instance provisioning orchestration
- Contains: Engine that coordinates providers, WireGuard, cloud-init
- Key files: `engine.go` (Provision method, Terminate method)

**internal/wireguard/**
- Purpose: WireGuard VPN management
- Contains: Peer configuration, key generation
- Key files: `manager.go` (AddPeer, RemovePeer, ListPeers), `keygen.go` (GenerateKeyPair)

**database/migrations/**
- Purpose: SQL schema versioning
- Contains: Migration files (one per version)
- Key files: `20250224_v0.sql` (tables: organizations, users, instances, ssh_keys, environments, usage_records)

**infra/**
- Purpose: Infrastructure and deployment templates
- Contains: Cloud-init bootstrap script, WireGuard config examples
- Key files: `cloud-init/` (template for instance bootstrap)

**docs/**
- Purpose: Architecture and design documentation
- Contains: High-level system design
- Key files: `ARCHITECTURE.md` (Phase 1 overview, system diagram, provider interface)

## Key File Locations

**Entry Points:**
- `cmd/gpuctl/main.go`: Wires dependencies, starts HTTP server + background workers, handles graceful shutdown

**Configuration:**
- `.env.example`: Template for required environment variables
- `internal/config/config.go`: Loads and validates config from environment

**Core Logic:**
- `internal/api/server.go`: HTTP server and route registration
- `internal/api/handlers.go`: Request handlers for all API endpoints
- `internal/provision/engine.go`: Instance provisioning orchestration
- `internal/availability/poller.go`: Background polling worker
- `internal/health/monitor.go`: Background health monitoring worker

**Data Layer:**
- `internal/db/pool.go`: pgx connection pool
- `internal/db/instances.go`: Instance CRUD and queries
- `internal/db/organizations.go`: Org/user CRUD
- `internal/db/ssh_keys.go`: SSH key management
- `database/migrations/20250224_v0.sql`: Database schema

**Provider Integration:**
- `internal/provider/provider.go`: Interface definition all providers implement
- `internal/provider/types.go`: Shared types (GPUOffering, ProvisionRequest)
- `internal/provider/runpod/adapter.go`: RunPod implementation
- `internal/provider/runpod/adapter_test.go`: RunPod tests

**Background Workers:**
- `internal/availability/poller.go`: Polls providers every 30s
- `internal/health/monitor.go`: Checks instance health every 60s

**Cross-Cutting Concerns:**
- `internal/auth/clerk.go`: JWT verification middleware
- `internal/wireguard/manager.go`: Peer management
- `internal/wireguard/keygen.go`: Key generation
- `internal/billing/stripe.go`: Stripe integration

**Testing:**
- `internal/provider/runpod/adapter_test.go`: Example test file

## Naming Conventions

**Files:**
- Lowercase snake_case with `.go` extension
- Packages use same name as directory
- Main entrypoint: `main.go` in `cmd/[binary]/`
- Example: `internal/db/instances.go` is package `db`

**Directories:**
- Lowercase, single-word where possible
- Functional organization (api, db, auth, billing, etc.)
- Provider-specific code in subdirectories: `internal/provider/runpod/`
- Example: `internal/availability/`, `internal/wireguard/`

**Functions:**
- Public functions: PascalCase (exported)
- Private functions: camelCase (unexported)
- Receivers: Short abbreviations (e.g., `func (p *Pool)`, `func (s *Server)`)
- Example: `func (s *Server) HandleListInstances()`, `func (e *Engine) Provision()`

**Types:**
- PascalCase (exported)
- Interface names end with "er": Provider, Cache, Monitor
- Request/Response types: `{Action}Request`, `{Action}Response`
- Example: `type Provider interface`, `type ProvisionRequest struct`

**Constants:**
- PascalCase (exported)
- Enum-like values: Prefixed with type name
- Example: `const GPUTypeH100SXM GPUType = "h100_sxm"`, `const TierOnDemand InstanceTier = "on_demand"`

## Where to Add New Code

**New Feature (e.g., volume management):**
- Primary code: Create new file `internal/[domain]/volume.go` or new package `internal/volume/`
- Database: Add tables/queries to `internal/db/volume.go`, update migration file
- API: Add handlers to `internal/api/handlers.go`, register routes in `server.go`
- Tests: Add `internal/[domain]/volume_test.go`

**New Provider Adapter (e.g., Lambda Labs):**
- Implementation: `internal/provider/lambda/adapter.go`
- Tests: `internal/provider/lambda/adapter_test.go`
- Registration: Update `cmd/gpuctl/main.go` provider map initialization
- No new files needed in other layers (uses existing Provider interface)

**Utility Functions:**
- Shared helpers across packages: Create `internal/util/` or `internal/helpers/` package
- Package-local utilities: Keep in the package file where used
- Example: WireGuard key generation is in `internal/wireguard/keygen.go` (used by provision engine)

**New Background Worker:**
- Implementation: `internal/[concern]/worker.go` or `internal/[concern]/monitor.go`
- Registration: Start goroutine in `cmd/gpuctl/main.go` main() function
- Context: Accept `context.Context` as first parameter, stop on context cancellation
- Example: `internal/availability/poller.go` follows this pattern

**New Database Table:**
- Schema: Add to `database/migrations/` as new timestamped file
- Queries: Implement in new file under `internal/db/` (e.g., `internal/db/domains.go`)
- Examples: `internal/db/instances.go`, `internal/db/organizations.go`

## Special Directories

**cmd/**
- Purpose: Executable binaries
- Generated: No
- Committed: Yes
- Notes: Only contains main functions; business logic in `internal/`

**internal/**
- Purpose: Private Go packages (not importable by external code)
- Generated: No
- Committed: Yes
- Notes: Enforces encapsulation — all business logic here

**database/migrations/**
- Purpose: SQL schema versioning
- Generated: No
- Committed: Yes
- Notes: Run via `python tools/migrate.py` or manual psql

**infra/**
- Purpose: Infrastructure configuration and templates
- Generated: No
- Committed: Yes
- Notes: Cloud-init template, WireGuard config examples

**tools/**
- Purpose: Development and operational tooling (Python)
- Generated: No
- Committed: Yes
- Notes: `migrate.py`, `seed.py` for local development

**docs/**
- Purpose: Architecture and design documentation
- Generated: No
- Committed: Yes
- Notes: `ARCHITECTURE.md` provides system design

---

*Structure analysis: 2026-02-24*
