---
phase: 04-auth-instance-lifecycle
verified: 2026-02-25T00:15:00Z
status: passed
score: 19/19 must-haves verified
re_verification:
  previous_status: gaps_found
  previous_score: 17/19
  gaps_closed:
    - "Idempotency-Key header on POST prevents duplicate instance creation — idempotency middleware now resolves Clerk org ID to internal UUID via GetOrgIDByClerkOrgID before all DB calls"
    - "DELETE /api/v1/instances/{id} terminates instance idempotently — WireGuard peer cleanup now calls wgMgr.RemovePeer with CIDR-stripped tunnel IP and computed external port"
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "POST /api/v1/instances with valid Clerk JWT"
    expected: "201 Created with InstanceResponse including price_per_hour"
    why_human: "Requires live Clerk JWT, RunPod API key, and PostgreSQL with v3 schema applied"
  - test: "GET /api/v1/instances with no Clerk JWT"
    expected: "401 with application/problem+json Content-Type and RFC 7807 body"
    why_human: "Requires live HTTP request with Clerk middleware active"
  - test: "GET /api/v1/instances/{id}/events with valid auth"
    expected: "text/event-stream response with initial state event and 30s keepalive pings"
    why_human: "SSE streaming requires live connection; cannot verify timing programmatically"
---

# Phase 4: Auth & Instance Lifecycle Verification Report

**Phase Goal:** Authenticated users can create, list, view, and terminate GPU instances through REST API endpoints, with a full state machine governing instance transitions, organization-scoped access control, rate limiting, idempotency, and pagination

**Verified:** 2026-02-25T00:15:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (Plan 04-04)

---

## Re-Verification Summary

Previous verification (`2026-02-24T23:50:00Z`) found 2 gaps blocking full goal achievement. Plan 04-04 closed both gaps via commits `e8a7650` and `00ec2c8`. This re-verification confirms both gaps are resolved and no regressions were introduced.

### Gaps Closed

**Gap 1 (was Blocker): Idempotency org_id UUID resolution**

`internal/api/idempotency.go` line 69 now calls `dbPool.GetOrgIDByClerkOrgID(ctx, claims.OrgID)` before any DB operation. The resolved `internalOrgID` (a proper UUID) is passed to all four DB calls:
- Line 93: `dbPool.GetIdempotencyKey(ctx, internalOrgID, key)`
- Line 125: `dbPool.CreateIdempotencyKey(ctx, internalOrgID, key, requestHash)`
- Line 130: `dbPool.GetIdempotencyKey(ctx, internalOrgID, key)` (race re-check)
- Line 167: `dbPool.CompleteIdempotencyKey(ctx, internalOrgID, key, ...)`

`claims.OrgID` (Clerk string) is no longer passed anywhere as a DB argument. Runtime PostgreSQL UUID type mismatch eliminated.

**Gap 2 (was Warning): WireGuard peer cleanup on termination**

`internal/provision/engine.go` lines 338-366 replace the previous stub log with actual cleanup:
1. Strips CIDR suffix via `strings.Cut(*inst.WGAddress, "/")` (handles PostgreSQL INET column format)
2. Parses tunnel IP via `net.ParseIP(addrStr)`
3. Computes external port via `wireguard.PortFromTunnelIP(tunnelIP)`
4. Calls `e.wgMgr.RemovePeer(ctx, *inst.WGPublicKey, tunnelIP, externalPort)` — actual WireGuard peer and iptables rule removal
5. Cleanup is best-effort: errors logged, termination does not fail

