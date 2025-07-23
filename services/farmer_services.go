package services

import (
	"context"
	"fmt"
	"strings"

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
	FetchFarmers(userId, farmerId, kisansathiUserId, fpoRegNo string) ([]models.Farmer, error)           // Updated to include user details
	FetchFarmersWithoutUserDetails(farmerId, kisansathiUserId, fpoRegNo string) ([]models.Farmer, error) // New method
	UpdateFarmer(farmerId string, req models.FarmerUpdateRequest) (*models.Farmer, error)
	FetchSubscribedFarmers(userId, kisansathiUserId string) ([]models.Farmer, error)
	SetSubscriptionStatus(farmerId string, subscribe bool) error

	AssignKisansathiToFarmers(kisansathiUserId string, farmerIds []string) error
}

// FarmerService handles business logic for farmers
type FarmerService struct {
	repo repositories.FarmerRepositoryInterface
	fpo  FPOServiceInterface
}

// NewFarmerService initializes a new FarmerService
func NewFarmerService(repo repositories.FarmerRepositoryInterface,
	fpoSvc FPOServiceInterface) *FarmerService {
	return &FarmerService{repo: repo, fpo: fpoSvc}
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

	var fpoRegNo *string
	if reg := strings.TrimSpace(req.FpoRegNo); reg != "" {
		if _, err := s.fpo.Get(reg); err != nil {
			return nil, nil, fmt.Errorf("unknown FPO reg-no: %s", reg)
		}
		fpoRegNo = &reg
	}

	f := &models.Farmer{
		UserId:           userId,
		KisansathiUserId: req.KisansathiUserId,
		FullName:         req.FullName,

		Gender:         req.Gender,
		SocialCategory: req.SocialCategory,
		FatherName:     req.FatherName,
		EquityShare:    req.EquityShare,
		TotalShare:     req.TotalShare,
		AreaType:       req.AreaType,

		FpoRegNo: fpoRegNo,

		IsActive: true,
		Type:     ftype,
	}

	created, err := s.repo.CreateFarmerEntry(f)
	if err != nil {
		return nil, nil, err
	}
	return created, userDetails, nil
}

func (s *FarmerService) ExistsForUser(userId string) (bool, error) {
	cnt, err := s.repo.CountByUserId(userId)
	return cnt > 0, err
}

func (s *FarmerService) FetchFarmers(userId, farmerId, kisansathiUserId, fpoRegNo string) ([]models.Farmer, error) {
	return s.repo.FetchFarmers(userId, farmerId, kisansathiUserId, fpoRegNo)
}

func (s *FarmerService) FetchFarmersWithoutUserDetails(farmerId, kisansathiUserId, fpoRegNo string) ([]models.Farmer, error) {
	return s.repo.FetchFarmers("", farmerId, kisansathiUserId, fpoRegNo)
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

// UpdateFarmer handles the business logic for updating a farmer's details.
func (s *FarmerService) UpdateFarmer(farmerId string, req models.FarmerUpdateRequest) (*models.Farmer, error) {
	// Create a map to hold only the fields that are being updated.
	updates := make(map[string]interface{})

	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.SocialCategory != nil {
		updates["social_category"] = *req.SocialCategory
	}
	if req.FatherName != nil {
		updates["father_name"] = *req.FatherName
	}
	if req.EquityShare != nil {
		updates["equity_share"] = *req.EquityShare
	}
	if req.TotalShare != nil {
		updates["total_share"] = *req.TotalShare
	}
	if req.AreaType != nil {
		updates["area_type"] = *req.AreaType
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsSubscribed != nil {
		updates["is_subscribed"] = *req.IsSubscribed
	}
	if req.Type != nil {
		ftype := entities.FarmerType(*req.Type)
		if !entities.FARMER_TYPES.IsValid(string(ftype)) {
			return nil, fmt.Errorf("invalid farmer type: %s", ftype)
		}
		updates["type"] = ftype
	}

	// Special handling for FpoRegNo to validate it
	if req.FpoRegNo != nil {
		if reg := strings.TrimSpace(*req.FpoRegNo); reg != "" {
			if _, err := s.fpo.Get(reg); err != nil {
				return nil, fmt.Errorf("unknown FPO reg-no: %s", reg)
			}
			updates["fpo_reg_no"] = &reg
		} else {
			updates["fpo_reg_no"] = nil // Allow unsetting the FPO
		}
	}

	// If there are no updates, we can return early.
	if len(updates) == 0 {
		// Or fetch and return the existing farmer record if preferred.
		// For now, we'll signal that no operation was performed.
		return nil, fmt.Errorf("no update fields provided")
	}

	return s.repo.UpdateFarmer(farmerId, updates)
}

// AssignKisansathiToFarmers assigns the KisansathiUserId to all specified farmers.
func (s *FarmerService) AssignKisansathiToFarmers(kisansathiUserId string, farmerIds []string) error {
	// Update farmers' KisansathiUserId
	if err := s.repo.UpdateKisansathiUserId(kisansathiUserId, farmerIds); err != nil {
		return fmt.Errorf("failed to assign Kisansathi UserId: %w", err)
	}
	return nil
}
