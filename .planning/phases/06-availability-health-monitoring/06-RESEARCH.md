# Phase 6: Availability + Health Monitoring - Research

**Researched:** 2026-02-25
**Domain:** Background polling, Redis caching, best-price provider selection, instance health monitoring, SSE event streaming, event persistence
**Confidence:** HIGH

## Summary

Phase 6 implements the background systems that make GPU.ai's real-time availability and reliability possible. There are four major subsystems: (1) an availability poller that queries all registered providers every 30 seconds and caches GPU offerings in Redis with 35-second TTL, (2) an API endpoint that reads the cache and returns aggregated, provider-stripped offerings with server-side filtering, (3) a best-price provider selection upgrade to the existing provisioning engine (replacing first-match with cheapest-first plus automatic fallback), and (4) a health monitor that polls provider APIs for running instance status, detects spot interruptions and failures, stops billing automatically, and emits events via SSE and to a Postgres event log.

The codebase already has strong foundations: `internal/availability/` has stub files with TODO comments, `internal/health/` has a stub monitor, the provisioning engine has a `selectProvider` method explicitly noting "Phase 6 adds best-price", the SSE broker and `StatusBroker` pattern already exist in `internal/api/sse.go`, and the billing session close mechanism is proven. The main work is implementing the stubs, upgrading SSE from per-instance to per-org, adding the `instance_events` table, and wiring everything into `main.go`.

**Primary recommendation:** Implement in five waves: (1) migration + Redis cache + poller, (2) availability API endpoint, (3) best-price provider selection with fallback, (4) health monitor + billing stop + event logging, (5) per-org SSE + REST event catch-up endpoint.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Availability API response is a flat list of offerings (single array), not grouped -- client decides presentation
- Rich fields per offering: gpu_model, vram_gb, cpu_cores, ram_gb, storage_gb, price_per_hour, region, tier, available_count, avg_uptime_pct
- Full server-side filtering via query params: gpu_model, region, tier, min/max price, min_vram
- Pricing shows GPU.ai's markup price only -- provider costs are internal, users see our price sheet
- Pure cost optimization for provider selection: cheapest provider wins
- Tiebreaker: higher-margin provider wins (existing priority list handles this -- no new mechanism)
- No health-based exclusion from selection -- health monitoring handles failures post-provisioning
- Automatic fallback: if selected provider fails to provision, retry with next cheapest provider (cap at 2-3 attempts)
- Health monitor polls provider APIs for instance status (no agent/heartbeat on instances)
- Spot interruption: immediately stop billing ticker and fire event notification
- Non-spot failure: retry health check 2-3 times over short window before declaring failure (avoids false positives from transient network blips)
- Instance ready detection: cloud-init callback from the instance itself (already designed in cloud-init template)
- No webhooks in this phase -- SSE for real-time dashboard, webhooks deferred to future phase
- Four event types: ready, interrupted, failed, terminated
- Events logged to `instance_events` table (audit log, SSE data source, future webhook source)
- SSE: one per-org stream (single connection streams all org instance events)
- Rich event payloads: event_type, instance_id, gpu_model, region, created_at, plus event-specific metadata (interruption reason, failure details)
- SSE is live-only, no replay support -- REST endpoint `GET /api/v1/events?since=<timestamp>` for catch-up on reconnect
- EventSource handles reconnection; client hits REST endpoint to backfill missed events and merges with live stream
- Instance events table schema already designed (in CONTEXT.md)

### Claude's Discretion
- Polling interval tuning and backoff strategy
- SSE connection management (keepalive, cleanup of stale connections)
- Exact retry timing for non-spot failure checks
- Event metadata schema details per event type

