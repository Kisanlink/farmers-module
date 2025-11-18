-- Migration: Add farm_count rollup field to farmers table
-- Purpose: Track the number of farms owned by each farmer automatically via database triggers

-- ============================================================================
-- PART 1: Add farm_count column to farmers table
-- ============================================================================

-- Add farm_count column with default value
ALTER TABLE farmers
ADD COLUMN IF NOT EXISTS farm_count INTEGER NOT NULL DEFAULT 0;

-- Add index for fast filtering and sorting by farm count
CREATE INDEX IF NOT EXISTS idx_farmers_farm_count ON farmers (farm_count);

-- Add comment for documentation
COMMENT ON COLUMN farmers.farm_count IS 'Number of farms owned by this farmer - maintained automatically by database triggers';

-- ============================================================================
-- PART 2: Update trigger function to maintain farm_count
-- ============================================================================

-- Drop existing trigger function and recreate with farm_count support
CREATE OR REPLACE FUNCTION update_farmer_total_acreage()
RETURNS TRIGGER AS $$
BEGIN
    -- For INSERT and UPDATE, update the farmer's total acreage and farm count
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        UPDATE farmers
        SET
            total_acreage_ha = (
                SELECT COALESCE(SUM(area_ha_computed), 0)
                FROM farms
                WHERE farmer_id = NEW.farmer_id
                AND deleted_at IS NULL
            ),
            farm_count = (
                SELECT COUNT(*)
                FROM farms
                WHERE farmer_id = NEW.farmer_id
                AND deleted_at IS NULL
            )
        WHERE id = NEW.farmer_id;

        -- If this is an UPDATE and farmer_id changed, update old farmer too
        IF (TG_OP = 'UPDATE' AND OLD.farmer_id != NEW.farmer_id) THEN
            UPDATE farmers
            SET
                total_acreage_ha = (
                    SELECT COALESCE(SUM(area_ha_computed), 0)
                    FROM farms
                    WHERE farmer_id = OLD.farmer_id
                    AND deleted_at IS NULL
                ),
                farm_count = (
                    SELECT COUNT(*)
                    FROM farms
                    WHERE farmer_id = OLD.farmer_id
                    AND deleted_at IS NULL
                )
            WHERE id = OLD.farmer_id;
        END IF;
    END IF;

    -- For DELETE, update the old farmer's total acreage and farm count
    IF (TG_OP = 'DELETE') THEN
        UPDATE farmers
        SET
            total_acreage_ha = (
                SELECT COALESCE(SUM(area_ha_computed), 0)
                FROM farms
                WHERE farmer_id = OLD.farmer_id
                AND deleted_at IS NULL
            ),
            farm_count = (
                SELECT COUNT(*)
                FROM farms
                WHERE farmer_id = OLD.farmer_id
                AND deleted_at IS NULL
            )
        WHERE id = OLD.farmer_id;
        RETURN OLD;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PART 3: Backfill existing data
-- ============================================================================

-- Update all existing farmers with their current farm counts
UPDATE farmers
SET farm_count = COALESCE(
    (SELECT COUNT(*)
     FROM farms
     WHERE farms.farmer_id = farmers.id
     AND farms.deleted_at IS NULL),
    0
);

-- ============================================================================
-- VERIFICATION QUERIES (commented out - for manual verification)
-- ============================================================================

-- Verify farm counts are correct
-- SELECT
--     f.id,
--     f.first_name,
--     f.last_name,
--     f.farm_count,
--     (SELECT COUNT(*) FROM farms WHERE farmer_id = f.id AND deleted_at IS NULL) as actual_count
-- FROM farmers f
-- WHERE f.farm_count != (SELECT COUNT(*) FROM farms WHERE farmer_id = f.id AND deleted_at IS NULL);
-- Expected: 0 rows (all counts should match)

-- View farmers with their farm counts
-- SELECT id, first_name, last_name, total_acreage_ha, farm_count
-- FROM farmers
-- ORDER BY farm_count DESC
-- LIMIT 10;
