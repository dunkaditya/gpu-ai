# Phase 4: Auth + Instance Lifecycle - Research

**Researched:** 2026-02-24
**Domain:** Clerk JWT auth, REST instance lifecycle, state machine, rate limiting, idempotency, SSE, pagination
**Confidence:** HIGH

## Summary

Phase 4 wires together authentication, instance CRUD, and cross-cutting API concerns (rate limiting, idempotency, pagination, SSE) on top of the existing foundation (Phases 1-3). The Clerk Go SDK v2 provides production-ready JWT middleware that extracts `user_id` (via `Subject`) and `org_id` (via `ActiveOrganizationID`) directly into request context. The instance state machine is a simple map-of-valid-transitions pattern (no library needed). Rate limiting uses `golang.org/x/time/rate` as the user specified. Idempotency keys are stored in PostgreSQL with a 24h TTL as decided. SSE uses Go's stdlib `http.Flusher` interface. Cursor-based pagination uses keyset pagination on `(created_at, instance_id)`.

The project already has `provider.CustomerInstance` (defense-by-omission struct), `provider.Instance.ToCustomer()`, and the full WireGuard + cloud-init pipeline from Phase 3. The existing `internal/api/handlers.go` is a TODO stub with the exact handler signatures needed. The `internal/auth/clerk.go` is also a TODO stub ready for implementation. The provisioning engine (`internal/provision/engine.go`) is a TODO stub that needs to be implemented as part of this phase.

**Primary recommendation:** Use the Clerk Go SDK v2 `RequireHeaderAuthorization()` middleware for all `/api/v1/*` routes, implement the state machine as a simple Go map of allowed transitions (no FSM library), and keep all cross-cutting concerns (rate limiting, idempotency, pagination) as composable middleware/helpers in `internal/api/`.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Single-step creation: POST /api/v1/instances creates the instance and returns confirmed `price_per_hour` in the response
- No separate quote/confirm step -- engine queries Redis cache for current pricing at creation time
- Optional `max_price_per_hour` field in request body -- engine rejects with 409 if current price exceeds it
- Request body: flat structure with `gpu_type` (e.g. "A100_80GB"), `gpu_count` (1-8), `region` (optional), `tier` ("spot" or "on-demand")
- `ssh_key_ids` is required -- reject if empty (no way to access instance = burning money)
- Optional `name` field for user display label; system always auto-generates branded hostname (e.g. gpu-xyz.gpu.ai)
- Full internal states: creating -> provisioning -> booting -> running -> stopping -> terminated, plus error
- Simplified external (API-facing) states: `starting` (maps to creating+provisioning+booting), `running`, `stopping`, `terminated`, `error`
- Terminate allowed from any non-terminal state -- system cancels provision with provider if still in progress
- Error state includes `error_reason` field explaining what went wrong
- Instances in error state are separate from terminated -- user can distinguish failed vs intentionally stopped
- SSE endpoint for real-time status push -- dashboard opens SSE connection after instance creation, server pushes status changes via Go channels
- Ready callback handler updates DB and writes to Go channel; SSE handler reads channel and pushes to client
- One goroutine per active SSE connection, cleaned up on connection close
- API/CLI users can poll GET /api/v1/instances/{id} as fallback
- Error format: RFC 7807 Problem Details (`type`, `title`, `status`, `detail`)
- Instance detail: nested `connection` object with `hostname`, `port`, `ssh_command`
- Lifecycle timestamps: `created_at`, `ready_at`, `terminated_at` -- all RFC 3339 format
- List responses: `{"data": [...], "cursor": "abc", "has_more": true}` -- no expensive COUNT query
- In-memory token bucket using `golang.org/x/time/rate` (Go stdlib), scoped per org
- No Redis for rate limiting -- single binary deployment, no shared state needed yet
- Minimal response headers: just 429 status + `Retry-After` header on rate limit hit
- `Idempotency-Key` header on POST /api/v1/instances
- Keys stored in PostgreSQL table with 24h TTL, scoped to org_id
- Duplicate request within TTL returns stored response -- prevents double-provision on retry
- DB-backed to survive server restarts

