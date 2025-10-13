package stage

import (
	"context"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// StageRepository provides data access methods for stages
type StageRepository struct {
	*base.BaseFilterableRepository[*stage.Stage]
	db *gorm.DB
}

// NewStageRepository creates a new stage repository
func NewStageRepository(db *gorm.DB) *StageRepository {
	baseRepo := base.NewBaseFilterableRepository[*stage.Stage]()
	return &StageRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// FindByName finds a stage by its name (case-insensitive)
func (r *StageRepository) FindByName(ctx context.Context, name string) (*stage.Stage, error) {
	var stg stage.Stage
	err := r.db.WithContext(ctx).
		Where("LOWER(stage_name) = LOWER(?)", name).
		Where("deleted_at IS NULL").
		First(&stg).Error
	if err != nil {
		return nil, err
	}
	return &stg, nil
}

// Search finds stages by name or description
func (r *StageRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*stage.Stage, int, error) {
	var stages []*stage.Stage
	var total int64

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	query := r.db.WithContext(ctx).Model(&stage.Stage{}).
		Where("(LOWER(stage_name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern).
		Where("deleted_at IS NULL")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("stage_name ASC").Find(&stages).Error; err != nil {
		return nil, 0, err
	}

	return stages, int(total), nil
}

// GetActiveStagesForLookup gets simplified stage data for dropdown/lookup
func (r *StageRepository) GetActiveStagesForLookup(ctx context.Context) ([]*StageLookup, error) {
	var stages []*StageLookup

	err := r.db.WithContext(ctx).
		Model(&stage.Stage{}).
		Select("id, stage_name, description").
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Order("stage_name ASC").
		Find(&stages).Error

	if err != nil {
		return nil, err
	}

	return stages, nil
}

// ListWithFilters lists stages with filters
func (r *StageRepository) ListWithFilters(ctx context.Context, filters StageFilters) ([]*stage.Stage, int, error) {
	var stages []*stage.Stage
	var total int64

	query := r.db.WithContext(ctx).Model(&stage.Stage{})

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("(LOWER(stage_name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}

	query = query.Where("deleted_at IS NULL")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).Order("stage_name ASC").Find(&stages).Error; err != nil {
		return nil, 0, err
	}

	return stages, int(total), nil
}

// StageFilters represents filters for stage queries
type StageFilters struct {
	Search   string
	IsActive *bool
	Page     int
	PageSize int
}

// StageLookup represents simplified stage data for lookups
type StageLookup struct {
	ID          string  `json:"id"`
	StageName   string  `json:"stage_name"`
	Description *string `json:"description"`
}
