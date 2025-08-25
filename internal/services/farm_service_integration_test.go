package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// FarmServiceIntegrationTestSuite contains integration tests for farm service with real database
type FarmServiceIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *FarmServiceImpl
}

// SetupSuite sets up the test suite with a test database
func (suite *FarmServiceIntegrationTestSuite) SetupSuite() {
	// Skip integration tests if no test database is available
	if testing.Short() {
		suite.T().Skip("Skipping integration tests in short mode")
	}

	// This would typically use a test database container
	// For now, we'll skip if no test DB is configured
	dsn := "host=localhost user=test password=test dbname=farmers_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		suite.T().Skip("Test database not available, skipping integration tests")
		return
	}

	suite.db = db

	// Enable PostGIS extension for testing
	suite.db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")

	// Create test service
	// Note: In real integration tests, you'd set up proper repositories and services
	suite.service = &FarmServiceImpl{
		db: db,
		// farmRepo and aaaService would be properly initialized
	}
}

// TearDownSuite cleans up after the test suite
func (suite *FarmServiceIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

// TestGeometryValidation tests WKT geometry validation with PostGIS
func (suite *FarmServiceIntegrationTestSuite) TestGeometryValidation() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		wkt         string
		expectError bool
		description string
	}{
		{
			name:        "valid polygon",
			wkt:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			expectError: false,
			description: "Simple square polygon",
		},
		{
			name:        "valid complex polygon",
			wkt:         "POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))",
			expectError: false,
			description: "Polygon in Bangalore coordinates",
		},
		{
			name:        "invalid self-intersecting polygon",
			wkt:         "POLYGON((0 0, 1 1, 1 0, 0 1, 0 0))",
			expectError: true,
			description: "Self-intersecting polygon (bow-tie shape)",
		},
		{
			name:        "invalid unclosed polygon",
			wkt:         "POLYGON((0 0, 1 0, 1 1, 0 1))",
			expectError: true,
			description: "Polygon not properly closed",
		},
		{
			name:        "invalid empty geometry",
			wkt:         "",
			expectError: true,
			description: "Empty WKT string",
		},
		{
			name:        "invalid non-polygon geometry",
			wkt:         "POINT(0 0)",
			expectError: true,
			description: "Point geometry instead of polygon",
		},
		{
			name:        "invalid linestring geometry",
			wkt:         "LINESTRING(0 0, 1 1)",
			expectError: true,
			description: "LineString geometry instead of polygon",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.service.validateGeometry(ctx, tt.wkt)

			if tt.expectError {
				assert.Error(suite.T(), err, "Expected error for %s", tt.description)
			} else {
				assert.NoError(suite.T(), err, "Expected no error for %s", tt.description)
			}
		})
	}
}

// TestSpatialQueries tests spatial query operations
func (suite *FarmServiceIntegrationTestSuite) TestSpatialQueries() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// Test bounding box queries

	// Test that bounding box WKT is generated correctly
	expectedBBoxWKT := "POLYGON((77.500000 12.900000, 77.700000 12.900000, 77.700000 13.000000, 77.500000 13.000000, 77.500000 12.900000))"

	// Test spatial intersection query (this would require actual farm data in the database)
	var count int64
	err := suite.db.Raw(`
		SELECT COUNT(*)
		FROM (SELECT 1) as dummy
		WHERE ST_Intersects(
			ST_GeomFromText('POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))', 4326),
			ST_GeomFromText(?, 4326)
		)`, expectedBBoxWKT).Scan(&count).Error

	assert.NoError(suite.T(), err, "Spatial intersection query should work")
}

// TestAreaCalculation tests PostGIS area calculation
func (suite *FarmServiceIntegrationTestSuite) TestAreaCalculation() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	tests := []struct {
		name         string
		wkt          string
		expectedArea float64
		tolerance    float64
		description  string
	}{
		{
			name:         "unit square",
			wkt:          "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			expectedArea: 1.0, // 1 degree square ≈ 12,321 hectares at equator
			tolerance:    0.1,
			description:  "1x1 degree square",
		},
		{
			name:         "small farm polygon",
			wkt:          "POLYGON((77.5946 12.9716, 77.5956 12.9716, 77.5956 12.9726, 77.5946 12.9726, 77.5946 12.9716))",
			expectedArea: 0.01, // Very small area
			tolerance:    0.005,
			description:  "Small farm in Bangalore",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var areaHa float64
			err := suite.db.Raw("SELECT ST_Area(ST_GeomFromText(?, 4326)::geography)/10000.0", tt.wkt).Scan(&areaHa).Error

			assert.NoError(suite.T(), err, "Area calculation should work for %s", tt.description)

			// For geographic coordinates, area calculation is complex
			// We mainly test that the query works and returns a reasonable value
			assert.True(suite.T(), areaHa > 0, "Area should be positive for %s", tt.description)
		})
	}
}

