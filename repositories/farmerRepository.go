package repositories

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// FarmerRepositoryInterface defines repository methods for farmers
type FarmerRepositoryInterface interface {
	CreateFarmerEntry(farmer *models.Farmer) (*models.Farmer, error)
	FetchFarmers(userId, farmerId, kisansathiUserId, fpoRegNo string) ([]models.Farmer, error)

	FetchSubscribedFarmers(userId, kisansathiUserId string) ([]models.Farmer, error)
	SetSubscriptionStatus(farmerId string, subscribe bool) error
	CountByUserId(id string) (int64, error)

	UpdateKisansathiUserId(kisansathiUserId string, farmerIds []string) error
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

func (r *FarmerRepository) FetchFarmers(userId, farmerId, kisansathiUserId string, fpoRegNo string) ([]models.Farmer, error) {
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
	if fpoRegNo != "" {
		query = query.Where("fpo_reg_no = ?", fpoRegNo)
	}

	// Execute the query
	if err := query.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}

// FetchSubscribedFarmers returns all farmers where is_subscribed = true,
// with optional filters on user_id or kisansathi_user_id
func (r *FarmerRepository) FetchSubscribedFarmers(
	userId, kisansathiUserId string,
) ([]models.Farmer, error) {
	var farmers []models.Farmer
	q := r.db.Model(&models.Farmer{}).
		Where("is_subscribed = ?", true)

	if userId != "" {
		q = q.Where("user_id = ?", userId)
	}
	if kisansathiUserId != "" {
		q = q.Where("kisansathi_user_id = ?", kisansathiUserId)
	}

	if err := q.Find(&farmers).Error; err != nil {
		return nil, err
	}
	return farmers, nil
}

// SetSubscriptionStatus toggles the is_subscribed column for a single farmer
func (r *FarmerRepository) SetSubscriptionStatus(
	farmerID string,
	subscribe bool,
) error {
	return r.db.
		Model(&models.Farmer{}).
		Where("id = ?", farmerID).
		Update("is_subscribed", subscribe).
		Error
}

func (r *FarmerRepository) SubscribeFarmer(farmerID string) error {
	return r.SetSubscriptionStatus(farmerID, true)
}

func (r *FarmerRepository) UnsubscribeFarmer(farmerID string) error {
	return r.SetSubscriptionStatus(farmerID, false)
}

func (r *FarmerRepository) CountByUserId(id string) (int64, error) {
	var n int64
	err := r.db.Model(&models.Farmer{}).
		Where("user_id = ?", id).
		Count(&n).Error
	return n, err
}

// UpdateKisansathiUserId updates the KisansathiUserId for the specified farmers.
func (r *FarmerRepository) UpdateKisansathiUserId(kisansathiUserId string, farmerIds []string) error {
	if len(farmerIds) == 0 {
		return nil
	}

	// Update the KisansathiUserId for the list of farmers
	if err := r.db.Model(&models.Farmer{}).
		Where("id IN (?)", farmerIds).
		Update("kisansathi_user_id", kisansathiUserId).Error; err != nil {
		return fmt.Errorf("failed to update KisansathiUserId: %w", err)
	}
	return nil
}
