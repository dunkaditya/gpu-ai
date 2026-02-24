# External Integrations

**Analysis Date:** 2026-02-24

## APIs & External Services

**GPU Provider APIs:**
- **RunPod**
  - What: Primary instant GPU provider (US)
  - SDK/Client: RunPod GraphQL API
  - Auth: Bearer token via `RUNPOD_API_KEY`
  - Base URL: `https://api.runpod.io/graphql`
  - Adapter: `internal/provider/runpod/adapter.go`
  - Boot time: ~15 seconds (FlashBoot)
  - Spot support: Via "Community Cloud" instances
  - Features: On-demand + spot, per-second billing, no egress fees

- **E2E Networks**
  - What: India-based GPU provider (cost arbitrage, 30-50% cheaper)
  - SDK/Client: E2E Networks REST API
  - Auth: Bearer token via `E2E_API_KEY`
  - Base URL: `https://api.e2enetworks.com/myaccount/api/v1/`
  - Rate limit: 5,000 requests/hr per token
  - Boot time: ~10 seconds
  - GPU fleet: H200 SXM, H100 80GB, A100 80GB/40GB, L40S, A40, L4
  - Data centers: Mumbai, Delhi, Bangalore
  - Certifications: SOC2, ISO 27001, PCI DSS
  - Future adapter: `internal/provider/e2e_networks/adapter.go` (not yet implemented)

- **Lambda Labs** (planned)
  - What: Alternative US provider
  - Adapter: `internal/provider/lambda/adapter.go` (not yet implemented)

**Provider Interface:**
- Location: `internal/provider/provider.go`
- All providers implement: `ListAvailable()`, `Provision()`, `GetStatus()`, `Terminate()`
- Response types in `internal/provider/types.go`

## Data Storage

**Databases:**

- **PostgreSQL**
  - Connection: `DATABASE_URL` env var
  - Client: `github.com/jackc/pgx/v5/pgxpool` (connection pool)
  - Location: `internal/db/pool.go`
  - Pool config: 5-20 connections (default)
  - Tables:
    - `organizations` - Customer accounts, Stripe customer IDs
    - `users` - Team members, roles (admin/member)
    - `ssh_keys` - SSH public keys per user
    - `instances` - Provisioned GPU instances, maps customer ID to upstream provider ID
    - `environments` - Saved Docker images + configs
    - `usage_records` - Billing records per instance
  - Key fields:
    - `instances.upstream_provider` - Hidden from customer, used internally (runpod, e2e_networks, lambda)
    - `instances.upstream_id` - Provider's instance ID (never exposed)
    - `instances.wireguard_public_key/private_key` - Encrypted tunnel keys
    - `organizations.stripe_customer_id` - Links org to Stripe for billing

- **Redis**
  - Connection: `REDIS_URL` env var
  - Client: Redis client (not yet fully specified in codebase)
  - Location: `internal/availability/cache.go`
  - Purpose: Cache GPU offerings + pricing from all providers
  - TTL: 60 seconds per offering entry
  - Key pattern: `gpu:{provider}:{gpu_type}:{tier}:{region}`
  - Polling interval: Every 30 seconds via `internal/availability/poller.go`

**File Storage:**
- **AWS ECR** (planned) - Customer Docker images for environment persistence
- Configuration: Not yet implemented, mentioned in `docs/ARCHITECTURE.md`

**Caching:**
- Redis (see above)
- TTL strategy: 60s for availability data, refreshed every 30s

## Authentication & Identity

**Auth Provider:**
- **Clerk**
  - SDK/Client: Clerk JWT verification
  - Auth: `CLERK_SECRET_KEY` env var
  - Location: `internal/auth/clerk.go`
  - Implementation: JWT middleware that:
    - Extracts Bearer token from Authorization header
    - Verifies JWT signature against Clerk JWKS
    - Parses claims: `user_id`, `org_id`, `email`, `role` (admin/member)
    - Injects Claims into request context
    - Returns 401 if invalid/missing
  - Routes: Public API routes require Clerk JWT (`GET/POST /api/v1/*`)

**Internal Service Auth:**
- Custom token-based auth for non-customer endpoints
- Token: `INTERNAL_API_TOKEN` env var
- Used by: Cloud-init callback (`POST /internal/instances/{id}/ready`), health checks (`POST /internal/instances/{id}/health`)
- Pattern: Bearer token in Authorization header

## Billing & Payments

