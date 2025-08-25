package services

import (
	"context"
	"fmt"
	"strings"

	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	farmRepo "github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmServiceImpl implements FarmService
type FarmServiceImpl struct {
	farmRepo   *farmRepo.FarmRepository
	aaaService AAAService
	db         *gorm.DB
}

// NewFarmService creates a new farm service
func NewFarmService(farmRepo *farmRepo.FarmRepository, aaaService AAAService, db *gorm.DB) FarmService {
	return &FarmServiceImpl{
		farmRepo:   farmRepo,
		aaaService: aaaService,
		db:         db,
	}
}

// CreateFarm implements W6: Create farm with WKT validation and PostGIS integration
func (s *FarmServiceImpl) CreateFarm(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateFarmRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for CreateFarm")
	}

	// Validate request
	if err := s.validateCreateFarmRequest(createReq); err != nil {
		return nil, err
	}

	// Check AAA permission for farm.create
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  createReq.AAAFarmerUserID,
		"resource": "farm",
		"action":   "create",
		"object":   createReq.AAAOrgID,
		"org_id":   createReq.AAAOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Validate geometry if provided
	var geometryWKT string
	if createReq.Geometry.WKT != "" {
		if err := s.validateGeometry(ctx, createReq.Geometry.WKT); err != nil {
			return nil, fmt.Errorf("geometry validation failed: %w", err)
		}
		geometryWKT = createReq.Geometry.WKT
	}

	// Create farm entity
	farm := &farmEntity.Farm{
		AAAFarmerUserID: createReq.AAAFarmerUserID,
		AAAOrgID:        createReq.AAAOrgID,
		Name:            fmt.Sprintf("Farm-%s", createReq.AAAFarmerUserID[:8]),
		Geometry:        geometryWKT,
		Metadata:        createReq.Metadata,
	}

	// Set name if provided in metadata
	if name, exists := createReq.Metadata["name"]; exists {
		farm.Name = name
	}

	// Create farm in database
	if err := s.farmRepo.Create(ctx, farm); err != nil {
		return nil, fmt.Errorf("failed to create farm: %w", err)
	}

	// Fetch the created farm to get computed area
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, farm.ID).Build()
	createdFarm, err := s.farmRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created farm: %w", err)
	}

	// Convert to response
	farmData := s.convertFarmToData(createdFarm)
	response := responses.NewFarmResponse(farmData, "Farm created successfully")
	response.SetRequestID(createReq.RequestID)

	return &response, nil
}

// UpdateFarm implements W7: Update farm with proper authorization
func (s *FarmServiceImpl) UpdateFarm(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateFarmRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for UpdateFarm")
	}

	// Get existing farm
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, updateReq.ID).Build()
	existingFarm, err := s.farmRepo.FindOne(ctx, filter)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get farm: %w", err)
	}

	// Check AAA permission for farm.update
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  existingFarm.AAAFarmerUserID,
		"resource": "farm",
		"action":   "update",
		"object":   existingFarm.ID,
		"org_id":   existingFarm.AAAOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Update fields
	if updateReq.Geometry != nil && updateReq.Geometry.WKT != "" {
		if err := s.validateGeometry(ctx, updateReq.Geometry.WKT); err != nil {
			return nil, fmt.Errorf("geometry validation failed: %w", err)
		}
		existingFarm.Geometry = updateReq.Geometry.WKT
	}

	if updateReq.Metadata != nil {
		existingFarm.Metadata = updateReq.Metadata
		// Update name if provided in metadata
		if name, exists := updateReq.Metadata["name"]; exists {
			existingFarm.Name = name
		}
	}

	// Update farm in database
	if err := s.farmRepo.Update(ctx, existingFarm); err != nil {
		return nil, fmt.Errorf("failed to update farm: %w", err)
	}

	// Fetch updated farm to get computed area
	filter = base.NewFilterBuilder().Where("id", base.OpEqual, existingFarm.ID).Build()
	updatedFarm, err := s.farmRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated farm: %w", err)
	}

	// Convert to response
	farmData := s.convertFarmToData(updatedFarm)
	response := responses.NewFarmResponse(farmData, "Farm updated successfully")
	response.SetRequestID(updateReq.RequestID)

	return &response, nil
}

