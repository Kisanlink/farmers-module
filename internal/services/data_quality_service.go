package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	farmRepo "github.com/Kisanlink/farmers-module/internal/repo/farm"
	farmerRepo "github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// DataQualityServiceImpl implements DataQualityService
type DataQualityServiceImpl struct {
	db                  *gorm.DB
	farmRepo            *farmRepo.FarmRepository
	farmerLinkageRepo   *farmerRepo.FarmerLinkRepository
	aaaService          AAAService
	notificationService NotificationService
}

// NewDataQualityService creates a new data quality service
func NewDataQualityService(
	db *gorm.DB,
	farmRepo *farmRepo.FarmRepository,
	farmerLinkageRepo *farmerRepo.FarmerLinkRepository,
	aaaService AAAService,
	notificationService NotificationService,
) DataQualityService {
	return &DataQualityServiceImpl{
		db:                  db,
		farmRepo:            farmRepo,
		farmerLinkageRepo:   farmerLinkageRepo,
		aaaService:          aaaService,
		notificationService: notificationService,
	}
}

// ValidateGeometry validates WKT geometry with PostGIS validation and SRID checks
func (s *DataQualityServiceImpl) ValidateGeometry(ctx context.Context, req interface{}) (interface{}, error) {
	validateReq, ok := req.(*requests.ValidateGeometryRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for ValidateGeometry")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can audit farm data quality
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "audit", "", validateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	response := &responses.ValidateGeometryResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: validateReq.RequestID,
			Message:   "Geometry validation completed",
		},
		WKT:      validateReq.WKT,
		IsValid:  false,
		Errors:   []string{},
		Warnings: []string{},
		SRID:     4326,
	}

	// Basic validation
	if validateReq.WKT == "" {
		response.Errors = append(response.Errors, "geometry cannot be empty")
		return response, nil
	}

	// If no database connection, do basic validation
	if s.db == nil {
		if !strings.HasPrefix(strings.ToUpper(validateReq.WKT), "POLYGON") {
			response.Errors = append(response.Errors, "only POLYGON geometries are supported")
		} else {
			response.IsValid = true
		}
		return response, nil
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := s.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("failed to check PostGIS availability: %v", err))
		// Treat as PostGIS not available and continue with basic validation
		postgisAvailable = false
	}

	if !postgisAvailable {
		response.Warnings = append(response.Warnings, "PostGIS extension not available, performing basic validation only")
		if !strings.HasPrefix(strings.ToUpper(validateReq.WKT), "POLYGON") {
			response.Errors = append(response.Errors, "only POLYGON geometries are supported")
		} else {
			response.IsValid = true
		}
		return response, nil
	}

	// Validate WKT format using PostGIS
	var isValid bool
	if err := s.db.Raw("SELECT ST_IsValid(ST_GeomFromText(?, 4326))", validateReq.WKT).Scan(&isValid).Error; err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("invalid WKT format: %v", err))
		return response, nil
	}
	if !isValid {
		response.Errors = append(response.Errors, "geometry is not valid according to PostGIS")
		return response, nil
	}

	// Check SRID enforcement (must be 4326)
	var srid int
	if err := s.db.Raw("SELECT ST_SRID(ST_GeomFromText(?, 4326))", validateReq.WKT).Scan(&srid).Error; err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("failed to get SRID: %v", err))
	} else if srid != 4326 {
		response.Errors = append(response.Errors, fmt.Sprintf("geometry must use SRID 4326, got %d", srid))
		return response, nil
	}

	// Check geometry type (must be POLYGON)
	var geomType string
	if err := s.db.Raw("SELECT ST_GeometryType(ST_GeomFromText(?, 4326))", validateReq.WKT).Scan(&geomType).Error; err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("failed to get geometry type: %v", err))
	} else if geomType != "ST_Polygon" {
		response.Errors = append(response.Errors, fmt.Sprintf("only POLYGON geometries are supported, got %s", geomType))
		return response, nil
	}

	// Check for self-intersections
	var hasIntersections bool
	if err := s.db.Raw("SELECT NOT ST_IsSimple(ST_GeomFromText(?, 4326))", validateReq.WKT).Scan(&hasIntersections).Error; err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("failed to check geometry simplicity: %v", err))
	} else if hasIntersections {
		response.Errors = append(response.Errors, "geometry has self-intersections")
		return response, nil
	}

	// Check area (should be reasonable for farm boundaries)
	var areaHa float64
	if err := s.db.Raw("SELECT ST_Area(ST_GeomFromText(?, 4326)::geography)/10000.0", validateReq.WKT).Scan(&areaHa).Error; err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("failed to calculate area: %v", err))
	} else {
		response.AreaHa = &areaHa
		if areaHa < 0.01 { // Less than 100 square meters
			response.Warnings = append(response.Warnings, fmt.Sprintf("geometry area is very small: %.4f hectares", areaHa))
		}
		if areaHa > 10000 { // More than 10,000 hectares
			response.Warnings = append(response.Warnings, fmt.Sprintf("geometry area is very large: %.2f hectares", areaHa))
		}
	}

	// Optional: Check if geometry is within India bounds (rough check)
	if validateReq.CheckBounds {
		var withinIndia bool
		indiaBounds := "POLYGON((68 6, 97 6, 97 37, 68 37, 68 6))" // Approximate India bounding box
		if err := s.db.Raw("SELECT ST_Within(ST_GeomFromText(?, 4326), ST_GeomFromText(?, 4326))", validateReq.WKT, indiaBounds).Scan(&withinIndia).Error; err != nil {
			response.Warnings = append(response.Warnings, fmt.Sprintf("failed to check India bounds: %v", err))
		} else if !withinIndia {
			response.Warnings = append(response.Warnings, "geometry appears to be outside India boundaries")
		}
	}

	// If we got here without errors, the geometry is valid
	if len(response.Errors) == 0 {
		response.IsValid = true
		response.Message = "Geometry validation passed"
	}

	return response, nil
}

