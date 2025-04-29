package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
)

// FarmerServiceInterface defines service methods for farmer operations
type FarmerServiceInterface interface {
	CreateFarmer(user_id string, req models.FarmerSignupRequest) (*models.Farmer, *pb.GetUserByIdResponse, error)
	// FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, *pb.GetUserByIdResponse, error) // Updated to include user details
	FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, error)          // Updated to include user details
	FetchFarmersWithoutUserDetails(farmer_id, kisansathi_user_id string) ([]models.Farmer, error) // New method
}

// FarmerService handles business logic for farmers
type FarmerService struct {
	Repo repositories.FarmerRepositoryInterface
}

// NewFarmerService initializes a new FarmerService
func NewFarmerService(repo repositories.FarmerRepositoryInterface) *FarmerService {
	return &FarmerService{
		Repo: repo,
	}
}

// CreateFarmer creates a new farmer entry
func (s *FarmerService) CreateFarmer(
	user_id string,
	req models.FarmerSignupRequest,
) (*models.Farmer, *pb.GetUserByIdResponse, error) {
	// Fetch user details using GetUserByIdClient
	user_details, err := GetUserByIdClient(context.Background(), user_id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user details: %w", err)
	}

	// Create farmer record
	new_farmer := &models.Farmer{
		UserId:           user_id,
		KisansathiUserId: req.KisansathiUserId,
		IsActive:         true,
	}

	created_farmer, err := s.Repo.CreateFarmerEntry(new_farmer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Return both user details and the created farmer
	return created_farmer, user_details, nil
}

// // FetchFarmersWithFilters fetches farmers with specific filters
// func (s *FarmerService) FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, *pb.GetUserByIdResponse, error) {
// 	// Fetch user details using GetUserByIdClient
// 	user_details, err := GetUserByIdClient(context.Background(), user_id)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to fetch user details: %w", err)
// 	}

// 	// Fetch farmers from the repository
// 	farmers, err := s.repo.FetchFarmers(user_id, farmer_id, kisansathi_user_id)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to fetch farmers: %w", err)
// 	}

// 	// Return both user details and the list of farmers
// 	return farmers, user_details, nil
// }

// // FetchFarmersWithoutUserDetails fetches farmers without user details
// func (s *FarmerService) FetchFarmersWithoutUserDetails(farmer_id, kisansathi_user_id string) ([]models.Farmer, error) {
// 	// Fetch farmers from the repository
// 	farmers, err := s.repo.FetchFarmers("", farmer_id, kisansathi_user_id)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch farmers: %w", err)
// 	}

// 	return farmers, nil
// }

func (s *FarmerService) FetchFarmers(user_id, farmer_id, kisansathi_user_id string) ([]models.Farmer, error) {
	return s.Repo.FetchFarmers(user_id, farmer_id, kisansathi_user_id)
}

func (s *FarmerService) FetchFarmersWithoutUserDetails(farmer_id, kisansathi_user_id string) ([]models.Farmer, error) {
	return s.Repo.FetchFarmers("", farmer_id, kisansathi_user_id)
}
