# Phase 3: Privacy Layer - Research

**Researched:** 2026-02-24
**Domain:** WireGuard VPN infrastructure, key management, IPAM, cloud-init templating, privacy filtering
**Confidence:** HIGH

## Summary

Phase 3 builds the privacy plumbing that hides all upstream provider details from the customer. This involves five core subsystems: (1) WireGuard key generation using the `wgctrl-go` library, (2) programmatic peer management on the proxy server via `wgctrl.Client.ConfigureDevice`, (3) PostgreSQL-backed IPAM for tunnel IP allocation from the 10.0.0.0/16 subnet, (4) cloud-init template rendering using Go's `text/template` with `embed`, and (5) customer-facing response types that structurally exclude upstream provider fields.

The Go ecosystem provides mature, official libraries for every piece: `wgctrl-go` (official WireGuard project) handles key generation and peer configuration, Go stdlib `crypto/aes` + `crypto/cipher` handles AES-256-GCM encryption for key storage, Go stdlib `text/template` + `embed` handles cloud-init rendering, and Go stdlib `net` handles IP address manipulation for IPAM. No third-party frameworks are needed beyond what the project already uses (pgx for Postgres, go-redis for Redis).

The biggest technical risk is WireGuard availability on RunPod GPU instances. RunPod runs Docker containers, not full VMs, and NET_ADMIN capability is unverified. The fallback is `wireguard-go` (userspace implementation) which requires no kernel module. The cloud-init template must account for this by attempting kernel WireGuard first, then falling back to `wireguard-go`. This concern was already flagged in STATE.md and must be addressed during implementation.

**Primary recommendation:** Use `golang.zx2c4.com/wireguard/wgctrl` for all WireGuard operations (key generation + peer management), Go stdlib for everything else (AES-GCM, template rendering, IPAM IP math), and PostgreSQL advisory locks for race-free IP allocation.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Single central proxy server -- one WireGuard endpoint that all GPU instances connect to
- All-in-one component: the proxy is both the WireGuard endpoint AND the SSH router (no separate services)
- Port mapping approach: `gpu-xxxx.gpu.ai:22` -> `proxy:100XX` -> WireGuard -> `10.0.0.X:22`
- Wildcard DNS: `*.gpu.ai` resolves to the proxy IP
- When WireGuard manager adds a peer, it also adds the iptables/routing rule; when it removes a peer, it removes the route -- one component, one source of truth
- WireGuard private keys stored encrypted (AES-256-GCM) in the instances table, encryption key from env var
- 10.0.0.0/16 subnet, sequential increment (last assigned + 1)
- 10.0.0.1 reserved for the proxy server itself
- `wg_address` column directly on the instances table with UNIQUE constraint
- Allocation uses `SELECT MAX(wg_address) + 1` with Postgres advisory lock to prevent races
- No immediate reuse on termination -- terminated instance records keep their wg_address; reclaim only when approaching exhaustion
- Can expand to 10.0.0.0/8 (16M addresses) later if needed
- Go `text/template` for rendering, template embedded in the binary via `embed`
- Cloud-init script installs and configures: WireGuard client, SSH keys + firewall lockdown, branded hostname, provider scrubbing, NVIDIA driver verification, disable unattended upgrades
- Ready callback: instance POSTs to `/internal/instances/{id}/ready` with `internal_token` auth and `gpu_info` from `nvidia-smi` in the body
- Unit test the template renderer in Phase 3
- Customer-facing API response structs defined in Phase 3 -- structurally exclude upstream fields (defense by omission, not middleware stripping)
- Separate internal types (with upstream_provider, upstream_id) vs customer types (only id, hostname, status, gpu_type, region)
- iptables drop to 169.254.169.254 only (not full link-local range, to avoid breaking networking)

### Claude's Discretion
- Exact WireGuard configuration parameters (keepalive, MTU, etc.)
- AES-256-GCM encryption implementation details for key storage
- Cloud-init script ordering and error handling within the script
- Exact iptables rule syntax and chain placement
- Template variable naming conventions

