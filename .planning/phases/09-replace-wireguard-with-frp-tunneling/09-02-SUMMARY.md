---
phase: 09-replace-wireguard-with-frp-tunneling
plan: 02
subsystem: infra
tags: [frp, tunneling, ssh, provisioning, api, wireguard-replacement]

# Dependency graph
requires:
  - phase: 09-replace-wireguard-with-frp-tunneling
    plan: 01
    provides: internal/tunnel package (types, ports, template, manager), FRP config fields
  - phase: 04-auth-instance-lifecycle
    provides: provisioning engine, API handlers, deps wiring
provides:
  - Provisioning engine uses tunnel.AllocatePort + tunnel.RenderBootstrap for FRP tunneling
  - API responses derive SSH connection info from FRP remote port
  - deps.go initializes tunnel.Manager instead of wireguard.Manager
  - serve.go manages FRP server lifecycle (start/stop)
affects: [09-03-PLAN (WireGuard code removal)]

# Tech tracking
tech-stack:
  added: []
  patterns: [FRP port allocation in provisioning flow, extractHost URL parsing for proxy hostname]

key-files:
  created: []
  modified:
    - internal/db/instances.go
    - internal/provision/engine.go
    - internal/api/handlers.go
    - cmd/gpuctl/deps.go
    - cmd/gpuctl/serve.go

key-decisions:
  - "FRPToken reuses per-instance internalToken for FRP auth -- no separate token system needed"
  - "FRP cleanup is automatic: frpc dies with instance, port freed by DB status change"
  - "extractHost uses net/url.Parse to extract hostname from GpuctlPublicURL"
  - "WG fields kept as nil in new instances for historical compatibility"

patterns-established:
  - "extractHost helper parses proxy hostname from GpuctlPublicURL (used in both engine and handlers)"
  - "FRP port allocation within transaction before provider retry loop"

requirements-completed: [FRP-04, FRP-05]

# Metrics
duration: 5min
completed: 2026-03-09
---

# Phase 09 Plan 02: Provisioning and API Rewiring Summary

**Provisioning engine, API handlers, and main wiring switched from WireGuard to FRP tunnel package with port-based SSH connection info**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-09T23:18:54Z
- **Completed:** 2026-03-09T23:24:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Provisioning engine allocates FRP remote port via tunnel.AllocatePort and renders bootstrap script via tunnel.RenderBootstrap
- API responses derive SSH command from FRP remote port + proxy host instead of WireGuard tunnel IP
- FRP tunnel manager initializes in deps.go and starts/stops alongside HTTP server in serve.go
- All WireGuard imports removed from engine.go, handlers.go, deps.go -- WG code still exists but is unused

## Task Commits

Each task was committed atomically:

1. **Task 1: Update DB layer and provisioning engine for FRP** - `146ab59` (feat)
2. **Task 2: Update API handlers and main wiring for FRP** - `11de923` (feat)

## Files Created/Modified
- `internal/db/instances.go` - Added FRPRemotePort field to Instance struct, updated all scan lists and CreateInstance INSERT
- `internal/provision/engine.go` - Replaced WG key gen/IPAM/AddPeer with tunnel.AllocatePort + tunnel.RenderBootstrap, removed WG cleanup from Terminate
- `internal/api/handlers.go` - Replaced WG proxy connection info with FRP remote port, added extractHost helper
- `cmd/gpuctl/deps.go` - Replaced wireguard.Manager/IPAM with tunnel.Manager initialization
- `cmd/gpuctl/serve.go` - Added FRP server start/stop lifecycle management

## Decisions Made
- FRP token per-instance reuses the existing internalToken (no new auth system required)
- FRP cleanup is entirely automatic -- frpc process dies when provider terminates instance, port is freed by partial unique index excluding terminated rows
- extractHost function placed in both engine.go and handlers.go (duplicated intentionally since each package needs it independently)
- WG fields (WGPublicKey, WGPrivateKeyEnc, WGAddress) kept as nil for new instances -- historical records retain their values

## Deviations from Plan

None - plan executed exactly as written. The Go format tool (goimports) automatically cleaned up unused imports during saves, which matched the planned import removals.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required. FRP env vars were already configured in plan 01.

## Next Phase Readiness
- All provisioning and API code now uses FRP tunneling instead of WireGuard
- WireGuard code (internal/wireguard/) is still present but completely unused by engine, handlers, and deps
- Plan 03 (WireGuard removal) can safely remove the internal/wireguard/ package and WG config fields

## Self-Check: PASSED

All 5 modified files verified present. All 2 task commits verified in git log.

---
*Phase: 09-replace-wireguard-with-frp-tunneling*
*Completed: 2026-03-09*
