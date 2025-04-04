package repositories

import (
	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
	FetchFarmers(filter models.FarmerFilter ) ([]models.Farmer, error)
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

func (r *FarmerRepository) FetchFarmers(filter models.FarmerFilter) ([]models.Farmer, error) {
	var farmers []models.Farmer
	query := r.db.Model(&models.Farmer{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.UserName != nil {
		query = query.Where("user_name LIKE ?", "%"+*filter.UserName+"%")
	}
	if filter.Email != nil {
		query = query.Where("email = ?", *filter.Email)
	}
	if filter.CountryCode != nil {
		query = query.Where("country_code = ?", *filter.CountryCode)
	}
	if filter.MobileNumber != nil {
		query = query.Where("mobile_number = ?", *filter.MobileNumber)
	}
	if filter.AadhaarNumber != nil {
		query = query.Where("aadhaar_number = ?", *filter.AadhaarNumber)
	}
	if filter.KisansathiUserID != nil {
		query = query.Where("kisansathi_user_id = ?", *filter.KisansathiUserID)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filter.CreatedBefore)
	}

	if err := query.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}