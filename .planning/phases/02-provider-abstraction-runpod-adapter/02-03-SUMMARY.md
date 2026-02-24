---
phase: 02-provider-abstraction-runpod-adapter
plan: 03
subsystem: api
tags: [go, runpod, graphql, provider, adapter, retry, mapping]

# Dependency graph
requires:
  - phase: 02-provider-abstraction-runpod-adapter
    provides: Provider interface with 5-method contract, GPU types, sentinel errors, registry
provides:
  - RunPod GraphQL HTTP client with retry logic and exponential backoff
  - RunPod adapter implementing all 5 Provider interface methods
  - GPU name mapping table (10 models) with bidirectional lookup
  - Region normalization from RunPod location strings to GPU.ai region codes
  - 12 unit tests with httptest mock servers (zero real API calls)
  - RunPodAPIKey optional config field
affects: [03-privacy-layer, 04-instance-lifecycle, 06-provisioning-engine]

# Tech tracking
tech-stack:
  added: []
  patterns: [GraphQL-over-HTTP with raw net/http, retry with exponential backoff, functional options for test injection, httptest mock servers]

key-files:
  created:
    - internal/provider/runpod/client.go
    - internal/provider/runpod/mapping.go
    - internal/provider/runpod/adapter.go
    - internal/provider/runpod/adapter_test.go
  modified:
    - internal/config/config.go

key-decisions:
  - "Raw net/http for GraphQL client (no library) matching project convention and RunPod CLI pattern"
  - "Functional ClientOption pattern (WithBaseURL, WithHTTPClient) for test injection"
  - "EU region mapping uses EU-XX prefix format to match RunPod location strings (EU-RO, EU-CZ)"
  - "bidPerGpu set to 0 for spot pods (lets RunPod set market price)"
  - "Default Docker image runpod/pytorch:latest when none specified"

patterns-established:
  - "GraphQL client pattern: Client.do() with response envelope handling, Client.doWithRetry() with backoff"
  - "httptest mock server pattern: setupTestAdapter() helper for all adapter tests"
  - "contextWithDefaultTimeout: applies default timeout only if context has no deadline"
  - "Bidirectional GPU name map: gpuNameMap for normalization, reverseGPUNameMap for provisioning"

requirements-completed: [PROV-03, PROV-04, PROV-05, PROV-06]

# Metrics
duration: 4min
completed: 2026-02-24
---

# Phase 2 Plan 03: RunPod Adapter Summary

**RunPod GraphQL adapter with retry logic, GPU/region normalization, on-demand and spot provisioning, and 12 httptest-based unit tests**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-24T18:02:25Z
- **Completed:** 2026-02-24T18:06:44Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Implemented RunPod GraphQL HTTP client with retry/backoff (3 attempts, exponential 1s/2s/4s)
- Full Provider interface implementation: ListAvailable, Provision (on-demand + spot), GetStatus, Terminate
- GPU name mapping covering 10 models with bidirectional lookup for provisioning reverse-mapping
- Region normalization handling RunPod location formats (US-TX-3, EU-RO-1, CA-MTL-1)
- 12 unit tests all using httptest mock servers -- zero real API calls
- Config updated with optional RunPodAPIKey field

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement RunPod GraphQL client and name mapping tables** - `17f0754` (feat)
2. **Task 2: Implement RunPod adapter and unit tests** - `314a135` (feat)

## Files Created/Modified
- `internal/provider/runpod/client.go` - GraphQL HTTP client with do/doWithRetry, retry logic, query/mutation constants
- `internal/provider/runpod/mapping.go` - GPU name map (10 models), reverse map, region normalization
- `internal/provider/runpod/adapter.go` - Full Provider interface implementation with all 5 methods
- `internal/provider/runpod/adapter_test.go` - 12 unit tests with httptest mock servers
- `internal/config/config.go` - Added optional RunPodAPIKey field

## Decisions Made
- Used raw net/http for GraphQL (matching project convention and RunPod CLI pattern)
- Functional options pattern (WithBaseURL, WithHTTPClient) enables clean test injection
- EU region map entries use EU-XX prefix format (EU-RO, EU-CZ) to match RunPod's location string format
- Spot bid set to 0 (lets RunPod set market price; availability poller in Phase 6 will refine)
- Default Docker image is runpod/pytorch:latest when caller doesn't specify

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed EU region mapping prefix collision**
- **Found during:** Task 2 (unit tests)
- **Issue:** Region map had "RO", "CZ" etc. as standalone entries, but RunPod locations use "EU-RO-1" format. Prefix matching hit "EU" (eu-west) before checking the country-specific entry.
- **Fix:** Added EU-prefixed entries (EU-RO, EU-CZ, EU-BG, EU-SE, EU-NO, EU-IS) to regionMap so longest-prefix matching works correctly.
- **Files modified:** internal/provider/runpod/mapping.go
- **Verification:** TestNormalizeRegion passes with EU-RO-1 -> eu-east
- **Committed in:** 314a135 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug fix)
**Impact on plan:** Minor mapping table fix to handle actual RunPod location format. No scope creep.

## Issues Encountered
None

## User Setup Required
None - RunPodAPIKey is optional and loaded from RUNPOD_API_KEY env var when present. All tests use httptest mock servers.

## Next Phase Readiness
- RunPod adapter ready for integration into provisioning engine (Phase 4/6)
- Provider registry can wire RunPod adapter at startup when RUNPOD_API_KEY is configured
- All sentinel errors (ErrNoCapacity, ErrInvalidGPUType) properly returned and testable
- Tests validate full request/response contract via httptest -- safe for CI

## Self-Check: PASSED

- All 5 created/modified files verified present on disk
- Both task commits (17f0754, 314a135) verified in git log
- `go test ./internal/provider/runpod/` -- 12/12 tests PASS
- `go test ./internal/provider/` -- 4/4 tests PASS (registry tests unaffected)
- `go build ./...` compiles cleanly

---
*Phase: 02-provider-abstraction-runpod-adapter*
*Completed: 2026-02-24*
