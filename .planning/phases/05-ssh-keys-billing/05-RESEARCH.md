# Phase 5: SSH Keys + Billing — Research

**Researched:** 2026-02-25
**Status:** Complete

## 1. Existing Codebase Inventory

### What Already Exists

**SSH Keys:**
- `ssh_keys` table exists in v0 schema: `ssh_key_id UUID PK, user_id UUID FK, name VARCHAR, public_key TEXT, fingerprint VARCHAR, created_at TIMESTAMPTZ`
- `internal/db/ssh_keys.go` has `SSHKey` struct and `GetSSHKeysByIDs()` — used by provisioning engine to resolve SSH key IDs
- `internal/provision/engine.go` already accepts `SSHKeyIDs []string` in `ProvisionRequest`, looks them up via `GetSSHKeysByIDs()`, and injects public keys into cloud-init
- `CreateInstanceRequest` in `handlers.go` already has `SSHKeyIDs []string` field with validation (`ssh_key_ids must not be empty`)
- Index `idx_ssh_keys_user_id` exists

**Billing:**
- `usage_records` table exists in v0 schema: `usage_record_id UUID PK, instance_id VARCHAR FK, org_id UUID FK, period_start TIMESTAMPTZ, period_end TIMESTAMPTZ, gpu_type VARCHAR, gpu_count INT, price_per_hour DECIMAL, total_cost DECIMAL, stripe_usage_record_id VARCHAR`
- `internal/billing/stripe.go` is a stub with TODO comments outlining planned Service struct
- `instances` table has `billing_start TIMESTAMPTZ` and `billing_end TIMESTAMPTZ` columns
- `TerminateInstance()` in `db/instances.go` sets `billing_end = NOW()` — termination stops billing timestamp
- `organizations` table has `stripe_customer_id VARCHAR` column

**NOT existing yet (must build):**
- SSH key CRUD handlers (add, list, delete)
- SSH key validation (format, fingerprint generation)
- SSH key org-scoping (current `ssh_keys` table references `user_id`, not `org_id`)
- Billing ledger (usage_records has wrong shape for per-second tracking)
- Billing ticker (60s background goroutine)
- Stripe Billing Meters integration
- Spending limit tables and enforcement
- Usage API endpoint
- Config entries for Stripe API key

### Codebase Patterns to Follow

1. **Constructor injection via Deps structs:** `ServerDeps`, `EngineDeps` pattern — new services follow same pattern
2. **Database access:** `internal/db/` with method receivers on `*Pool`, using `pgx.Row` scanning
3. **HTTP handlers:** methods on `*Server`, using `auth.ClaimsFromContext()` for auth, `writeProblem()` for errors, `writeJSON()` for success
4. **Middleware chain:** ClerkAuth -> RequireOrg -> RateLimiter for customer endpoints
5. **Pagination:** Keyset cursor with `ParsePageParams()` / `EncodeCursor()` / `DecodeCursor()`
6. **Org scoping:** All customer queries scope by `org_id` — look up via `GetOrgIDByClerkOrgID()`
7. **Background goroutines:** Pattern from `progressStatus()` — context.Background(), ticker, timeout
8. **Config loading:** `config.Load()` with optional env vars and validation

## 2. Schema Design

### SSH Keys — Minimal Changes Needed

The existing `ssh_keys` table is almost right. Key issues:
- Missing `org_id` column — need org-scoping for list/delete (user can see all org keys, not just their own? CONTEXT.md says max 50 per org)
- Missing key type validation at DB level
- The FK `user_id -> users(user_id) ON DELETE CASCADE` already exists from v0

**Migration additions:**
- Add `org_id UUID NOT NULL` FK to `ssh_keys` (matches org-scoping pattern in instances)
- Add `UNIQUE(org_id, fingerprint)` to prevent duplicate keys within org
- Add org-level count index for the 50-key limit enforcement

### Billing Ledger — Needs Redesign

