-- Rollback device_id column addition
-- Note: SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- This is a simplified rollback that just removes the index
-- Full rollback would require table recreation which is complex for SQLite
DROP INDEX IF EXISTS idx_ip_allocations_device_id;

-- Clear device_id values (column will remain but be unused)
UPDATE ip_allocations SET device_id = NULL;
