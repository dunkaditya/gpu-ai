---
phase: 12
slug: mobile-responsive-website-optimization
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-25
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Manual visual testing + Next.js build verification |
| **Config file** | none |
| **Quick run command** | `cd frontend && npm run build` |
| **Full suite command** | `cd frontend && npm run build && npm run lint` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd frontend && npm run build`
- **After every plan wave:** Run `cd frontend && npm run build && npm run lint`
- **Before `/gsd:verify-work`:** Full suite must be green + manual visual review at 375px
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 12-01-01 | 01 | 1 | MOBILE-01 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 12-01-02 | 01 | 1 | MOBILE-05 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 12-02-01 | 02 | 1 | MOBILE-03 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 12-02-02 | 02 | 1 | MOBILE-04 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 12-02-03 | 02 | 1 | MOBILE-02 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 12-03-01 | 03 | 2 | MOBILE-06 | manual | Visual inspection of touch targets | N/A | ⬜ pending |
| 12-03-02 | 03 | 2 | MOBILE-05 | manual | Browser DevTools at 375px width | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No test framework setup needed — this phase is purely CSS/layout changes verified by build success and manual visual testing.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Landing page renders correctly at 375px | MOBILE-01 | Visual layout verification | Open Chrome DevTools, set to 375px iPhone, check all landing sections |
| Dashboard sidebar opens/closes on mobile | MOBILE-02 | Interactive behavior | Toggle hamburger menu, verify overlay + slide-in |
| Billing tables show mobile cards | MOBILE-03 | Layout verification | Navigate to billing page at 375px, verify card layout |
| Forms usable at 375px | MOBILE-04 | Layout + interaction | Test settings form, launch modal at 375px |
| No horizontal scroll at 375px | MOBILE-05 | Viewport verification | Check every page at 375px for horizontal overflow |
| Touch targets >= 44px | MOBILE-06 | Measurement | Inspect button/link padding, verify >= 44px tap area |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
