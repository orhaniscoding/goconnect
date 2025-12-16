-- GoConnect Migration: Add device_id to ip_allocations
-- Version: 12
-- Date: 2025-12-13
-- Description: Changes IPAM model from user-based to device-based allocation
--              This allows multiple devices per user to have different IPs

-- Step 1: Add device_id column (nullable initially for migration)
ALTER TABLE ip_allocations ADD COLUMN IF NOT EXISTS device_id UUID;

-- Step 2: Create index for device-based lookups
CREATE INDEX IF NOT EXISTS idx_ip_allocations_device ON ip_allocations(network_id, device_id);

-- Step 3: Update existing allocations to link to user's first device (if any)
-- This is a best-effort migration - assumes one device per user in old data
UPDATE ip_allocations ip
SET device_id = (
    SELECT d.id 
    FROM devices d 
    WHERE d.user_id = ip.user_id 
    ORDER BY d.created_at ASC 
    LIMIT 1
)
WHERE ip.device_id IS NULL;

-- Step 4: Add unique constraint for network + device (replaces network + user uniqueness)
-- Keep user_id for backwards compatibility, but new allocations should use device_id
ALTER TABLE ip_allocations DROP CONSTRAINT IF EXISTS ip_allocations_network_user_unique;
ALTER TABLE ip_allocations ADD CONSTRAINT ip_allocations_network_device_unique 
    UNIQUE (network_id, device_id);

-- Step 5: Add foreign key constraint to devices table
-- Only if all device_ids are valid (may fail if orphaned allocations exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM ip_allocations WHERE device_id IS NOT NULL 
        AND device_id NOT IN (SELECT id FROM devices)
    ) THEN
        ALTER TABLE ip_allocations ADD CONSTRAINT ip_allocations_device_fk 
            FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add comment for documentation
COMMENT ON COLUMN ip_allocations.device_id IS 'Device that owns this IP allocation (replaces user_id for multi-device support)';
COMMENT ON COLUMN ip_allocations.user_id IS 'DEPRECATED: Use device_id instead. Kept for migration/audit purposes.';