// DeleteFarm implements W8: Delete farm with cascade delete
func (s *FarmServiceImpl) DeleteFarm(ctx context.Context, req interface{}) error {
	deleteReq, ok := req.(*requests.DeleteFarmRequest)
	if !ok {
		return fmt.Errorf("invalid request type for DeleteFarm")
	}

	// Get existing farm
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, deleteReq.ID).Build()
	existingFarm, err := s.farmRepo.FindOne(ctx, filter)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.ErrNotFound
		}
		return fmt.Errorf("failed to get farm: %w", err)
	}

	// Check AAA permission for farm.delete
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  existingFarm.AAAFarmerUserID,
		"resource": "farm",
		"action":   "delete",
		"object":   existingFarm.ID,
		"org_id":   existingFarm.AAAOrgID,
	})
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return common.ErrForbidden
	}

	// Delete farm (soft delete)
	if err := s.farmRepo.Delete(ctx, deleteReq.ID, existingFarm); err != nil {
		return fmt.Errorf("failed to delete farm: %w", err)
	}

	return nil
}

// ListFarms implements W9: List farms with spatial filtering and bounding box queries
func (s *FarmServiceImpl) ListFarms(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListFarmsRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for ListFarms")
	}

	// Check AAA permission for farm.list
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  listReq.AAAFarmerUserID,
		"resource": "farm",
		"action":   "list",
		"org_id":   listReq.AAAOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := make(map[string]interface{})
	if listReq.AAAFarmerUserID != "" {
		filters["aaa_farmer_user_id"] = listReq.AAAFarmerUserID
	}
	if listReq.AAAOrgID != "" {
		filters["aaa_org_id"] = listReq.AAAOrgID
	}

	// Build filter for database query
	filterBuilder := base.NewFilterBuilder().Page(listReq.Page, listReq.PageSize)

	// Add filters
	if listReq.AAAFarmerUserID != "" {
		filterBuilder = filterBuilder.Where("aaa_farmer_user_id", base.OpEqual, listReq.AAAFarmerUserID)
	}
	if listReq.AAAOrgID != "" {
		filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, listReq.AAAOrgID)
	}

	// Get farms
	farms, err := s.farmRepo.Find(ctx, filterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to list farms: %w", err)
	}

	// Get total count
	totalCount, err := s.farmRepo.Count(ctx, filterBuilder.Build(), &farmEntity.Farm{})
	if err != nil {
		return nil, fmt.Errorf("failed to count farms: %w", err)
	}

	// Apply area filters if specified
	if listReq.MinArea != nil || listReq.MaxArea != nil {
		farms = s.filterFarmsByArea(farms, listReq.MinArea, listReq.MaxArea)
		totalCount = int64(len(farms))
	}

	// Convert to response data
	farmDataList := make([]*responses.FarmData, len(farms))
	for i, farm := range farms {
		farmDataList[i] = s.convertFarmToData(farm)
	}

	response := responses.NewFarmListResponse(farmDataList, listReq.Page, listReq.PageSize, totalCount)
	response.SetRequestID(listReq.RequestID)

	return &response, nil
}

// GetFarm gets farm by ID
func (s *FarmServiceImpl) GetFarm(ctx context.Context, farmID string) (interface{}, error) {
	// Get farm
	filter := base.NewFilterBuilder().Where("id", base.OpEqual, farmID).Build()
	farm, err := s.farmRepo.FindOne(ctx, filter)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get farm: %w", err)
	}

	// Check AAA permission for farm.read
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  farm.AAAFarmerUserID,
		"resource": "farm",
		"action":   "read",
		"object":   farm.ID,
		"org_id":   farm.AAAOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Convert to response
	farmData := s.convertFarmToData(farm)
	response := responses.NewFarmResponse(farmData, "Farm retrieved successfully")

	return &response, nil
}

// ListFarmsByBoundingBox lists farms within a bounding box
func (s *FarmServiceImpl) ListFarmsByBoundingBox(ctx context.Context, bbox requests.BoundingBox, filters requests.ListFarmsRequest) (interface{}, error) {
	// Check AAA permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, map[string]interface{}{
		"subject":  filters.AAAFarmerUserID,
		"resource": "farm",
		"action":   "list",
		"org_id":   filters.AAAOrgID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build spatial query
	bboxWKT := fmt.Sprintf("POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
		bbox.MinLon, bbox.MinLat,
		bbox.MaxLon, bbox.MinLat,
		bbox.MaxLon, bbox.MaxLat,
		bbox.MinLon, bbox.MaxLat,
		bbox.MinLon, bbox.MinLat)

	var farms []*farmEntity.Farm
	query := s.db.Where("ST_Intersects(geometry, ST_GeomFromText(?, 4326))", bboxWKT)

	// Apply additional filters
	if filters.AAAFarmerUserID != "" {
		query = query.Where("aaa_farmer_user_id = ?", filters.AAAFarmerUserID)
	}
	if filters.AAAOrgID != "" {
		query = query.Where("aaa_org_id = ?", filters.AAAOrgID)
	}

	if err := query.Find(&farms).Error; err != nil {
		return nil, fmt.Errorf("failed to query farms by bounding box: %w", err)
	}

	// Convert to response data
	farmDataList := make([]*responses.FarmData, len(farms))
	for i, farm := range farms {
		farmDataList[i] = s.convertFarmToData(farm)
	}

	response := responses.NewFarmListResponse(farmDataList, 1, len(farms), int64(len(farms)))
	response.SetRequestID(filters.RequestID)

	return &response, nil
}

