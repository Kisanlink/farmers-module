package services

import (

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

// FarmServiceInterface defines service methods for farm operations
type FarmServiceInterface interface {
	CreateFarm( req models.FarmRequest) (*models.Farm, error)
}

// FarmService handles business logic for farms
type FarmService struct {
	repo repositories.FarmRepositoryInterface
}

// NewFarmService initializes a new FarmService
func NewFarmService(repo repositories.FarmRepositoryInterface) *FarmService {
	return &FarmService{repo: repo}
}
// CreateFarm creates a new farm entry
func (s *FarmService) CreateFarm(req models.FarmRequest) (*models.Farm, error) {
	newFarm := &models.Farm{
		FarmerID: req.FarmerID,
		Verified: req.Verified,
		Location: req.Location,
		Area:     req.Area,
		Locality: req.Locality,
	}

	// Insert farm record into database via repository
	err := s.repo.CreateFarm(newFarm)
	if err != nil {
		return nil, err
	}
	return newFarm, nil
}
