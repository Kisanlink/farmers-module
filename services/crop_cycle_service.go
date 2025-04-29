package services

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type CropCycleService struct {
	Repo     repositories.CropCycleRepositoryInterface
	FarmRepo repositories.FarmRepositoryInterface
}

func NewCropCycleService(
	repo repositories.CropCycleRepositoryInterface,
	farmRepo repositories.FarmRepositoryInterface,
) *CropCycleService {
	return &CropCycleService{Repo: repo, FarmRepo: farmRepo}
}

type CropCycleServiceInterface interface {
	CreateCropCycle(
		farm_id, crop_id string,
		start_date time.Time,
		end_date *time.Time,
		acreage float64,
		expected_quantity float64,
		quantity *float64,
		report string,
	) (*models.CropCycle, error)

	GetCropCycleById(id string) (*models.CropCycle, error)
	GetCropCycles(farm_id string, crop_id *string, status *string) ([]*models.CropCycle, error)
	UpdateCropCycleById(id string, end_date *time.Time, quantity *float64, report *string) (*models.CropCycle, error)
	ValidateCropCycleBelongsToFarm(cycle_id, farm_id string) (*models.CropCycle, error)
}

func (s *CropCycleService) CreateCropCycle(
	farm_id, crop_id string,
	start_date time.Time,
	end_date *time.Time,
	acreage float64,
	expected_quantity float64,
	quantity *float64,
	report string,
) (*models.CropCycle, error) {
	// Step 1: Validate required fields
	if farm_id == "" {
		return nil, fmt.Errorf("farm ID is required")
	}
	if crop_id == "" {
		return nil, fmt.Errorf("crop ID is required")
	}
	if start_date.IsZero() {
		return nil, fmt.Errorf("start date is required")
	}
	if acreage <= 0 {
		return nil, fmt.Errorf("acreage must be positive")
	}

	// Step 2: Fetch farm by farm_id
	farm, err := s.FarmRepo.GetFarmByID(farm_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch farm: %w", err)
	}

	// Step 3: Calculate total existing acreage for farm's crop cycles
	used_acreage, err := s.Repo.GetTotalAcreageByFarmID(farm_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch used acreage: %w", err)
	}

	// Step 4: Check if new acreage exceeds farm's area
	if used_acreage+acreage > farm.Area {
		return nil, fmt.Errorf("acreage exceeds available area on farm")
	}

	// Step 5: Create CropCycle record
	cycle := &models.CropCycle{
		FarmId:           farm_id,
		CropId:           crop_id,
		StartDate:        start_date,
		EndDate:          end_date,
		Acreage:          acreage,
		ExpectedQuantity: &expected_quantity,
		Report:           report,
	}

	if quantity != nil {
		cycle.Quantity = *quantity
	}

	// Status will be set in the repository layer based on whether EndDate is provided

	return s.Repo.Create(cycle)
}

func (s *CropCycleService) GetCropCycleById(id string) (*models.CropCycle, error) {
	return s.Repo.FindByID(id)
}

func (s *CropCycleService) GetCropCycles(farm_id string, crop_id *string, status *string) ([]*models.CropCycle, error) {
	// Validate farm_id
	if farm_id == "" {
		return nil, fmt.Errorf("farm ID is required")
	}

	// Validate status if provided
	if status != nil && *status != "" {
		if *status != models.CycleStatusOngoing && *status != models.CycleStatusCompleted {
			return nil, fmt.Errorf("invalid status: must be either ONGOING or COMPLETED")
		}
	}

	return s.Repo.FindByFarm(farm_id, crop_id, status)
}

func (s *CropCycleService) UpdateCropCycleById(
	id string,
	end_date *time.Time,
	quantity *float64,
	report *string,
) (*models.CropCycle, error) {
	// Step 1: Fetch crop cycle by id
	cycle, err := s.Repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	// Step 2: Validate inputs
	if end_date != nil {
		// Validate that end date is after start date
		if end_date.Before(cycle.StartDate) {
			return nil, fmt.Errorf("end date must be after start date")
		}
	}

	// Prepare report string
	var report_str string
	if report != nil {
		report_str = *report
	} else {
		report_str = cycle.Report
	}

	// Step 3 & 4: Update cycle and save to DB
	return s.Repo.UpdateCropCycleById(id, end_date, quantity, report_str)
}

// ValidateCropCycleBelongsToFarm checks if a cycle belongs to the specified farm
func (s *CropCycleService) ValidateCropCycleBelongsToFarm(cycle_id, farm_id string) (*models.CropCycle, error) {
	cycle, err := s.Repo.FindByID(cycle_id)
	if err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	if cycle.FarmId != farm_id {
		return nil, fmt.Errorf("crop cycle does not belong to the specified farm")
	}

	return cycle, nil
}
