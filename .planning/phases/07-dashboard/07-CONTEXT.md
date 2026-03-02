# Phase 7: Dashboard - Context

**Gathered:** 2026-03-02
**Status:** Ready for planning
**Source:** PRD Express Path (.planning/docs/FRONTEND_MULTI_DOMAIN.md)

<domain>
## Phase Boundary

Multi-domain routing for gpu.ai (marketing) + cloud.gpu.ai (dashboard) from a single Next.js app. Uses middleware + route groups to unify shared components, fonts, CSS, and build pipeline. Delivers a working cloud dashboard shell with sidebar, topbar, instances table (mock data), and settings placeholder.

</domain>

<decisions>
## Implementation Decisions

### Routing Architecture
- Use Next.js middleware (`src/middleware.ts`) for hostname-based routing — NOT separate apps
- Read `Host` header, strip port
- `cloud.gpu.ai` (or `?site=cloud` for local dev) → rewrite to `/(cloud)` route group
- All other hosts → rewrite to `/(marketing)` route group
- Skip `/api/`, `/_next/`, and static files in middleware

### Marketing Route Group `(marketing)/`
- `layout.tsx` — marketing-specific metadata (title, OG tags moved from current root layout)
- `page.tsx` — move current `src/app/page.tsx` verbatim (Navbar, Hero, sections, Footer)
- Delete old `src/app/page.tsx` after move

### Root Layout Simplification
- Keep: font loading (Vremena Grotesk + Necto Mono), `globals.css` import, html/body shell
- Move out: marketing metadata → `(marketing)/layout.tsx`
- Use template title: `{ template: "%s | GPU.ai", default: "GPU.ai" }`

### Mock Data Layer
- `src/lib/mock-data.ts` with `MockInstance` type + `MOCK_INSTANCES` array (5 instances with varied statuses)

### Cloud Components (`src/components/cloud/`)
- **DashboardSidebar** — nav links (Instances, Volumes, API Keys, Billing, Settings), active state via `usePathname()`, GPU.ai logo
- **DashboardTopbar** — breadcrumb placeholder + hardcoded user/sign-out (Clerk replaces later)
- **InstancesTable** — desktop table + mobile cards, uses existing theme classes
- **StatusBadge** — colored pill (running=green, stopped=gray, provisioning=purple, error=red)

### Cloud Route Group `(cloud)/`
- `layout.tsx` — `flex h-screen` shell with `<DashboardSidebar />` + `<DashboardTopbar />` + scrollable `<main>`
- `page.tsx` — `redirect("/instances")`
- `instances/page.tsx` — header + `<InstancesTable />` with mock data + "Launch Instance" button
- `settings/page.tsx` — placeholder text

### Local Dev
- `localhost:3001` → marketing (default)
- `localhost:3001?site=cloud` → cloud dashboard
- No `/etc/hosts` changes needed

### Claude's Discretion
- Specific Tailwind classes and styling details for cloud components
- Exact mock data values (GPU types, IPs, statuses)
- Mobile breakpoint handling for dashboard sidebar (collapsible vs hidden)
- Breadcrumb implementation details in topbar
- Exact color values for StatusBadge states (should align with existing design system)

</decisions>

<specifics>
## Specific Ideas

### File Actions Table
| File | Action |
|------|--------|
| `frontend/src/middleware.ts` | Create |
| `frontend/src/app/layout.tsx` | Modify (slim down metadata) |
| `frontend/src/app/page.tsx` | Delete (moved) |
| `frontend/src/app/(marketing)/layout.tsx` | Create |
| `frontend/src/app/(marketing)/page.tsx` | Create (content from old page.tsx) |
| `frontend/src/app/(cloud)/layout.tsx` | Create |
| `frontend/src/app/(cloud)/page.tsx` | Create |
| `frontend/src/app/(cloud)/instances/page.tsx` | Create |
| `frontend/src/app/(cloud)/settings/page.tsx` | Create |
| `frontend/src/components/cloud/DashboardSidebar.tsx` | Create |
| `frontend/src/components/cloud/DashboardTopbar.tsx` | Create |
| `frontend/src/components/cloud/InstancesTable.tsx` | Create |
| `frontend/src/components/cloud/StatusBadge.tsx` | Create |
| `frontend/src/lib/mock-data.ts` | Create |

### Verification Criteria
1. `npm run dev` starts without errors
2. `localhost:3001` shows the landing page (Navbar, Hero, sections, Footer)
3. `localhost:3001?site=cloud` shows dashboard with sidebar, topbar, instances table
4. `localhost:3001/instances?site=cloud` shows instances page
5. `localhost:3001/settings?site=cloud` shows settings placeholder
6. API proxy (`/api/v1/*`) still works (middleware skips it)
7. No TypeScript errors (`npx tsc --noEmit`)

</specifics>

<deferred>
## Deferred Ideas

- Clerk authentication integration (to be wired in later)
- Real API data replacing mock instances
- Volumes, API Keys, Billing pages (only nav links for now)
- `/etc/hosts` setup for true subdomain testing

</deferred>

---

*Phase: 07-dashboard*
*Context gathered: 2026-03-02 via PRD Express Path*
