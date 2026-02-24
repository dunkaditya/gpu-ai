# Phase 3: Privacy Layer - Context

**Gathered:** 2026-02-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete WireGuard-based privacy infrastructure: key generation, peer management, tunnel IP allocation, cloud-init template rendering, and ensuring upstream provider details never reach the customer. Covers PRIV-01 through PRIV-08.

Phase 3 builds the privacy plumbing. Phase 4 wires it into the API and instance lifecycle.

</domain>

<decisions>
## Implementation Decisions

### WireGuard Proxy Topology
- Single central proxy server — one WireGuard endpoint that all GPU instances connect to
- All-in-one component: the proxy is both the WireGuard endpoint AND the SSH router (no separate services)
- Port mapping approach: `gpu-xxxx.gpu.ai:22` -> `proxy:100XX` -> WireGuard -> `10.0.0.X:22`
- Wildcard DNS: `*.gpu.ai` resolves to the proxy IP
- When WireGuard manager adds a peer, it also adds the iptables/routing rule; when it removes a peer, it removes the route — one component, one source of truth
- WireGuard private keys stored encrypted (AES-256-GCM) in the instances table, encryption key from env var

### IPAM & Subnet Design
- 10.0.0.0/16 subnet, sequential increment (last assigned + 1)
- 10.0.0.1 reserved for the proxy server itself
- `wg_address` column directly on the instances table with UNIQUE constraint
- Allocation uses `SELECT MAX(wg_address) + 1` with Postgres advisory lock to prevent races
- No immediate reuse on termination — terminated instance records keep their wg_address; reclaim only when approaching exhaustion
- Can expand to 10.0.0.0/8 (16M addresses) later if needed — just a WireGuard AllowedIPs config change

### Init Template & Bootstrap
- Go `text/template` for rendering, template embedded in the binary via `embed`
- Cloud-init script installs and configures:
  1. **WireGuard client** — install package, write wg0.conf with server peer, bring up tunnel
  2. **SSH keys + firewall lockdown** — inject customer SSH keys, firewall allows SSH only over WireGuard tunnel (block direct SSH on public IP), disable root password auth, remove any backdoor/provider SSH keys (no persistent root access)
  3. **Branded hostname** — set hostname to `gpu-xxxx.gpu.ai`, configure /etc/hosts, set MOTD with GPU.ai branding
  4. **Provider scrubbing** — block metadata endpoint (iptables drop 169.254.169.254), remove provider env vars, clean /etc/motd of provider branding, remove provider CLI tools
  5. **NVIDIA driver verification** — confirm GPU is visible and drivers loaded (`nvidia-smi`), fail if GPU not detected
  6. **Disable unattended upgrades** — prevent auto-updates and auto-reboots that could interrupt training
- Ready callback: instance POSTs to `/internal/instances/{id}/ready` with `internal_token` auth and `gpu_info` from `nvidia-smi` in the body
- Unit test the template renderer in Phase 3 — render with sample inputs, verify output contains correct WG config, SSH keys, firewall rules, etc.

### Privacy Filtering (API)
- Customer-facing API response structs defined in Phase 3 — structurally exclude upstream fields (defense by omission, not middleware stripping)
- If the field doesn't exist in the response type, it can never leak
- Separate internal types (with upstream_provider, upstream_id) vs customer types (only id, hostname, status, gpu_type, region)
- Full instance scrub on the instance itself (handled by cloud-init above)
- iptables drop to 169.254.169.254 only (not full link-local range, to avoid breaking networking)

### Claude's Discretion
- Exact WireGuard configuration parameters (keepalive, MTU, etc.)
- AES-256-GCM encryption implementation details for key storage
- Cloud-init script ordering and error handling within the script
- Exact iptables rule syntax and chain placement
- Template variable naming conventions

</decisions>

<specifics>
## Specific Ideas

- Port mapping pattern: `gpu-4a7f.gpu.ai:22` -> `proxy:10022` -> WireGuard -> `10.0.0.5:22`
- Ready callback should include GPU info in body for hardware verification logging:
  ```bash
  GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader 2>/dev/null)
  curl -s -X POST "https://api.gpu.ai/internal/instances/${INSTANCE_ID}/ready" \
      -H "Authorization: Bearer ${INTERNAL_TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"gpu_info\": \"${GPU_INFO}\"}"
  ```
- 5-minute timeout for stuck `creating` instances — background check marks as `error` and terminates upstream (this is a Phase 4 lifecycle concern, but the timeout contract is decided here)
- Billing starts at callback time (`billing_start = NOW()` when ready received), not at provisioning time

</specifics>

<deferred>
## Deferred Ideas

- **Audit logging** — log every provisioning event, termination, SSH key change, and API call in a separate table. SOC 2 readiness. (Own phase)
- **Encryption at rest** — managed Postgres with TDE or encryption enabled. Infrastructure/deployment decision.
- **Per-region proxy servers** — deploy WireGuard proxies in each region for lower latency. Future scaling concern.

</deferred>

---

*Phase: 03-privacy-layer*
*Context gathered: 2026-02-24*
