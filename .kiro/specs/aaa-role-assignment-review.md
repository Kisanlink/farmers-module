# AAA Service Integration: Role Seeding and Assignment Review

**Date:** 2025-10-16
**Reviewer:** SDE-3 Backend Architect
**Status:** Analysis Complete - Implementation Required

## Executive Summary

This document provides a comprehensive analysis of the AAA (Authentication, Authorization, Auditing) service integration within the farmers-module, specifically focusing on role seeding and role assignment during entity creation workflows. The review identifies **critical gaps** in role assignment logic that create security vulnerabilities and operational inconsistencies.

**Critical Findings:**
- **CRITICAL GAP**: Farmer creation workflow does NOT assign `farmer` role to users
- **CRITICAL GAP**: FPO creation workflow assigns `CEO` role but lacks validation and proper error handling
- **PARTIAL IMPLEMENTATION**: KisanSathi role assignment works but has reliability issues
- **MISSING**: No role seeding on application startup
- **MISSING**: No idempotency checks for role assignment

---

## 1. Current State Analysis

### 1.1 Role Seeding Implementation

**File:** `/Users/kaushik/farmers-module/internal/services/aaa_service.go`

#### Current Implementation
```go
// Line 68-75: SeedRolesAndPermissions delegates to AAA client
func (s *AAAServiceImpl) SeedRolesAndPermissions(ctx context.Context) error {
    if s.client == nil {
        log.Println("AAA client not available, skipping seeding")
        return nil
    }
    return s.client.SeedRolesAndPermissions(ctx)
}
```

**File:** `/Users/kaushik/farmers-module/internal/clients/aaa/aaa_client.go`

```go
// Line 900-924: Actual seeding call to AAA service
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
    log.Println("AAA SeedRolesAndPermissions: Seeding roles and permissions")

    grpcReq := &pb.SeedRolesAndPermissionsRequest{
        Force: false, // Don't force reseed if data exists
    }

    response, err := c.catalogClient.SeedRolesAndPermissions(ctx, grpcReq)
    if err != nil {
        // Error handling...
    }

    log.Printf("Roles and permissions seeded successfully: %d roles, %d permissions created",
        response.RolesCreated, response.PermissionsCreated)
    return nil
}
```

#### Analysis

**Strengths:**
- Properly delegates to AAA service for centralized role management
- Includes graceful degradation when AAA client is unavailable
- Uses `Force: false` for idempotent seeding

**Critical Issues:**
1. **No Startup Seeding**: `SeedRolesAndPermissions` is NEVER called during application initialization (not in `main.go` or service factory)
2. **No Role Validation**: System does not verify that required roles (`farmer`, `fpo_ceo`, `kisansathi`) exist before attempting assignments
3. **No Explicit Role Definitions**: The roles are defined in AAA service but not documented in this codebase

**Required Roles (inferred from workflows):**
- `farmer` - Required for all farmer users
- `fpo_ceo` / `CEO` - Required for FPO chief executive officers
- `kisansathi` / `KisanSathi` - Required for field agents
- `fpo_manager` - Listed in AAA client validation (line 629)
- `admin`, `readonly` - Listed in AAA client validation (line 626, 630)

---

### 1.2 Farmer Creation Workflow

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_service.go`

#### Current Implementation Flow

```go
// Line 61-323: CreateFarmer workflow
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest)
    (*responses.FarmerResponse, error) {

    // Step 1: Validate input
    // Step 2: Determine AAA user ID (existing or create new)

    // Line 100-121: Create user in AAA (if needed)
    createUserReq := map[string]interface{}{
        "phone_number": req.Profile.PhoneNumber,
        "password":     s.defaultPassword,
        "country_code": req.Profile.CountryCode,
    }
    aaaUser, err := s.aaaService.CreateUser(ctx, createUserReq)

    // Step 3: Validate KisanSathi (if provided)
    // Line 210-241: KisanSathi role validation
    hasRole, err := s.aaaService.CheckUserRole(ctx, *req.KisanSathiUserID, "kisansathi")

    // Step 4: Create farmer entity in local DB
    farmer := farmerentity.NewFarmer()
    farmer.AAAUserID = aaaUserID
    // ...

    // ⚠️ CRITICAL GAP: NO ROLE ASSIGNMENT FOR FARMER
    // Missing: s.aaaService.AssignRole(ctx, aaaUserID, aaaOrgID, "farmer")

    return response, nil
}
```

#### Critical Gap Analysis

**Issue**: **NO `farmer` ROLE ASSIGNMENT**

**Impact:**
- **Security Risk**: Farmers cannot be distinguished from other user types in AAA
- **Authorization Failure**: Permission checks for farmer-specific resources will fail
- **Data Integrity**: Cannot reliably query "all farmers" from AAA
- **Audit Trail**: No record of farmer role assignment in AAA audit logs

**Expected Behavior:**
After user creation or retrieval, the system should:
```go
// After line 204 (after AAA user ID is determined)
err = s.aaaService.AssignRole(ctx, aaaUserID, aaaOrgID, "farmer")
if err != nil {
    log.Printf("Warning: Failed to assign farmer role: %v", err)
    // Decision: Should this be fatal or warning?
}
```

**Transaction Semantics Question:**
- Should farmer profile creation fail if role assignment fails?
- Current pattern: KisanSathi validation failure is non-fatal (line 215-238)
- Recommendation: Role assignment failure should be **non-fatal with warning** (eventual consistency)

---

### 1.3 FPO Creation Workflow

**File:** `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`

#### Current Implementation Flow

```go
// Line 37-231: CreateFPO workflow
func (s *FPOServiceImpl) CreateFPO(ctx context.Context, req interface{}) (interface{}, error) {

    // Step 1: Create or get CEO user in AAA
    // Line 64-99: CEO user creation/retrieval

    // Step 2: Validate CEO is not already CEO of another FPO
    // Line 102-109: Business Rule 1.2 enforcement
    isCEO, err := s.aaaService.CheckUserRole(ctx, ceoUserID, "CEO")
    if err != nil {
        log.Printf("Warning: Failed to check if user is already CEO: %v", err)
        // ⚠️ ISSUE: Continues despite validation failure
    } else if isCEO {
        return nil, fmt.Errorf("user is already CEO of another FPO")
    }

    // Step 3: Create organization in AAA
    orgResp, err := s.aaaService.CreateOrganization(ctx, createOrgReq)

    // Step 4: Assign CEO role
    // Line 134-138: Role assignment with poor error handling
    err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
    if err != nil {
        log.Printf("Warning: Failed to assign CEO role: %v", err)
        // ⚠️ ISSUE: Continues as "not critical"
    }

    // Step 5: Create user groups (directors, shareholders, etc.)
    // Step 6: Assign permissions to groups
    // Step 7: Store FPO reference locally with status

    return responseData, nil
}
```

#### Gap Analysis

**Issues Identified:**

1. **Non-Fatal Role Assignment (Line 134-138)**
   - **Problem**: CEO role assignment failure is logged as warning but not treated as critical
   - **Impact**: FPO created without proper CEO authorization
   - **Risk**: CEO cannot perform CEO-specific operations
   - **Recommendation**: Should be added to `setupErrors` and FPO marked as `PENDING_SETUP`

2. **Inconsistent Role Name (Line 75 vs 134)**
   - Line 75: Creates user with `role: "CEO"`
   - Line 134: Assigns role `"CEO"` (same)
   - Line 103: Checks for role `"CEO"`
   - **Concern**: Ensure AAA service uses consistent role names
   - **Verification Needed**: Confirm `CEO` vs `fpo_ceo` role naming in AAA

3. **Role Validation Bypass (Line 104-106)**
   - Validation failure is logged but execution continues
   - **Problem**: Could create FPO with invalid CEO
   - **Recommendation**: Should be fatal error (prevent duplicate CEO assignment)

4. **No Idempotency for Role Assignment**
   - If `AssignRole` is called twice, does AAA handle it gracefully?
   - **Recommendation**: Check role exists before assigning (or rely on AAA idempotency)

**Positive Aspects:**
- Uses `setupErrors` JSONB field to track partial failures (line 142, 156, 182)
- Marks FPO as `PENDING_SETUP` if issues occur (line 189-194)
- Provides `CompleteFPOSetup` retry mechanism (line 354-449)

---

### 1.4 KisanSathi Assignment Workflow

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_linkage_service.go`

