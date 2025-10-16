# Role Assignment Edge Cases Catalog

**Purpose:** Comprehensive catalog of edge cases, failure scenarios, and abuse paths for role assignment logic

**Last Updated:** 2025-10-16

---

## Category 1: Role Assignment Failures

### EC-001: Role Assignment Fails After User Creation

**Scenario:**
- AAA user created successfully
- Role assignment call to AAA fails (network error, AAA service down)
- Local entity (farmer/FPO) creation pending

**Current Behavior:**
- Farmer: User exists in AAA without role, farmer profile MAY be created (BUG)
- FPO: Organization exists, CEO has no role, FPO marked PENDING_SETUP (partial handling)

**Expected Behavior:**
- Transaction rolled back
- AAA user deleted or marked for cleanup
- Entity creation fails with clear error message
- Audit log records failure with user ID for manual cleanup

**Test Cases:**
```go
TestCreateFarmer_RoleAssignmentNetworkFailure()
TestCreateFarmer_RoleAssignmentAAAServiceDown()
TestCreateFPO_CEORoleAssignmentTimeout()
```

**Risk Level:** HIGH - Leaves users in invalid state
**Exploitability:** MEDIUM - Timing-based attack possible

---

### EC-002: Partial Role Assignment in Bulk Import

**Scenario:**
- Bulk import of 1000 farmers
- AAA service degraded (slow responses, intermittent failures)
- 800 farmers get role assigned
- 200 farmers created without roles

**Current Behavior:**
- No role assignment in bulk pipeline (BUG)
- All farmers created without roles
- Bulk operation reports success

**Expected Behavior:**
- RoleAssignmentStage fails for 200 farmers
- Those 200 marked as FAILED in processing_details
- Bulk operation status shows partial success
- Failed farmers can be retried

**Test Cases:**
```go
TestBulkImport_PartialRoleFailures()
TestBulkImport_AllRoleFailures()
TestBulkRetry_SucceedsOnSecondAttempt()
```

**Risk Level:** CRITICAL - Mass creation of invalid users
**Exploitability:** LOW - Requires AAA service degradation

---

### EC-003: Role Assignment Succeeds But Local Entity Fails

**Scenario:**
- AAA user created
- Role assigned successfully
- Local database operation fails (constraint violation, disk full)
- User exists in AAA with role but no local entity

**Current Behavior:**
- Not explicitly handled
- Orphaned AAA user with role

**Expected Behavior:**
- Detect local entity creation failure
- Attempt to rollback role assignment
- If rollback fails, mark user for cleanup
- Log incident for manual intervention

**Test Cases:**
```go
TestCreateFarmer_LocalDBFailureAfterRole()
TestCreateFarmer_RollbackRoleOnEntityFailure()
```

**Risk Level:** MEDIUM - Creates inconsistency
**Exploitability:** LOW - Requires database failure

---

## Category 2: Concurrent Operations

### EC-004: Race Condition on Farmer Creation

**Scenario:**
- Two API requests create farmer with same phone simultaneously
- Request A: Checks AAA → no user found
- Request B: Checks AAA → no user found
- Request A: Creates AAA user
- Request B: Creates AAA user (fails with conflict)
- Request B: Falls back to "get existing user"
- Both try to create local farmer profile

**Current Behavior:**
- Handled by idempotency in farmer_service.go
- One succeeds, one returns existing farmer
- BUT: Neither assigns farmer role (BUG)

**Expected Behavior:**
- One request succeeds completely (user + role + profile)
- Other request detects existing, verifies role, returns existing farmer
- If existing farmer lacks role, assign role before returning

**Test Cases:**
```go
TestConcurrentFarmerCreation_SamePhone_BothSucceedWithRole()
TestConcurrentFarmerCreation_ExistingUserGetsRole()
```

**Risk Level:** MEDIUM - Common in production
**Exploitability:** LOW - Not malicious, but creates broken state

---

### EC-005: CEO Role Race Condition

