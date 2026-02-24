---
phase: 02-provider-abstraction-runpod-adapter
plan: 01
subsystem: database
tags: [postgres, schema-migration, constraints, triggers]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: "v0 schema with organizations, users, ssh_keys, instances, environments, usage_records tables"
provides:
  - "v1 schema with self-documenting PKs ({table}_id naming convention)"
  - "NOT NULL and ON DELETE constraints preventing orphaned records"
  - "CHECK constraint on instances.status for state machine enforcement"
  - "UNIQUE constraint on instances.hostname"
  - "Composite unique index on (upstream_provider, upstream_id)"
  - "internal_token column for per-instance callback auth"
  - "updated_at column with auto-update trigger on instances"
  - "wg_private_key_enc removed (security fix)"
affects: [03-privacy-wireguard-overlay, 04-instance-lifecycle, 05-billing-metering]

# Tech tracking
tech-stack:
  added: []
  patterns: ["{table}_id PK naming convention", "update_updated_at trigger function pattern"]

key-files:
  created:
    - database/migrations/20250224_v1_schema_improvements.sql
  modified:
    - internal/db/instances.go
    - internal/db/organizations.go
    - internal/db/ssh_keys.go

key-decisions:
  - "Operations ordered: renames -> constraints -> column drops -> column adds -> triggers"
  - "CHECK constraint includes expanded state machine: creating, provisioning, booting, running, stopping, terminated, error"
  - "ON DELETE RESTRICT on instances prevents accidental cascade deletion of billing data"

patterns-established:
  - "{table}_id PK naming: All tables use self-documenting primary key column names"
  - "update_updated_at trigger: Reusable trigger function for auto-updating timestamps"

requirements-completed: [SCHEMA-01, SCHEMA-02, SCHEMA-03, SCHEMA-04]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 2 Plan 1: Schema v1 Improvements Summary

**v1 schema migration with self-documenting PKs, NOT NULL/ON DELETE constraints, status CHECK, hostname UNIQUE, and wg_private_key_enc removal**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T17:57:40Z
- **Completed:** 2026-02-24T17:59:28Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Created v1 schema migration covering all 4 SCHEMA requirements (01-04)
- Renamed all 6 table primary keys to self-documenting {table}_id format
- Added NOT NULL, ON DELETE, CHECK, and UNIQUE constraints for data integrity
- Removed wg_private_key_enc security liability from instances table
- Added internal_token and updated_at columns with auto-update trigger
- Updated all Go query helper stubs to reflect new column naming convention

## Task Commits

Each task was committed atomically:

1. **Task 1: Create v1 schema migration SQL** - `43f9424` (feat)
2. **Task 2: Update Go query helpers for new column names** - `9020d9d` (feat)

## Files Created/Modified
- `database/migrations/20250224_v1_schema_improvements.sql` - v1 schema migration with all improvements
- `internal/db/instances.go` - Updated Instance struct and query stubs with new PK names, removed wg_private_key_enc, added internal_token/updated_at
- `internal/db/organizations.go` - Updated Organization/User structs with new PK names
- `internal/db/ssh_keys.go` - Updated SSHKey struct with new PK name

## Decisions Made
- Ordered migration operations carefully: renames first, then constraint changes, then column drops, then column adds, then triggers -- prevents FK reference errors during migration
- CHECK constraint includes expanded status values (provisioning, booting, error) beyond original v0 values -- matches full state machine from architecture
- Used ON DELETE RESTRICT (not CASCADE) for instances FK constraints to prevent accidental data loss of billing records

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- v1 schema is ready for provider interface definition (02-02) and RunPod adapter (02-03)
- All Go db stubs reference correct column names for when query implementations are built in Phase 4
- internal_token column ready for per-instance callback authentication in provisioning flow

## Self-Check: PASSED

All files verified present. All commits verified in git log.

---
*Phase: 02-provider-abstraction-runpod-adapter*
*Completed: 2026-02-24*
