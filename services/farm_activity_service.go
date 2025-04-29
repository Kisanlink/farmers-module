package services

import (
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

// FarmActivityServiceInterface declares the methods for farm activity services.
type FarmActivityServiceInterface interface {
	CreateActivity(activity *models.FarmActivity) error
	GetActivitiesByFarmId(farm_id string) ([]*models.FarmActivity, error)
	GetActivitiesByCropCycle(cycle_id string) ([]*models.FarmActivity, error)
	GetActivityById(id string) (*models.FarmActivity, error)
	GetActivitiesByDateRange(farm_id string, start, end time.Time) ([]*models.FarmActivity, error)
	UpdateActivity(activity *models.FarmActivity) error
	DeleteActivity(id string) error
}

// FarmActivityService implements FarmActivityServiceInterface.
type FarmActivityService struct {
	Repo repositories.FarmActivityRepositoryInterface
}

// NewFarmActivityService creates a new instance of FarmActivityService.
func NewFarmActivityService(repo repositories.FarmActivityRepositoryInterface) *FarmActivityService {
	return &FarmActivityService{Repo: repo}
}

// CreateActivity creates a new farm activity.
func (s *FarmActivityService) CreateActivity(activity *models.FarmActivity) error {

	return s.Repo.CreateActivity(activity)

}

// GetActivitiesByFarmId retrieves activities for a given farm.
func (s *FarmActivityService) GetActivitiesByFarmId(farm_id string) ([]*models.FarmActivity, error) {
	return s.Repo.GetActivitiesByFarmId(farm_id)
}

// GetActivitiesByCropCycle retrieves activities for a given crop cycle.
func (s *FarmActivityService) GetActivitiesByCropCycle(cycle_id string) ([]*models.FarmActivity, error) {
	return s.Repo.GetActivitiesByCropCycle(cycle_id)
}

// GetActivityById retrieves a single activity by its ID.
func (s *FarmActivityService) GetActivityById(id string) (*models.FarmActivity, error) {
	return s.Repo.GetActivityById(id)
}

// GetActivitiesByDateRange retrieves activities for a farm within a given date range.
func (s *FarmActivityService) GetActivitiesByDateRange(farm_id string, start, end time.Time) ([]*models.FarmActivity, error) {
	return s.Repo.GetActivitiesByDateRange(farm_id, start, end)
}

// UpdateActivity updates an existing farm activity.
func (s *FarmActivityService) UpdateActivity(activity *models.FarmActivity) error {
	return s.Repo.UpdateActivity(activity)
}

// DeleteActivity deletes a farm activity by its ID.
func (s *FarmActivityService) DeleteActivity(id string) error {
	return s.Repo.DeleteActivity(id)
}
