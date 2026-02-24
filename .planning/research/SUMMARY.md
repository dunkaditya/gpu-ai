# Project Research Summary

**Project:** GPU.ai
**Domain:** GPU Cloud Aggregation Platform
**Researched:** 2026-02-24
**Confidence:** HIGH

## Executive Summary

GPU.ai is a GPU cloud aggregation platform that re-rents GPU instances from upstream providers (RunPod initially, E2E Networks later) behind a WireGuard privacy layer, presenting customers with a single branded cloud experience. Experts build this as a **modular Go monolith** (single binary, `gpuctl`) with a Provider adapter interface for extensibility, a Redis-backed availability cache for real-time GPU inventory, Stripe usage-based billing for per-second metering, and a Next.js dashboard. The architecture is well-established: adapter pattern for multi-provider abstraction, background goroutines for polling and health monitoring, async provisioning with poll-for-status, and raw SQL (pgx) for financial correctness. The technology stack is mature and current -- Go 1.24, pgx v5, go-redis v9, Stripe v84, Clerk v2, wgctrl-go, Next.js 16, shadcn/ui, Tailwind v4.

The recommended approach is to build foundation-first: prove the RunPod adapter works with Docker-based initialization (not cloud-init VMs), validate WireGuard can run inside RunPod containers (requires NET_ADMIN capability verification), then layer on billing, availability, and the dashboard. The core differentiator -- full provider abstraction with WireGuard privacy -- is also the highest-risk feature and must be validated empirically before building the rest of the product around it. Start billing at provision time (not at instance-ready) to avoid margin erosion, use a local billing ledger as source of truth (not Stripe), and batch meter events to Stripe every 60 seconds.

The three highest risks are: (1) RunPod uses Docker containers, not VMs, so the cloud-init bootstrap script in the architecture doc will not work and must be completely redesigned as a custom Docker image with env-var injection; (2) WireGuard requires NET_ADMIN capability inside RunPod containers, which must be verified on a real pod before writing any WireGuard Go code -- the fallback is SSH reverse tunnels; (3) billing accuracy is a distributed transaction across PostgreSQL, Stripe, and the upstream provider, with race conditions on termination that can cause over-charging or margin loss. All three are solvable, but each requires deliberate design rather than naive implementation.

## Key Findings

### Recommended Stack

The stack is constrained by project decisions (Go backend, stdlib net/http, no ORM, no frameworks) and aligns with current best practices. All library versions are verified against official releases as of February 2026.

**Core technologies:**
- **Go 1.24** with stdlib `net/http`: Single binary backend. Go 1.22+ routing patterns (`GET /path/{id}`) eliminate framework need. Swiss Table maps give free 2-3% CPU improvement. Tool directives in go.mod replace tools.go hack.
- **PostgreSQL 16+ via pgx v5.7.5**: ACID for billing, JSONB for provider metadata, pgcrypto for UUIDs. pgxpool for connection management. Raw SQL only -- no ORM.
- **Redis 7.x via go-redis v9.18.0**: Sub-millisecond reads for GPU availability cache. TTL-based key expiry. Ephemeral data only.
- **Stripe v84 (stripe-go)**: Billing Meters API for per-second usage metering. 1,000 events/sec limit in live mode. Replaces deprecated Usage Records API.
- **Clerk v2 (clerk-sdk-go)**: JWT verification with automatic JWKS caching. Go middleware pattern integration.
- **WireGuard via wgctrl-go**: Programmatic peer management on proxy server. Curve25519 key generation via golang.org/x/crypto.
- **Next.js 16**: App Router, React Server Components, Turbopack. First-class Clerk integration. shadcn/ui + Tailwind v4 for dashboard.
- **TanStack Query v5**: Server state management for GPU availability polling and instance status updates.

