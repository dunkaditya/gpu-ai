# Phase 6: Availability + Health Monitoring - Context

**Gathered:** 2026-02-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Background systems that continuously poll providers for GPU availability (cached in Redis), expose aggregated offerings via API, select the best-price provider for provisioning, monitor running instances for health and spot interruptions, and notify the dashboard via SSE on instance events. Webhook delivery to external URLs is deferred to a future phase.

</domain>

<decisions>
## Implementation Decisions

### Availability API response
- Flat list of offerings (single array), not grouped — client decides presentation
- Rich fields per offering: gpu_model, vram_gb, cpu_cores, ram_gb, storage_gb, price_per_hour, region, tier, available_count, avg_uptime_pct
- Full server-side filtering via query params: gpu_model, region, tier, min/max price, min_vram
- Pricing shows GPU.ai's markup price only — provider costs are internal, users see our price sheet

### Provider selection logic
- Pure cost optimization: cheapest provider wins
- Tiebreaker: higher-margin provider wins (existing priority list handles this — no new mechanism)
- No health-based exclusion from selection — health monitoring handles failures post-provisioning
- Automatic fallback: if selected provider fails to provision, retry with next cheapest provider (cap at 2-3 attempts)

### Health & interruption handling
- Health monitor polls provider APIs for instance status (no agent/heartbeat on instances)
- Spot interruption: immediately stop billing ticker and fire event notification
- Non-spot failure: retry health check 2-3 times over short window before declaring failure (avoids false positives from transient network blips)
- Instance ready detection: cloud-init callback from the instance itself (already designed in cloud-init template)

### Event notifications (SSE + event logging)
- No webhooks in this phase — SSE for real-time dashboard, webhooks deferred to future phase
- Four event types: ready, interrupted, failed, terminated
- Events logged to `instance_events` table (audit log, SSE data source, future webhook source)
- SSE: one per-org stream (single connection streams all org instance events)
- Rich event payloads: event_type, instance_id, gpu_model, region, created_at, plus event-specific metadata (interruption reason, failure details)
- SSE is live-only, no replay support — REST endpoint `GET /api/v1/events?since=<timestamp>` for catch-up on reconnect
- EventSource handles reconnection; client hits REST endpoint to backfill missed events and merges with live stream

### Claude's Discretion
- Polling interval tuning and backoff strategy
- SSE connection management (keepalive, cleanup of stale connections)
- Exact retry timing for non-spot failure checks
- Event metadata schema details per event type

</decisions>

<specifics>
## Specific Ideas

- Instance events table schema already designed:
  ```sql
  CREATE TABLE instance_events (
      event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      instance_id VARCHAR(32) NOT NULL REFERENCES instances(instance_id) ON DELETE RESTRICT,
      org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT,
      event_type VARCHAR(50) NOT NULL
          CHECK (event_type IN ('ready', 'interrupted', 'failed', 'terminated')),
      metadata JSONB,
      created_at TIMESTAMPTZ DEFAULT NOW()
  );
  CREATE INDEX idx_instance_events_instance ON instance_events(instance_id);
  CREATE INDEX idx_instance_events_org ON instance_events(org_id);
  ```
- SSE for live, REST for catch-up — no server-side event buffering needed, events already in Postgres

</specifics>

<deferred>
## Deferred Ideas

- Webhook delivery to org-configured callback URLs — future phase (will use instance_events table as data source)
- Provider health-based exclusion from selection — not needed for v1, revisit if failure rates become an issue

</deferred>

---

*Phase: 06-availability-health-monitoring*
*Context gathered: 2026-02-25*
