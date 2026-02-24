# Phase 2: Provider Abstraction + RunPod Adapter - Research

**Researched:** 2026-02-24
**Domain:** Provider interface design, RunPod GraphQL API integration, PostgreSQL schema migration
**Confidence:** HIGH

## Summary

Phase 2 has two distinct workstreams: (1) a database schema migration to improve the v0 schema with self-documenting PKs, constraints, and security fixes, and (2) the provider abstraction layer with a working RunPod adapter. The schema work is straightforward ALTER TABLE operations against a greenfield database with no production data. The provider work builds on existing scaffold code (interface, types, and stub adapter already in `internal/provider/`) and requires implementing a GraphQL client for RunPod's Pod API.

RunPod's GraphQL API at `https://api.runpod.io/graphql` uses Bearer token authentication and provides well-documented queries for GPU availability (`gpuTypes` with `lowestPrice`), pod creation (`podFindAndDeployOnDemand` for on-demand, `podRentInterruptable` for spot), pod status (`pod` query), and pod termination (`podTerminate` mutation). Their official Go CLI (`runpodctl`) uses raw `net/http` with JSON-marshaled GraphQL payloads -- no third-party GraphQL library -- which aligns perfectly with the project's "stdlib + minimal deps" convention.

**Primary recommendation:** Use raw `net/http` for the RunPod GraphQL client (matching the project convention and RunPod's own Go CLI pattern), implement the adapter with retry/backoff for transient errors, and write the schema migration as a single SQL file that applies all v1 improvements in one transaction.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- GPU.ai defines canonical GPU type IDs (e.g. `rtx_4090`, `a100_80gb`, `h100_80gb`); each adapter maps provider-specific names to GPU.ai canonical set
- GPU types are pulled dynamically from RunPod's API, not hardcoded -- adapter normalizes whatever RunPod offers via a name mapping table
- If an unmapped type appears, log a warning and skip or pass through a sanitized version
- GPU.ai defines its own region codes (e.g. `us-east`, `us-west`, `eu-west`); adapters map provider regions into GPU.ai regions
- Users also see the specific datacenter location (e.g. "San Jose, CA" under `us-west`)
- Two tiers only for v1: Spot and On-Demand (no savings plans, no reserved tier)
- Spot: passed through from RunPod Community Cloud / spot. Interruptible. US-only initially.
- On-Demand: passed through from RunPod Secure Cloud. Non-interruptible, per-second billing.
- Async fire-and-return provisioning: Provision returns immediately with upstream ID; status polling via GetStatus
- Adapter retries transient errors (5xx, timeouts) internally -- 3 attempts with backoff before surfacing error
- Typed `ErrNoCapacity` error for "GPU unavailable" (distinct from API failures)
- Respect RunPod's `Retry-After` header on 429 rate limits, with backoff
- Per-operation default timeouts (list: 10s, provision: 30s, status: 10s, terminate: 30s), overridable via context deadline
- Target RunPod GPU Cloud Pods only (not Serverless endpoints)
- One package per provider: `internal/provider/runpod/`, `internal/provider/e2e/`, etc.
- Each package is self-contained with its own API client code
- Provider registry wires adapters at startup from config
- Interface defined at `internal/provider/` package level, implementations in sub-packages

### Claude's Discretion
- GraphQL client approach for RunPod (raw HTTP vs graphql library)
- Exact GPU name mapping table contents
- Internal struct design for normalized GPU offerings
- Provider registry implementation details

### Deferred Ideas (OUT OF SCOPE)
- RunPod Serverless endpoint support -- potential v2 feature
- Savings plans / reserved tier -- requires own hardware or locked reseller agreement
- E2E Networks adapter -- v2 milestone (INDIA-01, INDIA-02, INDIA-03)
- Availability polling (Phase 6) will reduce ErrNoCapacity frequency
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| SCHEMA-01 | Rename all primary keys to self-documenting `{table}_id` format and update foreign key references | ALTER TABLE RENAME COLUMN + CASCADE FK updates; PostgreSQL auto-updates FK references on rename |
| SCHEMA-02 | Add NOT NULL constraints on mandatory FKs, explicit ON DELETE, CHECK on instances.status, UNIQUE on instances.hostname, composite unique index on (upstream_provider, upstream_id) | ALTER TABLE ADD CONSTRAINT + SET NOT NULL; all standard PostgreSQL DDL |
| SCHEMA-03 | Remove wg_private_key_enc column (security liability) | ALTER TABLE DROP COLUMN; straightforward since no production data |
| SCHEMA-04 | Add internal_token column to instances for per-instance callback auth, add updated_at column | ALTER TABLE ADD COLUMN; trigger function for updated_at auto-update |
| PROV-01 | Provider interface defines standard contract (Name, ListAvailable, Provision, GetStatus, Terminate) | Interface already exists in `internal/provider/provider.go`; needs review against CONTEXT.md decisions (timeouts, error types) |
| PROV-02 | Provider registry manages multiple adapters with lookup by name | Registry pattern from ARCHITECTURE.md; implement in `internal/provider/registry.go` |
| PROV-03 | RunPod adapter lists available GPU types with pricing via GraphQL API | RunPod `gpuTypes` query with `lowestPrice` subquery; needs GPU name mapping table + region mapping |
| PROV-04 | RunPod adapter provisions pod with custom Docker image and startup scripts | RunPod `podFindAndDeployOnDemand` (on-demand) / `podRentInterruptable` (spot) mutations via GraphQL |
| PROV-05 | RunPod adapter queries pod status by upstream ID | RunPod `pod(input: {podId: ...})` query; map `desiredStatus` to GPU.ai status enum |
| PROV-06 | RunPod adapter terminates pod by upstream ID | RunPod `podTerminate(input: {podId: ...})` mutation; returns void |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `net/http` | Go 1.24 | HTTP client for RunPod GraphQL API | Project convention (no frameworks); RunPod's own Go CLI uses raw `net/http` |
| Go stdlib `encoding/json` | Go 1.24 | JSON marshal/unmarshal for GraphQL payloads | Part of stdlib; GraphQL over HTTP is just JSON POST |
| Go stdlib `text/template` | Go 1.24 | GraphQL query templates (optional) | Cleaner than string concatenation for complex queries |
| `pgx/v5` | v5.8.0 | Database migration execution | Already in project; migration runner uses `psycopg2` but schema SQL is the same |
| `log/slog` | Go 1.24 | Structured logging | Already established in Phase 1 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `time` | Go 1.24 | Timeouts, backoff calculations | Per-operation timeouts (10s/30s), retry backoff |
| Go stdlib `errors` | Go 1.24 | Typed error values (ErrNoCapacity) | Sentinel errors for provider-specific failure modes |
| Go stdlib `sync` | Go 1.24 | Registry thread safety (RWMutex) | Registry may be read concurrently by availability poller |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Raw `net/http` for GraphQL | `machinebox/graphql` | Adds dependency for ~100 lines of code savings; not worth it given project convention of minimal deps |
| Raw `net/http` for GraphQL | `hasura/go-graphql-client` | Heavier dependency with struct-tag-based query building; overkill for 5 fixed queries |
| String literal queries | Code-generated GraphQL client | Only worth it with large schema; 5 queries don't justify generation tooling |

**Decision (Claude's Discretion): Use raw `net/http` for RunPod GraphQL.** Rationale: The project mandates stdlib + minimal deps. RunPod's own Go CLI (`runpodctl`) uses the same pattern -- a simple `Query()` function that JSON-marshals `{query, variables}` and POSTs to the GraphQL endpoint. The adapter only needs 5 fixed queries/mutations, making a library unnecessary.

## Architecture Patterns

### Recommended Project Structure
```
internal/provider/
├── provider.go              # Provider interface (already exists, needs refinement)
├── types.go                 # GPUOffering, ProvisionRequest, etc. (already exists, needs updates)
├── registry.go              # Provider registry (new file)
├── errors.go                # Sentinel errors: ErrNoCapacity, ErrProviderUnavailable (new file)
└── runpod/
    ├── adapter.go           # RunPod adapter implementing Provider interface
    ├── client.go            # RunPod GraphQL HTTP client (queries, mutations, auth)
    ├── mapping.go           # GPU type name mapping + region mapping tables
    └── adapter_test.go      # Unit tests with HTTP test server mocking RunPod API
```

### Pattern 1: GraphQL Client with Raw HTTP
**What:** Simple struct wrapping `*http.Client` with methods for each RunPod API operation
**When to use:** When consuming a GraphQL API with a small number of fixed queries

```go
// internal/provider/runpod/client.go
type Client struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

type graphQLRequest struct {
    Query     string                 `json:"query"`
    Variables map[string]interface{} `json:"variables"`
}

type graphQLResponse struct {
    Data   json.RawMessage  `json:"data"`
    Errors []graphQLError   `json:"errors"`
}

type graphQLError struct {
    Message string `json:"message"`
}

func (c *Client) do(ctx context.Context, req graphQLRequest, result interface{}) error {
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

    resp, err := c.httpClient.Do(httpReq)
    // ... handle response, check for GraphQL errors, unmarshal into result
}
```

### Pattern 2: Retry with Exponential Backoff
**What:** Internal retry logic for transient errors with configurable attempts and backoff
**When to use:** All RunPod API calls (per user decision: 3 attempts with backoff)

```go
func (a *Adapter) withRetry(ctx context.Context, op string, fn func(ctx context.Context) error) error {
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        lastErr = fn(ctx)
        if lastErr == nil {
            return nil
        }
        if !isRetryable(lastErr) {
            return lastErr
        }
        delay := time.Duration(1<<uint(attempt)) * time.Second
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
    }
    return fmt.Errorf("%s failed after 3 attempts: %w", op, lastErr)
}
```

### Pattern 3: GPU Name Mapping Table
**What:** Simple `map[string]GPUType` translating RunPod GPU IDs to GPU.ai canonical names
**When to use:** Every call to ListAvailable and Provision

```go
// internal/provider/runpod/mapping.go
var gpuNameMap = map[string]provider.GPUType{
    "NVIDIA GeForce RTX 4090":  provider.GPUTypeRTX4090,
    "NVIDIA RTX A6000":         provider.GPUTypeRTXA6000,
    "NVIDIA A40":               provider.GPUTypeA40,
    "NVIDIA L40S":              provider.GPUTypeL40S,
    "NVIDIA L4":                provider.GPUTypeL4,
    "NVIDIA A100 80GB PCIe":    provider.GPUTypeA10080GB,
    "NVIDIA A100-SXM4-80GB":    provider.GPUTypeA10080GB,
    "NVIDIA H100 80GB HBM3":    provider.GPUTypeH100SXM,
    "NVIDIA H100 PCIe":         provider.GPUTypeH100PCIE,
    "NVIDIA H200":              provider.GPUTypeH200SXM,
    // More mappings added as RunPod fleet expands
}

var regionMap = map[string]string{
    // RunPod location string -> GPU.ai region code
    // Populated from RunPod Machine.Location field
}
```

### Pattern 4: Provider Registry with Thread Safety
**What:** Map-based registry with `sync.RWMutex` for concurrent read access
**When to use:** Startup registration + concurrent reads from availability poller and API handlers

```go
// internal/provider/registry.go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *Registry) Register(p Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.providers[name]
    return p, ok
}

func (r *Registry) All() []Provider {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Provider, 0, len(r.providers))
    for _, p := range r.providers {
        out = append(out, p)
    }
    return out
}
```

### Pattern 5: Typed Error Values
**What:** Sentinel errors that callers can check with `errors.Is()`
**When to use:** Distinguishing capacity failures from API failures

```go
// internal/provider/errors.go
var (
    ErrNoCapacity         = errors.New("no GPU capacity available")
    ErrProviderUnavailable = errors.New("provider API unavailable")
    ErrInvalidGPUType     = errors.New("unsupported GPU type")
)
```

### Anti-Patterns to Avoid
- **Hardcoding GPU types in the adapter:** GPU types should come from the API dynamically, then be mapped via the mapping table. Never maintain a static list of what RunPod offers.
- **Synchronous provisioning:** Provision MUST return immediately with an upstream ID. Never block waiting for the pod to be ready.
- **Leaking RunPod types across the interface boundary:** The adapter should translate all RunPod-specific types into `provider.*` types. No RunPod structs should appear outside `internal/provider/runpod/`.
- **Global HTTP client:** Each adapter should own its `*http.Client` with appropriate timeouts. Don't share a global client across adapters.
- **Ignoring GraphQL errors field:** RunPod returns HTTP 200 even for GraphQL errors. Always check the `errors` array in the response JSON, not just the HTTP status code.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| GraphQL request/response cycle | Full GraphQL parser | Raw HTTP POST with JSON marshal | Only 5 fixed queries; a parser adds complexity for no benefit |
| HTTP retry with backoff | Custom retry from scratch | Adapt existing `db.ConnectWithRetry` pattern | Phase 1 established this pattern; reuse the approach (not the function itself, as adapter needs per-request retry) |
| UUID generation for `internal_token` | Custom random string | `crypto/rand` + `encoding/hex` or PostgreSQL `gen_random_uuid()` | Cryptographically secure tokens; don't use `math/rand` |
| Schema migration ordering | Custom version tracker | Existing `tools/migrate.py` with `schema_migrations` table | Already handles version tracking, ordering, and per-migration transactions |

**Key insight:** The RunPod GraphQL API is simple enough that a full GraphQL client library would add more complexity than it removes. Five static queries with variable substitution is a perfect fit for raw HTTP.

## Common Pitfalls

### Pitfall 1: GraphQL Errors with HTTP 200
**What goes wrong:** RunPod returns HTTP 200 OK even when the GraphQL query has errors (auth failures, invalid input, etc.). Code that only checks HTTP status misses these errors.
**Why it happens:** GraphQL spec allows errors in the response body alongside HTTP 200.
**How to avoid:** Always unmarshal the full response and check the `errors` array before processing `data`. Return an error if `len(response.Errors) > 0`.
**Warning signs:** Operations appearing to succeed but returning empty/nil data.

### Pitfall 2: RunPod GPU ID Mismatch
**What goes wrong:** RunPod GPU IDs are full names like `"NVIDIA H100 80GB HBM3"` and `"NVIDIA A100-SXM4-80GB"`, not short codes. The same logical GPU may have multiple IDs (e.g., A100 80GB has both PCIe and SXM4 variants with different IDs).
**Why it happens:** RunPod uses manufacturer marketing names as API identifiers.
**How to avoid:** Build the mapping table from actual RunPod `gpuTypes` query output, not assumptions. Log unmapped GPU types at WARN level. Handle both SXM and PCIe variants of the same GPU family.
**Warning signs:** ListAvailable returning empty results when GPUs should be available.

### Pitfall 3: Spot vs On-Demand Mutation Mismatch
**What goes wrong:** Using `podFindAndDeployOnDemand` for spot instances or `podRentInterruptable` for on-demand. The mutations are different and have different parameters (spot requires `bidPerGpu`).
**Why it happens:** RunPod has separate mutations for each pricing tier.
**How to avoid:** Branch on `InstanceTier` in the Provision method. On-demand uses `podFindAndDeployOnDemand` with `cloudType: SECURE`. Spot uses `podRentInterruptable` with a bid price.
**Warning signs:** Spot pods not being interruptible, or on-demand pods unexpectedly being reclaimed.

### Pitfall 4: Schema Migration on Existing v0 Data
**What goes wrong:** ALTER TABLE RENAME COLUMN on PKs can break FK references if not handled correctly.
**Why it happens:** Renaming a PK column requires all FK references to also update.
**How to avoid:** PostgreSQL automatically updates FK references when a column is renamed. However, indexes need to be explicitly dropped and recreated with new names. Run the entire migration in a single transaction so it's atomic. Since this is greenfield with no production data, the migration can also be written as DROP TABLE + CREATE TABLE if preferred (simpler, but loses migrate-forward discipline).
**Warning signs:** Foreign key constraint errors after migration.

### Pitfall 5: RunPod Docker vs Cloud-Init Model
**What goes wrong:** Assuming RunPod supports cloud-init style initialization. RunPod uses Docker images + startup scripts, not cloud-init.
**Why it happens:** The architecture document describes a cloud-init `bootstrap.sh`, but RunPod pods are Docker containers.
**How to avoid:** For RunPod, the "initialization" happens via: (1) a custom Docker image (`imageName` parameter), (2) startup commands in `dockerArgs`, and (3) environment variables (`env` array). The WireGuard/SSH setup must be baked into the Docker image or executed via startup commands, not cloud-init. This is a known concern flagged in STATE.md: "RunPod Docker initialization requires complete redesign from cloud-init."
**Warning signs:** Pods starting but missing WireGuard tunnel, SSH keys, or hostname configuration.

### Pitfall 6: Missing `updated_at` Trigger
**What goes wrong:** Adding an `updated_at` column without a trigger means it never gets updated automatically.
**Why it happens:** PostgreSQL doesn't have a built-in "on update" feature like MySQL's `ON UPDATE CURRENT_TIMESTAMP`.
**How to avoid:** Create a trigger function that sets `updated_at = NOW()` on every UPDATE, then attach it to the instances table.
**Warning signs:** `updated_at` always equaling `created_at`.

## Code Examples

### RunPod GraphQL: List GPU Types with Availability
```go
// Source: RunPod official docs (https://docs.runpod.io/sdks/graphql/manage-pods)
// + RunPod CLI source (https://github.com/runpod/runpodctl/blob/main/api/cloud.go)

const listGPUTypesQuery = `
query GpuTypes {
  gpuTypes {
    id
    displayName
    memoryInGb
    secureCloud
    communityCloud
    securePrice
    communityPrice
    secureSpotPrice
    communitySpotPrice
    lowestPrice(input: { gpuCount: 1 }) {
      minimumBidPrice
      uninterruptablePrice
      minVcpu
      minMemory
      stockStatus
      maxUnreservedGpuCount
    }
  }
}
`
```

### RunPod GraphQL: Create On-Demand Pod
```go
// Source: RunPod official docs + CLI source
const createOnDemandPodMutation = `
mutation createPod($input: PodFindAndDeployOnDemandInput!) {
  podFindAndDeployOnDemand(input: $input) {
    id
    costPerHr
    desiredStatus
    lastStatusChange
    machine {
      gpuDisplayName
      location
    }
  }
}
`

// Variables for on-demand:
// {
//   "input": {
//     "cloudType": "SECURE",
//     "gpuCount": 1,
//     "gpuTypeId": "NVIDIA GeForce RTX 4090",
//     "name": "gpu-xxxx",
//     "imageName": "gpuai/workspace:latest",
//     "containerDiskInGb": 40,
//     "volumeInGb": 0,
//     "startSsh": true,
//     "ports": "22/tcp",
//     "env": [
//       {"key": "SSH_PUBLIC_KEYS", "value": "ssh-rsa ..."},
//       {"key": "GPUAI_INSTANCE_ID", "value": "gpu-xxxx"},
//       {"key": "GPUAI_CALLBACK_URL", "value": "https://api.gpu.ai"},
//       {"key": "GPUAI_INTERNAL_TOKEN", "value": "token-xxx"}
//     ]
//   }
// }
```

### RunPod GraphQL: Create Spot Pod
```go
const createSpotPodMutation = `
mutation createSpotPod($input: PodRentInterruptableInput!) {
  podRentInterruptable(input: $input) {
    id
    costPerHr
    desiredStatus
    lastStatusChange
    machine {
      gpuDisplayName
      location
    }
  }
}
`

// Variables include bidPerGpu and cloudType: "COMMUNITY"
```

### RunPod GraphQL: Get Pod Status
```go
const getPodQuery = `
query Pod($input: PodFilter!) {
  pod(input: $input) {
    id
    desiredStatus
    lastStatusChange
    costPerHr
    runtime {
      uptimeInSeconds
      ports {
        ip
        isIpPublic
        privatePort
        publicPort
        type
      }
    }
    machine {
      gpuDisplayName
      location
    }
  }
}
`

// Variables: { "input": { "podId": "upstream-id-here" } }
```

### RunPod GraphQL: Terminate Pod
```go
const terminatePodMutation = `
mutation terminatePod($input: PodTerminateInput!) {
  podTerminate(input: $input)
}
`

// Variables: { "input": { "podId": "upstream-id-here" } }
// Note: podTerminate returns void (null)
```

### Schema Migration: v1 Improvements
```sql
-- Example: Rename PK columns to self-documenting format
ALTER TABLE organizations RENAME COLUMN id TO organization_id;
ALTER TABLE users RENAME COLUMN id TO user_id;
ALTER TABLE ssh_keys RENAME COLUMN id TO ssh_key_id;
ALTER TABLE instances RENAME COLUMN id TO instance_id;
ALTER TABLE environments RENAME COLUMN id TO environment_id;
ALTER TABLE usage_records RENAME COLUMN id TO usage_record_id;
-- PostgreSQL auto-updates FK references when source column is renamed

-- Add NOT NULL on mandatory FKs
ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE instances ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE instances ALTER COLUMN user_id SET NOT NULL;

-- Add ON DELETE behavior
ALTER TABLE users DROP CONSTRAINT users_org_id_fkey;
ALTER TABLE users ADD CONSTRAINT users_org_id_fkey
    FOREIGN KEY (org_id) REFERENCES organizations(organization_id) ON DELETE CASCADE;

-- CHECK constraint on status
ALTER TABLE instances ADD CONSTRAINT chk_instances_status
    CHECK (status IN ('creating', 'provisioning', 'booting', 'running', 'stopping', 'terminated', 'error'));

-- UNIQUE on hostname
ALTER TABLE instances ADD CONSTRAINT uq_instances_hostname UNIQUE (hostname);

-- Composite unique index
CREATE UNIQUE INDEX idx_instances_upstream
    ON instances(upstream_provider, upstream_id);

-- Drop security liability column
ALTER TABLE instances DROP COLUMN wg_private_key_enc;

-- Add new columns
ALTER TABLE instances ADD COLUMN internal_token VARCHAR(255);
ALTER TABLE instances ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_instances_updated_at
    BEFORE UPDATE ON instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

### RunPod Status Mapping
```go
// RunPod desiredStatus values -> GPU.ai status
var statusMap = map[string]string{
    "CREATED":  "creating",
    "RUNNING":  "running",
    "EXITED":   "terminated",
    // RunPod doesn't have a direct "provisioning" or "booting" --
    // CREATED covers the pre-running state
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| RunPod GraphQL-only API | RunPod REST API (beta) alongside GraphQL | 2025 | REST API offers similar functionality but GraphQL remains primary/documented; use GraphQL for now |
| API key in query parameter | Bearer token in Authorization header | 2025 | RunPod CLI now uses `Authorization: Bearer <key>`; more secure (not logged in URLs) |
| Fixed GPU type lists | Dynamic `gpuTypes` query with `lowestPrice` | Ongoing | GPU fleet changes regularly (added B200, RTX 5090, etc.); must query dynamically |

**Deprecated/outdated:**
- Passing API key as `?api_key=` query parameter: Still works but Bearer header is preferred
- `TierReserved` in current types.go: User decision says no reserved tier for v1; remove to avoid confusion

## Open Questions

1. **RunPod region data granularity**
   - What we know: RunPod's `Machine.Location` field provides datacenter location (e.g., "US-TX-3" or city names). The `lowestPrice` query has a `dataCenterId` filter.
   - What's unclear: The exact format of location strings and whether they're stable identifiers or display names. Also unclear how many distinct regions RunPod actually serves.
   - Recommendation: During implementation, query `gpuTypes` and log all unique `Machine.Location` values encountered. Build the region mapping table from observed data. Start with broad mappings (`US-*` -> `us-east`/`us-west` based on state) and refine.

2. **Spot bid pricing strategy**
   - What we know: `podRentInterruptable` requires a `bidPerGpu` parameter. `lowestPrice` returns `minimumBidPrice`.
   - What's unclear: Whether setting `bidPerGpu` to `minimumBidPrice` guarantees allocation, or if we should bid slightly higher.
   - Recommendation: Use `minimumBidPrice` from the `lowestPrice` query as the bid. If allocation fails, this is an `ErrNoCapacity` -- the availability poller (Phase 6) will reflect current spot pricing.

3. **RunPod rate limits for GraphQL**
   - What we know: RunPod returns 429 with `Retry-After` header. Serverless endpoints have per-endpoint rate limits with 10-second windows.
   - What's unclear: Exact rate limits for the Pod management GraphQL API (distinct from Serverless API).
   - Recommendation: Implement `Retry-After` header respect as decided by user. If 429 is received, backoff per the header. For the availability poller (Phase 6), a 30-second polling interval with a single request per cycle should be well within limits.

4. **GPUOffering struct -- datacenter location field**
   - What we know: User wants datacenter location visible (e.g., "San Jose, CA" under `us-west`).
   - What's unclear: The current `GPUOffering` struct has `Region string` but no `DatacenterLocation` field.
   - Recommendation: Add a `DatacenterLocation string` field to `GPUOffering` in `types.go`. Populated from RunPod's `Machine.Location` during ListAvailable, displayed to users alongside the GPU.ai region code.

## Sources

### Primary (HIGH confidence)
- [RunPod GraphQL Manage Pods Documentation](https://docs.runpod.io/sdks/graphql/manage-pods) - Pod CRUD mutations, query syntax, parameter types
- [RunPod GraphQL API Spec](https://graphql-spec.runpod.io/) - Full schema: GpuType fields, Pod fields, enum values, pricing fields
- [RunPod GPU Types Reference](https://docs.runpod.io/references/gpu-types) - Complete GPU ID mapping (41 GPU models with exact API IDs and VRAM)
- [RunPod Go CLI source: api/pod.go](https://github.com/runpod/runpodctl/blob/main/api/pod.go) - Official Go patterns for GraphQL queries, pod structs, variable construction
- [RunPod Go CLI source: api/query.go](https://github.com/runpod/runpodctl/blob/main/api/query.go) - Raw HTTP GraphQL client pattern with Bearer auth
- [RunPod Go CLI source: api/cloud.go](https://github.com/runpod/runpodctl/blob/main/api/cloud.go) - GPU availability query pattern with lowestPrice

### Secondary (MEDIUM confidence)
- [RunPod Pod Templates Overview](https://docs.runpod.io/pods/templates/overview) - Docker image + startup command model (not cloud-init)
- [RunPod REST API Blog Post](https://www.runpod.io/blog/runpod-rest-api-gpu-management) - REST API as future alternative to GraphQL
- [RunPod API Keys Documentation](https://docs.runpod.io/get-started/api-keys) - Bearer token authentication method

### Tertiary (LOW confidence)
- RunPod rate limits: Specific GraphQL rate limits not documented; 429 + Retry-After confirmed for serverless but assumed for pods based on HTTP standards

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using stdlib `net/http` is verified by project convention and RunPod CLI source code
- Architecture: HIGH - Interface/registry pattern already scaffolded in codebase; RunPod API thoroughly documented
- Pitfalls: HIGH - GraphQL-over-HTTP errors, Docker vs cloud-init model, GPU ID formats all verified from official sources
- Schema migration: HIGH - Standard PostgreSQL DDL; greenfield database with no production data risk

**Research date:** 2026-02-24
**Valid until:** 2026-03-24 (RunPod API is stable; GPU type IDs may expand but won't change)
