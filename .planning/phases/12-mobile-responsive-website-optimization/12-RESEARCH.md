# Phase 12: Mobile Responsive Website Optimization - Research

**Researched:** 2026-03-25
**Domain:** Mobile-first responsive CSS/Tailwind, Next.js 16, frontend layout optimization
**Confidence:** HIGH

## Summary

Phase 12 focuses on making the existing GPU.ai frontend fully mobile-responsive across all pages. After a thorough audit of every component and page in the frontend codebase, the project is in a **partially responsive state**: the landing page Navbar has a mobile menu, the dashboard has a mobile sidebar with hamburger toggle, and the InstancesTable has explicit mobile card views. However, significant responsive gaps exist across both the marketing site and cloud dashboard.

The current frontend stack is Next.js 16.1.6 with Tailwind CSS v4, React 19, and the `motion` library for animations. The design system uses a dark theme with CSS custom properties for colors, spacing on an 8-point grid, a major-third typescale with responsive breakpoints at 768px for headings, and two custom local fonts (Vremena Grotesk for display, Necto Mono for body/UI). The existing responsive breakpoints used in the codebase are `sm:` (640px), `md:` (768px), `lg:` (1024px), and `xl:` (1280px).

**Primary recommendation:** Systematically fix mobile issues in three passes: (1) landing/marketing pages, (2) cloud dashboard pages, (3) touch target and viewport polish. No new libraries needed -- all fixes are CSS/Tailwind class adjustments to existing components.

## Standard Stack

### Core (Already Installed -- No New Dependencies)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| tailwindcss | ^4 | Utility-first responsive CSS | Already in use, mobile-first by default |
| next | 16.1.6 | Framework with built-in viewport meta | Already configured |
| react | 19.2.3 | UI library | Already in use |
| clsx + tailwind-merge | ^2.1.1 / ^3.5.0 | Conditional class composition | Already in use via `cn()` utility |
| motion | ^12.38.0 | Animations | Already in use, respects `prefers-reduced-motion` |

### Supporting (No New Dependencies Needed)
This phase requires zero new npm packages. All responsive work is CSS/Tailwind class modifications to existing components.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Tailwind responsive classes | CSS container queries | Tailwind already in use, container queries would add complexity for no gain here |
| Manual media queries | Tailwind responsive prefixes | Already using Tailwind; manual queries only for the few custom CSS classes in globals.css |

## Architecture Patterns

### Responsive Breakpoint Strategy (Already Established)
```
Mobile-first approach (Tailwind default):
- Base styles: < 640px (mobile phones)
- sm: >= 640px (large phones / small tablets)
- md: >= 768px (tablets / small laptops)
- lg: >= 1024px (laptops / desktops)
- xl: >= 1280px (large desktops)
```

The project already uses this pattern. The `md:` breakpoint is the primary desktop/mobile split point for the cloud dashboard (sidebar hidden/shown). The `lg:` breakpoint is used for the landing page Navbar desktop/mobile split.

### Pattern 1: Dual Layout (Desktop Table + Mobile Cards)
**What:** Render completely different layouts for desktop and mobile using `hidden md:block` and `md:hidden`.
**When to use:** Complex data tables that cannot be made responsive with simple column hiding.
**Already used in:** `InstancesTable.tsx` (DesktopTable + MobileCards), `PricingTable.tsx` (desktop table + mobile cards).
```tsx
// Source: existing codebase pattern
<div className="hidden md:block">{/* Desktop table layout */}</div>
<div className="md:hidden">{/* Mobile card layout */}</div>
```

### Pattern 2: Responsive Grid Columns
**What:** Reduce grid columns on smaller screens.
**When to use:** Card grids, info panels, stat blocks.
**Already used in:** GPUAvailabilityTable cards, InstanceDetail info grid.
```tsx
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
```

