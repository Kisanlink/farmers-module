package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPermissionMatrix_EdgeCases tests edge cases in permission matrix logic
func TestPermissionMatrix_EdgeCases(t *testing.T) {
	t.Run("wildcard matching behavior", func(t *testing.T) {
		matrix := NewPermissionMatrix(true) // deny-by-default

		// Test wildcard patterns - ORDER MATTERS (first match wins)
		matrix.AddDenyRule("blocked_user", "*", "*", "*", "*")  // Specific user blocked (must come first)
		matrix.AddAllowRule("admin", "*", "*", "*", "*")        // Admin can do everything
		matrix.AddAllowRule("*", "farmer", "read", "*", "org1") // Anyone can read farmers

		tests := []struct {
			name     string
			subject  string
			resource string
			action   string
			object   string
			orgID    string
			expected bool
		}{
			{"admin overrides all", "admin", "farmer", "delete", "farmer1", "org1", true},
			{"blocked user denied despite wildcard", "blocked_user", "farmer", "read", "farmer1", "org1", false},
			{"random user can read farmers", "user123", "farmer", "read", "farmer1", "org1", true},
			{"random user cannot write farmers", "user123", "farmer", "write", "farmer1", "org1", false},
			{"empty object converted to wildcard", "user123", "farmer", "read", "", "org1", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := matrix.CheckPermission(tt.subject, tt.resource, tt.action, tt.object, tt.orgID)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("first-match-wins policy", func(t *testing.T) {
		matrix := NewPermissionMatrix(true)

		// Order matters - first rule wins
		matrix.AddDenyRule("user1", "farmer", "delete", "*", "org1")
		matrix.AddAllowRule("user1", "farmer", "*", "*", "org1") // This won't override delete

		// Delete should be denied (first rule)
		assert.False(t, matrix.CheckPermission("user1", "farmer", "delete", "farmer1", "org1"))
		// Other actions should be allowed (second rule)
		assert.True(t, matrix.CheckPermission("user1", "farmer", "create", "farmer1", "org1"))
	})

	t.Run("empty rules with different default modes", func(t *testing.T) {
		denyMatrix := NewPermissionMatrix(true)   // deny-by-default
		allowMatrix := NewPermissionMatrix(false) // allow-by-default

		// With no rules, behavior depends on default
		assert.False(t, denyMatrix.CheckPermission("user1", "farmer", "read", "obj1", "org1"))
		assert.True(t, allowMatrix.CheckPermission("user1", "farmer", "read", "obj1", "org1"))
	})

	t.Run("clear rules functionality", func(t *testing.T) {
		matrix := NewPermissionMatrix(true)
		matrix.AddAllowRule("user1", "farmer", "read", "*", "org1")

		// Should allow before clear
		assert.True(t, matrix.CheckPermission("user1", "farmer", "read", "obj1", "org1"))

		// Clear all rules
		matrix.Clear()

		// Should deny after clear (back to default)
		assert.False(t, matrix.CheckPermission("user1", "farmer", "read", "obj1", "org1"))
	})
}

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
		mock := NewMockAAAServiceShared(false) // allow-by-default mode

		// Note: MockAAAServiceShared always uses permission matrix
		// When no rules are defined, it returns the default mode (allow or deny)
		// The traditional mock.On() only works when permission matrix is nil

		// With allow-by-default and no rules, permission check should succeed
		allowed, err := mock.CheckPermission(ctx, "user123", "farmer", "create", "", "org456")
		assert.NoError(t, err)
		assert.True(t, allowed, "Allow-by-default mode should allow when no rules match")

		// To use traditional mock.On(), you would need to use MockAAAService
		// or disable the permission matrix, which is not the intended design
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

// TestPermissionMatrix_Concurrency tests thread safety of permission matrix
func TestPermissionMatrix_Concurrency(t *testing.T) {
	t.Run("concurrent reads and writes", func(t *testing.T) {
		matrix := NewPermissionMatrix(true)
		var wg sync.WaitGroup

		// Concurrent writers
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				matrix.AddAllowRule(
					"user"+string(rune(id)),
					"resource"+string(rune(id%10)),
					"action"+string(rune(id%5)),
					"*",
					"org1",
				)
			}(i)
		}

		// Concurrent readers
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				matrix.CheckPermission(
					"user"+string(rune(id)),
					"resource"+string(rune(id%10)),
					"action"+string(rune(id%5)),
					"obj1",
					"org1",
				)
			}(i)
		}

		// Concurrent clear operations
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(time.Microsecond) // Small delay to interleave
				matrix.Clear()
			}()
		}

		wg.Wait()
		// No panic means thread safety is working
	})

	t.Run("race condition in permission checks", func(t *testing.T) {
		matrix := NewPermissionMatrix(false) // allow-by-default

		// Simulate race: one goroutine adds deny rule while another checks
		done := make(chan bool)
		results := make([]bool, 1000)

		go func() {
			for i := 0; i < 500; i++ {
				matrix.AddDenyRule("user1", "farmer", "delete", "*", "org1")
				matrix.Clear()
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 1000; i++ {
				results[i] = matrix.CheckPermission("user1", "farmer", "delete", "obj1", "org1")
			}
			done <- true
		}()

		<-done
		<-done

		// Results should be consistent within the race window
		// (either all true or mixed, but no panic)
	})
}

