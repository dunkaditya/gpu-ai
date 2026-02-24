# Research Summary: GPU.ai

**Domain:** GPU Cloud Aggregation Platform
**Researched:** 2026-02-24
**Overall confidence:** HIGH

## Executive Summary

GPU.ai is a GPU cloud aggregation platform that provisions GPU instances from upstream providers (starting with RunPod) behind a WireGuard privacy layer, presenting a unified brand to customers. The architecture is a single Go binary (`gpuctl`) handling API, provisioning, billing, auth, and availability polling, paired with a Next.js customer dashboard. The technology stack is well-defined by project constraints and aligns cleanly with current Go ecosystem best practices.

The Go backend stack centers on Go 1.24 with stdlib `net/http` (1.22+ routing patterns providing method matching and path variables), pgx v5 for PostgreSQL, go-redis v9 for Redis caching, Stripe Go SDK v84 for usage-based billing, and Clerk Go SDK v2 for JWT authentication. WireGuard management uses the official wgctrl-go library for programmatic peer management, with key generation via golang.org/x/crypto/curve25519. All libraries are actively maintained, have stable APIs, and are production-proven.

The frontend stack uses Next.js 16 with App Router, shadcn/ui for dashboard components, Tailwind CSS v4 for styling, TanStack Query v5 for server state management, and Clerk's Next.js SDK for authentication. This is the standard 2025-2026 stack for SaaS dashboards and has excellent ecosystem support.

The most critical risks are billing accuracy (distributed transaction across Stripe/PostgreSQL/upstream provider), cloud-init reliability across providers (RunPod uses Docker-based instances, not traditional VMs), and WireGuard key security (encryption at rest is essential). The architecture's biggest strength is the Provider adapter interface, which makes adding new upstream providers a contained, testable change.

## Key Findings

**Stack:** Go 1.24 stdlib + pgx v5 + go-redis v9 + Stripe v84 + Clerk v2 + wgctrl-go. Next.js 16 + shadcn/ui + Tailwind v4. All actively maintained, all current versions.

**Architecture:** Single binary monolith with clear component boundaries (provider adapters, provisioning engine, availability cache). Dependency injection via constructors, stdlib middleware pattern, async provisioning with poll-for-status.

**Critical pitfall:** Billing inaccuracy from race conditions between upstream provider terminate, PostgreSQL billing_end update, and Stripe meter event submission. Must use state machine for instance lifecycle and reconciliation jobs.

## Implications for Roadmap

Based on research, suggested phase structure:

1. **Foundation + Data Layer** - Go scaffold, config loading, health endpoint, PostgreSQL schema with pgx pool, Redis connection
   - Addresses: Project scaffolding, database layer
   - Avoids: Building on unstable foundation

2. **Auth + Provider Adapter** - Clerk JWT middleware, RunPod adapter implementation
   - Addresses: Authentication required for all subsequent features, proves upstream provisioning works
   - Avoids: Building features without auth (security gap)

3. **Privacy Layer + Provisioning Engine** - WireGuard keygen, cloud-init template, provisioning orchestration, instance lifecycle API
   - Addresses: Core differentiator (privacy), core product (provisioning)
   - Avoids: Cloud-init failures by testing on real RunPod instances early

4. **Availability + Billing** - Redis availability poller, Stripe usage metering, billing API
   - Addresses: Revenue generation, real-time GPU availability
   - Avoids: Stripe rate limit issues by implementing batched meter events from the start

5. **Dashboard** - Next.js 16 frontend with Clerk auth, GPU availability UI, instance management, billing dashboard
   - Addresses: Customer-facing interface, completes the product
   - Avoids: Building UI before API is stable

6. **Environment Persistence + Polish** - Docker image save/deploy, error handling hardening, reconciliation jobs
   - Addresses: Differentiator feature, operational safety nets
   - Avoids: Premature optimization before core is working

**Phase ordering rationale:**
- Database and auth MUST come first -- every subsequent feature depends on them
- RunPod adapter should be proven BEFORE building the full provisioning engine around it
- WireGuard + cloud-init should be tested on real RunPod instances as early as possible (biggest integration risk)
- Billing comes after provisioning because you need running instances to test billing
- Dashboard comes last because the API is the primary interface; dashboard is a layer on top
- Environment persistence is deferred because it is a differentiator, not table stakes for beta

**Research flags for phases:**
- Phase 3 (WireGuard + Cloud-Init): Needs deeper research on RunPod's Docker-based instance model vs traditional cloud-init. RunPod uses startup scripts in Docker entrypoint, not systemd cloud-init.
- Phase 4 (Stripe Billing): Needs deeper research on Stripe Meters API v2 vs legacy usage records migration. v82+ SDK deprecated legacy billing.
- Phase 2 (RunPod Adapter): Needs investigation of RunPod REST vs GraphQL API coverage. REST is newer but may not cover all operations yet.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All library versions verified via GitHub releases and pkg.go.dev. Go 1.24, pgx v5, go-redis v9, Stripe v84, Clerk v2 are all current and stable. |
| Features | HIGH | Feature landscape mapped from competitor analysis (RunPod, Lambda, Vast.ai, CoreWeave). Table stakes and differentiators are clear. |
| Architecture | HIGH | Architecture is well-defined in docs/ARCHITECTURE.md. Patterns (adapter interface, dependency injection, middleware chain) are standard Go practices. |
| Pitfalls | MEDIUM | Billing race conditions and cloud-init reliability are well-understood risks. WireGuard proxy SPOF and RunPod API quirks are less well-documented -- based on community reports and training data. |

## Gaps to Address

- **RunPod Docker-based cloud-init:** RunPod instances are Docker containers, not VMs. The cloud-init script in the architecture doc assumes VM-style boot (apt-get, systemctl). Need to investigate RunPod's startup script mechanism and adapt cloud-init accordingly. This is the highest-risk gap.
- **WireGuard proxy server setup:** The architecture assumes a proxy server exists but does not specify how it is provisioned, configured, or monitored. Need infrastructure-as-code for the proxy server before Phase 3.
- **Stripe Meters API specifics:** The migration from legacy usage records to Billing Meters is recent. Need to verify exact meter event schema, aggregation behavior, and invoicing flow during Phase 4 research.
- **E2E Networks API:** Deferred to later milestone, but needs investigation before Phase 2. REST API documentation quality, rate limits, and cloud-init support are unknowns.
- **DNS wildcard for *.gpu.ai:** Required for branded hostnames but not addressed in research. Need DNS provider that supports wildcard A records pointing to proxy server IP.

---
*Research summary for: GPU.ai*
*Researched: 2026-02-24*
