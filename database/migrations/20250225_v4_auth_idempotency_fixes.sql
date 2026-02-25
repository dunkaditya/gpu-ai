-- GPU.ai Database Schema v4 — Auth & Idempotency Edge Case Fixes
--
-- Applies on top of v3 schema. Fixes:
--   AUTH-03: Drop email UNIQUE constraint that blocks multi-user organizations.
--            The email column is informational only — Clerk owns user identity via
--            the clerk_user_id UNIQUE constraint. Multiple users (even with the same
--            email) must be able to coexist across organizations.
--   AUTH-03: Make email nullable since Clerk JWTs may not always include email,
--            and EnsureOrgAndUser currently passes empty string for email.
--   CLEANUP: Drop redundant idx_users_email index (created in v0, recreated in v1).
--            Email lookups are not a hot query path now that Clerk manages identity.
--
-- Note: Transaction wrapping is handled by the migration runner (tools/migrate.py).

-- ============================================================
-- AUTH-03: Drop email UNIQUE constraint
-- ============================================================

-- PostgreSQL auto-names single-column UNIQUE constraints as {table}_{column}_key.
-- This constraint prevents creating a second user with the same (or empty) email,
-- blocking multi-user organizations entirely.
ALTER TABLE users DROP CONSTRAINT users_email_key;

-- ============================================================
-- AUTH-03: Make email nullable
-- ============================================================

-- Clerk handles identity; email is informational. NULL is more correct than empty
-- string when email is not available from JWT claims.
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- ============================================================
-- CLEANUP: Drop redundant email index
-- ============================================================

-- Originally created in v0, recreated in v1 after column rename.
-- No longer needed since email is not used for lookups or uniqueness.
DROP INDEX IF EXISTS idx_users_email;
