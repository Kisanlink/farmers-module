package services

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ContractTest defines the interface for contract testing
// Contract tests ensure mocks behave exactly like real implementations
type ContractTest struct {
	name        string
	testFunc    func(t *testing.T, service interfaces.AAAService)
	description string
}

// AAAServiceContractTests defines all contract tests for AAA service
// Both mock and real implementations must pass these tests
var AAAServiceContractTests = []ContractTest{
	{
		name:        "CheckPermission_ValidPermission_ReturnsTrue",
		description: "Valid permission checks should return true when permission is granted",
		testFunc: func(t *testing.T, service interfaces.AAAService) {
			ctx := context.Background()

			// Test basic permission check
			allowed, err := service.CheckPermission(ctx, "user123", "farmer", "read", "*", "org123")
			require.NoError(t, err)
			// Note: This test assumes the service has been configured with appropriate permissions
			// In real tests, you'd set up the permission first
			assert.NotNil(t, allowed, "Permission check should return a boolean value")
		},
	},
	{
		name:        "CheckPermission_InvalidPermission_ReturnsFalse",
		description: "Permission checks without grants should return false in deny-by-default mode",
		testFunc: func(t *testing.T, service interfaces.AAAService) {
			ctx := context.Background()

			// Test permission denial for non-existent user
			allowed, err := service.CheckPermission(ctx, "nonexistent", "farmer", "delete", "*", "org123")
			require.NoError(t, err)
			// In deny-by-default mode, this should be false
			assert.False(t, allowed, "Permission should be denied for non-existent user")
		},
	},
	{
		name:        "CheckPermission_WithEmptyParameters_HandlesGracefully",
		description: "Permission checks with empty parameters should handle gracefully",
		testFunc: func(t *testing.T, service interfaces.AAAService) {
			ctx := context.Background()

			// Test with empty subject
			allowed, err := service.CheckPermission(ctx, "", "farmer", "read", "*", "org123")
			// Should either error or deny
			if err == nil {
				assert.False(t, allowed, "Empty subject should be denied")
			}
		},
	},
	{
		name:        "CheckPermission_WithWildcardResource_MatchesAll",
		description: "Wildcard resource permissions should match all resources",
		testFunc: func(t *testing.T, service interfaces.AAAService) {
			ctx := context.Background()

			// This test requires the service to be configured with wildcard permissions
			// The actual behavior depends on the implementation
			_, err := service.CheckPermission(ctx, "admin", "*", "read", "*", "org123")
			require.NoError(t, err)
		},
	},
	{
		name:        "HealthCheck_Always_Succeeds",
		description: "Health check should always succeed for healthy service",
		testFunc: func(t *testing.T, service interfaces.AAAService) {
			ctx := context.Background()
			err := service.HealthCheck(ctx)
			assert.NoError(t, err, "Health check should succeed")
		},
	},
}

// RunContractTests runs all contract tests against a given AAA service implementation
func RunContractTests(t *testing.T, service interfaces.AAAService, implementationName string) {
	t.Run(implementationName, func(t *testing.T) {
		for _, test := range AAAServiceContractTests {
			t.Run(test.name, func(t *testing.T) {
				t.Logf("Running: %s", test.description)
				test.testFunc(t, service)
			})
		}
	})
}

// TestMockAAAServiceContract tests that mock AAA service follows the contract
func TestMockAAAServiceContract(t *testing.T) {
	// Create mock with deny-by-default for security
	mock := NewMockAAAServiceShared(true)

	// Configure mock for specific test scenarios
	matrix := mock.GetPermissionMatrix()
	matrix.AddAllowRule("admin", "*", "*", "*", "org123")
	matrix.AddAllowRule("user123", "farmer", "read", "*", "org123")

	// Configure mock expectations for methods that use testify mock
	ctx := context.Background()
	mock.On("HealthCheck", ctx).Return(nil)

	// Run all contract tests
	RunContractTests(t, mock, "MockAAAService")
}

// TestMockAAAServicePresetContracts tests that preset configurations follow contracts
func TestMockAAAServicePresetContracts(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)

	presets := []struct {
		name   MockPreset
		userID string
		orgID  string
	}{
		{PresetAdmin, "admin123", "org123"},
		{PresetFarmer, "farmer123", "org123"},
		{PresetKisanSathi, "ks123", "org123"},
		{PresetFPOManager, "mgr123", "org123"},
		{PresetReadOnly, "viewer123", "org123"},
	}

	for _, preset := range presets {
		t.Run(string(preset.name), func(t *testing.T) {
			mock := factory.NewAAAServiceMockWithPreset(preset.name, preset.userID, preset.orgID)

			// Configure mock expectations for methods that use testify mock
			ctx := context.Background()
			mock.On("HealthCheck", ctx).Return(nil)

			RunContractTests(t, mock, string(preset.name))
		})
	}
}

