package db

import (
	"context"
	"fmt"
	"log"

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
	"github.com/Kisanlink/farmers-module/internal/entities/fpo_config"
	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
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

	// Create custom ENUMs first (needed regardless of PostGIS availability)
	createEnums(gormDB)

	// Try to enable PostGIS extension (idempotent - safe to run even if already exists)
	log.Println("Attempting to enable PostGIS extension...")
	postgisAvailable := false
	if err := gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
		log.Printf("⚠️  Failed to create PostGIS extension: %v", err)
		log.Println("PostGIS may not be available on this PostgreSQL server")
		postgisAvailable = false
	} else {
		// Verify PostGIS was created successfully
		if err := gormDB.Raw(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')`).Scan(&postgisAvailable).Error; err != nil {
			log.Printf("Warning: Could not verify PostGIS availability: %v", err)
			postgisAvailable = false
		}
	}

	if !postgisAvailable {
		log.Println("❌ PostGIS not available - skipping spatial features (farms table)")
		log.Println("📝 To enable PostGIS on AWS RDS:")
		log.Println("   1. Ensure your RDS instance supports PostGIS (PostgreSQL 12+)")
		log.Println("   2. The database user must have rds_superuser role or CREATE privilege")
		log.Println("   3. PostGIS is installed by default on RDS, just needs CREATE EXTENSION")

		// Fix farmer constraints before AutoMigrate (idempotent)
		fixFarmerConstraints(gormDB)

		// Skip the farm entity that requires PostGIS geometry types
		// Migration order: independent tables first, then tables with FK dependencies
		models := []interface{}{
			// Independent master data tables
			&fpo.FPORef{},
			&fpo.FPOAuditLog{},
			&fpo_config.FPOConfig{},
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},
			&crop.Crop{},

			// Address table (no dependencies)
			&farmer.Address{},

			// Farmer tables (Farmer depends on Address via FK)
			&farmer.Farmer{},
			&farmer.FarmerLink{},

			// Crop variety (depends on Crop)
			&crop_variety.CropVariety{},

			// Stage tables (depends on Crop)
			&stage.Stage{},
			&stage.CropStage{},

			// Farm entity skipped (requires PostGIS)

			// Crop cycle (depends on Farm - skipped without PostGIS)
			// Farm activity (depends on CropCycle - skipped without PostGIS)

			// Bulk operations (last)
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}

		// Post-migration setup (without PostGIS features)
		setupPostMigration(gormDB)

		// Initialize ID counters from existing database records
		if err := initializeCounters(gormDB); err != nil {
			log.Printf("Warning: Failed to initialize ID counters: %v", err)
		}

		log.Println("✅ Database setup completed successfully (without PostGIS)")
		return nil
	} else {
		log.Println("✅ PostGIS extension enabled successfully")

		// Fix farmer constraints before AutoMigrate (idempotent)
		fixFarmerConstraints(gormDB)

		// AutoMigrate all models including farm (PostGIS enabled)
		// Migration order: independent tables first, then tables with FK dependencies
		models := []interface{}{
			// Independent master data tables
			&fpo.FPORef{},
			&fpo.FPOAuditLog{},
			&fpo_config.FPOConfig{},
			&soil_type.SoilType{},
			&irrigation_source.IrrigationSource{},
			&crop.Crop{},

			// Address table (no dependencies)
			&farmer.Address{},

			// Farmer tables (Farmer depends on Address via FK)
			&farmer.Farmer{},
			&farmer.FarmerLink{},

			// Farm (depends on Farmer, uses PostGIS)
			&farm.Farm{},

			// Crop variety (depends on Crop)
			&crop_variety.CropVariety{},

			// Stage tables (depends on Crop)
			&stage.Stage{},
			&stage.CropStage{},

			// Crop cycle (depends on Farm, Farmer, Crop, CropVariety)
			&crop_cycle.CropCycle{},

			// Farm activity (depends on CropCycle)
			&farm_activity.FarmActivity{},

			// Junction tables (depend on Farm and master tables)
			&farm_soil_type.FarmSoilType{},
			&farm_irrigation_source.FarmIrrigationSource{},

			// Bulk operations (last)
			&bulk.BulkOperation{},
			&bulk.ProcessingDetail{},
		}

		if err := postgresManager.AutoMigrateModels(ctx, models...); err != nil {
			return fmt.Errorf("failed to run AutoMigrate: %w", err)
		}

		// Post-migration setup (with PostGIS features)
		setupPostMigration(gormDB)
	}

	// Initialize ID counters from existing database records
	if err := initializeCounters(gormDB); err != nil {
		log.Printf("Warning: Failed to initialize ID counters: %v", err)
		// Don't fail the setup, but log the warning
	}

	log.Println("✅ Database setup completed successfully (with PostGIS)")
	return nil
}

