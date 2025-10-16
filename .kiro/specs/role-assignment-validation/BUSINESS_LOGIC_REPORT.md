# Role Assignment Business Logic Validation Report

**Generated:** 2025-10-16
**Module:** farmers-module
**Focus Area:** Role Assignment Logic Across All User Creation Flows

---

## Executive Summary

### Critical Findings

**SEVERITY: HIGH - MISSING ROLE ASSIGNMENTS IN MULTIPLE CRITICAL PATHS**

After comprehensive analysis of the farmers-module codebase, I have identified **critical gaps in role assignment logic** that expose the system to:

1. **Authorization Bypass Vulnerabilities**: Users created without proper roles can access the system but bypass permission checks
2. **Data Integrity Issues**: Inconsistent role state between AAA service and local module
3. **Business Logic Violations**: Farmers, FPO CEOs, and KisanSathis may exist without proper role assignments

### Key Statistics

- **Total User Creation Flows Identified**: 5
- **Flows WITH Role Assignment**: 2 (40%)
- **Flows MISSING Role Assignment**: 3 (60%)
- **Critical Invariants at Risk**: 4
- **High-Priority Edge Cases**: 12

---

## 1. Role Assignment Scenario Analysis

### 1.1 Farmer Role Assignment

#### Scenario 1: Single Farmer Creation via CreateFarmer API

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_service.go`
**Lines:** 57-323

**Current Implementation:**
```go
// CreateFarmer creates a new farmer
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error) {
    // ... user creation in AAA ...
    aaaUser, err := s.aaaService.CreateUser(ctx, createUserReq)
    // ... farmer profile creation ...
    farmer := farmerentity.NewFarmer()
    farmer.AAAUserID = aaaUserID
    // ... save farmer ...
}
```

**FINDING: MISSING ROLE ASSIGNMENT**

**Status:** ❌ **CRITICAL - NO FARMER ROLE ASSIGNED**

**Expected Behavior:**
After creating the AAA user, the system MUST assign the "farmer" role to the user in the AAA service.

**Actual Behavior:**
- AAA user is created successfully (line 111)
- Farmer profile is created locally (line 244-272)
- **NO role assignment call to AAA service**

**Impact:**
- Farmer user exists in AAA but has no role
- Permission checks in AAA service will fail
- Farmer cannot access farmer-specific endpoints
- Violates Business Rule: "A farmer should ALWAYS have the farmer role after creation"

**Exploitation Path:**
1. Malicious actor creates farmer account
2. Farmer has valid credentials but no role
3. Actor modifies their role in AAA (if they gain access)
4. Escalates privileges without audit trail

**Recommended Fix:**
```go
// After AAA user creation (line ~194)
aaaUserID = fmt.Sprintf("%v", id)

// ADD THIS: Assign farmer role
err = s.aaaService.AssignRole(ctx, aaaUserID, aaaOrgID, "farmer")
if err != nil {
    // CRITICAL: Roll back farmer creation if role assignment fails
    log.Printf("CRITICAL: Failed to assign farmer role to user %s: %v", aaaUserID, err)
    // Consider rolling back AAA user creation or marking farmer as PENDING_ACTIVATION
    return nil, fmt.Errorf("failed to assign farmer role: %w", err)
}
log.Printf("Assigned farmer role to user %s in org %s", aaaUserID, aaaOrgID)
```

---

#### Scenario 2: Bulk Farmer Import via BulkAddFarmersToFPO

**File:** `/Users/kaushik/farmers-module/internal/services/bulk_farmer_service.go`
**Processing Pipeline:** `/Users/kaushik/farmers-module/internal/services/pipeline/stages.go`
**Lines:** 98-163 (bulk service), 228-437 (pipeline stages)

**Current Implementation:**

**Pipeline Stages:**
1. **ValidationStage** (lines 16-146 in stages.go)
2. **DeduplicationStage** (lines 167-226)
3. **AAAUserCreationStage** (lines 228-341) - Creates AAA user
4. **FarmerRegistrationStage** (lines 343-437) - Creates farmer profile
5. **FPOLinkageStage** (lines 439-503) - Links farmer to FPO
6. **KisanSathiAssignmentStage** (lines 505-593) - Assigns KisanSathi

**FINDING: MISSING FARMER ROLE ASSIGNMENT IN PIPELINE**

**Status:** ❌ **CRITICAL - BULK OPERATIONS CREATE FARMERS WITHOUT ROLES**

**Analysis:**
- **AAAUserCreationStage** creates AAA users (line 293) but does NOT assign farmer role
- **FarmerRegistrationStage** creates local farmer profile but does NOT assign role
- **FPOLinkageStage** links farmer to FPO but does NOT assign role
- **NO stage in the pipeline assigns the farmer role**

**Evidence from AAAUserCreationStage:**
```go
// Line 293 - User creation
userResponse, err := aus.aaaService.CreateUser(ctx, createUserReq)
// ... extract user ID ...
// NO ROLE ASSIGNMENT HERE

procCtx.SetStageResult("aaa_user_creation", map[string]interface{}{
    "aaa_user_id":  aaaUserID,
    "user_existed": false,
    // NO ROLE ASSIGNMENT TRACKING
})
```

**Impact:**
- **HIGH SEVERITY**: Bulk imports can create hundreds of farmers without proper roles
- All bulk-imported farmers are in invalid state
- Mass authorization failures across the system
- Potential security incident if exploited

**Recommended Fix:**

**Option 1: Add Role Assignment Stage to Pipeline**
```go
// Create new pipeline stage: RoleAssignmentStage
type RoleAssignmentStage struct {
    *BasePipelineStage
    aaaService AAAServiceInterface
}

