# Phase 10: Frontend Dashboard with GPU Availability, Provisioning, and Instance Management - Research

**Researched:** 2026-03-09
**Domain:** Frontend dashboard redesign (Next.js 16, React 19, SWR, Tailwind CSS 4, Clerk)
**Confidence:** HIGH

## Summary

Phase 10 is a **redesign** of the existing Phase 7 cloud dashboard, not a greenfield build. All foundational code exists: routing (`/cloud/*`), API layer (`api.ts`, `types.ts`), SWR data fetching, Clerk auth integration, and a complete design system (Vremena Grotesk + Necto Mono, purple accents, `#09090f` dark background). The existing components are functional but basic -- the goal is to elevate them to production quality with a clean, spacious Lambda Labs aesthetic while using RunPod-style density where data scanning matters.

The primary work falls into five areas: (1) restructuring the sidebar navigation with "coming soon" items for future features, (2) converting the GPU availability page from a flat table to a card grid grouped by GPU model, (3) enhancing the instances table with clickable rows leading to a new `/cloud/instances/[id]` detail page plus inline rename and confirmation dialogs, (4) improving the launch modal with price confirmation and auto-filled GPU/region from availability cards, and (5) polishing existing pages (billing, SSH keys, settings) to match the elevated design standard.

**Critical gap:** The instance rename feature requires a new `PATCH /api/v1/instances/{id}` backend endpoint and a corresponding `UpdateInstanceName` DB function. Neither exists today. This must be built as part of this phase or scoped out.

**Primary recommendation:** Structure work as progressive enhancement of existing components rather than rewrites. Keep all existing SWR patterns, API functions, and type definitions. Add new routes (`/cloud/instances/[id]`), new components (GPU cards, confirmation dialogs, instance detail view), and refine existing components in place.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Clean & spacious layout (Lambda Labs style), not dense/data-rich (RunPod style)
- Generous whitespace, focused views, modern SaaS feel
- Exception: data tables (GPU availability, instances) can be denser where scanning matters
- Dark theme with existing design system (Vremena Grotesk + Necto Mono, purple accents, `#09090f` background)
- Phase 8 Vercel-style redesign was reverted -- do NOT use Geist fonts or Vercel visual patterns
- Reference RunPod and Lambda Labs dashboards for layout/UX patterns
- Sidebar nav structure at Claude's discretion -- reference RunPod and Lambda Labs
- Include "coming soon" items for future essential features (blank pages with coming-soon state)
- Minimal topbar: breadcrumb showing current page + Clerk user button (avatar, sign out) on the right
- Home page behavior at Claude's discretion (redirect to instances or overview dashboard)
- Launch from GPU availability table: click a GPU row -> pre-filled launch modal
- Launch modal shows price confirmation: estimated hourly cost, GPU specs (VRAM, RAM, storage), and tier
- Auto-attach all user's SSH keys to new instances (no key picker in modal)
- After launch: close modal, redirect to instances list, SWR auto-refreshes every 10s
- Enhanced instance table: name (editable/renamable), status badge, GPU type + count, region, hourly price, uptime, SSH command quick-copy, terminate action
- Rows clickable -> instance detail view at /cloud/instances/{id}
- Copy SSH command: one-click copy of `ssh -p PORT root@host`
- Terminate with confirmation dialog (prevent accidental deletion)
- Rename instances (custom names instead of just GPU type)
- Empty state: "No instances running" with CTA to GPU Availability page
- GPU Availability as card grid grouped by GPU model
- Each card: GPU name, VRAM, available count, spot + on-demand price side by side, region tags, launch button
- Filter bar with dropdowns: region, tier
- Sortable by price (ascending/descending)

### Claude's Discretion
- Exact sidebar nav items and ordering (reference RunPod/Lambda Labs, include coming-soon items)
- Whether to have an overview home page or redirect straight to instances
- Mobile responsive behavior for sidebar (collapsible vs hidden)
- Billing and SSH Keys page design (existing components exist, polish as needed)
- Settings page design
- Exact card layout and styling for GPU availability cards
- Instance detail page layout
- Loading skeletons and transition animations
- Error state designs