No stub log `"WireGuard cleanup would happen here"` remains in the file.

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Requests without valid Clerk JWT to /api/v1/* endpoints receive 401 | VERIFIED | `ClerkAuthMiddleware` with empty key returns 401; delegates to `clerkhttp.RequireHeaderAuthorization()` when key is set |
| 2 | Requests with valid JWT have user_id and org_id available in handler context | VERIFIED | `auth.ClaimsFromContext` wraps `clerk.SessionClaimsFromContext`, mapping `Subject->UserID` and `ActiveOrganizationID->OrgID` |
| 3 | Requests with valid JWT but no active organization receive 403 | VERIFIED | `RequireOrg` checks `ActiveOrganizationID == ""`, returns 403 with ProblemDetail |
| 4 | Error responses use RFC 7807 Problem Details format | VERIFIED | `writeProblem` in `errors.go` sets `Content-Type: application/problem+json`, encodes `ProblemDetail` struct |
| 5 | Requests exceeding rate limit receive 429 with Retry-After header | VERIFIED | `OrgRateLimiter.Middleware` sets `Retry-After: 1` and calls `writeProblem` with 429 |
| 6 | Database schema includes idempotency_keys table and new instance columns | VERIFIED | Migration `20250224_v3_auth_instance_lifecycle.sql` creates `idempotency_keys` and adds `name`, `ready_at`, `error_reason` to `instances` |
| 7 | Instance state machine validates transitions: creating->provisioning->booting->running->stopping->terminated, with error from any non-terminal state | VERIFIED | `state.go` defines `validTransitions` map with all 7 states; `CanTransition` enforces valid paths |
| 8 | DB queries for instances are always scoped by org_id (organization isolation) | VERIFIED | `GetInstanceForOrg` uses `WHERE instance_id = $1 AND org_id = $2`; all handler paths call this method |
| 9 | Instance list query supports cursor-based pagination with (created_at, instance_id) keyset | VERIFIED | `ListInstances` uses `(created_at, instance_id) < ($2, $3)` keyset with `ORDER BY created_at DESC, instance_id DESC LIMIT limit+1` |
| 10 | Provisioning engine creates WireGuard keys, renders cloud-init, calls provider, and writes instance to DB | VERIFIED | `engine.Provision` follows full flow: ID gen -> SSH keys -> provider select -> WG keygen -> cloud-init render -> provider.Provision -> db.CreateInstance |
| 11 | Termination updates instance status atomically using optimistic locking (WHERE status = $old) | VERIFIED | `UpdateInstanceStatus` uses `WHERE instance_id = $2 AND status = $3`; checks `RowsAffected() == 1` |
| 12 | Idempotency keys can be stored and retrieved from PostgreSQL with org_id scoping | VERIFIED | `GetOrgIDByClerkOrgID` called at line 69; `internalOrgID` UUID used in all four subsequent DB operations; `go build` and `go vet` pass |
| 13 | POST /api/v1/instances creates a GPU instance and returns CustomerInstance with price_per_hour | VERIFIED | `handleCreateInstance` decodes request, calls `engine.Provision`, returns 201 with `InstanceResponse` including `PricePerHour` |
| 14 | GET /api/v1/instances returns paginated list of instances for authenticated org | VERIFIED | `handleListInstances` uses `ParsePageParams`, `DecodeCursor`, `ListInstances`, encodes next cursor, returns `PageResult[InstanceResponse]` |
| 15 | GET /api/v1/instances/{id} returns instance details with nested connection object | VERIFIED | `handleGetInstance` calls `GetInstanceForOrg`, returns `InstanceResponse` with `ConnectionInfo` (hostname, port, ssh_command) |
| 16 | DELETE /api/v1/instances/{id} terminates instance idempotently (200 if already terminated) | VERIFIED | Idempotent return for already-terminated works; `wgMgr.RemovePeer` called with parsed IP and computed port; cleanup is best-effort |
| 17 | Idempotency-Key header on POST prevents duplicate instance creation | VERIFIED | `IdempotencyMiddleware` resolves Clerk org ID -> internal UUID via `GetOrgIDByClerkOrgID`; all DB calls use `internalOrgID`; no UUID type mismatch |
| 18 | POST /internal/instances/{id}/ready transitions instance from booting to running | VERIFIED | `handleInstanceReady` calls `db.SetInstanceRunning` (atomic `WHERE status = 'booting'`), publishes SSE event if updated |
| 19 | All /api/v1/* routes are protected by Clerk auth + org requirement + rate limiting | VERIFIED | `authChain = clerkAuth(requireOrg(rateLimiter.Middleware(h)))` applied to all 5 public routes in `server.go` |

**Score:** 19/19 truths verified

---

## Required Artifacts

### Plan 01 Artifacts

| Artifact | Status | Details |
|----------|--------|---------|
| `database/migrations/20250224_v3_auth_instance_lifecycle.sql` | VERIFIED | Creates idempotency_keys table, adds name/ready_at/error_reason to instances, adds clerk_org_id to organizations |
| `internal/auth/clerk.go` | VERIFIED | Exports `Claims` struct and `ClaimsFromContext`; wraps `clerk.SessionClaimsFromContext` |
| `internal/api/errors.go` | VERIFIED | Exports `ProblemDetail`, `writeProblem` (RFC 7807), `writeJSON` |
| `internal/api/middleware_auth.go` | VERIFIED | Exports `ClerkAuthMiddleware` and `RequireOrg` |
| `internal/api/middleware_rate.go` | VERIFIED | Exports `OrgRateLimiter` with `Middleware`, `StartCleanup`, `CleanupStale` |
| `internal/api/pagination.go` | VERIFIED | Exports `PageParams`, `PageResult[T]`, `ParsePageParams`, `EncodeCursor`, `DecodeCursor` |
| `internal/config/config.go` | VERIFIED | `ClerkSecretKey` field added, loaded from `CLERK_SECRET_KEY` env var |

### Plan 02 Artifacts

| Artifact | Status | Details |
|----------|--------|---------|
| `internal/provision/state.go` | VERIFIED | 7 states, `validTransitions` map, `CanTransition`, `ExternalState`, `IsTerminal` |
| `internal/db/instances.go` | VERIFIED | `Instance` struct with all v0-v3 columns; 8 CRUD methods with optimistic locking and keyset pagination |
| `internal/db/organizations.go` | VERIFIED | `Organization` and `User` structs; `GetOrganization`, `EnsureOrgAndUser`, `GetOrgIDByClerkOrgID` |
| `internal/db/idempotency.go` | VERIFIED | `IdempotencyKey` struct; `GetIdempotencyKey`, `CreateIdempotencyKey`, `CompleteIdempotencyKey`, `CleanupIdempotencyKeys` |
| `internal/db/ssh_keys.go` | VERIFIED | `SSHKey` struct; `GetSSHKeysByIDs` using `ANY($1)` |
| `internal/provision/engine.go` | VERIFIED | `Engine`, `NewEngine`, `Provision`, `Terminate`; WG cleanup calls `RemovePeer` with CIDR-stripped IP |

### Plan 03 Artifacts

| Artifact | Status | Details |
|----------|--------|---------|
| `internal/api/handlers.go` | VERIFIED | `handleCreateInstance`, `handleListInstances`, `handleGetInstance`, `handleDeleteInstance` with full implementations |
| `internal/api/handlers_internal.go` | VERIFIED | `handleInstanceReady` transitions booting->running and publishes SSE; `handleInstanceHealth` stub |
| `internal/api/idempotency.go` | VERIFIED | `IdempotencyMiddleware` resolves Clerk org ID to internal UUID via `GetOrgIDByClerkOrgID` before all DB calls |
| `internal/api/sse.go` | VERIFIED | `StatusBroker` with subscribe/unsubscribe/publish; `handleInstanceSSE` with keepalive and 30-min max |
| `internal/api/server.go` | VERIFIED | All 7 routes registered; auth chain applied to public routes; internal routes use LocalhostOnly + InternalAuthMiddleware |
| `cmd/gpuctl/main.go` | VERIFIED | Provider registry, RunPod conditional registration, `provision.NewEngine`, engine injected into `ServerDeps` |

### Plan 04 Artifacts (Gap Closure)

| Artifact | Status | Details |
|----------|--------|---------|
| `internal/api/idempotency.go` | VERIFIED | Line 69: `GetOrgIDByClerkOrgID` called; lines 93, 125, 130, 167: `internalOrgID` used in all DB calls; no `claims.OrgID` in DB arguments |
| `internal/provision/engine.go` | VERIFIED | Lines 338-366: `RemovePeer` called with `strings.Cut`-stripped CIDR address; best-effort error handling |

---

## Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `middleware_auth.go` | `auth/clerk.go` | `ClaimsFromContext` extracts Clerk session claims | WIRED | `RequireOrg` calls `clerk.SessionClaimsFromContext`; handlers call `auth.ClaimsFromContext` |
| `middleware_rate.go` | `auth/clerk.go` | Rate limiter reads `ActiveOrganizationID` | WIRED | Line 59 reads `claims.ActiveOrganizationID` for org bucketing |
| `idempotency.go` | `db/organizations.go` | Resolves Clerk org ID to internal UUID using `GetOrgIDByClerkOrgID` | WIRED | Line 69: `dbPool.GetOrgIDByClerkOrgID(ctx, claims.OrgID)` called before all DB operations |
| `provision/engine.go` | `db/instances.go` | Engine calls DB CRUD | WIRED | `e.db.CreateInstance` (line 252), `e.db.UpdateInstanceStatus` (lines 296, 401) |
| `provision/engine.go` | `provision/state.go` | Engine validates transitions | WIRED | `CanTransition(inst.Status, StateStopping)` (line 295) |
| `provision/engine.go` | `wireguard/manager.go` | Calls `RemovePeer` with parsed tunnel IP and computed external port | WIRED | Line 351: `e.wgMgr.RemovePeer(ctx, *inst.WGPublicKey, tunnelIP, externalPort)` |
| `provision/engine.go` | `db/ssh_keys.go` | Engine resolves SSH key IDs | WIRED | `e.db.GetSSHKeysByIDs(ctx, req.SSHKeyIDs)` (line 120) |
| `db/instances.go` | `api/pagination.go` | ListInstances uses keyset pagination | WIRED | `(created_at, instance_id) < ($2, $3)` keyset in SQL query |
| `handlers.go` | `provision/engine.go` | Handlers call Provision/Terminate | WIRED | `s.engine.Provision` (line 155), `s.engine.Terminate` (line 369) |
| `handlers.go` | `db/instances.go` | Handlers call list/get operations | WIRED | `s.db.ListInstances` (line 243), `s.db.GetInstanceForOrg` (lines 307, 351, 380) |
| `server.go` | `middleware_auth.go` | Route registration wraps with auth chain | WIRED | `clerkAuth(requireOrg(rateLimiter.Middleware(h)))` applied to all 5 public routes |
| `handlers_internal.go` | `sse.go` | Ready callback publishes to SSE broker | WIRED | `s.statusBroker.Publish(instanceID, StatusEvent{...})` (line 69) |
| `cmd/gpuctl/main.go` | `provision/engine.go` | main.go creates Engine | WIRED | `provision.NewEngine(provision.EngineDeps{...})` (line 81) |

---

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|---------|
| AUTH-01 | 04-01 | All customer API endpoints require valid Clerk JWT | SATISFIED | `ClerkAuthMiddleware` applied via `authChain` to all /api/v1/* routes |
| AUTH-02 | 04-01 | JWT verification extracts user_id and org_id into request context | SATISFIED | `auth.ClaimsFromContext` maps Clerk claims to `Claims{UserID, OrgID}` |
| AUTH-03 | 04-02 | Users can only access instances belonging to their organization | SATISFIED | `GetInstanceForOrg` enforces `org_id = $2` at query level |
| INST-01 | 04-02 | User can create a GPU instance specifying type, count, region, tier, SSH keys | SATISFIED | `handleCreateInstance` validates and passes all fields to `engine.Provision` |
| INST-02 | 04-02 | User can list their active instances with status and connection info | SATISFIED | `handleListInstances` returns paginated `InstanceResponse` with `ConnectionInfo` |
| INST-03 | 04-02 | User can get details of a specific instance by ID | SATISFIED | `handleGetInstance` returns `InstanceResponse` with all fields |
| INST-04 | 04-02, 04-04 | User can terminate an instance and billing stops | SATISFIED | `handleDeleteInstance` calls `engine.Terminate`; `TerminateInstance` sets `billing_end = NOW()`; WG peer cleaned up via `RemovePeer` |
| INST-05 | 04-02 | Instance follows state machine | SATISFIED | `state.go` enforces valid transitions; engine uses `CanTransition` before updates |
| INST-06 | 04-02, 04-04 | Instance termination is idempotent | SATISFIED | `handleDeleteInstance` returns 200 if already terminated; `TerminateInstance` uses `NOT IN ('terminated', 'error')` |
| INST-07 | 04-03 | Instance ready callback transitions status from booting to running | SATISFIED | `handleInstanceReady` calls `db.SetInstanceRunning` (atomic booting->running) |
| INST-08 | 04-02 | Instance creation response includes confirmed hourly cost | SATISFIED | `InstanceResponse.PricePerHour` populated from `offering.PricePerHour` |
| API-01 | 04-03 | POST /api/v1/instances creates a new GPU instance | SATISFIED | Route registered in `server.go`; `handleCreateInstance` returns 201 |
| API-02 | 04-03 | GET /api/v1/instances lists user's instances | SATISFIED | Route registered; `handleListInstances` returns `PageResult[InstanceResponse]` |
| API-03 | 04-03 | GET /api/v1/instances/{id} returns instance details | SATISFIED | Route registered; `handleGetInstance` returns `InstanceResponse` |
| API-04 | 04-03 | DELETE /api/v1/instances/{id} terminates an instance | SATISFIED | Route registered; `handleDeleteInstance` calls engine terminate |
| API-09 | 04-01 | Error responses never leak upstream provider details | SATISFIED | `InstanceResponse` struct has no upstream fields; generic error messages used for provisioning failures |
| API-10 | 04-02 | All list endpoints support cursor-based pagination | SATISFIED | `ListInstances` with keyset cursor; `ParsePageParams`/`EncodeCursor`/`DecodeCursor` utilities |
| API-11 | 04-03, 04-04 | POST /api/v1/instances accepts Idempotency-Key header | SATISFIED | `IdempotencyMiddleware` resolves Clerk org ID to internal UUID via `GetOrgIDByClerkOrgID`; all DB calls use UUID; no runtime type mismatch |
| API-12 | 04-01 | All customer API endpoints are rate-limited per org | SATISFIED | `NewOrgRateLimiter(rate.Every(100*time.Millisecond), 20)` applied via `authChain` |

**All 19 requirements: SATISFIED**

---

## Anti-Patterns Found

None. Previous blockers have been resolved:
- `"WireGuard cleanup would happen here"` stub removed (replaced with `RemovePeer` call)
- `claims.OrgID` no longer passed as UUID to any DB function

---

## Human Verification Required

### 1. Full Instance Lifecycle End-to-End

**Test:** With `CLERK_SECRET_KEY`, `RUNPOD_API_KEY`, and v3 migrations applied, make `POST /api/v1/instances` with valid JWT and organization-scoped token
**Expected:** 201 response with `id`, `status: "starting"`, `price_per_hour > 0`, `connection.hostname` matching `gpu-{id}.gpu.ai`
**Why human:** Requires live Clerk JWT, RunPod API key, PostgreSQL with migrated schema, and Redis

### 2. Auth Rejection

**Test:** Make `GET /api/v1/instances` with no Authorization header
**Expected:** 401 with `Content-Type: application/problem+json` and body `{"type": "https://api.gpu.ai/errors/...", "status": 401, ...}`
**Why human:** Requires running server with Clerk SDK active to verify the exact response format end-to-end

### 3. SSE Status Streaming

**Test:** Subscribe to `GET /api/v1/instances/{id}/events`, then trigger the ready callback at `POST /internal/instances/{id}/ready`
**Expected:** SSE event stream receives initial state event immediately, then a `"running"` status event after the callback
**Why human:** SSE connection timing and event delivery cannot be verified programmatically without an integration test harness

---

## Build Verification

```
go build ./cmd/gpuctl  — OK
go vet ./...           — OK
```

Committed at:
- `e8a7650` — fix(04-04): resolve Clerk org ID to internal UUID in idempotency middleware
- `00ec2c8` — feat(04-04): implement WireGuard peer cleanup on instance termination

---

_Verified: 2026-02-25T00:15:00Z_
_Verifier: Claude (gsd-verifier)_
