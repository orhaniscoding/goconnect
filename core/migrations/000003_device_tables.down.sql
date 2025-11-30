-- Drop indexes
DROP INDEX IF EXISTS idx_devices_platform;
DROP INDEX IF EXISTS idx_devices_active;
DROP INDEX IF EXISTS idx_devices_pubkey;
DROP INDEX IF EXISTS idx_devices_tenant;
DROP INDEX IF EXISTS idx_devices_user;

-- Drop devices table
DROP TABLE IF EXISTS devices;
