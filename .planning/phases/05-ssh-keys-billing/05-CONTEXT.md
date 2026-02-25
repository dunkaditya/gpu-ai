# Phase 5: SSH Keys + Billing - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

SSH key management (CRUD + injection into instances) and per-second billing with PostgreSQL ledger, Stripe metering, and per-org spending limits. Users manage keys, billing tracks usage automatically, and spending limits protect against bill shock.

</domain>

<decisions>
## Implementation Decisions

### SSH key selection at provisioning
- POST /instances accepts optional list of SSH key IDs
- Smart default: if no key IDs specified, auto-include all of the creating user's keys
- Provisioning blocked if zero SSH keys would be injected — return error "At least one SSH key required"
- No instance creation without SSH access configured

### SSH key limits and validation
- Max 50 SSH keys per org
- Accept RSA, Ed25519, ECDSA key formats only — reject DSA and other legacy types
- Keys stored with user-provided name for identification

### Spending limit enforcement — tiered model
- Opt-in only — no default spending limit for new orgs
- Monthly dollar cap that resets each billing cycle
- Tiered enforcement:
  - **80% of limit** → notify (webhook if configured)
  - **95% of limit** → notify + countdown warning
  - **100% of limit** → stop instances (not terminate — preserve local storage), block new instance creation
  - **+72 hours past 100%** → terminate stopped instances, free resources
- Key distinction: stop preserves state so customers can resume after adding funds; terminate is only the last resort

### Billing ticker architecture
- Single 60-second billing ticker handles both limit checks and Stripe reporting
- Limit enforcement runs first, then Stripe reporting — limit checks must never be blocked by Stripe API latency
- Near real-time spend calculation (every 60s via ticker)

### Billing start/stop timing
- Billing starts at **booting** state (when provider confirms pod is allocated), not at API request time or running state
- Billing stops at **DELETE request** time — customer stops paying the moment they request termination
- Delta between DELETE request and actual provider termination is platform cost to absorb

### Billing edge cases
- Failed provisions (never reached booting): $0 billing record created for audit trail, no charge
- Sub-second rounding: always round up to next second (ceil)
- Per-second usage tracked in PostgreSQL ledger with GPU type, count, duration, and cost

### Usage endpoint presentation
- GET /billing/usage returns per-instance session records by default (start, end, GPU type, duration, cost)
- ?summary=hourly for aggregated hourly view
- Currently-running instances included with real-time estimated cost
- Date filtering: both start/end date params AND preset periods (?period=current_month, ?period=last_30d)
- Costs displayed as per-hour rate with total in USD

### Claude's Discretion
- Stripe Billing Meter event structure and batch format
- Database schema design for billing ledger tables
- Notification delivery mechanism for spending limit warnings (email/webhook internals)
- SSH key parsing and validation implementation details

</decisions>

<specifics>
## Specific Ideas

- Stop vs terminate distinction at spending limit is a differentiator vs budget providers (Vast.ai) that kill instances without warning
- Billing ticker order matters: check limits first → report to Stripe second — never let Stripe latency block enforcement
- The 72-hour retention window after stop gives customers a reasonable safety net to add funds

</specifics>

<deferred>
## Deferred Ideas

- Email notifications for spending limit thresholds — requires email infrastructure (future phase or dashboard phase)
- Dashboard banner for spending limit warnings — Phase 7 (Dashboard)
- Billing alerts/notifications via multiple channels — v2

</deferred>

---

*Phase: 05-ssh-keys-billing*
*Context gathered: 2026-02-25*
