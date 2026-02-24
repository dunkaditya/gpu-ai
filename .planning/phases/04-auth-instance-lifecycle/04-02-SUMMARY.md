---
phase: 04-auth-instance-lifecycle
plan: 02
subsystem: provision
tags: [state-machine, db-crud, pagination, optimistic-locking, idempotency, provisioning-engine, wireguard]

# Dependency graph
requires:
  - phase: 04-auth-instance-lifecycle
    provides: "Clerk auth middleware, RFC 7807 errors, pagination utils, v3 schema migration"
  - phase: 03-privacy-layer
    provides: "WireGuard key encryption, IPAM, Manager, cloud-init template rendering"
  - phase: 02-provider-abstraction-runpod-adapter
    provides: "Provider interface, Registry, ProvisionRequest/ProvisionResult types"
provides:
  - "Instance state machine with 7 states and transition validation (CanTransition, ExternalState, IsTerminal)"
  - "Org-scoped instance CRUD with optimistic locking and keyset pagination"
  - "Organization/user auto-provisioning from Clerk IDs (EnsureOrgAndUser)"
  - "Idempotency key storage with conflict detection and TTL cleanup"
  - "SSH key batch lookup (GetSSHKeysByIDs)"
  - "Provisioning engine with Provision/Terminate orchestration"
affects: [04-03, 05-billing, 06-dashboard]

# Tech tracking
tech-stack:
  added: []
  patterns: [optimistic-locking, keyset-pagination, state-machine-transitions, engine-deps-injection]

key-files:
  created:
    - internal/provision/state.go
    - internal/db/idempotency.go
  modified:
    - internal/db/instances.go
    - internal/db/organizations.go
    - internal/db/ssh_keys.go
    - internal/provision/engine.go

key-decisions:
  - "State machine allows stopping from any non-terminal state (creating through running) and retry-terminate from error"
  - "EnsureOrgAndUser uses clerk_org_id as organization name placeholder for auto-creation"
  - "Provisioning engine uses EngineDeps struct injection matching project's ServerDeps pattern"
  - "WireGuard is fully conditional in engine: nil wgMgr skips key gen, IPAM, and cloud-init"
  - "Provider selection iterates registry with first-match (Phase 6 adds best-price)"
  - "isDuplicateKeyError checks SQLState 23505 via errors.As for pgx compatibility"

patterns-established:
  - "Optimistic locking: WHERE status = $old for atomic state transitions with race protection"
  - "Keyset pagination: (created_at, instance_id) < ($cursor, $cursorID) ORDER BY DESC"
  - "Engine deps injection: EngineDeps struct with optional nil fields for conditional features"
  - "Instance ID format: gpu-{8 hex chars} for branded identification"

requirements-completed: [AUTH-03, INST-01, INST-02, INST-03, INST-04, INST-05, INST-06, INST-08, API-10]

# Metrics
duration: 4min
completed: 2026-02-24
---

# Phase 4 Plan 02: Instance State Machine, DB Layer, and Provisioning Engine Summary

**Instance state machine with 7-state lifecycle, org-scoped DB CRUD with optimistic locking and keyset pagination, idempotency key storage, and full provisioning engine orchestrating provider calls, WireGuard, and DB persistence**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-24T23:29:19Z
- **Completed:** 2026-02-24T23:33:08Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Instance state machine with 7 states (creating, provisioning, booting, running, stopping, terminated, error), valid transition map, external state mapping, and terminal state detection
- Full org-scoped instance CRUD: create, get, get-for-org, list with keyset pagination, update with optimistic locking, terminate (idempotent), set error
- Organization/user auto-provisioning from Clerk IDs with upsert semantics for seamless first-call onboarding
- Idempotency key storage with duplicate detection (PG 23505), completion update, and TTL cleanup
- Provisioning engine orchestrating: ID generation, SSH key resolution, provider selection, WireGuard key gen, cloud-init rendering, provider API call, DB persistence, async status progression
- Terminate flow with optimistic locking, provider termination, DB marking, and error state fallback

## Task Commits

Each task was committed atomically:

1. **Task 1: State machine, DB instance CRUD, org queries, and idempotency storage** - `1fe4f87` (feat)
2. **Task 2: Provisioning engine and SSH key lookup** - `ca3854b` (feat)

## Files Created/Modified
- `internal/provision/state.go` - Instance state machine: 7 states, CanTransition, ExternalState, IsTerminal
- `internal/db/instances.go` - Full Instance struct and 8 CRUD methods with org scoping, optimistic locking, keyset pagination
- `internal/db/organizations.go` - Organization/User structs, GetOrganization, EnsureOrgAndUser, GetOrgIDByClerkOrgID
- `internal/db/idempotency.go` - IdempotencyKey struct, Get/Create/Complete/Cleanup methods with duplicate detection
- `internal/db/ssh_keys.go` - SSHKey struct and GetSSHKeysByIDs batch lookup using ANY($1)
- `internal/provision/engine.go` - Engine struct with EngineDeps injection, Provision and Terminate methods, provider selection

## Decisions Made
- State machine allows `stopping` from any non-terminal state and `stopping` from `error` for retry-terminate scenarios per CONTEXT.md guidance
- EnsureOrgAndUser stores Clerk org ID as the organization name placeholder during auto-creation; proper org naming deferred to org management features
- Engine uses EngineDeps struct (not functional options) matching the project's existing ServerDeps constructor injection pattern
- WireGuard operations are fully conditional: when wgMgr is nil, engine skips key generation, IPAM allocation, and cloud-init rendering
- Provider selection is first-match from registry (best-price optimization deferred to Phase 6 as planned)
- isDuplicateKeyError uses errors.As with SQLState() interface for pgx error detection instead of string matching

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required. All dependencies were established in Plan 01.

## Next Phase Readiness
- State machine and DB layer ready for handler wiring in Plan 03
- Provisioning engine ready to be called from HTTP handlers
- Idempotency storage ready for POST /instances deduplication
- Organization auto-provisioning ready for Clerk auth middleware integration

## Self-Check: PASSED

All 6 artifact files verified present on disk. Both task commits (1fe4f87, ca3854b) verified in git log.

---
*Phase: 04-auth-instance-lifecycle*
*Completed: 2026-02-24*
