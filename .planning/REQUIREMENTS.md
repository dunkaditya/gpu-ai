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

### Provider Integration

- [ ] **PROV-01**: Provider interface defines standard contract (Name, ListAvailable, Provision, GetStatus, Terminate)
- [ ] **PROV-02**: Provider registry manages multiple adapters with lookup by name
- [ ] **PROV-03**: RunPod adapter lists available GPU types with pricing via GraphQL API
- [ ] **PROV-04**: RunPod adapter provisions a pod with custom Docker image and startup scripts
- [ ] **PROV-05**: RunPod adapter queries pod status by upstream ID
- [ ] **PROV-06**: RunPod adapter terminates a pod by upstream ID

### Privacy Layer

- [ ] **PRIV-01**: WireGuard key pairs generated for each new instance
- [ ] **PRIV-02**: WireGuard peers added to proxy server programmatically via wgctrl-go
- [ ] **PRIV-03**: WireGuard peers removed from proxy server on instance termination
- [ ] **PRIV-04**: IPAM allocates unique WireGuard addresses from subnet pool backed by PostgreSQL
- [ ] **PRIV-05**: Instance init template renders with WireGuard config, SSH keys, hostname, firewall rules
- [ ] **PRIV-06**: Customer SSH connections route through WireGuard proxy with branded hostname
- [ ] **PRIV-07**: Upstream provider identity (name, IP, env vars, metadata endpoint) hidden from customer
- [ ] **PRIV-08**: All customer-facing API responses structurally exclude upstream provider details

### Instance Lifecycle

- [ ] **INST-01**: User can create a GPU instance specifying type, count, region, tier, SSH keys
- [ ] **INST-02**: User can list their active instances with status and connection info
- [ ] **INST-03**: User can get details of a specific instance by ID
- [ ] **INST-04**: User can terminate an instance and billing stops
- [ ] **INST-05**: Instance follows state machine (creating -> provisioning -> booting -> running -> stopping -> terminated)
- [ ] **INST-06**: Instance termination is idempotent (multiple calls produce same result)
- [ ] **INST-07**: Instance ready callback transitions status from booting to running

### Authentication

- [ ] **AUTH-01**: All customer API endpoints require valid Clerk JWT
- [ ] **AUTH-02**: JWT verification extracts user_id and org_id into request context
- [ ] **AUTH-03**: Users can only access instances belonging to their organization
- [x] **AUTH-04**: Internal endpoints restricted to localhost only

### SSH Key Management

- [ ] **SSHK-01**: User can add an SSH public key with a name
- [ ] **SSHK-02**: User can list their SSH keys
- [ ] **SSHK-03**: User can delete an SSH key
- [ ] **SSHK-04**: SSH keys are injected into new instances at provision time

### Billing

- [ ] **BILL-01**: Billing starts at instance provision request time
- [ ] **BILL-02**: Billing stops at instance termination time
- [ ] **BILL-03**: Per-second usage tracked in PostgreSQL billing ledger (source of truth)
- [ ] **BILL-04**: Usage batched and reported to Stripe Billing Meters every 60 seconds as integer GPU-seconds
- [ ] **BILL-05**: User can view their usage history and costs
- [ ] **BILL-06**: Billing records include GPU type, count, duration, and cost

### Availability

- [ ] **AVAIL-01**: Background poller queries all providers every 30 seconds
- [ ] **AVAIL-02**: GPU offerings cached in Redis with 35-second TTL
- [ ] **AVAIL-03**: User can view available GPU types with pricing by region and tier
- [ ] **AVAIL-04**: Availability response aggregates across providers without revealing provider identity
- [ ] **AVAIL-05**: Provisioning engine selects best-price provider automatically

### API

