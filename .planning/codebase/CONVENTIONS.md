# Coding Conventions

**Analysis Date:** 2026-02-24

## Naming Patterns

**Files:**
- `_test.go` suffix for test files (e.g., `adapter_test.go`)
- Descriptive names matching package responsibility: `pool.go`, `instances.go`, `manager.go`, `poller.go`
- Package names are lowercase, single word when possible: `api`, `db`, `auth`, `provider`, `health`
- Subpackages use domain structure: `internal/provider/runpod/`, `internal/wireguard/`

**Functions:**
- CamelCase starting with lowercase for unexported: `newAdapter()`, `newPool()`
- CamelCase starting with uppercase for exported: `NewAdapter()`, `NewPool()`, `ListAvailable()`, `GetInstance()`
- Handler functions prefixed with `Handle`: `HandleListInstances()`, `HandleCreateInstance()`, `HandleInstanceStatus()`
- Method names are descriptive and action-oriented: `Provision()`, `Terminate()`, `AddPeer()`, `RemovePeer()`, `CheckBillingStatus()`
- Middleware uses `Middleware` suffix in function signature: `authMiddleware()`, though typically composed in middleware chain

**Variables:**
- CamelCase for local variables: `ctx`, `err`, `port`, `mux`, `server`, `provider`, `logger`
- Single letter vars acceptable for loop counters and error returns: `ctx`, `i`, `err`
- Descriptive names for longer-lived vars: `writeTimeout`, `readTimeout`, `idleTimeout`, `shutdownCtx`

**Types:**
- CamelCase starting with uppercase for exported structs: `Provider`, `GPUOffering`, `ProvisionRequest`, `ProvisionResult`, `InstanceStatus`
- Singular names for struct types (not plural): `Instance` not `Instances`, `SSHKey` not `SSHKeys`
- Enums use full type names as prefix: `GPUType` enum with `GPUTypeH100SXM`, `GPUTypeA10080GB`
- Enum values use UPPER_SNAKE_CASE pattern within type: `TierOnDemand`, `TierSpot`, `TierReserved`

**Constants:**
- Type-specific naming for typed constants: `const (GPUTypeH100SXM GPUType = "h100_sxm")`
- Magic numbers avoided; use named constants

## Code Style

**Formatting:**
- Gofmt standard (enforced via Go toolchain)
- 80-character line preference observed in documentation
- Indentation: tabs (Go standard)
- Method receiver naming: short abbreviation or no abbreviation: `(a *Adapter)`, `(s *Server)`, `(p *Pool)`, `(m *Manager)`

**Linting:**
- Tool: `golangci-lint` (configured in Makefile via `make lint`)
- File: `.golangci.yml` or similar (not detected in codebase, using defaults)

**Error Handling:**
- Explicit error checking: `if err != nil { return err }`
- Error wrapping for context (though not yet fully implemented, standard Go practice)
- Fatal errors in main: `log.Fatalf()` for startup failures
- Package errors return as last return value: `func ListAvailable(ctx context.Context) ([]GPUOffering, error)`
- Middleware returns early on auth/validation failure with HTTP status codes

**Import Organization:**
- Standard library imports first
- Blank line separator
- Third-party imports (github.com packages)
- No explicit local imports in most files (internal packages in same module)
- Example from `main.go`:
  ```go
  import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
  )
  ```

**Path Aliases:**
- No aliases detected; standard relative imports within module

## Logging

**Framework:** `log/slog` (structured logging) preferred, though `log` package used in `main.go`

**Migration path:**
- Current: `log.Printf()`, `log.Println()`, `log.Fatalf()` in entry point
- Future: Structured logging via `slog.Logger` in all packages
- All package TODOs specify: `logger *slog.Logger` as struct field
- Logging expected at package initialization and error conditions

