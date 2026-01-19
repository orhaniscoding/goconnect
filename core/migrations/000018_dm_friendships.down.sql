-- GoConnect Migration: Remove DM channels and friendships
-- Version: 18 (down)

DROP INDEX IF EXISTS idx_dm_messages_author;
DROP INDEX IF EXISTS idx_dm_messages_channel;
DROP TABLE IF EXISTS dm_messages;

DROP INDEX IF EXISTS idx_dm_members_channel;
DROP INDEX IF EXISTS idx_dm_members_user;
DROP TABLE IF EXISTS dm_channel_members;

DROP TABLE IF EXISTS dm_channels;

DROP INDEX IF EXISTS idx_friendships_accepted;
DROP INDEX IF EXISTS idx_friendships_friend;
DROP INDEX IF EXISTS idx_friendships_user;
DROP TABLE IF EXISTS friendships;
