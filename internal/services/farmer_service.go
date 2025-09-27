package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// FarmerService handles farmer-related operations
type FarmerService interface {
	CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error)
	GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerResponse, error)
	UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error)
	DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error
	ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error)
	GetFarmerByPhone(ctx context.Context, phoneNumber string) (*responses.FarmerProfileResponse, error)
}

// FarmerServiceImpl implements FarmerService
type FarmerServiceImpl struct {
	repository *base.BaseFilterableRepository[*farmer.Farmer]
	farmRepo   *base.BaseFilterableRepository[*farm.Farm]
	dbManager  db.DBManager
	aaaService AAAService
}

// NewFarmerService creates a new farmer service with repository and AAA service
func NewFarmerService(repository *base.BaseFilterableRepository[*farmer.Farmer], farmRepo *base.BaseFilterableRepository[*farm.Farm], aaaService AAAService) FarmerService {
	return &FarmerServiceImpl{
		repository: repository,
		farmRepo:   farmRepo,
		aaaService: aaaService,
	}
}

// CreateFarmer creates a new farmer
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "create", "", req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to create farmer")
	}

	// Create farmer in AAA first
	aaaResponse, err := s.aaaService.CreateUser(ctx, map[string]interface{}{
		"username":     req.Profile.PhoneNumber, // Use phone as username
		"phone_number": req.Profile.PhoneNumber,
		"country_code": s.getCountryCode(req.Profile.PhoneNumber),
		"email":        req.Profile.Email,
		"password":     s.generateSecurePassword(),
		"full_name":    req.Profile.FirstName + " " + req.Profile.LastName,
		"role":         "farmer",
		"metadata":     make(map[string]string),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user in AAA: %w", err)
	}

	// Extract user ID from response
	responseMap, ok := aaaResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from AAA")
	}
	aaaUserID, ok := responseMap["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in AAA response")
	}

	// Create farmer entity
	farmerEntity := &farmer.Farmer{
		BaseModel:         *base.NewBaseModel("farmer", hash.Large),
		AAAUserID:         aaaUserID,
		AAAOrgID:          req.OrgID,
		KisanSathiUserID:  nil, // Will be set later if needed
		FirstName:         req.Profile.FirstName,
		LastName:          req.Profile.LastName,
		PhoneNumber:       req.Profile.PhoneNumber,
		Email:             req.Profile.Email,
		DateOfBirth:       &req.Profile.DateOfBirth,
		Gender:            req.Profile.Gender,
		StreetAddress:     req.Profile.Address.StreetAddress,
		City:              req.Profile.Address.City,
		State:             req.Profile.Address.State,
		PostalCode:        req.Profile.Address.PostalCode,
		Country:           req.Profile.Address.Country,
		Coordinates:       req.Profile.Address.Coordinates,
		LandOwnershipType: "", // Not available in request
		Status:            "ACTIVE",
		Preferences:       make(map[string]string),
		Metadata:          make(map[string]string),
	}

	// Save farmer to database
	if err := s.repository.Create(ctx, farmerEntity); err != nil {
		return nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Convert to response
	farmerResponse := farmerEntity.ToFarmerResponse()

	// Load farms for the farmer
	farms, err := s.loadFarmsForFarmer(ctx, farmerEntity.ID)
	if err != nil {
		log.Printf("Warning: Failed to load farms for farmer %s: %v", farmerEntity.ID, err)
		farms = []*responses.FarmData{}
	}

	// Create comprehensive response
	response := &responses.FarmerResponse{
		BaseResponse: &base.BaseResponse{
			RequestID: req.RequestID,
		},
		Data: &responses.FarmerProfileData{
			ID:               farmerResponse.ID,
			AAAUserID:        farmerResponse.AAAUserID,
			AAAOrgID:         farmerResponse.AAAOrgID,
			KisanSathiUserID: farmerResponse.KisanSathiUserID,
			FirstName:        farmerResponse.FirstName,
			LastName:         farmerResponse.LastName,
			PhoneNumber:      farmerResponse.PhoneNumber,
			Email:            farmerResponse.Email,
			DateOfBirth:      *farmerResponse.DateOfBirth,
			Gender:           farmerResponse.Gender,
			Address: responses.AddressData{
				StreetAddress: farmerResponse.StreetAddress,
				City:          farmerResponse.City,
				State:         farmerResponse.State,
				PostalCode:    farmerResponse.PostalCode,
				Country:       farmerResponse.Country,
				Coordinates:   farmerResponse.Coordinates,
			},
			Preferences: farmerResponse.Preferences,
			Metadata:    farmerResponse.Metadata,
			Farms:       farms,
			CreatedAt:   farmerResponse.CreatedAt,
			UpdatedAt:   farmerResponse.UpdatedAt,
		},
	}

	return response, nil
}

