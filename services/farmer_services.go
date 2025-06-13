package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
)

// FarmerServiceInterface defines service methods for farmer operations
type FarmerServiceInterface interface {
	CreateFarmer(userId string, req models.FarmerSignupRequest) (*models.Farmer, *pb.GetUserByIdResponse, error)
	// FetchFarmers(userId, farmerId, kisansathiUserId string) ([]models.Farmer, *pb.GetUserByIdResponse, error) // Updated to include user details
	FetchFarmers(userId, farmerId, kisansathiUserId string) ([]models.Farmer, error)           // Updated to include user details
	FetchFarmersWithoutUserDetails(farmerId, kisansathiUserId string) ([]models.Farmer, error) // New method

	FetchSubscribedFarmers(userId, kisansathiUserId string) ([]models.Farmer, error)
	SetSubscriptionStatus(farmerId string, subscribe bool) error
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

func (s *FarmerService) CreateFarmer(
	userId string,
	req models.FarmerSignupRequest,
) (*models.Farmer, *pb.GetUserByIdResponse, error) {

	userDetails, err := GetUserByIdClient(context.Background(), userId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user details: %w", err)
	}

	// Set farmerType to default 'OWNER' if req.Type is empty
	var farmerType entities.FarmerType
	if req.Type == "" {
		farmerType = entities.FARMER_TYPES.OTHER
	} else {
		farmerType = entities.FarmerType(req.Type)
	}

	newFarmer := &models.Farmer{
		UserId:           userId,
		KisansathiUserId: req.KisansathiUserId,
		IsActive:         true,
		Type:             farmerType,
	}

	createdFarmer, err := s.repo.CreateFarmerEntry(newFarmer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	return createdFarmer, userDetails, nil
}

func (s *FarmerService) FetchFarmers(userId, farmerId, kisansathiUserId string) ([]models.Farmer, error) {
	return s.repo.FetchFarmers(userId, farmerId, kisansathiUserId)
}

func (s *FarmerService) FetchFarmersWithoutUserDetails(farmerId, kisansathiUserId string) ([]models.Farmer, error) {
	return s.repo.FetchFarmers("", farmerId, kisansathiUserId)
}

func (s *FarmerService) FetchSubscribedFarmers(
	userId, kisansathiUserId string,
) ([]models.Farmer, error) {
	return s.repo.FetchSubscribedFarmers(userId, kisansathiUserId)
}

func (s *FarmerService) SetSubscriptionStatus(
	farmerId string,
	subscribe bool,
) error {
	return s.repo.SetSubscriptionStatus(farmerId, subscribe)
}