### Deferred Ideas (OUT OF SCOPE)
- Webhook delivery to org-configured callback URLs -- future phase (will use instance_events table as data source)
- Provider health-based exclusion from selection -- not needed for v1, revisit if failure rates become an issue
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AVAIL-01 | Background poller queries all providers every 30 seconds | Poller pattern (Pattern 1) with 30s ticker, concurrent provider polling, 35s Redis TTL. Stub already exists in `internal/availability/poller.go`. |
| AVAIL-02 | GPU offerings cached in Redis with 35-second TTL | Redis cache pattern (Pattern 2) using single aggregated key `gpu:offerings:all` instead of per-offering SCAN. 35s TTL ensures stale data expires between polls. |
| AVAIL-03 | User can view available GPU types with pricing by region and tier | API handler reads from Redis cache, strips provider identity, applies server-side filtering. Response uses `AvailableOffering` struct with rich fields per CONTEXT.md. |
| AVAIL-04 | Availability response aggregates across providers without revealing provider identity | `AvailableOffering` struct uses defense-by-omission (no Provider field). Offerings from all providers merged into single flat array. |
| AVAIL-05 | Provisioning engine selects best-price provider automatically | Upgrade `selectProvider` from first-match to sort-by-price with fallback. 2-3 retry attempts with next-cheapest provider. |
| HLTH-01 | Background goroutine monitors instance health every 60 seconds | Health monitor pattern (Pattern 4) polls provider `GetStatus` for all running/booting instances. Stub exists in `internal/health/monitor.go`. |
| HLTH-02 | Spot instance interruption detected and billing stopped automatically | Monitor detects "terminated"/"exited" status for spot instances, calls `CloseBillingSession`, transitions to error state, logs event to `instance_events`. |
| HLTH-03 | Instance ready callback received from booted instances | Already implemented in Phase 4 (`handleInstanceReady` in `handlers_internal.go`). Phase 6 adds event logging to `instance_events` table. |
| HLTH-04 | Spot interruption and instance failure events trigger webhook notification to org's configured callback URL | Webhooks deferred per CONTEXT.md. This phase logs events to `instance_events` table and emits via per-org SSE stream. REST catch-up endpoint provides event history. |
| API-05 | GET /api/v1/gpu/available returns GPU availability with optional filters | New handler with query param filtering: gpu_model, region, tier, min_price, max_price, min_vram. Behind auth chain. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-redis/v9 | v9.18.0 | Redis cache for GPU offerings | Already in go.mod, used for health check. Standard Go Redis client. |
| pgx/v5 | v5.8.0 | PostgreSQL for instance_events table | Already in go.mod, used throughout project. |
| stdlib net/http | Go 1.24 | SSE streaming, API endpoints | Project convention: no frameworks. |
| stdlib encoding/json | Go 1.24 | Redis value serialization, API responses | Already used throughout. |
| stdlib time | Go 1.24 | Ticker-based polling, TTL management | Standard for background goroutines. |
| stdlib sync | Go 1.24 | Thread-safe broker for SSE subscribers | Already used in StatusBroker pattern. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stdlib log/slog | Go 1.24 | Structured logging for poller/monitor | All background goroutines per project convention. |
| stdlib context | Go 1.24 | Cancellation propagation | All I/O functions per project convention. |
| stdlib sort | Go 1.24 | Price-based provider sorting | Best-price selection in provisioning engine. |
| stdlib strconv | Go 1.24 | Query param parsing | Availability API filter parsing. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Single Redis key for all offerings | Per-offering SCAN pattern | SCAN is O(N) and requires iteration; single key with JSON array is O(1) GET and atomic. The TODO stubs suggest per-key SCAN, but with <100 offerings, a single JSON blob is simpler and faster. |
| Polling provider APIs for health | Agent-based heartbeat on instances | User explicitly chose no agent/heartbeat -- poll provider APIs. Simpler, no instance-side dependency. |
| Per-instance SSE (current) | Per-org SSE (required) | Must upgrade: current SSE is per-instance (subscribe by instance_id). Phase 6 requires per-org stream. |

**Installation:**
No new dependencies required. All libraries already in go.mod.

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── availability/
│   ├── poller.go          # Background 30s poller goroutine
│   ├── cache.go           # Redis read/write for offerings
│   └── types.go           # AvailableOffering (customer-facing type)
├── health/
│   └── monitor.go         # Background 60s health check goroutine
├── api/
│   ├── handlers_gpu.go    # GET /api/v1/gpu/available handler
│   ├── handlers_events.go # GET /api/v1/events handler + per-org SSE
│   └── sse.go             # Upgraded: OrgBroker for per-org SSE streams
├── provision/
│   └── engine.go          # Upgraded selectProvider: best-price + fallback
├── db/
│   ├── instances.go       # Existing + ListRunningInstances (all orgs)
│   └── events.go          # NEW: instance_events CRUD
└── config/
    └── config.go          # No changes needed -- already has Redis, providers
