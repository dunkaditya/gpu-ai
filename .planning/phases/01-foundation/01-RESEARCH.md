# Phase 1: Foundation - Research

**Researched:** 2026-02-24
**Domain:** Go server scaffold, PostgreSQL/Redis connectivity, config loading, database migrations, health endpoint
**Confidence:** HIGH

## Summary

Phase 1 establishes the running Go binary (`gpuctl`) with database and Redis connectivity, environment config loading, a health endpoint, and the database schema applied via Python migrations. The codebase already has a scaffold with stub files across `internal/` packages, a basic `main.go` with a placeholder health endpoint, and a migration SQL file. The work is to flesh these stubs into working implementations.

The standard stack is well-defined by the project conventions: Go 1.22+ stdlib `net/http` with new routing patterns, `pgx/v5` for PostgreSQL, `go-redis/v9` for Redis, and `log/slog` for structured logging. No frameworks. The existing `main.go` already demonstrates graceful shutdown with `signal.NotifyContext` -- this pattern is correct and needs to be preserved while wiring real dependencies.

**Primary recommendation:** Implement config loading first (everything depends on it), then database pool and Redis client with retry logic, then health endpoint, then internal endpoint security, then Docker Compose for local dev, then finalize the migration file and Python runner.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Docker Compose file with Postgres and Redis for one-command dev setup (`docker compose up`)
- Developer runs `docker compose up -d` then `go run ./cmd/gpuctl`
- Retry connecting to Postgres and Redis with exponential backoff before giving up
- Env vars only -- no auto-loading of `.env` files (developer uses `source .env` or Docker sets them)
- Migrations handled by Python `tools/migrate.py` -- run separately, NOT auto-applied by gpuctl
- gpuctl expects the schema to already exist when it starts
- Date-prefixed migration files (keep existing pattern: `20250224_v0.sql`, `20250301_v1.sql`)
- Forward-only migrations -- no rollbacks, write a new migration to fix mistakes
- `/internal/*` endpoints secured with `INTERNAL_API_TOKEN` header (shared secret, already in `.env.example`)
- `/health` endpoint also behind the internal token (not publicly accessible)
- Health returns detailed JSON: `{"status": "ok", "db": "connected", "redis": "connected"}`