### Pattern 3: Stack-on-Mobile
**What:** Flex direction switches from row to column on mobile.
**When to use:** Side-by-side layouts that need to stack vertically.
```tsx
<div className="flex flex-col sm:flex-row gap-4">
```

### Anti-Patterns to Avoid
- **Fixed pixel widths without responsive alternatives:** Several components use `w-[180px]`, `w-[120px]` etc. in flex layouts -- these work because they are within a scrollable flex parent, but should be audited for narrow viewports.
- **Horizontal scrolling on mobile:** The `overflow-x-auto` wrapper on billing tables is correct, but the tables themselves lack any mobile-friendly alternative layout.
- **Touch targets under 44px:** Several small buttons (copy, edit icon) are under minimum touch target size.

## Audit: Responsive Issues by Component

### CRITICAL Issues (Broken or Unusable on Mobile)

#### 1. BillingDashboard.tsx -- Transaction History Table
**Problem:** The `TransactionHistorySection` renders a `<table>` with 5 columns (Date, Type, Description, Amount, Balance) wrapped in `overflow-x-auto`. On mobile, users must horizontally scroll a wide table. No mobile card alternative exists.
**Fix:** Add mobile card layout like InstancesTable pattern (hidden md:block for table, md:hidden for cards).

#### 2. BillingDashboard.tsx -- Usage Sessions Table
**Problem:** The Usage Sessions table has 7 columns (GPU, Count, Rate, Started, Ended, Duration, Cost). This is heavily unreadable on mobile with only horizontal scroll.
**Fix:** Add mobile card layout for sessions.

#### 3. Cloud Dashboard `p-6` Padding
**Problem:** The main content area in `cloud/layout.tsx` uses `p-6` (24px) padding on all sides. On mobile (320-375px viewport), this leaves very little content width.
**Fix:** Change to `p-4 md:p-6` for tighter mobile padding.

#### 4. Settings Page -- Form Layout
**Problem:** The spending limit form uses `flex gap-3 items-end` for the input + two buttons row. On narrow mobile, the three elements compress badly.
**Fix:** Stack form elements vertically on mobile with `flex flex-col sm:flex-row`.

### HIGH Issues (Poor but Functional on Mobile)

#### 5. Landing Page Hero -- PricingWidget
**Problem:** The PricingWidget has GPU tabs with 4 buttons in a row. On very narrow screens (<360px), tab text gets extremely cramped. The widget itself is fine at `max-w-[400px]` but the tabs need smaller text on mobile.
**Fix:** Reduce tab font size or use scrollable tabs on small screens.

#### 6. Landing Page TrustBar -- Dynamic Inset
**Problem:** TrustBar reads `.stripe-lines` element position to set `marginLeft`/`marginRight`. The stripe-lines are hidden on mobile (display:none at max-width 768px), so `getBoundingClientRect().left` returns 0, and the trust bar extends full-width. This actually works but the logos may be cut off on very small screens.
**Fix:** Ensure trust bar logos render well at small widths. Currently OK since it uses infinite scroll.

#### 7. GPUAvailabilityTable -- Filter Toolbar
**Problem:** The filter toolbar uses `flex flex-wrap gap-3`. On mobile, the category chips row (`overflow-x-auto flex-1 min-w-0`) sits between the search input and region/sort buttons. This wraps awkwardly on mid-width screens (400-640px).
**Fix:** Stack filter controls vertically on mobile: search full-width, category chips full-width scrollable, region + sort in a row.

#### 8. InstanceDetail -- Info Grid Cards
**Problem:** The info grid uses `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3`. This is fine. But the Connection card spans `md:col-span-1 lg:col-span-2` which means on tablet (md), the SSH command box can get tight.
**Fix:** Minor -- the SSH command already has `break-all` which handles overflow. Acceptable.

#### 9. About Page -- Team Member Grid Logos
**Problem:** Company logos in team member cards (NovacoreLogo with `-mr-12`, etc.) use negative margins and fixed sizes that could overflow on narrow mobile.
**Fix:** Audit logo sizes and remove negative margins on mobile.

