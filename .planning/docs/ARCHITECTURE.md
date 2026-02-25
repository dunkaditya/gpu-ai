# GPU.ai — Technical Architecture

## Overview

GPU.ai is a GPU cloud aggregation platform that provides a unified interface
for provisioning GPU instances across multiple upstream providers. Customers
interact with GPU.ai as if it were a single cloud provider — they never see
the upstream source.

**Phase 1:** Bare metal aggregation with re-rented upstream providers
**Phase 2:** Own hardware (India) + virtualization (MIG partitioning)
**Phase 3:** Predictive resource allocation — job-based pricing with memory prediction

This document covers the Phase 1 architecture in detail.

---

## Core Value Proposition

1. **Real-time availability + price comparison** across providers (US-West, India, Spot)
2. **One-click environment persistence** — Docker images stored in container registry, deployable across any provider
3. **Privacy layer** — customer never knows or sees the upstream provider
4. **India cost arbitrage** — Indian GPU providers are 30-50% cheaper than US equivalents

---

## Language Split

**Go (production services):** Single binary handles all live operations — API server,
provisioning engine, provider adapters, WireGuard management, availability polling,
health monitoring, Stripe billing, Clerk auth. Everything in the customer request
path or managing infrastructure in real-time.

**Python (offline tooling):** Database migrations, data analysis/reporting, usage
aggregation jobs, seed scripts, and eventually Phase 3 ML prediction models.
Runs on a schedule or manually, never in the hot path.

---

## Project Structure

```
gpu-ai/
├── CLAUDE.md
├── .gitignore
├── .env.example
├── .claude/
│   └── settings.json               # MCP servers (project-scoped)
│
├── cmd/                             # Go entrypoints
│   └── gpuctl/
│       └── main.go                  # Wires deps, starts server + goroutines
│
├── internal/                        # Go application code (not importable externally)
│   ├── api/
│   │   ├── server.go                # HTTP server setup, middleware, CORS
│   │   ├── router.go                # Route registration
│   │   └── handlers/
│   │       ├── instances.go         # POST/GET/DELETE /api/v1/instances
│   │       ├── gpu.go               # GET /api/v1/gpu/available
│   │       ├── ssh_keys.go          # CRUD /api/v1/ssh-keys
│   │       ├── billing.go           # GET /api/v1/billing/usage
│   │       └── health.go            # GET /health
│   │
│   ├── provider/
│   │   ├── provider.go              # Provider interface
│   │   ├── types.go                 # GPUOffering, ProvisionResult, InstanceStatus
│   │   ├── registry.go              # Adapter registry (map of active providers)
│   │   └── runpod/
│   │       ├── adapter.go           # RunPod adapter implementation
│   │       └── client.go            # RunPod HTTP client (API calls)
│   │
│   ├── provision/
│   │   └── engine.go                # Orchestration: query availability → pick adapter
│   │                                #   → generate WG keys → inject cloud-init
│   │                                #   → call adapter → write DB → return result
│   │
│   ├── availability/
│   │   ├── poller.go                # 30s ticker goroutine, polls all adapters → Redis
│   │   └── cache.go                 # Redis read/write helpers for GPU offerings
│   │
│   ├── wireguard/
│   │   ├── manager.go               # Add/remove WireGuard peers on proxy server
│   │   └── keygen.go                # WireGuard key pair generation
│   │
│   ├── billing/
│   │   ├── stripe.go                # Stripe customer/charge/usage-record operations
│   │   └── metering.go              # Per-second usage tracking, billing start/stop
│   │
│   ├── auth/
│   │   ├── clerk.go                 # Clerk JWT verification middleware
│   │   └── middleware.go            # Auth middleware, API key validation
│   │
│   ├── db/
│   │   ├── pool.go                  # pgx connection pool creation/teardown
│   │   ├── instances.go             # Instance CRUD queries
│   │   ├── organizations.go         # Org/user CRUD queries
│   │   └── ssh_keys.go              # SSH key CRUD queries
│   │
│   └── config/
│       └── config.go                # Env var loading into typed config struct
│
├── go.mod
├── go.sum
├── Makefile                         # build, run, test, lint targets
│
├── tools/                           # Python offline tooling
│   ├── requirements.txt
│   ├── migrate.py                   # Database migration runner
│   ├── seed.py                      # Dev/test data seeding
│   └── reports/                     # Usage analytics, billing reconciliation
│       └── .gitkeep
│
├── database/
│   └── migrations/                  # SQL migration files (applied by tools/migrate.py)
│       └── 001_init.sql             # Initial schema
│
├── infra/
│   ├── cloud-init/
│   │   └── bootstrap.sh             # Cloud-init template (variable injection by Go)
│   └── wireguard/
│       └── .gitkeep
│
├── frontend/                        # Next.js (created later)
│   └── .gitkeep
│
└── docs/
    └── ARCHITECTURE.md
```

