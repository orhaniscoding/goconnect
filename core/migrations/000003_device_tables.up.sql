-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('windows', 'macos', 'linux', 'android', 'ios')),
    pubkey TEXT NOT NULL UNIQUE,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active BOOLEAN NOT NULL DEFAULT FALSE,
    ip_address TEXT,
    daemon_ver TEXT,
    os_version TEXT,
    hostname TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disabled_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_devices_tenant ON devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_pubkey ON devices(pubkey);
CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(active, last_seen DESC) WHERE disabled_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_devices_platform ON devices(platform);
