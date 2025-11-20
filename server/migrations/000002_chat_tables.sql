-- Migration: Create chat tables
-- Version: 000002
-- Description: Chat messages with edit history and soft delete support

-- +goose Up
-- +goose StatementBegin

-- Chat messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    scope TEXT NOT NULL,  -- "host" or "network:<id>"
    tenant_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    body TEXT NOT NULL,
    attachments JSONB DEFAULT '[]'::jsonb,
    redacted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_chat_messages_scope_created 
    ON chat_messages(scope, created_at DESC) 
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_chat_messages_user_created 
    ON chat_messages(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_chat_messages_tenant 
    ON chat_messages(tenant_id);

-- Index for cursor pagination
CREATE INDEX IF NOT EXISTS idx_chat_messages_created 
    ON chat_messages(created_at DESC);

-- Chat message edit history
CREATE TABLE IF NOT EXISTS chat_message_edits (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    prev_body TEXT NOT NULL,
    new_body TEXT NOT NULL,
    editor_id TEXT NOT NULL,
    edited_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fetching edit history
CREATE INDEX IF NOT EXISTS idx_chat_edits_message 
    ON chat_message_edits(message_id, edited_at ASC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_chat_edits_message;
DROP TABLE IF EXISTS chat_message_edits;

DROP INDEX IF EXISTS idx_chat_messages_created;
DROP INDEX IF EXISTS idx_chat_messages_tenant;
DROP INDEX IF EXISTS idx_chat_messages_user_created;
DROP INDEX IF EXISTS idx_chat_messages_scope_created;
DROP TABLE IF EXISTS chat_messages;

-- +goose StatementEnd