---

## System Architecture

```
                    ┌──────────────────────────────┐
                    │       Customer Dashboard      │
                    │          (Next.js)            │
                    └──────────────┬───────────────┘
                                   │ HTTPS
                                   ▼
┌──────────────────────────────────────────────────────────────┐
│                    gpuctl (single Go binary)                  │
│                                                               │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────────┐  │
│  │  API     │ │ Billing  │ │   Auth   │ │  Availability   │  │
│  │ Handlers │ │ (Stripe) │ │ (Clerk)  │ │  Poller (30s)   │  │
│  └────┬─────┘ └──────────┘ └──────────┘ └────────┬────────┘  │
│       │                                           │           │
│       ▼                                           ▼           │
│  ┌──────────────┐                         ┌────────────┐     │
│  │ Provisioning │                         │   Redis    │     │
│  │   Engine     │                         │   Cache    │     │
│  └──────┬───────┘                         └────────────┘     │
│         │                                                     │
│    ┌────┴──────────────────────┐                             │
│    │    Provider Adapters      │                             │
│    │  ┌─────────┐ ┌────────┐  │                             │
│    │  │ RunPod  │ │  E2E   │  │                             │
│    │  └────┬────┘ └───┬────┘  │                             │
│    └───────┼──────────┼───────┘                             │
│            │          │                                      │
│  ┌─────────┴──────────┴────────┐     ┌──────────────────┐   │
│  │    WireGuard Manager        │     │    PostgreSQL     │   │
│  │  (peer add/remove/keygen)   │     │    (pgx pool)    │   │
│  └─────────────────────────────┘     └──────────────────┘   │
└──────────────────────────────────────────────────────────────┘
         │               │
         │ WireGuard      │ Upstream APIs
         ▼               ▼
┌──────────────┐  ┌─────────────────┐
│  GPU.ai      │  │ RunPod / E2E /  │
│  Proxy       │  │ Lambda APIs     │
│  Server      │  └─────────────────┘
└──────┬───────┘
       │ WireGuard tunnels
       ▼
┌─────────────────────────────────────┐
│       Upstream GPU Instances         │
│  (cloud-init: WG tunnel, SSH keys,  │
│   hostname, MOTD, firewall)         │
└─────────────────────────────────────┘
```

---

## Provider Adapter Interface

Defined in `internal/provider/provider.go`:

```go
type Provider interface {
    // Name returns the provider identifier (e.g., "runpod", "e2e")
    Name() string

    // ListAvailable polls the provider for current GPU inventory and pricing
    ListAvailable(ctx context.Context) ([]GPUOffering, error)

    // Provision creates a new instance with the given config.
    // The cloud-init script (WireGuard, SSH keys, hostname, MOTD, firewall)
    // is injected into the instance boot process.
    Provision(ctx context.Context, config InstanceConfig) (*ProvisionResult, error)

    // GetStatus returns the current status of an upstream instance
    GetStatus(ctx context.Context, upstreamID string) (*InstanceStatus, error)

    // Terminate destroys an upstream instance
    Terminate(ctx context.Context, upstreamID string) error
}
```

### Shared Types (`internal/provider/types.go`)

