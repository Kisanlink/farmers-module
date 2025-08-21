package farm

import (
	"context"

	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// FarmRepository defines the interface for farm data operations
type FarmRepository interface {
	// Farm-specific operations
	GetByFarmerAndOrg(ctx context.Context, farmerUserID, orgID string) ([]farmEntity.Farm, error)
	GetByOrg(ctx context.Context, orgID string) ([]farmEntity.Farm, error)
	GetByFarmer(ctx context.Context, farmerUserID string) ([]farmEntity.Farm, error)

	// Spatial operations
	GetFarmsInArea(ctx context.Context, bounds farmEntity.Geometry) ([]farmEntity.Farm, error)
	GetFarmsByAreaRange(ctx context.Context, minArea, maxArea float64) ([]farmEntity.Farm, error)
}

// farmRepository implements FarmRepository
type farmRepository struct {
	postgresManager *db.PostgresManager
}

// NewFarmRepository creates a new farm repository
func NewFarmRepository(dbManager db.DBManager) FarmRepository {
	return &farmRepository{
		postgresManager: dbManager.(*db.PostgresManager),
	}
}

// GetByFarmerAndOrg retrieves farms by farmer user ID and organization ID
func (r *farmRepository) GetByFarmerAndOrg(ctx context.Context, farmerUserID, orgID string) ([]farmEntity.Farm, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_farmer_user_id", Operator: base.OpEqual, Value: farmerUserID},
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
	}

	var farms []farmEntity.Farm
	err := r.postgresManager.List(ctx, filter, &farms)
	return farms, err
}

// GetByOrg retrieves all farms in an organization
func (r *farmRepository) GetByOrg(ctx context.Context, orgID string) ([]farmEntity.Farm, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
	}

	var farms []farmEntity.Farm
	err := r.postgresManager.List(ctx, filter, &farms)
	return farms, err
}

// GetByFarmer retrieves all farms owned by a farmer
func (r *farmRepository) GetByFarmer(ctx context.Context, farmerUserID string) ([]farmEntity.Farm, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_farmer_user_id", Operator: base.OpEqual, Value: farmerUserID},
			},
			Logic: base.LogicAnd,
		},
	}

	var farms []farmEntity.Farm
	err := r.postgresManager.List(ctx, filter, &farms)
	return farms, err
}

// GetFarmsInArea retrieves farms within a geographic area
func (r *farmRepository) GetFarmsInArea(ctx context.Context, bounds farmEntity.Geometry) ([]farmEntity.Farm, error) {
	// This would use PostGIS spatial queries
	// For now, return empty slice - implement with proper spatial queries
	return []farmEntity.Farm{}, nil
}

// GetFarmsByAreaRange retrieves farms within an area range
func (r *farmRepository) GetFarmsByAreaRange(ctx context.Context, minArea, maxArea float64) ([]farmEntity.Farm, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "area_ha", Operator: base.OpGreaterEqual, Value: minArea},
				{Field: "area_ha", Operator: base.OpLessEqual, Value: maxArea},
			},
			Logic: base.LogicAnd,
		},
	}

	var farms []farmEntity.Farm
	err := r.postgresManager.List(ctx, filter, &farms)
	return farms, err
}
