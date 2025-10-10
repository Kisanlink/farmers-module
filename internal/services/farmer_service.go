package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/internal/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmerService handles farmer-related operations
type FarmerService interface {
	CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error)
	GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerProfileResponse, error)
	UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error)
	DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error
	ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error)
}

// FarmerServiceImpl implements FarmerService
type FarmerServiceImpl struct {
	repository  *farmer.FarmerRepository
	aaaService  AAAService
	passwordGen *utils.PasswordGenerator
}

// NewFarmerService creates a new farmer service with repository and AAA service
func NewFarmerService(repository *farmer.FarmerRepository, aaaService AAAService) FarmerService {
	return &FarmerServiceImpl{
		repository:  repository,
		aaaService:  aaaService,
		passwordGen: utils.NewPasswordGenerator(),
	}
}

// CreateFarmer creates a new farmer
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error) {
	// Check if farmer already exists using filter
	existingFilter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, req.AAAUserID).
		Where("aaa_org_id", base.OpEqual, req.AAAOrgID).
		Build()

	existing, err := s.repository.FindOne(ctx, existingFilter)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("farmer already exists")
	}

	// Check if user exists in AAA by mobile number
	var aaaUserID string
	var aaaOrgID string

	if req.Profile.PhoneNumber != "" {
		// Business Rule 5.1: RegisterFarmer idempotency - return existing farmer if phone exists
		// Try to find existing user by mobile number
		aaaUser, err := s.aaaService.GetUserByMobile(ctx, req.Profile.PhoneNumber)
		if err != nil {
			// User doesn't exist, create new user in AAA
			log.Printf("User not found in AAA, creating new user with mobile: %s", req.Profile.PhoneNumber)

			// Generate secure password
			password, err := s.passwordGen.GenerateSecurePassword()
			if err != nil {
				return nil, fmt.Errorf("failed to generate secure password: %w", err)
			}

			// Log password generation (remove in production)
			log.Printf("Generated secure password for farmer %s", req.Profile.PhoneNumber)

			// Create user in AAA
			createUserReq := map[string]interface{}{
				"username":       fmt.Sprintf("farmer_%s", req.Profile.PhoneNumber),
				"mobile_number":  req.Profile.PhoneNumber,
				"password":       password,
				"country_code":   "+91", // TODO: Make configurable
				"aadhaar_number": "",    // TODO: Add to request if available
			}

			aaaUser, err = s.aaaService.CreateUser(ctx, createUserReq)
			if err != nil {
				return nil, fmt.Errorf("failed to create user in AAA: %w", err)
			}
		} else {
			// User exists in AAA - check if farmer profile exists locally
			// Extract AAA user ID from existing user
			if userMap, ok := aaaUser.(map[string]interface{}); ok {
				if id, exists := userMap["id"]; exists {
					aaaUserID = fmt.Sprintf("%v", id)
				}
			}

			// Try to get existing farmer profile
			if aaaUserID != "" {
				existingFarmerFilter := base.NewFilterBuilder().
					Where("aaa_user_id", base.OpEqual, aaaUserID).
					Build()
				existingFarmer, err := s.repository.FindOne(ctx, existingFarmerFilter)
				if err == nil && existingFarmer != nil {
					// Farmer profile exists - return it (idempotent operation)
					log.Printf("Farmer already registered with phone %s, returning existing profile", req.Profile.PhoneNumber)

					// Convert to response format
					var addressData responses.AddressData
					if existingFarmer.Address != nil {
						addressData = responses.AddressData{
							StreetAddress: existingFarmer.Address.StreetAddress,
							City:          existingFarmer.Address.City,
							State:         existingFarmer.Address.State,
							PostalCode:    existingFarmer.Address.PostalCode,
							Country:       existingFarmer.Address.Country,
							Coordinates:   existingFarmer.Address.Coordinates,
						}
					}

					farmerProfileData := &responses.FarmerProfileData{
						ID:               existingFarmer.GetID(),
						AAAUserID:        existingFarmer.AAAUserID,
						AAAOrgID:         existingFarmer.AAAOrgID,
						KisanSathiUserID: existingFarmer.KisanSathiUserID,
						FirstName:        existingFarmer.FirstName,
						LastName:         existingFarmer.LastName,
						PhoneNumber:      existingFarmer.PhoneNumber,
						Email:            existingFarmer.Email,
						DateOfBirth:      existingFarmer.DateOfBirth,
						Gender:           existingFarmer.Gender,
						Address:          addressData,
						Preferences:      existingFarmer.Preferences,
						Metadata:         existingFarmer.Metadata,
						Farms:            []*responses.FarmData{},
						CreatedAt:        existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
						UpdatedAt:        existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
					}

					response := responses.NewFarmerResponse(farmerProfileData, "Farmer already registered")
					return &response, nil
				}
				log.Printf("AAA user exists but no local farmer profile found, proceeding with registration")
			}
		}

		// Extract AAA user ID from response
		if userMap, ok := aaaUser.(map[string]interface{}); ok {
			if id, exists := userMap["id"]; exists {
				aaaUserID = fmt.Sprintf("%v", id)
			}
		}

		if aaaUserID == "" {
			return nil, fmt.Errorf("failed to get AAA user ID from response")
		}

		// For now, use the same org ID as provided in request.
		// Verify organization exists in AAA (best-effort; non-fatal if AAA unsupported)
		aaaOrgID = req.AAAOrgID
		if aaaOrgID != "" {
			if _, err := s.aaaService.GetOrganization(ctx, aaaOrgID); err != nil {
				log.Printf("Warning: could not verify AAA organization %s: %v", aaaOrgID, err)
			}
		}
	} else {
		// Use provided AAA user ID and org ID
		aaaUserID = req.AAAUserID
		aaaOrgID = req.AAAOrgID
	}

	// Business Rule 2.2: Validate KisanSathi if provided
	// If validation fails, set to NULL and return warning (don't fail registration)
	var kisanSathiWarning string
	var validatedKisanSathiID *string
	if req.KisanSathiUserID != nil && *req.KisanSathiUserID != "" {
		// Validate KisanSathi user exists and has the kisansathi role
		_, err := s.aaaService.GetUser(ctx, *req.KisanSathiUserID)
		if err != nil {
			log.Printf("Warning: KisanSathi validation failed for user %s: %v", *req.KisanSathiUserID, err)
			kisanSathiWarning = fmt.Sprintf("KisanSathi validation failed: user not found - %v", err)
			validatedKisanSathiID = nil
		} else {
			// Check if user has kisansathi role
			hasRole, err := s.aaaService.CheckUserRole(ctx, *req.KisanSathiUserID, "kisansathi")
			if err != nil {
				log.Printf("Warning: Failed to check KisanSathi role for user %s: %v", *req.KisanSathiUserID, err)
				kisanSathiWarning = fmt.Sprintf("KisanSathi role check failed: %v", err)
				validatedKisanSathiID = nil
			} else if !hasRole {
				log.Printf("Warning: User %s does not have kisansathi role", *req.KisanSathiUserID)
				kisanSathiWarning = "KisanSathi validation failed: user does not have kisansathi role"
				validatedKisanSathiID = nil
			} else {
				// Validation successful
				validatedKisanSathiID = req.KisanSathiUserID
				log.Printf("KisanSathi %s validated successfully", *req.KisanSathiUserID)
			}
		}

		// Log the warning if validation failed
		if kisanSathiWarning != "" {
			log.Printf("Farmer registration proceeding without KisanSathi assignment. User: %s, Warning: %s", aaaUserID, kisanSathiWarning)
		}
	} else {
		validatedKisanSathiID = nil
	}

	// Create new farmer profile
	farmerProfile := &entities.FarmerProfile{
		BaseModel:        *base.NewBaseModel("farmer_profile", hash.Medium),
		AAAUserID:        aaaUserID,
		AAAOrgID:         aaaOrgID,
		KisanSathiUserID: validatedKisanSathiID,
		FirstName:        req.Profile.FirstName,
		LastName:         req.Profile.LastName,
		PhoneNumber:      req.Profile.PhoneNumber,
		Email:            req.Profile.Email,
		DateOfBirth:      req.Profile.DateOfBirth,
		Gender:           req.Profile.Gender,
		Address: &entities.Address{
			StreetAddress: req.Profile.Address.StreetAddress,
			City:          req.Profile.Address.City,
			State:         req.Profile.Address.State,
			PostalCode:    req.Profile.Address.PostalCode,
			Country:       req.Profile.Address.Country,
			Coordinates:   req.Profile.Address.Coordinates,
		},
		Preferences: req.Profile.Preferences,
		Metadata:    req.Profile.Metadata,
		Status:      "ACTIVE",
	}
	farmerProfile.SetCreatedBy(req.UserID)

	// Save to repository
	if err := s.repository.Create(ctx, farmerProfile); err != nil {
		return nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Convert to response format
	// Handle address data safely
	var addressData responses.AddressData
	if farmerProfile.Address != nil {
		addressData = responses.AddressData{
			StreetAddress: farmerProfile.Address.StreetAddress,
			City:          farmerProfile.Address.City,
			State:         farmerProfile.Address.State,
			PostalCode:    farmerProfile.Address.PostalCode,
			Country:       farmerProfile.Address.Country,
			Coordinates:   farmerProfile.Address.Coordinates,
		}
	}

	farmerProfileData := &responses.FarmerProfileData{
		ID:               farmerProfile.GetID(),
		AAAUserID:        farmerProfile.AAAUserID,
		AAAOrgID:         farmerProfile.AAAOrgID,
		KisanSathiUserID: farmerProfile.KisanSathiUserID,
		FirstName:        farmerProfile.FirstName,
		LastName:         farmerProfile.LastName,
		PhoneNumber:      farmerProfile.PhoneNumber,
		Email:            farmerProfile.Email,
		DateOfBirth:      farmerProfile.DateOfBirth,
		Gender:           farmerProfile.Gender,
		Address:          addressData,
		Preferences:      farmerProfile.Preferences,
		Metadata:         farmerProfile.Metadata,
		Farms:            []*responses.FarmData{},
		CreatedAt:        farmerProfile.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        farmerProfile.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Prepare response message with warning if KisanSathi validation failed
	message := "Farmer created successfully"
	if kisanSathiWarning != "" {
		message = fmt.Sprintf("Farmer created successfully (warning: %s)", kisanSathiWarning)
	}

	response := responses.NewFarmerResponse(farmerProfileData, message)
	return &response, nil
}

// GetFarmer retrieves a farmer by ID, user_id, or user_id+org_id
func (s *FarmerServiceImpl) GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerProfileResponse, error) {
	var filter *base.Filter

	// Priority: farmer_id > aaa_user_id > aaa_user_id + aaa_org_id
	if req.FarmerID != "" {
		// Lookup by primary key (farmer_id)
		filter = base.NewFilterBuilder().
			Where("id", base.OpEqual, req.FarmerID).
			Build()
	} else if req.AAAUserID != "" {
		// Lookup by user_id, optionally filtered by org_id
		filterBuilder := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID)

		if req.AAAOrgID != "" {
			filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
		}

		filter = filterBuilder.Build()
	} else {
		return nil, fmt.Errorf("either farmer_id or aaa_user_id must be provided")
	}

	farmerProfile, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Convert to response format
	// Handle address data safely
	var addressData responses.AddressData
	if farmerProfile.Address != nil {
		addressData = responses.AddressData{
			StreetAddress: farmerProfile.Address.StreetAddress,
			City:          farmerProfile.Address.City,
			State:         farmerProfile.Address.State,
			PostalCode:    farmerProfile.Address.PostalCode,
			Country:       farmerProfile.Address.Country,
			Coordinates:   farmerProfile.Address.Coordinates,
		}
	}

	farmerProfileData := &responses.FarmerProfileData{
		ID:               farmerProfile.GetID(),
		AAAUserID:        farmerProfile.AAAUserID,
		AAAOrgID:         farmerProfile.AAAOrgID,
		KisanSathiUserID: farmerProfile.KisanSathiUserID,
		FirstName:        farmerProfile.FirstName,
		LastName:         farmerProfile.LastName,
		PhoneNumber:      farmerProfile.PhoneNumber,
		Email:            farmerProfile.Email,
		DateOfBirth:      farmerProfile.DateOfBirth,
		Gender:           farmerProfile.Gender,
		Address:          addressData,
		Preferences:      farmerProfile.Preferences,
		Metadata:         farmerProfile.Metadata,
		Farms:            []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:        farmerProfile.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        farmerProfile.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerProfileResponse(farmerProfileData, "Farmer retrieved successfully")
	return &response, nil
}