```

### Pattern 1: Background Poller (Availability)
**What:** Goroutine with time.Ticker that polls all providers concurrently and writes results to Redis cache.
**When to use:** AVAIL-01, AVAIL-02
**Implementation notes:**
- Poll all providers concurrently using `sync.WaitGroup` or `errgroup`
- Per-provider error isolation: log error and continue (one provider failure must not block others)
- Write all offerings as single JSON array to Redis key `gpu:offerings:all` with 35s TTL
- TTL of 35s (not 30s) ensures brief overlap where cached data is still valid during next poll
- Initial poll on startup (don't wait 30s for first data)
- Backoff strategy for failing providers: exponential backoff up to 5 minutes, reset on success

```go
// Source: follows existing billing ticker pattern in internal/billing/ticker.go
type Poller struct {
    registry *provider.Registry
    cache    *Cache
    interval time.Duration
    logger   *slog.Logger
}

func (p *Poller) Start(ctx context.Context) {
    // Poll immediately on startup
    p.poll(ctx)

    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            p.logger.Info("availability poller stopped")
            return
        case <-ticker.C:
            p.poll(ctx)
        }
    }
}

func (p *Poller) poll(ctx context.Context) {
    providers := p.registry.All()
    var mu sync.Mutex
    var allOfferings []provider.GPUOffering
    var wg sync.WaitGroup

    for _, prov := range providers {
        wg.Add(1)
        go func(prov provider.Provider) {
            defer wg.Done()
            offerings, err := prov.ListAvailable(ctx)
            if err != nil {
                p.logger.Error("poll failed", "provider", prov.Name(), "error", err)
                return
            }
            mu.Lock()
            allOfferings = append(allOfferings, offerings...)
            mu.Unlock()
        }(prov)
    }
    wg.Wait()

    if err := p.cache.SetOfferings(ctx, allOfferings); err != nil {
        p.logger.Error("failed to cache offerings", "error", err)
    }
}
```

### Pattern 2: Redis Cache (Single Key with JSON Array)
**What:** Store all provider offerings as a single JSON array under one Redis key, with 35s TTL.
**When to use:** AVAIL-02, AVAIL-03
**Why not per-key SCAN:** With <100 offerings total, a single `GET` + `json.Unmarshal` is simpler and faster than `SCAN` + N `GET`s. Also atomic -- no partial reads during writes.

```go
type Cache struct {
    redis *redis.Client
    ttl   time.Duration
}

const offeringsKey = "gpu:offerings:all"

func (c *Cache) SetOfferings(ctx context.Context, offerings []provider.GPUOffering) error {
    data, err := json.Marshal(offerings)
    if err != nil {
        return fmt.Errorf("marshal offerings: %w", err)
    }
    return c.redis.Set(ctx, offeringsKey, data, c.ttl).Err()
}

