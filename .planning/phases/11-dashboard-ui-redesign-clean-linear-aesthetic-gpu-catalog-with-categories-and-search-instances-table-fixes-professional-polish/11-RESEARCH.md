# Phase 11: Dashboard UI Redesign - Research

**Researched:** 2026-03-11
**Domain:** Frontend UI/UX - Next.js 16 + Tailwind CSS 4 + React 19 dashboard redesign
**Confidence:** HIGH

## Summary

Phase 11 transforms the existing cloud dashboard from a functional-but-generic interface into a polished, Linear-inspired product with professional visual design. The phase name specifies three distinct areas of work: (1) clean linear aesthetic applied across all dashboard pages, (2) GPU catalog redesign with categories and search, and (3) instances table fixes plus overall professional polish.

The existing codebase is well-structured with clean component boundaries (DashboardSidebar, GPUAvailabilityTable, GPUCard, InstancesTable, InstanceDetail, LaunchInstanceForm, SSHKeyManager, BillingDashboard, ConfirmDialog, StatusBadge) all using Tailwind CSS 4 with a custom design system in globals.css. The current design system uses Vremena Grotesk (sans) and Necto Mono (mono) fonts, a dark purple-accented theme, and CSS custom properties for all colors and spacing. The technology stack (Next.js 16, React 19, SWR, Tailwind CSS 4, Clerk) is stable and requires no changes -- this is purely a frontend visual and UX redesign.

The GPU availability API returns offerings with `gpu_model` names that map to well-defined categories (Blackwell, Hopper, Ampere, Ada Lovelace, Professional/Workstation, Consumer/Gaming, Legacy). The backend already supports filtering by `gpu_model`, `region`, and `tier` via query params. The frontend currently groups offerings by `gpu_model` into cards but lacks text search, category filtering, and proper catalog organization.

**Primary recommendation:** Apply the "Linear aesthetic" (restrained color palette, precise spacing, subtle borders, monochrome with minimal accent color) across all dashboard components. Restructure GPU availability as a searchable catalog with GPU family categories. Fix instances table column alignment, overflow, and responsiveness issues. Polish all pages for visual consistency and professional feel.

## Standard Stack

### Core (Already Installed -- No Changes)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Next.js | 16.1.6 | App Router framework | Already in use |
| React | 19.2.3 | UI library | Already in use |
| Tailwind CSS | ^4 | Utility-first styling | Already in use |
| SWR | ^2.4.1 | Data fetching + caching | Already in use |
| @clerk/nextjs | ^6.39.0 | Authentication | Already in use |
| clsx + tailwind-merge | ^2.1.1 / ^3.5.0 | Class name utilities | Already in use via cn() |

### Supporting (May Need to Install)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| use-debounce | ^10.0.4 | Search input debouncing | GPU catalog search field |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| use-debounce | Custom setTimeout debounce | use-debounce is 1KB, well-tested, provides useDebouncedCallback hook directly |
| lucide-react icons | Keep inline SVGs | Inline SVGs are already the pattern; switching icons would be a large diff for no functional gain |

**Installation (if search debounce added):**
```bash
cd frontend && npm install use-debounce
```

## Architecture Patterns

### Current Project Structure (frontend/src)
```
app/
  cloud/
    layout.tsx              -- Dashboard shell (sidebar + topbar + main)
    page.tsx                -- Redirects to /cloud/instances
    gpu-availability/       -- GPU catalog page
    instances/              -- Instances list page
    instances/[id]/         -- Instance detail page
    ssh-keys/               -- SSH key management
    billing/                -- Billing dashboard
    settings/               -- Spending limit settings
    api-keys/               -- Coming soon placeholder
    team/                   -- Coming soon placeholder
components/
  cloud/                    -- Dashboard components (10 components)
  landing/                  -- Marketing page components
  ui/                       -- Shared primitives (Button, Card, etc.)
lib/
  api.ts                    -- API client functions
  types.ts                  -- TypeScript interfaces matching Go backend
  utils.ts                  -- cn() utility
  constants.ts              -- Landing page constants
  mock-data.ts              -- Mock instance data
```

