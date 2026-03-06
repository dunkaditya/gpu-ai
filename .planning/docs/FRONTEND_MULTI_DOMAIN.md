# Multi-Domain Routing: gpu.ai + cloud.gpu.ai

## Context

The frontend currently serves a single landing page at `/`. We need to support two domains from one Next.js app:
- **gpu.ai** → marketing/landing page
- **cloud.gpu.ai** → dashboard/cloud console

Instead of splitting into two apps, we use Next.js middleware + route groups. This keeps shared components, fonts, CSS, and the build pipeline unified.

## File Structure (after changes)

```
src/
  middleware.ts                         # NEW — hostname-based routing
  app/
    layout.tsx                          # MODIFY — keep fonts/CSS, move metadata down
    globals.css                         # unchanged
    (marketing)/
      layout.tsx                        # NEW — marketing metadata
      page.tsx                          # MOVED from src/app/page.tsx
    (cloud)/
      layout.tsx                        # NEW — sidebar + topbar shell
      page.tsx                          # NEW — redirects to /instances
      instances/
        page.tsx                        # NEW — GPU instances list (mock data)
      settings/
        page.tsx                        # NEW — placeholder
  components/
    landing/                            # unchanged
    cloud/                              # NEW
      DashboardSidebar.tsx              # sidebar nav
      DashboardTopbar.tsx               # top bar with user placeholder
      InstancesTable.tsx                # instances table
      StatusBadge.tsx                   # status pill component
    ui/                                 # unchanged
  lib/
    constants.ts                        # unchanged
    utils.ts                            # unchanged
    mock-data.ts                        # NEW — mock instance data
```

## Implementation Steps

### 1. Create `src/middleware.ts`
- Read `Host` header, strip port
- If host matches `cloud.gpu.ai` (or `?site=cloud` for local dev) → rewrite to `/(cloud)` route group
- Otherwise → rewrite to `/(marketing)` route group
- Skip `/api/`, `/_next/`, static files

### 2. Create `src/app/(marketing)/` route group
- `layout.tsx` — marketing-specific metadata (title, OG tags from current root layout)
- `page.tsx` — move current `src/app/page.tsx` verbatim (Navbar, Hero, sections, Footer)
- Delete old `src/app/page.tsx`

### 3. Simplify `src/app/layout.tsx`
- Keep: font loading (Vremena Grotesk + Necto Mono), `globals.css` import, html/body shell
- Move out: marketing metadata → `(marketing)/layout.tsx`
- Use template title: `{ template: "%s | GPU.ai", default: "GPU.ai" }`

### 4. Create `src/lib/mock-data.ts`
- `MockInstance` type + `MOCK_INSTANCES` array (5 instances with varied statuses)

### 5. Create `src/components/cloud/` (4 components)
- **DashboardSidebar** — nav links (Instances, Volumes, API Keys, Billing, Settings), active state via `usePathname()`, GPU.ai logo
- **DashboardTopbar** — breadcrumb placeholder + hardcoded user/sign-out (Clerk replaces later)
- **InstancesTable** — desktop table + mobile cards, uses existing theme classes
- **StatusBadge** — colored pill (running=green, stopped=gray, provisioning=purple, error=red)

### 6. Create `src/app/(cloud)/` route group
- `layout.tsx` — `flex h-screen` shell with `<DashboardSidebar />` + `<DashboardTopbar />` + scrollable `<main>`
- `page.tsx` — `redirect("/instances")`
- `instances/page.tsx` — header + `<InstancesTable />` with mock data + "Launch Instance" button
- `settings/page.tsx` — placeholder text

## Local Dev

- `localhost:3001` → marketing (default)
- `localhost:3001?site=cloud` → cloud dashboard
- No `/etc/hosts` changes needed

## Key Files to Modify/Create

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

## Verification

1. `npm run dev` starts without errors
2. `localhost:3001` shows the landing page (Navbar, Hero, sections, Footer)
3. `localhost:3001?site=cloud` shows dashboard with sidebar, topbar, instances table
4. `localhost:3001/instances?site=cloud` shows instances page
5. `localhost:3001/settings?site=cloud` shows settings placeholder
6. API proxy (`/api/v1/*`) still works (middleware skips it)
7. No TypeScript errors (`npx tsc --noEmit`)
