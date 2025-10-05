package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTValidator_GenerateAndValidateToken(t *testing.T) {
	validator, err := NewJWTValidator("test-issuer", 1*time.Hour)
	require.NoError(t, err)

	t.Run("generates valid token", func(t *testing.T) {
		token, err := validator.GenerateToken("user123", "org456", []string{"admin", "farmer"})
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("validates correct token", func(t *testing.T) {
		token, err := validator.GenerateToken("user123", "org456", []string{"admin"})
		require.NoError(t, err)

		claims, err := validator.ValidateToken(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)

		sub, _ := (*claims)["sub"].(string)
		org, _ := (*claims)["org"].(string)

		assert.Equal(t, "user123", sub)
		assert.Equal(t, "org456", org)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := validator.ValidateToken("invalid.token.here")
		assert.Error(t, err)
	})

	t.Run("rejects token with wrong issuer", func(t *testing.T) {
		// Create a validator with different issuer
		validator2, err := NewJWTValidator("different-issuer", 1*time.Hour)
		require.NoError(t, err)

		token, err := validator2.GenerateToken("user123", "org456", []string{"admin"})
		require.NoError(t, err)

		// Try to validate with original validator
		_, err = validator.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		// Create validator with very short validity
		shortValidator, err := NewJWTValidator("test-issuer", 1*time.Millisecond)
		require.NoError(t, err)

		token, err := shortValidator.GenerateToken("user123", "org456", []string{"admin"})
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortValidator.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("extracts roles from token", func(t *testing.T) {
		roles := []string{"admin", "fpo_manager", "kisansathi"}
		token, err := validator.GenerateToken("user123", "org456", roles)
		require.NoError(t, err)

		claims, err := validator.ValidateToken(token)
		require.NoError(t, err)

		tokenRoles, ok := (*claims)["roles"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tokenRoles, 3)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		limiter := NewRateLimiter(10, 1*time.Minute)

		for i := 0; i < 10; i++ {
			allowed, err := limiter.Allow("test-key")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		limiter := NewRateLimiter(5, 1*time.Minute)

		// Use up the limit
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow("test-key")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// Next request should be blocked
		allowed, err := limiter.Allow("test-key")
		assert.Error(t, err)
		assert.False(t, allowed)
	})

	t.Run("resets limit after window", func(t *testing.T) {
		limiter := NewRateLimiter(2, 10*time.Millisecond)

		// Use up the limit
		limiter.Allow("test-key")
		limiter.Allow("test-key")

		// Should be blocked
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)

		// Wait for window to expire
		time.Sleep(15 * time.Millisecond)

		// Should be allowed again
		allowed, err := limiter.Allow("test-key")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("tracks different keys separately", func(t *testing.T) {
		limiter := NewRateLimiter(2, 1*time.Minute)

		// Use up limit for key1
		limiter.Allow("key1")
		limiter.Allow("key1")
		allowed, _ := limiter.Allow("key1")
		assert.False(t, allowed)

		// key2 should still be allowed
		allowed, err := limiter.Allow("key2")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("custom limits per key", func(t *testing.T) {
		limiter := NewRateLimiter(10, 1*time.Minute)

		// Set custom limit for specific key
		limiter.SetLimit("vip-key", 100)

		// Use the custom limit
		for i := 0; i < 50; i++ {
			allowed, err := limiter.Allow("vip-key")
			assert.NoError(t, err)
			assert.True(t, allowed)
		}
	})

	t.Run("reset clears limit", func(t *testing.T) {
		limiter := NewRateLimiter(2, 1*time.Minute)

		// Use up the limit
		limiter.Allow("test-key")
		limiter.Allow("test-key")

		// Should be blocked
		allowed, _ := limiter.Allow("test-key")
		assert.False(t, allowed)

		// Reset the limit
		limiter.Reset("test-key")

		// Should be allowed again
		allowed, err := limiter.Allow("test-key")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("can be disabled", func(t *testing.T) {
		limiter := NewRateLimiter(1, 1*time.Minute)

		// Use up the limit
		limiter.Allow("test-key")

		// Disable rate limiting
		limiter.Disable()

		// Should be allowed even though limit is exceeded
		allowed, err := limiter.Allow("test-key")
		assert.NoError(t, err)
		assert.True(t, allowed)

		// Re-enable and verify it works
		limiter.Enable()
		limiter.Reset("test-key")
		limiter.Allow("test-key")
		allowed, _ = limiter.Allow("test-key")
		assert.False(t, allowed)
	})
}

func TestAuditLogger(t *testing.T) {
	t.Run("logs events", func(t *testing.T) {
		logger := NewAuditLogger()

		event := AuditEvent{
			EventType: "permission_check",
			UserID:    "user123",
			OrgID:     "org456",
			Resource:  "farmer",
			Action:    "create",
			Result:    "success",
		}

		logger.LogEvent(event)

		events := logger.GetEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "permission_check", events[0].EventType)
	})

	t.Run("filters events by user", func(t *testing.T) {
		logger := NewAuditLogger()

		logger.LogEvent(AuditEvent{UserID: "user1", EventType: "action1"})
		logger.LogEvent(AuditEvent{UserID: "user2", EventType: "action2"})
		logger.LogEvent(AuditEvent{UserID: "user1", EventType: "action3"})

		user1Events := logger.GetEventsByUser("user1")
		assert.Len(t, user1Events, 2)
	})

	t.Run("filters events by type", func(t *testing.T) {
		logger := NewAuditLogger()

		logger.LogEvent(AuditEvent{EventType: "login", UserID: "user1"})
		logger.LogEvent(AuditEvent{EventType: "permission_check", UserID: "user1"})
		logger.LogEvent(AuditEvent{EventType: "login", UserID: "user2"})

		loginEvents := logger.GetEventsByType("login")
		assert.Len(t, loginEvents, 2)
	})

	t.Run("clears all events", func(t *testing.T) {
		logger := NewAuditLogger()

		logger.LogEvent(AuditEvent{EventType: "test1"})
		logger.LogEvent(AuditEvent{EventType: "test2"})

		assert.Len(t, logger.GetEvents(), 2)

		logger.Clear()

		assert.Len(t, logger.GetEvents(), 0)
	})

	t.Run("adds timestamps automatically", func(t *testing.T) {
		logger := NewAuditLogger()

		before := time.Now()
		logger.LogEvent(AuditEvent{EventType: "test"})
		after := time.Now()

		events := logger.GetEvents()
		require.Len(t, events, 1)

		eventTime := events[0].Timestamp
		assert.True(t, eventTime.After(before) || eventTime.Equal(before))
		assert.True(t, eventTime.Before(after) || eventTime.Equal(after))
	})
}

func TestSecurityEnhancedMockAAA(t *testing.T) {
	t.Run("creates with security enabled", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)
		assert.NotNil(t, mock)
		assert.True(t, mock.securityEnabled)
	})

	t.Run("JWT validation integration", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		// Generate a test token
		token, err := mock.GenerateTestToken("user123", "org456", []string{"admin"})
		require.NoError(t, err)

		// Validate the token
		ctx := context.Background()
		userInfo, err := mock.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, "user123", userInfo.UserID)
		assert.Equal(t, "org456", userInfo.OrgID)

		// Verify audit log
		events := mock.GetAuditEvents()
		assert.Greater(t, len(events), 0)

		// Find token validation event
		var found bool
		for _, event := range events {
			if event.EventType == "token_validation_success" {
				found = true
				assert.Equal(t, "user123", event.UserID)
				break
			}
		}
		assert.True(t, found, "Token validation should be audited")
	})

	t.Run("rate limiting integration", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		// Set low rate limit for testing
		mock.GetRateLimiter().SetLimit("check_permission:user123", 3)

		ctx := context.Background()

		// Configure permissions
		matrix := mock.GetPermissionMatrix()
		matrix.AddAllowRule("user123", "farmer", "read", "*", "org123")

		// Make requests up to the limit
		for i := 0; i < 3; i++ {
			_, err := mock.CheckPermission(ctx, "user123", "farmer", "read", "*", "org123")
			assert.NoError(t, err)
		}

		// Next request should be rate limited
		_, err = mock.CheckPermission(ctx, "user123", "farmer", "read", "*", "org123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")

		// Verify audit log contains rate limit event
		events := mock.GetAuditEvents()
		var rateLimitEvent bool
		for _, event := range events {
			if event.EventType == "rate_limit_exceeded" {
				rateLimitEvent = true
				assert.Equal(t, "user123", event.UserID)
				break
			}
		}
		assert.True(t, rateLimitEvent, "Rate limit violation should be audited")
	})

	t.Run("permission check audit logging", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()
		matrix := mock.GetPermissionMatrix()
		matrix.AddAllowRule("user123", "farmer", "create", "*", "org123")

		// Make permission check
		mock.CheckPermission(ctx, "user123", "farmer", "create", "*", "org123")

		// Verify audit log
		events := mock.GetAuditEvents()
		var permCheckEvent bool
		for _, event := range events {
			if event.EventType == "permission_check" {
				permCheckEvent = true
				assert.Equal(t, "user123", event.UserID)
				assert.Equal(t, "farmer", event.Resource)
				assert.Equal(t, "create", event.Action)
				assert.Equal(t, "success", event.Result)
				break
			}
		}
		assert.True(t, permCheckEvent, "Permission check should be audited")
	})

	t.Run("security can be disabled for simple tests", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(false)
		require.NoError(t, err)

		ctx := context.Background()
		matrix := mock.GetPermissionMatrix()
		matrix.AddAllowRule("user123", "farmer", "read", "*", "org123")

		// Make many requests without rate limiting
		for i := 0; i < 100; i++ {
			_, err := mock.CheckPermission(ctx, "user123", "farmer", "read", "*", "org123")
			assert.NoError(t, err)
		}

		// Audit log should be empty (security disabled)
		events := mock.GetAuditEvents()
		assert.Empty(t, events)
	})

	t.Run("security can be toggled", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()

		// Disable security
		mock.DisableSecurity()

		// Should work without audit logging
		mock.CheckPermission(ctx, "user", "resource", "action", "*", "org")
		assert.Empty(t, mock.GetAuditEvents())

		// Enable security
		mock.EnableSecurity()

		// Should now have audit logging
		mock.CheckPermission(ctx, "user", "resource", "action", "*", "org")
		assert.NotEmpty(t, mock.GetAuditEvents())
	})
}

// TestAttackScenarios tests common attack vectors
func TestAttackScenarios(t *testing.T) {
	t.Run("SQL injection attempt in permission check", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()
		sqlInjection := "admin' OR '1'='1"

		// Should not panic or allow bypass
		allowed, err := mock.CheckPermission(ctx, sqlInjection, "farmer", "delete", "*", "org123")
		assert.False(t, allowed)
		// The mock should handle this gracefully
	})

	t.Run("XSS attempt in user input", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()
		xssPayload := "<script>alert('xss')</script>"

		// Should handle XSS payloads without issues
		_, err = mock.CheckPermission(ctx, xssPayload, "farmer", "read", "*", "org123")
		assert.NoError(t, err) // Should not panic
	})

	t.Run("brute force rate limiting", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		// Set very low limit to simulate brute force protection
		mock.GetRateLimiter().SetLimit("check_permission:attacker", 5)

		ctx := context.Background()
		var blocked int

		// Simulate brute force attack
		for i := 0; i < 100; i++ {
			_, err := mock.CheckPermission(ctx, "attacker", "farmer", "read", "*", "org123")
			if err != nil {
				blocked++
			}
		}

		// Most requests should be blocked
		assert.Greater(t, blocked, 90, "Rate limiter should block brute force attempts")

		// Verify audit log captures attack
		events := mock.GetAuditEvents()
		var rateLimitEvents int
		for _, event := range events {
			if event.EventType == "rate_limit_exceeded" {
				rateLimitEvents++
			}
		}
		assert.Greater(t, rateLimitEvents, 0, "Attack attempts should be audited")
	})

	t.Run("token forgery attempt", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()

		// Try to use a forged token
		forgedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhdHRhY2tlciIsIm9yZyI6ImFsbCJ9.invalid"

		_, err = mock.ValidateToken(ctx, forgedToken)
		assert.Error(t, err, "Forged token should be rejected")

		// Verify audit log
		events := mock.GetAuditEvents()
		var validationFailure bool
		for _, event := range events {
			if event.EventType == "token_validation_failed" {
				validationFailure = true
				break
			}
		}
		assert.True(t, validationFailure, "Token forgery attempt should be audited")
	})

	t.Run("privilege escalation attempt", func(t *testing.T) {
		mock, err := NewSecurityEnhancedMockAAA(true)
		require.NoError(t, err)

		ctx := context.Background()
		matrix := mock.GetPermissionMatrix()

		// Setup farmer with limited permissions
		matrix.AddAllowRule("farmer", "farmer", "read", "farmer", "org123")

		// Farmer tries to escalate to admin permissions
		allowed, _ := mock.CheckPermission(ctx, "farmer", "admin", "create", "*", "org123")
		assert.False(t, allowed, "Privilege escalation should be denied")

		// Farmer tries to access different org
		allowed, _ = mock.CheckPermission(ctx, "farmer", "farmer", "read", "*", "different_org")
		assert.False(t, allowed, "Cross-org access should be denied")
	})
}

// Benchmark security overhead
func BenchmarkSecurityEnhancedMock_CheckPermission(b *testing.B) {
	mock, _ := NewSecurityEnhancedMockAAA(true)
	ctx := context.Background()
	matrix := mock.GetPermissionMatrix()
	matrix.AddAllowRule("user123", "farmer", "read", "*", "org123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.CheckPermission(ctx, "user123", "farmer", "read", fmt.Sprintf("farm%d", i%100), "org123")
	}
}

func BenchmarkSecurityEnhancedMock_ValidateToken(b *testing.B) {
	mock, _ := NewSecurityEnhancedMockAAA(true)
	token, _ := mock.GenerateTestToken("user123", "org456", []string{"admin"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ValidateToken(ctx, token)
	}
}