**Critical version notes:**
- pgx v5.7.5 (not v5.8.0) if targeting Go 1.22 minimum; v5.8.0 requires Go 1.24
- Stripe Billing Meters API is the required approach -- legacy Usage Records API was removed in Stripe API version 2025-03-31
- Use `wireguard-go` (userspace) inside containers, not kernel WireGuard module
- Use `NUMERIC(10,6)` for price storage, not `NUMERIC(10,4)` -- per-second rates need 6+ decimal places

### Expected Features

**Must have (table stakes -- users leave without these):**
- GPU instance provisioning (create/list/get/terminate) -- the fundamental product
- SSH access via branded hostnames (`gpu-4a7f.gpu.ai`) -- primary interaction model
- WireGuard privacy layer -- core differentiator, hides upstream provider identity
- Real-time GPU availability display -- users must see what is available before provisioning
- Per-second billing with transparent pricing -- industry standard (RunPod bills per-millisecond)
- User authentication and SSH key management -- via Clerk
- REST API for programmatic access -- power users and CI/CD pipelines depend on this
- Basic web dashboard -- instance management, GPU availability, billing summary
- Docker image support -- custom CUDA/PyTorch environments
- Multiple GPU types and region selection -- competitive necessity

**Should have (differentiators):**
- Cross-provider availability aggregation -- see GPUs from ALL providers in one view
- Automatic best-price routing -- cheapest provider selected transparently
- Unified billing across providers -- one Stripe invoice for multi-provider compute
- Branded GPU hostnames with full identity hiding -- `gpu-xxx.gpu.ai` not raw IPs

**Defer to v2+:**
- E2E Networks adapter (India cost arbitrage) -- build after RunPod flow is solid
- Spot instance cross-provider migration -- requires checkpoint/restore
- Serverless GPU endpoints -- massive engineering effort, compete where we cannot win
- Jupyter/web IDE -- SSH tunneling documented as workaround for v1
- GPU utilization monitoring -- users can run nvidia-smi via SSH
- Multi-node GPU clusters -- cannot span providers due to InfiniBand requirements

**Anti-features (do not build):**
- Community marketplace for hosting GPUs (Vast.ai model) -- fundamentally different business
- Kubernetes-native orchestration -- conflicts with single-instance model
- Real-time GPU metrics dashboard -- requires agent on every instance, data pipeline

### Architecture Approach

The system is a single Go binary monolith (`gpuctl`) with well-defined internal package boundaries. Components communicate via direct function calls (same process), with background goroutines for polling and health monitoring. The Provider adapter interface is the core extensibility mechanism -- adding a new upstream provider requires implementing 5 methods and zero changes to existing code. Provisioning follows a state machine pattern (`creating -> provisioning -> booting -> running -> stopping -> terminated`) with each transition recorded in PostgreSQL. All customer-facing API responses use separate response types that structurally exclude upstream provider details.

**Major components:**
1. **Provisioning Engine** (`internal/provision/`) -- The orchestration brain. Selects provider from availability cache, generates WireGuard keypair, renders init template, calls upstream adapter, writes DB record, registers WG peer. Most complex component.
2. **Provider Registry + Adapters** (`internal/provider/`) -- Interface with 5 methods (Name, ListAvailable, Provision, GetStatus, Terminate). Registry holds map of adapters. RunPod adapter uses GraphQL API.
3. **Availability Poller** (`internal/availability/`) -- 30-second ticker goroutine polling all providers concurrently, writing results to Redis with TTL.
4. **WireGuard Manager** (`internal/wireguard/`) -- Programmatic peer add/remove via wgctrl-go on the proxy server. IPAM for 10.0.0.0/16 subnet.
5. **Billing Service** (`internal/billing/`) -- Stripe integration for usage metering, invoice generation. Local ledger in PostgreSQL is source of truth.
6. **Auth Middleware** (`internal/auth/`) -- Clerk JWT verification, user/org context injection.
7. **HTTP API Layer** (`internal/api/`) -- Thin handlers that decode request, call service, encode response. Privacy-first response filtering.
8. **Health Monitor** (`internal/health/`) -- 60-second goroutine checking WireGuard handshake timestamps.
9. **Next.js Dashboard** (`frontend/`) -- Completely separate build. Communicates only via REST API.

