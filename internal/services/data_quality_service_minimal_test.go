package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/auth"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	farmerRepo "github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDataQualityService_ValidateGeometry_Minimal tests the ValidateGeometry method with minimal dependencies
func TestDataQualityService_ValidateGeometry_Minimal(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}
	mockNotificationService := &MockNotificationService{}

	// Setup AAA service mock to allow permission
	mockAAAService.On("CheckPermission", mock.Anything, "user123", "farm", "audit", "", "org123").Return(true, nil)

	// Create service with nil repositories (will use basic validation)
	service := NewDataQualityService(nil, nil, nil, mockAAAService, mockNotificationService)

	t.Run("Valid polygon geometry without PostGIS", func(t *testing.T) {
		req := &requests.ValidateGeometryRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "user123",
				OrgID:     "org123",
				RequestID: "req123",
			},
			WKT:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			CheckBounds: false,
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testuser",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		response, err := service.ValidateGeometry(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		validateResponse, ok := response.(*responses.ValidateGeometryResponse)
		assert.True(t, ok)
		assert.True(t, validateResponse.IsValid)
		assert.Equal(t, req.WKT, validateResponse.WKT)
		assert.Equal(t, req.RequestID, validateResponse.RequestID)
		assert.Equal(t, 4326, validateResponse.SRID)
	})

	t.Run("Empty WKT geometry", func(t *testing.T) {
		req := &requests.ValidateGeometryRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "user123",
				OrgID:     "org123",
				RequestID: "req124",
			},
			WKT:         "",
			CheckBounds: false,
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testuser",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		response, err := service.ValidateGeometry(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		validateResponse, ok := response.(*responses.ValidateGeometryResponse)
		assert.True(t, ok)
		assert.False(t, validateResponse.IsValid)
		assert.Contains(t, validateResponse.Errors, "geometry cannot be empty")
	})

	t.Run("Invalid geometry type", func(t *testing.T) {
		req := &requests.ValidateGeometryRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "user123",
				OrgID:     "org123",
				RequestID: "req125",
			},
			WKT:         "POINT(0 0)",
			CheckBounds: false,
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testuser",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		response, err := service.ValidateGeometry(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		validateResponse, ok := response.(*responses.ValidateGeometryResponse)
		assert.True(t, ok)
		assert.False(t, validateResponse.IsValid)
		assert.Contains(t, validateResponse.Errors, "only POLYGON geometries are supported")
	})

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}

// TestDataQualityService_ReconcileAAALinks_Minimal tests the ReconcileAAALinks method with minimal dependencies
func TestDataQualityService_ReconcileAAALinks_Minimal(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}
	mockNotificationService := &MockNotificationService{}

	// Setup AAA service mock to allow permission
	mockAAAService.On("CheckPermission", mock.Anything, "admin123", "admin", "maintain", "", "org123").Return(true, nil)

	// Create in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the FarmerLink model
	err = db.AutoMigrate(&farmerentity.FarmerLink{})
	assert.NoError(t, err)

	// Create real repository
	farmerLinkageRepo := &farmerRepo.FarmerLinkRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*farmerentity.FarmerLink](),
	}
	farmerLinkageRepo.SetDBManager(&mockDBManager{db: db})

	// Create service with real repository
	service := NewDataQualityService(db, nil, farmerLinkageRepo, mockAAAService, mockNotificationService)

	t.Run("Successful reconciliation with no links", func(t *testing.T) {
		req := &requests.ReconcileAAALinksRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "admin123",
				OrgID:     "org123",
				RequestID: "req123",
			},
			DryRun: false,
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testadmin",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		response, err := service.ReconcileAAALinks(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		reconcileResponse, ok := response.(*responses.ReconcileAAALinksResponse)
		assert.True(t, ok)
		assert.Equal(t, 0, reconcileResponse.ProcessedLinks)
		assert.Equal(t, 0, reconcileResponse.FixedLinks)
		assert.Equal(t, 0, reconcileResponse.BrokenLinks)
		assert.Equal(t, req.RequestID, reconcileResponse.RequestID)
	})

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}

// TestDataQualityService_RebuildSpatialIndexes_Minimal tests the RebuildSpatialIndexes method with minimal dependencies
func TestDataQualityService_RebuildSpatialIndexes_Minimal(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}
	mockNotificationService := &MockNotificationService{}

	// Setup AAA service mock to allow permission
	mockAAAService.On("CheckPermission", mock.Anything, "admin123", "admin", "maintain", "", "org123").Return(true, nil)

	// Create service with nil repositories (will report no PostGIS)
	service := NewDataQualityService(nil, nil, nil, mockAAAService, mockNotificationService)

	t.Run("Rebuild without database connection", func(t *testing.T) {
		req := &requests.RebuildSpatialIndexesRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "admin123",
				OrgID:     "org123",
				RequestID: "req123",
			},
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testadmin",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		response, err := service.RebuildSpatialIndexes(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		rebuildResponse, ok := response.(*responses.RebuildSpatialIndexesResponse)
		assert.True(t, ok)
		assert.Equal(t, req.RequestID, rebuildResponse.RequestID)
		assert.Empty(t, rebuildResponse.RebuiltIndexes)
		assert.Contains(t, rebuildResponse.Errors, "database connection not available")
	})

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}

// TestDataQualityService_DetectFarmOverlaps_Minimal tests the DetectFarmOverlaps method with minimal dependencies
func TestDataQualityService_DetectFarmOverlaps_Minimal(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}
	mockNotificationService := &MockNotificationService{}

	// Setup AAA service mock to allow permission
	mockAAAService.On("CheckPermission", mock.Anything, "user123", "farm", "audit", "", "org123").Return(true, nil)

	// Create service with nil repositories (will fail without PostGIS)
	service := NewDataQualityService(nil, nil, nil, mockAAAService, mockNotificationService)

	t.Run("Detect overlaps without database connection", func(t *testing.T) {
		req := &requests.DetectFarmOverlapsRequest{
			BaseRequest: requests.BaseRequest{
				UserID:    "user123",
				OrgID:     "org123",
				RequestID: "req123",
			},
		}

		// Setup context with user information
		ctx := context.Background()
		userCtx := &auth.UserContext{
			AAAUserID: req.UserID,
			Username:  "testuser",
			Roles:     []string{"admin"},
		}
		ctx = auth.SetUserInContext(ctx, userCtx)

		_, err := service.DetectFarmOverlaps(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection not available")
	})

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}