```go
type GPUType string

const (
    H100SXM   GPUType = "h100_sxm"
    H100PCIE  GPUType = "h100_pcie"
    H200SXM   GPUType = "h200_sxm"
    A100_80GB GPUType = "a100_80gb"
    A100_40GB GPUType = "a100_40gb"
    L40S      GPUType = "l40s"
    RTX4090   GPUType = "rtx_4090"
)

type InstanceTier string

const (
    OnDemand InstanceTier = "on_demand"
    Spot     InstanceTier = "spot"
    Reserved InstanceTier = "reserved"
)

type GPUOffering struct {
    Provider       string       `json:"provider"`
    GPUType        GPUType      `json:"gpu_type"`
    GPUCount       int          `json:"gpu_count"`
    VRAMPerGPU     int          `json:"vram_per_gpu_gb"`
    PricePerHour   float64      `json:"price_per_hour"`
    Tier           InstanceTier `json:"tier"`
    Region         string       `json:"region"`
    AvailableCount int          `json:"available_count"`
}

type InstanceConfig struct {
    GPUType       GPUType      `json:"gpu_type"`
    GPUCount      int          `json:"gpu_count"`
    Tier          InstanceTier `json:"tier"`
    Region        string       `json:"region"`
    SSHPublicKeys []string     `json:"ssh_public_keys"`
    DockerImage   string       `json:"docker_image,omitempty"`
}

type ProvisionResult struct {
    UpstreamID            string `json:"upstream_id"`
    UpstreamIP            string `json:"upstream_ip"`
    Provider              string `json:"provider"`
    Status                string `json:"status"`
    EstimatedReadySeconds int    `json:"estimated_ready_seconds"`
}

type InstanceStatus struct {
    UpstreamID string `json:"upstream_id"`
    Status     string `json:"status"`
    IP         string `json:"ip,omitempty"`
}
```

### Provider Registry (`internal/provider/registry.go`)

```go
type Registry struct {
    providers map[string]Provider
}

func NewRegistry() *Registry {
    return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) {
    r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
    p, ok := r.providers[name]
    return p, ok
}

func (r *Registry) All() []Provider {
    out := make([]Provider, 0, len(r.providers))
    for _, p := range r.providers {
        out = append(out, p)
    }
    return out
}
```

---

## Provisioning Flow (End to End)

```
1. Customer hits POST /api/v1/instances
   - Body: { gpu_type, gpu_count, region, tier, ssh_public_keys, docker_image? }

2. Auth middleware verifies Clerk JWT → extracts user_id, org_id

3. Handler validates input, calls provisioning engine

4. Provisioning engine:
   a. Queries Redis cache for matching GPU offerings
   b. Picks best available (price, region, availability)
   c. Generates WireGuard key pair for this instance
   d. Renders cloud-init template with:
      - WireGuard private key + proxy public key
      - SSH authorized keys
      - Instance ID for hostname
      - MOTD branding
      - Firewall rules
   e. Calls selected provider adapter's Provision() method
   f. Adapter sends POST to upstream API (e.g., RunPod create pod)
      with cloud-init script injected

5. Upstream instance boots (~15s RunPod, ~10s E2E)
   - cloud-init executes: WireGuard → SSH keys → hostname → firewall
   - WireGuard tunnel establishes to GPU.ai proxy

6. Provisioning engine writes to PostgreSQL:
   - Instance record: id, org_id, user_id, upstream_provider, upstream_id,
     wireguard keys, hostname, gpu config, pricing, status, billing_start

7. WireGuard manager adds peer to proxy server config

8. Instance calls back: POST /internal/instances/{id}/ready

9. Customer receives:
   {
     "instance_id": "gpu-4a7f",
     "hostname": "gpu-4a7f.gpu.ai",
     "ssh_command": "ssh user@gpu-4a7f.gpu.ai",
     "status": "running",
     "gpu_type": "h100_sxm",
     "gpu_count": 8,
     "price_per_hour": 2.12,
     "tier": "on_demand",
     "region": "us-west"
   }
```

---

## Privacy / Network Layer

Every customer connection routes through GPU.ai infrastructure. The customer
never interacts with or sees the upstream provider.

### WireGuard Tunnel Architecture

```
Customer's Machine
    │
    │ SSH to gpu-4a7f.gpu.ai (DNS → GPU.ai proxy IP)
    │
    ▼
┌─────────────────────────────┐
│   GPU.ai Proxy / Bastion    │
│                             │
│   Public IP: 203.0.113.10  │
│   WireGuard: 10.0.0.1      │
│                             │
│   Routes SSH to correct     │
│   upstream via WireGuard    │
└─────────────┬───────────────┘
              │ WireGuard (encrypted)
              ▼
┌─────────────────────────────┐
│   Upstream Instance          │
│   (e.g., RunPod pod)        │
│                             │
│   Real IP: hidden           │
│   WireGuard: 10.0.0.x      │
│   Hostname: gpu-4a7f.gpu.ai│
│   Firewall: WireGuard only  │
└─────────────────────────────┘
```

