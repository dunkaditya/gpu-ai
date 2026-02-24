---
phase: 02-provider-abstraction-runpod-adapter
plan: 02
subsystem: api
tags: [go, provider, interface, registry, concurrency]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: Go module structure and project conventions
provides:
  - Provider interface with 5-method contract (Name, ListAvailable, Provision, GetStatus, Terminate)
  - Normalized GPU types and instance tiers (on_demand, spot only)
  - Sentinel errors for provider failure modes (ErrNoCapacity, ErrProviderUnavailable, ErrInvalidGPUType)
  - Thread-safe provider registry with register, get, all, names operations
  - ProvisionRequest with InternalToken and CallbackURL for instance callbacks
  - PortMapping type for upstream instance port forwarding
affects: [02-03 RunPod adapter, 03 privacy-layer, 04 instance-lifecycle, 06 provisioning-engine]

# Tech tracking
tech-stack:
  added: []
  patterns: [sync.RWMutex for thread-safe registry, sentinel errors with errors.Is(), mock provider pattern for testing]

key-files:
  created:
    - internal/provider/provider.go
    - internal/provider/types.go
    - internal/provider/errors.go
    - internal/provider/registry.go
    - internal/provider/registry_test.go
  modified: []

key-decisions:
  - "Two tiers only for v1 (on_demand and spot) - TierReserved removed per CONTEXT.md"
  - "Async fire-and-return provisioning model - Provision returns upstream ID immediately, GetStatus polls separately"
  - "WireGuardPrivateKey removed from ProvisionRequest - key generation deferred to Phase 3"
  - "DatacenterLocation added to GPUOffering for drill-down placement info"
  - "Re-registration allowed in registry for config reload scenarios"

patterns-established:
  - "Provider interface: 5-method contract all GPU cloud adapters implement"
  - "Mock provider pattern: minimal struct implementing Provider for unit tests"
  - "Sentinel errors: package-level error vars with errors.Is() for typed failure handling"
  - "Registry pattern: sync.RWMutex-protected map with sorted Names() for startup logging"

requirements-completed: [PROV-01, PROV-02]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 2 Plan 02: Provider Interface & Registry Summary

**Clean 5-method provider interface with normalized GPU types, sentinel errors, and thread-safe registry with 4 passing tests**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T17:57:44Z
- **Completed:** 2026-02-24T17:59:40Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Refined Provider interface to match CONTEXT.md decisions (async provisioning, no cloud-init references)
- Updated type definitions: removed TierReserved, added DatacenterLocation, StockStatus, InternalToken, CallbackURL, StartupScript, PortMapping
- Created sentinel errors (ErrNoCapacity, ErrProviderUnavailable, ErrInvalidGPUType) for typed failure handling
- Implemented thread-safe provider registry with Register, Get, All, Names methods
- All 4 registry tests pass: register/get, all, names, re-register

## Task Commits

Each task was committed atomically:

1. **Task 1: Refine provider interface, types, and create error values** - `7fb949c` (feat)
2. **Task 2: Implement thread-safe provider registry with tests** - `ade93b5` (feat)

## Files Created/Modified
- `internal/provider/provider.go` - Provider interface with 5-method contract and updated doc comments
- `internal/provider/types.go` - Normalized GPU types, 2 tiers, updated request/result structs with new fields
- `internal/provider/errors.go` - Sentinel errors for ErrNoCapacity, ErrProviderUnavailable, ErrInvalidGPUType
- `internal/provider/registry.go` - Thread-safe provider registry using sync.RWMutex
- `internal/provider/registry_test.go` - 4 unit tests with mockProvider covering all registry operations

## Decisions Made
- Removed TierReserved per CONTEXT.md (only on_demand and spot for v1)
- Async fire-and-return provisioning model (Provision returns upstream ID, GetStatus polls)
- WireGuardPrivateKey removed from ProvisionRequest (key generation is Phase 3)
- DatacenterLocation and StockStatus added to GPUOffering for datacenter drill-down
- Registry allows re-registration (replaces existing provider by name) for config reload

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Provider interface ready for RunPod adapter implementation (Plan 02-03)
- Registry ready to wire providers at startup in main.go
- Sentinel errors ready for use in provisioning engine error handling
- All tests use mocks only (no real API calls or database connections)

---
*Phase: 02-provider-abstraction-runpod-adapter*
*Completed: 2026-02-24*
