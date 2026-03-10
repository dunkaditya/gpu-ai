# Phase 10: Frontend Dashboard with GPU Availability, Provisioning, and Instance Management - Context

**Gathered:** 2026-03-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Redesign and improve the cloud dashboard pages built in Phase 7. The landing page is NOT in scope (Phase 8 was reverted; landing page stays as-is). This phase delivers a polished, production-quality cloud dashboard for GPU availability browsing, instance provisioning, and instance management — with real API integration already wired from Phase 7.

</domain>

<decisions>
## Implementation Decisions

### Design Direction
- Clean & spacious layout (Lambda Labs style), not dense/data-rich (RunPod style)
- Generous whitespace, focused views, modern SaaS feel
- Exception: data tables (GPU availability, instances) can be denser where scanning matters
- Dark theme with existing design system (Vremena Grotesk + Necto Mono, purple accents, `#09090f` background)
- Phase 8 Vercel-style redesign was reverted — do NOT use Geist fonts or Vercel visual patterns
- Reference RunPod and Lambda Labs dashboards for layout/UX patterns

### Navigation & Layout
- Sidebar nav structure at Claude's discretion — reference RunPod and Lambda Labs for what's important
- Include "coming soon" items for future essential features (blank pages with coming-soon state)
- Minimal topbar: breadcrumb showing current page + Clerk user button (avatar, sign out) on the right
- Home page behavior at Claude's discretion (redirect to instances or overview dashboard)

### Provisioning Flow
- Launch from GPU availability table: click a GPU row → pre-filled launch modal with that GPU/region selected
- Launch modal shows price confirmation: estimated hourly cost, GPU specs (VRAM, RAM, storage), and tier before "Launch" button
- Auto-attach all user's SSH keys to new instances (no key picker in modal)
- After launch: close modal, redirect to instances list. New instance appears with status badge. SWR auto-refreshes every 10s

### Instance Management
- Enhanced table layout: name (editable/renamable), status badge, GPU type + count, region, hourly price, uptime, SSH command quick-copy, terminate action
- Rows clickable → instance detail view at /cloud/instances/{id} with full specs, uptime, billing, SSH command, events
- Copy SSH command: one-click copy of `ssh -p PORT root@host`
- Terminate with confirmation dialog (prevent accidental deletion)
- Rename instances (custom names instead of just GPU type)
- Empty state: "No instances running" message with CTA button to browse GPU Availability page

### GPU Availability Display
- Card grid grouped by GPU model (H100, A100, RTX 4090, etc.)
- Each card shows: GPU name, VRAM, available count, spot price, on-demand price side by side, region tags, launch button
- Filter bar with dropdowns: region, tier
- Sortable by price (ascending/descending)
- Clicking launch on a card opens the pre-filled launch modal

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

</decisions>

<specifics>
## Specific Ideas

- Reference RunPod dashboard and Lambda Labs dashboard for UX patterns and navigation structure
- Lambda Labs approach for overall feel: clean, spacious, developer-focused
- RunPod approach for data density: where information needs to be scannable
- GPU availability as card grid (not table) — each GPU model is a card with specs and pricing
- Provisioning flow inspired by Lambda Labs: select from available GPUs → confirm → launch
- Instance list inspired by both: table with quick actions (SSH copy, terminate) and clickable rows for detail

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `InstancesTable` component (`src/components/cloud/InstancesTable.tsx`): desktop table + mobile cards, needs enhancement
- `GPUAvailabilityTable` component (`src/components/cloud/GPUAvailabilityTable.tsx`): needs redesign from table to card grid
- `LaunchInstanceForm` modal (`src/components/cloud/LaunchInstanceForm.tsx`): needs price confirmation and pre-fill support (defaultGPU/defaultRegion props exist)
- `StatusBadge` component (`src/components/cloud/StatusBadge.tsx`): reusable as-is
- `DashboardSidebar` (`src/components/cloud/DashboardSidebar.tsx`): needs nav item updates
- `DashboardTopbar` (`src/components/cloud/DashboardTopbar.tsx`): needs Clerk user button integration
- `SSHKeyManager` (`src/components/cloud/SSHKeyManager.tsx`): exists, polish as needed
- `BillingDashboard` (`src/components/cloud/BillingDashboard.tsx`): exists, polish as needed
- `api.ts` with all API functions: fetcher, gracefulFetcher, createInstance, terminateInstance, fetchGPUAvailability, addSSHKey, deleteSSHKey, fetchBillingUsage
- `types.ts` with all TypeScript interfaces matching Go backend types

### Established Patterns
- SWR for data fetching with tuned refresh intervals (10s instances, 30s availability, 60s billing)
- `gracefulFetcher` returns null on 401/403 for empty states when backend unavailable
- Design system CSS variables in globals.css (colors, spacing, typescale)
- Server layout + client page pattern for metadata on SWR-powered routes
- Skeleton loading with `bg-card-hover` pulse animations

### Integration Points
- Clerk auth already wired: `ClerkProvider` in root layout, sign-in/sign-up pages exist
- Multi-domain routing via `proxy.ts`: `?site=cloud` for local dev
- API proxy at `/api/v1/*` routes to Go backend
- Cloud route group at `src/app/cloud/` with layout, instances, gpu-availability, ssh-keys, billing, settings pages

</code_context>

<deferred>
## Deferred Ideas

- Interactive pricing widget on landing page — separate concern from dashboard
- GPU templates gallery (one-click deploy environments) — future phase
- Real-time instance metrics/monitoring dashboard — requires agent on instances
- Team/org management UI — ADV-02 is a v2 requirement

</deferred>

---

*Phase: 10-frontend-dashboard-with-gpu-availability-provisioning-and-instance-management*
*Context gathered: 2026-03-09*