### Pattern 1: GPU Category Classification (Client-Side)
**What:** Map GPU model names to categories (families) for catalog navigation
**When to use:** GPU availability page for tab/chip filtering by GPU family
**Example:**
```typescript
// GPU family categories derived from provider/types.go GPUType enum
const GPU_CATEGORIES = {
  "Blackwell": ["b200", "b300", "rtx_pro_6000", "rtx_pro_4500", "rtx_5090", "rtx_5080"],
  "Hopper": ["h200_sxm", "h200_nvl", "h100_sxm", "h100_nvl", "h100_pcie"],
  "Ada Lovelace": ["l40s", "l40", "l4", "rtx_6000_ada", "rtx_5000_ada", "rtx_4000_ada", "rtx_2000_ada", "rtx_4090", "rtx_4080"],
  "Ampere": ["a100_80gb", "a100_40gb", "a40", "a30", "rtx_a6000", "rtx_a5000", "rtx_a4500", "rtx_a4000", "rtx_3090", "rtx_3080"],
  "Legacy": ["v100"],
} as const;

// Utility to classify a gpu_model string into a category
function getGPUCategory(gpuModel: string): string {
  const normalized = gpuModel.toLowerCase().replace(/[\s-]/g, '_');
  for (const [category, models] of Object.entries(GPU_CATEGORIES)) {
    if (models.some(m => normalized.includes(m))) return category;
  }
  return "Other";
}
```

### Pattern 2: Debounced Search with SWR
**What:** Text search that filters GPU offerings client-side with debounced input
**When to use:** GPU catalog search bar
**Example:**
```typescript
import { useDebouncedCallback } from 'use-debounce';

const [searchQuery, setSearchQuery] = useState("");
const [debouncedQuery, setDebouncedQuery] = useState("");

const handleSearch = useDebouncedCallback((value: string) => {
  setDebouncedQuery(value);
}, 200);

// Filter offerings in useMemo using debouncedQuery
const filtered = useMemo(() => {
  if (!debouncedQuery) return offerings;
  const q = debouncedQuery.toLowerCase();
  return offerings.filter(o =>
    o.gpu_model.toLowerCase().includes(q) ||
    o.region.toLowerCase().includes(q)
  );
}, [offerings, debouncedQuery]);
```

### Pattern 3: Linear Aesthetic CSS Principles
**What:** Restrained, high-contrast dark theme with minimal color accents
**When to use:** Across all dashboard components
**Key principles from linear.style research:**
- Extremely sparse color usage -- near-monochrome with one accent color
- Very subtle borders (1px, very low contrast difference from background)
- Precise, generous spacing (8px grid)
- Typography hierarchy through weight and opacity, not size variety
- Subtle hover states (background color shift, no dramatic transformations)
- No gradients on interactive elements (flat buttons with subtle hover)
- Film grain and glow effects should be toned down or removed for the dashboard

### Pattern 4: Category Tab Navigation
**What:** Horizontal chip/tab bar for GPU family filtering
**When to use:** GPU availability page above the card grid
**Example:**
```typescript
const categories = ["All", "Blackwell", "Hopper", "Ada Lovelace", "Ampere", "Legacy"];

// Horizontal scrollable chip bar
<div className="flex gap-2 overflow-x-auto pb-2">
  {categories.map(cat => (
    <button
      key={cat}
      onClick={() => setActiveCategory(cat)}
      className={cn(
        "px-3 py-1.5 rounded-md type-ui-xs font-medium whitespace-nowrap transition-colors",
        activeCategory === cat
          ? "bg-bg-card-hover text-text"
          : "text-text-dim hover:text-text-muted"
      )}
    >
      {cat}
    </button>
  ))}
</div>
```

### Anti-Patterns to Avoid
- **Gradient buttons everywhere:** The current `gradient-btn` class is overused. Linear aesthetic uses flat, subtle buttons. Reserve gradient for one primary CTA per page.
- **Decorative film grain on dashboard:** The `body::after` grain overlay is appropriate for the marketing landing page but adds visual noise to a functional dashboard. It should be scoped to the marketing layout only.
- **Over-styled select dropdowns:** The current `<select>` elements break visual consistency. Use custom dropdown components or styled alternatives.
- **Inconsistent card border radius:** Some cards use `rounded-xl` (12px), others `rounded-lg` (8px). Standardize on one size.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Search debounce | Custom setTimeout/clearTimeout | `use-debounce` | Edge cases with component unmount, race conditions |
| Class name merging | String concatenation | `cn()` (already using clsx + tailwind-merge) | Already established pattern |
| Data fetching | Raw fetch in useEffect | SWR (already using) | Already established pattern with 10s/30s/60s refresh intervals |
| GPU category data | Backend API changes | Client-side classification map | Backend GPU types are already well-structured; adding a category field to the API is unnecessary overhead for a UI concern |

