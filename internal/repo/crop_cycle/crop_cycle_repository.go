package crop_cycle

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CropCycleRepository defines the interface for crop cycle data operations
type CropCycleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, cycle *entities.CropCycle) error
	GetByID(ctx context.Context, id string) (*entities.CropCycle, error)
	Update(ctx context.Context, cycle *entities.CropCycle) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *base.Filter) ([]entities.CropCycle, error)

	// Crop cycle-specific operations
	GetByFarm(ctx context.Context, farmID string) ([]entities.CropCycle, error)
	GetBySeason(ctx context.Context, season string) ([]entities.CropCycle, error)
	GetByStatus(ctx context.Context, status string) ([]entities.CropCycle, error)
	GetActiveCycles(ctx context.Context) ([]entities.CropCycle, error)
}

// cropCycleRepository implements CropCycleRepository
type cropCycleRepository struct {
	dbManager db.DBManager
}

// NewCropCycleRepository creates a new crop cycle repository
func NewCropCycleRepository(dbManager db.DBManager) CropCycleRepository {
	return &cropCycleRepository{
		dbManager: dbManager,
	}
}

// Basic CRUD operations
func (r *cropCycleRepository) Create(ctx context.Context, cycle *entities.CropCycle) error {
	return r.dbManager.Create(ctx, cycle)
}

func (r *cropCycleRepository) GetByID(ctx context.Context, id string) (*entities.CropCycle, error) {
	var cycle entities.CropCycle
	err := r.dbManager.GetByID(ctx, id, &cycle)
	if err != nil {
		return nil, err
	}
	return &cycle, nil
}

func (r *cropCycleRepository) Update(ctx context.Context, cycle *entities.CropCycle) error {
	return r.dbManager.Update(ctx, cycle)
}

func (r *cropCycleRepository) Delete(ctx context.Context, id string) error {
	return r.dbManager.Delete(ctx, id, &entities.CropCycle{})
}

func (r *cropCycleRepository) List(ctx context.Context, filter *base.Filter) ([]entities.CropCycle, error) {
	var cycles []entities.CropCycle
	err := r.dbManager.List(ctx, filter, &cycles)
	return cycles, err
}

// GetByFarm retrieves all crop cycles for a specific farm
func (r *cropCycleRepository) GetByFarm(ctx context.Context, farmID string) ([]entities.CropCycle, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "farm_id", Operator: base.OpEqual, Value: farmID},
			},
			Logic: base.LogicAnd,
		},
	}

	return r.List(ctx, filter)
}

// GetBySeason retrieves all crop cycles for a specific season
func (r *cropCycleRepository) GetBySeason(ctx context.Context, season string) ([]entities.CropCycle, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "season", Operator: base.OpEqual, Value: season},
			},
			Logic: base.LogicAnd,
		},
	}

	return r.List(ctx, filter)
}

// GetByStatus retrieves all crop cycles with a specific status
func (r *cropCycleRepository) GetByStatus(ctx context.Context, status string) ([]entities.CropCycle, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "status", Operator: base.OpEqual, Value: status},
			},
			Logic: base.LogicAnd,
		},
	}

	return r.List(ctx, filter)
}

// GetActiveCycles retrieves all active crop cycles
func (r *cropCycleRepository) GetActiveCycles(ctx context.Context) ([]entities.CropCycle, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "status", Operator: base.OpEqual, Value: "ACTIVE"},
			},
			Logic: base.LogicAnd,
		},
	}

	return r.List(ctx, filter)
}
