# Phase 7: Dashboard - Research

**Researched:** 2026-03-02
**Domain:** Next.js multi-domain routing, Clerk authentication, dashboard UI, API integration
**Confidence:** HIGH

## Summary

Phase 7 delivers the cloud dashboard by introducing multi-domain routing (gpu.ai for marketing, cloud.gpu.ai for the dashboard) within the existing Next.js 16 app using `proxy.ts` and route groups. The existing frontend already has a complete landing page (Phase 8), a design system with Tailwind v4, custom fonts (Vremena Grotesk + Necto Mono), and reusable UI components. The critical technical finding is that Next.js 16 renamed `middleware.ts` to `proxy.ts` with a `proxy` named export -- this is a departure from the CONTEXT.md which references `middleware.ts`. The CONTEXT.md already provides a detailed file action table and component specification, so the implementation is well-scoped.

The API layer is complete: instances CRUD, SSH key management, billing/usage endpoints, GPU availability, and SSE events are all functional at `/api/v1/*`. The Next.js config already proxies `/api/v1/*` to `localhost:9090`. Clerk integration adds authentication via `@clerk/nextjs` with `clerkMiddleware()` composing inside the `proxy.ts` function. The dashboard shell (sidebar, topbar, table) uses mock data initially, with real API integration replacing mocks via SWR or direct fetch in later iterations.

**Primary recommendation:** Use `proxy.ts` (not `middleware.ts`) for hostname-based routing with `clerkMiddleware()` composed inside. Build the cloud route group `(cloud)/` with a sidebar+topbar layout shell and mock data, then progressively wire real API calls for each dashboard page. Keep Clerk integration minimal for this phase (wrap in ClerkProvider, protect cloud routes, sign-in/sign-up pages).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Use Next.js middleware (`src/middleware.ts`) for hostname-based routing -- NOT separate apps
- Read `Host` header, strip port
- `cloud.gpu.ai` (or `?site=cloud` for local dev) rewrite to `/(cloud)` route group
- All other hosts rewrite to `/(marketing)` route group
- Skip `/api/`, `/_next/`, and static files in middleware
- Marketing route group gets moved landing page content
- Root layout keeps font loading + globals.css, metadata moves to marketing layout
- Template title: `{ template: "%s | GPU.ai", default: "GPU.ai" }`
- Mock data layer in `src/lib/mock-data.ts` with `MockInstance` type + 5 instances
- Cloud components in `src/components/cloud/`: DashboardSidebar, DashboardTopbar, InstancesTable, StatusBadge
- Cloud route group: layout with sidebar+topbar shell, `/` redirects to `/instances`, instances page, settings placeholder
- Local dev: `localhost:3001` = marketing, `?site=cloud` = cloud dashboard

### Claude's Discretion
- Specific Tailwind classes and styling details for cloud components
- Exact mock data values (GPU types, IPs, statuses)
- Mobile breakpoint handling for dashboard sidebar (collapsible vs hidden)
- Breadcrumb implementation details in topbar
- Exact color values for StatusBadge states (should align with existing design system)

### Deferred Ideas (OUT OF SCOPE)
- Clerk authentication integration (to be wired in later)
- Real API data replacing mock instances
- Volumes, API Keys, Billing pages (only nav links for now)
- `/etc/hosts` setup for true subdomain testing
</user_constraints>

**IMPORTANT NOTE ON CONTEXT.md vs REALITY:** The CONTEXT.md references `src/middleware.ts`, but Next.js 16 (which this project uses -- `"next": "16.1.6"`) has deprecated `middleware.ts` in favor of `proxy.ts`. The file should be named `proxy.ts` with a `proxy` named export. The logic is identical -- only the filename and export name change. This is the single most important correction for the planner.