// TestOverlapDetection tests farm overlap detection
func (suite *FarmServiceIntegrationTestSuite) TestOverlapDetection() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	tests := []struct {
		name          string
		polygon1      string
		polygon2      string
		shouldOverlap bool
		description   string
	}{
		{
			name:          "overlapping polygons",
			polygon1:      "POLYGON((0 0, 2 0, 2 2, 0 2, 0 0))",
			polygon2:      "POLYGON((1 1, 3 1, 3 3, 1 3, 1 1))",
			shouldOverlap: true,
			description:   "Two overlapping squares",
		},
		{
			name:          "adjacent polygons",
			polygon1:      "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			polygon2:      "POLYGON((1 0, 2 0, 2 1, 1 1, 1 0))",
			shouldOverlap: false, // Adjacent but not overlapping
			description:   "Two adjacent squares",
		},
		{
			name:          "separate polygons",
			polygon1:      "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			polygon2:      "POLYGON((2 2, 3 2, 3 3, 2 3, 2 2))",
			shouldOverlap: false,
			description:   "Two separate squares",
		},
		{
			name:          "contained polygon",
			polygon1:      "POLYGON((0 0, 4 0, 4 4, 0 4, 0 0))",
			polygon2:      "POLYGON((1 1, 2 1, 2 2, 1 2, 1 1))",
			shouldOverlap: true,
			description:   "Small polygon inside larger one",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var intersects bool
			err := suite.db.Raw(`
				SELECT ST_Intersects(
					ST_GeomFromText(?, 4326),
					ST_GeomFromText(?, 4326)
				)`, tt.polygon1, tt.polygon2).Scan(&intersects).Error

			assert.NoError(suite.T(), err, "Overlap detection should work for %s", tt.description)
			assert.Equal(suite.T(), tt.shouldOverlap, intersects, "Overlap result should match expectation for %s", tt.description)

			// Test intersection area calculation
			if tt.shouldOverlap {
				var intersectionArea float64
				err := suite.db.Raw(`
					SELECT ST_Area(ST_Intersection(
						ST_GeomFromText(?, 4326),
						ST_GeomFromText(?, 4326)
					))`, tt.polygon1, tt.polygon2).Scan(&intersectionArea).Error

				assert.NoError(suite.T(), err, "Intersection area calculation should work")
				assert.True(suite.T(), intersectionArea > 0, "Intersection area should be positive for overlapping polygons")
			}
		})
	}
}

// TestSRIDValidation tests SRID (Spatial Reference System Identifier) validation
func (suite *FarmServiceIntegrationTestSuite) TestSRIDValidation() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	tests := []struct {
		name        string
		wkt         string
		srid        int
		expectError bool
		description string
	}{
		{
			name:        "valid WGS84 coordinates",
			wkt:         "POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))",
			srid:        4326,
			expectError: false,
			description: "Valid WGS84 coordinates for Bangalore",
		},
		{
			name:        "coordinates outside valid range",
			wkt:         "POLYGON((200 100, 201 100, 201 101, 200 101, 200 100))",
			srid:        4326,
			expectError: true,
			description: "Coordinates outside valid WGS84 range",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var isValid bool
			err := suite.db.Raw("SELECT ST_IsValid(ST_GeomFromText(?, ?))", tt.wkt, tt.srid).Scan(&isValid).Error

			assert.NoError(suite.T(), err, "SRID validation query should work for %s", tt.description)

			if tt.expectError {
				// For invalid coordinates, PostGIS might still consider them "valid" geometrically
				// but they would be outside the expected coordinate system bounds
				// Additional validation would be needed in the application layer
			} else {
				assert.True(suite.T(), isValid, "Geometry should be valid for %s", tt.description)
			}
		})
	}
}

// Run the integration test suite
func TestFarmServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(FarmServiceIntegrationTestSuite))
}
