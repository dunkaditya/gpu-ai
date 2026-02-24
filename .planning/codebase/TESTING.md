# Testing Patterns

**Analysis Date:** 2026-02-24

## Test Framework

**Runner:**
- Standard Go `testing` package (`testing.T`)
- Run via: `go test ./...`

**Assertion Library:**
- Standard library only (no assertion framework imported)
- Manual error checking: `if !condition { t.Error(...) }` or `t.Fatalf(...)`

**Run Commands:**
```bash
go test ./...              # Run all tests in all packages
go test ./internal/...     # Run tests for internal packages only
go test -v ./...           # Run with verbose output
go test -race ./...        # Run with race detector
go test -cover ./...       # Run with coverage analysis
make test                  # Run via Makefile (equivalent to go test ./...)
```

**Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test File Organization

**Location:**
- Co-located with source files (same package, same directory)
- Test file naming: `{module}_test.go` (e.g., `adapter_test.go` in `internal/provider/runpod/`)

**Naming:**
- Test functions: `Test{FunctionName}(t *testing.T)`
- Example: `TestListAvailable()`, `TestProvision()`, `TestGetStatus()`, `TestTerminate()`

**Structure:**
```
internal/provider/runpod/
├── adapter.go           # Implementation
└── adapter_test.go      # Tests
```

## Test File Pattern

**Current State:**
Test file currently contains commented-out test stubs with documentation of what should be tested:

From `internal/provider/runpod/adapter_test.go`:
```go
package runpod

// TODO: Implement RunPod adapter tests:
//
// func TestListAvailable(t *testing.T)
//   - Mock RunPod GraphQL API response
//   - Verify correct mapping of GPU types
//   - Verify Secure Cloud → on_demand, Community Cloud → spot
//   - Verify price and availability fields
//
// func TestProvision(t *testing.T)
//   - Mock RunPod create pod API
//   - Verify cloud-init script is included
//   - Verify correct GPU type and count in request
//   - Verify ProvisionResult fields
//
// func TestGetStatus(t *testing.T)
//   - Mock RunPod pod status API
//   - Verify status mapping (RUNNING → running, etc.)
//
// func TestTerminate(t *testing.T)
//   - Mock RunPod terminate API
//   - Verify success response
//   - Verify error handling for unknown pod
```

## Test Structure Pattern

**Expected Format (from TODO documentation):**
```go
func TestListAvailable(t *testing.T) {
    // Setup: Mock API responses
    // Execute: Call method under test
    // Assert: Verify outputs and behaviors
}

func TestProvision(t *testing.T) {
    // Setup: Create request object
    // Execute: Call Provision()
    // Assert: Verify ProvisionResult fields
}
```

**Conventions to Follow:**
- Test functions are methods on `*testing.T`
- First argument is always `t *testing.T`
- Test name immediately follows package `func Test`
- No test suites yet (flat function style)

## Mocking Strategy

**Framework:** Standard library `net/http/httptest` expected

**Mocking HTTP APIs:**
- Use `httptest.NewServer()` for HTTP mocks
- Capture requests and verify them
- Return predictable responses for testing

**Patterns from TODOs:**
```
- Mock RunPod GraphQL API response
- Verify cloud-init script is included
- Verify correct GPU type and count in request
```

**What to Mock:**
- HTTP API calls (RunPod, Stripe, etc.)
- External provider adapters
- Database queries (optional: can use real DB for integration tests)

**What NOT to Mock:**
- Internal functions (unit test the full call chain)
- Standard library functions
- Go primitives and data structures

## Test Data & Fixtures

**Test Data Approach:**
- Inline test data in test functions (not yet established pattern)
- Factory functions for complex objects (expected pattern)

**Expected Pattern:**
```go
func TestProvision(t *testing.T) {
    req := provider.ProvisionRequest{
        InstanceID:           "inst-123",
        GPUType:              provider.GPUTypeH100SXM,
        GPUCount:             1,
        Tier:                 provider.TierOnDemand,
        Region:               "us-east-1",
        SSHPublicKeys:        []string{"ssh-rsa AAAA..."},
        WireGuardPrivateKey:  "WG_PRIVATE_KEY",
        WireGuardAddress:     "10.0.0.5",
    }
    // ... test assertions
}
```

**Fixtures Location:**
- Test helper functions in same `_test.go` file
- No separate fixture files detected or required yet

## Coverage Requirements

**Requirements:** Not enforced (no CI/CD pipeline visible)

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Current Status:**
- No tests implemented yet
- Coverage: 0% (all packages are TODO stubs)
- Expected coverage targets: TBD per team standards

## Test Types

**Unit Tests:**
- Scope: Individual functions/methods
- Approach: Mock external dependencies (APIs, databases)
- Example: `TestListAvailable()` mocks RunPod GraphQL
- Framework: Standard `testing.T`

**Integration Tests:**
- Scope: Multiple components working together
- Approach: Use real database or test database
- Example: Full provisioning flow (engine → provider → database)
- Not yet structured; would co-locate with unit tests

**End-to-End Tests:**
- Framework: Not used yet
- Would test full API workflows with real/staging provider accounts
- Separate from unit/integration tests

**Database Tests:**
- Expected approach: Use test database or in-memory equivalent
- Run with `go test ./internal/db/...`
- Should test query builders and result mapping

## Testing Checklist (from Codebase)

**For RunPod Adapter Tests (`internal/provider/runpod/adapter_test.go`):**
```
✗ func TestListAvailable(t *testing.T)
  ✗ Mock RunPod GraphQL API response
  ✗ Verify correct mapping of GPU types
  ✗ Verify Secure Cloud → on_demand, Community Cloud → spot
  ✗ Verify price and availability fields

✗ func TestProvision(t *testing.T)
  ✗ Mock RunPod create pod API
  ✗ Verify cloud-init script is included
  ✗ Verify correct GPU type and count in request
  ✗ Verify ProvisionResult fields

✗ func TestGetStatus(t *testing.T)
  ✗ Mock RunPod pod status API
  ✗ Verify status mapping (RUNNING → running, etc.)

✗ func TestTerminate(t *testing.T)
  ✗ Mock RunPod terminate API
  ✗ Verify success response
  ✗ Verify error handling for unknown pod
```

## Common Test Patterns

**Async Testing:**
- Use `context.WithTimeout()` for timeout testing
- Example from codebase architecture:
  ```go
  shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  ```

**Error Testing:**
- Test both success and error cases
- Verify error types and messages
- Mocking should return errors to test error paths

**Concurrency Testing:**
- Use `go test -race ./...` to detect race conditions
- No explicit synchronization primitives expected in tests (use channels)

## Test Execution

**Command Line:**
```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific package
go test ./internal/provider/runpod

# Run verbose with coverage
go test -v -cover ./...

# Watch mode (using external tool)
# go run github.com/cosmtrek/air@latest (or similar)
```

**CI/CD:**
- `.golangci-lint` linting in Makefile: `make lint`
- No explicit CI pipeline detected (GitHub Actions expected)

---

*Testing analysis: 2026-02-24*
