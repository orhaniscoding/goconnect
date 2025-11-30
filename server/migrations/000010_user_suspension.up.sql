-- Add suspension tracking fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS suspended BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS suspended_reason TEXT,
ADD COLUMN IF NOT EXISTS suspended_by VARCHAR(255);

-- Add last_seen column for activity tracking
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS last_seen TIMESTAMP;

-- Add is_moderator and username columns if not exist
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS is_moderator BOOLEAN DEFAULT FALSE NOT NULL,
ADD COLUMN IF NOT EXISTS username VARCHAR(255);

-- Add indexes for admin queries
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin) WHERE is_admin = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_is_moderator ON users(is_moderator) WHERE is_moderator = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_suspended ON users(suspended) WHERE suspended = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_created_at_desc ON users(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_users_email_search ON users USING gin(to_tsvector('english', email));
CREATE INDEX IF NOT EXISTS idx_users_username_search ON users USING gin(to_tsvector('english', COALESCE(username, '')));

-- Add comments for documentation
COMMENT ON COLUMN users.suspended IS 'Whether the user account is suspended';
COMMENT ON COLUMN users.suspended_at IS 'When the user was suspended';
COMMENT ON COLUMN users.suspended_reason IS 'Reason for suspension';
COMMENT ON COLUMN users.suspended_by IS 'User ID of the admin who suspended this user';
COMMENT ON COLUMN users.last_seen IS 'Last time the user was active';
