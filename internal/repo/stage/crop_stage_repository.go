package stage

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropStageRepository provides data access methods for crop stages
type CropStageRepository struct {
	*base.BaseFilterableRepository[*stage.CropStage]
}

// NewCropStageRepository creates a new crop stage repository using BaseFilterableRepository
func NewCropStageRepository(dbManager interface{}) *CropStageRepository {
	baseRepo := base.NewBaseFilterableRepository[*stage.CropStage]()
	baseRepo.SetDBManager(dbManager)

	return &CropStageRepository{
		BaseFilterableRepository: baseRepo,
	}
}

// GetCropStages gets all stages for a crop in order with Stage relationship preloaded
func (r *CropStageRepository) GetCropStages(ctx context.Context, cropID string) ([]*stage.CropStage, error) {
	filter := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("is_active", base.OpEqual, true).
		Where("deleted_at", base.OpIsNull, nil).
		Preload("Stage").
		Sort("stage_order", "asc").
		Build()

	cropStages, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop stages: %w", err)
	}

	return cropStages, nil
}

// GetCropStageByID gets a specific crop stage with Stage relationship preloaded
func (r *CropStageRepository) GetCropStageByID(ctx context.Context, id string) (*stage.CropStage, error) {
	filter := base.NewFilterBuilder().
		Where("id", base.OpEqual, id).
		Where("deleted_at", base.OpIsNull, nil).
		Preload("Stage").
		Build()

	cropStage, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop stage: %w", err)
	}

	return cropStage, nil
}

// GetCropStageByCropAndStage gets a crop stage by crop and stage IDs with Stage relationship preloaded
func (r *CropStageRepository) GetCropStageByCropAndStage(ctx context.Context, cropID, stageID string) (*stage.CropStage, error) {
	filter := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("stage_id", base.OpEqual, stageID).
		Where("deleted_at", base.OpIsNull, nil).
		Preload("Stage").
		Build()

	cropStage, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop stage: %w", err)
	}

	return cropStage, nil
}

// CheckCropStageExists checks if a crop-stage combination exists
func (r *CropStageRepository) CheckCropStageExists(ctx context.Context, cropID, stageID string, excludeID ...string) (bool, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("stage_id", base.OpEqual, stageID).
		Where("deleted_at", base.OpIsNull, nil)

	if len(excludeID) > 0 && excludeID[0] != "" {
		filterBuilder = filterBuilder.Where("id", base.OpNotEqual, excludeID[0])
	}

	filter := filterBuilder.Build()

	count, err := r.BaseFilterableRepository.Count(ctx, filter, &stage.CropStage{})
	if err != nil {
		return false, fmt.Errorf("failed to check crop stage exists: %w", err)
	}

	return count > 0, nil
}

// CheckStageOrderExists checks if a stage order already exists for a crop
func (r *CropStageRepository) CheckStageOrderExists(ctx context.Context, cropID string, stageOrder int, excludeID ...string) (bool, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("stage_order", base.OpEqual, stageOrder).
		Where("deleted_at", base.OpIsNull, nil)

	if len(excludeID) > 0 && excludeID[0] != "" {
		filterBuilder = filterBuilder.Where("id", base.OpNotEqual, excludeID[0])
	}

	filter := filterBuilder.Build()

	count, err := r.BaseFilterableRepository.Count(ctx, filter, &stage.CropStage{})
	if err != nil {
		return false, fmt.Errorf("failed to check stage order exists: %w", err)
	}

	return count > 0, nil
}

// GetMaxStageOrder gets the maximum stage order for a crop
func (r *CropStageRepository) GetMaxStageOrder(ctx context.Context, cropID string) (int, error) {
	filter := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	// Get all crop stages for this crop
	cropStages, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get max stage order: %w", err)
	}

	// Find max order in-memory
	maxOrder := 0
	for _, cs := range cropStages {
		if cs.StageOrder > maxOrder {
			maxOrder = cs.StageOrder
		}
	}

	return maxOrder, nil
}

// ReorderStages updates stage orders for a crop
func (r *CropStageRepository) ReorderStages(ctx context.Context, cropID string, stageOrders map[string]int) error {
	// Fetch all crop stages that need to be updated
	for stageID, order := range stageOrders {
		filter := base.NewFilterBuilder().
			Where("crop_id", base.OpEqual, cropID).
			Where("stage_id", base.OpEqual, stageID).
			Where("deleted_at", base.OpIsNull, nil).
			Build()

		cropStage, err := r.BaseFilterableRepository.FindOne(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to find crop stage for reorder (stage_id=%s): %w", stageID, err)
		}

		// Update the stage order
		cropStage.StageOrder = order
		if err := r.BaseFilterableRepository.Update(ctx, cropStage); err != nil {
			return fmt.Errorf("failed to update stage order for stage %s: %w", stageID, err)
		}
	}

	return nil
}
