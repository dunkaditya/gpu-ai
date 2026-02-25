---
phase: 06-availability-health-monitoring
plan: 01
subsystem: availability, database
tags: [redis, poller, gpu-offerings, instance-events, markup-pricing]

# Dependency graph
requires:
  - phase: 05-ssh-keys-billing
    provides: "Billing sessions and instance lifecycle for events"
  - phase: 02-provider-abstraction-runpod-adapter
    provides: "Provider interface, Registry, GPUOffering struct"
provides:
  - "instance_events table with event_type CHECK constraint and indexes"
  - "AvailableOffering customer-facing type with defense-by-omission"
  - "Redis cache layer for GPU offerings (single JSON array key)"
  - "Background availability poller with concurrent provider querying"
  - "CreateInstanceEvent, ListInstanceEventsByOrg, ListInstanceEventsByInstance DB methods"
  - "GPUOffering extended with CPUCores, RAMGB, StorageGB fields"
affects: [06-02-availability-api, 06-03-health-monitor, 06-04-sse-events]

# Tech tracking
tech-stack:
  added: []
  patterns: [defense-by-omission for customer types, single-key JSON cache, concurrent provider polling]

key-files:
  created:
    - database/migrations/20250225_v6_availability_health.sql
    - internal/availability/types.go
    - internal/availability/cache.go
    - internal/availability/poller.go
    - internal/db/events.go
  modified:
    - internal/provider/types.go

key-decisions:
  - "Redis single-key JSON array (gpu:offerings:all) instead of per-offering SCAN pattern -- atomic reads, no partial data"
  - "35s TTL on cache entries with 30s poll interval -- brief overlap prevents stale reads"
  - "Markup pricing applied during cache write so customers never see wholesale prices"
  - "AvailableOffering uses defense-by-omission: Provider field structurally absent"
  - "org_id index includes created_at DESC for efficient REST catch-up endpoint queries"

patterns-established:
  - "Defense-by-omission: customer-facing types structurally exclude internal fields (matches CustomerInstance pattern)"
  - "Concurrent provider polling with per-provider error isolation and mutex-protected result aggregation"
  - "Immediate-on-startup polling before entering ticker loop"

requirements-completed: [AVAIL-01, AVAIL-02]

# Metrics
duration: 2min
completed: 2026-02-25
---

# Phase 6 Plan 01: Availability Foundation Summary

**Instance events table, Redis-cached GPU offerings with markup pricing, and concurrent background poller querying all providers every 30s**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-25T20:14:38Z
- **Completed:** 2026-02-25T20:16:32Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Schema migration v6 creates instance_events table with event_type CHECK constraint and composite indexes
- GPUOffering extended with CPUCores, RAMGB, StorageGB for richer availability API data
- AvailableOffering type with defense-by-omission (no Provider field can ever leak)
- Redis cache stores all offerings as single JSON array under gpu:offerings:all with 35s TTL
- Background poller queries all providers concurrently, applies 15% default markup, writes to cache
- Poller executes immediate poll on startup, then every 30s via ticker
- Per-provider errors logged and isolated -- one failing provider does not block others
- DB methods for creating and listing instance events (by org with since/limit, by instance)

## Task Commits

Each task was committed atomically:

1. **Task 1: Schema migration v6 + provider type updates + event DB methods** - `95f3de1` (feat)
2. **Task 2: Availability types, Redis cache, and background poller** - `7facbe6` (feat)

**Plan metadata:** `07b406f` (docs: complete plan)

## Files Created/Modified
- `database/migrations/20250225_v6_availability_health.sql` - Instance events table with CHECK constraint and indexes
- `internal/provider/types.go` - Added CPUCores, RAMGB, StorageGB to GPUOffering
- `internal/db/events.go` - InstanceEvent struct, Create/List methods for events
- `internal/availability/types.go` - AvailableOffering (customer-safe), ToAvailableOffering converter with markup
- `internal/availability/cache.go` - Redis Cache with SetOfferings/GetOfferings using single JSON key
- `internal/availability/poller.go` - Concurrent provider polling with immediate startup poll

## Decisions Made
- Redis single-key JSON array (gpu:offerings:all) instead of per-offering SCAN pattern for atomic reads
- 35s TTL on cache with 30s poll interval provides brief overlap preventing stale reads
- Markup pricing applied at cache write time so retail prices are pre-computed for API responses
- org_id index includes created_at DESC for efficient catch-up endpoint queries with ?since= parameter
- Default 40GB storage when provider does not specify (defaultStorageGB constant)
- Static uptime percentages by tier: 99.5% on-demand, 95.0% spot

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Redis cache and poller ready for availability API endpoints (Plan 02)
- Instance events table and DB methods ready for health monitor (Plan 03) and SSE streaming (Plan 04)
- ToAvailableOffering converter available for any future offering transformations

## Self-Check: PASSED

All 7 files verified present. Both task commits (95f3de1, 7facbe6) confirmed in git log.

---
*Phase: 06-availability-health-monitoring*
*Completed: 2026-02-25*
