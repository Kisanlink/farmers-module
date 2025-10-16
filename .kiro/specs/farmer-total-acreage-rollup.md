# Farmer Total Acreage Rollup Design

## Problem Statement

Currently, there is no efficient way to query the total acreage owned by a farmer across all their farms. Calculating this requires:
- Joining the `farms` table
- Summing the `area_ha_computed` column for each farmer
- This is inefficient for large datasets and frequent queries

We need a denormalized rollup column on the `farmers` table that maintains the total acreage automatically using database triggers.

## Proposed Solution: Trigger-Based Rollup

### Database Schema Changes

#### 1. Add `total_acreage_ha` column to `farmers` table

```sql
ALTER TABLE farmers
ADD COLUMN total_acreage_ha NUMERIC(12,4) DEFAULT 0.0 NOT NULL;

-- Add index for fast filtering/sorting by total acreage
CREATE INDEX idx_farmers_total_acreage ON farmers(total_acreage_ha);

-- Ensure farms.farmer_id is indexed for fast aggregation
CREATE INDEX IF NOT EXISTS idx_farms_farmer_id ON farms(farmer_id);
```

#### 2. Create trigger function to maintain rollup

```sql
-- Function to update farmer's total acreage
CREATE OR REPLACE FUNCTION update_farmer_total_acreage()
RETURNS TRIGGER AS $$
BEGIN
    -- For INSERT and UPDATE, update the farmer's total acreage
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
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for INSERT, UPDATE, DELETE
CREATE TRIGGER trigger_farm_insert_update_farmer_acreage
AFTER INSERT OR UPDATE OF farmer_id, geometry, deleted_at
ON farms
FOR EACH ROW
EXECUTE FUNCTION update_farmer_total_acreage();

CREATE TRIGGER trigger_farm_delete_update_farmer_acreage
AFTER DELETE
ON farms
FOR EACH ROW
EXECUTE FUNCTION update_farmer_total_acreage();
```

### Entity Design (Go)

Update the `Farmer` entity in `internal/entities/farmer/farmer_normalized.go`:

