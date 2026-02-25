---
phase: 05-ssh-keys-billing
plan: 01
subsystem: database
tags: [postgres, migration, ssh-keys, billing, spending-limits]

# Dependency graph
requires:
  - phase: 04.3-auth-idempotency-edge-cases
    provides: "v4 schema with email nullable and UNIQUE dropped"
provides:
  - "ssh_keys.org_id column with FK to organizations and UNIQUE(org_id, fingerprint)"
  - "billing_sessions table for per-second usage ledger"
  - "spending_limits table for per-org monthly caps with notification thresholds"
  - "instances.status CHECK constraint with 'stopped' state"
affects: [05-02, 05-03, 05-04, 05-05]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "ON DELETE RESTRICT for billing FK relationships (prevent accidental data loss)"
    - "Partial index for active session queries (WHERE ended_at IS NULL)"
    - "One-row-per-org pattern for spending_limits (UNIQUE on org_id)"

key-files:
  created:
    - "database/migrations/20250225_v5_ssh_keys_billing.sql"
  modified: []

key-decisions:
  - "ON DELETE RESTRICT for billing_sessions FKs to instances and organizations -- prevents accidental cascade deletion of billing records"
  - "ON DELETE CASCADE for ssh_keys and spending_limits org FKs -- keys and limits should be cleaned up with the org"
  - "stripe_reported_seconds column for delta-based Stripe usage metering -- avoids double-reporting"
  - "billing_cycle_start as explicit column rather than derived from timestamps -- enables flexible billing periods"

patterns-established:
  - "Partial index pattern: CREATE INDEX ... WHERE ended_at IS NULL for active session lookups"
  - "Trigger reuse: existing update_updated_at function applied to new tables"

requirements-completed: [SSHK-01, SSHK-02, SSHK-03, SSHK-04, BILL-01, BILL-02, BILL-03, BILL-06, BILL-07]

# Metrics
duration: 1min
completed: 2026-02-25
---

# Phase 5 Plan 01: Database Migration v5 Summary

**PostgreSQL v5 migration adding org-scoped SSH keys, billing_sessions ledger, and spending_limits with notification thresholds**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-25T17:46:45Z
- **Completed:** 2026-02-25T17:47:24Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- SSH keys org-scoping: added org_id FK and UNIQUE(org_id, fingerprint) constraint to ssh_keys table
- Billing infrastructure: created billing_sessions table with per-second usage tracking, Stripe reporting delta column, and partial index for active sessions
- Spending limits: created spending_limits table with monthly caps, notification threshold flags (80%/95%), and update_updated_at trigger
- Instance status expansion: updated CHECK constraint to include 'stopped' state for spending limit enforcement

## Task Commits

Each task was committed atomically:

1. **Task 1: Write schema migration v5** - `dba49a0` (feat)

## Files Created/Modified
- `database/migrations/20250225_v5_ssh_keys_billing.sql` - v5 migration with SSH key org-scoping, billing_sessions, spending_limits, and status CHECK update

## Decisions Made
- ON DELETE RESTRICT chosen for billing_sessions FKs (instances, organizations) to prevent accidental cascade deletion of billing records
- ON DELETE CASCADE for ssh_keys.org_id and spending_limits.org_id since these should be cleaned up when an org is deleted
- stripe_reported_seconds tracks already-reported usage to enable delta-based Stripe metering on each ticker cycle
- billing_cycle_start as explicit column enables flexible billing periods without deriving from timestamps
- Partial index on billing_sessions (WHERE ended_at IS NULL) optimizes active session lookups for the billing ticker

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Schema foundation complete for Phase 5 plans 02 (SSH key CRUD) and 03 (billing engine)
- Migration must be applied to dev database before integration testing
- Go build passes cleanly (no Go file changes in this plan)

## Self-Check: PASSED

- FOUND: database/migrations/20250225_v5_ssh_keys_billing.sql
- FOUND: .planning/phases/05-ssh-keys-billing/05-01-SUMMARY.md
- FOUND: commit dba49a0

---
*Phase: 05-ssh-keys-billing*
*Completed: 2026-02-25*
