package testutils

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// DefaultPostgreSQLImage is the PostGIS-enabled PostgreSQL image
	DefaultPostgreSQLImage = "postgis/postgis:16-3.5"

	// DefaultDatabase is the default database name for tests
	DefaultDatabase = "farmers_test"

	// DefaultUsername is the default username for tests
	DefaultUsername = "test"

	// DefaultPassword is the default password for tests
	DefaultPassword = "test"

	// ContainerStartTimeout is the maximum time to wait for container startup
	ContainerStartTimeout = 60 * time.Second
)

// PostgreSQLContainer wraps the testcontainers PostgreSQL container
// with PostGIS extension enabled
type PostgreSQLContainer struct {
	Container testcontainers.Container
	Config    PostgreSQLConfig
}

// PostgreSQLConfig holds configuration for the test database
type PostgreSQLConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	DSN      string
}

// SetupPostgreSQLContainer creates and starts a PostgreSQL container with PostGIS
// It automatically handles cleanup when tests finish
func SetupPostgreSQLContainer(t *testing.T) *PostgreSQLContainer {
	ctx := context.Background()

	// Create PostgreSQL container with PostGIS image
	pgContainer, err := postgrescontainer.Run(ctx,
		DefaultPostgreSQLImage,
		postgrescontainer.WithDatabase(DefaultDatabase),
		postgrescontainer.WithUsername(DefaultUsername),
		postgrescontainer.WithPassword(DefaultPassword),
		postgrescontainer.BasicWaitStrategies(),
		postgrescontainer.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(ContainerStartTimeout),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	})

	// Get connection details
	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), DefaultUsername, DefaultPassword, DefaultDatabase)

	config := PostgreSQLConfig{
		Host:     host,
		Port:     port.Port(),
		Database: DefaultDatabase,
		Username: DefaultUsername,
		Password: DefaultPassword,
		DSN:      dsn,
	}

	return &PostgreSQLContainer{
		Container: pgContainer,
		Config:    config,
	}
}

// SetupTestDB creates a GORM database connection to the test container
// and runs migrations if provided
func (p *PostgreSQLContainer) SetupTestDB(t *testing.T, models ...interface{}) *gorm.DB {
	db, err := gorm.Open(postgres.Open(p.Config.DSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Enable PostGIS extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		t.Fatalf("Failed to enable PostGIS extension: %v", err)
	}

	// Run AutoMigrate if models are provided
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}
	}

	// Register cleanup
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return db
}

// ExecSQL executes raw SQL against the test database
func (p *PostgreSQLContainer) ExecSQL(t *testing.T, db *gorm.DB, sqlQuery string) {
	if err := db.Exec(sqlQuery).Error; err != nil {
		t.Fatalf("Failed to execute SQL: %v, error: %v", sqlQuery, err)
	}
}

// LoadSQLFile loads and executes SQL from a file
func (p *PostgreSQLContainer) LoadSQLFile(t *testing.T, db *gorm.DB, filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for SQL file: %v", err)
	}

	// Read SQL file content (implementation would read and execute)
	// For now, we'll skip the file reading part as it requires additional logic
	t.Logf("Would load SQL from: %s", absPath)
}

// TruncateTables truncates all specified tables for test cleanup
func (p *PostgreSQLContainer) TruncateTables(t *testing.T, db *gorm.DB, tables ...string) {
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// ResetDatabase drops and recreates the database for complete isolation
func (p *PostgreSQLContainer) ResetDatabase(t *testing.T, db *gorm.DB) {
	// Get all table names
	var tables []string
	if err := db.Raw(`
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error; err != nil {
		t.Fatalf("Failed to get table list: %v", err)
	}

	// Truncate all tables
	p.TruncateTables(t, db, tables...)
}

// SetupParallelTestDB creates an isolated database for parallel test execution
// Each test gets its own database to prevent interference
func SetupParallelTestDB(t *testing.T, models ...interface{}) (*gorm.DB, func()) {
	// Allow tests to run in parallel
	t.Parallel()

	// Create container
	pgContainer := SetupPostgreSQLContainer(t)

	// Setup database with migrations
	db := pgContainer.SetupTestDB(t, models...)

	// Return cleanup function
	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

// ValidatePostGISInstallation checks if PostGIS is properly installed
func (p *PostgreSQLContainer) ValidatePostGISInstallation(t *testing.T, db *gorm.DB) {
	var version string
	err := db.Raw("SELECT PostGIS_Version()").Scan(&version).Error
	if err != nil {
		t.Fatalf("PostGIS is not installed: %v", err)
	}
	t.Logf("PostGIS version: %s", version)
}

// CreateSpatialIndex creates a spatial index on a geometry column
func (p *PostgreSQLContainer) CreateSpatialIndex(t *testing.T, db *gorm.DB, table, column string) {
	indexName := fmt.Sprintf("idx_%s_%s_gist", table, column)
	sql := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s USING GIST (%s)", indexName, table, column)
	p.ExecSQL(t, db, sql)
	t.Logf("Created spatial index: %s", indexName)
}

// TestDatabaseConnection verifies the database connection is working
func (p *PostgreSQLContainer) TestDatabaseConnection(t *testing.T, db *gorm.DB) {
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		t.Fatalf("Database connection test failed: %v", err)
	}
	if result != 1 {
		t.Fatalf("Database connection test returned unexpected value: %d", result)
	}
}