### Deferred Ideas (OUT OF SCOPE)
- Audit logging -- log every provisioning event, termination, SSH key change, and API call in a separate table. SOC 2 readiness. (Own phase)
- Encryption at rest -- managed Postgres with TDE or encryption enabled. Infrastructure/deployment decision.
- Per-region proxy servers -- deploy WireGuard proxies in each region for lower latency. Future scaling concern.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| PRIV-01 | WireGuard key pairs generated for each new instance | `wgctrl/wgtypes.GeneratePrivateKey()` + `Key.PublicKey()` -- pure Go, cryptographically secure, no shell-out needed |
| PRIV-02 | WireGuard peers added to proxy server programmatically via wgctrl-go | `wgctrl.Client.ConfigureDevice()` with `wgtypes.PeerConfig` -- official API, supports add with AllowedIPs |
| PRIV-03 | WireGuard peers removed from proxy server on instance termination | `wgctrl.Client.ConfigureDevice()` with `PeerConfig{Remove: true}` -- atomic removal by public key |
| PRIV-04 | IPAM allocates unique WireGuard addresses from subnet pool backed by PostgreSQL | `SELECT MAX(wg_address) + 1` with `pg_advisory_xact_lock` for transaction-scoped race prevention; `net.IP` arithmetic for increment |
| PRIV-05 | Instance init template renders with WireGuard config, SSH keys, hostname, firewall rules | Go `text/template` + `embed` -- template compiled once at startup, rendered per-instance with struct data |
| PRIV-06 | Customer SSH connections route through WireGuard proxy with branded hostname | Port mapping iptables rules added by WireGuard manager alongside peer; hostname set by cloud-init |
| PRIV-07 | Upstream provider identity (name, IP, env vars, metadata endpoint) hidden from customer | Cloud-init: metadata endpoint blocked (iptables drop 169.254.169.254), provider env vars removed, MOTD scrubbed, provider CLI tools removed |
| PRIV-08 | All customer-facing API responses structurally exclude upstream provider details | Separate Go struct types: `CustomerInstance` (safe fields only) vs `Instance` (all fields); upstream fields physically absent from customer types |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `golang.zx2c4.com/wireguard/wgctrl` | v0.0.0-20241231184526 | WireGuard device control (peer add/remove/list) | Official WireGuard project library; controls both kernel and userspace WG devices |
| `golang.zx2c4.com/wireguard/wgctrl/wgtypes` | (same module) | Key generation, type definitions (Key, PeerConfig, Config) | Part of wgctrl-go; `GeneratePrivateKey()` + `Key.PublicKey()` for pure-Go keygen |
| Go stdlib `crypto/aes` + `crypto/cipher` | Go 1.24 | AES-256-GCM encryption for WireGuard private key storage | Stdlib; no third-party crypto needed. GCM provides authenticated encryption |
| Go stdlib `crypto/rand` | Go 1.24 | Cryptographically secure random bytes for nonces and tokens | Stdlib; used for AES-GCM nonces and internal_token generation |
| Go stdlib `text/template` | Go 1.24 | Cloud-init script rendering | Stdlib; `text/template` (not `html/template`) since output is bash, not HTML |
| Go stdlib `embed` | Go 1.24 | Embed cloud-init template in binary | Stdlib; `//go:embed` directive compiles template into binary, no external file deps |
| Go stdlib `net` | Go 1.24 | IP address parsing, manipulation, subnet math | Stdlib; `net.IP`, `net.IPNet`, `net.ParseCIDR` for IPAM operations |
| `github.com/jackc/pgx/v5` | v5.8.0 | PostgreSQL advisory locks for IPAM race prevention | Already in go.mod; `pg_advisory_xact_lock` via raw SQL through pgx |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib `encoding/base64` | Go 1.24 | Base64 encode/decode WireGuard keys for storage | Keys are 32-byte arrays; base64 for DB/template string representation |
| Go stdlib `encoding/hex` | Go 1.24 | Hex encode AES-GCM ciphertext for DB storage | Nonce+ciphertext stored as hex string in instances.wg_private_key_enc (to be re-added) |
| Go stdlib `os/exec` | Go 1.24 | Execute iptables commands for port mapping rules | Only for iptables rule management on the proxy server |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `wgctrl-go` programmatic API | Shell out to `wg` command | Shell commands are fragile, harder to test, require `wg` binary on system |
| AES-256-GCM via stdlib | `nacl/secretbox` (NaCl) | NaCl is simpler but AES-GCM is the locked decision and is industry standard |
| Go stdlib `net.IP` arithmetic | `go-cidr` library | Extra dependency for simple sequential increment; stdlib is sufficient |
| `text/template` with `embed` | External template file | Embed bakes template into binary; no missing-file deployment risk |

**Installation:**
```bash
go get golang.zx2c4.com/wireguard/wgctrl@latest
```

