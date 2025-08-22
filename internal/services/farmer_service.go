package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
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
	repository *base.BaseFilterableRepository[*farmer.Farmer]
	dbManager  db.DBManager
}

// NewFarmerService creates a new farmer service with DBManager
func NewFarmerService(dbManager db.DBManager) FarmerService {
	repo := base.NewBaseFilterableRepository[*farmer.Farmer]()
	repo.SetDBManager(dbManager)

	return &FarmerServiceImpl{
		repository: repo,
		dbManager:  dbManager,
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

	// Create new farmer model
	farmer := farmer.NewFarmer()
	farmer.AAAUserID = req.AAAUserID
	farmer.AAAOrgID = req.AAAOrgID
	farmer.KisanSathiUserID = req.KisanSathiUserID
	farmer.FirstName = req.Profile.FirstName
	farmer.LastName = req.Profile.LastName
	farmer.PhoneNumber = req.Profile.PhoneNumber
	farmer.Email = req.Profile.Email
	farmer.DateOfBirth = req.Profile.DateOfBirth
	farmer.Gender = req.Profile.Gender
	farmer.StreetAddress = req.Profile.Address.StreetAddress
	farmer.City = req.Profile.Address.City
	farmer.State = req.Profile.Address.State
	farmer.PostalCode = req.Profile.Address.PostalCode
	farmer.Country = req.Profile.Address.Country
	farmer.Coordinates = req.Profile.Address.Coordinates
	farmer.Preferences = req.Profile.Preferences
	farmer.Metadata = req.Profile.Metadata
	farmer.SetCreatedBy(req.UserID)

	// Save to repository
	if err := s.repository.Create(ctx, farmer); err != nil {
		return nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Convert to response format
	farmerProfile := &responses.FarmerProfileData{
		AAAUserID:        farmer.AAAUserID,
		AAAOrgID:         farmer.AAAOrgID,
		KisanSathiUserID: farmer.KisanSathiUserID,
		FirstName:        farmer.FirstName,
		LastName:         farmer.LastName,
		PhoneNumber:      farmer.PhoneNumber,
		Email:            farmer.Email,
		DateOfBirth:      farmer.DateOfBirth,
		Gender:           farmer.Gender,
		Address: responses.AddressData{
			StreetAddress: farmer.StreetAddress,
			City:          farmer.City,
			State:         farmer.State,
			PostalCode:    farmer.PostalCode,
			Country:       farmer.Country,
			Coordinates:   farmer.Coordinates,
		},
		Preferences: farmer.Preferences,
		Metadata:    farmer.Metadata,
		Farms:       []*responses.FarmData{},
		CreatedAt:   farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerResponse(farmerProfile, "Farmer created successfully")
	return &response, nil
}

// GetFarmer retrieves a farmer by ID
func (s *FarmerServiceImpl) GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerProfileResponse, error) {
	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, req.AAAUserID).
		Where("aaa_org_id", base.OpEqual, req.AAAOrgID).
		Build()

	farmer, err := s.repository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Convert to response format
	farmerProfile := &responses.FarmerProfileData{
		AAAUserID:        farmer.AAAUserID,
		AAAOrgID:         farmer.AAAOrgID,
		KisanSathiUserID: farmer.KisanSathiUserID,
		FirstName:        farmer.FirstName,
		LastName:         farmer.LastName,
		PhoneNumber:      farmer.PhoneNumber,
		Email:            farmer.Email,
		DateOfBirth:      farmer.DateOfBirth,
		Gender:           farmer.Gender,
		Address: responses.AddressData{
			StreetAddress: farmer.StreetAddress,
			City:          farmer.City,
			State:         farmer.State,
			PostalCode:    farmer.PostalCode,
			Country:       farmer.Country,
			Coordinates:   farmer.Coordinates,
		},
		Preferences: farmer.Preferences,
		Metadata:    farmer.Metadata,
		Farms:       []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:   farmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   farmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerProfileResponse(farmerProfile, "Farmer retrieved successfully")
	return &response, nil
}

// UpdateFarmer updates an existing farmer
func (s *FarmerServiceImpl) UpdateFarmer(ctx context.Context, req *requests.UpdateFarmerRequest) (*responses.FarmerResponse, error) {
	// Find existing farmer
	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, req.AAAUserID).
		Where("aaa_org_id", base.OpEqual, req.AAAOrgID).
		Build()

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
		existingFarmer.DateOfBirth = req.Profile.DateOfBirth
	}
	if req.Profile.Gender != "" {
		existingFarmer.Gender = req.Profile.Gender
	}
	if req.Profile.Address.StreetAddress != "" {
		existingFarmer.StreetAddress = req.Profile.Address.StreetAddress
		existingFarmer.City = req.Profile.Address.City
		existingFarmer.State = req.Profile.Address.State
		existingFarmer.PostalCode = req.Profile.Address.PostalCode
		existingFarmer.Country = req.Profile.Address.Country
		existingFarmer.Coordinates = req.Profile.Address.Coordinates
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
	farmerProfile := &responses.FarmerProfileData{
		AAAUserID:        existingFarmer.AAAUserID,
		AAAOrgID:         existingFarmer.AAAOrgID,
		KisanSathiUserID: existingFarmer.KisanSathiUserID,
		FirstName:        existingFarmer.FirstName,
		LastName:         existingFarmer.LastName,
		PhoneNumber:      existingFarmer.PhoneNumber,
		Email:            existingFarmer.Email,
		DateOfBirth:      existingFarmer.DateOfBirth,
		Gender:           existingFarmer.Gender,
		Address: responses.AddressData{
			StreetAddress: existingFarmer.StreetAddress,
			City:          existingFarmer.City,
			State:         existingFarmer.State,
			PostalCode:    existingFarmer.PostalCode,
			Country:       existingFarmer.Country,
			Coordinates:   existingFarmer.Coordinates,
		},
		Preferences: existingFarmer.Preferences,
		Metadata:    existingFarmer.Metadata,
		Farms:       []*responses.FarmData{}, // TODO: Load actual farms
		CreatedAt:   existingFarmer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   existingFarmer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response := responses.NewFarmerResponse(farmerProfile, "Farmer updated successfully")
	return &response, nil
}

// DeleteFarmer deletes a farmer
func (s *FarmerServiceImpl) DeleteFarmer(ctx context.Context, req *requests.DeleteFarmerRequest) error {
	// Find existing farmer
	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, req.AAAUserID).
		Where("aaa_org_id", base.OpEqual, req.AAAOrgID).
		Build()

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

	// Get total count
	totalCount, err := s.repository.Count(ctx, filter.Build(), farmer.NewFarmer())
	if err != nil {
		return nil, fmt.Errorf("failed to count farmers: %w", err)
	}

	// Convert to response format
	var farmerProfiles []*responses.FarmerProfileData
	for _, f := range farmers {
		farmerProfile := &responses.FarmerProfileData{
			AAAUserID:        f.AAAUserID,
			AAAOrgID:         f.AAAOrgID,
			KisanSathiUserID: f.KisanSathiUserID,
			FirstName:        f.FirstName,
			LastName:         f.LastName,
			PhoneNumber:      f.PhoneNumber,
			Email:            f.Email,
			DateOfBirth:      f.DateOfBirth,
			Gender:           f.Gender,
			Address: responses.AddressData{
				StreetAddress: f.StreetAddress,
				City:          f.City,
				State:         f.State,
				PostalCode:    f.PostalCode,
				Country:       f.Country,
				Coordinates:   f.Coordinates,
			},
			Preferences: f.Preferences,
			Metadata:    f.Metadata,
			Farms:       []*responses.FarmData{}, // TODO: Load actual farms
			CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		farmerProfiles = append(farmerProfiles, farmerProfile)
	}

	response := responses.NewFarmerListResponse(farmerProfiles, req.Page, req.PageSize, totalCount)
	return &response, nil
}