- [ ] **API-01**: POST /api/v1/instances creates a new GPU instance
- [ ] **API-02**: GET /api/v1/instances lists user's instances
- [ ] **API-03**: GET /api/v1/instances/{id} returns instance details
- [ ] **API-04**: DELETE /api/v1/instances/{id} terminates an instance
- [ ] **API-05**: GET /api/v1/gpu/available returns GPU availability with optional filters
- [ ] **API-06**: GET/POST/DELETE /api/v1/ssh-keys manages SSH keys
- [ ] **API-07**: GET /api/v1/billing/usage returns billing history
- [x] **API-08**: GET /health returns service health status
- [ ] **API-09**: Error responses never leak upstream provider details

### Health Monitoring

- [ ] **HLTH-01**: Background goroutine monitors instance health every 60 seconds
- [ ] **HLTH-02**: Spot instance interruption detected and billing stopped automatically
- [ ] **HLTH-03**: Instance ready callback received from booted instances

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

- **ADV-01**: Spend limits and billing alerts
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
| PROV-01 | Phase 2 | Pending |
| PROV-02 | Phase 2 | Pending |
| PROV-03 | Phase 2 | Pending |
| PROV-04 | Phase 2 | Pending |
| PROV-05 | Phase 2 | Pending |
| PROV-06 | Phase 2 | Pending |
| PRIV-01 | Phase 3 | Pending |
| PRIV-02 | Phase 3 | Pending |
| PRIV-03 | Phase 3 | Pending |
| PRIV-04 | Phase 3 | Pending |
| PRIV-05 | Phase 3 | Pending |
| PRIV-06 | Phase 3 | Pending |
| PRIV-07 | Phase 3 | Pending |
| PRIV-08 | Phase 3 | Pending |
| INST-01 | Phase 4 | Pending |
| INST-02 | Phase 4 | Pending |
| INST-03 | Phase 4 | Pending |
| INST-04 | Phase 4 | Pending |
| INST-05 | Phase 4 | Pending |
| INST-06 | Phase 4 | Pending |
| INST-07 | Phase 4 | Pending |
| AUTH-01 | Phase 4 | Pending |
| AUTH-02 | Phase 4 | Pending |
| AUTH-03 | Phase 4 | Pending |
| AUTH-04 | Phase 1 | Complete |
| SSHK-01 | Phase 5 | Pending |
| SSHK-02 | Phase 5 | Pending |
| SSHK-03 | Phase 5 | Pending |
| SSHK-04 | Phase 5 | Pending |
| BILL-01 | Phase 5 | Pending |
| BILL-02 | Phase 5 | Pending |
| BILL-03 | Phase 5 | Pending |
| BILL-04 | Phase 5 | Pending |
| BILL-05 | Phase 5 | Pending |
| BILL-06 | Phase 5 | Pending |
| AVAIL-01 | Phase 6 | Pending |
| AVAIL-02 | Phase 6 | Pending |
| AVAIL-03 | Phase 6 | Pending |
| AVAIL-04 | Phase 6 | Pending |
| AVAIL-05 | Phase 6 | Pending |
| API-01 | Phase 4 | Pending |
| API-02 | Phase 4 | Pending |
| API-03 | Phase 4 | Pending |
| API-04 | Phase 4 | Pending |
| API-05 | Phase 6 | Pending |
| API-06 | Phase 5 | Pending |
| API-07 | Phase 5 | Pending |
| API-08 | Phase 1 | Complete |
| API-09 | Phase 4 | Pending |
| HLTH-01 | Phase 6 | Pending |
| HLTH-02 | Phase 6 | Pending |
| HLTH-03 | Phase 6 | Pending |
| DASH-01 | Phase 7 | Pending |
| DASH-02 | Phase 7 | Pending |
| DASH-03 | Phase 7 | Pending |
| DASH-04 | Phase 7 | Pending |
| DASH-05 | Phase 7 | Pending |
| DASH-06 | Phase 7 | Pending |
| DASH-07 | Phase 7 | Pending |
| DASH-08 | Phase 7 | Pending |

**Coverage:**
- v1 requirements: 65 total
- Mapped to phases: 65
- Unmapped: 0

---
*Requirements defined: 2026-02-24*
*Last updated: 2026-02-24 after roadmap creation*