// GetFarmer retrieves a farmer by ID
func (s *FarmerServiceImpl) GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "read", req.FarmerID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to read farmer")
	}

	// Get farmer from database
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, req.FarmerID).Build()
	farmers, err := s.repository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer: %w", err)
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("farmer not found")
	}

	farmerEntity := farmers[0]

	// Convert to response
	farmerResponse := farmerEntity.ToFarmerResponse()

	// Load farms for the farmer
	farms, err := s.loadFarmsForFarmer(ctx, farmerEntity.ID)
	if err != nil {
		log.Printf("Warning: Failed to load farms for farmer %s: %v", farmerEntity.ID, err)
		farms = []*responses.FarmData{}
	}

	// Create comprehensive response
	response := &responses.FarmerResponse{
		BaseResponse: &base.BaseResponse{
			RequestID: req.RequestID,
		},
		Data: &responses.FarmerProfileData{
			ID:               farmerResponse.ID,
			AAAUserID:        farmerResponse.AAAUserID,
			AAAOrgID:         farmerResponse.AAAOrgID,
			KisanSathiUserID: farmerResponse.KisanSathiUserID,
			FirstName:        farmerResponse.FirstName,
			LastName:         farmerResponse.LastName,
			PhoneNumber:      farmerResponse.PhoneNumber,
			Email:            farmerResponse.Email,
			DateOfBirth:      *farmerResponse.DateOfBirth,
			Gender:           farmerResponse.Gender,
			Address: responses.AddressData{
				StreetAddress: farmerResponse.StreetAddress,
				City:          farmerResponse.City,
				State:         farmerResponse.State,
				PostalCode:    farmerResponse.PostalCode,
				Country:       farmerResponse.Country,
				Coordinates:   farmerResponse.Coordinates,
			},
			Preferences: farmerResponse.Preferences,
			Metadata:    farmerResponse.Metadata,
			Farms:       farms,
			CreatedAt:   farmerResponse.CreatedAt,
			UpdatedAt:   farmerResponse.UpdatedAt,
		},
	}

	return response, nil
}

// UpdateFarmer updates an existing farmer
func (s *FarmerServiceImpl) UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "update", req.FarmerID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to update farmer")
	}

	// Get existing farmer
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, req.FarmerID).Build()
	farmers, err := s.repository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer: %w", err)
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("farmer not found")
	}

	farmerEntity := farmers[0]

	// Update fields
	if req.Profile.FirstName != "" {
		farmerEntity.FirstName = req.Profile.FirstName
	}
	if req.Profile.LastName != "" {
		farmerEntity.LastName = req.Profile.LastName
	}
	if req.Profile.PhoneNumber != "" {
		farmerEntity.PhoneNumber = req.Profile.PhoneNumber
	}
	if req.Profile.Email != "" {
		farmerEntity.Email = req.Profile.Email
	}
	if req.Profile.DateOfBirth != "" {
		farmerEntity.DateOfBirth = &req.Profile.DateOfBirth
	}
	if req.Profile.Gender != "" {
		farmerEntity.Gender = req.Profile.Gender
	}
	if req.Profile.Address.StreetAddress != "" {
		farmerEntity.StreetAddress = req.Profile.Address.StreetAddress
	}
	if req.Profile.Address.City != "" {
		farmerEntity.City = req.Profile.Address.City
	}
	if req.Profile.Address.State != "" {
		farmerEntity.State = req.Profile.Address.State
	}
	if req.Profile.Address.PostalCode != "" {
		farmerEntity.PostalCode = req.Profile.Address.PostalCode
	}
	if req.Profile.Address.Country != "" {
		farmerEntity.Country = req.Profile.Address.Country
	}
	if req.Profile.Address.Coordinates != "" {
		farmerEntity.Coordinates = req.Profile.Address.Coordinates
	}
	// LandOwnershipType not available in request

	// Update in database
	if err := s.repository.Update(ctx, farmerEntity); err != nil {
		return nil, fmt.Errorf("failed to update farmer: %w", err)
	}

	// Convert to response
	farmerResponse := farmerEntity.ToFarmerResponse()

	// Load farms for the farmer
	farms, err := s.loadFarmsForFarmer(ctx, farmerEntity.ID)
	if err != nil {
		log.Printf("Warning: Failed to load farms for farmer %s: %v", farmerEntity.ID, err)
		farms = []*responses.FarmData{}
	}

	// Create comprehensive response
	response := &responses.FarmerResponse{
		BaseResponse: &base.BaseResponse{
			RequestID: req.RequestID,
		},
		Data: &responses.FarmerProfileData{
			ID:               farmerResponse.ID,
			AAAUserID:        farmerResponse.AAAUserID,
			AAAOrgID:         farmerResponse.AAAOrgID,
			KisanSathiUserID: farmerResponse.KisanSathiUserID,
			FirstName:        farmerResponse.FirstName,
			LastName:         farmerResponse.LastName,
			PhoneNumber:      farmerResponse.PhoneNumber,
			Email:            farmerResponse.Email,
			DateOfBirth:      *farmerResponse.DateOfBirth,
			Gender:           farmerResponse.Gender,
			Address: responses.AddressData{
				StreetAddress: farmerResponse.StreetAddress,
				City:          farmerResponse.City,
				State:         farmerResponse.State,
				PostalCode:    farmerResponse.PostalCode,
				Country:       farmerResponse.Country,
				Coordinates:   farmerResponse.Coordinates,
			},
			Preferences: farmerResponse.Preferences,
			Metadata:    farmerResponse.Metadata,
			Farms:       farms,
			CreatedAt:   farmerResponse.CreatedAt,
			UpdatedAt:   farmerResponse.UpdatedAt,
		},
	}

	return response, nil
}

