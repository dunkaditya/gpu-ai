# Pitfalls Research

**Domain:** GPU cloud aggregation platform (re-renting upstream GPUs behind a WireGuard privacy layer with per-second Stripe billing)
**Researched:** 2026-02-24
**Confidence:** HIGH (verified across official docs, community reports, and multiple sources)

## Critical Pitfalls

### Pitfall 1: RunPod Uses Docker Containers, Not VMs -- Cloud-Init Will Not Work

**What goes wrong:**
The architecture doc (`docs/ARCHITECTURE.md`) designs the privacy layer around a traditional cloud-init boot script (`infra/cloud-init/bootstrap.sh`) that runs `apt-get install wireguard`, configures `systemctl`, sets iptables rules, and manages the full OS. RunPod pods are Docker containers, not VMs. There is no cloud-init support. The `apt-get`, `systemctl`, and `hostnamectl` commands in the bootstrap script will either fail outright or behave unpredictably in a container context.

**Why it happens:**
The architecture was designed assuming upstream providers give you bare-metal or VM instances where you control the OS boot process. RunPod's "pods" are containerized environments with a specific initialization model: a `start.sh` entrypoint that executes `pre_start.sh` and `post_start.sh` hook scripts, not cloud-init userdata.

**How to avoid:**
Completely redesign the instance initialization strategy for RunPod:
1. Build a custom Docker image that bakes in WireGuard tooling (wireguard-go or wireguard-tools), SSH server, and the GPU.ai initialization logic.
2. Use RunPod's `dockerStartCmd` or `pre_start.sh` hook script to configure WireGuard tunnel parameters at pod boot using environment variables injected via the RunPod API's `env` array.
3. Use the RunPod template/pod creation GraphQL mutations (`podFindAndDeployOnDemand`, `podRentInterruptable`) to pass instance-specific config (WireGuard keys, proxy endpoint, SSH keys) as environment variables.
4. Abstract the init mechanism behind the Provider interface so future VM-based providers (E2E Networks) can use actual cloud-init while RunPod uses its Docker-native approach.

**Warning signs:**
- Bootstrap script references `systemctl`, `hostnamectl`, `apt-get install` -- none of these work reliably in Docker containers
- No testing of the cloud-init script against actual RunPod pod creation
- Provider adapter treats all providers identically for instance initialization

**Phase to address:**
Phase 2 (RunPod adapter) and Phase 3 (WireGuard privacy layer) -- this must be resolved before any end-to-end provisioning works.

---

### Pitfall 2: WireGuard Requires NET_ADMIN Capability Inside RunPod Containers

**What goes wrong:**
WireGuard (even the userspace `wireguard-go` implementation) requires `NET_ADMIN` capability to create tunnel interfaces and manipulate routing tables. Standard Docker containers drop this capability. If RunPod pods do not grant `NET_ADMIN`, the entire privacy layer is impossible inside the upstream instance.

**Why it happens:**
Cloud providers restrict container capabilities for security isolation. WireGuard needs to create network interfaces (`wg0`), modify iptables rules, and manipulate routing -- all of which require elevated privileges that multi-tenant container platforms typically deny.

**How to avoid:**
1. **Verify RunPod capabilities first.** Before writing any WireGuard code, create a test RunPod pod and attempt to run `wireguard-go` or `wg-quick` to confirm `NET_ADMIN` is available. RunPod runs containers as root by default and appears to grant elevated capabilities, but this must be verified empirically.
2. **Use `wireguard-go` (userspace) instead of kernel WireGuard.** Even if the kernel module is loaded on the host, you cannot `modprobe wireguard` from inside a container. The userspace Go implementation only needs `/dev/net/tun` and `NET_ADMIN` -- no kernel module loading.
3. **Fallback plan: reverse tunnel.** If WireGuard cannot run inside the pod, use an SSH reverse tunnel from the pod back to the GPU.ai proxy server. This is slower and less elegant but requires zero special capabilities.

**Warning signs:**
- `wg-quick up wg0` fails with "Operation not permitted" in the container
- `ip link add wg0 type wireguard` fails (kernel module approach will not work)
- No `/dev/net/tun` device available inside the container

**Phase to address:**
Phase 2 (RunPod adapter) -- must be the very first validation before building the WireGuard manager. Build a proof-of-concept pod with WireGuard before writing any Go code.

---

### Pitfall 3: Stripe Meter Events Accept Only Positive Integers -- Per-Second GPU Billing Requires Unit Conversion

