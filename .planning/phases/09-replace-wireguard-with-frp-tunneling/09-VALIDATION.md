---
phase: 9
slug: replace-wireguard-with-frp-tunneling
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-09
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing (stdlib) |
| **Config file** | None — Go convention |
| **Quick run command** | `go test ./internal/tunnel/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/tunnel/... ./internal/provision/... ./internal/api/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | FRP-01 | unit | `go test ./internal/tunnel/ -run TestManagerStart` | ❌ W0 | ⬜ pending |
| 09-01-02 | 01 | 1 | FRP-02 | unit | `go test ./internal/tunnel/ -run TestAllocatePort` | ❌ W0 | ⬜ pending |
| 09-01-03 | 01 | 1 | FRP-03 | unit | `go test ./internal/tunnel/ -run TestRenderBootstrap` | ❌ W0 | ⬜ pending |
| 09-02-01 | 02 | 2 | FRP-04 | unit | `go test ./internal/provision/ -run TestProvision` | ❌ W0 | ⬜ pending |
| 09-02-02 | 02 | 2 | FRP-05 | unit | `go test ./internal/api/ -run TestInstanceToResponse` | ❌ W0 | ⬜ pending |
| 09-02-03 | 02 | 2 | FRP-06 | unit | `go test ./internal/config/ -run TestLoad` | ❌ W0 | ⬜ pending |
| 09-03-01 | 03 | 3 | FRP-07 | build | `go build ./...` | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/tunnel/manager_test.go` — stubs for FRP-01
- [ ] `internal/tunnel/ports_test.go` — stubs for FRP-02
- [ ] `internal/tunnel/template_test.go` — stubs for FRP-03

*Existing Go test infrastructure covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| frpc connects from RunPod container to frps | FRP-04 | Requires real RunPod environment | Provision a test instance, verify frpc establishes tunnel, SSH through proxy port |
| End-to-end SSH through FRP tunnel | FRP-05 | Requires running infrastructure | SSH to proxy_host:remote_port, verify connection routes to instance |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