**SECOND IMPORTANT NOTE:** The CONTEXT.md defers Clerk authentication, but requirements DASH-02 through DASH-08 require it. The phase must address Clerk setup to satisfy requirements. The CONTEXT.md was written before the full requirements mapping was established -- Clerk is required to fulfill DASH-02 ("User can sign up and log in via Clerk"). The planner should include Clerk integration as a separate plan/wave after the multi-domain routing and mock dashboard are complete.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DASH-01 | Landing page describes the product | Already complete (Phase 8). Moving to `(marketing)/page.tsx` preserves it. |
| DASH-02 | User can sign up and log in via Clerk | Clerk `@clerk/nextjs` with `ClerkProvider`, sign-in/sign-up catch-all routes, `clerkMiddleware()` in `proxy.ts` |
| DASH-03 | User can view real-time GPU availability with pricing | API endpoint `GET /api/v1/gpu/available` exists. Dashboard page fetches and displays `AvailableOffering[]` data. |
| DASH-04 | User can provision a GPU instance from the dashboard | API endpoint `POST /api/v1/instances` exists with `CreateInstanceRequest`. Dashboard form submits to this API. |
| DASH-05 | User can view and manage running instances | API endpoint `GET /api/v1/instances` returns `InstanceResponse[]`. InstancesTable component displays these. `DELETE /api/v1/instances/{id}` for termination. |
| DASH-06 | User can manage SSH keys | API endpoints `GET/POST/DELETE /api/v1/ssh-keys` exist. SSH keys page with list, add, and delete functionality. |
| DASH-07 | User can view billing usage and costs | API endpoint `GET /api/v1/billing/usage` with `period`, `summary` params. Billing page displays sessions and hourly buckets. |
| DASH-08 | Dashboard displays instance status with SSH connection command | `InstanceResponse.Connection.SSHCommand` field already returned by API. StatusBadge component shows status colors. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| next | 16.1.6 | Framework (already installed) | Project uses Next.js 16 with App Router, Turbopack |
| react / react-dom | 19.2.3 | UI library (already installed) | Already in project |
| @clerk/nextjs | latest | Authentication (sign-up, sign-in, session, org) | Project backend already uses Clerk JWTs; frontend needs matching SDK |
| tailwindcss | ^4 | Styling (already installed) | Already in project with custom design system |
| swr | ^2 | Client-side data fetching with caching and revalidation | Lightweight (5.3KB), Vercel-backed, perfect for Next.js dashboard polling/refresh patterns |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| clsx + tailwind-merge | (already installed) | Conditional class merging | Already in project via `cn()` utility |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SWR | TanStack Query | TanStack Query is 3x larger (16KB), has DevTools. SWR is simpler, Vercel-native, sufficient for this dashboard's needs. |
| proxy.ts hostname routing | Separate Next.js apps | Separate apps lose shared components, fonts, CSS, build pipeline. Single app with route groups is cleaner. |
| Clerk | NextAuth.js | Backend already uses Clerk JWTs (auth.ClaimsFromContext). Switching auth provider would require backend changes. |

**Installation:**
```bash
cd frontend
npm install @clerk/nextjs swr
```

## Architecture Patterns

### Recommended Project Structure
```
frontend/src/
  proxy.ts                          # Hostname-based routing + Clerk auth
  app/
    layout.tsx                      # Root: fonts, globals.css, ClerkProvider, html/body
    (marketing)/
      layout.tsx                    # Marketing metadata (title, OG tags)
      page.tsx                      # Landing page (moved from app/page.tsx)
    (cloud)/
      layout.tsx                    # Dashboard shell: sidebar + topbar + scrollable main
      page.tsx                      # Redirects to /instances
      instances/
        page.tsx                    # Instances list (mock then real API)
      gpu-availability/
        page.tsx                    # GPU availability table (DASH-03)
      ssh-keys/
        page.tsx                    # SSH key management (DASH-06)
      billing/
        page.tsx                    # Billing usage (DASH-07)
      settings/
        page.tsx                    # Placeholder
    sign-in/
      [[...sign-in]]/page.tsx       # Clerk sign-in (catch-all for multi-step flows)
    sign-up/
      [[...sign-up]]/page.tsx       # Clerk sign-up (catch-all for multi-step flows)
  components/
    cloud/
      DashboardSidebar.tsx          # Sidebar nav with active state
      DashboardTopbar.tsx           # Topbar with breadcrumb + user button
      InstancesTable.tsx            # Desktop table + mobile cards
      StatusBadge.tsx               # Colored status pill
      GPUAvailabilityTable.tsx      # GPU offerings display
      SSHKeyManager.tsx             # SSH key list + add/delete
      BillingDashboard.tsx          # Usage sessions + hourly chart
      LaunchInstanceForm.tsx        # GPU provisioning form
    landing/                        # Existing (unchanged)
    ui/                             # Existing (unchanged)
  lib/
    mock-data.ts                    # Mock instances for initial build
    api.ts                          # API client functions (typed fetch wrappers)
    constants.ts                    # Existing
    utils.ts                        # Existing
```

