-- 004_drop_aaa_farmer_user_id_column.sql
-- Migration to remove redundant aaa_farmer_user_id column from farms table
-- This column is replaced by the farmer_id foreign key reference

-- Drop the index on aaa_farmer_user_id first (if it exists)
DROP INDEX IF EXISTS farms_farmer_id_idx;

-- Drop the aaa_farmer_user_id column from farms table
-- Note: This assumes farmer_id column has been added and populated via migration 003
ALTER TABLE farms
DROP COLUMN IF EXISTS aaa_farmer_user_id;

-- Add comment documenting the change
COMMENT ON COLUMN farms.farmer_id IS 'Foreign key reference to farmers.id - replaces deprecated aaa_farmer_user_id';

-- Verify that farmer_id column exists and is properly indexed
-- The index should have been created in migration 003
DO $$
BEGIN
    -- Check if farmer_id column exists
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'farms'
        AND column_name = 'farmer_id'
    ) THEN
        RAISE EXCEPTION 'farmer_id column must exist before dropping aaa_farmer_user_id. Run migration 003 first.';
    END IF;

    -- Check if all farms have farmer_id populated
    IF EXISTS (
        SELECT 1
        FROM farms
        WHERE farmer_id IS NULL
    ) THEN
        RAISE WARNING 'Some farms have NULL farmer_id. Please populate farmer_id before proceeding.';
    END IF;
END $$;