func (c *Cache) GetOfferings(ctx context.Context) ([]provider.GPUOffering, error) {
    data, err := c.redis.Get(ctx, offeringsKey).Bytes()
    if err == redis.Nil {
        return nil, nil // No cached data yet
    }
    if err != nil {
        return nil, err
    }
    var offerings []provider.GPUOffering
    if err := json.Unmarshal(data, &offerings); err != nil {
        return nil, fmt.Errorf("unmarshal offerings: %w", err)
    }
    return offerings, nil
}
```

### Pattern 3: Customer-Facing Offering Type (Defense by Omission)
**What:** A struct that does not include the `Provider` field, ensuring provider identity can never leak.
**When to use:** AVAIL-04, API-05
**Implementation notes:**
- `AvailableOffering` struct has fields per CONTEXT.md: gpu_model, vram_gb, cpu_cores, ram_gb, storage_gb, price_per_hour, region, tier, available_count, avg_uptime_pct
- Note: `cpu_cores`, `ram_gb`, `storage_gb`, `avg_uptime_pct` are not currently in `provider.GPUOffering`. These need to either be added to the provider offering or set to reasonable defaults/computed values.
- Pricing: The user decided "GPU.ai's markup price only." The current `GPUOffering.PricePerHour` is the provider price. A markup function must be applied before caching or before returning to customer.

```go
// internal/availability/types.go
type AvailableOffering struct {
    GPUModel       string  `json:"gpu_model"`
    VRAMGB         int     `json:"vram_gb"`
    CPUCores       int     `json:"cpu_cores"`
    RAMGB          int     `json:"ram_gb"`
    StorageGB      int     `json:"storage_gb"`
    PricePerHour   float64 `json:"price_per_hour"`
    Region         string  `json:"region"`
    Tier           string  `json:"tier"`
    AvailableCount int     `json:"available_count"`
    AvgUptimePct   float64 `json:"avg_uptime_pct"`
}
```

### Pattern 4: Health Monitor (Provider API Polling)
**What:** Background goroutine that polls provider APIs for running instance status, detects interruptions and failures, stops billing, logs events.
**When to use:** HLTH-01, HLTH-02, HLTH-04
**Implementation notes:**
- Query DB for all instances with status in ('running', 'booting')
- For each instance, call provider's `GetStatus(ctx, upstreamID)`
- Spot interruption detection: RunPod returns `desiredStatus: "EXITED"` when a spot pod is interrupted. The adapter maps this to `status: "terminated"`. If instance tier is "spot" and provider reports terminated/exited, this is a spot interruption.
- Non-spot failure: if provider reports error/terminated for on-demand instance, retry 2-3 times over ~30s before declaring failure
- On failure/interruption: close billing session, transition instance to error state, log event to `instance_events`, emit SSE event
- Concurrency: process instances concurrently with bounded parallelism (e.g., `errgroup` with limit of 10)

```go
type Monitor struct {
    db       *db.Pool
    registry *provider.Registry
    engine   *provision.Engine
    logger   *slog.Logger
    interval time.Duration
    onEvent  func(event db.InstanceEvent) // SSE notification
}

