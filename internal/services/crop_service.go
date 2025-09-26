package services

import (
	"context"
	"fmt"

	cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
	cropStageEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_stage"
	cropVarietyEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropService handles crop master data operations
type CropService interface {
	// Crop operations
	CreateCrop(ctx context.Context, req *requests.CreateCropRequest) (*responses.CropResponse, error)
	ListCrops(ctx context.Context, req *requests.ListCropsRequest) (*responses.CropsListResponse, error)
	GetCrop(ctx context.Context, cropID string) (*responses.CropResponse, error)
	UpdateCrop(ctx context.Context, req *requests.UpdateCropRequest) (*responses.CropResponse, error)
	DeleteCrop(ctx context.Context, cropID string) error

	// Variety operations
	CreateVariety(ctx context.Context, req *requests.CreateVarietyRequest) (*responses.CropVarietyResponse, error)
	ListVarietiesByCrop(ctx context.Context, cropID string) (*responses.CropVarietiesListResponse, error)
	GetVariety(ctx context.Context, varietyID string) (*responses.CropVarietyResponse, error)
	UpdateVariety(ctx context.Context, req *requests.UpdateVarietyRequest) (*responses.CropVarietyResponse, error)
	DeleteVariety(ctx context.Context, varietyID string) error

	// Stage operations
	CreateStage(ctx context.Context, req *requests.CreateStageRequest) (*responses.CropStageResponse, error)
	ListStagesByCrop(ctx context.Context, cropID string) (*responses.CropStagesListResponse, error)
	GetStage(ctx context.Context, stageID string) (*responses.CropStageResponse, error)
	UpdateStage(ctx context.Context, req *requests.UpdateStageRequest) (*responses.CropStageResponse, error)
	DeleteStage(ctx context.Context, stageID string) error

	// Lookup operations
	GetLookupData(ctx context.Context, lookupType string) (*responses.LookupResponse, error)
}

// CropServiceImpl implements CropService
type CropServiceImpl struct {
	cropRepo    *base.BaseFilterableRepository[*cropEntity.Crop]
	varietyRepo *base.BaseFilterableRepository[*cropVarietyEntity.CropVariety]
	stageRepo   *base.BaseFilterableRepository[*cropStageEntity.CropStage]
	aaaService  AAAService
}

// NewCropService creates a new crop service
func NewCropService(
	cropRepo *base.BaseFilterableRepository[*cropEntity.Crop],
	varietyRepo *base.BaseFilterableRepository[*cropVarietyEntity.CropVariety],
	stageRepo *base.BaseFilterableRepository[*cropStageEntity.CropStage],
	aaaService AAAService,
) CropService {
	return &CropServiceImpl{
		cropRepo:    cropRepo,
		varietyRepo: varietyRepo,
		stageRepo:   stageRepo,
		aaaService:  aaaService,
	}
}

// CreateCrop creates a new crop
func (s *CropServiceImpl) CreateCrop(ctx context.Context, req *requests.CreateCropRequest) (*responses.CropResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "create", "", req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Create crop entity
	crop := cropEntity.NewCrop()
	crop.Name = req.Name
	crop.Category = req.Category
	crop.CropDurationDays = req.CropDurationDays
	crop.TypicalUnits = req.TypicalUnits
	crop.Seasons = req.Seasons
	crop.ImageURL = req.ImageURL
	crop.DocumentID = req.DocumentID
	crop.Metadata = req.Metadata

	// Validate
	if err := crop.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.cropRepo.Create(ctx, crop); err != nil {
		return nil, fmt.Errorf("failed to create crop: %w", err)
	}

	return responses.NewCropResponse(crop, "Crop created successfully"), nil
}

// ListCrops lists crops with optional filtering
func (s *CropServiceImpl) ListCrops(ctx context.Context, req *requests.ListCropsRequest) (*responses.CropsListResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "list", "", req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filter
	filter := base.NewFilterBuilder()

	if req.Category != nil {
		filter.Where("category", base.OpEqual, string(*req.Category))
	}

	if req.Season != nil {
		filter.Where("seasons", base.OpContains, string(*req.Season))
	}

	if req.Search != nil {
		filter.Where("name", base.OpLike, "%"+*req.Search+"%")
	}

	// Set pagination
	limit := 50 // default
	offset := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	filter.Limit(limit, offset)

	// Query database
	crops, err := s.cropRepo.Find(ctx, filter.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to list crops: %w", err)
	}

	return responses.NewCropsListResponse(crops, "Crops retrieved successfully"), nil
}

// GetCrop gets a specific crop by ID
func (s *CropServiceImpl) GetCrop(ctx context.Context, cropID string) (*responses.CropResponse, error) {
	// Note: FindByID not implemented in base repository
	return nil, fmt.Errorf("FindByID not implemented in base repository")
}

// UpdateCrop updates an existing crop
func (s *CropServiceImpl) UpdateCrop(ctx context.Context, req *requests.UpdateCropRequest) (*responses.CropResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "update", req.CropID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Note: FindByID not implemented in base repository
	return nil, fmt.Errorf("FindByID not implemented in base repository")
}

// DeleteCrop deletes a crop
func (s *CropServiceImpl) DeleteCrop(ctx context.Context, cropID string) error {
	// Get crop to check if it exists
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, cropID).Build()
	crops, err := s.cropRepo.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get crop: %w", err)
	}
	if len(crops) == 0 {
		return common.ErrNotFound
	}
	crop := crops[0]

	// Delete from database
	if err := s.cropRepo.Delete(ctx, crop.ID, crop); err != nil {
		return fmt.Errorf("failed to delete crop: %w", err)
	}
	return nil
}