### Claude's Discretion
- Which env vars are required vs optional for Phase 1 (DB + Redis clearly required; Clerk, Stripe, RunPod deferred)
- Whether to reject startup if `INTERNAL_API_TOKEN` is still set to the default `change-me` value
- Startup behavior when required config is missing (crash with clear error listing what's missing is recommended)
- Structured logging format and verbosity at startup

### Deferred Ideas (OUT OF SCOPE)
- Kubernetes deployment for HA of the control plane itself -- production deployment concern, not dev milestone
- Auto-migration on startup -- intentionally deferred in favor of explicit `tools/migrate.py` workflow
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| FOUND-01 | Go binary compiles and runs with health endpoint on configurable port | Go 1.22 stdlib `net/http` with `ServeMux`, port from `GPUCTL_PORT` env var with default `9090`. Existing `main.go` scaffold already has this pattern. |
| FOUND-02 | Config loads from environment variables with validation and sensible defaults | Stdlib `os.Getenv` / `os.LookupEnv` pattern. Config struct in `internal/config/config.go`. Required vars fail startup with clear error; optional vars get defaults. |
| FOUND-03 | PostgreSQL connection pool initializes and verifies connectivity on startup | `pgx/v5` pgxpool with `Ping()` verification. Exponential backoff retry before giving up. |
| FOUND-04 | Redis connection initializes and verifies connectivity on startup | `go-redis/v9` with `Ping()` verification. Same retry pattern as Postgres. |
| FOUND-05 | Database migrations apply cleanly to create full schema | Existing `20250224_v0.sql` needs updates (see Schema Gaps below). Python `tools/migrate.py` with `psycopg2` + `schema_migrations` tracking table. |
| API-08 | GET /health returns service health status | Returns JSON with db/redis status. Behind `INTERNAL_API_TOKEN` header auth (user decision). |
| AUTH-04 | Internal endpoints restricted to localhost only | User decision overrides: use `INTERNAL_API_TOKEN` shared secret header instead of IP-based restriction. Health endpoint also behind this token. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `net/http` | Go 1.22+ | HTTP server with method+wildcard routing | Project convention: no frameworks. Go 1.22 added `GET /path/{id}` patterns to stdlib `ServeMux`. |
| `github.com/jackc/pgx/v5` | v5.7+ (latest v5.8.0) | PostgreSQL driver + connection pool | Project convention specifies pgx. Most performant Go PG driver; pure Go, supports COPY, LISTEN/NOTIFY, binary protocol. |
| `github.com/jackc/pgx/v5/pgxpool` | (same module) | Connection pooling | Built-in to pgx. Concurrency-safe pool with health checks, configurable min/max connections. |
| `github.com/redis/go-redis/v9` | v9.18+ (latest v9.18.0) | Redis client | Project convention specifies go-redis. Official Redis client for Go with automatic connection pooling, RESP3 support. |
| `log/slog` | Go 1.21+ (stdlib) | Structured logging | Project convention specifies slog. JSON handler for production, text handler for dev. |
| `psycopg2-binary` | 2.9+ | Python PostgreSQL driver for migration runner | Already in `tools/requirements.txt`. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `click` | 8.1+ | Python CLI framework for migration tool | Already in `tools/requirements.txt`. For `--status`, `--target` flags. |
| `tabulate` | 0.9+ | Python table formatting for migration status output | Already in `tools/requirements.txt`. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib `net/http` | chi router | Chi adds middleware chaining, but project convention mandates stdlib. Go 1.22 closed the feature gap. |
| pgx | database/sql + lib/pq | database/sql is simpler but loses pgx-specific features (COPY, binary protocol, better performance). Project convention specifies pgx. |
| Manual env loading | godotenv, envconfig, viper | User decided against auto-loading .env files. Stdlib `os.Getenv` keeps it simple with zero deps. |
| Custom retry | cenkalti/backoff | Simple retry loop with sleep is sufficient for startup-only use; no need for a library for 10 lines of code. |

**Installation:**
```bash
# Go dependencies
go get github.com/jackc/pgx/v5
go get github.com/redis/go-redis/v9

# Python dependencies (already specified)
pip install -r tools/requirements.txt
```

## Architecture Patterns

### Recommended Project Structure (Phase 1 scope)
```
cmd/gpuctl/main.go              # Wire deps, start server + graceful shutdown
internal/config/config.go        # Env var loading into typed Config struct
internal/db/pool.go              # pgxpool wrapper with Ping + Close
internal/api/server.go           # HTTP server with routes + middleware
internal/api/middleware.go       # Internal token auth middleware (NEW file)
database/migrations/20250224_v0.sql  # Full schema (updated)
tools/migrate.py                 # Migration runner (implemented)
docker-compose.yml               # Postgres + Redis for local dev (NEW file)
```

### Pattern 1: Config Loading with Validation
**What:** Load all env vars into a typed struct at startup. Validate required fields. Fail fast with a clear error listing ALL missing vars (not just the first).
**When to use:** Always at startup, before any other initialization.
**Example:**
```go
// Source: Go stdlib os.Getenv / os.LookupEnv
type Config struct {
    Port              string
    DatabaseURL       string
    RedisURL          string
    InternalAPIToken  string
}

func Load() (*Config, error) {
    var missing []string

    cfg := &Config{
        Port: getEnvDefault("GPUCTL_PORT", "9090"),
    }

    cfg.DatabaseURL = os.Getenv("DATABASE_URL")
    if cfg.DatabaseURL == "" {
        missing = append(missing, "DATABASE_URL")
    }

    // ... check all required vars

    if len(missing) > 0 {
        return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
    }
    return cfg, nil
}
```

### Pattern 2: Connection Retry with Exponential Backoff
**What:** Retry database/Redis connections with increasing delays (1s, 2s, 4s, 8s...) up to a maximum number of attempts. Useful when Docker Compose services are still starting.
**When to use:** At startup when connecting to Postgres and Redis.
**Example:**
```go
// Source: common Go pattern, no external library needed
func connectWithRetry(ctx context.Context, connect func(ctx context.Context) error, name string, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = connect(ctx)
        if err == nil {
            return nil
        }
        delay := time.Duration(1<<uint(i)) * time.Second // 1s, 2s, 4s, 8s...
        if delay > 30*time.Second {
            delay = 30 * time.Second
        }
        slog.Warn("connection failed, retrying", "service", name, "attempt", i+1, "delay", delay, "error", err)
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
    }
    return fmt.Errorf("failed to connect to %s after %d attempts: %w", name, maxRetries, err)
}
```

### Pattern 3: Internal Token Auth Middleware
**What:** Middleware that checks for `INTERNAL_API_TOKEN` in the Authorization header (or a custom header). Rejects requests without valid token with 403.
**When to use:** On all `/internal/*` and `/health` routes (per user decision).
**Example:**
```go
// Source: standard Go middleware pattern
func InternalAuthMiddleware(token string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        provided := r.Header.Get("Authorization")
        if provided == "" || provided != "Bearer "+token {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Pattern 4: Health Check Handler
**What:** Handler that actively pings Postgres and Redis, returning per-dependency status.
**When to use:** GET /health endpoint.
**Example:**
```go
// Source: standard health check pattern
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    dbStatus := "connected"
    if err := s.db.Ping(ctx); err != nil {
        dbStatus = "disconnected"
    }

    redisStatus := "connected"
    if err := s.redis.Ping(ctx).Err(); err != nil {
        redisStatus = "disconnected"
    }

    status := "ok"
    statusCode := http.StatusOK
    if dbStatus != "connected" || redisStatus != "connected" {
        status = "degraded"
        statusCode = http.StatusServiceUnavailable
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "status": status,
        "db":     dbStatus,
        "redis":  redisStatus,
    })
}
```

### Pattern 5: Dependency Injection via Struct
**What:** Main function creates all dependencies and passes them to the Server struct. No globals, no init().
**When to use:** In `cmd/gpuctl/main.go` for wiring.
**Example:**
```go
// Source: standard Go dependency injection pattern (no framework)
type Server struct {
    cfg    *config.Config
    db     *db.Pool
    redis  *redis.Client
    logger *slog.Logger
    mux    *http.ServeMux
}
```

### Anti-Patterns to Avoid
- **Global database connections:** All connections must be passed through struct fields, not package-level vars.
- **log.Fatal in goroutines:** Use `slog.Error` and return errors; `log.Fatal` calls `os.Exit(1)` which skips deferred cleanup.
- **Swallowing connection errors:** If Postgres or Redis is unreachable after retries, the binary MUST exit with a non-zero code and clear error message.
- **Auto-loading .env files:** User explicitly decided against this. The developer uses `source .env` or Docker sets environment vars.
- **Auto-running migrations:** User explicitly decided against this. gpuctl expects the schema to exist.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| PostgreSQL connection pooling | Custom pool manager | pgxpool from pgx/v5 | Handles health checks, connection validation, min/max sizing, idle timeout |
| Redis connection pooling | Custom pool | go-redis built-in pool | Automatic pool management, configurable pool size |
| SQL migration tracking | Custom version tracking | schema_migrations table pattern | Simple, proven pattern: table with version + applied_at columns |
| HTTP routing with methods | Custom router/switch | Go 1.22 ServeMux patterns | `"GET /health"` syntax built into stdlib, handles method matching |
| Graceful shutdown | Custom signal handling | signal.NotifyContext + server.Shutdown | Already in existing main.go, stdlib handles edge cases correctly |

**Key insight:** Phase 1 is infrastructure wiring, not business logic. Every component has a well-established standard pattern. The risk is not complexity but incorrect wiring order or missing error handling.

## Common Pitfalls

### Pitfall 1: TIMESTAMPTZ vs TIMESTAMP in PostgreSQL
**What goes wrong:** The existing `20250224_v0.sql` uses `TIMESTAMP` (without timezone) for all time columns. The `ARCHITECTURE.md` schema uses `TIMESTAMPTZ`. Using `TIMESTAMP` loses timezone information, causing bugs when servers run in different timezones.
**Why it happens:** `TIMESTAMP` is shorter to type. Developers forget that PostgreSQL stores `TIMESTAMP` as-is without timezone context.
**How to avoid:** Use `TIMESTAMPTZ` for all time columns. The migration file needs updating.
**Warning signs:** Times showing wrong offset in logs, billing calculations off by hours.

### Pitfall 2: Missing clerk_user_id column in users table
**What goes wrong:** The existing `20250224_v0.sql` is missing the `clerk_user_id` column on the `users` table. The `ARCHITECTURE.md` schema includes it. Without this column, Phase 4 (auth) will require a schema migration that could have been avoided.
**Why it happens:** The migration was written before auth design was finalized.
**How to avoid:** Add `clerk_user_id VARCHAR(255) UNIQUE` to the users table now. Forward-looking schema design prevents future migration churn.
**Warning signs:** N/A -- will fail at Phase 4 if not addressed.

### Pitfall 3: Docker Compose Port Conflicts
**What goes wrong:** Default PostgreSQL port 5432 or Redis port 6379 conflicts with locally installed instances.
**Why it happens:** Developer already has Postgres/Redis running natively.
**How to avoid:** Docker Compose should use the standard ports (matching `.env.example`) but document how to change them. The `.env.example` already uses standard ports.
**Warning signs:** "address already in use" errors on `docker compose up`.

### Pitfall 4: Context Cancellation During Retry
**What goes wrong:** Retry loop doesn't check context cancellation, so the app hangs during shutdown if it's still retrying connections.
**Why it happens:** Simple `time.Sleep` loops don't respect context.
**How to avoid:** Use `select` with `<-ctx.Done()` and `<-time.After(delay)` in the retry loop (see Pattern 2 above).
**Warning signs:** App takes a long time to shut down during development when DB isn't running.

### Pitfall 5: pgxpool Default Max Connections Too Low
**What goes wrong:** pgxpool defaults to max 4 connections. Under load this causes connection wait timeouts.
**Why it happens:** pgxpool's conservative default is designed for single-user CLI tools, not servers.
**How to avoid:** Set `pool_max_conns=20` in the connection string or via `pgxpool.Config`. The existing stub suggests min 5, max 20.
**Warning signs:** "acquiring connection from pool" timeout errors under concurrent requests.

### Pitfall 6: Redis URL Parsing
**What goes wrong:** `go-redis` `ParseURL()` expects a specific format (`redis://host:port/db`). Some formats cause silent failures.
**Why it happens:** Redis URL format varies between libraries and platforms.
**How to avoid:** Use `redis.ParseURL()` from go-redis, which handles the standard `redis://` scheme. The `.env.example` already uses the correct format.
**Warning signs:** "invalid URL" errors or connecting to wrong database number.

