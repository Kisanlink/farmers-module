package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// StageRepository provides data access methods for stages
type StageRepository struct {
	*base.BaseFilterableRepository[*stage.Stage]
}

// NewStageRepository creates a new stage repository using BaseFilterableRepository
func NewStageRepository(dbManager interface{}) *StageRepository {
	baseRepo := base.NewBaseFilterableRepository[*stage.Stage]()
	baseRepo.SetDBManager(dbManager)

	return &StageRepository{
		BaseFilterableRepository: baseRepo,
	}
}

// FindByName finds a stage by its name (case-insensitive)
// Note: For true case-insensitive matching, consider normalizing stage names at creation
func (r *StageRepository) FindByName(ctx context.Context, name string) (*stage.Stage, error) {
	filter := base.NewFilterBuilder().
		Where("stage_name", base.OpEqual, name).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	stg, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	return stg, nil
}

// Search finds stages by name or description
// Note: This is a simplified search. For LIKE/pattern matching, consider using full-text search
func (r *StageRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*stage.Stage, int, error) {
	// For now, we'll do exact match on stage_name
	// TODO: Implement proper LIKE search support in kisanlink-db or use full-text search
	filter := base.NewFilterBuilder().
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	filter.Page = page
	filter.PageSize = pageSize
	filter.Sort = []base.SortField{{Field: "stage_name", Direction: "asc"}}

	stages, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Filter results by search term in-memory (temporary solution)
	var filtered []*stage.Stage
	searchLower := strings.ToLower(searchTerm)
	for _, stg := range stages {
		if strings.Contains(strings.ToLower(stg.StageName), searchLower) ||
			(stg.Description != nil && strings.Contains(strings.ToLower(*stg.Description), searchLower)) {
			filtered = append(filtered, stg)
		}
	}

	return filtered, len(filtered), nil
}

// GetActiveStagesForLookup gets simplified stage data for dropdown/lookup
func (r *StageRepository) GetActiveStagesForLookup(ctx context.Context) ([]*stage.Stage, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	filter.Sort = []base.SortField{{Field: "stage_name", Direction: "asc"}}

	stages, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active stages: %w", err)
	}

	return stages, nil
}

// ListWithFilters lists stages with filters using BaseFilterableRepository
func (r *StageRepository) ListWithFilters(ctx context.Context, filters StageFilters) ([]*stage.Stage, int, error) {
	// Build filter using FilterBuilder
	filterBuilder := base.NewFilterBuilder()

	// Apply is_active filter
	if filters.IsActive != nil {
		filterBuilder = filterBuilder.Where("is_active", base.OpEqual, *filters.IsActive)
	} else {
		filterBuilder = filterBuilder.Where("is_active", base.OpEqual, true)
	}

	// Apply soft delete filter
	filterBuilder = filterBuilder.Where("deleted_at", base.OpIsNull, nil)

	// Build final filter
	filter := filterBuilder.Build()
	filter.Page = filters.Page
	filter.PageSize = filters.PageSize
	filter.Sort = []base.SortField{
		{Field: "stage_name", Direction: "asc"},
	}

	// Use BaseFilterableRepository to find stages
	stages, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find stages: %w", err)
	}

	// Apply search filter in-memory if provided
	// TODO: Move LIKE search to database level when kisanlink-db supports it
	if filters.Search != "" {
		stages = r.filterBySearch(stages, filters.Search)
	}

	// Get total count
	count, err := r.BaseFilterableRepository.Count(ctx, filter, &stage.Stage{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count stages: %w", err)
	}

	// Adjust count if search filter was applied
	if filters.Search != "" {
		count = int64(len(stages))
	}

	return stages, int(count), nil
}

// filterBySearch filters stages by search term in-memory
func (r *StageRepository) filterBySearch(stages []*stage.Stage, searchTerm string) []*stage.Stage {
	var filtered []*stage.Stage
	searchLower := strings.ToLower(searchTerm)

	for _, stg := range stages {
		if strings.Contains(strings.ToLower(stg.StageName), searchLower) ||
			(stg.Description != nil && strings.Contains(strings.ToLower(*stg.Description), searchLower)) {
			filtered = append(filtered, stg)
		}
	}

	return filtered
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
