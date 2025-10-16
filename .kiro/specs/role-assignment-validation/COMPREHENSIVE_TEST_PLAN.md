# Comprehensive Role Assignment Test Plan

**Purpose:** Complete testing strategy for validating role assignment business logic across all user creation and management flows

**Target Coverage:** >95% code coverage for role assignment paths
**Estimated Effort:** 5-7 days (1 developer)

---

## Test Strategy

### Testing Pyramid

```
                    /\
                   /  \
                  / E2E \              5 tests (5%)
                 /------\
                /        \
               /Integration\           20 tests (20%)
              /------------\
             /              \
            /   Unit Tests   \        75 tests (75%)
           /------------------\
```

### Testing Levels

1. **Unit Tests (75%)**: Test individual functions in isolation with mocks
2. **Integration Tests (20%)**: Test service integration with real AAA client (test instance)
3. **End-to-End Tests (5%)**: Test complete workflows through API with real infrastructure

---

## Test Suite 1: Farmer Role Assignment - Unit Tests

### File: `internal/services/farmer_service_test.go`

#### Test Group 1.1: Single Farmer Creation with Role Assignment

```go
package services_test

import (
    "context"
    "fmt"
    "testing"

    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/services"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Test 1: Happy path - farmer created with role assigned
func TestCreateFarmer_AssignsRoleSuccessfully(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFarmerRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFarmerService(mockRepo, mockAAA, "password123")

    // Mock AAA user creation
    mockAAA.On("CreateUser", mock.Anything, mock.MatchedBy(func(req interface{}) bool {
        reqMap := req.(map[string]interface{})
        return reqMap["phone_number"] == "9876543210"
    })).Return(map[string]interface{}{
        "id": "user-abc-123",
        "username": "farmer_9876543210",
        "status": "active",
    }, nil)

    // Mock role assignment - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "user-abc-123", "org-456", "farmer").
        Return(nil)

    // Mock organization verification
    mockAAA.On("GetOrganization", mock.Anything, "org-456").
        Return(map[string]interface{}{"id": "org-456", "name": "Test FPO"}, nil)

    // Mock repository save
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockRepo.On("FindOne", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("not found"))

    // ACT
    req := &requests.CreateFarmerRequest{
        AAAOrgID: "org-456",
        Profile: requests.FarmerProfileData{
            FirstName:   "John",
            LastName:    "Doe",
            PhoneNumber: "9876543210",
            Email:       "john@example.com",
            CountryCode: "+91",
        },
    }
    req.UserID = "admin-123"

    response, err := service.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.NoError(t, err, "Farmer creation should succeed")
    assert.NotNil(t, response, "Response should not be nil")
    assert.Equal(t, "user-abc-123", response.Data.AAAUserID)

    // CRITICAL: Verify role was assigned
    mockAAA.AssertCalled(t, "AssignRole", mock.Anything, "user-abc-123", "org-456", "farmer")
    mockAAA.AssertNumberOfCalls(t, "AssignRole", 1)
}

// Test 2: Role assignment fails - farmer creation should be rolled back
func TestCreateFarmer_RoleAssignmentFails_RollsBack(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFarmerRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFarmerService(mockRepo, mockAAA, "password123")

    // Mock AAA user creation succeeds
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "user-xyz-789"}, nil)

    // Mock role assignment FAILS - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "user-xyz-789", "org-456", "farmer").
        Return(fmt.Errorf("AAA service unavailable: role assignment failed"))

    // Mock AAA user deletion for rollback
    mockAAA.On("DeleteUser", mock.Anything, "user-xyz-789").
        Return(nil)

    mockAAA.On("GetOrganization", mock.Anything, "org-456").
        Return(map[string]interface{}{"id": "org-456"}, nil)
    mockRepo.On("FindOne", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("not found"))

    // ACT
    req := &requests.CreateFarmerRequest{
        AAAOrgID: "org-456",
        Profile: requests.FarmerProfileData{
            FirstName:   "Jane",
            LastName:    "Smith",
            PhoneNumber: "9123456789",
            CountryCode: "+91",
        },
    }
    req.UserID = "admin-123"

    response, err := service.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.Error(t, err, "Farmer creation should fail when role assignment fails")
    assert.Contains(t, err.Error(), "role assignment failed")
    assert.Nil(t, response, "Response should be nil on failure")

    // Verify rollback occurred
    mockAAA.AssertCalled(t, "DeleteUser", mock.Anything, "user-xyz-789")

    // Verify farmer was NOT saved to repository
    mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

// Test 3: AAA user exists, farmer profile doesn't, role missing - should assign role
func TestCreateFarmer_ExistingUserWithoutRole_AssignsRole(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFarmerRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFarmerService(mockRepo, mockAAA, "password123")

    // Mock AAA user creation fails (user exists)
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(nil, fmt.Errorf("user already exists"))

    // Mock GetUserByMobile returns existing user
    mockAAA.On("GetUserByMobile", mock.Anything, "9876543210").
        Return(map[string]interface{}{
            "id": "existing-user-456",
            "phone_number": "9876543210",
        }, nil)

    // Mock CheckUserRole returns false (no farmer role)
    mockAAA.On("CheckUserRole", mock.Anything, "existing-user-456", "farmer").
        Return(false, nil)

    // Mock role assignment - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "existing-user-456", "org-789", "farmer").
        Return(nil)

    mockAAA.On("GetOrganization", mock.Anything, "org-789").
        Return(map[string]interface{}{"id": "org-789"}, nil)

    // Mock repository - no existing farmer
    mockRepo.On("FindOne", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("not found"))
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // ACT
    req := &requests.CreateFarmerRequest{
        AAAOrgID: "org-789",
        Profile: requests.FarmerProfileData{
            PhoneNumber: "9876543210",
            FirstName:   "Existing",
            LastName:    "User",
            CountryCode: "+91",
        },
    }
    req.UserID = "admin-123"

    response, err := service.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.NoError(t, err, "Should succeed when existing user gets farmer role")
    assert.NotNil(t, response)

    // CRITICAL: Verify role was assigned to existing user
    mockAAA.AssertCalled(t, "CheckUserRole", mock.Anything, "existing-user-456", "farmer")
    mockAAA.AssertCalled(t, "AssignRole", mock.Anything, "existing-user-456", "org-789", "farmer")
}

// Test 4: Idempotent creation - existing farmer already has role
func TestCreateFarmer_Idempotent_ExistingFarmerWithRole(t *testing.T) {
    // Test that when farmer already exists with role, operation is idempotent
    // Should return existing farmer without errors
    // Should NOT attempt to assign role again
    // ... implementation ...
}

// Test 5: Role assignment timeout - should fail and rollback
func TestCreateFarmer_RoleAssignmentTimeout_RollsBack(t *testing.T) {
    // Simulate network timeout during role assignment
    // Verify farmer creation fails
    // Verify AAA user is deleted (rollback)
    // ... implementation ...
}
```

