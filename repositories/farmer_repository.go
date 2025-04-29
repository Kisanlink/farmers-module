package repositories

import (
	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
	FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, error)
}

// FarmerRepository interacts with the database
type FarmerRepository struct {
	DB *gorm.DB
}

// NewFarmerRepository initializes a new repository
func NewFarmerRepository(db *gorm.DB) *FarmerRepository {
	return &FarmerRepository{DB: db}
}

// CreateFarmerEntry inserts a new farmer in the database
func (r *FarmerRepository) CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error) {
	if err := r.DB.Create(farmer).Error; err != nil {
		return nil, err
	}
	return farmer, nil
}

func (r *FarmerRepository) FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, error) {
	var farmers []models.Farmer
	query := r.DB.Model(&models.Farmer{})

	// Apply filters if query parameters are provided
	if user_id != "" {
		query = query.Where("user_id = ?", user_id)
	}
	if farmer_id != "" {
		query = query.Where("id = ?", farmer_id)
	}
	if kisansathi_user_id != "" {
		query = query.Where("kisansathi_user_id = ?", kisansathi_user_id)
	}

	// Execute the query
	if err := query.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}
