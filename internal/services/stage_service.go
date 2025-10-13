package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	stageEntity "github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/farmers-module/internal/repo/stage"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"gorm.io/gorm"
)

// StageServiceImpl implements StageService
type StageServiceImpl struct {
	stageRepo     *stage.StageRepository
	cropStageRepo *stage.CropStageRepository
	aaaService    AAAService
}

// NewStageService creates a new stage service
func NewStageService(
	stageRepo *stage.StageRepository,
	cropStageRepo *stage.CropStageRepository,
	aaaService AAAService,
) StageService {
	return &StageServiceImpl{
		stageRepo:     stageRepo,
		cropStageRepo: cropStageRepo,
		aaaService:    aaaService,
	}
}

// CreateStage implements StageService.CreateStage
func (s *StageServiceImpl) CreateStage(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateStageRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, createReq.UserID, "stage", "create", "", createReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Check if stage with same name already exists
	existingStage, err := s.stageRepo.FindByName(ctx, createReq.StageName)
	if err == nil && existingStage != nil {
		return nil, common.ErrAlreadyExists
	}

	// Create stage entity
	stageEnt := stageEntity.NewStage()
	stageEnt.StageName = createReq.StageName
	stageEnt.Description = createReq.Description
	if createReq.Properties != nil {
		stageEnt.Properties = createReq.Properties
	}

	// Validate the stage entity
	if err := stageEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.stageRepo.Create(ctx, stageEnt); err != nil {
		return nil, fmt.Errorf("failed to create stage: %w", err)
	}

	// Convert to response
	stageData := &responses.StageData{
		ID:          stageEnt.ID,
		StageName:   stageEnt.StageName,
		Description: stageEnt.Description,
		Properties:  stageEnt.Properties,
		IsActive:    stageEnt.IsActive,
		CreatedAt:   stageEnt.CreatedAt,
		UpdatedAt:   stageEnt.UpdatedAt,
	}

	return &responses.StageResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stage created successfully",
			RequestID: createReq.RequestID,
		},
		Data: stageData,
	}, nil
}

// GetStage implements StageService.GetStage
func (s *StageServiceImpl) GetStage(ctx context.Context, req interface{}) (interface{}, error) {
	getReq, ok := req.(*requests.GetStageRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, getReq.UserID, "stage", "read", getReq.ID, getReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get stage by ID
	var stageEnt stageEntity.Stage
	_, err = s.stageRepo.GetByID(ctx, getReq.ID, &stageEnt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}

	// Convert to response
	stageData := &responses.StageData{
		ID:          stageEnt.ID,
		StageName:   stageEnt.StageName,
		Description: stageEnt.Description,
		Properties:  stageEnt.Properties,
		IsActive:    stageEnt.IsActive,
		CreatedAt:   stageEnt.CreatedAt,
		UpdatedAt:   stageEnt.UpdatedAt,
	}

	return &responses.StageResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stage retrieved successfully",
			RequestID: getReq.RequestID,
		},
		Data: stageData,
	}, nil
}

// UpdateStage implements StageService.UpdateStage
func (s *StageServiceImpl) UpdateStage(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateStageRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, updateReq.UserID, "stage", "update", updateReq.ID, updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing stage
	var stageEnt stageEntity.Stage
	_, err = s.stageRepo.GetByID(ctx, updateReq.ID, &stageEnt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}

	// Update fields if provided
	if updateReq.StageName != nil {
		// Check if new name already exists
		existingStage, err := s.stageRepo.FindByName(ctx, *updateReq.StageName)
		if err == nil && existingStage != nil && existingStage.ID != updateReq.ID {
			return nil, common.ErrAlreadyExists
		}
		stageEnt.StageName = *updateReq.StageName
	}

	if updateReq.Description != nil {
		stageEnt.Description = updateReq.Description
	}

	if updateReq.Properties != nil {
		stageEnt.Properties = updateReq.Properties
	}

	if updateReq.IsActive != nil {
		stageEnt.IsActive = *updateReq.IsActive
	}

	// Validate the updated stage
	if err := stageEnt.Validate(); err != nil {
		return nil, err
	}

	// Save updates
	if err := s.stageRepo.Update(ctx, &stageEnt); err != nil {
		return nil, fmt.Errorf("failed to update stage: %w", err)
	}

	// Convert to response
	stageData := &responses.StageData{
		ID:          stageEnt.ID,
		StageName:   stageEnt.StageName,
		Description: stageEnt.Description,
		Properties:  stageEnt.Properties,
		IsActive:    stageEnt.IsActive,
		CreatedAt:   stageEnt.CreatedAt,
		UpdatedAt:   stageEnt.UpdatedAt,
	}

	return &responses.StageResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stage updated successfully",
			RequestID: updateReq.RequestID,
		},
		Data: stageData,
	}, nil
}

