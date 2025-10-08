package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/auth"
	cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
	cropVarietyEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/repo/crop"
	"github.com/Kisanlink/farmers-module/pkg/common"
)

// CropServiceImpl implements CropService
type CropServiceImpl struct {
	cropRepo        *crop.CropRepository
	cropVarietyRepo *crop.CropVarietyRepository
	aaaService      AAAService
}

// NewCropService creates a new crop service
func NewCropService(cropRepo *crop.CropRepository, cropVarietyRepo *crop.CropVarietyRepository, aaaService AAAService) CropService {
	return &CropServiceImpl{
		cropRepo:        cropRepo,
		cropVarietyRepo: cropVarietyRepo,
		aaaService:      aaaService,
	}
}

// CreateCrop implements CropService.CreateCrop
func (s *CropServiceImpl) CreateCrop(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop", "create", "", createReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Check if crop with same name already exists
	existingCrop, err := s.cropRepo.FindByName(ctx, createReq.Name)
	if err == nil && existingCrop != nil {
		return nil, common.ErrAlreadyExists
	}

	// Create crop entity
	cropEnt := cropEntity.NewCrop()
	cropEnt.Name = createReq.Name
	cropEnt.ScientificName = createReq.ScientificName
	cropEnt.Category = cropEntity.CropCategory(createReq.Category)
	cropEnt.DurationDays = createReq.DurationDays
	cropEnt.Seasons = createReq.Seasons
	cropEnt.Unit = createReq.Unit
	if createReq.Properties != nil {
		cropEnt.Properties = createReq.Properties
	}

	// Validate the crop entity
	if err := cropEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropRepo.Create(ctx, cropEnt); err != nil {
		return nil, fmt.Errorf("failed to create crop: %w", err)
	}

	// Convert to response
	cropData := &responses.CropData{
		ID:             cropEnt.ID,
		Name:           cropEnt.Name,
		ScientificName: cropEnt.ScientificName,
		Category:       string(cropEnt.Category),
		DurationDays:   cropEnt.DurationDays,
		Seasons:        cropEnt.Seasons,
		Unit:           cropEnt.Unit,
		Properties:     cropEnt.Properties,
		IsActive:       cropEnt.IsActive,
		CreatedAt:      cropEnt.CreatedAt,
		UpdatedAt:      cropEnt.UpdatedAt,
	}

	return &responses.CropResponse{
		Success:   true,
		Message:   "Crop created successfully",
		RequestID: createReq.RequestID,
		Data:      cropData,
	}, nil
}

// GetCrop implements CropService.GetCrop
func (s *CropServiceImpl) GetCrop(ctx context.Context, req interface{}) (interface{}, error) {
	getReq, ok := req.(*requests.GetCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop", "read", getReq.ID, getReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get crop with variety count
	cropWithCount, err := s.cropRepo.GetCropWithVarietyCount(ctx, getReq.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop: %w", err)
	}

	// Convert to response
	cropData := &responses.CropData{
		ID:             cropWithCount.ID,
		Name:           cropWithCount.Name,
		ScientificName: cropWithCount.ScientificName,
		Category:       string(cropWithCount.Category),
		DurationDays:   cropWithCount.DurationDays,
		Seasons:        cropWithCount.Seasons,
		Unit:           cropWithCount.Unit,
		Properties:     cropWithCount.Properties,
		IsActive:       cropWithCount.IsActive,
		CreatedAt:      cropWithCount.CreatedAt,
		UpdatedAt:      cropWithCount.UpdatedAt,
		VarietyCount:   cropWithCount.VarietyCount,
	}

	return &responses.CropResponse{
		Success:   true,
		Message:   "Crop retrieved successfully",
		RequestID: getReq.RequestID,
		Data:      cropData,
	}, nil
}

// UpdateCrop implements CropService.UpdateCrop
func (s *CropServiceImpl) UpdateCrop(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop", "update", updateReq.ID, updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing crop
	var cropEnt cropEntity.Crop
	_, err = s.cropRepo.GetByID(ctx, updateReq.ID, &cropEnt)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop: %w", err)
	}

	// Update fields if provided
	if updateReq.Name != nil {
		// Check if another crop with same name exists
		existingCrop, err := s.cropRepo.FindByName(ctx, *updateReq.Name)
		if err == nil && existingCrop != nil && existingCrop.ID != updateReq.ID {
			return nil, common.ErrAlreadyExists
		}
		cropEnt.Name = *updateReq.Name
	}
	if updateReq.ScientificName != nil {
		cropEnt.ScientificName = updateReq.ScientificName
	}
	if updateReq.Category != nil {
		cropEnt.Category = cropEntity.CropCategory(*updateReq.Category)
	}
	if updateReq.DurationDays != nil {
		cropEnt.DurationDays = updateReq.DurationDays
	}
	if updateReq.Seasons != nil {
		cropEnt.Seasons = updateReq.Seasons
	}
	if updateReq.Unit != nil {
		cropEnt.Unit = *updateReq.Unit
	}
	if updateReq.Properties != nil {
		cropEnt.Properties = updateReq.Properties
	}
	if updateReq.IsActive != nil {
		cropEnt.IsActive = *updateReq.IsActive
	}

	// Validate the updated crop entity
	if err := cropEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropRepo.Update(ctx, &cropEnt); err != nil {
		return nil, fmt.Errorf("failed to update crop: %w", err)
	}

	// Convert to response
	cropData := &responses.CropData{
		ID:             cropEnt.ID,
		Name:           cropEnt.Name,
		ScientificName: cropEnt.ScientificName,
		Category:       string(cropEnt.Category),
		DurationDays:   cropEnt.DurationDays,
		Seasons:        cropEnt.Seasons,
		Unit:           cropEnt.Unit,
		Properties:     cropEnt.Properties,
		IsActive:       cropEnt.IsActive,
		CreatedAt:      cropEnt.CreatedAt,
		UpdatedAt:      cropEnt.UpdatedAt,
	}

	return &responses.CropResponse{
		Success:   true,
		Message:   "Crop updated successfully",
		RequestID: updateReq.RequestID,
		Data:      cropData,
	}, nil
}

