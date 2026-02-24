-- GPU.ai Database Schema v3 — Auth & Instance Lifecycle Additions
--
-- Applies on top of v2 schema. Changes:
--   AUTH-01: Add clerk_org_id to organizations for Clerk org mapping
--   AUTH-02: Create idempotency_keys table for request deduplication
--   INST-01: Add name, ready_at, error_reason columns to instances
--   INST-02: Add keyset pagination index on instances(created_at, instance_id)
--
-- Note: Transaction wrapping is handled by the migration runner (tools/migrate.py).

-- ============================================================
-- AUTH-01: Add Clerk organization ID mapping
-- ============================================================

-- Maps Clerk's external organization ID to our internal organization.
-- UNIQUE constraint ensures one-to-one mapping.
ALTER TABLE organizations ADD COLUMN clerk_org_id VARCHAR(255) UNIQUE;

-- ============================================================
-- AUTH-02: Idempotency keys for request deduplication
-- ============================================================

CREATE TABLE idempotency_keys (
    idempotency_key VARCHAR(255) NOT NULL,
    org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE CASCADE,
    request_hash VARCHAR(64) NOT NULL,
    response_code INT,
    response_body JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (org_id, idempotency_key)
);

-- Index for cleanup of expired idempotency keys
CREATE INDEX idx_idempotency_keys_created ON idempotency_keys(created_at);

-- ============================================================
-- INST-01: Add instance lifecycle columns
-- ============================================================

-- Optional user-facing display label for the instance
ALTER TABLE instances ADD COLUMN name VARCHAR(255);

-- Timestamp when instance transitioned to running state
ALTER TABLE instances ADD COLUMN ready_at TIMESTAMPTZ;

-- Human-readable explanation when instance status is 'error'
ALTER TABLE instances ADD COLUMN error_reason TEXT;

-- ============================================================
-- INST-02: Keyset pagination index
-- ============================================================

-- Supports efficient cursor-based pagination over instances ordered by creation time
CREATE INDEX idx_instances_created_at_id ON instances(created_at DESC, instance_id DESC);