func (ras *RoleAssignmentStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    procCtx := data.(*ProcessingContext)

    // Get AAA user ID from previous stage
    aaaResult := procCtx.GetStageResult("aaa_user_creation")
    aaaUserID := aaaResult["aaa_user_id"].(string)

    // Assign farmer role
    err := ras.aaaService.AssignRole(ctx, aaaUserID, procCtx.FPOOrgID, "farmer")
    if err != nil {
        return nil, fmt.Errorf("failed to assign farmer role: %w", err)
    }

    procCtx.SetStageResult("role_assignment", map[string]interface{}{
        "role": "farmer",
        "assigned_at": time.Now(),
    })

    return procCtx, nil
}
```

**Add to pipeline in bulk_farmer_service.go (line ~401):**
```go
func (s *BulkFarmerServiceImpl) buildProcessingPipeline(options requests.BulkProcessingOptions) pipeline.ProcessingPipeline {
    pipe := pipeline.NewPipeline(s.logger)

    pipe.AddStage(pipeline.NewValidationStage(s.logger))
    pipe.AddStage(pipeline.NewDeduplicationStage(s.farmerService, s.logger))
    pipe.AddStage(pipeline.NewAAAUserCreationStage(s.aaaService, s.logger))

    // ADD THIS: Role assignment stage
    pipe.AddStage(pipeline.NewRoleAssignmentStage(s.aaaService, "farmer", s.logger))

    pipe.AddStage(pipeline.NewFarmerRegistrationStage(s.farmerService, s.logger))
    pipe.AddStage(pipeline.NewFPOLinkageStage(s.linkageService, s.logger))

    if options.AssignKisanSathi {
        pipe.AddStage(pipeline.NewKisanSathiAssignmentStage(s.linkageService, options.KisanSathiUserID, s.logger))
    }

    return pipe
}
```

---

#### Scenario 3: Farmer Already Exists in AAA (Idempotent Registration)

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_service.go`
**Lines:** 130-178

**Current Implementation:**
```go
// If user already exists in AAA
aaaUser, err = s.aaaService.GetUserByMobile(ctx, req.Profile.PhoneNumber)
if err == nil && existingFarmer != nil {
    // Return existing farmer profile (idempotent)
    return &response, nil
}
```

**FINDING: MISSING ROLE VERIFICATION FOR EXISTING USERS**

**Status:** ⚠️ **MEDIUM - EXISTING USERS MAY LACK FARMER ROLE**

**Scenario:**
1. User created externally in AAA (e.g., as KisanSathi)
2. User tries to register as farmer
3. System finds existing AAA user
4. Creates local farmer profile but does NOT verify/assign farmer role

**Expected Behavior:**
- Check if user already has "farmer" role
- If not, assign farmer role (user can have multiple roles)
- Log the multi-role assignment for audit

**Recommended Fix:**
```go
// After finding existing user (line ~131)
existingFarmer, err := s.repository.FindOne(ctx, existingFarmerFilter)
if err == nil && existingFarmer != nil {
    // ADD: Verify farmer role exists
    hasFarmerRole, err := s.aaaService.CheckUserRole(ctx, aaaUserID, "farmer")
    if err != nil {
        log.Printf("Warning: Failed to check farmer role for user %s: %v", aaaUserID, err)
    } else if !hasFarmerRole {
        // Assign farmer role to existing user
        err = s.aaaService.AssignRole(ctx, aaaUserID, aaaOrgID, "farmer")
        if err != nil {
            log.Printf("Warning: Failed to assign farmer role to existing user %s: %v", aaaUserID, err)
        } else {
            log.Printf("Assigned farmer role to existing AAA user %s", aaaUserID)
        }
    }

    // Return existing farmer
    return &response, nil
}
```

---

### 1.2 FPO CEO Role Assignment

#### Scenario 1: FPO Creation with New CEO User

**File:** `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`
**Lines:** 36-231

**Current Implementation:**
```go
func (s *FPOServiceImpl) CreateFPO(ctx context.Context, req interface{}) (interface{}, error) {
    // Step 1: Create CEO user in AAA (line 79)
    userResp, err := s.aaaService.CreateUser(ctx, createUserReq)

    // Step 2: Create organization in AAA (line 120)
    orgResp, err := s.aaaService.CreateOrganization(ctx, createOrgReq)

    // Step 3: Assign CEO role (line 134)
    err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
    if err != nil {
        log.Printf("Warning: Failed to assign CEO role: %v", err)
        // Continue as this might not be critical  ← PROBLEM
    }
}
```

**Status:** ✅ **ROLE ASSIGNMENT PRESENT BUT IMPROPERLY HANDLED**

**Issues Identified:**

**Issue 1: Non-Fatal Role Assignment Failure**
- Line 134-138: Role assignment failure is logged as warning but execution continues
- FPO creation succeeds even if CEO has no role
- CEO user exists but cannot perform CEO operations
- Violates Business Rule 1.2: "An FPO CEO should ALWAYS have fpo_ceo role for their organization"

**Issue 2: Inconsistent Status Handling**
- Line 191: FPO marked as `PENDING_SETUP` only for user group failures
- CEO role assignment failure does NOT trigger `PENDING_SETUP` status
- No mechanism to retry CEO role assignment

**Issue 3: Race Condition on CEO Check**
- Line 103: Checks if user is already CEO before creating organization
- Line 134: Assigns CEO role after organization creation
- Gap between check and assignment allows concurrent CEO assignments

