# Phase 8: Rebuild Frontend Landing Page to Match Vercel Homepage Design - Research

**Researched:** 2026-02-26
**Domain:** Frontend landing page rebuild (Next.js + Tailwind CSS, Vercel-style design)
**Confidence:** HIGH

## Summary

This phase rebuilds the GPU.ai landing page to mirror Vercel's homepage layout and visual style. The existing frontend is a Next.js 16 + Tailwind CSS v4 + React 19 app with a broken build (missing `@/lib/utils` and `@/lib/constants` modules). The current page has 14 landing components, custom fonts (Vremena Grotesk + Necto Mono), and a purple/green dark theme with particle canvas effects.

The target design replaces this with a Vercel-style structure: clean grid background with cross markers, rainbow glow gradient behind a triangle graphic, horizontal use-case tabs, feature pillar sections, CLI code demo, and multi-column footer. The user explicitly decided to "pull Vercel's frontend source code and adapt it directly" -- this means studying vercel.com/home's structure and recreating it with GPU.ai content, not literally copying proprietary code.

The existing stack (Next.js 16, React 19, Tailwind v4) is correct and does not need changing. The primary work is: (1) fix the broken build by creating missing lib files, (2) replace the custom fonts with Geist (Vercel's typeface), (3) rewrite globals.css for the new design system, (4) delete removed components (ComputeField, EffectsToggle, PricingWidget, PricingTable, TrustBar), (5) build 7 new sections matching Vercel's layout, (6) adapt content for GPU.ai.

**Primary recommendation:** Keep the existing Next.js 16 + Tailwind v4 stack. Replace fonts with Geist Sans/Mono via the `geist` npm package. Rewrite the page from scratch using the existing project structure, building each section (Navbar, Hero, UseCaseTabs, FeaturePillars, CLIDemo, FinalCTA, Footer) as individual components.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Use Vercel's exact layout structure with GPU.ai branding
- Content and graphics adapted to what GPU.ai actually does -- not a content clone
- Reference: vercel.com/home (current as of Feb 2026)
- Match Vercel's rainbow color theme
- Logo: simple "GPU.ai" text wordmark in clean typography, Vercel-style
- Pull Vercel's frontend source code and adapt it directly rather than rebuilding from scratch
- Section Structure (top to bottom):
  1. Navbar -- GPU.ai text wordmark + navigation links (match Vercel's nav structure)
  2. Hero -- Centered headline + subtitle + two CTAs + static metrics + Vercel triangle graphic with rainbow glow
  3. Use Case Tabs -- ML Training, Inference, Fine-tuning, Rendering, Research
  4. Feature Pillars -- "Source, Deploy, Scale" sections
  5. CLI Demo -- gpuctl command examples showing instance creation, status check, SSH
  6. Final CTA -- Action buttons
  7. Footer -- Multi-column links
- Removed from current page: PricingWidget, PricingTable, ComputeField, EffectsToggle, TrustBar, Templates section
- Hero headline: Vercel-style single concise line
- Hero CTAs: "Launch GPU Instance" (primary) + "Talk to our team" (secondary)
- Tone: technical-polished, matching Vercel's style
- Hero stats: static metrics (placeholder/aspirational numbers fine, no rotation needed)
- Match Vercel's grid background exactly -- grid lines with cross markers at intersections, dark cells, same spacing
- Match Vercel's rainbow glow exactly -- blue-green-red radial gradient behind the triangle graphic
- No scroll-triggered animations
- No ComputeField particles or EffectsToggle -- clean, Vercel-style static layout with hover states
- Dark theme throughout

### Claude's Discretion
- Exact hero headline copy
- Subtitle text
- Metrics/stats numbers and labels
- Specific content within each feature pillar section
- Nav link labels and structure
- Footer column organization
- CLI demo command examples
- Hover state implementations

### Deferred Ideas (OUT OF SCOPE)
- Interactive pricing widget -- separate phase or add back later
- GPU templates gallery (one-click deploy environments) -- future phase
- Customer logos / social proof -- add when real customers exist
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DASH-01 | Landing page describes the product | Full page rebuild with 7 sections describing GPU.ai's value proposition, use cases, features, and CLI interface |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| next | 16.1.6 | React meta-framework (SSR, routing, optimization) | Already installed; Vercel's own framework |
| react | 19.2.3 | UI component library | Already installed; current stable |
| react-dom | 19.2.3 | React DOM renderer | Already installed; matches React version |
| tailwindcss | ^4 | Utility-first CSS | Already installed; CSS-first config in v4 |
| geist | latest | Vercel's Geist Sans + Mono typeface | Vercel's official font; matches target design exactly |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| clsx | ^2.1.1 | Conditional className joining | Already installed; use in all components |
| tailwind-merge | ^3.5.0 | Merge Tailwind classes without conflicts | Already installed; use with clsx in cn() utility |
| @tailwindcss/postcss | ^4 | PostCSS plugin for Tailwind v4 | Already installed; build toolchain |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| geist npm package | next/font/google with Geist | Both work; npm package gives more control and matches Vercel's own usage |
| CSS grid background | SVG background pattern | CSS is simpler, more performant, and easier to customize |
| Framer Motion for tabs | CSS transitions only | Vercel uses animated tabs, but CONTEXT.md says no scroll-triggered animations; CSS hover states suffice |

**Installation:**
```bash
cd frontend && npm install geist
```

## Architecture Patterns

### Recommended Project Structure
```
frontend/src/
├── app/
│   ├── layout.tsx          # Root layout (Geist fonts, metadata)
│   ├── page.tsx            # Landing page composition
│   ├── globals.css         # Tailwind v4 theme + Vercel-style grid background
│   └── favicon.ico
├── components/
│   ├── landing/
│   │   ├── Navbar.tsx      # Fixed nav: wordmark + links + CTAs
│   │   ├── Hero.tsx        # Centered headline + CTAs + metrics + triangle
│   │   ├── UseCaseTabs.tsx # NEW: Horizontal tabs (ML Training, Inference, etc.)
│   │   ├── FeaturePillars.tsx # NEW: Source/Deploy/Scale sections
│   │   ├── CLIDemo.tsx     # NEW: gpuctl terminal demo (replaces CodeExample)
│   │   ├── FinalCTA.tsx    # Action buttons section
│   │   └── Footer.tsx      # Multi-column links
│   └── ui/
│       ├── Button.tsx      # Reusable button/link
│       ├── Container.tsx   # Max-width wrapper
│       └── index.ts        # Barrel exports
├── lib/
│   ├── utils.ts            # MISSING: cn() utility (clsx + tailwind-merge)
│   └── constants.ts        # MISSING: Page content data
```

### Pattern 1: CSS Grid Background with Cross Markers (Vercel-style)
**What:** A full-page background with subtle grid lines and small cross (+) markers at intersections
**When to use:** Behind all page sections for the Vercel aesthetic
**Example:**
```css
/* Grid background using repeating linear gradients + radial gradient crosses */
.grid-background {
  background-color: #000;
  background-image:
    /* Vertical lines */
    linear-gradient(
      to right,
      rgba(255, 255, 255, 0.03) 1px,
      transparent 1px
    ),
    /* Horizontal lines */
    linear-gradient(
      to bottom,
      rgba(255, 255, 255, 0.03) 1px,
      transparent 1px
    );
  background-size: 64px 64px;
}

/* Cross markers at intersections using pseudo-element with radial-gradient */
.grid-background::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  background-image: radial-gradient(
    circle 1px at center,
    rgba(255, 255, 255, 0.08) 1px,
    transparent 1px
  );
  background-size: 64px 64px;
}
```

### Pattern 2: Vercel Triangle with Rainbow Glow
**What:** The Vercel triangle logo rendered as SVG with a multi-color radial gradient glow behind it
**When to use:** Hero section centerpiece
**Example:**
```tsx
// Vercel triangle SVG
<svg viewBox="0 0 1024 1024" fill="white" width={180} height={180}>
  <path d="M512 128L896 832H128L512 128Z" />
</svg>

// Rainbow glow behind it (CSS)
.rainbow-glow {
  background: radial-gradient(
    ellipse at center,
    rgba(0, 112, 243, 0.3) 0%,     /* blue */
    rgba(0, 200, 150, 0.2) 25%,     /* teal/green */
    rgba(255, 50, 50, 0.15) 50%,    /* red */
    rgba(200, 100, 255, 0.1) 75%,   /* purple */
    transparent 100%
  );
  filter: blur(60px);
}
```

### Pattern 3: Horizontal Tab Section (Use Cases)
**What:** A row of clickable tabs that switch content below, styled like Vercel's category tabs
**When to use:** UseCaseTabs component
**Example:**
```tsx
"use client";
import { useState } from "react";

const TABS = [
  { id: "training", label: "ML Training", content: "..." },
  { id: "inference", label: "Inference", content: "..." },
  // ...
];

function UseCaseTabs() {
  const [active, setActive] = useState(0);
  return (
    <div>
      <div className="flex border-b border-white/10">
        {TABS.map((tab, i) => (
          <button
            key={tab.id}
            onClick={() => setActive(i)}
            className={`px-6 py-3 text-sm transition-colors ${
              i === active
                ? "border-b-2 border-white text-white"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="mt-8">{TABS[active].content}</div>
    </div>
  );
}
```

### Pattern 4: Terminal/CLI Code Demo
**What:** A styled terminal window showing gpuctl commands with syntax highlighting
**When to use:** CLIDemo section (maps to Vercel's AI Gateway code block)
**Example:**
```tsx
function CLIDemo() {
  return (
    <div className="rounded-lg border border-white/10 bg-[#0a0a0a]">
      {/* Terminal chrome */}
      <div className="flex items-center gap-2 border-b border-white/10 px-4 py-3">
        <div className="flex gap-1.5">
          <span className="h-3 w-3 rounded-full bg-[#ff5f56]" />
          <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
          <span className="h-3 w-3 rounded-full bg-[#27c93f]" />
        </div>
        <span className="ml-2 text-xs text-gray-500">terminal</span>
      </div>
      {/* Commands */}
      <div className="p-6 font-mono text-sm">
        <div>
          <span className="text-green-400">$</span>{" "}
          <span className="text-white">gpuctl launch --gpu h100 --count 4</span>
        </div>
        {/* ... more command lines */}
      </div>
    </div>
  );
}
```

### Pattern 5: Geist Font Integration with Next.js 16 + Tailwind v4
**What:** Replace current custom fonts with Vercel's Geist typeface
**When to use:** Root layout and globals.css
**Example:**
```tsx
// app/layout.tsx
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";

export default function RootLayout({ children }) {
  return (
    <html lang="en" className={`${GeistSans.variable} ${GeistMono.variable}`}>
      <body>{children}</body>
    </html>
  );
}
```
```css
/* globals.css - Tailwind v4 theme integration */
@import "tailwindcss";

@theme inline {
  --font-sans: var(--font-geist-sans), system-ui, sans-serif;
  --font-mono: var(--font-geist-mono), ui-monospace, monospace;
}
```

### Anti-Patterns to Avoid
- **Keeping ComputeField canvas:** The canvas particle animation is explicitly removed. Delete the component entirely.
- **Purple accent color scheme:** The current design uses purple (#7c6bf0) as primary accent. Vercel uses a white/black/gray palette with rainbow glow for emphasis. The Vercel style has no single accent color.
- **Film grain overlay:** The current body::after with noise texture is not part of Vercel's design. Remove it.
- **Stripe-style vertical border lines:** The current .stripe-borders/.stripe-lines layout is replaced by the grid background pattern. Remove.
- **Custom type scale classes:** The current type-display, type-h1, etc. classes assume the old font stack. Rewrite the typography system for Geist.
- **gradient-text and gradient-btn classes:** These use purple gradients from the old design. Replace with Vercel-style white text and clean button styling.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Font loading/optimization | Manual @font-face declarations | `geist` npm package with Next.js font loader | Handles subset, preload, display:swap, CSS variables automatically |
| Responsive grid background | Canvas-based grid rendering | CSS background-image with linear-gradient | Pure CSS is GPU-accelerated, no JS overhead, infinitely tiling |
| Class name merging | String concatenation or manual dedup | `cn()` from clsx + tailwind-merge | Handles Tailwind class conflicts correctly (e.g., px-4 vs px-6) |
| Dark theme toggling | Custom theme provider | Single dark theme (no toggle needed) | CONTEXT.md specifies "dark theme throughout" -- no light mode |

**Key insight:** The entire page is a static marketing landing page with zero data fetching. Every component can be a Server Component except the tab switcher and mobile nav toggle. Minimize "use client" directives to those two components only.

## Common Pitfalls

### Pitfall 1: Missing lib/ Directory Breaks Build
**What goes wrong:** The current frontend has imports to `@/lib/utils` (cn function) and `@/lib/constants` (page data) but neither file exists on disk. The build fails with MODULE_NOT_FOUND.
**Why it happens:** Phase 7 cleanup created components referencing these files but never created the files themselves.
**How to avoid:** Create `src/lib/utils.ts` with the cn() function and `src/lib/constants.ts` with all page content data as the very first task.
**Warning signs:** `next build` fails immediately with module resolution errors.

### Pitfall 2: Tailwind v4 CSS-First Config Confusion
**What goes wrong:** Tailwind v4 uses `@theme inline {}` in CSS instead of `tailwind.config.js`. Developers familiar with v3 try to create a config file.
**Why it happens:** Most tutorials and examples online still reference Tailwind v3 configuration.
**How to avoid:** All theme customization goes in `globals.css` using `@theme inline { }` blocks. No tailwind.config.js file needed. The project already uses this pattern correctly.
**Warning signs:** Creating a tailwind.config.js file; using `theme: { extend: {} }` syntax.

### Pitfall 3: Geist Font Variable Name Mismatch
**What goes wrong:** The `geist` npm package exposes CSS variables as `--font-geist-sans` and `--font-geist-mono`. If the @theme block maps to different variable names, fonts silently fall back to system fonts.
**Why it happens:** The current codebase uses `--font-vremena-grotesk` and `--font-necto-mono`. After switching to Geist, the variable names must match exactly.
**How to avoid:** Use `GeistSans.variable` and `GeistMono.variable` from the package (they generate the correct CSS variable names). Map them in @theme: `--font-sans: var(--font-geist-sans)`.
**Warning signs:** Text renders in system font instead of Geist; font-family in DevTools shows fallback.

### Pitfall 4: Grid Background Alignment Across Sections
**What goes wrong:** If the grid background is applied per-section with padding offsets, the grid lines don't align continuously across the full page.
**Why it happens:** Each section has its own padding/margin, so background patterns start at different offsets.
**How to avoid:** Apply the grid background to the `<body>` or a single full-page wrapper element, not individual sections. Sections are transparent and overlay on top.
**Warning signs:** Visible grid line discontinuities at section boundaries.

### Pitfall 5: Rainbow Glow Clipping
**What goes wrong:** The rainbow radial gradient behind the triangle gets clipped by `overflow: hidden` on parent containers.
**Why it happens:** The glow needs to extend beyond the hero section boundaries for the intended visual effect.
**How to avoid:** Use `overflow: visible` on the hero section and apply the glow as a positioned pseudo-element or absolutely-positioned div that bleeds into adjacent sections.
**Warning signs:** Glow appears as a hard-edged rectangle instead of a soft diffused radial gradient.

### Pitfall 6: Server Component vs Client Component Boundaries
**What goes wrong:** Marking parent components as "use client" forces all children to be client-rendered, increasing bundle size.
**Why it happens:** Easy to sprinkle "use client" everywhere when only specific interactive parts need it.
**How to avoid:** Only the UseCaseTabs (tab switching) and Navbar (mobile menu + scroll state) need "use client". All other components (Hero, FeaturePillars, CLIDemo, FinalCTA, Footer) should be Server Components.
**Warning signs:** Unnecessary "use client" directives; large client-side JS bundle for a static page.

## Code Examples

### cn() Utility (Required - Missing File)
```typescript
// src/lib/utils.ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

### Vercel-Style Color Palette (Dark Theme)
```css
/* globals.css @theme inline block */
@theme inline {
  /* Backgrounds */
  --color-bg: #000000;
  --color-bg-card: #0a0a0a;
  --color-bg-card-hover: #111111;

  /* Text */
  --color-text: #ededed;
  --color-text-muted: #a1a1a1;
  --color-text-dim: #666666;

  /* Borders */
  --color-border: rgba(255, 255, 255, 0.08);
  --color-border-light: rgba(255, 255, 255, 0.14);

  /* Fonts */
  --font-sans: var(--font-geist-sans), system-ui, sans-serif;
  --font-mono: var(--font-geist-mono), ui-monospace, monospace;
}
```

### Vercel-Style Nav
```tsx
// Simplified Vercel nav structure for GPU.ai
<nav className="fixed top-0 z-50 w-full border-b border-white/[0.08]">
  <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between px-6">
    <span className="text-lg font-semibold tracking-tight text-white">GPU.ai</span>
    <div className="hidden items-center gap-8 md:flex">
      {/* nav links */}
    </div>
    <div className="flex items-center gap-4">
      <a className="text-sm text-gray-400 hover:text-white">Log in</a>
      <a className="rounded-full bg-white px-4 py-1.5 text-sm font-medium text-black">
        Sign Up
      </a>
    </div>
  </div>
</nav>
```

### Vercel Homepage Structure (Full Page Composition)
```tsx
// app/page.tsx
import { Navbar } from "@/components/landing/Navbar";
import { Hero } from "@/components/landing/Hero";
import { UseCaseTabs } from "@/components/landing/UseCaseTabs";
import { FeaturePillars } from "@/components/landing/FeaturePillars";
import { CLIDemo } from "@/components/landing/CLIDemo";
import { FinalCTA } from "@/components/landing/FinalCTA";
import { Footer } from "@/components/landing/Footer";

export default function LandingPage() {
  return (
    <div className="grid-background">
      <Navbar />
      <main>
        <Hero />
        <UseCaseTabs />
        <FeaturePillars />
        <CLIDemo />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
```

## State of the Art

| Old Approach (Current) | Current Approach (Target) | When Changed | Impact |
|------------------------|--------------------------|--------------|--------|
| Custom fonts (Vremena Grotesk + Necto Mono) | Geist Sans + Geist Mono | Phase 8 | Full typography stack replacement |
| Purple accent color (#7c6bf0) | White/black/gray with rainbow glow | Phase 8 | Entire color palette rewrite |
| Stripe-style vertical border lines | CSS grid background with cross markers | Phase 8 | Background visual approach change |
| Canvas particle animation (ComputeField) | Static layout with hover states only | Phase 8 | Remove JS-heavy animation |
| PricingWidget + PricingTable | Removed (deferred) | Phase 8 | Simplify page; pricing comes later |
| Film grain overlay (body::after) | Clean Vercel-style background | Phase 8 | Remove texture overlay |
| Custom ChipLogo SVG | "GPU.ai" text wordmark | Phase 8 | Simpler logo treatment |
| tailwind.config.js (old convention) | @theme inline in globals.css (Tailwind v4) | Already done | Project already uses v4 pattern |

**Deprecated/outdated:**
- ComputeField.tsx: Canvas particle animation -- remove entirely
- EffectsToggle.tsx: Global effects state -- remove entirely
- PricingWidget.tsx: Interactive pricing comparison -- deferred
- PricingTable.tsx: Full pricing comparison table -- deferred
- TrustBar.tsx: Customer logo carousel -- deferred (no real customers yet)
- HowItWorks.tsx: 3-step cards -- replaced by FeaturePillars
- Features.tsx: 3-column feature grid -- replaced by FeaturePillars
- ChipLogo.tsx: Animated chip SVG logo -- replaced by text wordmark
- Counter.tsx: Animated number counter -- replaced by static metrics
- Card.tsx: General purpose card -- may not be needed in new design
- Section.tsx/SectionLabel.tsx: Section wrapper + label -- replace with Vercel-style section pattern
- Icon.tsx: SVG icon component -- keep if footer needs social icons

## Components to Create

| Component | Type | Description |
|-----------|------|-------------|
| Navbar | Client | Fixed nav with text wordmark, links, CTAs, mobile hamburger, scroll glassmorphism |
| Hero | Server | Centered headline, subtitle, 2 CTAs, static metrics row, triangle SVG + rainbow glow |
| UseCaseTabs | Client | 5 horizontal tabs (ML Training, Inference, Fine-tuning, Rendering, Research) with content panels |
| FeaturePillars | Server | 3 sections: Source (multi-provider aggregation), Deploy (instant provisioning), Scale (auto-scaling) |
| CLIDemo | Server | Terminal window with gpuctl command examples + syntax highlighting |
| FinalCTA | Server | "Launch GPU Instance" + "Talk to our team" buttons |
| Footer | Server | Multi-column link grid + copyright |

## Components to Delete

| Component | Reason |
|-----------|--------|
| ComputeField.tsx | Explicit removal per CONTEXT.md |
| EffectsToggle.tsx | Explicit removal per CONTEXT.md |
| PricingWidget.tsx | Deferred per CONTEXT.md |
| PricingTable.tsx | Deferred per CONTEXT.md |
| TrustBar.tsx | No equivalent in Vercel design per CONTEXT.md |
| HowItWorks.tsx | Replaced by FeaturePillars |
| Features.tsx | Replaced by FeaturePillars |
| ChipLogo.tsx | Replaced by text wordmark |
| Counter.tsx | Replaced by static text metrics |
| Card.tsx | Not needed in Vercel-style layout |
| Section.tsx | Replaced by simpler section wrappers |
| SectionLabel.tsx | Not used in Vercel-style design |

## Open Questions

1. **Vercel triangle graphic licensing**
   - What we know: The Vercel triangle/logo is trademarked. CONTEXT.md says "Use Vercel's triangle graphic as-is for the hero (placeholder -- replace later)"
   - What's unclear: Whether using the Vercel triangle even as a placeholder is legally acceptable for a production site
   - Recommendation: Create a GPU.ai-specific geometric shape (e.g., a stylized GPU/chip outline) or use a simple abstract triangle as a placeholder. The rainbow glow effect is the key visual -- the shape inside is secondary.

2. **Custom fonts removal**
   - What we know: Vremena Grotesk and Necto Mono .woff2 files exist in public/fonts/
   - What's unclear: Whether any other page (future dashboard) will use these fonts
   - Recommendation: Delete the font files in this phase. If needed later, they can be re-added. The Geist font family covers all use cases (sans for headings/body, mono for code).

3. **Logo images in public/logos/**
   - What we know: Seven customer logo PNGs exist for the TrustBar component being removed
   - What's unclear: Whether to delete them now or leave for potential future use
   - Recommendation: Leave the files for now; they take minimal space and could be reused. Don't reference them in the new page.

## Sources

### Primary (HIGH confidence)
- Vercel.com/home -- Direct page structure analysis via WebFetch (Feb 2026)
- Vercel.com/geist/colors -- Geist design system color palette documentation
- Vercel.com/geist/grid -- Geist grid component documentation
- Vercel.com/font -- Geist font family documentation and installation

### Secondary (MEDIUM confidence)
- NPM geist package -- Font installation: `npm install geist`, exports GeistSans and GeistMono, CSS variables --font-geist-sans and --font-geist-mono
- GitHub vercel/geist-font -- Font repository and integration examples
- Tailwind CSS v4 documentation -- @theme inline configuration pattern
- Next.js Font Optimization docs -- next/font integration patterns

### Tertiary (LOW confidence)
- Various Vercel clone tutorials (CodePen, GitHub) -- General layout patterns but specific Vercel 2026 design details unverified
- WebSearch results for grid background CSS patterns -- Technique is well-established but exact Vercel implementation details require browser inspection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Existing stack is correct; only adding geist font package
- Architecture: HIGH - Clear component structure mapped directly from Vercel homepage sections to CONTEXT.md requirements
- Pitfalls: HIGH - Identified from direct codebase analysis (broken build, missing files) and established patterns (Tailwind v4, font loading, CSS backgrounds)

**Research date:** 2026-02-26
**Valid until:** 2026-03-26 (stable domain -- Vercel homepage design changes infrequently)