This is the only new dependency. Everything else is Go stdlib or already in go.mod.

## Architecture Patterns

### Recommended Project Structure
```
internal/wireguard/
    keygen.go           # GenerateKeyPair(), EncryptPrivateKey(), DecryptPrivateKey()
    manager.go          # Manager struct: AddPeer(), RemovePeer(), ListPeers()
    ipam.go             # IPAM struct: AllocateAddress(), ReleaseAddress() (DB-backed)
    types.go            # KeyPair, Peer, IPAllocation types
    template.go         # TemplateRenderer: Render() with embedded cloud-init
    template_test.go    # Unit tests for template rendering
    keygen_test.go      # Unit tests for key generation + encryption
    ipam_test.go        # Unit tests for IPAM (needs test DB or mocked queries)
    manager_test.go     # Unit tests for Manager (interface-based testing)

internal/provider/
    types.go            # Existing -- add CustomerInstance response type here

infra/cloud-init/
    bootstrap.sh.tmpl   # Renamed from bootstrap.sh -- Go template syntax
```

### Pattern 1: WireGuard Key Generation
**What:** Generate a WireGuard key pair using wgctrl/wgtypes, return as base64 strings
**When to use:** Every new instance provisioning
**Example:**
```go
// Source: https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl/wgtypes
import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

type KeyPair struct {
    PrivateKey string // base64-encoded
    PublicKey  string // base64-encoded
}

func GenerateKeyPair() (*KeyPair, error) {
    privKey, err := wgtypes.GeneratePrivateKey()
    if err != nil {
        return nil, fmt.Errorf("generate private key: %w", err)
    }
    pubKey := privKey.PublicKey()
    return &KeyPair{
        PrivateKey: privKey.String(), // base64-encoded
        PublicKey:  pubKey.String(),  // base64-encoded
    }, nil
}
```
**Confidence:** HIGH -- `wgtypes.GeneratePrivateKey()` and `Key.PublicKey()` are documented in official pkg.go.dev docs and the `Key.String()` method returns base64.

### Pattern 2: AES-256-GCM Key Encryption
**What:** Encrypt WireGuard private keys before storing in the database
**When to use:** Before DB write (encrypt) and after DB read (decrypt)
**Example:**
```go
// Source: https://pkg.go.dev/crypto/cipher#NewGCM
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "io"
)

func EncryptPrivateKey(plaintext string, encryptionKey []byte) (string, error) {
    block, err := aes.NewCipher(encryptionKey) // encryptionKey must be 32 bytes
    if err != nil {
        return "", fmt.Errorf("create cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("create GCM: %w", err)
    }
    nonce := make([]byte, gcm.NonceSize()) // 12 bytes for standard GCM
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("generate nonce: %w", err)
    }
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil) // prepend nonce
    return hex.EncodeToString(ciphertext), nil
}

func DecryptPrivateKey(ciphertextHex string, encryptionKey []byte) (string, error) {
    data, err := hex.DecodeString(ciphertextHex)
    if err != nil {
        return "", fmt.Errorf("hex decode: %w", err)
    }
    block, err := aes.NewCipher(encryptionKey)
    if err != nil {
        return "", fmt.Errorf("create cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("create GCM: %w", err)
    }
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("decrypt: %w", err)
    }
    return string(plaintext), nil
}
```
**Confidence:** HIGH -- Standard Go crypto pattern, verified against official `crypto/cipher` docs and multiple community examples.

### Pattern 3: Programmatic Peer Management via wgctrl
**What:** Add/remove WireGuard peers on the proxy server without shelling out
**When to use:** Instance creation (add peer) and termination (remove peer)
**Example:**
```go
// Source: https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl
import (
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
    "net"
    "time"
)

type Manager struct {
    client        *wgctrl.Client
    interfaceName string // e.g., "wg0"
}

func (m *Manager) AddPeer(publicKeyBase64 string, tunnelIP net.IP) error {
    pubKey, err := wgtypes.ParseKey(publicKeyBase64)
    if err != nil {
        return fmt.Errorf("parse public key: %w", err)
    }
    keepalive := 25 * time.Second
    cfg := wgtypes.Config{
        Peers: []wgtypes.PeerConfig{{
            PublicKey:                   pubKey,
            ReplaceAllowedIPs:          true,
            AllowedIPs:                 []net.IPNet{{IP: tunnelIP, Mask: net.CIDRMask(32, 32)}},
            PersistentKeepaliveInterval: &keepalive,
        }},
    }
    return m.client.ConfigureDevice(m.interfaceName, cfg)
}

func (m *Manager) RemovePeer(publicKeyBase64 string) error {
    pubKey, err := wgtypes.ParseKey(publicKeyBase64)
    if err != nil {
        return fmt.Errorf("parse public key: %w", err)
    }
    cfg := wgtypes.Config{
        Peers: []wgtypes.PeerConfig{{
            PublicKey: pubKey,
            Remove:   true,
        }},
    }
    return m.client.ConfigureDevice(m.interfaceName, cfg)
}
```
**Confidence:** HIGH -- `ConfigureDevice` with `PeerConfig` is the primary documented API in pkg.go.dev. `Remove: true` is explicitly documented for peer removal.

