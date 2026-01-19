-- GoConnect Migration: Remove messages table
-- Version: 16 (down)

DROP INDEX IF EXISTS idx_messages_content_search;
DROP INDEX IF EXISTS idx_messages_mentions;
DROP INDEX IF EXISTS idx_messages_pinned;
DROP INDEX IF EXISTS idx_messages_reply;
DROP INDEX IF EXISTS idx_messages_thread;
DROP INDEX IF EXISTS idx_messages_author;
DROP INDEX IF EXISTS idx_messages_channel;
DROP TABLE IF EXISTS messages;