**What goes wrong:**
The architecture specifies per-second billing with `price_per_hour` stored as `NUMERIC(10, 4)`. Stripe's Billing Meters API only accepts positive integer values for usage events. You cannot send `0.000589` (the per-second cost of a $2.12/hr GPU). If you naively report 1 event per second with value=1 and set the unit price to $0.000589, Stripe will truncate to $0.00 and the customer gets free compute.

**Why it happens:**
Stripe's metering system was designed for countable events (API calls, messages, seats), not continuous time-based billing at sub-cent granularity. The integer constraint is a fundamental API limitation.

**How to avoid:**
1. **Report GPU-seconds as the unit, price in millicents.** Report `value: 1` per second of usage. Set the Stripe price to the per-second rate in the smallest representable unit. For a $2.12/hr GPU: per-second cost = $0.0005889. Stripe's minimum price unit is cents, so use `$0.06/100 seconds` or similar aggregation.
2. **Aggregate locally, report periodically.** Instead of reporting every second, accumulate usage in PostgreSQL and report to Stripe every 60 seconds (value = 60 GPU-seconds). This reduces API call volume (from 1000/s limit to ~17/s for 1000 concurrent instances) and makes integer math work.
3. **Use a local billing ledger as the source of truth.** Record exact start/stop timestamps in PostgreSQL. Calculate charges with full floating-point precision locally. Use Stripe only for payment collection, not as the billing calculation engine. Reconcile periodically.
4. **Never use Stripe as the metering source of truth.** Stripe meter events process asynchronously and "might not immediately reflect recently received meter events." Your own database must be authoritative.

**Warning signs:**
- Customer invoices show $0.00 for short sessions
- Billing amounts do not match expected `(duration_seconds * price_per_hour / 3600)` calculations
- Stripe usage summaries lag behind actual usage by minutes

**Phase to address:**
Phase 5 (Auth + Billing) -- design the billing data model and Stripe integration together, not separately.

---

### Pitfall 4: Availability Cache Shows GPU as Available, But Provisioning Fails (Stale Cache Race Condition)

**What goes wrong:**
The availability poller runs every 30 seconds and caches results in Redis with a 60-second TTL. A customer sees "12 H100 SXM available" and clicks provision. In the 0-30 seconds since the last poll, another customer (on RunPod directly, or another GPU.ai customer) has taken the last available instance. The provisioning call to RunPod fails, but GPU.ai already told the customer it was available.

**Why it happens:**
GPU availability is a shared, rapidly changing resource. A 30-second polling interval means data is always 0-30 seconds stale. RunPod's inventory is shared across all RunPod customers globally, not just GPU.ai users. Popular GPUs (H100, A100) can go from 12 available to 0 in seconds during peak demand.

**How to avoid:**
1. **Treat availability as a hint, not a guarantee.** The UI should say "likely available" or use a confidence indicator, never "12 available" as a hard number.
2. **Implement graceful provisioning failure.** When RunPod returns an error because the GPU is no longer available, catch it specifically, invalidate the Redis cache for that GPU type, and return a clear "GPU no longer available, try again" response -- not a 500 error.
3. **Invalidate cache on provisioning failure.** When any provision attempt fails due to capacity, immediately delete/update the relevant Redis key so subsequent requests see accurate data.
4. **Consider optimistic locking with retry.** On provision failure, automatically retry with the next-best offering (different region, different tier) if the customer's request allows flexibility.
5. **Fix the TTL mismatch.** The current architecture sets 60-second TTL with 30-second polling. If the poller fails once, stale data persists for a full 60 seconds. Set TTL to 35 seconds (poll interval + small buffer).

**Warning signs:**
- Provisioning failures spike during peak hours (US business hours, ML conference deadlines)
- Customer complaints about "phantom availability" -- seeing GPUs listed but unable to provision
- Redis cache TTL (60s) is longer than poll interval (30s), meaning stale data persists even after a fresh poll fails

**Phase to address:**
Phase 6 (Availability engine) and Phase 7 (API routes) -- the availability engine and provisioning flow must be designed together.

---

### Pitfall 5: Privacy Layer Leaks Upstream Provider Identity Through Multiple Channels

**What goes wrong:**
The architecture hides the upstream provider IP behind WireGuard and strips provider names from API responses. But provider identity can leak through dozens of other channels: DNS reverse lookups on the WireGuard tunnel endpoint, NVIDIA driver version strings unique to provider images, container environment variables set by RunPod (`RUNPOD_*` prefixed), Docker image layer metadata, `/etc/hosts` entries, cloud provider metadata endpoints (169.254.169.254), kernel version fingerprinting, GPU topology output (`nvidia-smi -q`), and timing signatures in network latency.

