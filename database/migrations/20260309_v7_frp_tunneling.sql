-- v7: Add FRP tunneling support
-- Adds frp_remote_port column for FRP-based SSH tunnel port mapping.
-- WireGuard columns (wg_public_key, wg_private_key_enc, wg_address) are kept
-- nullable for historical records; they are no longer populated for new instances.

BEGIN;

-- Add FRP remote port column (nullable; only set when FRP tunneling is configured)
ALTER TABLE instances ADD COLUMN frp_remote_port INTEGER;

-- Partial unique constraint: no two active instances may share the same FRP remote port.
-- Terminated and error instances release their ports for reuse.
CREATE UNIQUE INDEX instances_frp_remote_port_active
    ON instances (frp_remote_port)
    WHERE frp_remote_port IS NOT NULL AND status NOT IN ('terminated', 'error');

COMMIT;