**Scenario:**
- User U tries to become CEO of FPO-A and FPO-B simultaneously
- FPO-A creation: Checks if U is CEO → false
- FPO-B creation: Checks if U is CEO → false
- FPO-A creation: Assigns CEO role to U
- FPO-B creation: Assigns CEO role to U
- Both succeed, U is CEO of 2 FPOs

**Current Behavior:**
- CheckUserRole called before organization creation (line 103)
- Race window exists between check and assignment
- May allow multiple CEO roles (depends on AAA implementation)

**Expected Behavior:**
- Use optimistic locking or exclusive role assignment
- AAA service rejects second CEO role assignment
- FPO-B creation fails with clear error

**Test Cases:**
```go
TestConcurrentFPOCreation_SameCEO_OnlyOneSucceeds()
TestCEORoleExclusivity_PreventsDuplicateCEOs()
```

**Risk Level:** HIGH - Violates business rule
**Exploitability:** MEDIUM - Attacker could time requests

---

### EC-006: Parallel Bulk Import of Duplicate Farmers

**Scenario:**
- Upload CSV file with farmer phone=9876543210 at row 5 and row 105
- Parallel processing: Chunk 1 (rows 1-100) and Chunk 2 (rows 101-200)
- Both chunks process the duplicate phone simultaneously

**Current Behavior:**
- Deduplication stage exists but is STUBBED (line 208 in stages.go)
- Both chunks create AAA user
- One succeeds, one fails

**Expected Behavior:**
- Deduplication stage checks AAA service
- First instance proceeds
- Second instance detected as duplicate, skipped

**Test Cases:**
```go
TestBulkImport_DuplicateInFile_OnlyOneCreated()
TestBulkImport_DeduplicationAcrossChunks()
```

**Risk Level:** MEDIUM - Common data issue
**Exploitability:** LOW - Not malicious

---

## Category 3: Data Inconsistency

### EC-007: Farmer Deleted But Role Remains

**Scenario:**
- Farmer profile soft-deleted (deleted_at set)
- AAA user still exists
- Farmer role still assigned in AAA
- User can still authenticate

**Current Behavior:**
- No cleanup of AAA role on soft delete
- User retains farmer permissions
- Queries exclude soft-deleted farmers, but AAA doesn't know

**Expected Behavior:**
- On soft delete, remove farmer role from AAA OR
- Mark role as "inactive" in AAA
- On hard delete (GDPR), remove role and user

**Test Cases:**
```go
TestFarmerSoftDelete_RemovesRole()
TestFarmerHardDelete_RemovesUserAndRole()
TestDeletedFarmer_CannotAuthenticate()
```

**Risk Level:** MEDIUM - Security concern
**Exploitability:** MEDIUM - Deleted users retain access

---

### EC-008: AAA User Deleted Externally

**Scenario:**
- Admin uses AAA admin panel to delete user
- Farmer profile still exists in farmers-module
- Farmer's farms, crop cycles still exist
- System tries to query AAA for deleted user

**Current Behavior:**
- GetUser calls fail with "user not found"
- Operations referencing this farmer fail
- No automatic detection or cleanup

**Expected Behavior:**
- Reconciliation job detects orphaned farmer profiles
- Marks farmer as ORPHANED
- Alerts admin for manual intervention
- Optionally auto-cleanup if configured

**Test Cases:**
```go
TestReconciliation_DetectsOrphanedFarmers()
TestReconciliation_MarksOrphanedFarmersForCleanup()
```

**Risk Level:** LOW - Rare scenario
**Exploitability:** NONE - External admin action

---

### EC-009: Role Exists But User Has Wrong Organization

**Scenario:**
- Farmer F created in Org A with role
- Admin manually changes farmer's org to Org B in AAA
- Farmer profile still references Org A
- Farmer role still scoped to Org A

**Current Behavior:**
- Not detected
- Permission checks may fail or succeed incorrectly
- Data inconsistency