**Patterns:**
- Startup logging: `log.Printf("gpuctl starting on :%s", port)`
- Fatal errors: `log.Fatalf("server error: %v", err)`
- Clean shutdown logging: `log.Println("shutting down...")`
- Package loggers will use: `logger.Info()`, `logger.Error()`, `logger.Warn()` (slog style)

## Comments

**When to Comment:**
- Package-level documentation comments required: `// Package api provides the HTTP server...`
- Interface documentation: explain contract for implementers
- Complex algorithms: describe approach in comments (e.g., polling interval, cache key patterns)
- TODO items: used extensively for planned implementation (19 instances found)

**JSDoc/TSDoc:**
- Not applicable (Go codebase)
- Standard Go comment style: `// FunctionName does X`

**Documentation Comments:**
- All exported types documented: `// GPUOffering represents an available GPU configuration...`
- All exported functions documented above signature
- Comments precede the declaration they document

## Function Design

**Size:**
- Single-function handlers ~20-30 lines (from `main.go`)
- Interface methods expected to be concise with focused responsibility
- Longer orchestration logic (engine, provisioning) documented in comments rather than inline

**Parameters:**
- Context always first parameter for I/O operations: `func (p *Pool) CreateInstance(ctx context.Context, inst *Instance) error`
- Request/response types passed by value for small structs, by pointer for large ones
- Receiver method convention: `func (s *Server) HandleListInstances(w http.ResponseWriter, r *http.Request)`
- Variadic parameters not common in observed patterns

**Return Values:**
- Error always last return value: `func ListAvailable(ctx context.Context) ([]GPUOffering, error)`
- Named return values not used in observed patterns
- Multiple returns acceptable for result + error pattern

## Module Design

**Exports:**
- Constructor functions: `NewPool()`, `NewAdapter()`, `NewManager()`, `NewService()`
- Exported struct fields when needed, but often encapsulated via getter functions
- Interfaces exported, implementations sometimes unexported (`Provider` interface, multiple adapters)

**Package Responsibility:**
- `internal/api/` - HTTP handlers and server setup
- `internal/provider/` - Provider interface and type definitions
- `internal/provider/runpod/` - RunPod-specific adapter
- `internal/db/` - Database connection and query helpers
- `internal/auth/` - Authentication (Clerk JWT) verification
- `internal/billing/` - Stripe integration
- `internal/provision/` - Orchestration engine
- `internal/wireguard/` - WireGuard peer management
- `internal/availability/` - GPU availability polling and caching
- `internal/health/` - Instance health monitoring
- `internal/config/` - Environment configuration loading

**No Barrel Files:**
- Each package imports directly: `internal/provider`, `internal/provider/runpod`
- Not using index files or re-exports (not observed in patterns)

## Middleware Pattern

**Handler Pattern:**
- HTTP handlers: `func(w http.ResponseWriter, r *http.Request)`
- Middleware composition: `mux.Handle("GET /path", authMiddleware(s.HandleListInstances))`
- Middleware returns `http.Handler`: `func authMiddleware(next http.Handler) http.Handler`

**Context Propagation:**
- Request context passed through: `r.Context()` available to handlers
- Claims injected into context for downstream retrieval
- All I/O functions require context: no context-less operations in business logic

## Dependency Injection

**Pattern:**
- Constructor injection: `NewServer(deps ServerDeps)`, `NewEngine(...)`
- Struct fields hold dependencies: `pool *db.Pool`, `providers map[string]provider.Provider`
- No global state; all dependencies explicit in structs
- Main entry point wires dependencies (currently scaffolded in TODOs)

## Error Handling

**Strategy:**
- Early return on error
- Function terminates immediately when error occurs
- Logging at point of error (future slog pattern)
- HTTP handlers return appropriate status codes (not yet implemented)

**Patterns:**
- Explicit nil check: `if err != nil { return err }`
- Fatal on startup: `log.Fatalf()` for unrecoverable initialization
- Middleware returns 401/403 for auth failures (planned in auth middleware)

---

*Convention analysis: 2026-02-24*