#### Current Implementation Flow

```go
// Line 178-248: AssignKisanSathi workflow
func (s *FarmerLinkageServiceImpl) AssignKisanSathi(ctx context.Context, req interface{})
    (interface{}, error) {

    // Step 1: Validate input
    // Step 2: Check permissions
    hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID,
        "kisansathi", "assign", assignReq.AAAUserID, assignReq.AAAOrgID)

    // Step 3: Verify KisanSathi user exists
    _, err = s.aaaService.GetUser(ctx, assignReq.KisanSathiUserID)

    // Step 4: Ensure KisanSathi role
    // Line 211-214: ✅ CORRECT PATTERN
    err = s.ensureKisanSathiRole(ctx, assignReq.KisanSathiUserID, assignReq.AAAOrgID)
    if err != nil {
        return nil, fmt.Errorf("failed to ensure KisanSathi role: %w", err)
    }

    // Step 5: Update farmer link with KisanSathi assignment
    farmerLink.KisanSathiUserID = &assignReq.KisanSathiUserID

    return assignmentData, nil
}

// Line 582-609: ensureKisanSathiRole helper
func (s *FarmerLinkageServiceImpl) ensureKisanSathiRole(ctx context.Context,
    userID, orgID string) error {

    // ✅ Check if role already exists (idempotency)
    hasRole, err := s.aaaService.CheckUserRole(ctx, userID, "KisanSathi")
    if hasRole {
        return nil // Already has role
    }

    // Assign role
    err = s.aaaService.AssignRole(ctx, userID, orgID, "KisanSathi")
    if err != nil {
        return fmt.Errorf("failed to assign KisanSathi role: %w", err)
    }

    // ✅ Verify assignment succeeded
    hasRole, err = s.aaaService.CheckUserRole(ctx, userID, "KisanSathi")
    if !hasRole {
        return fmt.Errorf("KisanSathi role assignment verification failed")
    }

    return nil
}
```

#### Analysis

**Strengths (Best Practice Pattern):**
1. **Idempotency Check**: Verifies role exists before assignment (line 584-591)
2. **Post-Assignment Verification**: Confirms role was actually assigned (line 600-606)
3. **Fatal Error Handling**: Returns error if role assignment fails (line 212-214)
4. **Used in Multiple Workflows**:
   - `AssignKisanSathi` (line 211)
   - `ReassignOrRemoveKisanSathi` (line 286)
   - `CreateKisanSathiUser` (line 425)

**Issues:**
1. **Role Name Inconsistency**: Uses `"KisanSathi"` but farmer creation uses `"kisansathi"` (lowercase)
2. **Verification Overhead**: Makes 2 extra AAA calls per assignment (check before + verify after)
3. **No Retry Logic**: If verification fails, doesn't retry assignment
4. **Empty OrgID**: Line 367, 425 pass empty string for orgID (potential issue)

**Recommendation:**
- This pattern should be used for farmer and CEO role assignments
- Consider caching role check results to reduce AAA calls
- Standardize role names across codebase

---

### 1.5 AAA Client Role Assignment

**File:** `/Users/kaushik/farmers-module/internal/clients/aaa/aaa_client.go`

#### Role Assignment Implementation

```go
// Line 610-658: AssignRole method
func (c *Client) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
    log.Printf("AAA AssignRole: userID=%s, orgID=%s, role=%s", userID, orgID, roleName)

    // Input validation
    if userID == "" || orgID == "" || roleName == "" {
        return fmt.Errorf("missing required parameters")
    }

    // ⚠️ ISSUE: Role name validation is too restrictive
    // Line 625-634: Hardcoded valid roles
    validRoles := map[string]bool{
        "admin":       true,
        "farmer":      true,
        "kisansathi":  true,
        "fpo_manager": true,
        "readonly":    true,
    }
    if !validRoles[strings.ToLower(roleName)] {
        return fmt.Errorf("invalid role name: %s", roleName)
    }

    // gRPC call to AAA service
    grpcReq := &pb.AssignRoleRequest{
        UserId:   userID,
        OrgId:    orgID,
        RoleName: roleName,
    }
    response, err := c.roleClient.AssignRole(ctx, grpcReq)

    // Success check
    if response.StatusCode != 200 && response.StatusCode != 201 {
        return fmt.Errorf("failed to assign role: %s", response.Message)
    }

    return nil
}
```