### Pattern 1: proxy.ts with Hostname Routing + Clerk
**What:** Single proxy file that handles hostname-based rewrites AND Clerk authentication
**When to use:** Every request (filtered by matcher)
**Example:**
```typescript
// Source: https://nextjs.org/docs/app/api-reference/file-conventions/proxy
// Source: https://clerk.com/docs/reference/nextjs/clerk-middleware
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'
import { NextResponse, type NextRequest } from 'next/server'

const isCloudRoute = createRouteMatcher(['/(cloud)(.*)'])
const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/(marketing)(.*)',
])

export default clerkMiddleware(async (auth, req) => {
  const hostname = req.headers.get('host')?.split(':')[0] ?? ''
  const { searchParams, pathname } = req.nextUrl

  // Local dev: ?site=cloud query param override
  const isCloud = hostname === 'cloud.gpu.ai' || searchParams.get('site') === 'cloud'

  if (isCloud) {
    // Protect all cloud routes
    await auth.protect()
    // Rewrite to (cloud) route group
    const url = req.nextUrl.clone()
    url.pathname = `/(cloud)${pathname}`
    return NextResponse.rewrite(url)
  }

  // Default: marketing
  const url = req.nextUrl.clone()
  url.pathname = `/(marketing)${pathname}`
  return NextResponse.rewrite(url)
})

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico|fonts|.*\\.(?:svg|png|jpg|jpeg|gif|webp|woff2?|ico)$).*)',
  ],
}
```

### Pattern 2: Route Group Layouts
**What:** Separate root-level layouts for marketing and cloud domains
**When to use:** Different UI shells for different parts of the app
**Example:**
```typescript
// src/app/(cloud)/layout.tsx
export const metadata = { title: 'Dashboard' }

export default function CloudLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen">
      <DashboardSidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <DashboardTopbar />
        <main className="flex-1 overflow-y-auto bg-bg p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
```

### Pattern 3: Typed API Client
**What:** Centralized fetch wrappers that return typed responses
**When to use:** All dashboard API calls
**Example:**
```typescript
// src/lib/api.ts
const API_BASE = '/api/v1'

export async function fetchInstances(): Promise<InstanceResponse[]> {
  const res = await fetch(`${API_BASE}/instances`)
  if (!res.ok) throw new Error('Failed to fetch instances')
  const data = await res.json()
  return data.instances
}

export async function fetchGPUAvailability(filters?: {
  gpu_model?: string
  region?: string
  tier?: string
}): Promise<AvailableOffering[]> {
  const params = new URLSearchParams()
  if (filters?.gpu_model) params.set('gpu_model', filters.gpu_model)
  if (filters?.region) params.set('region', filters.region)
  if (filters?.tier) params.set('tier', filters.tier)
  const res = await fetch(`${API_BASE}/gpu/available?${params}`)
  if (!res.ok) throw new Error('Failed to fetch GPU availability')
  const data = await res.json()
  return data.available
}
```

### Pattern 4: SWR Data Fetching in Client Components
**What:** SWR hooks for real-time dashboard data with automatic refresh
**When to use:** Dashboard pages that display API data
**Example:**
```typescript
'use client'
import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(r => r.json())

export function InstancesPage() {
  const { data, error, isLoading, mutate } = useSWR('/api/v1/instances', fetcher, {
    refreshInterval: 10000, // Poll every 10 seconds for status updates
  })

  if (isLoading) return <LoadingSkeleton />
  if (error) return <ErrorState />

  return <InstancesTable instances={data.instances} onRefresh={() => mutate()} />
}
```

### Anti-Patterns to Avoid
- **Putting Clerk env vars in client code without NEXT_PUBLIC_ prefix:** Clerk publishable key MUST be `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`. Secret key stays server-side as `CLERK_SECRET_KEY`.
- **Using middleware.ts instead of proxy.ts:** Next.js 16 deprecated `middleware.ts`. Using the old name will cause a deprecation warning and will break in future versions.
- **Fetching API data in Server Components behind auth:** Server Components in `(cloud)/` route group run at build time for static pages. Dashboard data is per-user and should use client-side fetching (SWR) or Server Components with `await auth()`.
- **Multiple root layouts without a shared root:** If `(marketing)` and `(cloud)` each have their own root `layout.tsx`, navigation between them triggers full page reloads. Keep a single root `app/layout.tsx` that wraps `ClerkProvider` and loads fonts.
- **Hardcoding hostnames without local dev fallback:** Always check `?site=cloud` query param for local development, not just the `Host` header.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT verification | Custom JWT parser | `@clerk/nextjs` `clerkMiddleware()` | Clerk SDK handles token refresh, JWKS rotation, org context |
| Route protection | Custom auth checks per route | `createRouteMatcher()` + `auth.protect()` | Clerk's pattern is battle-tested, handles redirects automatically |
| Data fetching + caching | Custom fetch + useState + useEffect | SWR | SWR handles cache invalidation, revalidation, error retry, deduplication |
| Hostname routing | Custom Express server | Next.js proxy.ts + NextResponse.rewrite | Built-in, zero-config, works with Vercel deployment |
| Form state management | Custom useState per field | React controlled inputs or `useActionState` | React 19 has built-in form primitives |
| Loading states | Custom isLoading booleans | SWR's `isLoading`/`isValidating` | SWR tracks loading state automatically |

