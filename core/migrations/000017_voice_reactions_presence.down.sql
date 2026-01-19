-- GoConnect Migration: Remove voice states, reactions, and presence
-- Version: 17 (down)

DROP INDEX IF EXISTS idx_presence_last_seen;
DROP INDEX IF EXISTS idx_presence_status;
DROP TABLE IF EXISTS user_presence;

DROP INDEX IF EXISTS idx_voice_states_channel;
DROP TABLE IF EXISTS voice_states;

DROP INDEX IF EXISTS idx_reactions_user;
DROP INDEX IF EXISTS idx_reactions_message;
DROP TABLE IF EXISTS reactions;
