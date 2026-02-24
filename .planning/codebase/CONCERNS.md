# Codebase Concerns

**Analysis Date:** 2026-02-24

## Implementation Status

**CRITICAL: Codebase is a skeleton with TODOs, not production code.**

The entire codebase consists of:
- Package declarations and interface definitions
- Comprehensive TODO comments outlining what needs implementation
- One functional component: `/cmd/gpuctl/main.go` (health check only)
- Database schema defined but no ORM/query implementations
- Python tooling stubs (`tools/migrate.py`, `tools/seed.py`)

**Impact:** Zero functional features are implemented. No tests exist beyond stubs. This is a Phase 1 architecture blueprint, not a working system.

---

## Tech Debt & Implementation Gaps

### Core Services Completely Unimplemented

**Database Layer:**
- Files: `internal/db/pool.go`, `internal/db/instances.go`, `internal/db/organizations.go`, `internal/db/ssh_keys.go`
- Issue: All database query functions are stubbed with TODO comments. No pgx pool initialization, no SQL query methods exist.
- Impact: Cannot persist or retrieve any data from PostgreSQL
- Fix approach: Implement Pool struct with pgxpool, then implement query methods for each entity (Instance, Organization, User, SSHKey)
- Priority: Critical — blocking all data operations

**HTTP Server & API Handlers:**
- Files: `internal/api/server.go`, `internal/api/handlers.go`
- Issue: No HTTP routes registered, no handlers implemented. Only health check endpoint exists in main.go
- Impact: All API endpoints (instances, GPU availability, SSH keys, billing) are non-functional
- Fix approach: Implement Server struct, register routes, implement handlers for CRUD operations on instances, GPU availability queries, billing endpoints
- Priority: Critical — blocks user-facing API

**Authentication:**
- Files: `internal/auth/clerk.go`
- Issue: Clerk JWT verification middleware is completely unimplemented
- Impact: All endpoints will be unauthenticated; cannot verify customer identity or extract organization/user claims
- Fix approach: Implement Verifier struct with JWT parsing from Authorization header, validate signature against Clerk JWKS, extract claims into request context
- Priority: Critical — security-blocking

**Configuration Loading:**
- Files: `internal/config/config.go`
- Issue: Environment variable loading is stubbed. No validation of required fields.
- Impact: Application cannot load credentials (Clerk, Stripe, RunPod API keys, database URL, Redis URL)
- Fix approach: Implement Load() function to read env vars, validate required fields, return Config struct or error
- Priority: Critical — blocks startup

**Provider Adapter (RunPod):**
- Files: `internal/provider/runpod/adapter.go`
- Issue: Adapter interface implementation is stubbed. Cannot communicate with RunPod API
- Impact: Cannot provision instances, list available GPUs, check status, or terminate instances
- Fix approach: Implement HTTP client for RunPod GraphQL API, implement ListAvailable(), Provision(), GetStatus(), Terminate() methods, handle GraphQL responses
- Priority: Critical — core business logic

**WireGuard Management:**
- Files: `internal/wireguard/manager.go`, `internal/wireguard/keygen.go`
- Issue: No WireGuard command execution, no key generation, no peer management
- Impact: Cannot establish encrypted tunnels between customer and instances; instances cannot be configured with private keys
- Fix approach: Implement command execution wrappers for `wg set`, `wg show`, keygen; handle config file persistence
- Priority: Critical — privacy/security layer

**Provisioning Engine:**
- Files: `internal/provision/engine.go`
- Issue: Orchestration logic is stubbed. No provider selection, cloud-init injection, or WireGuard peer registration
- Impact: Cannot launch instances; entire provisioning flow is non-functional
- Fix approach: Implement Provision() to select provider, build cloud-init script from template, inject secrets, call provider adapter, register WireGuard peer
- Priority: Critical — core feature

**Availability Poller:**
- Files: `internal/availability/poller.go`, `internal/availability/cache.go`
- Issue: No polling loop, no Redis cache helpers, no concurrent provider queries
- Impact: Users cannot see available GPUs and pricing; availability cache is empty
- Fix approach: Implement Poller Run() with ticker, implement concurrent provider polling, implement Redis cache SetOffering() and GetOfferings() methods
- Priority: High — required for API availability endpoint

**Health Monitoring:**
- Files: `internal/health/monitor.go`
- Issue: No health check logic, no WireGuard peer status checking, no instance status updates
- Impact: Cannot detect unhealthy instances or stale connections
- Fix approach: Implement Monitor Run() with ticker, query running instances, check WireGuard last handshake times, mark instances unhealthy if timeout exceeded
- Priority: Medium — affects instance reliability detection