// UpdateFarmer updates an existing farmer
func (s *FarmerServiceImpl) UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error) {
	// Find existing farmer using flexible lookup
	var filter *base.Filter

	if req.FarmerID != "" {
		// Lookup by primary key (farmer_id)
		filter = base.NewFilterBuilder().
			Where("id", base.OpEqual, req.FarmerID).
			Build()
	} else if req.AAAUserID != "" {
		// Lookup by user_id, optionally filtered by org_id
		filterBuilder := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID)

		if req.AAAOrgID != "" {
			filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
		}

		filter = filterBuilder.Build()
	} else {
		return nil, fmt.Errorf("either farmer_id or aaa_user_id must be provided")
	}

	existingFarmerProfile, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Update fields if provided
	if req.Profile.FirstName != "" {
		existingFarmerProfile.FirstName = req.Profile.FirstName
	}
	if req.Profile.LastName != "" {
		existingFarmerProfile.LastName = req.Profile.LastName
	}
	if req.Profile.PhoneNumber != "" {
		existingFarmerProfile.PhoneNumber = req.Profile.PhoneNumber
	}
	if req.Profile.Email != "" {
		existingFarmerProfile.Email = req.Profile.Email
	}
	if req.Profile.DateOfBirth != "" {
		existingFarmerProfile.DateOfBirth = req.Profile.DateOfBirth
	}
	if req.Profile.Gender != "" {
		existingFarmerProfile.Gender = req.Profile.Gender
	}
	if req.Profile.Address.StreetAddress != "" {
		existingFarmerProfile.Address.StreetAddress = req.Profile.Address.StreetAddress
		existingFarmerProfile.Address.City = req.Profile.Address.City
		existingFarmerProfile.Address.State = req.Profile.Address.State
		existingFarmerProfile.Address.PostalCode = req.Profile.Address.PostalCode
		existingFarmerProfile.Address.Country = req.Profile.Address.Country
		existingFarmerProfile.Address.Coordinates = req.Profile.Address.Coordinates
	}
	if req.KisanSathiUserID != nil {
		existingFarmerProfile.KisanSathiUserID = req.KisanSathiUserID
	}

	// Update metadata
	if req.Profile.Metadata != nil {
		if existingFarmerProfile.Metadata == nil {
			existingFarmerProfile.Metadata = make(map[string]string)
		}
		for k, v := range req.Profile.Metadata {
			existingFarmerProfile.Metadata[k] = v
		}
	}

	existingFarmerProfile.SetUpdatedBy(req.UserID)

	// Save updated farmer
	if err := s.repository.Update(ctx, existingFarmerProfile); err != nil {
		return nil, fmt.Errorf("failed to update farmer: %w", err)
	}

	// Convert to response format
	farmerProfileData := &responses.FarmerProfileData{
		ID:               existingFarmerProfile.GetID(),
		AAAUserID:        existingFarmerProfile.AAAUserID,
		AAAOrgID:         existingFarmerProfile.AAAOrgID,
		KisanSathiUserID: existingFarmerProfile.KisanSathiUserID,
		FirstName:        existingFarmerProfile.FirstName,
		LastName:         existingFarmerProfile.LastName,
		PhoneNumber:      existingFarmerProfile.PhoneNumber,
		Email:            existingFarmerProfile.Email,
		DateOfBirth:      existingFarmerProfile.DateOfBirth,
		Gender:           existingFarmerProfile.Gender,
		Address: responses.AddressData{
			StreetAddress: existingFarmerProfile.Address.StreetAddress,
			City:          existingFarmerProfile.Address.City,
			State:         existingFarmerProfile.Address.State,
			PostalCode:    existingFarmerProfile.Address.PostalCode,
			Country:       existingFarmerProfile.Address.Country,
			Coordinates:   existingFarmerProfile.Address.Coordinates,
		},
		Preferences: existingFarmerProfile.Preferences,
		Metadata:    existingFarmerProfile.Metadata,
		Farms:       []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:   existingFarmerProfile.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   existingFarmerProfile.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerResponse(farmerProfileData, "Farmer updated successfully")
	return &response, nil
}

