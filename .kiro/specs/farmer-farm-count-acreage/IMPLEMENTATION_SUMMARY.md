# Farmer Farm Count & Acreage Auto-Update Implementation

## Overview

This implementation adds automatic tracking of a farmer's total number of farms (`farm_count`) and total acreage (`total_acreage_ha`) using PostgreSQL database triggers. These rollup fields are automatically maintained whenever farms are created, updated, or deleted.

## Changes Made

### 1. Entity Updates

**File:** `internal/entities/farmer/farmer_normalized.go`

Added `farm_count` field to the `Farmer` entity:

```go
// Rollup Fields - Maintained by database triggers
TotalAcreageHa float64 `json:"total_acreage_ha" gorm:"type:numeric(12,4);default:0.0;not null"`
FarmCount      int     `json:"farm_count" gorm:"type:integer;default:0;not null"`
```

### 2. GORM Hook Implementation

**File:** `internal/entities/farm/farm.go`

Implemented GORM hooks to automatically maintain farmer stats:

1. **AfterCreate**: Updates farmer stats when a new farm is created
2. **AfterUpdate**: Updates farmer stats when a farm is updated (including farmer_id changes)
3. **AfterDelete**: Updates farmer stats when a farm is deleted

#### GORM Hook Logic

The `updateFarmerStats()` function runs after farm operations:

- **AfterCreate**: Increments the farmer's count and adds the acreage
- **AfterUpdate**:
  - Updates the current farmer's stats
  - If `farmer_id` changed, also updates the previous farmer's stats
- **AfterDelete**: Decrements the farmer's count and subtracts the acreage

```go
func (f *Farm) AfterCreate(tx interface{}) error {
    return updateFarmerStats(tx, f.FarmerID)
}

func updateFarmerStats(tx interface{}, farmerID string) error {
    db, ok := tx.(*gorm.DB)
    if !ok {
        return nil // Silently fail if not a gorm.DB
    }

    // Calculate stats using GORM
    var stats struct {
        TotalAcreage float64
        FarmCount    int64
    }

    err := db.Model(&Farm{}).
        Select("COALESCE(SUM(area_ha_computed), 0) as total_acreage, COUNT(*) as farm_count").
        Where("farmer_id = ? AND deleted_at IS NULL", farmerID).
        Scan(&stats).Error

    if err != nil {
        return err
    }

    // Update farmer using GORM
    return db.Model(&farmer.Farmer{}).
        Where("id = ?", farmerID).
        Updates(map[string]interface{}{
            "total_acreage_ha": stats.TotalAcreage,
            "farm_count":       stats.FarmCount,
        }).Error
}
```

**File:** `internal/db/db.go`

Replaced SQL trigger setup with GORM-based backfill:

- Removed `setupFarmerAcreageRollupTriggers()` function and all SQL trigger code
- Added `backfillFarmerStats()` function that uses GORM to initialize existing farmer stats
- Uses pure GORM queries for all database operations (no raw SQL)

### 3. Database Indexes

Added index for fast filtering/sorting by `farm_count`:

```sql
CREATE INDEX IF NOT EXISTS idx_farmers_farm_count ON farmers (farm_count);
```

### 4. Migration Files

Created SQL migration files:

- **`migrations/004_add_farmer_farm_count.sql`**: Adds the column, index, and backfills data
- **`migrations/004_add_farmer_farm_count_rollback.sql`**: Reverts all changes

**Note:** The migration SQL files contain SQL trigger logic for backward compatibility, but the application now uses GORM hooks instead of SQL triggers.

## How It Works

### Automatic Updates

Updates are handled by GORM hooks in `internal/entities/farm/farm.go`:

1. **Farm Created:**
   ```
   POST /api/v1/farms
   {
     "farmer_id": "FMRR...",
     "geometry": "...",
     "area_ha": 2.5
   }
   ```
   → `AfterCreate` hook fires → `farmers.farm_count` increments by 1
   → `farmers.total_acreage_ha` increases by 2.5 ha

2. **Farm Updated (area changed):**
   ```
   PUT /api/v1/farms/{id}
   {
     "area_ha": 3.0  // was 2.5
   }
   ```
   → `AfterUpdate` hook fires → `farmers.total_acreage_ha` increases by 0.5 ha
   → `farmers.farm_count` stays the same

3. **Farm Transferred (farmer_id changed):**
   ```
   PUT /api/v1/farms/{id}
   {
     "farmer_id": "FMRR-new-farmer"  // changed
   }
   ```
   → `AfterUpdate` hook fires → **Old farmer**: count -1, acreage -3.0 ha
   → **New farmer**: count +1, acreage +3.0 ha

4. **Farm Deleted (soft delete):**
   ```
   DELETE /api/v1/farms/{id}
   ```
   → `AfterDelete` hook fires → `farmers.farm_count` decrements by 1
   → `farmers.total_acreage_ha` decreases by 3.0 ha

### Query Examples

