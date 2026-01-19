-- GoConnect Migration: Remove sections table
-- Version: 13 (down)

DROP TRIGGER IF EXISTS sections_updated_at_trigger ON sections;
DROP FUNCTION IF EXISTS update_sections_updated_at();
DROP INDEX IF EXISTS idx_sections_visibility;
DROP INDEX IF EXISTS idx_sections_position;
DROP INDEX IF EXISTS idx_sections_tenant;
DROP TABLE IF EXISTS sections;