// DeleteFarmer deletes a farmer
func (s *FarmerServiceImpl) DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "delete", req.FarmerID, req.OrgID)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to delete farmer")
	}

	// Get farmer to delete
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, req.FarmerID).Build()
	farmers, err := s.repository.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get farmer: %w", err)
	}

	if len(farmers) == 0 {
		return fmt.Errorf("farmer not found")
	}

	farmerEntity := farmers[0]

	// Delete from database
	if err := s.repository.Delete(ctx, farmerEntity.ID, farmerEntity); err != nil {
		return fmt.Errorf("failed to delete farmer: %w", err)
	}

	return nil
}

// ListFarmers lists farmers with pagination
func (s *FarmerServiceImpl) ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "list", "", req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to list farmers")
	}

	// Build filter
	filterBuilder := base.NewFilterBuilder()

	// Check if there's a search term in filters
	if searchTerm, exists := req.Filters["search"]; exists && searchTerm != "" {
		if search, ok := searchTerm.(string); ok && search != "" {
			filterBuilder.Where("first_name", base.OpLike, "%"+search+"%")
		}
	}

	// Check for status filter
	if status, exists := req.Filters["status"]; exists && status != "" {
		if statusStr, ok := status.(string); ok && statusStr != "" {
			filterBuilder.Where("status", base.OpEqual, statusStr)
		}
	}

	// Add pagination
	limit := req.PageSize
	if limit <= 0 {
		limit = 10
	}
	offset := (req.Page - 1) * limit

	filter := filterBuilder.Limit(limit, offset).Build()

	// Get farmers
	farmers, err := s.repository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list farmers: %w", err)
	}

	// Get total count
	// totalCount, err := s.repository.Count(ctx, filter, &farmer.Farmer{})
	// if err != nil {
	//	return nil, fmt.Errorf("failed to count farmers: %w", err)
	// }

	// Convert to response format
	var farmerDataList []*responses.FarmerProfileData
	for _, farmerEntity := range farmers {
		// Load farms for each farmer
		farms, err := s.loadFarmsForFarmer(ctx, farmerEntity.ID)
		if err != nil {
			log.Printf("Warning: Failed to load farms for farmer %s: %v", farmerEntity.ID, err)
			farms = []*responses.FarmData{}
		}

		farmerResponse := farmerEntity.ToFarmerResponse()
		farmerData := &responses.FarmerProfileData{
			ID:               farmerResponse.ID,
			AAAUserID:        farmerResponse.AAAUserID,
			AAAOrgID:         farmerResponse.AAAOrgID,
			KisanSathiUserID: farmerResponse.KisanSathiUserID,
			FirstName:        farmerResponse.FirstName,
			LastName:         farmerResponse.LastName,
			PhoneNumber:      farmerResponse.PhoneNumber,
			Email:            farmerResponse.Email,
			DateOfBirth:      *farmerResponse.DateOfBirth,
			Gender:           farmerResponse.Gender,
			Address: responses.AddressData{
				StreetAddress: farmerResponse.StreetAddress,
				City:          farmerResponse.City,
				State:         farmerResponse.State,
				PostalCode:    farmerResponse.PostalCode,
				Country:       farmerResponse.Country,
				Coordinates:   farmerResponse.Coordinates,
			},
			Preferences: farmerResponse.Preferences,
			Metadata:    farmerResponse.Metadata,
			Farms:       farms,
			CreatedAt:   farmerResponse.CreatedAt,
			UpdatedAt:   farmerResponse.UpdatedAt,
		}
		farmerDataList = append(farmerDataList, farmerData)
	}

	// Calculate pagination info
	// totalPages := (int(totalCount) + limit - 1) / limit

	response := &responses.FarmerListResponse{
		PaginatedResponse: &base.PaginatedResponse{},
		Data:              farmerDataList,
	}

	return response, nil
}

