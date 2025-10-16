package services

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AdministrativeServiceIntegrationTestSuite provides integration tests for the administrative service
type AdministrativeServiceIntegrationTestSuite struct {
	suite.Suite
	service AdministrativeService
	db      *gorm.DB
}

// SetupSuite sets up the test suite
func (suite *AdministrativeServiceIntegrationTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	suite.db = db

	// Create mock AAA service
	mockAAA := &MockAAAService{}

	// Setup default mock behaviors
	mockAAA.On("SeedRolesAndPermissions", context.Background()).Return(nil)
	mockAAA.On("HealthCheck", context.Background()).Return(nil)

	// Create service
	concreteService := NewAdministrativeService(nil, db, mockAAA, nil, nil)
	suite.service = NewAdministrativeServiceWrapper(concreteService)
}

// TearDownSuite cleans up after tests
func (suite *AdministrativeServiceIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// TestSeedRolesAndPermissions_Integration tests the seeding functionality
func (suite *AdministrativeServiceIntegrationTestSuite) TestSeedRolesAndPermissions_Integration() {
	suite.Run("successful seeding", func() {
		req := &requests.SeedRolesAndPermissionsRequest{
			Force:  false,
			DryRun: false,
		}

		result, err := suite.service.SeedRolesAndPermissions(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.SeedRolesAndPermissionsResponse)
		assert.True(suite.T(), ok, "Expected *responses.SeedRolesAndPermissionsResponse")
		assert.True(suite.T(), response.Success)
		assert.NotEmpty(suite.T(), response.Message)
		assert.True(suite.T(), response.Duration > 0)
		assert.False(suite.T(), response.Timestamp.IsZero())
	})

	suite.Run("seeding with force flag", func() {
		req := &requests.SeedRolesAndPermissionsRequest{
			Force:  true,
			DryRun: false,
		}

		result, err := suite.service.SeedRolesAndPermissions(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.SeedRolesAndPermissionsResponse)
		assert.True(suite.T(), ok, "Expected *responses.SeedRolesAndPermissionsResponse")
		assert.True(suite.T(), response.Success)
		assert.Contains(suite.T(), response.Message, "successfully")
	})
}

// TestHealthCheck_Integration tests the health check functionality
func (suite *AdministrativeServiceIntegrationTestSuite) TestHealthCheck_Integration() {
	suite.Run("basic health check", func() {
		req := &requests.HealthCheckRequest{}

		result, err := suite.service.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")
		assert.Equal(suite.T(), "healthy", response.Status)
		assert.NotEmpty(suite.T(), response.Components)
		assert.True(suite.T(), response.Duration > 0)
		assert.False(suite.T(), response.Timestamp.IsZero())
	})

	suite.Run("health check with component details", func() {
		req := &requests.HealthCheckRequest{
			Components: []string{"database", "aaa_service"},
		}

		result, err := suite.service.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")

		// Check that both components are present
		assert.Contains(suite.T(), response.Components, "database")
		assert.Contains(suite.T(), response.Components, "aaa_service")

		// Check database component
		dbHealth := response.Components["database"]
		assert.Equal(suite.T(), "PostgreSQL Database", dbHealth.Name)
		assert.NotEmpty(suite.T(), dbHealth.Status)
		assert.False(suite.T(), dbHealth.Timestamp.IsZero())

		// Check AAA service component
		aaaHealth := response.Components["aaa_service"]
		assert.Equal(suite.T(), "AAA Service", aaaHealth.Name)
		assert.Equal(suite.T(), "healthy", aaaHealth.Status)
		assert.False(suite.T(), aaaHealth.Timestamp.IsZero())
	})

	suite.Run("health check performance", func() {
		req := &requests.HealthCheckRequest{}

		start := time.Now()
		result, err := suite.service.HealthCheck(context.Background(), req)
		duration := time.Since(start)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")

		// Health check should complete within reasonable time
		assert.True(suite.T(), duration < time.Second, "Health check took too long: %v", duration)
		assert.True(suite.T(), response.Duration > 0)
		assert.True(suite.T(), response.Duration <= duration)
	})
}

// TestHealthCheck_DatabaseConnectivity tests database connectivity checking
func (suite *AdministrativeServiceIntegrationTestSuite) TestHealthCheck_DatabaseConnectivity() {
	suite.Run("database connectivity check", func() {
		req := &requests.HealthCheckRequest{}

		result, err := suite.service.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")

		// Check database component specifically
		dbHealth, exists := response.Components["database"]
		assert.True(suite.T(), exists, "Database component should be present")

		// For SQLite in-memory DB, it should be healthy
		assert.Equal(suite.T(), "healthy", dbHealth.Status)
		assert.Empty(suite.T(), dbHealth.Error)
		assert.NotEmpty(suite.T(), dbHealth.Message)
	})
}

// TestHealthCheck_ComponentFiltering tests component filtering functionality
func (suite *AdministrativeServiceIntegrationTestSuite) TestHealthCheck_ComponentFiltering() {
	suite.Run("filter specific components", func() {
		req := &requests.HealthCheckRequest{
			Components: []string{"database"},
		}

		result, err := suite.service.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")

		// Should still check all components (filtering is for display purposes)
		assert.Contains(suite.T(), response.Components, "database")
		assert.Contains(suite.T(), response.Components, "aaa_service")
	})
}

// TestSeedRolesAndPermissions_ErrorHandling tests error handling in seeding
func (suite *AdministrativeServiceIntegrationTestSuite) TestSeedRolesAndPermissions_ErrorHandling() {
	suite.Run("handle AAA service errors", func() {
		// Create a service with a failing AAA service
		mockAAA := &MockAAAService{}
		mockAAA.On("SeedRolesAndPermissions", context.Background()).Return(assert.AnError)

		concreteFailingService := NewAdministrativeService(nil, suite.db, mockAAA, nil, nil)
		failingService := NewAdministrativeServiceWrapper(concreteFailingService)

		req := &requests.SeedRolesAndPermissionsRequest{}

		result, err := failingService.SeedRolesAndPermissions(context.Background(), req)

		assert.Error(suite.T(), err)
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		if response, ok := result.(*responses.SeedRolesAndPermissionsResponse); ok {
			assert.False(suite.T(), response.Success)
			assert.NotEmpty(suite.T(), response.Error)
			assert.Contains(suite.T(), response.Message, "Failed")
		}
	})
}

// TestHealthCheck_ErrorScenarios tests various error scenarios
func (suite *AdministrativeServiceIntegrationTestSuite) TestHealthCheck_ErrorScenarios() {
	suite.Run("AAA service failure", func() {
		// Create a service with a failing AAA service
		mockAAA := &MockAAAService{}
		mockAAA.On("HealthCheck", context.Background()).Return(assert.AnError)

		concreteFailingService := NewAdministrativeService(nil, suite.db, mockAAA, nil, nil)
		failingService := NewAdministrativeServiceWrapper(concreteFailingService)

		req := &requests.HealthCheckRequest{}

		result, err := failingService.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err) // Health check itself shouldn't fail
		assert.NotNil(suite.T(), result)

		// Type assert to the expected response type
		response, ok := result.(*responses.HealthCheckResponse)
		assert.True(suite.T(), ok, "Expected *responses.HealthCheckResponse")
		assert.Equal(suite.T(), "unhealthy", response.Status)

		// AAA service should be marked as unhealthy
		aaaHealth := response.Components["aaa_service"]
		assert.Equal(suite.T(), "unhealthy", aaaHealth.Status)
		assert.NotEmpty(suite.T(), aaaHealth.Error)
	})

	suite.Run("nil AAA service", func() {
		// Create a service with nil AAA service
		nilAAAService := NewAdministrativeService(nil, suite.db, nil, nil, nil)

		req := &requests.HealthCheckRequest{}

		result, err := nilAAAService.HealthCheck(context.Background(), req)

		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), result)
		assert.Equal(suite.T(), "unhealthy", result.Status)

		// AAA service should be marked as unhealthy
		aaaHealth := result.Components["aaa_service"]
		assert.Equal(suite.T(), "unhealthy", aaaHealth.Status)
		assert.Contains(suite.T(), aaaHealth.Error, "not initialized")
	})
}

// TestConcurrentOperations tests concurrent access to the service
func (suite *AdministrativeServiceIntegrationTestSuite) TestConcurrentOperations() {
	suite.Run("concurrent health checks", func() {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		req := &requests.HealthCheckRequest{}

		// Launch multiple concurrent health checks
		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := suite.service.HealthCheck(context.Background(), req)
				results <- err
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(suite.T(), err, "Concurrent health check %d failed", i)
		}
	})

	suite.Run("concurrent seeding operations", func() {
		const numGoroutines = 5
		results := make(chan error, numGoroutines)

		req := &requests.SeedRolesAndPermissionsRequest{}

		// Launch multiple concurrent seeding operations
		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := suite.service.SeedRolesAndPermissions(context.Background(), req)
				results <- err
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(suite.T(), err, "Concurrent seeding operation %d failed", i)
		}
	})
}

// Run the integration test suite
func TestAdministrativeServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AdministrativeServiceIntegrationTestSuite))
}
