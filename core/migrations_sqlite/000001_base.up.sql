-- SQLite baseline schema for zero-config mode
-- Simplified types (TEXT/INTEGER/BOOLEAN) to maximize compatibility with modernc.org/sqlite
-- Triggers/extensions from Postgres are omitted; timestamps default to CURRENT_TIMESTAMP

PRAGMA foreign_keys = ON;

-- Tenants
CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    icon_url TEXT DEFAULT '',
    visibility TEXT DEFAULT 'private',
    access_type TEXT DEFAULT 'invite_only',
    password_hash TEXT,
    max_members INTEGER DEFAULT 0,
    member_count INTEGER DEFAULT 0,
    owner_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tenants_visibility ON tenants(visibility);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    locale TEXT DEFAULT 'en',
    is_admin BOOLEAN DEFAULT FALSE,
    is_moderator BOOLEAN DEFAULT FALSE,
    two_fa_key TEXT,
    two_fa_enabled BOOLEAN DEFAULT FALSE,
    recovery_codes TEXT,
    auth_provider TEXT,
    external_id TEXT,
    display_name TEXT,
    role TEXT DEFAULT 'user',
    totp_secret TEXT,
    totp_enabled BOOLEAN DEFAULT FALSE,
    recovery_used INTEGER DEFAULT 0,
    last_login_at DATETIME,
    username TEXT,
    full_name TEXT,
    bio TEXT,
    avatar_url TEXT,
    suspended BOOLEAN DEFAULT FALSE,
    suspended_at DATETIME,
    suspended_reason TEXT,
    suspended_by TEXT,
    last_seen DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, email)
);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_is_moderator ON users(is_moderator);
CREATE INDEX IF NOT EXISTS idx_users_two_fa_enabled ON users(two_fa_enabled);
CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider);
CREATE INDEX IF NOT EXISTS idx_users_suspended ON users(suspended);
CREATE INDEX IF NOT EXISTS idx_users_created_at_desc ON users(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen DESC);

-- Networks
CREATE TABLE IF NOT EXISTS networks (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    cidr TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    join_policy TEXT NOT NULL DEFAULT 'approval',
    dns TEXT,
    mtu INTEGER,
    split_tunnel BOOLEAN,
    moderation_redacted BOOLEAN DEFAULT FALSE,
    created_by TEXT NOT NULL,
    description TEXT DEFAULT '',
    required_role TEXT DEFAULT 'member',
    is_hidden BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE(tenant_id, name)
);
CREATE INDEX IF NOT EXISTS idx_networks_tenant_id ON networks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_networks_created_by ON networks(created_by);
CREATE INDEX IF NOT EXISTS idx_networks_visibility ON networks(visibility);
CREATE INDEX IF NOT EXISTS idx_networks_deleted_at ON networks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_networks_created_at ON networks(created_at);

