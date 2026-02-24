# Feature Research

**Domain:** GPU Cloud Aggregation Platform (re-renting upstream GPUs behind a unified interface with privacy layer)
**Researched:** 2026-02-24
**Confidence:** MEDIUM-HIGH (based on direct competitor product pages, comparison articles, and official documentation)

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist on any GPU cloud platform. Missing these means customers leave immediately.

#### 1. Core Compute & Access

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| SSH access to instances | Every GPU cloud offers SSH. It is the primary interaction model for ML engineers. | LOW | Cloud-init already handles SSH key injection. Route through WireGuard proxy. |
| GPU instance provisioning (create/terminate) | The fundamental product. Users need to spin up and destroy GPU instances on demand. | HIGH | Core provisioning engine with provider adapter pattern already designed. |
| Multiple GPU types (H100, A100, L40S, RTX 4090) | Customers expect choice. RunPod offers 30+ GPU types, Lambda offers 6+, Vast.ai offers 68 types. | MEDIUM | Depends on upstream provider catalog. RunPod alone covers most popular SKUs. |
| Instance status and lifecycle visibility | Users need to know if their instance is creating, running, stopping, or errored. | MEDIUM | Requires polling upstream status + WireGuard health check. |
| Web dashboard for instance management | Every competitor has a dashboard. CLI-only is not viable for most users. | HIGH | Next.js frontend already planned. Must show instances, status, GPU availability. |
| API for programmatic access | RunPod, Lambda, Vast.ai all provide REST APIs. Power users and CI/CD pipelines depend on API access. | MEDIUM | Go backend already API-first. REST endpoints defined in architecture doc. |

#### 2. Authentication & Account

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| User authentication (signup/login) | Non-negotiable. Every cloud platform requires accounts. | LOW | Clerk handles this -- JWT verification middleware. |
| SSH key management (CRUD) | Users manage SSH keys to access instances. RunPod, Lambda, all competitors have this. | LOW | Already in architecture. Simple CRUD with fingerprint display. |
| Organization/team accounts | Teams share billing and instances. RunPod, Lambda, CoreWeave all support org-level accounts. | MEDIUM | Schema already has organizations table. Clerk supports organizations. |

#### 3. Billing & Pricing

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Per-second (or per-minute) billing | Industry standard. RunPod bills per-millisecond. Lambda bills per-minute. Customers reject hourly billing. | HIGH | Stripe usage-based metering. Must track billing_start/billing_end precisely. |
| Transparent pricing display | Users expect to see GPU prices before provisioning. Every competitor shows a pricing page and in-dashboard pricing. | LOW | Availability engine already returns price_per_hour. Display in dashboard. |
| Usage history and cost breakdown | Users need to understand what they spent and on what. All competitors show usage reports. | MEDIUM | usage_records table exists. Build aggregation queries + dashboard UI. |
| Payment method management | Credit card on file before provisioning. Standard Stripe checkout flow. | LOW | Stripe customer portal handles this. |

#### 4. GPU Availability

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Real-time availability display | Users need to see what GPUs are available before trying to provision. RunPod and Vast.ai show live availability. | MEDIUM | Availability poller (30s) already designed. Redis cache with API endpoint. |
| Region selection | All providers offer region choice (us-west, us-east, eu, etc.). Users care about latency and data residency. | LOW | Architecture supports region in GPUOffering. Display in UI. |
| Tier selection (on-demand vs spot) | RunPod, Vast.ai, and hyperscalers all offer spot/interruptible pricing at 30-60% discount. | MEDIUM | Tier enum already defined (on_demand, spot, reserved). |

#### 5. Instance Environment

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Docker image support (custom environments) | RunPod, Lambda all support custom Docker images. Users need their own CUDA/PyTorch stacks. | MEDIUM | Cloud-init already handles docker pull + run. InstanceConfig has docker_image field. |
| Pre-installed ML frameworks (PyTorch, CUDA) | Lambda pre-installs Lambda Stack. RunPod offers 50+ templates. Users expect ready-to-use ML environments. | LOW | Provide curated default Docker images (pytorch/pytorch:latest-cuda, etc.). |

### Differentiators (Competitive Advantage)