**Billing Service:**
- Files: `internal/billing/stripe.go`
- Issue: No Stripe API integration, no usage metering, no webhook handling
- Impact: Cannot verify customer payment status, cannot meter usage, cannot send invoices
- Fix approach: Implement Service with Stripe client, implement CheckBillingStatus(), StartUsage(), StopUsage(), GetUsage(), HandleWebhook()
- Priority: High — blocks monetization

---

## Architectural Concerns

### WireGuard Private Key Storage

**Files:** `internal/db/instances.go` (schema comment)

**Issue:** WireGuard private keys are encrypted at rest according to schema comment, but:
1. No encryption/decryption implementation exists
2. Encryption algorithm/key derivation method not specified
3. Schema shows plaintext column: `wireguard_private_key VARCHAR(255)` — contradicts "encrypted at rest" promise
4. Database connection string may be exposed in logs

**Impact:** Private keys could be exposed if:
- Database is compromised
- Backup files are leaked
- Query logs are captured

**Recommendations:**
1. Implement envelope encryption: store encrypted key + key ID in database
2. Use AWS KMS, HashiCorp Vault, or similar HSM for key management
3. Never log database queries containing keys
4. Implement key rotation mechanism
5. Add database encryption at rest (PostgreSQL pgcrypto or transparent disk encryption)

**Priority:** Critical (before production launch)

---

### Stripe API Token Exposure

**Files:** `internal/billing/stripe.go`, `cmd/gpuctl/main.go`

**Issue:** Stripe secret key is loaded from environment variable but usage pattern not defined:
- Unknown if it's passed to functions as plain text
- Unknown if stored in logs or error messages
- Unknown if sent in HTTP requests without TLS

**Impact:** Stripe key compromise → customer funds fraudulently charged

**Recommendations:**
1. Create secrets management abstraction layer
2. Load secrets once at startup into secure struct
3. Pass secretRef IDs between functions, not values
4. Implement audit logging for any secret access
5. Rotate secrets quarterly minimum

**Priority:** Critical

---

### Cloud-Init Script Injection Risks

**Files:** `internal/provision/engine.go`, `infra/cloud-init/` (referenced but not examined)

**Issue:** TODO comment in engine.go mentions "Build cloud-init script from template" but:
1. No template file location specified
2. No escaping/validation of injected variables (instance ID, SSH keys, Docker image)
3. Variables injected into bash script — shell metacharacters could break script or cause injection

**Impact:** Malformed cloud-init could fail to boot instances, or injected SSH key could corrupt script execution

**Example risk:** SSH key containing `$(malicious_command)` would execute on instance

**Recommendations:**
1. Use structured templating (not string interpolation): Go templates with auto-escaping
2. Validate all inputs before injection:
   - Instance IDs: alphanumeric + hyphen only
   - SSH keys: valid OpenSSH format
   - Docker image: valid registry URL format
3. Test cloud-init script generation with fuzzing
4. Store approved cloud-init templates in version control with signed commits

**Priority:** High

---

### Concurrent Provider Polling Without Rate Limiting

**Files:** `internal/availability/poller.go` (TODO comment)

**Issue:** TODO mentions "For each provider (concurrently)" but:
1. No rate limiting implementation
2. No exponential backoff for failures
3. No timeout per provider
4. Could hammer upstream APIs if they're slow or flaky

**Example failure mode:** If RunPod API is slow, all goroutines could pile up and exhaust memory

**Impact:** Denial of service against GPU.ai's own infrastructure or upstream providers

**Recommendations:**
1. Implement worker pool pattern (e.g., 5 concurrent providers max)
2. Use context timeout per provider (e.g., 15 seconds)
3. Implement exponential backoff: 1s → 2s → 4s → 8s cap
4. Track failure counts and circuit-break providers that fail N times
5. Add prometheus metrics for poll latency and success rate

**Priority:** Medium → High (becomes critical at scale)

---

### Missing Error Handling Patterns

**Files:** All API/handler files have TODO comments without error handling spec

**Issue:** No documented approach for:
1. Provider API errors (rate limits, 5xx, timeouts)
2. Database errors (connection lost, constraint violations)
3. User input validation errors
4. WireGuard command failures

**Impact:** Unhandled errors crash goroutines, leaving orphaned database records (e.g., created instance but failed to set WireGuard peer)