func (m *Monitor) Start(ctx context.Context) {
    // Similar pattern to billing ticker
    m.checkAll(ctx) // Check immediately on startup
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.checkAll(ctx)
        }
    }
}
```

### Pattern 5: Per-Org SSE Broker
**What:** Upgrade the existing StatusBroker from per-instance to per-org subscription model.
**When to use:** SSE per-org streaming decision from CONTEXT.md
**Implementation notes:**
- Current `StatusBroker` subscribes by `instanceID` -- needs parallel `OrgBroker` that subscribes by `orgID`
- Keep the existing per-instance broker for backward compatibility (Phase 4 SSE)
- New SSE endpoint: `GET /api/v1/events` streams all org instance events
- Max connection duration: 30 minutes (same as existing SSE pattern)
- Keepalive: 30s comments (same as existing)
- No replay: client uses REST `GET /api/v1/events?since=<timestamp>` for catch-up

```go
type OrgEventBroker struct {
    mu          sync.RWMutex
    subscribers map[string][]chan InstanceEventPayload // org_id -> channels
}
```

### Pattern 6: Best-Price Provider Selection with Fallback
**What:** Sort matching providers by price (ascending), try cheapest first, fallback to next-cheapest on failure.
**When to use:** AVAIL-05
**Implementation notes:**
- Collect all matching offerings across all providers
- Sort by price ascending, tiebreak by margin (lower upstream cost = higher margin = preferred)
- Try top match, on provision failure try next, cap at 3 attempts
- The current `selectProvider` iterates registry and returns first match. Replace with sort + retry loop.

```go
func (e *Engine) selectProvider(ctx context.Context, req ProvisionRequest) (provider.Provider, *provider.GPUOffering, error) {
    // 1. Collect all matching offerings from all providers
    var candidates []struct {
        prov     provider.Provider
        offering provider.GPUOffering
    }
    for _, prov := range e.registry.All() {
        offerings, err := prov.ListAvailable(ctx)
        if err != nil {
            e.logger.Warn("provider unavailable", "provider", prov.Name())
            continue
        }
        for _, o := range offerings {
            if matches(o, req) {
                candidates = append(candidates, struct{...}{prov, o})
            }
        }
    }
    // 2. Sort by price ascending
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].offering.PricePerHour < candidates[j].offering.PricePerHour
    })
    // Return first candidate (Provision handles fallback retry)
    ...
}
```

### Anti-Patterns to Avoid
- **Per-key Redis SCAN for offerings:** With <100 GPU offerings, SCAN is unnecessary complexity. Use single key with JSON array for atomic reads.
- **Blocking health checks:** Health monitor must not block on a single slow provider. Use bounded concurrency with timeouts per check.
- **Polling instances individually for health when no instances exist:** Short-circuit if `ListRunningInstances` returns empty slice.
- **Mixing SSE event types:** Per-instance status events (Phase 4) and per-org instance events (Phase 6) serve different purposes. Keep both broker types separate.
- **Float comparison for price sorting:** Use `<` comparison on `float64` for sorting (acceptable for price comparison). Do not use equality checks on floats.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Redis TTL cache | Custom expiry tracking | go-redis `Set` with TTL duration | Redis handles TTL natively, no GC needed |
| JSON serialization | Custom binary format | stdlib `encoding/json` | Project convention, offerings are small (<10KB) |
| Concurrent polling | Manual goroutine management | `sync.WaitGroup` or `errgroup` | Proven patterns, handles panics and errors cleanly |
| SSE protocol | Custom streaming format | Existing `writeSSEEvent` helper + stdlib `fmt.Fprintf` | Already implemented in `sse.go`, spec-compliant |
| Timer-based polling | `time.Sleep` loops | `time.NewTicker` + `select` | Cancelable via context, consistent intervals, idiomatic Go |
| Price-based sorting | Custom sort algorithm | `sort.Slice` | Standard library, handles edge cases |

**Key insight:** The codebase already has proven patterns for every subsystem needed: billing ticker for background goroutines, SSE broker for event streaming, provider registry for multi-provider iteration, and optimistic locking for state transitions. Phase 6 applies these same patterns to new domains.

## Common Pitfalls

### Pitfall 1: Stale Cache Returning Empty Data on Startup
**What goes wrong:** Server starts, availability endpoint returns empty array because poller hasn't run yet.
**Why it happens:** Poller waits 30s before first tick.
**How to avoid:** Call `poll()` immediately on startup before entering the ticker loop. The `Start` method should execute one poll synchronously before the ticker starts.
**Warning signs:** First 30 seconds of API traffic returns `{"available": []}`.

### Pitfall 2: Provider Failure Cascading to Cache
**What goes wrong:** One provider's API is down, poller overwrites cache with partial data (offerings from working providers only), losing the failing provider's offerings entirely.
**Why it happens:** Poller collects offerings from all providers, then does a single cache SET replacing all previous data.
**How to avoid:** Only overwrite cache when at least one provider returns data. If all providers fail, keep the existing cached data (it has TTL and will expire naturally). Alternatively, merge per-provider: cache per-provider separately and aggregate on read. The single-key approach is simpler if we accept that a single provider failure drops its offerings from the aggregated view (which is correct -- if a provider is unreachable, its offerings are indeed unavailable).
**Warning signs:** Intermittent drops in available offering count.

### Pitfall 3: Health Monitor Creating Duplicate Events
**What goes wrong:** Health monitor detects same interruption on consecutive ticks and creates duplicate events in `instance_events`.
**Why it happens:** Instance is interrupted, monitor logs event, but instance status update fails or hasn't committed yet when next tick runs.
**How to avoid:** Use optimistic locking on instance status transition: only log the event if the status transition succeeds (same pattern as `UpdateInstanceStatus` with fromStatus check). If the transition returns `updated=false`, skip the event.
**Warning signs:** Multiple "interrupted" events for the same instance within seconds.

### Pitfall 4: SSE Connection Leak
**What goes wrong:** Client disconnects but server keeps goroutine alive, leading to goroutine leak.
**Why it happens:** SSE handler doesn't detect client disconnect.
**How to avoid:** Use `r.Context().Done()` channel to detect client disconnects (already implemented in current SSE handler). Add max duration timer (30 minutes, matching existing pattern). Clean up subscriber on all exit paths.
**Warning signs:** Goroutine count grows monotonically, memory leak over time.

### Pitfall 5: Race Between Health Monitor and User Termination
**What goes wrong:** User terminates instance while health monitor is processing the same instance, leading to conflicting state transitions.
**Why it happens:** Health monitor reads instance status, then user terminates, then health monitor tries to transition.
**How to avoid:** Use optimistic locking (`UpdateInstanceStatus` with fromStatus). If the CAS fails, re-read and check. Already proven pattern in the codebase.
**Warning signs:** Occasional "status already changed" log warnings (benign if handled correctly).

### Pitfall 6: RunPod Spot Interruption Detection
**What goes wrong:** Spot interruption not detected because adapter status mapping doesn't distinguish spot interruption from user-initiated stop.
**Why it happens:** RunPod returns `desiredStatus: "EXITED"` for both spot interruption and explicit stop. The current adapter maps this to `"terminated"`.
**How to avoid:** Cross-reference the instance's tier (stored in DB). If tier is "spot" and provider reports terminated/exited while we haven't initiated termination (instance status is still "running" in our DB), this is a spot interruption. Add an `IsInterrupted` or `Reason` field to `InstanceStatus` for better disambiguation if RunPod provides additional context in the future.
**Warning signs:** Spot interruptions logged as generic terminations.

### Pitfall 7: Markup Price Calculation Timing
**What goes wrong:** Provider prices cached without markup, customer sees provider wholesale prices.
**Why it happens:** `GPUOffering.PricePerHour` from providers is the cost price, not the retail price.
**How to avoid:** Apply markup during the cache write (in the poller, not in the API handler). This ensures cached data is always retail-ready. Alternatively, apply markup in the availability API handler before returning to customer.
**Warning signs:** Suspiciously low prices in availability API, prices matching provider docs exactly.

## Code Examples

Verified patterns from the existing codebase:

### Background Goroutine with Ticker (from billing/ticker.go)
```go
// Source: internal/billing/ticker.go - proven pattern
func (t *BillingTicker) Start(ctx context.Context) {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            t.logger.Info("billing ticker stopped")
            return
        case <-ticker.C:
            t.runTick(ctx)
        }
    }
}
```

### SSE Event Writing (from api/sse.go)
```go
// Source: internal/api/sse.go - existing pattern
func writeSSEEvent(w http.ResponseWriter, eventType string, data any) error {
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return err
    }
    _, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonBytes))
    return err
}
```

### Optimistic Locking State Transition (from db/instances.go)
```go
// Source: internal/db/instances.go - proven pattern for concurrent safety
func (p *Pool) UpdateInstanceStatus(ctx context.Context, instanceID, fromStatus, toStatus string) (bool, error) {
    tag, err := p.pool.Exec(ctx,
        `UPDATE instances SET status = $1, updated_at = NOW() WHERE instance_id = $2 AND status = $3`,
        toStatus, instanceID, fromStatus,
    )
    if err != nil {
        return false, err
    }
    return tag.RowsAffected() == 1, nil
}
```

### Wiring Background Goroutine in main.go (from existing billing ticker)
```go
// Source: cmd/gpuctl/main.go - proven pattern
go billingTicker.Start(ctx)
slog.Info("billing ticker started")
```

### Provider Registry Iteration (from provision/engine.go)
```go
// Source: internal/provision/engine.go - proven pattern
providers := e.registry.All()
for _, prov := range providers {
    offerings, err := prov.ListAvailable(ctx)
    if err != nil {
        e.logger.Warn("provider availability check failed",
            slog.String("provider", prov.Name()),
            slog.String("error", err.Error()),
        )
        continue
    }
    // process offerings...
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Per-key Redis SCAN (TODO stub) | Single key JSON array | Phase 6 design | O(1) read vs O(N) scan; atomic writes |
| First-match provider selection | Best-price with fallback | Phase 6 upgrade | Better pricing for customers, resilience |
| Per-instance SSE (Phase 4) | Per-org SSE + per-instance (Phase 6) | Phase 6 upgrade | Dashboard can stream all org events |
| No health monitoring (stub) | Provider API polling + event logging | Phase 6 implementation | Spot interruption detection, billing accuracy |

**Deprecated/outdated:**
- The SCAN-based cache pattern in the `internal/availability/cache.go` TODO stub should be replaced with single-key approach. The stubs were written early in the project before requirements were fully understood.
- The `internal/health/monitor.go` TODO references WireGuard handshake checking. The user decided against this -- use provider API polling instead.

## Open Questions

1. **Markup pricing mechanism**
   - What we know: User decided "GPU.ai's markup price only." Provider prices must not be shown to customers.
   - What's unclear: How is markup calculated? Fixed percentage? Per-GPU-type table? Is there a configuration source for markup rates?
   - Recommendation: For Phase 6, use a simple configurable markup percentage (e.g., 15%) applied during cache write. The provisioning engine already stores both `price_per_hour` (retail) and `upstream_price_per_hour` (cost) on instances. Use the same split for availability offerings. The planner should add a `PricingMarkupPct` config field or a `pricing.go` utility.

2. **CPU cores, RAM, storage fields for offerings**
   - What we know: CONTEXT.md requires `cpu_cores`, `ram_gb`, `storage_gb` in the availability response. Current `provider.GPUOffering` does not have these fields.
   - What's unclear: Do providers expose these per-offering? RunPod's `lowestPrice` has `minVcpu` and `minMemory`. Storage is configurable at provision time, not per-offering.
   - Recommendation: Add `CPUCores`, `RAMGB` fields to `provider.GPUOffering`. Populate from provider data where available (RunPod `minVcpu`/`minMemory`). For storage, use a default value (e.g., 40GB matching `defaultContainerDisk`). For `avg_uptime_pct`, use a static value (e.g., 99.5% for on-demand, 95.0% for spot) since we don't have historical data yet.

3. **Provider priority for tiebreaking**
   - What we know: CONTEXT.md says "higher-margin provider wins (existing priority list handles this -- no new mechanism)."
   - What's unclear: There is no explicit priority list in the codebase. The registry iterates providers in map order (effectively random).
   - Recommendation: For same-price offerings, tiebreak by margin (our_price - upstream_price). Higher margin = preferred. This naturally emerges from sorting by upstream price when retail prices are equal. No explicit priority list needed.

4. **RunPod spot interruption specifics**
   - What we know: RunPod returns `desiredStatus: "EXITED"` when a spot pod is reclaimed. Spot instances get 5s SIGTERM before SIGKILL.
   - What's unclear: Whether RunPod's API distinguishes between user-initiated stop and spot interruption in the `desiredStatus` or any other field.
   - Recommendation: Detect interruption by inference: if our DB says instance is "running" and tier is "spot" and provider says "terminated"/"exited", and we didn't initiate the termination, it's a spot interruption. This is the standard pattern for cloud providers that don't provide explicit interruption signals.

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/availability/poller.go`, `internal/availability/cache.go` -- TODO stubs with design hints
- Existing codebase: `internal/health/monitor.go` -- TODO stub with design hints
- Existing codebase: `internal/provision/engine.go` -- `selectProvider` method with Phase 6 comment
- Existing codebase: `internal/api/sse.go` -- proven SSE broker pattern
- Existing codebase: `internal/billing/ticker.go` -- proven background goroutine pattern
- Existing codebase: `internal/db/instances.go` -- proven optimistic locking pattern
- [go-redis documentation](https://pkg.go.dev/github.com/redis/go-redis/v9) -- SET/GET with TTL
- [Redis SCAN documentation](https://redis.io/docs/latest/commands/scan/) -- performance characteristics

### Secondary (MEDIUM confidence)
- [RunPod Manage Pods documentation](https://docs.runpod.io/sdks/graphql/manage-pods) -- pod status values: CREATED, RUNNING, EXITED
- [RunPod GraphQL API Spec](https://graphql-spec.runpod.io/) -- Pod type fields including desiredStatus, podType, runtime
- [RunPod Spot vs On-Demand](https://www.runpod.io/blog/spot-vs-on-demand-instances-runpod) -- 5s SIGTERM on spot interruption
- [Redis SCAN performance](https://redis.io/blog/faster-keys-and-scan-optimized/) -- batch sizes, pattern matching overhead

### Tertiary (LOW confidence)
- RunPod spot interruption specifics -- exact API behavior when spot instance is reclaimed is not explicitly documented. Inferred from desiredStatus="EXITED" and skypilot community discussions.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries already in go.mod, patterns proven in codebase
- Architecture: HIGH - all patterns mirror existing codebase patterns (billing ticker, SSE broker, optimistic locking)
- Pitfalls: HIGH - derived from actual codebase review and concurrency analysis
- RunPod interruption detection: MEDIUM - API behavior inferred from docs + community, not empirically verified
- Markup pricing: LOW - no existing mechanism, requires design decision

**Research date:** 2026-02-25
**Valid until:** 2026-03-25 (stable domain, no fast-moving dependencies)