Features that set GPU.ai apart from direct competitors. These are not expected, but they create the unique value proposition.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Complete provider abstraction / privacy layer** | Customer never sees RunPod, Lambda, etc. GPU.ai IS the cloud provider in their eyes. No other aggregator does this with full network privacy (WireGuard tunnels, branded hostnames, firewalled upstream IPs). Vast.ai is a marketplace -- you see the host. GPU.ai is a brand. | HIGH | Core differentiator. WireGuard proxy, cloud-init firewall rules, hostname branding, MOTD. This is the hardest and most valuable feature. |
| **Cross-provider availability aggregation** | See GPUs from ALL providers in one view. If RunPod has no H100s, but E2E does, the customer still gets an H100. No competitor offers this transparently -- they are all single-provider. | MEDIUM | Availability poller across multiple adapters, merged response with provider identity stripped. Already designed. |
| **India cost arbitrage** | E2E Networks GPUs are 30-50% cheaper than US equivalents. GPU.ai can offer "india-mumbai" region at dramatically lower prices. No US-focused competitor offers this. | MEDIUM | Requires E2E adapter (Phase 2 scope). But pricing advantage is a major differentiator once live. |
| **Environment persistence across providers** | Save your Docker environment, redeploy on any provider. If you trained on RunPod, deploy on E2E. Portable workspaces across underlying infrastructure. | HIGH | environments table exists. Needs container registry integration (push/pull images). Complex but very valuable. |
| **Unified billing across providers** | One Stripe invoice for compute from RunPod + E2E + future providers. No managing multiple accounts/billing. | MEDIUM | Already designed. Stripe integration aggregates all usage regardless of upstream provider. |
| **Branded GPU hostnames** | `gpu-4a7f.gpu.ai` instead of raw IP addresses. Professional, memorable, and hides upstream identity. SSH is `ssh user@gpu-4a7f.gpu.ai`. | LOW | Cloud-init sets hostname. DNS wildcard *.gpu.ai points to proxy. Cheap to implement, high perceived value. |
| **Automatic best-price routing** | When user requests "H100, us-west, on-demand", the engine picks the cheapest available across all providers. Customer gets best price without shopping around. | MEDIUM | Provisioning engine already designed to pick "best available (price, region, availability)". |
| **Spot instance with graceful migration** | When a spot instance is preempted, automatically migrate to another provider. No single-provider platform can do this -- they just terminate. | HIGH | Future differentiator. Requires checkpoint/restore support and multi-provider failover logic. Defer to v2+. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems for an aggregation platform specifically.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| **Serverless GPU endpoints** | RunPod Serverless and Vast.ai Serverless are popular. Users want to deploy inference endpoints that auto-scale. | Serverless requires deep integration with upstream provider serverless APIs (or building your own orchestrator). Massive engineering effort. Abstracts away from GPU.ai's core value of persistent instances with privacy. Competing directly with RunPod Serverless is unwinnable at launch. | Focus on persistent GPU instances. Revisit serverless in Phase 3 if there is demand. |
| **Multi-node GPU clusters with InfiniBand** | Large training jobs need 16-2000+ GPUs networked together. CoreWeave and Lambda specialize in this. | InfiniBand networking cannot span providers. Multi-node clusters require physical co-location and dedicated interconnect hardware. The privacy/WireGuard layer adds latency incompatible with distributed training. An aggregation platform cannot provide this. | Offer single-node multi-GPU instances (up to 8 GPUs). For multi-node, direct customers to a "Contact Us" reserved tier flow where ops can arrange dedicated multi-node from a single provider. |
| **Jupyter Notebook / web IDE** | RunPod and Lambda both offer in-browser Jupyter access. Users expect it. | Web IDE requires proxying HTTP traffic through the WireGuard tunnel and exposing ports. Adds significant proxy complexity. Every upstream provider has different port-forwarding behavior. Security surface area increases. | Provide SSH access only for v1. Document how to set up SSH tunneling for Jupyter (`ssh -L 8888:localhost:8888 user@gpu-xxx.gpu.ai`). Consider port-forwarding proxy in v1.x. |
| **Kubernetes-native orchestration** | CoreWeave is Kubernetes-native. Enterprise teams want kubectl access. | Kubernetes adds massive operational overhead. GPU.ai provisions bare-metal-like instances, not pods. K8s abstraction conflicts with the single-instance-per-customer model. | Keep the instance abstraction simple (create/terminate). If enterprise customers need K8s, they should use CoreWeave directly. |
| **Community Cloud / marketplace for host GPUs** | Vast.ai's core model. Let anyone host GPUs and earn money. | Running a two-sided marketplace is a fundamentally different business. Quality control, trust, payment splitting, compliance with distributed hosts. Massive operational burden. | Curate upstream providers carefully. Add providers through the adapter interface. Do not open to random hosts. |
| **Real-time GPU utilization monitoring** | Users want to see GPU%, VRAM%, temperature in their dashboard. Lambda and RunPod offer this. | Requires an agent running on each instance that reports metrics back. Adds complexity to cloud-init, creates a data pipeline for metrics ingestion, and requires real-time websocket/SSE streaming to dashboard. | For v1, do not build monitoring. Customers can SSH in and run `nvidia-smi`. Consider adding a lightweight metrics agent in v1.x that reports to the health endpoint. |
| **Snapshots / volume backups** | Hyperscalers offer VM snapshots. Users want to save state. | Snapshots require storage infrastructure and deep provider integration. Each upstream provider handles snapshots differently (or not at all). Cannot snapshot across providers. | Use Docker environment persistence instead. Users save their environment as a Docker image, which is portable across providers. This is better than snapshots anyway. |