**Key insight:** The entire auth + routing layer is handled by two libraries (Clerk + Next.js proxy). Custom solutions would need to handle JWT refresh, JWKS rotation, CSRF, redirect loops, and org-scoping -- all of which Clerk provides out of the box.

## Common Pitfalls

### Pitfall 1: middleware.ts vs proxy.ts Naming
**What goes wrong:** Using `middleware.ts` instead of `proxy.ts` in Next.js 16 triggers deprecation warnings and may cause confusion with Clerk's `clerkMiddleware` function name.
**Why it happens:** Most documentation and tutorials still reference `middleware.ts`. Next.js 16 renamed it.
**How to avoid:** Name the file `proxy.ts`, export function as `proxy` (or use `export default clerkMiddleware(...)` which handles both).
**Warning signs:** Console warning about deprecated middleware file.

### Pitfall 2: Route Group Conflicts
**What goes wrong:** Both `(marketing)/about/page.tsx` and `(cloud)/about/page.tsx` resolve to `/about` and cause a build error.
**Why it happens:** Route groups are organizational only -- the parenthesized name is not part of the URL path.
**How to avoid:** Ensure no two route groups define the same URL path. Cloud pages live at `/instances`, `/billing`, etc. Marketing pages live at `/`, `/sign-in`, `/sign-up`.
**Warning signs:** Next.js build error: "Conflicting app routes."

### Pitfall 3: ClerkProvider Placement
**What goes wrong:** `auth()` calls in Server Components fail with "Clerk can't detect usage of clerkMiddleware()" error.
**Why it happens:** ClerkProvider must wrap the entire app in the root layout, AND clerkMiddleware must be exported from proxy.ts.
**How to avoid:** 1) Export `clerkMiddleware()` from `proxy.ts`. 2) Wrap `{children}` with `<ClerkProvider>` in `app/layout.tsx`.
**Warning signs:** Runtime error mentioning clerkMiddleware not detected.

### Pitfall 4: API Proxy Interference
**What goes wrong:** The proxy.ts rewrites all requests including `/api/v1/*` calls, breaking the Next.js config rewrite to `localhost:9090`.
**Why it happens:** The matcher in proxy.ts is too broad and catches API routes.
**How to avoid:** Exclude `api` from the proxy.ts matcher: `'/((?!api|_next/static|...)'`. The existing `next.config.ts` rewrite handles `/api/v1/*` proxy to Go backend.
**Warning signs:** API calls return 404 or HTML instead of JSON.

### Pitfall 5: Full Page Reload Between Domains
**What goes wrong:** Navigating from marketing to cloud (or vice versa) triggers a full page reload if they use different root layouts.
**Why it happens:** Next.js route groups with separate root layouts cause full page reloads on cross-group navigation.
**How to avoid:** Use a single root `app/layout.tsx` (fonts + ClerkProvider + body). Marketing and cloud layouts are nested, not root. This is already the plan per CONTEXT.md.
**Warning signs:** Flash of white screen when navigating between marketing and cloud.

### Pitfall 6: Clerk Keyless Mode in Development
**What goes wrong:** Developer expects auth to work immediately but gets confused by keyless mode behavior.
**Why it happens:** Clerk SDK operates in keyless mode without env vars, using auto-generated temporary keys.
**How to avoid:** For development, keyless mode is fine. For testing with real auth flows, set `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` and `CLERK_SECRET_KEY` in `.env.local`.
**Warning signs:** Auth works but with a Clerk dev banner and temporary test accounts.

## Code Examples

Verified patterns from official sources:

### Hostname Detection in proxy.ts
```typescript
// Source: https://nextjs.org/docs/app/api-reference/file-conventions/proxy
export function proxy(request: NextRequest) {
  const hostname = request.headers.get('host')?.split(':')[0] ?? ''
  const { searchParams, pathname } = request.nextUrl

  // Skip static files and API routes
  // (handled by matcher config, but belt-and-suspenders)

  const isCloud = hostname === 'cloud.gpu.ai'
    || searchParams.get('site') === 'cloud'

  const url = request.nextUrl.clone()
  if (isCloud) {
    url.pathname = `/(cloud)${pathname}`
  } else {
    url.pathname = `/(marketing)${pathname}`
  }
  return NextResponse.rewrite(url)
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico|fonts|.*\\.(?:svg|png|jpg|jpeg|gif|webp|woff2?|ico)$).*)',
  ],
}
```

### Composing Clerk + Hostname Routing
```typescript
// Source: https://clerk.com/docs/reference/nextjs/clerk-middleware
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
])

export default clerkMiddleware(async (auth, req) => {
  const hostname = req.headers.get('host')?.split(':')[0] ?? ''
  const isCloud = hostname === 'cloud.gpu.ai'
    || req.nextUrl.searchParams.get('site') === 'cloud'

  // Protect cloud routes
  if (isCloud && !isPublicRoute(req)) {
    await auth.protect()
  }

  // Rewrite to appropriate route group
  const url = req.nextUrl.clone()
  url.pathname = isCloud
    ? `/(cloud)${req.nextUrl.pathname}`
    : `/(marketing)${req.nextUrl.pathname}`
  return NextResponse.rewrite(url)
})

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico|fonts|.*\\.(?:svg|png|jpg|jpeg|gif|webp|woff2?|ico)$).*)',
  ],
}
```

### Root Layout with ClerkProvider
```typescript
// Source: https://clerk.com/docs/nextjs/getting-started/quickstart
import { ClerkProvider } from '@clerk/nextjs'
import localFont from 'next/font/local'
import './globals.css'

const vremenaGrotesk = localFont({
  src: '../../public/fonts/vremena-grotesk.woff2',
  variable: '--font-vremena-grotesk',
  display: 'swap',
})

const nectoMono = localFont({
  src: '../../public/fonts/necto-mono.woff2',
  variable: '--font-necto-mono',
  display: 'swap',
})

export const metadata = {
  title: { template: '%s | GPU.ai', default: 'GPU.ai' },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider>
      <html lang="en" className="dark">
        <body className={`${vremenaGrotesk.variable} ${nectoMono.variable} antialiased`}>
          {children}
        </body>
      </html>
    </ClerkProvider>
  )
}
```

### Clerk Sign-In Page (Catch-All)
```typescript
// Source: https://clerk.com/docs/nextjs/guides/development/custom-sign-in-or-up-page
// File: app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-bg">
      <SignIn />
    </div>
  )
}
```