### Claude's Discretion
- Default and max page sizes for cursor-based pagination
- SSE endpoint path and event format details
- Exact rate limit thresholds per endpoint (creates vs reads)
- Internal state machine transition validation logic
- Auth middleware structure and context propagation pattern

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AUTH-01 | All customer API endpoints require valid Clerk JWT | Clerk Go SDK v2 `RequireHeaderAuthorization()` middleware; wraps all `/api/v1/*` routes |
| AUTH-02 | JWT verification extracts user_id and org_id into request context | `clerk.SessionClaimsFromContext()` provides `claims.Subject` (user_id) and `claims.ActiveOrganizationID` (org_id) |
| AUTH-03 | Users can only access instances belonging to their organization | DB queries filter by `org_id` extracted from JWT claims; handler verifies org ownership on single-instance lookups |
| INST-01 | User can create a GPU instance specifying type, count, region, tier, SSH keys | POST handler decodes request, validates, calls provisioning engine, returns `CustomerInstance` with `price_per_hour` |
| INST-02 | User can list their active instances with status and connection info | GET handler queries DB with `org_id` filter, cursor-based pagination, returns `CustomerInstance` array |
| INST-03 | User can get details of a specific instance by ID | GET handler fetches by `instance_id`, verifies `org_id` match, returns `CustomerInstance` with nested `connection` object |
| INST-04 | User can terminate an instance and billing stops | DELETE handler validates ownership, calls provider.Terminate(), updates DB status+billing_end+terminated_at |
| INST-05 | Instance follows state machine (creating -> provisioning -> booting -> running -> stopping -> terminated) | Map-based state machine with allowed transitions; internal states map to simplified external states |
| INST-06 | Instance termination is idempotent (multiple calls produce same result) | DELETE returns 200 if already terminated; idempotency-key middleware prevents duplicate provisions |
| INST-07 | Instance ready callback transitions status from booting to running | Internal endpoint `POST /internal/instances/{id}/ready` updates status, sets `ready_at`, writes to SSE channel |
| INST-08 | Instance creation response includes confirmed hourly cost | Engine queries Redis cache for current pricing, includes `price_per_hour` in response |
| API-01 | POST /api/v1/instances creates a new GPU instance | Handler implementation with request validation, provisioning engine call, response serialization |
| API-02 | GET /api/v1/instances lists user's instances | Handler with cursor-based pagination, org-scoped query, `CustomerInstance` mapping |
| API-03 | GET /api/v1/instances/{id} returns instance details | Handler with org ownership check, `CustomerInstance` with `connection` object |
| API-04 | DELETE /api/v1/instances/{id} terminates an instance | Handler with org ownership, idempotent termination, state machine transition |
| API-09 | Error responses never leak upstream provider details | RFC 7807 Problem Details struct excludes provider fields; `CustomerInstance` defense-by-omission already in codebase |
| API-10 | All list endpoints support cursor-based pagination | Keyset pagination on `(created_at, instance_id)` with base64-encoded cursor; default 20, max 100 |
| API-11 | POST /api/v1/instances accepts Idempotency-Key header | PostgreSQL `idempotency_keys` table with 24h TTL, org-scoped; middleware checks before handler executes |
| API-12 | All customer API endpoints are rate-limited per org | `golang.org/x/time/rate` token bucket per org_id in sync.Map; middleware returns 429 + Retry-After |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/clerk/clerk-sdk-go/v2` | v2.5.1 | JWT verification + auth middleware | Official Clerk SDK; provides `RequireHeaderAuthorization()`, `SessionClaimsFromContext()` |
| `golang.org/x/time/rate` | v0.14.0 | Token bucket rate limiter | Go extended stdlib; user-specified choice; no external deps |
| `github.com/jackc/pgx/v5` | v5.8.0 | PostgreSQL queries (already in go.mod) | Already used; instance CRUD queries, idempotency key storage |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis cache reads (already in go.mod) | Already used; availability cache reads for pricing |
| `net/http` (stdlib) | Go 1.24 | HTTP server, routing, SSE via Flusher | Project convention; no frameworks |
| `encoding/json` (stdlib) | Go 1.24 | JSON serialization | Project convention |
| `log/slog` (stdlib) | Go 1.24 | Structured logging | Project convention |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/base64` (stdlib) | Go 1.24 | Cursor encoding/decoding | Pagination cursor serialization |
| `crypto/rand` (stdlib) | Go 1.24 | Generate instance IDs and tokens | Instance ID generation (e.g., "gpu-a1b2"), internal tokens |
| `sync` (stdlib) | Go 1.24 | sync.Map for rate limiter storage | Per-org rate limiter map |
| `time` (stdlib) | Go 1.24 | TTL management, timestamps | Idempotency key expiry, rate limiter cleanup |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled state machine map | `looplab/fsm` or `qmuntal/stateless` | Libraries add dep for a ~30-line map; overkill for 7 states |
| Hand-rolled RFC 7807 struct | `mvmaasakkers/go-problemdetails` | Simple struct + marshal is ~20 lines; library adds unnecessary dep |
| In-memory rate limiter | Redis-based (`go-redis/redis_rate`) | User explicitly chose in-memory for single-binary deployment |
| `Idempotency-Key` in Redis | PostgreSQL table | User explicitly chose DB-backed for restart survival |

