---
phase: 05-ssh-keys-billing
plan: 04
subsystem: billing
tags: [stripe, spending-limits, ticker, goroutine, billing-metering]

# Dependency graph
requires:
  - phase: 05-03
    provides: Billing session lifecycle, Stripe BillingService, billing DB methods
provides:
  - 60-second billing ticker goroutine with spending limit enforcement
  - Per-org spending limit DB CRUD (GetSpendingLimit, UpsertSpendingLimit, DeleteSpendingLimit, UpdateSpendingLimitFlags, ResetMonthlySpend)
  - Spending limit check at instance provision time (blocks at-limit orgs with 402)
  - StopInstancesForOrg and TerminateStoppedInstancesForOrg engine operations
  - StateStopped state with proper state machine transitions
  - Stripe meter event reporting aggregated by org per tick
affects: [05-05, dashboard, api-handlers]

# Tech tracking
tech-stack:
  added: []
  patterns: [ticker-goroutine-with-context-cancel, spending-limit-threshold-tracking, limits-before-stripe-ordering]

key-files:
  created:
    - internal/billing/ticker.go
    - internal/db/spending_limits.go
  modified:
    - internal/provision/engine.go
    - internal/provision/state.go
    - internal/db/instances.go
    - internal/db/organizations.go
    - internal/api/handlers.go
    - cmd/gpuctl/main.go

key-decisions:
  - "Limits enforced BEFORE Stripe reporting in every tick -- prevents API latency from delaying protection"
  - "StateStopped added to state machine -- stopped preserves storage but suspends billing"
  - "72h auto-terminate after limit reached -- stopped instances terminated after grace period"
  - "Live spend check in checkSpendingLimit -- catches limit even if ticker hasn't run yet"

patterns-established:
  - "Ticker goroutine with context cancellation: Start(ctx) loops on time.Ticker, exits on ctx.Done()"
  - "Threshold flag tracking: notify_80_sent, notify_95_sent prevent duplicate notifications"
  - "Billing cycle reset: automatic monthly reset of spend counters and notification flags"

requirements-completed: [BILL-04, BILL-07]

# Metrics
duration: 3min
completed: 2026-02-25
---

# Phase 05 Plan 04: Billing Ticker & Spending Limits Summary

**60-second billing ticker with per-org spending limit enforcement (80%/95%/100% thresholds, stop/terminate) and Stripe meter event reporting**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-25T17:58:16Z
- **Completed:** 2026-02-25T18:02:04Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- 60-second billing ticker goroutine that enforces spending limits before reporting to Stripe
- Per-org spending limit CRUD operations with threshold notification flag tracking (80%, 95%, 100%)
- Instance stop at 100% limit with 72-hour grace period before auto-termination
- Provision-time spending limit check blocks new instances for at-limit orgs (HTTP 402)
- StateStopped added to state machine allowing running -> stopped -> terminated transitions

## Task Commits

Each task was committed atomically:

1. **Task 1: Spending limit DB methods and billing ticker** - `f14ad09` (feat)
2. **Task 2: Spending limit check at provision and main.go wiring** - `ea2072d` (feat)

**Plan metadata:** (pending) (docs: complete plan)

## Files Created/Modified
- `internal/billing/ticker.go` - BillingTicker with Start(), runTick(), enforceSpendingLimit(), reportToStripe()
- `internal/db/spending_limits.go` - SpendingLimit struct and CRUD: Get, Upsert, Delete, UpdateFlags, ResetMonthlySpend
- `internal/provision/engine.go` - checkSpendingLimit(), StopInstancesForOrg(), TerminateStoppedInstancesForOrg(), ErrSpendingLimitReached
- `internal/provision/state.go` - StateStopped constant, state machine transitions for stopped state
- `internal/db/instances.go` - ListRunningInstancesByOrg(), ListStoppedInstancesByOrg()
- `internal/db/organizations.go` - GetOrgStripeCustomerID()
- `internal/api/handlers.go` - ErrSpendingLimitReached mapped to HTTP 402
- `cmd/gpuctl/main.go` - BillingService creation, BillingTicker wiring with context cancellation

## Decisions Made
- Limits enforced BEFORE Stripe reporting in every tick -- prevents API latency from delaying protection
- StateStopped added to state machine -- stopped preserves storage but suspends billing, distinct from stopping/terminated
- 72h auto-terminate after limit reached -- stopped instances terminated after grace period
- Live spend check in checkSpendingLimit at provision time -- catches limit even if ticker hasn't run yet
- Billing cycle reset detects month rollover by comparing billing_cycle_start to current month start

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Billing ticker operational, spending limits enforced both at tick time and provision time
- Ready for Plan 05 (spending limits API endpoints)
- All `go build ./...` passes clean

## Self-Check: PASSED

All 8 key files verified present. Both task commits (f14ad09, ea2072d) verified in git log.

---
*Phase: 05-ssh-keys-billing*
*Completed: 2026-02-25*