#### Critical Issue: Missing `CEO` Role

**Problem:**
- Line 625-634: Valid roles list does NOT include `"CEO"` or `"fpo_ceo"`
- FPO creation calls `AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")` (line 134 of fpo_ref_service.go)
- **This call will FAIL with "invalid role name: CEO"**

**Impact:**
- FPO creation workflow assigns CEO role, but it's rejected by client validation
- Error is logged as warning and ignored
- CEOs never get proper role assignment

**Fix Required:**
```go
validRoles := map[string]bool{
    "admin":       true,
    "farmer":      true,
    "kisansathi":  true,
    "fpo_manager": true,
    "fpo_ceo":     true,  // Add this
    "CEO":         true,  // Or this (depends on AAA service role name)
    "readonly":    true,
}
```

---

## 2. Architectural Analysis

### 2.1 Role Assignment Strategy

#### Current Patterns Identified

| Pattern | Used In | Error Handling | Idempotency | Verification |
|---------|---------|----------------|-------------|--------------|
| **No Assignment** | Farmer creation | N/A | N/A | N/A |
| **Assign with Warning** | FPO CEO | Non-fatal log | No | No |
| **Ensure with Verification** | KisanSathi | Fatal error | Yes | Yes |

**Recommendation:** Adopt **"Ensure with Verification"** pattern universally

### 2.2 Transaction Management

#### Current State
- Role assignment is NOT part of database transaction
- Failure scenarios:
  1. Local DB write succeeds + AAA role assignment fails → Orphaned entity
  2. AAA role assignment succeeds + Local DB write fails → Orphaned AAA user
  3. Role check fails but assignment succeeds → Inconsistent state

#### Recommendation: Eventual Consistency Model

```go
// Proposed Pattern for Farmer Creation
func (s *FarmerServiceImpl) CreateFarmer(ctx context.Context, req) (*responses.FarmerResponse, error) {
    // Step 1: Create/get AAA user
    aaaUserID, err := s.getOrCreateAAAUser(ctx, req)
    if err != nil {
        return nil, err // Fatal: Cannot proceed without user
    }

    // Step 2: Create local farmer entity (primary operation)
    farmer, err := s.createFarmerEntity(ctx, aaaUserID, req)
    if err != nil {
        return nil, err // Fatal: Primary operation failed
    }

    // Step 3: Assign farmer role (best-effort, eventual consistency)
    roleErr := s.ensureFarmerRole(ctx, aaaUserID, req.AAAOrgID)
    if roleErr != nil {
        log.Printf("Warning: Role assignment failed for farmer %s: %v", farmer.ID, roleErr)
        // Store failure metadata for retry
        farmer.Metadata["role_assignment_error"] = roleErr.Error()
        farmer.Metadata["role_assignment_pending"] = "true"
        s.repository.Update(ctx, farmer)

        // Trigger async retry (optional)
        // s.roleAssignmentQueue.Enqueue(farmer.ID, aaaUserID, "farmer")
    }

    return s.buildResponse(farmer, roleErr), nil
}
```

**Key Principles:**
1. **Primary Operation is Fatal**: Entity creation must succeed
2. **Role Assignment is Best-Effort**: Log failure, store retry metadata
3. **Eventual Consistency**: Implement retry mechanism for failed role assignments
4. **Observability**: Track role assignment status in entity metadata or separate table

### 2.3 Error Handling Strategy

#### Failure Mode Analysis

| Failure Scenario | Current Behavior | Recommended Behavior |
|------------------|------------------|----------------------|
| Role already assigned | Unknown (no check) | Idempotent (succeed silently) |
| Invalid role name | Client-side validation error | Pass to AAA for authoritative check |
| AAA service unavailable | Degraded mode (skip) | Queue for retry, mark entity |
| Permission denied | Error varies by workflow | Log warning, mark for manual review |
| Verification fails | Not checked (except KisanSathi) | Retry once, then mark for review |

#### Proposed Error Taxonomy

```go
type RoleAssignmentError struct {
    UserID    string
    OrgID     string
    RoleName  string
    ErrorType RoleErrorType
    Message   string
    Retryable bool
}

type RoleErrorType int

const (
    RoleErrorAAUnavailable RoleErrorType = iota  // Retryable
    RoleErrorPermissionDenied                    // Manual review
    RoleErrorInvalidRole                         // Configuration error
    RoleErrorVerificationFailed                  // Retry once
    RoleErrorAlreadyAssigned                     // Success (idempotent)
)
```

### 2.4 Permission Granularity Review

#### Current Permission Model (inferred)

**Resources:**
- `farmer` - Farmer entities
- `kisansathi` - KisanSathi assignments
- `fpo` - FPO operations (used in group permissions, line 180 of fpo_ref_service.go)

**Actions:**
- `create`, `read`, `update`, `delete`, `list` - Standard CRUD
- `link`, `unlink` - Farmer-FPO linkage
- `assign`, `reassign` - KisanSathi assignment
- `manage`, `approve`, `inventory`, `reports`, `vote` - FPO group permissions

#### Gap Analysis

**Missing Permissions:**
1. **Farmer Self-Access**: Can farmers read their own profile?
   - Recommended: `farmer:read:self` permission
2. **FPO CEO Operations**: What can CEOs do?
   - Recommended: `farmer:*`, `fpo:manage`, `kisansathi:assign`
3. **Cross-Org Boundaries**: Can users from Org A access Org B data?
   - Current: `orgID` parameter in `CheckPermission` (line 752-825 of aaa_client.go)
   - Recommended: Enforce organization context in all permission checks

**Permission Check Patterns:**

```go
// Line 55-61 of farmer_linkage_service.go: Link farmer permission
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID,
    "farmer", "link", linkReq.AAAUserID, linkReq.AAAOrgID)

// Analysis:
// - subject: Authenticated user ID (from context)
// - resource: "farmer"
// - action: "link"
// - object: Target farmer user ID
// - orgID: Organization context
```

**Recommendation:**
- Document all permission tuples required for each workflow
- Create permission matrix mapping roles to allowed operations
- Implement permission check middleware for all API endpoints

---