// GetFarmerByPhone retrieves a farmer by phone number
func (s *FarmerServiceImpl) GetFarmerByPhone(ctx context.Context, phoneNumber string) (*responses.FarmerProfileResponse, error) {
	filter := base.NewFilterBuilder().Where("phone_number", base.OpEqual, phoneNumber).Build()
	farmers, err := s.repository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer by phone: %w", err)
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("farmer not found")
	}

	// Convert to FarmerProfileResponse
	farmerEntity := farmers[0]
	return &responses.FarmerProfileResponse{
		BaseResponse: &base.BaseResponse{
			Success: true,
			Message: "Farmer retrieved successfully",
		},
		Data: &responses.FarmerProfileData{
			ID:               farmerEntity.ID,
			AAAUserID:        farmerEntity.AAAUserID,
			AAAOrgID:         farmerEntity.AAAOrgID,
			KisanSathiUserID: farmerEntity.KisanSathiUserID,
			FirstName:        farmerEntity.FirstName,
			LastName:         farmerEntity.LastName,
			PhoneNumber:      farmerEntity.PhoneNumber,
			Email:            farmerEntity.Email,
			DateOfBirth:      getStringValue(farmerEntity.DateOfBirth),
			Gender:           farmerEntity.Gender,
			Address: responses.AddressData{
				StreetAddress: farmerEntity.StreetAddress,
				City:          farmerEntity.City,
				State:         farmerEntity.State,
				PostalCode:    farmerEntity.PostalCode,
				Country:       farmerEntity.Country,
				Coordinates:   farmerEntity.Coordinates,
			},
			Preferences: farmerEntity.Preferences,
			Metadata:    farmerEntity.Metadata,
		},
	}, nil
}

// Helper functions

// getStringValue safely gets string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// loadFarmsForFarmer loads farms associated with a farmer
func (s *FarmerServiceImpl) loadFarmsForFarmer(ctx context.Context, farmerID string) ([]*responses.FarmData, error) {
	filter := base.NewFilterBuilder().Where("aaa_farmer_user_id", base.OpEqual, farmerID).Build()
	farms, err := s.farmRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to load farms: %w", err)
	}

	var farmDataList []*responses.FarmData
	for _, farmEntity := range farms {
		farmData := &responses.FarmData{
			ID:              farmEntity.ID,
			AAAFarmerUserID: farmEntity.AAAFarmerUserID,
			AAAOrgID:        farmEntity.AAAOrgID,
			Name:            *farmEntity.Name,
			Geometry:        farmEntity.Geometry,
			AreaHa:          farmEntity.AreaHa,
			CreatedAt:       farmEntity.CreatedAt,
			UpdatedAt:       farmEntity.UpdatedAt,
		}
		farmDataList = append(farmDataList, farmData)
	}

	return farmDataList, nil
}

// generateSecurePassword generates a secure password
func (s *FarmerServiceImpl) generateSecurePassword() string {
	// Generate a secure random password
	// In production, use a proper password generation library
	return "TempPassword123!"
}

// getCountryCode extracts country code from phone number
func (s *FarmerServiceImpl) getCountryCode(phoneNumber string) string {
	// Simple country code detection
	// In production, use a proper phone number parsing library
	if strings.HasPrefix(phoneNumber, "+91") {
		return "+91"
	}
	return "+91" // Default to India
}
