# Database Schema Design - Crop Cycles Area Allocation

## Overview

This document provides the complete database schema changes required for implementing crop cycles area allocation and stage-based farm activity organization.

## Schema Changes Summary

1. **crop_cycles** - Add area_ha field for area allocation
2. **farm_activities** - Add crop_stage_id field for stage association
3. **farm_area_allocations** - New table for tracking area allocations (optional optimization)
4. Indexes, constraints, and triggers for data integrity

## Detailed Schema Changes

### 1. Crop Cycles Table Modifications

#### Add area_ha Column

```sql
-- Step 1: Add column as nullable initially
ALTER TABLE crop_cycles
ADD COLUMN area_ha DECIMAL(12,4);

-- Step 2: Add comment for documentation
COMMENT ON COLUMN crop_cycles.area_ha IS 'Allocated area in hectares for this crop cycle';

-- Step 3: After data migration, add constraints
ALTER TABLE crop_cycles
ADD CONSTRAINT chk_crop_cycles_positive_area
CHECK (area_ha IS NULL OR area_ha > 0);

-- Step 4: Create indexes for performance
CREATE INDEX idx_crop_cycles_farm_area
ON crop_cycles(farm_id, status, area_ha)
WHERE deleted_at IS NULL;

CREATE INDEX idx_crop_cycles_area_allocation
ON crop_cycles(farm_id, area_ha)
WHERE status IN ('PLANNED', 'ACTIVE') AND deleted_at IS NULL;
```

#### Composite Constraint for Area Validation

```sql
-- Create a function to validate total area allocation
CREATE OR REPLACE FUNCTION validate_crop_cycle_area()
RETURNS TRIGGER AS $$
DECLARE
    v_farm_area DECIMAL(12,4);
    v_total_allocated DECIMAL(12,4);
BEGIN
    -- Get farm's total area
    SELECT area_ha_computed INTO v_farm_area
    FROM farms
    WHERE id = NEW.farm_id AND deleted_at IS NULL;

    -- Calculate total allocated area (including the new/updated cycle)
    SELECT COALESCE(SUM(area_ha), 0) INTO v_total_allocated
    FROM crop_cycles
    WHERE farm_id = NEW.farm_id
        AND id != NEW.id
        AND status IN ('PLANNED', 'ACTIVE')
        AND deleted_at IS NULL;

    -- Add the current cycle's area
    v_total_allocated := v_total_allocated + COALESCE(NEW.area_ha, 0);

    -- Validate total allocation doesn't exceed farm area
    IF v_total_allocated > v_farm_area THEN
        RAISE EXCEPTION 'Total allocated area (%) exceeds farm area (%) for farm %',
            v_total_allocated, v_farm_area, NEW.farm_id
            USING ERRCODE = '23514'; -- check_violation
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for insert and update
CREATE TRIGGER trg_validate_crop_cycle_area
BEFORE INSERT OR UPDATE OF area_ha, status ON crop_cycles
FOR EACH ROW
WHEN (NEW.area_ha IS NOT NULL AND NEW.status IN ('PLANNED', 'ACTIVE'))
EXECUTE FUNCTION validate_crop_cycle_area();
```

### 2. Farm Activities Table Modifications

#### Add crop_stage_id Column

```sql
-- Step 1: Add column as nullable
ALTER TABLE farm_activities
ADD COLUMN crop_stage_id VARCHAR(20);

-- Step 2: Add foreign key constraint
ALTER TABLE farm_activities
ADD CONSTRAINT fk_farm_activities_crop_stage
FOREIGN KEY (crop_stage_id)
REFERENCES crop_stages(id)
ON DELETE SET NULL;

-- Step 3: Create indexes for performance
CREATE INDEX idx_farm_activities_crop_stage
ON farm_activities(crop_stage_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_farm_activities_cycle_stage
ON farm_activities(crop_cycle_id, crop_stage_id)
WHERE deleted_at IS NULL;

-- Step 4: Add comment for documentation
COMMENT ON COLUMN farm_activities.crop_stage_id IS 'Reference to the crop growth stage this activity belongs to';
```