#### 10. Free Trial Page -- Two-Column Layout
**Problem:** Uses `grid grid-cols-1 lg:grid-cols-[1fr_520px] lg:gap-24`. On mobile this stacks correctly, but the form card has `p-8 lg:p-10` and fixed width `520px` column. On mid-width screens (768-1024px), the single-column form is fine but padding could be tighter.
**Fix:** Reduce form padding on mobile: `p-5 sm:p-8 lg:p-10`.

### MEDIUM Issues (Cosmetic / Polish)

#### 11. Dashboard Topbar -- Breadcrumb Truncation
**Problem:** Breadcrumb text does not truncate on very narrow mobile screens. "Cloud > GPU Availability" could exceed available width when hamburger menu + balance pill + user avatar are all present.
**Fix:** Add `truncate` or `overflow-hidden` to breadcrumb container with max-width.

#### 12. Navbar Mobile Menu -- Company Dropdown
**Problem:** On mobile, the Company section shows "Company" as a static label with sub-links indented. This works but the indented links have `pl-3` which is subtle.
**Fix:** Minor visual polish only.

#### 13. Footer -- 5-Column Grid
**Problem:** Footer uses `grid gap-10 md:grid-cols-5`. On mobile, it stacks to single column which works. But 5 columns at `md:` (768px) is tight.
**Fix:** Change to `md:grid-cols-2 lg:grid-cols-5` for a 2-column intermediate layout.

#### 14. Pagination Controls on Mobile
**Problem:** The InstancesTable pagination row has per-page toggles (Show 10/25/50) and page navigation side by side. On narrow screens these can overflow.
**Fix:** Stack pagination vertically on mobile or hide per-page selector on small screens.

#### 15. SSHKeyManager -- Grid Layout
**Problem:** Uses `md:grid md:grid-cols-[1fr_1fr_auto_auto]`. On mobile, it falls back to `flex flex-col` which works correctly. Minor: the fingerprint code element could overflow on very narrow screens.
**Fix:** Add `truncate` or `break-all` to fingerprint display on mobile.

#### 16. LaunchInstanceForm Modal
**Problem:** The modal uses `max-w-lg mx-4`. On mobile this gives 16px margins on each side, which is fine. But on very small screens (320px), the inner padding of `p-6` (24px) leaves very little content width.
**Fix:** Reduce modal padding on mobile: `p-4 sm:p-6`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Responsive breakpoints | Custom media query system | Tailwind's built-in responsive prefixes | Already in use, consistent, well-tested |
| Mobile-friendly tables | Custom virtualized table | Dual layout pattern (table + cards) | Already proven in InstancesTable, simple to replicate |
| Touch-friendly buttons | Custom touch handlers | Minimum 44x44px tap targets via padding | CSS-only solution, follows Apple HIG |
| Mobile navigation | Custom drawer/sheet component | Existing DashboardSidebar mobile overlay | Already implemented with backdrop + slide-in |
| Viewport meta tag | Custom viewport handling | Next.js built-in viewport config | Handled automatically by the framework |

**Key insight:** Every responsive fix in this phase is a Tailwind class adjustment or HTML restructuring -- no JavaScript behavior changes, no new components, no new libraries.

## Common Pitfalls

### Pitfall 1: Breaking Desktop While Fixing Mobile
**What goes wrong:** Adding mobile-specific classes that inadvertently change desktop layout.
**Why it happens:** Tailwind mobile-first means base classes affect all sizes. Adding `flex-col` without the `sm:flex-row` counterpart breaks desktop.
**How to avoid:** Always test both mobile and desktop after each change. Use the pattern `flex flex-col sm:flex-row` (stack on mobile, row on desktop).
**Warning signs:** Desktop layout changes unexpectedly after mobile fixes.

