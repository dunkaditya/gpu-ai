# Integration Testing Strategy

Lessons from manual end-to-end provisioning (RTX 5090 on RunPod, March 2026) and the plan for automated integration tests.

## 1. Bugs Found During Manual Testing

These are the real failures we hit when provisioning an RTX 5090 pod through the full `gpuctl provision` flow. Integration tests should target every one of these.

### Bug 1: Docker Image `runpod/pytorch:latest` Removed

- **Symptom**: Pod stuck in a boot loop. RunPod logs showed `error pulling image: manifest for runpod/pytorch:latest not found`.
- **Root cause**: RunPod removed the `latest` tag from their `runpod/pytorch` registry. Our default image constant pointed to a tag that no longer existed.
- **Fix**: Changed default to a pinned tag: `runpod/pytorch:2.4.0-py3.11-cuda12.4.1-devel-ubuntu22.04` (`internal/provider/runpod/adapter.go:25`).
- **Test target**: Mock provider should verify that the image name sent in the create-pod payload is a pinned tag (not `:latest`). Real-provider smoke test catches if the pinned tag itself is removed.

### Bug 2: Missing `booting → running` Polling Transition

- **Symptom**: CLI showed `status=booting` indefinitely even though RunPod reported the pod as `RUNNING`.
- **Root cause**: The `progressStatus` goroutine handled `provisioning → booting` but then returned early, never polling again for the `booting → running` transition.
- **Fix**: Changed the goroutine to continue polling after `booting` until the pod reaches `running` (with IP and port populated).
- **Secondary issue**: We also transitioned to `running` as soon as RunPod said `RUNNING`, but before networking was actually ready (no IP assigned yet). Fixed by gating `booting → running` on the provider returning a non-empty IP.
- **Test target**: Mock provider should return `running` status without IP first, then with IP on next poll. Assert the engine stays in `booting` until IP is present.

### Bug 3: Postgres INET Column Scan Error

- **Symptom**: `can't scan into dest[5] (col: upstream_ip): cannot scan inet (OID 869) in binary format into **string`.
- **Root cause**: The `upstream_ip` column is Postgres `INET` type, but the Go struct field is `*string`. pgx can't scan `INET` binary format into a Go string directly.
- **Fix**: Cast in the SELECT query: `upstream_ip::TEXT` (`internal/db/instances.go:47`). The write path uses `$2::INET` to store it properly.
- **Test target**: Mock integration test with real Postgres should write an IP via `UpdateInstanceStatus`, then read it back via `GetInstanceByID`. Verifies the INET↔string round-trip works.

### Bug 4: SSH Key Env Var Name (`SSH_PUBLIC_KEYS` vs `PUBLIC_KEY`)

- **Symptom**: SSH to the pod prompted for a password despite SSH keys being passed.
- **Root cause**: We injected keys as the `SSH_PUBLIC_KEYS` env var, but RunPod's pytorch images only read `PUBLIC_KEY` for authorized key injection.
- **Fix**: Changed env var key from `SSH_PUBLIC_KEYS` to `PUBLIC_KEY` (`internal/provider/runpod/adapter.go:244`).
- **Test target**: Mock integration test should assert the GraphQL mutation payload contains `PUBLIC_KEY` (not `SSH_PUBLIC_KEYS`). Real-provider smoke test should verify passwordless SSH works.

## 2. Two-Layer Test Strategy

| Layer | Trigger | What It Tests | Runtime | Cost |
|-------|---------|---------------|---------|------|
| Mock integration tests | Every push (GitHub Actions CI) | Engine logic, DB queries, state machine transitions, provider adapter contract | ~30s | Free |
| Real-provider smoke test | Daily cron (GitHub Actions) | Actual RunPod API, Docker image availability, SSH connectivity, `nvidia-smi` | ~5 min | ~$0.10/run |

The mock layer catches code regressions immediately. The smoke layer catches provider API drift, image removal, networking changes — the kind of bugs that broke us during manual testing.

## 3. Mock Integration Test Design

Engine-level tests using **real Postgres + real Redis + stateful mock provider**. No HTTP layer — test the `provision.Engine` directly.

### Test infrastructure

```
testcontainers (or GitHub Actions services):
  - postgres:16      (real schema via migrations)
  - redis:7          (real caching)

mock provider:
  - In-memory stateful adapter implementing provider.Provider
  - Simulates: provisioning → running (with configurable delays)
  - Can inject failures: API errors, slow transitions, missing IPs
```

### Key test cases