---

#### Test Group 1.2: Concurrent Farmer Creation

```go
// Test 6: Concurrent creation with same phone - one succeeds with role
func TestCreateFarmer_Concurrent_SamePhone_OneSucceedsWithRole(t *testing.T) {
    t.Parallel()

    // ARRANGE
    numGoroutines := 10
    successCount := 0
    errors := make([]error, numGoroutines)
    results := make([]*responses.FarmerResponse, numGoroutines)

    // Shared AAA service mock (with mutex for thread safety)
    mockAAA := new(ThreadSafeMockAAAService)
    mockRepo := new(ThreadSafeMockFarmerRepository)
    service := services.NewFarmerService(mockRepo, mockAAA, "password123")

    // Configure mock to allow only ONE successful user creation
    userCreated := false
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(func(ctx context.Context, req interface{}) (interface{}, error) {
            if !userCreated {
                userCreated = true
                return map[string]interface{}{"id": "user-concurrent-123"}, nil
            }
            return nil, fmt.Errorf("user already exists")
        })

    mockAAA.On("GetUserByMobile", mock.Anything, "9111111111").
        Return(map[string]interface{}{"id": "user-concurrent-123"}, nil)

    mockAAA.On("AssignRole", mock.Anything, "user-concurrent-123", "org-concurrent", "farmer").
        Return(nil)

    mockAAA.On("CheckUserRole", mock.Anything, "user-concurrent-123", "farmer").
        Return(true, nil) // Role already assigned by first request

    // ... repository mocks ...

    // ACT
    var wg sync.WaitGroup
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            req := &requests.CreateFarmerRequest{
                AAAOrgID: "org-concurrent",
                Profile: requests.FarmerProfileData{
                    PhoneNumber: "9111111111",
                    FirstName:   "Concurrent",
                    LastName:    fmt.Sprintf("Test-%d", index),
                    CountryCode: "+91",
                },
            }
            req.UserID = fmt.Sprintf("admin-%d", index)

            resp, err := service.CreateFarmer(context.Background(), req)
            results[index] = resp
            errors[index] = err

            if err == nil {
                successCount++
            }
        }(i)
    }
    wg.Wait()

    // ASSERT
    assert.GreaterOrEqual(t, successCount, 1, "At least one request should succeed")

    // CRITICAL: Verify role assigned exactly ONCE
    roleAssignCalls := mockAAA.Calls("AssignRole", mock.Anything, "user-concurrent-123", "org-concurrent", "farmer")
    assert.Equal(t, 1, len(roleAssignCalls), "Role should be assigned exactly once")

    // Verify all successful responses have same user ID
    for i, resp := range results {
        if errors[i] == nil {
            assert.Equal(t, "user-concurrent-123", resp.Data.AAAUserID)
        }
    }
}
```

---

## Test Suite 2: Bulk Farmer Import - Unit Tests

### File: `internal/services/bulk_farmer_service_test.go`

#### Test Group 2.1: Role Assignment in Pipeline

