---
phase: 06-availability-health-monitoring
plan: 04
subsystem: api, infra
tags: [sse, events, goroutine, redis, health-monitor, availability-poller]

# Dependency graph
requires:
  - phase: 06-02
    provides: Availability cache, poller, and GPU availability API endpoint
  - phase: 06-03
    provides: Health monitor with spot interruption detection and event logging
provides:
  - Per-org SSE broker (OrgEventBroker) for real-time instance event streaming
  - REST catch-up endpoint for missed events (GET /api/v1/events?since=)
  - Availability poller goroutine wired in main.go (30s interval)
  - Health monitor goroutine wired in main.go (60s interval, SSE push via OnEvent)
  - PricingMarkupPct config field with 15% default
affects: [07-dashboard, frontend-integration]

# Tech tracking
tech-stack:
  added: []
  patterns: [org-scoped-sse-broker, combined-sse-rest-endpoint, goroutine-lifecycle-wiring]

key-files:
  created:
    - internal/api/handlers_events.go
  modified:
    - internal/api/sse.go
    - internal/api/server.go
    - internal/config/config.go
    - cmd/gpuctl/main.go

key-decisions:
  - "Combined SSE and REST catch-up in single GET /api/v1/events endpoint (since= param switches mode)"
  - "OrgEventBroker buffer size 20 (double StatusBroker's 10) for higher org-level event volume"
  - "Health monitor created after API server to use srv.PublishOrgEvent callback directly"

patterns-established:
  - "Combined endpoint pattern: single route handler dispatches to SSE or REST based on query params"
  - "Goroutine startup order: poller before server, monitor after server (dependency-driven)"

requirements-completed: [HLTH-04, API-05]

# Metrics
duration: 2min
completed: 2026-02-25
---

# Phase 6 Plan 4: SSE Events + Main.go Wiring Summary

**Per-org SSE event broker with REST catch-up endpoint, availability poller and health monitor goroutines wired in main.go**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-25T21:14:00Z
- **Completed:** 2026-02-25T21:17:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- OrgEventBroker enables per-org SSE subscription for real-time instance events (interruptions, failures)
- GET /api/v1/events endpoint serves both SSE streaming and REST catch-up (via ?since= param)
- Availability poller and health monitor started as background goroutines in main.go
- AvailCache wired through ServerDeps, PricingMarkupPct config with 15% default

## Task Commits

Each task was committed atomically:

1. **Task 1: Per-org SSE broker and event handlers** - `c175619` (feat)
2. **Task 2: Config update and main.go wiring** - `152fd0e` (feat)

## Files Created/Modified
- `internal/api/sse.go` - Added OrgEventBroker with Subscribe/Unsubscribe/Publish for per-org SSE
- `internal/api/handlers_events.go` - handleEvents combining SSE streaming and REST catch-up
- `internal/api/server.go` - orgEventBroker field, route registration, PublishOrgEvent method
- `internal/config/config.go` - PricingMarkupPct field with 15.0% default from PRICING_MARKUP_PCT env
- `cmd/gpuctl/main.go` - Availability poller, health monitor goroutines, AvailCache in ServerDeps

## Decisions Made
- Combined SSE and REST catch-up in single GET /api/v1/events endpoint -- since= query param switches mode, avoiding route proliferation
- OrgEventBroker buffer size 20 (vs StatusBroker's 10) for higher org-level event volume
- Health monitor created after API server to use srv.PublishOrgEvent callback directly (no post-construction setter needed)
- Startup order: poller before server (data available on first request), monitor after server (needs PublishOrgEvent)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required. PricingMarkupPct defaults to 15.0% and is optional.

## Next Phase Readiness
- Phase 6 complete: all 4 plans executed
- Availability polling, health monitoring, SSE events, and REST catch-up all wired and operational
- Ready for Phase 7 (Dashboard) -- SSE and REST event endpoints available for frontend consumption

## Self-Check: PASSED

All files verified present. All commit hashes verified in git log.

---
*Phase: 06-availability-health-monitoring*
*Completed: 2026-02-25*
