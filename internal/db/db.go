package db

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
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
	"gorm.io/gorm"
)

// Connect establishes a connection to PostgreSQL using kisanlink-db PostgresManager
func Connect(config *db.Config) *db.PostgresManager {
	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Create PostgresManager
	postgresManager := db.NewPostgresManager(config, logger)

	log.Println("PostgresManager created successfully")
	return postgresManager
}

// SetupDatabase runs migrations and setup for the farmers module
func SetupDatabase(postgresManager *db.PostgresManager) error {
	// First connect to the database
	ctx := context.Background()
	if err := postgresManager.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	// Note: We don't close the connection here as it's needed by the server
	// The connection will be closed when the main function exits

	// Get the GORM DB instance
	gormDB, err := postgresManager.GetDB(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get GORM DB: %w", err)
	}
	if gormDB == nil {
		return fmt.Errorf("GORM DB not available")
	}

	// Check if PostGIS is available before proceeding
	var postgisAvailable bool
	if err := gormDB.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		log.Printf("Warning: Could not check PostGIS availability: %v", err)
		postgisAvailable = false
	}

	// Create custom ENUMs first (needed regardless of PostGIS availability)
	createEnums(gormDB)

	if !postgisAvailable {
		log.Println("PostGIS not available - skipping spatial features")
		// Skip the farm entity that requires PostGIS geometry types
		models := []interface{}{
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{},          // Add the main Farmer model
			&entities.Address{},       // Add Address entity
			&entities.FarmerProfile{}, // Add FarmerProfile entity
			&crop.Crop{},
			&crop_variety.CropVariety{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},
			&farm_soil_type.FarmSoilType{},
			&farm_irrigation_source.FarmIrrigationSource{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}

		// Post-migration setup (without PostGIS features)
		setupPostMigration(gormDB)
	} else {
		// Enable PostGIS extension
		if err := gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
			log.Printf("Warning: PostGIS extension not available: %v", err)
			log.Println("Continuing without PostGIS - some spatial features may not work")
		} else {
			log.Println("PostGIS extension enabled successfully")
		}

		// AutoMigrate all models including farm
		models := []interface{}{
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{},          // Add the main Farmer model
			&entities.Address{},       // Add Address entity
			&entities.FarmerProfile{}, // Add FarmerProfile entity
			&farm.Farm{},
			&crop.Crop{},
			&crop_variety.CropVariety{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},
			&farm_soil_type.FarmSoilType{},
			&farm_irrigation_source.FarmIrrigationSource{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}

		// Post-migration setup (with PostGIS features)
		setupPostMigration(gormDB)
	}

	log.Println("Database setup completed successfully")
	return nil
}

// createEnums creates custom ENUM types for the database
func createEnums(gormDB *gorm.DB) {
	// Season enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE season AS ENUM ('RABI','KHARIF','ZAID','OTHER');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	// Crop category enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE crop_category AS ENUM ('CEREALS','PULSES','VEGETABLES','FRUITS','OIL_SEEDS','SPICES','CASH_CROPS','FODDER','MEDICINAL','OTHER');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	// Cycle status enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE cycle_status AS ENUM ('PLANNED','ACTIVE','COMPLETED','CANCELLED');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	// Activity status enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE activity_status AS ENUM ('PLANNED','COMPLETED','CANCELLED');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	// Link status enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE link_status AS ENUM ('ACTIVE','INACTIVE');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	// Farmer status enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE farmer_status AS ENUM ('ACTIVE','INACTIVE','SUSPENDED');
	EXCEPTION WHEN duplicate_object THEN NULL; END $$;`)

	log.Println("Custom ENUM types created successfully")
}

// setupPostMigration sets up computed columns, indexes, and constraints
func setupPostMigration(gormDB *gorm.DB) {
	// Check if PostGIS is available before setting up spatial features
	var postgisAvailable bool
	if err := gormDB.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
		log.Printf("Warning: Could not check PostGIS availability: %v", err)
		postgisAvailable = false
	}

	if postgisAvailable {
		// Add computed area column for farms with PostGIS
		gormDB.Exec(`ALTER TABLE farms ADD COLUMN IF NOT EXISTS area_ha_computed NUMERIC(12,4)
			GENERATED ALWAYS AS (ST_Area(geometry::geometry)/10000.0) STORED;`)

		// Create spatial indexes
		gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_geometry_gist ON farms USING GIST (geometry::geometry);`)

		// Add SRID validation constraint
		gormDB.Exec(`ALTER TABLE farms ADD CONSTRAINT IF NOT EXISTS farms_geometry_srid_check
			CHECK (ST_SRID(geometry) = 4326);`)

		// Add geometry validity constraint
		gormDB.Exec(`ALTER TABLE farms ADD CONSTRAINT IF NOT EXISTS farms_geometry_valid_check
			CHECK (ST_IsValid(geometry));`)

		log.Println("PostGIS spatial features configured")
	} else {
		log.Println("PostGIS not available - skipping spatial features")
	}

	// Create regular indexes (only create farm indexes if PostGIS is available)
	if postgisAvailable {
		gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (aaa_farmer_user_id);`)
		gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_fpo_id_idx ON farms (aaa_org_id);`)
		gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_created_at_idx ON farms (created_at);`)
	}

	// Create indexes for farmer tables
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS farmers_aaa_user_org_idx ON farmers (aaa_user_id, aaa_org_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmers_phone_idx ON farmers (phone_number);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmers_email_idx ON farmers (email);`)

	// Create indexes for farmer_profiles table
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS farmer_profiles_aaa_user_org_idx ON farmer_profiles (aaa_user_id, aaa_org_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_profiles_phone_idx ON farmer_profiles (phone_number);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_profiles_email_idx ON farmer_profiles (email);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_profiles_address_id_idx ON farmer_profiles (address_id);`)

	// Create indexes for addresses table
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS addresses_city_idx ON addresses (city);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS addresses_state_idx ON addresses (state);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS addresses_postal_code_idx ON addresses (postal_code);`)

	// Create indexes for farmer_links table
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS farmer_links_user_org_idx ON farmer_links (aaa_user_id, aaa_org_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_links_kisan_sathi_idx ON farmer_links (kisan_sathi_user_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_links_status_idx ON farmer_links (status);`)

	// Create indexes for fpo_refs table
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS fpo_refs_aaa_org_id_idx ON fpo_refs (aaa_org_id);`)

	// Create indexes for crops table
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS crops_name_idx ON crops (name);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crops_category_idx ON crops (category);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crops_is_active_idx ON crops (is_active);`)

	// Create indexes for crop_varieties table
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_varieties_crop_id_idx ON crop_varieties (crop_id);`)
	gormDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS crop_varieties_crop_name_idx ON crop_varieties (crop_id, name);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_varieties_is_active_idx ON crop_varieties (is_active);`)

	// Create indexes for crop_cycles table
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_farm_id_idx ON crop_cycles (farm_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_farmer_id_idx ON crop_cycles (farmer_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_crop_id_idx ON crop_cycles (crop_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_variety_id_idx ON crop_cycles (variety_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_season_idx ON crop_cycles (season);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_status_idx ON crop_cycles (status);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_start_date_idx ON crop_cycles (start_date);`)

	// Create indexes for farm_activities table
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_crop_cycle_id_idx ON farm_activities (crop_cycle_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_type_idx ON farm_activities (activity_type);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_status_idx ON farm_activities (status);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_created_by_idx ON farm_activities (created_by);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_planned_at_idx ON farm_activities (planned_at);`)

	log.Println("Post-migration setup completed")
}

// Close closes the database connection
func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error getting underlying sql.DB: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}
