package repositories

import (
	"context"
	"errors"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmRepositoryInterface defines database operations for farms
type FarmRepositoryInterface interface {
	CreateFarm( farm *models.Farm) error
	GetFarmByID(ctx context.Context, farmID string) (*models.Farm, error)
	GetFarmsByFarmerID(ctx context.Context, farmerID string) ([]models.Farm, error)
	UpdateFarm(ctx context.Context, farm *models.Farm) error
	DeleteFarm(ctx context.Context, farmID string) error
}

// FarmRepository handles actual database operations
type FarmRepository struct {
	db *gorm.DB
}

// NewFarmRepository initializes a FarmRepository
func NewFarmRepository(db *gorm.DB) *FarmRepository {
	return &FarmRepository{db: db}
}

// CreateFarm inserts a new farm into the database
func (r *FarmRepository) CreateFarm( farm *models.Farm) error {
	result := r.db.Create(farm)
	if result.Error != nil {
		log.Printf("Failed to create farm: %v", result.Error)
		return result.Error
	}
	return nil
}

// GetFarmByID fetches a farm by ID
func (r *FarmRepository) GetFarmByID(ctx context.Context, farmID string) (*models.Farm, error) {
	var farm models.Farm
	result := r.db.WithContext(ctx).First(&farm, "id = ?", farmID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("farm not found")
		}
		log.Printf("Failed to fetch farm: %v", result.Error)
		return nil, result.Error
	}
	return &farm, nil
}

// GetFarmsByFarmerID fetches all farms belonging to a farmer
func (r *FarmRepository) GetFarmsByFarmerID(ctx context.Context, farmerID string) ([]models.Farm, error) {
	var farms []models.Farm
	result := r.db.WithContext(ctx).Where("farmer_id = ?", farmerID).Find(&farms)
	if result.Error != nil {
		log.Printf("Failed to fetch farms: %v", result.Error)
		return nil, result.Error
	}
	return farms, nil
}

// UpdateFarm updates an existing farm entry
func (r *FarmRepository) UpdateFarm(ctx context.Context, farm *models.Farm) error {
	result := r.db.WithContext(ctx).Save(farm)
	if result.Error != nil {
		log.Printf("Failed to update farm: %v", result.Error)
		return result.Error
	}
	return nil
}

// DeleteFarm deletes a farm by ID
func (r *FarmRepository) DeleteFarm(ctx context.Context, farmID string) error {
	result := r.db.WithContext(ctx).Delete(&models.Farm{}, "id = ?", farmID)
	if result.Error != nil {
		log.Printf("Failed to delete farm: %v", result.Error)
		return result.Error
	}
	return nil
}
