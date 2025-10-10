# Database Migration Strategy

## Overview

This document describes the database migration strategy for the Farmers Module, including entity relationships, migration ordering, and PostGIS handling.

## Migration Execution Paths

The system supports three migration paths based on PostGIS availability:

### Path 1: PostGIS Available (Full Feature Set)
- **Condition**: PostGIS extension successfully created
- **Features**: All entities including spatial farm boundaries
- **Tables Created**: All tables in the system

### Path 2: PostGIS Not Available (Fallback)
- **Condition**: PostGIS extension check fails
- **Features**: All entities except farms
- **Tables Skipped**: `farms`, `crop_cycles`, `farm_activities`, junction tables

### Path 3: PostGIS Extension Creation Failed
- **Condition**: `CREATE EXTENSION postgis` fails
- **Features**: Same as Path 2
- **Tables Skipped**: Same as Path 2

## Migration Order (Dependency Graph)

Tables are migrated in strict dependency order to avoid foreign key constraint violations:

```
1. Independent Master Data Tables (no FK dependencies)
   - fpo_refs
   - soil_types
   - irrigation_sources
   - crops

2. Address Table (no dependencies)
   - addresses

3. Farmer Tables
   - farmers (depends on addresses via FK)
   - farmer_links (no dependencies)

4. Farm Table (PostGIS required)
   - farms (depends on farmers, uses PostGIS geometry)

5. Crop Variety Table
   - crop_varieties (depends on crops)

6. Crop Cycle Table
   - crop_cycles (depends on farms, farmers, crops, crop_varieties)

7. Farm Activity Table
   - farm_activities (depends on crop_cycles)

8. Junction Tables
   - farm_soil_types (depends on farms, soil_types)
   - farm_irrigation_sources (depends on farms, irrigation_sources)

9. Bulk Operations (last)
   - bulk_operations
   - processing_details
```

## Entity Status

### ✅ Active Entities (Included in Migration)

| Entity | Package | Table | PostGIS Required |
|--------|---------|-------|------------------|
| Address | `farmer` | `addresses` | No |
| Farmer | `farmer` | `farmers` | No |
| FarmerLink | `farmer` | `farmer_links` | No |
| FPORef | `fpo` | `fpo_refs` | No |
| Farm | `farm` | `farms` | **Yes** |
| Crop | `crop` | `crops` | No |
| CropVariety | `crop_variety` | `crop_varieties` | No |
| CropCycle | `crop_cycle` | `crop_cycles` | **Yes** (depends on Farm) |
| FarmActivity | `farm_activity` | `farm_activities` | **Yes** (depends on CropCycle) |
| SoilType | `soil_type` | `soil_types` | No |
| IrrigationSource | `irrigation_source` | `irrigation_sources` | No |
| FarmSoilType | `farm_soil_type` | `farm_soil_types` | **Yes** (depends on Farm) |
| FarmIrrigationSource | `farm_irrigation_source` | `farm_irrigation_sources` | **Yes** (depends on Farm) |
| BulkOperation | `bulk` | `bulk_operations` | No |
| ProcessingDetail | `bulk` | `processing_details` | No |

### ❌ Deprecated Entities (NOT Included in Migration)

| Entity | Package | Reason | Use Instead |
|--------|---------|--------|-------------|
| FarmerLegacy | `farmer` | Denormalized model with embedded address | `farmer.Farmer` |
| FarmerProfile | `entities` | Old structure using JSONB custom type | `farmer.Farmer` |
| Address (entities pkg) | `entities` | Duplicate definition | `farmer.Address` |

## PostGIS Support

### Farm Entity (Requires PostGIS)

The Farm entity uses PostGIS geometry types for spatial features:

```go
Geometry string `gorm:"type:geometry(POLYGON,4326)"`
```

PostGIS features include:
- Spatial indexing (GIST indexes)
- Area computation (computed column)
- SRID validation (4326 = WGS84)
- Geometry validity checks

### Address Entity (No PostGIS)

The Address entity stores coordinates as text (not PostGIS geometry):

```go
Coordinates string `gorm:"type:text"` // Stored as "lat,lng"
```

**Rationale**:
- Address coordinates don't require spatial operations
- Simpler for point location lookups
- Compatible with SQLite for testing
- Reduces PostGIS dependency surface

## GORM Type Specifications

### Explicit Type Tags

All entities use explicit GORM type tags to avoid type inference issues:

```go
// Correct - explicit type
Status string `gorm:"type:farmer_status;not null;default:'ACTIVE'"`

// Avoid - GORM infers type
Status string `gorm:"not null;default:'ACTIVE'"`
```

### JSONB Fields

JSONB fields use explicit type and default value:

```go
Preferences map[string]string `gorm:"type:jsonb;default:'{}'::jsonb"`
Metadata    map[string]string `gorm:"type:jsonb;default:'{}'::jsonb"`
```

### Custom ENUM Types

Custom PostgreSQL ENUM types are created first (before table creation):

- `farmer_status`: ACTIVE, INACTIVE, SUSPENDED
- `link_status`: ACTIVE, INACTIVE
- `season`: RABI, KHARIF, ZAID, OTHER
- `crop_category`: CEREALS, PULSES, VEGETABLES, etc.
- `cycle_status`: PLANNED, ACTIVE, COMPLETED, CANCELLED
- `activity_status`: PLANNED, COMPLETED, CANCELLED

