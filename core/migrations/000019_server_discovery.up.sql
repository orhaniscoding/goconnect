-- GoConnect Migration: Server discovery
-- Version: 19
-- Date: 2026-01-19
-- Description: Server discovery for finding and joining public servers

-- ═══════════════════════════════════════════════════════════════════════════
-- SERVER DISCOVERY (Public server listing and search)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS server_discovery (
    tenant_id       UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,

    -- Discovery settings
    enabled         BOOLEAN DEFAULT FALSE,

    -- Categorization
    category        VARCHAR(50),            -- gaming, education, music, tech, etc.
    tags            VARCHAR(50)[] DEFAULT '{}',

    -- Description
    short_description VARCHAR(300),

    -- Cached statistics (updated periodically)
    member_count    INTEGER DEFAULT 0,
    online_count    INTEGER DEFAULT 0,

    -- Featured/verified status
    featured        BOOLEAN DEFAULT FALSE,
    verified        BOOLEAN DEFAULT FALSE,

    -- Timestamps
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT discovery_category_check CHECK (
        category IS NULL OR category IN (
            'gaming', 'education', 'music', 'tech', 'art',
            'science', 'entertainment', 'sports', 'finance',
            'crypto', 'social', 'community', 'other'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_discovery_enabled ON server_discovery(enabled) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_discovery_category ON server_discovery(category) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_discovery_featured ON server_discovery(featured) WHERE enabled = TRUE AND featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_discovery_member_count ON server_discovery(member_count DESC) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_discovery_tags ON server_discovery USING GIN(tags) WHERE enabled = TRUE;

-- ═══════════════════════════════════════════════════════════════════════════
-- SERVER VANITY URLS (Custom URLs for premium servers)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS server_vanity_urls (
    tenant_id       UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    vanity_code     VARCHAR(32) NOT NULL UNIQUE,    -- e.g., "minecraft-tr"
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT vanity_code_format CHECK (vanity_code ~ '^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$')
);

CREATE INDEX IF NOT EXISTS idx_vanity_code ON server_vanity_urls(vanity_code);

-- Comments
COMMENT ON TABLE server_discovery IS 'Public server listing for discovery';
COMMENT ON TABLE server_vanity_urls IS 'Custom short URLs for servers (e.g., goconnect.io/g/minecraft-tr)';
COMMENT ON COLUMN server_discovery.category IS 'Primary category for filtering';
COMMENT ON COLUMN server_discovery.tags IS 'Additional tags for search (max 5 recommended)';
