-- GoConnect Initial Schema Migration
-- Version: 1.0
-- Author: orhaniscoding
-- Date: 2025-10-29

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tenants table (multi-tenancy support)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_created_at ON tenants(created_at);

-- Users table (authentication)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(512) NOT NULL, -- Argon2id: salt:hash format (base64)
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Networks table (VPN networks)
CREATE TABLE networks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    cidr VARCHAR(50) NOT NULL, -- e.g., "10.0.0.0/24"
    visibility VARCHAR(20) NOT NULL DEFAULT 'private' CHECK (visibility IN ('public', 'private')),
    join_policy VARCHAR(20) NOT NULL DEFAULT 'approval' CHECK (join_policy IN ('open', 'invite', 'approval')),
    dns VARCHAR(255), -- Optional DNS server
    mtu INT, -- Optional MTU
    split_tunnel BOOLEAN, -- Optional split tunnel flag
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE, -- soft delete
    moderation_redacted BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_networks_tenant_id ON networks(tenant_id);
CREATE INDEX idx_networks_created_by ON networks(created_by);
CREATE INDEX idx_networks_visibility ON networks(visibility);
CREATE INDEX idx_networks_deleted_at ON networks(deleted_at);
CREATE INDEX idx_networks_created_at ON networks(created_at);

-- Memberships table (network members with roles)
CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'member', 'read_only')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'banned', 'pending')),
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(network_id, user_id)
);

CREATE INDEX idx_memberships_network_id ON memberships(network_id);
CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_status ON memberships(status);
CREATE INDEX idx_memberships_role ON memberships(role);

-- Join requests table (pending approvals)
CREATE TABLE join_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied')),
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE(network_id, user_id)
);

CREATE INDEX idx_join_requests_network_id ON join_requests(network_id);
CREATE INDEX idx_join_requests_user_id ON join_requests(user_id);
CREATE INDEX idx_join_requests_status ON join_requests(status);

-- IP allocations table (IPAM)
CREATE TABLE ip_allocations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(50) NOT NULL, -- e.g., "10.0.0.5"
    allocated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(network_id, ip_address),
    UNIQUE(network_id, user_id)
);

CREATE INDEX idx_ip_allocations_network_id ON ip_allocations(network_id);
CREATE INDEX idx_ip_allocations_user_id ON ip_allocations(user_id);
CREATE INDEX idx_ip_allocations_ip_address ON ip_allocations(ip_address);

-- Idempotency keys table (24h TTL for mutation deduplication)
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response_body TEXT NOT NULL,
    response_status INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys(created_at);

-- Audit events table (immutable log)
CREATE TABLE audit_events (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255), -- hashed user ID
    object_type VARCHAR(50),
    object_id VARCHAR(255), -- hashed object ID
    network_id UUID REFERENCES networks(id) ON DELETE SET NULL,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    prev_hash VARCHAR(128), -- SHA-256 of previous event
    hash VARCHAR(128) NOT NULL -- SHA-256 of this event
);

CREATE INDEX idx_audit_events_tenant_id ON audit_events(tenant_id);
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_network_id ON audit_events(network_id);
CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_networks_updated_at BEFORE UPDATE ON networks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_memberships_updated_at BEFORE UPDATE ON memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE tenants IS 'Multi-tenancy: isolated workspaces for organizations';
COMMENT ON TABLE users IS 'User accounts with Argon2id password hashing';
COMMENT ON TABLE networks IS 'VPN networks with CIDR ranges';
COMMENT ON TABLE memberships IS 'Network memberships with RBAC roles';
COMMENT ON TABLE join_requests IS 'Pending network join requests requiring admin approval';
COMMENT ON TABLE ip_allocations IS 'IPAM: one IP per user per network';
COMMENT ON TABLE idempotency_keys IS 'Mutation deduplication with 24h TTL';
COMMENT ON TABLE audit_events IS 'Immutable audit log with hash chain integrity';
