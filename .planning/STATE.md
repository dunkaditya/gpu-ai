---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: in-progress
last_updated: "2026-02-25T20:22:45Z"
progress:
  total_phases: 8
  completed_phases: 8
  total_plans: 28
  completed_plans: 26
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-24)

**Core value:** Customers can find available GPUs across providers and provision them instantly through a single interface, with a privacy layer that completely hides the upstream provider.
**Current focus:** Phase 6: Availability & Health Monitoring

## Current Position

Phase: 6 (Availability & Health Monitoring)
Plan: 2 of 4 in current phase
Status: In Progress
Last activity: 2026-02-25 -- Completed 06-02 (Availability API & best-price selection)

Progress: [█████████░] 93%

## Performance Metrics

**Velocity:**
- Total plans completed: 26
- Average duration: 2.0min
- Total execution time: 0.80 hours

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
| 06-availability-health-monitoring | 2 | 5min | 2.5min |

**Recent Trend:**
- Last 5 plans: 05-03 (4min), 05-04 (3min), 05-05 (2min), 06-01 (2min), 06-02 (3min)
- Trend: stable

*Updated after each plan completion*
| Phase 05 P03 | 4min | 2 tasks | 4 files |
| Phase 05 P04 | 3min | 2 tasks | 8 files |
| Phase 05 P05 | 2min | 2 tasks | 2 files |
| Phase 06 P01 | 2min | 2 tasks | 6 files |
| Phase 06 P02 | 3min | 2 tasks | 3 files |

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

### Pending Todos

None yet.

### Blockers/Concerns

- RunPod Docker initialization requires complete redesign from cloud-init (research finding)
- WireGuard NET_ADMIN capability in RunPod containers unverified -- must validate empirically in Phase 3
- REQUIREMENTS.md stated 55 total requirements but actual count is 65 -- corrected in traceability update

## Session Continuity

Last session: 2026-02-25
Stopped at: Completed 06-02-PLAN.md (Availability API & best-price selection)
Resume file: None