**Expected Behavior:**
- Reconciliation detects org mismatch
- Alerts admin
- Optionally realigns role to correct org

**Test Cases:**
```go
TestReconciliation_DetectsOrgMismatch()
TestReconciliation_AlignsRoleToCorrectOrg()
```

**Risk Level:** LOW - Admin error, rare
**Exploitability:** NONE

---

## Category 4: Multi-Role Scenarios

### EC-010: User Is Both Farmer and KisanSathi

**Scenario:**
- User U is farmer in FPO-A
- User U becomes KisanSathi in FPO-A (same org)
- User now has both roles

**Current Behavior:**
- Roles are additive, both assigned
- Permission checks: unclear which role takes precedence

**Expected Behavior:**
- **NEEDS BUSINESS RULE CLARIFICATION**
- Option 1: Allow multi-role, use role hierarchy (KisanSathi > Farmer)
- Option 2: Forbid multi-role in same org
- Option 3: Require explicit role switching

**Test Cases:**
```go
TestMultiRole_FarmerAndKisanSathi_AllowedOrForbidden()
TestMultiRole_PermissionResolution()
```

**Risk Level:** MEDIUM - Unclear business logic
**Exploitability:** LOW - Legitimate use case

---

### EC-011: User Is Farmer in Org A, CEO in Org B

**Scenario:**
- User is farmer in FPO-A
- User becomes CEO of FPO-B
- User accesses farmer endpoints with CEO credentials

**Current Behavior:**
- Both roles assigned, org-scoped
- Permission checks should scope to org
- Potential for cross-org access if not properly scoped

**Expected Behavior:**
- Multi-org roles allowed
- Permission checks MUST include org context
- User can switch org context via API

**Test Cases:**
```go
TestMultiOrg_FarmerInAandCEOInB_CorrectPermissions()
TestMultiOrg_CannotAccessOrgADataWithOrgBCredentials()
```

**Risk Level:** HIGH - Authorization bypass risk
**Exploitability:** HIGH - Attacker could manipulate org context

---

### EC-012: CEO Replaced, Old CEO Retains Role

**Scenario:**
- FPO has CEO = User A
- Admin changes CEO to User B via AAA admin panel
- Organization updated in AAA
- BUT: User A still has CEO role in AAA (if role not auto-removed)

**Current Behavior:**
- Not handled by farmers-module
- Depends on AAA service behavior

**Expected Behavior:**
- When CEO changes, old CEO's CEO role removed
- New CEO role assigned
- Or: AAA service enforces single CEO per org

**Test Cases:**
```go
TestCEOChange_OldCEOLosesRole()
TestCEOChange_NewCEOGetsRole()
TestCEOChange_OnlyOneCEOActive()
```

**Risk Level:** MEDIUM - Access control issue
**Exploitability:** LOW - Requires AAA admin action

---

## Category 5: Permission Bypass Attacks

### EC-013: Role Assigned to Wrong Organization

**Scenario:**
- Attacker creates farmer in Org A
- Attacker manipulates request to assign farmer role in Org B
- Farmer profile created in Org A but role scoped to Org B

**Current Behavior:**
- AssignRole accepts orgID parameter
- No validation that orgID matches farmer's org

**Expected Behavior:**
- Validate that role's orgID matches entity's orgID
- Reject mismatched assignments

**Test Cases:**
```go
TestRoleAssignment_OrgMismatch_Rejected()
TestRoleAssignment_CorrectOrgRequired()
```

**Risk Level:** HIGH - Authorization bypass
**Exploitability:** HIGH - Direct API manipulation

---

### EC-014: Replaying Role Assignment Request

**Scenario:**
- Attacker intercepts role assignment request
- Request: AssignRole(userID=U, orgID=O, role=CEO)
- Attacker replays request with different user
- Request: AssignRole(userID=ATTACKER, orgID=O, role=CEO)

**Current Behavior:**
- No idempotency tokens
- No request signing
- May succeed if attacker has credentials

