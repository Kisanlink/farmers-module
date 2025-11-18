-- Migration: Sync ID column with aaa_org_id for FPO tables
-- Purpose: Ensure ID column always contains the same value as aaa_org_id
-- This allows using aaa_org_id as a natural primary key while maintaining BaseModel structure

-- ============================================================================
-- PART 1: Update fpo_refs table
-- ============================================================================

-- Update existing fpo_refs records to sync ID with aaa_org_id
UPDATE fpo_refs
SET id = aaa_org_id
WHERE id != aaa_org_id OR id IS NULL;

-- Add constraint to ensure ID and aaa_org_id are always in sync (PostgreSQL 12+)
-- This is enforced programmatically via BeforeCreate/BeforeUpdate hooks
-- but we add a comment for documentation
COMMENT ON COLUMN fpo_refs.id IS 'Primary key - always synced with aaa_org_id via BeforeCreate/BeforeUpdate hooks';
COMMENT ON COLUMN fpo_refs.aaa_org_id IS 'AAA Organization ID - business key that is copied to id column';

-- ============================================================================
-- PART 2: Update fpo_configs table
-- ============================================================================

-- Rename fpo_id to aaa_org_id
ALTER TABLE fpo_configs RENAME COLUMN fpo_id TO aaa_org_id;

-- Update existing fpo_configs records to sync ID with aaa_org_id
UPDATE fpo_configs
SET id = aaa_org_id
WHERE id != aaa_org_id OR id IS NULL;

-- Update index on aaa_org_id
DROP INDEX IF EXISTS idx_fpo_configs_fpo_id;
CREATE UNIQUE INDEX idx_fpo_configs_aaa_org_id ON fpo_configs(aaa_org_id) WHERE deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON COLUMN fpo_configs.id IS 'Primary key - always synced with aaa_org_id via BeforeCreate/BeforeUpdate hooks';
COMMENT ON COLUMN fpo_configs.aaa_org_id IS 'AAA Organization ID - business key that is copied to id column';

-- ============================================================================
-- VERIFICATION QUERIES (commented out - for manual verification)
-- ============================================================================

-- Verify fpo_refs sync
-- SELECT COUNT(*) as mismatched_count FROM fpo_refs WHERE id != aaa_org_id;
-- Expected: 0

-- Verify fpo_configs sync
-- SELECT COUNT(*) as mismatched_count FROM fpo_configs WHERE id != aaa_org_id;
-- Expected: 0

-- View sample data
-- SELECT id, aaa_org_id, name FROM fpo_refs LIMIT 5;
-- SELECT id, aaa_org_id, fpo_name FROM fpo_configs LIMIT 5;
