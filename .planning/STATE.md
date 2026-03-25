---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: completed
stopped_at: Completed 12-01-PLAN.md
last_updated: "2026-03-25T20:34:17.763Z"
last_activity: 2026-03-25 -- Completed 12-01 (Landing page mobile responsive fixes)
progress:
  total_phases: 15
  completed_phases: 13
  total_plans: 47
  completed_plans: 44
  percent: 94
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Customers can find available GPUs across providers and provision them instantly through a single interface, with a privacy layer that completely hides the upstream provider.
**Current focus:** Phase 12: Mobile Responsive Website Optimization

## Current Position

Phase: 12 (Mobile Responsive Website Optimization)
Plan: 1 of 3 in current phase
Status: Plan 12-01 Complete
Last activity: 2026-03-25 -- Completed 12-01 (Landing page mobile responsive fixes)

Progress: [█████████░] 94%

## Performance Metrics

**Velocity:**
- Total plans completed: 35
- Average duration: 2.1min
- Total execution time: 1.05 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation | 4 | 6min | 1.5min |
| 02-provider-abstraction | 3 | 7min | 2.3min |
| 03-privacy-layer | 3 | 6min | 2.0min |
| 04-auth-instance-lifecycle | 4 | 12min | 3.0min |
| 04.1-wireguard-integration-wiring | 2 | 4min | 2.0min |
| 04.2-instance-lifecycle-fix | 2 | 5min | 2.5min |
| 04.3-auth-idempotency-edge-cases | 1 | 3min | 3.0min |
| 05-ssh-keys-billing | 5 | 13min | 2.6min |
| 06-availability-health-monitoring | 4 | 9min | 2.3min |

**Recent Trend:**
- Last 5 plans: 07-03 (2min), 07-04 (5min), 09-01 (7min), 09-03 (7min), 10-01 (3min)
- Trend: stable

