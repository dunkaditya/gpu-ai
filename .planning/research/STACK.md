# Stack Research

**Domain:** GPU Cloud Aggregation Platform
**Researched:** 2026-02-24
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.24.x | Backend language, single binary | Latest stable (Feb 2025). Swiss Table maps for 2-3% CPU reduction, post-quantum crypto, tool directives in go.mod. Stdlib net/http with 1.22+ routing patterns eliminates need for any router framework. |
| Go stdlib `net/http` | 1.22+ patterns | HTTP server, routing, middleware | Go 1.22 added method matching (`GET /path/{id}`) and wildcard path variables to ServeMux. This is sufficient for the entire API surface. No framework needed. |
| PostgreSQL | 16+ | Primary data store | ACID compliance for billing/instance records. Mature pgcrypto for UUID generation. JSONB for flexible provider metadata. Proven at scale for SaaS platforms. |
| Redis | 7.x | Availability cache, ephemeral data | Sub-millisecond reads for GPU availability polling. TTL-based expiry (60s) matches the 30s poll interval naturally. Lightweight, zero-config for this use case. |
| Next.js | 16.x | Customer dashboard frontend | Latest stable (Oct 2025). App Router with React Server Components, Turbopack dev server, built-in React Compiler. First-class Clerk integration. |
| WireGuard | userspace/kernel | Privacy tunnel layer | Kernel-level performance, simple peer model (public/private key pairs), built-in keepalive. Industry standard for encrypted overlay networks. |

### Go Libraries

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| `github.com/jackc/pgx/v5` | v5.7.5 | PostgreSQL driver + connection pool | The de facto Go Postgres driver. pgxpool provides production-grade connection pooling. Native support for LISTEN/NOTIFY, COPY, and pgcrypto types. Use v5.7.5 (stable) rather than v5.8.0 which requires Go 1.24 and drops PG12 support -- use v5.8.0 only if Go 1.24 minimum is acceptable. |
| `github.com/redis/go-redis/v9` | v9.18.0 | Redis client | Official Redis Go client. v9 supports RESP3, connection pooling, pipelining, OpenTelemetry integration. Maintained by Redis org. v9.18.0 is latest stable (Feb 2026), includes Redis 8.6 support. |
| `github.com/stripe/stripe-go/v84` | v84.3.0 | Stripe billing SDK | Official Stripe Go SDK. v84 uses Billing Meters API (replaces deprecated usage records). Supports per-second metering via meter events. 1,000 events/sec in live mode. |
| `github.com/clerk/clerk-sdk-go/v2` | v2.5.1 | Clerk JWT auth | Official Clerk Go SDK. v2 provides JWT verification, JWKS caching, session claims extraction. Integrates with http.Handler middleware pattern. |
| `golang.zx2c4.com/wireguard/wgctrl` | latest | WireGuard peer management | Official WireGuard Go control library. Programmatic peer add/remove/configure across Linux, BSD, Windows. Cross-platform device detection. |
| `golang.org/x/crypto/curve25519` | latest | WireGuard key generation | Standard Go crypto library for Curve25519 key pairs. Use X25519 function with Basepoint (not deprecated ScalarBaseMult). |
| `log/slog` | stdlib | Structured logging | Go 1.21+ stdlib. JSON handler for production, text handler for dev. LogValuer interface for sensitive field redaction. No external dependency needed. |

