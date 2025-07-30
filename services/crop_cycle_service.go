package services

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type CropCycleService struct {
	repo     repositories.CropCycleRepositoryInterface
	farmRepo repositories.FarmRepositoryInterface
}

func NewCropCycleService(
	repo repositories.CropCycleRepositoryInterface,
	farmRepo repositories.FarmRepositoryInterface,
) *CropCycleService {
	return &CropCycleService{repo: repo, farmRepo: farmRepo}
}

type CropCycleServiceInterface interface {
	CreateCropCycle(
		farmID, cropID string,
		startDate time.Time,
		endDate *time.Time,
		acreage float64,
		expectedQuantity float64,
		quantity *float64,
		report string,
		noOfCrops *int,
	) (*models.CropCycle, error)

	GetCropCycleByID(id string) (*models.CropCycle, error)
	GetCropCycles(farmID string, cropID *string, status *string) ([]*models.CropCycle, error)
	UpdateCropCycleByID(id string, endDate *time.Time, quantity *float64, report *string) (*models.CropCycle, error)
	ValidateCropCycleBelongsToFarm(cycleID, farmID string) (*models.CropCycle, error)

	// Batch methods
	GetCropCyclesBatch(farmIDs []string, cropID *string, status *string) (map[string]interface{}, map[string]string)
}

func (s *CropCycleService) CreateCropCycle(
	farmID, cropID string,
	startDate time.Time,
	endDate *time.Time,
	acreage float64,
	expectedQuantity float64,
	quantity *float64,
	report string,
	noOfCrops *int) (*models.CropCycle, error) {
	// Step 1: Validate required fields
	if farmID == "" {
		return nil, fmt.Errorf("farm ID is required")
	}
	if cropID == "" {
		return nil, fmt.Errorf("crop ID is required")
	}
	if startDate.IsZero() {
		return nil, fmt.Errorf("start date is required")
	}
	if acreage <= 0 {
		return nil, fmt.Errorf("acreage must be positive")
	}

	// Step 2: Fetch farm by farm_id
	farm, err := s.farmRepo.GetFarmByID(farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch farm: %w", err)
	}

	// Step 3: Calculate total existing acreage for farm's crop cycles
	usedAcreage, err := s.repo.GetTotalAcreageByFarmID(farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch used acreage: %w", err)
	}

	// Step 4: Check if new acreage exceeds farm's area
	if usedAcreage+acreage > farm.Area {
		return nil, fmt.Errorf("acreage exceeds available area on farm")
	}

	// Step 5: Create CropCycle record
	cycle := &models.CropCycle{
		FarmID:           farmID,
		CropID:           cropID,
		StartDate:        startDate,
		EndDate:          endDate,
		Acreage:          acreage,
		ExpectedQuantity: &expectedQuantity,
		NoOfCrops:        noOfCrops,
		Report:           report,
	}

	if quantity != nil {
		cycle.Quantity = *quantity
	}

	// Status will be set in the repository layer based on whether EndDate is provided

	return s.repo.Create(cycle)
}

func (s *CropCycleService) GetCropCycleByID(id string) (*models.CropCycle, error) {
	return s.repo.FindByID(id)
}

func (s *CropCycleService) GetCropCycles(farmID string, cropID *string, status *string) ([]*models.CropCycle, error) {
	// Validate farmID
	if farmID == "" {
		return nil, fmt.Errorf("farm ID is required")
	}

	// Validate status if provided
	if status != nil && *status != "" {
		if *status != models.CycleStatusOngoing && *status != models.CycleStatusCompleted {
			return nil, fmt.Errorf("invalid status: must be either ONGOING or COMPLETED")
		}
	}

	return s.repo.FindByFarm(farmID, cropID, status)
}

func (s *CropCycleService) UpdateCropCycleByID(
	id string,
	endDate *time.Time,
	quantity *float64,
	report *string,
) (*models.CropCycle, error) {
	// Step 1: Fetch crop cycle by id
	cycle, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	// Step 2: Validate inputs
	if endDate != nil {
		// Validate that end date is after start date
		if endDate.Before(cycle.StartDate) {
			return nil, fmt.Errorf("end date must be after start date")
		}
	}

	// Prepare report string
	var reportStr string
	if report != nil {
		reportStr = *report
	} else {
		reportStr = cycle.Report
	}

	// Step 3 & 4: Update cycle and save to DB
	return s.repo.UpdateCropCycleByID(id, endDate, quantity, reportStr)
}

// ValidateCropCycleBelongsToFarm checks if a cycle belongs to the specified farm
func (s *CropCycleService) ValidateCropCycleBelongsToFarm(cycleID, farmID string) (*models.CropCycle, error) {
	cycle, err := s.repo.FindByID(cycleID)
	if err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	if cycle.FarmID != farmID {
		return nil, fmt.Errorf("crop cycle does not belong to the specified farm")
	}

	return cycle, nil
}

// GetCropCyclesBatch retrieves crop cycles for multiple farms
func (s *CropCycleService) GetCropCyclesBatch(farmIDs []string, cropID *string, status *string) (map[string]interface{}, map[string]string) {
	data := make(map[string]interface{})
	errors := make(map[string]string)

	for _, farmID := range farmIDs {
		cycles, err := s.GetCropCycles(farmID, cropID, status)
		if err != nil {
			errors[farmID] = err.Error()
		} else {
			data[farmID] = cycles
		}
	}

	return data, errors
}