**Key insight:** This phase is purely frontend -- no backend changes needed. All filtering, categorization, and search happen client-side using the existing `/api/v1/gpu/available` endpoint which already returns all offerings. The backend already supports `gpu_model`, `region`, and `tier` query params if server-side filtering is ever needed, but client-side filtering of the cached SWR data is simpler and avoids API changes.

## Common Pitfalls

### Pitfall 1: Instances Table Column Alignment with Full-Row Links
**What goes wrong:** The current InstancesTable uses `<tr>` with `<td colSpan={8}>` wrapping a single `<Link>`, then uses `<div>` children with fixed widths for columns. This breaks native table column alignment and causes SSH command truncation.
**Why it happens:** Making entire table rows clickable required wrapping content in a single Link, but HTML tables expect matching `<td>` elements per `<th>`.
**How to avoid:** Either switch to a CSS grid layout (div-based table) where the entire row is a single clickable element, or use proper `<td>` elements with onClick handler and `router.push()` instead of wrapping in Link.
**Warning signs:** Columns don't align with headers, SSH commands get truncated to 260px max-width.

### Pitfall 2: Film Grain Overlay Blocking Interactions on Dashboard
**What goes wrong:** The `body::after` pseudo-element with `z-index: 9999` and `pointer-events: none` covers the entire viewport. While `pointer-events: none` prevents click blocking, it can interfere with text selection and creates unnecessary rendering overhead on every page.
**Why it happens:** The grain effect was designed for the marketing landing page and applied globally.
**How to avoid:** Scope the grain effect to the marketing layout only, or conditionally disable it in the cloud dashboard layout.
**Warning signs:** Subtle text selection issues, unnecessary GPU compositing on dashboard pages.

### Pitfall 3: Search Input Causing Excessive Re-renders
**What goes wrong:** Direct `onChange` on search input triggers immediate state updates and re-renders of the entire GPU grid on every keystroke.
**Why it happens:** React re-renders the component tree on every state change.
**How to avoid:** Use debounced callback (200ms) for the search filter value. Keep the input value responsive but debounce the filter computation.
**Warning signs:** Laggy search on lower-end devices, jank during typing.

### Pitfall 4: Inconsistent Empty States Across Pages
**What goes wrong:** Each component has a slightly different empty state design, making the dashboard feel inconsistent.
**Why it happens:** Components were built incrementally across multiple phases with local design decisions.
**How to avoid:** Create a shared EmptyState component with consistent icon, title, subtitle, and action button pattern.
**Warning signs:** Different padding, icon sizes, and text styles across empty states.

### Pitfall 5: Dark Theme Color Contrast
**What goes wrong:** When refining the color palette toward more muted Linear-style tones, text-dim and border colors can become too similar to the background, failing WCAG contrast requirements.
**Why it happens:** Linear aesthetic prioritizes subtlety, but GPU.ai's `#09090f` background is already very dark, leaving little room for subtle border distinctions.
**How to avoid:** Test contrast ratios. Keep `text-muted` at minimum 4.5:1 contrast ratio against `bg`. Keep `text-dim` at minimum 3:1 (for non-essential labels).
**Warning signs:** Hard to read labels, invisible borders on certain monitors.

## Code Examples

### Example 1: Unified GPU Category Constants
```typescript
// lib/gpu-categories.ts
export interface GPUCategoryDef {
  label: string;
  description: string;
  models: string[]; // lowercase model identifiers to match against gpu_model
}

export const GPU_CATEGORIES: GPUCategoryDef[] = [
  {
    label: "Blackwell",
    description: "Latest generation",
    models: ["b200", "b300", "rtx pro 6000", "rtx pro 4500", "rtx 5090", "rtx 5080"],
  },
  {
    label: "Hopper",
    description: "Data center H-series",
    models: ["h200", "h100"],
  },
  {
    label: "Ada Lovelace",
    description: "L-series & RTX 40-series",
    models: ["l40s", "l40", "l4", "rtx 6000 ada", "rtx 5000 ada", "rtx 4000 ada", "rtx 2000 ada", "rtx 4090", "rtx 4080"],
  },
  {
    label: "Ampere",
    description: "A-series & RTX 30-series",
    models: ["a100", "a40", "a30", "a10", "rtx a6000", "rtx a5000", "rtx a4500", "rtx a4000", "rtx 3090", "rtx 3080"],
  },
  {
    label: "Legacy",
    description: "Previous generation",
    models: ["v100"],
  },
];

export function classifyGPU(gpuModel: string): string {
  const lower = gpuModel.toLowerCase();
  for (const cat of GPU_CATEGORIES) {
    if (cat.models.some(m => lower.includes(m))) {
      return cat.label;
    }
  }
  return "Other";
}
```

