---
phase: 05-ssh-keys-billing
plan: 03
subsystem: billing
tags: [stripe, billing, postgres, provisioning, metering]

# Dependency graph
requires:
  - phase: 05-ssh-keys-billing
    provides: "v5 schema with billing_sessions table and spending_limits"
provides:
  - "Billing DB methods: CreateBillingSession, CloseBillingSession, GetActiveBillingSessions, GetBillingSessionsByOrg, UpdateStripeReportedSeconds, GetOrgMonthSpendCents"
  - "BillingService struct with Stripe Billing Meter event reporting"
  - "Engine billing session lifecycle: create at booting, close at termination"
  - "Config fields: StripeAPIKey, StripeMeterEventName"
affects: [05-04, 05-05]

# Tech tracking
tech-stack:
  added: [stripe-go/v82]
  patterns:
    - "Non-fatal billing errors: billing failures logged but never block instance lifecycle"
    - "Zero billing session for audit trail on failed provisions"
    - "No-op service pattern: BillingService disabled when apiKey empty"

key-files:
  created:
    - "internal/db/billing.go"
    - "internal/billing/stripe.go"
  modified:
    - "internal/provision/engine.go"
    - "internal/config/config.go"

key-decisions:
  - "Billing session creation is non-fatal -- instance transitions to booting even if billing insert fails"
  - "CloseBillingSession is idempotent -- returns nil if no open session found"
  - "CEIL rounding on duration_seconds for sub-second billing accuracy"
  - "GetOrgMonthSpendCents returns cents (int64) to avoid float precision issues"
  - "BillingService uses no-op pattern when Stripe not configured (matches WG optional pattern)"

patterns-established:
  - "Non-fatal billing pattern: log error but continue instance lifecycle"
  - "Audit trail pattern: $0 billing sessions for failed provisions"

requirements-completed: [BILL-01, BILL-02, BILL-03, BILL-06]

# Metrics
duration: 4min
completed: 2026-02-25
---

# Phase 5 Plan 03: Billing Session Lifecycle Summary

**Billing session DB layer with 6 methods, engine integration at booting/termination, and Stripe BillingService with stripe-go/v82**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-25T17:50:38Z
- **Completed:** 2026-02-25T17:55:07Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Built billing DB layer with 6 methods covering full session lifecycle (create, close, query active, query by org, update Stripe reported seconds, calculate month spend)
- Wired billing into provisioning engine: session created at booting transition, closed at termination, $0 audit session for failed provisions
- Replaced billing/stripe.go stub with functional BillingService struct supporting Stripe Billing Meter event reporting
- Added StripeAPIKey and StripeMeterEventName to config with optional env var loading

## Task Commits

Each task was committed atomically:

1. **Task 1: Billing DB methods and config fields** - `2e84a36` (feat)
2. **Task 2: Wire billing sessions into engine and build Stripe service struct** - `2c1a347` (feat)

## Files Created/Modified
- `internal/db/billing.go` - BillingSession struct and 6 DB methods for full billing lifecycle
- `internal/billing/stripe.go` - BillingService with NewBillingService, ReportMeterEvents, MeterEventBatch
- `internal/provision/engine.go` - CreateBillingSession at booting, CloseBillingSession at terminate, createZeroBillingSession helper
- `internal/config/config.go` - StripeAPIKey, StripeMeterEventName fields and env var loading

## Decisions Made
- Billing session creation is non-fatal: if the INSERT fails, the instance still transitions to booting. A billing gap is acceptable vs preventing instance from reaching running state.
- CloseBillingSession is idempotent: returns nil if no open session exists, safe for failed provisions that may call close without a session.
- CEIL rounding in SQL ensures sub-second durations round up to the next second (per CONTEXT.md billing edge cases).
- GetOrgMonthSpendCents uses cents (int64) rather than float dollars to avoid precision issues in spending limit checks.
- BillingService follows the same optional-service pattern as WireGuard: no-op when API key not configured.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - Stripe configuration is optional. Set STRIPE_API_KEY and STRIPE_METER_EVENT_NAME environment variables when ready to enable billing metering.

## Next Phase Readiness
- Billing session infrastructure ready for Plan 04 (billing ticker with Stripe reporting and spending limit enforcement)
- Plan 05 (billing API endpoints) can query billing sessions via GetBillingSessionsByOrg
- stripe-go/v82 dependency available for ticker's ReportMeterEvents calls

## Self-Check: PASSED

- FOUND: internal/db/billing.go
- FOUND: internal/billing/stripe.go
- FOUND: internal/provision/engine.go (billing wiring)
- FOUND: internal/config/config.go (Stripe fields)
- FOUND: commit 2e84a36
- FOUND: commit 2c1a347

---
*Phase: 05-ssh-keys-billing*
*Completed: 2026-02-25*
