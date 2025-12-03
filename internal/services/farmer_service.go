package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/farmers-module/internal/constants"
	"github.com/Kisanlink/farmers-module/internal/entities"
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
	GetFarmerByUserID(ctx context.Context, aaaUserID string) (*responses.FarmerResponse, error)
	UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error)
	DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error
	ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error)
}

// FarmerServiceImpl implements FarmerService
type FarmerServiceImpl struct {
	repository       *farmer.FarmerRepository
	aaaService       AAAService
	fpoConfigService FPOConfigService
	defaultPassword  string
}

// NewFarmerService creates a new farmer service with repository, AAA service, and default password
func NewFarmerService(repository *farmer.FarmerRepository, aaaService AAAService, defaultPassword string) FarmerService {
	return &FarmerServiceImpl{
		repository:      repository,
		aaaService:      aaaService,
		defaultPassword: defaultPassword,
	}
}

// NewFarmerServiceWithFPOConfig creates a new farmer service with FPO config service
func NewFarmerServiceWithFPOConfig(repository *farmer.FarmerRepository, aaaService AAAService, fpoConfigService FPOConfigService, defaultPassword string) FarmerService {
	return &FarmerServiceImpl{
		repository:       repository,
		aaaService:       aaaService,
		fpoConfigService: fpoConfigService,
		defaultPassword:  defaultPassword,
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

	// If aaa_user_id is provided, check if farmer already exists (by user ID only)
	if req.AAAUserID != "" {
		existingFilter := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID).
			Build()

		existing, err := s.repository.FindOne(ctx, existingFilter)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("farmer already exists with aaa_user_id=%s", req.AAAUserID)
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
						SocialCategory:   existingFarmer.SocialCategory,
						AreaType:         existingFarmer.AreaType,
						TotalAcreageHa:   existingFarmer.TotalAcreageHa,
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
	// Note: Farmer is uniquely identified by aaa_user_id only
	// AAAOrgID is optional - stored for backward compatibility if provided
	farmer := farmerentity.NewFarmer()
	farmer.AAAUserID = aaaUserID
	farmer.AAAOrgID = aaaOrgID // Optional: primary org for backward compatibility
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

	// Link FPO configuration if requested
	if req.LinkFPOConfig && s.fpoConfigService != nil {
		if err := s.linkFPOConfigToFarmer(ctx, farmer, aaaOrgID); err != nil {
			// Log warning but don't fail farmer creation
			log.Printf("Warning: Failed to link FPO config to farmer %s: %v", aaaUserID, err)
			if farmer.Metadata == nil {
				farmer.Metadata = make(entities.JSONB)
			}
			farmer.Metadata["fpo_config_link_error"] = err.Error()
			farmer.Metadata["fpo_config_link_pending"] = "true"
			farmer.Metadata["fpo_config_link_attempted_at"] = time.Now().Format(time.RFC3339)

			// Best-effort metadata update (non-fatal if fails)
			if updateErr := s.repository.Update(ctx, farmer); updateErr != nil {
				log.Printf("Error: Failed to store FPO config link failure metadata: %v", updateErr)
			}
		}
	}

	// Ensure farmer role is assigned (best-effort with retry)
	// Following ADR-001: Role Assignment Strategy for Entity Creation
	roleErr := s.ensureFarmerRoleWithRetry(ctx, aaaUserID, aaaOrgID)
	if roleErr != nil {
		log.Printf("Warning: Failed to assign farmer role to user %s: %v", aaaUserID, roleErr)
		// Store failure metadata for reconciliation
		if farmer.Metadata == nil {
			farmer.Metadata = make(entities.JSONB)
		}
		farmer.Metadata["role_assignment_error"] = roleErr.Error()
		farmer.Metadata["role_assignment_pending"] = "true"
		farmer.Metadata["role_assignment_attempted_at"] = time.Now().Format(time.RFC3339)

		// Best-effort metadata update (non-fatal if fails)
		if updateErr := s.repository.Update(ctx, farmer); updateErr != nil {
			log.Printf("Error: Failed to store role assignment failure metadata: %v", updateErr)
		}
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
		SocialCategory:   farmer.SocialCategory,
		AreaType:         farmer.AreaType,
		TotalAcreageHa:   farmer.TotalAcreageHa,
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
	var filterBuilder *base.FilterBuilder

	// Priority: farmer_id > aaa_user_id > aaa_user_id + aaa_org_id
	if req.FarmerID != "" {
		// Lookup by primary key (farmer_id)
		filterBuilder = base.NewFilterBuilder().
			Where("id", base.OpEqual, req.FarmerID)
	} else if req.AAAUserID != "" {
		// Lookup by user_id, optionally filtered by org_id
		filterBuilder = base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, req.AAAUserID)

		if req.AAAOrgID != "" {
			filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
		}
	} else {
		return nil, fmt.Errorf("either farmer_id or aaa_user_id must be provided")
	}

	// Preload relationships: Address, FPOLinkages, and Farms
	filter := filterBuilder.
		Preload("Address").
		Preload("FPOLinkages").
		Preload("Farms").
		Build()

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

	// Convert FPO linkages
	var fpoLinkages []*responses.FarmerLinkData
	if len(farmer.FPOLinkages) > 0 {
		for _, link := range farmer.FPOLinkages {
			fpoLinkages = append(fpoLinkages, &responses.FarmerLinkData{
				ID:               link.GetID(),
				AAAUserID:        link.AAAUserID,
				AAAOrgID:         link.AAAOrgID,
				KisanSathiUserID: link.KisanSathiUserID,
				Status:           link.Status,
				CreatedAt:        link.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:        link.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	// Convert farms
	var farms []*responses.FarmData
	if len(farmer.Farms) > 0 {
		for _, farm := range farmer.Farms {
			var farmName string
			if farm.Name != nil {
				farmName = *farm.Name
			}
			farms = append(farms, &responses.FarmData{
				ID:        farm.ID,
				FarmerID:  farm.FarmerID,
				Name:      farmName,
				AreaHa:    farm.AreaHa,
				CreatedAt: farm.CreatedAt,
				UpdatedAt: farm.UpdatedAt,
			})
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
		SocialCategory:   farmer.SocialCategory,
		AreaType:         farmer.AreaType,
		TotalAcreageHa:   farmer.TotalAcreageHa,
		Address:          addressData,
		FPOLinkages:      fpoLinkages,
		Preferences:      farmer.Preferences,
		Metadata:         farmer.Metadata,
		Farms:            farms,
		CreatedAt:        farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerProfileResponse(farmerProfileData, "Farmer retrieved successfully")
	return &response, nil
}

// GetFarmerByUserID retrieves a farmer by AAA user ID
func (s *FarmerServiceImpl) GetFarmerByUserID(ctx context.Context, aaaUserID string) (*responses.FarmerResponse, error) {
	if aaaUserID == "" {
		return nil, fmt.Errorf("aaa_user_id is required")
	}

	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, aaaUserID).
		Build()

	farmer, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if farmer == nil {
		return nil, fmt.Errorf("farmer not found with aaa_user_id=%s", aaaUserID)
	}

	farmerProfileData := &responses.FarmerProfileData{
		ID:          farmer.ID,
		AAAUserID:   farmer.AAAUserID,
		FirstName:   farmer.FirstName,
		LastName:    farmer.LastName,
		PhoneNumber: farmer.PhoneNumber,
		Email:       farmer.Email,
	}

	return &responses.FarmerResponse{
		BaseResponse: base.NewSuccessResponse("Farmer found", farmerProfileData),
		Data:         farmerProfileData,
	}, nil
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
			existingFarmer.Metadata = make(entities.JSONB)
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
		SocialCategory:   existingFarmer.SocialCategory,
		AreaType:         existingFarmer.AreaType,
		TotalAcreageHa:   existingFarmer.TotalAcreageHa,
		Address:          addressData,
		Preferences:      existingFarmer.Preferences,
		Metadata:         existingFarmer.Metadata,
		Farms:            []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:        existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
// When aaa_org_id is provided, it filters farmers based on their linkage to the FPO organization
// through the farmer_links table instead of the direct aaa_org_id field in the farmers table
func (s *FarmerServiceImpl) ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error) {
	// Build filter for additional query parameters
	filter := base.NewFilterBuilder().
		Page(req.Page, req.PageSize)

	// Add KisanSathi filter if specified
	if req.KisanSathiUserID != "" {
		filter = filter.Where("kisan_sathi_user_id", base.OpEqual, req.KisanSathiUserID)
	}

	// Add phone number filter if specified
	if req.PhoneNumber != "" {
		filter = filter.Where("phone_number", base.OpEqual, req.PhoneNumber)
	}

	var farmers []*farmerentity.Farmer
	var totalCount int64
	var err error

	// If organization filter is specified, use the FPO linkage-based filtering
	if req.AAAOrgID != "" {
		// Use the custom FindByOrgID method that joins with farmer_links table
		farmers, err = s.repository.FindByOrgID(ctx, req.AAAOrgID, filter.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to list farmers by org_id: %w", err)
		}

		// Get count using the same join-based approach
		countFilter := base.NewFilterBuilder()
		if req.KisanSathiUserID != "" {
			countFilter = countFilter.Where("kisan_sathi_user_id", base.OpEqual, req.KisanSathiUserID)
		}
		if req.PhoneNumber != "" {
			countFilter = countFilter.Where("phone_number", base.OpEqual, req.PhoneNumber)
		}

		totalCount, err = s.repository.CountByOrgID(ctx, req.AAAOrgID, countFilter.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to count farmers by org_id: %w", err)
		}
	} else {
		// If no organization filter, use standard repository Find method
		farmers, err = s.repository.Find(ctx, filter.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to list farmers: %w", err)
		}

		// Get total count for pagination
		countFilter := base.NewFilterBuilder()
		if req.KisanSathiUserID != "" {
			countFilter = countFilter.Where("kisan_sathi_user_id", base.OpEqual, req.KisanSathiUserID)
		}
		if req.PhoneNumber != "" {
			countFilter = countFilter.Where("phone_number", base.OpEqual, req.PhoneNumber)
		}

		// Count without pagination
		allResults, err := s.repository.Find(ctx, countFilter.Build())
		if err != nil {
			return nil, fmt.Errorf("failed to count farmers: %w", err)
		}
		totalCount = int64(len(allResults))
	}

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
			SocialCategory:   farmer.SocialCategory,
			AreaType:         farmer.AreaType,
			TotalAcreageHa:   farmer.TotalAcreageHa,
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

// ensureFarmerRoleWithRetry attempts to assign the farmer role with a single retry
// Implements ADR-001: Role Assignment Strategy with synchronous retry
func (s *FarmerServiceImpl) ensureFarmerRoleWithRetry(ctx context.Context, userID, orgID string) error {
	// Attempt role assignment with one retry
	for attempt := 1; attempt <= 2; attempt++ {
		err := s.ensureFarmerRole(ctx, userID, orgID)
		if err == nil {
			return nil // Success
		}

		if attempt == 1 {
			log.Printf("Warning: Role assignment attempt %d failed for user %s, retrying: %v", attempt, userID, err)
			time.Sleep(500 * time.Millisecond) // Brief delay before retry
			continue
		}

		return fmt.Errorf("role assignment failed after %d attempts: %w", attempt, err)
	}
	return nil
}

// ensureFarmerRole ensures the user has the farmer role with idempotency and verification
// Implements the "Check-Assign-Verify" pattern from ADR-001
func (s *FarmerServiceImpl) ensureFarmerRole(ctx context.Context, userID, orgID string) error {
	// 1. Check if role already exists (idempotency)
	hasRole, err := s.aaaService.CheckUserRole(ctx, userID, constants.RoleFarmer)
	if err != nil {
		// If check fails, try assignment anyway (AAA may be degraded but functional)
		log.Printf("Warning: Failed to check farmer role for user %s: %v", userID, err)
	} else if hasRole {
		log.Printf("User %s already has farmer role, skipping assignment", userID)
		return nil
	}

	// 2. Assign farmer role
	err = s.aaaService.AssignRole(ctx, userID, orgID, constants.RoleFarmer)
	if err != nil {
		return fmt.Errorf("failed to assign farmer role: %w", err)
	}

	// 3. Verify assignment succeeded
	hasRole, err = s.aaaService.CheckUserRole(ctx, userID, constants.RoleFarmer)
	if err != nil {
		return fmt.Errorf("failed to verify farmer role assignment: %w", err)
	}
	if !hasRole {
		return fmt.Errorf("farmer role assignment verification failed (role not present)")
	}

	log.Printf("Successfully assigned and verified farmer role for user %s", userID)
	return nil
}

// linkFPOConfigToFarmer links FPO configuration to farmer by adding metadata
func (s *FarmerServiceImpl) linkFPOConfigToFarmer(ctx context.Context, farmer *farmerentity.Farmer, aaaOrgID string) error {
	// Verify FPO config service is available
	if s.fpoConfigService == nil {
		return fmt.Errorf("FPO config service not available")
	}

	// Check if FPO config exists for this organization
	fpoConfig, err := s.fpoConfigService.GetFPOConfig(ctx, aaaOrgID)
	if err != nil {
		return fmt.Errorf("failed to retrieve FPO config for org %s: %w", aaaOrgID, err)
	}

	// If config exists but is not configured, skip linking
	if fpoConfig.APIHealthStatus == "not_configured" {
		log.Printf("FPO config not configured for org %s, skipping link", aaaOrgID)
		return fmt.Errorf("FPO config not configured for org %s", aaaOrgID)
	}

	// Add FPO config metadata to farmer
	if farmer.Metadata == nil {
		farmer.Metadata = make(entities.JSONB)
	}

	farmer.Metadata["fpo_config_linked"] = true
	farmer.Metadata["fpo_config_id"] = fpoConfig.ID
	farmer.Metadata["fpo_name"] = fpoConfig.FPOName
	farmer.Metadata["fpo_config_linked_at"] = time.Now().Format(time.RFC3339)

	// Update farmer with metadata
	if err := s.repository.Update(ctx, farmer); err != nil {
		return fmt.Errorf("failed to update farmer with FPO config metadata: %w", err)
	}

	log.Printf("Successfully linked FPO config %s to farmer %s", fpoConfig.ID, farmer.GetID())
	return nil
}