### Pattern 4: PostgreSQL-Backed IPAM with Advisory Lock
**What:** Allocate sequential tunnel IPs from 10.0.0.0/16, preventing race conditions
**When to use:** During instance provisioning, before WireGuard peer setup
**Example:**
```go
// Uses pgx transaction with pg_advisory_xact_lock
func (ipam *IPAM) AllocateAddress(ctx context.Context, tx pgx.Tx) (net.IP, error) {
    // Transaction-scoped advisory lock (auto-released on commit/rollback)
    _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", ipamLockID)
    if err != nil {
        return nil, fmt.Errorf("acquire advisory lock: %w", err)
    }

    var maxAddr string
    err = tx.QueryRow(ctx,
        "SELECT COALESCE(MAX(host(wg_address)::inet), '10.0.0.1') FROM instances",
    ).Scan(&maxAddr)
    if err != nil {
        return nil, fmt.Errorf("query max address: %w", err)
    }

    nextIP := incrementIP(net.ParseIP(maxAddr))
    // Validate still in subnet
    if !ipam.subnet.Contains(nextIP) {
        return nil, fmt.Errorf("IPAM exhausted: next IP %s outside subnet %s", nextIP, ipam.subnet)
    }
    return nextIP, nil
}
```
**Confidence:** HIGH -- `pg_advisory_xact_lock` is standard PostgreSQL; pgx supports it natively. IP increment is simple byte arithmetic.

### Pattern 5: Embedded Template Rendering
**What:** Compile cloud-init template into binary, render per-instance
**When to use:** During provisioning to generate the startup script
**Example:**
```go
import (
    "embed"
    "text/template"
    "bytes"
)

//go:embed templates/bootstrap.sh.tmpl
var bootstrapTemplate string

var tmpl = template.Must(template.New("bootstrap").Parse(bootstrapTemplate))

type BootstrapData struct {
    InstanceID          string
    ProxyEndpoint       string
    ProxyPublicKey      string
    InstancePrivateKey  string
    InstanceAddress     string // e.g., "10.0.0.5/16"
    SSHAuthorizedKeys   string
    InternalToken       string
    Hostname            string
    DockerImage         string
}

func RenderBootstrap(data BootstrapData) (string, error) {
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("render bootstrap template: %w", err)
    }
    return buf.String(), nil
}
```
**Confidence:** HIGH -- `text/template` + `embed` is standard Go pattern. Template.Must at package init catches parse errors at startup.

### Pattern 6: Customer Response Types (Defense by Omission)
**What:** Separate struct types that physically cannot contain upstream fields
**When to use:** All customer-facing API responses
**Example:**
```go
// Internal type -- has everything, used by provisioning engine and DB layer
type Instance struct {
    InstanceID           string
    OrgID                string
    UserID               string
    UpstreamProvider     string  // NEVER serialized to customer
    UpstreamID           string  // NEVER serialized to customer
    UpstreamIP           string  // NEVER serialized to customer
    Hostname             string
    WGPublicKey          string
    WGAddress            string
    GPUType              string
    GPUCount             int
    Tier                 string
    Region               string
    PricePerHour         float64
    Status               string
    CreatedAt            time.Time
}

// Customer-facing type -- upstream fields structurally absent
type CustomerInstance struct {
    ID           string    `json:"id"`
    Hostname     string    `json:"hostname"`
    SSHCommand   string    `json:"ssh_command"`
    Status       string    `json:"status"`
    GPUType      string    `json:"gpu_type"`
    GPUCount     int       `json:"gpu_count"`
    Tier         string    `json:"tier"`
    Region       string    `json:"region"`
    PricePerHour float64   `json:"price_per_hour"`
    CreatedAt    time.Time `json:"created_at"`
}

// Conversion method
func (i *Instance) ToCustomer() CustomerInstance {
    return CustomerInstance{
        ID:           i.InstanceID,
        Hostname:     i.Hostname,
        SSHCommand:   fmt.Sprintf("ssh root@%s", i.Hostname),
        Status:       i.Status,
        GPUType:      i.GPUType,
        GPUCount:     i.GPUCount,
        Tier:         i.Tier,
        Region:       i.Region,
        PricePerHour: i.PricePerHour,
        CreatedAt:    i.CreatedAt,
    }
}
```
**Confidence:** HIGH -- This is a straightforward Go pattern. The key insight is that if a field does not exist in the struct, it cannot accidentally be serialized via `json.Marshal`.

