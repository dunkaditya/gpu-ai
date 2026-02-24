# Roadmap: GPU.ai

## Overview

GPU.ai goes from scaffolded codebase to working dev milestone in 7 phases. Foundation and data infrastructure come first, then the highest-risk integration (RunPod adapter), then the core differentiator (WireGuard privacy layer), then the product itself (authenticated instance lifecycle), then supporting systems (SSH keys, billing, availability, health monitoring), and finally the customer dashboard. Each phase delivers a coherent, testable capability. The ordering validates risky assumptions early -- if RunPod containers cannot run WireGuard, we discover that in Phase 3, not Phase 7.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation** - Compilable binary with config, database, Redis, migrations, and health endpoint
- [x] **Phase 2: Provider Abstraction + RunPod Adapter** - Provider interface, registry, and RunPod GraphQL adapter (completed 2026-02-24)
- [x] **Phase 3: Privacy Layer** - WireGuard key management, peer management, IPAM, init template, and privacy filtering (completed 2026-02-24)
- [ ] **Phase 4: Auth + Instance Lifecycle** - Clerk JWT auth, instance CRUD with state machine, and core API endpoints
- [ ] **Phase 5: SSH Keys + Billing** - SSH key management, per-second billing ledger, and Stripe metering
- [ ] **Phase 6: Availability + Health Monitoring** - Background GPU polling, Redis cache, best-price selection, and instance health checks
- [ ] **Phase 7: Dashboard** - Next.js customer dashboard with auth, GPU availability, instance management, billing, and SSH keys

## Phase Details

### Phase 1: Foundation
**Goal**: A running Go binary with verified database and Redis connectivity, environment config, applied schema, and a health endpoint proving the system is alive
**Depends on**: Nothing (first phase)
**Requirements**: FOUND-01, FOUND-02, FOUND-03, FOUND-04, FOUND-05, API-08, AUTH-04
**Success Criteria** (what must be TRUE):
  1. Running `go run ./cmd/gpuctl` starts the server on the configured port and responds to HTTP requests
  2. GET /health returns 200 with database and Redis connectivity status
  3. Environment variables load with validation errors on missing required values and sensible defaults for optional ones
  4. Database migrations apply cleanly from scratch and create the full schema (organizations, users, instances, ssh_keys, usage_records tables)
  5. Internal endpoints (health, callbacks) are restricted to localhost -- external requests are rejected
**Plans**: 4 plans

Plans:
- [x] 01-01-PLAN.md — Config loading, Docker Compose, Go dependencies
- [x] 01-02-PLAN.md — Database migration SQL fix and Python migration runner
- [x] 01-03-PLAN.md — Database pool, Redis client, API server, health endpoint, main.go wiring
- [x] 01-04-PLAN.md — Gap closure: Add localhost-only IP restriction to internal endpoints (AUTH-04)

### Phase 2: Provider Abstraction + RunPod Adapter
**Goal**: A clean provider interface that any GPU cloud can implement, with a working RunPod adapter that can list GPUs, provision pods, check status, and terminate. Also applies schema improvements (self-documenting PKs, constraints, security fixes) via a new migration before instance CRUD code is written.
**Depends on**: Phase 1
**Requirements**: SCHEMA-01, SCHEMA-02, SCHEMA-03, SCHEMA-04, PROV-01, PROV-02, PROV-03, PROV-04, PROV-05, PROV-06
**Success Criteria** (what must be TRUE):
  1. Database schema v1 migration applies cleanly: self-documenting PKs, NOT NULL/ON DELETE constraints, CHECK on status, UNIQUE on hostname, composite index on upstream, internal_token column, wg_private_key_enc removed
  2. Provider interface defines a 5-method contract (Name, ListAvailable, Provision, GetStatus, Terminate) that compiles and is usable by any adapter
  3. Provider registry holds multiple adapters and looks up providers by name
  4. RunPod adapter translates GPU availability queries into RunPod GraphQL API calls and returns normalized results
  5. RunPod adapter provisions a pod with a custom Docker image and startup configuration, and can query its status and terminate it
**Plans**: 3 plans

Plans:
- [ ] 02-01-PLAN.md — Schema v1 migration (self-documenting PKs, constraints, security fixes)
- [ ] 02-02-PLAN.md — Provider interface refinement, types update, registry with tests
- [ ] 02-03-PLAN.md — RunPod GraphQL adapter (client, mapping, all 5 methods, tests)

### Phase 3: Privacy Layer
**Goal**: Complete WireGuard-based privacy infrastructure that generates keys, manages peers, allocates tunnel IPs, renders init templates, and ensures no upstream provider details ever reach the customer
**Depends on**: Phase 2
**Requirements**: PRIV-01, PRIV-02, PRIV-03, PRIV-04, PRIV-05, PRIV-06, PRIV-07, PRIV-08
**Success Criteria** (what must be TRUE):
  1. WireGuard key pairs are generated for each new instance and stored securely
  2. WireGuard peers are programmatically added to the proxy server on instance creation and removed on termination
  3. IPAM allocates unique tunnel addresses from the 10.0.0.0/16 subnet and reclaims them on termination
  4. Instance init template renders correctly with WireGuard config, SSH keys, hostname, and firewall rules
  5. Customer SSH connections route through WireGuard proxy with branded hostname -- upstream provider IP, name, and metadata are never visible
**Plans**: 3 plans