// ReconcileAAALinks heals broken AAA references in farmer_links
// Business Rule 6.1: AAA-Local State Invariants
// - Every farmer_links with status='ACTIVE' MUST reference valid AAA user and org
// - Mark as ORPHANED when AAA references become invalid (drift detection)
func (s *DataQualityServiceImpl) ReconcileAAALinks(ctx context.Context, req interface{}) (interface{}, error) {
	reconcileReq, ok := req.(*requests.ReconcileAAALinksRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for ReconcileAAALinks")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can maintain AAA links
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "admin", "maintain", "", reconcileReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	response := &responses.ReconcileAAALinksResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: reconcileReq.RequestID,
			Message:   "AAA links reconciliation completed",
		},
		ProcessedLinks: 0,
		FixedLinks:     0,
		BrokenLinks:    0,
		Errors:         []string{},
	}

	// Get all farmer links (only ACTIVE ones need strict validation)
	filter := base.NewFilterBuilder().Where("status", base.OpEqual, "ACTIVE").Build()
	if reconcileReq.OrgID != "" {
		filter = base.NewFilterBuilder().
			Where("aaa_org_id", base.OpEqual, reconcileReq.OrgID).
			Where("status", base.OpEqual, "ACTIVE").
			Build()
	}

	farmerLinks, err := s.farmerLinkageRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer links: %w", err)
	}

	response.ProcessedLinks = len(farmerLinks)

	// Business Rule 6.2: Drift Detection and Recovery
	// Collect orphaned links for notification
	orphanedLinks := []*entities.FarmerLink{}

	// Check each link against AAA service
	for _, link := range farmerLinks {
		isDrifted := false
		var driftReason string

		// Invariant check: AAA user must exist
		_, err := s.aaaService.GetUser(ctx, link.AAAUserID)
		if err != nil {
			isDrifted = true
			driftReason = fmt.Sprintf("User %s not found in AAA", link.AAAUserID)
			response.BrokenLinks++
			response.Errors = append(response.Errors, driftReason)
		}

		// Invariant check: AAA organization must exist
		if !isDrifted {
			_, err := s.aaaService.GetOrganization(ctx, link.AAAOrgID)
			if err != nil {
				isDrifted = true
				driftReason = fmt.Sprintf("Organization %s not found in AAA", link.AAAOrgID)
				response.BrokenLinks++
				response.Errors = append(response.Errors, driftReason)
			}
		}

		// Recovery Action: Mark as ORPHANED when drift detected
		if isDrifted {
			if !reconcileReq.DryRun {
				link.Status = "ORPHANED"
				if updateErr := s.farmerLinkageRepo.Update(ctx, link); updateErr != nil {
					response.Errors = append(response.Errors, fmt.Sprintf("Failed to mark link as ORPHANED: %v", updateErr))
				} else {
					// Collect successfully orphaned links for notification
					orphanedLinks = append(orphanedLinks, link)
				}
			}
			continue
		}

		// If both AAA references exist and link was marked as ORPHANED, restore it
		if link.Status == "ORPHANED" || link.Status == "BROKEN" {
			if !reconcileReq.DryRun {
				link.Status = "ACTIVE"
				if updateErr := s.farmerLinkageRepo.Update(ctx, link); updateErr != nil {
					response.Errors = append(response.Errors, fmt.Sprintf("Failed to restore link: %v", updateErr))
				} else {
					response.FixedLinks++
				}
			} else {
				response.FixedLinks++ // Count what would be fixed in dry-run
			}
		}
	}

	// Business Rule 6.2: Notify FPO admin of data inconsistency
	if len(orphanedLinks) > 0 && !reconcileReq.DryRun {
		// Group orphaned links by organization for targeted notifications
		orgOrphanedLinks := make(map[string][]*entities.FarmerLink)
		for _, link := range orphanedLinks {
			orgOrphanedLinks[link.AAAOrgID] = append(orgOrphanedLinks[link.AAAOrgID], link)
		}

		// Send notification to each affected organization
		for orgID, links := range orgOrphanedLinks {
			if notifyErr := s.notificationService.SendOrphanedLinkAlert(ctx, orgID, links); notifyErr != nil {
				response.Errors = append(response.Errors, fmt.Sprintf("Failed to send notification to org %s: %v", orgID, notifyErr))
			}
		}
	}

	if reconcileReq.DryRun {
		response.Message = fmt.Sprintf("Dry run completed: %d links processed, %d would be fixed, %d orphaned",
			response.ProcessedLinks, response.FixedLinks, response.BrokenLinks)
	} else {
		response.Message = fmt.Sprintf("Reconciliation completed: %d links processed, %d fixed, %d orphaned",
			response.ProcessedLinks, response.FixedLinks, response.BrokenLinks)
	}

	return response, nil
}