### Frontend Libraries

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| `@clerk/nextjs` | latest | Auth UI components + middleware | Official Clerk Next.js SDK. ClerkProvider, SignIn/SignUp components, clerkMiddleware() for route protection. 30-minute setup to production auth. |
| `shadcn/ui` | latest (CLI 3.0) | UI component library | Not a package -- copy-paste components built on Radix UI + Tailwind. Full ownership of code, no version lock-in. Dashboard templates widely available. |
| Tailwind CSS | v4.x | Utility-first CSS | CSS-first configuration (no JS config file). 3-10x faster builds vs v3. Auto-scanning, @theme directive. shadcn/ui fully supports v4. |
| TanStack Query | v5 | Server state management | Automatic caching, background refetch, optimistic updates for GPU availability polling. Replaces manual fetch/state management. |
| Zustand | v5 | Client state management | Minimal API for local UI state (sidebar toggle, modal state). 1KB bundle. Avoid Redux/Context for simple client state. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| golangci-lint v2.7+ | Go linting | Meta-linter running 50+ linters. v2.x supports Go 1.24. Configure via `.golangci.yml`. |
| Turbopack | Next.js dev bundler | Built into Next.js 16. Filesystem caching now stable. Replaces webpack for dev. |
| Docker Compose | Local dev environment | PostgreSQL + Redis containers. Single `docker compose up` for dependencies. |
| `tools/migrate.py` | DB migrations | Python runner for SQL migration files. Keeps migration logic simple and auditable. |

## Go Module Dependencies

```bash
# go.mod should declare:
go 1.24.0

# Core dependencies
go get github.com/jackc/pgx/v5@v5.7.5
go get github.com/redis/go-redis/v9@v9.18.0
go get github.com/stripe/stripe-go/v84@v84.3.0
go get github.com/clerk/clerk-sdk-go/v2@v2.5.1
go get golang.zx2c4.com/wireguard/wgctrl@latest
go get golang.org/x/crypto@latest
```

## Frontend Installation

```bash
# Initialize Next.js 16 project
npx create-next-app@latest frontend --typescript --tailwind --app --turbopack

# Auth
npm install @clerk/nextjs

# State management
npm install @tanstack/react-query zustand

# UI (shadcn is CLI-based, not an npm install)
npx shadcn@latest init
npx shadcn@latest add button card table dialog sheet sidebar chart
```

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| HTTP Router | Go stdlib `net/http` | chi, Gin, Echo | Go 1.22+ ServeMux has method matching and path vars. Chi adds middleware chaining sugar but project constraint is "no frameworks." Gin/Echo are full frameworks with reflect-based routing -- unnecessary overhead and coupling. |
| Postgres Driver | pgx v5 | database/sql + lib/pq | lib/pq is in maintenance mode. pgx is faster, supports native types (UUID, INET, JSONB), has built-in connection pooling (pgxpool), and offers LISTEN/NOTIFY for future real-time features. |
| Redis Client | go-redis v9 | redigo | redigo is lower-level, requires manual connection pooling, lacks pipeline/context support. go-redis v9 is the official Redis client with RESP3, pipelining, and OpenTelemetry built in. |
| Auth | Clerk | Auth0, Supabase Auth, DIY JWT | Clerk has first-class Go SDK and Next.js SDK. Auth0 is more expensive at scale. Supabase Auth ties you to Supabase ecosystem. DIY JWT is weeks of work for auth, MFA, social login, org management. |
| Billing | Stripe | Paddle, LemonSqueezy | Stripe's Billing Meters API supports per-second usage metering natively. Paddle/LemonSqueezy are merchant-of-record (simpler tax but less control). GPU billing needs per-second granularity. |
| Frontend Framework | Next.js 16 | Remix, SvelteKit, plain React | Next.js has the best Clerk integration. React Server Components reduce client bundle for dashboard. Huge ecosystem of shadcn dashboard templates. Vercel deployment is trivial. |
| CSS | Tailwind v4 | CSS Modules, styled-components | shadcn/ui requires Tailwind. Utility-first is faster for dashboard development. v4 is CSS-first with massive perf improvements. |
| State (server) | TanStack Query v5 | SWR, manual fetch | TanStack Query has better devtools, mutation support, and optimistic updates. SWR is simpler but lacks advanced cache invalidation patterns needed for instance lifecycle updates. |
| State (client) | Zustand | Redux, Jotai, Context | Redux is overkill. Context causes re-renders. Zustand is 1KB, works outside React, persists to localStorage trivially. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| GORM / any Go ORM | ORMs hide query complexity, generate unpredictable SQL, and make billing-critical queries harder to audit. GPU instance billing must be exact. | Raw SQL with pgx. Write queries in `internal/db/*.go`. Full control over every query. |
| chi / Gin / Echo | Project constraint is stdlib-only. These add unnecessary abstraction over Go 1.22+ ServeMux which now handles method routing and path params natively. | `net/http.ServeMux` with 1.22+ patterns like `GET /api/v1/instances/{id}`. |
| lib/pq | Maintenance mode since 2023. No active development. Missing native type support. | pgx v5 -- actively maintained, faster, better type system. |
| redigo | No context support, manual pool management, no pipelining convenience. | go-redis v9 -- official client, modern API. |
| Next.js Pages Router | Legacy routing model. No React Server Components. Clerk and shadcn docs now default to App Router. | Next.js App Router (default since Next.js 13.4, mature in 16). |
| Tailwind v3 | Requires JS config file, slower builds, no CSS-first theming. shadcn CLI 3.0 defaults to v4. | Tailwind v4 -- CSS-first, auto-scanning, 3-10x faster. |
| NextAuth.js / Auth.js | Requires building auth UI, session management, and social login from scratch. No org management. More code, more bugs. | Clerk -- managed auth with prebuilt components, org support, Go SDK. |
| `encoding/json` for high-throughput | stdlib json is slow for hot paths. GPU availability API serving cached JSON to many clients. | Use stdlib for now. If profiling shows bottleneck, swap to `github.com/goccy/go-json` or `github.com/bytedance/sonic` as drop-in replacement. |