#### Validation Function for Stage Association

```sql
-- Function to validate that stage belongs to the crop cycle's crop
CREATE OR REPLACE FUNCTION validate_activity_stage()
RETURNS TRIGGER AS $$
DECLARE
    v_crop_id VARCHAR(20);
    v_stage_crop_id VARCHAR(20);
BEGIN
    -- Skip validation if no stage specified
    IF NEW.crop_stage_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- Get crop_id from the crop cycle
    SELECT c.crop_id INTO v_crop_id
    FROM crop_cycles cc
    JOIN crops c ON cc.crop_id = c.id
    WHERE cc.id = NEW.crop_cycle_id;

    -- Get crop_id from the crop stage
    SELECT crop_id INTO v_stage_crop_id
    FROM crop_stages
    WHERE id = NEW.crop_stage_id;

    -- Validate they match
    IF v_crop_id != v_stage_crop_id THEN
        RAISE EXCEPTION 'Stage % does not belong to the crop of cycle %',
            NEW.crop_stage_id, NEW.crop_cycle_id
            USING ERRCODE = '23503'; -- foreign_key_violation
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER trg_validate_activity_stage
BEFORE INSERT OR UPDATE OF crop_stage_id, crop_cycle_id ON farm_activities
FOR EACH ROW
EXECUTE FUNCTION validate_activity_stage();
```

### 3. Farm Area Allocations Table (Optional Optimization)

This table provides a denormalized view for faster area allocation queries and implements optimistic locking.

```sql
-- Create area allocations tracking table
CREATE TABLE farm_area_allocations (
    id VARCHAR(20) PRIMARY KEY DEFAULT generate_id('FAAL', 'medium'),
    farm_id VARCHAR(20) NOT NULL,
    total_area_ha DECIMAL(12,4) NOT NULL,
    allocated_area_ha DECIMAL(12,4) NOT NULL DEFAULT 0,
    available_area_ha DECIMAL(12,4) GENERATED ALWAYS AS (total_area_ha - allocated_area_ha) STORED,
    active_cycles_count INTEGER NOT NULL DEFAULT 0,
    planned_cycles_count INTEGER NOT NULL DEFAULT 0,
    last_validated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key
    CONSTRAINT fk_farm_area_allocations_farm
    FOREIGN KEY (farm_id) REFERENCES farms(id) ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT chk_farm_area_allocations_positive
    CHECK (total_area_ha > 0 AND allocated_area_ha >= 0),

    CONSTRAINT chk_farm_area_allocations_not_exceed
    CHECK (allocated_area_ha <= total_area_ha)
);

-- Unique index on farm_id
CREATE UNIQUE INDEX idx_farm_area_allocations_farm
ON farm_area_allocations(farm_id)
WHERE deleted_at IS NULL;

-- Index for queries
CREATE INDEX idx_farm_area_allocations_available
ON farm_area_allocations(available_area_ha)
WHERE deleted_at IS NULL;
```

#### Triggers to Maintain Area Allocations

