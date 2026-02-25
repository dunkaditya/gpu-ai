---
phase: 06-availability-health-monitoring
verified: 2026-02-25T22:00:00Z
status: passed_with_deferral
score: 5/6 success criteria verified (1 deferred by user decision)
re_verification: false
deferred:
  - truth: "Spot interruptions and instance failures trigger webhook notification to org's configured callback URL"
    status: deferred
    reason: "HLTH-04 webhook delivery intentionally deferred to future phase per user's Phase 6 scoping decision. SSE event delivery implemented instead. CONTEXT.md locked decision: 'No webhooks in this phase — SSE for real-time dashboard, webhooks deferred to future phase.'"
human_verification:
  - test: "SSE endpoint connection and real-time event delivery"
    expected: "GET /api/v1/events maintains open connection, delivers keepalive every 30s, delivers health events in real time"
    why_human: "Cannot verify real-time streaming behavior or keepalive timing programmatically without a live server"
  - test: "GPU availability endpoint with Redis cache populated"
    expected: "GET /api/v1/gpu/available returns non-empty list after provider poll interval, filters work correctly"
    why_human: "Requires live Redis + provider credentials to verify end-to-end cache population and filtering"
---

# Phase 6: Availability + Health Monitoring Verification Report

**Phase Goal:** Background systems continuously poll providers for GPU availability (cached in Redis), select the best-price provider for provisioning, monitor running instances for health and spot interruptions, and notify orgs via webhook on instance failures
**Verified:** 2026-02-25T22:00:00Z
**Status:** passed_with_deferral
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (from ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|---------|
| 1 | Background poller queries all registered providers every 30 seconds and caches results in Redis with 35-second TTL | VERIFIED | `internal/availability/poller.go`: `Start()` polls immediately then on 30s ticker; `cache.go` uses `offeringsKey = "gpu:offerings:all"` with 35s TTL via `NewCache(redisClient, 35*time.Second)` in main.go |
| 2 | GET /api/v1/gpu/available returns aggregated GPU offerings with pricing by region and tier — without revealing provider identity | VERIFIED | `internal/api/handlers_gpu.go` reads from `availCache.GetOfferings`, `AvailableOffering` struct has no `Provider` field (defense-by-omission), route registered behind `authChain` in `server.go:99` |
| 3 | Provisioning engine automatically selects the best-price provider when creating an instance | VERIFIED | `internal/provision/engine.go`: `selectProviderCandidates` uses `sort.SliceStable` by price ascending, `Provision` retries up to 3 candidates on failure |
| 4 | Health monitor detects spot instance interruptions and automatically stops billing | VERIFIED | `internal/health/monitor.go`: `handleSpotInterruption` calls `CloseBillingSession`, sets error state with optimistic locking, logs `interrupted` event |
| 5 | Instance ready callback transitions instance status from booting to running | VERIFIED | `internal/api/handlers_internal.go:45`: `SetInstanceRunning` called on POST to `/internal/instances/{id}/ready`, logs `ready` event to `instance_events` |
| 6 | Spot interruptions and instance failures trigger webhook notification to org's configured callback URL | FAILED | SSE events are published via `srv.PublishOrgEvent` (SSE broker), but no HTTP webhook POST is delivered to any org-configured external URL. No `webhook_url` column exists in the database schema. |

**Score: 5/6 truths verified**

---

### Required Artifacts (Level 1: Exists, Level 2: Substantive, Level 3: Wired)

| Artifact | Exists | Substantive | Wired | Status | Notes |
|----------|--------|-------------|-------|--------|-------|
| `database/migrations/20250225_v6_availability_health.sql` | YES | YES (20 lines, `instance_events` table, CHECK constraint, 2 indexes) | N/A (migration) | VERIFIED | `event_type CHECK (event_type IN ('ready', 'interrupted', 'failed', 'terminated'))` |
| `internal/availability/types.go` | YES | YES (63 lines, `AvailableOffering`, `ToAvailableOffering`, no `Provider` field) | YES (used by cache, poller, handlers_gpu) | VERIFIED | Defense-by-omission confirmed |
| `internal/availability/cache.go` | YES | YES (54 lines, `SetOfferings`, `GetOfferings`, single JSON key) | YES (poller calls `SetOfferings`, handler calls `GetOfferings`) | VERIFIED | Key = `"gpu:offerings:all"`, TTL passed from `NewCache` |
| `internal/availability/poller.go` | YES | YES (108 lines, `Start`, `poll`, concurrent providers, mutex) | YES (goroutine started in `main.go:129`) | VERIFIED | Immediate startup poll before ticker loop confirmed |
| `internal/db/events.go` | YES | YES (81 lines, `InstanceEvent`, `CreateInstanceEvent`, `ListInstanceEventsByOrg`, `ListInstanceEventsByInstance`) | YES (called from monitor, handlers_internal, engine) | VERIFIED | |
| `internal/provider/types.go` (GPUOffering extension) | YES | YES (`CPUCores`, `RAMGB`, `StorageGB` fields at lines 39-41) | YES (used by `ToAvailableOffering`) | VERIFIED | |
| `internal/api/handlers_gpu.go` | YES | YES (116 lines, `handleListGPUAvailability`, `matchesFilters`, 6 filter params) | YES (route registered in `server.go:99-100`) | VERIFIED | |
| `internal/health/monitor.go` | YES | YES (315 lines, `Start`, `checkAll`, `checkInstance`, `handleSpotInterruption`, `handleNonSpotFailure`) | YES (goroutine in `main.go:165`) | VERIFIED | |
| `internal/api/sse.go` (OrgEventBroker) | YES | YES (OrgEventBroker with `Subscribe`, `Unsubscribe`, `Publish`, buffer 20) | YES (wired in `server.go`, called from `PublishOrgEvent`) | VERIFIED | |
| `internal/api/handlers_events.go` | YES | YES (162 lines, `handleEvents`, `handleListEventsREST`, `handleOrgSSEStream`) | YES (route registered in `server.go:107-108`) | VERIFIED | |
| `cmd/gpuctl/main.go` (poller + monitor goroutines) | YES | YES (availPoller goroutine at line 129, healthMonitor goroutine at line 165) | YES (both receive ctx for cancellation) | VERIFIED | `AvailCache` passed to `ServerDeps` at line 138 |
| `internal/config/config.go` (PricingMarkupPct) | YES | YES (field at line 70, parsed from `PRICING_MARKUP_PCT` env, default 15.0) | YES (passed to `NewPoller` in main.go) | VERIFIED | |

---

### Key Link Verification

| From | To | Via | Status | Detail |
|------|----|-----|--------|--------|
| `internal/availability/poller.go` | `internal/provider/registry.go` | `registry.All()` | WIRED | Line 57: `providers := p.registry.All()` |
| `internal/availability/poller.go` | `internal/availability/cache.go` | `cache.SetOfferings` | WIRED | Line 100: `p.cache.SetOfferings(ctx, allOfferings)` |
| `internal/availability/cache.go` | go-redis | `redis.Set` with `"gpu:offerings:all"` key | WIRED | Line 36: `c.redis.Set(ctx, offeringsKey, data, c.ttl)` |
| `internal/api/handlers_gpu.go` | `internal/availability/cache.go` | `cache.GetOfferings` | WIRED | Line 61: `s.availCache.GetOfferings(ctx)` |
| `internal/api/server.go` | `internal/api/handlers_gpu.go` | Route `GET /api/v1/gpu/available` | WIRED | Line 99-100 of server.go |
| `internal/provision/engine.go` | `internal/provider/registry.go` | `sort.SliceStable` on candidates from `registry.All()` | WIRED | Line 643 in engine.go: `sort.SliceStable(candidates, ...)` |
| `cmd/gpuctl/main.go` | `internal/availability/poller.go` | `go availPoller.Start(ctx)` goroutine | WIRED | Line 129 of main.go |
| `cmd/gpuctl/main.go` | `internal/health/monitor.go` | `go healthMonitor.Start(ctx)` goroutine | WIRED | Line 165 of main.go |
| `cmd/gpuctl/main.go` | `internal/api/server.go` | `AvailCache` in `ServerDeps` | WIRED | Line 138 of main.go |
| `internal/api/handlers_events.go` | `internal/api/sse.go` | `orgEventBroker.Subscribe` | WIRED | Line 114 of handlers_events.go |
| `internal/api/handlers_events.go` | `internal/db/events.go` | `ListInstanceEventsByOrg` | WIRED | Line 76 of handlers_events.go |
| `internal/health/monitor.go` | `internal/provider/registry.go` | `registry.Get(upstreamProvider)` | WIRED | Line 110 of monitor.go |
| `internal/health/monitor.go` | `internal/db/instances.go` | `ListActiveInstances` | WIRED | Line 78 of monitor.go |
| `internal/health/monitor.go` | `internal/db/events.go` | `CreateInstanceEvent` for interruption/failure | WIRED | Lines 204, 305 of monitor.go |
| `internal/health/monitor.go` | `internal/db/billing.go` | `CloseBillingSession` on spot interruption | WIRED | Lines 184, 285 of monitor.go |
| `internal/api/handlers_internal.go` | `internal/db/events.go` | `CreateInstanceEvent` for ready event | WIRED | Line 79 of handlers_internal.go |
| `internal/health/monitor.go` | **org webhook URL (external HTTP POST)** | HTTP POST to org-configured URL | NOT WIRED | No webhook column, no HTTP delivery code exists |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| AVAIL-01 | 06-01 | Background poller queries all providers every 30 seconds | SATISFIED | `poller.go:Start()` polls on 30s ticker, `registry.All()` iterates all providers |
| AVAIL-02 | 06-01 | GPU offerings cached in Redis with 35-second TTL | SATISFIED | `cache.go:SetOfferings` sets `gpu:offerings:all` key with `c.ttl` (35s from `main.go:121`) |
| AVAIL-03 | 06-02 | User can view available GPU types with pricing by region and tier | SATISFIED | `handlers_gpu.go:handleListGPUAvailability` returns `GPUAvailabilityResponse{Available: filtered}` with region/tier fields |
| AVAIL-04 | 06-02 | Availability response aggregates across providers without revealing provider identity | SATISFIED | `AvailableOffering` struct has no `Provider` field (defense-by-omission confirmed in types.go) |
| AVAIL-05 | 06-02 | Provisioning engine selects best-price provider automatically | SATISFIED | `engine.go:selectProviderCandidates` uses `sort.SliceStable` by `PricePerHour`, Provision retries up to 3 candidates |
| HLTH-01 | 06-03 | Background goroutine monitors instance health every 60 seconds | SATISFIED | `monitor.go:Start()` uses 60s ticker (from `main.go:161: Interval: 60 * time.Second`) |
| HLTH-02 | 06-03 | Spot instance interruption detected and billing stopped automatically | SATISFIED | `handleSpotInterruption` calls `CloseBillingSession(ctx, inst.InstanceID, time.Now().UTC())` |
| HLTH-03 | 06-03 | Instance ready callback received from booted instances | SATISFIED | `handlers_internal.go:handleInstanceReady` at POST `/internal/instances/{id}/ready` transitions booting->running via `SetInstanceRunning` and logs `ready` event |
| HLTH-04 | 06-04 | Spot interruption and instance failure events trigger webhook notification to org's configured callback URL | BLOCKED | SSE events delivered, but no HTTP webhook POST implemented. No `webhook_url` on orgs. REQUIREMENTS.md description explicitly says "org's configured callback URL" — SSE is not a webhook. |
| API-05 | 06-02 | GET /api/v1/gpu/available returns GPU availability with optional filters | SATISFIED | Route registered in server.go, `matchesFilters` handles 6 query params (gpu_model, region, tier, min_price, max_price, min_vram) |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/api/handlers_internal.go` | 125 | Comment: "full last_seen update deferred to Phase 6" in `handleInstanceHealth` | INFO | health ping handler logs but does not update any `last_seen` or `last_health_ping` timestamp. Phase 6 was the deferred phase — this was not completed. Does not block core goal. |

---

### Human Verification Required

#### 1. SSE Real-Time Event Streaming

**Test:** Connect to `GET /api/v1/events` with a valid JWT, then trigger a spot interruption simulation
**Expected:** SSE connection stays open, keepalive comment appears every 30s, event data appears within one poll cycle after health monitor detects interruption
**Why human:** Cannot verify SSE real-time behavior, flushing, and timing without a live server

#### 2. GPU Availability with Live Cache

**Test:** Start the server with RunPod API key configured, wait 30s, then call `GET /api/v1/gpu/available`
**Expected:** Non-empty JSON array with `available` key, prices show 15% markup over RunPod wholesale prices
**Why human:** Requires live RunPod credentials and running server + Redis to verify end-to-end cache population

#### 3. Availability Filters End-to-End

**Test:** Call `GET /api/v1/gpu/available?gpu_model=RTX4090&tier=spot&max_price=5.0` with cached data
**Expected:** Only offerings matching all three filters in response (case-insensitive GPU model match)
**Why human:** Requires live Redis data with known offerings to verify filter correctness

---

### Gaps Summary

**1 gap blocking full goal achievement:**

**HLTH-04 — Webhook delivery not implemented.** The ROADMAP success criterion 6 states: "Spot interruptions and instance failures trigger webhook notification to org's configured callback URL." The 06-03 PLAN's `must_haves` truth states: "Event notification callback (onEvent) invoked for SSE integration." The 06-04 PLAN's `must_haves` truth states: "Health monitor OnEvent callback publishes to OrgEventBroker for SSE."

The implementation substitutes SSE event delivery for webhook delivery — this is NOT the same as the requirement. HLTH-04 and the ROADMAP success criterion explicitly require an HTTP POST to "org's configured callback URL." There is:
- No `webhook_url` column in the organizations table
- No DB method to retrieve an org's webhook URL
- No HTTP POST delivery in any code path
- Only an SSE broker (`PublishOrgEvent`) which is a different delivery mechanism

The organizations table has no column for a webhook endpoint. The phase scope in 06-PLAN.md claims HLTH-04 as "completed" but the actual delivery mechanism (external HTTP webhook) was never implemented — only in-process SSE was wired.

**Note on scope:** 06-03 PLAN's `must_haves.truths` item 8 reads "Event notification callback (onEvent) invoked for SSE integration" — this narrowed HLTH-04's scope to SSE-only in the plan, but the actual requirement (HLTH-04 + ROADMAP criterion 6) requires webhook delivery. The plans diverged from the requirement.

---

## Build Verification

`go build ./...` from project root: **EXIT 0** — all packages compile cleanly.

---

_Verified: 2026-02-25T22:00:00Z_
_Verifier: Claude (gsd-verifier)_