// DeleteCrop implements CropService.DeleteCrop
func (s *CropServiceImpl) DeleteCrop(ctx context.Context, req interface{}) (interface{}, error) {
	deleteReq, ok := req.(*requests.DeleteCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop", "delete", deleteReq.ID, deleteReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing crop
	var cropEnt cropEntity.Crop
	_, err = s.cropRepo.GetByID(ctx, deleteReq.ID, &cropEnt)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop: %w", err)
	}

	// Soft delete by setting is_active to false
	cropEnt.IsActive = false
	if err := s.cropRepo.Update(ctx, &cropEnt); err != nil {
		return nil, fmt.Errorf("failed to delete crop: %w", err)
	}

	return &responses.CropResponse{
		Success:   true,
		Message:   "Crop deleted successfully",
		RequestID: deleteReq.RequestID,
		Data:      nil,
	}, nil
}

// ListCrops implements CropService.ListCrops
func (s *CropServiceImpl) ListCrops(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListCropsRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop", "read", "", listReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := crop.CropFilters{
		Category: listReq.Category,
		Season:   listReq.Season,
		Seasons:  listReq.Seasons,
		IsActive: listReq.IsActive,
		Search:   listReq.Search,
		Page:     listReq.Page,
		PageSize: listReq.PageSize,
	}

	// Get crops
	crops, total, err := s.cropRepo.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list crops: %w", err)
	}

	// Convert to response data
	var cropDataList []*responses.CropData
	for _, cropEnt := range crops {
		cropData := &responses.CropData{
			ID:             cropEnt.ID,
			Name:           cropEnt.Name,
			ScientificName: cropEnt.ScientificName,
			Category:       string(cropEnt.Category),
			DurationDays:   cropEnt.DurationDays,
			Seasons:        cropEnt.Seasons,
			Unit:           cropEnt.Unit,
			Properties:     cropEnt.Properties,
			IsActive:       cropEnt.IsActive,
			CreatedAt:      cropEnt.CreatedAt,
			UpdatedAt:      cropEnt.UpdatedAt,
		}
		cropDataList = append(cropDataList, cropData)
	}

	return &responses.CropListResponse{
		Success:   true,
		Message:   "Crops retrieved successfully",
		RequestID: listReq.RequestID,
		Data:      cropDataList,
		Page:      listReq.Page,
		PageSize:  listReq.PageSize,
		Total:     total,
	}, nil
}

