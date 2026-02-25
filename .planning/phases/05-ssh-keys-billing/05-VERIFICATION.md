---
phase: 05-ssh-keys-billing
verified: 2026-02-25T18:30:00Z
status: passed
score: 13/13 must-haves verified
re_verification: false
---

# Phase 5: SSH Keys & Billing Verification Report

**Phase Goal:** Users can manage SSH keys that are injected into new instances, per-second billing tracks usage accurately in a PostgreSQL ledger with batched reporting to Stripe, and per-org spending limits prevent bill shock
**Verified:** 2026-02-25T18:30:00Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

All must-haves were consolidated across the five plans covering this phase.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can POST /api/v1/ssh-keys with name and public key, stored with fingerprint | VERIFIED | `handleCreateSSHKey` in `internal/api/handlers_ssh_keys.go:39-147`. Validates key, computes SHA256 fingerprint, calls `s.db.CreateSSHKey`. |
| 2 | User can GET /api/v1/ssh-keys and see all keys for their org | VERIFIED | `handleListSSHKeys` in `internal/api/handlers_ssh_keys.go:149-197`. Calls `s.db.ListSSHKeysByOrg`, returns `{"ssh_keys":[...]}`. |
| 3 | User can DELETE /api/v1/ssh-keys/{id} and the key is removed | VERIFIED | `handleDeleteSSHKey` in `internal/api/handlers_ssh_keys.go:199-243`. Returns 204 on success, 404 on not found. |
| 4 | POST /api/v1/instances with empty ssh_key_ids auto-includes all of creating user's keys | VERIFIED | `engine.go:139-153`. Empty `SSHKeyIDs` falls back to `e.db.GetSSHKeysByUserID`. `handlers.go:38-40` confirms SSHKeyIDs is optional in `Validate()`. |
| 5 | POST /api/v1/instances with zero resolvable keys returns "at least one SSH key required" | VERIFIED | `engine.go:152-154`. Returns `ErrSSHKeysNotFound`; `handlers.go:169-173` maps to 400 "ssh-keys-not-found". |
| 6 | RSA, Ed25519, and ECDSA keys accepted; DSA rejected | VERIFIED | `handlers_ssh_keys.go:31-95`. `allowedKeyTypes` map excludes `ssh-dss`. Returns 400 "unsupported_key_type". |
| 7 | Org-level 50-key limit enforced on add | VERIFIED | `handlers_ssh_keys.go:103-116`. Calls `s.db.CountSSHKeysByOrg`; returns 422 "key_limit_exceeded" when `count >= 50`. |
| 8 | A billing_session record is created when an instance transitions provisioning to booting | VERIFIED | `engine.go:680-696`. `CreateBillingSession` called after successful `UpdateInstanceStatus(provisioning->booting)`. |
| 9 | The billing_session record is closed when an instance is terminated | VERIFIED | `engine.go:403-410`. `CloseBillingSession` called after `TerminateInstance` with `time.Now().UTC()`. |
| 10 | Failed provisions create a $0 billing session for audit trail | VERIFIED | `engine.go:715-752`. `createZeroBillingSession` called on timeout (`engine.go:625`) and upstream failure (`engine.go:707`). Creates and immediately closes session with same timestamp. |
| 11 | Config struct has StripeAPIKey and StripeMeterEventName fields | VERIFIED | `config/config.go:59-65`. Both fields present, read from `STRIPE_API_KEY` and `STRIPE_METER_EVENT_NAME` env vars. |
| 12 | A 60-second billing ticker goroutine starts in main.go and runs for the server lifetime | VERIFIED | `main.go:130-140`. `NewBillingService` and `NewBillingTicker` created; `go billingTicker.Start(ctx)` launched with signal context. |
| 13 | Spending limits enforced BEFORE Stripe reporting per tick | VERIFIED | `ticker.go:79-85`. `enforceSpendingLimit` loop (Step 3) runs before `reportToStripe` (Step 4). Comment documents this criticality. |
| 14 | When org reaches 80%/95% of monthly limit, notify flags are set | VERIFIED | `ticker.go:181-209`. 80% and 95% thresholds call `UpdateSpendingLimitFlags` with appropriate flag values. |
| 15 | When org reaches 100% of monthly limit, running instances are stopped | VERIFIED | `ticker.go:158-178`. Calls `t.engine.StopInstancesForOrg`; engine.go:452-473 updates status running->stopped without calling provider.Terminate. |
| 16 | 72 hours after limit_reached_at, stopped instances are terminated | VERIFIED | `ticker.go:142-155`. Checks `time.Since(*limit.LimitReachedAt) > 72*time.Hour`, calls `t.engine.TerminateStoppedInstancesForOrg`. |
| 17 | New instance creation is blocked when org is at spending limit | VERIFIED | `engine.go:163-165`. `checkSpendingLimit` called before provider selection. `handlers.go:174-178` maps to 402 "spending_limit_reached". |
| 18 | Stripe meter events reported every 60s as aggregated GPU-seconds | VERIFIED | `ticker.go:213-296`. `reportToStripe` computes unreported seconds per session, aggregates by org, calls `t.stripe.ReportMeterEvents`. |
| 19 | stripe_reported_seconds updated to prevent double-reporting | VERIFIED | `ticker.go:288-294`. Calls `t.db.UpdateStripeReportedSeconds` after successful Stripe report. |
| 20 | GET /api/v1/billing/usage returns per-instance sessions with real-time cost | VERIFIED | `handlers_billing.go:69-216`. Active sessions compute `EstimatedCost` using `math.Ceil(time.Since(sess.StartedAt).Seconds()) / 3600.0 * price * gpuCount`. |
| 21 | Period presets and RFC3339 date range filters work | VERIFIED | `handlers_billing.go:95-145`. `current_month`, `last_30d` presets; RFC3339 `start`/`end` params; mutually exclusive with 400 on conflict. |
| 22 | Hourly aggregation mode returns HourlyUsageResponse | VERIFIED | `handlers_billing.go:196-209`. `aggregateHourlyBuckets` function (lines 218-289) distributes GPU-seconds across hour boundaries. |
| 23 | PUT/GET/DELETE /api/v1/billing/spending-limit CRUD works | VERIFIED | `handlers_billing.go:291-418`. All three handlers implemented and returning correct status codes (200/200/204). |