```go
// TestProvisionHappyPath
// Seed user + SSH key → engine.Provision() → poll until running
// Assert: DB instance status=running, upstream_ip is set, SSH port > 0

// TestProvisionDockerImage
// Verify the create-pod payload uses pinned image tag, not :latest

// TestBootingToRunningRequiresIP
// Mock returns status=running but empty IP → assert engine stays in booting
// Mock returns status=running with IP → assert engine transitions to running

// TestSSHKeyInjection
// Provision with 2 SSH keys → assert mock received PUBLIC_KEY env var
// with both keys newline-joined

// TestINETRoundTrip
// Write upstream_ip via UpdateInstanceStatus → read via GetInstanceByID
// Assert the IP string survives the INET cast round-trip

// TestProvisionFailure
// Mock provider returns error on create → assert instance status=failed
// Assert error is logged and propagated

// TestTermination
// Provision → running → engine.Terminate() → assert provider.Terminate called
// Assert DB status=terminated
```

### Structure

```
internal/integration/          # Integration test package (build tag: integration)
  engine_test.go               # Tests above
  mock_provider.go             # Stateful mock implementing provider.Provider
  testutil.go                  # DB setup, migration runner, cleanup
```

Run with: `go test -tags=integration ./internal/integration/ -v`

## 4. Real-Provider Smoke Test Design

A GitHub Actions cron workflow that provisions a real GPU, verifies it works, and terminates it. Essentially automating the manual test session we ran.

### Workflow

```yaml
# .github/workflows/smoke-test.yml
name: GPU Smoke Test
on:
  schedule:
    - cron: '0 6 * * *'    # Daily at 6am UTC
  workflow_dispatch: {}      # Manual trigger

jobs:
  smoke:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    env:
      RUNPOD_API_KEY: ${{ secrets.RUNPOD_API_KEY }}
      DATABASE_URL: ${{ secrets.SMOKE_DATABASE_URL }}
      REDIS_URL: ${{ secrets.SMOKE_REDIS_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }

      - name: Build gpuctl
        run: go build -o gpuctl ./cmd/gpuctl

      - name: Seed test user
        run: |
          psql "$DATABASE_URL" -f tools/seed_smoke_test.sql

      - name: Provision GPU
        id: provision
        run: |
          # Use cheapest available GPU, spot tier
          ./gpuctl provision \
            --gpu-type=rtx_4090 \
            --gpu-count=1 \
            --tier=spot \
            --user-id=$SMOKE_USER_ID \
            --org-id=$SMOKE_ORG_ID \
            --yes \
            --json | tee provision.json
          echo "instance_id=$(jq -r .instance_id provision.json)" >> "$GITHUB_OUTPUT"
          echo "ssh_host=$(jq -r .ssh_host provision.json)" >> "$GITHUB_OUTPUT"
          echo "ssh_port=$(jq -r .ssh_port provision.json)" >> "$GITHUB_OUTPUT"

      - name: Verify SSH + nvidia-smi
        run: |
          ssh -o StrictHostKeyChecking=no \
              -i ~/.ssh/smoke_test_key \
              -p ${{ steps.provision.outputs.ssh_port }} \
              root@${{ steps.provision.outputs.ssh_host }} \
              'nvidia-smi && echo "SMOKE_TEST_OK"'

      - name: Terminate instance
        if: always()
        run: |
          ./gpuctl terminate ${{ steps.provision.outputs.instance_id }} --yes

      - name: Alert on failure
        if: failure()
        # Slack/PagerDuty webhook — provider API probably changed
        run: echo "Smoke test failed — check RunPod API or image availability"
```

### What this catches

- RunPod API breaking changes or auth issues
- Docker image tag removal (the `pytorch:latest` bug)
- SSH key injection failures (the `PUBLIC_KEY` bug)
- Networking issues (pod running but no IP assigned)
- GPU driver problems (nvidia-smi failure)

### Cost control

- Use spot tier (cheapest, ~$0.20/hr for RTX 4090)
- 10-minute timeout kills stuck pods
- `if: always()` on terminate step prevents leaked pods
- Budget alert on RunPod account for safety

## 5. Infrastructure Note

**K8s is unnecessary for v1.** The production stack is:

- 1 VM running `gpuctl` (API server + WireGuard proxy endpoint)
- Managed Postgres + Redis (or co-located on the VM for v1)

That's it. RunPod pods tunnel back to the single WireGuard endpoint. All customer SSH traffic routes through there. There's nothing to orchestrate — it's one binary and a WireGuard interface.

A VM with systemd provides process restarts and health checks. No container orchestration needed.

### When K8s would make sense

- **Multiple WireGuard proxy nodes across regions** — if customers in EU/Asia need low-latency SSH, you'd run proxy servers in each region. K8s simplifies deploying identical nodes across regions.
- **High availability** — if the single `gpuctl` process dying means all SSH sessions drop.
- **Horizontal scaling** — if thousands of concurrent instances overload one API server.

Until then, multi-region could also be solved with DNS-based routing + a few VMs running the same binary — still no K8s required. The trigger for K8s adoption is when managing 3+ proxy nodes manually becomes painful.