The existing `usage_records` table uses period-based billing (period_start/end). CONTEXT.md specifies:
- Per-second billing with billing start at **booting** state (not API request time — CONTEXT.md overrides success criteria #2)
- Records tied to instance sessions, not arbitrary periods
- Need GPU-seconds tracking for Stripe reporting

**New approach — billing_sessions table:**

```sql
CREATE TABLE billing_sessions (
    billing_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id VARCHAR(12) NOT NULL REFERENCES instances(instance_id) ON DELETE RESTRICT,
    org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE RESTRICT,
    gpu_type VARCHAR(50) NOT NULL,
    gpu_count INT NOT NULL,
    price_per_hour DECIMAL(10, 6) NOT NULL,  -- 6 decimal places for per-second precision
    started_at TIMESTAMPTZ NOT NULL,          -- set when instance reaches 'booting'
    ended_at TIMESTAMPTZ,                     -- set on termination
    duration_seconds BIGINT,                  -- computed on close: ceil(extract(epoch from ended_at - started_at))
    total_cost DECIMAL(12, 6),                -- computed on close
    stripe_reported_seconds BIGINT DEFAULT 0, -- how many GPU-seconds already reported to Stripe
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Why a new table instead of altering usage_records:**
- `usage_records` has the wrong semantics (period-based, not session-based)
- Clean schema is better than patching wrong shape
- Old `usage_records` table can remain for future use or be dropped

### Spending Limits

```sql
CREATE TABLE spending_limits (
    spending_limit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE REFERENCES organizations(organization_id) ON DELETE CASCADE,
    monthly_limit_cents BIGINT NOT NULL,       -- cents to avoid float issues
    current_month_spend_cents BIGINT DEFAULT 0, -- cached, updated by ticker
    billing_cycle_start TIMESTAMPTZ NOT NULL,   -- resets monthly
    notify_80_sent BOOLEAN DEFAULT FALSE,
    notify_95_sent BOOLEAN DEFAULT FALSE,
    limit_reached_at TIMESTAMPTZ,               -- when 100% was hit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## 3. Stripe Billing Meters Integration

### Stripe Billing Meters API (2024+)

Stripe Billing Meters is the modern replacement for metered billing. Key concepts:

1. **Meter** — defines what you're measuring (GPU-seconds). Created once in Stripe dashboard or API.
2. **Meter Events** — individual usage reports sent via `POST /v2/billing/meter_events`
3. **Event payload:** `{ event_name: "gpu_seconds", payload: { stripe_customer_id: "cus_xxx", value: "3600" }, timestamp: unix_ts }`

### Implementation approach:

- **Go SDK:** Use `github.com/stripe/stripe-go/v82` — provides typed access to Stripe API
- **Meter Event batching:** Collect GPU-seconds per-org over 60s ticker interval, send as single event per org
- **Idempotency:** Stripe meter events use `identifier` field for deduplication — use `{org_id}:{timestamp_bucket}` as identifier
- **Error handling:** Failed meter reports should be retried; track `stripe_reported_seconds` in billing_sessions to avoid double-reporting

### Config additions needed:
- `STRIPE_API_KEY` — Stripe secret key (required for billing)
- `STRIPE_METER_EVENT_NAME` — the meter event name configured in Stripe (e.g., "gpu_seconds")
- Both optional — billing features disabled if not configured (matches WG optional pattern)

## 4. Billing Ticker Architecture

Per CONTEXT.md: Single 60-second ticker handles both limit checks and Stripe reporting. Limit enforcement runs first.

### Ticker flow:
```
Every 60 seconds:
  1. Query all active billing sessions (ended_at IS NULL)
  2. For each org with active sessions:
     a. Calculate current spend: sum(elapsed_seconds * price_per_second * gpu_count)
     b. Check spending limit thresholds (80%, 95%, 100%)
     c. If 100% hit: stop instances (not terminate), block new creation
     d. If +72h past 100%: terminate stopped instances
  3. For each active session:
     a. Calculate unreported GPU-seconds: elapsed - stripe_reported_seconds
     b. Batch by org, send to Stripe Billing Meters
     c. Update stripe_reported_seconds
```

### Key implementation details:
- **Limit checks before Stripe:** Prevents Stripe API latency from delaying enforcement
- **Stop vs terminate:** Add `StateStopped = "stopped"` to state machine — preserves local storage. Add to DB CHECK constraint.
- **Instance creation blocking:** Engine.Provision must check spending limit before proceeding
- **72-hour auto-terminate:** Ticker checks `limit_reached_at` and terminates if > 72h

### New state machine addition:
- Add `stopped` state between `running` and `terminated`
- Transitions: `running -> stopped` (spending limit), `stopped -> running` (limit cleared), `stopped -> terminated` (72h expiry)
- External state mapping: `stopped` -> `"stopped"` (customer-visible)

## 5. SSH Key Validation

### Supported formats (from CONTEXT.md):
- **RSA:** `ssh-rsa AAAA...`
- **Ed25519:** `ssh-ed25519 AAAA...`
- **ECDSA:** `ecdsa-sha2-nistp{256,384,521} AAAA...`

### Validation approach:
- Parse with `golang.org/x/crypto/ssh` — `ssh.ParseAuthorizedKey()` handles all standard formats
- Reject if key type not in allowed list (blocks DSA, etc.)
- Compute fingerprint: `ssh.FingerprintSHA256(pubKey)` — returns `SHA256:base64...`
- Already in dependencies: `golang.org/x/crypto v0.43.0` (used by wireguard)

### Smart default at provisioning (from CONTEXT.md):
- If no `ssh_key_ids` in POST /instances, auto-include ALL of creating user's keys
- If zero keys would be injected, return error
- This changes current validation: `ssh_key_ids` becomes optional (not required), but at least one key must resolve

## 6. API Endpoints

### SSH Key Endpoints (API-06):
```
POST   /api/v1/ssh-keys          — Add SSH key
GET    /api/v1/ssh-keys          — List SSH keys (paginated)
DELETE /api/v1/ssh-keys/{id}     — Delete SSH key
```

### Billing Endpoint (API-07):
```
GET    /api/v1/billing/usage     — Usage history
  ?summary=hourly               — Aggregated hourly view
  ?start=RFC3339&end=RFC3339    — Date range filter
  ?period=current_month         — Preset period filter
  ?period=last_30d              — Preset period filter
```

### Spending Limit Endpoints (from CONTEXT.md):
```
PUT    /api/v1/billing/spending-limit  — Set/update spending limit
GET    /api/v1/billing/spending-limit  — Get current spending limit
DELETE /api/v1/billing/spending-limit  — Remove spending limit
```

## 7. Integration Points

### Billing start timing (CONTEXT.md decision):
- Billing starts at **booting** state — when provider confirms pod is allocated
- This means: in `progressStatus()`, when transitioning `provisioning -> booting`, create a billing_session record
- NOT at instance creation time (overrides success criteria #2 wording)

### Billing end timing (CONTEXT.md decision):
- Billing stops at DELETE request time — `TerminateInstance()` already sets `billing_end = NOW()`
- Need to also close the billing_session record at the same time

### Failed provisions:
- If instance never reaches booting: create $0 billing session for audit trail
- Can be handled in `SetInstanceError()` — if no billing_session exists, create one with zero cost

### SSH key injection changes:
- Current: `SSHKeyIDs` required in `CreateInstanceRequest`
- New: `SSHKeyIDs` optional; if empty, auto-include user's keys
- Need `GetSSHKeysByUserID()` DB method

## 8. Dependency Additions

New Go dependencies needed:
- `github.com/stripe/stripe-go/v82` — Stripe API client for billing meters
- No other new deps needed (golang.org/x/crypto already present for SSH key parsing)

## 9. Risk Assessment

### Low Risk:
- SSH key CRUD — straightforward, table already exists, patterns well established
- SSH key validation — `golang.org/x/crypto/ssh` is battle-tested
- Usage endpoint — read-only query with pagination

### Medium Risk:
- Billing ticker correctness — per-second accuracy, rounding, edge cases at month boundaries
- Spending limit enforcement — must not have gaps where instances run unbilled
- Stripe meter event reliability — network failures, retries, idempotency

### Mitigation:
- Billing session records as source of truth (not Stripe)
- Stripe reporting is async and best-effort (catch up on next tick)
- Spending limit checks use cached values (updated every 60s), not real-time Stripe queries

## RESEARCH COMPLETE

---
*Phase: 05-ssh-keys-billing*
*Research completed: 2026-02-25*
