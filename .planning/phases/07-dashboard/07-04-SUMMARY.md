---
phase: 07-dashboard
plan: 04
subsystem: ui
tags: [nextjs, react, swr, api-client, dashboard, gpu-availability, ssh-keys, billing, typescript]

# Dependency graph
requires:
  - phase: 07-dashboard-02
    provides: Dashboard shell layout, InstancesTable, StatusBadge, mock data layer
  - phase: 07-dashboard-03
    provides: Clerk authentication protecting cloud routes
provides:
  - TypeScript interfaces matching all Go backend API responses
  - Typed API client with SWR fetcher for all /api/v1/* endpoints
  - GPU availability page with filtering, sorting, and launch-from-row
  - SSH key management page with add/delete operations
  - Billing dashboard with period selector and usage sessions table
  - Instances page upgraded from mock data to live SWR polling
  - LaunchInstanceForm modal for instance provisioning
affects: []

# Tech tracking
tech-stack:
  added: [swr]
  patterns: [SWR data fetching with auto-refresh, typed API client layer, modal overlay pattern]

key-files:
  created:
    - frontend/src/lib/types.ts
    - frontend/src/lib/api.ts
    - frontend/src/components/cloud/GPUAvailabilityTable.tsx
    - frontend/src/components/cloud/LaunchInstanceForm.tsx
    - frontend/src/components/cloud/SSHKeyManager.tsx
    - frontend/src/components/cloud/BillingDashboard.tsx
    - frontend/src/app/(cloud)/gpu-availability/page.tsx
    - frontend/src/app/(cloud)/ssh-keys/page.tsx
    - frontend/src/app/(cloud)/billing/page.tsx
  modified:
    - frontend/src/components/cloud/InstancesTable.tsx
    - frontend/src/components/cloud/StatusBadge.tsx
    - frontend/src/app/(cloud)/instances/page.tsx
    - frontend/package.json

key-decisions:
  - "SWR with refreshInterval for auto-polling: 10s for instances, 30s for availability, 60s for billing"
  - "StatusBadge and InstancesTable migrated from MockInstance to InstanceResponse type for type safety"
  - "LaunchInstanceForm as modal overlay with backdrop blur, reusable from both instances page and GPU availability table"
  - "Skeleton loading animations using bg-card-hover pulse pattern instead of generic spinners"

patterns-established:
  - "API client pattern: fetcher for SWR, named functions for mutations (createInstance, terminateInstance)"
  - "Dashboard page pattern: server layout for metadata + client page for data fetching"
  - "Error state pattern: red error text with retry button calling SWR mutate()"
  - "Empty state pattern: icon + message + actionable CTA"

requirements-completed: [DASH-03, DASH-04, DASH-06, DASH-07]

# Metrics
duration: 5min
completed: 2026-03-02
---

# Phase 7 Plan 04: Dashboard API Integration Summary

**SWR-powered dashboard pages with typed API client fetching GPU availability, instances, SSH keys, and billing from Go backend endpoints**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-02T21:45:37Z
- **Completed:** 2026-03-02T21:51:01Z
- **Tasks:** 2
- **Files modified:** 18

## Accomplishments
- Created complete TypeScript type layer matching all Go backend API response structs (InstanceResponse, AvailableOffering, SSHKeyResponse, BillingSessionResponse)
- Built typed API client with SWR fetcher and mutation functions for all /api/v1 endpoints
- Delivered 4 fully-functional dashboard pages: GPU availability with filtering/sorting, instances with live polling, SSH key management, billing usage with period selector
- Migrated instances page from static mock data to real-time SWR data fetching with 10-second auto-refresh

## Task Commits

Each task was committed atomically:

1. **Task 1: Install SWR, create API types and client** - `f20e866` (feat)
2. **Task 2: Build dashboard pages with real API data** - `d6322bf` (feat)

## Files Created/Modified
- `frontend/package.json` - Added swr dependency
- `frontend/src/lib/types.ts` - TypeScript interfaces for all API response types
- `frontend/src/lib/api.ts` - Typed fetch wrappers and SWR fetcher
- `frontend/src/components/cloud/StatusBadge.tsx` - Migrated from MockInstance to InstanceResponse type
- `frontend/src/components/cloud/InstancesTable.tsx` - Added terminate action, onRefresh callback, InstanceResponse type
- `frontend/src/components/cloud/GPUAvailabilityTable.tsx` - GPU offerings table with filter/sort and launch action
- `frontend/src/components/cloud/LaunchInstanceForm.tsx` - Modal form for instance provisioning
- `frontend/src/components/cloud/SSHKeyManager.tsx` - SSH key list with add/delete operations
- `frontend/src/components/cloud/BillingDashboard.tsx` - Usage sessions table with summary cards and period selector
- `frontend/src/app/(cloud)/instances/page.tsx` - Converted to client component with SWR
- `frontend/src/app/(cloud)/instances/layout.tsx` - Server component for metadata
- `frontend/src/app/(cloud)/gpu-availability/page.tsx` - GPU availability page
- `frontend/src/app/(cloud)/gpu-availability/layout.tsx` - GPU availability metadata
- `frontend/src/app/(cloud)/ssh-keys/page.tsx` - SSH keys page
- `frontend/src/app/(cloud)/ssh-keys/layout.tsx` - SSH keys metadata
- `frontend/src/app/(cloud)/billing/page.tsx` - Billing page
- `frontend/src/app/(cloud)/billing/layout.tsx` - Billing metadata

## Decisions Made
- SWR auto-refresh intervals tuned per page urgency: 10s instances (status changes matter), 30s availability (less volatile), 60s billing (slow-changing)
- StatusBadge and InstancesTable migrated from MockInstance to InstanceResponse -- types are structurally identical so no logic changes needed
- LaunchInstanceForm built as a reusable modal overlay with backdrop blur, used from both the instances page header button and GPU availability table row launch buttons
- Skeleton loading uses bg-card-hover pulse animations matching the design system, not generic spinners
- Server layout + client page pattern for metadata: each route has a layout.tsx (server) exporting Metadata and a page.tsx (client) with SWR hooks

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required

None - API endpoints are proxied via next.config.ts rewrites to localhost:9090. Backend must be running for data to load.

## Next Phase Readiness
- All dashboard pages are complete with real API integration
- Phase 07 (Dashboard) is fully delivered: routing, shell, auth, and API integration
- Mock data layer (mock-data.ts) can be removed in a future cleanup

## Self-Check: PASSED

All files verified present, both task commits (f20e866, d6322bf) found in git log. TypeScript compilation passes with zero errors.

---
*Phase: 07-dashboard*
*Completed: 2026-03-02*
