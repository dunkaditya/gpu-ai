---
phase: 11
slug: dashboard-ui-redesign-clean-linear-aesthetic-gpu-catalog-with-categories-and-search-instances-table-fixes-professional-polish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None (no frontend test framework configured — visual UI redesign) |
| **Config file** | none |
| **Quick run command** | `cd frontend && npm run build` |
| **Full suite command** | `cd frontend && npm run build && npm run lint` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd frontend && npm run build`
- **After every plan wave:** Run `cd frontend && npm run build && npm run lint`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 11-01-01 | 01 | 1 | UI-01 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 11-01-02 | 01 | 1 | UI-02 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 11-01-03 | 01 | 1 | UI-03 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 11-01-04 | 01 | 1 | UI-04 | build | `cd frontend && npm run build` | ✅ | ⬜ pending |
| 11-01-05 | 01 | 1 | UI-05 | manual-only | Visual inspection | N/A | ⬜ pending |
| 11-01-06 | 01 | 1 | UI-06 | manual-only | Visual inspection | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. No new test framework needed — this is a pure visual/UX redesign. Build + lint infrastructure is sufficient for automated validation.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Linear aesthetic CSS applied | UI-01 | Visual design quality | Inspect dashboard pages for clean dark theme, muted borders, proper spacing |
| GPU category tabs filter correctly | UI-02 | Interactive UI behavior | Click each category tab, verify correct GPU subset shown |
| Search filters GPU offerings | UI-03 | Interactive UI behavior | Type GPU names in search, verify real-time filtering |
| Instances table columns align | UI-04 | Visual layout correctness | Compare header columns with row columns, verify alignment |
| Consistent polish across pages | UI-05 | Cross-page visual consistency | Navigate all dashboard pages, check typography/color/spacing consistency |
| Sidebar/topbar refined | UI-06 | Visual design quality | Inspect navigation components for linear aesthetic |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