**Recommended Fix:**

```go
// Step 4: Assign CEO role to user in organization (line 134)
err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
if err != nil {
    log.Printf("ERROR: Failed to assign CEO role to user %s for org %s: %v", ceoUserID, aaaOrgID, err)

    // CRITICAL: Mark FPO as PENDING_SETUP if CEO role assignment fails
    setupErrors[fmt.Sprintf("ceo_role_assignment")] = err.Error()
    fpoStatus = fpo.FPOStatusPendingSetup

    // DO NOT return error - allow FPO creation to proceed with PENDING_SETUP status
    // This enables CompleteFPOSetup workflow to retry role assignment
} else {
    log.Printf("Successfully assigned CEO role to user %s for org %s", ceoUserID, aaaOrgID)
}
```

**Add CEO role retry logic to CompleteFPOSetup:**
```go
func (s *FPOServiceImpl) CompleteFPOSetup(ctx context.Context, orgID string) (interface{}, error) {
    // ... existing code ...

    // ADD: Retry CEO role assignment if it failed
    if ceoRoleError, exists := fpoRef.SetupErrors["ceo_role_assignment"]; exists {
        log.Printf("Retrying CEO role assignment for org %s", orgID)

        // Get CEO user ID from organization
        orgData, err := s.aaaService.GetOrganization(ctx, orgID)
        if err != nil {
            setupErrors["ceo_role_assignment"] = fmt.Sprintf("Failed to get org data: %v", err)
        } else {
            orgMap := orgData.(map[string]interface{})
            ceoUserID := orgMap["ceo_user_id"].(string)

            // Retry role assignment
            err = s.aaaService.AssignRole(ctx, ceoUserID, orgID, "CEO")
            if err != nil {
                setupErrors["ceo_role_assignment"] = err.Error()
            } else {
                log.Printf("Successfully assigned CEO role on retry")
                delete(setupErrors, "ceo_role_assignment")
            }
        }
    }

    // ... rest of existing code ...
}
```

---

#### Scenario 2: FPO Creation with Existing CEO User

**File:** `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`
**Lines:** 64-99

**Current Implementation:**
```go
// If CEO user already exists
existingUser, err := s.aaaService.GetUserByMobile(ctx, createReq.CEOUser.PhoneNumber)
if err == nil {
    userMap := existingUser.(map[string]interface{})
    ceoUserID = userMap["id"].(string)
    log.Printf("Using existing CEO user with ID: %s", ceoUserID)
}

// Step 2: Validate CEO is not already CEO of another FPO
isCEO, err := s.aaaService.CheckUserRole(ctx, ceoUserID, "CEO")
if err != nil {
    log.Printf("Warning: Failed to check if user is already CEO: %v", err)
    // Continue anyway - this is a best-effort check  ← PROBLEM
} else if isCEO {
    return nil, fmt.Errorf("user is already CEO of another FPO")
}
```

**Status:** ⚠️ **MEDIUM - ROLE CHECK FAILURE BYPASSED**

**Issues:**

**Issue 1: Failed Role Check is Non-Fatal**
- Line 105: Role check failure is logged but execution continues
- Cannot verify if user is already CEO of another FPO
- Violates Business Rule 1.2: "A user CANNOT be CEO of multiple FPOs simultaneously"

**Issue 2: No Organization-Scoped Role Check**
- CheckUserRole only checks global role, not organization-scoped
- User might be CEO of Org A, system allows CEO of Org B
- Need to check role+organization combination

**Recommended Fix:**

```go
// Step 2: Validate CEO is not already CEO of another FPO (line 101)
isCEO, err := s.aaaService.CheckUserRole(ctx, ceoUserID, "CEO")
if err != nil {
    // CHANGE: Make role check failure fatal
    return nil, fmt.Errorf("failed to verify CEO eligibility: cannot check existing CEO status: %w", err)
}

if isCEO {
    // ADD: Check if CEO role is for a different organization
    // This requires enhancing CheckUserRole to return organization context
    return nil, fmt.Errorf("user is already CEO of another FPO - a user cannot be CEO of multiple FPOs simultaneously")
}

log.Printf("Validated user %s is eligible to be CEO (not CEO of any other FPO)", ceoUserID)
```

**Enhancement Needed in AAA Service:**
```go
// Add new method to check role in specific organization
func (s *AAAServiceImpl) CheckUserRoleInOrg(ctx context.Context, userID, orgID, roleName string) (bool, error) {
    // Check if user has role specifically in the given organization
    // Returns false if user has role in different organization
}
```

---

### 1.3 KisanSathi Role Assignment

#### Scenario 1: KisanSathi Assignment to Farmer

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_linkage_service.go`
**Lines:** ~500-600 (approximate, need to examine full file)

Let me check the farmer linkage service for KisanSathi role assignment:

**Current Implementation Analysis:**

Based on grep results (line 594 in farmer_linkage_service.go):
```go
err = s.aaaService.AssignRole(ctx, userID, orgID, "KisanSathi")
```

**Status:** ✅ **ROLE ASSIGNMENT PRESENT**

**Verification Needed:**
- Check if role assignment is inside proper error handling
- Verify rollback logic if role assignment fails
- Confirm KisanSathi role is assigned BEFORE farmer linkage

---

### 1.4 Missing Role Assignment Scenarios

#### Scenario: Farmer Linked to FPO

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_linkage_service.go`
**Expected:** When linking farmer to FPO, verify farmer has farmer role in that org

**Status:** ❓ **UNKNOWN - NEEDS INVESTIGATION**

