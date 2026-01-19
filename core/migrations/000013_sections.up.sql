-- GoConnect Migration: Add sections table
-- Version: 13
-- Date: 2026-01-19
-- Description: Sections allow organizing large servers into sub-groups
--              Each section can have its own moderators and channels

-- ═══════════════════════════════════════════════════════════════════════════
-- SECTION (Alt-Server / Bölüm) - Organizing large servers
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    icon            VARCHAR(255),           -- Emoji or URL
    position        INTEGER NOT NULL DEFAULT 0,

    -- Visibility
    visibility      VARCHAR(20) NOT NULL DEFAULT 'visible',  -- visible, hidden, archived

    -- Timestamps
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMP,

    -- Constraints
    CONSTRAINT sections_visibility_check CHECK (visibility IN ('visible', 'hidden', 'archived')),
    CONSTRAINT sections_name_length CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
    UNIQUE(tenant_id, name)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_sections_tenant ON sections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sections_position ON sections(tenant_id, position);
CREATE INDEX IF NOT EXISTS idx_sections_visibility ON sections(tenant_id, visibility) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_sections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sections_updated_at_trigger
    BEFORE UPDATE ON sections
    FOR EACH ROW
    EXECUTE FUNCTION update_sections_updated_at();

-- Comments for documentation
COMMENT ON TABLE sections IS 'Sections organize channels within a tenant (like Discord categories)';
COMMENT ON COLUMN sections.visibility IS 'visible = shown to all, hidden = only for authorized, archived = read-only';
COMMENT ON COLUMN sections.position IS 'Display order (lower = higher in list)';
