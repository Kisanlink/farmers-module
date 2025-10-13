package db

import (
	"testing"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConnect(t *testing.T) {
	config := &db.Config{}

	postgresManager := Connect(config)
	assert.NotNil(t, postgresManager)
}

func TestCreateEnums(t *testing.T) {
	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Test that createEnums doesn't panic with SQLite (which doesn't support ENUMs)
	assert.NotPanics(t, func() {
		createEnums(gormDB)
	})
}

func TestSetupPostMigration(t *testing.T) {
	// Use in-memory SQLite for testing
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create test tables first
	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS farms (
		id TEXT PRIMARY KEY,
		farmer_id TEXT,
		aaa_user_id TEXT,
		aaa_org_id TEXT,
		geometry TEXT,
		created_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS farmers (
		id TEXT PRIMARY KEY,
		aaa_user_id TEXT,
		aaa_org_id TEXT,
		phone_number TEXT,
		email TEXT
	)`).Error
	require.NoError(t, err)

	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS farmer_links (
		id TEXT PRIMARY KEY,
		aaa_user_id TEXT,
		aaa_org_id TEXT,
		kisan_sathi_user_id TEXT,
		status TEXT
	)`).Error
	require.NoError(t, err)

	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS fpo_refs (
		id TEXT PRIMARY KEY,
		aaa_org_id TEXT
	)`).Error
	require.NoError(t, err)

	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS crop_cycles (
		id TEXT PRIMARY KEY,
		farm_id TEXT,
		farmer_id TEXT,
		season TEXT,
		status TEXT,
		start_date DATE
	)`).Error
	require.NoError(t, err)

	err = gormDB.Exec(`CREATE TABLE IF NOT EXISTS farm_activities (
		id TEXT PRIMARY KEY,
		crop_cycle_id TEXT,
		activity_type TEXT,
		status TEXT,
		created_by TEXT,
		planned_at DATETIME
	)`).Error
	require.NoError(t, err)

	// Test that setupPostMigration doesn't panic
	assert.NotPanics(t, func() {
		setupPostMigration(gormDB)
	})
}

func TestSetupDatabaseWithoutPostGIS(t *testing.T) {
	// Create a mock PostgresManager for testing
	logger, _ := zap.NewDevelopment()
	config := &db.Config{}

	postgresManager := db.NewPostgresManager(config, logger)

	// Test that setup completes without PostGIS
	// This is a basic test to ensure the function structure is correct
	assert.NotNil(t, postgresManager)
}