### Deferred Ideas (OUT OF SCOPE)
- Interactive pricing widget on landing page -- separate concern from dashboard
- GPU templates gallery (one-click deploy environments) -- future phase
- Real-time instance metrics/monitoring dashboard -- requires agent on instances
- Team/org management UI -- ADV-02 is a v2 requirement
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DASH-01 | Landing page describes the product | Already complete from Phase 7. Landing page is NOT in scope for this phase (per CONTEXT.md). |
| DASH-02 | User can sign up and log in via Clerk | Already complete. ClerkProvider in root layout, sign-in/sign-up pages exist, proxy.ts protects /cloud routes. |
| DASH-03 | User can view real-time GPU availability with pricing | Requires redesign from table to card grid. SWR fetching exists (30s refresh). New card component needed. |
| DASH-04 | User can provision a GPU instance from the dashboard | Exists but needs enhancement: pre-fill from card click, price confirmation, auto-SSH-key attachment. |
| DASH-05 | User can view and manage running instances | Requires enhancement: clickable rows, new detail page, rename, confirmation dialogs, uptime display. |
| DASH-06 | User can manage SSH keys | Exists. Polish styling to match new design standard. |
| DASH-07 | User can view billing usage and costs | Exists. Polish styling to match new design standard. |
| DASH-08 | Dashboard displays instance status with SSH connection command | Exists. Enhance with one-click copy, detail view connection info, and improved visual treatment. |
</phase_requirements>

## Standard Stack

### Core (Already Installed)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js | 16.1.6 | App Router, server/client components, API proxy rewrites | Already in use, project convention |
| React | 19.2.3 | UI framework | Already in use |
| SWR | 2.4.1 | Data fetching with stale-while-revalidate, mutation, optimistic updates | Already in use, Vercel ecosystem match |
| @clerk/nextjs | 6.39.0 | Authentication, UserButton component, route protection | Already in use |
| Tailwind CSS | 4.x | Utility-first CSS with @theme inline design tokens | Already in use |
| clsx + tailwind-merge | 2.1.1 / 3.5.0 | Conditional class composition via `cn()` utility | Already in use |

### No Additional Libraries Needed
The existing stack covers all requirements. Do NOT add:
- UI component libraries (shadcn/ui, Radix, Headless UI) -- project uses custom components
- Animation libraries (Framer Motion, Motion) -- CSS animations suffice for dashboard
- State management (Zustand, Jotai) -- SWR handles all server state; React state handles local state
- Form libraries (React Hook Form, Formik) -- forms are simple enough for controlled components

### Fonts (Already Configured)
| Font | Role | Loading |
|------|------|---------|
| Vremena Grotesk | Display/headings (--font-sans) | Local woff2 via next/font/local |
| Necto Mono | Body/UI/code (--font-mono) | Local woff2 via next/font/local |

## Architecture Patterns

### Existing Project Structure (Preserve)
```
frontend/src/
  app/
    layout.tsx              # Root: ClerkProvider, fonts, metadata
    globals.css             # Design system tokens, type classes, utilities
    cloud/
      layout.tsx            # Sidebar + topbar shell (server component)
      instances/
        layout.tsx          # Server metadata
        page.tsx            # Client page with SWR
        [id]/               # NEW: instance detail route
          page.tsx           # NEW: instance detail client page
      gpu-availability/
        layout.tsx          # Server metadata
        page.tsx            # Client page with SWR
      ssh-keys/
        layout.tsx
        page.tsx
      billing/
        layout.tsx
        page.tsx
      settings/
        page.tsx
  components/cloud/
    DashboardSidebar.tsx    # ENHANCE: add nav items, coming-soon badges
    DashboardTopbar.tsx     # ENHANCE: breadcrumb for nested routes
    InstancesTable.tsx      # ENHANCE: clickable rows, rename, uptime
    GPUAvailabilityTable.tsx # REDESIGN: table -> card grid
    LaunchInstanceForm.tsx  # ENHANCE: price confirmation, pre-fill
    StatusBadge.tsx         # KEEP AS-IS: already well-built
    SSHKeyManager.tsx       # POLISH: match new design
    BillingDashboard.tsx    # POLISH: match new design
    ConfirmDialog.tsx       # NEW: reusable confirmation modal
    GPUCard.tsx             # NEW: individual GPU availability card
    InstanceDetail.tsx      # NEW: full instance detail view
  lib/
    api.ts                  # EXTEND: add renameInstance function
    types.ts                # EXTEND: add RenameInstanceRequest
    utils.ts                # cn() utility -- keep as-is
  proxy.ts                  # Clerk middleware -- keep as-is
```