-- Memberships
CREATE TABLE IF NOT EXISTS memberships (
    id TEXT PRIMARY KEY,
    network_id TEXT NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active',
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(network_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_memberships_network_id ON memberships(network_id);
CREATE INDEX IF NOT EXISTS idx_memberships_user_id ON memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_memberships_status ON memberships(status);
CREATE INDEX IF NOT EXISTS idx_memberships_role ON memberships(role);

-- Join requests
CREATE TABLE IF NOT EXISTS join_requests (
    id TEXT PRIMARY KEY,
    network_id TEXT NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    reviewed_at DATETIME,
    reviewed_by TEXT
);
CREATE INDEX IF NOT EXISTS idx_join_requests_network_id ON join_requests(network_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_user_id ON join_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_join_requests_status ON join_requests(status);

-- IP allocations (IPAM)
CREATE TABLE IF NOT EXISTS ip_allocations (
    id TEXT PRIMARY KEY,
    network_id TEXT NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address TEXT NOT NULL,
    allocated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(network_id, ip_address),
    UNIQUE(network_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_ip_allocations_network_id ON ip_allocations(network_id);
CREATE INDEX IF NOT EXISTS idx_ip_allocations_user_id ON ip_allocations(user_id);
CREATE INDEX IF NOT EXISTS idx_ip_allocations_ip_address ON ip_allocations(ip_address);

-- Idempotency keys
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key TEXT PRIMARY KEY,
    response_body TEXT NOT NULL,
    response_status INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_created_at ON idempotency_keys(created_at);

-- Audit events
CREATE TABLE IF NOT EXISTS audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id TEXT,
    event_type TEXT NOT NULL,
    actor_id TEXT,
    object_type TEXT,
    object_id TEXT,
    network_id TEXT,
    metadata TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    prev_hash TEXT,
    hash TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_id ON audit_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_events_network_id ON audit_events(network_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);

-- Chat messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    scope TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    body TEXT NOT NULL,
    attachments TEXT DEFAULT '[]',
    redacted BOOLEAN DEFAULT FALSE,
    deleted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_scope_created ON chat_messages(scope, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_created ON chat_messages(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_tenant ON chat_messages(tenant_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created ON chat_messages(created_at DESC);

CREATE TABLE IF NOT EXISTS chat_message_edits (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    prev_body TEXT NOT NULL,
    new_body TEXT NOT NULL,
    editor_id TEXT NOT NULL,
    edited_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chat_edits_message ON chat_message_edits(message_id, edited_at ASC);

-- Devices
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    platform TEXT NOT NULL,
    pubkey TEXT NOT NULL UNIQUE,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT FALSE,
    ip_address TEXT,
    daemon_ver TEXT,
    os_version TEXT,
    hostname TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    disabled_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_pubkey ON devices(pubkey);
CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(active, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_devices_platform ON devices(platform);

-- Peers
CREATE TABLE IF NOT EXISTS peers (
    id TEXT PRIMARY KEY,
    network_id TEXT NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,
    preshared_key TEXT,
    endpoint TEXT,
    allowed_ips TEXT NOT NULL,
    persistent_keepalive INTEGER DEFAULT 0,
    last_handshake DATETIME,
    rx_bytes INTEGER DEFAULT 0,
    tx_bytes INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    disabled_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_peers_network_id ON peers(network_id);
CREATE INDEX IF NOT EXISTS idx_peers_device_id ON peers(device_id);
CREATE INDEX IF NOT EXISTS idx_peers_tenant_id ON peers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_peers_public_key ON peers(public_key);
CREATE INDEX IF NOT EXISTS idx_peers_active ON peers(network_id, active);
CREATE INDEX IF NOT EXISTS idx_peers_last_handshake ON peers(last_handshake DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_network_device_unique ON peers(network_id, device_id);

-- Recovery codes
CREATE TABLE IF NOT EXISTS recovery_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_recovery_codes_user ON recovery_codes(user_id);

-- Invite tokens
CREATE TABLE IF NOT EXISTS invite_tokens (
    id TEXT PRIMARY KEY,
    network_id TEXT NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    expires_at DATETIME,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_network ON invite_tokens(network_id);

-- IP rules
CREATE TABLE IF NOT EXISTS ip_rules (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    cidr TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ip_rules_tenant ON ip_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ip_rules_action ON ip_rules(action);

-- Tenant multi-membership
CREATE TABLE IF NOT EXISTS tenant_members (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    nickname TEXT,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    banned_at DATETIME,
    banned_by TEXT,
    UNIQUE(tenant_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant ON tenant_members(tenant_id);

CREATE TABLE IF NOT EXISTS tenant_invites (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    max_uses INTEGER DEFAULT 1,
    use_count INTEGER DEFAULT 0,
    revoked_at DATETIME,
    expires_at DATETIME,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_code ON tenant_invites(code);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant ON tenant_invites(tenant_id);

CREATE TABLE IF NOT EXISTS tenant_announcements (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author_id TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tenant_announcements_tenant ON tenant_announcements(tenant_id);

CREATE TABLE IF NOT EXISTS tenant_chat_messages (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    edited_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_tenant_chat_tenant ON tenant_chat_messages(tenant_id, created_at);

-- Posts
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    image_url TEXT,
    likes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);
