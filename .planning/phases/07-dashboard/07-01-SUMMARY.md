---
phase: 07-dashboard
plan: 01
subsystem: ui
tags: [nextjs, proxy, routing, route-groups, multi-domain]

# Dependency graph
requires:
  - phase: 08-landing-page
    provides: Landing page components (Navbar, Hero, TrustBar, PricingTable, etc.)
provides:
  - proxy.ts with hostname-based routing to (marketing) and (cloud) route groups
  - (marketing) route group with landing page and marketing metadata
  - Simplified root layout with template title pattern
affects: [07-02, 07-03, 07-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [proxy.ts hostname routing, Next.js route groups, template title metadata]

key-files:
  created:
    - frontend/src/proxy.ts
    - frontend/src/app/(marketing)/layout.tsx
    - frontend/src/app/(marketing)/page.tsx
  modified:
    - frontend/src/app/layout.tsx

key-decisions:
  - "proxy.ts (not middleware.ts) for Next.js 16 hostname routing convention"
  - "Template title pattern in root layout for flexible per-group page titles"
  - "Marketing metadata (title, OG tags) moved to (marketing)/layout.tsx"

patterns-established:
  - "Route group pattern: (marketing) and (cloud) separate UI domains within single Next.js app"
  - "Local dev routing: ?site=cloud query param for testing cloud routes without /etc/hosts"

requirements-completed: [DASH-01]

# Metrics
duration: 2min
completed: 2026-03-02
---

# Phase 7 Plan 01: Multi-Domain Routing Summary

**proxy.ts with hostname-based routing separating marketing (gpu.ai) and cloud (cloud.gpu.ai) route groups, landing page migrated to (marketing) route group**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-02T21:24:00Z
- **Completed:** 2026-03-02T21:25:47Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Created proxy.ts with hostname routing (cloud.gpu.ai -> (cloud), default -> (marketing)) and ?site=cloud local dev support
- Migrated landing page verbatim into (marketing) route group with marketing-specific metadata
- Simplified root layout to template title pattern, removing marketing-specific metadata

## Task Commits

Each task was committed atomically:

1. **Task 1: Create proxy.ts and (marketing) route group** - `6c829e4` (feat)
2. **Task 2: Simplify root layout and delete old page.tsx** - `e71f522` (refactor)

## Files Created/Modified
- `frontend/src/proxy.ts` - Hostname-based routing with matcher excluding API/static files
- `frontend/src/app/(marketing)/layout.tsx` - Marketing metadata (title, description, openGraph)
- `frontend/src/app/(marketing)/page.tsx` - Landing page content (verbatim from old page.tsx)
- `frontend/src/app/layout.tsx` - Simplified to template title, kept fonts and html/body shell
- `frontend/src/app/page.tsx` - Deleted (content moved to (marketing)/page.tsx)

## Decisions Made
- Used proxy.ts with named `proxy` export per Next.js 16 convention (not middleware.ts)
- Template title `{ template: "%s | GPU.ai", default: "GPU.ai" }` allows each route group to set its own page title
- Marketing metadata (title, description, openGraph) moved to (marketing)/layout.tsx so cloud route group can have its own metadata

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- .next build cache referenced deleted page.tsx causing stale TypeScript error - resolved by cleaning .next directory before tsc check

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- (marketing) route group with landing page is ready
- proxy.ts routing foundation in place for (cloud) route group (Plan 02)
- Root layout simplified and ready for ClerkProvider wrapping (Plan 03)

## Self-Check: PASSED

All files verified present, old page.tsx confirmed deleted, both task commits (6c829e4, e71f522) found in git log.

---
*Phase: 07-dashboard*
*Completed: 2026-03-02*