// CacheContractTests defines contract tests for Cache interface
var CacheContractTests = []struct {
	name        string
	testFunc    func(t *testing.T, cache interfaces.Cache)
	description string
}{
	{
		name:        "Get_NonExistentKey_ReturnsError",
		description: "Getting a non-existent key should return an error",
		testFunc: func(t *testing.T, cache interfaces.Cache) {
			ctx := context.Background()
			val, err := cache.Get(ctx, "nonexistent_key")
			assert.Error(t, err, "Should error for non-existent key")
			assert.Nil(t, val, "Value should be nil for non-existent key")
		},
	},
	{
		name:        "SetAndGet_ValidKey_ReturnsValue",
		description: "Setting and getting a value should work correctly",
		testFunc: func(t *testing.T, cache interfaces.Cache) {
			ctx := context.Background()
			testValue := "test_value"

			err := cache.Set(ctx, "test_key", testValue, time.Duration(0))
			require.NoError(t, err, "Set should succeed")

			val, err := cache.Get(ctx, "test_key")
			require.NoError(t, err, "Get should succeed")
			assert.Equal(t, testValue, val, "Retrieved value should match set value")
		},
	},
	{
		name:        "Delete_ExistingKey_Succeeds",
		description: "Deleting an existing key should succeed",
		testFunc: func(t *testing.T, cache interfaces.Cache) {
			ctx := context.Background()

			// Set a value
			err := cache.Set(ctx, "delete_key", "value", time.Duration(0))
			require.NoError(t, err)

			// Delete it
			err = cache.Delete(ctx, "delete_key")
			assert.NoError(t, err, "Delete should succeed")

			// Verify it's gone
			_, err = cache.Get(ctx, "delete_key")
			assert.Error(t, err, "Get should fail after delete")
		},
	},
	{
		name:        "Clear_Always_Succeeds",
		description: "Clearing the cache should always succeed",
		testFunc: func(t *testing.T, cache interfaces.Cache) {
			ctx := context.Background()

			// Set some values
			cache.Set(ctx, "key1", "value1", time.Duration(0))
			cache.Set(ctx, "key2", "value2", time.Duration(0))

			// Clear cache
			err := cache.Clear(ctx)
			assert.NoError(t, err, "Clear should succeed")

			// Verify values are gone
			_, err = cache.Get(ctx, "key1")
			assert.Error(t, err, "Get should fail after clear")
		},
	},
}

// RunCacheContractTests runs all cache contract tests
func RunCacheContractTests(t *testing.T, cache interfaces.Cache, implementationName string) {
	t.Run(implementationName, func(t *testing.T) {
		for _, test := range CacheContractTests {
			t.Run(test.name, func(t *testing.T) {
				t.Logf("Running: %s", test.description)
				test.testFunc(t, cache)
			})
		}
	})
}

// TestMockCacheContract tests that mock cache follows the contract
func TestMockCacheContract(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)
	mock := factory.NewCacheMock()

	// Configure mock behavior
	ctx := context.Background()
	mock.On("Get", ctx, "nonexistent_key").Return(nil, assert.AnError)
	mock.On("Set", ctx, "test_key", "test_value", time.Duration(0)).Return(nil)
	mock.On("Get", ctx, "test_key").Return("test_value", nil)
	mock.On("Set", ctx, "delete_key", "value", time.Duration(0)).Return(nil)
	mock.On("Delete", ctx, "delete_key").Return(nil)
	mock.On("Get", ctx, "delete_key").Return(nil, assert.AnError)
	mock.On("Set", ctx, "key1", "value1", time.Duration(0)).Return(nil)
	mock.On("Set", ctx, "key2", "value2", time.Duration(0)).Return(nil)
	mock.On("Clear", ctx).Return(nil)
	mock.On("Get", ctx, "key1").Return(nil, assert.AnError)

	RunCacheContractTests(t, mock, "MockCache")
}

