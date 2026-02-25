# GPU.ai — Project Conventions

## Architecture

GPU.ai is a GPU cloud aggregation platform — a single Go binary (`gpuctl`) that serves the public API, manages provider adapters, availability polling, WireGuard tunnels, billing, and auth.

Python is used only for tooling (migrations, seeding, reports) in `tools/`.

See `planning/docs/ARCHITECTURE.md` for full system design – ONLY REFERENCE THIS DOCUMENT WHEN YOU ARE NOT SURE ABOUT THE PROJECT STRUCTURE OR THE TECH STACK.

## Build & Run

```bash
go build ./cmd/gpuctl             # compile
go run ./cmd/gpuctl               # run on :9090
go test ./...                     # run tests
make build                        # or use Makefile
```

### Python tooling (`tools/`)
```bash
pip install -r tools/requirements.txt
python tools/migrate.py           # run DB migrations
python tools/seed.py              # seed dev data
```

## Code Style

### Go
- Go 1.22+, use stdlib `net/http` with new routing patterns (`GET /path/{id}`)
- `internal/` packages for all business logic
- Structured logging via `log/slog`
- Context propagation on all functions that do I/O
- No frameworks — stdlib + minimal deps (go-redis, pgx)

## Project Structure

```
cmd/gpuctl/main.go          — Entrypoint: wires deps, starts server + goroutines
internal/api/               — HTTP server, routes, middleware, handlers
internal/provider/          — Provider adapter interface + implementations (RunPod, etc.)
internal/availability/      — Redis-cached GPU polling (30s interval)
internal/provision/         — Provisioning orchestration engine
internal/wireguard/         — WireGuard peer management + key generation
internal/billing/           — Stripe integration (usage metering, payments)
internal/auth/              — Clerk JWT verification middleware
internal/db/                — pgx connection pool + SQL queries
internal/health/            — Periodic instance health monitoring
internal/config/            — Environment variable loading
database/migrations/        — SQL migration files
tools/                      — Python tooling (migrate, seed, reports)
infra/cloud-init/           — Cloud-init bootstrap template
```

## Environment

Copy `.env.example` to `.env` and fill in values. Required for local dev:
- `DATABASE_URL` — PostgreSQL connection string
- `REDIS_URL` — Redis connection string