```sql
-- Get farmers with the most farms
SELECT id, first_name, last_name, farm_count, total_acreage_ha
FROM farmers
WHERE deleted_at IS NULL
ORDER BY farm_count DESC
LIMIT 10;

-- Find farmers with large holdings
SELECT id, first_name, last_name, farm_count, total_acreage_ha
FROM farmers
WHERE total_acreage_ha > 10.0
AND deleted_at IS NULL;

-- Get average farm size per farmer
SELECT
    id,
    first_name,
    last_name,
    farm_count,
    total_acreage_ha,
    CASE
        WHEN farm_count > 0 THEN ROUND((total_acreage_ha / farm_count)::numeric, 2)
        ELSE 0
    END as avg_farm_size_ha
FROM farmers
WHERE deleted_at IS NULL
AND farm_count > 0;
```

## Benefits

1. **Performance**: No need to JOIN and COUNT farms table every time
2. **Consistency**: GORM hooks ensure data is always accurate when using the application
3. **Real-time**: Updates happen automatically, no batch jobs needed
4. **Simple API**: Frontend gets `farm_count` and `total_acreage_ha` directly in farmer response
5. **Maintainable**: Pure Go/GORM implementation, no raw SQL to maintain
6. **Testable**: Easier to test with GORM hooks than database triggers
7. **Portable**: Works across different database engines supported by GORM

## API Response Example

```json
GET /api/v1/identity/farmers/id/FMRR...

{
  "success": true,
  "data": {
    "id": "FMRR...",
    "aaa_user_id": "USER...",
    "aaa_org_id": "ORGN...",
    "first_name": "Rajesh",
    "last_name": "Kumar",
    "total_acreage_ha": 5.75,
    "farm_count": 3,
    "status": "ACTIVE",
    ...
  }
}
```

## Testing

### Manual Testing Steps

1. **Create a farmer**
   ```bash
   POST /api/v1/identity/farmers
   ```
   Verify: `farm_count = 0`, `total_acreage_ha = 0.0`

2. **Add a farm**
   ```bash
   POST /api/v1/farms
   {
     "farmer_id": "<farmer_id>",
     "area_ha": 2.5
   }
   ```
   Verify: `farm_count = 1`, `total_acreage_ha = 2.5`

3. **Add another farm**
   ```bash
   POST /api/v1/farms
   {
     "farmer_id": "<farmer_id>",
     "area_ha": 1.25
   }
   ```
   Verify: `farm_count = 2`, `total_acreage_ha = 3.75`

4. **Delete a farm**
   ```bash
   DELETE /api/v1/farms/{id}
   ```
   Verify: `farm_count = 1`, `total_acreage_ha = 2.5`

### Database Verification

```sql
-- Check that all farmers have correct counts
SELECT
    f.id,
    f.farm_count,
    (SELECT COUNT(*) FROM farms WHERE farmer_id = f.id AND deleted_at IS NULL) as actual_count
FROM farmers f
WHERE f.farm_count != (SELECT COUNT(*) FROM farms WHERE farmer_id = f.id AND deleted_at IS NULL);
-- Expected: 0 rows
```

## Migration Instructions

### Apply Migration

When you run the application, the migration will be applied automatically via `SetupDatabase()`.

Alternatively, run manually:
```bash
psql -d farmers_db -f migrations/004_add_farmer_farm_count.sql
```

### Rollback Migration

```bash
psql -d farmers_db -f migrations/004_add_farmer_farm_count_rollback.sql
```

⚠️ **Note**: After rollback, you must also revert the code changes to the `Farmer` entity.

## Future Enhancements

1. **Additional Rollup Fields**:
   - `total_crop_cycles` - Total number of crop cycles across all farms
   - `active_crop_cycles` - Current active crop cycles
   - `total_production_kg` - Historical total production

2. **Historical Tracking**:
   - Store snapshots of `farm_count` and `total_acreage_ha` over time
   - Track growth trends

3. **Alerts**:
   - Notify when a farmer's land holdings exceed certain thresholds
   - Alert on sudden drops in acreage (potential data quality issues)

## Related Files

- **Entity**: `internal/entities/farmer/farmer_normalized.go`
- **Database Setup**: `internal/db/db.go`
- **Migration**: `migrations/004_add_farmer_farm_count.sql`
- **Rollback**: `migrations/004_add_farmer_farm_count_rollback.sql`
- **Farm Entity**: `internal/entities/farm/farm.go`

## Notes

- Both `total_acreage_ha` and `farm_count` are maintained by GORM hooks in `farm.go`
- GORM hooks run automatically on Create/Update/Delete operations through the application
- The `area_ha_computed` column in farms is computed from the PostGIS geometry
- Soft-deleted farms (where `deleted_at IS NOT NULL`) are excluded from counts
- **Important**: Direct SQL updates to the farms table will NOT trigger the GORM hooks. Always use the application's GORM models for farm operations to maintain consistency.
- For data consistency after direct SQL operations, run the backfill function or restart the application
