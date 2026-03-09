---
phase: 09-replace-wireguard-with-frp-tunneling
plan: 01
subsystem: infra
tags: [frp, tunneling, ssh, wireguard-replacement, tcp-proxy]

# Dependency graph
requires:
  - phase: 03-privacy-layer
    provides: WireGuard IPAM pattern, bootstrap template pattern, encryption
provides:
  - internal/tunnel package with types, port allocator, bootstrap template, and FRP manager
  - v7 database migration adding frp_remote_port column
  - FRP config loading (FRP_BIND_PORT, FRP_TOKEN, FRP_ALLOW_PORTS)
affects: [09-02-PLAN (provisioning engine rewiring), 09-03-PLAN (WireGuard removal)]

# Tech tracking
tech-stack:
  added: [github.com/fatedier/frp@v0.67.0]
  patterns: [embedded frps server, port-based allocation replacing IP-based IPAM, frpc bootstrap]

key-files:
  created:
    - internal/tunnel/types.go
    - internal/tunnel/ports.go
    - internal/tunnel/ports_test.go
    - internal/tunnel/template.go
    - internal/tunnel/template_test.go
    - internal/tunnel/manager.go
    - internal/tunnel/manager_test.go
    - internal/tunnel/templates/bootstrap.sh.tmpl
    - database/migrations/20260309_v7_frp_tunneling.sql
  modified:
    - internal/config/config.go
    - go.mod
    - go.sum

key-decisions:
  - "Port range 10000-10255 matching existing WG port formula for consistency"
  - "FRP token auth (not OIDC) for simplicity in machine-to-machine auth"
  - "Embedded frps via Go library maintaining single-binary architecture"
  - "Advisory lock ID 0x4650525054 (FRPPT) for port allocation serialization"
  - "PortsRange type from frp/pkg/config/types used directly for AllowPorts config"

patterns-established:
  - "tunnel.Manager wraps frps Service with Start(ctx)/Close() lifecycle"
  - "Port allocation via SELECT MAX(frp_remote_port) with advisory lock (parallel to IPAM pattern)"
  - "Bootstrap template downloads frpc binary from GitHub CDN"

requirements-completed: [FRP-01, FRP-02, FRP-03, FRP-06]

# Metrics
duration: 7min
completed: 2026-03-09
---

# Phase 09 Plan 01: FRP Tunnel Infrastructure Summary

**FRP tunnel package with embedded frps manager, port allocator, bootstrap template, and config loading replacing WireGuard internals**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-09T23:07:59Z
- **Completed:** 2026-03-09T23:14:35Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Created complete internal/tunnel package with types, port allocator, template renderer, and FRP manager
- Database migration v7 adds frp_remote_port column with partial unique index for active instances
- Bootstrap template downloads frpc, configures SSH tunneling, and sends ready callback
- Config loads FRP_BIND_PORT (default 7000), FRP_TOKEN, FRP_ALLOW_PORTS (default 10000-10255)
- All existing WireGuard code untouched -- builds and tests continue to pass

## Task Commits

Each task was committed atomically (TDD: test then feat):

1. **Task 1: Database migration, tunnel types, port allocator, and bootstrap template**
   - `8f89bb6` (test) - Failing tests for port allocator and bootstrap template
   - `2b0eaa3` (feat) - Migration, types, ports, template, bootstrap script

2. **Task 2: FRP manager and config loading**
   - `39f4f4e` (test) - Failing tests for FRP manager
   - `3964302` (feat) - Manager implementation and config FRP fields

## Files Created/Modified
- `database/migrations/20260309_v7_frp_tunneling.sql` - Adds frp_remote_port column with partial unique index
- `internal/tunnel/types.go` - BootstrapData struct, port range constants, advisory lock ID
- `internal/tunnel/ports.go` - AllocatePort with PostgreSQL advisory lock serialization
- `internal/tunnel/ports_test.go` - Tests for sequential allocation, exhaustion, terminated reclamation
- `internal/tunnel/template.go` - RenderBootstrap with validation and embedded template
- `internal/tunnel/template_test.go` - Tests for frpc TOML config, SSH keys, callback URL
- `internal/tunnel/manager.go` - FRP server lifecycle manager wrapping frpserver.Service
- `internal/tunnel/manager_test.go` - Tests for valid/invalid manager creation
- `internal/tunnel/templates/bootstrap.sh.tmpl` - Bootstrap script with frpc download and SSH tunneling
- `internal/config/config.go` - Added FRPBindPort, FRPToken, FRPAllowPorts fields
- `go.mod` - Added github.com/fatedier/frp@v0.67.0
- `go.sum` - Updated with FRP transitive dependencies

## Decisions Made
- Port range 10000-10255 matches existing WG port formula for consistency during migration
- FRP token auth chosen over OIDC for simplicity in machine-to-machine authentication
- Embedded frps via Go library (not subprocess) to maintain single-binary architecture
- PortsRange type from frp/pkg/config/types used directly (not v1 package) for AllowPorts config
- Empty FRP_TOKEN disables FRP tunneling (same optional pattern as WireGuard)
- Ephemeral ports used in manager tests to avoid port conflicts

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed PortsRange import path**
- **Found during:** Task 2 (FRP manager implementation)
- **Issue:** Plan referenced v1.PortsRange but the actual type is in frp/pkg/config/types package
- **Fix:** Added types import and used types.PortsRange instead of v1.PortsRange
- **Files modified:** internal/tunnel/manager.go
- **Verification:** go build ./... succeeds
- **Committed in:** 3964302 (Task 2 commit)

**2. [Rule 1 - Bug] Used ephemeral ports in manager tests**
- **Found during:** Task 2 (FRP manager tests)
- **Issue:** Port 7000 was already in use on dev machine, causing test failures
- **Fix:** Added ephemeralPort() helper that finds available ports via net.Listen
- **Files modified:** internal/tunnel/manager_test.go
- **Verification:** All manager tests pass reliably
- **Committed in:** 3964302 (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** Both auto-fixes necessary for correctness. No scope creep.

## Issues Encountered
None beyond the auto-fixed deviations above.

## User Setup Required
None - no external service configuration required. FRP env vars are optional.

## Next Phase Readiness
- internal/tunnel package is complete and ready for plan 02 (provisioning engine rewiring)
- All existing WireGuard code continues to work (no breakage)
- Config struct has both WG and FRP fields for gradual migration

## Self-Check: PASSED

All 10 files verified present. All 4 task commits verified in git log.

---
*Phase: 09-replace-wireguard-with-frp-tunneling*
*Completed: 2026-03-09*
