---
phase: 03-privacy-layer
plan: 01
subsystem: wireguard
tags: [wireguard, aes-256-gcm, encryption, wgctrl-go, privacy, types]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: "Config struct, migration runner, database schema v0/v1"
  - phase: 02-provider-abstraction-runpod-adapter
    provides: "Provider types (GPUOffering, ProvisionRequest, ProvisionResult)"
provides:
  - "WireGuard key pair generation via wgctrl-go"
  - "AES-256-GCM encryption/decryption for private key storage"
  - "KeyPair type for WireGuard keys"
  - "CustomerInstance type excluding upstream provider details"
  - "Instance type with ToCustomer() conversion"
  - "Schema migration for wg_private_key_enc and wg_address UNIQUE"
  - "Config fields for WG encryption key, proxy endpoint, proxy public key, interface name"
affects: [03-privacy-layer, 04-instance-lifecycle]

# Tech tracking
tech-stack:
  added: [golang.zx2c4.com/wireguard/wgctrl, crypto/aes, crypto/cipher]
  patterns: [AES-256-GCM with random nonce prepended to ciphertext, defense-by-omission response types, hex-encoded encryption key from env]

key-files:
  created:
    - database/migrations/20250224_v2_privacy_layer.sql
    - internal/wireguard/types.go
    - internal/wireguard/keygen.go
    - internal/wireguard/keygen_test.go
  modified:
    - internal/config/config.go
    - internal/provider/types.go
    - .env.example
    - go.mod
    - go.sum

key-decisions:
  - "AES-256-GCM with random 12-byte nonce prepended to ciphertext, hex-encoded for storage"
  - "WG_ENCRYPTION_KEY validated as 64 hex chars with decoded bytes stored on Config struct"
  - "CustomerInstance uses defense-by-omission: upstream fields structurally absent, not filtered"
  - "Test for invalid AES key uses 15 bytes (not 16) since AES-128 with 16 bytes is valid"

patterns-established:
  - "Defense by omission: customer-facing types exclude sensitive fields at the struct level"
  - "Hex-encoded encryption keys from environment with validation in config loader"
  - "Nonce prepended to ciphertext for self-contained encrypted blobs"

requirements-completed: [PRIV-01, PRIV-08]

# Metrics
duration: 2min
completed: 2026-02-24
---

# Phase 3 Plan 1: Privacy Layer Foundation Summary

**WireGuard key generation via wgctrl-go with AES-256-GCM encryption, privacy-enforcing CustomerInstance types, and schema migration for encrypted key storage**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-24T22:12:27Z
- **Completed:** 2026-02-24T22:15:08Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- WireGuard key pair generation using wgctrl-go library (no shell-out)
- AES-256-GCM encryption/decryption with random nonces for private key storage
- CustomerInstance type that structurally excludes upstream provider, ID, and IP fields
- Schema migration adding wg_private_key_enc column and wg_address UNIQUE constraint
- Config validation for 64-char hex encryption key with decoded byte storage
- 9 comprehensive tests covering keygen, encryption round-trip, nonce uniqueness, error cases, and upstream field exclusion

## Task Commits

Each task was committed atomically:

1. **Task 1: Schema migration and config updates for privacy layer** - `5446912` (feat)
2. **Task 2: WireGuard key generation, AES-256-GCM encryption, and privacy types with tests** - `76cbc11` (feat)

## Files Created/Modified
- `database/migrations/20250224_v2_privacy_layer.sql` - Privacy layer schema: wg_private_key_enc column, wg_address UNIQUE
- `internal/wireguard/types.go` - KeyPair struct for WireGuard key pairs
- `internal/wireguard/keygen.go` - GenerateKeyPair, EncryptPrivateKey, DecryptPrivateKey functions
- `internal/wireguard/keygen_test.go` - 9 tests covering keygen, encryption, and privacy types
- `internal/config/config.go` - WGEncryptionKey (hex-validated), WGProxyEndpoint, WGProxyPublicKey, WGInterfaceName
- `internal/provider/types.go` - CustomerInstance, Instance types, ToCustomer() conversion
- `.env.example` - WG_ENCRYPTION_KEY, WG_PROXY_ENDPOINT, WG_PROXY_PUBLIC_KEY, WG_INTERFACE_NAME
- `go.mod` / `go.sum` - Added wgctrl-go and transitive dependencies

## Decisions Made
- AES-256-GCM with random 12-byte nonce prepended to ciphertext, hex-encoded for database storage
- WG_ENCRYPTION_KEY validated as exactly 64 hex characters; decoded bytes stored as WGEncryptionKeyBytes on Config struct
- CustomerInstance uses defense-by-omission pattern: upstream fields are structurally absent from the type, not filtered by middleware
- Test for invalid AES key length uses 15 bytes instead of plan's 16 bytes, since AES-128 (16 bytes) is a valid key size

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed TestEncryptWithInvalidKeyLength using 15-byte key instead of 16**
- **Found during:** Task 2 (keygen tests)
- **Issue:** Plan specified 16-byte key expecting aes.NewCipher error, but AES-128 accepts 16-byte keys
- **Fix:** Changed test to use 15-byte key which is truly invalid for AES
- **Files modified:** internal/wireguard/keygen_test.go
- **Verification:** Test passes, correctly expects error from aes.NewCipher with invalid key length
- **Committed in:** 76cbc11 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug fix)
**Impact on plan:** Necessary correction for test correctness. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required. The new WG_* environment variables are documented in .env.example but not needed for test execution.

## Next Phase Readiness
- Key generation and encryption primitives ready for peer management (Plan 02)
- CustomerInstance type ready for API response serialization (Plan 03 / Phase 04)
- Schema migration ready for deployment
- Config fields ready for cloud-init template rendering

---
*Phase: 03-privacy-layer*
*Completed: 2026-02-24*
