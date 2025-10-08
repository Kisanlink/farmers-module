package crop

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// CropRepository provides data access methods for crops
type CropRepository struct {
	*base.BaseFilterableRepository[*crop.Crop]
	db *gorm.DB
}

// NewCropRepository creates a new crop repository
func NewCropRepository(dbManager interface{}) *CropRepository {
	baseRepo := base.NewBaseFilterableRepository[*crop.Crop]()
	baseRepo.SetDBManager(dbManager)

	// Get the GORM DB instance for custom queries
	var db *gorm.DB
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			db = gormDB
		}
	}

	return &CropRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// FindByName finds a crop by its name (case-insensitive)
func (r *CropRepository) FindByName(ctx context.Context, name string) (*crop.Crop, error) {
	var crop crop.Crop
	err := r.db.WithContext(ctx).
		Where("LOWER(name) = LOWER(?)", name).
		Where("is_active = ?", true).
		First(&crop).Error
	if err != nil {
		return nil, err
	}
	return &crop, nil
}

// FindByCategory finds crops by category
func (r *CropRepository) FindByCategory(ctx context.Context, category crop.CropCategory, page, pageSize int) ([]*crop.Crop, int, error) {
	var crops []*crop.Crop
	var total int64

	query := r.db.WithContext(ctx).Model(&crop.Crop{}).
		Where("category = ?", category).
		Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&crops).Error; err != nil {
		return nil, 0, err
	}

	return crops, int(total), nil
}

// FindBySeason finds crops that can be grown in a specific season
func (r *CropRepository) FindBySeason(ctx context.Context, season string, page, pageSize int) ([]*crop.Crop, int, error) {
	var crops []*crop.Crop
	var total int64

	query := r.db.WithContext(ctx).Model(&crop.Crop{}).
		Where("seasons::jsonb @> ?", fmt.Sprintf(`["%s"]`, season)).
		Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&crops).Error; err != nil {
		return nil, 0, err
	}

	return crops, int(total), nil
}

// Search finds crops by name or scientific name
func (r *CropRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*crop.Crop, int, error) {
	var crops []*crop.Crop
	var total int64

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	query := r.db.WithContext(ctx).Model(&crop.Crop{}).
		Where("(LOWER(name) LIKE ? OR LOWER(scientific_name) LIKE ?)", searchPattern, searchPattern).
		Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&crops).Error; err != nil {
		return nil, 0, err
	}

	return crops, int(total), nil
}

// FindWithFilters finds crops with multiple filters
func (r *CropRepository) FindWithFilters(ctx context.Context, filters CropFilters) ([]*crop.Crop, int, error) {
	var crops []*crop.Crop
	var total int64

	query := r.db.WithContext(ctx).Model(&crop.Crop{})

	// Apply filters
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.Season != "" {
		query = query.Where("seasons::jsonb @> ?", fmt.Sprintf(`["%s"]`, filters.Season))
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}

	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(scientific_name) LIKE ?)", searchPattern, searchPattern)
	}

	if len(filters.Seasons) > 0 {
		var seasonConditions []string
		var seasonArgs []interface{}
		for _, season := range filters.Seasons {
			seasonConditions = append(seasonConditions, "seasons::jsonb @> ?")
			seasonArgs = append(seasonArgs, fmt.Sprintf(`["%s"]`, season))
		}
		query = query.Where(strings.Join(seasonConditions, " OR "), seasonArgs...)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).Order("name ASC").Find(&crops).Error; err != nil {
		return nil, 0, err
	}

	return crops, int(total), nil
}

// GetCropWithVarietyCount gets crop with count of active varieties
func (r *CropRepository) GetCropWithVarietyCount(ctx context.Context, cropID string) (*CropWithVarietyCount, error) {
	var result CropWithVarietyCount

	err := r.db.WithContext(ctx).
		Model(&crop.Crop{}).
		Select("crops.*, COALESCE(variety_counts.count, 0) as variety_count").
		Joins("LEFT JOIN (SELECT crop_id, COUNT(*) as count FROM crop_varieties WHERE is_active = true GROUP BY crop_id) variety_counts ON crops.id = variety_counts.crop_id").
		Where("crops.id = ?", cropID).
		Where("crops.is_active = ?", true).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetActiveCropsForLookup gets simplified crop data for dropdown/lookup
func (r *CropRepository) GetActiveCropsForLookup(ctx context.Context, category, season string) ([]*CropLookup, error) {
	var crops []*CropLookup

	query := r.db.WithContext(ctx).
		Model(&crop.Crop{}).
		Select("id, name, category, seasons, unit").
		Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if season != "" {
		query = query.Where("seasons::jsonb @> ?", fmt.Sprintf(`["%s"]`, season))
	}

	err := query.Order("name ASC").Find(&crops).Error
	if err != nil {
		return nil, err
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