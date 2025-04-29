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
	DB *gorm.DB
}

// NewFarmActivityRepository creates a new instance of FarmActivityRepository.
func NewFarmActivityRepository(db *gorm.DB) *FarmActivityRepository {
	return &FarmActivityRepository{DB: db}
}

// FarmActivityRepositoryInterface defines repository methods for FarmActivity.
type FarmActivityRepositoryInterface interface {
	CreateActivity(activity *models.FarmActivity) error
	GetActivitiesByFarmId(farm_id string) ([]*models.FarmActivity, error)
	GetActivitiesByCropCycle(cycle_id string) ([]*models.FarmActivity, error)
	GetActivityById(id string) (*models.FarmActivity, error)
	GetActivitiesByDateRange(farm_id string, start, end time.Time) ([]*models.FarmActivity, error)
	UpdateActivity(activity *models.FarmActivity) error
	DeleteActivity(id string) error
}

var err_start_before_cycle = errors.New("activity start date is before crop cycle start date")

// CreateActivity creates a new farm activity record.
func (r *FarmActivityRepository) CreateActivity(activity *models.FarmActivity) error {
	// 1) load the parent crop cycle to read its StartDate
	var cycle models.CropCycle
	if err := r.DB.
		Where("id = ?", activity.CropCycleId).
		First(&cycle).Error; err != nil {
		return fmt.Errorf("failed to fetch crop cycle %q: %w", activity.CropCycleId, err)
	}

	// 2) enforce the business rule
	if activity.StartDate.Before(cycle.StartDate) {
		return err_start_before_cycle
	}

	// 3) metadata + insert
	activity.Id = utils.GenerateActId()
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if err := r.DB.Create(activity).Error; err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}
	return nil
}

// GetActivitiesByFarmId retrieves activities for a given farm ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivitiesByFarmId(farm_id string) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity
	if err := r.DB.
		Where("farm_id = ?", farm_id).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for farm_id %s: %w", farm_id, err)
	}
	return activities, nil
}

// GetActivitiesByCropCycle retrieves activities for a given crop cycle ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivitiesByCropCycle(cycle_id string) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity
	if err := r.DB.
		Where("crop_cycle_id = ?", cycle_id).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for crop_cycle_id %s: %w", cycle_id, err)
	}
	return activities, nil
}

// GetActivityById retrieves a single activity by its ID with CropCycle preloaded.
func (r *FarmActivityRepository) GetActivityById(id string) (*models.FarmActivity, error) {
	var activity models.FarmActivity
	if err := r.DB.
		Where("id = ?", id).
		First(&activity).Error; err != nil {
		return nil, fmt.Errorf("failed to get activity by id %s: %w", id, err)
	}
	return &activity, nil
}

// GetActivitiesByDateRange retrieves activities for a given farm ID between the provided start and end dates.
// This method filters records based on the DATE portion of the created_at timestamp.
func (r *FarmActivityRepository) GetActivitiesByDateRange(farm_id string, start, end time.Time) ([]*models.FarmActivity, error) {
	var activities []*models.FarmActivity

	// Format dates as YYYY-MM-DD to compare only date portions.
	start_str := start.Format("2006-01-02")
	end_str := end.Format("2006-01-02")

	if err := r.DB.
		Where("farm_id = ? AND DATE(created_at) BETWEEN ? AND ?", farm_id, start_str, end_str).
		Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities for farm_id %s in date range: %w", farm_id, err)
	}
	return activities, nil
}

// UpdateActivity updates an existing activity record. It automatically updates the UpdatedAt timestamp.
func (r *FarmActivityRepository) UpdateActivity(activity *models.FarmActivity) error {
	activity.UpdatedAt = time.Now()
	if err := r.DB.Save(activity).Error; err != nil {
		return fmt.Errorf("failed to update activity with id %s: %w", activity.Id, err)
	}
	return nil
}

// DeleteActivity removes an activity record from the database by its ID.
func (r *FarmActivityRepository) DeleteActivity(id string) error {
	if err := r.DB.
		Where("id = ?", id).
		Delete(&models.FarmActivity{}).Error; err != nil {
		return fmt.Errorf("failed to delete activity with id %s: %w", id, err)
	}
	return nil
}
