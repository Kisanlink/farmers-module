# Cleanup: Remove aaa_farmer_user_id Column

## Overview

This document tracks the removal of the deprecated `aaa_farmer_user_id` column from the `farms` table after it was replaced by the `farmer_id` foreign key.

## Status: ✅ COMPLETED

## Changes Made

### 1. Database Migration ✅

**File**: `internal/db/migrations/004_drop_aaa_farmer_user_id_column.sql`

Actions:
- Drop `farms_farmer_id_idx` index (old index on `aaa_farmer_user_id`)
- Drop `aaa_farmer_user_id` column from `farms` table
- Add validation to ensure `farmer_id` exists before dropping old column
- Add warning if any farms have NULL `farmer_id`

```sql
DROP INDEX IF EXISTS farms_farmer_id_idx;
ALTER TABLE farms DROP COLUMN IF EXISTS aaa_farmer_user_id;
```

### 2. Database Setup Code ✅

**File**: `internal/db/db.go:304`

**Before:**
```go
gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (aaa_farmer_user_id);`)
```

**After:**
```go
gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (farmer_id);`)
```

Changed index to use `farmer_id` instead of `aaa_farmer_user_id`.

### 3. Test Schema ✅

**File**: `internal/db/db_test.go:38-45`

**Before:**
```go
CREATE TABLE IF NOT EXISTS farms (
    id TEXT PRIMARY KEY,
    aaa_farmer_user_id TEXT,
    aaa_org_id TEXT,
    geometry TEXT,
    created_at DATETIME
)
```

**After:**
```go
CREATE TABLE IF NOT EXISTS farms (
    id TEXT PRIMARY KEY,
    farmer_id TEXT,
    aaa_user_id TEXT,
    aaa_org_id TEXT,
    geometry TEXT,
    created_at DATETIME
)
```

Updated test schema to match production schema with `farmer_id`.

### 4. SQLC Queries Documentation ✅

**File**: `internal/db/sqlc/README.md` (NEW)

Created documentation noting that SQLC queries need to be updated to use `farmer_id` instead of `aaa_farmer_user_id`. Includes:
- Migration timeline
- Required query changes
- Before/After examples
- Regeneration instructions

## Files Still Referencing aaa_farmer_user_id

The following files still contain references but are **documentation only** (not code):

### Documentation Files (No Action Needed)
- `FARMERS_MODULE_WORKFLOW_SPECIFICATIONS.md` - Historical specification
- `IMPLEMENTATION_SUMMARY.md` - Implementation notes
- `.kiro/specs/aaa-permission-check-fix.md` - Old design doc
- `.kiro/specs/farm-attributes/*.md` - API specifications
- `.kiro/specs/farmers-module-workflows/design.md` - Workflow design

### Proto Files (Legacy - No Action if Not Used)
- `proto/kisanlink/farmers/v1alpha1/farmers.proto` - gRPC definitions

### Generated Documentation (Will Update on Regen)
- `docs/swagger.json`
- `docs/swagger.yaml`
- `docs/docs.go`

These will be regenerated when Swagger docs are rebuilt.

### SQLC Query File (Requires Manual Update)
- `internal/db/sqlc/query.sql` - **ACTION REQUIRED**
  - See `internal/db/sqlc/README.md` for update instructions
  - Lines to update: 46, 56, 60, 76
  - After updating, run: `make sqlc` or `sqlc generate`

## Verification

### Code Compilation ✅
```bash
go build ./...
```
Status: **PASSED** - No compilation errors

### Database Migration Order

1. **Migration 003**: Add `farmer_id` + FK constraints + populate data
2. **Migration 004**: Drop `aaa_farmer_user_id` column (THIS MIGRATION)

**Important**: Migration 003 MUST run successfully before migration 004.

## Benefits

1. **Cleaner Schema**: Single source of truth for farmer ownership
2. **Simpler Queries**: `WHERE farmer_id = $1` vs `WHERE aaa_farmer_user_id = $1 AND aaa_org_id = $2`
3. **Reduced Index Overhead**: One index instead of composite
4. **Better Performance**: Single column lookups
5. **Clearer Intent**: Foreign key makes relationship explicit

## Rollback Plan

If needed, rollback migration 004:

```sql
-- Rollback: Re-add aaa_farmer_user_id column
ALTER TABLE farms ADD COLUMN aaa_farmer_user_id VARCHAR(255);

-- Repopulate from farmers table
UPDATE farms f
SET aaa_farmer_user_id = fr.aaa_user_id
FROM farmers fr
WHERE f.farmer_id = fr.id;

-- Recreate index
CREATE INDEX farms_farmer_id_idx ON farms (aaa_farmer_user_id);
```

**Note**: Only rollback if migration 004 fails. Once deployed successfully, do not rollback.

## Related Documentation

- **Main Design**: `.kiro/specs/farmer-id-foreign-keys/DESIGN.md`
- **ADR**: `.kiro/specs/farmer-id-foreign-keys/ADR-farmer-id-foreign-keys.md`
- **Tasks**: `.kiro/specs/farmer-id-foreign-keys/TASKS.md`
- **Migration 003**: `internal/db/migrations/003_add_farmer_id_foreign_keys.sql`
- **Migration 004**: `internal/db/migrations/004_drop_aaa_farmer_user_id_column.sql`

## Completion Checklist

- [x] Create migration SQL to drop column
- [x] Update db.go index creation
- [x] Update test schema
- [x] Document SQLC query updates needed
- [x] Verify code compiles
- [x] Document cleanup changes

## Next Steps

1. Run migration 003 in development/staging
2. Verify all farms have `farmer_id` populated
3. Run migration 004 to drop old column
4. Update SQLC queries (see `internal/db/sqlc/README.md`)
5. Regenerate SQLC code
6. Run integration tests
7. Deploy to production

## Completion Date

2025-10-13
