# SQLC Queries - Update Required

## ⚠️ IMPORTANT: Schema Changes

The `aaa_farmer_user_id` column has been **removed** from the `farms` table and **replaced** with `farmer_id` as a foreign key to the `farmers` table.

### Migration Timeline

1. **Migration 003**: Added `farmer_id` column with foreign key constraints
2. **Migration 004**: Dropped `aaa_farmer_user_id` column

### Required Actions

The SQLC queries in `query.sql` currently reference `aaa_farmer_user_id` and need to be updated:

#### Files Affected

- `query.sql` (lines 46, 56, 60, 76)

#### Changes Required

Replace all occurrences of `aaa_farmer_user_id` with `farmer_id`:

**Before:**
```sql
-- name: CreateFarm :one
INSERT INTO farms (aaa_farmer_user_id, aaa_org_id, geom, metadata, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFarmByFarmer :one
SELECT * FROM farms
WHERE id = $1 AND aaa_farmer_user_id = $2 AND aaa_org_id = $3;

-- name: ListFarmsByFarmer :many
SELECT * FROM farms
WHERE aaa_farmer_user_id = $1 AND aaa_org_id = $2
ORDER BY created_at DESC;

-- name: DeleteFarm :exec
DELETE FROM farms
WHERE id = $1 AND aaa_farmer_user_id = $2 AND aaa_org_id = $3;
```

**After:**
```sql
-- name: CreateFarm :one
INSERT INTO farms (farmer_id, aaa_user_id, aaa_org_id, geom, metadata, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFarmByFarmer :one
SELECT * FROM farms
WHERE id = $1 AND farmer_id = $2;

-- name: ListFarmsByFarmer :many
SELECT * FROM farms
WHERE farmer_id = $1
ORDER BY created_at DESC;

-- name: DeleteFarm :exec
DELETE FROM farms
WHERE id = $1 AND farmer_id = $2;
```

### Regenerate SQLC Code

After updating `query.sql`, regenerate the Go code:

```bash
# If using sqlc
make sqlc

# Or directly
sqlc generate
```

### Benefits of the Change

1. **Referential Integrity**: Foreign key constraint ensures valid farmer references
2. **Simpler Queries**: Single `farmer_id` instead of `aaa_farmer_user_id + aaa_org_id`
3. **Performance**: Single indexed column lookup
4. **Cleaner API**: One parameter instead of two for farmer-scoped operations

### Migration Compatibility

- **Backward Compatibility**: None - this is a breaking schema change
- **Data Migration**: Handled by migration 003 (populates `farmer_id` from existing data)
- **Cleanup**: Migration 004 removes the old column

### Related Documentation

- Design: `.kiro/specs/farmer-id-foreign-keys/DESIGN.md`
- ADR: `.kiro/specs/farmer-id-foreign-keys/ADR-farmer-id-foreign-keys.md`
- Migrations: `internal/db/migrations/003_*.sql` and `004_*.sql`