## 3. Security Analysis

### 3.1 Threat Model (STRIDE Analysis)

#### Spoofing
- **Threat**: User creates farmer without `farmer` role, then claims farmer privileges
- **Current**: No role assigned → Vulnerable
- **Mitigation**: Mandatory role assignment with verification

#### Tampering
- **Threat**: User modifies farmer entity to bypass role checks
- **Current**: Local DB has no role data (relies on AAA)
- **Mitigation**: Acceptable (AAA is source of truth)

#### Repudiation
- **Threat**: No audit trail of role assignments
- **Current**: AAA audit service integration exists but not verified
- **Mitigation**: Verify AAA logs all role changes

#### Information Disclosure
- **Threat**: User lists farmers across organizations
- **Current**: Permission checks include `orgID` parameter (good)
- **Mitigation**: Ensure all queries filter by authenticated user's org

#### Denial of Service
- **Threat**: AAA service unavailable blocks all operations
- **Current**: Graceful degradation (logs warning, continues)
- **Mitigation**: Implement circuit breaker pattern

#### Elevation of Privilege
- **Threat**: User without CEO role performs CEO operations
- **Current**: Vulnerable (CEO role assignment fails silently)
- **Mitigation**: Fix CEO role validation, enforce role checks

### 3.2 OWASP ASVS Compliance

#### V4: Access Control

| Requirement | Status | Gap |
|-------------|--------|-----|
| 4.1.1: Enforce access control on trusted service layer | ⚠️ Partial | Missing role assignment |
| 4.1.2: All user/data attributes used for access control are in a single, trusted location | ✅ Pass | AAA service is source of truth |
| 4.1.3: Principle of least privilege | ⚠️ Partial | Overly permissive in degraded mode |
| 4.1.5: Access controls fail securely | ❌ Fail | Degraded mode allows operations without role |

#### V4.2: Operation Level Access Control

| Requirement | Status | Gap |
|-------------|--------|-----|
| 4.2.1: Sensitive data/APIs protected with effective access control | ⚠️ Partial | Role assignment gaps |
| 4.2.2: Direct object references protected by access control checks | ✅ Pass | Permission checks include object ID |

### 3.3 Attack Scenarios

#### Scenario 1: Unauthorized Farmer Data Access

**Attack:**
1. Attacker registers as farmer in FPO A
2. Farmer creation succeeds WITHOUT `farmer` role
3. Attacker crafts request to access farmer data in FPO B
4. Permission check: `CheckPermission(attackerID, "farmer", "read", targetFarmerID, orgB)`
5. **Vulnerability**: What if attacker's lack of `farmer` role causes permission check to fail open?

**Current Risk:** Medium (depends on AAA service default-deny policy)

**Mitigation:**
- Assign `farmer` role during creation
- Verify role assignment before returning success
- Ensure AAA service defaults to deny

#### Scenario 2: CEO Privilege Escalation

**Attack:**
1. Attacker creates FPO with themselves as CEO
2. CEO role assignment fails (invalid role name in client)
3. FPO is marked ACTIVE despite CEO role missing
4. Attacker cannot perform CEO operations but FPO exists

**Current Risk:** Low (operations fail, but data is inconsistent)

**Mitigation:**
- Fix `CEO` role in valid roles list
- Make CEO role assignment fatal (prevent FPO activation)

---

## 4. Implementation Gaps

### 4.1 Critical Gaps (P0 - Blocking)

1. **No Farmer Role Assignment**
   - **File**: `internal/services/farmer_service.go`
   - **Line**: After line 204 (after AAA user ID determined)
   - **Fix**: Add `s.ensureFarmerRole(ctx, aaaUserID, aaaOrgID)`

2. **CEO Role Not in Valid Roles List**
   - **File**: `internal/clients/aaa/aaa_client.go`
   - **Line**: 625-634
   - **Fix**: Add `"CEO"` or `"fpo_ceo"` to `validRoles` map

3. **No Role Seeding on Startup**
   - **File**: `cmd/farmers-service/main.go`
   - **Line**: After line 89 (after routes setup)
   - **Fix**: Call `serviceFactory.AAAService.SeedRolesAndPermissions(context.Background())`

### 4.2 High Priority Gaps (P1 - Major Issues)

4. **CEO Role Assignment Non-Fatal**
   - **File**: `internal/services/fpo_ref_service.go`
   - **Line**: 134-138
   - **Fix**: Add to `setupErrors`, mark FPO as `PENDING_SETUP`

5. **Role Assignment Idempotency**
   - **File**: `internal/services/farmer_service.go`, `fpo_ref_service.go`
   - **Fix**: Implement `ensureRole` pattern (check before assign)

6. **No Role Assignment Retry Mechanism**
   - **File**: All services
   - **Fix**: Implement async retry queue or reconciliation job

### 4.3 Medium Priority Gaps (P2 - Improvements)

7. **Role Name Consistency**
   - **Files**: Multiple
   - **Issue**: `"KisanSathi"` vs `"kisansathi"`, `"CEO"` vs `"fpo_ceo"`
   - **Fix**: Standardize to lowercase, document in constants

8. **Verification Overhead**
   - **File**: `internal/services/farmer_linkage_service.go`
   - **Line**: 582-609
   - **Fix**: Cache role check results (with TTL)

9. **Permission Documentation**
   - **Files**: All services
   - **Fix**: Document required permissions for each workflow

### 4.4 Low Priority Gaps (P3 - Nice to Have)

10. **Observability**
    - **Fix**: Add metrics for role assignment success/failure rates
    - **Fix**: Create dashboard for pending role assignments

11. **Integration Tests**
    - **Fix**: Add tests for role assignment failures
    - **Fix**: Test AAA service unavailable scenarios

---

## 5. Proposed Fixes

### 5.1 Fix 1: Farmer Role Assignment

**File:** `/Users/kaushik/farmers-module/internal/services/farmer_service.go`

**Change:** Add role assignment after user creation

```go
// After line 204 (after aaaUserID and aaaOrgID are determined)

// Ensure farmer role is assigned (with verification)
err = s.ensureFarmerRole(ctx, aaaUserID, aaaOrgID)
if err != nil {
    log.Printf("Warning: Failed to assign farmer role to user %s: %v", aaaUserID, err)
    // Store in metadata for retry
    if farmer.Metadata == nil {
        farmer.Metadata = make(entities.JSONB)
    }
    farmer.Metadata["role_assignment_error"] = err.Error()
    farmer.Metadata["role_assignment_pending"] = "true"
    farmer.Metadata["role_assignment_attempted_at"] = time.Now().Format(time.RFC3339)
}
```

