-- GPU.ai Database Schema v1 — Schema Improvements Migration
--
-- Applies on top of v0 schema. Changes:
--   SCHEMA-01: Rename primary keys to self-documenting {table}_id format
--   SCHEMA-02: Add NOT NULL, ON DELETE, CHECK, and UNIQUE constraints
--   SCHEMA-03: Remove wg_private_key_enc security liability
--   SCHEMA-04: Add internal_token, updated_at columns + auto-update trigger
--
-- Note: Transaction wrapping is handled by the migration runner (tools/migrate.py).

-- ============================================================
-- SCHEMA-01: Rename primary keys to self-documenting format
-- ============================================================

ALTER TABLE organizations RENAME COLUMN id TO organization_id;
ALTER TABLE users RENAME COLUMN id TO user_id;
ALTER TABLE ssh_keys RENAME COLUMN id TO ssh_key_id;
ALTER TABLE instances RENAME COLUMN id TO instance_id;
ALTER TABLE environments RENAME COLUMN id TO environment_id;
ALTER TABLE usage_records RENAME COLUMN id TO usage_record_id;

-- Recreate indexes that reference old column names
DROP INDEX IF EXISTS idx_instances_org_id;
DROP INDEX IF EXISTS idx_instances_status;
DROP INDEX IF EXISTS idx_instances_user_id;
DROP INDEX IF EXISTS idx_users_org_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_ssh_keys_user_id;
DROP INDEX IF EXISTS idx_environments_org_id;
DROP INDEX IF EXISTS idx_usage_records_org_id;
DROP INDEX IF EXISTS idx_usage_records_instance_id;

CREATE INDEX idx_instances_org_id ON instances(org_id);
CREATE INDEX idx_instances_status ON instances(status);
CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_users_org_id ON users(org_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_ssh_keys_user_id ON ssh_keys(user_id);
CREATE INDEX idx_environments_org_id ON environments(org_id);
CREATE INDEX idx_usage_records_org_id ON usage_records(org_id);
CREATE INDEX idx_usage_records_instance_id ON usage_records(instance_id);

-- ============================================================
-- SCHEMA-02: Add constraints
-- ============================================================

-- NOT NULL constraints
ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE instances ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE instances ALTER COLUMN user_id SET NOT NULL;

-- Drop existing FK constraints and recreate with explicit ON DELETE behavior.
-- PostgreSQL names auto-generated FK constraints as {table}_{column}_fkey.

-- users.org_id -> ON DELETE CASCADE (org deleted = users deleted)
ALTER TABLE users DROP CONSTRAINT users_org_id_fkey;
ALTER TABLE users ADD CONSTRAINT users_org_id_fkey
    FOREIGN KEY (org_id) REFERENCES organizations(organization_id) ON DELETE CASCADE;

-- instances.org_id -> ON DELETE RESTRICT (can't delete org with active instances)
ALTER TABLE instances DROP CONSTRAINT instances_org_id_fkey;
ALTER TABLE instances ADD CONSTRAINT instances_org_id_fkey
    FOREIGN KEY (org_id) REFERENCES organizations(organization_id) ON DELETE RESTRICT;

-- instances.user_id -> ON DELETE RESTRICT (can't delete user with active instances)
ALTER TABLE instances DROP CONSTRAINT instances_user_id_fkey;
ALTER TABLE instances ADD CONSTRAINT instances_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE RESTRICT;

-- usage_records.instance_id -> ON DELETE RESTRICT (can't delete instance with billing records)
ALTER TABLE usage_records DROP CONSTRAINT usage_records_instance_id_fkey;
ALTER TABLE usage_records ADD CONSTRAINT usage_records_instance_id_fkey
    FOREIGN KEY (instance_id) REFERENCES instances(instance_id) ON DELETE RESTRICT;

-- usage_records.org_id -> ON DELETE RESTRICT (can't delete org with billing records)
ALTER TABLE usage_records DROP CONSTRAINT usage_records_org_id_fkey;
ALTER TABLE usage_records ADD CONSTRAINT usage_records_org_id_fkey
    FOREIGN KEY (org_id) REFERENCES organizations(organization_id) ON DELETE RESTRICT;

-- environments.org_id -> ON DELETE CASCADE (org deleted = envs deleted)
ALTER TABLE environments DROP CONSTRAINT environments_org_id_fkey;
ALTER TABLE environments ADD CONSTRAINT environments_org_id_fkey
    FOREIGN KEY (org_id) REFERENCES organizations(organization_id) ON DELETE CASCADE;

-- environments.user_id -> ON DELETE SET NULL (user deleted but env preserved for org)
ALTER TABLE environments DROP CONSTRAINT environments_user_id_fkey;
ALTER TABLE environments ADD CONSTRAINT environments_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL;

-- CHECK constraint on instances.status (valid state machine values)
ALTER TABLE instances ADD CONSTRAINT instances_status_check
    CHECK (status IN ('creating', 'provisioning', 'booting', 'running', 'stopping', 'terminated', 'error'));

-- UNIQUE constraint on instances.hostname
ALTER TABLE instances ADD CONSTRAINT instances_hostname_unique UNIQUE (hostname);

-- Composite unique index on (upstream_provider, upstream_id) to prevent duplicate upstream references
CREATE UNIQUE INDEX idx_instances_upstream_unique ON instances(upstream_provider, upstream_id);

-- ============================================================
-- SCHEMA-03: Remove wg_private_key_enc (security fix)
-- ============================================================

ALTER TABLE instances DROP COLUMN wg_private_key_enc;

-- ============================================================
-- SCHEMA-04: Add new columns + updated_at trigger
-- ============================================================

-- Per-instance callback authentication token
ALTER TABLE instances ADD COLUMN internal_token VARCHAR(255);

-- Change-tracking timestamp
ALTER TABLE instances ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Trigger function for auto-updating updated_at on row modification
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to instances table
CREATE TRIGGER trg_instances_updated_at
    BEFORE UPDATE ON instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