**Recommendations:**
1. Define error categories: retryable vs fatal
2. Implement circuit breaker for provider adapters
3. Use transaction semantics: provision succeeds or rolls back entirely
4. Log errors with context: user ID, request ID, operation, attempt number
5. Return user-friendly error messages (never expose internal stack traces)

**Priority:** High

---

### No Request Validation

**Files:** `internal/api/handlers.go` (TODO comments)

**Issue:** Handler stubs don't define validation for:
1. `CreateInstanceRequest` — GPU type, count, region validity
2. `CreateSSHKeyRequest` — SSH key format validation
3. Query parameters — type, tier, region filtering

**Impact:** Invalid requests reach business logic, causing cryptic errors or unexpected behavior

**Example:** GPU count of 0 or negative should be rejected at API boundary

**Fix approach:**
1. Define request DTOs with validation tags (e.g., `gpu_count min:1 max:8`)
2. Use validator library (e.g., `go-playground/validator`)
3. Return 400 with field-level error messages
4. Log validation failures for security monitoring

**Priority:** Medium

---

### Missing Database Migrations Tooling

**Files:** `tools/migrate.py`, `database/migrations/20250224_v0.sql`

**Issue:**
1. Migration runner is a Python TODO stub
2. Only one migration file exists (v0)
3. No version tracking table schema defined in migration
4. No rollback mechanism documented
5. Tools not called from Go code — no way to run migrations from `make` or CI/CD

**Impact:** Cannot safely evolve schema; dev/prod schema drift possible

**Recommendations:**
1. Implement migration runner in Go or call Python from Go
2. Add schema_migrations table to v0 migration
3. Implement rollback for each migration (DOWN script)
4. Integrate into startup: `gpuctl migrate` command before starting server
5. Test migrations in CI with test database

**Priority:** High

---

### No Test Coverage

**Files:** `internal/provider/runpod/adapter_test.go` (24 lines, all TODO)

**Issue:**
- Only one test file exists, and it's empty TODO stubs
- No unit tests for any package
- No integration tests
- No mocks for external services (RunPod, Stripe, Clerk)

**Impact:** Cannot safely refactor code, bugs go undetected, provider adapters untested

**Test coverage gaps:**
- Provider adapters (RunPod, E2E API integration)
- Database queries (CRUD operations, edge cases)
- Billing calculations and Stripe webhook handling
- WireGuard peer management (adding/removing peers)
- Authentication middleware (valid tokens, expired tokens, missing headers)
- Provisioning orchestration (end-to-end flow)

**Recommendations:**
1. Add test for RunPod adapter (mock GraphQL responses)
2. Add tests for database queries (using PostgreSQL test containers)
3. Add tests for billing service (mock Stripe client)
4. Add tests for provisioning engine (mock provider adapters + WireGuard)
5. Set CI/CD gate: 70%+ coverage minimum
6. Use table-driven tests for multiple scenarios

**Priority:** Critical (test-first development going forward)

---

### Environment Configuration Not Validated

**Files:** `internal/config/config.go`, `.env.example`

**Issue:**
1. Config.Load() is not implemented
2. No validation of required fields (DATABASE_URL, REDIS_URL, secrets)
3. No example default values for optional fields
4. No documentation of what each var means or expected format

**Impact:** Application starts with missing credentials, fails at first attempt to use that service

**Example:** Missing `CLERK_SECRET_KEY` causes auth middleware to panic

**Recommendations:**
1. Implement config validation at startup
2. Fatal error if any required var is missing
3. Print which vars are missing and where to set them
4. Add config logging on startup (mask secrets): "Loaded config: DATABASE_URL=postgres://host/db, REDIS_URL=redis://..., RUNPOD_API_KEY=***"
5. Document all env vars in CLAUDE.md or docs/CONFIG.md

**Priority:** Medium

---

### Redis Connection Pooling

**Files:** `internal/availability/cache.go`, `cmd/gpuctl/main.go`

**Issue:**
- No Redis client initialization in main.go
- No connection pooling configuration
- No handling of Redis connection loss

**Impact:** If Redis is unavailable at startup, application starts but crashes on first cache access

**Recommendations:**
1. Initialize Redis client in main.go (before starting poller)
2. Implement retry logic with exponential backoff for Redis connection
3. Optional graceful degradation: serve stale cache data if Redis is temporarily down
4. Health check endpoint should verify Redis connectivity
5. Metrics: track Redis connection pool usage, cache hit/miss rates

