-- GoConnect Migration: Enhanced messages table
-- Version: 16
-- Date: 2026-01-19
-- Description: Channel-based messaging with replies, threads, mentions, and attachments

-- ═══════════════════════════════════════════════════════════════════════════
-- MESSAGES (Channel-based messaging)
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id      UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    author_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    content         TEXT NOT NULL,

    -- Reply/Thread support
    reply_to_id     UUID REFERENCES messages(id) ON DELETE SET NULL,
    thread_id       UUID REFERENCES messages(id) ON DELETE SET NULL,    -- Thread parent message

    -- Attachments (JSON array of file metadata)
    attachments     JSONB DEFAULT '[]',

    -- Embeds (link previews, rich content)
    embeds          JSONB DEFAULT '[]',

    -- Mentions (arrays of UUIDs)
    mentions        UUID[] DEFAULT '{}',             -- @user mentions
    mention_roles   UUID[] DEFAULT '{}',             -- @role mentions
    mention_everyone BOOLEAN DEFAULT FALSE,

    -- Flags
    pinned          BOOLEAN DEFAULT FALSE,
    edited_at       TIMESTAMP,

    -- E2E encryption (for DMs)
    encrypted       BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_author ON messages(author_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_thread ON messages(thread_id) WHERE thread_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_reply ON messages(reply_to_id) WHERE reply_to_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_pinned ON messages(channel_id, pinned) WHERE pinned = TRUE AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_messages_mentions ON messages USING GIN(mentions) WHERE deleted_at IS NULL;

-- Full-text search index for message content
CREATE INDEX IF NOT EXISTS idx_messages_content_search ON messages USING GIN(to_tsvector('english', content)) WHERE deleted_at IS NULL AND encrypted = FALSE;

-- Comments
COMMENT ON TABLE messages IS 'Channel messages with rich features like replies, threads, and attachments';
COMMENT ON COLUMN messages.attachments IS 'JSON array: [{id, filename, size, content_type, url}]';
COMMENT ON COLUMN messages.embeds IS 'JSON array: [{type, title, description, url, thumbnail}]';
COMMENT ON COLUMN messages.encrypted IS 'True if content is E2E encrypted (for DMs)';
