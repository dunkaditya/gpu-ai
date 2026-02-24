# GPU.ai

## What This Is

GPU.ai is a GPU cloud aggregation platform that provides a unified interface for provisioning GPU instances across multiple upstream providers. Customers interact with GPU.ai as a single cloud provider — they never see the upstream source. A single Go binary (`gpuctl`) handles all backend operations, paired with a Next.js customer dashboard.

## Core Value

Customers can find available GPUs across providers and provision them instantly through a single interface, with a privacy layer that completely hides the upstream provider.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Real-time GPU availability and price comparison across providers
- [ ] One-click GPU instance provisioning via RunPod adapter
- [ ] WireGuard privacy layer hiding upstream provider from customer
- [ ] Cloud-init boot script for tunnel setup, SSH keys, hostname, firewall
- [ ] Clerk JWT authentication on all customer API endpoints
- [ ] Stripe usage-based billing (per-second metering)
- [ ] SSH key management (CRUD)
- [ ] Instance lifecycle management (create, list, get, terminate)
- [ ] Redis-cached availability polling (30s interval)
- [ ] PostgreSQL persistence for instances, users, orgs, billing
- [ ] Docker environment persistence (save/deploy images across providers)
- [ ] Next.js dashboard: landing page, auth flow, instance management, billing UI

### Out of Scope

- E2E Networks adapter — deferred to next milestone
- Phase 2 own hardware / MIG virtualization — future
- Phase 3 predictive resource allocation — future
- Production deployment / infra — dev milestone only
- Closed beta operations — comes after dev milestone
- NVIDIA Confidential Computing / TEE — Phase 2
- Reserved tier (Novacore/CTRLS) — requires ops-assisted flow, not automated

## Context

- Existing codebase has full directory structure scaffolded but no working code
- Architecture doc (`docs/ARCHITECTURE.md`) defines the complete Phase 1 design including provider interface, database schema, API endpoints, cloud-init template, and WireGuard tunnel architecture
- Go 1.22+ with stdlib `net/http` routing patterns — no frameworks
- Three-tier provider strategy (Instant US, Instant India, Reserved) but only Instant US (RunPod) in this milestone
- RunPod API uses REST + GraphQL, per-second billing, ~15s boot time
- India providers (E2E Networks) offer 30-50% cost savings — key differentiator for later

## Constraints

- **Tech stack**: Go 1.22+ stdlib for backend, Next.js for frontend, PostgreSQL + Redis for storage — as defined in architecture doc
- **No frameworks**: stdlib `net/http` with new routing patterns, pgx for Postgres, go-redis for Redis — minimal deps
- **Privacy**: Customer must never see upstream provider identity, IP, or metadata
- **Single binary**: All Go backend services compile into one `gpuctl` binary
- **Python for tooling only**: Migrations, seeds, reports — never in the hot path

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| RunPod first, E2E later | Prove full stack with one provider before adding complexity | — Pending |
| WireGuard for privacy layer | Encrypted tunnels, kernel-level performance, simple peer management | — Pending |
| Clerk for auth | Managed auth with JWT verification, avoids building auth from scratch | — Pending |
| Stripe usage billing | Per-second metering matches GPU pricing model, industry standard | — Pending |
| Single Go binary | Operational simplicity, no service mesh needed for Phase 1 | — Pending |
| Dev milestone target | Get code working locally before worrying about deployment | — Pending |

---
*Last updated: 2026-02-24 after initialization*
