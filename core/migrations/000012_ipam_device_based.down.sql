-- Rollback: Revert device-based IPAM
-- Version: 12

-- Step 1: Remove foreign key constraint
ALTER TABLE ip_allocations DROP CONSTRAINT IF EXISTS ip_allocations_device_fk;

-- Step 2: Remove unique constraint
ALTER TABLE ip_allocations DROP CONSTRAINT IF EXISTS ip_allocations_network_device_unique;

-- Step 3: Restore old unique constraint (network + user)
ALTER TABLE ip_allocations ADD CONSTRAINT ip_allocations_network_user_unique 
    UNIQUE (network_id, user_id);

-- Step 4: Drop device index
DROP INDEX IF EXISTS idx_ip_allocations_device;

-- Step 5: Drop device_id column
ALTER TABLE ip_allocations DROP COLUMN IF EXISTS device_id;
