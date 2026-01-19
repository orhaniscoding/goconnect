-- GoConnect Migration: Add roles and permissions tables
-- Version: 15
-- Date: 2026-01-19
-- Description: Granular RBAC system with role hierarchy and channel overrides

-- ═══════════════════════════════════════════════════════════════════════════
-- PERMISSION DEFINITIONS (Enum-like table for all possible permissions)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS permission_definitions (
    id              VARCHAR(100) PRIMARY KEY,   -- 'server.manage', 'channel.send_messages'
    category        VARCHAR(50) NOT NULL,       -- 'server', 'network', 'channel', 'member', 'voice'
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    default_value   BOOLEAN DEFAULT FALSE,

    CONSTRAINT permission_category_check CHECK (category IN ('server', 'network', 'channel', 'member', 'voice'))
);

-- ═══════════════════════════════════════════════════════════════════════════
-- ROLES (Custom roles per tenant)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    name            VARCHAR(100) NOT NULL,
    color           VARCHAR(7),             -- Hex color (#FF5733)
    icon            VARCHAR(255),           -- Emoji or URL
    position        INTEGER NOT NULL,       -- Hierarchy (higher = more powerful)

    -- System roles
    is_default      BOOLEAN DEFAULT FALSE,  -- Auto-assigned to new members
    is_admin        BOOLEAN DEFAULT FALSE,  -- Has all permissions

    -- Display options
    mentionable     BOOLEAN DEFAULT FALSE,
    hoist           BOOLEAN DEFAULT FALSE,  -- Show separately in member list

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT roles_name_length CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
    CONSTRAINT roles_color_format CHECK (color IS NULL OR color ~ '^#[0-9A-Fa-f]{6}$'),
    CONSTRAINT roles_position_positive CHECK (position >= 0),
    UNIQUE(tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roles_tenant ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_position ON roles(tenant_id, position DESC);
CREATE INDEX IF NOT EXISTS idx_roles_default ON roles(tenant_id, is_default) WHERE is_default = TRUE;

-- ═══════════════════════════════════════════════════════════════════════════
-- ROLE PERMISSIONS (Which permissions each role has)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id   VARCHAR(100) NOT NULL REFERENCES permission_definitions(id) ON DELETE CASCADE,
    allowed         BOOLEAN NOT NULL,           -- true = allow, false = deny

    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- USER ROLES (Which roles each user has in a tenant)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS user_roles (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id         UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by     UUID REFERENCES users(id) ON DELETE SET NULL,

    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- CHANNEL PERMISSION OVERRIDES (Per-channel role/user permission overrides)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS channel_permission_overrides (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,

    -- Target: either role OR user (exactly one)
    role_id         UUID REFERENCES roles(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,

    permission_id   VARCHAR(100) NOT NULL REFERENCES permission_definitions(id) ON DELETE CASCADE,
    allowed         BOOLEAN,                    -- true = allow, false = deny, NULL = inherit

    -- Constraints
    CONSTRAINT override_target_check CHECK (
        (CASE WHEN role_id IS NOT NULL THEN 1 ELSE 0 END +
         CASE WHEN user_id IS NOT NULL THEN 1 ELSE 0 END) = 1
    ),
    -- Unique per channel + target + permission
    UNIQUE(channel_id, role_id, permission_id),
    UNIQUE(channel_id, user_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_overrides_channel ON channel_permission_overrides(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_overrides_role ON channel_permission_overrides(role_id) WHERE role_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_channel_overrides_user ON channel_permission_overrides(user_id) WHERE user_id IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════════
-- SEED DATA: Default permission definitions
-- ═══════════════════════════════════════════════════════════════════════════

INSERT INTO permission_definitions (id, category, name, description, default_value) VALUES
-- Server Permissions
('server.manage', 'server', 'Manage Server', 'Edit server settings, name, icon', FALSE),
('server.delete', 'server', 'Delete Server', 'Permanently delete the server', FALSE),
('server.view_audit_log', 'server', 'View Audit Log', 'View server audit log', FALSE),
('server.manage_roles', 'server', 'Manage Roles', 'Create, edit, delete roles', FALSE),
('server.manage_channels', 'server', 'Manage Channels', 'Create, edit, delete channels', FALSE),
('server.manage_sections', 'server', 'Manage Sections', 'Create, edit, delete sections', FALSE),
('server.kick_members', 'server', 'Kick Members', 'Remove members from server', FALSE),
('server.ban_members', 'server', 'Ban Members', 'Ban members from server', FALSE),
('server.create_invite', 'server', 'Create Invite', 'Create server invite links', TRUE),

-- Network Permissions
('network.create', 'network', 'Create Network', 'Create VPN networks', FALSE),
('network.manage', 'network', 'Manage Network', 'Edit network settings', FALSE),
('network.delete', 'network', 'Delete Network', 'Delete networks', FALSE),
('network.connect', 'network', 'Connect to Network', 'Join VPN networks', TRUE),
('network.kick_peers', 'network', 'Kick Peers', 'Remove peers from network', FALSE),
('network.ban_peers', 'network', 'Ban Peers', 'Ban peers from network', FALSE),
('network.approve_join', 'network', 'Approve Join Requests', 'Approve pending join requests', FALSE),

-- Channel Permissions
('channel.view', 'channel', 'View Channel', 'See the channel', TRUE),
('channel.send_messages', 'channel', 'Send Messages', 'Send messages in text channels', TRUE),
('channel.embed_links', 'channel', 'Embed Links', 'Links show preview', TRUE),
('channel.attach_files', 'channel', 'Attach Files', 'Upload files', TRUE),
('channel.add_reactions', 'channel', 'Add Reactions', 'Add emoji reactions', TRUE),
('channel.mention_everyone', 'channel', 'Mention Everyone', 'Use @everyone and @here', FALSE),
('channel.manage_messages', 'channel', 'Manage Messages', 'Delete any message, pin messages', FALSE),
('channel.manage_threads', 'channel', 'Manage Threads', 'Manage thread settings', FALSE),

-- Voice Permissions
('voice.connect', 'voice', 'Connect', 'Join voice channels', TRUE),
('voice.speak', 'voice', 'Speak', 'Talk in voice channels', TRUE),
('voice.mute_members', 'voice', 'Mute Members', 'Server mute other members', FALSE),
('voice.deafen_members', 'voice', 'Deafen Members', 'Server deafen other members', FALSE),
('voice.move_members', 'voice', 'Move Members', 'Move members between channels', FALSE),
('voice.priority_speaker', 'voice', 'Priority Speaker', 'Be heard over others', FALSE)

ON CONFLICT (id) DO NOTHING;

-- ═══════════════════════════════════════════════════════════════════════════
-- Triggers for updated_at
-- ═══════════════════════════════════════════════════════════════════════════

CREATE OR REPLACE FUNCTION update_roles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER roles_updated_at_trigger
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_roles_updated_at();

-- Comments
COMMENT ON TABLE roles IS 'Custom roles with hierarchical permissions';
COMMENT ON TABLE permission_definitions IS 'All available permissions in the system';
COMMENT ON TABLE role_permissions IS 'Maps permissions to roles';
COMMENT ON TABLE user_roles IS 'Maps roles to users';
COMMENT ON TABLE channel_permission_overrides IS 'Per-channel permission overrides for roles/users';