### What the Customer Sees
- **Hostname:** `gpu-4a7f.gpu.ai`
- **SSH:** `ssh user@gpu-4a7f.gpu.ai`
- **MOTD:** GPU.ai branded welcome message
- **IP:** GPU.ai proxy IP only

### What's Hidden
- Upstream provider identity
- Upstream instance IP
- Upstream instance ID
- Provider-specific metadata

---

## Cloud-Init Boot Script

Located at `infra/cloud-init/bootstrap.sh`. The Go provisioning engine reads
this template, injects instance-specific variables via `text/template`, and
passes it to the provider adapter.

```bash
#!/bin/bash
set -euo pipefail

# Variables injected by Go provisioning engine
INSTANCE_ID="{{.InstanceID}}"
PROXY_ENDPOINT="{{.ProxyEndpoint}}"
PROXY_PUBLIC_KEY="{{.ProxyPublicKey}}"
INSTANCE_PRIVATE_KEY="{{.InstancePrivateKey}}"
INSTANCE_ADDRESS="{{.InstanceAddress}}"
SSH_AUTHORIZED_KEYS="{{.SSHAuthorizedKeys}}"
DOCKER_IMAGE="{{.DockerImage}}"

# 1. Install WireGuard
apt-get update -qq && apt-get install -y -qq wireguard

# 2. Configure WireGuard tunnel
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = ${INSTANCE_PRIVATE_KEY}
Address = ${INSTANCE_ADDRESS}

[Peer]
PublicKey = ${PROXY_PUBLIC_KEY}
Endpoint = ${PROXY_ENDPOINT}:51820
AllowedIPs = 10.0.0.0/24
PersistentKeepalive = 25
EOF

chmod 600 /etc/wireguard/wg0.conf
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

# 3. Set hostname
hostnamectl set-hostname "gpu-${INSTANCE_ID}.gpu.ai"

# 4. Set MOTD
cat > /etc/motd << 'MOTD'

  ██████╗ ██████╗ ██╗   ██╗   █████╗ ██╗
 ██╔════╝ ██╔══██╗██║   ██║  ██╔══██╗██║
 ██║  ███╗██████╔╝██║   ██║  ███████║██║
 ██║   ██║██╔═══╝ ██║   ██║  ██╔══██║██║
 ╚██████╔╝██║     ╚██████╔╝  ██║  ██║██║
  ╚═════╝ ╚═╝      ╚═════╝   ╚═╝  ╚═╝╚═╝

 Welcome to GPU.ai
 Instance: gpu-${INSTANCE_ID}
 Support: support@gpu.ai

MOTD

# 5. Configure SSH keys
mkdir -p /root/.ssh
echo "${SSH_AUTHORIZED_KEYS}" > /root/.ssh/authorized_keys
chmod 700 /root/.ssh
chmod 600 /root/.ssh/authorized_keys

# 6. Firewall — only allow traffic via WireGuard
iptables -A INPUT -i wg0 -j ACCEPT
iptables -A INPUT -p udp --dport 51820 -j ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -j DROP

# 7. Optional: pull and run Docker image
if [ -n "${DOCKER_IMAGE}" ]; then
    docker pull "${DOCKER_IMAGE}"
    docker run -d --gpus all --name workspace "${DOCKER_IMAGE}"
fi

# 8. Signal ready
curl -s -X POST "https://api.gpu.ai/internal/instances/${INSTANCE_ID}/ready" \
    -H "Authorization: Bearer {{.InternalToken}}"
```

---

## Availability Engine

Background goroutine that polls all registered provider adapters every 30
seconds and caches results in Redis.

### Poller (`internal/availability/poller.go`)

```go
func (p *Poller) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.poll(ctx)
        }
    }
}

func (p *Poller) poll(ctx context.Context) {
    for _, provider := range p.registry.All() {
        offerings, err := provider.ListAvailable(ctx)
        if err != nil {
            log.Error("poll failed", "provider", provider.Name(), "err", err)
            continue
        }
        for _, o := range offerings {
            key := fmt.Sprintf("gpu:%s:%s:%s:%s", o.Provider, o.GPUType, o.Tier, o.Region)
            data, _ := json.Marshal(o)
            p.redis.Set(ctx, key, data, 60*time.Second)
        }
    }
}
```

