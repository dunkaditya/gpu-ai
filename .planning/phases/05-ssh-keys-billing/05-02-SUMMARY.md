---
phase: 05-ssh-keys-billing
plan: 02
subsystem: api
tags: [ssh-keys, golang-x-crypto, fingerprint, crud, provisioning]

# Dependency graph
requires:
  - phase: 05-ssh-keys-billing
    provides: "v5 schema with ssh_keys.org_id column and UNIQUE(org_id, fingerprint)"
provides:
  - "SSH key CRUD API: POST/GET/DELETE /api/v1/ssh-keys"
  - "SSH key DB methods: CreateSSHKey, ListSSHKeysByOrg, DeleteSSHKey, GetSSHKeysByUserID, CountSSHKeysByOrg"
  - "SSH key validation using golang.org/x/crypto/ssh (RSA, Ed25519, ECDSA accepted; DSA rejected)"
  - "Provisioning smart default: empty ssh_key_ids auto-includes user's keys"
  - "Org 50-key limit enforcement"
affects: [05-03, 05-04, 05-05]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "SSH key validation via golang.org/x/crypto/ssh ParseAuthorizedKey + FingerprintSHA256"
    - "Smart default pattern: optional field with auto-fill fallback in engine layer"
    - "Duplicate key detection via pgx SQLState 23505 mapped to domain sentinel error"

key-files:
  created:
    - "internal/api/handlers_ssh_keys.go"
  modified:
    - "internal/db/ssh_keys.go"
    - "internal/api/server.go"
    - "internal/api/handlers.go"
    - "internal/provision/engine.go"

key-decisions:
  - "EnsureOrgAndUser used in create handler (needs both orgID and userID), GetOrgIDByClerkOrgID used in list/delete (only needs orgID)"
  - "SSH key validation uses golang.org/x/crypto/ssh ParseAuthorizedKey -- not regex -- for correctness"
  - "Key type allowlist (not blocklist) -- only ssh-rsa, ssh-ed25519, ecdsa-sha2-nistp{256,384,521} accepted"
  - "SSHKeyIDs made optional in CreateInstanceRequest -- engine layer handles fallback to user's keys"

patterns-established:
  - "Domain sentinel errors (ErrDuplicateKey) mapped from DB-level errors (SQLSTATE 23505)"
  - "Smart provisioning defaults: optional request fields with engine-level auto-fill"

requirements-completed: [SSHK-01, SSHK-02, SSHK-03, SSHK-04, API-06]

# Metrics
duration: 3min
completed: 2026-02-25
---

# Phase 5 Plan 02: SSH Key CRUD API Summary

**SSH key CRUD with golang.org/x/crypto validation, org-scoped 50-key limit, and provisioning smart default for auto-including user keys**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-25T17:50:37Z
- **Completed:** 2026-02-25T17:53:59Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Full SSH key CRUD API: POST (add with validation), GET (list by org), DELETE (scoped to org) endpoints
- SSH key validation using golang.org/x/crypto/ssh: accepts RSA/Ed25519/ECDSA, rejects DSA and others
- Smart provisioning default: empty ssh_key_ids in POST /instances auto-includes all of the creating user's keys
- Org 50-key limit with 422 response, duplicate fingerprint detection with 409 response

## Task Commits

Each task was committed atomically:

1. **Task 1: SSH key DB methods and validation** - `2e84a36` (feat)
2. **Task 2: SSH key handlers, routes, and provisioning smart default** - `edf2fe3` (feat)

## Files Created/Modified
- `internal/db/ssh_keys.go` - Added OrgID to SSHKey struct, ErrDuplicateKey sentinel, CreateSSHKey/ListSSHKeysByOrg/DeleteSSHKey/GetSSHKeysByUserID/CountSSHKeysByOrg methods, updated GetSSHKeysByIDs
- `internal/api/handlers_ssh_keys.go` - SSH key HTTP handlers: create (with validation, fingerprinting, limit check), list, delete
- `internal/api/server.go` - Wired POST/GET/DELETE /api/v1/ssh-keys routes with authChain
- `internal/api/handlers.go` - Made SSHKeyIDs optional in CreateInstanceRequest, removed required validation
- `internal/provision/engine.go` - Smart default: GetSSHKeysByUserID fallback when SSHKeyIDs empty

## Decisions Made
- Used EnsureOrgAndUser for create handler (needs userID for key ownership), GetOrgIDByClerkOrgID for list/delete (only needs orgID for scoping)
- Key type validation uses allowlist approach (explicitly listing accepted types) rather than blocklist -- safer against unknown future key types
- SSHKeyIDs field made optional at handler validation level, with enforcement pushed to provisioning engine layer -- cleaner separation of concerns
- Fingerprint computed server-side using ssh.FingerprintSHA256 (SHA256:base64 format) -- consistent regardless of client

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- SSH key CRUD complete, ready for billing engine (05-03) and spending limits (05-04)
- All three SSH key endpoints registered and fully functional
- Provisioning engine updated with smart default for SSH keys
- go build ./... and go test ./... pass cleanly

## Self-Check: PASSED

- FOUND: internal/db/ssh_keys.go
- FOUND: internal/api/handlers_ssh_keys.go
- FOUND: internal/api/server.go
- FOUND: internal/api/handlers.go
- FOUND: internal/provision/engine.go
- FOUND: .planning/phases/05-ssh-keys-billing/05-02-SUMMARY.md
- FOUND: commit 2e84a36
- FOUND: commit edf2fe3

---
*Phase: 05-ssh-keys-billing*
*Completed: 2026-02-25*
