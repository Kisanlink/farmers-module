package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestMockFactory_NewAAAServiceMock(t *testing.T) {
	t.Run("deny-all mode creates secure mock", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeDenyAll)
		mock := factory.NewAAAServiceMock()

		assert.NotNil(t, mock)
		assert.NotNil(t, mock.GetPermissionMatrix())
		assert.True(t, mock.GetPermissionMatrix().defaultDeny)
	})

	t.Run("allow-all mode creates permissive mock", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeAllowAll)
		mock := factory.NewAAAServiceMock()

		assert.NotNil(t, mock)
		assert.NotNil(t, mock.GetPermissionMatrix())
		assert.False(t, mock.GetPermissionMatrix().defaultDeny)
	})

	t.Run("custom mode creates deny-by-default mock", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeCustom)
		mock := factory.NewAAAServiceMock()

		assert.NotNil(t, mock)
		assert.True(t, mock.GetPermissionMatrix().defaultDeny)
	})
}

func TestMockFactory_NewAAAServiceMockWithPreset(t *testing.T) {
	ctx := context.Background()
	factory := NewMockFactory(PermissionModeDenyAll)

	tests := []struct {
		name            string
		preset          MockPreset
		userID          string
		orgID           string
		testPermissions []permissionTest
		expectedAllowed int // Number of permissions that should be allowed
	}{
		{
			name:   "admin preset allows everything",
			preset: PresetAdmin,
			userID: "admin123",
			orgID:  "org456",
			testPermissions: []permissionTest{
				{"admin123", "farmer", "create", "", "org456", true},
				{"admin123", "farmer", "delete", "", "org456", true},
				{"admin123", "farm", "update", "farm123", "org456", true},
				{"admin123", "cycle", "end", "cycle789", "org456", true},
			},
			expectedAllowed: 4,
		},
		{
			name:   "farmer preset allows limited operations",
			preset: PresetFarmer,
			userID: "farmer123",
			orgID:  "org456",
			testPermissions: []permissionTest{
				{"farmer123", "farmer", "read", "", "org456", true},
				{"farmer123", "farmer", "update", "farmer123", "org456", true},
				{"farmer123", "farm", "create", "", "org456", true},
				{"farmer123", "cycle", "start", "", "org456", true},
				{"farmer123", "farmer", "delete", "", "org456", false}, // Should be denied
				{"farmer123", "fpo", "update", "", "org456", false},    // Should be denied
			},
			expectedAllowed: 4,
		},
		{
			name:   "kisansathi preset allows farmer management",
			preset: PresetKisanSathi,
			userID: "ks123",
			orgID:  "org456",
			testPermissions: []permissionTest{
				{"ks123", "farmer", "create", "", "org456", true},
				{"ks123", "farmer", "delete", "", "org456", true},
				{"ks123", "farm", "update", "farm123", "org456", true},
				{"ks123", "cycle", "end", "cycle789", "org456", true},
				{"ks123", "fpo", "delete", "", "org456", false}, // Should be denied
			},
			expectedAllowed: 4,
		},
		{
			name:   "fpo_manager preset allows org management",
			preset: PresetFPOManager,
			userID: "mgr123",
			orgID:  "org456",
			testPermissions: []permissionTest{
				{"mgr123", "fpo", "update", "", "org456", true},
				{"mgr123", "farmer", "create", "", "org456", true},
				{"mgr123", "farm", "delete", "farm123", "org456", true},
				{"mgr123", "user", "create", "", "org456", true},
			},
			expectedAllowed: 4,
		},
		{
			name:   "readonly preset denies write operations",
			preset: PresetReadOnly,
			userID: "viewer123",
			orgID:  "org456",
			testPermissions: []permissionTest{
				{"viewer123", "farmer", "read", "", "org456", true},
				{"viewer123", "farm", "list", "", "org456", true},
				{"viewer123", "farmer", "create", "", "org456", false},
				{"viewer123", "farm", "update", "", "org456", false},
				{"viewer123", "farmer", "delete", "", "org456", false},
			},
			expectedAllowed: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := factory.NewAAAServiceMockWithPreset(tt.preset, tt.userID, tt.orgID)
			mock.GetPermissionMatrix().logDenials = false // Disable logging for tests

			allowed := 0
			for _, perm := range tt.testPermissions {
				result, err := mock.CheckPermission(ctx,
					perm.subject, perm.resource, perm.action, perm.object, perm.orgID)

				assert.NoError(t, err)
				if result {
					allowed++
				}
				assert.Equal(t, perm.expected, result,
					"Permission check failed for %s %s %s %s %s",
					perm.subject, perm.resource, perm.action, perm.object, perm.orgID)
			}

			assert.Equal(t, tt.expectedAllowed, allowed,
				"Expected %d permissions to be allowed, got %d",
				tt.expectedAllowed, allowed)
		})
	}
}