### Pitfall 7: Missing gen_random_uuid() Extension
**What goes wrong:** The existing migration does NOT include `CREATE EXTENSION IF NOT EXISTS "pgcrypto"` but uses `gen_random_uuid()` in column defaults.
**Why it happens:** In PostgreSQL 13+, `gen_random_uuid()` is built-in without pgcrypto. But older versions need the extension.
**How to avoid:** Include `CREATE EXTENSION IF NOT EXISTS "pgcrypto"` at the top of the migration for compatibility. The ARCHITECTURE.md schema includes this.
**Warning signs:** "function gen_random_uuid() does not exist" error when running migrations on PostgreSQL < 13.

## Code Examples

Verified patterns from official sources and stdlib documentation:

### pgxpool Connection Setup
```go
// Source: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool
import "github.com/jackc/pgx/v5/pgxpool"

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("parsing database URL: %w", err)
    }
    config.MaxConns = 20
    config.MinConns = 5

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("creating pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("pinging database: %w", err)
    }

    return pool, nil
}
```

### go-redis Connection Setup
```go
// Source: https://redis.io/docs/latest/develop/clients/go/connect/
import "github.com/redis/go-redis/v9"

func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, fmt.Errorf("parsing redis URL: %w", err)
    }

    client := redis.NewClient(opt)

    if err := client.Ping(ctx).Err(); err != nil {
        client.Close()
        return nil, fmt.Errorf("pinging redis: %w", err)
    }

    return client, nil
}
```

