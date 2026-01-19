-- GoConnect Migration: Remove channels table
-- Version: 14 (down)

DROP TRIGGER IF EXISTS channels_updated_at_trigger ON channels;
DROP FUNCTION IF EXISTS update_channels_updated_at();
DROP INDEX IF EXISTS idx_channels_type;
DROP INDEX IF EXISTS idx_channels_position;
DROP INDEX IF EXISTS idx_channels_network;
DROP INDEX IF EXISTS idx_channels_section;
DROP INDEX IF EXISTS idx_channels_tenant;
DROP TABLE IF EXISTS channels;
