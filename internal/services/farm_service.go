package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/auth"
	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	farmRepo "github.com/Kisanlink/farmers-module/internal/repo/farm"
	farmerRepo "github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FarmServiceImpl implements FarmService
type FarmServiceImpl struct {
	farmRepo   *farmRepo.FarmRepository
	farmerRepo *farmerRepo.FarmerRepository
	aaaService AAAService
	db         *gorm.DB
}

// NewFarmService creates a new farm service
func NewFarmService(farmRepo *farmRepo.FarmRepository, farmerRepo *farmerRepo.FarmerRepository, aaaService AAAService, db *gorm.DB) FarmService {
	return &FarmServiceImpl{
		farmRepo:   farmRepo,
		farmerRepo: farmerRepo,
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can create farms
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "create", "", createReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Resolve farmer_id if not provided
	farmerID := createReq.FarmerID
	if farmerID == "" && createReq.AAAUserID != "" {
		// Look up farmer_id using farmer repository
		filter := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, createReq.AAAUserID).
			Where("aaa_org_id", base.OpEqual, createReq.AAAOrgID).
			Build()

		farmer, err := s.farmerRepo.FindOne(ctx, filter)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("no farmer found for aaa_user_id: %s and aaa_org_id: %s", createReq.AAAUserID, createReq.AAAOrgID)
			}
			return nil, fmt.Errorf("failed to lookup farmer: %w", err)
		}
		farmerID = farmer.ID
	}

	// Validate geometry if provided
	var geometryWKT string
	if createReq.Geometry.WKT != "" {
		if err := s.validateGeometry(ctx, createReq.Geometry.WKT); err != nil {
			return nil, fmt.Errorf("geometry validation failed: %w", err)
		}
		geometryWKT = createReq.Geometry.WKT
	}

	// Create farm entity with proper BaseModel initialization
	farm := farmEntity.NewFarm()
	farm.FarmerID = farmerID
	farm.AAAUserID = createReq.AAAUserID
	farm.AAAOrgID = createReq.AAAOrgID
	farm.Name = createReq.Name
	farm.OwnershipType = farmEntity.OwnershipType(createReq.OwnershipType)
	farm.Geometry = geometryWKT
	farm.SoilTypeID = createReq.SoilTypeID
	farm.PrimaryIrrigationSourceID = createReq.PrimaryIrrigationSourceID
	farm.BoreWellCount = createReq.BoreWellCount
	farm.OtherIrrigationDetails = createReq.OtherIrrigationDetails
	farm.Metadata = createReq.Metadata

	// Set name from metadata if provided (legacy support)
	if nameVal, exists := createReq.Metadata["name"]; exists && farm.Name == nil {
		if name, ok := nameVal.(string); ok {
			farm.Name = &name
		}
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can update this farm
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "update", existingFarm.ID, existingFarm.AAAOrgID)
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

	// Update name if provided
	if updateReq.Name != nil {
		existingFarm.Name = updateReq.Name
	}

	// Update ownership type if provided
	if updateReq.OwnershipType != nil {
		existingFarm.OwnershipType = farmEntity.OwnershipType(*updateReq.OwnershipType)
	}

	// Update soil type if provided
	if updateReq.SoilTypeID != nil {
		existingFarm.SoilTypeID = updateReq.SoilTypeID
	}

	// Update irrigation source if provided
	if updateReq.PrimaryIrrigationSourceID != nil {
		existingFarm.PrimaryIrrigationSourceID = updateReq.PrimaryIrrigationSourceID
	}

	// Update bore well count if provided
	if updateReq.BoreWellCount != nil {
		existingFarm.BoreWellCount = *updateReq.BoreWellCount
	}

	// Update other irrigation details if provided
	if updateReq.OtherIrrigationDetails != nil {
		existingFarm.OtherIrrigationDetails = updateReq.OtherIrrigationDetails
	}

	if updateReq.Metadata != nil {
		existingFarm.Metadata = updateReq.Metadata
		// Update name if provided in metadata (legacy support)
		if nameVal, exists := updateReq.Metadata["name"]; exists && updateReq.Name == nil {
			if name, ok := nameVal.(string); ok {
				existingFarm.Name = &name
			}
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can delete this farm
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "delete", existingFarm.ID, existingFarm.AAAOrgID)
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can list farms
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "list", "", listReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := make(map[string]interface{})
	if listReq.FarmerID != "" {
		filters["farmer_id"] = listReq.FarmerID
	}
	if listReq.AAAUserID != "" {
		filters["aaa_user_id"] = listReq.AAAUserID
	}
	if listReq.AAAOrgID != "" {
		filters["aaa_org_id"] = listReq.AAAOrgID
	}

	// Build filter for database query
	filterBuilder := base.NewFilterBuilder().Page(listReq.Page, listReq.PageSize)

	// Add filters
	if listReq.FarmerID != "" {
		filterBuilder = filterBuilder.Where("farmer_id", base.OpEqual, listReq.FarmerID)
	}
	if listReq.AAAUserID != "" {
		filterBuilder = filterBuilder.Where("aaa_user_id", base.OpEqual, listReq.AAAUserID)
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can read this farm
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "read", farm.ID, farm.AAAOrgID)
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
	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can list farms
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "list", "", filters.AAAOrgID)
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
	if filters.FarmerID != "" {
		query = query.Where("farmer_id = ?", filters.FarmerID)
	}
	if filters.AAAUserID != "" {
		query = query.Where("aaa_user_id = ?", filters.AAAUserID)
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
	// Use the request's built-in validation
	if err := req.Validate(); err != nil {
		return err
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

	// Business Rule 7.1: Farm geometry business validation
	const (
		MaxFarmSizeHa = 100.0 // Maximum farm size in hectares
		MinFarmSizeHa = 0.01  // Minimum farm size in hectares (100 sq meters)
	)

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

	// Business Rule 7.1: Validate farm area size
	// Calculate area in hectares using ST_Area with geography cast
	var areaHa float64
	if err := s.db.Raw(`
		SELECT ST_Area(ST_GeomFromText(?, 4326)::geography) / 10000.0 AS area_ha
	`, wkt).Scan(&areaHa).Error; err != nil {
		return fmt.Errorf("failed to calculate farm area: %w", err)
	}

	if areaHa > MaxFarmSizeHa {
		return fmt.Errorf("farm size %.2f ha exceeds maximum allowed size of %.2f ha", areaHa, MaxFarmSizeHa)
	}
	if areaHa < MinFarmSizeHa {
		return fmt.Errorf("farm size %.2f ha is below minimum required size of %.2f ha", areaHa, MinFarmSizeHa)
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
	farmName := ""
	if farm.Name != nil {
		farmName = *farm.Name
	}
	return &responses.FarmData{
		ID:             farm.ID,
		FarmerID:       farm.FarmerID,
		AAAUserID:      farm.AAAUserID,
		AAAOrgID:       farm.AAAOrgID,
		Name:           farmName,
		Geometry:       farm.Geometry,
		AreaHa:         farm.AreaHa,
		AreaHaComputed: farm.AreaHaComputed,
		Metadata:       farm.Metadata,
		CreatedAt:      farm.CreatedAt,
		UpdatedAt:      farm.UpdatedAt,
	}
}