```go
// Test 7: Bulk import assigns role to all farmers
func TestBulkImport_AssignsRolesToAllFarmers(t *testing.T) {
    // ARRANGE
    mockBulkRepo := new(MockBulkOperationRepository)
    mockProcRepo := new(MockProcessingDetailRepository)
    mockFarmerSvc := new(MockFarmerService)
    mockLinkageSvc := new(MockFarmerLinkageService)
    mockAAA := new(MockAAAService)
    logger := newTestLogger()

    service := services.NewBulkFarmerService(
        mockBulkRepo, mockProcRepo, mockFarmerSvc, mockLinkageSvc, mockAAA, logger,
    )

    // Prepare bulk data (5 farmers)
    farmers := []*requests.FarmerBulkData{
        {FirstName: "Farmer1", LastName: "Test", PhoneNumber: "9000000001"},
        {FirstName: "Farmer2", LastName: "Test", PhoneNumber: "9000000002"},
        {FirstName: "Farmer3", LastName: "Test", PhoneNumber: "9000000003"},
        {FirstName: "Farmer4", LastName: "Test", PhoneNumber: "9000000004"},
        {FirstName: "Farmer5", LastName: "Test", PhoneNumber: "9000000005"},
    }

    // Mock AAA responses for each farmer
    for i, farmer := range farmers {
        userID := fmt.Sprintf("user-bulk-%d", i+1)

        // Mock user creation
        mockAAA.On("CreateUser", mock.Anything, mock.MatchedBy(func(req interface{}) bool {
            reqMap := req.(map[string]interface{})
            return reqMap["phone_number"] == farmer.PhoneNumber
        })).Return(map[string]interface{}{"id": userID}, nil)

        // Mock role assignment - CRITICAL TEST
        mockAAA.On("AssignRole", mock.Anything, userID, "org-bulk", "farmer").
            Return(nil).Once()
    }

    // Mock repository operations
    mockBulkRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockBulkRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
    mockBulkRepo.On("UpdateProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    mockProcRepo.On("CreateBatch", mock.Anything, mock.Anything).Return(nil)
    mockProcRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

    // Mock farmer service (creates local profile)
    mockFarmerSvc.On("CreateFarmer", mock.Anything, mock.Anything).
        Return(&responses.FarmerResponse{
            Data: &responses.FarmerProfileData{ID: "farmer-123"},
        }, nil)

    // Mock linkage service
    mockLinkageSvc.On("LinkFarmerToFPO", mock.Anything, mock.Anything).Return(nil)

    // ACT
    req := &requests.BulkFarmerAdditionRequest{
        FPOOrgID:       "org-bulk",
        UserID:         "admin-bulk",
        ProcessingMode: "sync",
        InputFormat:    "json",
        Data:           marshalFarmersToJSON(farmers),
    }

    result, err := service.BulkAddFarmersToFPO(context.Background(), req)

    // Wait for async processing to complete
    time.Sleep(2 * time.Second)

    // ASSERT
    assert.NoError(t, err, "Bulk import should succeed")
    assert.NotNil(t, result)

    // CRITICAL: Verify role assigned to ALL farmers
    for i := 1; i <= 5; i++ {
        userID := fmt.Sprintf("user-bulk-%d", i)
        mockAAA.AssertCalled(t, "AssignRole", mock.Anything, userID, "org-bulk", "farmer")
    }

    // Verify total role assignment calls
    roleAssignCalls := mockAAA.FilteredCalls("AssignRole")
    assert.Equal(t, 5, len(roleAssignCalls), "Should have 5 role assignments")
}

// Test 8: Partial role assignment failures - marks records as failed
func TestBulkImport_PartialRoleFailures_MarksRecordsFailed(t *testing.T) {
    // ARRANGE
    // ... setup similar to Test 7 ...

    // Mock role assignment: first 3 succeed, last 2 fail
    mockAAA.On("AssignRole", mock.Anything, "user-bulk-1", "org-bulk", "farmer").Return(nil)
    mockAAA.On("AssignRole", mock.Anything, "user-bulk-2", "org-bulk", "farmer").Return(nil)
    mockAAA.On("AssignRole", mock.Anything, "user-bulk-3", "org-bulk", "farmer").Return(nil)
    mockAAA.On("AssignRole", mock.Anything, "user-bulk-4", "org-bulk", "farmer").
        Return(fmt.Errorf("AAA service unavailable"))
    mockAAA.On("AssignRole", mock.Anything, "user-bulk-5", "org-bulk", "farmer").
        Return(fmt.Errorf("AAA service unavailable"))

    // ACT
    result, err := service.BulkAddFarmersToFPO(context.Background(), req)
    time.Sleep(2 * time.Second)

    // ASSERT
    assert.NoError(t, err, "Bulk operation should be initiated")

    // Verify operation status
    status, _ := service.GetBulkOperationStatus(context.Background(), result.OperationID)
    assert.Equal(t, 5, status.Progress.Processed)
    assert.Equal(t, 3, status.Progress.Successful)
    assert.Equal(t, 2, status.Progress.Failed)

    // Verify failed records can be retried
    assert.True(t, status.CanRetry)
}

// Test 9: All role assignments fail - entire bulk operation fails
func TestBulkImport_AllRolesFail_BulkOperationFails(t *testing.T) {
    // Test scenario where AAA service is completely down
    // All role assignments fail
    // Bulk operation status should be FAILED
    // All processing details should be FAILED
    // ... implementation ...
}
```