// createEnums creates custom ENUM types for the database
func createEnums(gormDB *gorm.DB) {
	// Season enum
	gormDB.Exec(`DO $$ BEGIN
		CREATE TYPE season AS ENUM ('RABI','KHARIF','ZAID','PERENNIAL','OTHER');
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

// fixFarmerConstraints fixes farmer-related constraints before AutoMigrate
// This is needed because:
// 1. Farmers are now uniquely identified by aaa_user_id only (not composite aaa_user_id + aaa_org_id)
// 2. The old FK constraint fk_farmers_fpo_linkages referenced the composite key
// 3. AutoMigrate won't drop existing constraints, so we do it here (idempotent)
func fixFarmerConstraints(gormDB *gorm.DB) {
	log.Println("Fixing farmer constraints...")

	// Drop old FK constraint on farmer_links that references composite key
	gormDB.Exec(`ALTER TABLE farmer_links DROP CONSTRAINT IF EXISTS fk_farmers_fpo_linkages`)

	// Drop old unique constraint on farmers that used composite (aaa_user_id, aaa_org_id)
	gormDB.Exec(`DROP INDEX IF EXISTS idx_farmer_unique`)

	// Make aaa_org_id nullable on farmers table (it's now optional)
	gormDB.Exec(`ALTER TABLE farmers ALTER COLUMN aaa_org_id DROP NOT NULL`)

	log.Println("Farmer constraints fixed successfully")
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
		// Check if area_ha_computed column exists and drop it to recreate with correct formula
		// This ensures we use geography (accurate area) instead of geometry (square degrees)
		var columnExists bool
		gormDB.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'farms' AND column_name = 'area_ha_computed'
		)`).Scan(&columnExists)

		if columnExists {
			// Drop the existing generated column to recreate with correct formula
			log.Println("Recreating area_ha_computed column with correct geography-based formula")
			gormDB.Exec(`ALTER TABLE farms DROP COLUMN IF EXISTS area_ha_computed;`)
		}

		// Add computed area column for farms with PostGIS using geography for accurate area calculation
		// ST_Area(geometry::geography) returns area in square meters, divide by 10000 for hectares
		gormDB.Exec(`ALTER TABLE farms ADD COLUMN IF NOT EXISTS area_ha_computed NUMERIC(12,4)
			GENERATED ALWAYS AS (ST_Area(geometry::geography)/10000.0) STORED;`)

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
		gormDB.Exec(`CREATE INDEX IF NOT EXISTS farms_farmer_id_idx ON farms (farmer_id);`)
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

	// Create indexes for farmers total_acreage_ha and farm_count (for fast filtering/sorting)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS idx_farmers_total_acreage ON farmers (total_acreage_ha);`)
	gormDB.Exec(`CREATE INDEX IF NOT EXISTS idx_farmers_farm_count ON farmers (farm_count);`)

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

	// Backfill farmer stats from existing farms (only if farms table exists)
	if postgisAvailable {
		backfillFarmerStats(gormDB)
	}

	log.Println("Post-migration setup completed")
}

// backfillFarmerStats updates all existing farmers' total_acreage_ha and farm_count fields
// using GORM. After this backfill, the GORM hooks in farm.go will maintain these fields automatically.
func backfillFarmerStats(gormDB *gorm.DB) {
	log.Println("Backfilling farmer total acreage and farm count from existing farms using GORM...")

	// Get all unique farmer IDs from farms table
	var farmerIDs []string
	if err := gormDB.Model(&farm.Farm{}).
		Distinct("farmer_id").
		Where("deleted_at IS NULL").
		Pluck("farmer_id", &farmerIDs).Error; err != nil {
		log.Printf("Warning: Failed to get farmer IDs for backfill: %v", err)
		return
	}

	log.Printf("Found %d farmers with farms to backfill", len(farmerIDs))

	// Update each farmer's stats using GORM
	successCount := 0
	for _, farmerID := range farmerIDs {
		// Calculate stats for this farmer
		var stats struct {
			TotalAcreage float64
			FarmCount    int64
		}

		err := gormDB.Model(&farm.Farm{}).
			Select("COALESCE(SUM(area_ha_computed), 0) as total_acreage, COUNT(*) as farm_count").
			Where("farmer_id = ? AND deleted_at IS NULL", farmerID).
			Scan(&stats).Error

		if err != nil {
			log.Printf("Warning: Failed to calculate stats for farmer %s: %v", farmerID, err)
			continue
		}

		// Update farmer using GORM
		err = gormDB.Model(&farmer.Farmer{}).
			Where("id = ?", farmerID).
			Updates(map[string]interface{}{
				"total_acreage_ha": stats.TotalAcreage,
				"farm_count":       stats.FarmCount,
			}).Error

		if err != nil {
			log.Printf("Warning: Failed to update farmer %s: %v", farmerID, err)
			continue
		}

		successCount++
	}

	log.Printf("✅ Backfilled total_acreage_ha and farm_count for %d farmers using GORM", successCount)
	log.Println("✅ Future updates will be handled by GORM hooks in farm.go")
}