### Pitfall 2: Forgetting Touch Targets
**What goes wrong:** Buttons and interactive elements too small for finger taps.
**Why it happens:** Desktop designs optimize for mouse precision (8x8px click targets work fine with a cursor).
**How to avoid:** Minimum 44x44px touch target area. Use padding to increase tap area without changing visual size: `p-2` on icon buttons gives adequate touch targets.
**Warning signs:** Users report difficulty tapping small buttons on mobile.

### Pitfall 3: Horizontal Overflow
**What goes wrong:** Content wider than viewport creates horizontal scrollbar on the page.
**Why it happens:** Fixed-width elements, long unbroken strings (SSH commands, IDs), wide tables.
**How to avoid:** Use `overflow-hidden` on body/main containers, `break-all` or `truncate` on long strings, `overflow-x-auto` on wide tables as last resort.
**Warning signs:** Page scrolls horizontally on mobile devices.

### Pitfall 4: Fixed Positioning on Mobile Safari
**What goes wrong:** `position: fixed` elements (modals, navbars) misbehave when iOS Safari shows/hides the URL bar.
**Why it happens:** iOS Safari's dynamic viewport height changes when scrolling.
**How to avoid:** Use `dvh` units for full-height fixed elements, or `min-h-screen` with fallback. The current `h-screen` on the cloud layout sidebar is fine because it uses flexbox.
**Warning signs:** Fixed elements jump or overlap content on iOS Safari scroll.

### Pitfall 5: z-index Stacking on Mobile Overlays
**What goes wrong:** Mobile sidebar overlay (z-50) conflicts with modals (z-50) or confirm dialogs.
**Why it happens:** Multiple fixed/absolute positioned layers competing.
**How to avoid:** The current z-index scheme works: sidebar overlay z-50, modals z-50, film grain z-9999. Ensure mobile sidebar is closed before opening modals.
**Warning signs:** Overlapping UI elements when sidebar and modal are both open.

## Code Examples

### Example 1: Converting a Table to Mobile Cards (Billing Transactions)
```tsx
// Source: pattern from existing InstancesTable.tsx, applied to BillingDashboard
// Desktop: keep existing <table>
<div className="hidden md:block">
  <div className="overflow-x-auto">
    <table className="w-full">...</table>
  </div>
</div>

// Mobile: card layout
<div className="md:hidden space-y-3">
  {transactions.map((tx) => (
    <div key={tx.id} className="bg-bg-card rounded-[10px] border border-border p-4 space-y-2">
      <div className="flex items-center justify-between">
        <span className="type-ui-xs text-text-muted">{formatDate(tx.created_at)}</span>
        <span className={cn("type-ui-2xs font-medium rounded-full px-2 py-0.5", info.color)}>
          {info.label}
        </span>
      </div>
      <div className="flex items-center justify-between">
        <span className="type-ui-sm text-text-muted">{tx.description || "--"}</span>
        <span className={cn("type-ui-sm font-mono", tx.amount_cents >= 0 ? "text-green" : "text-text")}>
          {tx.amount_cents >= 0 ? "+" : ""}${formatCents(Math.abs(tx.amount_cents))}
        </span>
      </div>
    </div>
  ))}
</div>
```

### Example 2: Responsive Padding
```tsx
// Source: Tailwind mobile-first pattern
// Before (too much padding on mobile):
<main className="flex-1 overflow-y-auto p-6">{children}</main>

// After (responsive padding):
<main className="flex-1 overflow-y-auto p-4 md:p-6">{children}</main>
```

### Example 3: Stacking Form Controls on Mobile
```tsx
// Source: Tailwind responsive flex pattern
// Before (cramped on mobile):
<form className="flex gap-3 items-end">...</form>

// After (stacks on mobile):
<form className="flex flex-col gap-3 sm:flex-row sm:items-end">...</form>
```