// DeleteStage implements StageService.DeleteStage
func (s *StageServiceImpl) DeleteStage(ctx context.Context, req interface{}) (interface{}, error) {
	deleteReq, ok := req.(*requests.DeleteStageRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, deleteReq.UserID, "stage", "delete", deleteReq.ID, deleteReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing stage
	var stageEnt stageEntity.Stage
	_, err = s.stageRepo.GetByID(ctx, deleteReq.ID, &stageEnt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}

	// Soft delete the stage
	if err := s.stageRepo.Delete(ctx, deleteReq.ID, &stageEnt); err != nil {
		return nil, fmt.Errorf("failed to delete stage: %w", err)
	}

	return &responses.BaseResponse{
		Success:   true,
		Message:   "Stage deleted successfully",
		RequestID: deleteReq.RequestID,
	}, nil
}

// ListStages implements StageService.ListStages
func (s *StageServiceImpl) ListStages(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListStagesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, listReq.UserID, "stage", "list", "", listReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Set default pagination values
	if listReq.Page < 1 {
		listReq.Page = 1
	}
	if listReq.PageSize < 1 {
		listReq.PageSize = 20
	}
	if listReq.PageSize > 100 {
		listReq.PageSize = 100
	}

	// Build filters
	filters := stage.StageFilters{
		Search:   listReq.Search,
		IsActive: listReq.IsActive,
		Page:     listReq.Page,
		PageSize: listReq.PageSize,
	}

	// Get stages with filters
	stages, total, err := s.stageRepo.ListWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list stages: %w", err)
	}

	// Convert to response
	stageDataList := make([]*responses.StageData, len(stages))
	for i, stageEnt := range stages {
		stageDataList[i] = &responses.StageData{
			ID:          stageEnt.ID,
			StageName:   stageEnt.StageName,
			Description: stageEnt.Description,
			Properties:  stageEnt.Properties,
			IsActive:    stageEnt.IsActive,
			CreatedAt:   stageEnt.CreatedAt,
			UpdatedAt:   stageEnt.UpdatedAt,
		}
	}

	return &responses.StageListResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stages retrieved successfully",
			RequestID: listReq.RequestID,
		},
		Data:     stageDataList,
		Page:     listReq.Page,
		PageSize: listReq.PageSize,
		Total:    total,
	}, nil
}

// GetStageLookup implements StageService.GetStageLookup
func (s *StageServiceImpl) GetStageLookup(ctx context.Context, req interface{}) (interface{}, error) {
	lookupReq, ok := req.(*requests.GetStageLookupRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, lookupReq.UserID, "stage", "list", "", lookupReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get active stages for lookup
	stages, err := s.stageRepo.GetActiveStagesForLookup(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage lookup data: %w", err)
	}

	// Convert to response
	lookupDataList := make([]*responses.StageLookupData, len(stages))
	for i, stageLookup := range stages {
		lookupDataList[i] = &responses.StageLookupData{
			ID:          stageLookup.ID,
			StageName:   stageLookup.StageName,
			Description: stageLookup.Description,
		}
	}

	return &responses.StageLookupResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stage lookup data retrieved successfully",
			RequestID: lookupReq.RequestID,
		},
		Data: lookupDataList,
	}, nil
}