### Example 2: Search + Category Filter Component Structure
```typescript
// Composable filter state for GPU catalog
const [searchQuery, setSearchQuery] = useState("");
const [activeCategory, setActiveCategory] = useState("All");
const [regionFilter, setRegionFilter] = useState("");
const [sortDir, setSortDir] = useState<"asc" | "desc">("asc");

// Debounced search
const debouncedSearch = useDebouncedCallback((val: string) => {
  setDebouncedQuery(val);
}, 200);

// Chained filters: category -> search -> region -> sort
const filteredCards = useMemo(() => {
  let result = offerings.filter(o => o.tier === "on_demand");

  // Category filter
  if (activeCategory !== "All") {
    result = result.filter(o => classifyGPU(o.gpu_model) === activeCategory);
  }

  // Text search
  if (debouncedQuery) {
    const q = debouncedQuery.toLowerCase();
    result = result.filter(o =>
      o.gpu_model.toLowerCase().includes(q) ||
      o.region.toLowerCase().includes(q)
    );
  }

  // Region filter
  if (regionFilter) {
    result = result.filter(o => o.region === regionFilter);
  }

  return result;
}, [offerings, activeCategory, debouncedQuery, regionFilter]);
```

### Example 3: Fixed Instances Table with CSS Grid
```typescript
// Using CSS grid instead of <table> for proper clickable rows
<div className="grid" style={{ gridTemplateColumns: "180px 120px 100px 100px 100px 80px 1fr auto" }}>
  {/* Header row */}
  <div className="contents">
    <div className="px-4 py-3 type-ui-2xs text-text-dim uppercase">Name</div>
    <div className="px-4 py-3 type-ui-2xs text-text-dim uppercase">GPU</div>
    {/* ... more headers ... */}
  </div>

  {/* Data rows */}
  {instances.map(instance => (
    <Link
      key={instance.id}
      href={`/cloud/instances/${instance.id}`}
      className="contents group"
    >
      <div className="px-4 py-3 group-hover:bg-bg-card transition-colors">
        {/* Name cell content */}
      </div>
      {/* ... more cells ... */}
    </Link>
  ))}
</div>
```

