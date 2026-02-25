---
phase: 06-availability-health-monitoring
plan: 02
subsystem: api, provision
tags: [availability-api, price-sorting, provider-selection, fallback-retry, server-side-filtering]

# Dependency graph
requires:
  - phase: 06-availability-health-monitoring
    plan: 01
    provides: "Redis cache layer for GPU offerings, AvailableOffering type"
  - phase: 04-auth-instance-lifecycle
    provides: "Auth middleware chain, Server struct, provisioning engine"
provides:
  - "GET /api/v1/gpu/available endpoint with 6 query param filters behind auth"
  - "Best-price provider selection via sort.SliceStable (replaces first-match)"
  - "Fallback retry: Provision tries up to 3 providers on failure"
  - "selectProviderCandidates returns all sorted matches for retry loop"
  - "AvailCache field on Server and ServerDeps for availability cache injection"
affects: [06-03-health-monitor, 06-04-sse-events, cmd/gpuctl/main.go]

# Tech tracking
tech-stack:
  added: []
  patterns: [sort.SliceStable for stable price sorting with implicit tiebreak, retry loop with candidate list]

key-files:
  created:
    - internal/api/handlers_gpu.go
  modified:
    - internal/api/server.go
    - internal/provision/engine.go

key-decisions:
  - "sort.SliceStable for price sorting preserves registry iteration order as tiebreaker -- no explicit priority list needed"
  - "Provider retry is limited to provider.Provision call only -- WG setup happens once before retry loop"
  - "Max 3 provision attempts across different providers to limit latency"
  - "Price cap checked against cheapest candidate only (first in sorted list)"

patterns-established:
  - "selectProviderCandidates returns full sorted list; selectProvider wraps it for backward compatibility"
  - "Fallback retry pattern: loop over sorted candidates with max attempts, log each failure"

requirements-completed: [AVAIL-03, AVAIL-04, AVAIL-05, API-05]

# Metrics
duration: 3min
completed: 2026-02-25
---

# Phase 6 Plan 02: Availability API & Best-Price Selection Summary

**GPU availability endpoint with 6-filter server-side search, plus price-sorted provider selection with automatic fallback retry (max 3 providers)**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-25T20:19:46Z
- **Completed:** 2026-02-25T20:22:45Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- GET /api/v1/gpu/available endpoint reads Redis cache and returns filtered GPU offerings behind auth chain
- All 6 server-side filters implemented: gpu_model (case-insensitive), region, tier, min_price, max_price, min_vram
- selectProvider upgraded from first-match to sort.SliceStable by price ascending with registry-order tiebreak
- Provision retries with next-cheapest provider on failure, capped at 3 attempts
- WireGuard setup happens once before retry loop; only provider API call is retried
- Empty array returned (not error) when no offerings match filters or cache is empty

## Task Commits

Each task was committed atomically:

1. **Task 1: GPU availability API handler with filtering** - `a5b5321` (feat)
2. **Task 2: Best-price provider selection with fallback retry** - `7c3e99c` (feat)

## Files Created/Modified
- `internal/api/handlers_gpu.go` - handleListGPUAvailability handler with query param filtering and matchesFilters helper
- `internal/api/server.go` - Added availCache field to Server, AvailCache to ServerDeps, registered GET /api/v1/gpu/available route
- `internal/provision/engine.go` - providerCandidate type, selectProviderCandidates with sort.SliceStable, Provision fallback retry loop

## Decisions Made
- sort.SliceStable preserves registry iteration order as equal-price tiebreaker (no explicit priority list mechanism needed, per CONTEXT.md)
- WG setup (key gen, IPAM, AddPeer) runs once before retry loop; only provider.Provision call retried with different providers
- Max 3 provision attempts limits latency while providing resilience against single-provider failures
- Price cap checked against cheapest candidate (first in sorted list) before attempting any provisioning

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Availability API ready for dashboard consumption (Phase 7)
- Best-price selection and fallback retry active for all new provisions
- main.go needs AvailCache wired into ServerDeps (will happen in 06-04 main.go wiring plan)
- Health monitor (Plan 03) and SSE events (Plan 04) are next

## Self-Check: PASSED

All 4 files verified present. Both task commits (a5b5321, 7c3e99c) confirmed in git log.

---
*Phase: 06-availability-health-monitoring*
*Completed: 2026-02-25*