**Priority:** Medium

---

### Concurrent Map Access in Provider Adapters

**Files:** `cmd/gpuctl/main.go` (comment line 26)

**Issue:** `providers := map[string]provider.Provider{"runpod": ...}` suggests map of providers but:
1. No thread-safe access mechanism documented
2. No locking if adapters are added/removed after startup

**Impact:** Race conditions if providers map is mutated while pollers access it

**Recommendations:**
1. Use `sync.RWMutex` to protect provider map access
2. OR make provider map immutable after startup
3. Add race detector to test: `go test -race ./...`

**Priority:** Medium

---

### Upstream Instance IP Leakage Risk

**Files:** Database schema, `internal/db/instances.go`

**Issue:** Schema includes `upstream_ip INET` which violates privacy promise. If API handler accidentally returns this field:

```go
// WRONG: exposes upstream IP to customer
json.Marshal(instance) // includes upstream_ip
```

**Impact:** Customer discovers which upstream provider they're using (violates privacy promise)

**Recommendations:**
1. Create separate response DTOs that exclude upstream fields
2. Use tags: `json:"-"` for sensitive fields in instance struct
3. Code review checklist: "no upstream_ fields in API responses"
4. Test: verify API responses don't contain upstream fields

**Priority:** High

---

### No Graceful Shutdown Handling for Goroutines

**Files:** `cmd/gpuctl/main.go`

**Issue:**
- Poller, monitor, and provisioning services spawned as goroutines
- No context cancellation mechanism defined in packages
- If goroutines are still running, they're force-killed on shutdown

**Impact:** Incomplete operations, inconsistent state, orphaned resources

**Example:** Provisioning goroutine creating instance while shutdown signal is sent

**Recommendations:**
1. Pass context from main.go to all long-running services
2. Services should respect context cancellation and clean up
3. Use sync.WaitGroup to wait for goroutine completion before exiting
4. Implement graceful shutdown timeout (e.g., 30 seconds then force kill)
5. Test: send SIGTERM while provisioning and verify cleanup

**Priority:** High

---

### Missing Observability

**Files:** `cmd/gpuctl/main.go`, all packages

**Issue:**
- No logging strategy defined (mentioned `log/slog` but not implemented)
- No metrics/tracing
- No request ID correlation
- No structured logging in any handler/service

**Impact:** Impossible to debug production issues, no visibility into system behavior

**Recommendations:**
1. Implement slog initialization in main.go with JSON output
2. Add request ID middleware to inject into all logs
3. Log at entry/exit of all public functions with arguments and results
4. Add metrics: request latency, error rates, queue depths
5. Use OpenTelemetry for distributed tracing across services

**Priority:** High (before beta)

---

### Spot Instance Interruption Not Handled

**Files:** Architecture doc (Section: Spot Instance Behavior), no implementation

**Issue:**
- Architecture mentions "5-second SIGTERM warning" for spot interruptions
- No handler in health monitor or instance code to gracefully shut down on SIGTERM
- No cloud-init script logic to signal GPU.ai proxy on interruption

**Impact:** When spot instance is interrupted:
1. GPU.ai doesn't know immediately (status stays "running")
2. Customer attempts to SSH and gets connection refused
3. Orphaned WireGuard peer stays registered on proxy

**Recommendations:**
1. Add SIGTERM handler to cloud-init: notify GPU.ai API `/internal/instances/{id}/interrupted`
2. API marks instance as "spot_interrupted"
3. Health monitor removes WireGuard peer on interruption
4. Include graceful shutdown in instance bootstrap scripts
5. Document to customers: "Spot instances have 5-second termination notice"

**Priority:** High (impacts spot tier reliability)

---

### No Rate Limiting on Public API

**Files:** `internal/api/handlers.go` (not implemented)

**Issue:**
- No rate limiting specified for `GET /api/v1/instances`, `POST /api/v1/instances`, etc.
- Malicious user could flood API with provision requests

**Impact:** DoS attack; billing system flooded with invalid provisions

**Recommendations:**
1. Implement per-user rate limiting (extract user ID from JWT claims)
2. Different limits per endpoint: provision (1/sec), list (100/sec), status (10/sec)
3. Use token bucket algorithm or sliding window
4. Return 429 Too Many Requests with Retry-After header
5. Metrics: track rate limit violations by user

**Priority:** Medium → High (before public beta)

---

### Billing Start/End Race Condition

**Files:** `internal/api/handlers.go` (HandleCreateInstance TODO), `internal/provision/engine.go`, `internal/health/monitor.go`

