package stage

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// CropStageRepository provides data access methods for crop stages
type CropStageRepository struct {
	*base.BaseFilterableRepository[*stage.CropStage]
	db *gorm.DB
}

// NewCropStageRepository creates a new crop stage repository
func NewCropStageRepository(db *gorm.DB) *CropStageRepository {
	baseRepo := base.NewBaseFilterableRepository[*stage.CropStage]()
	return &CropStageRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// GetCropStages gets all stages for a crop in order
func (r *CropStageRepository) GetCropStages(ctx context.Context, cropID string) ([]*stage.CropStage, error) {
	var cropStages []*stage.CropStage

	err := r.db.WithContext(ctx).
		Preload("Stage", "deleted_at IS NULL").
		Where("crop_id = ?", cropID).
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Order("stage_order ASC").
		Find(&cropStages).Error

	if err != nil {
		return nil, err
	}

	return cropStages, nil
}

// GetCropStageByID gets a specific crop stage
func (r *CropStageRepository) GetCropStageByID(ctx context.Context, id string) (*stage.CropStage, error) {
	var cropStage stage.CropStage

	err := r.db.WithContext(ctx).
		Preload("Stage").
		Preload("Crop").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&cropStage).Error

	if err != nil {
		return nil, err
	}

	return &cropStage, nil
}

// GetCropStageByCropAndStage gets a crop stage by crop and stage IDs
func (r *CropStageRepository) GetCropStageByCropAndStage(ctx context.Context, cropID, stageID string) (*stage.CropStage, error) {
	var cropStage stage.CropStage

	err := r.db.WithContext(ctx).
		Where("crop_id = ?", cropID).
		Where("stage_id = ?", stageID).
		Where("deleted_at IS NULL").
		First(&cropStage).Error

	if err != nil {
		return nil, err
	}

	return &cropStage, nil
}

// CheckCropStageExists checks if a crop-stage combination exists
func (r *CropStageRepository) CheckCropStageExists(ctx context.Context, cropID, stageID string, excludeID ...string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&stage.CropStage{}).
		Where("crop_id = ?", cropID).
		Where("stage_id = ?", stageID).
		Where("deleted_at IS NULL")

	if len(excludeID) > 0 && excludeID[0] != "" {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckStageOrderExists checks if a stage order already exists for a crop
func (r *CropStageRepository) CheckStageOrderExists(ctx context.Context, cropID string, stageOrder int, excludeID ...string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&stage.CropStage{}).
		Where("crop_id = ?", cropID).
		Where("stage_order = ?", stageOrder).
		Where("deleted_at IS NULL")

	if len(excludeID) > 0 && excludeID[0] != "" {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMaxStageOrder gets the maximum stage order for a crop
func (r *CropStageRepository) GetMaxStageOrder(ctx context.Context, cropID string) (int, error) {
	var maxOrder int
	err := r.db.WithContext(ctx).
		Model(&stage.CropStage{}).
		Select("COALESCE(MAX(stage_order), 0)").
		Where("crop_id = ?", cropID).
		Where("deleted_at IS NULL").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}

	return maxOrder, nil
}

// ReorderStages updates stage orders for a crop
func (r *CropStageRepository) ReorderStages(ctx context.Context, cropID string, stageOrders map[string]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for stageID, order := range stageOrders {
			err := tx.Model(&stage.CropStage{}).
				Where("crop_id = ?", cropID).
				Where("stage_id = ?", stageID).
				Where("deleted_at IS NULL").
				Update("stage_order", order).Error

			if err != nil {
				return fmt.Errorf("failed to update stage order for stage %s: %w", stageID, err)
			}
		}
		return nil
	})
}
