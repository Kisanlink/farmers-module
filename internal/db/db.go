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
	// Get the GORM DB instance
	gormDB, err := postgresManager.GetDB(context.Background(), false)
	if err != nil {
		return fmt.Errorf("failed to get GORM DB: %w", err)
	}
	if gormDB == nil {
		return fmt.Errorf("GORM DB not available")
	}

	// Enable PostGIS extension
	if err := gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
		return fmt.Errorf("failed to enable PostGIS extension: %w", err)
	}

	// Create custom ENUMs
	createEnums(gormDB)

	// AutoMigrate models using kisanlink-db
	models := []interface{}{
		&fpo.FPORef{},
		&farmer.FarmerLink{},
		&farm.Farm{},
		&crop_cycle.CropCycle{},
		&farm_activity.FarmActivity{},
	}

	if err := postgresManager.AutoMigrateModels(context.Background(), models...); err != nil {
		return fmt.Errorf("failed to run AutoMigrate: %w", err)
	}

	// Post-migration setup
	setupPostMigration(gormDB)

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
	// Add computed area column for farms
	gormDB.Exec(`ALTER TABLE farms ADD COLUMN IF NOT EXISTS area_ha NUMERIC(12,4)
		GENERATED ALWAYS AS (ST_Area(geom::geography)/10000.0) STORED;`)

	// Create spatial indexes
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_geom_gist ON farms USING GIST (geom);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (farmer_id);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_fpo_id_idx ON farms (fpo_id);`)

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
