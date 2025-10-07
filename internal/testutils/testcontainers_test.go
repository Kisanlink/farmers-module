package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupPostgreSQLContainer(t *testing.T) {
	// Skip if running in CI without Docker
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("creates and starts PostgreSQL container", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		require.NotNil(t, pgContainer)
		require.NotNil(t, pgContainer.Container)

		// Verify configuration
		assert.NotEmpty(t, pgContainer.Config.Host)
		assert.NotEmpty(t, pgContainer.Config.Port)
		assert.Equal(t, DefaultDatabase, pgContainer.Config.Database)
		assert.Equal(t, DefaultUsername, pgContainer.Config.Username)
		assert.NotEmpty(t, pgContainer.Config.DSN)
	})
}

func TestPostgreSQLContainer_SetupTestDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("creates database connection and enables PostGIS", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		require.NotNil(t, db)

		// Verify PostGIS extension is enabled
		var extExists bool
		err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'postgis')").Scan(&extExists).Error
		require.NoError(t, err)
		assert.True(t, extExists, "PostGIS extension should be installed")
	})

	t.Run("runs migrations for provided models", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)

		// Define a test model
		type TestModel struct {
			ID   string `gorm:"primaryKey"`
			Name string
		}

		db := pgContainer.SetupTestDB(t, &TestModel{})
		require.NotNil(t, db)

		// Verify table was created
		var tableExists bool
		err := db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'test_models')").Scan(&tableExists).Error
		require.NoError(t, err)
		assert.True(t, tableExists, "TestModel table should be created")
	})
}

func TestPostgreSQLContainer_ValidatePostGISInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("validates PostGIS is installed and working", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Should not panic or fail
		pgContainer.ValidatePostGISInstallation(t, db)
	})
}

func TestPostgreSQLContainer_TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("verifies database connection", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Should not panic or fail
		pgContainer.TestDatabaseConnection(t, db)
	})
}

func TestPostgreSQLContainer_ExecSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("executes SQL queries", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Create a test table
		pgContainer.ExecSQL(t, db, "CREATE TABLE test_exec (id SERIAL PRIMARY KEY, name VARCHAR(100))")

		// Verify table exists
		var tableExists bool
		err := db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'test_exec')").Scan(&tableExists).Error
		require.NoError(t, err)
		assert.True(t, tableExists)
	})
}

func TestPostgreSQLContainer_TruncateTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("truncates specified tables", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Create and populate a test table
		pgContainer.ExecSQL(t, db, "CREATE TABLE test_truncate (id SERIAL PRIMARY KEY, name VARCHAR(100))")
		pgContainer.ExecSQL(t, db, "INSERT INTO test_truncate (name) VALUES ('test1'), ('test2')")

		// Verify data exists
		var count int64
		db.Raw("SELECT COUNT(*) FROM test_truncate").Scan(&count)
		assert.Equal(t, int64(2), count)

		// Truncate table
		pgContainer.TruncateTables(t, db, "test_truncate")

		// Verify table is empty
		db.Raw("SELECT COUNT(*) FROM test_truncate").Scan(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestPostgreSQLContainer_CreateSpatialIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("creates spatial index on geometry column", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Create table with geometry column
		pgContainer.ExecSQL(t, db, `
			CREATE TABLE test_spatial (
				id SERIAL PRIMARY KEY,
				location GEOMETRY(POINT, 4326)
			)
		`)

		// Create spatial index
		pgContainer.CreateSpatialIndex(t, db, "test_spatial", "location")

		// Verify index exists
		var indexExists bool
		err := db.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM pg_indexes
				WHERE tablename = 'test_spatial'
				AND indexname = 'idx_test_spatial_location_gist'
			)
		`).Scan(&indexExists).Error
		require.NoError(t, err)
		assert.True(t, indexExists, "Spatial index should be created")
	})
}

func TestSetupParallelTestDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	// Define a test model
	type ParallelTestModel struct {
		ID   string `gorm:"primaryKey"`
		Name string
	}

	t.Run("parallel test 1", func(t *testing.T) {
		db, cleanup := SetupParallelTestDB(t, &ParallelTestModel{})
		defer cleanup()

		require.NotNil(t, db)

		// Each parallel test should have its own isolated database
		err := db.Create(&ParallelTestModel{ID: "1", Name: "Test 1"}).Error
		assert.NoError(t, err)
	})

	t.Run("parallel test 2", func(t *testing.T) {
		db, cleanup := SetupParallelTestDB(t, &ParallelTestModel{})
		defer cleanup()

		require.NotNil(t, db)

		// This should not conflict with parallel test 1
		err := db.Create(&ParallelTestModel{ID: "1", Name: "Test 2"}).Error
		assert.NoError(t, err)
	})
}

func TestPostgreSQLContainer_ResetDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestContainers test in short mode")
	}

	t.Run("resets database by truncating all tables", func(t *testing.T) {
		pgContainer := SetupPostgreSQLContainer(t)
		db := pgContainer.SetupTestDB(t)

		// Create multiple tables with data
		pgContainer.ExecSQL(t, db, "CREATE TABLE reset_test1 (id SERIAL PRIMARY KEY, name VARCHAR(100))")
		pgContainer.ExecSQL(t, db, "CREATE TABLE reset_test2 (id SERIAL PRIMARY KEY, value VARCHAR(100))")
		pgContainer.ExecSQL(t, db, "INSERT INTO reset_test1 (name) VALUES ('test')")
		pgContainer.ExecSQL(t, db, "INSERT INTO reset_test2 (value) VALUES ('value')")

		// Reset database
		pgContainer.ResetDatabase(t, db)

		// Verify all tables are empty
		var count1, count2 int64
		db.Raw("SELECT COUNT(*) FROM reset_test1").Scan(&count1)
		db.Raw("SELECT COUNT(*) FROM reset_test2").Scan(&count2)
		assert.Equal(t, int64(0), count1)
		assert.Equal(t, int64(0), count2)
	})
}
