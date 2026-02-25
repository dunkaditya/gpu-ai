# Requirements: GPU.ai

**Defined:** 2026-02-24
**Core Value:** Customers can find available GPUs across providers and provision them instantly through a single interface, with a privacy layer that completely hides the upstream provider.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Foundation

- [x] **FOUND-01**: Go binary compiles and runs with health endpoint on configurable port
- [x] **FOUND-02**: Config loads from environment variables with validation and sensible defaults
- [x] **FOUND-03**: PostgreSQL connection pool initializes and verifies connectivity on startup
- [x] **FOUND-04**: Redis connection initializes and verifies connectivity on startup
- [x] **FOUND-05**: Database migrations apply cleanly to create full schema

### Schema Improvements

- [x] **SCHEMA-01**: Rename all primary keys to self-documenting `{table}_id` format and update foreign key references
- [x] **SCHEMA-02**: Add NOT NULL constraints on mandatory foreign keys, explicit ON DELETE behavior on all FKs, CHECK constraint on instances.status, UNIQUE on instances.hostname, composite unique index on (upstream_provider, upstream_id)
- [x] **SCHEMA-03**: Remove wg_private_key_enc column (ephemeral key, security liability to store)
- [x] **SCHEMA-04**: Add internal_token column to instances for per-instance callback authentication, add updated_at column to instances

### Provider Integration

- [x] **PROV-01**: Provider interface defines standard contract (Name, ListAvailable, Provision, GetStatus, Terminate)
- [x] **PROV-02**: Provider registry manages multiple adapters with lookup by name
- [x] **PROV-03**: RunPod adapter lists available GPU types with pricing via GraphQL API
- [x] **PROV-04**: RunPod adapter provisions a pod with custom Docker image and startup scripts
- [x] **PROV-05**: RunPod adapter queries pod status by upstream ID
- [x] **PROV-06**: RunPod adapter terminates a pod by upstream ID

### Privacy Layer

- [x] **PRIV-01**: WireGuard key pairs generated for each new instance
- [x] **PRIV-02**: WireGuard peers added to proxy server programmatically via wgctrl-go
- [x] **PRIV-03**: WireGuard peers removed from proxy server on instance termination
- [x] **PRIV-04**: IPAM allocates unique WireGuard addresses from subnet pool backed by PostgreSQL
- [x] **PRIV-05**: Instance init template renders with WireGuard config, SSH keys, hostname, firewall rules
- [x] **PRIV-06**: Customer SSH connections route through WireGuard proxy with branded hostname
- [x] **PRIV-07**: Upstream provider identity (name, IP, env vars, metadata endpoint) hidden from customer
- [x] **PRIV-08**: All customer-facing API responses structurally exclude upstream provider details

### Instance Lifecycle

- [x] **INST-01**: User can create a GPU instance specifying type, count, region, tier, SSH keys
- [x] **INST-02**: User can list their active instances with status and connection info
- [x] **INST-03**: User can get details of a specific instance by ID
- [x] **INST-04**: User can terminate an instance and billing stops
- [x] **INST-05**: Instance follows state machine (creating -> provisioning -> booting -> running -> stopping -> terminated)
- [x] **INST-06**: Instance termination is idempotent (multiple calls produce same result)
- [x] **INST-07**: Instance ready callback transitions status from booting to running
- [x] **INST-08**: Instance creation response includes confirmed hourly cost so user knows what they're paying before resources are allocated

### Authentication

- [x] **AUTH-01**: All customer API endpoints require valid Clerk JWT
- [x] **AUTH-02**: JWT verification extracts user_id and org_id into request context
- [x] **AUTH-03**: Users can only access instances belonging to their organization
- [x] **AUTH-04**: Internal endpoints restricted to localhost only

### SSH Key Management

- [x] **SSHK-01**: User can add an SSH public key with a name
- [x] **SSHK-02**: User can list their SSH keys
- [x] **SSHK-03**: User can delete an SSH key
- [x] **SSHK-04**: SSH keys are injected into new instances at provision time

### Billing