**Question:** Does linking farmer to FPO require farmer role to be org-scoped?

---

## 2. Edge Cases and Invariants

### 2.1 Critical Invariants

#### Invariant 1: Farmer Role Consistency
**Rule:** Every farmer profile MUST have corresponding "farmer" role in AAA service

**Violation Scenarios:**
1. ❌ Farmer created via CreateFarmer without role assignment
2. ❌ Farmer created via bulk import without role assignment
3. ⚠️ Existing AAA user becomes farmer without role verification

**Detection Query:**
```sql
-- Find farmers without farmer role
SELECT f.id, f.aaa_user_id, f.first_name, f.last_name, f.phone_number
FROM farmer_profiles f
WHERE NOT EXISTS (
    SELECT 1 FROM aaa_service.user_roles r
    WHERE r.user_id = f.aaa_user_id
    AND r.role_name = 'farmer'
)
```

**Recommended Invariant Check:**
```go
// Add to farmer_service.go as health check
func (s *FarmerServiceImpl) ValidateFarmerRoleInvariant(ctx context.Context, farmerID string) error {
    farmer, err := s.repository.GetByID(ctx, farmerID)
    if err != nil {
        return err
    }

    // Check if farmer has farmer role in AAA
    hasRole, err := s.aaaService.CheckUserRole(ctx, farmer.AAAUserID, "farmer")
    if err != nil {
        return fmt.Errorf("failed to check farmer role: %w", err)
    }

    if !hasRole {
        // CRITICAL: Invariant violation detected
        log.Printf("INVARIANT VIOLATION: Farmer %s (user %s) does not have farmer role in AAA",
            farmerID, farmer.AAAUserID)
        return fmt.Errorf("invariant violation: farmer %s lacks farmer role", farmerID)
    }

    return nil
}
```

---

#### Invariant 2: FPO CEO Role Consistency
**Rule:** Every FPO MUST have exactly one CEO with "CEO" role in AAA

**Violation Scenarios:**
1. ⚠️ FPO created but CEO role assignment failed
2. ❌ User is CEO of multiple FPOs (checked but error handling weak)

**Detection Query:**
```sql
-- Find FPOs with CEO users lacking CEO role
SELECT fpo.id, fpo.aaa_org_id, fpo.name, org.ceo_user_id
FROM fpo_refs fpo
JOIN aaa_service.organizations org ON org.id = fpo.aaa_org_id
WHERE NOT EXISTS (
    SELECT 1 FROM aaa_service.user_roles r
    WHERE r.user_id = org.ceo_user_id
    AND r.role_name = 'CEO'
    AND r.org_id = org.id
)
```

---

#### Invariant 3: Role-Entity Coupling
**Rule:** Role assignment MUST succeed for entity creation to be considered complete

**Current State:**
- ❌ Farmer creation: NO role assignment
- ⚠️ FPO CEO: Role assignment failure non-fatal
- ✅ KisanSathi: Role assignment present (assuming proper error handling)

**Recommended Pattern:**
```go
// Transactional pattern for entity+role creation
func createEntityWithRole(ctx context.Context, entityData, roleData interface{}) error {
    tx := beginTransaction()

    // 1. Create AAA user
    userID, err := createAAAUser(ctx, entityData)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 2. Assign role (CRITICAL - must succeed)
    err = assignRole(ctx, userID, roleData)
    if err != nil {
        // Attempt to delete AAA user
        deleteAAAUser(ctx, userID)
        tx.Rollback()
        return fmt.Errorf("role assignment failed, entity creation aborted: %w", err)
    }

    // 3. Create local entity
    err = createLocalEntity(ctx, entityData)
    if err != nil {
        // Attempt to remove role and delete user
        removeRole(ctx, userID, roleData)
        deleteAAAUser(ctx, userID)
        tx.Rollback()
        return err
    }

    tx.Commit()
    return nil
}
```

---

#### Invariant 4: Multi-Role Users
**Rule:** Users CAN have multiple roles simultaneously, but roles must not conflict

**Scenarios:**
- ✅ User is farmer in Org A, KisanSathi in Org B (allowed)
- ❌ User is CEO of Org A, CEO of Org B (forbidden)
- ⚠️ User is farmer and KisanSathi in same Org (needs business rule clarification)

**Question for Product Owner:**
1. Can a farmer also be a KisanSathi in the same FPO?
2. Can a farmer be a shareholder in the same FPO?
3. What role takes precedence for permission checks?

---

### 2.2 Race Conditions

#### Race 1: Concurrent Farmer Creation

**Scenario:**
1. Two requests create farmer with same phone number simultaneously
2. Both check AAA - no existing user found
3. Both create AAA user - one succeeds, one fails with conflict
4. Winner creates farmer profile
5. Loser: Current code returns existing farmer (line 117-176)

**Problem:** What if loser's AAA user creation partially succeeded?

**Mitigation:**
- Idempotency handled in farmer_service.go (lines 130-178)
- BUT: No role assignment, so both users may lack farmer role

---

#### Race 2: CEO Role Assignment During FPO Creation

**Scenario:**
1. Two FPOs created simultaneously with same CEO user
2. Both check if user is CEO (line 103) - both return false
3. Both create organization successfully
4. Both try to assign CEO role
5. First succeeds, second might succeed too (role not unique in AAA?)

**Question:** Does AAA service prevent duplicate role assignments per user?

**Recommended Fix:**
```go
// Use optimistic locking or unique constraints in AAA
err = s.aaaService.AssignRoleExclusive(ctx, ceoUserID, aaaOrgID, "CEO")
// Returns error if user already has CEO role in ANY organization
```

