-- Rollback Migration: Sync ID column with aaa_org_id
-- WARNING: Rolling back this migration will break the application!
-- The application code expects aaa_org_id to be used as the primary identifier

-- ============================================================================
-- PART 1: Rollback fpo_configs changes
-- ============================================================================

-- Rename aaa_org_id back to fpo_id
ALTER TABLE fpo_configs RENAME COLUMN aaa_org_id TO fpo_id;

-- Recreate original index
DROP INDEX IF EXISTS idx_fpo_configs_aaa_org_id;
CREATE UNIQUE INDEX idx_fpo_configs_fpo_id ON fpo_configs(fpo_id) WHERE deleted_at IS NULL;

-- Remove comments
COMMENT ON COLUMN fpo_configs.id IS NULL;
COMMENT ON COLUMN fpo_configs.fpo_id IS NULL;

-- ============================================================================
-- PART 2: Rollback fpo_refs changes
-- ============================================================================

-- Remove comments
COMMENT ON COLUMN fpo_refs.id IS NULL;
COMMENT ON COLUMN fpo_refs.aaa_org_id IS NULL;

-- Note: We don't change the ID values back because:
-- 1. The old ID values were generated IDs that we don't have a way to restore
-- 2. The application code now expects ID = aaa_org_id
-- 3. Rolling back this migration without rolling back code changes will break the application

-- ============================================================================
-- POST-ROLLBACK WARNING
-- ============================================================================

-- After rolling back this migration, you MUST also:
-- 1. Rollback the application code changes
-- 2. Regenerate unique IDs for the id column if needed
-- 3. Update all foreign key references
