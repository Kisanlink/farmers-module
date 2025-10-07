package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BehaviorDriftDetector detects when mock behavior differs from expected real behavior
type BehaviorDriftDetector struct {
	violations []BehaviorViolation
}

// BehaviorViolation represents a detected behavioral difference
type BehaviorViolation struct {
	TestName         string
	ExpectedBehavior string
	ActualBehavior   string
	Severity         string // "critical", "high", "medium", "low"
	Description      string
}

// NewBehaviorDriftDetector creates a new drift detector
func NewBehaviorDriftDetector() *BehaviorDriftDetector {
	return &BehaviorDriftDetector{
		violations: make([]BehaviorViolation, 0),
	}
}

// RecordViolation records a behavioral violation
func (d *BehaviorDriftDetector) RecordViolation(violation BehaviorViolation) {
	d.violations = append(d.violations, violation)
}

// HasViolations checks if any violations were detected
func (d *BehaviorDriftDetector) HasViolations() bool {
	return len(d.violations) > 0
}

// GetViolations returns all detected violations
func (d *BehaviorDriftDetector) GetViolations() []BehaviorViolation {
	return d.violations
}

// PrintReport prints a formatted report of all violations
func (d *BehaviorDriftDetector) PrintReport(t *testing.T) {
	if !d.HasViolations() {
		t.Log("✅ No behavior drift detected - mock matches expected real behavior")
		return
	}

	t.Errorf("⚠️ Behavior drift detected: %d violations found", len(d.violations))
	for i, violation := range d.violations {
		t.Errorf("\nViolation %d/%d:", i+1, len(d.violations))
		t.Errorf("  Test: %s", violation.TestName)
		t.Errorf("  Severity: %s", violation.Severity)
		t.Errorf("  Description: %s", violation.Description)
		t.Errorf("  Expected: %s", violation.ExpectedBehavior)
		t.Errorf("  Actual: %s", violation.ActualBehavior)
	}
}

