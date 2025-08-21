package farm_activity

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// FarmActivityRepository defines the interface for farm activity data operations
type FarmActivityRepository interface {
	// Farm activity-specific operations
	GetByCycle(ctx context.Context, cycleID string) ([]farm_activity.FarmActivity, error)
	GetByType(ctx context.Context, activityType string) ([]farm_activity.FarmActivity, error)
	GetPlannedActivities(ctx context.Context) ([]farm_activity.FarmActivity, error)
	GetCompletedActivities(ctx context.Context) ([]farm_activity.FarmActivity, error)
}

// farmActivityRepository implements FarmActivityRepository
type farmActivityRepository struct {
	postgresManager *db.PostgresManager
}

// NewFarmActivityRepository creates a new farm activity repository
func NewFarmActivityRepository(dbManager db.DBManager) FarmActivityRepository {
	return &farmActivityRepository{
		postgresManager: dbManager.(*db.PostgresManager),
	}
}

// GetByCycle retrieves all activities for a specific crop cycle
func (r *farmActivityRepository) GetByCycle(ctx context.Context, cycleID string) ([]farm_activity.FarmActivity, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "cycle_id", Operator: base.OpEqual, Value: cycleID},
			},
			Logic: base.LogicAnd,
		},
	}

	var activities []farm_activity.FarmActivity
	err := r.postgresManager.List(ctx, filter, &activities)
	return activities, err
}

// GetByType retrieves all activities of a specific type
func (r *farmActivityRepository) GetByType(ctx context.Context, activityType string) ([]farm_activity.FarmActivity, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "activity_type", Operator: base.OpEqual, Value: activityType},
			},
			Logic: base.LogicAnd,
		},
	}

	var activities []farm_activity.FarmActivity
	err := r.postgresManager.List(ctx, filter, &activities)
	return activities, err
}

// GetPlannedActivities retrieves all planned activities
func (r *farmActivityRepository) GetPlannedActivities(ctx context.Context) ([]farm_activity.FarmActivity, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "completed_at", Operator: base.OpIsNull, Value: nil},
			},
			Logic: base.LogicAnd,
		},
	}

	var activities []farm_activity.FarmActivity
	err := r.postgresManager.List(ctx, filter, &activities)
	return activities, err
}

// GetCompletedActivities retrieves all completed activities
func (r *farmActivityRepository) GetCompletedActivities(ctx context.Context) ([]farm_activity.FarmActivity, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "completed_at", Operator: base.OpIsNotNull, Value: nil},
			},
			Logic: base.LogicAnd,
		},
	}

	var activities []farm_activity.FarmActivity
	err := r.postgresManager.List(ctx, filter, &activities)
	return activities, err
}
