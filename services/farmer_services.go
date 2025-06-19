package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/kisanlink/protobuf/pb-aaa"
)

// FarmerServiceInterface defines service methods for farmer operations
type FarmerServiceInterface interface {
	CreateFarmer(userId string, req models.FarmerSignupRequest) (*models.Farmer, *pb.GetUserByIdResponse, error)
	ExistsForUser(userId string) (bool, error) // Checks if a farmer exists for the given user ID
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
		return nil, nil, fmt.Errorf("AAA lookup: %w", err)
	}

	ftype := entities.FARMER_TYPES.OTHER
	if req.Type != "" {
		ftype = entities.FarmerType(req.Type)
	}

	f := &models.Farmer{
		UserId:           userId,
		KisansathiUserId: req.KisansathiUserId,

		Gender:         req.Gender,
		SocialCategory: req.SocialCategory,
		FatherName:     req.FatherName,
		EquityShare:    req.EquityShare,
		TotalShare:     req.TotalShare,
		AreaType:       req.AreaType,

		IsFPO:    req.IsFPO,
		State:    nullable(req.State),
		District: nullable(req.District),
		Block:    nullable(req.Block),
		IaName:   nullable(req.IaName),
		CbbName:  nullable(req.CbbName),
		FpoName:  nullable(req.FpoName),
		FpoRegNo: nullable(req.FpoRegNo),

		IsActive: true,
		Type:     ftype,
	}

	created, err := s.repo.CreateFarmerEntry(f)
	if err != nil {
		return nil, nil, err
	}
	return created, userDetails, nil
}

func nullable(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func (s *FarmerService) ExistsForUser(userId string) (bool, error) {
	cnt, err := s.repo.CountByUserId(userId)
	return cnt > 0, err
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