## Feature Dependencies

```
[Auth (Clerk JWT)]
    |-- requires --> [User/Org database schema]
    |-- enables --> [SSH Key Management]
    |-- enables --> [Instance Provisioning]
    |-- enables --> [Billing]

[Provider Adapter (RunPod)]
    |-- requires --> [Provider Interface definition]
    |-- enables --> [Instance Provisioning]
    |-- enables --> [Availability Polling]

[Availability Polling]
    |-- requires --> [Provider Adapter(s)]
    |-- requires --> [Redis Cache]
    |-- enables --> [Real-time Availability Display]
    |-- enables --> [Best-price Routing]

[Instance Provisioning]
    |-- requires --> [Provider Adapter]
    |-- requires --> [WireGuard Key Generation]
    |-- requires --> [Cloud-init Template]
    |-- requires --> [Auth]
    |-- requires --> [Database (instances table)]
    |-- enables --> [Instance Lifecycle Management]
    |-- enables --> [Billing Start/Stop]

[WireGuard Privacy Layer]
    |-- requires --> [Proxy/Bastion Server infrastructure]
    |-- requires --> [Cloud-init with WG config]
    |-- requires --> [DNS wildcard *.gpu.ai]
    |-- enables --> [SSH Access through branded hostname]
    |-- enables --> [Provider Identity Hiding]

[Billing (Stripe)]
    |-- requires --> [Auth (for customer identity)]
    |-- requires --> [Instance Lifecycle (billing_start/end)]
    |-- enables --> [Usage History]
    |-- enables --> [Invoice Generation]

[Dashboard (Next.js)]
    |-- requires --> [API endpoints (all of them)]
    |-- requires --> [Auth (Clerk frontend SDK)]
    |-- enhances --> [Every other feature (visual interface)]

[Environment Persistence]
    |-- requires --> [Container Registry]
    |-- requires --> [Instance Provisioning (to deploy images)]
    |-- requires --> [Docker support on instances]
    |-- conflicts with --> Snapshots (choose one persistence model)

[India Cost Arbitrage]
    |-- requires --> [E2E Networks Adapter]
    |-- enhances --> [Cross-provider Availability]
    |-- enhances --> [Best-price Routing]
```

### Dependency Notes

- **Instance Provisioning requires WireGuard**: The privacy layer is not optional. Without WireGuard, customers see upstream IPs and the core value proposition collapses.
- **Billing requires Instance Lifecycle**: Cannot meter usage without knowing when instances start and stop. billing_start must be set on provision, billing_end on termination.
- **Dashboard enhances everything but blocks nothing**: The API is the primary interface. Dashboard is a UI layer on top. Can launch with API-only and add dashboard incrementally.
- **Environment Persistence conflicts with Snapshots**: Pick one persistence model. Docker images are portable across providers; snapshots are not. Docker wins for an aggregation platform.

## MVP Definition

### Launch With (v1 -- Closed Beta)

Minimum viable product for 10-20 beta users to validate the concept.