**Key patterns:**
- Dependency injection via constructors (main.go is composition root)
- Background goroutine with context cancellation for polling/monitoring
- Privacy-first response filtering (separate internal/external types)
- Async provisioning (return 201 immediately, poll for status)
- State machine for instance lifecycle

### Critical Pitfalls

1. **RunPod uses Docker containers, not VMs -- cloud-init will not work.** The bootstrap.sh script uses `systemctl`, `apt-get`, `hostnamectl` which do not work in containers. Must build a custom Docker image and use RunPod's `pre_start.sh`/`post_start.sh` hooks with env var injection. This is a complete redesign of the initialization strategy, not a minor tweak. Recovery cost: HIGH (1-2 weeks).

2. **WireGuard requires NET_ADMIN inside RunPod containers.** Even userspace `wireguard-go` needs NET_ADMIN capability and `/dev/net/tun`. Must verify empirically on a real RunPod pod before writing any WireGuard Go code. If unavailable, fall back to SSH reverse tunnels (slower, less elegant). Recovery cost: HIGH (2-3 weeks if fallback needed).

3. **Stripe meter events accept only positive integers.** Per-second cost of a $2.12/hr GPU is $0.000589 -- cannot be sent as a meter event value. Solution: report GPU-seconds as integer count (value=60 every minute), use local PostgreSQL ledger as billing source of truth, Stripe for payment collection only. Never trust Stripe for billing calculation.

4. **Billing race condition on instance termination.** Termination is a distributed transaction across upstream provider, PostgreSQL, and Stripe. Any component can fail independently. Solution: state machine (never skip states), record `billing_end` in PostgreSQL BEFORE calling upstream terminate (safer failure mode: we eat cost rather than overcharge customer), idempotent terminate endpoint, reconciliation job.

5. **Privacy layer leaks through multiple channels.** Even with WireGuard, provider identity leaks via: RunPod `RUNPOD_*` environment variables, cloud metadata endpoint (169.254.169.254), DNS resolution, `/etc/hosts`, `nvidia-smi` output, traceroute, kernel version. Must scrub all env vars, block metadata endpoint, force DNS through WireGuard tunnel, and run automated privacy audit on every provider integration.

6. **Per-second billing drift causes margin loss.** RunPod starts charging when the pod starts. GPU.ai's `billing_start` may lag 15-45 seconds (waiting for boot + WireGuard + ready callback). At $2.12/hr, 30 seconds of unbilled time per provision is $0.018/provision. At scale: $6,500/year. Solution: start billing at provision request, not at ready callback.

7. **Stale availability cache race condition.** 30-second polling means data is 0-30 seconds stale. Customer sees "available," clicks provision, GPU is gone. Solution: treat availability as a hint (not guarantee), invalidate cache on provisioning failure, graceful error handling with retry suggestions.

## Implications for Roadmap

Based on combined research, the following phase structure respects dependency ordering, addresses the highest-risk pitfalls early, and groups features by architectural cohesion.

### Phase 1: Project Foundation
**Rationale:** Every component depends on config loading, database connectivity, and the project scaffold. This must come first.
**Delivers:** Compilable Go binary, health endpoint, PostgreSQL schema, Redis connection, CI basics.
**Addresses:** Config loading (`internal/config/`), database pool (`internal/db/pool.go`), migrations (`database/migrations/001_init.sql`), health endpoint.
**Avoids:** Building on unstable foundation.
**Stack elements:** Go 1.24, pgx v5, go-redis v9, Docker Compose for local dev.

