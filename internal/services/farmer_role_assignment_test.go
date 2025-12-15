package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/constants"
	"github.com/stretchr/testify/assert"
)

// TestFarmerService_EnsureFarmerRole_Success tests successful farmer role assignment
func TestFarmerService_EnsureFarmerRole_Success(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil, // Not needed for this test
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations for successful role assignment
	// 1. Check role doesn't exist initially
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	// 2. Assign role
	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(nil).Once()

	// Execute
	err := service.ensureFarmerRole(ctx, userID, orgID)

	// Verify
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRole_AlreadyExists tests idempotent role assignment
func TestFarmerService_EnsureFarmerRole_AlreadyExists(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations - role already exists
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(true, nil).Once()

	// AssignRole should NOT be called (idempotent check prevents it)

	// Execute
	err := service.ensureFarmerRole(ctx, userID, orgID)

	// Verify
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
	// Verify AssignRole was NOT called
	mockAAA.AssertNotCalled(t, "AssignRole")
}

// TestFarmerService_EnsureFarmerRole_AssignmentFails tests failure handling
func TestFarmerService_EnsureFarmerRole_AssignmentFails(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(assert.AnError).Once()

	// Execute
	err := service.ensureFarmerRole(ctx, userID, orgID)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign farmer role")
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRole_AlreadyAssignedError tests handling of "already assigned" error
// This handles eventual consistency where CheckUserRole says false but AssignRole says already assigned
func TestFarmerService_EnsureFarmerRole_AlreadyAssignedError(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations - simulating eventual consistency issue
	// 1. Check role says not present
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	// 2. Assign role fails with "already assigned" (eventual consistency)
	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(fmt.Errorf("role already assigned to user")).Once()

	// Execute
	err := service.ensureFarmerRole(ctx, userID, orgID)

	// Verify - should succeed because "already assigned" is treated as success
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRoleWithRetry_SucceedsOnFirstAttempt tests retry logic
func TestFarmerService_EnsureFarmerRoleWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations for successful first attempt
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(nil).Once()

	// Execute
	err := service.ensureFarmerRoleWithRetry(ctx, userID, orgID)

	// Verify
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRoleWithRetry_SucceedsOnSecondAttempt tests retry success
func TestFarmerService_EnsureFarmerRoleWithRetry_SucceedsOnSecondAttempt(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// First attempt fails
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(assert.AnError).Once()

	// Second attempt succeeds
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(nil).Once()

	// Execute
	err := service.ensureFarmerRoleWithRetry(ctx, userID, orgID)

	// Verify
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRoleWithRetry_FailsAfterTwoAttempts tests retry exhaustion
func TestFarmerService_EnsureFarmerRoleWithRetry_FailsAfterTwoAttempts(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Both attempts fail
	// First attempt
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(assert.AnError).Once()

	// Second attempt
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, nil).Once()

	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(assert.AnError).Once()

	// Execute
	err := service.ensureFarmerRoleWithRetry(ctx, userID, orgID)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role assignment failed after 2 attempts")
	mockAAA.AssertExpectations(t)
}

// TestFarmerService_EnsureFarmerRole_CheckRoleFails_ContinuesWithAssignment tests degraded mode
func TestFarmerService_EnsureFarmerRole_CheckRoleFails_ContinuesWithAssignment(t *testing.T) {
	// Setup
	mockAAA := new(MockAAAService)

	service := &FarmerServiceImpl{
		repository:      nil,
		aaaService:      mockAAA,
		defaultPassword: "test123",
	}

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	// Mock expectations
	// 1. Initial check fails (AAA degraded)
	mockAAA.On("CheckUserRole", ctx, userID, constants.RoleFarmer).
		Return(false, assert.AnError).Once()

	// 2. Assignment should still be attempted
	mockAAA.On("AssignRole", ctx, userID, orgID, constants.RoleFarmer).
		Return(nil).Once()

	// Execute
	err := service.ensureFarmerRole(ctx, userID, orgID)

	// Verify - should succeed despite initial check failure
	assert.NoError(t, err)
	mockAAA.AssertExpectations(t)
}

// TestRoleConstants tests that role constants are properly defined
func TestRoleConstants(t *testing.T) {
	// Verify all expected roles are defined
	assert.Equal(t, "farmer", constants.RoleFarmer)
	assert.Equal(t, "kisansathi", constants.RoleKisanSathi)
	assert.Equal(t, "CEO", constants.RoleFPOCEO)
	assert.Equal(t, "fpo_manager", constants.RoleFPOManager)
	assert.Equal(t, "admin", constants.RoleAdmin)
	assert.Equal(t, "super_admin", constants.RoleSuperAdmin)
	assert.Equal(t, "readonly", constants.RoleReadOnly)

	// Verify AllRoles returns all expected roles
	allRoles := constants.AllRoles()
	assert.Len(t, allRoles, 7)
	assert.Contains(t, allRoles, constants.RoleFarmer)
	assert.Contains(t, allRoles, constants.RoleKisanSathi)
	assert.Contains(t, allRoles, constants.RoleFPOCEO)
	assert.Contains(t, allRoles, constants.RoleSuperAdmin)

	// Verify IsValidRole works correctly
	assert.True(t, constants.IsValidRole(constants.RoleFarmer))
	assert.True(t, constants.IsValidRole(constants.RoleFPOCEO))
	assert.False(t, constants.IsValidRole("invalid_role"))
	assert.False(t, constants.IsValidRole(""))

	// Verify display names
	assert.Equal(t, "Farmer", constants.GetRoleDisplayName(constants.RoleFarmer))
	assert.Equal(t, "FPO CEO", constants.GetRoleDisplayName(constants.RoleFPOCEO))

	// Verify descriptions exist
	farmerDesc := constants.GetRoleDescription(constants.RoleFarmer)
	assert.NotEmpty(t, farmerDesc)
	assert.Contains(t, farmerDesc, "agricultural practitioner")
}
