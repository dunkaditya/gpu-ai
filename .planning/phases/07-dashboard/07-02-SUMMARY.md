---
phase: 07-dashboard
plan: 02
subsystem: ui
tags: [nextjs, react, dashboard, sidebar, table, mock-data, route-groups]

# Dependency graph
requires:
  - phase: 07-dashboard-01
    provides: proxy.ts hostname routing, (marketing) and (cloud) route groups
provides:
  - Dashboard shell layout with sidebar navigation and topbar
  - Mock instance data layer (MockInstance type + 5 test instances)
  - InstancesTable with desktop table and mobile card views
  - StatusBadge component with color-coded status indicators
  - Cloud route group pages (instances, settings, root redirect)
affects: [07-03, 07-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [dashboard shell layout, sidebar active state via usePathname, site param preservation for local dev routing]

key-files:
  created:
    - frontend/src/lib/mock-data.ts
    - frontend/src/components/cloud/StatusBadge.tsx
    - frontend/src/components/cloud/DashboardSidebar.tsx
    - frontend/src/components/cloud/DashboardTopbar.tsx
    - frontend/src/components/cloud/InstancesTable.tsx
    - frontend/src/app/(cloud)/layout.tsx
    - frontend/src/app/(cloud)/page.tsx
    - frontend/src/app/(cloud)/instances/page.tsx
    - frontend/src/app/(cloud)/settings/page.tsx
  modified: []

key-decisions:
  - "StatusBadge uses design system CSS variables (green-dim, purple-dim) with animate-pulse-dot for starting state"
  - "DashboardSidebar preserves ?site=cloud query param via useSearchParams for local dev routing compatibility"
  - "InstancesTable provides both desktop table and mobile card layout with responsive breakpoints"

patterns-established:
  - "Cloud component pattern: components/cloud/ directory for dashboard-specific components"
  - "Mock data layer: lib/mock-data.ts with typed interfaces matching backend API response shapes"
  - "Dashboard shell: sidebar + topbar + scrollable main area in (cloud)/layout.tsx"

requirements-completed: [DASH-05, DASH-08]

# Metrics
duration: 2min
completed: 2026-03-02
---

# Phase 7 Plan 02: Dashboard Shell & Instances Table Summary

**Cloud dashboard shell with sidebar navigation, instances table showing 5 mock GPUs with status badges and SSH copy, settings placeholder**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-02T21:34:31Z
- **Completed:** 2026-03-02T21:37:00Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- Built complete dashboard shell with sidebar navigation (5 nav items with SVG icons, pathname-based active state) and topbar with breadcrumb
- Created InstancesTable with desktop table and mobile card views, SSH command copy-to-clipboard, and empty state
- Created mock data layer with 5 instances covering all statuses (running, starting, error, terminated) matching backend API shape
- StatusBadge pill component with color-coded indicators: green/running, purple-pulse/starting, amber/stopping, gray/terminated, red/error

## Task Commits

Each task was committed atomically:

1. **Task 1: Create mock data and cloud UI components** - `3ddcc9a` (feat)
2. **Task 2: Create cloud route group pages and verify build** - `472b238` (feat)

## Files Created/Modified
- `frontend/src/lib/mock-data.ts` - MockInstance type and 5 mock instances with varied statuses
- `frontend/src/components/cloud/StatusBadge.tsx` - Colored status pill with dot indicator and pulse animation
- `frontend/src/components/cloud/DashboardSidebar.tsx` - Sidebar nav with 5 items, pathname active state, site param preservation
- `frontend/src/components/cloud/DashboardTopbar.tsx` - Top bar with breadcrumb navigation and user avatar placeholder
- `frontend/src/components/cloud/InstancesTable.tsx` - Desktop table + mobile cards with SSH copy button and empty state
- `frontend/src/app/(cloud)/layout.tsx` - Dashboard shell with sidebar + topbar + scrollable main area
- `frontend/src/app/(cloud)/page.tsx` - Root redirect to /instances
- `frontend/src/app/(cloud)/instances/page.tsx` - Instances list page with mock data and Launch Instance button
- `frontend/src/app/(cloud)/settings/page.tsx` - Settings placeholder page

## Decisions Made
- StatusBadge uses globals.css design tokens (green-dim, purple-dim, etc.) with the existing animate-pulse-dot keyframe for the starting state
- DashboardSidebar preserves the ?site=cloud query param via useSearchParams hook so local dev routing works when clicking nav links
- InstancesTable provides both a dense desktop table view (Bloomberg-terminal style) and compact mobile card view with responsive md breakpoint
- DashboardTopbar derives page label from pathname lookup map rather than passing props
- Logo link in sidebar navigates back to marketing site (with ?site=marketing in local dev)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Dashboard shell is ready for Clerk auth integration (Plan 03)
- Mock data layer ready to be swapped with real API calls (Plan 04)
- All cloud components use design system tokens and cn() utility for consistent styling
- Route group pages can be extended with additional routes (GPU Availability, SSH Keys, Billing)

## Self-Check: PASSED

All 9 created files verified present. Both task commits (3ddcc9a, 472b238) found in git log. TypeScript compilation passes with zero errors.

---
*Phase: 07-dashboard*
*Completed: 2026-03-02*