---

#### Test Group 2.2: Bulk Retry with Role Assignment

```go
// Test 10: Retry failed records - assigns missing roles
func TestBulkRetry_AssignsMissingRoles(t *testing.T) {
    // ARRANGE
    // Initial bulk import: 3 farmers created without roles (due to bug)
    // Retry operation should assign roles to those 3 farmers

    // ... setup ...

    // Mock processing repo to return 3 failed records
    mockProcRepo.On("GetRetryableRecords", mock.Anything, "bulk-op-original").
        Return([]*bulk.ProcessingDetail{
            {RecordIndex: 0, InputData: map[string]interface{}{/* farmer 1 data */}},
            {RecordIndex: 1, InputData: map[string]interface{}{/* farmer 2 data */}},
            {RecordIndex: 2, InputData: map[string]interface{}{/* farmer 3 data */}},
        }, nil)

    // Mock role assignments on retry
    mockAAA.On("AssignRole", mock.Anything, mock.Anything, mock.Anything, "farmer").
        Return(nil).Times(3)

    // ACT
    retryReq := &requests.RetryBulkOperationRequest{
        OperationID: "bulk-op-original",
    }
    retryResult, err := service.RetryFailedRecords(context.Background(), retryReq)

    time.Sleep(2 * time.Second)

    // ASSERT
    assert.NoError(t, err)

    // CRITICAL: Verify 3 role assignments
    roleAssignCalls := mockAAA.FilteredCalls("AssignRole")
    assert.Equal(t, 3, len(roleAssignCalls))

    // Verify retry operation succeeded
    retryStatus, _ := service.GetBulkOperationStatus(context.Background(), retryResult.OperationID)
    assert.Equal(t, 3, retryStatus.Progress.Successful)
    assert.Equal(t, 0, retryStatus.Progress.Failed)
}
```

---

## Test Suite 3: FPO CEO Role Assignment - Unit Tests

### File: `internal/services/fpo_ref_service_test.go`

#### Test Group 3.1: CEO Role Assignment

