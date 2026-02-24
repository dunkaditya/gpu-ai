---
phase: 03-privacy-layer
plan: 03
subsystem: wireguard
tags: [cloud-init, text-template, go-embed, wireguard, privacy-scrubbing, bootstrap, shell-injection-prevention]

# Dependency graph
requires:
  - phase: 03-privacy-layer
    provides: "WireGuard key generation, AES-256-GCM encryption, Config with proxy endpoint/public key"
provides:
  - "Embedded cloud-init bootstrap template with 7 sections (WireGuard, SSH, firewall, hostname, provider scrubbing, NVIDIA, ready callback)"
  - "BootstrapData struct for template rendering inputs"
  - "ValidateBootstrapData with shell injection prevention"
  - "RenderBootstrap renderer with validation-first pattern"
affects: [04-instance-lifecycle]

# Tech tracking
tech-stack:
  added: [text/template, embed]
  patterns: [go:embed for binary-embedded templates, validation-before-render, shell injection regex guards]

key-files:
  created:
    - internal/wireguard/templates/bootstrap.sh.tmpl
    - internal/wireguard/template.go
    - internal/wireguard/template_test.go

key-decisions:
  - "text/template used instead of html/template to avoid HTML-escaping bash characters"
  - "Single-quoted heredocs (WGEOF, SSHEOF) prevent bash variable expansion -- Go template expands before script runs"
  - "SSH key validation combines format regex with shell injection character blocklist"
  - "CallbackURL is pre-rendered by Go code (full URL), not constructed in bash"

patterns-established:
  - "go:embed for templates: embed directive with text/template parsing at package init via template.Must"
  - "Validation-before-render: RenderBootstrap calls ValidateBootstrapData before template execution"
  - "Shell injection prevention: regex-based input validation for all user-influenced template fields"

requirements-completed: [PRIV-05, PRIV-07]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 3 Plan 3: Cloud-Init Template Summary

**Embedded Go text/template for GPU instance bootstrap with WireGuard tunnel, SSH lockdown, provider scrubbing, NVIDIA verification, and shell injection validation**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T22:17:28Z
- **Completed:** 2026-02-24T22:19:57Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Cloud-init template with all 7 bootstrap sections embedded in binary via go:embed
- WireGuard kernel-to-wireguard-go userspace fallback for container environments
- Provider identity scrubbing: metadata endpoint block, env var removal, CLI tool removal, service disabling
- Input validation prevents shell injection via instance IDs, SSH key comments, and hostnames
- 10 comprehensive tests covering rendering, multi-key SSH, validation, 3 injection vectors, and output correctness

## Task Commits

Each task was committed atomically:

1. **Task 1: Cloud-init Go template with full privacy scrubbing and WireGuard fallback** - `4d6c46a` (feat)
2. **Task 2: Template renderer with input validation and comprehensive tests** - `5da8a9e` (feat)

## Files Created/Modified
- `internal/wireguard/templates/bootstrap.sh.tmpl` - Go text/template cloud-init script with 7 sections (166 lines)
- `internal/wireguard/template.go` - BootstrapData struct, ValidateBootstrapData, RenderBootstrap with go:embed
- `internal/wireguard/template_test.go` - 10 tests covering rendering, validation, injection prevention, and output correctness

## Decisions Made
- Used text/template (not html/template) per CONTEXT.md -- html/template would HTML-escape bash characters like <, >, &
- Single-quoted heredocs (WGEOF, SSHEOF) in template prevent bash variable expansion; Go template variables are expanded before the script runs on the instance
- SSH key validation combines format regex with explicit shell injection character blocklist ($, backtick, ;, |, &)
- CallbackURL is the full pre-rendered URL from Go code, not constructed in bash, to avoid injection risk

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required. Template is embedded in the binary at compile time.

## Next Phase Readiness
- Cloud-init template ready for provisioning engine to render per-instance bootstrap scripts
- BootstrapData struct ready to be populated from Config (proxy endpoint, public key) and per-instance data (keys, addresses)
- Template renderer exported as RenderBootstrap for use by provisioning orchestration in Phase 4
- All 28 wireguard package tests pass (18 pre-existing + 10 new)

## Self-Check: PASSED

All files exist. All commits verified.

---
*Phase: 03-privacy-layer*
*Completed: 2026-02-24*