### Go 1.22 ServeMux Routing
```go
// Source: https://go.dev/blog/routing-enhancements
mux := http.NewServeMux()

// Method matching
mux.HandleFunc("GET /health", handleHealth)
mux.HandleFunc("POST /internal/instances/{id}/ready", handleReady)

// Path parameter extraction
func handleReady(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
}
```

### slog Structured Logging Setup
```go
// Source: https://go.dev/blog/slog
import "log/slog"

// JSON handler for production-style output
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Usage
slog.Info("server starting", "port", cfg.Port)
slog.Error("connection failed", "service", "postgres", "error", err)
```

### Docker Compose for Local Dev
```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: gpuai
      POSTGRES_PASSWORD: gpuai
      POSTGRES_DB: gpuai
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

### Python Migration Runner Pattern
```python
# Source: standard migration runner pattern with psycopg2
import psycopg2
import glob
import os

def run_migrations(conn, migrations_dir):
    with conn.cursor() as cur:
        cur.execute("""
            CREATE TABLE IF NOT EXISTS schema_migrations (
                version VARCHAR(255) PRIMARY KEY,
                applied_at TIMESTAMPTZ DEFAULT NOW()
            )
        """)
        cur.execute("SELECT version FROM schema_migrations ORDER BY version")
        applied = {row[0] for row in cur.fetchall()}

    files = sorted(glob.glob(os.path.join(migrations_dir, "*.sql")))
    for f in files:
        version = os.path.basename(f)
        if version not in applied:
            with open(f) as sql_file:
                with conn.cursor() as cur:
                    cur.execute(sql_file.read())
                    cur.execute(
                        "INSERT INTO schema_migrations (version) VALUES (%s)",
                        (version,)
                    )
            conn.commit()
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Chi/Gorilla for HTTP routing | Go 1.22 stdlib ServeMux with method+wildcard patterns | Go 1.22, Feb 2024 | No external router needed; `"GET /path/{id}"` works in stdlib |
| `log` package | `log/slog` structured logging | Go 1.21, Aug 2023 | JSON-structured logs out of the box, no external logging library needed |
| `lib/pq` PostgreSQL driver | `pgx/v5` native driver | pgx v5, Oct 2022 | Better performance, connection pooling, PostgreSQL-specific features |
| `go-redis/redis/v9` import path | `redis/go-redis/v9` import path | 2023 | Repository moved to Redis GitHub org; new import path |
| Docker Compose v1 (`docker-compose`) | Docker Compose v2 (`docker compose`) | 2022 | Integrated into Docker CLI, no separate binary |