```go
// Test 11: FPO creation assigns CEO role
func TestCreateFPO_AssignsCEORole(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFPORefRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFPOService(mockRepo, mockAAA)

    // Mock CEO user creation
    mockAAA.On("GetUserByMobile", mock.Anything, "9999999999").
        Return(nil, fmt.Errorf("user not found"))
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "ceo-user-123"}, nil)

    // Mock CEO role check (not CEO of other org)
    mockAAA.On("CheckUserRole", mock.Anything, "ceo-user-123", "CEO").
        Return(false, nil)

    // Mock organization creation
    mockAAA.On("CreateOrganization", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"org_id": "fpo-org-456"}, nil)

    // Mock CEO role assignment - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "ceo-user-123", "fpo-org-456", "CEO").
        Return(nil)

    // Mock user group creation
    mockAAA.On("CreateUserGroup", mock.Anything, mock.Anything).
        Return(map[string]interface{}{
            "group_id": "group-123",
            "name": "directors",
            "org_id": "fpo-org-456",
            "created_at": time.Now().Format(time.RFC3339),
        }, nil)
    mockAAA.On("AssignPermissionToGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockRepo.On("FindOne", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("not found"))

    // ACT
    req := &requests.CreateFPORequest{
        Name:           "Test FPO",
        RegistrationNo: "FPO-12345",
        CEOUser: requests.CEOUserData{
            FirstName:   "CEO",
            LastName:    "Test",
            PhoneNumber: "9999999999",
            Email:       "ceo@test.com",
            Password:    "secure123",
        },
    }

    result, err := service.CreateFPO(context.Background(), req)

    // ASSERT
    assert.NoError(t, err)
    assert.NotNil(t, result)

    // CRITICAL: Verify CEO role was assigned
    mockAAA.AssertCalled(t, "AssignRole", mock.Anything, "ceo-user-123", "fpo-org-456", "CEO")
    mockAAA.AssertNumberOfCalls(t, "AssignRole", 1)

    // Verify FPO status is ACTIVE
    resultData := result.(*responses.CreateFPOData)
    assert.Equal(t, "ACTIVE", resultData.Status)
}

// Test 12: CEO role assignment fails - FPO marked PENDING_SETUP
func TestCreateFPO_CEORoleFails_MarksPendingSetup(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFPORefRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFPOService(mockRepo, mockAAA)

    // Mock successful user and org creation
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "ceo-user-fail-123"}, nil)
    mockAAA.On("CheckUserRole", mock.Anything, "ceo-user-fail-123", "CEO").
        Return(false, nil)
    mockAAA.On("CreateOrganization", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"org_id": "fpo-org-fail-456"}, nil)

    // Mock CEO role assignment FAILS - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "ceo-user-fail-123", "fpo-org-fail-456", "CEO").
        Return(fmt.Errorf("AAA service error: insufficient permissions"))

    // Mock user groups and permissions (these should still proceed)
    mockAAA.On("CreateUserGroup", mock.Anything, mock.Anything).
        Return(map[string]interface{}{
            "group_id": "group-fail-123",
            "name": "directors",
            "org_id": "fpo-org-fail-456",
            "created_at": time.Now().Format(time.RFC3339),
        }, nil)
    mockAAA.On("AssignPermissionToGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    mockRepo.On("Create", mock.Anything, mock.Anything).
        Run(func(args mock.Arguments) {
            fpoRef := args.Get(1).(*fpo.FPORef)
            // Verify status is PENDING_SETUP
            assert.Equal(t, fpo.FPOStatusPendingSetup, fpoRef.Status)
            // Verify setup errors contain CEO role failure
            assert.Contains(t, fpoRef.SetupErrors, "ceo_role_assignment")
        }).
        Return(nil)

    // ACT
    req := &requests.CreateFPORequest{
        Name:           "Test FPO Fail",
        RegistrationNo: "FPO-FAIL-123",
        CEOUser: requests.CEOUserData{
            FirstName:   "CEO",
            LastName:    "Fail",
            PhoneNumber: "9888888888",
            Email:       "ceo.fail@test.com",
            Password:    "secure123",
        },
    }

    result, err := service.CreateFPO(context.Background(), req)

    // ASSERT
    // IMPORTANT: Should NOT return error (FPO creation proceeds with PENDING status)
    assert.NoError(t, err, "FPO creation should succeed despite role failure")
    assert.NotNil(t, result)

    // CRITICAL: Verify role assignment was attempted
    mockAAA.AssertCalled(t, "AssignRole", mock.Anything, "ceo-user-fail-123", "fpo-org-fail-456", "CEO")

    // Verify FPO status is PENDING_SETUP
    resultData := result.(*responses.CreateFPOData)
    assert.Equal(t, "PENDING_SETUP", resultData.Status)

    // Verify repository was called with PENDING_SETUP status
    mockRepo.AssertExpectations(t)
}

// Test 13: CompleteFPOSetup retries CEO role assignment
func TestCompleteFPOSetup_RetriesCEORole(t *testing.T) {
    // ARRANGE
    mockRepo := new(MockFPORefRepository)
    mockAAA := new(MockAAAService)
    service := services.NewFPOService(mockRepo, mockAAA)

    // Mock FPO in PENDING_SETUP status with CEO role error
    mockRepo.On("FindOne", mock.Anything, mock.Anything).
        Return(&fpo.FPORef{
            ID:             "fpo-pending-123",
            AAAOrgID:       "org-pending-456",
            Name:           "Pending FPO",
            RegistrationNo: "FPO-PENDING-123",
            Status:         fpo.FPOStatusPendingSetup,
            SetupErrors: map[string]interface{}{
                "ceo_role_assignment": "AAA service error: insufficient permissions",
            },
        }, nil)

    // Mock get organization (to retrieve CEO user ID)
    mockAAA.On("GetOrganization", mock.Anything, "org-pending-456").
        Return(map[string]interface{}{
            "org_id":      "org-pending-456",
            "ceo_user_id": "ceo-pending-789",
        }, nil)

    // Mock CEO role assignment SUCCEEDS on retry - CRITICAL TEST
    mockAAA.On("AssignRole", mock.Anything, "ceo-pending-789", "org-pending-456", "CEO").
        Return(nil)

    // Mock user group creation (for other setup errors)
    mockAAA.On("CreateUserGroup", mock.Anything, mock.Anything).
        Return(map[string]interface{}{
            "group_id":   "group-retry-123",
            "name":       "directors",
            "org_id":     "org-pending-456",
            "created_at": time.Now().Format(time.RFC3339),
        }, nil)
    mockAAA.On("AssignPermissionToGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    mockRepo.On("Create", mock.Anything, mock.Anything).
        Run(func(args mock.Arguments) {
            fpoRef := args.Get(1).(*fpo.FPORef)
            // Verify status changed to ACTIVE
            assert.Equal(t, fpo.FPOStatusActive, fpoRef.Status)
            // Verify CEO role error removed from setup errors
            assert.NotContains(t, fpoRef.SetupErrors, "ceo_role_assignment")
        }).
        Return(nil)

    // ACT
    result, err := service.CompleteFPOSetup(context.Background(), "org-pending-456")

    // ASSERT
    assert.NoError(t, err)
    assert.NotNil(t, result)

    // CRITICAL: Verify CEO role assignment was retried
    mockAAA.AssertCalled(t, "AssignRole", mock.Anything, "ceo-pending-789", "org-pending-456", "CEO")

    // Verify FPO status is now ACTIVE
    resultData := result.(*responses.FPORefData)
    assert.Equal(t, "ACTIVE", resultData.Status)

    mockRepo.AssertExpectations(t)
}

// Test 14: User already CEO of another FPO - creation rejected
func TestCreateFPO_UserAlreadyCEO_Rejected(t *testing.T) {
    // Test that CheckUserRole detects existing CEO role
    // CreateFPO should return error
    // No organization should be created
    // ... implementation ...
}
```

