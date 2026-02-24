---
phase: 04-auth-instance-lifecycle
plan: 01
subsystem: auth
tags: [clerk, jwt, rfc7807, rate-limiting, pagination, middleware]

# Dependency graph
requires:
  - phase: 03-privacy-layer
    provides: "Database schema v2 with WireGuard encryption columns"
  - phase: 01-foundation
    provides: "Config struct, API server, internal auth middleware"
provides:
  - "Clerk JWT authentication middleware (ClerkAuthMiddleware)"
  - "RequireOrg middleware for organization enforcement"
  - "RFC 7807 ProblemDetail error response helpers"
  - "Per-organization rate limiting (OrgRateLimiter)"
  - "Cursor-based pagination utilities (EncodeCursor/DecodeCursor)"
  - "Database schema v3: idempotency_keys table, instance lifecycle columns, clerk_org_id"
affects: [04-02, 04-03, 05-billing, 06-dashboard]

# Tech tracking
tech-stack:
  added: [clerk-sdk-go/v2, golang.org/x/time/rate]
  patterns: [clerk-session-claims, rfc7807-errors, org-rate-limiting, cursor-pagination]

key-files:
  created:
    - database/migrations/20250224_v3_auth_instance_lifecycle.sql
    - internal/api/errors.go
    - internal/api/middleware_auth.go
    - internal/api/middleware_rate.go
    - internal/api/pagination.go
  modified:
    - internal/config/config.go
    - internal/auth/clerk.go

key-decisions:
  - "Clerk SDK v2 RequireHeaderAuthorization used directly instead of custom JWT parsing"
  - "Empty CLERK_SECRET_KEY returns 401 (not silent pass-through) for dev safety"
  - "ClaimsFromContext wraps Clerk SessionClaimsFromContext mapping Subject->UserID and ActiveOrganizationID->OrgID"
  - "Rate limiter uses sync.Map with limiterEntry wrapper for thread-safe lastSeen tracking"
  - "Cursor encodes RFC3339 timestamp + ID as base64url (no padding)"

patterns-established:
  - "RFC 7807 errors: all API errors use writeProblem with application/problem+json"
  - "Auth middleware chain: ClerkAuthMiddleware -> RequireOrg -> rate limiter -> handler"
  - "Keyset pagination: EncodeCursor/DecodeCursor with (created_at DESC, id DESC)"

requirements-completed: [AUTH-01, AUTH-02, API-09, API-12]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 4 Plan 01: Auth & Cross-cutting Infrastructure Summary

**Clerk JWT auth middleware, RFC 7807 errors, per-org rate limiting, and cursor-based pagination with v3 schema migration**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T23:23:38Z
- **Completed:** 2026-02-24T23:26:20Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Database schema v3 migration adds idempotency_keys table, instance lifecycle columns (name, ready_at, error_reason), clerk_org_id mapping, and keyset pagination index
- Clerk JWT authentication middleware using official SDK v2 with RequireHeaderAuthorization and RequireOrg enforcement
- RFC 7807 Problem Details error format with writeProblem helper used consistently across all middleware
- Per-organization rate limiting with token bucket algorithm, stale entry cleanup, and Retry-After header on 429
- Cursor-based pagination utilities with base64url-encoded (created_at, id) tuples for keyset queries

## Task Commits

Each task was committed atomically:

1. **Task 1: Database migration and config update** - `cc72fbf` (feat)
2. **Task 2: Auth middleware, error helpers, rate limiter, and pagination utilities** - `24c2e0d` (feat)

## Files Created/Modified
- `database/migrations/20250224_v3_auth_instance_lifecycle.sql` - Schema v3: idempotency_keys, instance lifecycle columns, clerk_org_id, pagination index
- `internal/config/config.go` - Added optional ClerkSecretKey field loaded from CLERK_SECRET_KEY env var
- `internal/auth/clerk.go` - Claims struct and ClaimsFromContext wrapping Clerk SDK session claims
- `internal/api/errors.go` - ProblemDetail struct, writeProblem (RFC 7807), writeJSON helpers
- `internal/api/middleware_auth.go` - ClerkAuthMiddleware (JWT verification) and RequireOrg (org enforcement)
- `internal/api/middleware_rate.go` - OrgRateLimiter with per-org token bucket, cleanup goroutine
- `internal/api/pagination.go` - PageParams, PageResult, ParsePageParams, EncodeCursor, DecodeCursor
- `go.mod` / `go.sum` - Added clerk-sdk-go/v2, golang.org/x/time dependencies

## Decisions Made
- Used Clerk SDK v2 RequireHeaderAuthorization directly instead of custom JWT parsing -- leverages well-tested JWKS fetching, caching, and verification
- Empty CLERK_SECRET_KEY returns explicit 401 error (not silent pass-through) -- prevents accidental unauthenticated access in dev
- ClaimsFromContext wraps Clerk's SessionClaimsFromContext, mapping Subject to UserID and ActiveOrganizationID to OrgID -- single context key, no dual-key confusion
- Rate limiter uses sync.Map with limiterEntry wrapper for thread-safe lastSeen tracking -- avoids global mutex on hot path
- Cursor format uses RFC3339 timestamp + pipe + ID, base64url encoded (no padding) -- compact, URL-safe, deterministic

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed missing go.sum entries after dependency upgrade**
- **Found during:** Task 2 (build verification)
- **Issue:** Adding clerk-sdk-go/v2 upgraded golang.org/x/text, causing pgx module to have missing go.sum entries
- **Fix:** Ran `go mod tidy` to regenerate go.sum
- **Files modified:** go.sum
- **Verification:** `go build ./...` and `go vet ./...` pass
- **Committed in:** 24c2e0d (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Standard dependency resolution. No scope creep.

## Issues Encountered
None beyond the go.sum dependency resolution above.

## User Setup Required

Clerk authentication requires external configuration. For Phase 4:
- Set `CLERK_SECRET_KEY` environment variable (from Clerk Dashboard -> Configure -> API Keys -> Secret keys)
- Create a Clerk application and enable Organizations feature
- Without CLERK_SECRET_KEY, all /api/v1/* endpoints return 401

## Next Phase Readiness
- Auth middleware chain ready for handler wiring in Plan 03
- Error format standardized for all API endpoints
- Rate limiter ready to attach to route groups
- Pagination utilities ready for list endpoints (instances, SSH keys)
- Schema supports idempotency keys for safe POST retries

## Self-Check: PASSED

All 7 artifact files verified present on disk. Both task commits (cc72fbf, 24c2e0d) verified in git log.

---
*Phase: 04-auth-instance-lifecycle*
*Completed: 2026-02-24*