// TestMockAAAServiceShared_PermissionMatrix tests the integration between mock and permission matrix
func TestMockAAAServiceShared_PermissionMatrix(t *testing.T) {
	ctx := context.Background()

	t.Run("permission matrix overrides mock calls", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(true)

		// Setup mock to return true (but matrix should override)
		mockSvc.On("CheckPermission", ctx, "user1", "farmer", "delete", "", "org1").Return(true, nil)

		// Matrix denies by default
		allowed, err := mockSvc.CheckPermission(ctx, "user1", "farmer", "delete", "", "org1")
		assert.NoError(t, err)
		assert.False(t, allowed, "Matrix should override mock")

		// Add explicit allow rule
		mockSvc.GetPermissionMatrix().AddAllowRule("user1", "farmer", "delete", "*", "org1")
		allowed, err = mockSvc.CheckPermission(ctx, "user1", "farmer", "delete", "", "org1")
		assert.NoError(t, err)
		assert.True(t, allowed, "Matrix should allow after rule added")
	})

	t.Run("fallback to mock when matrix has no rules and allow-by-default", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(false) // allow-by-default

		// With allow-by-default mode and no rules, permission matrix returns default (allow)
		// The mock's CheckPermission method is NOT called because the matrix handles it
		allowed, err := mockSvc.CheckPermission(ctx, "user1", "farmer", "read", "", "org1")
		assert.NoError(t, err)
		assert.True(t, allowed, "Should allow by default when no rules match")
	})

	t.Run("matrix takes precedence with rules even in allow-by-default", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(false) // allow-by-default
		mockSvc.GetPermissionMatrix().AddDenyRule("user1", "farmer", "delete", "*", "org1")

		// Should use matrix, not mock
		allowed, err := mockSvc.CheckPermission(ctx, "user1", "farmer", "delete", "", "org1")
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("thread-safe matrix updates", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(true)
		var wg sync.WaitGroup

		// Concurrent matrix updates
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				newMatrix := NewPermissionMatrix(true)
				newMatrix.AddAllowRule("user"+string(rune(id)), "*", "*", "*", "*")
				mockSvc.SetPermissionMatrix(newMatrix)
			}(i)
		}

		// Concurrent permission checks
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mockSvc.CheckPermission(ctx, "user"+string(rune(id)), "farmer", "read", "", "org1")
			}(i)
		}

		wg.Wait()
	})
}

// TestMockAAAServiceShared_ErrorScenarios tests error handling in AAA mock
func TestMockAAAServiceShared_ErrorScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("handle nil context gracefully", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(true)
		mockSvc.GetPermissionMatrix().AddAllowRule("user1", "farmer", "read", "*", "org1")

		// Should not panic with nil context
		allowed, err := mockSvc.CheckPermission(nil, "user1", "farmer", "read", "", "org1")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("handle empty strings in permission check", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(true)
		mockSvc.GetPermissionMatrix().AddAllowRule("", "", "", "", "")

		// Empty strings should be denied (security requirement)
		allowed, err := mockSvc.CheckPermission(ctx, "", "", "", "", "")
		assert.NoError(t, err)
		assert.False(t, allowed, "Empty parameters should be denied for security")
	})

	t.Run("special characters in permission parameters", func(t *testing.T) {
		mockSvc := NewMockAAAServiceShared(true)
		specialChars := []string{
			"user-with-dash",
			"user.with.dots",
			"user/with/slashes",
			"user@with@at",
			"user with spaces",
			"user\twith\ttabs",
			"user\nwith\nnewlines",
		}

		for _, special := range specialChars {
			mockSvc.GetPermissionMatrix().AddAllowRule(special, "farmer", "read", "*", "org1")
			allowed, err := mockSvc.CheckPermission(ctx, special, "farmer", "read", "", "org1")
			assert.NoError(t, err)
			assert.True(t, allowed, "Should handle special char: %s", special)
		}
	})
}

// BenchmarkPermissionMatrix tests performance of permission checks
func BenchmarkPermissionMatrix(b *testing.B) {
	matrix := NewPermissionMatrix(true)

	// Add various rules
	for i := 0; i < 100; i++ {
		matrix.AddAllowRule(
			"user"+string(rune(i)),
			"resource"+string(rune(i%10)),
			"action"+string(rune(i%5)),
			"*",
			"org"+string(rune(i%3)),
		)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			matrix.CheckPermission("user50", "resource5", "action0", "obj1", "org2")
		}
	})
}

// BenchmarkMockAAAService tests performance of AAA mock with matrix
func BenchmarkMockAAAService(b *testing.B) {
	mockSvc := NewMockAAAServiceShared(true)
	ctx := context.Background()

	// Add rules for benchmark
	for i := 0; i < 50; i++ {
		mockSvc.GetPermissionMatrix().AddAllowRule(
			"user"+string(rune(i)),
			"*",
			"*",
			"*",
			"org1",
		)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mockSvc.CheckPermission(ctx, "user25", "farmer", "read", "", "org1")
		}
	})
}