// CreateCropVariety implements CropService.CreateCropVariety
func (s *CropServiceImpl) CreateCropVariety(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateCropVarietyRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop_variety", "create", "", createReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Check if crop exists
	var crop cropEntity.Crop
	_, err = s.cropRepo.GetByID(ctx, createReq.CropID, &crop)
	if err != nil {
		return nil, fmt.Errorf("crop not found: %w", err)
	}

	// Check if variety with same name already exists for this crop
	exists, err := s.cropVarietyRepo.CheckVarietyExists(ctx, createReq.CropID, createReq.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check variety existence: %w", err)
	}
	if exists {
		return nil, common.ErrAlreadyExists
	}

	// Create crop variety entity
	varietyEnt := cropVarietyEntity.NewCropVariety()
	varietyEnt.CropID = createReq.CropID
	varietyEnt.Name = createReq.Name
	varietyEnt.Description = createReq.Description
	varietyEnt.DurationDays = createReq.DurationDays
	varietyEnt.YieldPerAcre = createReq.YieldPerAcre
	varietyEnt.YieldPerTree = createReq.YieldPerTree
	if createReq.Properties != nil {
		varietyEnt.Properties = createReq.Properties
	}

	// Validate the variety entity
	if err := varietyEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropVarietyRepo.Create(ctx, varietyEnt); err != nil {
		return nil, fmt.Errorf("failed to create crop variety: %w", err)
	}

	// Convert to response
	varietyData := &responses.CropVarietyData{
		ID:                 varietyEnt.ID,
		CropID:             varietyEnt.CropID,
		Name:               varietyEnt.Name,
		Description:        varietyEnt.Description,
		DurationDays: varietyEnt.DurationDays,
		YieldPerAcre: varietyEnt.YieldPerAcre,
		YieldPerTree: varietyEnt.YieldPerTree,
		Properties:   varietyEnt.Properties,
		IsActive:     varietyEnt.IsActive,
		CreatedAt:    varietyEnt.CreatedAt,
		UpdatedAt:    varietyEnt.UpdatedAt,
	}

	return &responses.CropVarietyResponse{
		Success:   true,
		Message:   "Crop variety created successfully",
		RequestID: createReq.RequestID,
		Data:      varietyData,
	}, nil
}

// GetCropVariety implements CropService.GetCropVariety
func (s *CropServiceImpl) GetCropVariety(ctx context.Context, req interface{}) (interface{}, error) {
	getReq, ok := req.(*requests.GetCropVarietyRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop_variety", "read", getReq.ID, getReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get crop variety
	var varietyEnt cropVarietyEntity.CropVariety
	_, err = s.cropVarietyRepo.GetByID(ctx, getReq.ID, &varietyEnt)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop variety: %w", err)
	}

	// Convert to response
	varietyData := &responses.CropVarietyData{
		ID:                 varietyEnt.ID,
		CropID:             varietyEnt.CropID,
		Name:               varietyEnt.Name,
		Description:        varietyEnt.Description,
		DurationDays: varietyEnt.DurationDays,
		YieldPerAcre: varietyEnt.YieldPerAcre,
		YieldPerTree: varietyEnt.YieldPerTree,
		Properties:   varietyEnt.Properties,
		IsActive:     varietyEnt.IsActive,
		CreatedAt:    varietyEnt.CreatedAt,
		UpdatedAt:    varietyEnt.UpdatedAt,
	}

	// Add crop name if crop relationship is loaded
	if varietyEnt.Crop.Name != "" {
		varietyData.CropName = varietyEnt.Crop.Name
	}

	return &responses.CropVarietyResponse{
		Success:   true,
		Message:   "Crop variety retrieved successfully",
		RequestID: getReq.RequestID,
		Data:      varietyData,
	}, nil
}