---

## Test Suite 4: Integration Tests

### File: `internal/services/integration_role_test.go`

These tests use a real AAA service test instance.

```go
// +build integration

package services_test

import (
    "context"
    "testing"

    "github.com/Kisanlink/farmers-module/internal/config"
    "github.com/Kisanlink/farmers-module/internal/services"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type RoleIntegrationTestSuite struct {
    suite.Suite
    aaaService    services.AAAService
    farmerService services.FarmerService
    fpoService    services.FPOService
    testOrgID     string
    testUserIDs   []string // Track created users for cleanup
}

func (suite *RoleIntegrationTestSuite) SetupSuite() {
    // Initialize real AAA service client (pointing to test instance)
    cfg := &config.Config{
        AAA: config.AAAConfig{
            Enabled:     true,
            GRPCAddress: "localhost:50051", // Test AAA instance
        },
    }

    suite.aaaService = services.NewAAAService(cfg)

    // Create test organization
    orgResp, err := suite.aaaService.CreateOrganization(context.Background(), map[string]interface{}{
        "name":        "Integration Test FPO",
        "description": "Test organization for role integration tests",
        "type":        "FPO",
    })
    suite.NoError(err)

    orgMap := orgResp.(map[string]interface{})
    suite.testOrgID = orgMap["org_id"].(string)
}

func (suite *RoleIntegrationTestSuite) TearDownSuite() {
    // Cleanup: Delete all created users
    for _, userID := range suite.testUserIDs {
        suite.aaaService.DeleteUser(context.Background(), userID)
    }

    // Delete test organization
    // ... cleanup code ...
}

func (suite *RoleIntegrationTestSuite) TestFarmerCreation_RoleInAAA() {
    t := suite.T()

    // ARRANGE
    // ... setup farmer service with real AAA client ...

    // ACT
    req := &requests.CreateFarmerRequest{
        AAAOrgID: suite.testOrgID,
        Profile: requests.FarmerProfileData{
            FirstName:   "Integration",
            LastName:    "Test",
            PhoneNumber: "9123456780",
            CountryCode: "+91",
        },
    }
    req.UserID = "admin-integration"

    response, err := suite.farmerService.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.NoError(t, err)
    assert.NotNil(t, response)

    suite.testUserIDs = append(suite.testUserIDs, response.Data.AAAUserID)

    // CRITICAL: Verify role exists in AAA service
    hasRole, err := suite.aaaService.CheckUserRole(
        context.Background(),
        response.Data.AAAUserID,
        "farmer",
    )
    assert.NoError(t, err)
    assert.True(t, hasRole, "Farmer role should exist in AAA service")

    // Verify user can authenticate and has farmer permissions
    // ... additional verification ...
}

func (suite *RoleIntegrationTestSuite) TestBulkImport_AllFarmersHaveRoles() {
    // Create 50 farmers via bulk import
    // Verify all 50 have farmer role in AAA
    // ... implementation ...
}

func (suite *RoleIntegrationTestSuite) TestFPOCreation_CEOHasRole() {
    // Create FPO via API
    // Verify CEO user has CEO role in AAA service
    // ... implementation ...
}

func TestRoleIntegrationSuite(t *testing.T) {
    suite.Run(t, new(RoleIntegrationTestSuite))
}
```

---

## Test Suite 5: Invariant Validation Tests

### File: `internal/services/invariant_test.go`

```go
// Test 15: Invariant - All farmers have farmer role
func TestInvariant_AllFarmersHaveRole(t *testing.T) {
    // ARRANGE
    ctx := context.Background()

    // Create 50 farmers through various paths
    // - 20 via single CreateFarmer
    // - 20 via bulk import
    // - 10 via existing AAA user

    // ACT
    // Query all farmers from database
    farmers, err := farmerRepo.Find(ctx, base.NewFilterBuilder().Build())
    assert.NoError(t, err)

    violationCount := 0
    var violations []string

    for _, farmer := range farmers {
        // Check if farmer has role in AAA
        hasRole, err := aaaService.CheckUserRole(ctx, farmer.AAAUserID, "farmer")
        if err != nil || !hasRole {
            violationCount++
            violations = append(violations, fmt.Sprintf(
                "Farmer %s (user %s) lacks farmer role",
                farmer.ID, farmer.AAAUserID,
            ))
        }
    }

    // ASSERT
    assert.Equal(t, 0, violationCount, "No farmers should lack farmer role")
    if violationCount > 0 {
        t.Logf("Invariant violations found:\n%s", strings.Join(violations, "\n"))
    }
}

// Test 16: Invariant - All FPO CEOs have CEO role
func TestInvariant_AllCEOsHaveRole(t *testing.T) {
    // Similar to Test 15, but for FPO CEOs
    // ... implementation ...
}

// Test 17: Invariant - No orphaned roles
func TestInvariant_NoOrphanedRoles(t *testing.T) {
    // Query AAA for all users with farmer role
    // Verify each has corresponding farmer profile in farmers-module
    // ... implementation ...
}
```

