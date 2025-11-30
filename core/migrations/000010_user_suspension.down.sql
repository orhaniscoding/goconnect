-- Remove indexes
DROP INDEX IF EXISTS idx_users_username_search;
DROP INDEX IF EXISTS idx_users_email_search;
DROP INDEX IF EXISTS idx_users_last_seen;
DROP INDEX IF EXISTS idx_users_created_at_desc;
DROP INDEX IF EXISTS idx_users_suspended;
DROP INDEX IF EXISTS idx_users_is_moderator;
DROP INDEX IF EXISTS idx_users_is_admin;

-- Remove columns
ALTER TABLE users 
DROP COLUMN IF EXISTS username,
DROP COLUMN IF EXISTS is_moderator,
DROP COLUMN IF EXISTS last_seen,
DROP COLUMN IF EXISTS suspended_by,
DROP COLUMN IF EXISTS suspended_reason,
DROP COLUMN IF EXISTS suspended_at,
DROP COLUMN IF EXISTS suspended;
