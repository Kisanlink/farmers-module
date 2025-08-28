package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// ConcreteAdministrativeService defines the concrete interface for administrative operations
type ConcreteAdministrativeService interface {
	// SeedRolesAndPermissions triggers a complete reseed of AAA resources, actions, and role bindings
	SeedRolesAndPermissions(ctx context.Context, req *requests.SeedRolesAndPermissionsRequest) (*responses.SeedRolesAndPermissionsResponse, error)

	// HealthCheck verifies database connectivity and AAA service availability
	HealthCheck(ctx context.Context, req *requests.HealthCheckRequest) (*responses.HealthCheckResponse, error)
}

// AdministrativeServiceImpl implements ConcreteAdministrativeService
type AdministrativeServiceImpl struct {
	postgresManager *db.PostgresManager
	gormDB          *gorm.DB
	aaaService      AAAService
}

// NewAdministrativeService creates a new administrative service
func NewAdministrativeService(postgresManager *db.PostgresManager, gormDB *gorm.DB, aaaService AAAService) ConcreteAdministrativeService {
	return &AdministrativeServiceImpl{
		postgresManager: postgresManager,
		gormDB:          gormDB,
		aaaService:      aaaService,
	}
}

// SeedRolesAndPermissions implements the seeding of roles and permissions
func (s *AdministrativeServiceImpl) SeedRolesAndPermissions(ctx context.Context, req *requests.SeedRolesAndPermissionsRequest) (*responses.SeedRolesAndPermissionsResponse, error) {
	startTime := time.Now()

	// Call AAA service to seed roles and permissions
	err := s.aaaService.SeedRolesAndPermissions(ctx)
	if err != nil {
		return &responses.SeedRolesAndPermissionsResponse{
			Success:   false,
			Message:   "Failed to seed roles and permissions",
			Error:     err.Error(),
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}, fmt.Errorf("failed to seed roles and permissions: %w", err)
	}

	return &responses.SeedRolesAndPermissionsResponse{
		Success:   true,
		Message:   "Roles and permissions seeded successfully",
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
	}, nil
}

// HealthCheck implements comprehensive health checking
func (s *AdministrativeServiceImpl) HealthCheck(ctx context.Context, req *requests.HealthCheckRequest) (*responses.HealthCheckResponse, error) {
	startTime := time.Now()
	response := &responses.HealthCheckResponse{
		Status:     "healthy",
		Timestamp:  time.Now(),
		Components: make(map[string]responses.ComponentHealth),
	}

	overallHealthy := true

	// Check database connectivity
	dbHealth := s.checkDatabaseHealth(ctx)
	response.Components["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		overallHealthy = false
	}

	// Check AAA service connectivity
	aaaHealth := s.checkAAAServiceHealth(ctx)
	response.Components["aaa_service"] = aaaHealth
	if aaaHealth.Status != "healthy" {
		overallHealthy = false
	}

	// Set overall status
	if !overallHealthy {
		response.Status = "unhealthy"
	}

	response.Duration = time.Since(startTime)
	return response, nil
}

// checkDatabaseHealth checks database connectivity and basic operations
func (s *AdministrativeServiceImpl) checkDatabaseHealth(ctx context.Context) responses.ComponentHealth {
	health := responses.ComponentHealth{
		Name:      "PostgreSQL Database",
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Check GORM DB connection
	if s.gormDB != nil {
		sqlDB, err := s.gormDB.DB()
		if err != nil {
			health.Status = "unhealthy"
			health.Error = fmt.Sprintf("Failed to get underlying SQL DB: %v", err)
			return health
		}

		// Ping the database
		if err := sqlDB.PingContext(ctx); err != nil {
			health.Status = "unhealthy"
			health.Error = fmt.Sprintf("Database ping failed: %v", err)
			return health
		}

		// Get database stats
		stats := sqlDB.Stats()
		health.Details["open_connections"] = stats.OpenConnections
		health.Details["in_use"] = stats.InUse
		health.Details["idle"] = stats.Idle
		health.Details["max_open_connections"] = stats.MaxOpenConnections
	}

	// Check PostgresManager if available
	if s.postgresManager != nil {
		db, err := s.postgresManager.GetDB(ctx, false)
		if err != nil {
			health.Status = "unhealthy"
			health.Error = fmt.Sprintf("PostgresManager connection failed: %v", err)
			return health
		}

		// Test a simple query
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			health.Status = "unhealthy"
			health.Error = fmt.Sprintf("Database query test failed: %v", err)
			return health
		}

		health.Details["query_test"] = "passed"
	}

	health.Message = "Database is healthy and responsive"
	return health
}

// checkAAAServiceHealth checks AAA service connectivity
func (s *AdministrativeServiceImpl) checkAAAServiceHealth(ctx context.Context) responses.ComponentHealth {
	health := responses.ComponentHealth{
		Name:      "AAA Service",
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if s.aaaService == nil {
		health.Status = "unhealthy"
		health.Error = "AAA service not initialized"
		health.Message = "AAA service is not available"
		return health
	}

	// Call AAA service health check
	err := s.aaaService.HealthCheck(ctx)
	if err != nil {
		health.Status = "unhealthy"
		health.Error = fmt.Sprintf("AAA service health check failed: %v", err)
		health.Message = "AAA service is not responding"
		return health
	}

	health.Message = "AAA service is healthy and responsive"
	health.Details["connectivity"] = "ok"
	return health
}