```sql
-- Function to update area allocations on crop cycle changes
CREATE OR REPLACE FUNCTION update_farm_area_allocation()
RETURNS TRIGGER AS $$
BEGIN
    -- Handle INSERT
    IF TG_OP = 'INSERT' AND NEW.area_ha IS NOT NULL AND NEW.status IN ('PLANNED', 'ACTIVE') THEN
        INSERT INTO farm_area_allocations (farm_id, total_area_ha, allocated_area_ha)
        SELECT
            NEW.farm_id,
            f.area_ha_computed,
            COALESCE(SUM(cc.area_ha), 0)
        FROM farms f
        LEFT JOIN crop_cycles cc ON f.id = cc.farm_id
            AND cc.status IN ('PLANNED', 'ACTIVE')
            AND cc.deleted_at IS NULL
        WHERE f.id = NEW.farm_id
        GROUP BY f.id, f.area_ha_computed
        ON CONFLICT (farm_id) WHERE deleted_at IS NULL
        DO UPDATE SET
            allocated_area_ha = EXCLUDED.allocated_area_ha,
            version = farm_area_allocations.version + 1,
            last_validated_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP;
    END IF;

    -- Handle UPDATE
    IF TG_OP = 'UPDATE' THEN
        -- Recalculate if area or status changed
        IF (OLD.area_ha IS DISTINCT FROM NEW.area_ha) OR
           (OLD.status IS DISTINCT FROM NEW.status) THEN
            UPDATE farm_area_allocations
            SET allocated_area_ha = (
                SELECT COALESCE(SUM(cc.area_ha), 0)
                FROM crop_cycles cc
                WHERE cc.farm_id = NEW.farm_id
                    AND cc.status IN ('PLANNED', 'ACTIVE')
                    AND cc.deleted_at IS NULL
            ),
            version = version + 1,
            last_validated_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
            WHERE farm_id = NEW.farm_id;
        END IF;
    END IF;

    -- Handle DELETE
    IF TG_OP = 'DELETE' AND OLD.area_ha IS NOT NULL AND OLD.status IN ('PLANNED', 'ACTIVE') THEN
        UPDATE farm_area_allocations
        SET allocated_area_ha = allocated_area_ha - OLD.area_ha,
            version = version + 1,
            last_validated_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE farm_id = OLD.farm_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER trg_update_farm_area_allocation
AFTER INSERT OR UPDATE OR DELETE ON crop_cycles
FOR EACH ROW
EXECUTE FUNCTION update_farm_area_allocation();
```

### 4. Materialized Views for Performance

#### Farm Area Summary View

```sql
-- Materialized view for quick area summaries
CREATE MATERIALIZED VIEW mv_farm_area_summary AS
SELECT
    f.id as farm_id,
    f.farmer_id,
    f.name as farm_name,
    f.area_ha_computed as total_area_ha,
    COALESCE(alloc.allocated_area, 0) as allocated_area_ha,
    f.area_ha_computed - COALESCE(alloc.allocated_area, 0) as available_area_ha,
    COALESCE(alloc.active_cycles, 0) as active_cycles_count,
    COALESCE(alloc.planned_cycles, 0) as planned_cycles_count,
    COALESCE(alloc.crops, '[]'::jsonb) as allocated_crops
FROM farms f
LEFT JOIN LATERAL (
    SELECT
        SUM(cc.area_ha) FILTER (WHERE cc.status IN ('PLANNED', 'ACTIVE')) as allocated_area,
        COUNT(*) FILTER (WHERE cc.status = 'ACTIVE') as active_cycles,
        COUNT(*) FILTER (WHERE cc.status = 'PLANNED') as planned_cycles,
        jsonb_agg(
            jsonb_build_object(
                'cycle_id', cc.id,
                'crop_name', c.name,
                'variety_name', cv.name,
                'area_ha', cc.area_ha,
                'status', cc.status,
                'season', cc.season
            ) ORDER BY cc.area_ha DESC
        ) FILTER (WHERE cc.status IN ('PLANNED', 'ACTIVE')) as crops
    FROM crop_cycles cc
    LEFT JOIN crops c ON cc.crop_id = c.id
    LEFT JOIN crop_varieties cv ON cc.variety_id = cv.id
    WHERE cc.farm_id = f.id AND cc.deleted_at IS NULL
) alloc ON true
WHERE f.deleted_at IS NULL;

-- Indexes for the materialized view
CREATE UNIQUE INDEX idx_mv_farm_area_summary_farm ON mv_farm_area_summary(farm_id);
CREATE INDEX idx_mv_farm_area_summary_available ON mv_farm_area_summary(available_area_ha);
CREATE INDEX idx_mv_farm_area_summary_farmer ON mv_farm_area_summary(farmer_id);

-- Refresh function
CREATE OR REPLACE FUNCTION refresh_farm_area_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_farm_area_summary;
END;
$$ LANGUAGE plpgsql;
```

#### Stage Activity Progress View