// UpdateCropVariety implements CropService.UpdateCropVariety
func (s *CropServiceImpl) UpdateCropVariety(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateCropVarietyRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop_variety", "update", updateReq.ID, updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing variety
	var varietyEnt cropVarietyEntity.CropVariety
	_, err = s.cropVarietyRepo.GetByID(ctx, updateReq.ID, &varietyEnt)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop variety: %w", err)
	}

	// Update fields if provided
	if updateReq.Name != nil {
		// Check if another variety with same name exists for this crop
		exists, err := s.cropVarietyRepo.CheckVarietyExists(ctx, varietyEnt.CropID, *updateReq.Name, updateReq.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check variety existence: %w", err)
		}
		if exists {
			return nil, common.ErrAlreadyExists
		}
		varietyEnt.Name = *updateReq.Name
	}
	if updateReq.Description != nil {
		varietyEnt.Description = updateReq.Description
	}
	if updateReq.DurationDays != nil {
		varietyEnt.DurationDays = updateReq.DurationDays
	}
	if updateReq.YieldPerAcre != nil {
		varietyEnt.YieldPerAcre = updateReq.YieldPerAcre
	}
	if updateReq.YieldPerTree != nil {
		varietyEnt.YieldPerTree = updateReq.YieldPerTree
	}
	if updateReq.Properties != nil {
		varietyEnt.Properties = updateReq.Properties
	}
	if updateReq.IsActive != nil {
		varietyEnt.IsActive = *updateReq.IsActive
	}

	// Validate the updated variety entity
	if err := varietyEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropVarietyRepo.Update(ctx, &varietyEnt); err != nil {
		return nil, fmt.Errorf("failed to update crop variety: %w", err)
	}

	// Convert to response
	varietyData := &responses.CropVarietyData{
		ID:                 varietyEnt.ID,
		CropID:             varietyEnt.CropID,
		Name:               varietyEnt.Name,
		Description:        varietyEnt.Description,
		DurationDays: varietyEnt.DurationDays,
		YieldPerAcre: varietyEnt.YieldPerAcre,
		YieldPerTree: varietyEnt.YieldPerTree,
		Properties:   varietyEnt.Properties,
		IsActive:     varietyEnt.IsActive,
		CreatedAt:    varietyEnt.CreatedAt,
		UpdatedAt:    varietyEnt.UpdatedAt,
	}

	return &responses.CropVarietyResponse{
		Success:   true,
		Message:   "Crop variety updated successfully",
		RequestID: updateReq.RequestID,
		Data:      varietyData,
	}, nil
}

// DeleteCropVariety implements CropService.DeleteCropVariety
func (s *CropServiceImpl) DeleteCropVariety(ctx context.Context, req interface{}) (interface{}, error) {
	deleteReq, ok := req.(*requests.DeleteCropVarietyRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop_variety", "delete", deleteReq.ID, deleteReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing variety
	var varietyEnt cropVarietyEntity.CropVariety
	_, err = s.cropVarietyRepo.GetByID(ctx, deleteReq.ID, &varietyEnt)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop variety: %w", err)
	}

	// Soft delete by setting is_active to false
	varietyEnt.IsActive = false
	if err := s.cropVarietyRepo.Update(ctx, &varietyEnt); err != nil {
		return nil, fmt.Errorf("failed to delete crop variety: %w", err)
	}

	return &responses.CropVarietyResponse{
		Success:   true,
		Message:   "Crop variety deleted successfully",
		RequestID: deleteReq.RequestID,
		Data:      nil,
	}, nil
}

// ListCropVarieties implements CropService.ListCropVarieties
func (s *CropServiceImpl) ListCropVarieties(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListCropVarietiesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "crop_variety", "read", "", listReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := crop.CropVarietyFilters{
		CropID:   listReq.CropID,
		IsActive: listReq.IsActive,
		Search:   listReq.Search,
		Page:     listReq.Page,
		PageSize: listReq.PageSize,
	}

	// Get varieties with crop info
	varieties, total, err := s.cropVarietyRepo.GetVarietiesWithCropInfo(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list crop varieties: %w", err)
	}

	// Convert to response data
	var varietyDataList []*responses.CropVarietyData
	for _, varietyWithCrop := range varieties {
		varietyData := &responses.CropVarietyData{
			ID:                 varietyWithCrop.ID,
			CropID:             varietyWithCrop.CropID,
			CropName:           varietyWithCrop.CropName,
			Name:               varietyWithCrop.Name,
			Description:        varietyWithCrop.Description,
			DurationDays: varietyWithCrop.DurationDays,
			YieldPerAcre: varietyWithCrop.YieldPerAcre,
			YieldPerTree: varietyWithCrop.YieldPerTree,
			Properties:   varietyWithCrop.Properties,
			IsActive:     varietyWithCrop.IsActive,
			CreatedAt:          varietyWithCrop.CreatedAt,
			UpdatedAt:          varietyWithCrop.UpdatedAt,
		}
		varietyDataList = append(varietyDataList, varietyData)
	}

	return &responses.CropVarietyListResponse{
		Success:   true,
		Message:   "Crop varieties retrieved successfully",
		RequestID: listReq.RequestID,
		Data:      varietyDataList,
		Page:      listReq.Page,
		PageSize:  listReq.PageSize,
		Total:     total,
	}, nil
}

