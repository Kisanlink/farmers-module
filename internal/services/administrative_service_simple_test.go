package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdministrativeService_Basic(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create a simple mock AAA service
	mockAAA := &MockAAAServiceShared{}

	// Create service
	service := NewAdministrativeService(nil, db, mockAAA, nil, nil)
	assert.NotNil(t, service)

	// Test SeedRolesAndPermissions
	t.Run("SeedRolesAndPermissions", func(t *testing.T) {
		// Set up mock expectation for SeedRolesAndPermissions
		mockAAA.On("SeedRolesAndPermissions", context.Background(), false).Return(nil)

		req := &requests.SeedRolesAndPermissionsRequest{
			Force:  false,
			DryRun: false,
		}

		result, err := service.SeedRolesAndPermissions(context.Background(), req)

		// Should not error with proper mock setup
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.False(t, result.Timestamp.IsZero())
		assert.True(t, result.Duration >= 0)
		assert.Equal(t, "Roles and permissions seeded successfully", result.Message)

		// Verify mock was called
		mockAAA.AssertExpectations(t)
	})

	// Test HealthCheck
	t.Run("HealthCheck", func(t *testing.T) {
		// Set up mock expectation for HealthCheck
		mockAAA.On("HealthCheck", context.Background()).Return(nil)

		req := &requests.HealthCheckRequest{}

		result, err := service.HealthCheck(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Status)
		assert.NotEmpty(t, result.Components)
		assert.True(t, result.Duration > 0)
		assert.False(t, result.Timestamp.IsZero())

		// Should have database and aaa_service components
		assert.Contains(t, result.Components, "database")
		assert.Contains(t, result.Components, "aaa_service")

		// Verify mock was called
		mockAAA.AssertExpectations(t)
	})
}

func TestAdministrativeServiceWrapper_Basic(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create a simple mock AAA service
	mockAAA := &MockAAAServiceShared{}

	// Create concrete service and wrapper
	concreteService := NewAdministrativeService(nil, db, mockAAA, nil, nil)
	wrapper := NewAdministrativeServiceWrapper(concreteService)
	assert.NotNil(t, wrapper)

	// Test wrapper functionality
	t.Run("SeedRolesAndPermissions via wrapper", func(t *testing.T) {
		// Set up mock expectation for SeedRolesAndPermissions
		mockAAA.On("SeedRolesAndPermissions", context.Background(), true).Return(nil)

		req := map[string]interface{}{
			"force":   true,
			"dry_run": false,
		}

		result, err := wrapper.SeedRolesAndPermissions(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Should be able to handle map input and convert to proper request

		// Verify mock was called
		mockAAA.AssertExpectations(t)
	})

	t.Run("HealthCheck via wrapper", func(t *testing.T) {
		// Set up mock expectation for HealthCheck
		mockAAA.On("HealthCheck", context.Background()).Return(nil)

		result, err := wrapper.HealthCheck(context.Background(), nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify mock was called
		mockAAA.AssertExpectations(t)
	})
}