- [ ] **Auth + SSH key management** -- Users can sign up, manage SSH keys (Clerk + CRUD)
- [ ] **RunPod adapter** -- Prove provisioning works with one upstream provider
- [ ] **GPU availability display** -- Show what GPUs are available in real-time (Redis poller)
- [ ] **Instance provisioning (create/list/get/terminate)** -- Core lifecycle
- [ ] **WireGuard privacy layer** -- Branded hostnames, hidden upstream IPs, firewalled instances
- [ ] **Per-second billing** -- Stripe usage metering, transparent pricing
- [ ] **API endpoints** -- Full REST API for all operations
- [ ] **Basic dashboard** -- Landing page, auth flow, instance list, GPU availability, billing summary

### Add After Validation (v1.x)

Features to add once core is working and beta feedback comes in.

- [ ] **Docker environment persistence** -- Save/deploy custom Docker images across instances (trigger: users asking for environment portability)
- [ ] **Port forwarding / HTTP proxy** -- Access web services (Jupyter, TensorBoard) running on instances through the proxy (trigger: multiple users requesting Jupyter access)
- [ ] **Spend limits and billing alerts** -- Email notifications when spending exceeds thresholds (trigger: users getting surprise bills)
- [ ] **Lightweight instance health metrics** -- GPU utilization, uptime, basic metrics via health endpoint (trigger: users wanting visibility without SSH)
- [ ] **Team/org member management** -- Invite team members, shared billing, basic RBAC (trigger: team accounts signing up)
- [ ] **CLI tool** -- `gpuai create --gpu h100 --region us-west` (trigger: power users wanting automation)

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] **E2E Networks adapter (India)** -- 30-50% cost savings, but adds provider complexity. Defer until RunPod flow is solid.
- [ ] **Spot instance migration** -- Auto-migrate preempted spot instances to another provider. Requires checkpoint/restore.
- [ ] **Multi-node cluster support** -- Contact-us flow for reserved tier (Novacore/CTRLS). Ops-assisted, not automated.
- [ ] **Serverless inference endpoints** -- Only if persistent instances prove insufficient for inference customers.
- [ ] **Predictive resource allocation (Phase 3)** -- Job-based pricing with memory prediction. Requires Phase 2 own hardware + MIG.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Instance provisioning (create/terminate) | HIGH | HIGH | P1 |
| WireGuard privacy layer | HIGH | HIGH | P1 |
| SSH access via branded hostname | HIGH | MEDIUM | P1 |
| Auth (Clerk JWT) | HIGH | LOW | P1 |
| SSH key management | HIGH | LOW | P1 |
| Real-time GPU availability | HIGH | MEDIUM | P1 |
| Per-second billing (Stripe) | HIGH | HIGH | P1 |
| REST API | HIGH | MEDIUM | P1 |
| Basic dashboard (Next.js) | HIGH | HIGH | P1 |
| Docker image support | MEDIUM | LOW | P1 |
| Instance status/lifecycle | HIGH | MEDIUM | P1 |
| Transparent pricing display | MEDIUM | LOW | P1 |
| Usage history / cost breakdown | MEDIUM | MEDIUM | P2 |
| Environment persistence (save/deploy) | MEDIUM | HIGH | P2 |
| Port forwarding (Jupyter proxy) | MEDIUM | HIGH | P2 |
| Billing alerts / spend limits | MEDIUM | MEDIUM | P2 |
| Team/org member management | MEDIUM | MEDIUM | P2 |
| CLI tool | MEDIUM | MEDIUM | P2 |
| GPU health metrics | LOW | MEDIUM | P2 |
| E2E Networks adapter | HIGH | HIGH | P2 |
| Spot migration across providers | HIGH | HIGH | P3 |
| Multi-node clusters | MEDIUM | HIGH | P3 |
| Serverless endpoints | MEDIUM | HIGH | P3 |