// ValidateGeometry validates WKT geometry with SRID enforcement and integrity checks
func (s *FarmServiceImpl) ValidateGeometry(ctx context.Context, wkt string) error {
	return s.validateGeometry(ctx, wkt)
}

// Helper methods

func (s *FarmServiceImpl) validateCreateFarmRequest(req *requests.CreateFarmRequest) error {
	if req.AAAFarmerUserID == "" {
		return fmt.Errorf("farmer user ID is required")
	}
	if req.AAAOrgID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if req.Geometry.WKT == "" {
		return fmt.Errorf("geometry is required")
	}
	return nil
}

func (s *FarmServiceImpl) validateGeometry(ctx context.Context, wkt string) error {
	if wkt == "" {
		return common.ErrInvalidFarmGeometry
	}

	// If no database connection, do basic validation
	if s.db == nil {
		// Basic WKT format validation without PostGIS
		if !strings.HasPrefix(strings.ToUpper(wkt), "POLYGON") {
			return fmt.Errorf("only POLYGON geometries are supported")
		}
		return nil
	}

	// Check if PostGIS is available
	var postgisAvailable bool
	if err := s.db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		return fmt.Errorf("failed to check PostGIS availability: %w", err)
	}

	if !postgisAvailable {
		// Basic WKT format validation without PostGIS
		if !strings.HasPrefix(strings.ToUpper(wkt), "POLYGON") {
			return fmt.Errorf("only POLYGON geometries are supported")
		}
		return nil
	}

	// Validate WKT format using PostGIS
	var isValid bool
	if err := s.db.Raw("SELECT ST_IsValid(ST_GeomFromText(?, 4326))", wkt).Scan(&isValid).Error; err != nil {
		return fmt.Errorf("invalid WKT format: %w", err)
	}
	if !isValid {
		return fmt.Errorf("geometry is not valid")
	}

	// Check geometry type (must be POLYGON)
	var geomType string
	if err := s.db.Raw("SELECT ST_GeometryType(ST_GeomFromText(?, 4326))", wkt).Scan(&geomType).Error; err != nil {
		return fmt.Errorf("failed to get geometry type: %w", err)
	}
	if geomType != "ST_Polygon" {
		return fmt.Errorf("only POLYGON geometries are supported, got %s", geomType)
	}

	// Check for self-intersections
	var hasIntersections bool
	if err := s.db.Raw("SELECT NOT ST_IsSimple(ST_GeomFromText(?, 4326))", wkt).Scan(&hasIntersections).Error; err != nil {
		return fmt.Errorf("failed to check geometry simplicity: %w", err)
	}
	if hasIntersections {
		return fmt.Errorf("geometry has self-intersections")
	}

	// Optional: Check if geometry is within India bounds (rough check)
	var withinIndia bool
	indiaBounds := "POLYGON((68 6, 97 6, 97 37, 68 37, 68 6))" // Approximate India bounding box
	if err := s.db.Raw("SELECT ST_Within(ST_GeomFromText(?, 4326), ST_GeomFromText(?, 4326))", wkt, indiaBounds).Scan(&withinIndia).Error; err != nil {
		// Don't fail if this check fails, just log a warning
		return nil
	}
	if !withinIndia {
		// This is just a warning, not an error
		// Could be logged or returned as a warning in the response
	}

	return nil
}

func (s *FarmServiceImpl) filterFarmsByArea(farms []*farmEntity.Farm, minArea, maxArea *float64) []*farmEntity.Farm {
	if minArea == nil && maxArea == nil {
		return farms
	}

	filtered := make([]*farmEntity.Farm, 0)
	for _, farm := range farms {
		if minArea != nil && farm.AreaHa < *minArea {
			continue
		}
		if maxArea != nil && farm.AreaHa > *maxArea {
			continue
		}
		filtered = append(filtered, farm)
	}
	return filtered
}

func (s *FarmServiceImpl) convertFarmToData(farm *farmEntity.Farm) *responses.FarmData {
	return &responses.FarmData{
		ID:              farm.ID,
		AAAFarmerUserID: farm.AAAFarmerUserID,
		AAAOrgID:        farm.AAAOrgID,
		Name:            farm.Name,
		Geometry:        farm.Geometry,
		AreaHa:          farm.AreaHa,
		Metadata:        farm.Metadata,
		CreatedAt:       farm.CreatedAt,
		UpdatedAt:       farm.UpdatedAt,
	}
}