```sql
-- Materialized view for stage-wise activity progress
CREATE MATERIALIZED VIEW mv_stage_activity_progress AS
SELECT
    cc.id as crop_cycle_id,
    cs.id as crop_stage_id,
    cs.stage_order,
    s.stage_name,
    cs.duration_days,
    cs.duration_unit,
    COUNT(fa.id) as total_activities,
    COUNT(fa.id) FILTER (WHERE fa.status = 'COMPLETED') as completed_activities,
    COUNT(fa.id) FILTER (WHERE fa.status = 'IN_PROGRESS') as in_progress_activities,
    COUNT(fa.id) FILTER (WHERE fa.status = 'PLANNED') as planned_activities,
    ROUND(
        CASE
            WHEN COUNT(fa.id) > 0 THEN
                (COUNT(fa.id) FILTER (WHERE fa.status = 'COMPLETED')::numeric / COUNT(fa.id)::numeric) * 100
            ELSE 0
        END, 2
    ) as completion_percentage,
    MIN(fa.completed_at) as first_activity_completed,
    MAX(fa.completed_at) as last_activity_completed
FROM crop_cycles cc
JOIN crop_stages cs ON cc.crop_id = cs.crop_id
JOIN stages s ON cs.stage_id = s.id
LEFT JOIN farm_activities fa ON cc.id = fa.crop_cycle_id
    AND fa.crop_stage_id = cs.id
    AND fa.deleted_at IS NULL
WHERE cc.deleted_at IS NULL
    AND cs.is_active = true
    AND cc.status IN ('ACTIVE', 'COMPLETED')
GROUP BY cc.id, cs.id, cs.stage_order, s.stage_name, cs.duration_days, cs.duration_unit;

-- Indexes
CREATE INDEX idx_mv_stage_activity_cycle ON mv_stage_activity_progress(crop_cycle_id);
CREATE INDEX idx_mv_stage_activity_stage ON mv_stage_activity_progress(crop_stage_id);
CREATE INDEX idx_mv_stage_activity_order ON mv_stage_activity_progress(crop_cycle_id, stage_order);
```

### 5. Performance Indexes

```sql
-- Additional indexes for common queries

-- For finding cycles by area range
CREATE INDEX idx_crop_cycles_area_range
ON crop_cycles(area_ha)
WHERE status IN ('PLANNED', 'ACTIVE') AND deleted_at IS NULL;

-- For area allocation queries with farmer
CREATE INDEX idx_crop_cycles_farmer_area
ON crop_cycles(farmer_id, farm_id, area_ha)
WHERE status IN ('PLANNED', 'ACTIVE') AND deleted_at IS NULL;

-- For activity stage queries with date range
CREATE INDEX idx_farm_activities_stage_date
ON farm_activities(crop_stage_id, planned_at)
WHERE deleted_at IS NULL;

-- For finding activities by stage and status
CREATE INDEX idx_farm_activities_stage_status
ON farm_activities(crop_stage_id, status)
WHERE deleted_at IS NULL;

-- Partial index for active allocations
CREATE INDEX idx_crop_cycles_active_allocation
ON crop_cycles(farm_id, area_ha)
WHERE status = 'ACTIVE' AND area_ha IS NOT NULL AND deleted_at IS NULL;
```

### 6. Database Functions for Common Operations

#### Get Available Area Function

```sql
CREATE OR REPLACE FUNCTION get_available_farm_area(p_farm_id VARCHAR(20))
RETURNS DECIMAL(12,4) AS $$
DECLARE
    v_total_area DECIMAL(12,4);
    v_allocated_area DECIMAL(12,4);
BEGIN
    -- Get farm's total area
    SELECT area_ha_computed INTO v_total_area
    FROM farms
    WHERE id = p_farm_id AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Farm % not found', p_farm_id;
    END IF;

    -- Get allocated area
    SELECT COALESCE(SUM(area_ha), 0) INTO v_allocated_area
    FROM crop_cycles
    WHERE farm_id = p_farm_id
        AND status IN ('PLANNED', 'ACTIVE')
        AND deleted_at IS NULL;

    RETURN v_total_area - v_allocated_area;
END;
$$ LANGUAGE plpgsql STABLE;
```

#### Check Area Availability Function

