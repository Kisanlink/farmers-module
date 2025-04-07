package repositories

import (
	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
	FetchFarmers(userID, farmerID, kisansathiUserID string) ([]models.Farmer, error)
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


func (r *FarmerRepository) FetchFarmers(userID, farmerID, kisansathiUserID string) ([]models.Farmer, error) {
	var farmers []models.Farmer
	query := r.db.Model(&models.Farmer{})

	// Apply filters if query parameters are provided
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if farmerID != "" {
		query = query.Where("id = ?", farmerID)
	}
	if kisansathiUserID != "" {
		query = query.Where("kisansathi_user_id = ?", kisansathiUserID)
	}

	// Execute the query
	if err := query.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}
