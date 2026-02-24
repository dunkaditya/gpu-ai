---
phase: 03-privacy-layer
plan: 02
subsystem: wireguard
tags: [wireguard, ipam, wgctrl-go, iptables, peer-management, port-mapping, advisory-lock, postgresql]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: "Config struct, database schema with wg_address column"
  - phase: 03-privacy-layer
    plan: 01
    provides: "WireGuard key generation, AES-256-GCM encryption, KeyPair type, wgctrl-go dependency"
provides:
  - "IPAM: PostgreSQL-backed IP allocation from 10.0.0.0/16 with advisory lock"
  - "WireGuard Manager: programmatic peer add/remove via wgctrl-go ConfigureDevice"
  - "iptables DNAT/FORWARD rules added/removed atomically with WireGuard peers"
  - "PortFromTunnelIP: derives external SSH port from tunnel IP address"
  - "WGClient interface for testable WireGuard device control"
  - "CommandRunner interface for testable iptables execution"
affects: [03-privacy-layer, 04-instance-lifecycle]

# Tech tracking
tech-stack:
  added: [os/exec]
  patterns: [PostgreSQL advisory lock for serialized allocation, interface-based dependency injection for WireGuard and iptables, atomic rollback on partial failure]

key-files:
  created:
    - internal/wireguard/ipam.go
    - internal/wireguard/ipam_test.go
    - internal/wireguard/manager_test.go
  modified:
    - internal/wireguard/manager.go

key-decisions:
  - "Advisory lock constant 0x475055414950414D (GPUAIPAM) for deterministic lock ID"
  - "IPAM uses pgx.Tx passed by caller; lock is transaction-scoped and auto-releases"
  - "Manager accepts WGClient and CommandRunner interfaces for full testability"
  - "iptables DNAT failure triggers automatic rollback of WireGuard peer"
  - "RemovePeer treats iptables cleanup as best-effort (logs but does not fail)"
  - "Port formula: 10000 + ip[2]*256 + ip[3] maps full /16 to ports 10002-75535"

patterns-established:
  - "Interface injection: WGClient and CommandRunner enable unit testing without root, real WG device, or real iptables"
  - "Atomic rollback: AddPeer rolls back WG peer if iptables setup fails"
  - "Best-effort cleanup: RemovePeer logs iptables errors but always removes the WG peer"
  - "Mock pgx.Tx: minimal struct implementing only Exec/QueryRow for IPAM tests"

requirements-completed: [PRIV-02, PRIV-03, PRIV-04, PRIV-06]

# Metrics
duration: 3min
completed: 2026-02-24
---

# Phase 3 Plan 2: IPAM & WireGuard Manager Summary

**PostgreSQL-backed IPAM with advisory lock IP allocation and WireGuard peer manager with atomic iptables port-mapping rules via wgctrl-go**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-24T22:17:35Z
- **Completed:** 2026-02-24T22:21:08Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- IPAM allocates sequential IPs from 10.0.0.0/16 with pg_advisory_xact_lock preventing races
- WireGuard Manager adds/removes peers programmatically via wgctrl-go ConfigureDevice (no shell-out to wg)
- iptables DNAT and FORWARD rules managed atomically alongside WireGuard peers
- Rollback on partial failure: if iptables fails after WG peer added, peer is removed
- 17 new unit tests total: 9 IPAM tests (allocation, carry, exhaustion, lock failure) + 8 Manager tests (add/remove/list/port mapping, failure paths)
- All tests pass without root, real WireGuard device, or real database

## Task Commits

Each task was committed atomically:

1. **Task 1: PostgreSQL-backed IPAM with advisory lock and unit tests** - `701682c` (feat)
2. **Task 2: WireGuard peer manager with iptables port mapping and unit tests** - `6289309` (feat)

## Files Created/Modified
- `internal/wireguard/ipam.go` - IPAM struct with AllocateAddress (advisory lock + sequential increment), incrementIP, IsProxyAddress, SubnetCIDR
- `internal/wireguard/ipam_test.go` - 9 tests: NewIPAM, invalid CIDR, incrementIP carry, proxy address, subnet CIDR, mock tx allocation, exhaustion, lock failure, query failure
- `internal/wireguard/manager.go` - Manager struct with AddPeer (WG + iptables), RemovePeer (iptables + WG), ListPeers, PortFromTunnelIP; WGClient and CommandRunner interfaces
- `internal/wireguard/manager_test.go` - 8 tests: add peer (full verification), invalid key, WG failure, iptables failure + rollback, remove peer, non-fatal iptables removal, list peers, port derivation

## Decisions Made
- Advisory lock uses constant `0x475055414950414D` (ASCII "GPUAIPAM") for a deterministic, memorable lock ID
- IPAM receives `pgx.Tx` from caller rather than managing its own connection -- fits the pattern where provisioning orchestrator owns the transaction
- Manager constructor takes WGClient and CommandRunner as parameters rather than creating them internally -- enables complete mock injection
- AddPeer always adds DNAT rule first, then FORWARD rule; if either fails, all prior work is rolled back
- RemovePeer removes iptables rules before WG peer (iptables errors are non-fatal); peer is always removed even if iptables cleanup fails
- Port formula `10000 + ip[2]*256 + ip[3]` provides a direct mapping from the full /16 range without collisions

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required. IPAM and Manager are runtime components that use existing database and WireGuard infrastructure.

## Next Phase Readiness
- IPAM and Manager ready for integration into provisioning orchestration (Phase 4)
- PortFromTunnelIP available for cloud-init template rendering (Plan 03)
- WGClient interface ready to wrap real wgctrl.Client in production
- CommandRunner interface ready to use os/exec in production

## Self-Check: PASSED

All files exist. All commits verified.

---
*Phase: 03-privacy-layer*
*Completed: 2026-02-24*
