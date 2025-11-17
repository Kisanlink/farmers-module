# Testing Plan: AAA Role Assignment Implementation

**Date:** 2025-10-16
**Related:** [AAA Role Assignment Review](./.kiro/specs/aaa-role-assignment-review.md)
**Related:** [ADR-001 Role Assignment Strategy](./.kiro/specs/adr-role-assignment-strategy.md)

---

## Executive Summary

This document provides a comprehensive testing strategy for the AAA role assignment implementation across farmer, FPO, and KisanSathi creation workflows. The testing plan covers unit tests, integration tests, contract tests, security tests, and operational validations.

---

## 1. Unit Tests

### 1.1 Farmer Role Assignment (`farmer_service_test.go`)

#### Test: `TestEnsureFarmerRole_Success`
**Purpose:** Verify successful role assignment with idempotency

```go
func TestEnsureFarmerRole_Success(t *testing.T) {
    // Setup
    mockAAA := &MockAAAService{}
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once()
    mockAAA.On("AssignRole", mock.Anything, "user-123", "org-456", constants.RoleFarmer).
        Return(nil).Once()
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(true, nil).Once()

    service := &FarmerServiceImpl{aaaService: mockAAA}

    // Execute
    err := service.ensureFarmerRole(context.Background(), "user-123", "org-456")

    // Assert
    assert.NoError(t, err)
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** No error, all AAA calls made in correct sequence

#### Test: `TestEnsureFarmerRole_AlreadyExists`
**Purpose:** Verify idempotency when role already assigned

```go
func TestEnsureFarmerRole_AlreadyExists(t *testing.T) {
    // Setup
    mockAAA := &MockAAAService{}
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(true, nil).Once()

    service := &FarmerServiceImpl{aaaService: mockAAA}

    // Execute
    err := service.ensureFarmerRole(context.Background(), "user-123", "org-456")

    // Assert
    assert.NoError(t, err)
    mockAAA.AssertExpectations(t)
    // Verify AssignRole was NOT called
    mockAAA.AssertNotCalled(t, "AssignRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
```

**Expected Result:** No error, AssignRole not called (idempotent)

#### Test: `TestEnsureFarmerRole_AssignmentFailure`
**Purpose:** Verify error handling when role assignment fails

```go
func TestEnsureFarmerRole_AssignmentFailure(t *testing.T) {
    // Setup
    mockAAA := &MockAAAService{}
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once()
    mockAAA.On("AssignRole", mock.Anything, "user-123", "org-456", constants.RoleFarmer).
        Return(errors.New("AAA service unavailable")).Once()

    service := &FarmerServiceImpl{aaaService: mockAAA}

    // Execute
    err := service.ensureFarmerRole(context.Background(), "user-123", "org-456")

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to assign farmer role")
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** Error returned, verification not attempted

#### Test: `TestEnsureFarmerRole_VerificationFailure`
**Purpose:** Verify error handling when verification fails

```go
func TestEnsureFarmerRole_VerificationFailure(t *testing.T) {
    // Setup
    mockAAA := &MockAAAService{}
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once()
    mockAAA.On("AssignRole", mock.Anything, "user-123", "org-456", constants.RoleFarmer).
        Return(nil).Once()
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once() // Role not present after assignment

    service := &FarmerServiceImpl{aaaService: mockAAA}

    // Execute
    err := service.ensureFarmerRole(context.Background(), "user-123", "org-456")

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "verification failed")
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** Error returned indicating verification failure

#### Test: `TestEnsureFarmerRoleWithRetry_SuccessOnSecondAttempt`
**Purpose:** Verify retry logic handles transient failures

```go
func TestEnsureFarmerRoleWithRetry_SuccessOnSecondAttempt(t *testing.T) {
    // Setup
    mockAAA := &MockAAAService{}
    service := &FarmerServiceImpl{aaaService: mockAAA}

    // First attempt fails
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once()
    mockAAA.On("AssignRole", mock.Anything, "user-123", "org-456", constants.RoleFarmer).
        Return(errors.New("transient error")).Once()

    // Second attempt succeeds
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Once()
    mockAAA.On("AssignRole", mock.Anything, "user-123", "org-456", constants.RoleFarmer).
        Return(nil).Once()
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(true, nil).Once()

    // Execute
    err := service.ensureFarmerRoleWithRetry(context.Background(), "user-123", "org-456")

    // Assert
    assert.NoError(t, err)
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** No error, role assigned on second attempt

#### Test: `TestCreateFarmer_RoleAssignmentPendingMetadata`
**Purpose:** Verify metadata stored when role assignment fails

```go
func TestCreateFarmer_RoleAssignmentPendingMetadata(t *testing.T) {
    // Setup
    mockRepo := &MockFarmerRepository{}
    mockAAA := &MockAAAService{}
    service := &FarmerServiceImpl{
        repository: mockRepo,
        aaaService: mockAAA,
        defaultPassword: "test123",
    }

    // User creation succeeds
    mockAAA.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "user-123"}, nil).Once()

    // Repository create succeeds
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(f *farmerentity.Farmer) bool {
        return f.AAAUserID == "user-123"
    })).Return(nil).Once()

    // Role assignment fails
    mockAAA.On("CheckUserRole", mock.Anything, "user-123", constants.RoleFarmer).
        Return(false, nil).Times(2)
    mockAAA.On("AssignRole", mock.Anything, "user-123", mock.Anything, constants.RoleFarmer).
        Return(errors.New("AAA unavailable")).Times(2)

    // Repository update with metadata
    mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *farmerentity.Farmer) bool {
        return f.Metadata["role_assignment_pending"] == "true" &&
               f.Metadata["role_assignment_error"] != nil
    })).Return(nil).Once()

    req := &requests.CreateFarmerRequest{
        AAAOrgID: "org-456",
        Profile: requests.FarmerProfileData{
            PhoneNumber: "9876543210",
            CountryCode: "+91",
            FirstName: "Test",
            LastName: "Farmer",
        },
    }

    // Execute
    response, err := service.CreateFarmer(context.Background(), req)

    // Assert
    assert.NoError(t, err) // Farmer creation succeeds despite role failure
    assert.NotNil(t, response)
    mockRepo.AssertExpectations(t)
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** Farmer created, metadata stored, no error returned

---

### 1.2 FPO CEO Role Assignment (`fpo_ref_service_test.go`)

#### Test: `TestCreateFPO_CEORoleAssignmentFailure`
**Purpose:** Verify FPO marked PENDING_SETUP if CEO role fails

```go
func TestCreateFPO_CEORoleAssignmentFailure(t *testing.T) {
    // Setup
    mockRepo := &MockFPORefRepository{}
    mockAAA := &MockAAAService{}
    service := &FPOServiceImpl{
        fpoRefRepo: mockRepo,
        aaaService: mockAAA,
    }

    // User and org creation succeed
    mockAAA.On("GetUserByMobile", mock.Anything, "9876543210").
        Return(map[string]interface{}{"id": "ceo-user-123"}, nil).Once()
    mockAAA.On("CheckUserRole", mock.Anything, "ceo-user-123", "CEO").
        Return(false, nil).Once()
    mockAAA.On("CreateOrganization", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"org_id": "org-789"}, nil).Once()

    // CEO role assignment fails
    mockAAA.On("AssignRole", mock.Anything, "ceo-user-123", "org-789", "CEO").
        Return(errors.New("role assignment failed")).Once()

    // User group creation succeeds (simplified)
    mockAAA.On("CreateUserGroup", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"group_id": "group-123"}, nil).Times(4)
    mockAAA.On("AssignPermissionToGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil).Times(10)

    // FPO stored with PENDING_SETUP status
    mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(fpo *fpo.FPORef) bool {
        return fpo.Status == fpo.FPOStatusPendingSetup &&
               fpo.SetupErrors["ceo_role_assignment"] != nil
    })).Return(nil).Once()

    req := &requests.CreateFPORequest{
        Name: "Test FPO",
        RegistrationNo: "FPO123",
        CEOUser: requests.CEOUserData{
            FirstName: "CEO",
            LastName: "User",
            PhoneNumber: "9876543210",
        },
    }

    // Execute
    response, err := service.CreateFPO(context.Background(), req)

    // Assert
    assert.NoError(t, err) // FPO creation succeeds
    assert.NotNil(t, response)
    responseData := response.(*responses.CreateFPOData)
    assert.Equal(t, fpo.FPOStatusPendingSetup.String(), responseData.Status)
    mockRepo.AssertExpectations(t)
    mockAAA.AssertExpectations(t)
}
```

**Expected Result:** FPO created with PENDING_SETUP status, setupErrors populated

---

### 1.3 Role Constants Tests (`constants/roles_test.go`)

#### Test: `TestAllRoles_ReturnsExpectedRoles`
```go
func TestAllRoles_ReturnsExpectedRoles(t *testing.T) {
    roles := constants.AllRoles()

    assert.Len(t, roles, 6)
    assert.Contains(t, roles, constants.RoleFarmer)
    assert.Contains(t, roles, constants.RoleKisanSathi)
    assert.Contains(t, roles, constants.RoleFPOCEO)
    assert.Contains(t, roles, constants.RoleFPOManager)
    assert.Contains(t, roles, constants.RoleAdmin)
    assert.Contains(t, roles, constants.RoleReadOnly)
}
```

#### Test: `TestIsValidRole`
```go
func TestIsValidRole(t *testing.T) {
    tests := []struct {
        name     string
        roleName string
        expected bool
    }{
        {"Valid farmer role", constants.RoleFarmer, true},
        {"Valid CEO role", constants.RoleFPOCEO, true},
        {"Invalid role", "invalid_role", false},
        {"Empty role", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := constants.IsValidRole(tt.roleName)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

---

## 2. Integration Tests

### 2.1 End-to-End Farmer Creation with Role Assignment

**Test:** `TestIntegration_CreateFarmer_WithRoleVerification`

**Setup:**
- Real database connection
- Mock AAA service (or test AAA instance)

**Steps:**
1. Create farmer with phone number
2. Verify farmer entity created in DB
3. Verify AAA user created
4. Verify `farmer` role assigned to AAA user
5. Query AAA to confirm role exists

**Assertions:**
- Farmer record exists in DB with correct AAA user ID
- AAA user has `farmer` role
- No `role_assignment_pending` in metadata

**Cleanup:**
- Delete farmer from DB
- Remove AAA user and role assignment

---

### 2.2 AAA Service Unavailable Scenario

**Test:** `TestIntegration_CreateFarmer_AAAUnavailable`

**Setup:**
- Real database connection
- AAA service mock configured to fail

**Steps:**
1. Attempt to create farmer
2. AAA role assignment fails
3. Farmer creation succeeds

**Assertions:**
- Farmer entity created in DB
- Metadata contains `role_assignment_pending = true`
- Metadata contains error message

**Recovery Test:**
1. Enable AAA service
2. Call role reconciliation endpoint (future)
3. Verify role assigned
4. Verify metadata cleared

---

### 2.3 FPO Creation with CEO Role

**Test:** `TestIntegration_CreateFPO_CEORoleAssigned`

**Setup:**
- Real database and AAA connection

**Steps:**
1. Create FPO with CEO user details
2. Verify FPO organization created in AAA
3. Verify CEO user created
4. Verify `CEO` role assigned to CEO user
5. Verify user groups created
6. Verify FPO status is ACTIVE

**Assertions:**
- FPO record in DB with ACTIVE status
- AAA organization exists
- CEO user has `CEO` role
- All user groups created
- No setupErrors

---

### 2.4 Role Seeding on Startup

**Test:** `TestIntegration_Startup_RoleSeeding`

**Setup:**
- Clean AAA instance (no roles)
- Start application

**Steps:**
1. Start application (call main)
2. Wait for startup logs
3. Query AAA service for roles

**Assertions:**
- Log shows "Seeding AAA roles and permissions..."
- Log shows "Successfully seeded..."
- AAA contains all required roles: farmer, kisansathi, CEO, fpo_manager, admin, readonly

**Alternative: Seeding Failure**
1. Start with AAA unavailable
2. Verify warning logged
3. Verify application continues to start

---

## 3. Contract Tests (AAA Service API)

### 3.1 AssignRole Contract

**Test:** `TestContract_AssignRole_Success`

```go
func TestContract_AssignRole_Success(t *testing.T) {
    client := aaa.NewClient(testConfig)

    // Create test user
    userID := createTestUser(t, client)
    orgID := createTestOrg(t, client)

    // Execute
    err := client.AssignRole(context.Background(), userID, orgID, "farmer")

    // Assert
    assert.NoError(t, err)

    // Verify
    user, err := client.GetUser(context.Background(), userID)
    assert.NoError(t, err)
    assert.Contains(t, user.UserRoles, "farmer")

    // Cleanup
    deleteTestUser(t, client, userID)
    deleteTestOrg(t, client, orgID)
}
```

### 3.2 AssignRole Idempotency

**Test:** `TestContract_AssignRole_Idempotent`

```go
func TestContract_AssignRole_Idempotent(t *testing.T) {
    client := aaa.NewClient(testConfig)
    userID := createTestUser(t, client)
    orgID := createTestOrg(t, client)

    // First assignment
    err1 := client.AssignRole(context.Background(), userID, orgID, "farmer")
    assert.NoError(t, err1)

    // Second assignment (duplicate)
    err2 := client.AssignRole(context.Background(), userID, orgID, "farmer")
    assert.NoError(t, err2) // Should succeed (idempotent)

    // Verify
    user, _ := client.GetUser(context.Background(), userID)
    assert.Len(t, user.UserRoles, 1) // Only one role entry

    // Cleanup
    deleteTestUser(t, client, userID)
    deleteTestOrg(t, client, orgID)
}
```

### 3.3 SeedRolesAndPermissions Contract

**Test:** `TestContract_SeedRoles_CreatesRequiredRoles`

```go
func TestContract_SeedRoles_CreatesRequiredRoles(t *testing.T) {
    client := aaa.NewClient(testConfig)

    // Execute
    err := client.SeedRolesAndPermissions(context.Background())
    assert.NoError(t, err)

    // Verify all required roles exist
    requiredRoles := []string{
        constants.RoleFarmer,
        constants.RoleKisanSathi,
        constants.RoleFPOCEO,
        constants.RoleFPOManager,
        constants.RoleAdmin,
        constants.RoleReadOnly,
    }

    for _, role := range requiredRoles {
        exists := verifyRoleExists(t, client, role)
        assert.True(t, exists, "Role %s should exist after seeding", role)
    }
}
```

---

## 4. Security Tests

### 4.1 Unauthorized Role Assignment

**Test:** `TestSecurity_CannotAssignRoleWithoutPermission`

**Setup:**
- Create farmer user (no admin permissions)
- Attempt to assign role to another user

**Expected:** Permission denied error

---

### 4.2 Role Escalation Prevention

**Test:** `TestSecurity_FarmerCannotAssignCEORole`

**Setup:**
- Authenticate as farmer
- Attempt to create FPO with self as CEO

**Expected:** Proper role assignment validation

---

### 4.3 Cross-Organization Access

**Test:** `TestSecurity_CannotAccessFarmerInDifferentOrg`

**Setup:**
- Create farmer in Org A
- Authenticate as user in Org B
- Attempt to read farmer data

**Expected:** Permission denied (org context enforced)

---

## 5. Operational Tests

### 5.1 Role Assignment Metrics

**Test:** `TestOperational_RoleAssignmentMetrics`

**Steps:**
1. Create 10 farmers
2. Query Prometheus metrics endpoint
3. Verify `farmer_role_assignments_total{status="success"}` = 10

---

### 5.2 Pending Role Assignment Query

**Test:** `TestOperational_QueryPendingAssignments`

**Setup:**
- Create 5 farmers with role assignment failures
- Create 5 farmers with successful assignment

**Steps:**
```sql
SELECT COUNT(*)
FROM farmers
WHERE metadata->>'role_assignment_pending' = 'true'
```

**Expected:** Returns 5

---

### 5.3 Role Reconciliation Endpoint (Future)

**Test:** `TestOperational_ReconcileRoles`

**Setup:**
- Create farmers with pending role assignments
- Enable AAA service

**Steps:**
1. Call `/admin/reconcile-roles` endpoint
2. Verify role assignments retried
3. Verify metadata cleared

---

## 6. Performance Tests

### 6.1 Role Assignment Latency

**Test:** `TestPerformance_RoleAssignmentLatency`

**Benchmark:**
```go
func BenchmarkEnsureFarmerRole(b *testing.B) {
    service := setupFarmerService()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        userID := fmt.Sprintf("user-%d", i)
        _ = service.ensureFarmerRole(ctx, userID, "org-123")
    }
}
```

**Acceptance Criteria:** < 200ms p95 latency

---

### 6.2 Concurrent Role Assignments

**Test:** `TestPerformance_ConcurrentRoleAssignments`

**Steps:**
1. Create 100 farmers concurrently
2. Verify all role assignments succeed
3. No race conditions or deadlocks

---

## 7. Regression Tests

### 7.1 Existing Functionality Not Broken

**Test:** `TestRegression_KisanSathiRoleStillWorks`

**Purpose:** Verify KisanSathi role assignment (already working) not broken by changes

**Steps:**
1. Create KisanSathi user
2. Assign to farmer
3. Verify role assigned

**Expected:** Same behavior as before

---

## 8. Test Execution Plan

### Phase 1: Pre-Deployment (Week 1)
- [ ] Run all unit tests (target: 100% pass rate)
- [ ] Run integration tests against staging AAA
- [ ] Run contract tests against test AAA instance
- [ ] Verify all tests green before merge

### Phase 2: Staging Deployment (Week 2)
- [ ] Deploy to staging environment
- [ ] Run smoke tests (create farmer, FPO, KisanSathi)
- [ ] Verify role seeding logs
- [ ] Test AAA unavailable scenario (stop AAA service temporarily)
- [ ] Verify pending role assignment metadata stored

### Phase 3: Production Deployment (Week 3)
- [ ] Deploy to production (off-peak hours)
- [ ] Monitor role assignment success rate
- [ ] Verify no errors in application logs
- [ ] Check Prometheus metrics for anomalies

### Phase 4: Post-Deployment Validation (Week 3)
- [ ] Query production DB for pending role assignments
- [ ] If any pending: investigate and manually assign
- [ ] Run data migration script for existing entities
- [ ] Verify all existing farmers have farmer role

---

## 9. Test Data Requirements

### Test Users
- **Farmer User**: phone=9876543210, country_code=+91
- **CEO User**: phone=9876543211, country_code=+91
- **KisanSathi User**: phone=9876543212, country_code=+91
- **Admin User**: phone=9876543213, country_code=+91 (for manual operations)

### Test Organizations
- **Test FPO**: name="Test FPO", registration_number="TEST001"
- **Test Org 2**: name="Test FPO 2", registration_number="TEST002" (for cross-org tests)

### Cleanup
- All test data must be cleaned up after tests
- Use transaction rollback for DB tests
- Delete AAA users/orgs after contract tests

---

## 10. Continuous Integration

### CI Pipeline Stages
1. **Lint**: Run golangci-lint (includes pre-commit checks)
2. **Unit Tests**: Run all unit tests with coverage
3. **Contract Tests**: Run against test AAA instance
4. **Integration Tests**: Run against staging database
5. **Build**: Compile application
6. **Deploy to Staging**: Automatic on main branch

### Coverage Requirements
- Unit test coverage: > 80%
- Integration test coverage: > 60%
- Critical paths (role assignment): 100%

### Automated Checks
- [ ] All tests must pass before merge
- [ ] No decrease in code coverage
- [ ] No critical security vulnerabilities (gosec)
- [ ] No high-severity linting errors

---

## 11. Monitoring and Alerting Validation

### Test Alerts Trigger Correctly

**Test:** `TestOperational_HighPendingRoleAssignmentAlert`

**Steps:**
1. Create 150 farmers with role assignment failures (exceed alert threshold of 100)
2. Wait for alert to fire
3. Verify PagerDuty incident created

**Expected:** Alert fires within 5 minutes

---

## 12. Rollback Plan Testing

### Test Rollback Scenario

**Test:** `TestOperational_Rollback`

**Steps:**
1. Deploy new version with role assignment
2. Simulate critical issue (e.g., all role assignments fail)
3. Roll back to previous version
4. Verify farmers can still be created
5. Verify no data corruption

**Expected:** Graceful rollback, no errors

---

## Summary

This testing plan provides comprehensive coverage of the role assignment implementation across all workflows. Execute tests in phases (unit → integration → contract → operational) to ensure system reliability and security. All tests should be automated and run in CI pipeline before production deployment.

**Test Count:**
- Unit Tests: 15+
- Integration Tests: 6+
- Contract Tests: 3+
- Security Tests: 3+
- Operational Tests: 3+
- Performance Tests: 2+
- Regression Tests: 1+

**Total Estimated Effort:**
- Writing tests: 3-4 days
- Running tests: 1 day
- Fixing issues: 2-3 days
- **Total: ~7-8 days**

---

**Document Version:** 1.0
**Last Updated:** 2025-10-16
**Owner:** Backend Architecture Team
**Status:** Ready for Implementation
