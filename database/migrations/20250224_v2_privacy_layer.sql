-- GPU.ai Database Schema v2 — Privacy Layer Additions
--
-- Applies on top of v1 schema. Changes:
--   PRIV-01: Re-add wg_private_key_enc for encrypted WireGuard key storage
--   PRIV-02: Add UNIQUE constraint on wg_address for tunnel IP uniqueness
--
-- Note: Transaction wrapping is handled by the migration runner (tools/migrate.py).

-- ============================================================
-- PRIV-01: Re-add wg_private_key_enc with proper encryption
-- ============================================================

-- Stores AES-256-GCM encrypted WireGuard private key (hex-encoded nonce+ciphertext).
-- Re-added from SCHEMA-03 removal, now with proper encryption.
ALTER TABLE instances ADD COLUMN wg_private_key_enc TEXT;

-- ============================================================
-- PRIV-02: Add UNIQUE constraint on wg_address
-- ============================================================

-- wg_address column exists from v0 as INET type. Add uniqueness constraint
-- to prevent duplicate tunnel IP assignments during IPAM allocation.
ALTER TABLE instances ADD CONSTRAINT instances_wg_address_unique UNIQUE (wg_address);