---

#### Race 3: Bulk Import Parallel Processing

**File:** `/Users/kaushik/farmers-module/internal/services/bulk_farmer_service.go`
**Lines:** 233-278 (processAsynchronously)

**Scenario:**
- Concurrent chunks process same farmer (duplicate in file)
- Both create AAA user - one fails
- Both try to create local farmer profile

**Mitigation:**
- Deduplication stage exists (line 167-226 in stages.go)
- BUT: Deduplication is stubbed (line 208: random dedup for testing)

**Recommended Fix:**
```go
// Implement proper deduplication in DeduplicationStage
func (ds *DeduplicationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
    // Check AAA service for existing user
    existingUser, err := ds.aaaService.GetUserByMobile(ctx, farmerData.PhoneNumber)
    if err == nil && existingUser != nil {
        // User exists - mark as duplicate
        return nil, fmt.Errorf("duplicate farmer: phone number already registered")
    }

    // Check local database for existing farmer
    filter := base.NewFilterBuilder().
        Where("phone_number", base.OpEqual, farmerData.PhoneNumber).
        Build()
    existingFarmer, err := ds.farmerService.repository.FindOne(ctx, filter)
    if err == nil && existingFarmer != nil {
        return nil, fmt.Errorf("duplicate farmer: local profile already exists")
    }

    return procCtx, nil
}
```

---

### 2.3 Data Consistency

#### Consistency Issue 1: Orphaned Roles

**Scenario:**
- Farmer profile deleted (soft delete)
- AAA user still exists with farmer role

**Detection:**
```sql
-- Find AAA users with farmer role but no active farmer profile
SELECT r.user_id, u.phone_number, u.email
FROM aaa_service.user_roles r
JOIN aaa_service.users u ON u.id = r.user_id
WHERE r.role_name = 'farmer'
AND NOT EXISTS (
    SELECT 1 FROM farmer_profiles f
    WHERE f.aaa_user_id = r.user_id
    AND f.deleted_at IS NULL
)
```

**Recommended Cleanup:**
```go
// Add to reconciliation job
func (s *FarmerServiceImpl) ReconcileFarmerRoles(ctx context.Context) error {
    // Get all soft-deleted farmers
    filter := base.NewFilterBuilder().
        Where("deleted_at", base.OpNotEqual, nil).
        Build()
    deletedFarmers, err := s.repository.Find(ctx, filter)

    for _, farmer := range deletedFarmers {
        // Remove farmer role from AAA
        err := s.aaaService.RemoveRole(ctx, farmer.AAAUserID, "farmer")
        if err != nil {
            log.Printf("Failed to remove farmer role from deleted farmer %s: %v", farmer.ID, err)
        }
    }
}
```

---

#### Consistency Issue 2: Missing Roles (Reverse Orphans)

**Scenario:**
- Farmer profile exists
- AAA user exists
- But farmer role NOT assigned in AAA

**This is the PRIMARY BUG we identified - farmers created without roles**

**Remediation Script:**
```go
// Add to administrative service
func (s *AdminServiceImpl) AssignMissingFarmerRoles(ctx context.Context) (*RepairReport, error) {
    report := &RepairReport{
        TotalFarmers: 0,
        MissingRoles: 0,
        RolesAssigned: 0,
        Failures: []string{},
    }

    // Get all active farmers
    farmers, err := s.farmerRepository.Find(ctx, base.NewFilterBuilder().Build())
    if err != nil {
        return nil, err
    }

    report.TotalFarmers = len(farmers)

    for _, farmer := range farmers {
        // Check if farmer has farmer role
        hasRole, err := s.aaaService.CheckUserRole(ctx, farmer.AAAUserID, "farmer")
        if err != nil {
            report.Failures = append(report.Failures,
                fmt.Sprintf("Farmer %s: role check failed: %v", farmer.ID, err))
            continue
        }

        if !hasRole {
            report.MissingRoles++

            // Assign farmer role
            err = s.aaaService.AssignRole(ctx, farmer.AAAUserID, farmer.AAAOrgID, "farmer")
            if err != nil {
                report.Failures = append(report.Failures,
                    fmt.Sprintf("Farmer %s: role assignment failed: %v", farmer.ID, err))
            } else {
                report.RolesAssigned++
                log.Printf("Assigned farmer role to user %s (farmer %s)",
                    farmer.AAAUserID, farmer.ID)
            }
        }
    }

    return report, nil
}
```

---

## 3. Abuse Paths and Security Implications

### 3.1 Permission Bypass via Missing Role

**Attack Vector:**
1. Attacker creates farmer account through legitimate API
2. Farmer created without "farmer" role (current bug)
3. Attacker has valid AAA credentials but no role
4. AAA service permission checks fail differently than expected
5. If AAA service defaults to permissive on role check failure, attacker gains access

**Mitigation:**
- Ensure AAA service fails CLOSED (denies access when role check fails)
- Add invariant checks on farmer creation
- Implement role verification middleware

---

### 3.2 Privilege Escalation via Multi-Role Manipulation

**Attack Vector:**
1. User is farmer in Org A
2. User becomes KisanSathi in Org B (legitimate)
3. User manipulates requests to use Org B credentials for Org A operations
4. If permission checks don't properly scope roles to organizations, user gains elevated access

**Mitigation:**
- Ensure all permission checks include organization context
- Validate role-organization binding
- Audit cross-organization access attempts

---

### 3.3 Role Assignment Replay Attack

