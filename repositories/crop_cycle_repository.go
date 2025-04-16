package repositories

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

// CropCycleRepository handles operations for the CropCycle model.
type CropCycleRepository struct {
	db *gorm.DB
}

// NewCropCycleRepository creates a new instance of CropCycleRepository.
func NewCropCycleRepository(db *gorm.DB) *CropCycleRepository {
	return &CropCycleRepository{db: db}
}

// CropCycleRepositoryInterface defines repository methods for CropCycle.
type CropCycleRepositoryInterface interface {
	CreateCropCycle(cycle *models.CropCycle) error
	GetCropCyclesByFarmID(farmID string) ([]*models.CropCycle, error)
	GetCropCyclesByCropID(cropID string) ([]*models.CropCycle, error)
	GetCropCyclesByFarmAndCropID(farmID, cropID string) ([]*models.CropCycle, error)
	GetCropCycleByID(id string) (*models.CropCycle, error)
	UpdateCropCycle(cycle *models.CropCycle) error
	DeleteCropCycle(id string) error
}

// CreateCropCycle creates a new crop cycle record.
func (r *CropCycleRepository) CreateCropCycle(cycle *models.CropCycle) error {
	cycle.Id = utils.Generate10DigitId()
	cycle.CreatedAt = time.Now()
	cycle.UpdatedAt = time.Now()

	if err := r.db.Create(cycle).Error; err != nil {
		return fmt.Errorf("failed to create crop cycle: %w", err)
	}
	return nil
}

// GetCropCyclesByFarmID retrieves crop cycles for a given farm ID.
func (r *CropCycleRepository) GetCropCyclesByFarmID(farmID string) ([]*models.CropCycle, error) {
	var cycles []*models.CropCycle
	if err := r.db.Where("farm_id = ?", farmID).Find(&cycles).Error; err != nil {
		return nil, fmt.Errorf("failed to get crop cycles for farm_id %s: %w", farmID, err)
	}
	return cycles, nil
}

// GetCropCyclesByCropID retrieves crop cycles for a given crop ID.
func (r *CropCycleRepository) GetCropCyclesByCropID(cropID string) ([]*models.CropCycle, error) {
	var cycles []*models.CropCycle
	if err := r.db.Where("crop_id = ?", cropID).Find(&cycles).Error; err != nil {
		return nil, fmt.Errorf("failed to get crop cycles for crop_id %s: %w", cropID, err)
	}
	return cycles, nil
}

// GetCropCycleByID retrieves a single crop cycle by its ID.
func (r *CropCycleRepository) GetCropCycleByID(id string) (*models.CropCycle, error) {
	var cycle models.CropCycle
	if err := r.db.Where("id = ?", id).First(&cycle).Error; err != nil {
		return nil, fmt.Errorf("failed to get crop cycle with id %s: %w", id, err)
	}
	return &cycle, nil
}

// UpdateCropCycle updates an existing crop cycle record.
func (r *CropCycleRepository) UpdateCropCycle(cycle *models.CropCycle) error {
	cycle.UpdatedAt = time.Now()
	if err := r.db.Save(cycle).Error; err != nil {
		return fmt.Errorf("failed to update crop cycle with id %s: %w", cycle.Id, err)
	}
	return nil
}

// DeleteCropCycle deletes a crop cycle record by its ID.
func (r *CropCycleRepository) DeleteCropCycle(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.CropCycle{}).Error; err != nil {
		return fmt.Errorf("failed to delete crop cycle with id %s: %w", id, err)
	}
	return nil
}

func (r *CropCycleRepository) GetCropCyclesByFarmAndCropID(farmID, cropID string) ([]*models.CropCycle, error) {
	var cycles []*models.CropCycle
	err := r.db.Where("farm_id = ? AND crop_id = ?", farmID, cropID).Find(&cycles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crop cycles by farm_id and crop_id: %w", err)
	}
	return cycles, nil
}
