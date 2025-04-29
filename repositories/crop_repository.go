package repositories

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

// CropRepository handles CRUD operations for the Crop model.
type CropRepository struct {
	DB *gorm.DB
}

// NewCropRepository creates a new instance of CropRepository.
func NewCropRepository(db *gorm.DB) *CropRepository {
	return &CropRepository{DB: db}
}

// CropRepositoryInterface defines repository methods for Crop.
type CropRepositoryInterface interface {
	CreateCrop(crop *models.Crop) error
	GetAllCrops(name string, page, page_size int) ([]*models.Crop, int64, error) // Updated signature
	GetCropById(id string) (*models.Crop, error)
	UpdateCrop(crop *models.Crop) error
	DeleteCrop(id string) error
}

// CreateCrop creates a new crop record.
func (r *CropRepository) CreateCrop(crop *models.Crop) error {
	crop.Id = utils.Generate10DigitId() // Generate a unique 10-digit ID.
	crop.CreatedAt = time.Now()
	crop.UpdatedAt = time.Now()

	if err := r.DB.Create(crop).Error; err != nil {
		return fmt.Errorf("failed to create crop: %w", err)
	}
	return nil
}

func (r *CropRepository) GetAllCrops(name string, page, page_size int) ([]*models.Crop, int64, error) {
	var crops []*models.Crop
	var total int64

	query := r.DB.Model(&models.Crop{})

	if name != "" {
		// PostgreSQL case-insensitive search using ILIKE
		query = query.Where("crop_name ILIKE ?", "%"+name+"%")
	}

	// Count total records first
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count crops: %w", err)
	}

	// Apply pagination
	if page > 0 && page_size > 0 {
		offset := (page - 1) * page_size
		query = query.Offset(offset).Limit(page_size)
	}

	if err := query.Find(&crops).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get crops: %w", err)
	}

	return crops, total, nil
}

// GetCropById retrieves a single crop record by its ID.
func (r *CropRepository) GetCropById(id string) (*models.Crop, error) {
	var crop models.Crop
	if err := r.DB.Where("id = ?", id).First(&crop).Error; err != nil {
		return nil, fmt.Errorf("failed to get crop by id %s: %w", id, err)
	}
	return &crop, nil
}

// UpdateCrop updates an existing crop record. It automatically updates the UpdatedAt timestamp.
func (r *CropRepository) UpdateCrop(crop *models.Crop) error {
	crop.UpdatedAt = time.Now()
	if err := r.DB.Save(crop).Error; err != nil {
		return fmt.Errorf("failed to update crop with id %s: %w", crop.Id, err)
	}
	return nil
}

// DeleteCrop deletes a crop record by its ID.
func (r *CropRepository) DeleteCrop(id string) error {
	if err := r.DB.
		Where("id = ?", id).
		Delete(&models.Crop{}).Error; err != nil {
		return fmt.Errorf("failed to delete crop with id %s: %w", id, err)
	}
	return nil
}
