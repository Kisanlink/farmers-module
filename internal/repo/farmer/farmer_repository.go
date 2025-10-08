package farmer

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmerRepository wraps BaseFilterableRepository with custom methods
type FarmerRepository struct {
	*base.BaseFilterableRepository[*entities.FarmerProfile]
	db *gorm.DB
}

// FarmerLinkRepository wraps BaseFilterableRepository with custom methods
type FarmerLinkRepository struct {
	*base.BaseFilterableRepository[*entities.FarmerLink]
	db *gorm.DB
}

// NewFarmerRepository creates a new farmer repository using BaseFilterableRepository
func NewFarmerRepository(dbManager interface{}) *FarmerRepository {
	baseRepo := base.NewBaseFilterableRepository[*entities.FarmerProfile]()
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

	return &FarmerRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *FarmerRepository) Count(ctx context.Context, filter *base.Filter, model *entities.FarmerProfile) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&entities.FarmerProfile{}).WithContext(ctx)

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

// NewFarmerLinkRepository creates a new farmer link repository using BaseFilterableRepository
func NewFarmerLinkRepository(dbManager interface{}) *FarmerLinkRepository {
	baseRepo := base.NewBaseFilterableRepository[*entities.FarmerLink]()
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

	return &FarmerLinkRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *FarmerLinkRepository) Count(ctx context.Context, filter *base.Filter, model *entities.FarmerLink) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&entities.FarmerLink{}).WithContext(ctx)

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
