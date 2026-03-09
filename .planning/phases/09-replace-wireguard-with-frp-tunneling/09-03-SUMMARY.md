---
phase: 09-replace-wireguard-with-frp-tunneling
plan: 03
subsystem: infra
tags: [wireguard, frp, cleanup, go-mod, dependency-removal]

requires:
  - phase: 09-02
    provides: "FRP-based provisioning wiring in engine, handlers, and deps"
provides:
  - "Clean codebase with zero WireGuard code or imports"
  - "Removed wgctrl direct dependency from go.mod"
  - "Config struct without WG fields"
  - "Clean provider.ProvisionRequest without WireGuardAddress"
affects: []

tech-stack:
  added: []
  patterns:
    - "FRP tunneling is now the sole SSH access mechanism (replaces WireGuard)"

key-files:
  created: []
  modified:
    - "go.mod"
    - "go.sum"
    - "internal/config/config.go"
    - "internal/provider/types.go"

key-decisions:
  - "WireGuard indirect deps (golang.zx2c4.com/wireguard, wintun) remain as transitive FRP library dependencies -- not removable"
  - "WG config env vars (WG_ENCRYPTION_KEY, WG_PROXY_ENDPOINT, WG_PROXY_PUBLIC_KEY, WG_INTERFACE_NAME) fully removed from config"
  - "WireGuardAddress field removed from provider.ProvisionRequest as no code sets it anymore"

patterns-established:
  - "Single tunnel implementation: FRP only, no WireGuard fallback"

requirements-completed: [FRP-07]

duration: 7min
completed: 2026-03-09
---

# Phase 09 Plan 03: Delete WireGuard Package Summary

**Removed entire internal/wireguard/ package (10 files, ~2000 LOC), wgctrl dependency, and all WG config fields -- completing the WireGuard to FRP migration**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-09T23:19:34Z
- **Completed:** 2026-03-09T23:26:31Z
- **Tasks:** 1
- **Files modified:** 14

## Accomplishments
- Deleted entire internal/wireguard/ directory (10 files: manager, IPAM, keygen, template, types, tests, bootstrap template)
- Removed WG config fields from Config struct (WGEncryptionKey, WGEncryptionKeyBytes, WGProxyEndpoint, WGProxyPublicKey, WGInterfaceName)
- Removed WG validation block from config.Load() (all-or-nothing check, hex decode)
- Removed WireGuardAddress from provider.ProvisionRequest
- Ran go mod tidy to remove wgctrl direct dependency and related indirect deps
- Verified: go build, go test, go vet all pass with zero WireGuard code

## Task Commits

Each task was committed atomically:

1. **Task 1: Delete WireGuard package and remove wgctrl dependency** - `ef2c947` (feat)

**Plan metadata:** pending (docs: complete plan)

## Files Created/Modified
- `internal/wireguard/` - Entire directory deleted (10 files)
- `internal/config/config.go` - Removed WG fields, WG validation, encoding/hex import
- `internal/provider/types.go` - Removed WireGuardAddress from ProvisionRequest
- `go.mod` - Removed wgctrl direct dependency
- `go.sum` - Updated after go mod tidy

## Decisions Made
- WireGuard indirect deps (golang.zx2c4.com/wireguard, wintun) remain in go.mod as they are transitive dependencies of the FRP library (fatedier/frp) -- these cannot be removed without removing FRP itself
- Removed encoding/hex import from config.go since it was only used for WG_ENCRYPTION_KEY hex decoding
- WireGuardAddress removed from provider.ProvisionRequest since no code sets it after engine.go uses FRP

## Deviations from Plan

None - plan executed exactly as written. The plan anticipated that files from plan 02 might need fixing (step 2: "If any remain, fix them"), but plan 02 was already committed (commits 146ab59, 11de923) so no additional fixes were needed.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 09 (Replace WireGuard with FRP Tunneling) is complete
- The codebase now uses FRP as the sole tunnel implementation
- All three plans executed: infrastructure (09-01), wiring (09-02), cleanup (09-03)
- Ready for deployment: re-build gpuctl-linux and deploy to droplet with FRP_TOKEN env var

---
*Phase: 09-replace-wireguard-with-frp-tunneling*
*Completed: 2026-03-09*