### API Response Types (from existing Go backend)
```typescript
// Mirroring internal/api/handlers.go InstanceResponse
interface InstanceResponse {
  id: string
  name?: string
  status: 'starting' | 'running' | 'stopping' | 'terminated' | 'error'
  gpu_type: string
  gpu_count: number
  tier: 'spot' | 'on_demand'
  region: string
  price_per_hour: number
  connection?: {
    hostname: string
    port: number
    ssh_command: string
  }
  error_reason?: string
  created_at: string
  ready_at?: string
  terminated_at?: string
}

// Mirroring internal/availability/types.go AvailableOffering
interface AvailableOffering {
  gpu_model: string
  vram_gb: number
  cpu_cores: number
  ram_gb: number
  storage_gb: number
  price_per_hour: number
  region: string
  tier: 'spot' | 'on_demand'
  available_count: number
  avg_uptime_pct: number
}

// Mirroring internal/api/handlers_ssh_keys.go SSHKeyResponse
interface SSHKeyResponse {
  id: string
  name: string
  fingerprint: string
  created_at: string
}

// Mirroring internal/api/handlers_billing.go
interface BillingSessionResponse {
  id: string
  instance_id: string
  gpu_type: string
  gpu_count: number
  price_per_hour: number
  started_at: string
  ended_at?: string
  duration_seconds?: number
  total_cost?: number
  estimated_cost?: number
  is_active: boolean
}

interface UsageResponse {
  sessions: BillingSessionResponse[]
  total_cost: number
  currency: string
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `middleware.ts` with `middleware` export | `proxy.ts` with `proxy` export | Next.js 16 (Oct 2025) | File rename + export rename; logic unchanged. Deprecated `middleware.ts` still works but shows warning. |
| `experimental.ppr` flag | `cacheComponents: true` in next.config | Next.js 16 (Oct 2025) | PPR flag removed entirely. Not needed for this phase. |
| Clerk `authMiddleware()` | Clerk `clerkMiddleware()` | Clerk SDK 5+ | `authMiddleware` deprecated; `clerkMiddleware` is the current API. All routes public by default. |
| sync `params`/`searchParams` | async `await params` / `await searchParams` | Next.js 16 | Must use `await` for params in Server Components. |
| sync `cookies()`/`headers()` | async `await cookies()` / `await headers()` | Next.js 16 | All these APIs are now async in Next.js 16. |

**Deprecated/outdated:**
- `middleware.ts`: Renamed to `proxy.ts` in Next.js 16. Still works but deprecated.
- `authMiddleware()` from Clerk: Replaced by `clerkMiddleware()`.
- `experimental.dynamicIO`: Renamed to `cacheComponents`.
- Sync params/searchParams access: Must be awaited in Next.js 16.

## Open Questions

1. **Sign-in/sign-up page placement with route groups**
   - What we know: Clerk needs catch-all routes at `/sign-in/[[...sign-in]]/page.tsx` and `/sign-up/[[...sign-up]]/page.tsx`. These need to be accessible from both marketing and cloud domains.
   - What's unclear: Whether these pages should live at the root `app/` level (outside both route groups) or inside `(marketing)/`. If at root, the proxy.ts rewrite logic needs to handle them specially.
   - Recommendation: Place sign-in/sign-up at root `app/` level (outside route groups). The proxy.ts should NOT rewrite paths starting with `/sign-in` or `/sign-up` to either route group. This makes them accessible from both domains.

2. **ClerkProvider and dark theme styling**
   - What we know: Clerk's pre-built components (`<SignIn />`, `<UserButton />`) have their own styling.
   - What's unclear: How well Clerk components integrate with the dark theme and custom fonts.
   - Recommendation: Use Clerk's `appearance` prop on `<ClerkProvider>` to pass dark theme configuration. Clerk supports `baseTheme` from `@clerk/themes` or custom CSS variables.

3. **Real API data vs mock data phasing**
   - What we know: CONTEXT.md defers "Real API data replacing mock instances". Requirements DASH-03 through DASH-08 require functional API integration.
   - What's unclear: Whether the phase should ship with mock data only or wire real API calls.
   - Recommendation: Build with mock data first (Plan 1), then wire real API calls (Plan 2+). The mock data layer provides a working UI to iterate on, and the API client layer replaces mocks cleanly. This satisfies CONTEXT.md's phased approach while ultimately fulfilling all DASH requirements.

## Sources

### Primary (HIGH confidence)
- [Next.js 16 Blog Post](https://nextjs.org/blog/next-16) - Confirmed proxy.ts rename, deprecation of middleware.ts, Node.js runtime default
- [Next.js proxy.ts API Reference](https://nextjs.org/docs/app/api-reference/file-conventions/proxy) - Full API reference with matcher, examples, migration guide
- [Next.js Route Groups](https://nextjs.org/docs/app/api-reference/file-conventions/route-groups) - Confirmed route group conventions, multiple layout support, caveats
- [Clerk clerkMiddleware Reference](https://clerk.com/docs/reference/nextjs/clerk-middleware) - Full API with createRouteMatcher, auth.protect, composition with custom logic
- [Clerk Next.js Quickstart](https://clerk.com/docs/nextjs/getting-started/quickstart) - ClerkProvider setup, proxy.ts integration, keyless mode
- [Clerk Custom Sign-In Page](https://clerk.com/docs/nextjs/guides/development/custom-sign-in-or-up-page) - Catch-all route pattern, env variables

### Secondary (MEDIUM confidence)
- [Next.js Multi-Tenant Guide](https://nextjs.org/docs/app/guides/multi-tenant) - Points to Platforms Starter Kit for hostname routing patterns
- [Existing codebase analysis] - Direct examination of Go API handlers, response types, and Next.js frontend structure

### Tertiary (LOW confidence)
- None. All findings verified with primary sources.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified against official docs and existing project setup
- Architecture: HIGH - proxy.ts pattern confirmed in Next.js 16 official docs; Clerk composition pattern confirmed in Clerk docs; route groups well-documented
- Pitfalls: HIGH - All pitfalls derived from official documentation caveats sections

**Research date:** 2026-03-02
**Valid until:** 2026-04-02 (30 days -- stable stack, well-documented patterns)