**Score:** 23/23 observable truths verified (covers all 13 plan-defined must-haves across 5 plans)

---

### Required Artifacts

| Artifact | Min Lines | Actual Lines | Status | Details |
|----------|-----------|--------------|--------|---------|
| `database/migrations/20250225_v5_ssh_keys_billing.sql` | — | 103 | VERIFIED | All 4 sections: ssh_keys.org_id, instances status check update, billing_sessions, spending_limits. Trigger applied. |
| `internal/db/ssh_keys.go` | 80 | 143 | VERIFIED | 5 DB methods: CreateSSHKey, ListSSHKeysByOrg, DeleteSSHKey, GetSSHKeysByUserID, CountSSHKeysByOrg. ErrDuplicateKey sentinel. |
| `internal/api/handlers_ssh_keys.go` | 60 | 243 | VERIFIED | 3 handlers: handleCreateSSHKey, handleListSSHKeys, handleDeleteSSHKey. Key validation uses golang.org/x/crypto/ssh. |
| `internal/db/billing.go` | 80 | 151 | VERIFIED | 6 DB methods: CreateBillingSession, CloseBillingSession, GetActiveBillingSessions, GetBillingSessionsByOrg, UpdateStripeReportedSeconds, GetOrgMonthSpendCents. |
| `internal/billing/stripe.go` | 40 | 86 | VERIFIED | BillingService struct, NewBillingService, ReportMeterEvents. No-op when disabled. |
| `internal/billing/ticker.go` | 120 | 296 | VERIFIED | BillingTicker with Start(), runTick(), enforceSpendingLimit(), reportToStripe(). |
| `internal/db/spending_limits.go` | 60 | 118 | VERIFIED | 5 DB methods: GetSpendingLimit, UpsertSpendingLimit, DeleteSpendingLimit, UpdateSpendingLimitFlags, ResetMonthlySpend. |
| `internal/api/handlers_billing.go` | 100 | 445 | VERIFIED | 4 handlers: handleGetUsage, handleSetSpendingLimit, handleGetSpendingLimit, handleDeleteSpendingLimit. All response types defined. |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/api/handlers_ssh_keys.go` | `internal/db/ssh_keys.go` | `s.db.CreateSSHKey / ListSSHKeysByOrg / DeleteSSHKey` | WIRED | Lines 126, 174, 227 |
| `internal/provision/engine.go` | `internal/db/ssh_keys.go` | `GetSSHKeysByUserID fallback` | WIRED | Line 147 |
| `internal/provision/engine.go` | `internal/db/billing.go` | `CreateBillingSession at booting transition` | WIRED | Lines 689, 738 |
| `internal/provision/engine.go` | `internal/db/billing.go` | `CloseBillingSession in Terminate` | WIRED | Lines 404, 746 |
| `internal/config/config.go` | `internal/billing/stripe.go` | `Config.StripeAPIKey passed to NewBillingService` | WIRED | `main.go:130` passes `cfg.StripeAPIKey` |
| `internal/billing/ticker.go` | `internal/db/billing.go` | `GetActiveBillingSessions, UpdateStripeReportedSeconds` | WIRED | Lines 61, 289 |
| `internal/billing/ticker.go` | `internal/db/spending_limits.go` | `GetSpendingLimit, UpdateSpendingLimitFlags` | WIRED | Lines 91, 171, 187, 203 |
| `cmd/gpuctl/main.go` | `internal/billing/ticker.go` | `ticker.Start(ctx)` | WIRED | Line 139: `go billingTicker.Start(ctx)` |
| `internal/api/handlers_billing.go` | `internal/db/billing.go` | `GetBillingSessionsByOrg with date filters` | WIRED | Line 148 |
| `internal/api/handlers_billing.go` | `internal/db/spending_limits.go` | `UpsertSpendingLimit / GetSpendingLimit / DeleteSpendingLimit` | WIRED | Lines 326, 362, 402 |
| `internal/api/server.go` | `internal/api/handlers_billing.go` | `4 billing routes registered` | WIRED | Lines 83-90: GET billing/usage, PUT/GET/DELETE billing/spending-limit |
| `billing_sessions` | `instances` | `instance_id FK ON DELETE RESTRICT` | WIRED | Migration line 54: `REFERENCES instances(instance_id) ON DELETE RESTRICT` |
| `billing_sessions` | `organizations` | `org_id FK ON DELETE RESTRICT` | WIRED | Migration line 55: `REFERENCES organizations(organization_id) ON DELETE RESTRICT` |
| `spending_limits` | `organizations` | `org_id FK UNIQUE ON DELETE CASCADE` | WIRED | Migration line 87: `REFERENCES organizations(organization_id) ON DELETE CASCADE` |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| SSHK-01 | 05-01, 05-02 | User can add an SSH public key with a name | SATISFIED | `handleCreateSSHKey` + `CreateSSHKey` DB method |
| SSHK-02 | 05-01, 05-02 | User can list their SSH keys | SATISFIED | `handleListSSHKeys` + `ListSSHKeysByOrg` DB method |
| SSHK-03 | 05-01, 05-02 | User can delete an SSH key | SATISFIED | `handleDeleteSSHKey` + `DeleteSSHKey` DB method |
| SSHK-04 | 05-02 | SSH keys are injected into new instances at provision time | SATISFIED | `engine.go:157-265`: sshPubKeys injected into `SSHAuthorizedKeys` (cloud-init) and `SSHPublicKeys` (provider) |
| BILL-01 | 05-01, 05-03 | Billing starts at instance provision request time | SATISFIED (with clarification) | Plan 05-CONTEXT.md clarifies "billing starts at booting state." Implementation correctly bills from booting (`engine.go:680-696`). The requirement text in REQUIREMENTS.md is imprecise; the phase CONTEXT.md is authoritative. |
| BILL-02 | 05-03 | Billing stops at instance termination time | SATISFIED | `engine.go:403-410`: `CloseBillingSession` called in `Terminate()` with `time.Now().UTC()` |
| BILL-03 | 05-01, 05-03 | Per-second usage tracked in PostgreSQL billing ledger | SATISFIED | `billing_sessions` table with `duration_seconds`, `total_cost`. CEIL-based rounding in `CloseBillingSession`. |
| BILL-04 | 05-04 | Usage batched and reported to Stripe Billing Meters every 60 seconds | SATISFIED | `ticker.go:213-296` aggregates GPU-seconds per org and calls `ReportMeterEvents` every tick |
| BILL-05 | 05-05 | User can view their usage history and costs | SATISFIED | `GET /api/v1/billing/usage` with date filtering, period presets, hourly aggregation |
| BILL-06 | 05-01, 05-03 | Billing records include GPU type, count, duration, and cost | SATISFIED | `billing_sessions` schema: `gpu_type`, `gpu_count`, `duration_seconds`, `total_cost`, `price_per_hour` |
| BILL-07 | 05-01, 05-04 | Configurable per-org spending limit with automatic instance termination | SATISFIED | `spending_limits` table + ticker enforcement: stop at 100%, terminate at 72h |
| API-06 | 05-02 | GET/POST/DELETE /api/v1/ssh-keys manages SSH keys | SATISFIED | 3 routes in `server.go:75-80` behind authChain |
| API-07 | 05-05 | GET /api/v1/billing/usage returns billing history | SATISFIED | Route at `server.go:83-84` behind authChain; full handler in `handlers_billing.go` |

All 13 requirement IDs fully satisfied. No orphaned requirements detected.

---

### Anti-Patterns Found

None detected. Scanned `handlers_billing.go`, `handlers_ssh_keys.go`, `ticker.go`, `billing/stripe.go`, `db/billing.go`, `db/spending_limits.go`, `db/ssh_keys.go`, and `provision/engine.go` for TODO/FIXME/placeholder patterns and empty implementations.

---

### Build and Test Status

- `go build ./...` — PASSES (verified)
- `go test ./...` — PASSES: `internal/api`, `internal/provision`, `internal/provider`, `internal/provider/runpod`, `internal/wireguard`

---

### Human Verification Required

The following items require human verification (cannot be verified programmatically):

#### 1. SSH Key Injection End-to-End

**Test:** Provision a real instance with a registered SSH key. Attempt SSH login to the instance hostname.
**Expected:** SSH key accepted; login succeeds.
**Why human:** Requires a live RunPod pod with cloud-init executing. Cannot verify from code alone.

#### 2. Stripe Meter Event Delivery

**Test:** With `STRIPE_API_KEY` and `STRIPE_METER_EVENT_NAME` configured, run server with an active billing session for 60+ seconds. Check Stripe Dashboard for meter events.
**Expected:** GPU-seconds appear in Stripe Billing Meters for the org's customer ID.
**Why human:** Requires valid Stripe credentials and a live connection; cannot mock from code inspection.

#### 3. Spending Limit Stop Behavior

**Test:** Set a spending limit below current spend. Wait for ticker to fire. Confirm running instances transition to "stopped" (not terminated) in DB.
**Expected:** Instances show status "stopped", data preserved, no provider termination call.
**Why human:** Requires live DB + running instances + real-time observation.

#### 4. 72-Hour Auto-Terminate

**Test:** Manually set `limit_reached_at` to >72 hours ago in DB. Trigger a ticker cycle. Confirm stopped instances are terminated.
**Expected:** `Terminate()` called for each stopped instance; instances reach "terminated" state.
**Why human:** Requires time-travel or DB manipulation in a live environment.

---

### Summary

Phase 5 goal is fully achieved. All required artifacts exist with substantive implementations (no stubs), all key links are wired and verified through code inspection, and the project compiles and passes all tests cleanly.

Notable implementation details confirmed:
- SSH key validation uses `golang.org/x/crypto/ssh` (not regex), correctly rejects DSA
- Billing sessions use CEIL rounding for sub-second accuracy per spec
- Spending limit enforcement runs BEFORE Stripe reporting in each tick (critical order preserved)
- Stripe reporting is a true no-op when `STRIPE_API_KEY` is empty (enabled flag checked)
- Smart SSH key default (auto-include user's keys when none specified) is wired into the provisioning engine

One semantic note: BILL-01 in REQUIREMENTS.md says "billing starts at provision request time" but the phase CONTEXT.md (the authoritative spec for this phase) clarifies billing starts at "booting" state when the provider confirms pod allocation. The implementation follows the CONTEXT.md spec, which is correct behavior.

---

_Verified: 2026-02-25T18:30:00Z_
_Verifier: Claude (gsd-verifier)_