**Installation:**
```bash
go get github.com/clerk/clerk-sdk-go/v2@v2.5.1
go get golang.org/x/time/rate
```

## Architecture Patterns

### Recommended Project Structure
```
internal/api/
├── server.go           # Server struct, NewServer, route registration
├── middleware.go        # LocalhostOnly, InternalAuthMiddleware (existing)
├── middleware_auth.go   # Clerk JWT middleware wrapper + claims helper
├── middleware_rate.go   # Per-org rate limiting middleware
├── handlers.go          # Existing TODO stubs -> implement instance handlers
├── handlers_internal.go # Internal callback handlers (ready, health)
├── idempotency.go       # Idempotency-Key middleware + DB helpers
├── pagination.go        # Cursor encode/decode, pagination params
├── sse.go              # SSE handler for instance status streaming
├── errors.go           # RFC 7807 ProblemDetail struct + helpers
├── middleware_test.go   # Existing tests
├── handlers_test.go     # Handler unit tests
└── idempotency_test.go  # Idempotency tests

internal/auth/
├── clerk.go            # Existing TODO -> Clerk verifier + Claims type + context helpers

internal/provision/
├── engine.go           # Existing TODO -> Full provisioning orchestration

internal/db/
├── pool.go             # Existing connection pool
├── instances.go        # Existing TODO -> Instance CRUD queries
├── idempotency.go      # Idempotency key CRUD queries
└── organizations.go    # Existing TODO -> Org/user queries
```

### Pattern 1: Clerk JWT Auth Middleware Chain
**What:** Wrap all `/api/v1/*` routes with Clerk's `RequireHeaderAuthorization()`, then extract claims in handlers via a thin project-specific helper.
**When to use:** Every customer-facing endpoint.
**Example:**
```go
// Source: https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2/http
import (
    "github.com/clerk/clerk-sdk-go/v2"
    clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

// In server.go route registration:
clerkAuth := clerkhttp.RequireHeaderAuthorization()

// Wrap each handler:
s.mux.Handle("POST /api/v1/instances",
    clerkAuth(rateLimiter(http.HandlerFunc(s.handleCreateInstance))))
s.mux.Handle("GET /api/v1/instances",
    clerkAuth(rateLimiter(http.HandlerFunc(s.handleListInstances))))

// In handler:
func (s *Server) handleListInstances(w http.ResponseWriter, r *http.Request) {
    claims, ok := clerk.SessionClaimsFromContext(r.Context())
    if !ok {
        writeProblem(w, http.StatusUnauthorized, "unauthorized", "Missing or invalid session")
        return
    }
    orgID := claims.ActiveOrganizationID
    userID := claims.Subject
    // ... query DB with orgID
}
```

