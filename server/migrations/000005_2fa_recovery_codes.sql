-- GoConnect 2FA Recovery Codes Migration
-- Version: 5.0
-- Author: orhaniscoding
-- Date: 2025-11-25

-- Add 2FA columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_fa_key VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_fa_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_codes TEXT[]; -- Array of hashed recovery codes
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_moderator BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- Create index for 2FA enabled users
CREATE INDEX IF NOT EXISTS idx_users_two_fa_enabled ON users(two_fa_enabled) WHERE two_fa_enabled = TRUE;

-- Create index for auth provider lookups
CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider) WHERE auth_provider IS NOT NULL;

COMMENT ON COLUMN users.two_fa_key IS 'TOTP secret key for 2FA';
COMMENT ON COLUMN users.two_fa_enabled IS 'Whether 2FA is enabled for this user';
COMMENT ON COLUMN users.recovery_codes IS 'Hashed one-time recovery codes (8 codes, Argon2id hashed)';
COMMENT ON COLUMN users.is_moderator IS 'Whether user can moderate chat/content';
COMMENT ON COLUMN users.auth_provider IS 'Authentication provider: local, google, github, oidc';
COMMENT ON COLUMN users.external_id IS 'External ID from auth provider';
