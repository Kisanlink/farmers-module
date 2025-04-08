package repositories

import (
	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
	FetchFarmers(userId, farmerId, kisansathiUserId string) ([]models.Farmer, error)
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

func (r *FarmerRepository) FetchFarmers(userId, farmerId, kisansathiUserId string) ([]models.Farmer, error) {
	var farmers []models.Farmer
	query := r.db.Model(&models.Farmer{})

	// Apply filters if query parameters are provided
	if userId != "" {
		query = query.Where("user_id = ?", userId)
	}
	if farmerId != "" {
		query = query.Where("id = ?", farmerId)
	}
	if kisansathiUserId != "" {
		query = query.Where("kisansathi_user_id = ?", kisansathiUserId)
	}

	// Execute the query
	if err := query.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}