### Pattern 2: Instance State Machine (Map-based)
**What:** Define valid transitions as a map; validate before any DB update.
**When to use:** Every state change in the instance lifecycle.
**Example:**
```go
// Internal states (stored in DB)
const (
    StateCreating     = "creating"
    StateProvisioning = "provisioning"
    StateBooting      = "booting"
    StateRunning      = "running"
    StateStopping     = "stopping"
    StateTerminated   = "terminated"
    StateError        = "error"
)

// Valid transitions: from -> []to
var validTransitions = map[string][]string{
    StateCreating:     {StateProvisioning, StateError, StateStopping},
    StateProvisioning: {StateBooting, StateError, StateStopping},
    StateBooting:      {StateRunning, StateError, StateStopping},
    StateRunning:      {StateStopping, StateError},
    StateStopping:     {StateTerminated, StateError},
    // Terminal states: no outgoing transitions
    StateTerminated:   {},
    StateError:        {StateStopping}, // Allow retry-terminate from error
}

func CanTransition(from, to string) bool {
    allowed, ok := validTransitions[from]
    if !ok { return false }
    for _, s := range allowed {
        if s == to { return true }
    }
    return false
}

// Map internal state to external (API-facing) state
func ExternalState(internal string) string {
    switch internal {
    case StateCreating, StateProvisioning, StateBooting:
        return "starting"
    case StateRunning:
        return "running"
    case StateStopping:
        return "stopping"
    case StateTerminated:
        return "terminated"
    case StateError:
        return "error"
    default:
        return "unknown"
    }
}
```

### Pattern 3: Per-Org Rate Limiter
**What:** Maintain a `sync.Map` of `org_id -> *rate.Limiter`. Middleware extracts org from claims and checks `Allow()`.
**When to use:** All customer API endpoints.
**Example:**
```go
import "golang.org/x/time/rate"

type OrgRateLimiter struct {
    limiters sync.Map // org_id -> *rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewOrgRateLimiter(r rate.Limit, burst int) *OrgRateLimiter {
    return &OrgRateLimiter{rate: r, burst: burst}
}

func (o *OrgRateLimiter) GetLimiter(orgID string) *rate.Limiter {
    if v, ok := o.limiters.Load(orgID); ok {
        return v.(*rate.Limiter)
    }
    limiter := rate.NewLimiter(o.rate, o.burst)
    actual, _ := o.limiters.LoadOrStore(orgID, limiter)
    return actual.(*rate.Limiter)
}

func (o *OrgRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims, ok := clerk.SessionClaimsFromContext(r.Context())
        if !ok {
            next.ServeHTTP(w, r) // auth middleware handles this
            return
        }
        limiter := o.GetLimiter(claims.ActiveOrganizationID)
        if !limiter.Allow() {
            w.Header().Set("Retry-After", "1")
            writeProblem(w, http.StatusTooManyRequests, "rate_limited",
                "Too many requests. Try again later.")
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Pattern 4: SSE with http.Flusher
**What:** Stream instance status updates to the client using Server-Sent Events. One goroutine per connection, communicating via Go channels.
**When to use:** Real-time status push for instance creation/termination.
**Example:**
```go
func (s *Server) handleInstanceSSE(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "SSE not supported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("X-Accel-Buffering", "no")

    instanceID := r.PathValue("id")
    ch := s.statusBroker.Subscribe(instanceID)
    defer s.statusBroker.Unsubscribe(instanceID, ch)

    // Send current state immediately
    // ...

    keepAlive := time.NewTicker(30 * time.Second)
    defer keepAlive.Stop()

    for {
        select {
        case <-r.Context().Done():
            return
        case event := <-ch:
            fmt.Fprintf(w, "event: status\ndata: %s\n\n", event)
            flusher.Flush()
        case <-keepAlive.C:
            fmt.Fprintf(w, ": keepalive\n\n")
            flusher.Flush()
        }
    }
}
```

### Pattern 5: Cursor-Based Pagination (Keyset)
**What:** Encode `(created_at, instance_id)` as a base64 cursor. Use `WHERE (created_at, instance_id) < ($cursor_time, $cursor_id)` for the next page.
**When to use:** List endpoints (instances, future ssh-keys, usage).
**Example:**
```go
type PageParams struct {
    Cursor string // base64-encoded "created_at|instance_id"
    Limit  int    // default 20, max 100
}

type PageResult[T any] struct {
    Data    []T    `json:"data"`
    Cursor  string `json:"cursor,omitempty"`
    HasMore bool   `json:"has_more"`
}