type permissionTest struct {
	subject  string
	resource string
	action   string
	object   string
	orgID    string
	expected bool
}

func TestMockFactory_ValidateMockInterface(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)

	t.Run("AAA service mock implements interface", func(t *testing.T) {
		mock := factory.NewAAAServiceMock()
		var expectedInterface *interfaces.AAAService
		err := factory.ValidateMockInterface(mock, expectedInterface)
		assert.NoError(t, err)
	})

	t.Run("cache mock implements interface", func(t *testing.T) {
		mock := factory.NewCacheMock()
		var expectedInterface *interfaces.Cache
		err := factory.ValidateMockInterface(mock, expectedInterface)
		assert.NoError(t, err)
	})

	t.Run("event emitter mock implements interface", func(t *testing.T) {
		mock := factory.NewEventEmitterMock()
		var expectedInterface *interfaces.EventEmitter
		err := factory.ValidateMockInterface(mock, expectedInterface)
		assert.NoError(t, err)
	})

	t.Run("database mock implements interface", func(t *testing.T) {
		mock := factory.NewDatabaseMock()
		var expectedInterface *interfaces.Database
		err := factory.ValidateMockInterface(mock, expectedInterface)
		assert.NoError(t, err)
	})
}

func TestMockFactory_AllMockCreators(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)

	t.Run("creates AAA service mock", func(t *testing.T) {
		mock := factory.NewAAAServiceMock()
		assert.NotNil(t, mock)
	})

	t.Run("creates cache mock", func(t *testing.T) {
		mock := factory.NewCacheMock()
		assert.NotNil(t, mock)
	})

	t.Run("creates event emitter mock", func(t *testing.T) {
		mock := factory.NewEventEmitterMock()
		assert.NotNil(t, mock)
	})

	t.Run("creates database mock", func(t *testing.T) {
		mock := factory.NewDatabaseMock()
		assert.NotNil(t, mock)
	})

	t.Run("creates farmer linkage repo mock", func(t *testing.T) {
		mock := factory.NewFarmerLinkageRepoMock()
		assert.NotNil(t, mock)
	})

	t.Run("creates data quality service mock", func(t *testing.T) {
		mock := factory.NewDataQualityServiceMock()
		assert.NotNil(t, mock)
	})
}

func TestDefaultMockFactory(t *testing.T) {
	t.Run("default factory is secure by default", func(t *testing.T) {
		assert.Equal(t, PermissionModeDenyAll, DefaultMockFactory.DefaultPermissionMode)
		mock := DefaultMockFactory.NewAAAServiceMock()
		assert.True(t, mock.GetPermissionMatrix().defaultDeny)
	})
}

func TestTestMockFactory(t *testing.T) {
	t.Run("test factory is permissive", func(t *testing.T) {
		assert.Equal(t, PermissionModeAllowAll, TestMockFactory.DefaultPermissionMode)
		mock := TestMockFactory.NewAAAServiceMock()
		assert.False(t, mock.GetPermissionMatrix().defaultDeny)
	})
}

// TestMockFactory_SecurityBestPractices demonstrates security best practices
func TestMockFactory_SecurityBestPractices(t *testing.T) {
	t.Run("use deny-all for security tests", func(t *testing.T) {
		// For security-sensitive tests, always use deny-all
		factory := NewMockFactory(PermissionModeDenyAll)
		mock := factory.NewAAAServiceMock()

		// Without explicit permissions, everything is denied
		ctx := context.Background()
		allowed, err := mock.CheckPermission(ctx, "user123", "farmer", "delete", "", "org456")
		assert.NoError(t, err)
		assert.False(t, allowed, "Should deny by default")

		// Add explicit permission
		mock.GetPermissionMatrix().AddAllowRule("user123", "farmer", "delete", "*", "org456")
		allowed, err = mock.CheckPermission(ctx, "user123", "farmer", "delete", "", "org456")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should allow with explicit rule")
	})

	t.Run("use presets for common scenarios", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeDenyAll)

		// Use presets instead of configuring permissions manually
		farmerMock := factory.NewAAAServiceMockWithPreset(PresetFarmer, "farmer123", "org456")
		adminMock := factory.NewAAAServiceMockWithPreset(PresetAdmin, "admin123", "org456")

		ctx := context.Background()

		// Farmer cannot delete
		allowed, _ := farmerMock.CheckPermission(ctx, "farmer123", "farmer", "delete", "", "org456")
		assert.False(t, allowed)

		// Admin can delete
		allowed, _ = adminMock.CheckPermission(ctx, "admin123", "farmer", "delete", "", "org456")
		assert.True(t, allowed)
	})
}