**Issue:**
- Instance created (status='creating')
- Billing not started yet (billing_start is NULL)
- Cloud-init calls `/internal/instances/{id}/ready`
- Handler sets status='running' and billing_start=NOW()
- But if handler crashes, billing_start stays NULL forever

**Impact:** Customer's instance runs unbilled or double-charged if retry happens

**Recommendations:**
1. Use database transactions: set status and billing_start atomically
2. Or use temporal semantics: billing starts based on cloud-init callback timestamp
3. Billing service queries for instances where status='running' and billing_start IS NULL → logs alert
4. Test: simulate handler crash and verify cleanup process

**Priority:** High

---

## Security Considerations

### Clerk JWT Validation Not Implemented

**Files:** `internal/auth/clerk.go`

**Risk:** If implementation cuts corners:
- Doesn't validate signature against Clerk JWKS
- Doesn't check token expiration
- Doesn't validate organization claims

**Recommendations:**
1. Use Clerk's official Go SDK if available
2. Validate signature, expiration, and required claims
3. Cache JWKS for performance but refresh periodically
4. Log auth failures with rate limiting to prevent log spam

**Priority:** Critical

---

### WireGuard Key Generation

**Files:** `internal/wireguard/keygen.go`

**Risk:** If implementation is naive:
- Uses weak random source (math/rand instead of crypto/rand)
- Doesn't use proper key derivation

**Recommendations:**
1. Use `crypto/rand` for all key generation
2. Test with cryptographic audit if possible
3. Document key format (Ed25519, not RSA)

**Priority:** High

---

### SQL Injection Risk in Database Layer

**Files:** All `internal/db/*.go` files (not yet implemented)

**Risk:** If queries are built with string concatenation instead of prepared statements

**Recommendations:**
1. Use pgx prepared statements exclusively
2. Never interpolate user input into query strings
3. Code review all database code before merge
4. Use database-level RLS (Row Level Security) to prevent org-level data leakage

**Priority:** Critical

---

## Performance Bottlenecks

### Availability Cache TTL Too Low

**Files:** `internal/availability/cache.go` (TODO)

**Issue:** Architecture doc mentions 60s TTL for Redis cache. At 30s poll interval + 60s TTL = ~90s stale data possible.

**Impact:**
- User sees GPU that was available 90s ago but isn't anymore
- Provision attempt fails
- Poor UX

**Recommendations:**
1. Extend TTL to 120s (double poll interval)
2. Implement smart cache invalidation: on provider error, keep old data longer
3. Add "last_updated" timestamp to offering so UI can show freshness
4. Consider 10-15s TTL for high-demand GPUs (H100, A100)

**Priority:** Medium

---

### No Pagination on List Endpoints

**Files:** `internal/api/handlers.go` (HandleListInstances, HandleListAvailable)

**Issue:**
- TODO comment shows `ListInstances()` returning slice, but no pagination mentioned
- If customer has 1000+ instances, returns all in one response

**Impact:** Large response bodies, slow serialization, high memory usage

**Recommendations:**
1. Implement cursor-based pagination (better than offset for distributed data)
2. Default page size: 20 items
3. Query params: ?page=0&limit=20
4. Return next_cursor in response

**Priority:** Medium

---

### No Caching of Organization Lookups

**Files:** `internal/db/organizations.go`

**Issue:** Every request extracts org_id from JWT and calls database. No caching.

**Impact:** Database query for every API request

**Recommendations:**
1. Cache org metadata in memory (updated on cache miss)
2. Or use Redis: `org:{org_id}` key with 1-hour TTL
3. Invalidate cache on org update

**Priority:** Low (not critical until thousands of concurrent users)

---

## Scaling Limits

### Single PostgreSQL Database

**Files:** Configuration, no sharding strategy documented

**Issue:**
- No documented plan for scaling writes (instance creation, status updates)
- No read replicas mentioned
- As customers scale to 10k+ instances, single database becomes bottleneck

**Recommendations:**
1. Document planned scaling: use PostgreSQL native replication
2. Read replicas for analytics queries (billing, usage)
3. Implement connection pooling (pgBouncer) in front of database
4. Plan for sharding if needed (by org_id)

**Priority:** Low (Phase 2 concern)

---

### No Distributed Tracing for Multi-Provider Scenarios

**Files:** `internal/availability/poller.go`

**Issue:** When provisioning across multiple regions/providers, no way to track request through all systems

