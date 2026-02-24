# Phase 2: Provider Abstraction + RunPod Adapter - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Clean provider interface that any GPU cloud can implement, plus a working RunPod adapter that can list GPUs, provision pods, check status, and terminate. Also applies schema improvements (self-documenting PKs, constraints, security fixes) via a new migration before instance CRUD code is written in Phase 4.

</domain>

<decisions>
## Implementation Decisions

### GPU type normalization
- GPU.ai defines canonical GPU type IDs (e.g. `rtx_4090`, `a100_80gb`, `h100_80gb`)
- Each adapter maps provider-specific names to GPU.ai canonical set
- Customers always see consistent names regardless of provider
- GPU types are pulled dynamically from RunPod's API, not hardcoded — adapter normalizes whatever RunPod offers via a name mapping table
- If an unmapped type appears, log a warning and skip or pass through a sanitized version

### Region normalization
- GPU.ai defines its own region codes (e.g. `us-east`, `us-west`, `eu-west`)
- Adapters map provider regions into GPU.ai regions
- Users also see the specific datacenter location (e.g. "San Jose, CA" under `us-west`) so they can drill into exact placement before provisioning

### Pricing tiers
- Two tiers only for v1: **Spot** and **On-Demand**
- Spot: passed through from RunPod Community Cloud / spot. Interruptible. Priced at or slightly above RunPod's spot rate. India (E2E) may not have spot equivalent — US-only initially.
- On-Demand: passed through from RunPod Secure Cloud / E2E on-demand. Non-interruptible, per-second billing. This is the core product.
- No savings plans — creates liability (guaranteeing capacity on someone else's infra, risk of upstream price changes mid-commitment). Defer until own hardware or locked reseller agreement.
- No reserved/supercluster tier — that's a sales motion ("Contact Us"), not a self-serve tier. Handled manually with custom pricing.

### Provisioning model
- Async fire-and-return: Provision returns immediately with an upstream ID
- Status polling happens separately via GetStatus
- Matches how cloud providers actually work (provisioning takes 30s-5min)

### Error handling & retries
- Adapter retries transient errors (5xx, timeouts) internally — 3 attempts with backoff before surfacing error to caller
- Typed `ErrNoCapacity` error for "GPU unavailable" (distinct from API failures) so provisioning engine can try another provider
- Respect RunPod's `Retry-After` header on 429 rate limits, with backoff
- Per-operation default timeouts (list: 10s, provision: 30s, status: 10s, terminate: 30s), overridable via context deadline

### RunPod specifics
- Target RunPod GPU Cloud **Pods only** (Community Cloud + Secure Cloud)
- Not Serverless endpoints (different product for inference, not SSH-in-and-work)
- All GPU types supported dynamically — pull from RunPod API, normalize names
- Reference RunPod API docs for interface contract details during research/planning

### Provider extensibility
- One package per provider: `internal/provider/runpod/`, `internal/provider/e2e/`, etc.
- Each package is self-contained with its own API client code
- Provider registry wires adapters at startup from config
- Interface defined at `internal/provider/` package level, implementations in sub-packages

### Claude's Discretion
- GraphQL client approach for RunPod (raw HTTP vs graphql library)
- Exact GPU name mapping table contents
- Internal struct design for normalized GPU offerings
- Provider registry implementation details

</decisions>

<specifics>
## Specific Ideas

- Reference RunPod API documentation as source of truth for all RunPod-specific decisions
- GPU type mapping should be a simple map that grows — not a complex taxonomy
- Region detail (datacenter location) is a nice product touch that differentiates from competitors showing just "US"

</specifics>

<deferred>
## Deferred Ideas

- RunPod Serverless endpoint support — potential v2 feature for inference-focused users
- Savings plans / reserved tier — requires own hardware or locked reseller agreement
- E2E Networks adapter — v2 milestone (INDIA-01, INDIA-02, INDIA-03)
- Availability polling (Phase 6) will reduce ErrNoCapacity frequency

</deferred>

---

*Phase: 02-provider-abstraction-runpod-adapter*
*Context gathered: 2026-02-24*