---

## Test Suite 6: Concurrency and Race Condition Tests

### File: `internal/services/concurrency_test.go`

```go
// Test 18: Concurrent farmer creation with same phone
func TestConcurrency_FarmerCreation_SamePhone(t *testing.T) {
    // Already covered in Test 6
}

// Test 19: Concurrent FPO creation with same CEO
func TestConcurrency_FPOCreation_SameCEO_OnlyOneSucceeds(t *testing.T) {
    t.Parallel()

    // ARRANGE
    numGoroutines := 5
    successCount := 0
    results := make([]interface{}, numGoroutines)
    errors := make([]error, numGoroutines)

    mockAAA := new(ThreadSafeMockAAAService)
    mockRepo := new(ThreadSafeMockFPORefRepository)
    service := services.NewFPOService(mockRepo, mockAAA)

    // Configure mock to allow CEO role assignment only ONCE
    ceoRoleAssigned := false
    mockAAA.On("AssignRole", mock.Anything, "ceo-concurrent-123", mock.Anything, "CEO").
        Return(func(ctx context.Context, userID, orgID, role string) error {
            if !ceoRoleAssigned {
                ceoRoleAssigned = true
                return nil
            }
            return fmt.Errorf("user already has CEO role in another organization")
        })

    // ... other mocks ...

    // ACT
    var wg sync.WaitGroup
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            req := &requests.CreateFPORequest{
                Name:           fmt.Sprintf("FPO-%d", index),
                RegistrationNo: fmt.Sprintf("REG-%d", index),
                CEOUser: requests.CEOUserData{
                    FirstName:   "CEO",
                    LastName:    "Concurrent",
                    PhoneNumber: "9777777777", // Same CEO
                    Email:       "ceo@concurrent.com",
                    Password:    "secure123",
                },
            }

            result, err := service.CreateFPO(context.Background(), req)
            results[index] = result
            errors[index] = err

            if err == nil {
                successCount++
            }
        }(i)
    }
    wg.Wait()

    // ASSERT
    assert.Equal(t, 1, successCount, "Exactly one FPO creation should succeed")

    // Verify only one CEO role assignment succeeded
    roleAssignCalls := mockAAA.FilteredCalls("AssignRole", "CEO")
    assert.Equal(t, 1, len(roleAssignCalls), "CEO role should be assigned exactly once")

    // Verify other 4 requests failed with appropriate error
    failureCount := 0
    for _, err := range errors {
        if err != nil {
            assert.Contains(t, err.Error(), "already has CEO role")
            failureCount++
        }
    }
    assert.Equal(t, 4, failureCount)
}

// Test 20: Concurrent bulk imports with overlapping data
func TestConcurrency_BulkImport_OverlappingFarmers(t *testing.T) {
    // Two bulk imports initiated simultaneously
    // Both contain same 5 farmers
    // Verify only 5 farmers created (not 10)
    // Verify exactly 5 farmer role assignments
    // ... implementation ...
}
```

---

## Test Suite 7: End-to-End Tests

### File: `e2e/role_assignment_e2e_test.go`

```go
// +build e2e

package e2e_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestE2E_FarmerCreation_FullFlow(t *testing.T) {
    // ARRANGE
    baseURL := "http://localhost:8000" // Test instance

    // ACT
    // 1. Create farmer via API
    reqBody := map[string]interface{}{
        "aaa_org_id": "test-org-123",
        "profile": map[string]interface{}{
            "first_name":   "E2E",
            "last_name":    "Test",
            "phone_number": "9555555555",
            "country_code": "+91",
            "email":        "e2e@test.com",
        },
    }

    reqJSON, _ := json.Marshal(reqBody)
    resp, err := http.Post(
        baseURL+"/api/v1/farmers",
        "application/json",
        bytes.NewBuffer(reqJSON),
    )

    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var createResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&createResp)
    resp.Body.Close()

    farmerData := createResp["data"].(map[string]interface{})
    farmerID := farmerData["id"].(string)
    aaaUserID := farmerData["aaa_user_id"].(string)

    // 2. Verify farmer can authenticate
    loginResp, err := http.Post(
        baseURL+"/api/v1/auth/login",
        "application/json",
        bytes.NewBuffer([]byte(`{"phone_number":"9555555555","password":"<default>"}`)),
    )
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, loginResp.StatusCode)

    var loginData map[string]interface{}
    json.NewDecoder(loginResp.Body).Decode(&loginData)
    loginResp.Body.Close()

    token := loginData["token"].(string)
    assert.NotEmpty(t, token)

    // 3. Verify farmer can access farmer endpoints
    req, _ := http.NewRequest("GET", baseURL+"/api/v1/farmers/"+farmerID, nil)
    req.Header.Set("Authorization", "Bearer "+token)
    getResp, err := http.DefaultClient.Do(req)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, getResp.StatusCode)

    // 4. Verify farmer has correct permissions
    req, _ = http.NewRequest("POST", baseURL+"/api/v1/farms", bytes.NewBuffer([]byte(`{
        "farmer_id": "`+farmerID+`",
        "name": "Test Farm",
        "area_ha": 2.5
    }`)))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    createFarmResp, err := http.DefaultClient.Do(req)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, createFarmResp.StatusCode)

    // CRITICAL: Verify role in AAA service directly
    checkRoleResp, err := http.Get(
        baseURL + "/api/v1/admin/check-role?user_id=" + aaaUserID + "&role=farmer",
    )
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, checkRoleResp.StatusCode)

    var roleCheckData map[string]interface{}
    json.NewDecoder(checkRoleResp.Body).Decode(&roleCheckData)
    checkRoleResp.Body.Close()

    assert.True(t, roleCheckData["has_role"].(bool), "Farmer should have farmer role")
}

func TestE2E_BulkImport_FullFlow(t *testing.T) {
    // Upload CSV with 100 farmers
    // Poll bulk operation status until complete
    // Verify all 100 farmers created
    // Verify all 100 have farmer role
    // ... implementation ...
}
```