// SQL for next page (descending by created_at):
// SELECT ... FROM instances
// WHERE org_id = $1
//   AND (created_at, instance_id) < ($2, $3)
// ORDER BY created_at DESC, instance_id DESC
// LIMIT $4 + 1  -- fetch one extra to determine has_more
```

### Pattern 6: Idempotency-Key Middleware
**What:** Check `Idempotency-Key` header on POST requests. If key exists in DB and completed, return cached response. If key is new, execute handler and store response.
**When to use:** POST /api/v1/instances only (per CONTEXT.md).
**Example:**
```go
// DB schema for idempotency keys:
// CREATE TABLE idempotency_keys (
//     idempotency_key VARCHAR(255) NOT NULL,
//     org_id          UUID NOT NULL,
//     response_code   INT,
//     response_body   JSONB,
//     created_at      TIMESTAMPTZ DEFAULT NOW(),
//     PRIMARY KEY (org_id, idempotency_key)
// );
// CREATE INDEX idx_idempotency_keys_created ON idempotency_keys(created_at);

func IdempotencyMiddleware(db *db.Pool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := r.Header.Get("Idempotency-Key")
            if key == "" {
                next.ServeHTTP(w, r)
                return
            }
            // Check for existing key in DB...
            // If found and completed: replay stored response
            // If found and in-progress: return 409 Conflict
            // If not found: create record, execute handler, store response
        })
    }
}
```

### Pattern 7: RFC 7807 Problem Details
**What:** Standardized error response format with `type`, `title`, `status`, `detail`.
**When to use:** Every error response from customer-facing endpoints.
**Example:**
```go
type ProblemDetail struct {
    Type   string `json:"type"`
    Title  string `json:"title"`
    Status int    `json:"status"`
    Detail string `json:"detail"`
}

