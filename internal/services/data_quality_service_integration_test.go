package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	farmRepo "github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DataQualityServiceIntegrationTestSuite is the test suite for data quality service integration tests
type DataQualityServiceIntegrationTestSuite struct {
	suite.Suite
	db                *gorm.DB
	service           DataQualityService
	farmRepo          *farmRepo.FarmRepository
	farmerLinkageRepo *base.BaseFilterableRepository[*entities.FarmerLink]
	mockAAAService    *MockAAAService
}

// SetupSuite sets up the test suite
func (suite *DataQualityServiceIntegrationTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&entities.FarmerLink{})
	suite.Require().NoError(err)

	suite.db = db

	// Create repositories
	suite.farmRepo = farmRepo.NewFarmRepository(nil) // No DB manager for this test
	suite.farmerLinkageRepo = base.NewBaseFilterableRepository[*entities.FarmerLink]()
	suite.farmerLinkageRepo.SetDBManager(&mockDBManager{db: db})

	// Create mock AAA service
	suite.mockAAAService = &MockAAAService{}

	// Create service
	suite.service = NewDataQualityService(db, suite.farmRepo, suite.farmerLinkageRepo, suite.mockAAAService)
}

// TearDownSuite tears down the test suite
func (suite *DataQualityServiceIntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest sets up each test
func (suite *DataQualityServiceIntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.Exec("DELETE FROM farmer_links")

	// Reset mock expectations
	suite.mockAAAService.ExpectedCalls = nil
	suite.mockAAAService.Calls = nil
}

// mockDBManager implements the interface expected by BaseFilterableRepository
type mockDBManager struct {
	db *gorm.DB
}

func (m *mockDBManager) GetDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	return m.db, nil
}

func TestDataQualityServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DataQualityServiceIntegrationTestSuite))
}

func (suite *DataQualityServiceIntegrationTestSuite) TestValidateGeometry_WithoutPostGIS() {
	// Setup AAA service mock
	suite.mockAAAService.On("CheckPermission", mock.Anything, "user123", "farm", "audit", "", "org123").Return(true, nil)

	tests := []struct {
		name          string
		wkt           string
		checkBounds   bool
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "Valid polygon",
			wkt:           "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			checkBounds:   false,
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "Empty WKT",
			wkt:           "",
			checkBounds:   false,
			expectedValid: false,
			expectedError: false,
		},
		{
			name:          "Invalid geometry type",
			wkt:           "POINT(0 0)",
			checkBounds:   false,
			expectedValid: false,
			expectedError: false,
		},
		{
			name:          "Valid polygon with bounds check",
			wkt:           "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			checkBounds:   true,
			expectedValid: true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := &requests.ValidateGeometryRequest{
				BaseRequest: requests.BaseRequest{
					UserID:    "user123",
					OrgID:     "org123",
					RequestID: "req123",
				},
				WKT:         tt.wkt,
				CheckBounds: tt.checkBounds,
			}

			response, err := suite.service.ValidateGeometry(context.Background(), req)

			if tt.expectedError {
				suite.Error(err)
			} else {
				suite.NoError(err)
				suite.NotNil(response)

				validateResponse, ok := response.(*responses.ValidateGeometryResponse)
				suite.True(ok)
				suite.Equal(tt.expectedValid, validateResponse.IsValid)
				suite.Equal(tt.wkt, validateResponse.WKT)
				suite.Equal("req123", validateResponse.RequestID)

				if !tt.expectedValid && tt.wkt != "" {
					suite.NotEmpty(validateResponse.Errors)
				}
			}
		})
	}
}

func (suite *DataQualityServiceIntegrationTestSuite) TestReconcileAAALinks_Integration() {
	// Skip this test for now as it requires complex repository setup
	// The functionality is tested in the minimal test instead
	suite.T().Skip("Skipping integration test - repository setup is complex")
}

func (suite *DataQualityServiceIntegrationTestSuite) TestRebuildSpatialIndexes_WithoutPostGIS() {
	// Setup AAA service mock
	suite.mockAAAService.On("CheckPermission", mock.Anything, "admin123", "admin", "maintain", "", "org123").Return(true, nil)

	req := &requests.RebuildSpatialIndexesRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    "admin123",
			OrgID:     "org123",
			RequestID: "req123",
		},
	}

	response, err := suite.service.RebuildSpatialIndexes(context.Background(), req)

	suite.NoError(err)
	suite.NotNil(response)

	rebuildResponse, ok := response.(*responses.RebuildSpatialIndexesResponse)
	suite.True(ok)
	suite.Equal("req123", rebuildResponse.RequestID)
	suite.Empty(rebuildResponse.RebuiltIndexes) // No PostGIS available
	suite.NotEmpty(rebuildResponse.Errors)      // Should have PostGIS not available error
	suite.Contains(rebuildResponse.Errors[0], "failed to check PostGIS availability")
}

func (suite *DataQualityServiceIntegrationTestSuite) TestDetectFarmOverlaps_WithoutPostGIS() {
	// Setup AAA service mock
	suite.mockAAAService.On("CheckPermission", mock.Anything, "user123", "farm", "audit", "", "org123").Return(true, nil)

	req := &requests.DetectFarmOverlapsRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    "user123",
			OrgID:     "org123",
			RequestID: "req123",
		},
		MinOverlapAreaHa: nil,
		Limit:            nil,
	}

	_, err := suite.service.DetectFarmOverlaps(context.Background(), req)

	// Should fail because PostGIS is not available
	suite.Error(err)
	suite.Contains(err.Error(), "failed to check PostGIS availability")
}

func (suite *DataQualityServiceIntegrationTestSuite) TestPermissionDenied() {
	// Reset the mock to clear any previous expectations
	suite.mockAAAService.ExpectedCalls = nil
	suite.mockAAAService.Calls = nil

	// Setup AAA service mock to deny permission
	suite.mockAAAService.On("CheckPermission", mock.Anything, "user123", "farm", "audit", "", "org123").Return(false, nil)

	req := &requests.ValidateGeometryRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    "user123",
			OrgID:     "org123",
			RequestID: "req123",
		},
		WKT:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		CheckBounds: false,
	}

	response, err := suite.service.ValidateGeometry(context.Background(), req)

	suite.Error(err)
	suite.Nil(response)
	suite.Contains(err.Error(), "forbidden")
}
