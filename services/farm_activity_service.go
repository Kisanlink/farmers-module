package services

import (
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

// FarmActivityServiceInterface declares the methods for farm activity services.
type FarmActivityServiceInterface interface {
	CreateActivity(activity *models.FarmActivity) error
	GetActivitiesByFarmID(farmID string) ([]*models.FarmActivity, error)
	GetActivitiesByCropCycle(cycleID string) ([]*models.FarmActivity, error)
	GetActivityByID(id string) (*models.FarmActivity, error)
	GetActivitiesByDateRange(farmID string, start, end time.Time) ([]*models.FarmActivity, error)
	UpdateActivity(activity *models.FarmActivity) error
	DeleteActivity(id string) error

	// Batch methods
	GetActivitiesByFarmIDsBatch(farmIDs []string) (map[string]interface{}, map[string]string)
	GetActivitiesByDateRangeBatch(farmIDs []string, start, end time.Time) (map[string]interface{}, map[string]string)
}

// FarmActivityService implements FarmActivityServiceInterface.
type FarmActivityService struct {
	repo repositories.FarmActivityRepositoryInterface
}

// NewFarmActivityService creates a new instance of FarmActivityService.
func NewFarmActivityService(repo repositories.FarmActivityRepositoryInterface) *FarmActivityService {
	return &FarmActivityService{repo: repo}
}

// CreateActivity creates a new farm activity.
func (s *FarmActivityService) CreateActivity(activity *models.FarmActivity) error {

	return s.repo.CreateActivity(activity)

}

// GetActivitiesByFarmID retrieves activities for a given farm.
func (s *FarmActivityService) GetActivitiesByFarmID(farmID string) ([]*models.FarmActivity, error) {
	return s.repo.GetActivitiesByFarmID(farmID)
}

// GetActivitiesByCropCycle retrieves activities for a given crop cycle.
func (s *FarmActivityService) GetActivitiesByCropCycle(cycleID string) ([]*models.FarmActivity, error) {
	return s.repo.GetActivitiesByCropCycle(cycleID)
}

// GetActivityByID retrieves a single activity by its ID.
func (s *FarmActivityService) GetActivityByID(id string) (*models.FarmActivity, error) {
	return s.repo.GetActivityByID(id)
}

// GetActivitiesByDateRange retrieves activities for a farm within a given date range.
func (s *FarmActivityService) GetActivitiesByDateRange(farmID string, start, end time.Time) ([]*models.FarmActivity, error) {
	return s.repo.GetActivitiesByDateRange(farmID, start, end)
}

// UpdateActivity updates an existing farm activity.
func (s *FarmActivityService) UpdateActivity(activity *models.FarmActivity) error {
	return s.repo.UpdateActivity(activity)
}

// DeleteActivity deletes a farm activity by its ID.
func (s *FarmActivityService) DeleteActivity(id string) error {
	return s.repo.DeleteActivity(id)
}

// GetActivitiesByFarmIDsBatch retrieves activities for multiple farms
func (s *FarmActivityService) GetActivitiesByFarmIDsBatch(farmIDs []string) (map[string]interface{}, map[string]string) {
	data := make(map[string]interface{})
	errors := make(map[string]string)

	for _, farmID := range farmIDs {
		activities, err := s.GetActivitiesByFarmID(farmID)
		if err != nil {
			errors[farmID] = err.Error()
		} else {
			data[farmID] = activities
		}
	}

	return data, errors
}

// GetActivitiesByDateRangeBatch retrieves activities for multiple farms within a date range
func (s *FarmActivityService) GetActivitiesByDateRangeBatch(farmIDs []string, start, end time.Time) (map[string]interface{}, map[string]string) {
	data := make(map[string]interface{})
	errors := make(map[string]string)

	for _, farmID := range farmIDs {
		activities, err := s.GetActivitiesByDateRange(farmID, start, end)
		if err != nil {
			errors[farmID] = err.Error()
		} else {
			data[farmID] = activities
		}
	}

	return data, errors
}