**Add Helper Method:**

```go
// ensureFarmerRole ensures the user has the farmer role with idempotency and verification
func (s *FarmerServiceImpl) ensureFarmerRole(ctx context.Context, userID, orgID string) error {
    // Check if role already exists (idempotency)
    hasRole, err := s.aaaService.CheckUserRole(ctx, userID, "farmer")
    if err != nil {
        // If check fails, try assignment anyway (AAA may be temporarily unavailable)
        log.Printf("Warning: Failed to check farmer role for user %s: %v", userID, err)
    } else if hasRole {
        log.Printf("User %s already has farmer role, skipping assignment", userID)
        return nil
    }

    // Assign farmer role
    err = s.aaaService.AssignRole(ctx, userID, orgID, "farmer")
    if err != nil {
        return fmt.Errorf("failed to assign farmer role: %w", err)
    }

    // Verify assignment succeeded
    hasRole, err = s.aaaService.CheckUserRole(ctx, userID, "farmer")
    if err != nil {
        return fmt.Errorf("failed to verify farmer role assignment: %w", err)
    }
    if !hasRole {
        return fmt.Errorf("farmer role assignment verification failed")
    }

    log.Printf("Successfully assigned and verified farmer role for user %s", userID)
    return nil
}
```

### 5.2 Fix 2: CEO Role in Valid Roles List

**File:** `/Users/kaushik/farmers-module/internal/clients/aaa/aaa_client.go`

**Change:** Add CEO role to validation

```go
// Line 625-634: Update validRoles map
validRoles := map[string]bool{
    "admin":       true,
    "farmer":      true,
    "kisansathi":  true,
    "fpo_manager": true,
    "fpo_ceo":     true,  // Add FPO CEO role
    "ceo":         true,  // Add alternative name (case-insensitive check)
    "readonly":    true,
}
```

**Consideration:** Verify with AAA service which role name is canonical

### 5.3 Fix 3: CEO Role Assignment Error Handling

**File:** `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`

**Change:** Treat CEO role assignment failure as critical setup error

```go
// Line 134-138: Update error handling
err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
if err != nil {
    log.Printf("Error: Failed to assign CEO role to user %s: %v", ceoUserID, err)
    setupErrors["ceo_role_assignment"] = err.Error()
    // Don't return error, continue with setup to create FPO in PENDING_SETUP state
}
```

**Rationale:** CEO role is critical, but shouldn't block FPO creation entirely (allow retry via `CompleteFPOSetup`)

### 5.4 Fix 4: Role Seeding on Startup

**File:** `/Users/kaushik/farmers-module/cmd/farmers-service/main.go`

**Change:** Seed roles after service initialization

```go
// After line 89 (after routes.SetupRoutes)

// Seed AAA roles and permissions on startup
log.Println("Seeding AAA roles and permissions...")
seedCtx, seedCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer seedCancel()

if err := serviceFactory.AAAService.SeedRolesAndPermissions(seedCtx); err != nil {
    // Non-fatal: Log warning and continue
    // Roles may already exist, or AAA service may be temporarily unavailable
    log.Printf("Warning: Failed to seed AAA roles and permissions: %v", err)
    log.Println("Application will continue, but role assignments may fail if roles don't exist")
} else {
    log.Println("Successfully seeded AAA roles and permissions")
}
```

**Rationale:** Non-fatal on startup (AAA service may not be ready), but ensures roles exist before operations begin

### 5.5 Fix 5: Role Name Constants

**File:** `/Users/kaushik/farmers-module/internal/constants/roles.go` (new file)

**Create:** Centralized role name constants

```go
package constants

// AAA Role Names
// These must match the role names in the AAA service
const (
    RoleFarmer     = "farmer"       // Assigned to all farmer users
    RoleKisanSathi = "kisansathi"   // Assigned to field agents (note: lowercase)
    RoleFPOCEO     = "CEO"          // Assigned to FPO chief executive officers
    RoleFPOManager = "fpo_manager"  // Assigned to FPO managers
    RoleAdmin      = "admin"        // System administrators
    RoleReadOnly   = "readonly"     // Read-only access
)

// RoleDisplayNames maps internal role names to human-readable names
var RoleDisplayNames = map[string]string{
    RoleFarmer:     "Farmer",
    RoleKisanSathi: "KisanSathi",
    RoleFPOCEO:     "FPO CEO",
    RoleFPOManager: "FPO Manager",
    RoleAdmin:      "Administrator",
    RoleReadOnly:   "Read-Only User",
}
```

**Usage:** Replace all hardcoded role strings with constants

```go
// Before
s.aaaService.AssignRole(ctx, userID, orgID, "farmer")

// After
import "github.com/Kisanlink/farmers-module/internal/constants"
s.aaaService.AssignRole(ctx, userID, orgID, constants.RoleFarmer)
```

---

## 6. Architecture Decision Record (ADR)

### ADR-001: Role Assignment Strategy for Entity Creation

**Status:** Proposed
**Date:** 2025-10-16
**Context:**

The farmers-module creates entities (farmers, FPOs) that map to AAA users and organizations. These users require specific roles (farmer, fpo_ceo, kisansathi) to perform authorized operations. The question is: should role assignment be part of the critical path (fatal error if fails) or best-effort (eventual consistency)?

**Decision:**

Adopt a **hybrid eventual consistency model** with the following rules:

1. **Primary Entity Creation is Fatal**: If user/org creation fails, entire workflow fails
2. **Role Assignment is Best-Effort**: If role assignment fails:
   - Log error with sufficient detail for debugging
   - Store failure metadata in entity (for retry/reconciliation)
   - Mark entity status appropriately (e.g., PENDING_SETUP for FPOs)
   - Return success with warning message to client
   - Implement async retry mechanism (future enhancement)

3. **Idempotency First**: Always check if role exists before assigning
4. **Verification Required**: After assignment, verify role was actually assigned
5. **Graceful Degradation**: If AAA is unavailable, continue with operations (allow eventual reconciliation)