## Foreign Key Relationships

### Farmer → Address (One-to-One, Optional)

```go
// Farmer entity
AddressID *string  `gorm:"type:varchar(255)"`
Address   *Address `gorm:"foreignKey:AddressID;constraint:OnDelete:SET NULL"`
```

**Behavior**: When address is deleted, farmer's `address_id` is set to NULL.

### CropCycle → Farm, Farmer (Many-to-One, Required)

```go
// CropCycle entity
FarmID   string `gorm:"type:varchar(255);not null"`
FarmerID string `gorm:"type:varchar(255);not null"`
```

**Behavior**: Crop cycles require both farm and farmer to exist.

### FarmActivity → CropCycle (Many-to-One, Required)

```go
// FarmActivity entity
CropCycleID string `gorm:"type:varchar(255);not null"`
```

**Behavior**: Activities are tied to specific crop cycles.

## Index Strategy

### Unique Indexes

```sql
-- Farmer uniqueness per organization
CREATE UNIQUE INDEX farmers_aaa_user_org_idx ON farmers (aaa_user_id, aaa_org_id);

-- FPO uniqueness
CREATE UNIQUE INDEX fpo_refs_aaa_org_id_idx ON fpo_refs (aaa_org_id);

-- Crop name uniqueness
CREATE UNIQUE INDEX crops_name_idx ON crops (name);
```

### Performance Indexes

```sql
-- Farmer lookups
CREATE INDEX farmers_phone_idx ON farmers (phone_number);
CREATE INDEX farmers_email_idx ON farmers (email);

-- Farm lookups
CREATE INDEX farms_farmer_id_idx ON farms (aaa_farmer_user_id);
CREATE INDEX farms_fpo_id_idx ON farms (aaa_org_id);

-- Crop cycle queries
CREATE INDEX crop_cycles_farm_id_idx ON crop_cycles (farm_id);
CREATE INDEX crop_cycles_season_idx ON crop_cycles (season);
CREATE INDEX crop_cycles_status_idx ON crop_cycles (status);
```

### Spatial Indexes (PostGIS Only)

```sql
CREATE INDEX farms_geometry_gist ON farms USING GIST (geometry::geometry);
```

## Testing Strategy

### Unit Tests

- SQLite in-memory database for non-PostGIS entities
- Skip entities requiring PostGIS (Farm, CropCycle, FarmActivity)
- Test validation logic and entity structure

### Integration Tests

- TestContainers with PostgreSQL + PostGIS
- Full migration with all entities
- Test spatial operations and relationships

## Migration Rollback

To rollback migration:

```sql
-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS processing_details CASCADE;
DROP TABLE IF EXISTS bulk_operations CASCADE;
DROP TABLE IF EXISTS farm_irrigation_sources CASCADE;
DROP TABLE IF EXISTS farm_soil_types CASCADE;
DROP TABLE IF EXISTS farm_activities CASCADE;
DROP TABLE IF EXISTS crop_cycles CASCADE;
DROP TABLE IF EXISTS crop_varieties CASCADE;
DROP TABLE IF EXISTS farms CASCADE;
DROP TABLE IF EXISTS farmer_links CASCADE;
DROP TABLE IF EXISTS farmers CASCADE;
DROP TABLE IF EXISTS addresses CASCADE;
DROP TABLE IF EXISTS crops CASCADE;
DROP TABLE IF EXISTS irrigation_sources CASCADE;
DROP TABLE IF EXISTS soil_types CASCADE;
DROP TABLE IF EXISTS fpo_refs CASCADE;

-- Drop custom ENUM types
DROP TYPE IF EXISTS farmer_status CASCADE;
DROP TYPE IF EXISTS link_status CASCADE;
DROP TYPE IF EXISTS season CASCADE;
DROP TYPE IF EXISTS crop_category CASCADE;
DROP TYPE IF EXISTS cycle_status CASCADE;
DROP TYPE IF EXISTS activity_status CASCADE;
```

## Best Practices

1. **Always check PostGIS availability** before migrating farm-related entities
2. **Migrate in dependency order** to avoid FK constraint violations
3. **Use explicit GORM type tags** for all fields
4. **Create ENUMs first** before table creation
5. **Add indexes after table creation** via `setupPostMigration()`
6. **Test migrations** with both PostGIS enabled and disabled
7. **Document breaking changes** when updating entity structure

## Troubleshooting

### Issue: "invalid geometry" error on addresses table

**Cause**: Trying to use PostGIS geometry type on addresses
**Solution**: addresses.coordinates uses `type:text`, not PostGIS geometry

### Issue: Foreign key constraint violation during migration

**Cause**: Tables created out of dependency order
**Solution**: Check migration order matches dependency graph above

### Issue: "near 'Point': syntax error" in tests

**Cause**: SQLite doesn't support PostGIS types
**Solution**: Skip PostGIS-dependent entities in SQLite tests

## Version History

- **v1.0.0** (2025-01-10): Initial migration strategy with normalized Farmer entity
- **v0.9.0** (2024-12-XX): Legacy migration with FarmerProfile (deprecated)