// GetCropLookupData implements CropService.GetCropLookupData
func (s *CropServiceImpl) GetCropLookupData(ctx context.Context, req interface{}) (interface{}, error) {
	lookupReq, ok := req.(*requests.GetCropLookupRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Get crops for lookup
	crops, err := s.cropRepo.GetActiveCropsForLookup(ctx, lookupReq.Category, lookupReq.Season)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop lookup data: %w", err)
	}

	// Convert to response data
	var cropLookupDataList []*responses.CropLookupData
	for _, cropLookup := range crops {
		cropLookupData := &responses.CropLookupData{
			ID:       cropLookup.ID,
			Name:     cropLookup.Name,
			Category: cropLookup.Category,
			Seasons:  cropLookup.Seasons,
			Unit:     cropLookup.Unit,
		}
		cropLookupDataList = append(cropLookupDataList, cropLookupData)
	}

	return &responses.CropLookupResponse{
		Success:   true,
		Message:   "Crop lookup data retrieved successfully",
		RequestID: lookupReq.RequestID,
		Data:      cropLookupDataList,
	}, nil
}

// GetVarietyLookupData implements CropService.GetVarietyLookupData
func (s *CropServiceImpl) GetVarietyLookupData(ctx context.Context, req interface{}) (interface{}, error) {
	lookupReq, ok := req.(*requests.GetVarietyLookupRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Get varieties for lookup
	varieties, err := s.cropVarietyRepo.GetActiveVarietiesForLookup(ctx, lookupReq.CropID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variety lookup data: %w", err)
	}

	// Convert to response data
	var varietyLookupDataList []*responses.CropVarietyLookupData
	for _, varietyLookup := range varieties {
		varietyLookupData := &responses.CropVarietyLookupData{
			ID:           varietyLookup.ID,
			Name:         varietyLookup.Name,
			DurationDays: varietyLookup.DurationDays,
		}
		varietyLookupDataList = append(varietyLookupDataList, varietyLookupData)
	}

	return &responses.CropVarietyLookupResponse{
		Success:   true,
		Message:   "Variety lookup data retrieved successfully",
		RequestID: lookupReq.RequestID,
		Data:      varietyLookupDataList,
	}, nil
}

// GetCropCategories implements CropService.GetCropCategories
func (s *CropServiceImpl) GetCropCategories(ctx context.Context) (interface{}, error) {
	categories := cropEntity.GetValidCategories()
	var categoryStrings []string
	for _, category := range categories {
		categoryStrings = append(categoryStrings, string(category))
	}

	return &responses.CropCategoriesResponse{
		Success: true,
		Message: "Crop categories retrieved successfully",
		Data:    categoryStrings,
	}, nil
}

// GetCropSeasons implements CropService.GetCropSeasons
func (s *CropServiceImpl) GetCropSeasons(ctx context.Context) (interface{}, error) {
	seasons := cropEntity.GetValidSeasons()
	var seasonStrings []string
	for _, season := range seasons {
		seasonStrings = append(seasonStrings, string(season))
	}

	return &responses.CropSeasonsResponse{
		Success: true,
		Message: "Crop seasons retrieved successfully",
		Data:    seasonStrings,
	}, nil
}

// SeedInitialCropData implements CropService.SeedInitialCropData
func (s *CropServiceImpl) SeedInitialCropData(ctx context.Context) error {
	// Define crop data with varieties
	cropData := []struct {
		name        string
		category    cropEntity.CropCategory
		seasons     []string
		description string
		varieties   []struct {
			name        string
			duration    int
			description string
		}
	}{
		{
			name:        "Rice",
			category:    cropEntity.CropCategoryCereals,
			seasons:     []string{string(cropEntity.SeasonKharif), string(cropEntity.SeasonRabi)},
			description: "Staple cereal crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Basmati", 120, "Premium aromatic rice variety"},
				{"IR64", 115, "High yielding non-basmati variety"},
				{"Sona Masuri", 110, "Medium grain variety"},
			},
		},
		{
			name:        "Wheat",
			category:    cropEntity.CropCategoryCereals,
			seasons:     []string{string(cropEntity.SeasonRabi)},
			description: "Major cereal crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"HD2967", 140, "High yielding wheat variety"},
				{"PBW343", 145, "Popular wheat variety in Punjab"},
				{"DBW88", 135, "Disease resistant variety"},
			},
		},
		{
			name:        "Maize",
			category:    cropEntity.CropCategoryCereals,
			seasons:     []string{string(cropEntity.SeasonKharif), string(cropEntity.SeasonRabi)},
			description: "Versatile cereal crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"NK6240", 90, "High yielding hybrid"},
				{"P3396", 95, "Premium hybrid variety"},
			},
		},
		{
			name:        "Cotton",
			category:    cropEntity.CropCategoryCashCrops,
			seasons:     []string{string(cropEntity.SeasonKharif)},
			description: "Major cash crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Bt Cotton", 180, "Genetically modified variety"},
				{"Desi Cotton", 160, "Traditional variety"},
			},
		},
		{
			name:        "Sugarcane",
			category:    cropEntity.CropCategoryCashCrops,
			seasons:     []string{string(cropEntity.SeasonKharif), string(cropEntity.SeasonRabi)},
			description: "Sugar producing crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Co 86032", 365, "High sugar content variety"},
				{"Co 238", 350, "Disease resistant variety"},
			},
		},
		{
			name:        "Tomato",
			category:    cropEntity.CropCategoryVegetables,
			seasons:     []string{string(cropEntity.SeasonKharif), string(cropEntity.SeasonRabi), string(cropEntity.SeasonZaid)},
			description: "Popular vegetable crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Pusa Ruby", 90, "Determinate variety"},
				{"Arka Vikas", 85, "High yielding variety"},
			},
		},
		{
			name:        "Onion",
			category:    cropEntity.CropCategoryVegetables,
			seasons:     []string{string(cropEntity.SeasonKharif), string(cropEntity.SeasonRabi)},
			description: "Essential vegetable crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Nashik Red", 120, "Storage type variety"},
				{"Pusa White Flat", 110, "White variety"},
			},
		},
		{
			name:        "Soybean",
			category:    cropEntity.CropCategoryPulses,
			seasons:     []string{string(cropEntity.SeasonKharif)},
			description: "Protein rich legume",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"JS335", 95, "Popular variety"},
				{"MACS1407", 100, "High yielding variety"},
			},
		},
		{
			name:        "Chickpea",
			category:    cropEntity.CropCategoryPulses,
			seasons:     []string{string(cropEntity.SeasonRabi)},
			description: "Important pulse crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Kabuli", 120, "Large seeded variety"},
				{"Desi", 110, "Small seeded variety"},
			},
		},
		{
			name:        "Mustard",
			category:    cropEntity.CropCategoryOilSeeds,
			seasons:     []string{string(cropEntity.SeasonRabi)},
			description: "Oilseed crop",
			varieties: []struct {
				name        string
				duration    int
				description string
			}{
				{"Pusa Bold", 135, "High oil content"},
				{"Kranti", 140, "Disease resistant"},
			},
		},
	}

	// Seed crops and varieties
	for _, cropInfo := range cropData {
		// Check if crop already exists
		existingCrop, err := s.cropRepo.FindByName(ctx, cropInfo.name)
		if err == nil && existingCrop != nil {
			// Crop exists, skip
			continue
		}

		// Create crop
		crop := &cropEntity.Crop{
			Name:        cropInfo.name,
			Category:    cropInfo.category,
			Seasons:     cropInfo.seasons,
			IsActive:    true,
		}

		// Save crop
		if err := s.cropRepo.Create(ctx, crop); err != nil {
			return fmt.Errorf("failed to create crop %s: %w", cropInfo.name, err)
		}

		// Create varieties for this crop
		for _, varietyInfo := range cropInfo.varieties {
			variety := &cropVarietyEntity.CropVariety{
				CropID:       crop.ID,
				Name:         varietyInfo.name,
				DurationDays: &varietyInfo.duration,
				Description:  &varietyInfo.description,
				IsActive:     true,
			}

			if err := s.cropVarietyRepo.Create(ctx, variety); err != nil {
				return fmt.Errorf("failed to create variety %s for crop %s: %w",
					varietyInfo.name, cropInfo.name, err)
			}
		}
	}

	return nil
}