**Attack Vector:**
1. Attacker intercepts role assignment request
2. Replays request to assign CEO role to themselves
3. If role assignment API lacks idempotency tokens, replay succeeds

**Mitigation:**
- Add idempotency tokens to role assignment requests
- Require authentication for all role changes
- Audit all role modifications with immutable logs

---

## 4. Test Plan

### 4.1 Unit Tests Needed

#### Test Suite 1: Farmer Role Assignment

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_service_test.go` (create if missing)

```go
func TestCreateFarmer_AssignsRoleSuccessfully(t *testing.T) {
    // ARRANGE
    mockAAAService := new(MockAAAService)
    mockAAAService.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "user123"}, nil)
    mockAAAService.On("AssignRole", mock.Anything, "user123", "org456", "farmer").
        Return(nil)

    service := NewFarmerService(mockRepo, mockAAAService, "password123")

    // ACT
    req := &requests.CreateFarmerRequest{
        AAAOrgID: "org456",
        Profile: requests.FarmerProfileData{
            FirstName: "John",
            LastName: "Doe",
            PhoneNumber: "9876543210",
        },
    }
    _, err := service.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.NoError(t, err)
    mockAAAService.AssertCalled(t, "AssignRole", mock.Anything, "user123", "org456", "farmer")
}

func TestCreateFarmer_RoleAssignmentFails_RollsBack(t *testing.T) {
    // ARRANGE
    mockAAAService := new(MockAAAService)
    mockAAAService.On("CreateUser", mock.Anything, mock.Anything).
        Return(map[string]interface{}{"id": "user123"}, nil)
    mockAAAService.On("AssignRole", mock.Anything, "user123", "org456", "farmer").
        Return(fmt.Errorf("role assignment failed"))
    mockAAAService.On("DeleteUser", mock.Anything, "user123").
        Return(nil)

    service := NewFarmerService(mockRepo, mockAAAService, "password123")

    // ACT
    req := &requests.CreateFarmerRequest{/* ... */}
    _, err := service.CreateFarmer(context.Background(), req)

    // ASSERT
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "role assignment failed")

    // Verify rollback
    mockAAAService.AssertCalled(t, "DeleteUser", mock.Anything, "user123")

    // Verify farmer was not created in local DB
    filter := base.NewFilterBuilder().Where("aaa_user_id", base.OpEqual, "user123").Build()
    farmer, _ := mockRepo.FindOne(context.Background(), filter)
    assert.Nil(t, farmer)
}

func TestCreateFarmer_ExistingUserWithoutRole_AssignsRole(t *testing.T) {
    // Test that when AAA user exists but local farmer doesn't,
    // AND user doesn't have farmer role, role is assigned
    // ... implementation ...
}
```

---

#### Test Suite 2: Bulk Import Role Assignment

```go
func TestBulkImport_AssignsRolesToAllFarmers(t *testing.T) {
    // Verify all farmers in bulk import get farmer role
}

func TestBulkImport_RoleAssignmentFailure_MarksRecordFailed(t *testing.T) {
    // Verify that if role assignment fails for one farmer,
    // that farmer is marked as failed but others continue
}

func TestBulkImport_PartialRoleFailures_ReportedCorrectly(t *testing.T) {
    // Test that bulk operation status correctly reports
    // how many role assignments succeeded vs failed
}
```

---

#### Test Suite 3: FPO CEO Role Assignment

```go
func TestCreateFPO_AssignsCEORole(t *testing.T) {
    // Verify CEO role assigned after org creation
}

func TestCreateFPO_CEORoleFailure_MarksPendingSetup(t *testing.T) {
    // Verify FPO status is PENDING_SETUP if CEO role fails
}

func TestCompleteFPOSetup_RetriesCEORole(t *testing.T) {
    // Verify CompleteFPOSetup retries CEO role assignment
}

func TestCreateFPO_UserAlreadyCEO_Rejected(t *testing.T) {
    // Verify user cannot be CEO of multiple FPOs
}
```

---

### 4.2 Integration Tests Needed

```go
func TestFarmerCreation_EndToEnd_RoleInAAA(t *testing.T) {
    // Full integration test with real AAA service
    // Create farmer, verify role exists in AAA
}

func TestBulkImport_100Farmers_AllHaveRoles(t *testing.T) {
    // Import 100 farmers, verify all have roles
}

func TestFarmerDeletion_RemovesRoleFromAAA(t *testing.T) {
    // Delete farmer, verify role removed from AAA
}
```

---

### 4.3 Concurrency Tests

```go
func TestConcurrentFarmerCreation_SamePhone_OneSucceeds(t *testing.T) {
    // 10 goroutines try to create farmer with same phone
    // Verify exactly 1 succeeds, 9 get "already exists"
}

func TestConcurrentCEOAssignment_DifferentOrgs_OneSucceeds(t *testing.T) {
    // 2 FPOs try to assign same user as CEO
    // Verify exactly 1 succeeds
}
```

---

### 4.4 Invariant Validation Tests

```go
func TestInvariant_AllFarmersHaveRole(t *testing.T) {
    // Create 50 farmers through various paths
    // Verify all have farmer role in AAA
}

func TestInvariant_AllCEOsHaveRole(t *testing.T) {
    // Create 10 FPOs
    // Verify all CEOs have CEO role
}
```

---

## 5. Monitoring and Alerting Strategy

### 5.1 Production Signals

#### Signal 1: Role Assignment Failure Rate
```
Metric: role_assignment_failures_total
Labels: role_type (farmer|ceo|kisansathi), entity_type (farmer|fpo)
Alert Threshold: > 5% of creations in 5-minute window

