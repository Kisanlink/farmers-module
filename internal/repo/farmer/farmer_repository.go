package farmer

import (
	"context"
	"fmt"
	"log"

	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmerRepository wraps BaseFilterableRepository with custom methods for normalized Farmer entity
type FarmerRepository struct {
	*base.BaseFilterableRepository[*farmerentity.Farmer]
	db *gorm.DB
}

// FarmerLinkRepository wraps BaseFilterableRepository with custom methods
type FarmerLinkRepository struct {
	*base.BaseFilterableRepository[*farmerentity.FarmerLink]
	db *gorm.DB
}

// NewFarmerRepository creates a new farmer repository using BaseFilterableRepository with normalized Farmer entity
func NewFarmerRepository(dbManager interface{}) *FarmerRepository {
	baseRepo := base.NewBaseFilterableRepository[*farmerentity.Farmer]()
	baseRepo.SetDBManager(dbManager)

	// Get the GORM DB instance
	var db *gorm.DB
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			db = gormDB
		}
	}

	return &FarmerRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *FarmerRepository) Count(ctx context.Context, filter *base.Filter, model *farmerentity.Farmer) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&farmerentity.Farmer{}).WithContext(ctx)

	// Apply filters
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			query = query.Where(condition.Field+" "+string(condition.Operator)+" ?", condition.Value)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// FindByOrgID retrieves farmers linked to a specific organization through the farmer_links table
// This method performs a JOIN between farmers and farmer_links tables to filter by aaa_org_id
func (r *FarmerRepository) FindByOrgID(ctx context.Context, aaaOrgID string, filter *base.Filter) ([]*farmerentity.Farmer, error) {
	log.Printf("[FindByOrgID] Called with aaaOrgID=%s, db=%v", aaaOrgID, r.db != nil)

	if r.db == nil {
		log.Printf("[FindByOrgID] ERROR: database connection not available")
		return nil, fmt.Errorf("database connection not available")
	}

	// Start with the farmers table and join with farmer_links
	// Note: We must explicitly filter out soft-deleted records for BOTH tables since BaseModel uses
	// *time.Time (not gorm.DeletedAt), which means GORM doesn't automatically apply soft delete filtering
	query := r.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Joins("INNER JOIN farmer_links ON farmers.aaa_user_id = farmer_links.aaa_user_id AND farmer_links.deleted_at IS NULL").
		Where("farmer_links.aaa_org_id = ?", aaaOrgID).
		Where("farmers.deleted_at IS NULL")

	// Apply additional filters if provided
	if filter != nil {
		// Apply filter conditions
		if filter.Group.Conditions != nil {
			for _, condition := range filter.Group.Conditions {
				// Skip aaa_org_id filter as it's already applied via the join
				if condition.Field == "aaa_org_id" {
					continue
				}

				// Apply farmer table filters with proper table prefix
				switch condition.Operator {
				case base.OpEqual:
					query = query.Where("farmers."+condition.Field+" = ?", condition.Value)
				case base.OpNotEqual:
					query = query.Where("farmers."+condition.Field+" != ?", condition.Value)
				case base.OpIn:
					query = query.Where("farmers."+condition.Field+" IN ?", condition.Value)
				case base.OpContains:
					query = query.Where("farmers."+condition.Field+" LIKE ?", "%"+fmt.Sprint(condition.Value)+"%")
				case base.OpStartsWith:
					query = query.Where("farmers."+condition.Field+" LIKE ?", fmt.Sprint(condition.Value)+"%")
				case base.OpEndsWith:
					query = query.Where("farmers."+condition.Field+" LIKE ?", "%"+fmt.Sprint(condition.Value))
				case base.OpIsNull:
					query = query.Where("farmers." + condition.Field + " IS NULL")
				case base.OpIsNotNull:
					query = query.Where("farmers." + condition.Field + " IS NOT NULL")
				default:
					// For other operators, apply without table prefix (GORM will handle it)
					query = query.Where(condition.Field+" = ?", condition.Value)
				}
			}
		}

		// Apply sorting
		if filter.Sort != nil {
			for _, sortField := range filter.Sort {
				order := "farmers." + sortField.Field + " " + sortField.Direction
				query = query.Order(order)
			}
		}

		// Apply pagination
		if filter.Page > 0 && filter.PageSize > 0 {
			offset := (filter.Page - 1) * filter.PageSize
			query = query.Limit(filter.PageSize).Offset(offset)
		}

		// Apply preloads (relationships)
		if filter.Preloads != nil {
			for _, preload := range filter.Preloads {
				if len(preload.Conditions) > 0 {
					query = query.Preload(preload.Relation, preload.Conditions...)
				} else {
					query = query.Preload(preload.Relation)
				}
			}
		}
	}

	// Execute the query
	var farmers []*farmerentity.Farmer
	if err := query.Find(&farmers).Error; err != nil {
		log.Printf("[FindByOrgID] ERROR: query failed: %v", err)
		return nil, fmt.Errorf("failed to find farmers by org_id: %w", err)
	}

	log.Printf("[FindByOrgID] Found %d farmers for org_id=%s", len(farmers), aaaOrgID)
	return farmers, nil
}

