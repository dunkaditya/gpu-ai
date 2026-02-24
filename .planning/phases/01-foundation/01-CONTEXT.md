# Phase 1: Foundation - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

A running Go binary (`gpuctl`) with verified database and Redis connectivity, environment config loading, applied schema, and a health endpoint proving the system is alive. Internal endpoints are secured with a shared secret token. Docker Compose provides local dev dependencies.

</domain>

<decisions>
## Implementation Decisions

### Local Dev Setup
- Docker Compose file with Postgres and Redis for one-command dev setup (`docker compose up`)
- Developer runs `docker compose up -d` then `go run ./cmd/gpuctl`

### Startup Failure Behavior
- Retry connecting to Postgres and Redis with exponential backoff before giving up
- Env vars only — no auto-loading of `.env` files (developer uses `source .env` or Docker sets them)

### Migration Workflow
- Migrations handled by Python `tools/migrate.py` — run separately, NOT auto-applied by gpuctl
- gpuctl expects the schema to already exist when it starts
- Date-prefixed migration files (keep existing pattern: `20250224_v0.sql`, `20250301_v1.sql`)
- Forward-only migrations — no rollbacks, write a new migration to fix mistakes

### Internal Endpoint Security
- `/internal/*` endpoints secured with `INTERNAL_API_TOKEN` header (shared secret, already in `.env.example`)
- `/health` endpoint also behind the internal token (not publicly accessible)
- Health returns detailed JSON: `{"status": "ok", "db": "connected", "redis": "connected"}`

### Claude's Discretion
- Which env vars are required vs optional for Phase 1 (DB + Redis clearly required; Clerk, Stripe, RunPod deferred)
- Whether to reject startup if `INTERNAL_API_TOKEN` is still set to the default `change-me` value
- Startup behavior when required config is missing (crash with clear error listing what's missing is recommended)
- Structured logging format and verbosity at startup

</decisions>

<specifics>
## Specific Ideas

- Control plane architecture: gpuctl is the orchestrator, providers (RunPod etc.) manage actual compute. No Kubernetes needed.
- Docker Compose should match the connection strings in `.env.example` (same ports, database name, credentials)

</specifics>

<deferred>
## Deferred Ideas

- Kubernetes deployment for HA of the control plane itself — production deployment concern, not dev milestone
- Auto-migration on startup — intentionally deferred in favor of explicit `tools/migrate.py` workflow

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-02-24*