**Consequences:**

**Positive:**
- System remains available even if AAA role service is temporarily down
- Clear separation of concerns (entity creation vs role assignment)
- Enables retry mechanisms and reconciliation processes
- Better observability (can track pending role assignments)

**Negative:**
- Eventual consistency means brief period where entity exists without proper role
- Requires additional metadata storage and monitoring
- More complex error handling and recovery logic

**Alternatives Considered:**

1. **Fatal Role Assignment**: Fail entity creation if role assignment fails
   - Rejected: Too brittle, single point of failure
   - Exception: Use for critical roles (e.g., CEO role for FPO)

2. **No Role Assignment**: Rely on external process to assign roles
   - Rejected: Too fragile, no clear ownership

3. **Synchronous Retry**: Retry role assignment immediately on failure
   - Rejected: Adds latency, doesn't handle extended AAA outages

**Implementation Notes:**
- Add `role_assignment_pending` flag to entity metadata
- Create reconciliation job to retry pending assignments (future)
- Monitor role assignment success rate in observability dashboard

---

### ADR-002: Idempotent Role Assignment Pattern

**Status:** Proposed
**Date:** 2025-10-16
**Context:**

Role assignment operations may be retried (e.g., after network failure, during reconciliation). We need to ensure that retrying role assignment doesn't cause errors or inconsistencies.

**Decision:**

Implement the **"Check-Assign-Verify" pattern** for all role assignments:

```go
func ensureRole(ctx, userID, orgID, roleName string) error {
    // 1. Check: Does user already have the role?
    hasRole, err := CheckUserRole(ctx, userID, roleName)
    if err != nil {
        // Log warning, continue with assignment
    } else if hasRole {
        return nil // Idempotent: already has role
    }

    // 2. Assign: Give user the role
    err = AssignRole(ctx, userID, orgID, roleName)
    if err != nil {
        return fmt.Errorf("assignment failed: %w", err)
    }

    // 3. Verify: Confirm role was assigned
    hasRole, err = CheckUserRole(ctx, userID, roleName)
    if err != nil || !hasRole {
        return fmt.Errorf("verification failed")
    }

    return nil
}
```

**Consequences:**

**Positive:**
- Safe to retry role assignment operations
- Clear verification of success
- Works even if CheckUserRole fails initially

**Negative:**
- 2 additional AAA service calls per assignment
- Higher latency for role assignment operations

**Mitigations:**
- Cache role check results with TTL (future optimization)
- Accept higher latency for correctness guarantee

**Alternatives Considered:**
1. **Assign Without Check**: Rely on AAA service idempotency
   - Rejected: No verification, unclear if AAA guarantees idempotency
2. **Assign Without Verify**: Skip verification step
   - Rejected: No confidence that assignment succeeded

---

### ADR-003: Role Seeding Timing

**Status:** Proposed
**Date:** 2025-10-16
**Context:**

Roles must exist in AAA service before they can be assigned. When should role seeding occur?

**Decision:**

Implement **multi-layered seeding strategy**:

1. **On Application Startup** (Non-Fatal)
   - Attempt to seed roles during main.go initialization
   - Log warning if seeding fails
   - Continue startup (roles may already exist, or AAA may be temporarily down)

2. **On-Demand Seeding** (Future Enhancement)
   - If role assignment fails with "role not found" error, trigger seeding
   - Retry assignment after seeding

3. **Manual Administrative Endpoint** (Required)
   - Provide `/admin/seed-roles` endpoint for manual seeding
   - Use for disaster recovery and initial setup

**Consequences:**

**Positive:**
- Roles are seeded automatically in typical deployment
- Graceful handling of race conditions (startup before AAA ready)
- Manual override available for troubleshooting

**Negative:**
- Startup seeding failure is silent (may go unnoticed)
- Requires monitoring to detect missing roles

**Mitigations:**
- Add health check that verifies all required roles exist
- Alert if role assignment failures spike (indicates missing roles)

---

## 7. Testing Recommendations

### 7.1 Unit Tests Required

1. **Farmer Role Assignment**
   - Test: Farmer created with existing AAA user → Role assigned
   - Test: Farmer created with new AAA user → User created, role assigned
   - Test: Role assignment fails → Farmer created with pending metadata
   - Test: Role already exists → Idempotent (no error)

2. **CEO Role Assignment**
   - Test: FPO created → CEO role assigned
   - Test: CEO role assignment fails → FPO marked PENDING_SETUP
   - Test: User already CEO of another FPO → Error returned

3. **KisanSathi Role Assignment**
   - Test: KisanSathi assigned → Role verified
   - Test: Role verification fails → Error returned
   - Test: KisanSathi reassigned → Old assignment removed

### 7.2 Integration Tests Required

1. **End-to-End Farmer Creation**
   - Test: Create farmer → Verify role in AAA service
   - Test: Create farmer with KisanSathi → Verify both roles

2. **End-to-End FPO Creation**
   - Test: Create FPO → Verify CEO role, user groups, permissions

3. **AAA Service Unavailable**
   - Test: Create farmer when AAA down → Farmer created, role pending
   - Test: Retry role assignment when AAA recovers

### 7.3 Contract Tests Required

1. **AAA Service Role API**
   - Test: AssignRole with valid inputs → Returns success
   - Test: AssignRole with duplicate → Returns success (idempotent)
   - Test: CheckUserRole → Returns accurate result

2. **AAA Service Seeding API**
   - Test: SeedRolesAndPermissions → Creates all required roles
   - Test: Re-seed → Idempotent (no duplicates)

### 7.4 Security Tests Required

1. **Authorization Bypass**
   - Test: User without farmer role cannot access farmer resources
   - Test: User cannot read farmers from different organization

2. **Privilege Escalation**
   - Test: Non-CEO user cannot perform CEO operations
   - Test: KisanSathi cannot assign themselves to farmers

---

## 8. Monitoring and Observability

### 8.1 Metrics to Track

1. **Role Assignment Metrics**
   ```
   farmer_role_assignments_total{status="success|failure"}
   ceo_role_assignments_total{status="success|failure"}
   kisansathi_role_assignments_total{status="success|failure"}
   role_assignment_duration_seconds{role="farmer|ceo|kisansathi"}
   role_assignment_pending_count{role="farmer|ceo|kisansathi"}
   ```