### Anti-Patterns to Avoid
- **Shell-out to `wg genkey`/`wg set`:** Fragile, harder to test, requires WireGuard tools installed on build machine. Use wgctrl-go library instead.
- **Middleware-based privacy filtering:** Stripping fields in middleware (e.g., JSON transformer) is error-prone -- one missed endpoint leaks data. Use separate struct types with no upstream fields.
- **Storing plaintext WireGuard private keys:** The decision mandates AES-256-GCM encryption. Store ciphertext + nonce in the DB column, derive encryption key from env var.
- **Random IP allocation for IPAM:** Random selection within a /16 would eventually collide. Sequential with MAX+1 is simple and deterministic.
- **Using `html/template` for cloud-init:** Would HTML-escape bash special characters (`<`, `>`, `&`), breaking the script. Use `text/template`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| WireGuard key generation | Curve25519 math from scratch | `wgtypes.GeneratePrivateKey()` | Cryptographic operations are easy to get wrong; official library uses Go's crypto/rand correctly |
| WireGuard peer management | Parse/write wg0.conf files manually | `wgctrl.Client.ConfigureDevice()` | Config file format may change; API handles kernel and userspace WG devices transparently |
| AES-256-GCM encryption | Custom encryption wrapper | Go stdlib `crypto/aes` + `crypto/cipher` | Authenticated encryption with standard nonce handling; battle-tested implementation |
| IP address increment | String parsing/manipulation | `net.IP` byte arithmetic | Handles IPv4/IPv6, edge cases (overflow), and subnet boundary checking |

**Key insight:** Every piece of this phase has a stdlib or official solution. The WireGuard project maintains wgctrl-go specifically for programmatic control. Go's crypto stdlib provides everything needed for AES-256-GCM. There is zero need for third-party frameworks.

## Common Pitfalls

### Pitfall 1: wgctrl.Client Requires Root / CAP_NET_ADMIN on the Proxy
**What goes wrong:** `wgctrl.New()` and `ConfigureDevice()` fail with permission errors when not running as root
**Why it happens:** WireGuard device control requires CAP_NET_ADMIN capability on Linux
**How to avoid:** The gpuctl binary on the proxy server must run as root or with CAP_NET_ADMIN. For development/testing, use interface-based mocking of the WireGuard client
**Warning signs:** `operation not permitted` errors from wgctrl calls

### Pitfall 2: AES-GCM Nonce Reuse
**What goes wrong:** Reusing a nonce with the same key completely breaks GCM's security guarantees
**Why it happens:** Using a fixed or predictable nonce instead of `crypto/rand`
**How to avoid:** Always generate a fresh random nonce via `io.ReadFull(rand.Reader, nonce)` for every encryption. Prepend the nonce to the ciphertext so it is available for decryption. With random 12-byte nonces, the collision probability is negligible for fewer than 2^32 encryptions per key (far more than our instance count)
**Warning signs:** Unit tests should verify each encryption of the same plaintext produces different ciphertext

### Pitfall 3: WireGuard on RunPod Containers (NET_ADMIN / Kernel Module)
**What goes wrong:** RunPod pods are Docker containers, not full VMs. Kernel WireGuard requires the `wireguard` kernel module to be loaded on the host AND the container needs NET_ADMIN capability. Neither is guaranteed.
**Why it happens:** RunPod documentation does not specify container capabilities or host kernel module availability
**How to avoid:** The cloud-init template should try kernel WireGuard first (`apt-get install wireguard`), and if `wg-quick up wg0` fails, fall back to `wireguard-go` (userspace implementation that requires no kernel module and no NET_ADMIN). The userspace implementation is slower but functionally identical.
**Warning signs:** `RTNETLINK answers: Operation not permitted` or `wireguard module not found` during boot
**Mitigation in template:**
```bash
# Try kernel WireGuard first, fall back to userspace
if ! modprobe wireguard 2>/dev/null; then
    # Install wireguard-go userspace implementation
    wget -qO /usr/local/bin/wireguard-go <wireguard-go-binary-url>
    chmod +x /usr/local/bin/wireguard-go
    export WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go
fi
```

