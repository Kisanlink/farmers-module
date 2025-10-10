package farmer

import (
	"context"
	"fmt"

	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmerRepository wraps BaseFilterableRepository with custom methods for normalized Farmer entity
type FarmerRepository struct {
	*base.BaseFilterableRepository[*farmerentity.Farmer]
	db *gorm.DB
}

// FarmerLinkRepository wraps BaseFilterableRepository with custom methods
type FarmerLinkRepository struct {
	*base.BaseFilterableRepository[*farmerentity.FarmerLink]
	db *gorm.DB
}

// NewFarmerRepository creates a new farmer repository using BaseFilterableRepository with normalized Farmer entity
func NewFarmerRepository(dbManager interface{}) *FarmerRepository {
	baseRepo := base.NewBaseFilterableRepository[*farmerentity.Farmer]()
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
func (r *FarmerRepository) Count(ctx context.Context, filter *base.Filter, model *farmerentity.Farmer) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&farmerentity.Farmer{}).WithContext(ctx)

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
	baseRepo := base.NewBaseFilterableRepository[*farmerentity.FarmerLink]()
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
func (r *FarmerLinkRepository) Count(ctx context.Context, filter *base.Filter, model *farmerentity.FarmerLink) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&farmerentity.FarmerLink{}).WithContext(ctx)

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