- [x] **BILL-01**: Billing starts at instance provision request time
- [x] **BILL-02**: Billing stops at instance termination time
- [x] **BILL-03**: Per-second usage tracked in PostgreSQL billing ledger (source of truth)
- [x] **BILL-04**: Usage batched and reported to Stripe Billing Meters every 60 seconds as integer GPU-seconds
- [x] **BILL-05**: User can view their usage history and costs
- [x] **BILL-06**: Billing records include GPU type, count, duration, and cost
- [x] **BILL-07**: Configurable per-org spending limit with automatic instance termination when exceeded

### Availability

- [x] **AVAIL-01**: Background poller queries all providers every 30 seconds
- [x] **AVAIL-02**: GPU offerings cached in Redis with 35-second TTL
- [ ] **AVAIL-03**: User can view available GPU types with pricing by region and tier
- [ ] **AVAIL-04**: Availability response aggregates across providers without revealing provider identity
- [ ] **AVAIL-05**: Provisioning engine selects best-price provider automatically

### API

- [x] **API-01**: POST /api/v1/instances creates a new GPU instance
- [x] **API-02**: GET /api/v1/instances lists user's instances
- [x] **API-03**: GET /api/v1/instances/{id} returns instance details
- [x] **API-04**: DELETE /api/v1/instances/{id} terminates an instance
- [ ] **API-05**: GET /api/v1/gpu/available returns GPU availability with optional filters
- [x] **API-06**: GET/POST/DELETE /api/v1/ssh-keys manages SSH keys
- [x] **API-07**: GET /api/v1/billing/usage returns billing history
- [x] **API-08**: GET /health returns service health status
- [x] **API-09**: Error responses never leak upstream provider details
- [x] **API-10**: All list endpoints support cursor-based pagination with configurable page size
- [x] **API-11**: POST /api/v1/instances accepts Idempotency-Key header to prevent duplicate instance creation on network retries
- [x] **API-12**: All customer API endpoints are rate-limited per org (prevent runaway scripts from creating dozens of instances)

### Health Monitoring

- [ ] **HLTH-01**: Background goroutine monitors instance health every 60 seconds
- [ ] **HLTH-02**: Spot instance interruption detected and billing stopped automatically
- [ ] **HLTH-03**: Instance ready callback received from booted instances
- [ ] **HLTH-04**: Spot interruption and instance failure events trigger webhook notification to org's configured callback URL

### Dashboard

- [ ] **DASH-01**: Landing page describes the product
- [ ] **DASH-02**: User can sign up and log in via Clerk
- [ ] **DASH-03**: User can view real-time GPU availability with pricing
- [ ] **DASH-04**: User can provision a GPU instance from the dashboard
- [ ] **DASH-05**: User can view and manage running instances
- [ ] **DASH-06**: User can manage SSH keys
- [ ] **DASH-07**: User can view billing usage and costs
- [ ] **DASH-08**: Dashboard displays instance status with SSH connection command

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### India Provider

- **INDIA-01**: E2E Networks adapter implements Provider interface
- **INDIA-02**: India region available in availability and provisioning
- **INDIA-03**: Cross-region price comparison shows India cost advantage

### Environment Persistence

- **ENV-01**: User can save a running instance's Docker image to registry
- **ENV-02**: User can provision a new instance from a saved environment
- **ENV-03**: Environments are portable across providers

### Advanced Features

- ~~**ADV-01**: Spend limits and billing alerts~~ *(promoted to v1 as BILL-07)*
- **ADV-02**: Team/org management with role-based access
- **ADV-03**: CLI tool for programmatic instance management

## Out of Scope

