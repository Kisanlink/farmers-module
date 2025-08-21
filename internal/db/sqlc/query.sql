-- name: CreateFPORef :one
INSERT INTO fpo_refs (aaa_org_id, business_config)
VALUES ($1, $2)
RETURNING *;

-- name: GetFPORef :one
SELECT * FROM fpo_refs
WHERE aaa_org_id = $1;

-- name: UpdateFPORef :one
UPDATE fpo_refs
SET business_config = $2, updated_at = now()
WHERE aaa_org_id = $1
RETURNING *;

-- name: DeleteFPORef :exec
DELETE FROM fpo_refs
WHERE aaa_org_id = $1;

-- name: CreateFarmerLink :one
INSERT INTO farmer_links (aaa_user_id, aaa_org_id, kisan_sathi_user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFarmerLink :one
SELECT * FROM farmer_links
WHERE aaa_user_id = $1 AND aaa_org_id = $2;

-- name: UpdateFarmerLink :one
UPDATE farmer_links
SET kisan_sathi_user_id = $3, status = $4, updated_at = now()
WHERE aaa_user_id = $1 AND aaa_org_id = $2
RETURNING *;

-- name: ListFarmerLinksByOrg :many
SELECT * FROM farmer_links
WHERE aaa_org_id = $1
ORDER BY created_at DESC;

-- name: ListFarmerLinksByKisanSathi :many
SELECT * FROM farmer_links
WHERE kisan_sathi_user_id = $1
ORDER BY created_at DESC;

-- name: CreateFarm :one
INSERT INTO farms (aaa_farmer_user_id, aaa_org_id, geom, metadata, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFarm :one
SELECT * FROM farms
WHERE id = $1;

-- name: GetFarmByFarmer :one
SELECT * FROM farms
WHERE id = $1 AND aaa_farmer_user_id = $2 AND aaa_org_id = $3;

-- name: ListFarmsByFarmer :many
SELECT * FROM farms
WHERE aaa_farmer_user_id = $1 AND aaa_org_id = $2
ORDER BY created_at DESC;

-- name: ListFarmsByOrg :many
SELECT * FROM farms
WHERE aaa_org_id = $1
ORDER BY created_at DESC;

-- name: UpdateFarm :one
UPDATE farms
SET geom = COALESCE($2, geom), metadata = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteFarm :exec
DELETE FROM farms
WHERE id = $1 AND aaa_farmer_user_id = $2 AND aaa_org_id = $3;

-- name: CreateCropCycle :one
INSERT INTO crop_cycles (farm_id, season, start_date, planned_crops)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCropCycle :one
SELECT * FROM crop_cycles
WHERE id = $1;

-- name: ListCropCyclesByFarm :many
SELECT * FROM crop_cycles
WHERE farm_id = $1
ORDER BY start_date DESC;

-- name: UpdateCropCycle :one
UPDATE crop_cycles
SET season = COALESCE($2, season),
    status = COALESCE($3, status),
    start_date = COALESCE($4, start_date),
    end_date = $5,
    planned_crops = COALESCE($6, planned_crops),
    outcome = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: EndCropCycle :one
UPDATE crop_cycles
SET status = 'COMPLETED', end_date = $2, outcome = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCropCycle :exec
DELETE FROM crop_cycles
WHERE id = $1;

-- name: CreateFarmActivity :one
INSERT INTO farm_activities (cycle_id, activity_type, planned_at, metadata, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFarmActivity :one
SELECT * FROM farm_activities
WHERE id = $1;

-- name: ListFarmActivitiesByCycle :many
SELECT * FROM farm_activities
WHERE cycle_id = $1
ORDER BY planned_at ASC, created_at DESC;

-- name: UpdateFarmActivity :one
UPDATE farm_activities
SET activity_type = COALESCE($2, activity_type),
    planned_at = $3,
    completed_at = $4,
    metadata = COALESCE($5, metadata),
    output = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CompleteFarmActivity :one
UPDATE farm_activities
SET completed_at = $2, output = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteFarmActivity :exec
DELETE FROM farm_activities
WHERE id = $1;

-- name: GetFarmWithDetails :one
SELECT
    f.*,
    cc.id as cycle_id,
    cc.season,
    cc.status as cycle_status,
    cc.start_date,
    cc.end_date,
    cc.planned_crops,
    cc.outcome
FROM farms f
LEFT JOIN crop_cycles cc ON f.id = cc.farm_id
WHERE f.id = $1
LIMIT 1;