**Why it happens:**
Privacy is hard because it requires defense in depth. Blocking the obvious leak (IP address) while ignoring metadata, environment, and fingerprinting channels creates a false sense of security. RunPod specifically injects `RUNPOD_*` environment variables into every pod and exposes them via its initialization system.

**How to avoid:**
1. **Scrub RunPod environment variables.** In the startup script, `unset` all `RUNPOD_*` variables before the customer gains SSH access. Do this in `post_start.sh` after RunPod's init completes.
2. **Block cloud metadata endpoints.** The iptables firewall must block outbound requests to `169.254.169.254` (cloud metadata service) from within the WireGuard tunnel.
3. **Sanitize `/etc/hosts`, `/etc/resolv.conf`, MOTD.** Remove any provider-specific entries. Replace with GPU.ai branding.
4. **Custom NVIDIA driver reporting.** While you cannot change `nvidia-smi` output, you can alias or wrap it. More practically, accept that a determined technical user can fingerprint the hardware -- the goal is that casual inspection reveals nothing.
5. **Control DNS resolution.** Route all DNS through the WireGuard tunnel to GPU.ai's DNS server. Block direct DNS queries from the instance to the internet.
6. **Audit checklist.** Before marking a provider integration as complete, run a full privacy audit: SSH in, run `env`, `cat /etc/hosts`, `cat /etc/resolv.conf`, `curl 169.254.169.254`, `nvidia-smi -q`, `uname -a`, `traceroute 8.8.8.8`, and verify nothing reveals the upstream provider.

**Warning signs:**
- Running `env | grep -i runpod` inside an instance returns results
- `curl http://169.254.169.254/` returns cloud metadata
- DNS resolution uses the upstream provider's DNS servers instead of GPU.ai's
- `traceroute` reveals upstream provider's network path before entering the WireGuard tunnel

**Phase to address:**
Phase 3 (WireGuard privacy layer) -- build the privacy audit checklist as an automated test that runs against every new provider integration.

---

### Pitfall 6: Spot Instance Interruption Kills Customer Workload with Only 5-Second Warning

**What goes wrong:**
RunPod spot (Community Cloud interruptible) instances send SIGTERM followed by SIGKILL with only a 5-second grace period. The customer's GPU training job, inference server, or development environment is destroyed. If billing has been active, the customer is charged for time used but loses all unsaved work. WireGuard tunnel drops. The instance record in PostgreSQL shows "running" but the upstream pod no longer exists.

**Why it happens:**
RunPod spot instances are priced 40-70% cheaper because they can be reclaimed when on-demand capacity is needed or when another user outbids. The 5-second warning is dramatically shorter than AWS's 2-minute spot interruption notice.

**How to avoid:**
1. **Health monitoring must detect interruptions within seconds.** The health poller should check instance status every 15-30 seconds. When an instance disappears, immediately update PostgreSQL status to "interrupted" and stop billing.
2. **Expose spot tier risks to customers.** The API and UI must clearly communicate that spot instances can be terminated at any time. Require explicit acknowledgment.
3. **Implement automatic billing stop on interruption.** When health check detects the instance is gone, `billing_end` must be set immediately. Do not wait for the customer to explicitly terminate.
4. **Consider a callback/webhook from the instance.** The startup script should install a SIGTERM handler that calls back to GPU.ai's API before the pod dies: `trap 'curl -X POST https://api.gpu.ai/internal/instances/${ID}/interrupted' SIGTERM`. This gives GPU.ai near-instant notification, but the 5-second window makes this unreliable.
5. **Volume persistence documentation.** Clearly document that RunPod volumes survive spot interruption but compute state does not. Recommend customers use persistent volumes for checkpoints.

**Warning signs:**
- Instances showing "running" in the database but unreachable via WireGuard
- Billing continues for terminated spot instances
- Customer complaints about unexpected termination without notification
- Health check interval is longer than the 5-second SIGTERM window

**Phase to address:**
Phase 2 (RunPod adapter) for spot handling, Phase 5 (Billing) for interruption billing, and Phase 8 (Health monitoring) for detection.

---

### Pitfall 7: Per-Second Billing Drift Between GPU.ai and RunPod Causes Margin Loss

**What goes wrong:**
GPU.ai charges the customer per-second starting from `billing_start` in PostgreSQL. RunPod also charges per-second from the moment the pod starts. If GPU.ai's `billing_start` timestamp is even 30 seconds later than RunPod's actual pod start (waiting for WireGuard tunnel establishment, startup script completion, "ready" callback), GPU.ai absorbs that cost. At $2.12/hr for an H100, 30 seconds of unbilled time costs $0.018 per provision. At 1000 provisions/day, that is $18/day or $6,500/year of margin erosion.