### Example 4: Shared EmptyState Component
```typescript
interface EmptyStateProps {
  icon: React.ReactNode;
  title: string;
  description?: string;
  action?: { label: string; href?: string; onClick?: () => void };
}

function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-10 h-10 rounded-lg bg-bg-card-hover flex items-center justify-center mb-3 text-text-dim">
        {icon}
      </div>
      <p className="type-ui-sm text-text-muted">{title}</p>
      {description && <p className="type-ui-2xs text-text-dim mt-1">{description}</p>}
      {action && (
        action.href
          ? <Link href={action.href} className="mt-4 type-ui-xs text-text-muted hover:text-text transition-colors">
              {action.label}
            </Link>
          : <button onClick={action.onClick} className="mt-4 type-ui-xs text-text-muted hover:text-text transition-colors">
              {action.label}
            </button>
      )}
    </div>
  );
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Gradient buttons everywhere | Flat buttons with subtle hover (Linear style) | 2024-2025 | Reduced visual noise, more professional |
| Bold purple accent on all interactive elements | Monochrome with accent reserved for primary CTA | 2025 | Cleaner, less fatiguing UI |
| Card-heavy layouts | Clean list/table layouts with card details | 2025 | Better information density |
| Complex filter panels | Inline chip bars + search | 2025 | Faster filtering, less UI overhead |
| Custom scrollbars | Native scrollbars or none | 2025 | Better cross-platform consistency |

**Deprecated/outdated:**
- Heavy glassmorphism effects: Trend is moving toward flat, precise surfaces
- Animated gradient borders: Appropriate for marketing pages, not functional dashboards
- Film grain overlays on app surfaces: Marketing effect, not app chrome

## Phase Requirements

Since no explicit requirement IDs were provided, these are derived from the phase name and description:

| ID | Description | Research Support |
|----|-------------|-----------------|
| UI-01 | Clean linear aesthetic applied to all dashboard pages | Linear design research (monochrome, precise spacing, subtle borders), existing globals.css design tokens |
| UI-02 | GPU catalog page with category tabs for GPU families | GPU_CATEGORIES classification from provider/types.go, category chip bar pattern |
| UI-03 | GPU catalog search functionality | Debounced search pattern with use-debounce, client-side filtering of SWR-cached data |
| UI-04 | Instances table alignment and overflow fixes | CSS grid layout pattern replacing colSpan hack, proper SSH command display |
| UI-05 | Professional visual polish across all pages | Shared EmptyState component, consistent border-radius, button styling, spacing normalization |
| UI-06 | Dashboard-specific chrome (sidebar, topbar) refinement | Linear aesthetic navigation patterns, active state indicators |

## Open Questions

1. **GPU category granularity**
   - What we know: GPU types map to 5 architecture families (Blackwell, Hopper, Ada Lovelace, Ampere, Legacy)
   - What's unclear: Should categories also separate Data Center vs Consumer/Gaming GPUs within each family? (e.g., H100 SXM is data center, RTX 4090 is consumer)
   - Recommendation: Start with architecture-based categories. If the catalog grows large enough to warrant sub-categories, add "Data Center" and "Consumer" as secondary filters later.

2. **Film grain scope**
   - What we know: The grain overlay is applied globally via `body::after` in globals.css
   - What's unclear: Should the grain be completely removed from dashboard, or kept with reduced opacity?
   - Recommendation: Remove it from the dashboard entirely (scope to marketing layout). The Linear aesthetic is about precision, and grain adds unwanted texture.

3. **Button style migration**
   - What we know: Current `gradient-btn` class is used in 8+ locations across the dashboard
   - What's unclear: How aggressively to flatten button styles -- keep gradient for primary CTAs only, or go fully flat?
   - Recommendation: Use flat buttons as default, reserve gradient (or solid accent) for one primary action per page (e.g., "Launch Instance" button).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | None (no frontend tests configured) |
| Config file | None |
| Quick run command | `cd frontend && npm run build` |
| Full suite command | `cd frontend && npm run build && npm run lint` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UI-01 | Linear aesthetic CSS applied | manual-only | Visual inspection | N/A |
| UI-02 | GPU category tabs filter correctly | manual-only | Visual inspection + interaction test | N/A |
| UI-03 | Search filters GPU offerings | manual-only | Type in search, verify filtered results | N/A |
| UI-04 | Instances table columns align | manual-only | Visual inspection | N/A |
| UI-05 | Consistent polish across pages | manual-only | Visual walkthrough of all dashboard pages | N/A |
| UI-06 | Sidebar/topbar refined | manual-only | Visual inspection | N/A |

### Sampling Rate
- **Per task commit:** `cd frontend && npm run build` (ensures no TypeScript/build errors)
- **Per wave merge:** `cd frontend && npm run build && npm run lint`
- **Phase gate:** Build succeeds + visual inspection of all dashboard pages

### Wave 0 Gaps
None -- this is a pure visual/UX redesign. The existing build + lint infrastructure is sufficient. No new test framework needed since all validation is visual.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: All 10 cloud dashboard components, globals.css design system, lib/types.ts, lib/api.ts, provider/types.go GPU type definitions
- Codebase analysis: internal/provider/runpod/mapping.go for complete GPU model mapping
- Codebase analysis: internal/availability/types.go for API response structure

### Secondary (MEDIUM confidence)
- [Linear Style](https://linear.style/) -- Linear's official design system reference (dark theme colors, typography, component patterns)
- [Linear Design blog](https://linear.app/now/how-we-redesigned-the-linear-ui) -- Linear's design philosophy and evolution
- [LogRocket: Linear Design Trend](https://blog.logrocket.com/ux-design/linear-design/) -- Analysis of the Linear aesthetic as a SaaS design trend
- [Next.js Dashboard Search Tutorial](https://nextjs.org/learn/dashboard-app/adding-search-and-pagination) -- Official Next.js pattern for search with debounce

### Tertiary (LOW confidence)
- WebSearch results on dashboard design trends for 2025-2026 (multiple sources, general patterns)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- No stack changes, pure visual redesign of existing components
- Architecture: HIGH -- Component structure stays the same, patterns are well-established React/Next.js
- Pitfalls: HIGH -- Identified from direct codebase analysis of existing bugs (colSpan alignment, grain overlay scope)
- GPU categories: MEDIUM -- Derived from Go backend types, but exact UX mapping needs user validation

**Research date:** 2026-03-11
**Valid until:** 2026-04-11 (stable -- no dependency changes, pure UI work)