**Billing Provider:**
- **Stripe**
  - SDK/Client: Stripe API
  - Auth: `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET` env vars
  - Location: `internal/billing/stripe.go`
  - Features:
    - Usage metering (GPU hours billed)
    - Payment method validation
    - Invoice generation
    - Unpaid invoice detection
  - Webhooks:
    - `invoice.paid` - Handle successful payment
    - `invoice.payment_failed` - Handle failed payment
    - `customer.subscription.deleted` - Handle cancelled subscription
  - Webhook route: `POST /api/v1/billing/webhook` (signature verified, no auth required)
  - Pricing: Margin model: pay upstream provider at spot/on-demand rate, charge customer higher rate
    - `instances.price_per_hour` - What customer is charged
    - `instances.upstream_price_per_hour` - What we pay upstream
    - Margin: difference used for operating costs + profit
  - Usage tracking: `usage_records` table stores start/end times, computed costs

## Monitoring & Observability

**Error Tracking:**
- Not detected in codebase

**Logs:**
- Structured logging via `log/slog` (stdlib, Go 1.21+)
- Logger: `*slog.Logger` pattern throughout codebase
- Example usage in `cmd/gpuctl/main.go`: `log.Printf()` for startup/shutdown messages
- Future: All packages (`internal/*`) will use `slog.Logger` for structured logs

**Health Checks:**
- Endpoint: `GET /health` (no auth required)
- Response: `{"status":"ok"}`
- Location: `cmd/gpuctl/main.go`
- Monitoring: `internal/health/monitor.go` - Periodic instance health checks from upstream

## CI/CD & Deployment

**Hosting:**
- Not specified in codebase (self-hosted Linux VPS or cloud VMs)
- Deployment artifact: Single binary `gpuctl`

**CI Pipeline:**
- Not detected (no GitHub Actions, GitLab CI, or similar config files)

**Build Process:**
- Command: `go build -o gpuctl ./cmd/gpuctl`
- Testing: `go test ./...`
- Linting: `golangci-lint run`
- Makefile: See `Makefile` for convenience targets

## Environment Configuration

**Required env vars (production):**
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `CLERK_SECRET_KEY` - Clerk JWT verification
- `STRIPE_SECRET_KEY` - Stripe payments
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook signature verification
- `RUNPOD_API_KEY` - RunPod provider access
- `E2E_API_KEY` - E2E Networks provider access
- `WG_PROXY_PUBLIC_KEY` - WireGuard tunnel public key
- `WG_PROXY_ENDPOINT` - WireGuard tunnel endpoint (proxy IP)
- `WG_SUBNET` - WireGuard network range (default: `10.0.0.0/24`)
- `GPUCTL_PORT` - Server port (default: `9090`)
- `INTERNAL_API_TOKEN` - Token for internal callbacks

**Secrets location:**
- All secrets stored as environment variables (12-factor app pattern)
- `.env` file for local development (copy from `.env.example`)
- Production: Use container secret management (Docker Secrets, Kubernetes Secrets, AWS Systems Manager, etc.)

## Webhooks & Callbacks

**Incoming (Stripe):**
- `POST /api/v1/billing/webhook` - Stripe event notifications
  - Events: `invoice.paid`, `invoice.payment_failed`, `customer.subscription.deleted`
  - Signature verification required (Stripe HMAC)
  - No Clerk authentication needed

**Outgoing (Cloud-Init Callback):**
- `POST /internal/instances/{id}/ready` - Called by cloud-init script after instance boots
  - Auth: `INTERNAL_API_TOKEN` Bearer token
  - Triggered by: `internal/infra/cloud-init/template` (bootstrap script)
  - Signals: Instance is fully provisioned and WireGuard tunnel is ready

**Outgoing (Instance Health):**
- `POST /internal/instances/{id}/health` - Called by running instance for health monitoring
  - Auth: `INTERNAL_API_TOKEN` Bearer token
  - Triggered by: Health check process on instance
  - Purpose: Track uptime, detect hung instances

## VPN & Network Layer

**WireGuard Tunnel:**
- Location: `internal/wireguard/manager.go`
- Purpose: Privacy layer - customer traffic encrypted, upstream provider never visible
- Configuration:
  - Interface: `wg0` (default, configurable)
  - Network: `10.0.0.0/24` (configurable via `WG_SUBNET`)
  - Proxy endpoint: `WG_PROXY_ENDPOINT` (proxy server public IP)
  - Proxy public key: `WG_PROXY_PUBLIC_KEY`
  - Per-instance keys: Generated by `internal/wireguard/keygen.go`
  - Persistent keepalive: 25 seconds
- Flow:
  1. Instance boots, cloud-init generates WireGuard keypair
  2. Instance configures `/etc/wireguard/wg0.conf` with proxy details
  3. Instance starts WireGuard tunnel (`systemctl start wg-quick@wg0`)
  4. Proxy server detects new peer via `wg show` command
  5. Manager registers peer in database (`instances.wireguard_public_key`, `instances.wireguard_address`)
  6. Customer SSH to `gpu-{id}.gpu.ai` → resolves to proxy IP → routed through tunnel to instance
- Firewall rules (on instance): `iptables` - only allow traffic via WireGuard tunnel, block direct access

---

*Integration audit: 2026-02-24*
