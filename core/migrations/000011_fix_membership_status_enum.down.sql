-- Rollback: Revert membership status enum fix
-- Version: 11

-- Step 1: Update all 'approved' status back to 'active'
UPDATE memberships SET status = 'active' WHERE status = 'approved';

-- Step 2: Drop the new CHECK constraint
ALTER TABLE memberships DROP CONSTRAINT IF EXISTS memberships_status_check;

-- Step 3: Add old CHECK constraint with 'active' instead of 'approved'
ALTER TABLE memberships ADD CONSTRAINT memberships_status_check 
    CHECK (status IN ('active', 'banned', 'pending'));