2. **AAA Service Health**
   ```
   aaa_service_available{service="user|role|authz"}
   aaa_request_duration_seconds{endpoint="assign_role|check_role|seed"}
   aaa_request_errors_total{endpoint="...", error_type="timeout|unavailable|denied"}
   ```

3. **Entity Creation Metrics**
   ```
   farmers_created_total{status="success|partial|failure"}
   fpos_created_total{status="active|pending_setup|failed"}
   kisansathi_assignments_total{status="success|failure"}
   ```

### 8.2 Alerts to Configure

1. **High Priority Alerts**
   - Role assignment failure rate > 5% (15 min window)
   - AAA service unavailable > 5 minutes
   - Pending role assignments > 100 entities

2. **Medium Priority Alerts**
   - Role seeding failed on startup
   - FPO in PENDING_SETUP for > 24 hours
   - Role verification failures > 10% (1 hour window)

### 8.3 Logging Requirements

1. **Structured Log Fields**
   ```go
   log.Printf("Role assignment attempt: user_id=%s, org_id=%s, role=%s, operation=assign",
       userID, orgID, roleName)
   log.Printf("Role assignment success: user_id=%s, org_id=%s, role=%s, duration_ms=%d",
       userID, orgID, roleName, duration)
   log.Printf("Role assignment failure: user_id=%s, org_id=%s, role=%s, error=%s",
       userID, orgID, roleName, err)
   ```

2. **Correlation IDs**
   - Include request ID in all AAA service calls
   - Track role assignment across retries with same correlation ID

---

## 9. Migration Plan

### 9.1 Phase 1: Fix Critical Gaps (Week 1)

**Goal:** Resolve blocking issues preventing proper role assignment

**Tasks:**
1. Add CEO role to valid roles list (aaa_client.go)
2. Implement farmer role assignment (farmer_service.go)
3. Update CEO role assignment error handling (fpo_ref_service.go)
4. Add role seeding on startup (main.go)
5. Deploy to development environment

**Success Criteria:**
- All new farmers get `farmer` role
- All new FPOs have CEOs with `CEO` role
- No errors in startup logs related to role seeding

### 9.2 Phase 2: Add Idempotency (Week 2)

**Goal:** Make role assignment safe to retry

**Tasks:**
1. Create `ensureFarmerRole` helper method
2. Refactor KisanSathi pattern into reusable helper
3. Add role constants file
4. Replace hardcoded strings with constants
5. Add unit tests for idempotency

**Success Criteria:**
- Role assignment can be safely retried
- All role operations are idempotent
- Test coverage > 80%

### 9.3 Phase 3: Data Migration (Week 3)

**Goal:** Assign roles to existing entities

**Tasks:**
1. Create migration script to find entities without roles
2. Run dry-run migration in staging
3. Execute migration in production (off-peak hours)
4. Verify all entities have correct roles

**Migration Script:**
```go
func MigrateExistingFarmers(ctx context.Context) error {
    // 1. Find all farmers without farmer role
    farmers, err := getAllFarmers(ctx)

    for _, farmer := range farmers {
        hasRole, _ := aaaService.CheckUserRole(ctx, farmer.AAAUserID, "farmer")
        if !hasRole {
            log.Printf("Migrating farmer %s (user %s)", farmer.ID, farmer.AAAUserID)
            err := aaaService.AssignRole(ctx, farmer.AAAUserID, farmer.AAAOrgID, "farmer")
            if err != nil {
                log.Printf("Migration failed for farmer %s: %v", farmer.ID, err)
                // Store in migration_failures table for manual review
            }
        }
    }

    return nil
}
```

**Success Criteria:**
- 100% of farmers have `farmer` role
- 100% of FPO CEOs have `CEO` role
- Migration log reviewed for failures

### 9.4 Phase 4: Monitoring and Alerts (Week 4)

**Goal:** Detect and respond to role assignment issues

**Tasks:**
1. Add Prometheus metrics (section 8.1)
2. Configure Grafana dashboards
3. Set up alerts (section 8.2)
4. Create runbook for role assignment failures

**Success Criteria:**
- All metrics are collected and visible in Grafana
- Alerts trigger correctly in test scenarios
- Runbook tested by on-call team

### 9.5 Phase 5: Eventual Consistency (Future)

**Goal:** Implement retry mechanism for failed role assignments

**Tasks:**
1. Create role_assignment_queue table
2. Implement background worker to process queue
3. Add retry logic with exponential backoff
4. Implement reconciliation job (daily)

**Success Criteria:**
- Failed role assignments are automatically retried
- Retry success rate > 95%
- Pending role assignments < 1% of total entities

---

## 10. Summary and Recommendations

### 10.1 Executive Summary

**Current State:**
- Role seeding is implemented but NEVER called
- Farmer creation does NOT assign `farmer` role (critical security gap)
- FPO CEO role assignment fails silently due to validation bug
- KisanSathi role assignment works correctly (best practice pattern)

**Risk Level:** **HIGH** - Security and operational risks due to missing role assignments

**Impact:**
- Farmers cannot be authorized for farmer-specific operations
- Permission checks may fail or allow unauthorized access
- Inconsistent authorization state across system

### 10.2 Immediate Actions Required (Priority 0)

1. **Fix CEO Role Validation** (1 hour)
   - Add `"CEO"` to valid roles list in `aaa_client.go`
   - Deploy immediately to prevent FPO creation failures

2. **Implement Farmer Role Assignment** (4 hours)
   - Add `ensureFarmerRole` method to `farmer_service.go`
   - Call during farmer creation workflow
   - Add unit tests

3. **Enable Role Seeding on Startup** (2 hours)
   - Add seeding call to `main.go`
   - Make it non-fatal (log warning if fails)
   - Test in development environment

4. **Deploy and Monitor** (2 hours)
   - Deploy fixes to staging
   - Verify role assignments in AAA service
   - Deploy to production with monitoring

**Total Estimated Time:** 1 business day

### 10.3 Short-Term Actions (Priority 1 - Within 2 Weeks)

5. **Implement Idempotent Role Assignment Pattern**
   - Extract `ensureRole` helper method
   - Apply pattern to all role assignments
   - Add verification step

6. **Create Role Constants File**
   - Centralize role name definitions
   - Replace hardcoded strings
   - Document role purpose and permissions