**Deprecated/outdated:**
- `lib/pq`: Still maintained but pgx is the recommended PostgreSQL driver for new Go projects
- `go-redis/redis/v9` import: Old import path still works but `redis/go-redis/v9` is canonical
- `docker-compose` (hyphenated): Use `docker compose` (space) for Compose v2

## Schema Gaps in Existing Migration

The existing `database/migrations/20250224_v0.sql` has several discrepancies vs the `ARCHITECTURE.md` reference schema. These should be fixed in this phase:

| Issue | Current (v0.sql) | Should Be (ARCHITECTURE.md) | Impact |
|-------|-------------------|------------------------------|--------|
| Missing pgcrypto extension | Not present | `CREATE EXTENSION IF NOT EXISTS "pgcrypto"` | Fails on PostgreSQL < 13 |
| Timestamp type | `TIMESTAMP` | `TIMESTAMPTZ` | Timezone-unaware times cause bugs |
| Missing clerk_user_id | Not on users table | `clerk_user_id VARCHAR(255) UNIQUE` | Blocks Phase 4 auth integration |
| SSH keys ON DELETE | No cascade rule | `ON DELETE CASCADE` | Orphaned SSH keys when user deleted |
| WG column names | `wireguard_public_key`, `wireguard_private_key` | `wg_public_key`, `wg_private_key_enc` | Column name mismatch with code; `_enc` suffix clarifies encryption |
| WG address column | `wireguard_address` | `wg_address` | Column name mismatch with code |
| Usage records stripe field | `stripe_invoice_id` | `stripe_usage_record_id` | Semantic mismatch (usage records vs invoices) |
| Index names | `idx_instances_org_id` etc. | `idx_instances_org` etc. | Minor naming difference |