func writeProblem(w http.ResponseWriter, status int, errType, detail string) {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ProblemDetail{
        Type:   "https://api.gpu.ai/errors/" + errType,
        Title:  http.StatusText(status),
        Status: status,
        Detail: detail,
    })
}
```

### Anti-Patterns to Avoid
- **Leaking provider in error messages:** Never include upstream error details in customer-facing responses. Log the full error internally, return a generic ProblemDetail externally.
- **COUNT(*) for pagination:** Never use `SELECT COUNT(*)` for total count in list responses. Use `has_more` flag by fetching `limit + 1` rows.
- **Serializing `provider.Instance` directly:** Always use `CustomerInstance` (defense-by-omission). Never pass the internal Instance struct to `json.Marshal` in an HTTP response.
- **Blocking SSE goroutines:** Always select on `r.Context().Done()` to clean up when the client disconnects. Never block indefinitely on a channel without context cancellation.
- **Shared rate limiter state across orgs:** Each org gets its own `rate.Limiter` instance. Do not use a single global limiter.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT verification | Custom JWKS fetching + JWT parsing | `clerk-sdk-go/v2` RequireHeaderAuthorization | JWKS caching, key rotation, clock skew, standard claims extraction |
| Token bucket rate limiting | Custom counter/timer | `golang.org/x/time/rate` | Thread-safe, well-tested, handles burst and refill correctly |
| WireGuard key generation | Manual crypto | Already implemented in `internal/wireguard/keygen.go` | Exists from Phase 3 |
| Cloud-init rendering | String concatenation | Already implemented in `internal/wireguard/template.go` | Exists from Phase 3 with input validation |
| Instance ID generation | Sequential IDs | `crypto/rand` hex bytes with `gpu-` prefix | Unpredictable, collision-resistant, branded |

**Key insight:** This phase is mostly wiring together existing building blocks (provider interface, WireGuard, cloud-init) with new middleware (auth, rate limiting, idempotency) and DB queries. The only genuinely new complexity is the provisioning engine orchestration and the SSE status broker.

## Common Pitfalls

### Pitfall 1: Clerk org_id May Be Empty
**What goes wrong:** `claims.ActiveOrganizationID` is empty string when the user has no active organization selected in their Clerk session.
**Why it happens:** Clerk only populates org claims when the user has an active organization. New users or users not in any org will have an empty `ActiveOrganizationID`.
**How to avoid:** After extracting claims, check that `ActiveOrganizationID` is non-empty. Return 403 with a clear error message ("organization required") if empty. This also covers the case where a user exists but hasn't joined an org yet.
**Warning signs:** Tests pass with hardcoded org_id but fail with real Clerk tokens; instances created without org_id.

### Pitfall 2: State Machine Race Conditions
**What goes wrong:** Two concurrent requests both read status="booting" and both try to transition to "running", or a terminate request races with a ready callback.
**Why it happens:** Read-then-write without atomicity. DB reads current status, Go validates transition, DB updates -- but between read and write, another request can change the state.
**How to avoid:** Use optimistic locking: `UPDATE instances SET status = $new WHERE instance_id = $id AND status = $old RETURNING instance_id`. If zero rows affected, the state changed between read and write -- retry or return conflict.
**Warning signs:** Duplicate "running" log entries; terminated instances receiving "ready" callback after termination.

### Pitfall 3: SSE Connection Leak
**What goes wrong:** Goroutines accumulate when clients disconnect without the server detecting it.
**Why it happens:** The HTTP request context is not cancelled if the connection drops silently (e.g., mobile network switch). Without keep-alive pings, the write may block or succeed into a buffered pipe.
**How to avoid:** 1) Always send periodic keep-alive comments (`:keepalive\n\n`). 2) Check write errors -- a failed write means the client disconnected. 3) Use `r.Context().Done()` as the primary exit signal. 4) Set a maximum SSE connection duration (e.g., 30 minutes) to prevent zombie connections.
**Warning signs:** Increasing goroutine count over time; OOM on the server.

### Pitfall 4: Idempotency Key Without Request Fingerprint
**What goes wrong:** Client sends same idempotency key with different request bodies. Server returns the cached response from the first request, silently ignoring the changed parameters.
**Why it happens:** Only checking the key without verifying the request body matches.
**How to avoid:** Store a hash of the request body alongside the idempotency key. On duplicate key with different body hash, return 422 Unprocessable Entity explaining the mismatch.
**Warning signs:** Users report "wrong instance type" when retrying with different parameters.

### Pitfall 5: Rate Limiter Memory Leak
**What goes wrong:** `sync.Map` accumulates limiter entries for orgs that are no longer active, growing unboundedly.
**Why it happens:** Limiters are created on first request but never cleaned up.
**How to avoid:** Run a periodic cleanup goroutine (every 5 minutes) that removes entries not accessed in the last 10 minutes. Wrap the limiter in a struct with a `lastSeen` timestamp.
**Warning signs:** Gradually increasing memory usage; sync.Map growing without bound.

### Pitfall 6: Pagination Cursor Tampering
**What goes wrong:** Client modifies the base64 cursor to access instances from other orgs, or crafts cursors that cause SQL injection.
**Why it happens:** Cursor is decoded and used directly in SQL without validation.
**How to avoid:** 1) Always include `org_id` in the WHERE clause regardless of cursor content. 2) Parse cursor values as strongly-typed Go values (time.Time, string) before using in parameterized queries. 3) Never interpolate cursor values into SQL strings.
**Warning signs:** Security audit finds org data leakage via cursor manipulation.

## Code Examples

### Complete Instance Creation Handler Flow
```go
// Source: Project-specific pattern combining Clerk SDK + provisioning engine