### Phase 2: Provider Interface + RunPod Adapter
**Rationale:** The RunPod adapter is the highest-integration-risk component. RunPod uses Docker containers, not VMs. Cloud-init will not work. WireGuard may not be possible inside containers. These must be validated empirically BEFORE building the provisioning engine around assumptions that may be wrong.
**Delivers:** Provider interface definition, RunPod adapter (GraphQL client), proof-of-concept Docker image that boots on RunPod with WireGuard and SSH, validated NET_ADMIN capability.
**Addresses:** Provider interface, RunPod API integration, Docker-based initialization.
**Avoids:** Pitfall 1 (cloud-init failure), Pitfall 2 (NET_ADMIN unavailability).
**CRITICAL:** This phase must include a hands-on spike: create a real RunPod pod with a custom Docker image, verify WireGuard runs, verify SSH works through a tunnel. Do not proceed to Phase 3 without this validation.

### Phase 3: WireGuard Privacy Layer + Instance Init
**Rationale:** The privacy layer is the core differentiator and depends on Phase 2 validation. Once we know WireGuard works in RunPod containers, build the full privacy layer: key generation, peer management, IPAM, DNS, firewall rules, env scrubbing.
**Delivers:** WireGuard manager, key generation, IPAM (use /16 from day one), proxy server setup, custom Docker image with full privacy layer, privacy audit checklist.
**Addresses:** Branded hostnames, provider identity hiding, SSH access routing.
**Avoids:** Pitfall 5 (privacy leaks), Pitfall 9 (address space exhaustion with /24).
**Implements:** WireGuard Manager component, IPAM, init template system.

### Phase 4: Database Schema + Instance Lifecycle
**Rationale:** Instance management is the core product and requires the full database schema. The state machine pattern must be established here, before billing hooks into it.
**Delivers:** Full database schema (organizations, users, instances, ssh_keys, usage_records), instance CRUD, state machine (`creating -> provisioning -> booting -> running -> stopping -> terminated`), idempotent termination.
**Addresses:** Instance provisioning API, SSH key management, instance status tracking.
**Avoids:** Pitfall 8 (billing race on termination -- state machine prevents it), Pitfall 10 (float precision -- use NUMERIC(10,6) and integer cents in Go).

### Phase 5: Auth + Billing
**Rationale:** Auth gates all customer operations. Billing is tightly coupled to instance lifecycle (billing_start on provision, billing_end on terminate) and must be designed together with the state machine from Phase 4. Grouping these ensures the billing ledger is correct from day one.
**Delivers:** Clerk JWT middleware, auth context injection, Stripe customer creation, Billing Meter setup, per-second metering (batched to 60-second reports), local billing ledger in PostgreSQL, billing reconciliation report (Python tooling).
**Addresses:** Authentication, per-second billing, transparent pricing, usage history.
**Avoids:** Pitfall 3 (Stripe integer-only meter values), Pitfall 7 (billing drift), Pitfall 8 (termination race).
**Uses:** Stripe v84 Billing Meters API (not legacy Usage Records), Clerk v2 SDK.

### Phase 6: Availability Engine + Provisioning Orchestration
**Rationale:** The availability poller and provisioning engine depend on the provider adapter (Phase 2), WireGuard manager (Phase 3), database (Phase 4), and billing hooks (Phase 5). This phase wires everything together into the complete provisioning flow.
**Delivers:** Availability poller (30s concurrent polling), Redis cache with 35s TTL, best-price selection, full provisioning orchestration (select provider -> WG keys -> init template -> upstream API -> DB -> WG peer -> billing start), cache invalidation on provisioning failure.
**Addresses:** Real-time GPU availability, automatic best-price routing, cross-provider aggregation.
**Avoids:** Pitfall 4 (stale cache race condition), Performance trap (sequential polling).

### Phase 7: API Layer + Error Handling
**Rationale:** With all business logic in place, build the thin HTTP layer on top. This phase focuses on API correctness, privacy-first response filtering, error handling, and security hardening.
**Delivers:** All REST API endpoints, privacy-first response types, generic error messages (never leak upstream details), internal endpoint security, rate limiting, CORS configuration.
**Addresses:** REST API, programmatic access, security hardening.
**Avoids:** Pitfall 5 (leaking upstream errors), Security mistake (internal endpoints publicly accessible).

