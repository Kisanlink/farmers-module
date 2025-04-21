package repositories

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

type CropCycleRepository struct {
	db *gorm.DB
}

func NewCropCycleRepository(db *gorm.DB) *CropCycleRepository {
	return &CropCycleRepository{db: db}
}

type CropCycleRepositoryInterface interface {
	Create(cycle *models.CropCycle) (*models.CropCycle, error)
	GetTotalAcreageByFarmID(farmID string) (float64, error)

	FindByID(id string) (*models.CropCycle, error)
	Update(cycle *models.CropCycle) (*models.CropCycle, error)
	Delete(id string) error

	FindByFarm(farmID string, cropID *string, status *string) ([]*models.CropCycle, error)
	GetCropCyclesByFarmIDAndStatus(farmID, status string) ([]*models.CropCycle, error)
	UpdateCropCycleByID(id string, endDate *time.Time, quantity *float64, report string) (*models.CropCycle, error)
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

	if err := r.db.Omit("Crop").Create(cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to create crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.db.Preload("Crop").First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to load created crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) GetTotalAcreageByFarmID(farmID string) (float64, error) {
	var total float64
	err := r.db.Model(&models.CropCycle{}).
		Where("farm_id = ?", farmID).
		Select("COALESCE(SUM(acreage), 0)").
		Scan(&total).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get total acreage: %w", err)
	}
	return total, nil
}

func (r *CropCycleRepository) FindByID(id string) (*models.CropCycle, error) {
	var cycle models.CropCycle
	if err := r.db.Preload("Crop").
		First(&cycle, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}
	return &cycle, nil
}

func (r *CropCycleRepository) FindByFarm(farmID string, cropID *string, status *string) ([]*models.CropCycle, error) {
	var cycles []*models.CropCycle
	query := r.db.Preload("Crop").Where("farm_id = ?", farmID)

	if cropID != nil {
		query = query.Where("crop_id = ?", *cropID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Find(&cycles).Error; err != nil {
		return nil, fmt.Errorf("failed to query crop cycles: %w", err)
	}
	return cycles, nil
}

func (r *CropCycleRepository) GetCropCyclesByFarmIDAndStatus(farmID, status string) ([]*models.CropCycle, error) {
	return r.FindByFarm(farmID, nil, &status)
}

func (r *CropCycleRepository) UpdateCropCycleByID(id string, endDate *time.Time, quantity *float64, report string) (*models.CropCycle, error) {
	var cycle models.CropCycle
	if err := r.db.First(&cycle, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	if endDate != nil {
		cycle.EndDate = endDate
		cycle.Status = models.CycleStatusCompleted
	}

	if quantity != nil {
		cycle.Quantity = *quantity
	}

	cycle.Report = report
	cycle.UpdatedAt = time.Now()

	if err := r.db.Save(&cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to update crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.db.Preload("Crop").First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) Update(cycle *models.CropCycle) (*models.CropCycle, error) {
	cycle.UpdatedAt = time.Now()
	if err := r.db.Save(cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to update crop cycle: %w", err)
	}

	var full models.CropCycle
	if err := r.db.Preload("Crop").
		First(&full, "id = ?", cycle.Id).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated crop cycle: %w", err)
	}

	return &full, nil
}

func (r *CropCycleRepository) Delete(id string) error {
	if err := r.db.Delete(&models.CropCycle{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete crop cycle: %w", err)
	}
	return nil
}