func (s *Server) handleCreateInstance(w http.ResponseWriter, r *http.Request) {
    // 1. Extract auth claims (Clerk middleware already verified JWT)
    claims, ok := clerk.SessionClaimsFromContext(r.Context())
    if !ok {
        writeProblem(w, http.StatusUnauthorized, "unauthorized", "Invalid session")
        return
    }
    if claims.ActiveOrganizationID == "" {
        writeProblem(w, http.StatusForbidden, "org_required", "Active organization required")
        return
    }

    // 2. Decode and validate request body
    var req CreateInstanceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeProblem(w, http.StatusBadRequest, "invalid_body", "Invalid JSON request body")
        return
    }
    if err := req.Validate(); err != nil {
        writeProblem(w, http.StatusBadRequest, "validation_error", err.Error())
        return
    }

    // 3. Query availability cache for pricing
    // 4. Check max_price_per_hour constraint
    // 5. Generate instance ID, WireGuard keys, render cloud-init
    // 6. Call provisioning engine
    // 7. Write instance to DB with status="creating"
    // 8. Return CustomerInstance with price_per_hour

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(instanceResponse)
}
```

### Idempotency Key DB Schema
```sql
-- Migration for Phase 4: idempotency_keys table
CREATE TABLE idempotency_keys (
    idempotency_key VARCHAR(255) NOT NULL,
    org_id          UUID         NOT NULL REFERENCES organizations(organization_id),
    request_hash    VARCHAR(64)  NOT NULL, -- SHA-256 of request body
    response_code   INT,
    response_body   JSONB,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    PRIMARY KEY (org_id, idempotency_key)
);

-- Index for TTL cleanup (delete keys older than 24h)
CREATE INDEX idx_idempotency_keys_created ON idempotency_keys(created_at);
```

### Instance Response Shape (Customer-Facing)
```json
{
    "id": "gpu-a1b2c3",
    "name": "my-training-run",
    "status": "starting",
    "gpu_type": "a100_80gb",
    "gpu_count": 4,
    "tier": "on_demand",
    "region": "us-west",
    "price_per_hour": 2.12,
    "connection": {
        "hostname": "gpu-a1b2c3.gpu.ai",
        "port": 10042,
        "ssh_command": "ssh root@gpu-a1b2c3.gpu.ai -p 10042"
    },
    "error_reason": null,
    "created_at": "2026-02-24T10:30:00Z",
    "ready_at": null,
    "terminated_at": null
}
```

### List Response Shape
```json
{
    "data": [
        { "id": "gpu-a1b2c3", "status": "running", "..." : "..." },
        { "id": "gpu-d4e5f6", "status": "starting", "..." : "..." }
    ],
    "cursor": "eyJjcmVhdGVkX2F0IjoiMjAyNi0wMi0yNFQxMDozMDowMFoiLCJpZCI6ImdwdS1kNGU1ZjYifQ==",
    "has_more": true
}
```

### SSE Event Format
```
event: status
data: {"instance_id":"gpu-a1b2c3","status":"starting","internal_status":"provisioning","timestamp":"2026-02-24T10:30:05Z"}

event: status
data: {"instance_id":"gpu-a1b2c3","status":"starting","internal_status":"booting","timestamp":"2026-02-24T10:30:10Z"}

event: status
data: {"instance_id":"gpu-a1b2c3","status":"running","internal_status":"running","timestamp":"2026-02-24T10:30:25Z"}

: keepalive