```sql
CREATE OR REPLACE FUNCTION check_area_availability(
    p_farm_id VARCHAR(20),
    p_requested_area DECIMAL(12,4),
    p_exclude_cycle_id VARCHAR(20) DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_available_area DECIMAL(12,4);
    v_current_cycle_area DECIMAL(12,4);
BEGIN
    -- Get available area
    v_available_area := get_available_farm_area(p_farm_id);

    -- If updating existing cycle, add back its current allocation
    IF p_exclude_cycle_id IS NOT NULL THEN
        SELECT area_ha INTO v_current_cycle_area
        FROM crop_cycles
        WHERE id = p_exclude_cycle_id
            AND farm_id = p_farm_id
            AND status IN ('PLANNED', 'ACTIVE')
            AND deleted_at IS NULL;

        v_available_area := v_available_area + COALESCE(v_current_cycle_area, 0);
    END IF;

    RETURN p_requested_area <= v_available_area;
END;
$$ LANGUAGE plpgsql STABLE;
```

### 7. Migration Scripts

#### Phase 1: Initial Schema Changes

```sql
-- Migration: 2025_01_14_add_area_allocation.sql

BEGIN;

-- Add area_ha to crop_cycles
ALTER TABLE crop_cycles ADD COLUMN IF NOT EXISTS area_ha DECIMAL(12,4);

-- Add crop_stage_id to farm_activities
ALTER TABLE farm_activities ADD COLUMN IF NOT EXISTS crop_stage_id VARCHAR(20);

-- Create farm_area_allocations table
CREATE TABLE IF NOT EXISTS farm_area_allocations (
    -- table definition as above
);

-- Create functions
-- Create all functions as defined above

-- Create triggers
-- Create all triggers as defined above

-- Create indexes
-- Create all indexes as defined above

COMMIT;
```

#### Phase 2: Data Migration

```sql
-- Migration: 2025_01_15_migrate_existing_data.sql

BEGIN;

-- Set default area for existing active cycles (equal distribution)
WITH farm_cycle_counts AS (
    SELECT
        farm_id,
        COUNT(*) as cycle_count,
        MAX(f.area_ha_computed) as farm_area
    FROM crop_cycles cc
    JOIN farms f ON cc.farm_id = f.id
    WHERE cc.status IN ('PLANNED', 'ACTIVE')
        AND cc.area_ha IS NULL
        AND cc.deleted_at IS NULL
    GROUP BY farm_id
)
UPDATE crop_cycles cc
SET area_ha = ROUND(fcc.farm_area / fcc.cycle_count, 4)
FROM farm_cycle_counts fcc
WHERE cc.farm_id = fcc.farm_id
    AND cc.status IN ('PLANNED', 'ACTIVE')
    AND cc.area_ha IS NULL
    AND cc.deleted_at IS NULL;

-- Initialize farm_area_allocations table
INSERT INTO farm_area_allocations (farm_id, total_area_ha, allocated_area_ha)
SELECT
    f.id,
    f.area_ha_computed,
    COALESCE(SUM(cc.area_ha), 0)
FROM farms f
LEFT JOIN crop_cycles cc ON f.id = cc.farm_id
    AND cc.status IN ('PLANNED', 'ACTIVE')
    AND cc.deleted_at IS NULL
WHERE f.deleted_at IS NULL
GROUP BY f.id, f.area_ha_computed
ON CONFLICT (farm_id) WHERE deleted_at IS NULL DO NOTHING;

-- Create materialized views
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_farm_area_summary AS
-- view definition as above

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_stage_activity_progress AS
-- view definition as above

-- Refresh views
REFRESH MATERIALIZED VIEW mv_farm_area_summary;
REFRESH MATERIALIZED VIEW mv_stage_activity_progress;

COMMIT;
```

#### Phase 3: Apply Constraints

