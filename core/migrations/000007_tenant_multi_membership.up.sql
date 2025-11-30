-- GoConnect Multi-Tenant Membership Migration
-- Version: 7
-- Author: AI Assistant
-- Date: 2025-11-25
-- Description: Adds tenant multi-membership system

-- =====================================================
-- Part 1: Enhance tenants table with new fields
-- =====================================================

-- Add new columns to tenants table
ALTER TABLE tenants
ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS icon_url TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) DEFAULT 'private' CHECK (visibility IN ('public', 'unlisted', 'private')),
ADD COLUMN IF NOT EXISTS access_type VARCHAR(20) DEFAULT 'invite_only' CHECK (access_type IN ('open', 'password', 'invite_only')),
ADD COLUMN IF NOT EXISTS password_hash VARCHAR(512),
ADD COLUMN IF NOT EXISTS max_members INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users(id);

-- Update existing tenants to set owner_id from first admin user
UPDATE tenants t
SET owner_id = (
    SELECT u.id FROM users u 
    WHERE u.tenant_id = t.id AND u.is_admin = true 
    ORDER BY u.created_at ASC 
    LIMIT 1
)
WHERE owner_id IS NULL;

-- If no admin found, use the first user
UPDATE tenants t
SET owner_id = (
    SELECT u.id FROM users u 
    WHERE u.tenant_id = t.id 
    ORDER BY u.created_at ASC 
    LIMIT 1
)
WHERE owner_id IS NULL;

-- Create index for tenant discovery
CREATE INDEX IF NOT EXISTS idx_tenants_visibility ON tenants(visibility);
CREATE INDEX IF NOT EXISTS idx_tenants_owner_id ON tenants(owner_id);

-- =====================================================
-- Part 2: Tenant Members table (N:N relationship)
-- =====================================================

CREATE TABLE IF NOT EXISTS tenant_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'moderator', 'vip', 'member')),
    nickname VARCHAR(100),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant ON tenant_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_role ON tenant_members(tenant_id, role);

-- Migrate existing users to tenant_members
-- Each user's current tenant_id becomes their membership
INSERT INTO tenant_members (tenant_id, user_id, role, joined_at, updated_at)
SELECT 
    u.tenant_id,
    u.id,
    CASE 
        WHEN u.is_admin THEN 'admin'
        WHEN u.is_moderator THEN 'moderator'
        ELSE 'member'
    END,
    u.created_at,
    NOW()
FROM users u
WHERE u.tenant_id IS NOT NULL
ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- Set owner role for tenant owners
UPDATE tenant_members tm
SET role = 'owner'
FROM tenants t
WHERE tm.tenant_id = t.id AND tm.user_id = t.owner_id;

-- =====================================================
-- Part 3: Tenant Invites table (Steam-like codes)
-- =====================================================

CREATE TABLE IF NOT EXISTS tenant_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(20) UNIQUE NOT NULL,
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tenant_invites_code ON tenant_invites(code);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant ON tenant_invites(tenant_id);

-- =====================================================
-- Part 4: Tenant Announcements table
-- =====================================================

CREATE TABLE IF NOT EXISTS tenant_announcements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id),
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_announcements_tenant ON tenant_announcements(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_announcements_pinned ON tenant_announcements(tenant_id, is_pinned, created_at DESC);

-- =====================================================
-- Part 5: Tenant Chat Messages table
-- =====================================================

CREATE TABLE IF NOT EXISTS tenant_chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    edited_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tenant_chat_tenant ON tenant_chat_messages(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tenant_chat_user ON tenant_chat_messages(user_id);

-- =====================================================
-- Part 6: Enhance Networks table for role-based access
-- =====================================================

ALTER TABLE networks
ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS required_role VARCHAR(20) DEFAULT 'member' CHECK (required_role IN ('owner', 'admin', 'moderator', 'vip', 'member')),
ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_networks_required_role ON networks(tenant_id, required_role);

-- =====================================================
-- Part 7: Utility functions
-- =====================================================

-- Function to get member count for a tenant
CREATE OR REPLACE FUNCTION get_tenant_member_count(p_tenant_id UUID)
RETURNS INTEGER AS $$
    SELECT COUNT(*)::INTEGER FROM tenant_members WHERE tenant_id = p_tenant_id;
$$ LANGUAGE SQL STABLE;

-- Function to check if user has required role
CREATE OR REPLACE FUNCTION user_has_tenant_role(
    p_user_id UUID,
    p_tenant_id UUID,
    p_required_role VARCHAR(20)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_user_role VARCHAR(20);
    v_required_level INTEGER;
    v_user_level INTEGER;
BEGIN
    -- Get user's role in tenant
    SELECT role INTO v_user_role
    FROM tenant_members
    WHERE user_id = p_user_id AND tenant_id = p_tenant_id;
    
    IF v_user_role IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- Role hierarchy: owner=100, admin=80, moderator=60, vip=40, member=20
    v_required_level := CASE p_required_role
        WHEN 'owner' THEN 100
        WHEN 'admin' THEN 80
        WHEN 'moderator' THEN 60
        WHEN 'vip' THEN 40
        WHEN 'member' THEN 20
        ELSE 0
    END;
    
    v_user_level := CASE v_user_role
        WHEN 'owner' THEN 100
        WHEN 'admin' THEN 80
        WHEN 'moderator' THEN 60
        WHEN 'vip' THEN 40
        WHEN 'member' THEN 20
        ELSE 0
    END;
    
    RETURN v_user_level >= v_required_level;
END;
$$ LANGUAGE plpgsql STABLE;

-- =====================================================
-- Part 8: Views for common queries
-- =====================================================

-- View for tenant list with member count
CREATE OR REPLACE VIEW tenant_with_stats AS
SELECT 
    t.*,
    COALESCE(mc.member_count, 0) AS member_count
FROM tenants t
LEFT JOIN (
    SELECT tenant_id, COUNT(*) AS member_count
    FROM tenant_members
    GROUP BY tenant_id
) mc ON t.id = mc.tenant_id;

-- View for user's tenants with their role
CREATE OR REPLACE VIEW user_tenant_memberships AS
SELECT 
    tm.user_id,
    tm.tenant_id,
    tm.role,
    tm.nickname,
    tm.joined_at,
    t.name AS tenant_name,
    t.description AS tenant_description,
    t.icon_url AS tenant_icon_url,
    t.visibility AS tenant_visibility
FROM tenant_members tm
JOIN tenants t ON t.id = tm.tenant_id;

-- =====================================================
-- Comments for documentation
-- =====================================================

COMMENT ON TABLE tenant_members IS 'User memberships in tenants (N:N relationship)';
COMMENT ON TABLE tenant_invites IS 'Invitation codes for joining tenants';
COMMENT ON TABLE tenant_announcements IS 'Admin announcements visible to all tenant members';
COMMENT ON TABLE tenant_chat_messages IS 'General chat messages in tenant';

COMMENT ON COLUMN tenants.visibility IS 'public=discoverable, unlisted=link only, private=invite only';
COMMENT ON COLUMN tenants.access_type IS 'open=anyone, password=needs password, invite_only=needs code';
COMMENT ON COLUMN tenant_members.role IS 'owner>admin>moderator>vip>member';
COMMENT ON COLUMN networks.required_role IS 'Minimum role required to access this network';
COMMENT ON COLUMN networks.is_hidden IS 'Hidden networks not shown in list, but accessible if user has permission';
