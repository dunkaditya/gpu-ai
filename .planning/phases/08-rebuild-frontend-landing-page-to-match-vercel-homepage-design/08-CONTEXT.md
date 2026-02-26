# Phase 8: Rebuild Frontend Landing Page to Match Vercel Homepage Design - Context

**Gathered:** 2026-02-26
**Status:** Ready for planning

<domain>
## Phase Boundary

Rebuild the existing GPU.ai landing page to match the layout, structure, and visual style of Vercel's homepage (vercel.com/home as of Feb 2026). Content and graphics are adapted for GPU.ai's product, but the page structure mirrors Vercel's exactly. No dashboard pages, no auth flows, no pricing widget — just the public-facing landing page.

</domain>

<decisions>
## Implementation Decisions

### Design Fidelity
- Use Vercel's exact layout structure with GPU.ai branding
- Content and graphics adapted to what GPU.ai actually does — not a content clone
- Reference: vercel.com/home (current as of Feb 2026)
- Match Vercel's rainbow color theme
- Logo: simple "GPU.ai" text wordmark in clean typography, Vercel-style
- Pull Vercel's frontend source code and adapt it directly rather than rebuilding from scratch

### Section Structure (top to bottom)
1. **Navbar** — GPU.ai text wordmark + navigation links (match Vercel's nav structure)
2. **Hero** — Centered headline + subtitle + two CTAs + static metrics + Vercel triangle graphic with rainbow glow
3. **Use Case Tabs** — ML Training, Inference, Fine-tuning, Rendering, Research (maps to Vercel's AI Apps/Web Apps/Ecommerce/Marketing/Platforms tabs)
4. **Feature Pillars** — "Source, Deploy, Scale" sections (maps to Vercel's "Your Product, Delivered" / infrastructure sections)
5. **CLI Demo** — gpuctl command examples showing instance creation, status check, SSH (maps to Vercel's AI Gateway code demo)
6. **Final CTA** — Action buttons
7. **Footer** — Multi-column links

**Removed from current page:**
- PricingWidget (drop for this phase)
- PricingTable section (drop for this phase)
- ComputeField animated background (remove)
- EffectsToggle (remove)
- TrustBar (no equivalent in Vercel's current design)
- Templates section (skip — GPU.ai doesn't have templates yet)

### Content & Messaging
- Hero headline: Vercel-style single concise line (Claude writes, conveying GPU.ai's value prop)
- Hero CTAs: "Launch GPU Instance" (primary) + "Talk to our team" (secondary)
- Tone: technical-polished, matching Vercel's style — developer-facing, confident, concise
- Hero stats: static metrics (e.g., cost savings, deploy time, GPU count) — placeholder/aspirational numbers fine, no rotation needed
- Body copy throughout adapted to GPU.ai's actual capabilities

### Visual Effects
- Match Vercel's grid background exactly — grid lines with cross markers at intersections, dark cells, same spacing
- Match Vercel's rainbow glow exactly — blue-green-red radial gradient behind the triangle graphic
- No scroll-triggered animations (Vercel doesn't use them)
- No ComputeField particles or EffectsToggle — clean, Vercel-style static layout with hover states
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

</decisions>

<specifics>
## Specific Ideas

- Pull Vercel's frontend source code (vercel.com/home) and adapt it directly — don't rebuild from scratch
- Use Vercel's triangle graphic as-is for the hero (placeholder — replace later)
- Use case tabs: ML Training, Inference, Fine-tuning, Rendering, Research
- Feature pillars follow "Source, Deploy, Scale" narrative
- CLI demo should show gpuctl commands (not REST API)
- Hero CTAs are exactly: "Launch GPU Instance" and "Talk to our team"

</specifics>

<deferred>
## Deferred Ideas

- Interactive pricing widget — separate phase or add back later
- GPU templates gallery (one-click deploy environments) — future phase
- Customer logos / social proof — add when real customers exist

</deferred>

---

*Phase: 08-rebuild-frontend-landing-page-to-match-vercel-homepage-design*
*Context gathered: 2026-02-26*
