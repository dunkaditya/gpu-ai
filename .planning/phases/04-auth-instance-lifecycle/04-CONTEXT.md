# Phase 4: Auth + Instance Lifecycle - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Authenticated REST API for creating, listing, viewing, and terminating GPU instances. Includes Clerk JWT auth, org-scoped access, instance state machine, rate limiting, idempotency, cursor-based pagination, and SSE for real-time status updates. SSH key management and billing are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Instance Creation Flow
- Single-step creation: POST /api/v1/instances creates the instance and returns confirmed `price_per_hour` in the response
- No separate quote/confirm step — engine queries Redis cache for current pricing at creation time
- Optional `max_price_per_hour` field in request body — engine rejects with 409 if current price exceeds it
- Request body: flat structure with `gpu_type` (e.g. "A100_80GB"), `gpu_count` (1-8), `region` (optional), `tier` ("spot" or "on-demand")
- `ssh_key_ids` is required — reject if empty (no way to access instance = burning money)
- Optional `name` field for user display label; system always auto-generates branded hostname (e.g. gpu-xyz.gpu.ai)

### State Machine
- Full internal states: creating → provisioning → booting → running → stopping → terminated, plus error
- Simplified external (API-facing) states: `starting` (maps to creating+provisioning+booting), `running`, `stopping`, `terminated`, `error`
- Terminate allowed from any non-terminal state — system cancels provision with provider if still in progress
- Error state includes `error_reason` field explaining what went wrong
- Instances in error state are separate from terminated — user can distinguish failed vs intentionally stopped

### Status Updates
- SSE endpoint for real-time status push — dashboard opens SSE connection after instance creation, server pushes status changes via Go channels
- Ready callback handler updates DB and writes to Go channel; SSE handler reads channel and pushes to client
- One goroutine per active SSE connection, cleaned up on connection close
- API/CLI users can poll GET /api/v1/instances/{id} as fallback

### API Response Shape
- Error format: RFC 7807 Problem Details (`type`, `title`, `status`, `detail`)
- Instance detail: nested `connection` object with `hostname`, `port`, `ssh_command`
- Lifecycle timestamps: `created_at`, `ready_at`, `terminated_at` — all RFC 3339 format
- List responses: `{"data": [...], "cursor": "abc", "has_more": true}` — no expensive COUNT query

### Rate Limiting
- In-memory token bucket using `golang.org/x/time/rate` (Go stdlib), scoped per org
- No Redis for rate limiting — single binary deployment, no shared state needed yet
- Minimal response headers: just 429 status + `Retry-After` header on rate limit hit
- Move to Redis-backed when multi-instance deployment requires shared state

### Idempotency
- `Idempotency-Key` header on POST /api/v1/instances
- Keys stored in PostgreSQL table with 24h TTL, scoped to org_id
- Duplicate request within TTL returns stored response — prevents double-provision on retry
- DB-backed to survive server restarts

### Claude's Discretion
- Default and max page sizes for cursor-based pagination
- SSE endpoint path and event format details
- Exact rate limit thresholds per endpoint (creates vs reads)
- Internal state machine transition validation logic
- Auth middleware structure and context propagation pattern

</decisions>

<specifics>
## Specific Ideas

- Rate limiter: use `golang.org/x/time/rate` specifically — it's Go's standard token bucket
- SSE is "just HTTP" — no WebSocket upgrade, EventSource built into every browser
- Dashboard experience: click provision → spinner → flips to running with SSH command ready to copy (~15 seconds)
- Error responses must never leak upstream provider identity (provider name, IP, metadata)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-auth-instance-lifecycle*
*Context gathered: 2026-02-24*
