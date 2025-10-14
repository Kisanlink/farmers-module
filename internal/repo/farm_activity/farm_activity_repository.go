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

// Find overrides the base Find method to preload CropStage relationship
func (r *FarmActivityRepository) Find(ctx context.Context, filter *base.Filter) ([]*farm_activity.FarmActivity, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var activities []*farm_activity.FarmActivity
	query := r.db.WithContext(ctx).Preload("CropStage").Preload("CropStage.Stage")

	// Apply filters
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			query = query.Where(condition.Field+" "+string(condition.Operator)+" ?", condition.Value)
		}
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		query = query.Limit(filter.Limit)
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	if err := query.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to find activities: %w", err)
	}

	return activities, nil
}

// GetByID overrides the base GetByID method to preload CropStage relationship
func (r *FarmActivityRepository) GetByID(ctx context.Context, id string, dest *farm_activity.FarmActivity) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("database connection not available")
	}

	err := r.db.WithContext(ctx).
		Preload("CropStage").
		Preload("CropStage.Stage").
		Where("id = ? AND deleted_at IS NULL", id).
		First(dest).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get activity by ID: %w", err)
	}

	return true, nil
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

// StageCompletionStat represents completion statistics for a crop stage
type StageCompletionStat struct {
	CropStageID          string
	StageID              string
	StageName            string
	StageOrder           int
	DurationDays         *int
	TotalActivities      int
	CompletedActivities  int
	InProgressActivities int
	PlannedActivities    int
	CompletionPercent    float64
}

// GetStageCompletionStats retrieves stage-wise activity completion statistics for a crop cycle
// This method is meant to be called from the service layer which will pass crop stages
// The service layer should use CropStageRepository to get the stages first
func (r *FarmActivityRepository) GetStageCompletionStats(ctx context.Context, cropCycleID string, cropStages interface{}) ([]*StageCompletionStat, error) {
	// Cast the interface to the expected type
	type CropStageInfo struct {
		ID           string
		StageID      string
		StageName    string
		StageOrder   int
		DurationDays *int
	}

	var stageInfos []*CropStageInfo

	// Type assertion to extract stage info
	switch v := cropStages.(type) {
	case []*CropStageInfo:
		stageInfos = v
	default:
		return nil, fmt.Errorf("invalid crop stages type")
	}

	var stats []*StageCompletionStat

	// For each crop stage, get activities and calculate statistics
	for _, stageInfo := range stageInfos {
		// Use BaseFilterableRepository pattern to get activities for this stage
		activityFilter := base.NewFilterBuilder().
			Where("crop_cycle_id", base.OpEqual, cropCycleID).
			Where("crop_stage_id", base.OpEqual, stageInfo.ID).
			Build()

		activities, err := r.Find(ctx, activityFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get activities for stage %s: %w", stageInfo.ID, err)
		}

		// Calculate statistics in-memory
		stat := &StageCompletionStat{
			CropStageID:          stageInfo.ID,
			StageID:              stageInfo.StageID,
			StageName:            stageInfo.StageName,
			StageOrder:           stageInfo.StageOrder,
			DurationDays:         stageInfo.DurationDays,
			TotalActivities:      len(activities),
			CompletedActivities:  0,
			InProgressActivities: 0,
			PlannedActivities:    0,
		}

		// Count by status
		for _, activity := range activities {
			switch activity.Status {
			case "COMPLETED":
				stat.CompletedActivities++
			case "IN_PROGRESS":
				stat.InProgressActivities++
			case "PLANNED":
				stat.PlannedActivities++
			}
		}

		// Calculate completion percentage
		if stat.TotalActivities > 0 {
			stat.CompletionPercent = float64(stat.CompletedActivities) / float64(stat.TotalActivities) * 100
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