| Feature | Reason |
|---------|--------|
| Serverless GPU endpoints | Massive engineering effort, competes where GPU.ai cannot win |
| Multi-node GPU clusters | InfiniBand cannot span providers |
| Kubernetes orchestration | Conflicts with single-instance model |
| Community GPU marketplace | Fundamentally different business model (Vast.ai) |
| Jupyter/web IDE | SSH tunneling documented as v1 workaround |
| Real-time GPU metrics dashboard | Requires agent on every instance + data pipeline |
| Production deployment/infra | Dev milestone only |
| E2E Networks adapter | Deferred to v2 milestone |
| NVIDIA Confidential Computing | Phase 2 future |
| Reserved tier (Novacore/CTRLS) | Requires ops-assisted flow |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Complete |
| FOUND-02 | Phase 1 | Complete |
| FOUND-03 | Phase 1 | Complete |
| FOUND-04 | Phase 1 | Complete |
| FOUND-05 | Phase 1 | Complete |
| SCHEMA-01 | Phase 2 | Complete |
| SCHEMA-02 | Phase 2 | Complete |
| SCHEMA-03 | Phase 2 | Complete |
| SCHEMA-04 | Phase 2 | Complete |
| PROV-01 | Phase 2 | Complete |
| PROV-02 | Phase 2 | Complete |
| PROV-03 | Phase 2 | Complete |
| PROV-04 | Phase 2 | Complete |
| PROV-05 | Phase 4.2 | Complete |
| PROV-06 | Phase 2 | Complete |
| PRIV-01 | Phase 4.1 | Complete |
| PRIV-02 | Phase 4.1 | Complete |
| PRIV-03 | Phase 4.1 | Complete |
| PRIV-04 | Phase 4.1 | Complete |
| PRIV-05 | Phase 4.1 | Complete |
| PRIV-06 | Phase 4.1 | Complete |
| PRIV-07 | Phase 3 | Complete |
| PRIV-08 | Phase 3 | Complete |
| INST-01 | Phase 4 | Complete |
| INST-02 | Phase 4 | Complete |
| INST-03 | Phase 4 | Complete |
| INST-04 | Phase 4 | Complete |
| INST-05 | Phase 4.2 | Complete |
| INST-06 | Phase 4 | Complete |
| INST-07 | Phase 4.2 | Complete |
| INST-08 | Phase 4 | Complete |
| AUTH-01 | Phase 4 | Complete |
| AUTH-02 | Phase 4 | Complete |
| AUTH-03 | Phase 4.3 | Complete |
| AUTH-04 | Phase 1 | Complete |
| SSHK-01 | Phase 5 | Complete |
| SSHK-02 | Phase 5 | Complete |
| SSHK-03 | Phase 5 | Complete |
| SSHK-04 | Phase 5 | Complete |
| BILL-01 | Phase 5 | Complete |
| BILL-02 | Phase 5 | Complete |
| BILL-03 | Phase 5 | Complete |
| BILL-04 | Phase 5 | Complete |
| BILL-05 | Phase 5 | Complete |
| BILL-06 | Phase 5 | Complete |
| BILL-07 | Phase 5 | Complete |
| AVAIL-01 | Phase 6 | Complete |
| AVAIL-02 | Phase 6 | Complete |
| AVAIL-03 | Phase 6 | Pending |
| AVAIL-04 | Phase 6 | Pending |
| AVAIL-05 | Phase 6 | Pending |
| API-01 | Phase 4 | Complete |
| API-02 | Phase 4 | Complete |
| API-03 | Phase 4 | Complete |
| API-04 | Phase 4 | Complete |
| API-05 | Phase 6 | Pending |
| API-06 | Phase 5 | Complete |
| API-07 | Phase 5 | Complete |
| API-08 | Phase 1 | Complete |
| API-09 | Phase 4 | Complete |
| API-10 | Phase 4 | Complete |
| API-11 | Phase 4.3 | Complete |
| API-12 | Phase 4 | Complete |
| HLTH-01 | Phase 6 | Pending |
| HLTH-02 | Phase 6 | Pending |
| HLTH-03 | Phase 6 | Pending |
| HLTH-04 | Phase 6 | Pending |
| DASH-01 | Phase 7 | Pending |
| DASH-02 | Phase 7 | Pending |
| DASH-03 | Phase 7 | Pending |
| DASH-04 | Phase 7 | Pending |
| DASH-05 | Phase 7 | Pending |
| DASH-06 | Phase 7 | Pending |
| DASH-07 | Phase 7 | Pending |
| DASH-08 | Phase 7 | Pending |

**Coverage:**
- v1 requirements: 75 total
- Mapped to phases: 75
- Unmapped: 0

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-24 after roadmap creation*
