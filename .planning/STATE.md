# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Customers can find available GPUs across providers and provision them instantly through a single interface, with a privacy layer that completely hides the upstream provider.
**Current focus:** Phase 2: Provider Abstraction + RunPod Adapter

## Current Position

Phase: 2 of 7 (Provider Abstraction + RunPod Adapter)
Plan: 2 of 3 in current phase
Status: In Progress
Last activity: 2026-02-24 -- Completed 02-02 (Provider Interface & Registry)

Progress: [███░░░░░░░] 27%

## Performance Metrics

**Velocity:**
- Total plans completed: 6
- Average duration: 1.5min
- Total execution time: 0.15 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation | 4 | 6min | 1.5min |
| 02-provider-abstraction | 2 | 3min | 1.5min |

**Recent Trend:**
- Last 5 plans: 01-02 (2min), 01-03 (2min), 01-04 (1min), 02-01 (1min), 02-02 (2min)
- Trend: stable

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: 7-phase standard-depth structure derived from 65 requirements
- [Roadmap]: RunPod adapter in Phase 2 (early risk validation) -- no API keys yet so unit tests cover interface/registry only
- [Roadmap]: Privacy layer isolated in Phase 3 before instance lifecycle depends on it
- [Roadmap]: Dashboard last (Phase 7) -- API must be stable before building frontend
- [01-01]: Phase 1 Config struct scoped to 4 fields only (Port, DatabaseURL, RedisURL, InternalAPIToken)
- [01-01]: Go version upgraded from 1.22.0 to 1.24.0 (required by pgx/v5 dependency)
- [01-02]: Edit initial migration directly (no production data exists, greenfield)
- [01-02]: TIMESTAMPTZ for all time columns to avoid timezone bugs
- [01-02]: wg_private_key_enc suffix to clarify encryption at rest
- [01-02]: Per-migration transactions with rollback on error
- [01-03]: Redis client used directly (no wrapper) -- go-redis Client sufficient for Phase 1
- [01-03]: Health endpoint behind InternalAuthMiddleware -- prevents unauthenticated probing
- [01-03]: ConnectWithRetry as package-level function in db -- reusable for any service
- [01-03]: ServerDeps struct for constructor injection -- clean dependency boundary for testing
- [01-04]: 404 Not Found for rejected non-loopback IPs -- avoids revealing endpoint existence to scanners
- [01-04]: net.SplitHostPort for IP extraction -- handles both IPv4 and IPv6 RemoteAddr correctly
- [01-04]: LocalhostOnly as outermost middleware -- rejects external IPs before token check
- [02-02]: Two tiers only for v1 (on_demand and spot) -- TierReserved removed per CONTEXT.md
- [02-02]: Async fire-and-return provisioning model -- Provision returns upstream ID, GetStatus polls separately
- [02-02]: WireGuardPrivateKey removed from ProvisionRequest -- key generation deferred to Phase 3
- [02-02]: DatacenterLocation added to GPUOffering for datacenter drill-down info
- [02-02]: Re-registration allowed in registry for config reload scenarios
- [Phase 02-01]: Operations ordered: renames -> constraints -> column drops -> column adds -> triggers
- [Phase 02-01]: CHECK constraint includes expanded state machine: creating, provisioning, booting, running, stopping, terminated, error
- [Phase 02-01]: ON DELETE RESTRICT on instances prevents accidental cascade deletion of billing data

### Pending Todos

None yet.

### Blockers/Concerns

- RunPod Docker initialization requires complete redesign from cloud-init (research finding)
- WireGuard NET_ADMIN capability in RunPod containers unverified -- must validate empirically in Phase 3
- REQUIREMENTS.md stated 55 total requirements but actual count is 65 -- corrected in traceability update

## Session Continuity

Last session: 2026-02-24
Stopped at: Re-executed 02-01-PLAN.md (Schema v1 Improvements) -- updated migration and Go stubs
Resume file: None
