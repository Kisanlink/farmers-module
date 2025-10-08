package crop_cycle

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// CropCycleRepository wraps BaseFilterableRepository with custom methods
type CropCycleRepository struct {
	*base.BaseFilterableRepository[*crop_cycle.CropCycle]
	db *gorm.DB
}

// NewRepository creates a new crop cycle repository using BaseFilterableRepository
func NewRepository(dbManager interface{}) *CropCycleRepository {
	baseRepo := base.NewBaseFilterableRepository[*crop_cycle.CropCycle]()
	baseRepo.SetDBManager(dbManager)

	// Get the GORM DB instance
	var db *gorm.DB
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			db = gormDB
		}
	}

	return &CropCycleRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *CropCycleRepository) Count(ctx context.Context, filter *base.Filter, model *crop_cycle.CropCycle) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&crop_cycle.CropCycle{}).WithContext(ctx)

	// Apply filters
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			query = query.Where(condition.Field+" "+string(condition.Operator)+" ?", condition.Value)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