### Pattern 1: Server Layout + Client Page (Existing)
**What:** Server components define metadata in `layout.tsx`, client components handle SWR data in `page.tsx`.
**When to use:** Every route page that fetches data.
**Example:**
```typescript
// layout.tsx (server - metadata only)
export const metadata: Metadata = { title: "Instances" };
export default function Layout({ children }) { return children; }

// page.tsx (client - SWR data fetching)
"use client";
export default function Page() {
  const { data, mutate } = useSWR("/api/v1/instances", fetcher, { refreshInterval: 10000 });
  return <InstancesTable instances={data?.data ?? []} onRefresh={() => mutate()} />;
}
```

### Pattern 2: Dynamic Route with useParams (NEW for Instance Detail)
**What:** Next.js App Router dynamic segment `[id]` with client-side `useParams()`.
**When to use:** Instance detail page at `/cloud/instances/[id]`.
**Example:**
```typescript
// app/cloud/instances/[id]/page.tsx
"use client";
import { useParams } from "next/navigation";
import useSWR from "swr";

export default function InstanceDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data } = useSWR(`/api/v1/instances/${id}`, fetcher, { refreshInterval: 10000 });
  return <InstanceDetail instance={data} />;
}
```

### Pattern 3: SWR Mutation for Actions (Existing, Extend)
**What:** Trigger mutations (terminate, rename) then revalidate SWR cache.
**When to use:** Any destructive or state-changing action.
**Example:**
```typescript
async function handleTerminate(instanceId: string) {
  await terminateInstance(instanceId);
  // Revalidate the instances list
  mutate("/api/v1/instances");
}
```

### Pattern 4: Modal with Backdrop (Existing)
**What:** Fixed-position overlay with backdrop blur and centered modal card.
**When to use:** Launch instance form, confirmation dialogs, rename dialogs.
**Example:** See existing `LaunchInstanceForm.tsx` -- portal-style fixed overlay with `bg-bg/80 backdrop-blur-sm`.

### Pattern 5: Card Grid for GPU Availability (NEW)
**What:** CSS Grid of cards grouped by GPU model, each card showing specs and pricing.
**When to use:** GPU Availability page redesign.
**Key CSS:** `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4`

### Anti-Patterns to Avoid
- **Do NOT use Geist fonts or Vercel visual patterns** -- Phase 8 was reverted for this reason
- **Do NOT add component libraries** -- project uses custom Tailwind components throughout
- **Do NOT change the API proxy pattern** -- `next.config.ts` rewrites `/api/v1/*` to Go backend
- **Do NOT break the `gracefulFetcher` pattern** -- it enables dashboard rendering when backend is down
- **Do NOT use `window.confirm()`** -- replace with custom `ConfirmDialog` component per CONTEXT.md
- **Do NOT use `useRouter()` for navigation where `<Link>` suffices** -- Next.js prefetches `<Link>` destinations

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Clipboard copy | Custom clipboard logic | `navigator.clipboard.writeText()` wrapped in `CopyButton` (already exists) | Browser API handles permissions and fallbacks |
| Data fetching/caching | Custom fetch + useState | SWR with `fetcher`/`gracefulFetcher` (already exists) | SWR handles cache, dedup, revalidation, race conditions |
| Route protection | Custom auth middleware | `proxy.ts` with Clerk `auth.protect()` (already exists) | Clerk handles JWT validation, redirect flows |
| CSS class merging | Manual string concat | `cn()` utility (already exists) | `tailwind-merge` resolves conflicting utility classes |
| Form validation | Custom validation library | HTML5 `required`, `min`, `max` + manual checks | Forms are simple (2-3 fields each) |
| Time formatting | Moment.js/day.js | `Intl.DateTimeFormat` / `toLocaleDateString()` | No dependencies needed for date display |
| Skeleton loading | Loading library | Tailwind `animate-pulse` on `bg-bg-card-hover` divs (existing pattern) | Already established project convention |

## Common Pitfalls

