package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	irrigationRepo "github.com/Kisanlink/farmers-module/internal/repo/irrigation_source"
	soilTypeRepo "github.com/Kisanlink/farmers-module/internal/repo/soil_type"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// ConcreteAdministrativeService defines the concrete interface for administrative operations
type ConcreteAdministrativeService interface {
	// SeedRolesAndPermissions triggers a complete reseed of AAA resources, actions, and role bindings
	SeedRolesAndPermissions(ctx context.Context, req *requests.SeedRolesAndPermissionsRequest) (*responses.SeedRolesAndPermissionsResponse, error)

	// SeedLookupData seeds master lookup data (soil types, irrigation sources, etc.)
	SeedLookupData(ctx context.Context, req *requests.SeedLookupsRequest) (*responses.SeedLookupsResponse, error)

	// HealthCheck verifies database connectivity and AAA service availability
	HealthCheck(ctx context.Context, req *requests.HealthCheckRequest) (*responses.HealthCheckResponse, error)
}

// AdministrativeServiceImpl implements ConcreteAdministrativeService
type AdministrativeServiceImpl struct {
	postgresManager      *db.PostgresManager
	gormDB               *gorm.DB
	aaaService           AAAService
	soilTypeRepo         *soilTypeRepo.SoilTypeRepository
	irrigationSourceRepo *irrigationRepo.IrrigationSourceRepository
}

// NewAdministrativeService creates a new administrative service
func NewAdministrativeService(
	postgresManager *db.PostgresManager,
	gormDB *gorm.DB,
	aaaService AAAService,
	soilTypeRepo *soilTypeRepo.SoilTypeRepository,
	irrigationSourceRepo *irrigationRepo.IrrigationSourceRepository,
) ConcreteAdministrativeService {
	return &AdministrativeServiceImpl{
		postgresManager:      postgresManager,
		gormDB:               gormDB,
		aaaService:           aaaService,
		soilTypeRepo:         soilTypeRepo,
		irrigationSourceRepo: irrigationSourceRepo,
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

// SeedLookupData seeds master lookup data including soil types and irrigation sources
func (s *AdministrativeServiceImpl) SeedLookupData(ctx context.Context, req *requests.SeedLookupsRequest) (*responses.SeedLookupsResponse, error) {
	startTime := time.Now()
	response := &responses.SeedLookupsResponse{
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Track seeded counts
	soilTypesSeeded := 0
	irrigationSourcesSeeded := 0

	// Seed soil types
	if req.SeedSoilTypes {
		count, err := s.seedSoilTypes(ctx)
		if err != nil {
			response.Success = false
			response.Message = "Failed to seed lookup data"
			response.Error = fmt.Sprintf("Soil types seeding failed: %v", err)
			response.Duration = time.Since(startTime)
			return response, fmt.Errorf("failed to seed soil types: %w", err)
		}
		soilTypesSeeded = count
		response.Details["soil_types_seeded"] = soilTypesSeeded
	}

	// Seed irrigation sources
	if req.SeedIrrigationSources {
		count, err := s.seedIrrigationSources(ctx)
		if err != nil {
			response.Success = false
			response.Message = "Failed to seed lookup data"
			response.Error = fmt.Sprintf("Irrigation sources seeding failed: %v", err)
			response.Duration = time.Since(startTime)
			return response, fmt.Errorf("failed to seed irrigation sources: %w", err)
		}
		irrigationSourcesSeeded = count
		response.Details["irrigation_sources_seeded"] = irrigationSourcesSeeded
	}

	response.Success = true
	response.Message = fmt.Sprintf("Successfully seeded lookup data: %d soil types, %d irrigation sources", soilTypesSeeded, irrigationSourcesSeeded)
	response.Duration = time.Since(startTime)

	return response, nil
}

// seedSoilTypes seeds predefined soil types using repository
func (s *AdministrativeServiceImpl) seedSoilTypes(ctx context.Context) (int, error) {
	// Use predefined soil types from the entity
	predefinedSoilTypes := []*soil_type.SoilType{
		{Name: "BLACK", Description: "Black soil - rich in clay content, good for cotton cultivation"},
		{Name: "RED", Description: "Red soil - well-drained, suitable for various crops"},
		{Name: "SANDY", Description: "Sandy soil - well-drained but low water retention"},
		{Name: "LOAMY", Description: "Loamy soil - ideal mixture of sand, silt, and clay"},
		{Name: "ALLUVIAL", Description: "Alluvial soil - fertile soil deposited by rivers"},
		{Name: "MIXED", Description: "Mixed soil types - combination of different soil types"},
	}

	// Initialize base models for each soil type
	for _, st := range predefinedSoilTypes {
		// Initialize using base.NewBaseModel
		st.BaseModel = *base.NewBaseModel("SOIL", st.GetTableSize())
		st.Properties = make(map[string]interface{})
	}

	// Use repository to create batch (it will skip duplicates)
	if err := s.soilTypeRepo.CreateBatch(ctx, predefinedSoilTypes); err != nil {
		return 0, fmt.Errorf("failed to seed soil types: %w", err)
	}

	// Count total soil types in database
	var count int64
	if s.gormDB != nil {
		s.gormDB.WithContext(ctx).Model(&soil_type.SoilType{}).Count(&count)
	}

	return int(count), nil
}

// seedIrrigationSources seeds predefined irrigation sources using repository
func (s *AdministrativeServiceImpl) seedIrrigationSources(ctx context.Context) (int, error) {
	// Use predefined irrigation sources from the entity
	predefinedSources := []*irrigation_source.IrrigationSource{
		{Name: "BOREWELL", Description: "Borewell irrigation system", RequiresCount: true},
		{Name: "FLOOD_IRRIGATION", Description: "Flood irrigation method", RequiresCount: false},
		{Name: "DRIP_IRRIGATION", Description: "Drip irrigation system", RequiresCount: false},
		{Name: "CANAL", Description: "Canal irrigation", RequiresCount: false},
		{Name: "RAINFED", Description: "Rain-fed agriculture", RequiresCount: false},
		{Name: "OTHER", Description: "Other irrigation sources", RequiresCount: false},
	}

	// Initialize base models for each irrigation source
	for _, is := range predefinedSources {
		// Initialize using base.NewBaseModel
		is.BaseModel = *base.NewBaseModel("IRRG", is.GetTableSize())
		is.Properties = make(map[string]interface{})
	}

	// Use repository to create batch (it will skip duplicates)
	if err := s.irrigationSourceRepo.CreateBatch(ctx, predefinedSources); err != nil {
		return 0, fmt.Errorf("failed to seed irrigation sources: %w", err)
	}

	// Count total irrigation sources in database
	var count int64
	if s.gormDB != nil {
		s.gormDB.WithContext(ctx).Model(&irrigation_source.IrrigationSource{}).Count(&count)
	}

	return int(count), nil
}
