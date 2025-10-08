package farm_activity

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmActivityRepository wraps BaseFilterableRepository with custom methods
type FarmActivityRepository struct {
	*base.BaseFilterableRepository[*farm_activity.FarmActivity]
	db *gorm.DB
}

// NewFarmActivityRepository creates a new farm activity repository using BaseFilterableRepository
func NewFarmActivityRepository(dbManager interface{}) *FarmActivityRepository {
	baseRepo := base.NewBaseFilterableRepository[*farm_activity.FarmActivity]()
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

	return &FarmActivityRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *FarmActivityRepository) Count(ctx context.Context, filter *base.Filter, model *farm_activity.FarmActivity) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&farm_activity.FarmActivity{}).WithContext(ctx)

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