**Why it happens:**
The provisioning flow has multiple asynchronous steps: API call to RunPod -> pod boots (~15s) -> startup script runs -> WireGuard establishes -> ready callback fires -> billing starts. RunPod starts billing at step 1. GPU.ai starts billing at the last step. The gap is real cost.

**How to avoid:**
1. **Start billing at provision request, not at ready callback.** The customer clicked "provision" -- billing starts when the upstream provider starts charging, not when the instance becomes usable. Document this clearly.
2. **Track upstream billing start separately.** Store both `upstream_billing_start` (when RunPod starts charging) and `customer_billing_start` in the database. Reconcile the difference as a known cost of doing business or pass it through.
3. **Minimize boot-to-ready time.** Pre-bake as much as possible into the Docker image to reduce startup script execution time. Every second saved is margin preserved.
4. **Build a reconciliation report.** The `tools/reports/` Python tooling should compare GPU.ai billing records against RunPod's billing API to detect systematic margin leakage.

**Warning signs:**
- `billing_start` is consistently 15-45 seconds after `created_at` for instances
- Margin calculations show less profit than expected at the per-instance level
- No mechanism to query RunPod's actual billing start for comparison

**Phase to address:**
Phase 5 (Billing) -- billing timestamps must be designed alongside the provisioning engine, not as an afterthought.

---

### Pitfall 8: Billing Race Condition on Instance Termination

**What goes wrong:**
Instance terminates but `billing_end` is not recorded atomically. Network failures between the terminate call and database update leave billing running. Customer is charged for GPU time they are not using. Or worse: billing stops but the upstream instance keeps running, consuming cost with no revenue.

**Why it happens:**
Billing start/stop is a distributed transaction across three systems: upstream provider API, PostgreSQL, and Stripe. Any one can fail independently. Developers treat this as a simple "update column" operation.

**How to avoid:**
1. Use a state machine for instance lifecycle: `creating -> running -> stopping -> terminated`. Never skip states.
2. Record `billing_end` in PostgreSQL FIRST, then call upstream terminate. If upstream terminate fails, retry with exponential backoff. An instance with billing stopped but still running costs GPU.ai money but does not overcharge the customer -- this is the safer failure mode.
3. Run a reconciliation job (Python tooling in `tools/`) that compares active upstream instances against PostgreSQL records. Alert on mismatches.
4. Idempotent terminate: calling `DELETE /instances/{id}` multiple times must be safe.

**Warning signs:**
- Instances in "running" state with no upstream equivalent
- Instances in "stopping" state for more than 5 minutes
- Usage records that span implausible durations (more than 30 days continuous)

