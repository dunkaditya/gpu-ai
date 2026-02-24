---
phase: 01-foundation
plan: 04
subsystem: api, middleware
tags: [localhost-restriction, ip-filtering, defense-in-depth, net-splithost, middleware-chain]

# Dependency graph
requires:
  - phase: 01-foundation/03
    provides: "InternalAuthMiddleware, API Server with health endpoint, ServerDeps pattern"
provides:
  - "LocalhostOnly middleware rejecting non-loopback IPs with 404"
  - "Defense-in-depth: LocalhostOnly layered before InternalAuthMiddleware on /health"
  - "7 middleware tests (5 LocalhostOnly + 2 InternalAuthMiddleware regression)"
affects: [02-provider-adapter, 03-privacy-layer, 04-instance-lifecycle]

# Tech tracking
tech-stack:
  added: []
  patterns: [localhost-only-ip-restriction, defense-in-depth-middleware-chaining, 404-for-hidden-endpoints]

key-files:
  created:
    - internal/api/middleware_test.go
  modified:
    - internal/api/middleware.go
    - internal/api/server.go

key-decisions:
  - "404 Not Found for rejected IPs instead of 403 -- avoids revealing endpoint existence to external scanners"
  - "net.SplitHostPort for IP extraction -- handles both IPv4 and IPv6 RemoteAddr formats correctly"
  - "LocalhostOnly as outermost middleware -- rejects external IPs before token check runs"

patterns-established:
  - "LocalhostOnly middleware: defense-in-depth IP restriction for internal endpoints"
  - "Middleware chaining order: network restriction (LocalhostOnly) -> auth (InternalAuthMiddleware) -> handler"

requirements-completed: [AUTH-04]

# Metrics
duration: 1min
completed: 2026-02-24
---

# Phase 1 Plan 4: Localhost IP Restriction Summary

**LocalhostOnly middleware restricting /health to 127.0.0.1 and ::1 with 404 rejection, layered before Bearer token auth as defense-in-depth**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-24T16:21:42Z
- **Completed:** 2026-02-24T16:22:56Z
- **Tasks:** 1
- **Files modified:** 3

## Accomplishments
- LocalhostOnly middleware using net.SplitHostPort to extract and validate remote IP
- Non-loopback requests rejected with 404 Not Found (hides endpoint from scanners)
- Health route wrapped as LocalhostOnly(InternalAuthMiddleware(...)) for defense-in-depth
- 7 passing tests: IPv4/IPv6 loopback allow, external IPv4/IPv6 reject, malformed reject, plus 2 InternalAuthMiddleware regression tests

## Task Commits

Each task was committed atomically:

1. **Task 1: Add LocalhostOnly middleware and wire into health route with tests** - `98ce268` (feat)

**Plan metadata:** `faa62a9` (docs: complete plan)

## Files Created/Modified
- `internal/api/middleware.go` - Added LocalhostOnly function using net.SplitHostPort for IP validation
- `internal/api/server.go` - Wrapped health route with LocalhostOnly as outermost middleware
- `internal/api/middleware_test.go` - 7 tests covering LocalhostOnly + InternalAuthMiddleware

## Decisions Made
- Used 404 Not Found (not 403 Forbidden) for rejected IPs to avoid leaking endpoint existence to external scanners
- Used net.SplitHostPort for robust IP extraction that handles both IPv4 and IPv6 RemoteAddr formats
- Placed LocalhostOnly as outermost middleware so external IPs are rejected before token verification runs

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- AUTH-04 gap fully closed: internal endpoints now reject non-loopback IPs regardless of token
- All Phase 1 foundation work complete including this gap closure
- Middleware chain pattern extensible for future internal endpoints
- Ready to proceed to Phase 2 (Provider Adapters)

## Self-Check: PASSED

- All 3 source files exist on disk (middleware.go, server.go, middleware_test.go)
- Task commit (98ce268) verified in git log
- `go build ./cmd/gpuctl` succeeds
- `go vet ./...` passes with no errors

---
*Phase: 01-foundation*
*Completed: 2026-02-24*
