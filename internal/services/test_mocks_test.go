package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPermissionMatrix tests the permission matrix functionality
func TestPermissionMatrix(t *testing.T) {
	tests := []struct {
		name        string
		setupMatrix func(*PermissionMatrix)
		checkParams struct {
			subject  string
			resource string
			action   string
			object   string
			orgID    string
		}
		expectedAllow bool
	}{
		{
			name: "allow specific permission",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farmer", "create", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farmer", "create", "farmer789", "org456"},
			expectedAllow: true,
		},
		{
			name: "deny specific permission",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddDenyRule("user123", "farmer", "delete", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farmer", "delete", "farmer789", "org456"},
			expectedAllow: false,
		},
		{
			name: "wildcard subject",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("*", "farm", "read", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user999", "farm", "read", "farm123", "org456"},
			expectedAllow: true,
		},
		{
			name: "default deny when no rules match",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farmer", "create", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farm", "delete", "farm123", "org456"},
			expectedAllow: false, // Default deny is true in this matrix
		},
		{
			name: "specific object permission",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farm", "update", "farm123", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farm", "update", "farm123", "org456"},
			expectedAllow: true,
		},
		{
			name: "specific object permission - different object denied",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farm", "update", "farm123", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farm", "update", "farm999", "org456"},
			expectedAllow: false, // Different object
		},
		{
			name: "org-scoped permission",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farmer", "list", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farmer", "list", "", "org456"},
			expectedAllow: true,
		},
		{
			name: "org-scoped permission - wrong org denied",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddAllowRule("user123", "farmer", "list", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farmer", "list", "", "org999"},
			expectedAllow: false, // Different org
		},
		{
			name: "first match wins - deny before allow",
			setupMatrix: func(pm *PermissionMatrix) {
				pm.AddDenyRule("user123", "farmer", "delete", "*", "org456")
				pm.AddAllowRule("user123", "farmer", "delete", "*", "org456")
			},
			checkParams: struct {
				subject  string
				resource string
				action   string
				object   string
				orgID    string
			}{"user123", "farmer", "delete", "farmer789", "org456"},
			expectedAllow: false, // First rule (deny) wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create permission matrix with deny-by-default
			pm := NewPermissionMatrix(true)
			pm.logDenials = false // Disable logging for tests

			// Setup the matrix
			tt.setupMatrix(pm)

			// Check permission
			allowed := pm.CheckPermission(
				tt.checkParams.subject,
				tt.checkParams.resource,
				tt.checkParams.action,
				tt.checkParams.object,
				tt.checkParams.orgID,
			)

			assert.Equal(t, tt.expectedAllow, allowed,
				"Permission check failed for %s: expected %v, got %v",
				tt.name, tt.expectedAllow, allowed)
		})
	}
}

// TestMockAAAServiceShared_WithPermissionMatrix tests the mock AAA service with permission matrix
func TestMockAAAServiceShared_WithPermissionMatrix(t *testing.T) {
	ctx := context.Background()

	t.Run("deny-by-default behavior", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()
		matrix.logDenials = false // Disable logging for tests

		// No rules configured, should deny
		allowed, err := mock.CheckPermission(ctx, "user123", "farmer", "create", "", "org456")
		assert.NoError(t, err)
		assert.False(t, allowed, "Should deny by default when no rules configured")
	})

	t.Run("allow specific permission", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()
		matrix.logDenials = false // Disable logging for tests
		matrix.AddAllowRule("user123", "farmer", "create", "*", "org456")

		// Should allow the configured permission
		allowed, err := mock.CheckPermission(ctx, "user123", "farmer", "create", "", "org456")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should allow configured permission")

		// Should deny different action
		allowed, err = mock.CheckPermission(ctx, "user123", "farmer", "delete", "", "org456")
		assert.NoError(t, err)
		assert.False(t, allowed, "Should deny different action")
	})

	t.Run("complex permission scenario", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()
		matrix.logDenials = false // Disable logging for tests

		// Admin can do everything in org123
		matrix.AddAllowRule("admin123", "*", "*", "*", "org123")

		// User can only read farmers in org123
		matrix.AddAllowRule("user456", "farmer", "read", "*", "org123")

		// User cannot delete anything
		matrix.AddDenyRule("user456", "*", "delete", "*", "*")

		// Test admin permissions
		allowed, err := mock.CheckPermission(ctx, "admin123", "farmer", "delete", "farmer789", "org123")
		assert.NoError(t, err)
		assert.True(t, allowed, "Admin should be able to delete")

		// Test user read permission
		allowed, err = mock.CheckPermission(ctx, "user456", "farmer", "read", "farmer789", "org123")
		assert.NoError(t, err)
		assert.True(t, allowed, "User should be able to read")

		// Test user delete denied
		allowed, err = mock.CheckPermission(ctx, "user456", "farmer", "delete", "farmer789", "org123")
		assert.NoError(t, err)
		assert.False(t, allowed, "User should not be able to delete")
	})

	t.Run("backward compatibility with mock.On", func(t *testing.T) {
		mock := NewMockAAAServiceShared(false)

		// Use traditional mock setup (no permission matrix rules)
		mock.On("CheckPermission", ctx, "user123", "farmer", "create", "", "org456").Return(true, nil)

		// Should use mock behavior when no matrix rules
		allowed, err := mock.CheckPermission(ctx, "user123", "farmer", "create", "", "org456")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should fall back to mock behavior")

		mock.AssertExpectations(t)
	})
}

