# Technology Stack

**Analysis Date:** 2026-02-24

## Languages

**Primary:**
- Go 1.22+ - Main binary (`gpuctl`), all production business logic in `internal/` packages

**Secondary:**
- Python 3.x - Tooling only, used for database migrations, data seeding, and reporting in `tools/`

## Runtime

**Environment:**
- Linux (primary deployment target)
- Supports cloud-init for instance provisioning

**Package Manager:**
- Go modules (`go.mod`)
- pip (for Python tooling dependencies)

## Frameworks

**Core:**
- stdlib `net/http` - HTTP server with Go 1.22+ routing patterns (`GET /path/{id}`)
- No web frameworks (Go) or API frameworks (FastAPI, etc.) — stdlib only

**Build/Dev:**
- `golangci-lint` - Linting tool (referenced in `Makefile`)
- `make` - Build system (`Makefile`)

## Key Dependencies

**Critical:**
- `github.com/jackc/pgx/v5/pgxpool` - PostgreSQL connection pooling (referenced in `internal/db/pool.go`)
- Redis client - For GPU availability caching (imported but not yet specified, used in `internal/availability/cache.go`)
- Stripe SDK - Payment and usage metering integration (referenced in `internal/billing/stripe.go`)
- Clerk SDK - JWT verification for authentication (referenced in `internal/auth/clerk.go`)

**Infrastructure:**
- No explicit Go dependencies listed in `go.mod` yet (file shows minimal content)
- Python dependencies in `tools/requirements.txt`:
  - `psycopg2-binary>=2.9` - PostgreSQL adapter for Python migrations
  - `click>=8.1` - CLI framework for tooling scripts
  - `tabulate>=0.9` - Output formatting for reports

## Configuration

**Environment:**
- All configuration via environment variables (see `.env.example`)
- Required vars:
  - `DATABASE_URL` - PostgreSQL connection string (format: `postgresql://user:pass@host:port/dbname`)
  - `REDIS_URL` - Redis connection string (format: `redis://host:port/0`)
  - `GPUCTL_PORT` - Server port (default: `9090`)
  - `CLERK_SECRET_KEY` - Clerk authentication secret
  - `STRIPE_SECRET_KEY` - Stripe API secret key
  - `STRIPE_WEBHOOK_SECRET` - Stripe webhook signing secret
  - `RUNPOD_API_KEY` - RunPod provider API key
  - `E2E_API_KEY` - E2E Networks provider API key
  - `WG_PROXY_PUBLIC_KEY` - WireGuard proxy server public key
  - `WG_PROXY_ENDPOINT` - WireGuard proxy public IP/endpoint
  - `WG_SUBNET` - WireGuard network subnet (default: `10.0.0.0/24`)
  - `INTERNAL_API_TOKEN` - Token for internal service callbacks (cloud-init, health checks)

**Build:**
- `Makefile` - Targets: `build`, `run`, `test`, `lint`, `clean`
- No build config files (Webpack, tsconfig, etc.) — pure Go project

## Platform Requirements

**Development:**
- Go 1.22+
- PostgreSQL 12+
- Redis 6+
- golangci-lint (for linting)
- `wg` command-line tool (for WireGuard management in `internal/wireguard/`)

**Production:**
- Linux host running `gpuctl` binary on port 9090
- PostgreSQL 12+ for persistent data
- Redis 6+ for GPU availability cache
- WireGuard kernel module and tools (for VPN tunnel management)
- Python 3.7+ available for administrative tools (migrations, seeding)

## Server Behavior

**HTTP Server:**
- Location: `cmd/gpuctl/main.go`
- Default port: `9090` (configurable via `GPUCTL_PORT`)
- Graceful shutdown: Handles SIGINT/SIGTERM with 10-second timeout
- Timeouts: Read 10s, Write 30s, Idle 60s

---

*Stack analysis: 2026-02-24*
