package repositories

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CropRepository handles CRUD operations for the Crop model.
type CropRepository struct {
	db *gorm.DB
}

// NewCropRepository creates a new instance of CropRepository.
func NewCropRepository(db *gorm.DB) *CropRepository {
	return &CropRepository{db: db}
}

// CropRepositoryInterface defines repository methods for Crop.
type CropRepositoryInterface interface {
	CreateCrop(crop *models.Crop) error
	GetAllCrops(name string, page, pageSize int) ([]*models.Crop, int64, error) // Updated signature
	GetCropByID(id string) (*models.Crop, error)
	UpdateCrop(crop *models.Crop) error
	DeleteCrop(id string) error
}

func (r *CropRepository) CreateCrop(crop *models.Crop) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// The BeforeCreate hook will generate the ID and set timestamps.

		// 1. We store the stages from the request in a temporary variable.
		// This prevents GORM's default behavior of trying to create new master Stage records.
		stagesToLink := crop.Stages
		// We MUST set crop.Stages to nil before creating the crop to avoid the cascade.
		crop.Stages = nil

		// 2. Create the Crop record itself, without any associations.
		if err := tx.Create(crop).Error; err != nil {
			return fmt.Errorf("failed to create crop: %w", err)
		}

		// 3. If there are stages to link from the request, create the association records now.
		if len(stagesToLink) > 0 {
			// Now that the crop exists and has an ID, we assign that ID to each association record.
			for i := range stagesToLink {
				stagesToLink[i].CropID = crop.Id
			}

			// Create the records in the 'crop_stages' join table.
			if err := tx.Create(&stagesToLink).Error; err != nil {
				return fmt.Errorf("failed to create crop stage associations: %w", err)
			}
		}

		// 4. Re-fetch the entire crop, now with its stages properly preloaded.
		// This ensures the data returned to the user is complete.
		return tx.Preload("Stages.Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("crop_stages.\"order\" ASC")
		}).First(crop, "id = ?", crop.Id).Error
	})
}

func (r *CropRepository) GetAllCrops(name string, page, pageSize int) ([]*models.Crop, int64, error) {
	var crops []*models.Crop
	var total int64

	query := r.db.Model(&models.Crop{})

	if name != "" {
		query = query.Where("crop_name ILIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count crops: %w", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Preload stages in the correct order for each crop.
	// Also preload the master stage info for each association.
	err := query.Preload("Stages.Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("crop_stages.\"order\" ASC")
	}).Find(&crops).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get crops: %w", err)
	}

	return crops, total, nil
}

// GetCropByID retrieves a single crop record by its ID.
func (r *CropRepository) GetCropByID(id string) (*models.Crop, error) {
	var crop models.Crop
	// Preload stages in the correct order.
	err := r.db.Preload("Stages.Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("crop_stages.\"order\" ASC")
	}).Where("id = ?", id).First(&crop).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get crop by id %s: %w", id, err)
	}
	return &crop, nil
}

// UpdateCrop updates an existing crop record. It automatically updates the UpdatedAt timestamp.
func (r *CropRepository) UpdateCrop(crop *models.Crop) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		crop.UpdatedAt = time.Now()

		// 1. Update the crop itself. We use Omit to prevent GORM from touching the association.
		if err := tx.Model(crop).Omit("Stages").Save(crop).Error; err != nil {
			return fmt.Errorf("failed to update crop with id %s: %w", crop.Id, err)
		}

		// 2. Delete existing stage associations for this crop.
		if err := tx.Where("crop_id = ?", crop.Id).Delete(&models.CropStage{}).Error; err != nil {
			return fmt.Errorf("failed to delete old stages for crop %s: %w", crop.Id, err)
		}

		// 3. Insert the new stage associations, if any.
		if len(crop.Stages) > 0 {
			for i := range crop.Stages {
				crop.Stages[i].CropID = crop.Id
			}
			if err := tx.Create(&crop.Stages).Error; err != nil {
				return fmt.Errorf("failed to create new stages for crop %s: %w", crop.Id, err)
			}
		}

		// Re-fetch the fully updated crop data to return to the user.
		return tx.Preload("Stages.Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("crop_stages.\"order\" ASC")
		}).First(crop, "id = ?", crop.Id).Error
	})
}

// DeleteCrop deletes a crop record by its ID.
func (r *CropRepository) DeleteCrop(id string) error {
	if err := r.db.Select(clause.Associations).Delete(&models.Crop{Base: models.Base{Id: id}}).Error; err != nil {
		return fmt.Errorf("failed to delete crop with id %s: %w", id, err)
	}
	return nil
}
