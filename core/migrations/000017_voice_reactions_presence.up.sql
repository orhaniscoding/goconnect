-- GoConnect Migration: Voice states, reactions, and presence
-- Version: 17
-- Date: 2026-01-19
-- Description: Real-time features for voice channels, emoji reactions, and user presence

-- ═══════════════════════════════════════════════════════════════════════════
-- REACTIONS (Emoji reactions on messages)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS reactions (
    message_id      UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji           VARCHAR(100) NOT NULL,      -- Unicode emoji or custom emoji ID
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_reactions_message ON reactions(message_id);
CREATE INDEX IF NOT EXISTS idx_reactions_user ON reactions(user_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- VOICE STATES (Who is in which voice channel)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS voice_states (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,

    -- User-controlled states
    self_mute       BOOLEAN DEFAULT FALSE,
    self_deaf       BOOLEAN DEFAULT FALSE,

    -- Server-controlled states (by moderators)
    server_mute     BOOLEAN DEFAULT FALSE,
    server_deaf     BOOLEAN DEFAULT FALSE,

    -- Connection info
    connected_at    TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT voice_channel_type CHECK (
        EXISTS (SELECT 1 FROM channels WHERE id = channel_id AND type = 'voice')
    )
);

CREATE INDEX IF NOT EXISTS idx_voice_states_channel ON voice_states(channel_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- USER PRESENCE (Online status and activity)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS user_presence (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'offline',  -- online, idle, dnd, invisible, offline
    custom_status   VARCHAR(128),

    -- Activity
    activity_type   VARCHAR(50),        -- playing, listening, watching, streaming
    activity_name   VARCHAR(128),

    -- Last seen
    last_seen       TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Per-device status
    desktop_status  VARCHAR(20),
    mobile_status   VARCHAR(20),
    web_status      VARCHAR(20),

    -- Constraints
    CONSTRAINT presence_status_check CHECK (status IN ('online', 'idle', 'dnd', 'invisible', 'offline')),
    CONSTRAINT presence_activity_check CHECK (activity_type IS NULL OR activity_type IN ('playing', 'listening', 'watching', 'streaming'))
);

CREATE INDEX IF NOT EXISTS idx_presence_status ON user_presence(status) WHERE status != 'offline';
CREATE INDEX IF NOT EXISTS idx_presence_last_seen ON user_presence(last_seen);

-- Comments
COMMENT ON TABLE reactions IS 'Emoji reactions on messages';
COMMENT ON TABLE voice_states IS 'Current voice channel connections (one per user)';
COMMENT ON TABLE user_presence IS 'User online status and activity';
COMMENT ON COLUMN user_presence.status IS 'online = active, idle = away, dnd = do not disturb, invisible = hidden online';
