---
phase: 06-availability-health-monitoring
plan: 03
subsystem: health, api, provision
tags: [health-monitor, spot-interruption, event-logging, billing, concurrency]

# Dependency graph
requires:
  - phase: 06-01
    provides: "Availability poller, instance_events table, event DB methods"
  - phase: 05-ssh-keys-billing
    provides: "Billing session management, CloseBillingSession"
  - phase: 04-auth-instance-lifecycle
    provides: "Instance lifecycle, provision engine, handlers_internal"
provides:
  - "Background health monitor polling provider APIs every 60s"
  - "Spot interruption detection with immediate billing stop"
  - "Non-spot failure detection with 3 retries over 30s"
  - "Instance event logging for ready, terminated, interrupted, failed events"
  - "OnEvent callback for SSE integration"
  - "ListActiveInstances DB method for running/booting instances"
affects: [06-04, dashboard, sse-events]

# Tech tracking
tech-stack:
  added: []
  patterns: ["bounded concurrency with semaphore channel", "optimistic locking for duplicate prevention", "non-fatal event logging pattern"]

key-files:
  created:
    - internal/health/monitor.go
  modified:
    - internal/db/instances.go
    - internal/api/handlers_internal.go
    - internal/provision/engine.go

key-decisions:
  - "Optimistic locking prevents duplicate event logging on concurrent status transitions"
  - "Non-fatal event logging: failure to log event does not block primary operation"
  - "Bounded concurrency (max 10) via semaphore channel for parallel provider checks"

patterns-established:
  - "Health monitor OnEvent callback pattern: same post-construction wiring as Engine.SetOnStatusChange"
  - "Event metadata as JSON map with gpu_type, region, tier for SSE consumers"

requirements-completed: [HLTH-01, HLTH-02, HLTH-03, HLTH-04]

# Metrics
duration: 2min
completed: 2026-02-25
---

# Phase 6 Plan 3: Health Monitor Summary

**Background health monitor with spot interruption detection, non-spot retry logic, and instance event logging across ready/terminated/interrupted/failed lifecycle transitions**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-25T20:19:58Z
- **Completed:** 2026-02-25T20:22:24Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Health monitor polls all running/booting instances every 60s with bounded concurrency (max 10)
- Spot interruption: immediately closes billing, sets error state, logs "interrupted" event
- Non-spot failure: retries 3 times over 30s, then closes billing, sets error, logs "failed" event
- Instance ready callback logs "ready" event to instance_events
- Instance termination logs "terminated" event to instance_events
- Optimistic locking prevents duplicate events on concurrent status transitions
- OnEvent callback wired for SSE integration

## Task Commits

Each task was committed atomically:

1. **Task 1: ListActiveInstances DB method and health monitor implementation** - `510ef7e` (feat)
2. **Task 2: Add event logging to instance ready callback and termination** - `0ed59af` (feat)

## Files Created/Modified
- `internal/health/monitor.go` - Full health monitor: Start, checkAll, checkInstance, handleSpotInterruption, handleNonSpotFailure
- `internal/db/instances.go` - Added ListActiveInstances for running/booting instances across all orgs
- `internal/api/handlers_internal.go` - Added "ready" event logging in handleInstanceReady
- `internal/provision/engine.go` - Added "terminated" event logging in Terminate

## Decisions Made
- Optimistic locking prevents duplicate event logging: only log if status transition succeeds via UpdateInstanceStatus
- Non-fatal event logging: event persistence failure does not block the primary operation (instance running, terminated, etc.)
- Bounded concurrency uses semaphore channel pattern (max 10 goroutines) to avoid overwhelming provider APIs
- OnEvent callback follows same post-construction wiring pattern as Engine.SetOnStatusChange

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Health monitor ready to be started in main.go (wired as goroutine with context)
- Event logging in place for all major lifecycle transitions (ready, terminated, interrupted, failed)
- OnEvent callback ready for SSE broker integration in Phase 06-04

## Self-Check: PASSED

- All 4 files verified present on disk
- Both task commits verified in git log (510ef7e, 0ed59af)
- `go build ./...` passes

---
*Phase: 06-availability-health-monitoring*
*Completed: 2026-02-25*