**Phase to address:**
Phase 4 (Database + Instance Management) for the state machine, Phase 5 (Billing) for the billing logic, Phase 7 (API routes) for idempotent termination.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Storing WireGuard private keys as plaintext in PostgreSQL | Simpler implementation, no encryption layer needed | Key compromise in a DB breach exposes all active tunnels; attacker can impersonate instances | Never -- use `pgcrypto` `pgp_sym_encrypt()` from day one, as the schema already has `wg_private_key_enc` column |
| Single WireGuard proxy server | Simple architecture, one server to manage | Single point of failure; all customer traffic routes through one box; outage = total platform outage | MVP only -- plan for multi-proxy by Phase 2 deployment |
| Polling RunPod every 30s for all GPU types | Simple timer goroutine, minimal code | Wastes API calls for GPU types no one requests; may hit rate limits as provider count grows | MVP -- add demand-based polling later |
| Hardcoding RunPod as the only provider | Faster to build without provider abstraction | Adding E2E Networks requires refactoring if the Provider interface was not properly defined upfront | Never -- the Provider interface is already designed; implement it from the start |
| Using `float64` for prices in Go | Simpler code, no external library | Floating-point arithmetic errors accumulate in billing calculations; `2.12 * 3600` may not equal expected value | Never -- use `shopspring/decimal` or integer cents from day one |
| Using `NUMERIC(10,4)` for per-second price storage | Matches hourly pricing well | Per-second rate for $2.12/hr is $0.000588... which requires 6+ decimal places. 4 decimal places truncates and accumulates rounding errors over hours | Never -- use `NUMERIC(10,6)` or store prices only at hourly granularity and compute per-second in application code |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| RunPod GraphQL API | Assuming `gpuCount` requested will always be fulfilled; creating a pod with `gpuCount=8` can return internal server error even when `maxGpuCountCommunityCloud=8` | Always handle partial fulfillment. Check actual GPU count in the response. Validate against `securePrice` (zero means unavailable). |
| RunPod GraphQL API | Using a single API key for all operations (polling + provisioning) and hitting rate limits | Use separate API keys for availability polling vs. provisioning. Implement exponential backoff. RunPod rate limits are not publicly documented -- build in throttling from day one. |
| RunPod GraphQL API | Assuming cloud-init/userdata support. RunPod pods are Docker containers with `start.sh` -> `pre_start.sh` / `post_start.sh` hook scripts. | Build a custom GPU.ai Docker image. Inject per-instance config via env vars in the `podFindAndDeployOnDemand` mutation's `env` parameter. |
| Stripe Billing Meters | Using legacy Usage Records API (`/v1/subscription_items/{id}/usage_records`) | Use the new Billing Meters API. Legacy API was removed in Stripe API version 2025-03-31. The new API uses `/v2/billing/meter_events`. |
| Stripe Billing Meters | Treating meter events as synchronous confirmation of billing | Meter events process asynchronously. Aggregated usage "might not immediately reflect recently received meter events." Your database is the source of truth, not Stripe. |
| Stripe Billing Meters | Not using idempotency keys for meter events | Always include an idempotency key (e.g., `instance_id:timestamp_bucket`). Network retries without idempotency keys cause double-billing. |
| Stripe Billing Meters | Timestamp drift -- events more than 35 days old are rejected, events more than 5 minutes in the future are rejected | Ensure NTP is configured on the server. Validate timestamps before sending. Buffer events locally if Stripe is temporarily unreachable. |
| Stripe Billing Meters | Sending fractional values for usage -- Stripe only accepts positive integers | Report GPU-seconds as integer counts. Set price per unit at the per-second rate. Aggregate locally and report in batches (e.g., 60 GPU-seconds every minute). |
| Clerk JWT | Caching JWKS indefinitely and never refreshing | Clerk's Go SDK caches JWK for 1 hour by default. If doing manual verification, fetch fresh JWKS when a key ID in the JWT header does not match any cached key. Refresh on Clerk Secret Key rotation. |
| Clerk JWT | Not handling clock skew between servers | Use `clerk.WithLeeway()` to add 30-60 seconds of tolerance for `exp` and `nbf` claim validation. |
| Redis (availability cache) | Setting TTL longer than poll interval, creating stale-after-refresh windows | Set Redis key TTL to 35 seconds (poll interval + 5s buffer), not 60 seconds as in the current architecture. Keys should expire before the next poll if the poller crashes. |
| WireGuard | Using kernel WireGuard module inside Docker containers | Use `wireguard-go` (userspace implementation). Requires `NET_ADMIN` capability and `/dev/net/tun` but no kernel module loading. Verify availability on RunPod before building. |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Polling all providers sequentially in one goroutine | Availability data is stale by the time the last provider is polled; if one provider times out, all providers' data is delayed | Poll each provider in its own goroutine with independent timeouts and error handling | When adding a second provider (E2E Networks); a slow provider blocks the fast one |
| No connection pooling for RunPod API calls | Under load, each provision request opens a new HTTP connection, causing TCP handshake overhead and potential connection exhaustion | Use `http.Client` with a configured `Transport` that limits `MaxIdleConnsPerHost` and reuses connections | At ~50 concurrent provisions; Go's default transport handles this well but explicit configuration prevents surprises |
| Reporting every second to Stripe for every active instance | At 1000 concurrent instances, that is 1000 meter events/second -- exactly at Stripe's API limit with zero headroom | Batch usage reporting: accumulate locally and report every 60 seconds. 1000 instances = ~17 events/s, well within limits | At ~500 concurrent instances if you report per-second; design for batch from day one |
| Single Redis instance for availability cache | Redis failure means no availability data and all provision requests fail because they cannot query cached offerings | Implement a local in-memory fallback cache. If Redis is down, serve last-known data with a "stale" flag. | Any Redis maintenance window or network blip; even 5 seconds of Redis downtime causes customer-visible failures |
| WireGuard proxy handling all SSH forwarding | A single proxy server has finite bandwidth and connection limits; each SSH session is a persistent connection consuming resources | Design proxy for horizontal scaling from the start (multiple proxy IPs, GeoDNS routing); even if you deploy one initially, the architecture should support N | At ~200 concurrent SSH sessions with active data transfer (ML training output streaming) |
| pgx connection pool exhaustion | All API requests block. Health checks fail. Cascading timeouts. | Set `MaxConns` to 10-20 for Phase 1. Use `context.WithTimeout(ctx, 5*time.Second)` on all queries. Monitor `pool.Stat()`. | Under sustained load with slow queries or lock contention |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| WireGuard private keys stored unencrypted in PostgreSQL | Database breach exposes all tunnel keys; attacker can decrypt all WireGuard traffic or impersonate instances | Encrypt with `pgp_sym_encrypt()` using a key from environment variable, not stored in DB. The schema already has `wg_private_key_enc` -- honor the `_enc` suffix. |
| Internal endpoints (`/internal/instances/{id}/ready`) accessible from the internet | Attacker can mark arbitrary instances as "ready," bypassing the actual boot process; can manipulate instance lifecycle | Bind internal endpoints to `127.0.0.1` only, or require a shared secret (internal bearer token) validated per-request. The cloud-init script uses `Authorization: Bearer {{.InternalToken}}` -- ensure this token is cryptographically random per-instance. |
| Cloud metadata endpoint (169.254.169.254) accessible from inside the instance | Customer can query the upstream provider's metadata service, revealing provider identity, instance type, region, and potentially IAM credentials | Block `169.254.169.254` in iptables rules within the container. Verify this works in RunPod's container environment. |
| RunPod API key exposed in environment variables inside the pod | Customer SSHs in, runs `env`, and sees the RunPod API key. They can now directly control RunPod resources. | Never pass the RunPod API key to the pod. All RunPod API calls happen from the `gpuctl` server, not from inside instances. The pod only receives its own WireGuard keys and SSH keys. |
| `InternalToken` reuse across instances | If one instance's internal token is compromised (e.g., from instance logs), it can be used to send lifecycle events for other instances | Generate a unique, cryptographically random token per instance. Store the hash in PostgreSQL. Validate by hashing the presented token and comparing. |
| DNS queries from instances bypass WireGuard tunnel | Customer's DNS queries go directly to the upstream provider's DNS resolver, leaking that the instance is on RunPod/E2E Networks | Force all DNS through the WireGuard tunnel by setting the instance's `/etc/resolv.conf` to point to the WireGuard gateway (10.0.0.1). Block UDP/TCP port 53 to all destinations except the WireGuard interface. |
| Upstream error messages passed through to customers | RunPod GraphQL error messages contain provider-specific details ("RunPod pod creation failed: insufficient A100 capacity in US-TX") | Wrap all upstream errors in generic GPU.ai error messages. Log the original error internally with `slog`. Return "GPU provisioning failed -- try again or select a different configuration" to the customer. |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Showing exact availability count ("12 available") when data may be 30s stale | Customer sees 12, tries to provision, gets "unavailable" error. Feels like a broken platform. | Show availability as a qualitative indicator: "High availability" / "Limited" / "Scarce." Or show count with a "last updated X seconds ago" timestamp. |
| Not explaining spot instance risks at provision time | Customer provisions a spot H100, starts a 3-day training run, gets evicted after 2 hours. Blames GPU.ai. | Require explicit opt-in for spot tier with a clear warning: "This instance can be terminated with 5 seconds notice. Save your work frequently." Show on-demand price alongside for comparison. |
| Billing starts but instance is not yet SSH-accessible | Customer is charged from provision request but cannot actually use the GPU for 15-45 seconds during boot. Feels unfair. | Either (a) start billing only when the instance reports "ready" and eat the upstream cost, or (b) clearly communicate "billing starts at provision, instance ready in ~15s" and show both timestamps in the UI. |
| WireGuard tunnel drops silently and customer cannot SSH | Customer thinks the platform is broken. No indication of what went wrong. | Implement a connection status indicator in the dashboard. If the health check detects the tunnel is down, show "Connection issue -- reconnecting" rather than just timing out. Provide a "reconnect" button. |
| Instance termination is ambiguous about data loss | Customer terminates an instance and loses all data that was not on a persistent volume. | Show a confirmation dialog that explicitly lists what will be lost: "All data not on a persistent volume will be deleted. This cannot be undone." |

