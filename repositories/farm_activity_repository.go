package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

// FarmActivityRepository handles CRUD operations for FarmActivity.
type FarmActivityRepository struct {
	db *gorm.DB
}

// NewFarmActivityRepository creates a new instance of FarmActivityRepository.
func NewFarmActivityRepository(db *gorm.DB) *FarmActivityRepository {
	return &FarmActivityRepository{db: db}
}

// FarmActivityRepositoryInterface defines repository methods for FarmActivity.
type FarmActivityRepositoryInterface interface {
	CreateActivity(activity *models.FarmActivity) error
	GetActivitiesByFarmID(farmID string) ([]*models.FarmActivity, error)
	GetActivitiesByCropCycle(cycleID string) ([]*models.FarmActivity, error)
	GetActivityByID(id string) (*models.FarmActivity, error)
	GetActivitiesByDateRange(farmID string, start, end time.Time) ([]*models.FarmActivity, error)
	UpdateActivity(activity *models.FarmActivity) error
	DeleteActivity(id string) error
}

var ErrStartBeforeCycle = errors.New("activity start date is before crop cycle start date")

// CreateActivity creates a new farm activity record.
func (r *FarmActivityRepository) CreateActivity(activity *models.FarmActivity) error {
	// 1) load the parent crop cycle to read its StartDate
	var cycle models.CropCycle
	if err := r.db.
		Where("id = ?", activity.CropCycleID).
		First(&cycle).Error; err != nil {
		return fmt.Errorf("failed to fetch crop cycle %q: %w", activity.CropCycleID, err)
	}

	// 2) enforce the business rule
	if activity.StartDate.Before(cycle.StartDate) {
		return ErrStartBeforeCycle
	}

	// 3) metadata + insert
	activity.Id = utils.GenerateActId()
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if err := r.db.Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

// GetActivitiesByFarmID retrieves activities for a given farm ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivitiesByFarmID(farmID string) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity
	if err := r.db.
		Where("farm_id = ?", farmID).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for farm_id %s: %w", farmID, err)
	}
	return activities, nil
}

// GetActivitiesByCropCycle retrieves activities for a given crop cycle ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivitiesByCropCycle(cycleID string) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity
	if err := r.db.
		Where("crop_cycle_id = ?", cycleID).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for crop_cycle_id %s: %w", cycleID, err)
	}
	return activities, nil
}

// GetActivityByID retrieves a single activity by its ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivityByID(id string) (*models.FarmActivity, error) {
	var activity models.FarmActivity
	if err := r.db.
		Where("id = ?", id).
		First(&activity).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity by id %s: %w", id, err)
	}
	return &activity, nil
}

// GetActivitiesByDateRange retrieves activities for a given farm ID between the provided start and end dates.
// This method filters records based on the DATE portion of the created_at timestamp.
func (r *FarmActivityRepository) GetActivitiesByDateRange(farmID string, start, end time.Time) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity

	// Format dates as YYYY-MM-DD to compare only date portions.
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	if err := r.db.
		Where("farm_id = ? AND DATE(created_at) BETWEEN ? AND ?", farmID, startStr, endStr).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for farm_id %s in date range: %w", farmID, err)
	}
	return activities, nil
}

// UpdateActivity updates an existing activity record. It automatically updates the UpdatedAt timestamp.
func (r *FarmActivityRepository) UpdateActivity(activity *models.FarmActivity) error {
	activity.UpdatedAt = time.Now()
	if err := r.db.Save(activity).Error; err != nil {
		return fmt.Errorf("failed to update activity with id %s: %w", activity.Id, err)
	}
	return nil
}

// DeleteActivity removes an activity record from the database by its ID.
func (r *FarmActivityRepository) DeleteActivity(id string) error {
	if err := r.db.
		Where("id = ?", id).
		Delete(&models.FarmActivity{}).Error; err != nil {
		return fmt.Errorf("failed to delete activity with id %s: %w", id, err)
	}
	return nil
}