```sql
-- Migration: 2025_01_16_apply_constraints.sql

BEGIN;

-- Make area_ha required for new cycles (after ensuring all existing have values)
ALTER TABLE crop_cycles
ALTER COLUMN area_ha SET NOT NULL;

-- Add foreign key for crop_stage_id
ALTER TABLE farm_activities
ADD CONSTRAINT fk_farm_activities_crop_stage
FOREIGN KEY (crop_stage_id)
REFERENCES crop_stages(id)
ON DELETE SET NULL;

-- Enable triggers
ALTER TABLE crop_cycles ENABLE TRIGGER trg_validate_crop_cycle_area;
ALTER TABLE farm_activities ENABLE TRIGGER trg_validate_activity_stage;

COMMIT;
```

### 8. Rollback Scripts

```sql
-- Rollback: Remove area allocation features

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_validate_crop_cycle_area ON crop_cycles;
DROP TRIGGER IF EXISTS trg_validate_activity_stage ON farm_activities;
DROP TRIGGER IF EXISTS trg_update_farm_area_allocation ON crop_cycles;

-- Drop functions
DROP FUNCTION IF EXISTS validate_crop_cycle_area();
DROP FUNCTION IF EXISTS validate_activity_stage();
DROP FUNCTION IF EXISTS update_farm_area_allocation();
DROP FUNCTION IF EXISTS get_available_farm_area(VARCHAR);
DROP FUNCTION IF EXISTS check_area_availability(VARCHAR, DECIMAL, VARCHAR);

-- Drop materialized views
DROP MATERIALIZED VIEW IF EXISTS mv_stage_activity_progress;
DROP MATERIALIZED VIEW IF EXISTS mv_farm_area_summary;

-- Drop indexes
DROP INDEX IF EXISTS idx_crop_cycles_farm_area;
DROP INDEX IF EXISTS idx_crop_cycles_area_allocation;
DROP INDEX IF EXISTS idx_farm_activities_crop_stage;
DROP INDEX IF EXISTS idx_farm_activities_cycle_stage;

-- Drop table
DROP TABLE IF EXISTS farm_area_allocations;

-- Remove columns
ALTER TABLE farm_activities DROP COLUMN IF EXISTS crop_stage_id;
ALTER TABLE crop_cycles DROP COLUMN IF EXISTS area_ha;

COMMIT;
```

## Database Performance Considerations

### Query Optimization Tips

1. **Use partial indexes** for status-based queries
2. **Leverage materialized views** for complex aggregations
3. **Implement connection pooling** for concurrent operations
4. **Use EXPLAIN ANALYZE** to optimize slow queries
5. **Consider partitioning** for large activity tables

### Maintenance Tasks

```sql
-- Weekly maintenance script
-- Run as scheduled job

-- Update statistics
ANALYZE crop_cycles;
ANALYZE farm_activities;
ANALYZE farm_area_allocations;

-- Refresh materialized views
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_farm_area_summary;
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_stage_activity_progress;

-- Reindex if needed (during maintenance window)
REINDEX INDEX CONCURRENTLY idx_crop_cycles_farm_area;
REINDEX INDEX CONCURRENTLY idx_farm_activities_crop_stage;
```

## Security Considerations

1. **Row-Level Security** - Consider implementing RLS for multi-tenant isolation
2. **Audit Triggers** - Add triggers to log all area modifications
3. **Permission Checks** - Ensure application validates permissions before DB operations
4. **SQL Injection Prevention** - Use parameterized queries exclusively

## Monitoring Queries

```sql
-- Monitor area allocation usage
SELECT
    COUNT(DISTINCT farm_id) as farms_using_allocation,
    AVG(allocated_area_ha / total_area_ha * 100) as avg_utilization_percent,
    MAX(allocated_area_ha / total_area_ha * 100) as max_utilization_percent,
    COUNT(*) FILTER (WHERE allocated_area_ha = total_area_ha) as fully_allocated_farms
FROM farm_area_allocations
WHERE deleted_at IS NULL;

-- Monitor trigger performance
SELECT
    schemaname,
    tablename,
    tgname as trigger_name,
    tgenabled as enabled,
    tgtype
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_tables pt ON c.relname = pt.tablename
WHERE tablename IN ('crop_cycles', 'farm_activities')
ORDER BY tablename, trigger_name;
```
