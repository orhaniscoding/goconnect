-- GoConnect Migration: Add channels table
-- Version: 14
-- Date: 2026-01-19
-- Description: Channels provide text/voice communication within tenants
--              Channels can belong to tenant, section, or network

-- ═══════════════════════════════════════════════════════════════════════════
-- CHANNEL (Text/Voice Channels)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Hierarchy (exactly one must be set)
    tenant_id       UUID REFERENCES tenants(id) ON DELETE CASCADE,
    section_id      UUID REFERENCES sections(id) ON DELETE CASCADE,
    network_id      UUID REFERENCES networks(id) ON DELETE CASCADE,

    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    type            VARCHAR(20) NOT NULL DEFAULT 'text',   -- 'text', 'voice', 'announcement'
    position        INTEGER NOT NULL DEFAULT 0,

    -- Voice channel settings
    bitrate         INTEGER DEFAULT 64000,  -- 64kbps default
    user_limit      INTEGER DEFAULT 0,      -- 0 = unlimited

    -- Moderation
    slowmode        INTEGER DEFAULT 0,      -- Seconds between messages (0 = off)
    nsfw            BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,

    -- Constraints
    CONSTRAINT channels_type_check CHECK (type IN ('text', 'voice', 'announcement')),
    CONSTRAINT channels_name_length CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
    CONSTRAINT channels_bitrate_range CHECK (bitrate >= 8000 AND bitrate <= 384000),
    CONSTRAINT channels_user_limit_range CHECK (user_limit >= 0 AND user_limit <= 99),
    CONSTRAINT channels_slowmode_range CHECK (slowmode >= 0 AND slowmode <= 21600),
    -- Exactly one parent must be set
    CONSTRAINT channels_has_parent CHECK (
        (CASE WHEN tenant_id IS NOT NULL THEN 1 ELSE 0 END +
         CASE WHEN section_id IS NOT NULL THEN 1 ELSE 0 END +
         CASE WHEN network_id IS NOT NULL THEN 1 ELSE 0 END) = 1
    )
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_channels_tenant ON channels(tenant_id) WHERE tenant_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_section ON channels(section_id) WHERE section_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_network ON channels(network_id) WHERE network_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_channels_position ON channels(COALESCE(tenant_id, section_id, network_id), position);
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_channels_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER channels_updated_at_trigger
    BEFORE UPDATE ON channels
    FOR EACH ROW
    EXECUTE FUNCTION update_channels_updated_at();

-- Comments for documentation
COMMENT ON TABLE channels IS 'Text and voice channels for communication';
COMMENT ON COLUMN channels.type IS 'text = chat, voice = audio, announcement = read-only broadcasts';
COMMENT ON COLUMN channels.bitrate IS 'Voice quality in bps (64000 = 64kbps)';
COMMENT ON COLUMN channels.slowmode IS 'Minimum seconds between user messages (0 = disabled)';
COMMENT ON COLUMN channels.nsfw IS 'Age-restricted content warning';
