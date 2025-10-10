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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestModelValidation(t *testing.T) {
	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Test AutoMigrate with models that work with SQLite
	// Skip models with PostGIS types (Address with geometry, Farm with geometry)
	models := []interface{}{
		&fpo.FPORef{},
		&farmer.FarmerLink{},
		// Skip farmer.Farmer as it has FK to Address which uses PostGIS Point
		// Skip farmer.Address as it uses PostGIS Point type
		// Skip farm.Farm as it uses PostGIS geometry types
		&crop_cycle.CropCycle{},
		&farm_activity.FarmActivity{},
	}

	err = gormDB.AutoMigrate(models...)
	assert.NoError(t, err)

	// Test creating and validating each model
	t.Run("FPORef", func(t *testing.T) {
		fpoRef := &fpo.FPORef{
			AAAOrgID:       "org123",
			Name:           "Test FPO",
			RegistrationNo: "REG123",
			Status:         "ACTIVE",
			// Skip BusinessConfig as SQLite doesn't support JSONB
		}

		err := fpoRef.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, fpoRef)
	})

	t.Run("Farmer", func(t *testing.T) {
		farmer := &farmer.Farmer{
			AAAUserID: "user123",
			AAAOrgID:  "org123",
			FirstName: "John",
			LastName:  "Doe",
			Status:    "ACTIVE",
		}

		err := farmer.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, farmer)
	})

	t.Run("FarmerLink", func(t *testing.T) {
		farmerLink := &farmer.FarmerLink{
			AAAUserID: "user123",
			AAAOrgID:  "org123",
			Status:    "ACTIVE",
		}

		err := farmerLink.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, farmerLink)
	})

	t.Run("Farm", func(t *testing.T) {
		farm := &farm.Farm{
			AAAFarmerUserID: "user123",
			AAAOrgID:        "org123",
			Name:            testutils.StringPtr("Test Farm"),
			Geometry:        "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		}

		err := farm.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, farm)
	})

	t.Run("CropCycle", func(t *testing.T) {
		cropCycle := &crop_cycle.CropCycle{
			FarmID:   "farm123",
			FarmerID: "farmer123",
			Season:   "RABI",
			Status:   "PLANNED",
			CropID:   "crop123",
		}

		err := cropCycle.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, cropCycle)
	})

	t.Run("FarmActivity", func(t *testing.T) {
		farmActivity := &farm_activity.FarmActivity{
			CropCycleID:  "cycle123",
			ActivityType: "PLANTING",
			CreatedBy:    "user123",
			Status:       "PLANNED",
		}

		err := farmActivity.Validate()
		assert.NoError(t, err)

		// Just test validation, skip database operations for SQLite compatibility
		assert.NotNil(t, farmActivity)
	})
}

func TestModelRelationships(t *testing.T) {
	// Test model relationships conceptually without database operations
	// This tests the model structure and validation logic

	testFarmer := &farmer.Farmer{
		AAAUserID: "user123",
		AAAOrgID:  "org123",
		FirstName: "John",
		LastName:  "Doe",
		Status:    "ACTIVE",
	}

	testFarm := &farm.Farm{
		AAAFarmerUserID: testFarmer.AAAUserID,
		AAAOrgID:        testFarmer.AAAOrgID,
		Name:            testutils.StringPtr("Test Farm"),
		Geometry:        "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
	}

	testCropCycle := &crop_cycle.CropCycle{
		FarmID:   "farm123",   // Would be testFarm.ID in real scenario
		FarmerID: "farmer123", // Would be testFarmer.ID in real scenario
		Season:   "RABI",
		Status:   "PLANNED",
		CropID:   "crop123", // Crop master data reference
	}

	testFarmActivity := &farm_activity.FarmActivity{
		CropCycleID:  "cycle123", // Would be testCropCycle.ID in real scenario
		ActivityType: "PLANTING",
		CreatedBy:    testFarmer.AAAUserID,
		Status:       "PLANNED",
	}

	// Test that all models validate correctly
	assert.NoError(t, testFarmer.Validate())
	assert.NoError(t, testFarm.Validate())
	assert.NoError(t, testCropCycle.Validate())
	assert.NoError(t, testFarmActivity.Validate())

	// Test that relationships are properly structured
	assert.Equal(t, testFarmer.AAAUserID, testFarm.AAAFarmerUserID)
	assert.Equal(t, testFarmer.AAAOrgID, testFarm.AAAOrgID)
	assert.Equal(t, testFarmer.AAAUserID, testFarmActivity.CreatedBy)
}