```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Clerk SDK v1 (`clerkinc/clerk-sdk-go`) | Clerk SDK v2 (`clerk/clerk-sdk-go/v2`) | 2024 | New module path, `SessionClaimsFromContext` instead of custom context key |
| Offset-based pagination | Cursor/keyset pagination | Widely adopted 2020+ | Consistent performance at any depth; required for real-time data |
| Custom JWT parsing with `golang-jwt` | Clerk SDK built-in verification | Always for Clerk | Handles JWKS rotation, caching, clock skew automatically |
| RFC 7807 (Problem Details) | RFC 9457 (supersedes 7807, same format) | 2023 | Identical wire format; 9457 is the current standard |

**Deprecated/outdated:**
- `clerkinc/clerk-sdk-go` (v1): Replaced by `clerk/clerk-sdk-go/v2`. Different module path.
- Manual JWKS fetching: The Clerk SDK handles this automatically with 1-hour caching.

## Open Questions

1. **Clerk API Key Configuration**
   - What we know: The SDK requires `clerk.SetKey("sk_...")` or the `CLERK_SECRET_KEY` env var to be set for JWKS fetching.
   - What's unclear: Whether the project already has a Clerk account provisioned or if this is a dev-only concern.
   - Recommendation: Add `CLERK_SECRET_KEY` to config.go as an optional field (empty = skip Clerk auth, for local dev). This allows testing without a Clerk account.

2. **SSH Key IDs vs SSH Key Content**
   - What we know: CONTEXT.md says `ssh_key_ids` is required in the create request. The `ssh_keys` table exists but its CRUD is Phase 5.
   - What's unclear: Whether Phase 4 should implement the SSH key lookup (`ssh_key_ids -> public key content`) or defer to Phase 5.
   - Recommendation: Phase 4 should implement the DB query to resolve `ssh_key_ids` to public key content for the provisioning engine. Phase 5 adds the management CRUD. If no SSH keys exist yet, the create request will naturally fail validation.

3. **Instance Name Column**
   - What we know: CONTEXT.md specifies optional `name` field for user display label. The current DB schema has no `name` column on instances.
   - What's unclear: Needs a migration to add it.
   - Recommendation: Add `name VARCHAR(255)` to the instances table in the Phase 4 migration, alongside the `idempotency_keys` table and any other schema changes.

4. **`ready_at` Timestamp Column**
   - What we know: CONTEXT.md specifies `ready_at` in the API response timestamps. The current DB schema has `created_at` and `terminated_at` but no `ready_at`.
   - What's unclear: Needs a migration.
   - Recommendation: Add `ready_at TIMESTAMPTZ` to the instances table in the Phase 4 migration.

5. **`error_reason` Column**
   - What we know: CONTEXT.md specifies error state includes `error_reason` field. Not in current schema.
   - What's unclear: Needs a migration.
   - Recommendation: Add `error_reason TEXT` to the instances table in the Phase 4 migration.

6. **Rate Limiter Thresholds**
   - What we know: User deferred to Claude's discretion.
   - What's unclear: Exact values.
   - Recommendation: Instance creation: 10 req/min (burst 5). Read endpoints: 60 req/min (burst 20). These can be tuned via env vars later.

7. **Default and Max Page Sizes**
   - What we know: User deferred to Claude's discretion.
   - What's unclear: Exact values.
   - Recommendation: Default: 20, Max: 100. Standard API pagination sizes.

## Sources

### Primary (HIGH confidence)
- [Clerk Go SDK v2 HTTP package](https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2/http) - Middleware API, AuthorizationOption types
- [Clerk Go SDK v2 JWT package](https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2/jwt) - Verify function, VerifyParams
- [Clerk Go SDK v2 main package](https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2) - SessionClaims struct, SessionClaimsFromContext
- [Clerk session verification guide](https://clerk.com/docs/guides/sessions/verifying) - Official Go middleware examples
- [clerk-sdk-go/jwt.go source](https://github.com/clerk/clerk-sdk-go/blob/v2/jwt.go) - SessionClaims, RegisteredClaims, Claims struct definitions
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) - Limiter API, NewLimiter, Allow/Wait/Reserve

### Secondary (MEDIUM confidence)
- [Brandur: Stripe-like Idempotency Keys in Postgres](https://brandur.org/idempotency-keys) - Idempotency key schema, lifecycle, atomic phases
- [Keyset cursors for Postgres pagination](https://www.stacksync.com/blog/keyset-cursors-postgres-pagination-fast-accurate-scalable) - Keyset pagination implementation
- [Alex Edwards: Rate limiting HTTP requests in Go](https://www.alexedwards.net/blog/how-to-rate-limit-http-requests) - Per-client rate limiter pattern with sync.Map
- [Thoughtbot: Writing an SSE server in Go](https://thoughtbot.com/blog/writing-a-server-sent-events-server-in-go) - SSE handler pattern with http.Flusher

### Tertiary (LOW confidence)
- None -- all findings verified with primary or secondary sources.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Clerk SDK v2 verified via official docs and pkg.go.dev; rate limiter is Go extended stdlib; all other deps already in go.mod
- Architecture: HIGH - Patterns follow established Go stdlib conventions (middleware chaining, http.Flusher, parameterized SQL); state machine is trivial map-based validation
- Pitfalls: HIGH - Race conditions in state transitions and SSE leaks are well-documented patterns; idempotency key fingerprinting is established practice (Stripe, Brandur article)

**Research date:** 2026-02-24
**Valid until:** 2026-03-24 (Clerk SDK stable; Go stdlib patterns stable; all patterns well-established)