### Example 4: Footer Intermediate Grid
```tsx
// Source: responsive grid column pattern
// Before:
<div className="grid gap-10 md:grid-cols-5">

// After (2-col tablet, 5-col desktop):
<div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-5">
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Separate mobile sites (m.site.com) | Responsive design via CSS | 2010+ | Single codebase, all devices |
| Bootstrap grid system | Tailwind utility-first responsive | 2019+ | More granular control, smaller CSS |
| `vh` units for mobile | `dvh` (dynamic viewport height) | 2022+ | Fixes iOS Safari viewport issues |
| JavaScript resize listeners | CSS Container Queries | 2023+ | Component-level responsiveness (not needed here) |

**Deprecated/outdated:**
- Using `@media` queries manually when Tailwind responsive prefixes do the same thing
- Separate mobile/desktop components when responsive CSS suffices (only use dual-layout for tables)

## Open Questions

1. **Minimum supported viewport width**
   - What we know: Most mobile devices are 360px+ wide. iPhone SE is 375px. Very old devices are 320px.
   - What's unclear: Whether 320px support is required
   - Recommendation: Target 360px minimum, test at 375px (iPhone SE). Do not actively optimize for 320px.

2. **Tablet landscape orientation**
   - What we know: The `md:` breakpoint at 768px covers iPad portrait. iPad landscape is 1024px (hits `lg:`).
   - What's unclear: Whether any specific tablet optimization is needed
   - Recommendation: The existing breakpoint scheme handles tablets well. No special work needed.

3. **Mobile Safari URL bar interaction**
   - What we know: The cloud dashboard uses `h-screen` for the layout which can cause issues with iOS Safari's dynamic viewport
   - What's unclear: Whether this is causing actual problems
   - Recommendation: Test on real iOS device. If issues found, switch to `h-dvh` with `h-screen` fallback.

## Validation Architecture

> Note: `workflow.nyquist_validation` is not set in config.json, treating as enabled.

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Manual visual testing (no automated UI test framework detected) |
| Config file | none |
| Quick run command | `cd frontend && npm run build` (catches build errors) |
| Full suite command | `cd frontend && npm run build && npm run lint` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| MOBILE-01 | Landing page renders correctly at 375px width | manual | Visual testing in browser DevTools | N/A |
| MOBILE-02 | Dashboard sidebar opens/closes on mobile | manual | Visual testing | N/A |
| MOBILE-03 | Billing tables have mobile card layouts | manual | Visual testing | N/A |
| MOBILE-04 | All forms are usable on 375px viewport | manual | Visual testing | N/A |
| MOBILE-05 | No horizontal scroll on any page at 375px | manual | Visual testing in browser DevTools | N/A |
| MOBILE-06 | Touch targets >= 44px on interactive elements | manual | Visual inspection | N/A |

### Sampling Rate
- **Per task commit:** `cd frontend && npm run build` (verify no build breakage)
- **Per wave merge:** `cd frontend && npm run build && npm run lint`
- **Phase gate:** Full build passes, manual visual review of all pages at 375px width

### Wave 0 Gaps
None -- no test infrastructure to set up. This phase is purely CSS/layout changes verified by build success and manual visual testing.

## Sources

### Primary (HIGH confidence)
- Direct codebase audit: Read and analyzed every `.tsx` and `.css` file in the frontend
- Tailwind CSS v4 responsive design: built-in mobile-first breakpoint system (sm/md/lg/xl)
- Next.js 16 viewport handling: automatic viewport meta tag

### Secondary (MEDIUM confidence)
- Apple Human Interface Guidelines: 44pt minimum touch target size
- WCAG 2.1 SC 2.5.5: Target size minimum 44x44 CSS pixels

### Tertiary (LOW confidence)
- None -- all findings are from direct codebase analysis

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all dependencies already installed, no new packages
- Architecture: HIGH - patterns already established in codebase (dual layout, responsive grid)
- Pitfalls: HIGH - well-known CSS responsive design patterns
- Issue audit: HIGH - based on direct reading of every component file

**Research date:** 2026-03-25
**Valid until:** Indefinite (CSS responsive patterns are stable)
