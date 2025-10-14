package crop_cycle

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/pkg/common"
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

// AreaAllocationSummary represents area allocation summary for a farm
type AreaAllocationSummary struct {
	FarmID             string
	TotalAreaHa        float64
	AllocatedAreaHa    float64
	AvailableAreaHa    float64
	ActiveCyclesCount  int64
	PlannedCyclesCount int64
}

// GetTotalAllocatedArea calculates the total allocated area for a farm
// excluding a specific cycle ID (useful for updates)
func (r *CropCycleRepository) GetTotalAllocatedArea(ctx context.Context, farmID string, excludeCycleID string) (float64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	var totalArea float64
	query := r.db.WithContext(ctx).
		Model(&crop_cycle.CropCycle{}).
		Select("COALESCE(SUM(area_ha), 0)").
		Where("farm_id = ? AND status IN (?) AND deleted_at IS NULL",
			farmID, []string{"PLANNED", "ACTIVE"})

	if excludeCycleID != "" {
		query = query.Where("id != ?", excludeCycleID)
	}

	if err := query.Scan(&totalArea).Error; err != nil {
		return 0, err
	}

	return totalArea, nil
}

// ValidateAreaAllocation validates that the requested area doesn't exceed farm capacity
// Uses pessimistic locking with SELECT FOR UPDATE to prevent race conditions
func (r *CropCycleRepository) ValidateAreaAllocation(ctx context.Context, farmID string, cycleID string, requestedArea float64) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Start transaction with SERIALIZABLE isolation
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock farm record for update
		var farm farmEntity.Farm
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ? AND deleted_at IS NULL", farmID).
			First(&farm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return common.ErrNotFound
			}
			return err
		}

		// Calculate current allocation (excluding the cycle being updated)
		var totalAllocated float64
		if err := tx.Model(&crop_cycle.CropCycle{}).
			Where("farm_id = ? AND status IN (?) AND deleted_at IS NULL",
				farmID, []string{"PLANNED", "ACTIVE"}).
			Where("id != ?", cycleID).
			Select("COALESCE(SUM(area_ha), 0)").
			Scan(&totalAllocated).Error; err != nil {
			return err
		}

		// Calculate available area
		availableArea := farm.AreaHa - totalAllocated

		// Validate
		if requestedArea > availableArea {
			return &common.AreaExceededError{
				FarmID:        farmID,
				FarmArea:      farm.AreaHa,
				RequestedArea: requestedArea,
				AvailableArea: availableArea,
				AllocatedArea: totalAllocated,
			}
		}

		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
}

// GetAreaAllocationSummary retrieves area allocation summary for a farm
func (r *CropCycleRepository) GetAreaAllocationSummary(ctx context.Context, farmID string) (*AreaAllocationSummary, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Get farm details
	var farm farmEntity.Farm
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", farmID).
		First(&farm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, err
	}

	// Calculate allocated area and counts
	var allocatedArea float64
	var activeCycles int64
	var plannedCycles int64

	if err := r.db.WithContext(ctx).
		Model(&crop_cycle.CropCycle{}).
		Where("farm_id = ? AND status IN (?) AND deleted_at IS NULL",
			farmID, []string{"PLANNED", "ACTIVE"}).
		Select("COALESCE(SUM(area_ha), 0)").
		Scan(&allocatedArea).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&crop_cycle.CropCycle{}).
		Where("farm_id = ? AND status = ? AND deleted_at IS NULL", farmID, "ACTIVE").
		Count(&activeCycles).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&crop_cycle.CropCycle{}).
		Where("farm_id = ? AND status = ? AND deleted_at IS NULL", farmID, "PLANNED").
		Count(&plannedCycles).Error; err != nil {
		return nil, err
	}

	return &AreaAllocationSummary{
		FarmID:             farmID,
		TotalAreaHa:        farm.AreaHa,
		AllocatedAreaHa:    allocatedArea,
		AvailableAreaHa:    farm.AreaHa - allocatedArea,
		ActiveCyclesCount:  activeCycles,
		PlannedCyclesCount: plannedCycles,
	}, nil
}
