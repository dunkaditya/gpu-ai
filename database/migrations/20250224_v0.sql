-- GPU.ai Database Schema v0 (updated to match ARCHITECTURE.md reference)
-- Initial schema for Phase 1

-- Enable pgcrypto for gen_random_uuid() compatibility with PostgreSQL < 13
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Organizations / customers
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users within organizations
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    clerk_user_id VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'member',  -- admin, member
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- SSH keys
CREATE TABLE ssh_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255),
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- GPU instances
CREATE TABLE instances (
    id VARCHAR(12) PRIMARY KEY,  -- e.g., "gpu-4a7f"
    org_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),

    -- Upstream (hidden from customer)
    upstream_provider VARCHAR(50) NOT NULL,  -- runpod, e2e, lambda
    upstream_id VARCHAR(255) NOT NULL,
    upstream_ip INET,

    -- GPU.ai facing
    hostname VARCHAR(255) NOT NULL,  -- gpu-4a7f.gpu.ai
    wg_public_key VARCHAR(255),
    wg_private_key_enc VARCHAR(255),  -- encrypted at rest
    wg_address INET,  -- 10.0.0.x

    -- Configuration
    gpu_type VARCHAR(50) NOT NULL,
    gpu_count INT NOT NULL,
    tier VARCHAR(20) NOT NULL,  -- on_demand, spot
    region VARCHAR(50) NOT NULL,

    -- Billing
    price_per_hour DECIMAL(10, 4) NOT NULL,
    upstream_price_per_hour DECIMAL(10, 4) NOT NULL,  -- what we pay
    billing_start TIMESTAMPTZ,
    billing_end TIMESTAMPTZ,

    -- Status
    status VARCHAR(20) DEFAULT 'creating',  -- creating, running, stopping, terminated
    created_at TIMESTAMPTZ DEFAULT NOW(),
    terminated_at TIMESTAMPTZ
);

-- Saved environments
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    docker_image VARCHAR(512),  -- registry URL
    description TEXT,
    is_shared BOOLEAN DEFAULT FALSE,  -- shared within org
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Usage records for billing
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id VARCHAR(12) REFERENCES instances(id),
    org_id UUID REFERENCES organizations(id),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    gpu_type VARCHAR(50),
    gpu_count INT,
    price_per_hour DECIMAL(10, 4),
    total_cost DECIMAL(10, 4),
    stripe_usage_record_id VARCHAR(255)
);

-- Indexes
CREATE INDEX idx_instances_org_id ON instances(org_id);
CREATE INDEX idx_instances_status ON instances(status);
CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_users_org_id ON users(org_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_ssh_keys_user_id ON ssh_keys(user_id);
CREATE INDEX idx_environments_org_id ON environments(org_id);
CREATE INDEX idx_usage_records_org_id ON usage_records(org_id);
CREATE INDEX idx_usage_records_instance_id ON usage_records(instance_id);