```go
type Farmer struct {
    base.BaseModel

    // AAA Integration (External System IDs)
    AAAUserID        string  `json:"aaa_user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
    AAAOrgID         string  `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
    KisanSathiUserID *string `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`

    // Personal Information
    FirstName   string  `json:"first_name" gorm:"type:varchar(255);not null"`
    LastName    string  `json:"last_name" gorm:"type:varchar(255);not null"`
    PhoneNumber string  `json:"phone_number" gorm:"type:varchar(50)"`
    Email       string  `json:"email" gorm:"type:varchar(255)"`
    DateOfBirth *string `json:"date_of_birth" gorm:"type:date"`
    Gender      string  `json:"gender" gorm:"type:varchar(50)"`

    // Address (Normalized via Foreign Key)
    AddressID *string  `json:"address_id" gorm:"type:varchar(255)"`
    Address   *Address `json:"address,omitempty" gorm:"foreignKey:AddressID;constraint:OnDelete:SET NULL"`

    // Additional Fields
    LandOwnershipType string `json:"land_ownership_type" gorm:"type:varchar(100)"`
    SocialCategory    string `json:"social_category" gorm:"type:varchar(50)"`
    AreaType          string `json:"area_type" gorm:"type:varchar(50)"`
    Status            string `json:"status" gorm:"type:farmer_status;not null;default:'ACTIVE'"`

    // Rollup Fields - Maintained by triggers
    TotalAcreageHa float64 `json:"total_acreage_ha" gorm:"type:numeric(12,4);default:0.0;not null;index"`

    // Relationships
    FPOLinkages []*FarmerLink `json:"fpo_linkages,omitempty" gorm:"foreignKey:AAAUserID,AAAOrgID;references:AAAUserID,AAAOrgID"`
    Farms       []FarmRef     `json:"farms,omitempty" gorm:"foreignKey:FarmerID;references:ID"`

    // Flexible Data (JSONB for extensibility)
    Preferences entities.JSONB `json:"preferences" gorm:"type:jsonb;default:'{}'::jsonb;serializer:json"`
    Metadata    entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}'::jsonb;serializer:json"`
}
```

## Performance Optimization

### 1. Index Strategy
- **farms.farmer_id**: Already indexed in `setupPostMigration()` (line 319 in db.go)
- **farmers.total_acreage_ha**: New index for fast filtering/sorting

### 2. Delta Updates (Future Optimization)
For very large datasets, consider implementing delta-based updates:
- Store the old area value in a temporary variable
- Calculate delta: `new_area - old_area`
- Update farmer: `total_acreage_ha = total_acreage_ha + delta`

This avoids full SUM() recalculation on every change.

**Current Implementation**: Full SUM() recalculation (simpler, works well for moderate datasets)
**Future Optimization**: Delta updates (for very large farms per farmer)

### 3. Trigger Efficiency
- Triggers fire only on relevant columns: `farmer_id`, `geometry`, `deleted_at`
- Uses `COALESCE()` to handle NULL cases
- Single UPDATE per affected farmer
- Handles farmer_id changes (moving farms between farmers)

## Migration Strategy

### Migration File: `006_add_farmer_total_acreage_rollup.sql`

1. Add column with default value
2. Create indexes
3. Backfill existing data
4. Create trigger function
5. Create triggers

## Benefits

1. **Query Performance**: No need to join and aggregate farms table
2. **Consistency**: Automatically maintained by database triggers
3. **Accuracy**: Single source of truth, updated atomically
4. **Scalability**: Indexed for fast filtering and sorting
5. **Reliability**: Database-level enforcement, no application logic needed

## Use Cases

### 1. Query farmers by total acreage
```sql
-- Farmers with more than 10 hectares
SELECT * FROM farmers
WHERE total_acreage_ha > 10.0
ORDER BY total_acreage_ha DESC;
```

### 2. FPO-level aggregations
```sql
-- Total acreage per FPO
SELECT aaa_org_id, SUM(total_acreage_ha) as fpo_total_acreage
FROM farmers
WHERE status = 'ACTIVE'
GROUP BY aaa_org_id;
```

### 3. Farmer ranking by land ownership
```sql
-- Top 10 farmers by acreage
SELECT first_name, last_name, total_acreage_ha
FROM farmers
ORDER BY total_acreage_ha DESC
LIMIT 10;
```

### 4. API Response Enhancement
```json
{
  "id": "FMRR00000001",
  "first_name": "John",
  "last_name": "Doe",
  "total_acreage_ha": 25.75,
  "farms": [
    {"id": "FARM00000001", "area_ha": 10.5},
    {"id": "FARM00000002", "area_ha": 15.25}
  ]
}
```

## Testing Strategy

### 1. Unit Tests
- Test trigger function logic
- Verify rollup calculations

### 2. Integration Tests
- Insert new farm → verify farmer total updates
- Update farm geometry → verify area recalculation
- Delete farm → verify farmer total decreases
- Move farm between farmers → verify both farmers update
- Soft delete farm → verify total excludes deleted farms

### 3. Performance Tests
- Measure query performance before/after rollup
- Test with large datasets (10K+ farms)

## Rollback Strategy

If issues arise, the migration can be rolled back:

```sql
-- Drop triggers
DROP TRIGGER IF EXISTS trigger_farm_insert_update_farmer_acreage ON farms;
DROP TRIGGER IF EXISTS trigger_farm_delete_update_farmer_acreage ON farms;

-- Drop function
DROP FUNCTION IF EXISTS update_farmer_total_acreage();

-- Drop column
ALTER TABLE farmers DROP COLUMN IF EXISTS total_acreage_ha;

-- Drop index
DROP INDEX IF EXISTS idx_farmers_total_acreage;
```

## Implementation Checklist

- [ ] Create migration file `006_add_farmer_total_acreage_rollup.sql`
- [ ] Add `TotalAcreageHa` field to `Farmer` entity
- [ ] Update farmer responses to include `total_acreage_ha`
- [ ] Add integration tests for trigger behavior
- [ ] Test rollback procedure
- [ ] Update API documentation
- [ ] Deploy to staging environment
- [ ] Verify performance improvements
- [ ] Deploy to production

## Related Files

- Entity: `internal/entities/farmer/farmer_normalized.go`
- Database: `internal/db/db.go` (setupPostMigration function)
- Migration: `internal/db/migrations/006_add_farmer_total_acreage_rollup.sql`
- Tests: `internal/services/farmer_service_test.go`