*Updated after each plan completion*
| Phase 05 P03 | 4min | 2 tasks | 4 files |
| Phase 05 P04 | 3min | 2 tasks | 8 files |
| Phase 05 P05 | 2min | 2 tasks | 2 files |
| Phase 06 P01 | 2min | 2 tasks | 6 files |
| Phase 06 P02 | 3min | 2 tasks | 3 files |
| Phase 06 P03 | 2min | 2 tasks | 4 files |
| Phase 06 P04 | 2min | 2 tasks | 5 files |
| Phase 08 P01 | 4min | 2 tasks | 29 files |
| Phase 08 P02 | 2min | 2 tasks | 8 files |
| Phase 07 P01 | 2min | 2 tasks | 4 files |
| Phase 07 P02 | 2min | 2 tasks | 9 files |
| Phase 07 P03 | 2min | 2 tasks | 6 files |
| Phase 07 P04 | 5min | 2 tasks | 18 files |
| Phase 09 P01 | 7min | 2 tasks | 12 files |
| Phase 09 P03 | 7min | 1 tasks | 14 files |
| Phase 09 P02 | 5min | 2 tasks | 5 files |
| Phase 10 P01 | 3min | 2 tasks | 11 files |
| Phase 10 P02 | 4min | 2 tasks | 5 files |
| Phase 10 P03 | 3min | 2 tasks | 4 files |
| Phase 11 P01 | 3min | 2 tasks | 8 files |
| Phase 11 P02 | 3min | 2 tasks | 5 files |
| Phase 11 P03 | 5min | 2 tasks | 6 files |
| Phase 12 P01 | 2min | 2 tasks | 4 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: 7-phase standard-depth structure derived from 65 requirements
- [Roadmap]: RunPod adapter in Phase 2 (early risk validation) -- no API keys yet so unit tests cover interface/registry only
- [Roadmap]: Privacy layer isolated in Phase 3 before instance lifecycle depends on it
- [Roadmap]: Dashboard last (Phase 7) -- API must be stable before building frontend
- [01-01]: Phase 1 Config struct scoped to 4 fields only (Port, DatabaseURL, RedisURL, InternalAPIToken)
- [01-01]: Go version upgraded from 1.22.0 to 1.24.0 (required by pgx/v5 dependency)
- [01-02]: Edit initial migration directly (no production data exists, greenfield)
- [01-02]: TIMESTAMPTZ for all time columns to avoid timezone bugs
- [01-02]: wg_private_key_enc suffix to clarify encryption at rest
- [01-02]: Per-migration transactions with rollback on error
- [01-03]: Redis client used directly (no wrapper) -- go-redis Client sufficient for Phase 1
- [01-03]: Health endpoint behind InternalAuthMiddleware -- prevents unauthenticated probing
- [01-03]: ConnectWithRetry as package-level function in db -- reusable for any service
- [01-03]: ServerDeps struct for constructor injection -- clean dependency boundary for testing
- [01-04]: 404 Not Found for rejected non-loopback IPs -- avoids revealing endpoint existence to scanners
- [01-04]: net.SplitHostPort for IP extraction -- handles both IPv4 and IPv6 RemoteAddr correctly
- [01-04]: LocalhostOnly as outermost middleware -- rejects external IPs before token check
- [02-02]: Two tiers only for v1 (on_demand and spot) -- TierReserved removed per CONTEXT.md
- [02-02]: Async fire-and-return provisioning model -- Provision returns upstream ID, GetStatus polls separately
- [02-02]: WireGuardPrivateKey removed from ProvisionRequest -- key generation deferred to Phase 3
- [02-02]: DatacenterLocation added to GPUOffering for datacenter drill-down info
- [02-02]: Re-registration allowed in registry for config reload scenarios
- [Phase 02-01]: Operations ordered: renames -> constraints -> column drops -> column adds -> triggers
- [Phase 02-01]: CHECK constraint includes expanded state machine: creating, provisioning, booting, running, stopping, terminated, error
- [Phase 02-01]: ON DELETE RESTRICT on instances prevents accidental cascade deletion of billing data
- [02-03]: Raw net/http for RunPod GraphQL client -- no library, matches project convention and RunPod CLI
- [02-03]: Functional ClientOption pattern (WithBaseURL, WithHTTPClient) for test injection
- [02-03]: EU region mapping uses EU-XX prefix format to match RunPod location strings
- [02-03]: bidPerGpu set to 0 for spot pods -- lets RunPod set market price
- [02-03]: Default Docker image runpod/pytorch:latest when none specified in ProvisionRequest
- [03-01]: AES-256-GCM with random 12-byte nonce prepended to ciphertext, hex-encoded for storage
- [03-01]: WG_ENCRYPTION_KEY validated as 64 hex chars with decoded bytes stored on Config struct
- [03-01]: CustomerInstance uses defense-by-omission: upstream fields structurally absent, not filtered
- [03-01]: Test for invalid AES key uses 15 bytes (not 16) since AES-128 with 16 bytes is valid
- [03-02]: Advisory lock constant 0x475055414950414D (GPUAIPAM) for deterministic lock ID
- [03-02]: IPAM receives pgx.Tx from caller; advisory lock is transaction-scoped and auto-releases
- [03-02]: Manager accepts WGClient and CommandRunner interfaces for full testability without root
- [03-02]: AddPeer rolls back WG peer if iptables DNAT/FORWARD setup fails
- [03-02]: RemovePeer treats iptables cleanup as best-effort (logs but does not fail)
- [03-02]: Port formula 10000 + ip[2]*256 + ip[3] maps full /16 to ports 10002-75535
- [03-03]: text/template used instead of html/template to avoid HTML-escaping bash characters
- [03-03]: Single-quoted heredocs prevent bash variable expansion; Go template expands before script runs
- [03-03]: SSH key validation combines format regex with shell injection character blocklist
- [03-03]: CallbackURL is pre-rendered by Go code (full URL), not constructed in bash
- [04-01]: Clerk SDK v2 RequireHeaderAuthorization used directly instead of custom JWT parsing
- [04-01]: Empty CLERK_SECRET_KEY returns 401 (not silent pass-through) for dev safety
- [04-01]: ClaimsFromContext wraps Clerk SessionClaimsFromContext mapping Subject->UserID and ActiveOrganizationID->OrgID
- [04-01]: Rate limiter uses sync.Map with limiterEntry wrapper for thread-safe lastSeen tracking
- [04-01]: Cursor format uses RFC3339 timestamp + pipe + ID, base64url encoded (no padding)
- [04-02]: State machine allows stopping from any non-terminal state and retry-terminate from error
- [04-02]: EnsureOrgAndUser uses clerk_org_id as org name placeholder for auto-creation
- [04-02]: EngineDeps struct injection matching ServerDeps pattern (not functional options)
- [04-02]: WireGuard is fully conditional in engine: nil wgMgr skips key gen, IPAM, and cloud-init
- [04-02]: Provider selection iterates registry with first-match (Phase 6 adds best-price)
- [04-02]: isDuplicateKeyError checks SQLState 23505 via errors.As for pgx compatibility
- [04-03]: InstanceResponse uses defense-by-omission: no provider fields exist in struct
- [04-03]: SSE max connection duration 30 minutes with client reconnect expected
- [04-03]: WriteTimeout set to 0 for SSE support; per-handler timeouts deferred to production
- [04-03]: Idempotency middleware uses SHA-256 body hash to detect key reuse with different bodies
- [04-03]: Internal ready callback is idempotent: returns 200 even if already transitioned
- [04-03]: Rate limiter: 10 req/s sustained with burst of 20 per org
- [04-04]: Resolve Clerk org ID at middleware layer, not DB layer -- keeps DB functions UUID-only
- [04-04]: WG cleanup is best-effort: errors logged but termination succeeds regardless
- [04-04]: Strip CIDR suffix from INET column values before net.ParseIP to handle PostgreSQL format
- [04.1-01]: All-or-nothing WG config: all three WG vars present together or all absent -- prevents misconfiguration
- [04.1-01]: WG init gated on WGEncryptionKeyBytes != nil (decoded bytes, not string) -- matches research Pattern 1
- [04.1-01]: IPAM hardcoded to 10.0.0.0/16 subnet -- matches existing IPAM tests and engine code
- [04.1-02]: AddPeer placed after IPAM allocation but before provider.Provision -- peer must exist on proxy before instance boots
- [04.1-02]: RemovePeer rollback uses wgPubKey (outer scope) not kp.PublicKey (inner scope) -- correct Go variable scoping
- [04.1-02]: dockerArgs only set when startupScript non-empty -- backward compatibility for tests without WG
- [04.2-01]: InstanceTokenAuth validates per-instance tokens via DB lookup, replacing LocalhostOnly + InternalAuthMiddleware on callback routes
- [04.2-01]: GpuctlPublicURL is optional config -- empty falls back to branded hostname for dev
- [04.2-01]: Redundant per-handler token checks removed -- middleware handles all auth
- [04.2-02]: buildCallbackURL extracted as package-level function for testability
- [04.2-02]: SetOnStatusChange setter avoids Engine<->Server circular dependency in main.go
- [04.2-02]: 5-second polling interval with 10-minute timeout balances responsiveness with resource usage
- [04.3-01]: Pattern 2 (Eager Creation) for idempotency middleware -- avoids pass-through gap where retries could create duplicates
- [04.3-01]: EnsureOrg extracted as standalone function -- reusable in both middleware and EnsureOrgAndUser
- [04.3-01]: NULLIF converts empty email to NULL in upsert queries -- cleaner than storing empty strings
- [04.3-01]: RETURNING user_id on user upsert -- gets internal UUID in single query round-trip
- [05-01]: ON DELETE RESTRICT for billing_sessions FKs to instances and organizations -- prevents accidental cascade deletion of billing records
- [05-01]: ON DELETE CASCADE for ssh_keys and spending_limits org FKs -- keys and limits should be cleaned up with the org
- [05-01]: stripe_reported_seconds column for delta-based Stripe usage metering -- avoids double-reporting
- [05-01]: Partial index on billing_sessions (WHERE ended_at IS NULL) for active session lookups
- [05-02]: SSH key validation uses golang.org/x/crypto/ssh ParseAuthorizedKey -- not regex -- for correctness
- [05-02]: Key type allowlist (ssh-rsa, ssh-ed25519, ecdsa-sha2-nistp*) -- safer than blocklist against unknown future types
- [05-02]: SSHKeyIDs made optional in CreateInstanceRequest -- engine layer handles fallback to user's keys via GetSSHKeysByUserID
- [05-02]: ErrDuplicateKey sentinel mapped from SQLSTATE 23505 -- same pattern as isDuplicateKeyError in idempotency.go
- [Phase 05]: Billing session creation non-fatal -- instance transitions to booting even if billing INSERT fails
- [Phase 05]: CloseBillingSession idempotent -- returns nil if no open session (safe for failed provisions)
- [Phase 05]: GetOrgMonthSpendCents uses cents (int64) to avoid float precision issues in spending limit checks
- [Phase 05]: BillingService no-op when Stripe not configured -- matches WireGuard optional service pattern
- [Phase 05]: Limits enforced BEFORE Stripe reporting in every tick -- prevents API latency from delaying protection
- [Phase 05]: StateStopped added to state machine -- stopped preserves storage but suspends billing
- [Phase 05]: 72h auto-terminate after limit reached -- stopped instances terminated after grace period
- [Phase 05]: Live spend check in checkSpendingLimit at provision time -- catches limit even if ticker hasn't run yet
- [05-05]: EnsureOrgAndUser for write endpoints, GetOrgIDByClerkOrgID for read-only -- consistent with SSH key handler patterns
- [05-05]: Period and start/end params mutually exclusive with 400 error on conflict
- [05-05]: Hourly aggregation walks sessions across hour boundaries for precise bucket distribution
- [06-01]: Redis single-key JSON array (gpu:offerings:all) instead of per-offering SCAN pattern -- atomic reads, no partial data
- [06-01]: 35s TTL on cache entries with 30s poll interval -- brief overlap prevents stale reads
- [06-01]: Markup pricing applied during cache write so customers never see wholesale prices
- [06-01]: AvailableOffering uses defense-by-omission: Provider field structurally absent
- [06-01]: org_id index includes created_at DESC for efficient REST catch-up endpoint queries
- [06-02]: sort.SliceStable for price sorting preserves registry iteration order as tiebreaker -- no explicit priority list needed
- [06-02]: Provider retry limited to provider.Provision call only -- WG setup happens once before retry loop
- [06-02]: Max 3 provision attempts across different providers to limit latency
- [06-02]: Price cap checked against cheapest candidate only (first in sorted list)
- [06-03]: Optimistic locking prevents duplicate event logging on concurrent status transitions
- [06-03]: Non-fatal event logging: failure to log event does not block primary operation
- [06-03]: Bounded concurrency (max 10) via semaphore channel for parallel provider checks
- [06-04]: Combined SSE and REST catch-up in single GET /api/v1/events endpoint (since= param switches mode)
- [06-04]: OrgEventBroker buffer size 20 for higher org-level event volume
- [06-04]: Health monitor created after API server to use srv.PublishOrgEvent callback directly
- [06-04]: Goroutine startup order: poller before server, monitor after server (dependency-driven)
- [08-01]: Geist font loaded via npm package (geist) not local woff2 files
- [08-01]: Component deletions pulled into Task 1 to unblock build (old components imported changed modules)
- [08-01]: Root .gitignore lib/ pattern extended with !frontend/src/lib/ exception
- [08-01]: Button component uses anchor element per Vercel CTA link pattern
- [08-02]: Only Navbar and UseCaseTabs use 'use client'; all other sections are server components
- [08-02]: UseCaseTabs heading added for section context
- [08-02]: FeaturePillars uses dot bullets to differentiate from UseCaseTabs checkmarks
- [07-01]: proxy.ts (not middleware.ts) for Next.js 16 hostname routing convention
- [07-01]: Template title pattern in root layout for flexible per-group page titles
- [07-01]: Marketing metadata (title, OG tags) moved to (marketing)/layout.tsx
- [07-02]: StatusBadge uses design system CSS variables (green-dim, purple-dim) with animate-pulse-dot for starting state
- [07-02]: DashboardSidebar preserves ?site=cloud query param via useSearchParams for local dev routing compatibility
- [07-02]: InstancesTable provides both desktop table and mobile card layout with responsive breakpoints
- [07-03]: Named proxy export preserved (clerkMiddleware assigned to export const proxy) for Next.js 16 convention
- [07-03]: ClerkProvider wraps outside html tag per Clerk docs for full auth context coverage
- [07-03]: Keyless mode for dev: empty CLERK env vars in .env.local let Clerk auto-generate temporary keys
- [07-04]: SWR refresh intervals tuned per page urgency: 10s instances, 30s availability, 60s billing
- [07-04]: LaunchInstanceForm reusable modal from both instances page and GPU availability table
- [07-04]: Server layout + client page pattern for metadata on SWR-powered routes
- [07-04]: Skeleton loading with bg-card-hover pulse animations matching design system
- [09-01]: Embedded frps via Go library maintaining single-binary architecture (not subprocess)
- [09-01]: FRP token auth (not OIDC) for simplicity in machine-to-machine auth
- [09-01]: Port range 10000-10255 matches existing WG port formula for consistency during migration
- [09-01]: Advisory lock ID 0x4650525054 (FRPPT) for port allocation serialization
- [09-01]: Empty FRP_TOKEN disables FRP tunneling (same optional pattern as WireGuard)
- [Phase 09]: FRPToken reuses per-instance internalToken for FRP auth
- [Phase 09]: FRP cleanup is automatic: frpc dies with instance, port freed by DB status change
- [Phase 09]: extractHost helper parses proxy hostname from GpuctlPublicURL using net/url.Parse
- [09-03]: WireGuard indirect deps (wireguard, wintun) remain as transitive FRP library dependencies -- not removable without removing FRP
- [09-03]: WG config env vars fully removed from Config struct and Load() validation
- [10-01]: Cloud layout converted to client component for mobile sidebar state sharing between sidebar and topbar
- [10-01]: Sidebar icons extracted to named components for readability and reuse
- [10-01]: Coming-soon pages use server components with static content (no client-side JS needed)
- [10-02]: Filters apply BEFORE grouping: flat offerings filtered by region/tier, then grouped into GPUCardData by gpu_model
- [10-02]: GPUCard launches cheapest available offering when Launch button clicked
- [10-02]: Launch modal pre-filled mode omits ssh_key_ids -- backend auto-attaches all user SSH keys
- [10-02]: GPU availability page converted to client component for GPUAvailabilityTable
- [Phase 10-03]: ConfirmDialog replaces window.confirm for all terminate actions
- [Phase 10-03]: SWR globalMutate used on detail page to invalidate list cache key for cross-page consistency
- [Phase 11]: Film grain moved from body::after to .film-grain::after -- dashboard renders grain-free
- [Phase 11]: btn-primary is solid purple with no gradient or glow -- clean Linear aesthetic
- [Phase 11]: Sidebar active state uses 2px left border instead of background highlight -- minimal visual weight
- [Phase 11]: Breadcrumb separator changed from / to > with 40% opacity -- subtler hierarchy
- [Phase 11-02]: useDebouncedCallback for search (not useDebounce hook) -- direct callback control
- [Phase 11-02]: Category chips as horizontal scrollable row -- quick visual scanning over dropdown
- [Phase 11-02]: Focus rings use border-light not purple -- subtle non-colored focus matching Linear
- [Phase 11-03]: CSS grid with display:contents on Link elements for clickable rows with proper column alignment
- [Phase 11-03]: SSH command column uses minmax(200px, 2fr) -- flexible width instead of truncated 260px
- [Phase 11-03]: Billing summary cards use standard border (no colored left accents) for Linear aesthetic
- [Phase 11-03]: Progress bars use bg-text-muted for under-70% state -- reduce purple in non-CTA contexts
- [Phase 12-01]: PricingTable already had mobile cards -- no changes needed
- [Phase 12-01]: Footer uses sm:grid-cols-2 lg:grid-cols-5 instead of direct md:grid-cols-5 jump
- [Phase 12-01]: NovacoreLogo negative margin reduced from -mr-12 to -mr-4 on mobile to prevent overflow

### Roadmap Evolution

- Phase 8 added: Rebuild frontend landing page to match Vercel homepage design
- Phase 9 added: Replace WireGuard with FRP tunneling
- Phase 10 added: Frontend dashboard with GPU availability, provisioning, and instance management
- Phase 11 added: Dashboard UI redesign — clean linear aesthetic, GPU catalog with categories and search, instances table fixes, professional polish
- Phase 12 added: Mobile responsive website optimization

### Pending Todos

None yet.

### Blockers/Concerns

- RunPod Docker initialization requires complete redesign from cloud-init (research finding)
- WireGuard NET_ADMIN capability in RunPod containers unverified -- must validate empirically in Phase 3
- REQUIREMENTS.md stated 55 total requirements but actual count is 65 -- corrected in traceability update

## Session Continuity

Last session: 2026-03-25T20:34:17.759Z
Stopped at: Completed 12-01-PLAN.md
Resume file: None
