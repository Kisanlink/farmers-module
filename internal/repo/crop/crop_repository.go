package crop

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropRepository provides data access methods for crops
type CropRepository struct {
	*base.BaseFilterableRepository[*crop.Crop]
}

// NewCropRepository creates a new crop repository using BaseFilterableRepository
func NewCropRepository(dbManager interface{}) *CropRepository {
	baseRepo := base.NewBaseFilterableRepository[*crop.Crop]()
	baseRepo.SetDBManager(dbManager)

	return &CropRepository{
		BaseFilterableRepository: baseRepo,
	}
}

// FindByName finds a crop by its name
func (r *CropRepository) FindByName(ctx context.Context, name string) (*crop.Crop, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Where("is_active", base.OpEqual, true).
		Build()

	cropEntity, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	return cropEntity, nil
}

// FindByCategory finds crops by category
func (r *CropRepository) FindByCategory(ctx context.Context, category crop.CropCategory, page, pageSize int) ([]*crop.Crop, int, error) {
	filter := base.NewFilterBuilder().
		Where("category", base.OpEqual, category).
		Where("is_active", base.OpEqual, true).
		Build()

	filter.Page = page
	filter.PageSize = pageSize

	crops, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find crops by category: %w", err)
	}

	count, err := r.BaseFilterableRepository.Count(ctx, filter, &crop.Crop{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count crops: %w", err)
	}

	return crops, int(count), nil
}

// FindBySeason finds crops that can be grown in a specific season
// Note: Season filtering requires JSONB support. Returns all crops and filters in-memory.
// TODO: Add JSONB operator support to kisanlink-db for database-level filtering
func (r *CropRepository) FindBySeason(ctx context.Context, season string, page, pageSize int) ([]*crop.Crop, int, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Build()

	// Get all active crops (no pagination yet)
	crops, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find crops: %w", err)
	}

	// Filter by season in-memory
	filtered := make([]*crop.Crop, 0)
	for _, c := range crops {
		if containsSeason(c.Seasons, season) {
			filtered = append(filtered, c)
		}
	}

	total := len(filtered)

	// Apply pagination manually since we filtered in-memory
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*crop.Crop{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// Search finds crops by name or scientific name
// Note: LIKE search requires custom SQL. Returns all crops and filters in-memory.
// TODO: Add LIKE operator support to kisanlink-db for database-level search
func (r *CropRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*crop.Crop, int, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Build()

	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	// Get all active crops (no pagination yet)
	crops, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find crops: %w", err)
	}

	// Filter by search term in-memory
	filtered := filterBySearch(crops, searchTerm)

	total := len(filtered)

	// Apply pagination manually since we filtered in-memory
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*crop.Crop{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// FindWithFilters finds crops with multiple filters
func (r *CropRepository) FindWithFilters(ctx context.Context, filters CropFilters) ([]*crop.Crop, int, error) {
	filterBuilder := base.NewFilterBuilder()

	// Apply category filter
	if filters.Category != "" {
		filterBuilder = filterBuilder.Where("category", base.OpEqual, filters.Category)
	}

	// Apply is_active filter
	if filters.IsActive != nil {
		filterBuilder = filterBuilder.Where("is_active", base.OpEqual, *filters.IsActive)
	} else {
		filterBuilder = filterBuilder.Where("is_active", base.OpEqual, true)
	}

	filter := filterBuilder.Build()
	filter.Page = filters.Page
	filter.PageSize = filters.PageSize
	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	crops, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find crops: %w", err)
	}

	// Apply in-memory filters for season and search
	if filters.Season != "" {
		crops = filterBySeason(crops, filters.Season)
	}

	if len(filters.Seasons) > 0 {
		crops = filterBySeasons(crops, filters.Seasons)
	}

	if filters.Search != "" {
		crops = filterBySearch(crops, filters.Search)
	}

	total := len(crops)
	return crops, total, nil
}

// GetCropWithVarietyCount gets crop with count of active varieties
// Note: Variety count aggregation needs to be done in service layer
func (r *CropRepository) GetCropWithVarietyCount(ctx context.Context, cropID string) (*CropWithVarietyCount, error) {
	filter := base.NewFilterBuilder().
		Where("id", base.OpEqual, cropID).
		Where("is_active", base.OpEqual, true).
		Build()

	cropEntity, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop: %w", err)
	}

	// TODO: Move variety count logic to service layer
	result := &CropWithVarietyCount{
		Crop:         *cropEntity,
		VarietyCount: 0, // Service layer should populate this
	}

	return result, nil
}

// GetActiveCropsForLookup gets simplified crop data for dropdown/lookup
func (r *CropRepository) GetActiveCropsForLookup(ctx context.Context, category, season string) ([]*crop.Crop, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true)

	if category != "" {
		filterBuilder = filterBuilder.Where("category", base.OpEqual, category)
	}

	filter := filterBuilder.Build()
	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	crops, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get crops for lookup: %w", err)
	}

	// Filter by season in-memory if provided
	if season != "" {
		crops = filterBySeason(crops, season)
	}

	return crops, nil
}

// CropFilters represents filters for crop queries
type CropFilters struct {
	Category string
	Season   string
	Seasons  []string
	IsActive *bool
	Search   string
	Page     int
	PageSize int
}

// CropWithVarietyCount represents a crop with variety count
type CropWithVarietyCount struct {
	crop.Crop
	VarietyCount int `json:"variety_count"`
}

// CropLookup represents simplified crop data for lookups
type CropLookup struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Seasons  []string `json:"seasons"`
	Unit     string   `json:"unit"`
}