// TestAAA_ServiceBehaviorParity validates mock behavior matches real service expectations
func TestAAA_ServiceBehaviorParity(t *testing.T) {
	detector := NewBehaviorDriftDetector()
	ctx := context.Background()

	t.Run("permission check behavior", func(t *testing.T) {
		// Create deny-by-default mock
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()

		// Test 1: Deny-by-default behavior
		allowed, err := mock.CheckPermission(ctx, "unknown_user", "farmer", "delete", "*", "org123")
		if err != nil {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_unknown_user",
				ExpectedBehavior: "should not error, should return false",
				ActualBehavior:   fmt.Sprintf("returned error: %v", err),
				Severity:         "high",
				Description:      "Mock should handle unknown users gracefully without errors",
			})
		} else if allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_unknown_user",
				ExpectedBehavior: "should return false for unknown user",
				ActualBehavior:   "returned true",
				Severity:         "critical",
				Description:      "Deny-by-default policy violated - security issue",
			})
		}

		// Test 2: Wildcard matching behavior
		matrix.AddAllowRule("admin", "*", "*", "*", "org123")
		allowed, err = mock.CheckPermission(ctx, "admin", "farmer", "create", "resource123", "org123")
		if err != nil {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_wildcard",
				ExpectedBehavior: "wildcard permission should match specific resource",
				ActualBehavior:   fmt.Sprintf("returned error: %v", err),
				Severity:         "high",
				Description:      "Wildcard permission matching failed",
			})
		} else if !allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_wildcard",
				ExpectedBehavior: "wildcard should match all resources",
				ActualBehavior:   "returned false",
				Severity:         "high",
				Description:      "Wildcard permission not working as expected",
			})
		}

		// Test 3: First-match-wins rule precedence
		matrix.Clear()
		matrix.AddDenyRule("user123", "farmer", "delete", "*", "org123")
		matrix.AddAllowRule("user123", "farmer", "delete", "*", "org123")

		allowed, err = mock.CheckPermission(ctx, "user123", "farmer", "delete", "*", "org123")
		if err == nil && allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_precedence",
				ExpectedBehavior: "first matching rule should win (deny)",
				ActualBehavior:   "second rule was applied (allow)",
				Severity:         "critical",
				Description:      "Rule precedence not working correctly - security issue",
			})
		}

		// Test 4: Empty object handling
		matrix.Clear()
		matrix.AddAllowRule("user123", "farmer", "read", "*", "org123")
		allowed, err = mock.CheckPermission(ctx, "user123", "farmer", "read", "", "org123")
		if err != nil {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "permission_check_empty_object",
				ExpectedBehavior: "empty object should be treated as wildcard",
				ActualBehavior:   fmt.Sprintf("returned error: %v", err),
				Severity:         "medium",
				Description:      "Empty object parameter handling incorrect",
			})
		}
	})

	t.Run("preset configuration behavior", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeDenyAll)

		// Test admin preset
		adminMock := factory.NewAAAServiceMockWithPreset(PresetAdmin, "admin123", "org123")

		// Admin should have all permissions in their org
		testCases := []struct {
			resource string
			action   string
			expected bool
		}{
			{"farmer", "create", true},
			{"farmer", "delete", true},
			{"farm", "update", true},
			{"cycle", "start", true},
			{"fpo", "manage", true},
		}

		for _, tc := range testCases {
			allowed, err := adminMock.CheckPermission(ctx, "admin123", tc.resource, tc.action, "*", "org123")
			if err != nil || allowed != tc.expected {
				detector.RecordViolation(BehaviorViolation{
					TestName:         fmt.Sprintf("admin_preset_%s_%s", tc.resource, tc.action),
					ExpectedBehavior: fmt.Sprintf("admin should have %s permission on %s", tc.action, tc.resource),
					ActualBehavior:   fmt.Sprintf("allowed=%v, err=%v", allowed, err),
					Severity:         "high",
					Description:      "Admin preset permissions incomplete",
				})
			}
		}

		// Admin should NOT have permissions in different org
		allowed, err := adminMock.CheckPermission(ctx, "admin123", "farmer", "create", "*", "different_org")
		if err == nil && allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "admin_preset_org_isolation",
				ExpectedBehavior: "admin should only have permissions in their org",
				ActualBehavior:   "has permissions in different org",
				Severity:         "critical",
				Description:      "Organization isolation violated - security issue",
			})
		}
	})

	t.Run("farmer preset behavior", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeDenyAll)
		farmerMock := factory.NewAAAServiceMockWithPreset(PresetFarmer, "farmer123", "org123")

		// Farmer should NOT have delete permissions
		allowed, err := farmerMock.CheckPermission(ctx, "farmer123", "farmer", "delete", "*", "org123")
		if err == nil && allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "farmer_preset_delete_restriction",
				ExpectedBehavior: "farmer should not have delete permissions",
				ActualBehavior:   "has delete permissions",
				Severity:         "critical",
				Description:      "Farmer preset too permissive - security issue",
			})
		}

		// Farmer should have read permissions
		allowed, err = farmerMock.CheckPermission(ctx, "farmer123", "farmer", "read", "*", "org123")
		if err != nil || !allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "farmer_preset_read_permission",
				ExpectedBehavior: "farmer should have read permissions",
				ActualBehavior:   fmt.Sprintf("allowed=%v, err=%v", allowed, err),
				Severity:         "high",
				Description:      "Farmer preset missing expected permissions",
			})
		}
	})

	t.Run("readonly preset behavior", func(t *testing.T) {
		factory := NewMockFactory(PermissionModeDenyAll)
		readonlyMock := factory.NewAAAServiceMockWithPreset(PresetReadOnly, "viewer123", "org123")

		// Readonly should have read and list only
		writeActions := []string{"create", "update", "delete", "manage"}
		for _, action := range writeActions {
			allowed, err := readonlyMock.CheckPermission(ctx, "viewer123", "farmer", action, "*", "org123")
			if err == nil && allowed {
				detector.RecordViolation(BehaviorViolation{
					TestName:         fmt.Sprintf("readonly_preset_write_action_%s", action),
					ExpectedBehavior: fmt.Sprintf("readonly should not have %s permission", action),
					ActualBehavior:   "has write permission",
					Severity:         "critical",
					Description:      "Readonly preset allowing write operations - security issue",
				})
			}
		}

		// Should have read permission
		allowed, err := readonlyMock.CheckPermission(ctx, "viewer123", "farmer", "read", "*", "org123")
		if err != nil || !allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "readonly_preset_read_permission",
				ExpectedBehavior: "readonly should have read permission",
				ActualBehavior:   fmt.Sprintf("allowed=%v, err=%v", allowed, err),
				Severity:         "medium",
				Description:      "Readonly preset missing expected permissions",
			})
		}
	})

	t.Run("permission matrix rule ordering", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()

		// Add rules in specific order
		matrix.AddDenyRule("user", "resource", "action", "object", "org")
		matrix.AddAllowRule("user", "resource", "action", "object", "org")
		matrix.AddAllowRule("user", "resource", "*", "*", "org")

		// First matching rule should win (deny)
		allowed, _ := mock.CheckPermission(ctx, "user", "resource", "action", "object", "org")
		if allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "rule_ordering_first_match_wins",
				ExpectedBehavior: "first deny rule should win",
				ActualBehavior:   "later allow rule was applied",
				Severity:         "critical",
				Description:      "Rule evaluation order incorrect - security issue",
			})
		}
	})

	// Print final report
	detector.PrintReport(t)
	assert.False(t, detector.HasViolations(), "Behavior drift detected - mock does not match expected real service behavior")
}

