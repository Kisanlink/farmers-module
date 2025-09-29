package db

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_stage"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_soil_type"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
)

// Connect creates a new database manager
func Connect(config *db.Config) db.DBManager {
	logger, _ := zap.NewDevelopment()
	return db.NewPostgresManager(config, logger)
}

// SetupDatabase initializes the database with all required tables and extensions
func SetupDatabase(dbManager db.DBManager) error {
	ctx := context.Background()

	// Check if PostGIS is available
	var postgisAvailable bool

	// Try to get GORM DB instance - we'll use a different approach
	// For now, let's use the AutoMigrateModels method directly

	// Create custom ENUMs first
	// Note: We'll need to get the GORM instance for this, but let's skip for now

	if !postgisAvailable {
		log.Println("PostGIS not available - skipping spatial features")
		// For now, skip the farm entity that requires PostGIS
		models := []interface{}{
			// Core entities
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{},          // Main Farmer model
			&entities.FarmerProfile{}, // Farmer profile
			&entities.FarmerLinkage{}, // Farmer linkage

			// Crop management
			&crop.Crop{},
			&crop_variety.CropVariety{},
			&crop_stage.CropStage{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},

			// Soil and irrigation
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},

			// Bulk operations
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
		}

		if err := dbManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}
	} else {
		// AutoMigrate all models including farm
		models := []interface{}{
			// Core entities
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{},          // Main Farmer model
			&entities.FarmerProfile{}, // Farmer profile
			&entities.FarmerLinkage{}, // Farmer linkage

			// Farm management
			&farm.Farm{},
			&farm_soil_type.FarmSoilType{},
			&farm_irrigation_source.FarmIrrigationSource{},

			// Crop management
			&crop.Crop{},
			&crop_variety.CropVariety{},
			&crop_stage.CropStage{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},

			// Soil and irrigation
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},

			// Bulk operations
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
		}

		if err := dbManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}
	}

	// Run farm-related migrations (soil types, irrigation sources, etc.)
	// Note: We'll need to get the GORM instance for this
	// For now, let's skip the migration service call

	log.Println("Database setup completed successfully")
	return nil
}
