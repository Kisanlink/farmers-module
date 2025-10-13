package crop

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropVarietyRepository provides data access methods for crop varieties
type CropVarietyRepository struct {
	*base.BaseFilterableRepository[*crop_variety.CropVariety]
}

// NewCropVarietyRepository creates a new crop variety repository using BaseFilterableRepository
func NewCropVarietyRepository(dbManager interface{}) *CropVarietyRepository {
	baseRepo := base.NewBaseFilterableRepository[*crop_variety.CropVariety]()
	baseRepo.SetDBManager(dbManager)

	return &CropVarietyRepository{
		BaseFilterableRepository: baseRepo,
	}
}

// FindByCropID finds all varieties for a specific crop
func (r *CropVarietyRepository) FindByCropID(ctx context.Context, cropID string, page, pageSize int) ([]*crop_variety.CropVariety, int, error) {
	filter := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("is_active", base.OpEqual, true).
		Build()

	filter.Page = page
	filter.PageSize = pageSize
	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	varieties, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find varieties: %w", err)
	}

	count, err := r.BaseFilterableRepository.Count(ctx, filter, &crop_variety.CropVariety{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count varieties: %w", err)
	}

	return varieties, int(count), nil
}

// FindByCropIDAndName finds a variety by crop ID and name
func (r *CropVarietyRepository) FindByCropIDAndName(ctx context.Context, cropID, name string) (*crop_variety.CropVariety, error) {
	filter := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("name", base.OpEqual, name).
		Where("is_active", base.OpEqual, true).
		Build()

	variety, err := r.BaseFilterableRepository.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find variety: %w", err)
	}

	return variety, nil
}

// Search finds varieties by name or description
// Uses in-memory filtering for LIKE pattern matching
func (r *CropVarietyRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*crop_variety.CropVariety, int, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Build()

	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	// Get all active varieties (no pagination yet)
	varieties, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find varieties: %w", err)
	}

	// Filter by search term in-memory
	filtered := filterVarietiesBySearch(varieties, searchTerm)

	total := len(filtered)

	// Apply pagination manually
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*crop_variety.CropVariety{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// FindWithFilters finds varieties with multiple filters
func (r *CropVarietyRepository) FindWithFilters(ctx context.Context, filters CropVarietyFilters) ([]*crop_variety.CropVariety, int, error) {
	filterBuilder := base.NewFilterBuilder()

	// Apply crop_id filter
	if filters.CropID != "" {
		filterBuilder = filterBuilder.Where("crop_id", base.OpEqual, filters.CropID)
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

	varieties, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find varieties: %w", err)
	}

	// Apply search filter in-memory if provided
	if filters.Search != "" {
		varieties = filterVarietiesBySearch(varieties, filters.Search)
	}

	total := len(varieties)
	return varieties, total, nil
}

// GetActiveVarietiesForLookup gets simplified variety data for dropdown/lookup
func (r *CropVarietyRepository) GetActiveVarietiesForLookup(ctx context.Context, cropID string) ([]*crop_variety.CropVariety, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true)

	if cropID != "" {
		filterBuilder = filterBuilder.Where("crop_id", base.OpEqual, cropID)
	}

	filter := filterBuilder.Build()
	filter.Sort = []base.SortField{{Field: "name", Direction: "asc"}}

	varieties, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get varieties for lookup: %w", err)
	}

	return varieties, nil
}

// GetVarietiesWithCropInfo gets varieties with crop information
// Note: JOIN operations need to be done in service layer
func (r *CropVarietyRepository) GetVarietiesWithCropInfo(ctx context.Context, filters CropVarietyFilters) ([]*crop_variety.CropVariety, int, error) {
	return r.FindWithFilters(ctx, filters)
}

// CheckVarietyExists checks if a variety exists for a crop
func (r *CropVarietyRepository) CheckVarietyExists(ctx context.Context, cropID, name string, excludeID ...string) (bool, error) {
	filterBuilder := base.NewFilterBuilder().
		Where("crop_id", base.OpEqual, cropID).
		Where("name", base.OpEqual, name)

	if len(excludeID) > 0 && excludeID[0] != "" {
		filterBuilder = filterBuilder.Where("id", base.OpNotEqual, excludeID[0])
	}

	filter := filterBuilder.Build()

	count, err := r.BaseFilterableRepository.Count(ctx, filter, &crop_variety.CropVariety{})
	if err != nil {
		return false, fmt.Errorf("failed to check variety exists: %w", err)
	}

	return count > 0, nil
}

// filterVarietiesBySearch filters varieties by name or description
func filterVarietiesBySearch(varieties []*crop_variety.CropVariety, searchTerm string) []*crop_variety.CropVariety {
	filtered := make([]*crop_variety.CropVariety, 0)
	searchLower := strings.ToLower(searchTerm)

	for _, v := range varieties {
		nameMatch := strings.Contains(strings.ToLower(v.Name), searchLower)
		descMatch := v.Description != nil && strings.Contains(strings.ToLower(*v.Description), searchLower)

		if nameMatch || descMatch {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// CropVarietyFilters represents filters for crop variety queries
type CropVarietyFilters struct {
	CropID   string
	IsActive *bool
	Search   string
	Page     int
	PageSize int
}

// CropVarietyLookup represents simplified variety data for lookups
type CropVarietyLookup struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DurationDays *int   `json:"duration_days,omitempty"`
}

// VarietyWithCropInfo represents variety data with crop information
type VarietyWithCropInfo struct {
	crop_variety.CropVariety
	CropName string `json:"crop_name"`
}
