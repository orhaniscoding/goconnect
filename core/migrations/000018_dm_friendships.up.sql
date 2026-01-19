-- GoConnect Migration: DM channels and friendships
-- Version: 18
-- Date: 2026-01-19
-- Description: Direct messaging and friend system

-- ═══════════════════════════════════════════════════════════════════════════
-- FRIENDSHIPS (User-to-user relationships)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS friendships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, accepted, blocked

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    accepted_at     TIMESTAMP,

    -- Constraints
    CONSTRAINT friendships_status_check CHECK (status IN ('pending', 'accepted', 'blocked')),
    CONSTRAINT friendships_no_self CHECK (user_id != friend_id),
    UNIQUE(user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friendships_user ON friendships(user_id, status);
CREATE INDEX IF NOT EXISTS idx_friendships_friend ON friendships(friend_id, status);
CREATE INDEX IF NOT EXISTS idx_friendships_accepted ON friendships(user_id) WHERE status = 'accepted';

-- ═══════════════════════════════════════════════════════════════════════════
-- DM CHANNELS (Direct message conversations)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS dm_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            VARCHAR(20) NOT NULL DEFAULT 'dm',  -- 'dm' (1:1), 'group_dm' (multi-user)

    -- Group DM only
    name            VARCHAR(100),
    icon            VARCHAR(255),
    owner_id        UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dm_type_check CHECK (type IN ('dm', 'group_dm')),
    CONSTRAINT dm_group_requires_owner CHECK (type != 'group_dm' OR owner_id IS NOT NULL)
);

-- ═══════════════════════════════════════════════════════════════════════════
-- DM CHANNEL MEMBERS (Who is in each DM channel)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS dm_channel_members (
    channel_id      UUID NOT NULL REFERENCES dm_channels(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Per-user settings
    muted           BOOLEAN DEFAULT FALSE,

    joined_at       TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (channel_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_dm_members_user ON dm_channel_members(user_id);
CREATE INDEX IF NOT EXISTS idx_dm_members_channel ON dm_channel_members(channel_id);

-- ═══════════════════════════════════════════════════════════════════════════
-- DM MESSAGES (Messages in DM channels - stored separately for E2E encryption)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS dm_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id      UUID NOT NULL REFERENCES dm_channels(id) ON DELETE CASCADE,
    author_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- E2E encrypted content (server cannot read)
    content         TEXT NOT NULL,          -- Encrypted payload
    encrypted       BOOLEAN NOT NULL DEFAULT TRUE,

    -- Attachments (encrypted metadata)
    attachments     JSONB DEFAULT '[]',

    -- Reply support
    reply_to_id     UUID REFERENCES dm_messages(id) ON DELETE SET NULL,

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at       TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dm_messages_channel ON dm_messages(channel_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_dm_messages_author ON dm_messages(author_id) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE friendships IS 'User friend relationships (pending, accepted, blocked)';
COMMENT ON TABLE dm_channels IS 'Direct message channels (1:1 or group)';
COMMENT ON TABLE dm_channel_members IS 'Members of DM channels';
COMMENT ON TABLE dm_messages IS 'E2E encrypted direct messages';
COMMENT ON COLUMN dm_messages.content IS 'E2E encrypted - server cannot read';