**Recommendation:** Since migrations are forward-only and this is the initial schema (no production data exists), the simplest fix is to update `20250224_v0.sql` directly to match the ARCHITECTURE.md reference schema. This is a greenfield project with no deployed databases.

## Open Questions

1. **INTERNAL_API_TOKEN default value rejection**
   - What we know: `.env.example` has `INTERNAL_API_TOKEN=change-me`. User marked this as Claude's discretion.
   - What's unclear: Whether to reject startup if the value is literally `change-me`.
   - Recommendation: YES -- reject startup with `change-me` value. A running server with a known default token is a security risk. Log a clear error: "INTERNAL_API_TOKEN must be changed from default value".

2. **AUTH-04 vs CONTEXT.md conflict**
   - What we know: AUTH-04 says "Internal endpoints restricted to localhost only." CONTEXT.md says "secured with INTERNAL_API_TOKEN header (shared secret)."
   - What's unclear: Which takes precedence.
   - Recommendation: Follow CONTEXT.md (user's explicit decision). The token-based approach is more flexible and works when gpuctl runs behind a reverse proxy or in Docker where "localhost" semantics differ. The requirement should be interpreted as "internal endpoints are not publicly accessible" rather than literally "IP-restricted to 127.0.0.1."

3. **Phase 1 required vs optional env vars**
   - What we know: User says "DB + Redis clearly required; Clerk, Stripe, RunPod deferred."
   - Recommendation: Phase 1 required: `DATABASE_URL`, `REDIS_URL`, `INTERNAL_API_TOKEN`. Phase 1 optional with defaults: `GPUCTL_PORT` (default `9090`). All other vars (Clerk, Stripe, RunPod, E2E, WireGuard) should be ignored in Phase 1 config loading -- they'll be added in later phases.

## Sources

### Primary (HIGH confidence)
- [Go 1.22 Routing Enhancements](https://go.dev/blog/routing-enhancements) - Method matching, wildcards in stdlib ServeMux
- [Go slog blog post](https://go.dev/blog/slog) - Structured logging in stdlib
- [pgxpool package docs](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) - Connection pool API, configuration
- [go-redis official docs](https://redis.io/docs/latest/develop/clients/go/) - Redis client setup, ParseURL
- [go-redis connect guide](https://redis.io/docs/latest/develop/clients/go/connect/) - Connection patterns

### Secondary (MEDIUM confidence)
- [pgx GitHub repository](https://github.com/jackc/pgx) - Version v5.7+/v5.8.0 confirmed via GitHub tags
- [go-redis GitHub releases](https://github.com/redis/go-redis/releases) - v9.18.0 confirmed as latest
- [Better Stack slog guide](https://betterstack.com/community/guides/logging/logging-in-go/) - Production logging patterns
- [Better Stack pgx guide](https://betterstack.com/community/guides/scaling-go/postgresql-pgx-golang/) - pgx connection patterns
- [Go graceful shutdown patterns](https://victoriametrics.com/blog/go-graceful-shutdown/) - signal.NotifyContext patterns

### Tertiary (LOW confidence)
- Docker Compose configuration: Based on standard patterns, not verified against specific Docker Compose version

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries specified by project conventions (CLAUDE.md), versions confirmed via official package registries
- Architecture: HIGH - Patterns are well-established Go idioms; existing scaffold aligns with conventions
- Pitfalls: HIGH - Schema gaps verified by direct comparison of two files in the repo; connection pitfalls from official pgx/redis docs

**Research date:** 2026-02-24
**Valid until:** 2026-03-24 (30 days -- stable domain, no fast-moving dependencies)