**Expected Behavior:**
- Role assignment requires authenticated admin
- Idempotency tokens prevent replay
- Audit log tracks all role changes

**Test Cases:**
```go
TestRoleAssignment_RequiresAuthentication()
TestRoleAssignment_IdempotencyTokenPreventsReplay()
```

**Risk Level:** HIGH - Privilege escalation
**Exploitability:** MEDIUM - Requires intercepted credentials

---

### EC-015: No Role Check Before Entity Creation

**Scenario:**
- Farmer creation flow:
  1. Create AAA user
  2. Create local farmer profile
  3. Link to FPO
- NO role check at any step
- Farmer exists without proper permissions

**Current Behavior:**
- Exactly what happens now (BUG)

**Expected Behavior:**
- Role assignment is part of creation transaction
- Entity not considered "created" until role assigned

**Test Cases:**
```go
TestFarmerCreation_FailsIfRoleNotAssigned()
TestFarmerCreation_TransactionalRoleAssignment()
```

**Risk Level:** CRITICAL - Core bug
**Exploitability:** HIGH - All farmers affected

---

## Category 6: Bulk Operations Specific

### EC-016: Bulk Import Aborted Mid-Processing

**Scenario:**
- Bulk import of 1000 farmers
- 500 processed successfully
- Server crashes / process killed
- 500 farmers in unknown state

**Current Behavior:**
- Goroutines may continue in background
- No transaction management across bulk
- Some farmers may be partially created

**Expected Behavior:**
- Bulk operation status persists
- Can resume from last successful index
- Partial farmers marked for cleanup

**Test Cases:**
```go
TestBulkImport_Resumable()
TestBulkImport_CrashRecovery()
```

**Risk Level:** MEDIUM - Data corruption
**Exploitability:** LOW - Operational issue

---

### EC-017: Bulk Import Retry Creates Duplicates

**Scenario:**
- Bulk import fails at 80% completion
- User retries entire operation
- First 80% already created
- Retry creates duplicates