### Pitfall 4: text/template vs html/template
**What goes wrong:** Using `html/template` for bash scripts causes HTML entity encoding of `<`, `>`, `&` characters, breaking iptables rules and bash conditionals
**Why it happens:** `html/template` is designed for HTML output and auto-escapes for XSS prevention
**How to avoid:** Always use `text/template` for non-HTML output. The CONTEXT.md decision already specifies this.
**Warning signs:** Rendered scripts contain `&lt;` instead of `<`, `&amp;` instead of `&`

### Pitfall 5: IPAM Race Condition Without Advisory Lock
**What goes wrong:** Two concurrent provisioning requests get the same `MAX(wg_address) + 1`, causing a UNIQUE constraint violation
**Why it happens:** Without locking, both transactions read the same MAX value before either inserts
**How to avoid:** Use `pg_advisory_xact_lock(constant_id)` at the start of the IPAM allocation query. This is transaction-scoped (auto-released on commit/rollback) and blocks concurrent allocators. The lock ID should be a fixed constant (e.g., hash of 'ipam_allocate').
**Warning signs:** Sporadic `duplicate key value violates unique constraint` errors on wg_address under concurrent provisioning

### Pitfall 6: Template Injection / Shell Escaping
**What goes wrong:** If SSH public keys or instance IDs contain shell metacharacters, the rendered cloud-init script could execute arbitrary commands
**Why it happens:** `text/template` does not escape for shell context
**How to avoid:** Validate all template inputs before rendering: SSH keys must match `^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp\d+) [A-Za-z0-9+/=]+ .*$`, instance IDs are alphanumeric with hyphens only. Use `{{.Field}}` inside single-quoted strings in the template where possible.
**Warning signs:** Template test with adversarial input (keys containing `$(malicious)`) should be a test case

### Pitfall 7: wg_private_key_enc Column Was Removed in Schema v1
**What goes wrong:** SCHEMA-03 explicitly removed the `wg_private_key_enc` column as a "security liability." But CONTEXT.md now mandates storing encrypted WireGuard private keys.
**Why it happens:** Phase 2 removed the column because key generation was deferred to Phase 3. Now Phase 3 needs it back, but with proper encryption.
**How to avoid:** A new migration must re-add `wg_private_key_enc TEXT` to the instances table. The column now stores AES-256-GCM encrypted ciphertext (hex-encoded), not plaintext. Document this clearly.
**Warning signs:** Missing column error when trying to store encrypted keys

## Code Examples

Verified patterns from official sources:

### WireGuard Key Generation and Base64 Round-Trip
```go
// Source: https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl/wgtypes
import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

// Generate new key pair
privKey, err := wgtypes.GeneratePrivateKey()
if err != nil {
    return err
}
pubKey := privKey.PublicKey()

// Serialize to base64 strings for storage/transmission
privKeyStr := privKey.String()  // base64-encoded 32 bytes
pubKeyStr := pubKey.String()    // base64-encoded 32 bytes

// Parse back from base64
parsedKey, err := wgtypes.ParseKey(privKeyStr)
if err != nil {
    return err
}
// parsedKey == privKey
```

### Adding a Peer with AllowedIPs and Keepalive
```go
// Source: https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl
import (
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
    "net"
    "time"
)

client, err := wgctrl.New()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

peerPubKey, _ := wgtypes.ParseKey("base64PublicKeyHere=")
keepalive := 25 * time.Second

err = client.ConfigureDevice("wg0", wgtypes.Config{
    Peers: []wgtypes.PeerConfig{{
        PublicKey:                   peerPubKey,
        ReplaceAllowedIPs:          true,
        AllowedIPs:                 []net.IPNet{{
            IP:   net.ParseIP("10.0.0.5"),
            Mask: net.CIDRMask(32, 32),
        }},
        PersistentKeepaliveInterval: &keepalive,
    }},
})
```

