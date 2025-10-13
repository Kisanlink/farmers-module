package farm

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmRepository extends BaseFilterableRepository with geospatial operations
type FarmRepository struct {
	*base.BaseFilterableRepository[*farm.Farm]
	db *gorm.DB
}

// NewFarmRepository creates a new farm repository with geospatial capabilities
func NewFarmRepository(dbManager interface{}) *FarmRepository {
	baseRepo := base.NewBaseFilterableRepository[*farm.Farm]()
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

	return &FarmRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// ListByBoundingBox lists farms within a bounding box using spatial queries
func (r *FarmRepository) ListByBoundingBox(ctx context.Context, bbox requests.BoundingBox, filters map[string]interface{}) ([]*farm.Farm, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		return nil, fmt.Errorf("failed to check PostGIS availability: %w", err)
	}

	if !postgisAvailable {
		// Fallback to basic filtering without spatial operations
		// Build filter from map
		filterBuilder := base.NewFilterBuilder()
		for key, value := range filters {
			filterBuilder = filterBuilder.Where(key, base.OpEqual, value)
		}

		farms, err := r.BaseFilterableRepository.Find(ctx, filterBuilder.Build())
		return farms, err
	}

	// Build spatial query with PostGIS
	bboxWKT := fmt.Sprintf("POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
		bbox.MinLon, bbox.MinLat,
		bbox.MaxLon, bbox.MinLat,
		bbox.MaxLon, bbox.MaxLat,
		bbox.MinLon, bbox.MaxLat,
		bbox.MinLon, bbox.MinLat)

	var farms []*farm.Farm
	query := r.db.Where("ST_Intersects(geometry, ST_GeomFromText(?, 4326))", bboxWKT)

	// Apply additional filters
	for key, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	if err := query.Find(&farms).Error; err != nil {
		return nil, fmt.Errorf("failed to query farms by bounding box: %w", err)
	}

	return farms, nil
}

// ValidateGeometry validates WKT geometry using PostGIS
func (r *FarmRepository) ValidateGeometry(ctx context.Context, wkt string) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		return fmt.Errorf("failed to check PostGIS availability: %w", err)
	}

	if !postgisAvailable {
		// Basic validation without PostGIS
		if wkt == "" {
			return fmt.Errorf("geometry cannot be empty")
		}
		return nil
	}

	// Validate using PostGIS
	var isValid bool
	if err := r.db.Raw("SELECT ST_IsValid(ST_GeomFromText(?, 4326))", wkt).Scan(&isValid).Error; err != nil {
		return fmt.Errorf("invalid WKT format: %w", err)
	}
	if !isValid {
		return fmt.Errorf("geometry is not valid")
	}

	return nil
}

// Note: Count method is inherited from BaseFilterableRepository
// which properly delegates to PostgresManager.Count()
// No need to override it here

// CheckOverlap checks if a geometry overlaps with existing farms
func (r *FarmRepository) CheckOverlap(ctx context.Context, wkt string, excludeFarmID string, orgID string) (bool, []string, float64, error) {
	if r.db == nil {
		return false, nil, 0, fmt.Errorf("database connection not available")
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := r.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		return false, nil, 0, fmt.Errorf("failed to check PostGIS availability: %w", err)
	}

	if !postgisAvailable {
		// Cannot check overlap without PostGIS
		return false, nil, 0, nil
	}

	// Query for overlapping farms
	query := `
		SELECT id, ST_Area(ST_Intersection(geometry, ST_GeomFromText(?, 4326)))/10000.0 as overlap_area
		FROM farms
		WHERE aaa_org_id = ?
		AND ST_Intersects(geometry, ST_GeomFromText(?, 4326))
		AND deleted_at IS NULL`

	args := []interface{}{wkt, orgID, wkt}

	if excludeFarmID != "" {
		query += " AND id != ?"
		args = append(args, excludeFarmID)
	}

	rows, err := r.db.Raw(query, args...).Rows()
	if err != nil {
		return false, nil, 0, fmt.Errorf("failed to check overlap: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't return it as it's in defer
		}
	}()

	var overlappingFarms []string
	var totalOverlapArea float64

	for rows.Next() {
		var farmID string
		var overlapArea float64
		if err := rows.Scan(&farmID, &overlapArea); err != nil {
			return false, nil, 0, fmt.Errorf("failed to scan overlap result: %w", err)
		}
		overlappingFarms = append(overlappingFarms, farmID)
		totalOverlapArea += overlapArea
	}

	hasOverlap := len(overlappingFarms) > 0
	return hasOverlap, overlappingFarms, totalOverlapArea, nil
}
