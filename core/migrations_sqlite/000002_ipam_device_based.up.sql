-- Add device_id column to ip_allocations for device-based IP allocation
-- This matches the Postgres migration 000012_ipam_device_based

ALTER TABLE ip_allocations ADD COLUMN device_id TEXT DEFAULT NULL;

-- Create index on device_id for efficient lookups
CREATE INDEX IF NOT EXISTS idx_ip_allocations_device_id ON ip_allocations(device_id);

-- Copy existing user_id values to device_id for backward compatibility
UPDATE ip_allocations SET device_id = user_id WHERE device_id IS NULL;