### Phase 8: Health Monitoring + Ops
**Rationale:** Once instances are running, detect failures. Spot interruption handling, WireGuard tunnel monitoring, billing stop on instance death.
**Delivers:** Health monitor goroutine (60s), spot interruption detection, automatic billing stop on instance death, reconciliation tooling, alerting.
**Addresses:** Instance health monitoring, spot instance handling, operational safety.
**Avoids:** Pitfall 6 (spot interruption billing continues after pod death).

### Phase 9: Next.js Dashboard
**Rationale:** The API is the primary interface. The dashboard is a UI layer on top and should be built against a stable, tested API. Building it last means no wasted frontend work when API contracts change.
**Delivers:** Complete customer dashboard -- auth flow, GPU availability view, instance management, billing summary, SSH key management.
**Addresses:** Web dashboard (table stakes), user experience.
**Uses:** Next.js 16, Clerk Next.js SDK, shadcn/ui, TanStack Query v5, Zustand.

### Phase 10: Polish + Differentiators
**Rationale:** After core product is working end-to-end, add differentiator features and operational hardening based on beta feedback.
**Delivers:** Docker environment persistence, spend limits, billing alerts, team/org management, CLI tool considerations.
**Addresses:** v1.x features from feature research.

### Phase Ordering Rationale

- **Foundation first (Phase 1):** Everything depends on config, database, and the project scaffold.
- **Highest-risk integration second (Phase 2):** RunPod Docker containers and WireGuard NET_ADMIN are the biggest unknowns. Validate them before building 8 phases of code on top of assumptions.
- **Core differentiator third (Phase 3):** The WireGuard privacy layer IS the product. If it does not work, the product concept needs to change.
- **State machine before billing (Phase 4 before 5):** The instance lifecycle state machine must exist before billing hooks into it. Billing correctness depends on reliable state transitions.
- **API layer after business logic (Phase 7):** Thin HTTP handlers are easy to write once the underlying services are stable. Writing handlers first leads to constant rework.
- **Dashboard last (Phase 9):** The API must be stable before building a frontend against it. Dashboard enhances everything but blocks nothing.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2 (RunPod Adapter):** Needs hands-on spike with real RunPod API. Docker-based initialization is a complete departure from cloud-init. GraphQL vs REST API coverage needs investigation. NET_ADMIN verification is mandatory.
- **Phase 3 (WireGuard Privacy Layer):** If Phase 2 spike shows WireGuard cannot run in RunPod containers, this phase needs a complete re-architecture to SSH reverse tunnels.
- **Phase 5 (Billing):** Stripe Billing Meters API v2 specifics need investigation -- exact event schema, aggregation windows, integer value semantics, invoicing timing. The migration from legacy Usage Records is recent and documentation may have gaps.
- **Phase 6 (Availability Engine):** RunPod's `gpuTypes` GraphQL query response format needs investigation for availability count accuracy and rate limiting behavior.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Foundation):** Go project scaffold, pgx pool, config loading -- well-documented, no unknowns.
- **Phase 4 (Database + Instance Lifecycle):** Standard SQL schema design, state machine pattern is well-documented (Google Compute Engine lifecycle is the reference).
- **Phase 7 (API Layer):** Go stdlib HTTP handlers, middleware chain -- extremely well-documented.
- **Phase 9 (Dashboard):** Next.js + Clerk + shadcn/ui -- massive ecosystem of tutorials and templates.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All library versions verified via GitHub releases and pkg.go.dev. Go 1.24 (1 year old), pgx v5 (mature), go-redis v9 (official), Stripe v84 (official), Clerk v2 (official). No speculative technology choices. |
| Features | MEDIUM-HIGH | Feature landscape mapped from 6+ competitor product pages (RunPod, Lambda, Vast.ai, CoreWeave) and third-party comparisons. Table stakes are clear. Differentiators are validated by competitive gap analysis. Some pricing data from third-party sources. |
| Architecture | HIGH | Architecture follows established patterns (adapter, state machine, DI). Referenced implementations from Tailscale (WireGuard overlay), Google Compute Engine (instance lifecycle), AWS (idempotent APIs). Project structure aligns with Go community conventions. |
| Pitfalls | HIGH | Critical pitfalls verified across official docs (RunPod pod initialization, Stripe Meters API, WireGuard capabilities). 8 critical pitfalls identified with prevention strategies and phase mappings. Integration gotchas sourced from official documentation. |

