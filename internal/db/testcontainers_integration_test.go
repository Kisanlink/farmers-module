package db

import (
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostGISIntegration tests PostGIS-specific functionality using TestContainers
func TestPostGISIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostGIS integration test in short mode")
	}

	// Setup PostgreSQL container with PostGIS
	pgContainer := testutils.SetupPostgreSQLContainer(t)

	// Define all models for migration
	models := []interface{}{
		&fpo.FPORef{},
		&farmer.FarmerLink{},
		&farmer.Farmer{},
		&farm.Farm{},
		&crop_cycle.CropCycle{},
		&farm_activity.FarmActivity{},
	}

	// Setup database with migrations
	db := pgContainer.SetupTestDB(t, models...)

	// Validate PostGIS installation
	pgContainer.ValidatePostGISInstallation(t, db)

	t.Run("validates geometry data types", func(t *testing.T) {
		// Create a test farm with geometry
		testFarm := &farm.Farm{
			AAAUserID: "user123",
			AAAOrgID:  "org123",
			Name:      testutils.StringPtr("Test Farm"),
			Geometry:  "POLYGON((77.5946 12.9716, 77.6046 12.9716, 77.6046 12.9816, 77.5946 12.9816, 77.5946 12.9716))",
		}

		// Insert the farm
		err := db.Create(testFarm).Error
		require.NoError(t, err)

		// Retrieve the farm
		var retrieved farm.Farm
		err = db.Where("id = ?", testFarm.ID).First(&retrieved).Error
		require.NoError(t, err)

		// Verify geometry is preserved
		assert.NotEmpty(t, retrieved.Geometry)
	})

	t.Run("performs spatial queries", func(t *testing.T) {
		// Create spatial index on geometry column
		pgContainer.CreateSpatialIndex(t, db, "farms", "geometry")

		// Test ST_Area calculation
		var area float64
		err := db.Raw(`
			SELECT ST_Area(ST_GeomFromText(?, 4326)::geography) / 10000 as area_hectares
		`, "POLYGON((0 0, 0 0.01, 0.01 0.01, 0.01 0, 0 0))").Scan(&area).Error
		require.NoError(t, err)
		assert.Greater(t, area, 0.0, "Area should be greater than 0")
	})

	t.Run("validates WKT geometry format", func(t *testing.T) {
		testCases := []struct {
			name        string
			wkt         string
			expectValid bool
		}{
			{
				name:        "valid polygon",
				wkt:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				expectValid: true,
			},
			{
				name:        "invalid self-intersecting polygon",
				wkt:         "POLYGON((0 0, 1 1, 1 0, 0 1, 0 0))",
				expectValid: false,
			},
			{
				name:        "valid multipolygon",
				wkt:         "MULTIPOLYGON(((0 0, 1 0, 1 1, 0 1, 0 0)))",
				expectValid: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var isValid bool
				err := db.Raw(`
					SELECT ST_IsValid(ST_GeomFromText(?, 4326))
				`, tc.wkt).Scan(&isValid).Error

				if tc.expectValid {
					require.NoError(t, err)
					assert.True(t, isValid, "Geometry should be valid")
				} else {
					// Self-intersecting polygons may not error but will be invalid
					if err == nil {
						assert.False(t, isValid, "Geometry should be invalid")
					}
				}
			})
		}
	})
}

// TestDatabaseMigrations tests all model migrations using TestContainers
func TestDatabaseMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database migration test in short mode")
	}

	pgContainer := testutils.SetupPostgreSQLContainer(t)

	t.Run("migrates all models successfully", func(t *testing.T) {
		models := []interface{}{
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{},
			&farm.Farm{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},
		}

		db := pgContainer.SetupTestDB(t, models...)

		// Verify all tables exist
		expectedTables := []string{
			"fpo_refs",
			"farmer_links",
			"farmers",
			"farms",
			"crop_cycles",
			"farm_activities",
		}

		for _, table := range expectedTables {
			var exists bool
			err := db.Raw(`
				SELECT EXISTS (
					SELECT 1 FROM information_schema.tables
					WHERE table_name = ?
				)
			`, table).Scan(&exists).Error
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", table)
		}
	})
}