### IP Address Increment for IPAM
```go
// Source: Go stdlib net.IP is a byte slice
func incrementIP(ip net.IP) net.IP {
    // Work on a copy to avoid modifying the original
    result := make(net.IP, len(ip))
    copy(result, ip)

    // Use 4-byte form for IPv4
    ip4 := result.To4()
    if ip4 != nil {
        result = ip4
    }

    // Increment from last byte, carrying over
    for i := len(result) - 1; i >= 0; i-- {
        result[i]++
        if result[i] != 0 {
            break // no carry needed
        }
    }
    return result
}
```

### PostgreSQL Advisory Lock for IPAM
```go
// Source: https://www.postgresql.org/docs/current/explicit-locking.html
// pg_advisory_xact_lock is transaction-scoped -- released on COMMIT or ROLLBACK

const ipamLockID int64 = 0x4750554149_4950414D // "GPUAI_IPAM" as int64

func allocateNextAddress(ctx context.Context, tx pgx.Tx) (string, error) {
    // Acquire transaction-scoped advisory lock
    if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", ipamLockID); err != nil {
        return "", err
    }

    var maxAddr string
    err := tx.QueryRow(ctx,
        `SELECT COALESCE(
            host(MAX(wg_address)),
            '10.0.0.1'
         ) FROM instances WHERE wg_address IS NOT NULL`,
    ).Scan(&maxAddr)
    if err != nil {
        return "", err
    }

    nextIP := incrementIP(net.ParseIP(maxAddr))
    return nextIP.String(), nil
}
```

### Embedded Template with text/template
```go
// Source: https://pkg.go.dev/embed, https://pkg.go.dev/text/template
import (
    _ "embed"
    "text/template"
    "bytes"
)

//go:embed templates/bootstrap.sh.tmpl
var bootstrapTmplStr string

var bootstrapTmpl = template.Must(
    template.New("bootstrap").Parse(bootstrapTmplStr),
)

type BootstrapParams struct {
    InstanceID         string
    ProxyEndpoint      string
    ProxyPublicKey     string
    InstancePrivateKey string
    InstanceAddress    string
    SSHAuthorizedKeys  string
    InternalToken      string
    Hostname           string
}

func RenderBootstrap(params BootstrapParams) (string, error) {
    var buf bytes.Buffer
    if err := bootstrapTmpl.Execute(&buf, params); err != nil {
        return "", fmt.Errorf("render bootstrap: %w", err)
    }
    return buf.String(), nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Shell out to `wg genkey` / `wg set` | `wgctrl-go` programmatic API | 2020+ (library matured) | Testable, no binary dependency, cross-platform |
| Kernel WireGuard only | Kernel + `wireguard-go` userspace fallback | wireguard-go stable since 2021 | Works in containers without kernel module |
| Manual wg0.conf editing | `wgctrl.ConfigureDevice()` live API | wgctrl-go v0.0.0-20200609 | Atomic peer changes without config file parsing |
| IP allocation with SELECT FOR UPDATE | pg_advisory_xact_lock | PostgreSQL 8.2+ | Less deadlock risk, simpler than row-level locks |
| Go 1.16 embed | Go 1.24 embed (unchanged API) | N/A | Stable API since Go 1.16, no changes needed |

**Deprecated/outdated:**
- The original `wireguard` npm package and various wrappers are irrelevant (this is Go)
- The `wg` CLI tool still works but programmatic control via wgctrl-go is preferred for production systems
- `wg_private_key_enc` was removed in SCHEMA-03 but needs to be re-added with proper encryption in Phase 3

## Open Questions

1. **RunPod NET_ADMIN / WireGuard kernel module availability**
   - What we know: RunPod pods are Docker containers. WireGuard kernel module requires host support. NET_ADMIN capability is undocumented for RunPod.
   - What's unclear: Whether RunPod's host kernel has the wireguard module loaded, and whether containers get NET_ADMIN.
   - Recommendation: Design the cloud-init template to try kernel WireGuard first and fall back to wireguard-go userspace. This handles both cases. Empirical validation needed once a RunPod API key is available. This was already flagged in STATE.md as a blocker to validate.

2. **WG_ENCRYPTION_KEY environment variable**
   - What we know: AES-256-GCM requires a 32-byte key. The CONTEXT.md says "encryption key from env var."
   - What's unclear: Exact env var name, format (hex? base64? raw?), and whether to validate at startup.
   - Recommendation: Add `WG_ENCRYPTION_KEY` to config.go, require it as hex-encoded 32 bytes (64 hex chars), validate length at startup. Add to `.env.example`. This is Claude's discretion per CONTEXT.md.

3. **Port mapping iptables integration with WireGuard Manager**
   - What we know: The decision says "When WireGuard manager adds a peer, it also adds the iptables/routing rule."
   - What's unclear: Exact iptables rule syntax for port mapping (DNAT from proxy:100XX to 10.0.0.X:22).
   - Recommendation: The Manager.AddPeer() method should also call `iptables -t nat -A PREROUTING -p tcp --dport {port} -j DNAT --to-destination {tunnelIP}:22` and a matching FORWARD rule. Port number can be derived from the last octet(s) of the tunnel IP (e.g., 10.0.0.5 -> port 10005). Removal must delete both rules. This is Claude's discretion.

4. **Schema migration: re-adding wg_private_key_enc**
   - What we know: SCHEMA-03 removed `wg_private_key_enc`. Phase 3 needs to store encrypted keys.
   - What's unclear: Whether to re-add the same column name or use a new name.
   - Recommendation: Re-add as `wg_private_key_enc TEXT` via a new migration. The column name is already established in the codebase conventions. A comment in the migration should clarify this stores AES-256-GCM ciphertext, not plaintext.

5. **Cloud-init template location**
   - What we know: The existing template is at `infra/cloud-init/bootstrap.sh`. The decision says embed via `//go:embed`.
   - What's unclear: Whether to embed from `infra/cloud-init/` or move the template into `internal/wireguard/templates/`.
   - Recommendation: Keep the template at `infra/cloud-init/bootstrap.sh.tmpl` (rename to .tmpl to signal it is a Go template, not a standalone script). Use `//go:embed ../../infra/cloud-init/bootstrap.sh.tmpl` from `internal/wireguard/template.go`, OR move a copy to `internal/wireguard/templates/bootstrap.sh.tmpl` since `//go:embed` cannot traverse above the module root. The latter is cleaner. Validate with a test.