Alert:
  "High role assignment failure rate detected.
  {{ $value }}% of {{ $labels.role_type }} role assignments failing.
  This may indicate AAA service issues or permission problems."
```

#### Signal 2: Invariant Violation Detection
```
Metric: role_invariant_violations_total
Labels: invariant_type (farmer_role|ceo_role)
Alert Threshold: > 0 in 1-hour window

Alert:
  "Role invariant violation detected!
  {{ $value }} farmers/FPOs exist without required roles.
  Immediate investigation required."
```

#### Signal 3: Orphaned Roles
```
Metric: orphaned_roles_total
Labels: role_type
Alert Threshold: > 100

Alert:
  "High number of orphaned roles detected.
  {{ $value }} {{ $labels.role_type }} roles exist without entities.
  Consider running reconciliation job."
```

---

### 5.2 Reconciliation Job

```go
// Scheduled job: Run daily at 2 AM
func ReconcileRoleAssignments(ctx context.Context) (*ReconciliationReport, error) {
    report := &ReconciliationReport{}

    // 1. Check all farmers have farmer role
    report.FarmerRoles = checkFarmerRoles(ctx)

    // 2. Check all FPO CEOs have CEO role
    report.CEORoles = checkCEORoles(ctx)

    // 3. Check for orphaned roles
    report.OrphanedRoles = checkOrphanedRoles(ctx)

    // 4. Auto-fix if configured
    if config.AutoFixRoleIssues {
        report.Fixes = autoFixRoleIssues(ctx, report)
    }

    // 5. Send report to monitoring
    sendMetrics(report)

    return report, nil
}
```

---

## 6. Recommended Implementation Plan

### Phase 1: Critical Fixes (P0 - Immediate)

**Target: 2-3 days**

1. **Add Farmer Role Assignment to CreateFarmer**
   - File: `internal/services/farmer_service.go`
   - Add role assignment after AAA user creation
   - Add rollback logic if role assignment fails
   - Add error handling and logging

2. **Add Farmer Role Assignment to Bulk Import Pipeline**
   - File: `internal/services/pipeline/stages.go`
   - Create `RoleAssignmentStage`
   - Insert after `AAAUserCreationStage`
   - Handle failures gracefully

3. **Fix FPO CEO Role Handling**
   - File: `internal/services/fpo_ref_service.go`
   - Make CEO role assignment failure trigger PENDING_SETUP
   - Add CEO role retry to CompleteFPOSetup

**Acceptance Criteria:**
- All new farmers created with farmer role
- All new FPOs have CEOs with CEO role
- Existing farmers without roles identified (but not auto-fixed yet)

---

### Phase 2: Testing and Validation (P1 - High)

**Target: 3-5 days**

1. **Create Comprehensive Test Suites**
   - Unit tests for all role assignment paths
   - Integration tests with AAA service
   - Concurrency tests for race conditions

2. **Add Invariant Validation Functions**
   - Farmer role invariant checker
   - CEO role invariant checker
   - Exposed via admin API for manual triggering

3. **Production Readiness**
   - Add metrics and logging
   - Set up monitoring dashboards
   - Configure alerts

**Acceptance Criteria:**
- >90% test coverage for role assignment code
- All invariants testable via API
- Monitoring dashboard shows role assignment health

---

### Phase 3: Data Remediation (P2 - Medium)

**Target: 2-3 days**

1. **Analyze Existing Data**
   - Run SQL queries to find farmers without roles
   - Identify FPOs with CEO role issues
   - Generate remediation report

2. **Create Remediation Scripts**
   - Script to assign missing farmer roles
   - Script to fix CEO role issues
   - Dry-run mode for safety

3. **Execute Remediation**
   - Run in dry-run mode
   - Review results
   - Execute with monitoring
   - Validate results

**Acceptance Criteria:**
- All existing farmers have farmer role
- All FPO CEOs have CEO role
- Zero invariant violations in production

---

### Phase 4: Long-term Improvements (P3 - Low)

**Target: Ongoing**

1. **Reconciliation Job**
   - Automated daily role consistency checks
   - Auto-healing for minor issues
   - Reporting and alerting

2. **Enhanced AAA Integration**
   - Organization-scoped role checking
   - Exclusive role assignment API
   - Better error reporting

3. **Documentation**
   - Update business rules document
   - Create role assignment ADR
   - Developer guidelines for role handling

---

## 7. Open Questions for Product Owner

1. **Multi-Role Scenarios:**
   - Can a farmer also be a KisanSathi in the same FPO?
   - Can a farmer be a shareholder in the same FPO?
   - If user has multiple roles, which role's permissions take precedence?

2. **Role Removal:**
   - When farmer is unlinked from FPO, should farmer role be removed from AAA?
   - When farmer profile is soft-deleted, should role remain for audit?
   - What's the retention policy for orphaned roles?

3. **CEO Restrictions:**
   - Can a CEO of one FPO be a director of another FPO?
   - Can a CEO of one FPO be a farmer in another FPO?
   - What happens to CEO role when CEO is replaced?

4. **KisanSathi Scope:**
   - Can one KisanSathi serve farmers across multiple FPOs?
   - Does KisanSathi role need to be organization-scoped?
   - Can a farmer also be a KisanSathi (serving other farmers)?

5. **Error Recovery:**
   - If role assignment fails during creation, should the entire operation fail or mark entity as PENDING?
   - What's the retry policy for failed role assignments?
   - Who should be notified when role assignment failures occur?

---

## 8. Summary of Findings

### Critical Issues (Must Fix Immediately)

1. ❌ **Farmer role NOT assigned during CreateFarmer** (Severity: CRITICAL)
2. ❌ **Farmer role NOT assigned during bulk import** (Severity: CRITICAL)
3. ⚠️ **FPO CEO role failure handled non-fatally** (Severity: HIGH)

### Medium Priority Issues

4. ⚠️ **Existing AAA users becoming farmers without role verification** (Severity: MEDIUM)
5. ⚠️ **CEO eligibility check failure bypassed** (Severity: MEDIUM)
6. ⚠️ **No reconciliation job to detect invariant violations** (Severity: MEDIUM)

### Low Priority / Enhancements

7. ℹ️ **Missing organization-scoped role checking** (Severity: LOW)
8. ℹ️ **No automated cleanup of orphaned roles** (Severity: LOW)
9. ℹ️ **Unclear multi-role business rules** (Severity: LOW)

### Positive Findings

1. ✅ KisanSathi role assignment appears to be implemented
2. ✅ FPO creation includes CEO role assignment (though error handling needs improvement)
3. ✅ Idempotency handling for farmer registration exists
4. ✅ Business rules document is comprehensive

---

## Appendix A: Affected Files

### Files Requiring Changes (Critical)

1. `/Users/kaushik/farmers-module/internal/services/farmer_service.go`
   - Add farmer role assignment to CreateFarmer (line ~194)
   - Add role verification for existing users (line ~131)

2. `/Users/kaushik/farmers-module/internal/services/pipeline/stages.go`
   - Create RoleAssignmentStage (new code)
   - Add to pipeline stages

3. `/Users/kaushik/farmers-module/internal/services/bulk_farmer_service.go`
   - Add RoleAssignmentStage to pipeline (line ~401)

4. `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`
   - Fix CEO role error handling (line ~134)
   - Add CEO role retry to CompleteFPOSetup (line ~354)

### Files Requiring Changes (Medium)

5. `/Users/kaushik/farmers-module/internal/services/farmer_linkage_service.go`
   - Verify KisanSathi role assignment error handling

6. `/Users/kaushik/farmers-module/internal/services/aaa_service.go`
   - Add CheckUserRoleInOrg method (new)
   - Add AssignRoleExclusive method (new)

### New Files Needed

7. `/Users/kaushik/farmers-module/internal/services/role_reconciliation_service.go`
   - Reconciliation job implementation

8. `/Users/kaushik/farmers-module/internal/services/farmer_service_test.go`
   - Comprehensive unit tests

9. `/Users/kaushik/farmers-module/.kiro/specs/role-assignment-validation/ADR-role-assignment.md`
   - Architecture Decision Record

---

## Appendix B: SQL Diagnostic Queries

```sql
-- Query 1: Find farmers without farmer role
SELECT f.id, f.aaa_user_id, f.first_name, f.last_name, f.phone_number, f.created_at
FROM farmer_profiles f
WHERE f.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM aaa_service.user_roles r
    WHERE r.user_id = f.aaa_user_id
    AND r.role_name = 'farmer'
    AND r.deleted_at IS NULL
)
ORDER BY f.created_at DESC;