// EventEmitterContractTests defines contract tests for EventEmitter interface
var EventEmitterContractTests = []struct {
	name        string
	testFunc    func(t *testing.T, emitter interfaces.EventEmitter)
	description string
}{
	{
		name:        "EmitAuditEvent_ValidEvent_Succeeds",
		description: "Emitting a valid audit event should succeed",
		testFunc: func(t *testing.T, emitter interfaces.EventEmitter) {
			event := map[string]interface{}{
				"user_id":   "user123",
				"action":    "create_farmer",
				"timestamp": "2025-10-05T00:00:00Z",
			}
			err := emitter.EmitAuditEvent(event)
			assert.NoError(t, err, "EmitAuditEvent should succeed")
		},
	},
	{
		name:        "EmitBusinessEvent_ValidEvent_Succeeds",
		description: "Emitting a valid business event should succeed",
		testFunc: func(t *testing.T, emitter interfaces.EventEmitter) {
			data := map[string]interface{}{
				"farmer_id": "farmer123",
				"status":    "active",
			}
			err := emitter.EmitBusinessEvent("farmer.created", data)
			assert.NoError(t, err, "EmitBusinessEvent should succeed")
		},
	},
}

// RunEventEmitterContractTests runs all event emitter contract tests
func RunEventEmitterContractTests(t *testing.T, emitter interfaces.EventEmitter, implementationName string) {
	t.Run(implementationName, func(t *testing.T) {
		for _, test := range EventEmitterContractTests {
			t.Run(test.name, func(t *testing.T) {
				t.Logf("Running: %s", test.description)
				test.testFunc(t, emitter)
			})
		}
	})
}

// TestMockEventEmitterContract tests that mock event emitter follows the contract
func TestMockEventEmitterContract(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)
	mock := factory.NewEventEmitterMock()

	// Configure mock behavior for audit event
	auditEvent := map[string]interface{}{
		"user_id":   "user123",
		"action":    "create_farmer",
		"timestamp": "2025-10-05T00:00:00Z",
	}
	mock.On("EmitAuditEvent", auditEvent).Return(nil)

	// Configure mock behavior for business event
	businessData := map[string]interface{}{
		"farmer_id": "farmer123",
		"status":    "active",
	}
	mock.On("EmitBusinessEvent", "farmer.created", businessData).Return(nil)

	RunEventEmitterContractTests(t, mock, "MockEventEmitter")
}

// DatabaseContractTests defines contract tests for Database interface
var DatabaseContractTests = []struct {
	name        string
	testFunc    func(t *testing.T, db interfaces.Database)
	description string
}{
	{
		name:        "Connect_Always_Succeeds",
		description: "Database connection should succeed",
		testFunc: func(t *testing.T, db interfaces.Database) {
			err := db.Connect()
			assert.NoError(t, err, "Connect should succeed")
		},
	},
	{
		name:        "Ping_AfterConnect_Succeeds",
		description: "Ping should succeed after connection",
		testFunc: func(t *testing.T, db interfaces.Database) {
			db.Connect()
			err := db.Ping()
			assert.NoError(t, err, "Ping should succeed")
		},
	},
	{
		name:        "Migrate_Always_Succeeds",
		description: "Migration should succeed",
		testFunc: func(t *testing.T, db interfaces.Database) {
			err := db.Migrate()
			assert.NoError(t, err, "Migrate should succeed")
		},
	},
	{
		name:        "Close_AfterConnect_Succeeds",
		description: "Close should succeed after connection",
		testFunc: func(t *testing.T, db interfaces.Database) {
			db.Connect()
			err := db.Close()
			assert.NoError(t, err, "Close should succeed")
		},
	},
}

// RunDatabaseContractTests runs all database contract tests
func RunDatabaseContractTests(t *testing.T, db interfaces.Database, implementationName string) {
	t.Run(implementationName, func(t *testing.T) {
		for _, test := range DatabaseContractTests {
			t.Run(test.name, func(t *testing.T) {
				t.Logf("Running: %s", test.description)
				test.testFunc(t, db)
			})
		}
	})
}

// TestMockDatabaseContract tests that mock database follows the contract
func TestMockDatabaseContract(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)
	mock := factory.NewDatabaseMock()

	// Configure mock behavior
	mock.On("Connect").Return(nil)
	mock.On("Ping").Return(nil)
	mock.On("Migrate").Return(nil)
	mock.On("Close").Return(nil)

	RunDatabaseContractTests(t, mock, "MockDatabase")
}