// AssignStageToCrop implements StageService.AssignStageToCrop
func (s *StageServiceImpl) AssignStageToCrop(ctx context.Context, req interface{}) (interface{}, error) {
	assignReq, ok := req.(*requests.AssignStageToCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, assignReq.UserID, "crop_stage", "create", "", assignReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Check if stage exists
	var stage stageEntity.Stage
	_, err = s.stageRepo.GetByID(ctx, assignReq.StageID, &stage)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stage not found: %w", common.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}

	// Check if crop-stage combination already exists
	exists, err := s.cropStageRepo.CheckCropStageExists(ctx, assignReq.CropID, assignReq.StageID)
	if err != nil {
		return nil, fmt.Errorf("failed to check crop stage existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("stage already assigned to this crop: %w", common.ErrAlreadyExists)
	}

	// Check if stage order already exists for this crop
	orderExists, err := s.cropStageRepo.CheckStageOrderExists(ctx, assignReq.CropID, assignReq.StageOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to check stage order: %w", err)
	}
	if orderExists {
		return nil, fmt.Errorf("stage order %d already exists for this crop: %w", assignReq.StageOrder, common.ErrAlreadyExists)
	}

	// Create crop stage entity
	cropStageEnt := stageEntity.NewCropStage()
	cropStageEnt.CropID = assignReq.CropID
	cropStageEnt.StageID = assignReq.StageID
	cropStageEnt.StageOrder = assignReq.StageOrder
	cropStageEnt.DurationDays = assignReq.DurationDays

	if assignReq.DurationUnit != "" {
		cropStageEnt.DurationUnit = stageEntity.DurationUnit(assignReq.DurationUnit)
	}

	if assignReq.Properties != nil {
		cropStageEnt.Properties = assignReq.Properties
	}

	// Validate the crop stage entity
	if err := cropStageEnt.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropStageRepo.Create(ctx, cropStageEnt); err != nil {
		return nil, fmt.Errorf("failed to assign stage to crop: %w", err)
	}

	// Get the created crop stage with stage details
	cropStage, err := s.cropStageRepo.GetCropStageByID(ctx, cropStageEnt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop stage: %w", err)
	}

	// Convert to response
	cropStageData := s.convertCropStageToResponse(cropStage)

	return &responses.CropStageResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Stage assigned to crop successfully",
			RequestID: assignReq.RequestID,
		},
		Data: cropStageData,
	}, nil
}

// UpdateCropStage implements StageService.UpdateCropStage
func (s *StageServiceImpl) UpdateCropStage(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateCropStageRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, updateReq.UserID, "crop_stage", "update", "", updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing crop stage
	cropStage, err := s.cropStageRepo.GetCropStageByCropAndStage(ctx, updateReq.CropID, updateReq.StageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get crop stage: %w", err)
	}

	// Update fields if provided
	if updateReq.StageOrder != nil {
		// Check if new stage order already exists
		orderExists, err := s.cropStageRepo.CheckStageOrderExists(ctx, updateReq.CropID, *updateReq.StageOrder, cropStage.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check stage order: %w", err)
		}
		if orderExists {
			return nil, fmt.Errorf("stage order %d already exists for this crop: %w", *updateReq.StageOrder, common.ErrAlreadyExists)
		}
		cropStage.StageOrder = *updateReq.StageOrder
	}

	if updateReq.DurationDays != nil {
		cropStage.DurationDays = updateReq.DurationDays
	}

	if updateReq.DurationUnit != nil {
		cropStage.DurationUnit = stageEntity.DurationUnit(*updateReq.DurationUnit)
	}

	if updateReq.Properties != nil {
		cropStage.Properties = updateReq.Properties
	}

	if updateReq.IsActive != nil {
		cropStage.IsActive = *updateReq.IsActive
	}

	// Validate the updated crop stage
	if err := cropStage.Validate(); err != nil {
		return nil, err
	}

	// Save updates
	if err := s.cropStageRepo.Update(ctx, cropStage); err != nil {
		return nil, fmt.Errorf("failed to update crop stage: %w", err)
	}

	// Get updated crop stage with stage details
	updatedCropStage, err := s.cropStageRepo.GetCropStageByID(ctx, cropStage.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated crop stage: %w", err)
	}

	// Convert to response
	cropStageData := s.convertCropStageToResponse(updatedCropStage)

	return &responses.CropStageResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Crop stage updated successfully",
			RequestID: updateReq.RequestID,
		},
		Data: cropStageData,
	}, nil
}