**Impact:** Hard to debug: why did provisioning on RunPod fail but E2E succeeded?

**Recommendations:**
1. Use OpenTelemetry for distributed tracing
2. Inject trace ID into all logs and external service calls
3. Correlate RunPod API calls with E2E API calls via trace ID

**Priority:** Medium (Phase 1B improvement)

---

## Missing Critical Features

### No Audit Logging

**Files:** No audit package exists

**Issue:** No record of:
- Who provisioned which instance and when
- Who deleted an instance
- Billing changes
- Auth failures

**Impact:** Compliance risk (if customers are enterprises), security blind spot

**Recommendations:**
1. Create `internal/audit` package
2. Log to separate table: `audit_logs(id, actor, action, resource, details, timestamp)`
3. Immutable append-only design
4. Implement access control: admins can view audit logs, members cannot

**Priority:** High (before enterprise customers)

---

### No Backup/Disaster Recovery Plan

**Files:** No docs or implementation

**Issue:**
- If PostgreSQL corrupted, no recovery
- If Redis lost, availability cache empty but system continues
- If instance data lost, customers don't know what they have

**Recommendations:**
1. Document backup strategy: daily PostgreSQL backups to S3
2. Test restore process monthly
3. Implement point-in-time recovery (WAL archiving)
4. Cache should be ephemeral (it's OK to lose, regenerates on poll)

**Priority:** Medium (before production)

---

### No Instance Timeouts / Auto-Stop

**Files:** Architecture mentions this as feature gap

**Issue:**
- Customer forgets to stop instance
- Runs up bill indefinitely
- No safeguards

**Recommendations:**
1. Implement idle detection: no SSH activity for N hours → auto-stop
2. Or max-lifetime: instance runs for max 7 days then must be manually renewed
3. Send reminder emails before auto-stop
4. Implement as separate service that queries instances and terminates after threshold

**Priority:** Medium (UX improvement)

---

## Testing Gaps

**Critical test categories missing:**

1. **End-to-End Provisioning Flow:**
   - Create instance → cloud-init callback → status=running → health check passes
   - Files to test: all integration points

2. **Spot Instance Interruption:**
   - Instance receives SIGTERM → notifies API → status changes → peer removed
   - Files: health monitor, cloud-init

3. **Provider Failover:**
   - RunPod unavailable → failover to E2E → provision succeeds
   - Files: availability poller, provisioning engine

4. **Concurrent Provision Requests:**
   - Multiple users provision simultaneously → no resource conflicts
   - Files: database layer, WireGuard manager

5. **Database Transaction Rollback:**
   - Provision succeeds but billing fails → state cleaned up
   - Files: all DB operations

6. **Stripe Webhook Security:**
   - Verify signature validation prevents webhook spoofing
   - Files: billing service

7. **JWT Token Validation:**
   - Expired token rejected, invalid signature rejected, valid token accepted
   - Files: auth middleware

8. **Rate Limiting:**
   - User exceeds limit → gets 429 response
   - Files: API middleware

**Priority:** Critical — implement alongside feature implementation

---

## Recommendations by Priority

### Critical (Before Any User Can Use System)

1. Implement database pool and all query methods
2. Implement HTTP server and all API handlers
3. Implement Clerk JWT authentication
4. Implement RunPod provider adapter
5. Implement WireGuard key generation and peer management
6. Implement provisioning orchestration engine
7. Implement configuration loading with validation
8. Implement health check and graceful shutdown
9. Add secret management (no plaintext keys in code)
10. Add SQL injection prevention (prepared statements only)

### High (Before Beta Launch)

1. Implement availability poller and Redis cache
2. Implement Stripe integration and billing checks
3. Implement cloud-init script generation with injection prevention
4. Implement error handling patterns across all services
5. Implement graceful shutdown for all goroutines
6. Implement structured logging with slog
7. Implement database migrations with tooling
8. Implement test suite (70%+ coverage target)
9. Implement spot instance interruption handling
10. Implement audit logging
11. Implement request validation at API boundary
12. Implement rate limiting per user

### Medium (Phase 1 Polish)

1. Implement request ID correlation and distributed tracing
2. Implement pagination for list endpoints
3. Implement health check endpoint with Redis/database verification
4. Implement instance timeout/auto-stop feature
5. Implement provider failover logic (if primary provider down)
6. Document all environment variables and configuration
7. Implement metrics and monitoring dashboards
8. Test provider adapters with real API credentials (E2E Networks)

---

*Concerns audit: 2026-02-24*