// DeleteFarmer deletes a farmer
func (s *FarmerServiceImpl) DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error {
	// Find existing farmer using flexible lookup
	var filter *base.Filter

	if req.FarmerID != "" {
		// Lookup by primary key (farmer_id)
		filter = base.NewFilterBuilder().
			Where("id", base.OpEqual, req.FarmerID).
			Build()
	} else if req.AAAUserID != "" {
		// Lookup by user_id, optionally filtered by org_id
		filterBuilder := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID)

		if req.AAAOrgID != "" {
			filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
		}

		filter = filterBuilder.Build()
	} else {
		return fmt.Errorf("either farmer_id or aaa_user_id must be provided")
	}

	existingFarmerProfile, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("farmer not found: %w", err)
	}

	// Perform soft delete
	if err := s.repository.SoftDelete(ctx, existingFarmerProfile.GetID(), req.UserID); err != nil {
		return fmt.Errorf("failed to delete farmer: %w", err)
	}

	return nil
}

// ListFarmers lists farmers with filtering
func (s *FarmerServiceImpl) ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error) {
	// Build filter for database query
	filter := base.NewFilterBuilder().
		Page(req.Page, req.PageSize)

	// Add organization filter if specified
	if req.AAAOrgID != "" {
		filter = filter.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
	}

	// Add KisanSathi filter if specified
	if req.KisanSathiUserID != "" {
		filter = filter.Where("kisan_sathi_user_id", base.OpEqual, req.KisanSathiUserID)
	}

	// Query farmers from repository
	farmerProfiles, err := s.repository.Find(ctx, filter.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to list farmers: %w", err)
	}

	// Get total count - use a temporary count query
	countFilter := base.NewFilterBuilder()
	if req.AAAOrgID != "" {
		countFilter = countFilter.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
	}
	if req.KisanSathiUserID != "" {
		countFilter = countFilter.Where("kisan_sathi_user_id", base.OpEqual, req.KisanSathiUserID)
	}

	// Count without pagination
	allResults, err := s.repository.Find(ctx, countFilter.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to count farmers: %w", err)
	}
	totalCount := int64(len(allResults))

	// Convert to response format
	var farmerProfilesData []*responses.FarmerProfileData
	for _, fp := range farmerProfiles {
		// Handle address data safely
		var addressData responses.AddressData
		if fp.Address != nil {
			addressData = responses.AddressData{
				StreetAddress: fp.Address.StreetAddress,
				City:          fp.Address.City,
				State:         fp.Address.State,
				PostalCode:    fp.Address.PostalCode,
				Country:       fp.Address.Country,
				Coordinates:   fp.Address.Coordinates,
			}
		}

		farmerProfileData := &responses.FarmerProfileData{
			ID:               fp.GetID(),
			AAAUserID:        fp.AAAUserID,
			AAAOrgID:         fp.AAAOrgID,
			KisanSathiUserID: fp.KisanSathiUserID,
			FirstName:        fp.FirstName,
			LastName:         fp.LastName,
			PhoneNumber:      fp.PhoneNumber,
			Email:            fp.Email,
			DateOfBirth:      fp.DateOfBirth,
			Gender:           fp.Gender,
			Address:          addressData,
			Preferences:      fp.Preferences,
			Metadata:         fp.Metadata,
			Farms:            []*responses.FarmData{}, // TODO: Load actual farms
			CreatedAt:        fp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        fp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		farmerProfilesData = append(farmerProfilesData, farmerProfileData)
	}

	response := responses.NewFarmerListResponse(farmerProfilesData, req.Page, req.PageSize, totalCount)
	return &response, nil
}