7. **Add Integration Tests**
   - Test end-to-end role assignment
   - Test AAA service unavailable scenario
   - Test idempotency

8. **Migrate Existing Data**
   - Create migration script
   - Assign roles to existing farmers and FPOs
   - Validate migration results

### 10.4 Medium-Term Actions (Priority 2 - Within 1 Month)

9. **Implement Monitoring and Alerts**
   - Add Prometheus metrics
   - Create Grafana dashboards
   - Configure PagerDuty alerts

10. **Document Permission Model**
    - Create permission matrix (roles → operations)
    - Document required permissions for each workflow
    - Update API documentation

11. **Improve Error Handling**
    - Standardize error responses
    - Add error metadata (retryable, etc.)
    - Implement circuit breaker for AAA calls

### 10.5 Long-Term Actions (Priority 3 - Future)

12. **Implement Eventual Consistency**
    - Create retry queue for failed role assignments
    - Implement background reconciliation job
    - Add admin dashboard for pending assignments

13. **Add Role Assignment Audit Trail**
    - Log all role assignment attempts
    - Create audit report API
    - Integrate with centralized audit service

14. **Optimize Performance**
    - Cache role check results
    - Batch role assignments
    - Reduce AAA service call overhead

### 10.6 Key Architectural Decisions

1. **Adopt Eventual Consistency Model**
   - Role assignment failures should not block entity creation
   - Store failure metadata for retry
   - Implement async retry mechanism

2. **Use "Check-Assign-Verify" Pattern**
   - Always check if role exists before assigning (idempotency)
   - Verify role was actually assigned (reliability)
   - Accept 2 extra AAA calls for correctness guarantee

3. **Centralize Role Definitions**
   - Create constants file for role names
   - Document role purpose and permissions
   - Enforce usage via linting

4. **Prioritize Security Over Availability**
   - Fail secure (deny by default)
   - Don't skip permission checks even in degraded mode
   - Exception: Role assignment can fail gracefully (eventual consistency)

---

## 11. Appendices

### Appendix A: Role and Permission Matrix

| Role | Resource | Actions | Scope |
|------|----------|---------|-------|
| `farmer` | farmer (self) | read, update | Own profile |
| `farmer` | farm (own) | create, read, update, delete | Linked to own profile |
| `farmer` | crop_cycle (own) | create, read, update, end | Linked to own farms |
| `kisansathi` | farmer (assigned) | read, update | Farmers assigned to them |
| `kisansathi` | farm (assigned) | read, update | Farms of assigned farmers |
| `kisansathi` | crop_cycle (assigned) | read, update | Cycles of assigned farmers |
| `CEO` | farmer (org) | create, read, update, delete | All farmers in organization |
| `CEO` | fpo (own) | manage | Own organization |
| `CEO` | kisansathi (org) | assign, reassign, remove | Organization KisanSathis |
| `fpo_manager` | farmer (org) | read | All farmers in organization |
| `fpo_manager` | inventory (org) | create, read, update | Organization inventory |
| `admin` | * | * | All resources |

**Notes:**
- (self) = User's own resources
- (own) = Resources they created
- (assigned) = Resources assigned to them by manager/CEO
- (org) = All resources within their organization
- (*) = All resources system-wide

### Appendix B: AAA Service API Contract

**AssignRole**
```protobuf
message AssignRoleRequest {
    string user_id = 1;      // Required: User to assign role to
    string org_id = 2;       // Required: Organization context
    string role_name = 3;    // Required: Role to assign
}

message AssignRoleResponse {
    int32 status_code = 1;   // 200 = success, 201 = created, 4xx/5xx = error
    string message = 2;      // Human-readable message
}

// Idempotency: Calling twice with same parameters should succeed both times
// Authorization: Caller must have permission to assign roles in target org
```

**CheckUserRole**
```protobuf
message GetUserRequest {
    string id = 1;           // User ID to check
}

message GetUserResponse {
    int32 status_code = 1;
    User user = 2;
    string message = 3;
}

message User {
    string id = 1;
    string username = 2;
    repeated UserRoleV2 user_roles = 3;  // All roles for user
}

message UserRoleV2 {
    string role_name = 1;    // Name of role (e.g., "farmer", "CEO")
    string org_id = 2;       // Organization context
}

// Check logic: Iterate user_roles, return true if role_name matches
```

**SeedRolesAndPermissions**
```protobuf
message SeedRolesAndPermissionsRequest {
    bool force = 1;          // If true, re-seed even if data exists
}

message SeedRolesAndPermissionsResponse {
    int32 status_code = 1;
    string message = 2;
    int32 roles_created = 3;       // Number of new roles created
    int32 permissions_created = 4; // Number of new permissions created
}

// Idempotency: Safe to call multiple times (force=false will skip if exists)
// Required Roles: Should seed at minimum: admin, farmer, kisansathi, fpo_ceo
```

### Appendix C: File Change Summary

| File | Change Type | Lines Changed | Priority |
|------|-------------|---------------|----------|
| `internal/services/farmer_service.go` | Add method + call | ~30 | P0 |
| `internal/clients/aaa/aaa_client.go` | Update map | ~3 | P0 |
| `internal/services/fpo_ref_service.go` | Update error handling | ~5 | P0 |
| `cmd/farmers-service/main.go` | Add seeding call | ~10 | P0 |
| `internal/constants/roles.go` | New file | ~30 | P1 |
| `internal/services/farmer_linkage_service.go` | Extract helper | ~20 | P1 |
| Various | Replace strings with constants | ~50 | P1 |

**Total Estimated Changes:** ~150 lines of production code, ~200 lines of test code

### Appendix D: References

- [OWASP ASVS v4.0](https://owasp.org/www-project-application-security-verification-standard/)
- [STRIDE Threat Model](https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats)
- [Farmers Module Product Spec](./.kiro/steering/product.md)
- [Farmers Module Tech Stack](./.kiro/steering/tech.md)
- [AAA Service Integration Spec](./.kiro/specs/aaa-permission-check-fix.md)

---

**Document Version:** 1.0
**Last Updated:** 2025-10-16
**Next Review:** After P0 fixes deployed
**Owner:** Backend Architecture Team
**Reviewers:** Security Team, Product Team, DevOps Team