// CreateVariety creates a new crop variety
func (s *CropServiceImpl) CreateVariety(ctx context.Context, req *requests.CreateVarietyRequest) (*responses.CropVarietyResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "create_variety", req.CropID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Verify crop exists
	cropFilter := base.NewFilterBuilder().Where("id", base.OpEqual, req.CropID).Build()
	crops, err := s.cropRepo.Find(ctx, cropFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to verify crop: %w", err)
	}
	if len(crops) == 0 {
		return nil, fmt.Errorf("crop not found")
	}

	// Create variety entity
	variety := cropVarietyEntity.NewCropVariety()
	variety.CropID = req.CropID
	variety.VarietyName = req.VarietyName
	variety.DurationDays = req.DurationDays
	variety.Characteristics = req.Characteristics
	variety.Metadata = req.Metadata

	// Validate
	if err := variety.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.varietyRepo.Create(ctx, variety); err != nil {
		return nil, fmt.Errorf("failed to create variety: %w", err)
	}

	return responses.NewCropVarietyResponse(variety, "Crop variety created successfully"), nil
}

// ListVarietiesByCrop lists varieties for a specific crop
func (s *CropServiceImpl) ListVarietiesByCrop(ctx context.Context, cropID string) (*responses.CropVarietiesListResponse, error) {
	// Build filter
	filter := base.NewFilterBuilder().Where("crop_id", base.OpEqual, cropID)

	// Query database
	varieties, err := s.varietyRepo.Find(ctx, filter.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to list varieties: %w", err)
	}

	return responses.NewCropVarietiesListResponse(varieties, "Crop varieties retrieved successfully"), nil
}

// GetVariety gets a specific variety by ID
func (s *CropServiceImpl) GetVariety(ctx context.Context, varietyID string) (*responses.CropVarietyResponse, error) {
	// Get variety from database
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, varietyID).Build()
	varieties, err := s.varietyRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get variety: %w", err)
	}
	if len(varieties) == 0 {
		return nil, common.ErrNotFound
	}
	variety := varieties[0]

	return responses.NewCropVarietyResponse(variety, "Crop variety retrieved successfully"), nil
}

// UpdateVariety updates an existing variety
func (s *CropServiceImpl) UpdateVariety(ctx context.Context, req *requests.UpdateVarietyRequest) (*responses.CropVarietyResponse, error) {
	// Get existing variety
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, req.VarietyID).Build()
	varieties, err := s.varietyRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get variety: %w", err)
	}
	if len(varieties) == 0 {
		return nil, common.ErrNotFound
	}
	variety := varieties[0]

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "update_variety", variety.CropID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Update fields
	if req.VarietyName != nil {
		variety.VarietyName = *req.VarietyName
	}
	if req.DurationDays != nil {
		variety.DurationDays = req.DurationDays
	}
	if req.Characteristics != nil {
		variety.Characteristics = req.Characteristics
	}
	if req.Metadata != nil {
		variety.Metadata = req.Metadata
	}

	// Validate
	if err := variety.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.varietyRepo.Update(ctx, variety); err != nil {
		return nil, fmt.Errorf("failed to update variety: %w", err)
	}

	return responses.NewCropVarietyResponse(variety, "Crop variety updated successfully"), nil
}

// DeleteVariety deletes a variety
func (s *CropServiceImpl) DeleteVariety(ctx context.Context, varietyID string) error {
	// Get variety to check if it exists
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, varietyID).Build()
	varieties, err := s.varietyRepo.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get variety: %w", err)
	}
	if len(varieties) == 0 {
		return common.ErrNotFound
	}
	variety := varieties[0]

	// Delete from database
	if err := s.varietyRepo.Delete(ctx, variety.ID, variety); err != nil {
		return fmt.Errorf("failed to delete variety: %w", err)
	}

	return nil
}