**Priority key:**
- P1: Must have for closed beta launch
- P2: Should have, add based on user feedback
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | RunPod | Lambda | Vast.ai | CoreWeave | GPU.ai Approach |
|---------|--------|--------|---------|-----------|-----------------|
| Instance types | Pods (persistent), Serverless | On-demand instances, 1-Click Clusters | On-demand, Interruptible, Auction | K8s-native VMs | Persistent instances only (v1) |
| GPU selection | 30+ types | ~6 types (high-end focus) | 68 types from marketplace | Enterprise-grade (H100, B200, GB200) | Whatever upstream providers offer, merged into unified catalog |
| Billing granularity | Per-millisecond | Per-minute | Per-second | Per-hour (contracts) | Per-second via Stripe metering |
| Access methods | SSH, Jupyter, Web Terminal, VSCode | SSH, Jupyter, Web Terminal | SSH, Jupyter | K8s kubectl, SSH | SSH only (v1), port forwarding (v1.x) |
| Storage | Container disk, volumes, network volumes | Persistent storage | Disk-based | Kubernetes PVCs | Docker images for environment persistence |
| Networking | Private cross-DC, no egress fees | InfiniBand (clusters), no egress | Standard | InfiniBand, NVLink | WireGuard tunnels (privacy layer), no egress |
| Templates | 50+ pre-configured | Lambda Stack pre-installed | Community templates | None (BYO K8s) | Curated default Docker images |
| Multi-node | Instant Clusters (2-8 nodes) | 1-Click Clusters (16-2000+) | No built-in | Native (rack-scale) | Not supported (v1). Reserved tier contact-us (v2) |
| Privacy | None (you see RunPod branding) | None (you see Lambda branding) | None (you see host details) | None (you see CoreWeave) | **Full abstraction**: branded hostnames, hidden IPs, WireGuard firewall |
| Provider aggregation | Single provider (RunPod) | Single provider (Lambda) | Marketplace (many hosts) | Single provider (CoreWeave) | **Multi-provider**: RunPod + E2E + future, transparent to customer |
| Monitoring | Dashboard metrics | Dashboard + guest agent metrics | Basic stats | Kubernetes monitoring | None (v1). nvidia-smi via SSH. Lightweight metrics (v1.x) |
| Compliance | SOC 2 Type II | SOC 2 | SOC 2 Type I & II | SOC 2, ISO 27001 | Inherits upstream compliance. Own SOC 2 when at scale. |
| Spot instances | Community Cloud (spot-like) | Not offered | Interruptible + Auction | Not offered | Pass-through spot tier from upstream. Migration across providers (v2+). |

## Sources

- [RunPod Cloud GPUs product page](https://www.runpod.io/product/cloud-gpus) -- HIGH confidence, official product page
- [RunPod Pricing](https://www.runpod.io/pricing) -- HIGH confidence, official pricing
- [RunPod Documentation - Pods Overview](https://docs.runpod.io/pods/overview) -- HIGH confidence, official docs
- [RunPod Serverless](https://www.runpod.io/product/serverless) -- HIGH confidence, official product page
- [RunPod REST API blog post](https://www.runpod.io/blog/runpod-rest-api-gpu-management) -- HIGH confidence, official blog
- [Lambda AI Pricing](https://lambda.ai/pricing) -- HIGH confidence, official pricing
- [Lambda AI Instances](https://lambda.ai/instances) -- HIGH confidence, official product page
- [Vast.ai Year in Review 2025](https://vast.ai/article/vast-ai-2025-year-in-review) -- HIGH confidence, official blog
- [Vast.ai Pricing](https://vast.ai/pricing) -- HIGH confidence, official page
- [CoreWeave AI Infrastructure](https://www.coreweave.com/ai-infrastructure) -- HIGH confidence, official page
- [GPU Price Comparison 2026 - getdeploying.com](https://getdeploying.com/gpus) -- MEDIUM confidence, third-party comparison
- [Northflank RunPod vs Vast.ai comparison](https://northflank.com/blog/runpod-vs-vastai-northflank) -- MEDIUM confidence, third-party analysis
- [RunPod Top 12 GPU Providers Guide](https://www.runpod.io/articles/guides/top-cloud-gpu-providers) -- MEDIUM confidence, RunPod-authored (biased toward RunPod)
- [Hyperstack Cloud GPU Providers Ranking](https://www.hyperstack.cloud/blog/case-study/top-cloud-gpu-providers) -- MEDIUM confidence, competitor-authored
- [DigitalOcean Best Cloud GPU Platforms](https://www.digitalocean.com/resources/articles/best-cloud-gpu-platforms) -- MEDIUM confidence, third-party guide
- [ThunderCompute Spot Instance Interruption Rates](https://www.thundercompute.com/blog/should-i-use-cloud-gpu-spot-instances) -- MEDIUM confidence, competitor analysis
- [Cast AI GPU Price Report](https://cast.ai/reports/gpu-price/) -- MEDIUM confidence, independent report

---
*Feature research for: GPU Cloud Aggregation Platform*
*Researched: 2026-02-24*
