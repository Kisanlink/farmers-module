package services

import (

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

// FarmerServiceInterface defines service methods for farmer operations
type FarmerServiceInterface interface {
	CreateFarmer( userID string, req models.FarmerSignupRequest) (*models.Farmer, error)
	FetchFarmers() ([]models.Farmer, error)
}

// FarmerService handles business logic for farmers
type FarmerService struct {
	repo repositories.FarmerRepositoryInterface
}

// NewFarmerService initializes a new FarmerService
func NewFarmerService(repo repositories.FarmerRepositoryInterface) *FarmerService {
	return &FarmerService{
		repo: repo,
	}
}

// CreateFarmer creates a new farmer entry
func (s *FarmerService) CreateFarmer(
	userID string,
	req models.FarmerSignupRequest,
) (*models.Farmer, error) {
	// Create farmer record
	newFarmer := &models.Farmer{
		UserID:           userID,
		KisansathiUserID: req.KisansathiUserID,
		IsActive:         true,
	}

	return s.repo.CreateFarmerEntry(newFarmer)
}

// FetchFarmers fetches all farmers from the database
func (s *FarmerService) FetchFarmers() ([]models.Farmer, error) {
	return s.repo.FetchFarmers()
}