// initializeCounters initializes ID counters for all tables from existing database records
func initializeCounters(gormDB *gorm.DB) error {
	// Define tables to initialize with their identifiers and sizes
	tables := []struct {
		TableName  string
		Identifier string
		Size       hash.TableSize
	}{
		{"farmers", "FMRR", hash.Large},
		{"addresses", "ADDR", hash.Medium},
		{"farmer_links", "FMLK", hash.Large},
		{"farms", "FARM", hash.Medium},       // Must match Farm.GetTableSize()
		{"crop_cycles", "CRCY", hash.Medium}, // Must match CropCycle.GetTableSize()
		{"farm_activities", "FACT", hash.XLarge},
		{"fpo_refs", "FPOR", hash.Medium},
		{"crops", "CROP", hash.Small},
		{"crop_varieties", "CVAR", hash.Medium},
		{"soil_types", "SOIL", hash.Tiny},
		{"irrigation_sources", "IRRG", hash.Tiny},
		{"bulk_operations", "BLKO", hash.Medium},
		{"bulk_processing_details", "BLKD", hash.Large},
	}

	for _, table := range tables {
		// Query all IDs from the table using Raw SQL for explicit control
		var ids []string
		query := fmt.Sprintf("SELECT id FROM %s ORDER BY id", table.TableName)
		if err := gormDB.Raw(query).Scan(&ids).Error; err != nil {
			// If table doesn't exist or error occurs, skip it
			log.Printf("Skipping counter initialization for %s: %v", table.TableName, err)
			continue
		}

		// Initialize counter using the hash package function
		log.Printf("Initializing counter for %s (%s) with %d existing IDs",
			table.TableName, table.Identifier, len(ids))

		// Debug: Log first few and last few IDs if any exist
		if len(ids) > 0 {
			sampleSize := 3
			if len(ids) < sampleSize {
				sampleSize = len(ids)
			}
			log.Printf("  First IDs from %s: %v", table.TableName, ids[:sampleSize])
			if len(ids) > sampleSize {
				lastIdx := len(ids)
				log.Printf("  Last IDs from %s: %v", table.TableName, ids[lastIdx-sampleSize:])
			}
		} else {
			// Double-check if there are actually records but query failed to get IDs
			var count int64
			if countErr := gormDB.Table(table.TableName).Count(&count).Error; countErr == nil && count > 0 {
				log.Printf("⚠️  WARNING: Found %d records in %s but query returned 0 IDs! Counter may not initialize correctly.", count, table.TableName)
			}
		}

		// Call the initialization function from kisanlink-db
		hash.InitializeGlobalCountersFromDatabase(table.Identifier, ids, table.Size)

		log.Printf("✓ Counter initialized for %s: %d existing records",
			table.TableName, len(ids))
	}

	log.Println("✅ All ID counters initialized successfully")
	return nil
}

// ResetCounter manually resets a single counter based on existing database records
// This is useful for debugging counter issues
func ResetCounter(gormDB *gorm.DB, tableName, identifier string, size hash.TableSize) error {
	var ids []string
	query := fmt.Sprintf("SELECT id FROM %s ORDER BY id", tableName)
	if err := gormDB.Raw(query).Scan(&ids).Error; err != nil {
		return fmt.Errorf("failed to query IDs from %s: %w", tableName, err)
	}

	log.Printf("Resetting counter for %s (%s) with %d existing IDs", tableName, identifier, len(ids))
	if len(ids) > 0 {
		sampleSize := 5
		if len(ids) < sampleSize {
			sampleSize = len(ids)
		}
		log.Printf("  First IDs: %v", ids[:sampleSize])
		if len(ids) > sampleSize {
			lastIdx := len(ids)
			log.Printf("  Last IDs: %v", ids[lastIdx-sampleSize:])
		}
	}

	hash.InitializeGlobalCountersFromDatabase(identifier, ids, size)
	log.Printf("✓ Counter reset for %s: %d existing records", tableName, len(ids))

	return nil
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