// RebuildSpatialIndexes rebuilds GIST indexes for database maintenance
func (s *DataQualityServiceImpl) RebuildSpatialIndexes(ctx context.Context, req interface{}) (interface{}, error) {
	rebuildReq, ok := req.(*requests.RebuildSpatialIndexesRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for RebuildSpatialIndexes")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can maintain spatial indexes
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "admin", "maintain", "", rebuildReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	response := &responses.RebuildSpatialIndexesResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: rebuildReq.RequestID,
			Message:   "Spatial indexes rebuild completed",
		},
		RebuiltIndexes: []string{},
		Errors:         []string{},
	}

	if s.db == nil {
		response.Errors = append(response.Errors, "database connection not available")
		return response, nil
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := s.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("failed to check PostGIS availability: %v", err))
		// Treat as PostGIS not available
		postgisAvailable = false
	}

	if !postgisAvailable {
		response.Errors = append(response.Errors, "PostGIS extension not available")
		return response, nil
	}

	// List of spatial indexes to rebuild
	spatialIndexes := []struct {
		table  string
		index  string
		column string
	}{
		{"farms", "idx_farms_geometry", "geometry"},
	}

	for _, idx := range spatialIndexes {
		// Check if index exists
		var indexExists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM pg_indexes WHERE tablename = ? AND indexname = ?)`
		if err := s.db.Raw(checkQuery, idx.table, idx.index).Scan(&indexExists).Error; err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("failed to check index %s: %v", idx.index, err))
			continue
		}

		if indexExists {
			// Drop existing index
			dropQuery := fmt.Sprintf("DROP INDEX IF EXISTS %s", idx.index)
			if err := s.db.Exec(dropQuery).Error; err != nil {
				response.Errors = append(response.Errors, fmt.Sprintf("failed to drop index %s: %v", idx.index, err))
				continue
			}
		}

		// Create/recreate the spatial index
		createQuery := fmt.Sprintf("CREATE INDEX %s ON %s USING GIST (%s)", idx.index, idx.table, idx.column)
		if err := s.db.Exec(createQuery).Error; err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("failed to create index %s: %v", idx.index, err))
			continue
		}

		response.RebuiltIndexes = append(response.RebuiltIndexes, idx.index)
	}

	// Update statistics for spatial tables
	statisticsQueries := []string{
		"ANALYZE farms",
	}

	for _, query := range statisticsQueries {
		if err := s.db.Exec(query).Error; err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("failed to update statistics: %v", err))
		}
	}

	if len(response.Errors) == 0 {
		response.Message = fmt.Sprintf("Successfully rebuilt %d spatial indexes", len(response.RebuiltIndexes))
	} else {
		response.Message = fmt.Sprintf("Rebuilt %d indexes with %d errors", len(response.RebuiltIndexes), len(response.Errors))
	}

	return response, nil
}

// DetectFarmOverlaps detects spatial intersections between farm boundaries
func (s *DataQualityServiceImpl) DetectFarmOverlaps(ctx context.Context, req interface{}) (interface{}, error) {
	detectReq, ok := req.(*requests.DetectFarmOverlapsRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for DetectFarmOverlaps")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can audit farm overlaps
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "audit", "", detectReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	response := &responses.DetectFarmOverlapsResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: detectReq.RequestID,
			Message:   "Farm overlap detection completed",
		},
		Overlaps:      []responses.FarmOverlap{},
		TotalOverlaps: 0,
	}

	if s.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := s.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		return nil, fmt.Errorf("failed to check PostGIS availability: %w", err)
	}

	if !postgisAvailable {
		return nil, fmt.Errorf("PostGIS extension not available for overlap detection")
	}

	// Build query to find overlapping farms
	query := `
		SELECT
			f1.id as farm1_id,
			f1.name as farm1_name,
			f1.aaa_farmer_user_id as farm1_farmer_id,
			f2.id as farm2_id,
			f2.name as farm2_name,
			f2.aaa_farmer_user_id as farm2_farmer_id,
			ST_Area(ST_Intersection(f1.geometry, f2.geometry))/10000.0 as overlap_area_ha,
			ST_Area(f1.geometry)/10000.0 as farm1_area_ha,
			ST_Area(f2.geometry)/10000.0 as farm2_area_ha
		FROM farms f1
		JOIN farms f2 ON f1.id < f2.id
		WHERE f1.aaa_org_id = ?
		AND f2.aaa_org_id = ?
		AND f1.deleted_at IS NULL
		AND f2.deleted_at IS NULL
		AND ST_Intersects(f1.geometry, f2.geometry)
		AND ST_Area(ST_Intersection(f1.geometry, f2.geometry)) > 0`

	args := []interface{}{detectReq.OrgID, detectReq.OrgID}

	// Add minimum overlap area filter if specified
	if detectReq.MinOverlapAreaHa != nil && *detectReq.MinOverlapAreaHa > 0 {
		query += " AND ST_Area(ST_Intersection(f1.geometry, f2.geometry))/10000.0 >= ?"
		args = append(args, *detectReq.MinOverlapAreaHa)
	}

	// Add limit if specified
	if detectReq.Limit != nil && *detectReq.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, *detectReq.Limit)
	}

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to detect overlaps: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't return it as it's in defer
		}
	}()

	for rows.Next() {
		var overlap responses.FarmOverlap
		if err := rows.Scan(
			&overlap.Farm1ID,
			&overlap.Farm1Name,
			&overlap.Farm1FarmerID,
			&overlap.Farm2ID,
			&overlap.Farm2Name,
			&overlap.Farm2FarmerID,
			&overlap.OverlapAreaHa,
			&overlap.Farm1AreaHa,
			&overlap.Farm2AreaHa,
		); err != nil {
			return nil, fmt.Errorf("failed to scan overlap result: %w", err)
		}

		// Calculate overlap percentages
		if overlap.Farm1AreaHa > 0 {
			overlap.OverlapPercentageFarm1 = (overlap.OverlapAreaHa / overlap.Farm1AreaHa) * 100
		}
		if overlap.Farm2AreaHa > 0 {
			overlap.OverlapPercentageFarm2 = (overlap.OverlapAreaHa / overlap.Farm2AreaHa) * 100
		}

		response.Overlaps = append(response.Overlaps, overlap)
	}

	response.TotalOverlaps = len(response.Overlaps)

	if response.TotalOverlaps == 0 {
		response.Message = "No farm overlaps detected"
	} else {
		response.Message = fmt.Sprintf("Detected %d farm overlaps", response.TotalOverlaps)
	}

	return response, nil
}