### Pitfall 1: Instance Rename Requires Backend Work
**What goes wrong:** Frontend builds a rename UI but there's no `PATCH /api/v1/instances/{id}` endpoint.
**Why it happens:** The Go backend currently has no rename/update-name handler, no route, and no DB function.
**How to avoid:** Include a task that adds `PATCH /api/v1/instances/{id}` to the Go backend (route, handler, DB function) before the frontend rename feature.
**Warning signs:** 405 Method Not Allowed when trying to PATCH.

### Pitfall 2: GPU Card Grouping Logic
**What goes wrong:** The availability API returns flat offerings (one per GPU+region+tier combo), but cards need grouping by GPU model with spot/on-demand prices side by side.
**Why it happens:** The `AvailableOffering` type has one `price_per_hour` and one `tier` per entry. To show spot + on-demand prices in one card, the frontend must group/aggregate.
**How to avoid:** Group offerings client-side by `gpu_model` in a `useMemo`. For each group, pick the spot price and on-demand price from the respective entries. Handle cases where only one tier is available.
**Warning signs:** Cards showing only one price when both tiers exist.

### Pitfall 3: Breadcrumb for Nested Routes
**What goes wrong:** The topbar breadcrumb currently uses a flat `pathLabels` map and won't work for `/cloud/instances/[id]`.
**Why it happens:** Dynamic routes don't appear in a static map.
**How to avoid:** Parse the pathname segments. For known patterns like `/cloud/instances/{uuid}`, display "Instances / Instance Detail" or the instance name.
**Warning signs:** Breadcrumb showing "Dashboard" on the instance detail page.

### Pitfall 4: SWR Cache Key Consistency
**What goes wrong:** Mutating an instance (rename, terminate) on the detail page doesn't update the instances list, or vice versa.
**Why it happens:** SWR caches by key. `/api/v1/instances` and `/api/v1/instances/{id}` are different keys.
**How to avoid:** After mutations, call `mutate()` on both the specific instance key AND the list key. Use `useSWRConfig().mutate` to invalidate globally if needed.
**Warning signs:** Stale data after navigation between list and detail views.

### Pitfall 5: Price Confirmation Needs GPU Specs
**What goes wrong:** The launch modal needs to show VRAM, RAM, storage specs, but `CreateInstanceRequest` only sends `gpu_type`, not full specs.
**Why it happens:** The form currently uses free-text input for GPU type. When launched from a card, we have the full `AvailableOffering` data, but it needs to be passed through.
**How to avoid:** When launching from a GPU card, pass the full `AvailableOffering` object (or relevant fields) to the modal as props. Show price + specs in a confirmation summary before the submit button.
**Warning signs:** Modal shows price but not VRAM/RAM/storage.

### Pitfall 6: Mobile Sidebar
**What goes wrong:** Sidebar is `hidden md:flex` currently, completely invisible on mobile with no hamburger menu.
**Why it happens:** Phase 7 deferred mobile sidebar.
**How to avoid:** Add a hamburger button in the topbar that toggles a mobile sidebar overlay. Use React state in the cloud layout (needs to become a client component wrapper or use a shared context).
**Warning signs:** No navigation possible on mobile devices.

### Pitfall 7: Confirm Dialog Focus Trap
**What goes wrong:** Users can tab to elements behind the confirmation dialog, or press Enter on a hidden button.
**Why it happens:** Modals without focus management leak focus to the background.
**How to avoid:** Auto-focus the cancel/confirm button on mount. Trap focus within the dialog. Close on Escape key. Use `role="dialog"` and `aria-modal="true"`.
**Warning signs:** Keyboard users can interact with content behind the modal.

## Code Examples

### GPU Card Component Structure
```typescript
// Derived from AvailableOffering[] grouped by gpu_model
interface GPUCardData {
  gpu_model: string;
  vram_gb: number;
  cpu_cores: number;
  ram_gb: number;
  storage_gb: number;
  regions: string[];
  spot_price?: number;        // lowest spot price across regions
  on_demand_price?: number;   // lowest on-demand price across regions
  total_available: number;    // sum of available_count across all entries
  offerings: AvailableOffering[]; // raw data for launch
}

// Grouping logic in useMemo:
const grouped = useMemo(() => {
  const map = new Map<string, AvailableOffering[]>();
  for (const o of offerings) {
    const existing = map.get(o.gpu_model) ?? [];
    existing.push(o);
    map.set(o.gpu_model, existing);
  }
  return Array.from(map.entries()).map(([model, items]) => ({
    gpu_model: model,
    vram_gb: items[0].vram_gb,
    cpu_cores: items[0].cpu_cores,
    ram_gb: items[0].ram_gb,
    storage_gb: items[0].storage_gb,
    regions: [...new Set(items.map(i => i.region))],
    spot_price: Math.min(...items.filter(i => i.tier === 'spot').map(i => i.price_per_hour).concat([Infinity])),
    on_demand_price: Math.min(...items.filter(i => i.tier === 'on_demand').map(i => i.price_per_hour).concat([Infinity])),
    total_available: items.reduce((sum, i) => sum + i.available_count, 0),
    offerings: items,
  }));
}, [offerings]);
```