### Customer-Facing Response

The API handler reads from Redis and strips provider identity:

```json
{
  "available": [
    {
      "gpu_type": "h100_sxm",
      "gpu_count": 8,
      "vram_per_gpu_gb": 80,
      "price_per_hour": 2.12,
      "tier": "on_demand",
      "region": "us-west",
      "available_count": 12
    },
    {
      "gpu_type": "h100_sxm",
      "gpu_count": 8,
      "vram_per_gpu_gb": 80,
      "price_per_hour": 1.40,
      "tier": "on_demand",
      "region": "india-mumbai",
      "available_count": 8
    }
  ]
}
```

---

## Database Schema

Applied via SQL migration files in `database/migrations/`. Managed by
`tools/migrate.py`.

```sql
-- 001_init.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Organizations
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    clerk_user_id VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- SSH keys
CREATE TABLE ssh_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255),
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- GPU instances
CREATE TABLE instances (
    id VARCHAR(12) PRIMARY KEY,
    org_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),

    -- Upstream (hidden from customer)
    upstream_provider VARCHAR(50) NOT NULL,
    upstream_id VARCHAR(255) NOT NULL,
    upstream_ip INET,

    -- GPU.ai facing
    hostname VARCHAR(255) NOT NULL,
    wg_public_key VARCHAR(255),
    wg_private_key_enc TEXT,
    wg_address INET,

    -- Configuration
    gpu_type VARCHAR(50) NOT NULL,
    gpu_count INT NOT NULL,
    tier VARCHAR(20) NOT NULL,
    region VARCHAR(50) NOT NULL,

    -- Billing
    price_per_hour NUMERIC(10, 4) NOT NULL,
    upstream_price_per_hour NUMERIC(10, 4) NOT NULL,
    billing_start TIMESTAMPTZ,
    billing_end TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) DEFAULT 'creating',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    terminated_at TIMESTAMPTZ
);

-- Saved environments
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    docker_image VARCHAR(512),
    description TEXT,
    is_shared BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Usage records
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id VARCHAR(12) REFERENCES instances(id),
    org_id UUID REFERENCES organizations(id),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    gpu_type VARCHAR(50),
    gpu_count INT,
    price_per_hour NUMERIC(10, 4),
    total_cost NUMERIC(10, 4),
    stripe_usage_record_id VARCHAR(255)
);

-- Indexes
CREATE INDEX idx_instances_org ON instances(org_id);
CREATE INDEX idx_instances_status ON instances(status);
CREATE INDEX idx_instances_user ON instances(user_id);
CREATE INDEX idx_usage_org ON usage_records(org_id);
CREATE INDEX idx_usage_instance ON usage_records(instance_id);
CREATE INDEX idx_ssh_keys_user ON ssh_keys(user_id);
```

---

## API Endpoints

All served by gpuctl on `:8080`.

```
# Health
GET    /health

# Instances (Clerk JWT required)
POST   /api/v1/instances
GET    /api/v1/instances
GET    /api/v1/instances/{id}
DELETE /api/v1/instances/{id}

# Availability
GET    /api/v1/gpu/available
GET    /api/v1/gpu/available?type=h100_sxm&tier=spot

# SSH Keys
GET    /api/v1/ssh-keys
POST   /api/v1/ssh-keys
DELETE /api/v1/ssh-keys/{id}

# Environments
GET    /api/v1/environments
POST   /api/v1/environments
DELETE /api/v1/environments/{id}

# Billing
GET    /api/v1/billing/usage
GET    /api/v1/billing/invoices

# Internal (localhost only)
POST   /internal/instances/{id}/ready
POST   /internal/instances/{id}/health
```

---

## Component Stack

