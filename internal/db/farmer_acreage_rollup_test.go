package db

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// FarmerAcreageRollupTestSuite contains integration tests for farmer total acreage and farm count rollup
// Tests verify that GORM hooks in farm.go automatically maintain farmer.total_acreage_ha and farmer.farm_count
type FarmerAcreageRollupTestSuite struct {
	suite.Suite
	pgContainer *testutils.PostgreSQLContainer
	db          *gorm.DB
}

// SetupSuite sets up the test suite with a test database
func (suite *FarmerAcreageRollupTestSuite) SetupSuite() {
	// Skip integration tests if no test database is available
	if testing.Short() {
		suite.T().Skip("Skipping integration tests in short mode")
	}

	// Setup PostgreSQL container with PostGIS
	suite.pgContainer = testutils.SetupPostgreSQLContainer(suite.T())

	// Setup database with migrations
	suite.db = suite.pgContainer.SetupTestDB(suite.T())

	// Create ENUMs needed for the database
	suite.setupEnums()

	// Run migrations for required entities
	models := []interface{}{
		&farmer.Address{},
		&farmer.Farmer{},
		&farm.Farm{},
	}

	err := suite.db.AutoMigrate(models...)
	require.NoError(suite.T(), err, "Failed to run migrations")

	// Setup post-migration features (computed columns, indexes)
	suite.setupPostMigration()

	// Note: No need to setup SQL triggers - GORM hooks in farm.go will handle updates automatically

	// Validate PostGIS is installed
	suite.pgContainer.ValidatePostGISInstallation(suite.T(), suite.db)
}

