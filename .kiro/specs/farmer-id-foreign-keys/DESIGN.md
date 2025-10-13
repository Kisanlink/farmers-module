# Farmer ID Foreign Key Relationships

## Overview

This specification documents the addition of `farmer_id` foreign key relationships to all user-related tables in the farmers module. This ensures proper referential integrity and establishes clear ownership relationships between farmers and their data.

## Problem Statement

Previously, several tables in the farmers module used `aaa_user_id` and `aaa_org_id` to identify user ownership but lacked direct foreign key relationships to the `farmers` table. This created the following issues:

1. **No referential integrity**: Database couldn't enforce data consistency
2. **Complex queries**: Required multiple joins through AAA fields instead of direct relationships
3. **Inconsistent data model**: Some tables had proper FK relationships while others didn't
4. **Cascading deletes**: Couldn't automatically clean up related data when a farmer is deleted

## Solution

Add `farmer_id` foreign key column with proper GORM relationship tags to all user-related tables:

### Tables Modified

#### 1. `farms` Table
- **Added**: `farmer_id VARCHAR(255) NOT NULL`
- **Foreign Key**: References `farmers.id` with `ON DELETE CASCADE`
- **Index**: Added for query performance
- **Relationship**: Added `Farmer` relationship field in GORM model (currently commented out to avoid circular import)

**Rationale**: Every farm must belong to a specific farmer. Direct FK relationship enables:
- Efficient farmer → farms queries
- Automatic cleanup when farmer is deleted
- Data integrity enforcement

#### 2. `crop_cycles` Table
- **Modified**: Changed `farmer_id` from `UUID` to `VARCHAR(255) NOT NULL`
- **Foreign Key**: Added constraint to `farmers.id` with `ON DELETE CASCADE`
- **Relationship**: Added `Farmer` relationship field with proper GORM tags
- **Index**: Already existed, retained

**Rationale**: Crop cycles already had `farmer_id` but lacked:
- Foreign key constraint for referential integrity
- Proper GORM relationship definition
- Consistent data type with other IDs

#### 3. `farm_activities` Table
- **Added**: `farmer_id VARCHAR(255) NOT NULL`
- **Foreign Key**: References `farmers.id` with `ON DELETE CASCADE`
- **Index**: Added for query performance
- **Relationship**: Added `Farmer` relationship field with proper GORM tags

**Rationale**: Activities belong to farmers through crop cycles, but direct FK:
- Enables efficient farmer → activities queries without joining through crop_cycle
- Maintains data consistency
- Supports audit and reporting requirements

### Tables Not Modified

The following tables were analyzed but determined not to require `farmer_id`:

#### Junction Tables
- **`farm_soil_types`**: Gets farmer through `farm_id` → `farms` → `farmer_id`
- **`farm_irrigation_sources`**: Gets farmer through `farm_id` → `farms` → `farmer_id`

**Rationale**: These are junction tables linking farms to master data. Adding `farmer_id` would be redundant and violate normalization principles.

#### Organization-Level Tables
- **`bulk_operations`**: FPO-level operations, identified by `fpo_org_id`
- **`addresses`**: Reusable address records, can be shared

**Rationale**: These tables operate at organization level or are shared resources, not farmer-specific.

#### Link Tables
- **`farmer_links`**: Uses `aaa_user_id` + `aaa_org_id` composite, doesn't need farmer_id as it's used to establish the farmer relationship

## Implementation Details

### Model Changes

