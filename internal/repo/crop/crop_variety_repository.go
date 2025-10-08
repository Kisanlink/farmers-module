package crop

import (
	"context"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// CropVarietyRepository provides data access methods for crop varieties
type CropVarietyRepository struct {
	*base.BaseFilterableRepository[*crop_variety.CropVariety]
	db *gorm.DB
}

// NewCropVarietyRepository creates a new crop variety repository
func NewCropVarietyRepository(dbManager interface{}) *CropVarietyRepository {
	baseRepo := base.NewBaseFilterableRepository[*crop_variety.CropVariety]()
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

	return &CropVarietyRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// FindByCropID finds all varieties for a specific crop
func (r *CropVarietyRepository) FindByCropID(ctx context.Context, cropID string, page, pageSize int) ([]*crop_variety.CropVariety, int, error) {
	var varieties []*crop_variety.CropVariety
	var total int64

	query := r.db.WithContext(ctx).Model(&crop_variety.CropVariety{}).
		Where("crop_id = ?", cropID).
		Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with crop data
	offset := (page - 1) * pageSize
	if err := query.Preload("Crop").Offset(offset).Limit(pageSize).Order("name ASC").Find(&varieties).Error; err != nil {
		return nil, 0, err
	}

	return varieties, int(total), nil
}

// FindByCropIDAndName finds a variety by crop ID and name
func (r *CropVarietyRepository) FindByCropIDAndName(ctx context.Context, cropID, name string) (*crop_variety.CropVariety, error) {
	var variety crop_variety.CropVariety
	err := r.db.WithContext(ctx).
		Where("crop_id = ?", cropID).
		Where("LOWER(name) = LOWER(?)", name).
		Where("is_active = ?", true).
		Preload("Crop").
		First(&variety).Error
	if err != nil {
		return nil, err
	}
	return &variety, nil
}

// Search finds varieties by name or description
func (r *CropVarietyRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*crop_variety.CropVariety, int, error) {
	var varieties []*crop_variety.CropVariety
	var total int64

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	query := r.db.WithContext(ctx).Model(&crop_variety.CropVariety{}).
		Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern).
		Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with crop data
	offset := (page - 1) * pageSize
	if err := query.Preload("Crop").Offset(offset).Limit(pageSize).Order("name ASC").Find(&varieties).Error; err != nil {
		return nil, 0, err
	}

	return varieties, int(total), nil
}

// FindWithFilters finds varieties with multiple filters
func (r *CropVarietyRepository) FindWithFilters(ctx context.Context, filters CropVarietyFilters) ([]*crop_variety.CropVariety, int, error) {
	var varieties []*crop_variety.CropVariety
	var total int64

	query := r.db.WithContext(ctx).Model(&crop_variety.CropVariety{})

	// Apply filters
	if filters.CropID != "" {
		query = query.Where("crop_id = ?", filters.CropID)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}

	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with crop data
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Preload("Crop").Offset(offset).Limit(filters.PageSize).Order("name ASC").Find(&varieties).Error; err != nil {
		return nil, 0, err
	}

	return varieties, int(total), nil
}

// GetActiveVarietiesForLookup gets simplified variety data for dropdown/lookup
func (r *CropVarietyRepository) GetActiveVarietiesForLookup(ctx context.Context, cropID string) ([]*CropVarietyLookup, error) {
	var varieties []*CropVarietyLookup

	query := r.db.WithContext(ctx).
		Model(&crop_variety.CropVariety{}).
		Select("id, name, duration_days").
		Where("is_active = ?", true)

	if cropID != "" {
		query = query.Where("crop_id = ?", cropID)
	}

	err := query.Order("name ASC").Find(&varieties).Error
	if err != nil {
		return nil, err
	}

	return varieties, nil
}

// GetVarietiesWithCropInfo gets varieties with crop information
func (r *CropVarietyRepository) GetVarietiesWithCropInfo(ctx context.Context, filters CropVarietyFilters) ([]*VarietyWithCropInfo, int, error) {
	var varieties []*VarietyWithCropInfo
	var total int64

	query := r.db.WithContext(ctx).
		Model(&crop_variety.CropVariety{}).
		Select(`
			crop_varieties.id,
			crop_varieties.crop_id,
			crop_varieties.name,
			crop_varieties.description,
			crop_varieties.duration_days,
			crop_varieties.yield_per_acre,
			crop_varieties.yield_per_tree,
			crop_varieties.properties,
			crop_varieties.is_active,
			crop_varieties.created_at,
			crop_varieties.updated_at,
			crops.name as crop_name
		`).
		Joins("JOIN crops ON crop_varieties.crop_id = crops.id")

	// Apply filters
	if filters.CropID != "" {
		query = query.Where("crop_varieties.crop_id = ?", filters.CropID)
	}

	if filters.IsActive != nil {
		query = query.Where("crop_varieties.is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("crop_varieties.is_active = ?", true)
	}

	if filters.Search != "" {
		searchPattern := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("(LOWER(crop_varieties.name) LIKE ? OR LOWER(crop_varieties.description) LIKE ?)", searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).Order("crop_varieties.name ASC").Find(&varieties).Error; err != nil {
		return nil, 0, err
	}

	return varieties, int(total), nil
}

// CheckVarietyExists checks if a variety exists for a crop
func (r *CropVarietyRepository) CheckVarietyExists(ctx context.Context, cropID, name string, excludeID ...string) (bool, error) {
	query := r.db.WithContext(ctx).
		Model(&crop_variety.CropVariety{}).
		Where("crop_id = ?", cropID).
		Where("LOWER(name) = LOWER(?)", name)

	if len(excludeID) > 0 && excludeID[0] != "" {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
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