// CreateStage creates a new crop stage
func (s *CropServiceImpl) CreateStage(ctx context.Context, req *requests.CreateStageRequest) (*responses.CropStageResponse, error) {
	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "create_stage", req.CropID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Verify crop exists
	cropFilter := base.NewFilterBuilder().Where("id", base.OpEqual, req.CropID).Build()
	crops, err := s.cropRepo.Find(ctx, cropFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to verify crop: %w", err)
	}
	if len(crops) == 0 {
		return nil, fmt.Errorf("crop not found")
	}

	// Create stage entity
	stage := cropStageEntity.NewCropStage()
	stage.CropID = req.CropID
	stage.StageName = req.StageName
	stage.StageOrder = req.StageOrder
	stage.TypicalDurationDays = req.TypicalDurationDays
	stage.Description = req.Description
	stage.Metadata = req.Metadata

	// Validate
	if err := stage.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.stageRepo.Create(ctx, stage); err != nil {
		return nil, fmt.Errorf("failed to create stage: %w", err)
	}

	return responses.NewCropStageResponse(stage, "Crop stage created successfully"), nil
}

// ListStagesByCrop lists stages for a specific crop
func (s *CropServiceImpl) ListStagesByCrop(ctx context.Context, cropID string) (*responses.CropStagesListResponse, error) {
	// Build filter
	filter := base.NewFilterBuilder().Where("crop_id", base.OpEqual, cropID)

	// Query database
	stages, err := s.stageRepo.Find(ctx, filter.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to list stages: %w", err)
	}

	return responses.NewCropStagesListResponse(stages, "Crop stages retrieved successfully"), nil
}

// GetStage gets a specific stage by ID
func (s *CropServiceImpl) GetStage(ctx context.Context, stageID string) (*responses.CropStageResponse, error) {
	// Get stage from database
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, stageID).Build()
	stages, err := s.stageRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}
	if len(stages) == 0 {
		return nil, common.ErrNotFound
	}
	stage := stages[0]

	return responses.NewCropStageResponse(stage, "Crop stage retrieved successfully"), nil
}

// UpdateStage updates an existing stage
func (s *CropServiceImpl) UpdateStage(ctx context.Context, req *requests.UpdateStageRequest) (*responses.CropStageResponse, error) {
	// Get existing stage
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, req.StageID).Build()
	stages, err := s.stageRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}
	if len(stages) == 0 {
		return nil, common.ErrNotFound
	}
	stage := stages[0]

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, req.UserID, "crop", "update_stage", stage.CropID, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Update fields
	if req.StageName != nil {
		stage.StageName = *req.StageName
	}
	if req.StageOrder != nil {
		stage.StageOrder = *req.StageOrder
	}
	if req.TypicalDurationDays != nil {
		stage.TypicalDurationDays = req.TypicalDurationDays
	}
	if req.Description != nil {
		stage.Description = req.Description
	}
	if req.Metadata != nil {
		stage.Metadata = req.Metadata
	}

	// Validate
	if err := stage.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.stageRepo.Update(ctx, stage); err != nil {
		return nil, fmt.Errorf("failed to update stage: %w", err)
	}

	return responses.NewCropStageResponse(stage, "Crop stage updated successfully"), nil
}

// DeleteStage deletes a stage
func (s *CropServiceImpl) DeleteStage(ctx context.Context, stageID string) error {
	// Get stage to check if it exists
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, stageID).Build()
	stages, err := s.stageRepo.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get stage: %w", err)
	}
	if len(stages) == 0 {
		return common.ErrNotFound
	}
	stage := stages[0]

	// Delete from database
	if err := s.stageRepo.Delete(ctx, stage.ID, stage); err != nil {
		return fmt.Errorf("failed to delete stage: %w", err)
	}

	return nil
}

// GetLookupData gets lookup data for dropdowns
func (s *CropServiceImpl) GetLookupData(ctx context.Context, lookupType string) (*responses.LookupResponse, error) {
	var items []string

	switch lookupType {
	case "categories":
		for _, category := range cropEntity.GetValidCategories() {
			items = append(items, string(category))
		}
	case "units":
		for _, unit := range cropEntity.GetValidUnits() {
			items = append(items, string(unit))
		}
	case "seasons":
		for _, season := range cropEntity.GetValidSeasons() {
			items = append(items, string(season))
		}
	default:
		return nil, fmt.Errorf("invalid lookup type: %s", lookupType)
	}

	return responses.NewLookupResponse(lookupType, items, "Lookup data retrieved successfully"), nil
}
