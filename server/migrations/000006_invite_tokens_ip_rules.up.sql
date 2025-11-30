-- GoConnect Invite Tokens and IP Rules Migration
-- Version: 6.0
-- Author: orhaniscoding
-- Date: 2025-11-25

-- Invite Tokens table (for network invitations)
CREATE TABLE IF NOT EXISTS invite_tokens (
    id VARCHAR(64) PRIMARY KEY,
    network_id UUID NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    uses_max INT NOT NULL DEFAULT 0, -- 0 = unlimited
    uses_left INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_invite_tokens_network_id ON invite_tokens(network_id);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_tenant_id ON invite_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_token ON invite_tokens(token);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_expires_at ON invite_tokens(expires_at) WHERE revoked_at IS NULL;

COMMENT ON TABLE invite_tokens IS 'Network invitation tokens for sharing join links';
COMMENT ON COLUMN invite_tokens.uses_max IS '0 means unlimited uses';
COMMENT ON COLUMN invite_tokens.uses_left IS 'Remaining uses, decrements on each use';
COMMENT ON COLUMN invite_tokens.revoked_at IS 'When the token was revoked (NULL = active)';

-- IP Rules table (for tenant IP allowlist/denylist)
CREATE TABLE IF NOT EXISTS ip_rules (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('allow', 'deny')),
    cidr VARCHAR(50) NOT NULL, -- IP or CIDR range (e.g., "192.168.1.0/24")
    description VARCHAR(255),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE -- Optional expiration
);

CREATE INDEX IF NOT EXISTS idx_ip_rules_tenant_id ON ip_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ip_rules_type ON ip_rules(type);
CREATE INDEX IF NOT EXISTS idx_ip_rules_expires_at ON ip_rules(expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE ip_rules IS 'IP allow/deny rules per tenant for access control';
COMMENT ON COLUMN ip_rules.type IS 'allow = whitelist, deny = blacklist';
COMMENT ON COLUMN ip_rules.cidr IS 'IP address or CIDR range';
COMMENT ON COLUMN ip_rules.expires_at IS 'Optional expiration time for temporary rules';
