-- v6: Availability + Health Monitoring tables
-- Phase 6: instance_events for audit log / SSE / future webhooks

BEGIN;

-- Instance events table: audit log, SSE data source, future webhook source.
CREATE TABLE instance_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id VARCHAR(32) NOT NULL REFERENCES instances(instance_id) ON DELETE RESTRICT,
    org_id UUID NOT NULL REFERENCES organizations(organization_id) ON DELETE RESTRICT,
    event_type VARCHAR(50) NOT NULL
        CHECK (event_type IN ('ready', 'interrupted', 'failed', 'terminated')),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_instance_events_instance ON instance_events(instance_id);
CREATE INDEX idx_instance_events_org_created ON instance_events(org_id, created_at DESC);

COMMIT;