#### Farm Entity (`internal/entities/farm/farm.go`)
```go
type Farm struct {
    base.BaseModel
    FarmerID  string `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
    // ... other fields

    // Relationships (circular import issue - to be resolved)
    // Farmer *farmer.Farmer `json:"farmer,omitempty" gorm:"foreignKey:FarmerID;references:ID"`
}
```

#### CropCycle Entity (`internal/entities/crop_cycle/crop_cycle.go`)
```go
type CropCycle struct {
    base.BaseModel
    FarmerID string `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
    // ... other fields

    // Relationships
    Farmer *farmer.Farmer `json:"farmer,omitempty" gorm:"foreignKey:FarmerID;references:ID;constraint:OnDelete:CASCADE"`
}
```

#### FarmActivity Entity (`internal/entities/farm_activity/farm_activity.go`)
```go
type FarmActivity struct {
    base.BaseModel
    FarmerID string `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
    // ... other fields

    // Relationships
    Farmer *farmer.Farmer `json:"farmer,omitempty" gorm:"foreignKey:FarmerID;references:ID;constraint:OnDelete:CASCADE"`
}
```

### Migration File

Created: `internal/db/migrations/003_add_farmer_id_foreign_keys.sql`

Key operations:
1. Add `farmer_id` column to `farms` and `farm_activities`
2. Alter `crop_cycles.farmer_id` type from UUID to VARCHAR(255)
3. Add NOT NULL constraints
4. Create indexes for performance
5. Add foreign key constraints with CASCADE delete
6. Backfill existing data with farmer IDs
7. Add documentation comments

### Validation Updates

Updated validation methods to require `farmer_id`:

```go
func (f *Farm) Validate() error {
    if f.FarmerID == "" {
        return common.ErrInvalidFarmData
    }
    // ... other validations
}
```

## Data Migration Strategy

The migration includes UPDATE statements to populate `farmer_id` in existing records:

1. **farms**: Map using `aaa_user_id` + `aaa_org_id` → `farmers.id`
2. **crop_cycles**: Already populated, but adds FK constraint
3. **farm_activities**: Derive from parent `crop_cycles.farmer_id`

**Important**: Review and test data migration SQL before running in production.

## Benefits

1. **Data Integrity**: Foreign key constraints prevent orphaned records
2. **Query Performance**: Direct relationships eliminate complex joins
3. **Cascading Deletes**: Automatic cleanup of related data
4. **Cleaner Code**: Proper GORM relationships simplify repository queries
5. **Audit Trail**: Direct farmer ownership is explicit in schema
6. **Consistency**: All user tables follow same pattern

## Considerations

### Circular Import Issue (Farm → Farmer)

The `Farm` entity currently cannot include the `Farmer` relationship field due to circular import:
- `farm` package needs to import `farmer` package
- But this may create issues if there are cross-dependencies

**Resolution Options**:
1. Keep relationship one-directional (CropCycle and FarmActivity → Farmer only)
2. Create interface/abstraction layer
3. Restructure package organization (future refactor)

### Backward Compatibility

- Existing queries using `aaa_user_id` still work
- Both `aaa_user_id` and `farmer_id` are maintained
- Services need gradual migration to use `farmer_id` for queries

### Performance Impact

- Added indexes on `farmer_id` columns minimize performance impact
- Foreign key constraints add minimal overhead for writes
- Query performance improves with direct relationships

## Testing Requirements

1. **Unit Tests**: Validate entity validation with `farmer_id`
2. **Integration Tests**: Test cascade delete behavior
3. **Migration Tests**: Verify data migration correctness
4. **Repository Tests**: Update queries to use `farmer_id`
5. **API Tests**: Ensure endpoints properly populate `farmer_id`

## Rollout Plan

1. ✅ Update entity models with `farmer_id` fields
2. ✅ Create migration SQL file
3. ✅ Verify code compiles
4. 🔄 Update service layer to populate `farmer_id`
5. 🔄 Update repository queries to use `farmer_id`
6. 🔄 Run migration in development environment
7. 🔄 Test cascade delete behavior
8. 🔄 Update API handlers/tests
9. 🔄 Code review and QA
10. 🔄 Production deployment

## Related Documentation

- Database schema: `internal/db/db.go`
- Migration service: `internal/db/migration_service.go`
- Entity models: `internal/entities/*/`

## Decision Records

See: `ADR-farmer-id-foreign-keys.md` for architectural decision rationale.