// RemoveStageFromCrop implements StageService.RemoveStageFromCrop
func (s *StageServiceImpl) RemoveStageFromCrop(ctx context.Context, req interface{}) (interface{}, error) {
	removeReq, ok := req.(*requests.RemoveStageFromCropRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, removeReq.UserID, "crop_stage", "delete", "", removeReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing crop stage
	cropStage, err := s.cropStageRepo.GetCropStageByCropAndStage(ctx, removeReq.CropID, removeReq.StageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get crop stage: %w", err)
	}

	// Soft delete the crop stage
	if err := s.cropStageRepo.Delete(ctx, cropStage.ID, cropStage); err != nil {
		return nil, fmt.Errorf("failed to remove stage from crop: %w", err)
	}

	return &responses.BaseResponse{
		Success:   true,
		Message:   "Stage removed from crop successfully",
		RequestID: removeReq.RequestID,
	}, nil
}

// GetCropStages implements StageService.GetCropStages
func (s *StageServiceImpl) GetCropStages(ctx context.Context, req interface{}) (interface{}, error) {
	getReq, ok := req.(*requests.GetCropStagesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, getReq.UserID, "crop_stage", "list", "", getReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get crop stages
	cropStages, err := s.cropStageRepo.GetCropStages(ctx, getReq.CropID)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop stages: %w", err)
	}

	// Convert to response
	cropStageDataList := make([]*responses.CropStageData, len(cropStages))
	for i, cropStage := range cropStages {
		cropStageDataList[i] = s.convertCropStageToResponse(cropStage)
	}

	return &responses.CropStagesResponse{
		BaseResponse: &responses.BaseResponse{
			Success:   true,
			Message:   "Crop stages retrieved successfully",
			RequestID: getReq.RequestID,
		},
		Data: cropStageDataList,
	}, nil
}

// ReorderCropStages implements StageService.ReorderCropStages
func (s *StageServiceImpl) ReorderCropStages(ctx context.Context, req interface{}) (interface{}, error) {
	reorderReq, ok := req.(*requests.ReorderCropStagesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, reorderReq.UserID, "crop_stage", "update", "", reorderReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Validate stage orders
	if len(reorderReq.StageOrders) == 0 {
		return nil, fmt.Errorf("stage orders cannot be empty: %w", common.ErrInvalidInput)
	}

	// Validate that all stage IDs exist for this crop
	for stageID := range reorderReq.StageOrders {
		exists, err := s.cropStageRepo.CheckCropStageExists(ctx, reorderReq.CropID, stageID)
		if err != nil {
			return nil, fmt.Errorf("failed to check crop stage existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("stage %s not assigned to this crop: %w", stageID, common.ErrNotFound)
		}
	}

	// Reorder stages
	if err := s.cropStageRepo.ReorderStages(ctx, reorderReq.CropID, reorderReq.StageOrders); err != nil {
		return nil, fmt.Errorf("failed to reorder crop stages: %w", err)
	}

	return &responses.BaseResponse{
		Success:   true,
		Message:   "Crop stages reordered successfully",
		RequestID: reorderReq.RequestID,
	}, nil
}

// convertCropStageToResponse converts a crop stage entity to response data
func (s *StageServiceImpl) convertCropStageToResponse(cropStage *stageEntity.CropStage) *responses.CropStageData {
	data := &responses.CropStageData{
		ID:           cropStage.ID,
		CropID:       cropStage.CropID,
		StageID:      cropStage.StageID,
		StageOrder:   cropStage.StageOrder,
		DurationDays: cropStage.DurationDays,
		DurationUnit: string(cropStage.DurationUnit),
		Properties:   cropStage.Properties,
		IsActive:     cropStage.IsActive,
		CreatedAt:    cropStage.CreatedAt,
		UpdatedAt:    cropStage.UpdatedAt,
	}

	// Add stage details if loaded
	if cropStage.Stage != nil {
		data.StageName = cropStage.Stage.StageName
		data.Description = cropStage.Stage.Description
	}

	return data
}