## "Looks Done But Isn't" Checklist

- [ ] **WireGuard tunnel:** Tunnel "connects" (handshake succeeds) but no traffic flows -- verify AllowedIPs, IP forwarding on proxy, NAT masquerade rules, and MTU settings (WireGuard overhead reduces effective MTU by 60-80 bytes)
- [ ] **Privacy layer:** API responses strip provider name but `env`, `/etc/hosts`, `nvidia-smi`, DNS, traceroute, and metadata endpoint all leak identity -- run full privacy audit
- [ ] **Billing integration:** Stripe subscription created and meter events sent, but invoice shows $0.00 because integer rounding killed sub-cent amounts -- verify with a real 5-minute session and check the invoice
- [ ] **Availability polling:** Poller runs and Redis keys exist, but keys use 60s TTL while polling is 30s -- stale data survives one poll cycle after inventory changes. Verify TTL <= poll interval + small buffer
- [ ] **Spot handling:** Provisioning works for spot instances but no SIGTERM handler, no health check, no automatic billing stop on interruption -- test by manually terminating a spot pod on RunPod and verify GPU.ai's response
- [ ] **Instance ready callback:** Callback fires and status updates to "running" but no verification that the WireGuard tunnel is actually functional -- add a tunnel connectivity check (ping 10.0.0.1) before reporting ready
- [ ] **SSH access:** Customer can SSH via `gpu-4a7f.gpu.ai` but the hostname resolves to the proxy IP only if DNS is configured -- verify DNS A record creation is part of the provisioning flow, not a manual step
- [ ] **Firewall rules:** iptables rules applied inside container but the container restarts (RunPod restart) and iptables rules are lost -- persist rules or re-apply on every startup
- [ ] **Database precision:** `price_per_hour NUMERIC(10,4)` loses precision for sub-cent per-second calculations -- verify arithmetic: `2.12 / 3600 = 0.000588...` which needs at least 6 decimal places
- [ ] **WireGuard address allocation:** Addresses assigned from `10.0.0.0/24` subnet but this limits to 253 concurrent instances -- use `/16` or `/12` subnet from the start
- [ ] **RunPod init:** Architecture doc's `bootstrap.sh` uses cloud-init patterns (`systemctl`, `apt-get`) but RunPod uses Docker containers with `pre_start.sh`/`post_start.sh` -- verify the init mechanism actually works on a real RunPod pod

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Cloud-init approach fails on RunPod (containers not VMs) | HIGH | Redesign initialization as custom Docker image + env var injection. Requires building and publishing a GPU.ai base image to a container registry. 1-2 weeks of work. |
| WireGuard cannot run in RunPod containers | HIGH | Fall back to SSH reverse tunnels or a userspace networking solution (e.g., `wireguard-go` if `/dev/net/tun` is available). Requires re-architecting the privacy layer. 2-3 weeks. |
| Billing amounts are wrong due to integer/float issues | MEDIUM | Fix the unit conversion and re-issue corrected invoices via Stripe. Customer trust impact depends on whether over- or under-charged. 2-3 days to fix code, 1 week to reconcile. |
| Stale availability causes provisioning failures | LOW | Add cache invalidation on failure and graceful error handling. Purely a code change, no architectural rework. 1-2 days. |
| Privacy leak discovered (env vars, DNS, metadata) | MEDIUM | Add scrubbing steps to startup script. Each leak source is an independent fix. 1-3 days per leak source. But if customers have already noticed, trust damage is done. |
| Spot interruption billing continues after pod death | MEDIUM | Fix health check to detect terminated pods. Retroactively credit affected customers via Stripe. 2-3 days to fix, 1 week to audit and credit. |
| WireGuard address space exhausted (10.0.0.0/24) | MEDIUM | Migrate to larger subnet. Requires updating proxy config, all active tunnels, and the address allocation logic. Can be done with a rolling migration but is disruptive. 3-5 days. |
| Billing race on termination | MEDIUM | Implement state machine with idempotent transitions. Retroactively reconcile with upstream. 3-5 days to fix, ongoing for reconciliation. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| RunPod uses containers, not VMs (no cloud-init) | Phase 2: RunPod Adapter | Deploy a test pod with custom Docker image and verify `pre_start.sh` executes correctly |
| WireGuard needs NET_ADMIN in container | Phase 2: RunPod Adapter | Run `wireguard-go` in a RunPod pod and confirm tunnel handshake before writing any Go WireGuard code |
| Stripe integer-only meter values | Phase 5: Billing | Create a test Stripe meter, send 300 events (simulating 5 minutes), generate an invoice, and verify the charged amount matches `300 * per_second_rate` |
| Stale availability cache race | Phase 6: Availability Engine | Simulate a scenario where a GPU becomes unavailable between polls; verify the provision failure triggers cache invalidation |
| Privacy leaks through env/DNS/metadata | Phase 3: WireGuard Privacy Layer | Automated privacy audit script that SSHs into a test instance and checks all leak vectors |
| Spot interruption handling | Phase 2: RunPod Adapter + Phase 8: Health Monitoring | Terminate a spot pod via RunPod API and verify GPU.ai detects it within 30 seconds and stops billing |
| Billing drift (upstream vs. customer) | Phase 5: Billing | Compare `billing_start` timestamps against RunPod pod creation timestamps for 10 test instances; margin leakage should be < 5 seconds |
| Billing race on termination | Phase 4: Database + Phase 5: Billing | Test concurrent termination requests; verify billing_end is set exactly once and upstream is terminated |
| WireGuard address space exhaustion | Phase 3: WireGuard Privacy Layer | Use `10.0.0.0/16` (65,534 addresses) from day one; verify address allocation assigns unique IPs and reclaims on termination |
| Float precision in billing | Phase 4: Database + Phase 5: Billing | Unit test: `Decimal("2.12").div(3600).mul(3600)` must equal `Decimal("2.12")` exactly |
| Internal endpoint security | Phase 3: WireGuard Privacy Layer + Phase 7: API Routes | Port scan the public IP and verify `/internal/*` endpoints are not accessible from outside |

