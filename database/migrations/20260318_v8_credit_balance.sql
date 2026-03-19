-- v8: Credit Balance Billing System
-- Adds org_balances, transactions ledger, and credit_codes tables.

-- org_balances: one row per org, tracks prepaid credit balance and auto-pay config.
-- No CHECK (balance_cents >= 0) — allow slight overdraft from 60s tick window.
CREATE TABLE org_balances (
    org_id UUID PRIMARY KEY REFERENCES organizations(organization_id) ON DELETE CASCADE,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    auto_pay_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_pay_threshold_cents BIGINT NOT NULL DEFAULT 0,
    auto_pay_amount_cents BIGINT NOT NULL DEFAULT 0,
    auto_pay_last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- transactions: immutable ledger of all balance changes.
CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE RESTRICT,
    type VARCHAR(30) NOT NULL CHECK (type IN (
        'credit_purchase', 'auto_pay', 'credit_code', 'usage_deduction', 'adjustment'
    )),
    amount_cents BIGINT NOT NULL,
    balance_after_cents BIGINT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    reference_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_org_created ON transactions(org_id, created_at DESC);
CREATE INDEX idx_transactions_reference ON transactions(reference_id) WHERE reference_id IS NOT NULL;

-- credit_codes: redeemable promo codes (GPU-XXXX-XXXX format).
CREATE TABLE credit_codes (
    credit_code_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL UNIQUE,     -- SHA-256 hash of GPU-XXXX-XXXX
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    redeemed_at TIMESTAMPTZ,
    redeemed_by_org_id UUID REFERENCES organizations(organization_id),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