// TestModelRelationshipsWithRealDB tests model relationships with actual database
func TestModelRelationshipsWithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping model relationship test in short mode")
	}

	pgContainer := testutils.SetupPostgreSQLContainer(t)

	models := []interface{}{
		&fpo.FPORef{},
		&farmer.FarmerLink{},
		&farmer.Farmer{},
		&farm.Farm{},
		&crop_cycle.CropCycle{},
		&farm_activity.FarmActivity{},
	}

	db := pgContainer.SetupTestDB(t, models...)

	t.Run("creates and retrieves farmer with FPO relationship", func(t *testing.T) {
		// Create FPO
		testFPO := &fpo.FPORef{
			AAAOrgID:       "org123",
			Name:           "Test FPO",
			RegistrationNo: "REG123",
			Status:         "ACTIVE",
		}
		err := db.Create(testFPO).Error
		require.NoError(t, err)

		// Create Farmer linked to FPO
		testFarmer := &farmer.Farmer{
			AAAUserID: "user123",
			AAAOrgID:  testFPO.AAAOrgID,
			FirstName: "John",
			LastName:  "Doe",
			Status:    "ACTIVE",
		}
		err = db.Create(testFarmer).Error
		require.NoError(t, err)

		// Retrieve and verify
		var retrieved farmer.Farmer
		err = db.Where("id = ?", testFarmer.ID).First(&retrieved).Error
		require.NoError(t, err)
		assert.Equal(t, testFPO.AAAOrgID, retrieved.AAAOrgID)
	})

	t.Run("creates complete farm workflow", func(t *testing.T) {
		// Create FPO
		testFPO := &fpo.FPORef{
			AAAOrgID:       "org456",
			Name:           "Workflow Test FPO",
			RegistrationNo: "WF123",
			Status:         "ACTIVE",
		}
		err := db.Create(testFPO).Error
		require.NoError(t, err)

		// Create Farmer
		testFarmer := &farmer.Farmer{
			AAAUserID: "farmer456",
			AAAOrgID:  testFPO.AAAOrgID,
			FirstName: "Jane",
			LastName:  "Smith",
			Status:    "ACTIVE",
		}
		err = db.Create(testFarmer).Error
		require.NoError(t, err)

		// Create Farm
		testFarm := &farm.Farm{
			AAAUserID: testFarmer.AAAUserID,
			AAAOrgID:  testFPO.AAAOrgID,
			Name:      testutils.StringPtr("Workflow Test Farm"),
			Geometry:  "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		}
		err = db.Create(testFarm).Error
		require.NoError(t, err)

		// Create Crop Cycle
		testCycle := &crop_cycle.CropCycle{
			FarmID:   testFarm.ID,
			FarmerID: testFarmer.ID,
			Season:   "RABI",
			Status:   "PLANNED",
		}
		err = db.Create(testCycle).Error
		require.NoError(t, err)

		// Create Farm Activity
		testActivity := &farm_activity.FarmActivity{
			CropCycleID:  testCycle.ID,
			ActivityType: "PLANTING",
			CreatedBy:    testFarmer.AAAUserID,
			Status:       "PLANNED",
		}
		err = db.Create(testActivity).Error
		require.NoError(t, err)

		// Verify complete workflow
		var activityCount int64
		err = db.Model(&farm_activity.FarmActivity{}).
			Where("crop_cycle_id = ?", testCycle.ID).
			Count(&activityCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), activityCount)
	})
}

// TestParallelDatabaseOperations demonstrates parallel test execution
func TestParallelDatabaseOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parallel database test in short mode")
	}

	t.Run("parallel farmer creation", func(t *testing.T) {
		db, cleanup := testutils.SetupParallelTestDB(t, &farmer.Farmer{})
		defer cleanup()

		testFarmer := &farmer.Farmer{
			AAAUserID: "parallel_user_1",
			AAAOrgID:  "parallel_org_1",
			FirstName: "Parallel",
			LastName:  "Test1",
			Status:    "ACTIVE",
		}

		err := db.Create(testFarmer).Error
		assert.NoError(t, err)
	})

	t.Run("parallel farm creation", func(t *testing.T) {
		db, cleanup := testutils.SetupParallelTestDB(t, &farm.Farm{})
		defer cleanup()

		testFarm := &farm.Farm{
			AAAUserID: "parallel_user_2",
			AAAOrgID:  "parallel_org_2",
			Name:      testutils.StringPtr("Parallel Farm"),
			Geometry:  "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		}

		err := db.Create(testFarm).Error
		assert.NoError(t, err)
	})
}

// TestSpatialIndexes tests spatial index creation and usage
func TestSpatialIndexes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping spatial index test in short mode")
	}

	pgContainer := testutils.SetupPostgreSQLContainer(t)
	db := pgContainer.SetupTestDB(t, &farm.Farm{})

	t.Run("creates and uses spatial index", func(t *testing.T) {
		// Create spatial index
		pgContainer.CreateSpatialIndex(t, db, "farms", "geometry")

		// Insert test farms
		farms := []*farm.Farm{
			{
				AAAUserID: "user1",
				AAAOrgID:  "org1",
				Name:      testutils.StringPtr("Farm 1"),
				Geometry:  "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			},
			{
				AAAUserID: "user2",
				AAAOrgID:  "org1",
				Name:      testutils.StringPtr("Farm 2"),
				Geometry:  "POLYGON((2 2, 3 2, 3 3, 2 3, 2 2))",
			},
		}

		for _, f := range farms {
			err := db.Create(f).Error
			require.NoError(t, err)
		}

		// Perform spatial query using the index
		var nearbyFarms []farm.Farm
		err := db.Raw(`
			SELECT * FROM farms
			WHERE ST_DWithin(
				geometry::geography,
				ST_GeomFromText('POINT(0.5 0.5)', 4326)::geography,
				1000
			)
		`).Scan(&nearbyFarms).Error
		require.NoError(t, err)
		assert.Greater(t, len(nearbyFarms), 0, "Should find nearby farms")
	})
}
