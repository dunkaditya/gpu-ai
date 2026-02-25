-- GPU.ai Database Schema v5 — SSH Keys Org-Scoping & Billing Infrastructure
--
-- Applies on top of v4 schema. Changes:
--   SSHK-01: Add org_id to ssh_keys for org-level key management
--   SSHK-02: Unique fingerprint per org (prevent duplicate keys within same org)
--   BILL-01: Update instances.status CHECK to include 'stopped' state
--   BILL-03: Create billing_sessions table for per-second usage ledger
--   BILL-06: Create spending_limits table for per-org monthly caps
--   BILL-07: Spending limit notification tracking columns (80%, 95% thresholds)
--
-- Note: Transaction wrapping is handled by the migration runner (tools/migrate.py).

-- ============================================================
-- SSHK-01: Add org_id to ssh_keys for org-level key management
-- ============================================================

-- SSH keys are scoped to organizations, not individual users.
-- This enables org-wide key management and sharing across org members.
ALTER TABLE ssh_keys ADD COLUMN org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE CASCADE;

-- ============================================================
-- SSHK-02: Unique fingerprint per org (prevent duplicate keys)
-- ============================================================

-- Prevent the same key from being added twice within an organization.
-- Different orgs may use the same key (e.g., user belongs to multiple orgs).
ALTER TABLE ssh_keys ADD CONSTRAINT ssh_keys_org_fingerprint_unique UNIQUE (org_id, fingerprint);

-- Index for fast org-scoped key listing (GET /ssh-keys)
CREATE INDEX idx_ssh_keys_org_id ON ssh_keys (org_id);

-- ============================================================
-- BILL-01: Update instances.status CHECK to include 'stopped'
-- ============================================================

-- The 'stopped' state is needed for spending limit enforcement:
-- at 100% of limit, instances are stopped (not terminated) to preserve
-- local storage and allow customers to resume after adding funds.
ALTER TABLE instances DROP CONSTRAINT instances_status_check;
ALTER TABLE instances ADD CONSTRAINT instances_status_check
    CHECK (status IN ('creating', 'provisioning', 'booting', 'running', 'stopping', 'stopped', 'terminated', 'error'));

-- ============================================================
-- BILL-03: billing_sessions — per-second usage ledger
-- ============================================================

-- Each billing session tracks a single instance's billable period.
-- Billing starts at 'booting' state and ends at DELETE request time.
-- duration_seconds and total_cost are computed at session close.
-- stripe_reported_seconds tracks what has been reported to Stripe's
-- usage meter, enabling delta-based reporting on each ticker cycle.
CREATE TABLE billing_sessions (
    billing_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id VARCHAR(12) NOT NULL REFERENCES instances(instance_id) ON DELETE RESTRICT,
    org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE RESTRICT,
    gpu_type VARCHAR(50) NOT NULL,
    gpu_count INT NOT NULL,
    price_per_hour DECIMAL(10, 6) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_seconds BIGINT,
    total_cost DECIMAL(12, 6),
    stripe_reported_seconds BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for listing sessions by org (GET /billing/usage)
CREATE INDEX idx_billing_sessions_org_id ON billing_sessions (org_id);

-- Index for looking up sessions by instance (billing start/stop)
CREATE INDEX idx_billing_sessions_instance_id ON billing_sessions (instance_id);

-- Partial index for finding active (open) sessions per org (ticker queries)
CREATE INDEX idx_billing_sessions_active ON billing_sessions (org_id) WHERE ended_at IS NULL;

-- ============================================================
-- BILL-06/BILL-07: spending_limits — per-org monthly caps
-- ============================================================

-- One spending limit record per organization (opt-in).
-- monthly_limit_cents: the dollar cap in cents for the billing cycle.
-- current_month_spend_cents: running total updated by the billing ticker.
-- notify_80_sent / notify_95_sent: prevent duplicate notifications per cycle.
-- limit_reached_at: timestamp when 100% was hit (triggers 72h countdown).
CREATE TABLE spending_limits (
    spending_limit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE REFERENCES organizations(organization_id) ON DELETE CASCADE,
    monthly_limit_cents BIGINT NOT NULL,
    current_month_spend_cents BIGINT NOT NULL DEFAULT 0,
    billing_cycle_start TIMESTAMPTZ NOT NULL,
    notify_80_sent BOOLEAN NOT NULL DEFAULT FALSE,
    notify_95_sent BOOLEAN NOT NULL DEFAULT FALSE,
    limit_reached_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Apply the existing update_updated_at trigger to spending_limits
-- (trigger function created in v1 migration)
CREATE TRIGGER spending_limits_updated_at
    BEFORE UPDATE ON spending_limits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
