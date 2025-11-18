-- Rollback Migration: Remove farm_count field and revert trigger function

-- ============================================================================
-- PART 1: Revert trigger function to original (without farm_count)
-- ============================================================================

CREATE OR REPLACE FUNCTION update_farmer_total_acreage()
RETURNS TRIGGER AS $$
BEGIN
    -- For INSERT and UPDATE, update the farmer's total acreage only
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        UPDATE farmers
        SET total_acreage_ha = (
            SELECT COALESCE(SUM(area_ha_computed), 0)
            FROM farms
            WHERE farmer_id = NEW.farmer_id
            AND deleted_at IS NULL
        )
        WHERE id = NEW.farmer_id;

        -- If this is an UPDATE and farmer_id changed, update old farmer too
        IF (TG_OP = 'UPDATE' AND OLD.farmer_id != NEW.farmer_id) THEN
            UPDATE farmers
            SET total_acreage_ha = (
                SELECT COALESCE(SUM(area_ha_computed), 0)
                FROM farms
                WHERE farmer_id = OLD.farmer_id
                AND deleted_at IS NULL
            )
            WHERE id = OLD.farmer_id;
        END IF;
    END IF;

    -- For DELETE, update the old farmer's total acreage
    IF (TG_OP = 'DELETE') THEN
        UPDATE farmers
        SET total_acreage_ha = (
            SELECT COALESCE(SUM(area_ha_computed), 0)
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
-- PART 2: Drop farm_count column and index
-- ============================================================================

-- Drop index
DROP INDEX IF EXISTS idx_farmers_farm_count;

-- Remove comment
COMMENT ON COLUMN farmers.farm_count IS NULL;

-- Drop column
ALTER TABLE farmers DROP COLUMN IF EXISTS farm_count;

-- ============================================================================
-- NOTE
-- ============================================================================
-- After rolling back this migration, you must also rollback the code changes
-- that reference the farm_count field in the Farmer entity.