### Confirmation Dialog Pattern
```typescript
// Reusable ConfirmDialog component
interface ConfirmDialogProps {
  title: string;
  message: string;
  confirmLabel?: string;
  confirmVariant?: 'danger' | 'primary';
  onConfirm: () => void | Promise<void>;
  onCancel: () => void;
}

// Usage: Replace window.confirm() calls
<ConfirmDialog
  title="Terminate Instance"
  message="This will permanently destroy the instance and stop billing. This action cannot be undone."
  confirmLabel="Terminate"
  confirmVariant="danger"
  onConfirm={() => terminateInstance(id)}
  onCancel={() => setShowConfirm(false)}
/>
```

### Instance Rename Inline Edit
```typescript
// Inline editable name with pencil icon
function EditableName({ instance, onRename }: { instance: InstanceResponse; onRename: (name: string) => void }) {
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(instance.name ?? '');

  if (!editing) {
    return (
      <div className="flex items-center gap-2 group">
        <span className="type-ui-sm text-text font-medium">{instance.name ?? instance.id.slice(0, 12)}</span>
        <button
          onClick={() => setEditing(true)}
          className="opacity-0 group-hover:opacity-100 text-text-dim hover:text-text transition-all"
        >
          {/* pencil icon */}
        </button>
      </div>
    );
  }

  return (
    <input
      autoFocus
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onBlur={() => { onRename(value); setEditing(false); }}
      onKeyDown={(e) => {
        if (e.key === 'Enter') { onRename(value); setEditing(false); }
        if (e.key === 'Escape') setEditing(false);
      }}
      className="bg-bg border border-purple/50 rounded px-2 py-0.5 type-ui-sm text-text font-medium outline-none"
    />
  );
}
```

### Backend Rename Endpoint (Go)
```go
// Required new endpoint: PATCH /api/v1/instances/{id}
// Handler:
func (s *Server) handleUpdateInstance(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    var req struct {
        Name *string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    claims := auth.ClaimsFromContext(r.Context())
    // Update name in DB (org-scoped)
    if err := s.db.UpdateInstanceName(r.Context(), id, claims.OrgID, req.Name); err != nil {
        writeError(w, http.StatusInternalServerError, "failed to update instance")
        return
    }
    // Return updated instance
    inst, _ := s.db.GetInstance(r.Context(), id)
    writeJSON(w, http.StatusOK, s.instanceToResponse(inst))
}

// DB function:
func (p *Pool) UpdateInstanceName(ctx context.Context, instanceID, orgID string, name *string) error {
    _, err := p.pool.Exec(ctx,
        `UPDATE instances SET name = $1, updated_at = NOW() WHERE instance_id = $2 AND organization_id = (SELECT organization_id FROM organizations WHERE clerk_org_id = $3)`,
        name, instanceID, orgID)
    return err
}
```

### Sidebar Navigation Structure (Recommended)
```typescript
const navItems = [
  // Primary - daily use
  { label: "Instances",        href: "/cloud/instances",        icon: ServerIcon },
  { label: "GPU Availability", href: "/cloud/gpu-availability", icon: ChartIcon },

  // Management
  { label: "SSH Keys",         href: "/cloud/ssh-keys",         icon: KeyIcon },
  { label: "Billing",          href: "/cloud/billing",          icon: CreditCardIcon },

  // Coming soon
  { label: "API Keys",         href: "/cloud/api-keys",         icon: CodeIcon,   comingSoon: true },
  { label: "Team",             href: "/cloud/team",             icon: UsersIcon,  comingSoon: true },

  // Bottom
  { label: "Settings",         href: "/cloud/settings",         icon: GearIcon },
];
```