// setupEnums creates required ENUM types
func (suite *FarmerAcreageRollupTestSuite) setupEnums() {
	// Create farmer_status enum
	suite.db.Exec(`DO $$ BEGIN
		CREATE TYPE farmer_status AS ENUM ('ACTIVE','INACTIVE','SUSPENDED');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)
}

// setupPostMigration sets up computed columns and indexes
func (suite *FarmerAcreageRollupTestSuite) setupPostMigration() {
	// Add computed area column for farms using geography for accurate area calculation
	suite.db.Exec(`ALTER TABLE farms ADD COLUMN IF NOT EXISTS area_ha_computed NUMERIC(12,4)
		GENERATED ALWAYS AS (ST_Area(geometry::geography)/10000.0) STORED;`)

	// Create spatial indexes
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS farms_geometry_gist ON farms USING GIST (geometry::geometry);`)

	// Add geometry constraints
	suite.db.Exec(`ALTER TABLE farms ADD CONSTRAINT IF NOT EXISTS farms_geometry_srid_check
		CHECK (ST_SRID(geometry) = 4326);`)
	suite.db.Exec(`ALTER TABLE farms ADD CONSTRAINT IF NOT EXISTS farms_geometry_valid_check
		CHECK (ST_IsValid(geometry));`)

	// Create regular indexes
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (farmer_id);`)
	suite.db.Exec(`CREATE INDEX IF NOT EXISTS idx_farmers_total_acreage ON farmers (total_acreage_ha);`)
}

// TearDownTest cleans up after each test
func (suite *FarmerAcreageRollupTestSuite) TearDownTest() {
	// Clean up test data between tests
	suite.db.Exec("DELETE FROM farms")
	suite.db.Exec("DELETE FROM farmers")
	suite.db.Exec("DELETE FROM addresses")
}

// Helper function to create a test farmer
func (suite *FarmerAcreageRollupTestSuite) createTestFarmer(name string) *farmer.Farmer {
	testFarmer := farmer.NewFarmer()
	testFarmer.AAAUserID = fmt.Sprintf("user_%s", name)
	testFarmer.AAAOrgID = "test_org"
	testFarmer.FirstName = name
	testFarmer.LastName = "Test"
	testFarmer.PhoneNumber = "1234567890"
	testFarmer.Email = fmt.Sprintf("%s@test.com", name)
	testFarmer.Status = "ACTIVE"
	testFarmer.TotalAcreageHa = 0.0 // Should start at 0

	err := suite.db.Create(testFarmer).Error
	require.NoError(suite.T(), err, "Failed to create test farmer")

	return testFarmer
}

// Helper function to create a test farm with specific area
func (suite *FarmerAcreageRollupTestSuite) createTestFarm(farmerID string, name string, areaApproxHa float64) *farm.Farm {
	// Create a polygon that will have approximately the specified area
	// Using a square polygon near the equator where degree calculations are simpler
	// Note: At equator, 1 degree ≈ 111 km, so for area calculation:
	// Area = width_deg * height_deg * (111 km)^2 / 100 (to convert to hectares)
	// For 1 hectare, we need roughly 0.003 degree sides
	sideDegrees := 0.003 * areaApproxHa // Rough approximation

	geometry := fmt.Sprintf("POLYGON((0 0, %f 0, %f %f, 0 %f, 0 0))",
		sideDegrees, sideDegrees, sideDegrees, sideDegrees)

	testFarm := farm.NewFarm()
	testFarm.FarmerID = farmerID
	testFarm.AAAUserID = "test_user"
	testFarm.AAAOrgID = "test_org"
	farmName := name
	testFarm.Name = &farmName
	testFarm.Geometry = geometry
	testFarm.OwnershipType = farm.OwnershipOwn

	err := suite.db.Create(testFarm).Error
	require.NoError(suite.T(), err, "Failed to create test farm")

	// Reload to get computed area
	suite.db.First(testFarm, "id = ?", testFarm.ID)

	return testFarm
}

// Helper function to get farmer's current total acreage
func (suite *FarmerAcreageRollupTestSuite) getFarmerTotalAcreage(farmerID string) float64 {
	var totalAcreage float64
	err := suite.db.Model(&farmer.Farmer{}).
		Where("id = ?", farmerID).
		Pluck("total_acreage_ha", &totalAcreage).Error
	require.NoError(suite.T(), err, "Failed to get farmer total acreage")
	return totalAcreage
}

// Test 1: Initial state - New farmer has total_acreage_ha = 0.0
func (suite *FarmerAcreageRollupTestSuite) TestInitialFarmerHasZeroAcreage() {
	// Create a new farmer
	testFarmer := suite.createTestFarmer("initial_test")

	// Verify initial total acreage is 0
	totalAcreage := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.Equal(suite.T(), 0.0, totalAcreage, "New farmer should have total_acreage_ha = 0.0")
}

// Test 2: Insert farm - Creating a farm updates the farmer's total acreage
func (suite *FarmerAcreageRollupTestSuite) TestInsertFarmUpdatesFarmerTotal() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("insert_test")

	// Create a farm
	testFarm := suite.createTestFarm(testFarmer.ID, "Test Farm 1", 2.5)

	// Verify farmer's total acreage is updated
	totalAcreage := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Due to geography calculations, exact area may vary slightly
	assert.InDelta(suite.T(), testFarm.AreaHa, totalAcreage, 0.01,
		"Farmer total acreage should equal farm area after insert")
}

// Test 3: Update farm geometry - Changing farm boundary recalculates area and updates farmer total
func (suite *FarmerAcreageRollupTestSuite) TestUpdateFarmGeometryUpdatesFarmerTotal() {
	// Create a farmer and farm
	testFarmer := suite.createTestFarmer("update_geometry_test")
	testFarm := suite.createTestFarm(testFarmer.ID, "Test Farm", 1.0)

	initialTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Update farm geometry to a larger area
	newGeometry := "POLYGON((0 0, 0.006 0, 0.006 0.006, 0 0.006, 0 0))" // ~4x larger
	err := suite.db.Model(&farm.Farm{}).
		Where("id = ?", testFarm.ID).
		Update("geometry", newGeometry).Error
	require.NoError(suite.T(), err, "Failed to update farm geometry")

	// Get updated total
	newTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Verify total increased
	assert.Greater(suite.T(), newTotal, initialTotal,
		"Farmer total acreage should increase after farm area increases")
}

// Test 4: Insert multiple farms - Adding multiple farms correctly sums all areas
func (suite *FarmerAcreageRollupTestSuite) TestMultipleFarmsSumCorrectly() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("multiple_farms_test")

	// Create multiple farms with different areas
	farm1 := suite.createTestFarm(testFarmer.ID, "Farm 1", 1.5)
	farm2 := suite.createTestFarm(testFarmer.ID, "Farm 2", 2.0)
	farm3 := suite.createTestFarm(testFarmer.ID, "Farm 3", 0.5)

	// Get farmer's total acreage
	totalAcreage := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Calculate expected sum
	expectedSum := farm1.AreaHa + farm2.AreaHa + farm3.AreaHa

	// Verify total equals sum of all farms
	assert.InDelta(suite.T(), expectedSum, totalAcreage, 0.1,
		"Farmer total acreage should equal sum of all farm areas")
}

// Test 5: Delete farm - Hard deleting a farm decreases farmer's total acreage
func (suite *FarmerAcreageRollupTestSuite) TestHardDeleteFarmUpdatesFarmerTotal() {
	// Create a farmer with two farms
	testFarmer := suite.createTestFarmer("delete_test")
	farm1 := suite.createTestFarm(testFarmer.ID, "Farm 1", 2.0)
	farm2 := suite.createTestFarm(testFarmer.ID, "Farm 2", 3.0)

	initialTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Hard delete one farm
	err := suite.db.Unscoped().Delete(&farm.Farm{}, "id = ?", farm1.ID).Error
	require.NoError(suite.T(), err, "Failed to hard delete farm")

	// Get updated total
	newTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Verify total decreased by the deleted farm's area
	assert.InDelta(suite.T(), farm2.AreaHa, newTotal, 0.01,
		"Farmer total should equal remaining farm area after deletion")
	assert.Less(suite.T(), newTotal, initialTotal,
		"Farmer total should decrease after farm deletion")
}

// Test 6: Soft delete farm - Soft deleting a farm excludes it from total
func (suite *FarmerAcreageRollupTestSuite) TestSoftDeleteFarmExcludesFromTotal() {
	// Create a farmer with two farms
	testFarmer := suite.createTestFarmer("soft_delete_test")
	farm1 := suite.createTestFarm(testFarmer.ID, "Farm 1", 2.0)
	farm2 := suite.createTestFarm(testFarmer.ID, "Farm 2", 3.0)

	initialTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Soft delete one farm (set deleted_at)
	now := time.Now()
	err := suite.db.Model(&farm.Farm{}).
		Where("id = ?", farm1.ID).
		Update("deleted_at", now).Error
	require.NoError(suite.T(), err, "Failed to soft delete farm")

	// Get updated total
	newTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Verify soft-deleted farm is excluded from total
	assert.InDelta(suite.T(), farm2.AreaHa, newTotal, 0.01,
		"Farmer total should exclude soft-deleted farm")
	assert.Less(suite.T(), newTotal, initialTotal,
		"Farmer total should decrease after soft deletion")
}

// Test 7: Undelete farm - Clearing deleted_at includes the farm in total again
func (suite *FarmerAcreageRollupTestSuite) TestUndeleteFarmIncludesInTotal() {
	// Create a farmer with a soft-deleted farm
	testFarmer := suite.createTestFarmer("undelete_test")
	testFarm := suite.createTestFarm(testFarmer.ID, "Farm 1", 2.0)

	// Soft delete the farm
	now := time.Now()
	err := suite.db.Model(&farm.Farm{}).
		Where("id = ?", testFarm.ID).
		Update("deleted_at", now).Error
	require.NoError(suite.T(), err)

	// Verify farm is excluded
	deletedTotal := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.Equal(suite.T(), 0.0, deletedTotal, "Soft-deleted farm should be excluded")

	// Undelete the farm (clear deleted_at)
	err = suite.db.Model(&farm.Farm{}).
		Where("id = ?", testFarm.ID).
		Update("deleted_at", gorm.Expr("NULL")).Error
	require.NoError(suite.T(), err, "Failed to undelete farm")

	// Get updated total
	restoredTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Verify farm is included again
	assert.InDelta(suite.T(), testFarm.AreaHa, restoredTotal, 0.01,
		"Undeleted farm should be included in total again")
}

// Test 8: Move farm between farmers - Changing farmer_id updates both old and new farmer totals
func (suite *FarmerAcreageRollupTestSuite) TestMoveFarmBetweenFarmers() {
	// Create two farmers
	farmer1 := suite.createTestFarmer("farmer1")
	farmer2 := suite.createTestFarmer("farmer2")

	// Create farms for farmer1
	farm1 := suite.createTestFarm(farmer1.ID, "Farm 1", 2.0)
	farm2 := suite.createTestFarm(farmer1.ID, "Farm 2", 3.0)

	// Verify initial totals
	farmer1InitialTotal := suite.getFarmerTotalAcreage(farmer1.ID)
	farmer2InitialTotal := suite.getFarmerTotalAcreage(farmer2.ID)

	assert.InDelta(suite.T(), farm1.AreaHa+farm2.AreaHa, farmer1InitialTotal, 0.01,
		"Farmer1 should have both farms initially")
	assert.Equal(suite.T(), 0.0, farmer2InitialTotal,
		"Farmer2 should have no farms initially")

	// Move farm1 from farmer1 to farmer2
	err := suite.db.Model(&farm.Farm{}).
		Where("id = ?", farm1.ID).
		Update("farmer_id", farmer2.ID).Error
	require.NoError(suite.T(), err, "Failed to move farm between farmers")

	// Get updated totals
	farmer1NewTotal := suite.getFarmerTotalAcreage(farmer1.ID)
	farmer2NewTotal := suite.getFarmerTotalAcreage(farmer2.ID)

	// Verify both farmers' totals are updated correctly
	assert.InDelta(suite.T(), farm2.AreaHa, farmer1NewTotal, 0.01,
		"Farmer1 should only have farm2 after move")
	assert.InDelta(suite.T(), farm1.AreaHa, farmer2NewTotal, 0.01,
		"Farmer2 should have farm1 after move")
}

// Test 9: Edge case - farmer with no farms should have total_acreage_ha = 0.0
func (suite *FarmerAcreageRollupTestSuite) TestFarmerWithNoFarms() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("no_farms")

	// Create and then delete all farms
	suite.createTestFarm(testFarmer.ID, "Farm 1", 2.0)
	suite.createTestFarm(testFarmer.ID, "Farm 2", 3.0)

	// Verify farmer has farms initially
	initialTotal := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.Greater(suite.T(), initialTotal, 0.0, "Farmer should have farms initially")

	// Delete all farms
	err := suite.db.Unscoped().Delete(&farm.Farm{}, "farmer_id = ?", testFarmer.ID).Error
	require.NoError(suite.T(), err, "Failed to delete all farms")

	// Verify total is now 0
	finalTotal := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.Equal(suite.T(), 0.0, finalTotal,
		"Farmer with no farms should have total_acreage_ha = 0.0")
}

// Test 10: Concurrent operations - Multiple farms being added/removed concurrently
func (suite *FarmerAcreageRollupTestSuite) TestConcurrentFarmOperations() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("concurrent_test")

	// Number of concurrent operations
	numOperations := 10
	var wg sync.WaitGroup
	wg.Add(numOperations)

	// Channel to collect created farm IDs
	farmIDs := make(chan string, numOperations)

	// Concurrently create farms
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			defer wg.Done()

			// Create a farm with area = index + 1 hectares
			farmName := fmt.Sprintf("Concurrent Farm %d", index)
			geometry := fmt.Sprintf("POLYGON((0 0, %f 0, %f %f, 0 %f, 0 0))",
				0.003*float64(index+1), 0.003*float64(index+1),
				0.003*float64(index+1), 0.003*float64(index+1))

			testFarm := farm.NewFarm()
			testFarm.FarmerID = testFarmer.ID
			testFarm.AAAUserID = "test_user"
			testFarm.AAAOrgID = "test_org"
			testFarm.Name = &farmName
			testFarm.Geometry = geometry
			testFarm.OwnershipType = farm.OwnershipOwn

			err := suite.db.Create(testFarm).Error
			if err != nil {
				suite.T().Errorf("Failed to create farm concurrently: %v", err)
				return
			}

			farmIDs <- testFarm.ID
		}(i)
	}

	// Wait for all farms to be created
	wg.Wait()
	close(farmIDs)

	// Collect all farm IDs
	var allFarmIDs []string
	for id := range farmIDs {
		allFarmIDs = append(allFarmIDs, id)
	}

	// Verify all farms were created
	assert.Equal(suite.T(), numOperations, len(allFarmIDs),
		"All concurrent farm creations should succeed")

	// Get the final total acreage
	finalTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Calculate expected total by summing actual farm areas
	var expectedTotal float64
	err := suite.db.Model(&farm.Farm{}).
		Where("farmer_id = ? AND deleted_at IS NULL", testFarmer.ID).
		Select("COALESCE(SUM(area_ha_computed), 0)").
		Scan(&expectedTotal).Error
	require.NoError(suite.T(), err)

	// Verify the trigger-maintained total matches the calculated sum
	assert.InDelta(suite.T(), expectedTotal, finalTotal, 0.01,
		"Trigger-maintained total should match calculated sum after concurrent operations")

	// Now concurrently delete half the farms
	halfCount := numOperations / 2
	wg.Add(halfCount)

	for i := 0; i < halfCount; i++ {
		go func(farmID string) {
			defer wg.Done()

			// Soft delete the farm
			now := time.Now()
			err := suite.db.Model(&farm.Farm{}).
				Where("id = ?", farmID).
				Update("deleted_at", now).Error
			if err != nil {
				suite.T().Errorf("Failed to delete farm concurrently: %v", err)
			}
		}(allFarmIDs[i])
	}

	// Wait for deletions to complete
	wg.Wait()

	// Get the updated total after deletions
	afterDeleteTotal := suite.getFarmerTotalAcreage(testFarmer.ID)

	// Verify total decreased after deletions
	assert.Less(suite.T(), afterDeleteTotal, finalTotal,
		"Total should decrease after concurrent deletions")

	// Calculate expected total after deletions
	var expectedAfterDelete float64
	err = suite.db.Model(&farm.Farm{}).
		Where("farmer_id = ? AND deleted_at IS NULL", testFarmer.ID).
		Select("COALESCE(SUM(area_ha_computed), 0)").
		Scan(&expectedAfterDelete).Error
	require.NoError(suite.T(), err)

	// Verify the trigger-maintained total still matches
	assert.InDelta(suite.T(), expectedAfterDelete, afterDeleteTotal, 0.01,
		"Trigger-maintained total should remain accurate after concurrent deletions")
}

// Additional test: Verify total never goes negative
func (suite *FarmerAcreageRollupTestSuite) TestTotalNeverNegative() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("never_negative")

	// Try various operations that might cause issues
	// 1. Create and delete farms
	farm1 := suite.createTestFarm(testFarmer.ID, "Farm 1", 2.0)
	suite.db.Unscoped().Delete(&farm.Farm{}, "id = ?", farm1.ID)

	total := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.GreaterOrEqual(suite.T(), total, 0.0,
		"Total acreage should never be negative")

	// 2. Delete non-existent farm (should not affect total)
	suite.db.Unscoped().Delete(&farm.Farm{}, "id = ?", "non-existent-id")

	total = suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.GreaterOrEqual(suite.T(), total, 0.0,
		"Total acreage should remain non-negative after invalid operations")
}

// Table-driven test for various polygon sizes and shapes
func (suite *FarmerAcreageRollupTestSuite) TestVariousPolygonShapes() {
	testCases := []struct {
		name        string
		geometry    string
		description string
	}{
		{
			name:        "small_square",
			geometry:    "POLYGON((0 0, 0.001 0, 0.001 0.001, 0 0.001, 0 0))",
			description: "Small square polygon",
		},
		{
			name:        "rectangle",
			geometry:    "POLYGON((0 0, 0.002 0, 0.002 0.001, 0 0.001, 0 0))",
			description: "Rectangular polygon",
		},
		{
			name:        "triangle",
			geometry:    "POLYGON((0 0, 0.002 0, 0.001 0.002, 0 0))",
			description: "Triangular polygon",
		},
		{
			name:        "pentagon",
			geometry:    "POLYGON((0 0.001, 0.001 0, 0.002 0.001, 0.0015 0.002, 0.0005 0.002, 0 0.001))",
			description: "Pentagon polygon",
		},
		{
			name:        "real_world_bangalore",
			geometry:    "POLYGON((77.5946 12.9716, 77.5956 12.9716, 77.5956 12.9726, 77.5946 12.9726, 77.5946 12.9716))",
			description: "Real-world coordinates in Bangalore",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create a farmer for this test case
			testFarmer := suite.createTestFarmer(fmt.Sprintf("shape_%s", tc.name))

			// Create a farm with the specific geometry
			testFarm := farm.NewFarm()
			testFarm.FarmerID = testFarmer.ID
			testFarm.AAAUserID = "test_user"
			testFarm.AAAOrgID = "test_org"
			farmName := tc.description
			testFarm.Name = &farmName
			testFarm.Geometry = tc.geometry
			testFarm.OwnershipType = farm.OwnershipOwn

			err := suite.db.Create(testFarm).Error
			require.NoError(suite.T(), err, "Failed to create farm with %s", tc.description)

			// Reload to get computed area
			suite.db.First(testFarm, "id = ?", testFarm.ID)

			// Verify the area is calculated and positive
			assert.Greater(suite.T(), testFarm.AreaHa, 0.0,
				"Farm area should be positive for %s", tc.description)

			// Verify farmer's total equals farm area
			totalAcreage := suite.getFarmerTotalAcreage(testFarmer.ID)
			assert.InDelta(suite.T(), testFarm.AreaHa, totalAcreage, 0.01,
				"Farmer total should equal farm area for %s", tc.description)
		})
	}
}

// Test that verifies the trigger handles NULL geometry gracefully
func (suite *FarmerAcreageRollupTestSuite) TestNullGeometryHandling() {
	// Create a farmer
	testFarmer := suite.createTestFarmer("null_geometry")

	// Try to create a farm with NULL geometry (should fail due to constraints)
	// But if it somehow gets through, the trigger should handle it
	testFarm := farm.NewFarm()
	testFarm.FarmerID = testFarmer.ID
	testFarm.AAAUserID = "test_user"
	testFarm.AAAOrgID = "test_org"
	farmName := "Null Geometry Farm"
	testFarm.Name = &farmName
	testFarm.Geometry = "" // Empty geometry
	testFarm.OwnershipType = farm.OwnershipOwn

	// This should fail due to validation
	err := suite.db.Create(testFarm).Error
	assert.Error(suite.T(), err, "Creating farm with empty geometry should fail")

	// Farmer's total should remain 0
	totalAcreage := suite.getFarmerTotalAcreage(testFarmer.ID)
	assert.Equal(suite.T(), 0.0, totalAcreage,
		"Farmer total should remain 0 when farm creation fails")
}

// Run the test suite
func TestFarmerAcreageRollupSuite(t *testing.T) {
	suite.Run(t, new(FarmerAcreageRollupTestSuite))
}
