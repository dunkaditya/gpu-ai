# Database Schema Fixes

Apply these changes to `database/migrations/001_init.sql`.

---

## 1. Rename primary keys to be self-documenting

Every table's primary key should be named `{table_singular}_id` so it matches
the foreign key name used in other tables. No more ambiguous `id` columns.

```
organizations.id  → organizations.org_id
users.id          → users.user_id
ssh_keys.id       → ssh_keys.ssh_key_id
instances.id      → instances.instance_id
environments.id   → environments.environment_id
usage_records.id  → usage_records.usage_record_id
```

Update all foreign key references to match (e.g., `REFERENCES organizations(org_id)`).

---

## 2. Use VARCHAR(32) for instance_id

Change `instances.instance_id` from `VARCHAR(12)` to `VARCHAR(32)` to allow
for longer IDs as the platform scales.

---

## 3. Use NUMERIC consistently instead of DECIMAL

Replace all `DECIMAL(10, 4)` with `NUMERIC(10, 4)`. They're equivalent in
Postgres but NUMERIC is the SQL standard. Be consistent across the schema.

---

## 4. Add updated_at to instances

```sql
updated_at TIMESTAMPTZ DEFAULT NOW()
```

Update this on every status change (creating → running → stopping → terminated).

---

## 5. Add NOT NULL to foreign keys that must always exist

These relationships are mandatory — no orphan records allowed:

```sql
-- instances
org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT,
user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,

-- users
org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT,
```

---

## 6. Add explicit ON DELETE behavior everywhere

Make intent clear on all foreign keys:

```sql
-- ssh_keys (already has CASCADE — correct, delete keys when user deleted)
user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE

-- instances (RESTRICT — can't delete user/org with active instances)
org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT
user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT

-- environments (CASCADE — delete envs when user/org deleted)
org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE CASCADE
user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE

-- usage_records (RESTRICT — never lose billing records)
instance_id VARCHAR(32) NOT NULL REFERENCES instances(instance_id) ON DELETE RESTRICT
org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT
```

---

## 7. Add comment explaining duplicated fields in usage_records

The `gpu_type`, `gpu_count`, and `price_per_hour` fields in `usage_records`
duplicate data from `instances`. This is intentional — they snapshot the
billing parameters at the time of the usage period. Instance pricing could
theoretically change (e.g., spot price fluctuation), and billing records
must reflect what was actually charged, not the instance's current state.

Add SQL comments:

```sql
-- Usage records for billing
-- Fields like gpu_type, gpu_count, price_per_hour are snapshotted here
-- because they reflect the billing parameters at the time of the usage period,
-- not the instance's current state.
CREATE TABLE usage_records (
```

---

## 8. Add CHECK constraint on instances.status

```sql
status VARCHAR(20) DEFAULT 'creating'
    CHECK (status IN ('creating', 'running', 'stopping', 'terminated', 'error')),
```

---

## 9. Add UNIQUE constraint on instances.hostname

Two instances must never share a hostname:

```sql
hostname VARCHAR(255) NOT NULL UNIQUE,
```

---

## 10. Add unique composite index on upstream provider + upstream ID

Prevents accidentally creating two records for the same upstream instance,
and speeds up lookups when checking status or processing callbacks:

```sql
CREATE UNIQUE INDEX idx_instances_upstream
    ON instances(upstream_provider, upstream_id);
```

---

## 11. Remove wg_private_key_enc column

The WireGuard private key is ephemeral — generated at provision time, injected
into cloud-init, and never needed again. Storing it is a security liability
with no upside. Remove entirely.

Keep `wg_public_key` (proxy needs it for peer config) and `wg_address` (tunnel routing).

```sql
-- REMOVE this line:
wg_private_key_enc TEXT,

-- KEEP these:
wg_public_key VARCHAR(255),
wg_address INET,
```

---

## 12. Internal callback authentication

Instance ready/health callbacks authenticate via a per-instance token.
Generated at provision time (`crypto/rand` in Go), injected into cloud-init,
and validated on callback. Works regardless of network topology (proxies,
load balancers, etc.) unlike WireGuard source IP checks.

Add column to instances table:

```sql
internal_token VARCHAR(255) NOT NULL,
```

The handler checks:

```
request header "Authorization: Bearer <token>" == instance.internal_token
```

---

## Summary of final instances table

After all fixes applied:

```sql
CREATE TABLE instances (
    instance_id VARCHAR(32) PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES organizations(org_id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,

    -- Upstream (hidden from customer)
    upstream_provider VARCHAR(50) NOT NULL,
    upstream_id VARCHAR(255) NOT NULL,
    upstream_ip INET,

    -- GPU.ai facing
    hostname VARCHAR(255) NOT NULL UNIQUE,
    wg_public_key VARCHAR(255),
    wg_address INET,
    internal_token VARCHAR(255) NOT NULL,

    -- Configuration
    gpu_type VARCHAR(50) NOT NULL,
    gpu_count INT NOT NULL,
    tier VARCHAR(20) NOT NULL,
    region VARCHAR(50) NOT NULL,

    -- Billing
    price_per_hour NUMERIC(10, 4) NOT NULL,
    upstream_price_per_hour NUMERIC(10, 4) NOT NULL,
    billing_start TIMESTAMPTZ,
    billing_end TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) DEFAULT 'creating'
        CHECK (status IN ('creating', 'running', 'stopping', 'terminated', 'error')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    terminated_at TIMESTAMPTZ
);
```