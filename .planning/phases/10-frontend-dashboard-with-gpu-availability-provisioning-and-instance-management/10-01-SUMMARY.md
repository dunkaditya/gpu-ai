---
phase: 10-frontend-dashboard
plan: 01
subsystem: ui, api
tags: [react, nextjs, go, sidebar, breadcrumb, confirm-dialog, patch-endpoint]

# Dependency graph
requires:
  - phase: 07-frontend-cloud-dashboard
    provides: Dashboard layout, sidebar, topbar, instances page
  - phase: 04-auth-instance-lifecycle
    provides: Instance CRUD handlers pattern
provides:
  - PATCH /api/v1/instances/{id} rename endpoint
  - UpdateInstanceName DB function
  - renameInstance frontend API function
  - ConfirmDialog reusable component with accessibility
  - Redesigned sidebar with 7 nav items and coming-soon badges
  - Mobile sidebar overlay with hamburger toggle
  - Dynamic breadcrumb parsing for all routes
  - /cloud redirect to /cloud/instances
  - /cloud/api-keys and /cloud/team coming-soon pages
affects: [10-02, 10-03, 10-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [mobile-sidebar-overlay, dynamic-breadcrumb, coming-soon-page-pattern]

key-files:
  created:
    - frontend/src/components/cloud/ConfirmDialog.tsx
    - frontend/src/app/cloud/page.tsx
    - frontend/src/app/cloud/api-keys/page.tsx
    - frontend/src/app/cloud/team/page.tsx
  modified:
    - internal/api/handlers.go
    - internal/api/server.go
    - internal/db/instances.go
    - frontend/src/lib/api.ts
    - frontend/src/components/cloud/DashboardSidebar.tsx
    - frontend/src/components/cloud/DashboardTopbar.tsx
    - frontend/src/app/cloud/layout.tsx

key-decisions:
  - "Cloud layout converted to client component for mobile sidebar state sharing between sidebar and topbar"
  - "Sidebar icons extracted to named components for readability and reuse"
  - "Coming-soon pages use server components with static content (no client-side JS needed)"

patterns-established:
  - "Coming-soon page pattern: icon + heading + description + pulsing badge"
  - "Mobile sidebar overlay: fixed z-50, backdrop-blur, slide-in, auto-close on nav"
  - "Dynamic breadcrumb: segment map + UUID detection for instance detail routes"

requirements-completed: [DASH-01, DASH-02, DASH-05]

# Metrics
duration: 3min
completed: 2026-03-10
---

# Phase 10 Plan 01: Dashboard Shell Summary

**PATCH rename endpoint, redesigned sidebar with 7 nav items and mobile hamburger, dynamic breadcrumb, ConfirmDialog with focus trap, and coming-soon pages**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-10T06:00:27Z
- **Completed:** 2026-03-10T06:03:38Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- PATCH /api/v1/instances/{id} endpoint for instance rename with org-scoped DB function
- Reusable ConfirmDialog component with role="dialog", aria-modal, focus trap, Escape close, and loading state
- Sidebar redesigned with primary/management sections, divider, 2 coming-soon badges, and mobile overlay
- Dynamic breadcrumb parsing handles all routes including /cloud/instances/{uuid}
- /cloud redirects to /cloud/instances; API Keys and Team coming-soon pages exist

## Task Commits

Each task was committed atomically:

1. **Task 1: Backend rename endpoint + frontend API + ConfirmDialog** - `0245e91` (feat)
2. **Task 2: Sidebar redesign, topbar breadcrumb, layout, coming-soon pages** - `80a62fc` (feat)

## Files Created/Modified
- `internal/api/handlers.go` - Added handleUpdateInstance PATCH handler
- `internal/api/server.go` - Registered PATCH /api/v1/instances/{id} route
- `internal/db/instances.go` - Added UpdateInstanceName function
- `frontend/src/lib/api.ts` - Added renameInstance function
- `frontend/src/components/cloud/ConfirmDialog.tsx` - Reusable confirmation dialog with accessibility
- `frontend/src/components/cloud/DashboardSidebar.tsx` - Redesigned with sections, badges, mobile overlay
- `frontend/src/components/cloud/DashboardTopbar.tsx` - Dynamic breadcrumb and hamburger button
- `frontend/src/app/cloud/layout.tsx` - Client wrapper for mobile sidebar state
- `frontend/src/app/cloud/page.tsx` - Redirect to /cloud/instances
- `frontend/src/app/cloud/api-keys/page.tsx` - Coming-soon page
- `frontend/src/app/cloud/team/page.tsx` - Coming-soon page

## Decisions Made
- Cloud layout converted to client component for mobile sidebar state sharing between sidebar and topbar
- Sidebar icons extracted to named components for readability and reuse
- Coming-soon pages use server components with static content (no client-side JS needed)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Rename endpoint ready for instances table inline editing in Plan 02
- ConfirmDialog ready for terminate confirmation in Plan 02
- Sidebar shell complete for all dashboard pages
- Mobile navigation functional

## Self-Check: PASSED

All 11 files verified present. Both task commits (0245e91, 80a62fc) verified in git log.

---
*Phase: 10-frontend-dashboard*
*Completed: 2026-03-10*