## Sources

### Primary (HIGH confidence)
- [wgtypes package - pkg.go.dev](https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl/wgtypes) - Key type, GeneratePrivateKey, PublicKey, PeerConfig, Config structs
- [wgctrl package - pkg.go.dev](https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl) - Client, ConfigureDevice, Device, Close methods
- [crypto/cipher - pkg.go.dev](https://pkg.go.dev/crypto/cipher) - NewGCM, Seal, Open for AES-GCM
- [crypto/aes - pkg.go.dev](https://pkg.go.dev/crypto/aes) - NewCipher for AES block cipher
- [text/template - pkg.go.dev](https://pkg.go.dev/text/template) - Template parsing and execution
- [PostgreSQL 18 Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html) - Advisory lock documentation
- [WireGuard/wgctrl-go GitHub](https://github.com/WireGuard/wgctrl-go) - Official WireGuard Go control library source

### Secondary (MEDIUM confidence)
- [AES-256 GCM Encryption Example in Golang (GitHub Gist)](https://gist.github.com/kkirsche/e28da6754c39d5e7ea10) - Community AES-GCM pattern, verified against stdlib docs
- [PG advisory locks in Go with built-in hashes (brandur.org)](https://brandur.org/fragments/pg-advisory-locks-with-go-hash) - Advisory lock usage patterns in Go
- [Using Go Embed Package for Template Rendering](https://andrew-mccall.com/blog/2025/01/using-go-embed-package-for-template-rendering/) - embed + template integration patterns
- [RunPod Initialization Scripts (DeepWiki)](https://deepwiki.com/runpod/containers/6.2-initialization-scripts) - RunPod container startup process (pre_start.sh, post_start.sh)

### Tertiary (LOW confidence)
- RunPod container capabilities (NET_ADMIN) - NOT documented anywhere found. Empirical validation required. FLAG FOR VALIDATION.
- [wireguard-go userspace implementation](https://git.zx2c4.com/wireguard-go) - Fallback for containers without kernel WireGuard. Known to work but performance characteristics in GPU workloads unverified.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - wgctrl-go is the official WireGuard project library; all other components are Go stdlib
- Architecture: HIGH - Patterns follow established Go conventions and the project's existing codebase patterns (ServerDeps injection, pgx usage, stdlib net/http)
- Pitfalls: HIGH - WireGuard peer management, AES-GCM nonce handling, and PostgreSQL advisory locks are well-documented domains with known failure modes
- RunPod WireGuard compatibility: LOW - Undocumented; requires empirical validation

**Research date:** 2026-02-24
**Valid until:** 2026-03-24 (stable domain; wgctrl-go and Go stdlib change slowly)
