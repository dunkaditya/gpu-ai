---
phase: 04-auth-instance-lifecycle
plan: 04
subsystem: api, infra
tags: [idempotency, wireguard, org-id, uuid, iptables, peer-cleanup]

# Dependency graph
requires:
  - phase: 04-auth-instance-lifecycle/03
    provides: "Idempotency middleware, provisioning engine with WG stub"
  - phase: 03-privacy-layer/02
    provides: "WireGuard manager with RemovePeer, PortFromTunnelIP, IPAM"
provides:
  - "Runtime-correct idempotency middleware resolving Clerk org ID to internal UUID"
  - "WireGuard peer and iptables cleanup on instance termination"
affects: [05-billing-metering, 06-dashboard]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Clerk-to-UUID resolution at middleware boundary before DB operations"
    - "Best-effort WG cleanup on termination (log errors, don't fail)"
    - "CIDR suffix stripping for INET column values before net.ParseIP"

key-files:
  created: []
  modified:
    - internal/api/idempotency.go
    - internal/provision/engine.go

key-decisions:
  - "Resolve Clerk org ID at middleware layer, not DB layer -- keeps DB functions UUID-only"
  - "WG cleanup is best-effort: errors logged but termination succeeds regardless"
  - "Strip CIDR suffix from INET column values before parsing to handle PostgreSQL INET format"

patterns-established:
  - "Clerk-to-UUID resolution: always call GetOrgIDByClerkOrgID before passing org scope to DB functions"
  - "Best-effort cleanup: resource cleanup after terminal state changes logs errors but does not revert the state change"

requirements-completed: [API-11, INST-04, INST-06]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 4 Plan 4: Gap Closure Summary

**Idempotency org_id UUID resolution and WireGuard peer cleanup on instance termination**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T23:58:08Z
- **Completed:** 2026-02-25T00:00:06Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Fixed runtime PostgreSQL UUID type mismatch in idempotency middleware by resolving Clerk org ID to internal UUID via GetOrgIDByClerkOrgID
- Implemented actual WireGuard peer removal on instance termination using RemovePeer with parsed tunnel IP and computed external port
- Added CIDR suffix handling for PostgreSQL INET column values in WG address parsing

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix idempotency middleware org_id type mismatch** - `e8a7650` (fix)
2. **Task 2: Implement WireGuard peer cleanup on instance termination** - `00ec2c8` (feat)

## Files Created/Modified
- `internal/api/idempotency.go` - Added GetOrgIDByClerkOrgID resolution, replaced all claims.OrgID in DB calls with internalOrgID
- `internal/provision/engine.go` - Replaced WG cleanup stub with RemovePeer call, added CIDR suffix stripping for INET addresses, added "net" import

## Decisions Made
- Resolve Clerk org ID at middleware layer (not DB layer) to keep DB functions UUID-only
- WG cleanup is best-effort: errors are logged but do not fail the termination flow
- Strip CIDR suffix from PostgreSQL INET column values using strings.Cut before net.ParseIP

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] CIDR suffix handling for WG address parsing**
- **Found during:** Task 2 (WireGuard peer cleanup)
- **Issue:** Plan used `net.ParseIP(*inst.WGAddress)` directly, but WGAddress from the PostgreSQL INET column includes a CIDR suffix (e.g., "10.0.0.2/16"). `net.ParseIP` returns nil for CIDR-formatted strings.
- **Fix:** Added `strings.Cut(*inst.WGAddress, "/")` to strip the CIDR suffix before parsing
- **Files modified:** internal/provision/engine.go
- **Verification:** `go build` and `go vet` pass
- **Committed in:** 00ec2c8 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Essential for correctness -- without this fix, WG cleanup would silently fail for every instance because the IP would never parse. No scope creep.

## Issues Encountered
None beyond the deviation documented above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 4 is now fully complete with all verification gaps closed
- Idempotency middleware correctly scopes keys by internal org UUID
- WireGuard peers and iptables rules are cleaned up on termination
- Ready for Phase 5 (billing/metering) which depends on correct instance lifecycle

## Self-Check: PASSED

- [x] internal/api/idempotency.go exists
- [x] internal/provision/engine.go exists
- [x] 04-04-SUMMARY.md exists
- [x] Commit e8a7650 exists (Task 1)
- [x] Commit 00ec2c8 exists (Task 2)

---
*Phase: 04-auth-instance-lifecycle*
*Completed: 2026-02-24*
