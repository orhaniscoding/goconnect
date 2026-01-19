-- GoConnect Migration: Remove roles and permissions tables
-- Version: 15 (down)

DROP TRIGGER IF EXISTS roles_updated_at_trigger ON roles;
DROP FUNCTION IF EXISTS update_roles_updated_at();

DROP INDEX IF EXISTS idx_channel_overrides_user;
DROP INDEX IF EXISTS idx_channel_overrides_role;
DROP INDEX IF EXISTS idx_channel_overrides_channel;
DROP TABLE IF EXISTS channel_permission_overrides;

DROP INDEX IF EXISTS idx_user_roles_role;
DROP INDEX IF EXISTS idx_user_roles_user;
DROP TABLE IF EXISTS user_roles;

DROP INDEX IF EXISTS idx_role_permissions_role;
DROP TABLE IF EXISTS role_permissions;

DROP INDEX IF EXISTS idx_roles_default;
DROP INDEX IF EXISTS idx_roles_position;
DROP INDEX IF EXISTS idx_roles_tenant;
DROP TABLE IF EXISTS roles;

DROP TABLE IF EXISTS permission_definitions;