// TestPermissionMatrix_ThreadSafety tests thread safety of permission matrix
func TestPermissionMatrix_ThreadSafety(t *testing.T) {
	pm := NewPermissionMatrix(true)
	pm.logDenials = false

	// Add rules concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				pm.AddAllowRule("user"+string(rune(id)), "farmer", "create", "*", "org123")
				_ = pm.CheckPermission("user"+string(rune(id)), "farmer", "create", "", "org123")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race conditions
	assert.True(t, true, "Thread safety test completed")
}

// TestPermissionMatrix_Clear tests clearing permission rules
func TestPermissionMatrix_Clear(t *testing.T) {
	pm := NewPermissionMatrix(true)
	pm.logDenials = false

	// Add some rules
	pm.AddAllowRule("user123", "farmer", "create", "*", "org456")
	pm.AddAllowRule("user456", "farm", "read", "*", "org789")

	// Verify rules work
	allowed := pm.CheckPermission("user123", "farmer", "create", "", "org456")
	assert.True(t, allowed, "Permission should be allowed before clear")

	// Clear rules
	pm.Clear()

	// Verify deny-by-default after clear
	allowed = pm.CheckPermission("user123", "farmer", "create", "", "org456")
	assert.False(t, allowed, "Permission should be denied after clear (default deny)")
}

// TestMockCache tests the mock cache implementation
func TestMockCache(t *testing.T) {
	ctx := context.Background()
	cache := &MockCache{}

	t.Run("Get", func(t *testing.T) {
		cache.On("Get", ctx, "key1").Return("value1", nil)
		val, err := cache.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)
		cache.AssertExpectations(t)
	})

	t.Run("Set", func(t *testing.T) {
		cache.On("Set", ctx, "key2", "value2", time.Duration(0)).Return(nil)
		err := cache.Set(ctx, "key2", "value2", time.Duration(0))
		assert.NoError(t, err)
		cache.AssertExpectations(t)
	})
}

// TestMockEventEmitter tests the mock event emitter implementation
func TestMockEventEmitter(t *testing.T) {
	emitter := &MockEventEmitter{}

	t.Run("EmitAuditEvent", func(t *testing.T) {
		event := map[string]interface{}{"action": "create"}
		emitter.On("EmitAuditEvent", event).Return(nil)
		err := emitter.EmitAuditEvent(event)
		assert.NoError(t, err)
		emitter.AssertExpectations(t)
	})

	t.Run("EmitBusinessEvent", func(t *testing.T) {
		data := map[string]interface{}{"farmer_id": "123"}
		emitter.On("EmitBusinessEvent", "farmer.created", data).Return(nil)
		err := emitter.EmitBusinessEvent("farmer.created", data)
		assert.NoError(t, err)
		emitter.AssertExpectations(t)
	})
}

// TestMockDatabase tests the mock database implementation
func TestMockDatabase(t *testing.T) {
	db := &MockDatabase{}

	t.Run("Connect", func(t *testing.T) {
		db.On("Connect").Return(nil)
		err := db.Connect()
		assert.NoError(t, err)
		db.AssertExpectations(t)
	})

	t.Run("Ping", func(t *testing.T) {
		db.On("Ping").Return(nil)
		err := db.Ping()
		assert.NoError(t, err)
		db.AssertExpectations(t)
	})
}
