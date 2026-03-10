---
phase: 10-frontend-dashboard
plan: 03
subsystem: ui
tags: [react, next.js, swr, instances, detail-page, inline-edit, confirm-dialog]

requires:
  - phase: 10-01
    provides: "Dashboard shell with sidebar, layout, ConfirmDialog, StatusBadge"
provides:
  - "Enhanced InstancesTable with clickable rows, inline rename, uptime, ConfirmDialog"
  - "InstanceDetail component with GPU config, connection, billing, metadata cards"
  - "/cloud/instances/[id] route with SWR data fetching"
affects: [10-04]

tech-stack:
  added: []
  patterns: ["EditableName inline editing with pencil icon", "SWR cache invalidation across list+detail keys", "ConfirmDialog for destructive actions"]

key-files:
  created:
    - frontend/src/components/cloud/InstanceDetail.tsx
    - frontend/src/app/cloud/instances/[id]/page.tsx
    - frontend/src/app/cloud/instances/[id]/layout.tsx
  modified:
    - frontend/src/components/cloud/InstancesTable.tsx

key-decisions:
  - "EditableName inlined in both table and detail (not extracted to shared) for simplicity"
  - "formatUptime and CopyButton duplicated across table and detail for self-contained components"
  - "ConfirmDialog replaces window.confirm for all terminate actions"
  - "SWR globalMutate used on detail page to invalidate list cache key for cross-page consistency"

patterns-established:
  - "EditableName: inline editing with pencil icon on hover, Enter to save, Escape to cancel"
  - "ConfirmDialog pattern for all destructive actions (no window.confirm)"
  - "Detail page SWR pattern: globalMutate to invalidate sibling cache keys"

requirements-completed: [DASH-05, DASH-08]

duration: 3min
completed: 2026-03-10
---

# Phase 10 Plan 03: Instances Table & Detail Page Summary

**Enhanced instances table with clickable rows, inline rename, uptime display, terminate confirmation dialog, and full instance detail page at /cloud/instances/[id]**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-10T06:06:42Z
- **Completed:** 2026-03-10T06:10:14Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Table rows are clickable and navigate to /cloud/instances/{id} detail page
- Inline editable instance name with pencil icon hover, auto-focus, Enter/Escape handling
- Uptime column calculates duration from ready_at/created_at (Xh Ym / Xd Yh format)
- Terminate action uses ConfirmDialog instead of window.confirm()
- Empty state shows "No instances running" with CTA link to GPU Availability
- Instance detail page shows GPU config, connection/SSH, cost/billing, and metadata cards
- SSH command displayed in prominent monospace block with one-click copy
- Terminated instances show muted state with "Launch New Instance" CTA
- Detail page rename/terminate invalidates both detail and list SWR cache keys

## Task Commits

Each task was committed atomically:

1. **Task 1: Enhanced InstancesTable with clickable rows, inline rename, uptime, ConfirmDialog** - `a238885` (feat)
2. **Task 2: Instance detail page at /cloud/instances/[id]** - `b0cd6b2` (feat)

## Files Created/Modified
- `frontend/src/components/cloud/InstancesTable.tsx` - Enhanced table with clickable rows, EditableName, uptime column, ConfirmDialog for terminate, empty state CTA
- `frontend/src/components/cloud/InstanceDetail.tsx` - Full instance detail view with GPU, connection, billing, metadata cards
- `frontend/src/app/cloud/instances/[id]/page.tsx` - Client page with SWR fetch, skeleton loading, error state
- `frontend/src/app/cloud/instances/[id]/layout.tsx` - Server metadata layout for instance detail

## Decisions Made
- EditableName component inlined in both InstancesTable and InstanceDetail rather than extracting to a shared component -- keeps each component self-contained and avoids premature abstraction
- formatUptime utility duplicated between table and detail for the same reason
- Used useSWRConfig().mutate (globalMutate) on detail page to invalidate the /api/v1/instances list cache key after rename/terminate, ensuring list page shows fresh data on navigation back
- ConfirmDialog pattern adopted project-wide for all destructive actions (replaces all window.confirm calls)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Instance management UI complete with list and detail views
- Ready for plan 10-04 (final integration and polish)

## Self-Check: PASSED

All 4 files verified present on disk. Both commit hashes (a238885, b0cd6b2) verified in git log.

---
*Phase: 10-frontend-dashboard*
*Completed: 2026-03-10*
