package repositories

import (
	"github.com/Kisanlink/farmers-module/models"

	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
}

// FarmerRepository interacts with the database
type FarmerRepository struct {
	db *gorm.DB
}

// NewFarmerRepository initializes a new repository
func NewFarmerRepository(db *gorm.DB) *FarmerRepository {
	return &FarmerRepository{db: db}
}



// CreateFarmerEntry inserts a new farmer in the database
func (r *FarmerRepository) CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error) {
	if err := r.db.Create(farmer).Error; err != nil {
		return nil, err
	}
	return farmer, nil
}