## Stack Patterns by Use Case

**For the RunPod adapter (GraphQL + REST):**
- Use stdlib `net/http.Client` for HTTP calls to RunPod API
- For GraphQL mutations (pod creation), use raw HTTP POST with JSON body -- do NOT add a GraphQL client library for one provider
- RunPod is migrating to REST API; prefer REST endpoints where available
- Confidence: MEDIUM (RunPod API is actively evolving)

**For WireGuard key generation:**
- Use `golang.org/x/crypto/curve25519.X25519(scalar, Basepoint)` to derive public keys
- Generate private keys with `crypto/rand.Read(privateKey[:])` (32 bytes)
- Apply WireGuard key clamping: `privateKey[0] &= 248; privateKey[31] &= 127; privateKey[31] |= 64`
- Confidence: HIGH (standard WireGuard protocol)

**For WireGuard peer management:**
- Use `wgctrl.New()` to get a client, then `client.ConfigureDevice()` to add/remove peers
- This manages the WireGuard interface on the proxy server programmatically
- Alternative: shell out to `wg set` commands if wgctrl gives trouble on the target OS
- Confidence: MEDIUM (wgctrl works well on Linux, less tested on other platforms)

**For Stripe per-second billing:**
- Create a Billing Meter with `sum` aggregation for GPU-seconds
- Send meter events via `POST /v2/billing/meter_events` with customer ID, value (seconds), timestamp
- Rate limit: 1,000 events/sec live mode. Batch meter events per-instance per-minute to stay well under limit
- Use Stripe Checkout for initial payment method capture
- Confidence: HIGH (Stripe Meters API is GA, documented)

**For Clerk JWT middleware:**
- Use `clerk.Verify()` from clerk-sdk-go/v2/jwt to verify tokens
- Extract `clerk.SessionClaims` for user_id and org_id
- Cache JWKS (JSON Web Key Set) -- SDK does this automatically
- Token comes from `Authorization: Bearer <token>` header (cross-origin from Next.js dashboard)
- Confidence: HIGH (Clerk Go SDK v2 is stable, well-documented)

**For Next.js dashboard:**
- App Router with server components for static dashboard shell
- Client components for real-time GPU availability (TanStack Query polling)
- Clerk `<ClerkProvider>` wrapping app, `clerkMiddleware()` protecting routes
- shadcn/ui sidebar layout with collapsible navigation
- Confidence: HIGH (well-established patterns)

