---
phase: 11-dashboard-ui-redesign
plan: 03
subsystem: ui
tags: [css-grid, react, tailwind, linear-aesthetic, instances-table, polish]

# Dependency graph
requires:
  - phase: 11-01
    provides: Design tokens (btn-primary, btn-secondary, border-radius), EmptyState component, scoped film grain
provides:
  - CSS grid instances table with proper column alignment and clickable rows
  - All dashboard pages using btn-primary/btn-secondary (zero gradient-btn usage)
  - Shared EmptyState component adopted across instances, billing, SSH key pages
  - Consistent 10px border radius and muted focus rings across all dashboard pages
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [CSS grid with display:contents for clickable table rows, EmptyState component pattern for all empty states]

key-files:
  created: []
  modified:
    - frontend/src/components/cloud/InstancesTable.tsx
    - frontend/src/components/cloud/InstanceDetail.tsx
    - frontend/src/components/cloud/BillingDashboard.tsx
    - frontend/src/components/cloud/SSHKeyManager.tsx
    - frontend/src/app/cloud/instances/page.tsx
    - frontend/src/app/cloud/settings/page.tsx

key-decisions:
  - "CSS grid with display:contents on Link elements for clickable table rows with proper column alignment"
  - "SSH command column uses minmax(200px, 2fr) instead of truncated max-w-[260px] for flexible width"
  - "Billing summary cards use standard border (no colored left accents) for clean Linear aesthetic"
  - "Period selector active state uses bg-bg-card-hover (muted) instead of bg-purple-dim"
  - "Progress bars use bg-text-muted instead of bg-purple for non-CTA contexts"
  - "Input focus rings use focus:ring-border-light instead of focus:ring-purple/50"

patterns-established:
  - "CSS grid + display:contents for tables that need clickable full-row links"
  - "EmptyState component as universal empty state pattern across all dashboard pages"
  - "btn-primary for all dashboard CTAs, red/danger styling only for destructive actions"
  - "focus:ring-1 focus:ring-border-light for subtle input focus indicators"

requirements-completed: [UI-04, UI-05]

# Metrics
duration: 5min
completed: 2026-03-12
---

# Phase 11 Plan 03: Instances Table Fix and Dashboard Polish Summary

**CSS grid instances table with proper column alignment, shared EmptyState adoption across all pages, gradient-btn to btn-primary migration, consistent 10px radius and muted focus rings**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-12T01:44:22Z
- **Completed:** 2026-03-12T01:50:11Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Replaced HTML table with CSS grid layout using display:contents on Link elements for clickable rows with proper column alignment
- Removed colSpan hack and fixed SSH command truncation (now uses flexible minmax column instead of 260px max-width)
- Migrated all gradient-btn usage to btn-primary/btn-secondary across all dashboard pages (zero gradient-btn remaining)
- Adopted shared EmptyState component across instances table, billing dashboard, and SSH key manager
- Applied consistent 10px border radius, muted focus rings, and neutral retry button colors throughout

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix instances table with CSS grid and polish instances pages** - `477e3f1` (feat)
2. **Task 2: Polish billing, SSH keys, and settings pages** - `a369c34` (feat)

## Files Created/Modified
- `frontend/src/components/cloud/InstancesTable.tsx` - Rewritten desktop table from HTML table to CSS grid with display:contents Link rows
- `frontend/src/components/cloud/InstanceDetail.tsx` - Updated to rounded-[10px], replaced gradient-btn with btn-primary
- `frontend/src/app/cloud/instances/page.tsx` - Updated to rounded-[10px], btn-primary, muted retry colors
- `frontend/src/components/cloud/BillingDashboard.tsx` - Removed colored left borders, muted period selector, EmptyState, rounded-[10px]
- `frontend/src/components/cloud/SSHKeyManager.tsx` - btn-primary/btn-secondary, EmptyState, subtle focus rings, rounded-[10px]
- `frontend/src/app/cloud/settings/page.tsx` - btn-primary, muted progress bar, subtle focus rings, rounded-[10px]

## Decisions Made
- CSS grid with display:contents chosen over flexbox-in-tr hack for proper column alignment while preserving Link semantics
- SSH command column uses minmax(200px, 2fr) grid column to use available space dynamically
- Billing summary cards use standard border on all sides (no colored left accents) to match Linear aesthetic
- Period selector active state uses bg-bg-card-hover for muted look instead of purple highlight
- Progress bars use bg-text-muted for under-70% state to reduce purple in non-CTA contexts
- Input focus rings use border-light color throughout for subtle, non-purple focus indicators

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All dashboard pages now use consistent Linear aesthetic with flat buttons, muted colors, and 10px radii
- Phase 11 dashboard UI redesign complete -- all 3 plans executed

## Self-Check: PASSED

All modified files verified present. All commit hashes verified in git log.

---
*Phase: 11-dashboard-ui-redesign*
*Completed: 2026-03-12*
