package services

import (
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"

	"github.com/google/uuid"
)

// FarmerServiceInterface defines service methods for farmer operations
type FarmerServiceInterface interface {
	CreateFarmer(userID string, req models.FarmerSignupRequest) (*models.Farmer, error)
}

// FarmerService handles business logic for farmers
type FarmerService struct {
	repo repositories.FarmerRepositoryInterface
}

// NewFarmerService initializes a new FarmerService
func NewFarmerService(repo repositories.FarmerRepositoryInterface) *FarmerService {
	return &FarmerService{repo: repo}
}

// CreateFarmer creates a new farmer entry
func (s *FarmerService) CreateFarmer(userID string, req models.FarmerSignupRequest) (*models.Farmer, error) {
	newFarmer := &models.Farmer{
		UserID:          uuid.MustParse(userID),
		KisanSathiUserID: req.KisansathiUserID,
		IsActive:        true,
	}

	return s.repo.CreateFarmerEntry(newFarmer)
}