### Uptime Calculation
```typescript
// Calculate uptime from created_at / ready_at for running instances
function formatUptime(instance: InstanceResponse): string {
  if (instance.status === 'terminated') return '--';
  const start = instance.ready_at ?? instance.created_at;
  const elapsed = Date.now() - new Date(start).getTime();
  const hours = Math.floor(elapsed / 3600000);
  const minutes = Math.floor((elapsed % 3600000) / 60000);
  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  return `${hours}h ${minutes}m`;
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `window.confirm()` | Custom modal dialogs | Industry standard | Better UX, consistent styling, accessibility |
| Table for all data | Card grids for visual data, tables for scannable data | 2023+ GPU cloud UIs | Lambda Labs uses cards for GPU selection, tables for instances |
| Free-text GPU type input | Pre-filled from availability data | RunPod/Lambda pattern | Eliminates typos, shows confirmation before commit |
| Static page navigation | Sidebar with active states + breadcrumbs | Standard SaaS pattern | Clear wayfinding in multi-page dashboards |

**Design reference precedents:**
- Lambda Labs: Clean, spacious, developer-focused. Sidebar with Instances, Storage, SSH keys, API keys, Usage, Team, Settings. Minimal chrome.
- RunPod: Denser data views. Pods, Templates, Storage, Serverless in sidebar. Action buttons top-left. Recently unified Secure/Community Cloud into single view.

## Open Questions

1. **Instance Rename Backend Scope**
   - What we know: No PATCH endpoint, no DB function exists. CONTEXT.md explicitly requests rename.
   - What's unclear: Should the Go backend changes be part of this phase or a separate mini-phase?
   - Recommendation: Include as first plan in this phase -- it's a small addition (1 route, 1 handler, 1 DB function, ~50 lines Go) and the frontend depends on it.

2. **Home Page Redirect vs Overview**
   - What we know: CONTEXT.md leaves this at Claude's discretion.
   - What's unclear: Whether `/cloud` should redirect to `/cloud/instances` or show an overview dashboard.
   - Recommendation: Redirect `/cloud` to `/cloud/instances`. An overview page adds complexity without adding value when there are only a few pages. Lambda Labs does this -- clicking "Cloud" goes straight to instances.

3. **API Keys Coming-Soon Page**
   - What we know: Lambda Labs has API Keys in sidebar. CONTEXT.md says include "coming soon" items.
   - What's unclear: Exact "coming soon" items beyond what Lambda/RunPod show.
   - Recommendation: Add "API Keys" and "Team" as coming-soon items (matches Lambda Labs sidebar). Simple pages with a "Coming Soon" state and brief description.

## Sources

### Primary (HIGH confidence)
- Existing codebase (`frontend/src/`) -- complete component audit performed
- Go backend API handlers (`internal/api/handlers.go`, `internal/api/server.go`) -- route and response shape verified
- Design system (`globals.css`) -- all CSS variables, type classes, and animations documented
- Types and API layer (`lib/types.ts`, `lib/api.ts`) -- all interfaces and fetch functions verified

### Secondary (MEDIUM confidence)
- [RunPod UI Navigation Update](https://www.runpod.io/blog/runpod-ui-navigation-update) -- sidebar structure: Pods, Templates, Storage, Serverless
- [Lambda Cloud Dashboard Docs](https://3c5f6dba.lambda-docs.pages.dev/cloud/cloud-dashboard/) -- sidebar: Instances, Storage, SSH keys, API keys, Usage, Team, Settings
- [SWR Mutation & Revalidation Docs](https://swr.vercel.app/docs/mutation) -- useSWRMutation, optimistic updates, cache invalidation patterns
- [Next.js Dynamic Routes](https://nextjs.org/docs/app/api-reference/file-conventions/dynamic-routes) -- `[id]` segments, `useParams()` in client components

### Tertiary (LOW confidence)
- None -- all findings verified against codebase or official documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries already installed and in use; no new dependencies needed
- Architecture: HIGH -- extending existing patterns (server layout + client page, SWR fetching, Tailwind components)
- Pitfalls: HIGH -- identified through codebase audit (missing backend endpoint, grouping logic, breadcrumb for dynamic routes)

**Research date:** 2026-03-09
**Valid until:** 2026-04-09 (stable -- no fast-moving dependencies)
