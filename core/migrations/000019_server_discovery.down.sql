-- GoConnect Migration: Remove server discovery
-- Version: 19 (down)

DROP INDEX IF EXISTS idx_vanity_code;
DROP TABLE IF EXISTS server_vanity_urls;

DROP INDEX IF EXISTS idx_discovery_tags;
DROP INDEX IF EXISTS idx_discovery_member_count;
DROP INDEX IF EXISTS idx_discovery_featured;
DROP INDEX IF EXISTS idx_discovery_category;
DROP INDEX IF EXISTS idx_discovery_enabled;
DROP TABLE IF EXISTS server_discovery;
