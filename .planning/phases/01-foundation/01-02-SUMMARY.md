---
phase: 01-foundation
plan: 02
subsystem: database
tags: [postgresql, migrations, python, schema, psycopg2, click]

# Dependency graph
requires:
  - phase: none
    provides: "Greenfield project - no prior phase dependencies"
provides:
  - "Complete v0 database schema with TIMESTAMPTZ, pgcrypto, clerk_user_id, WireGuard columns"
  - "Python migration runner (tools/migrate.py) with --status and --target flags"
  - "schema_migrations tracking table pattern for forward-only migrations"
affects: [02-providers, 03-privacy, 04-auth, 05-billing]

# Tech tracking
tech-stack:
  added: [psycopg2-binary, click, tabulate]
  patterns: [forward-only-migrations, schema-migrations-tracking, date-prefixed-migration-files]

key-files:
  created: []
  modified:
    - database/migrations/20250224_v0.sql
    - tools/migrate.py

key-decisions:
  - "Edit initial migration directly since no production data exists (greenfield)"
  - "TIMESTAMPTZ for all time columns to avoid timezone bugs"
  - "wg_private_key_enc suffix to clarify encryption at rest"
  - "Each migration runs in its own transaction with rollback on error"

patterns-established:
  - "Forward-only migrations: never rollback, write new migration to fix mistakes"
  - "schema_migrations table tracks applied versions with TIMESTAMPTZ timestamps"
  - "Migration files scanned via glob sorted by filename (date prefix ensures order)"

requirements-completed: [FOUND-05]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 1 Plan 2: Database Schema & Migration Runner Summary

**PostgreSQL v0 schema with TIMESTAMPTZ, pgcrypto, clerk_user_id, and Python migration runner with --status/--target CLI**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T11:20:58Z
- **Completed:** 2026-02-24T11:22:58Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Fixed all 7 schema gaps between v0.sql and ARCHITECTURE.md reference (TIMESTAMPTZ, pgcrypto, clerk_user_id, ON DELETE CASCADE, WireGuard column names, stripe field name)
- Implemented full Python migration runner with schema_migrations tracking, --status table display, --target version targeting, per-migration transactions with rollback

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix database migration SQL to match architecture reference** - `bffb884` (feat)
2. **Task 2: Implement Python migration runner** - `379932a` (feat)

## Files Created/Modified
- `database/migrations/20250224_v0.sql` - Complete v0 schema with all tables, indexes, TIMESTAMPTZ, pgcrypto, clerk_user_id, correct WireGuard/Stripe column names, ON DELETE CASCADE
- `tools/migrate.py` - Migration runner: connects to DATABASE_URL, creates schema_migrations table, applies pending .sql files in order, supports --status and --target flags

## Decisions Made
- Edited the initial migration directly rather than creating a new migration file, since this is a greenfield project with no deployed databases (safe and simpler)
- Used TIMESTAMPTZ for all 10 time columns to prevent timezone-related bugs across different server timezones
- Named WireGuard private key column `wg_private_key_enc` with `_enc` suffix to clarify it stores encrypted data at rest
- Renamed `stripe_invoice_id` to `stripe_usage_record_id` for semantic correctness (usage records, not invoices)
- Each migration executes in its own transaction so a failure rolls back only that migration, not prior successful ones

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Installed Python dependencies for verification**
- **Found during:** Task 2 (verification step)
- **Issue:** click, psycopg2-binary, tabulate not installed in local Python environment; --help command failed with ModuleNotFoundError
- **Fix:** Ran `pip3 install --break-system-packages -r tools/requirements.txt` to install dependencies
- **Files modified:** None (system packages only)
- **Verification:** `python3 tools/migrate.py --help` runs successfully
- **Committed in:** Not committed (runtime dependency, not a code change)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Minor -- dependency installation required for verification only. No code changes needed.

## Issues Encountered
None -- both tasks executed cleanly.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Database schema is ready for all subsequent phases (auth in Phase 4 has clerk_user_id, billing in Phase 5 has stripe_usage_record_id)
- Migration runner is ready for use: `python tools/migrate.py` to apply, `--status` to check
- Requires PostgreSQL running and DATABASE_URL set before migrations can be applied (Docker Compose from Plan 03 will provide this)

## Self-Check: PASSED

All files verified present on disk. All commit hashes verified in git log.

---
*Phase: 01-foundation*
*Completed: 2026-02-24*
