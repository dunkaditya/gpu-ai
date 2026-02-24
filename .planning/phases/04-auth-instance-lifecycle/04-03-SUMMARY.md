---
phase: 04-auth-instance-lifecycle
plan: 03
subsystem: api
tags: [handlers, idempotency, sse, crud, rest, middleware, wireguard]

# Dependency graph
requires:
  - phase: 04-auth-instance-lifecycle
    provides: "Clerk auth middleware, RFC 7807 errors, rate limiter, pagination, v3 schema"
  - phase: 04-auth-instance-lifecycle
    provides: "Instance state machine, DB CRUD, provisioning engine, idempotency storage"
provides:
  - "Instance CRUD handlers (create, list, get, delete) at /api/v1/instances"
  - "Idempotency middleware for POST deduplication via Idempotency-Key header"
  - "SSE StatusBroker with subscribe/unsubscribe/publish for real-time status updates"
  - "SSE handler at /api/v1/instances/{id}/events with keepalive and 30-min max"
  - "Internal ready callback at /internal/instances/{id}/ready (booting->running)"
  - "Internal health ping stub at /internal/instances/{id}/health"
  - "Full route registration with Clerk auth + RequireOrg + rate limiter chain"
  - "main.go wiring: provider registry, RunPod adapter, provisioning engine"
affects: [05-billing, 06-dashboard, 07-frontend]

# Tech tracking
tech-stack:
  added: []
  patterns: [idempotency-middleware, sse-status-streaming, response-capture-writer, auth-chain-routing]

key-files:
  created:
    - internal/api/idempotency.go
    - internal/api/sse.go
    - internal/api/handlers_internal.go
  modified:
    - internal/api/handlers.go
    - internal/api/server.go
    - cmd/gpuctl/main.go

key-decisions:
  - "InstanceResponse uses defense-by-omission: no provider fields exist in struct, cannot leak"
  - "SSE max connection duration 30 minutes with client reconnect expected"
  - "WriteTimeout set to 0 for SSE support; per-handler timeouts deferred to production hardening"
  - "Idempotency middleware uses SHA-256 body hash to detect key reuse with different bodies"
  - "Internal ready callback is idempotent: returns 200 even if already transitioned"
  - "Rate limiter: 10 req/s sustained with burst of 20 per org"

patterns-established:
  - "Auth chain: ClerkAuth -> RequireOrg -> RateLimiter -> Handler for all /api/v1/* routes"
  - "Response capture writer pattern for idempotency middleware response recording"
  - "SSE event loop: select on context.Done, channel event, keepalive ticker, max duration timer"
  - "Org resolution: ClaimsFromContext -> GetOrgIDByClerkOrgID for handler org scoping"

requirements-completed: [API-01, API-02, API-03, API-04, API-11, INST-07]

# Metrics
duration: 4min
completed: 2026-02-24
---

# Phase 4 Plan 03: API Handlers, Idempotency, SSE Streaming Summary

**Instance CRUD API with idempotency middleware, SSE real-time status streaming, and full route registration with Clerk auth chain wiring provider registry and provisioning engine in main.go**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-24T23:36:04Z
- **Completed:** 2026-02-24T23:40:20Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Instance CRUD handlers: POST create (with validation, engine call, 201 response), GET list (cursor-based pagination), GET detail (org-scoped), DELETE terminate (idempotent, 200 for already-terminated)
- Idempotency middleware: SHA-256 body hashing, response capture and replay, race condition handling with retry, 422 for key reuse with different body
- SSE status broker with subscribe/unsubscribe/publish and non-blocking event dispatch
- SSE handler with org ownership verification, initial state event, 30s keepalive pings, 30-min max duration
- Internal ready callback transitioning booting->running with SSE event publication
- Full route registration: 7 routes with proper middleware chains (Clerk auth, org check, rate limiter for public; localhost + internal token for callbacks)
- main.go wires provider registry (RunPod conditional on API key), provisioning engine, and passes to server

## Task Commits

Each task was committed atomically:

1. **Task 1: Instance handlers, idempotency middleware, and route registration** - `08746ae` (feat)
2. **Task 2: SSE status broker, SSE handler, and internal ready callback** - `a01e69c` (feat)

## Files Created/Modified
- `internal/api/handlers.go` - Instance CRUD handlers: create, list, get, delete with request validation and InstanceResponse mapping
- `internal/api/idempotency.go` - IdempotencyMiddleware with SHA-256 hashing, response capture, and replay
- `internal/api/sse.go` - StatusBroker (subscribe/unsubscribe/publish) and SSE handler with keepalive
- `internal/api/handlers_internal.go` - Internal ready callback (booting->running) and health ping stub
- `internal/api/server.go` - Route registration with auth chain, rate limiter, engine and broker fields
- `cmd/gpuctl/main.go` - Provider registry setup, RunPod registration, engine creation, server wiring

## Decisions Made
- InstanceResponse uses defense-by-omission: the struct has no upstream provider fields, making it impossible to accidentally serialize provider details to customers
- SSE connections have 30-minute max duration after which the server closes the connection; clients are expected to reconnect
- WriteTimeout set to 0 (disabled) to support SSE long-lived connections; per-handler timeouts for non-SSE routes deferred to production hardening
- Idempotency middleware uses SHA-256 body hash comparison to detect key reuse with different request bodies (returns 422)
- Internal ready callback returns 200 OK even if instance already transitioned (idempotent callback design)
- Rate limiter configured at 10 req/s sustained with burst of 20 per organization

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - all external dependencies (Clerk, RunPod) were configured in Plan 01 and Plan 02. This plan only wires existing components together.

## Next Phase Readiness
- All instance lifecycle API endpoints operational: create, list, get, delete, SSE events, internal ready callback
- Auth middleware chain protecting all /api/v1/* routes
- Provisioning engine integrated end-to-end from API to provider
- Phase 04 complete: ready for Phase 05 (Billing & Usage Metering)
- SSE broker ready for additional event types in future phases

## Self-Check: PASSED

All 6 artifact files verified present on disk. Both task commits (08746ae, a01e69c) verified in git log.

---
*Phase: 04-auth-instance-lifecycle*
*Completed: 2026-02-24*
