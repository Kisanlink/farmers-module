package services

import (
	"context"
	"fmt"
	"log"

	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
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
	repository      *farmer.FarmerRepository
	aaaService      AAAService
	defaultPassword string
}

// NewFarmerService creates a new farmer service with repository, AAA service, and default password
func NewFarmerService(repository *farmer.FarmerRepository, aaaService AAAService, defaultPassword string) FarmerService {
	return &FarmerServiceImpl{
		repository:      repository,
		aaaService:      aaaService,
		defaultPassword: defaultPassword,
	}
}

// Helper function to safely dereference DateOfBirth pointer
func safeDerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper function to convert string to pointer (returns nil for empty string)
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateFarmer creates a new farmer
// Supports two workflows:
// 1. Existing AAA user: Provide aaa_user_id + aaa_org_id
// 2. New AAA user: Provide country_code + phone_number + aaa_org_id (auto-creates/finds user in AAA)
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error) {
	// Validate request: Either aaa_user_id OR (country_code + phone_number) must be provided
	if req.AAAUserID == "" && (req.Profile.CountryCode == "" || req.Profile.PhoneNumber == "") {
		return nil, fmt.Errorf("either aaa_user_id or (country_code + phone_number) must be provided")
	}

	// Validate org_id is always required
	if req.AAAOrgID == "" {
		return nil, fmt.Errorf("aaa_org_id is required")
	}

	// If aaa_user_id is provided, check if farmer already exists
	if req.AAAUserID != "" {
		existingFilter := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID).
			Where("aaa_org_id", base.OpEqual, req.AAAOrgID).
			Build()

		existing, err := s.repository.FindOne(ctx, existingFilter)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("farmer already exists")
		}
	}

	// Determine AAA user ID - either use provided or create from phone number
	var aaaUserID string
	var aaaOrgID string

	// Workflow 1: Use existing AAA user ID if provided
	if req.AAAUserID != "" {
		aaaUserID = req.AAAUserID
		aaaOrgID = req.AAAOrgID
	} else if req.Profile.PhoneNumber != "" {
		// Workflow 2: Create or find AAA user by phone number
		// Business Rule 5.1: RegisterFarmer idempotency - return existing farmer if phone exists

		// Attempt to create user in AAA
		log.Printf("Creating user in AAA with mobile: %s, country_code: %s", req.Profile.PhoneNumber, req.Profile.CountryCode)

		createUserReq := map[string]interface{}{
			"phone_number": req.Profile.PhoneNumber,
			"password":     s.defaultPassword,
			"country_code": req.Profile.CountryCode,
		}

		// Add username if provided, otherwise AAA will auto-generate one
		if req.Profile.Username != "" {
			createUserReq["username"] = req.Profile.Username
		}

		aaaUser, err := s.aaaService.CreateUser(ctx, createUserReq)
		if err != nil {
			// Check if it's a conflict error (user already exists)
			log.Printf("Failed to create user in AAA (likely already exists), attempting to get existing user: %v", err)

			// Try to get existing user by mobile number
			aaaUser, err = s.aaaService.GetUserByMobile(ctx, req.Profile.PhoneNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to create or retrieve user from AAA: %w", err)
			}
			log.Printf("Found existing user in AAA with mobile: %s", req.Profile.PhoneNumber)

			// Extract AAA user ID from existing user
			if userMap, ok := aaaUser.(map[string]interface{}); ok {
				if id, exists := userMap["id"]; exists {
					aaaUserID = fmt.Sprintf("%v", id)
				}
			}

			// Check if farmer profile already exists locally (idempotent operation)
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
						DateOfBirth:      safeDerefString(existingFarmer.DateOfBirth),
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
		} else {
			// User created successfully in AAA
			log.Printf("User created successfully in AAA with mobile: %s", req.Profile.PhoneNumber)
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

	// Create new farmer with normalized address
	farmer := farmerentity.NewFarmer()
	farmer.AAAUserID = aaaUserID
	farmer.AAAOrgID = aaaOrgID
	farmer.KisanSathiUserID = validatedKisanSathiID
	farmer.FirstName = req.Profile.FirstName
	farmer.LastName = req.Profile.LastName
	farmer.PhoneNumber = req.Profile.PhoneNumber
	farmer.Email = req.Profile.Email
	farmer.DateOfBirth = stringToPtr(req.Profile.DateOfBirth)
	farmer.Gender = req.Profile.Gender
	farmer.Preferences = req.Profile.Preferences
	farmer.Metadata = req.Profile.Metadata
	farmer.Status = "ACTIVE"
	farmer.SetCreatedBy(req.UserID)

	// Create address entity if address data is provided
	if req.Profile.Address.StreetAddress != "" || req.Profile.Address.City != "" {
		address := farmerentity.NewAddress()
		address.StreetAddress = req.Profile.Address.StreetAddress
		address.City = req.Profile.Address.City
		address.State = req.Profile.Address.State
		address.PostalCode = req.Profile.Address.PostalCode
		address.Country = req.Profile.Address.Country
		address.Coordinates = req.Profile.Address.Coordinates
		address.SetCreatedBy(req.UserID)

		// Set the address for the farmer (establishes FK relationship)
		farmer.SetAddress(address)
	}

	// Save to repository
	if err := s.repository.Create(ctx, farmer); err != nil {
		return nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Convert to response format
	// Handle address data safely
	var addressData responses.AddressData
	if farmer.Address != nil {
		addressData = responses.AddressData{
			StreetAddress: farmer.Address.StreetAddress,
			City:          farmer.Address.City,
			State:         farmer.Address.State,
			PostalCode:    farmer.Address.PostalCode,
			Country:       farmer.Address.Country,
			Coordinates:   farmer.Address.Coordinates,
		}
	}

	farmerProfileData := &responses.FarmerProfileData{
		ID:               farmer.GetID(),
		AAAUserID:        farmer.AAAUserID,
		AAAOrgID:         farmer.AAAOrgID,
		KisanSathiUserID: farmer.KisanSathiUserID,
		FirstName:        farmer.FirstName,
		LastName:         farmer.LastName,
		PhoneNumber:      farmer.PhoneNumber,
		Email:            farmer.Email,
		DateOfBirth:      safeDerefString(farmer.DateOfBirth),
		Gender:           farmer.Gender,
		Address:          addressData,
		Preferences:      farmer.Preferences,
		Metadata:         farmer.Metadata,
		Farms:            []*responses.FarmData{},
		CreatedAt:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	farmer, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Convert to response format
	// Handle address data safely
	var addressData responses.AddressData
	if farmer.Address != nil {
		addressData = responses.AddressData{
			StreetAddress: farmer.Address.StreetAddress,
			City:          farmer.Address.City,
			State:         farmer.Address.State,
			PostalCode:    farmer.Address.PostalCode,
			Country:       farmer.Address.Country,
			Coordinates:   farmer.Address.Coordinates,
		}
	}

	farmerProfileData := &responses.FarmerProfileData{
		ID:               farmer.GetID(),
		AAAUserID:        farmer.AAAUserID,
		AAAOrgID:         farmer.AAAOrgID,
		KisanSathiUserID: farmer.KisanSathiUserID,
		FirstName:        farmer.FirstName,
		LastName:         farmer.LastName,
		PhoneNumber:      farmer.PhoneNumber,
		Email:            farmer.Email,
		DateOfBirth:      safeDerefString(farmer.DateOfBirth),
		Gender:           farmer.Gender,
		Address:          addressData,
		Preferences:      farmer.Preferences,
		Metadata:         farmer.Metadata,
		Farms:            []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	existingFarmer, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Update fields if provided
	if req.Profile.FirstName != "" {
		existingFarmer.FirstName = req.Profile.FirstName
	}
	if req.Profile.LastName != "" {
		existingFarmer.LastName = req.Profile.LastName
	}
	if req.Profile.PhoneNumber != "" {
		existingFarmer.PhoneNumber = req.Profile.PhoneNumber
	}
	if req.Profile.Email != "" {
		existingFarmer.Email = req.Profile.Email
	}
	if req.Profile.DateOfBirth != "" {
		existingFarmer.DateOfBirth = stringToPtr(req.Profile.DateOfBirth)
	}
	if req.Profile.Gender != "" {
		existingFarmer.Gender = req.Profile.Gender
	}
	if req.Profile.Address.StreetAddress != "" {
		existingFarmer.Address.StreetAddress = req.Profile.Address.StreetAddress
		existingFarmer.Address.City = req.Profile.Address.City
		existingFarmer.Address.State = req.Profile.Address.State
		existingFarmer.Address.PostalCode = req.Profile.Address.PostalCode
		existingFarmer.Address.Country = req.Profile.Address.Country
		existingFarmer.Address.Coordinates = req.Profile.Address.Coordinates
	}
	if req.KisanSathiUserID != nil {
		existingFarmer.KisanSathiUserID = req.KisanSathiUserID
	}

	// Update metadata
	if req.Profile.Metadata != nil {
		if existingFarmer.Metadata == nil {
			existingFarmer.Metadata = make(map[string]string)
		}
		for k, v := range req.Profile.Metadata {
			existingFarmer.Metadata[k] = v
		}
	}

	existingFarmer.SetUpdatedBy(req.UserID)

	// Save updated farmer
	if err := s.repository.Update(ctx, existingFarmer); err != nil {
		return nil, fmt.Errorf("failed to update farmer: %w", err)
	}

	// Convert to response format
	farmerProfileData := &responses.FarmerProfileData{
		ID:               existingFarmer.GetID(),
		AAAUserID:        existingFarmer.AAAUserID,
		AAAOrgID:         existingFarmer.AAAOrgID,
		KisanSathiUserID: existingFarmer.KisanSathiUserID,
		FirstName:        existingFarmer.FirstName,
		LastName:         existingFarmer.LastName,
		PhoneNumber:      existingFarmer.PhoneNumber,
		Email:            existingFarmer.Email,
		DateOfBirth:      safeDerefString(existingFarmer.DateOfBirth),
		Gender:           existingFarmer.Gender,
		Address: responses.AddressData{
			StreetAddress: existingFarmer.Address.StreetAddress,
			City:          existingFarmer.Address.City,
			State:         existingFarmer.Address.State,
			PostalCode:    existingFarmer.Address.PostalCode,
			Country:       existingFarmer.Address.Country,
			Coordinates:   existingFarmer.Address.Coordinates,
		},
		Preferences: existingFarmer.Preferences,
		Metadata:    existingFarmer.Metadata,
		Farms:       []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:   existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

	existingFarmer, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("farmer not found: %w", err)
	}

	// Perform soft delete
	if err := s.repository.SoftDelete(ctx, existingFarmer.GetID(), req.UserID); err != nil {
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
	farmers, err := s.repository.Find(ctx, filter.Build())
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
	for _, farmer := range farmers {
		// Handle address data safely
		var addressData responses.AddressData
		if farmer.Address != nil {
			addressData = responses.AddressData{
				StreetAddress: farmer.Address.StreetAddress,
				City:          farmer.Address.City,
				State:         farmer.Address.State,
				PostalCode:    farmer.Address.PostalCode,
				Country:       farmer.Address.Country,
				Coordinates:   farmer.Address.Coordinates,
			}
		}

		farmerProfileData := &responses.FarmerProfileData{
			ID:               farmer.GetID(),
			AAAUserID:        farmer.AAAUserID,
			AAAOrgID:         farmer.AAAOrgID,
			KisanSathiUserID: farmer.KisanSathiUserID,
			FirstName:        farmer.FirstName,
			LastName:         farmer.LastName,
			PhoneNumber:      farmer.PhoneNumber,
			Email:            farmer.Email,
			DateOfBirth:      safeDerefString(farmer.DateOfBirth),
			Gender:           farmer.Gender,
			Address:          addressData,
			Preferences:      farmer.Preferences,
			Metadata:         farmer.Metadata,
			Farms:            []*responses.FarmData{}, // TODO: Load actual farms
			CreatedAt:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		farmerProfilesData = append(farmerProfilesData, farmerProfileData)
	}

	response := responses.NewFarmerListResponse(farmerProfilesData, req.Page, req.PageSize, totalCount)
	return &response, nil
}