**Overall confidence:** HIGH

### Gaps to Address

- **RunPod Docker initialization (HIGHEST PRIORITY):** The cloud-init approach in `docs/ARCHITECTURE.md` and `infra/cloud-init/bootstrap.sh` will not work on RunPod. Must be completely redesigned. This is the biggest gap between current architecture and reality. Must validate with a real RunPod pod in Phase 2.

- **WireGuard in RunPod containers:** NET_ADMIN capability and /dev/net/tun availability are unconfirmed. If unavailable, the entire privacy layer architecture changes. Verify empirically before Phase 3.

- **Proxy server infrastructure:** The architecture assumes a WireGuard proxy server exists but does not specify provisioning, configuration, monitoring, or failover. Need infrastructure-as-code before Phase 3. Single proxy is a SPOF.

- **DNS wildcard for `*.gpu.ai`:** Required for branded hostnames. Needs DNS provider supporting wildcard A records. Not addressed in any research file. Must resolve before Phase 3.

- **RunPod API rate limits:** Not publicly documented. The availability poller will call RunPod every 30 seconds. Under load, provisioning adds more API calls. Need to discover limits empirically and build in throttling/backoff from day one.

- **Price storage precision:** Current schema uses `NUMERIC(10,4)` but per-second rates need 6+ decimal places. Must change to `NUMERIC(10,6)` in the migration. Use integer cents in Go code to avoid floating-point errors in billing calculations.

- **E2E Networks API:** Deferred to future milestone but completely unresearched. REST API documentation quality, rate limits, cloud-init support, and GPU SKU catalog are unknowns.

## Sources

### Primary (HIGH confidence)
- RunPod Pod Management GraphQL API docs -- provisioning mutations, availability queries
- RunPod Pricing docs -- spot instance interruption behavior (5-second SIGTERM)
- RunPod Templates docs -- container initialization scripts (pre_start.sh/post_start.sh)
- Stripe Billing Meters API docs -- meter events, integer values, 1000/sec rate limit
- Stripe Usage-Based Billing migration guide -- legacy to Meters API transition
- Clerk Go SDK v2 docs -- JWT verification, JWKS caching
- wgctrl-go package docs -- ConfigureDevice API for peer management
- Go 1.24 release notes -- Swiss Tables, tool directives, routing patterns
- Google Compute Engine Instance Lifecycle -- state machine reference
- AWS Idempotent APIs guide -- retry safety patterns
- Redis Real-Time Inventory -- caching patterns

### Secondary (MEDIUM confidence)
- RunPod community wiki (deepwiki.com) -- container initialization details
- GPU price comparison (getdeploying.com) -- cross-provider pricing data
- Competitor analysis articles (Northflank, DigitalOcean, Hyperstack) -- feature comparisons
- Tailscale architecture blog -- WireGuard overlay network reference
- NetBird architecture analysis -- WireGuard + Go at scale
- Stripe limitations analysis (withorb.com) -- usage-based billing edge cases

### Tertiary (LOW confidence)
- RunPod community discussions (answeroverflow.com) -- API quirks
- RunPod custom container blog posts -- practitioner experiences
- GPU security risks analysis (secureworld.io) -- hosting security considerations

---
*Research completed: 2026-02-24*
*Ready for roadmap: yes*
