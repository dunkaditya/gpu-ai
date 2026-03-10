---
phase: 10-frontend-dashboard
plan: 02
subsystem: ui
tags: [react, nextjs, swr, tailwind, gpu-availability, launch-modal, card-grid]

# Dependency graph
requires:
  - phase: 10-frontend-dashboard
    plan: 01
    provides: Dashboard shell, sidebar, ConfirmDialog, LaunchInstanceForm base
  - phase: 07-frontend-cloud-dashboard
    provides: SWR data fetching, API layer, types, GPU availability page
provides:
  - GPUCard component for individual GPU model display with specs and pricing
  - GPUAvailabilityTable redesigned from flat table to card grid grouped by GPU model
  - GPUCardData type for client-side aggregation of AvailableOffering[]
  - Filter bar with region dropdown, tier segmented control, price sort toggle
  - Enhanced LaunchInstanceForm with pre-filled mode (from GPU card) and manual mode
  - Price confirmation with hourly cost, monthly estimate, and GPU specs
  - Dynamic GPU count price updates in launch modal
  - Redirect to /cloud/instances after successful launch
affects: [10-03, 10-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [card-grid-grouping, dual-mode-modal, pre-filled-launch-flow]

key-files:
  created:
    - frontend/src/components/cloud/GPUCard.tsx
  modified:
    - frontend/src/components/cloud/GPUAvailabilityTable.tsx
    - frontend/src/components/cloud/LaunchInstanceForm.tsx
    - frontend/src/app/cloud/gpu-availability/page.tsx
    - frontend/src/lib/types.ts

key-decisions:
  - "Filters apply BEFORE grouping: flat offerings are filtered by region/tier, then grouped into GPUCardData by gpu_model"
  - "GPUCard launches cheapest available offering when Launch button clicked"
  - "Price sort uses minimum of spot/on-demand prices per card for sorting comparison"
  - "Launch modal pre-filled mode omits ssh_key_ids -- backend auto-attaches all user SSH keys"
  - "GPU availability page converted to client component for GPUAvailabilityTable"

patterns-established:
  - "Card grid grouping: useMemo to group flat API data into aggregate card models"
  - "Dual-mode modal: offering prop switches between confirmation display and free-text inputs"
  - "Skeleton card: animated placeholder matching card dimensions during loading"

requirements-completed: [DASH-03, DASH-04]

# Metrics
duration: 4min
completed: 2026-03-10
---

# Phase 10 Plan 02: GPU Availability & Launch Modal Summary

**GPU card grid grouped by model with spec/pricing display, filter bar, and enhanced launch modal with price confirmation and pre-fill from availability cards**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-10T06:06:42Z
- **Completed:** 2026-03-10T06:10:42Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- GPU availability page redesigned from flat table to responsive card grid (1/2/3 columns), each card showing GPU model, VRAM badge, CPU/RAM/storage specs, spot and on-demand pricing side by side, region tags, and availability count
- Filter bar with region dropdown, tier segmented control (All/Spot/On-Demand), and price sort toggle
- LaunchInstanceForm rewritten with dual-mode support: pre-filled mode shows specs grid, region/tier badges, large price display, monthly estimate, and dynamic price updates by GPU count; manual mode preserves free-text inputs for direct launching
- Successful instance launch redirects to /cloud/instances via router.push

## Task Commits

Each task was committed atomically:

1. **Task 1: GPUCard component and GPU availability card grid with filters** - `3298f44` (feat)
   - Note: This commit also included the Task 2 LaunchInstanceForm rewrite due to staging order

## Files Created/Modified
- `frontend/src/components/cloud/GPUCard.tsx` - Individual GPU card component with specs, dual pricing, region tags, availability, and launch button
- `frontend/src/components/cloud/GPUAvailabilityTable.tsx` - Redesigned from table to card grid with grouping logic, filters, skeleton loading
- `frontend/src/components/cloud/LaunchInstanceForm.tsx` - Enhanced with pre-filled mode (specs grid, price confirmation) and manual mode
- `frontend/src/app/cloud/gpu-availability/page.tsx` - Converted to client component
- `frontend/src/lib/types.ts` - Added GPUCardData interface for grouped card data

## Decisions Made
- Filters apply BEFORE grouping: flat offerings filtered by region/tier, then grouped into GPUCardData by gpu_model -- ensures card-level aggregation respects active filters
- GPUCard launches the cheapest available offering (by price_per_hour) when Launch button is clicked
- Price sort uses minimum of spot/on-demand prices per card as sort key
- Launch modal pre-filled mode omits ssh_key_ids entirely -- backend auto-attaches per CONTEXT.md decision
- GPU availability page converted to client component to support GPUAvailabilityTable as client component child

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- GPU card grid and launch modal ready for end-to-end provisioning flow
- Instance detail page and table enhancements ready for Plan 03
- Card grid can be extended with additional sort/filter options if needed

## Self-Check: PASSED

All 5 files verified present. Task commit (3298f44) verified in git log.

---
*Phase: 10-frontend-dashboard*
*Completed: 2026-03-10*