**Current Behavior:**
- Retry operation creates new bulk_operation record
- May attempt to recreate farmers
- Deduplication stage should catch (but it's stubbed)

**Expected Behavior:**
- Retry uses same operation ID
- Skips already-successful farmers
- Only retries failed/pending farmers

**Test Cases:**
```go
TestBulkRetry_SkipsSuccessfulFarmers()
TestBulkRetry_OnlyRetriesFailures()
```

**Risk Level:** MEDIUM - Duplicate data
**Exploitability:** LOW - User error

---

### EC-018: Bulk Import File Contains Malicious Data

**Scenario:**
- CSV file contains SQL injection in farmer name
- File contains script tags in email
- File contains extremely long strings
- File contains invalid characters

**Current Behavior:**
- ValidationStage checks required fields and formats
- May not sanitize malicious content
- GORM should escape SQL
- But XSS risk if data displayed without escaping

**Expected Behavior:**
- Strict input validation
- Sanitization of all string fields
- Length limits enforced
- Character set validation

**Test Cases:**
```go
TestBulkImport_SQLInjection_Rejected()
TestBulkImport_XSSPayload_Sanitized()
TestBulkImport_OversizeStrings_Truncated()
```

**Risk Level:** MEDIUM - Security vulnerability
**Exploitability:** MEDIUM - Standard injection attack

---

## Category 7: AAA Service Integration

### EC-019: AAA Service Returns Unexpected Data

**Scenario:**
- AssignRole call to AAA returns 200 OK
- But response body is empty or malformed
- Farmers-module assumes role assigned
- Role actually NOT assigned in AAA

**Current Behavior:**
- May not validate response structure
- Assumes success based on HTTP status

**Expected Behavior:**
- Validate response body
- Verify role assignment via CheckUserRole
- Retry if validation fails

**Test Cases:**
```go
TestRoleAssignment_AAA_InvalidResponse_Retried()
TestRoleAssignment_AAA_VerifiesAssignment()
```

**Risk Level:** MEDIUM - Silent failure
**Exploitability:** LOW - AAA bug required

---

### EC-020: AAA Service Has Eventual Consistency

**Scenario:**
- AssignRole returns success
- Immediately call CheckUserRole
- Returns false (role not yet propagated)
- Entity creation fails despite role actually assigned

**Current Behavior:**
- No retry logic for consistency checks
- May fail unnecessarily

**Expected Behavior:**
- Retry CheckUserRole with exponential backoff
- Timeout after reasonable period
- Log eventual consistency delays

**Test Cases:**
```go
TestRoleAssignment_AAA_EventualConsistency_Retried()
TestRoleAssignment_AAA_ConsistencyTimeout()
```

**Risk Level:** LOW - Operational issue
**Exploitability:** NONE

---

### EC-021: AAA Service Rate Limits Requests

**Scenario:**
- Bulk import of 10,000 farmers
- Each farmer requires 2 AAA calls (CreateUser, AssignRole)
- 20,000 requests to AAA
- AAA rate limit: 100 req/sec
- Bulk import fails with 429 Too Many Requests

**Current Behavior:**
- No rate limiting awareness
- Fails fast on 429 errors
- Entire bulk operation may fail

**Expected Behavior:**
- Respect AAA rate limits
- Backoff on 429 errors
- Throttle bulk processing if needed
- Estimate completion time considering rate limits

**Test Cases:**
```go
TestBulkImport_AAA_RateLimited_Backoff()
TestBulkImport_AAA_ThrottlesRequests()
```

**Risk Level:** MEDIUM - Bulk operations fail
**Exploitability:** NONE

---

## Category 8: Audit and Compliance

### EC-022: Role Assignment Without Audit Trail

**Scenario:**
- Role assigned to user
- No audit log of who assigned, when, why
- Compliance violation (SOC2, GDPR)

**Current Behavior:**
- AAA service may log role changes
- Farmers-module does not log role assignments
- No correlation between entity creation and role assignment

**Expected Behavior:**
- All role assignments logged with:
  - Actor (who assigned)
  - Subject (user receiving role)
  - Resource (organization)
  - Timestamp
  - Reason (entity creation, manual assignment, etc.)
  - Correlation ID (link to entity creation)

**Test Cases:**
```go
TestRoleAssignment_CreatesAuditLog()
TestAuditLog_IncludesCorrelationID()
```

**Risk Level:** MEDIUM - Compliance issue
**Exploitability:** NONE

---

### EC-023: GDPR Right to be Forgotten

**Scenario:**
- User requests data deletion (GDPR)
- Farmer profile hard-deleted
- AAA user and roles should also be deleted
- But role deletion may fail

**Current Behavior:**
- Not implemented
- Soft delete does not remove AAA data
- Hard delete flow unclear

**Expected Behavior:**
- GDPR deletion workflow:
  1. Delete farmer profile
  2. Delete all related entities (farms, cycles, activities)
  3. Remove all roles from AAA
  4. Delete AAA user
  5. Log deletion in immutable audit log (without PII)
- Partial failures handled gracefully

**Test Cases:**
```go
TestGDPR_Deletion_RemovesAllData()
TestGDPR_Deletion_PartialFailureHandled()
```

**Risk Level:** HIGH - Legal compliance
**Exploitability:** NONE

---

## Category 9: Business Logic Violations

### EC-024: Farmer Created Without FPO Linkage

**Scenario:**
- CreateFarmer called with valid data
- Farmer profile created
- No FPO linkage specified
- Farmer exists in system but not linked to any FPO

**Current Behavior:**
- Allowed (aaa_org_id can be any org)
- Farmer not required to be linked to FPO
- Business rule unclear

**Expected Behavior:**
- **NEEDS BUSINESS RULE CLARIFICATION**
- Option 1: Farmers MUST be linked to FPO
- Option 2: Farmers can exist without FPO (independent farmers)
- Option 3: Farmers created in PENDING status until linked

**Test Cases:**
```go
TestFarmerCreation_WithoutFPO_AllowedOrForbidden()
```

**Risk Level:** MEDIUM - Business logic issue
**Exploitability:** NONE

---

### EC-025: KisanSathi Assigned to Farmer in Different Org

**Scenario:**
- Farmer F in Org A
- KisanSathi K in Org B
- System allows K to be assigned to F
- Cross-org KisanSathi assignment

**Current Behavior:**
- Not validated
- May be allowed

**Expected Behavior:**
- **NEEDS BUSINESS RULE CLARIFICATION**
- Option 1: KisanSathi must be in same org as farmer
- Option 2: KisanSathi can serve farmers across orgs

**Test Cases:**
```go
TestKisanSathi_CrossOrg_AllowedOrForbidden()
```

**Risk Level:** MEDIUM - Business logic issue
**Exploitability:** LOW

---

## Edge Case Priority Matrix

| Edge Case ID | Category | Risk Level | Exploitability | Priority | Notes |
|--------------|----------|------------|----------------|----------|-------|
| EC-001 | Role Failures | HIGH | MEDIUM | P0 | Critical path failure |
| EC-002 | Role Failures | CRITICAL | LOW | P0 | Bulk operations broken |
| EC-015 | Permission Bypass | CRITICAL | HIGH | P0 | Core bug, all farmers affected |
| EC-005 | Concurrency | HIGH | MEDIUM | P1 | Business rule violation |
| EC-011 | Multi-Role | HIGH | HIGH | P1 | Authorization bypass risk |
| EC-013 | Permission Bypass | HIGH | HIGH | P1 | Org mismatch attack |
| EC-007 | Data Inconsistency | MEDIUM | MEDIUM | P2 | Security concern |
| EC-003 | Role Failures | MEDIUM | LOW | P2 | Orphaned data |
| EC-004 | Concurrency | MEDIUM | LOW | P2 | Common scenario |
| EC-010 | Multi-Role | MEDIUM | LOW | P2 | Needs business rule |
| EC-024 | Business Logic | MEDIUM | NONE | P2 | Needs clarification |
| EC-006 | Concurrency | MEDIUM | LOW | P3 | Data quality issue |
| EC-016 | Bulk Operations | MEDIUM | LOW | P3 | Operational robustness |
| EC-019 | AAA Integration | MEDIUM | LOW | P3 | Silent failure |
| EC-008 | Data Inconsistency | LOW | NONE | P4 | Rare scenario |
| EC-020 | AAA Integration | LOW | NONE | P4 | Operational issue |

**Priority Definitions:**
- **P0 (Critical):** Security vulnerability or critical business logic failure
- **P1 (High):** Important business rule or authorization issue
- **P2 (Medium):** Data quality or operational issue
- **P3 (Low):** Edge case or rare scenario
- **P4 (Nice-to-have):** Enhancement or defensive coding

---

## Recommended Actions by Priority

### P0 - Immediate (This Sprint)
1. Fix EC-001, EC-002, EC-015: Implement role assignment in all creation paths
2. Add transactional rollback for role assignment failures
3. Add role assignment to bulk import pipeline

### P1 - Next Sprint
4. Fix EC-005: Implement exclusive CEO role assignment
5. Fix EC-011, EC-013: Add org-scoped permission validation
6. Add comprehensive authorization tests

### P2 - Within 4 Weeks
7. Implement reconciliation job (EC-007, EC-008)
8. Add multi-role business rules (EC-010, EC-024)
9. Enhance bulk deduplication (EC-006)

### P3 - Backlog
10. Add resumable bulk operations (EC-016)
11. Add AAA response validation (EC-019)
12. Implement rate limiting awareness (EC-021)

### P4 - Future
13. GDPR compliance workflow (EC-023)
14. Eventual consistency handling (EC-020)
15. Advanced audit logging (EC-022)

---

**End of Edge Cases Catalog**