// CountByOrgID counts farmers linked to a specific organization through the farmer_links table
func (r *FarmerRepository) CountByOrgID(ctx context.Context, aaaOrgID string, filter *base.Filter) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// Start with the farmers table and join with farmer_links
	// Note: We must explicitly filter out soft-deleted records for BOTH tables since BaseModel uses
	// *time.Time (not gorm.DeletedAt), which means GORM doesn't automatically apply soft delete filtering
	query := r.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Joins("INNER JOIN farmer_links ON farmers.aaa_user_id = farmer_links.aaa_user_id AND farmer_links.deleted_at IS NULL").
		Where("farmer_links.aaa_org_id = ?", aaaOrgID).
		Where("farmers.deleted_at IS NULL")

	// Apply additional filters if provided
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			// Skip aaa_org_id filter as it's already applied via the join
			if condition.Field == "aaa_org_id" {
				continue
			}

			// Apply farmer table filters with proper table prefix
			switch condition.Operator {
			case base.OpEqual:
				query = query.Where("farmers."+condition.Field+" = ?", condition.Value)
			case base.OpNotEqual:
				query = query.Where("farmers."+condition.Field+" != ?", condition.Value)
			case base.OpIn:
				query = query.Where("farmers."+condition.Field+" IN ?", condition.Value)
			case base.OpContains:
				query = query.Where("farmers."+condition.Field+" LIKE ?", "%"+fmt.Sprint(condition.Value)+"%")
			default:
				query = query.Where("farmers."+condition.Field+" = ?", condition.Value)
			}
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count farmers by org_id: %w", err)
	}

	return count, nil
}

// NewFarmerLinkRepository creates a new farmer link repository using BaseFilterableRepository
func NewFarmerLinkRepository(dbManager interface{}) *FarmerLinkRepository {
	baseRepo := base.NewBaseFilterableRepository[*farmerentity.FarmerLink]()
	baseRepo.SetDBManager(dbManager)

	// Get the GORM DB instance
	var db *gorm.DB
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			db = gormDB
		}
	}

	return &FarmerLinkRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// Count overrides the base Count method to properly set the model
func (r *FarmerLinkRepository) Count(ctx context.Context, filter *base.Filter, model *farmerentity.FarmerLink) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	query := r.db.Model(&farmerentity.FarmerLink{}).WithContext(ctx)

	// Apply filters
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			query = query.Where(condition.Field+" "+string(condition.Operator)+" ?", condition.Value)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// FindUnscoped finds farmer links including soft-deleted records
func (r *FarmerLinkRepository) FindUnscoped(ctx context.Context, filter *base.Filter) ([]*farmerentity.FarmerLink, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := r.db.Unscoped().Model(&farmerentity.FarmerLink{}).WithContext(ctx)

	// Apply filters
	if filter != nil && filter.Group.Conditions != nil {
		for _, condition := range filter.Group.Conditions {
			switch condition.Operator {
			case base.OpEqual:
				query = query.Where(condition.Field+" = ?", condition.Value)
			case base.OpNotEqual:
				query = query.Where(condition.Field+" != ?", condition.Value)
			case base.OpIn:
				query = query.Where(condition.Field+" IN ?", condition.Value)
			default:
				query = query.Where(condition.Field+" = ?", condition.Value)
			}
		}
	}

	var results []*farmerentity.FarmerLink
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to find farmer links (unscoped): %w", err)
	}

	return results, nil
}

// Restore restores a soft-deleted farmer link by clearing deleted_at and updating status
func (r *FarmerLinkRepository) Restore(ctx context.Context, entity *farmerentity.FarmerLink) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Use Unscoped to update soft-deleted records
	result := r.db.Unscoped().WithContext(ctx).
		Model(&farmerentity.FarmerLink{}).
		Where("id = ?", entity.ID).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"deleted_by": nil,
			"status":     entity.Status,
			"updated_at": r.db.NowFunc(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to restore farmer link: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no farmer link found with id: %s", entity.ID)
	}

	return nil
}