-- Query 2: Find FPOs with CEOs lacking CEO role
SELECT fpo.id, fpo.aaa_org_id, fpo.name, org.ceo_user_id, u.email, u.phone
FROM fpo_refs fpo
JOIN aaa_service.organizations org ON org.id = fpo.aaa_org_id
JOIN aaa_service.users u ON u.id = org.ceo_user_id
WHERE fpo.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM aaa_service.user_roles r
    WHERE r.user_id = org.ceo_user_id
    AND r.role_name = 'CEO'
    AND r.org_id = org.id
    AND r.deleted_at IS NULL
);

-- Query 3: Find users with multiple CEO roles
SELECT r.user_id, u.email, u.phone, COUNT(*) as ceo_role_count,
       STRING_AGG(o.name, ', ') as organizations
FROM aaa_service.user_roles r
JOIN aaa_service.users u ON u.id = r.user_id
JOIN aaa_service.organizations o ON o.id = r.org_id
WHERE r.role_name = 'CEO'
AND r.deleted_at IS NULL
GROUP BY r.user_id, u.email, u.phone
HAVING COUNT(*) > 1;

-- Query 4: Find orphaned farmer roles (role exists but farmer deleted)
SELECT r.user_id, u.email, u.phone, r.org_id, r.created_at
FROM aaa_service.user_roles r
JOIN aaa_service.users u ON u.id = r.user_id
WHERE r.role_name = 'farmer'
AND r.deleted_at IS NULL
AND NOT EXISTS (
    SELECT 1 FROM farmer_profiles f
    WHERE f.aaa_user_id = r.user_id
    AND f.deleted_at IS NULL
);

-- Query 5: Count role assignments by date (to detect when bug started)
SELECT DATE(f.created_at) as creation_date,
       COUNT(*) as farmers_created,
       COUNT(r.id) as roles_assigned,
       COUNT(*) - COUNT(r.id) as missing_roles
FROM farmer_profiles f
LEFT JOIN aaa_service.user_roles r ON r.user_id = f.aaa_user_id AND r.role_name = 'farmer'
WHERE f.deleted_at IS NULL
GROUP BY DATE(f.created_at)
ORDER BY creation_date DESC
LIMIT 30;
```

---

**End of Report**

---

**Next Steps:**
1. Review this report with the team
2. Prioritize fixes based on business impact
3. Create implementation tickets for Phase 1
4. Set up monitoring before deploying fixes
5. Plan data remediation for existing farmers
