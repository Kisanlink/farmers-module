package repositories

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

type CropCycleRepository struct {
	DB *gorm.DB
}

func NewCropCycleRepository(db *gorm.DB) *CropCycleRepository {
	return &CropCycleRepository{DB: db}
}

type CropCycleRepositoryInterface interface {
	Create(cycle *models.CropCycle) (*models.CropCycle, error)
	GetTotalAcreageByFarmID(farm_id string) (float64, error)

	FindByID(id string) (*models.CropCycle, error)
	Update(cycle *models.CropCycle) (*models.CropCycle, error)
	Delete(id string) error

	FindByFarm(farm_id string, crop_id *string, status *string) ([]*models.CropCycle, error)
	GetCropCyclesByFarmIDAndStatus(farm_id, status string) ([]*models.CropCycle, error)
	UpdateCropCycleById(id string, end_date *time.Time, quantity *float64, report string) (*models.CropCycle, error)
}

func (r *CropCycleRepository) Create(cycle *models.CropCycle) (*models.CropCycle, error) {
	cycle.Id = utils.GenerateCycleId()
	now := time.Now()
	cycle.CreatedAt = now
	cycle.UpdatedAt = now

	if cycle.EndDate == nil {
		cycle.Status = models.CycleStatusOngoing
	} else {
		cycle.Status = models.CycleStatusCompleted
	}

	if err := r.DB.Omit("Crop").Create(cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to create crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.DB.Preload("Crop").First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to load created crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) GetTotalAcreageByFarmID(farm_id string) (float64, error) {
	var total float64
	err := r.DB.Model(&models.CropCycle{}).
		Where("farm_id = ?", farm_id).
		Select("COALESCE(SUM(acreage), 0)").
		Scan(&total).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get total acreage: %w", err)
	}
	return total, nil
}

func (r *CropCycleRepository) FindByID(id string) (*models.CropCycle, error) {
	var cycle models.CropCycle
	if err := r.DB.Preload("Crop").
		First(&cycle, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}
	return &cycle, nil
}

func (r *CropCycleRepository) FindByFarm(farm_id string, crop_id *string, status *string) ([]*models.CropCycle, error) {
	var cycles []*models.CropCycle
	query := r.DB.Preload("Crop").Where("farm_id = ?", farm_id)

	if crop_id != nil {
		query = query.Where("crop_id = ?", *crop_id)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Find(&cycles).Error; err != nil {
		return nil, fmt.Errorf("failed to query crop cycles: %w", err)
	}
	return cycles, nil
}

func (r *CropCycleRepository) GetCropCyclesByFarmIDAndStatus(farm_id, status string) ([]*models.CropCycle, error) {
	return r.FindByFarm(farm_id, nil, &status)
}

func (r *CropCycleRepository) UpdateCropCycleById(id string, end_date *time.Time, quantity *float64, report string) (*models.CropCycle, error) {
	var cycle models.CropCycle
	if err := r.DB.First(&cycle, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	if end_date != nil {
		cycle.EndDate = end_date
		cycle.Status = models.CycleStatusCompleted
	}

	if quantity != nil {
		cycle.Quantity = *quantity
	}

	cycle.Report = report
	cycle.UpdatedAt = time.Now()

	if err := r.DB.Save(&cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to update crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.DB.Preload("Crop").First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) Update(cycle *models.CropCycle) (*models.CropCycle, error) {
	cycle.UpdatedAt = time.Now()
	if err := r.DB.Save(cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to update crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.DB.Preload("Crop").
		First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) Delete(id string) error {
	if err := r.DB.Delete(&models.CropCycle{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete crop cycle: %w", err)
	}
	return nil
}