// TestEdgeCaseBehavior tests edge cases that might differ between mock and real
func TestEdgeCaseBehavior(t *testing.T) {
	detector := NewBehaviorDriftDetector()
	ctx := context.Background()

	t.Run("null and empty string handling", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)

		// Test with empty strings - should deny but not error
		allowed, err := mock.CheckPermission(ctx, "", "", "", "", "")

		// Empty parameters should be denied (returned false) but not cause an error
		// This is the correct behavior - the permission check validates and returns false
		if err != nil {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "empty_parameters",
				ExpectedBehavior: "should deny empty parameters without error",
				ActualBehavior:   "returned error for empty parameters",
				Severity:         "low",
				Description:      "Permission check should deny gracefully, not error",
			})
		}

		if allowed {
			detector.RecordViolation(BehaviorViolation{
				TestName:         "empty_parameters",
				ExpectedBehavior: "should deny empty parameters",
				ActualBehavior:   "allowed empty parameters",
				Severity:         "medium",
				Description:      "Empty parameters should always be denied for security",
			})
		}
	})

	t.Run("special character handling", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()

		// Test with special characters
		specialChars := []string{"*", "**", ".", ":", "/", "\\"}
		for _, char := range specialChars {
			matrix.Clear()
			matrix.AddAllowRule("user", char, "action", "*", "org")

			allowed, err := mock.CheckPermission(ctx, "user", char, "action", "*", "org")
			if err != nil || !allowed {
				detector.RecordViolation(BehaviorViolation{
					TestName:         fmt.Sprintf("special_char_handling_%s", char),
					ExpectedBehavior: "should handle special characters in resource names",
					ActualBehavior:   fmt.Sprintf("failed for char: %s", char),
					Severity:         "medium",
					Description:      "Special character handling may differ from real service",
				})
			}
		}
	})

	t.Run("concurrent access behavior", func(t *testing.T) {
		mock := NewMockAAAServiceShared(true)
		matrix := mock.GetPermissionMatrix()

		// Simulate concurrent permission matrix modifications
		done := make(chan bool, 2)

		go func() {
			for i := 0; i < 100; i++ {
				matrix.AddAllowRule(fmt.Sprintf("user%d", i), "resource", "action", "*", "org")
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				_, _ = mock.CheckPermission(ctx, fmt.Sprintf("user%d", i), "resource", "action", "*", "org")
			}
			done <- true
		}()

		<-done
		<-done

		// If we got here without panic, concurrent access is handled
		t.Log("✅ Concurrent access handled without panic")
	})

	detector.PrintReport(t)
}

// TestMockFactoryValidation validates the mock factory creates compliant mocks
func TestMockFactoryValidation(t *testing.T) {
	factory := NewMockFactory(PermissionModeDenyAll)

	t.Run("factory creates interface-compliant mocks", func(t *testing.T) {
		// Validate AAA service mock
		aaaService := factory.NewAAAServiceMock()
		assert.NotNil(t, aaaService)
		assert.NotNil(t, aaaService.GetPermissionMatrix())

		// Validate cache mock
		cache := factory.NewCacheMock()
		assert.NotNil(t, cache)

		// Validate event emitter mock
		eventEmitter := factory.NewEventEmitterMock()
		assert.NotNil(t, eventEmitter)

		// Validate database mock
		database := factory.NewDatabaseMock()
		assert.NotNil(t, database)

		// Validate farmer linkage repo mock
		farmerRepo := factory.NewFarmerLinkageRepoMock()
		assert.NotNil(t, farmerRepo)

		// Validate data quality service mock
		dataQuality := factory.NewDataQualityServiceMock()
		assert.NotNil(t, dataQuality)
	})

	t.Run("deny-all factory creates secure mocks", func(t *testing.T) {
		secureFactory := NewMockFactory(PermissionModeDenyAll)
		mock := secureFactory.NewAAAServiceMock()

		ctx := context.Background()
		allowed, err := mock.CheckPermission(ctx, "any_user", "any_resource", "any_action", "*", "any_org")

		assert.NoError(t, err)
		assert.False(t, allowed, "Deny-all factory should create secure deny-by-default mocks")
	})

	t.Run("allow-all factory creates permissive mocks", func(t *testing.T) {
		permissiveFactory := NewMockFactory(PermissionModeAllowAll)
		mock := permissiveFactory.NewAAAServiceMock()

		ctx := context.Background()
		allowed, err := mock.CheckPermission(ctx, "any_user", "any_resource", "any_action", "*", "any_org")

		assert.NoError(t, err)
		assert.True(t, allowed, "Allow-all factory should create permissive mocks")
	})
}
