-- GoConnect Migration: Fix Membership Status Enum
-- Version: 11
-- Date: 2025-12-13
-- Description: Changes membership status 'active' to 'approved' to match domain model

-- Step 1: Update all existing 'active' status to 'approved'
UPDATE memberships SET status = 'approved' WHERE status = 'active';

-- Step 2: Drop the old CHECK constraint
ALTER TABLE memberships DROP CONSTRAINT IF EXISTS memberships_status_check;

-- Step 3: Add new CHECK constraint with 'approved' instead of 'active'
ALTER TABLE memberships ADD CONSTRAINT memberships_status_check 
    CHECK (status IN ('approved', 'banned', 'pending'));

-- Add comment for documentation
COMMENT ON COLUMN memberships.status IS 'Membership status: approved (active member), banned (removed), pending (awaiting approval)';