Plans:
- [ ] 03-01-PLAN.md — Schema migration, config updates, WireGuard key generation, AES-256-GCM encryption, privacy types (CustomerInstance)
- [ ] 03-02-PLAN.md — PostgreSQL-backed IPAM with advisory lock, WireGuard peer manager with iptables port mapping
- [ ] 03-03-PLAN.md — Cloud-init Go template with privacy scrubbing, template renderer with input validation

### Phase 4: Auth + Instance Lifecycle
**Goal**: Authenticated users can create, list, view, and terminate GPU instances through REST API endpoints, with a full state machine governing instance transitions, organization-scoped access control, rate limiting, idempotency, and pagination
**Depends on**: Phase 1, Phase 2, Phase 3
**Requirements**: AUTH-01, AUTH-02, AUTH-03, INST-01, INST-02, INST-03, INST-04, INST-05, INST-06, INST-07, INST-08, API-01, API-02, API-03, API-04, API-09, API-10, API-11, API-12
**Success Criteria** (what must be TRUE):
  1. All customer API endpoints reject requests without a valid Clerk JWT and extract user_id/org_id into request context
  2. POST /api/v1/instances creates a GPU instance that progresses through the state machine (creating -> provisioning -> booting -> running) and returns confirmed hourly cost
  3. GET /api/v1/instances returns only instances belonging to the authenticated user's organization with cursor-based pagination
  4. DELETE /api/v1/instances/{id} terminates an instance, is idempotent (multiple calls produce same result), and transitions state to terminated
  5. All API error responses and instance details structurally exclude upstream provider identity -- no provider name, IP, or metadata leaks
  6. POST /api/v1/instances accepts Idempotency-Key header to prevent duplicate creation on network retries
  7. All customer API endpoints are rate-limited per organization
**Plans**: TBD

Plans:
- [ ] 04-01: TBD
- [ ] 04-02: TBD
- [ ] 04-03: TBD

### Phase 5: SSH Keys + Billing
**Goal**: Users can manage SSH keys that are injected into new instances, per-second billing tracks usage accurately in a PostgreSQL ledger with batched reporting to Stripe, and per-org spending limits prevent bill shock
**Depends on**: Phase 4
**Requirements**: SSHK-01, SSHK-02, SSHK-03, SSHK-04, BILL-01, BILL-02, BILL-03, BILL-04, BILL-05, BILL-06, BILL-07, API-06, API-07
**Success Criteria** (what must be TRUE):
  1. User can add, list, and delete SSH public keys via API, and keys are injected into new instances at provision time
  2. Billing starts at instance provision request time and stops at termination time -- no unbilled gaps
  3. Per-second usage is tracked in the PostgreSQL billing ledger with GPU type, count, duration, and cost
  4. Usage is batched and reported to Stripe Billing Meters every 60 seconds as integer GPU-seconds
  5. User can retrieve their billing usage history and costs via GET /api/v1/billing/usage
  6. Per-org spending limit can be configured, and instances are automatically terminated when the limit is exceeded
**Plans**: TBD

Plans:
- [ ] 05-01: TBD
- [ ] 05-02: TBD
- [ ] 05-03: TBD

### Phase 6: Availability + Health Monitoring
**Goal**: Background systems continuously poll providers for GPU availability (cached in Redis), select the best-price provider for provisioning, monitor running instances for health and spot interruptions, and notify orgs via webhook on instance failures
**Depends on**: Phase 2, Phase 4
**Requirements**: AVAIL-01, AVAIL-02, AVAIL-03, AVAIL-04, AVAIL-05, HLTH-01, HLTH-02, HLTH-03, HLTH-04, API-05
**Success Criteria** (what must be TRUE):
  1. Background poller queries all registered providers every 30 seconds and caches results in Redis with 35-second TTL
  2. GET /api/v1/gpu/available returns aggregated GPU offerings with pricing by region and tier -- without revealing provider identity
  3. Provisioning engine automatically selects the best-price provider when creating an instance
  4. Health monitor detects spot instance interruptions and automatically stops billing
  5. Instance ready callback transitions instance status from booting to running
  6. Spot interruptions and instance failures trigger webhook notification to org's configured callback URL
**Plans**: TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD
- [ ] 06-03: TBD

### Phase 7: Dashboard
**Goal**: A complete Next.js customer dashboard where users can sign up, browse GPU availability, provision and manage instances, manage SSH keys, and view billing -- all backed by the stable API
**Depends on**: Phase 4, Phase 5, Phase 6
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04, DASH-05, DASH-06, DASH-07, DASH-08
**Success Criteria** (what must be TRUE):
  1. Landing page describes the product and guides visitors to sign up
  2. User can sign up and log in via Clerk, and authenticated routes are protected
  3. User can view real-time GPU availability with pricing and provision an instance from the dashboard
  4. User can view and manage running instances with status indicators and SSH connection commands
  5. User can manage SSH keys and view billing usage and costs from the dashboard
**Plans**: TBD

Plans:
- [ ] 07-01: TBD
- [ ] 07-02: TBD
- [ ] 07-03: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 4/4 | Complete | 2026-02-24 |
| 2. Provider Abstraction + RunPod Adapter | 0/3 | Complete    | 2026-02-24 |
| 3. Privacy Layer | 0/3 | Complete    | 2026-02-24 |
| 4. Auth + Instance Lifecycle | 0/3 | Not started | - |
| 5. SSH Keys + Billing | 0/3 | Not started | - |
| 6. Availability + Health Monitoring | 0/3 | Not started | - |
| 7. Dashboard | 0/3 | Not started | - |