---

## Test Execution Strategy

### Phase 1: Critical Path Tests (Day 1)
Run these tests FIRST to validate core functionality:
- Test 1: TestCreateFarmer_AssignsRoleSuccessfully
- Test 2: TestCreateFarmer_RoleAssignmentFails_RollsBack
- Test 7: TestBulkImport_AssignsRolesToAllFarmers
- Test 11: TestCreateFPO_AssignsCEORole

**Acceptance:** All 4 tests must pass before proceeding

---

### Phase 2: Error Handling Tests (Day 2)
- Test 3: TestCreateFarmer_ExistingUserWithoutRole_AssignsRole
- Test 8: TestBulkImport_PartialRoleFailures_MarksRecordsFailed
- Test 12: TestCreateFPO_CEORoleFails_MarksPendingSetup
- Test 13: TestCompleteFPOSetup_RetriesCEORole

**Acceptance:** >90% pass rate, failures analyzed

---

### Phase 3: Concurrency Tests (Day 3)
- Test 6: TestCreateFarmer_Concurrent_SamePhone_OneSucceedsWithRole
- Test 19: TestConcurrency_FPOCreation_SameCEO_OnlyOneSucceeds
- Test 20: TestConcurrency_BulkImport_OverlappingFarmers

**Acceptance:** No race conditions, correct handling

---

### Phase 4: Integration Tests (Day 4)
- TestFarmerCreation_RoleInAAA
- TestBulkImport_AllFarmersHaveRoles
- TestFPOCreation_CEOHasRole

**Acceptance:** All tests pass with real AAA service

---

### Phase 5: Invariant & E2E Tests (Day 5)
- Test 15: TestInvariant_AllFarmersHaveRole
- Test 16: TestInvariant_AllCEOsHaveRole
- TestE2E_FarmerCreation_FullFlow
- TestE2E_BulkImport_FullFlow

**Acceptance:** Zero invariant violations, E2E flows work

---

## Continuous Integration Setup

### Test Pipeline Configuration

```yaml
# .github/workflows/role-assignment-tests.yml
name: Role Assignment Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.24
      - name: Run Unit Tests
        run: |
          go test ./internal/services -v -coverprofile=coverage.out
          go tool cover -func=coverage.out
      - name: Check Coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 90" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 90%"
            exit 1
          fi

  integration-tests:
    runs-on: ubuntu-latest
    services:
      aaa-service:
        image: kisanlink/aaa-service:test
        ports:
          - 50051:50051
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.24
      - name: Run Integration Tests
        run: go test -tags=integration ./internal/services -v

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Test Environment
        run: docker-compose -f docker-compose.test.yml up -d
      - name: Run E2E Tests
        run: go test -tags=e2e ./e2e -v
      - name: Teardown
        run: docker-compose -f docker-compose.test.yml down
```

---

## Test Metrics and Reporting

### Coverage Goals

| Component | Target Coverage | Critical Paths |
|-----------|----------------|----------------|
| farmer_service.go | 95% | CreateFarmer: 100% |
| bulk_farmer_service.go | 90% | Pipeline execution: 100% |
| fpo_ref_service.go | 95% | CreateFPO, CompleteFPOSetup: 100% |
| pipeline/stages.go | 85% | RoleAssignmentStage: 100% |

### Test Execution Time Budget

- Unit Tests: <10 seconds
- Integration Tests: <60 seconds
- E2E Tests: <180 seconds
- Total CI Pipeline: <300 seconds (5 minutes)

---

## Next Steps After Testing

1. **Automated Invariant Checks**: Deploy daily reconciliation job
2. **Production Monitoring**: Set up alerts for role assignment failures
3. **Regression Suite**: Add these tests to CI/CD pipeline
4. **Load Testing**: Test role assignment under high concurrency (1000 farmers/min)
5. **Security Audit**: Penetration testing of role assignment flows

---

**End of Test Plan**
