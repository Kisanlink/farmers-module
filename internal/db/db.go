package db

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
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

	if !postgisAvailable {
		log.Println("PostGIS not available - skipping spatial features")
		// For now, skip the farm entity that requires PostGIS
		models := []interface{}{
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{}, // Add the main Farmer model
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}
	} else {
		// Enable PostGIS extension
		if err := gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
			log.Printf("Warning: PostGIS extension not available: %v", err)
			log.Println("Continuing without PostGIS - some spatial features may not work")
		} else {
			log.Println("PostGIS extension enabled successfully")
		}

		// Create custom ENUMs
		createEnums(gormDB)

		// AutoMigrate all models including farm
		models := []interface{}{
			&fpo.FPORef{},
			&farmer.FarmerLink{},
			&farmer.Farmer{}, // Add the main Farmer model
			&farm.Farm{},
			&crop_cycle.CropCycle{},
			&farm_activity.FarmActivity{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}

		// Post-migration setup
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
		log.Println("PostGIS spatial features configured")
	} else {
		log.Println("PostGIS not available - skipping spatial features")
	}

	// Create regular indexes
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (aaa_farmer_user_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_fpo_id_idx ON farms (aaa_org_id);`)

	// Create indexes for other tables
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_links_aaa_user_id_idx ON farmer_links (aaa_user_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farmer_links_aaa_org_id_idx ON farmer_links (aaa_org_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_farm_id_idx ON crop_cycles (farm_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS crop_cycles_farmer_id_idx ON crop_cycles (farmer_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farm_activities_crop_cycle_id_idx ON farm_activities (crop_cycle_id);`)

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