## Sources

- [RunPod Pod Management - GraphQL API](https://docs.runpod.io/sdks/graphql/manage-pods) -- HIGH confidence, official docs
- [RunPod Pricing and Spot Instance Details](https://docs.runpod.io/pods/pricing) -- HIGH confidence, official docs
- [RunPod Templates Overview](https://docs.runpod.io/pods/templates/overview) -- HIGH confidence, official docs
- [RunPod Container Initialization Scripts](https://deepwiki.com/runpod/containers/6.2-initialization-scripts) -- MEDIUM confidence, community wiki based on official containers
- [RunPod Pod Creation and Configuration](https://deepwiki.com/runpod/docs/3.1-pod-creation-and-configuration) -- MEDIUM confidence, community wiki
- [RunPod Custom Container Setup](https://www.njordy.com/2026/01/02/runpod-custom-container/) -- MEDIUM confidence, practitioner blog
- [Stripe Recording Usage API](https://docs.stripe.com/billing/subscriptions/usage-based/recording-usage-api) -- HIGH confidence, official docs
- [Stripe Billing Meters API](https://docs.stripe.com/api/billing/meter) -- HIGH confidence, official docs
- [Stripe Limitations for Usage-Based Billing](https://www.withorb.com/blog/stripe-limitations-for-usage-based-billing) -- MEDIUM confidence, third-party analysis
- [Stripe Legacy Usage Records Migration](https://docs.stripe.com/billing/subscriptions/usage-based-legacy/migration-guide) -- HIGH confidence, official docs
- [Cloud-init Failure States](https://docs.cloud-init.io/en/latest/explanation/failure_states.html) -- HIGH confidence, official docs
- [WireGuard Userspace Implementation (wireguard-go)](https://www.wireguard.com/xplatform/) -- HIGH confidence, official WireGuard project
- [WireGuard in Docker Containers](https://github.com/masipcat/wireguard-go-docker) -- MEDIUM confidence, community project
- [WireGuard Security Hardening](https://contabo.com/blog/hardening-your-wireguard-security-a-comprehensive-guide/) -- MEDIUM confidence, hosting provider blog
- [Clerk Go SDK JWT Verification](https://clerk.com/docs/guides/sessions/verifying) -- HIGH confidence, official docs
- [Redis Race Conditions](https://redis.io/glossary/redis-race-condition/) -- HIGH confidence, official Redis docs
- [GPU Cloud Security Risks](https://www.secureworld.io/industry-news/gpu-hosting-llms-unseen-backdoor) -- MEDIUM confidence, security publication
- [WireGuard Endpoints and IP Addresses](https://www.procustodibus.com/blog/2021/01/wireguard-endpoints-and-ip-addresses/) -- MEDIUM confidence, WireGuard management platform blog
- [WireGuard Tunnel Connection Drop Issues](https://forums.freebsd.org/threads/wireguard-tunnel-intermittent-connection-drops.99705/) -- MEDIUM confidence, community forums
- [GPU Cloud Billing Architecture](https://rafay.co/ai-and-cloud-native-blog/gpu-cloud-billing-from-usage-metering-to-billing/) -- MEDIUM confidence, infrastructure vendor blog
- [RunPod API Community Discussions](https://www.answeroverflow.com/m/1376872820704809031) -- LOW confidence, community reports

---
*Pitfalls research for: GPU cloud aggregation platform*
*Researched: 2026-02-24*
