-- Remove 2FA columns
ALTER TABLE users DROP COLUMN IF EXISTS external_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS is_moderator;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_codes;
ALTER TABLE users DROP COLUMN IF EXISTS two_fa_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS two_fa_key;

DROP INDEX IF EXISTS idx_users_auth_provider;
DROP INDEX IF EXISTS idx_users_two_fa_enabled;
