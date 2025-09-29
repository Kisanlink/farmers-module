package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
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
	logger     *zap.Logger
}

// NewFarmerService creates a new farmer service with repository and AAA service
func NewFarmerService(repository *base.BaseFilterableRepository[*farmer.Farmer], farmRepo *base.BaseFilterableRepository[*farm.Farm], aaaService AAAService) FarmerService {
	logger, _ := zap.NewProduction() // Use production logger configuration
	return &FarmerServiceImpl{
		repository: repository,
		farmRepo:   farmRepo,
		aaaService: aaaService,
		logger:     logger,
	}
}

// CreateFarmer creates a new farmer
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error) {
	s.logger.Info("Creating new farmer",
		zap.String("user_id", req.UserID),
		zap.String("org_id", req.OrgID),
		zap.String("phone", req.Profile.PhoneNumber))

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "farmer", "create", "", req.OrgID)
	if err != nil {
		s.logger.Error("Permission check failed",
			zap.String("user_id", req.UserID),
			zap.Error(err))
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		s.logger.Warn("Insufficient permissions to create farmer",
			zap.String("user_id", req.UserID),
			zap.String("org_id", req.OrgID))
		return nil, fmt.Errorf("insufficient permissions to create farmer")
	}

	// Create farmer in AAA first
	securePassword := s.generateSecurePassword()
	s.logger.Info("Creating AAA user for farmer",
		zap.String("phone", req.Profile.PhoneNumber),
		zap.String("email", req.Profile.Email))

	aaaResponse, err := s.aaaService.CreateUser(ctx, map[string]interface{}{
		"username":     req.Profile.PhoneNumber, // Use phone as username
		"phone_number": req.Profile.PhoneNumber,
		"country_code": s.getCountryCode(req.Profile.PhoneNumber),
		"email":        req.Profile.Email,
		"password":     securePassword,
		"full_name":    req.Profile.FirstName + " " + req.Profile.LastName,
		"role":         "farmer",
		"metadata":     make(map[string]string),
	})
	if err != nil {
		s.logger.Error("Failed to create AAA user",
			zap.String("phone", req.Profile.PhoneNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create user in AAA: %w", err)
	}

	s.logger.Info("Successfully created AAA user",
		zap.String("phone", req.Profile.PhoneNumber))

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

// generateSecurePassword generates a cryptographically secure password for new farmers
func (s *FarmerServiceImpl) generateSecurePassword() string {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
		symbols   = "!@#$%^&*_+-="
		allChars  = lowercase + uppercase + digits + symbols
	)

	// Generate a 12-character password with at least one character from each category
	password := make([]byte, 12)

	// Add one lowercase letter
	if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowercase)))); err == nil {
		password[0] = lowercase[randInt.Int64()]
	} else {
		// Fallback to first character if random generation fails
		password[0] = lowercase[0]
	}

	// Add one uppercase letter
	if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(uppercase)))); err == nil {
		password[1] = uppercase[randInt.Int64()]
	} else {
		password[1] = uppercase[0]
	}

	// Add one digit
	if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits)))); err == nil {
		password[2] = digits[randInt.Int64()]
	} else {
		password[2] = digits[0]
	}

	// Add one symbol
	if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(symbols)))); err == nil {
		password[3] = symbols[randInt.Int64()]
	} else {
		password[3] = symbols[0]
	}

	// Fill remaining positions with random characters from all categories
	for i := 4; i < len(password); i++ {
		if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars)))); err == nil {
			password[i] = allChars[randInt.Int64()]
		} else {
			// Fallback to a default character if random generation fails
			password[i] = allChars[i%len(allChars)]
		}
	}

	// Shuffle the password using Fisher-Yates algorithm to randomize positions
	for i := len(password) - 1; i > 0; i-- {
		if randInt, err := rand.Int(rand.Reader, big.NewInt(int64(i+1))); err == nil {
			j := randInt.Int64()
			password[i], password[j] = password[j], password[i]
		}
	}

	return string(password)
}

// getCountryCode extracts country code from phone number
func (s *FarmerServiceImpl) getCountryCode(phoneNumber string) string {
	// Remove any whitespace and normalize the phone number
	phoneNumber = strings.TrimSpace(phoneNumber)

	// Handle different phone number formats
	if strings.HasPrefix(phoneNumber, "+") {
		// International format: +91XXXXXXXXXX
		if strings.HasPrefix(phoneNumber, "+91") && len(phoneNumber) == 13 {
			return "+91"
		}
		if strings.HasPrefix(phoneNumber, "+1") && len(phoneNumber) == 12 {
			return "+1"
		}
		if strings.HasPrefix(phoneNumber, "+44") && len(phoneNumber) == 13 {
			return "+44"
		}
		// Add more country codes as needed
	} else if strings.HasPrefix(phoneNumber, "91") && len(phoneNumber) == 12 {
		// Indian format without +: 91XXXXXXXXXX
		return "+91"
	} else if strings.HasPrefix(phoneNumber, "0") && len(phoneNumber) == 11 {
		// Indian format with leading 0: 0XXXXXXXXXX
		return "+91"
	} else if len(phoneNumber) == 10 {
		// Indian format without country code: XXXXXXXXXX
		return "+91"
	}

	// Default to India for unrecognized formats
	// In production, consider logging unknown formats for analysis
	return "+91"
}