| Component | Tech | Location |
|-----------|------|----------|
| API Server | Go (Chi or stdlib) | `internal/api/` |
| Provider Adapters | Go | `internal/provider/` |
| Provisioning Engine | Go | `internal/provision/` |
| Availability Poller | Go (goroutine + Redis) | `internal/availability/` |
| WireGuard Manager | Go | `internal/wireguard/` |
| Billing | Go + Stripe SDK | `internal/billing/` |
| Auth | Go + Clerk JWT | `internal/auth/` |
| Database | Go + pgx | `internal/db/` |
| Proxy / Bastion | Linux + WireGuard | `infra/` |
| Cloud-Init | Bash (Go-templated) | `infra/cloud-init/` |
| Migrations | SQL + Python runner | `database/` + `tools/` |
| Frontend | Next.js (later) | `frontend/` |

---

## Three-Tier Provider Strategy

| Tier | Provider | Use Case | Provisioning | Margin |
|------|----------|----------|--------------|--------|
| Instant US | RunPod | On-demand/spot training, inference, dev | Seconds | Thin/zero |
| Instant India | E2E Networks | Cost-sensitive training, data sovereignty | ~10 seconds | Strong |
| Reserved | Novacore/CTRLS | 64+ GPU multi-node, long commitments | 1-2 months | Best |

Instant tiers are fully automated via upstream APIs.
Reserved tier is ops-assisted with a "Contact Us" flow.

---

## RunPod Adapter Specifics

- **API:** REST + GraphQL endpoints
- **Auth:** Bearer token via `RUNPOD_API_KEY`
- **Spot:** Community Cloud (lower cost, can be reclaimed with 5s SIGTERM)
- **On-demand:** Secure Cloud (guaranteed non-interruption)
- **Billing:** Per-second, no egress fees
- **Boot time:** ~15 seconds (FlashBoot)
- **Cloud-init:** Via custom Docker images + startup scripts

## E2E Networks Adapter Specifics

- **API:** REST `https://api.e2enetworks.com/myaccount/api/v1/`
- **Rate limit:** 5,000 req/hr per token
- **GPU fleet:** H200 SXM, H100 80GB SXM, A100 80GB/40GB, L40S
- **Boot time:** ~10 seconds
- **Data centers:** Mumbai, Delhi, Bangalore
- **Pricing:** Starting ₹49/hr, up to 60% cheaper than hyperscalers

---

## Security (Phase 1)

- WireGuard encryption on all instance traffic
- SSH key-only authentication (password auth disabled)
- Instance firewall: only WireGuard tunnel traffic allowed
- Hostname/IP sanitization — no upstream information exposed
- Clerk JWT verification on all customer API endpoints
- Internal endpoints restricted to localhost
- WireGuard private keys encrypted at rest in PostgreSQL
- Per-second billing metering to prevent abuse

### Future (Phase 2)
- NVIDIA Confidential Computing (TEE on Hopper/Blackwell)
- Hardware attestation reports
- CloudHSM key management gated on GPU attestation

---

## Build Order

1. **Go project scaffold** — `cmd/gpuctl/main.go`, config loading, health endpoint
2. **RunPod adapter** — prove provisioning works end-to-end via upstream API
3. **WireGuard privacy layer** — proxy server + cloud-init tunnel setup
4. **Database + instance management** — pgx pool, instance CRUD, migrations
5. **Auth + billing** — Clerk JWT middleware, Stripe usage billing
6. **Availability engine** — Redis poller across providers
7. **API routes** — full instance lifecycle (create, list, get, terminate)
8. **Landing page + dashboard** — Next.js frontend
9. **Environment persistence** — Docker image save/deploy across providers
10. **E2E Networks adapter** — India provider integration
11. **Closed beta** — 10-20 users, iterate

---

## Phase 2 & 3 (Future)

**Phase 2: Own Hardware + Virtualization**
- Deploy on Novacore/CTRLS bare metal in India
- NVIDIA MIG partitioning for GPU memory slicing
- Add `gpuai_native` provider adapter to existing registry
- Same API, billing, auth — just another provider option with best margins

**Phase 3: Predictive Resource Allocation**
- Customers submit jobs, not infrastructure requests
- GPU.ai predicts memory/compute/time requirements from job metadata + historical data
- Phase 2 virtualization enables exact-sized memory slices (not fixed GPU tiers)
- Pricing based on predicted resource consumption over flexible time windows
- OOM failures auto-restart from checkpoint on GPU.ai's dime
- Customer stops thinking about GPUs, starts thinking about job completion
- Margin depends on prediction accuracy — better model = tighter allocations = higher utilization
