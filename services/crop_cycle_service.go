package services

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type CropCycleServiceInterface interface {
	CreateCropCycle(farmId, cropId string, startDate, endDate time.Time, acreage, expectedQuantity, quantity float64, report string) (*models.CropCycle, error)
	GetCropCyclesByFarmID(farmID string) ([]*models.CropCycle, error)
	GetCropCyclesByCropID(cropID string) ([]*models.CropCycle, error)
	GetCropCycleByID(id string) (*models.CropCycle, error)
	GetCropCyclesByFarmAndCropID(farmID, cropID string) ([]*models.CropCycle, error)
}

type CropCycleService struct {
	repo repositories.CropCycleRepositoryInterface
}

func NewCropCycleService(repo repositories.CropCycleRepositoryInterface) *CropCycleService {
	return &CropCycleService{repo: repo}
}

func (s *CropCycleService) CreateCropCycle(
	farmId, cropId string,
	startDate, endDate time.Time, acreage, expectedQuantity, quantity float64, report string,
) (*models.CropCycle, error) {
	cycle := &models.CropCycle{
		FarmID:           farmId,
		CropID:           cropId,
		StartDate:        &startDate,
		EndDate:          &endDate,
		Acreage:          acreage,
		ExpectedQuantity: expectedQuantity,
		Quantity:         quantity,
		Report:           report,
	}

	res, err := s.repo.CreateCropCycle(cycle)
	if err != nil {
		return nil, fmt.Errorf("failed to create crop cycle: %w", err)
	}
	return res, nil
}

func (s *CropCycleService) GetCropCyclesByFarmID(farmID string) ([]*models.CropCycle, error) {
	return s.repo.GetCropCyclesByFarmID(farmID)
}

func (s *CropCycleService) GetCropCyclesByCropID(cropID string) ([]*models.CropCycle, error) {
	return s.repo.GetCropCyclesByCropID(cropID)
}

func (s *CropCycleService) GetCropCycleByID(id string) (*models.CropCycle, error) {
	return s.repo.GetCropCycleByID(id)
}

func (s *CropCycleService) GetCropCyclesByFarmAndCropID(farmID, cropID string) ([]*models.CropCycle, error) {
	return s.repo.GetCropCyclesByFarmAndCropID(farmID, cropID)
}