## Version Compatibility Matrix

| Package | Go Version | PostgreSQL | Redis | Notes |
|---------|------------|------------|-------|-------|
| pgx v5.7.5 | Go 1.21+ | PG 12+ | -- | Safe for Go 1.22-1.24 |
| pgx v5.8.0 | Go 1.24+ | PG 13+ | -- | Only if targeting Go 1.24 minimum |
| go-redis v9.18.0 | Go 1.21+ | -- | Redis 6+ | RESP3 requires Redis 6+. Includes Redis 8.6 support. |
| stripe-go v84 | Go 1.15+ | -- | -- | Very permissive Go version requirement |
| clerk-sdk-go v2.5.1 | Go 1.21+ | -- | -- | Requires Go 1.21+ for slog compatibility |
| wgctrl | Go 1.20+ | -- | -- | Linux kernel WireGuard preferred for production |
| Next.js 16 | -- | -- | -- | Requires Node.js 18.18+ |
| Tailwind v4 | -- | -- | -- | Requires PostCSS via @tailwindcss/postcss |

## Go Module Version Decision

**Recommendation: Use `go 1.24` in go.mod.**

Rationale:
- Go 1.24 was released Feb 2025, now one year old and well-proven
- Enables tool directives in go.mod (replace tools.go hack for golangci-lint)
- Swiss Table maps give free 2-3% CPU improvement
- pgx v5.8.0 can be used (latest), but v5.7.5 also works
- All other dependencies support Go 1.24

If keeping Go 1.22 minimum (as currently in go.mod), use pgx v5.7.5 and lose tool directives. The project constraint says "Go 1.22+" so either works, but 1.24 is strictly better for new projects.

## Sources

- [pgx GitHub](https://github.com/jackc/pgx) -- v5.7.5/v5.8.0 version tags confirmed via GitHub
- [go-redis pkg.go.dev](https://pkg.go.dev/github.com/redis/go-redis/v9) -- v9.18.0 latest confirmed (Feb 16, 2026)
- [go-redis releases](https://github.com/redis/go-redis/releases) -- Release history and changelog
- [stripe-go releases](https://github.com/stripe/stripe-go/releases) -- v84.3.0 stable confirmed (Jan 2025)
- [clerk-sdk-go releases](https://github.com/clerk/clerk-sdk-go/releases) -- v2.5.1 confirmed (Jan 2025)
- [Clerk Go JWT docs](https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2/jwt) -- Verify function API
- [wgctrl-go GitHub](https://github.com/WireGuard/wgctrl-go) -- ConfigureDevice API
- [Go 1.24 release notes](https://go.dev/doc/go1.24) -- Feb 2025 release, Swiss Tables, tool directives
- [Go 1.22 routing enhancements](https://go.dev/blog/routing-enhancements) -- Method matching, wildcards in ServeMux
- [Stripe Billing Meters API](https://docs.stripe.com/api/billing/meter) -- GA metering API
- [Stripe usage-based billing guide](https://docs.stripe.com/billing/subscriptions/usage-based) -- Implementation patterns
- [Next.js 16 blog post](https://nextjs.org/blog/next-16) -- Oct 2025 release
- [Tailwind CSS v4](https://tailwindcss.com/docs/compatibility) -- Jan 2025 release
- [shadcn/ui changelog](https://ui.shadcn.com/docs/changelog) -- CLI 3.0, Tailwind v4 support
- [RunPod REST API blog](https://www.runpod.io/blog/runpod-rest-api-gpu-management) -- New REST API alongside GraphQL
- [Go slog official blog](https://go.dev/blog/slog) -- Structured logging patterns
- [golangci-lint releases](https://github.com/golangci/golangci-lint/releases) -- v2.7.2, Go 1.24 support

---
*Stack research for: GPU Cloud Aggregation Platform*
*Researched: 2026-02-24*
