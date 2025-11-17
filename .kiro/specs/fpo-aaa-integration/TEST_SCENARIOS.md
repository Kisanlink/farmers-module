# FPO Creation Test Scenarios

## Overview

Comprehensive test scenarios for FPO creation with AAA service integration, covering happy paths, edge cases, error conditions, and recovery scenarios.

## Test Environment Setup

### Prerequisites
- AAA service accessible and healthy
- Database with migrations applied
- Valid JWT token with FPO creation permissions
- Test user accounts in AAA service

### Test Data

```json
{
  "valid_fpo": {
    "name": "Test Farmers Producer Company",
    "registration_number": "FPO/TEST/2024/001",
    "description": "Test FPO for integration testing",
    "ceo_user": {
      "first_name": "Test",
      "last_name": "CEO",
      "phone_number": "+919999999001",
      "email": "test.ceo@example.com",
      "password": "TestPass@123"
    },
    "business_config": {
      "max_farmers": 500,
      "procurement_enabled": true
    }
  },
  "existing_ceo": {
    "phone_number": "+919999999002",
    "existing_fpo_id": "org_existing_123"
  }
}
```

## Happy Path Scenarios

### TC-HP-001: Create FPO with New CEO User

**Objective**: Verify successful FPO creation with a new CEO user

**Preconditions**:
- CEO phone number not registered in AAA
- Valid authentication token

**Test Steps**:
1. Send POST request to `/api/v1/identity/fpo/create`
2. Use phone number that doesn't exist in AAA
3. Provide all required fields

**Request**:
```json
{
  "name": "Rampur FPO",
  "registration_number": "FPO/MP/2024/001",
  "ceo_user": {
    "first_name": "Rajesh",
    "last_name": "Sharma",
    "phone_number": "+919876543210",
    "email": "rajesh@example.com",
    "password": "SecurePass@123"
  }
}
```

**Expected Result**:
- HTTP 201 Created
- Response contains:
  - `fpo_id`: Generated local ID
  - `aaa_org_id`: AAA organization ID
  - `ceo_user_id`: Newly created user ID
  - `user_groups`: Array of 4 groups
  - `status`: "ACTIVE"
- CEO user created in AAA
- Organization created in AAA
- 4 user groups created
- CEO role assigned
- Local FPO reference stored

**Verification**:
```sql
SELECT * FROM fpo_refs WHERE aaa_org_id = '{aaa_org_id}';
-- Should return 1 row with status = 'ACTIVE'
```

---

### TC-HP-002: Create FPO with Existing CEO User

**Objective**: Verify FPO creation reuses existing CEO user

**Preconditions**:
- CEO phone number already registered in AAA
- User is NOT CEO of any other FPO

**Test Steps**:
1. Create user in AAA first (or use existing)
2. Send POST request with same phone number
3. Password field can be different (will be ignored)

**Request**:
```json
{
  "name": "Sagar FPO",
  "registration_number": "FPO/MP/2024/002",
  "ceo_user": {
    "first_name": "Existing",
    "last_name": "User",
    "phone_number": "+919876543211",
    "password": "AnyPassword@123"
  }
}
```

**Expected Result**:
- HTTP 201 Created
- `ceo_user_id`: Existing user ID (not newly created)
- Organization created and linked to existing user
- Status: "ACTIVE"

**Verification**:
- No duplicate user created in AAA
- CEO role assigned to organization
- User count in AAA unchanged

---

### TC-HP-003: Create FPO with Business Config

**Objective**: Verify business configuration is stored correctly

**Request**:
```json
{
  "name": "Organic FPO",
  "registration_number": "FPO/MH/2024/003",
  "ceo_user": {
    "first_name": "Organic",
    "last_name": "Farmer",
    "phone_number": "+919876543212",
    "password": "SecurePass@123"
  },
  "business_config": {
    "max_farmers": 1000,
    "procurement_enabled": true,
    "credit_limit": 5000000,
    "focus_crops": ["wheat", "rice"]
  },
  "metadata": {
    "state": "Maharashtra",
    "established_year": 2024
  }
}
```

**Expected Result**:
- HTTP 201 Created
- `business_config` stored as JSONB in database
- All nested values preserved

**Verification**:
```sql
SELECT business_config FROM fpo_refs WHERE aaa_org_id = '{aaa_org_id}';
-- Should return JSONB with all fields
```

## Edge Case Scenarios

### TC-EC-001: Minimal Request (Required Fields Only)

**Objective**: Verify FPO creation with only required fields

**Request**:
```json
{
  "name": "Minimal FPO",
  "registration_number": "FPO/TEST/2024/MIN",
  "ceo_user": {
    "first_name": "Min",
    "last_name": "CEO",
    "phone_number": "+919999999003",
    "password": "Pass@123"
  }
}
```

**Expected Result**:
- HTTP 201 Created
- Optional fields null/empty in database
- No errors due to missing optional fields

---

### TC-EC-002: Very Long FPO Name

**Objective**: Test field length limits

**Test Cases**:
1. Name exactly at 255 char limit: Should succeed
2. Name > 255 chars: Should fail validation

**Request** (255 chars):
```json
{
  "name": "A{253 more chars}",
  "registration_number": "FPO/TEST/2024/LONG",
  "ceo_user": {...}
}
```

**Expected Result**:
- 255 chars: HTTP 201 Created
- 256+ chars: HTTP 400 Bad Request

---

### TC-EC-003: Special Characters in FPO Name

**Objective**: Test name field character validation

**Test Cases**:
```json
{
  "name": "Farmers' Producer & Marketing Co."  // Apostrophe, ampersand, period
}
```

**Expected Result**:
- HTTP 201 Created
- Special characters preserved in database

---

### TC-EC-004: Phone Number Formats

**Objective**: Verify various phone number formats

**Test Cases**:
1. `+919876543210` - No hyphens: Should succeed
2. `+91-9876543210` - With hyphen: Should succeed
3. `+91 9876 543210` - With spaces: Depends on validation
4. `9876543210` - No country code: Should fail

**Expected Result**:
- Valid E.164 formats accepted
- Invalid formats rejected with 400

## Error Scenarios

### TC-ERR-001: Missing Required Fields

**Objective**: Verify validation for missing required fields

**Test Cases**:

**Missing Name**:
```json
{
  "registration_number": "FPO/TEST/2024/ERR01",
  "ceo_user": {...}
}
```
**Expected**: HTTP 400, error: "FPO name is required"

**Missing Registration Number**:
```json
{
  "name": "Test FPO",
  "ceo_user": {...}
}
```
**Expected**: HTTP 400, error: "FPO registration number is required"

**Missing CEO First Name**:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/TEST/2024/ERR02",
  "ceo_user": {
    "last_name": "Sharma",
    "phone_number": "+919999999004",
    "password": "Pass@123"
  }
}
```
**Expected**: HTTP 400, error: "CEO user details are required"

---

### TC-ERR-002: Invalid Phone Number Format

**Objective**: Test phone number validation

**Request**:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/TEST/2024/ERR03",
  "ceo_user": {
    "first_name": "Test",
    "last_name": "User",
    "phone_number": "not-a-phone",
    "password": "Pass@123"
  }
}
```

**Expected Result**:
- HTTP 400 Bad Request
- Error message indicating invalid phone format

---

### TC-ERR-003: CEO Already CEO of Another FPO

**Objective**: Verify business rule BR-1.2 enforcement

**Preconditions**:
- User exists with phone `+919999999005`
- User already has CEO role in org `org_existing_001`

**Request**:
```json
{
  "name": "Second FPO",
  "registration_number": "FPO/TEST/2024/ERR04",
  "ceo_user": {
    "first_name": "Existing",
    "last_name": "CEO",
    "phone_number": "+919999999005",
    "password": "Pass@123"
  }
}
```

**Expected Result**:
- HTTP 400 Bad Request
- Error: "user is already CEO of another FPO - a user cannot be CEO of multiple FPOs simultaneously"

**Verification**:
- No organization created in AAA
- No FPO reference created in database

---

### TC-ERR-004: Wrong Field Name (registration_number)

**Objective**: Test common field name error

**Request**:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/TEST/2024/ERR05",  // Wrong field name
  "ceo_user": {...}
}
```

**Expected Result**:
- HTTP 400 Bad Request
- Error: "FPO registration number is required"
- (Because `registration_number` is missing)

---

### TC-ERR-005: Unauthorized Request

**Objective**: Verify authentication enforcement

**Request**: Same as TC-HP-001 but without Authorization header

**Expected Result**:
- HTTP 401 Unauthorized
- Error: "Invalid or missing authentication token"

---

### TC-ERR-006: Insufficient Permissions

**Objective**: Verify authorization check

**Preconditions**:
- User authenticated but lacks `create` permission on `fpo` resource

**Expected Result**:
- HTTP 403 Forbidden
- Error: "Insufficient permissions to create FPO"

## AAA Service Failure Scenarios

### TC-AAA-001: AAA Service Unavailable

**Objective**: Verify behavior when AAA service is down

**Setup**:
- Stop AAA service or block network access

**Request**: Valid FPO creation request

**Expected Result**:
- HTTP 500 Internal Server Error
- Error: "Failed to create organization: AAA service unavailable"
- No local FPO reference created

**Verification**:
- Operation should fail fast
- No partial state in database

---

### TC-AAA-002: User Creation Fails

**Objective**: Test AAA user creation failure handling

**Setup**:
- Mock AAA service to return error on RegisterRequest

**Expected Result**:
- HTTP 500 Internal Server Error
- Error: "failed to create CEO user: {AAA error}"
- No organization created
- No local FPO reference

---

### TC-AAA-003: Organization Creation Fails

**Objective**: Test AAA organization creation failure

**Setup**:
- User creation succeeds
- Organization creation fails

**Expected Result**:
- HTTP 500 Internal Server Error
- Error: "failed to create organization: {AAA error}"
- User exists but no organization
- No local FPO reference

**Cleanup Required**:
- Orphan user in AAA (manual cleanup needed)

---

### TC-AAA-004: Role Assignment Fails

**Objective**: Test partial setup with role assignment failure

**Setup**:
- User and organization creation succeed
- Role assignment fails

**Expected Result**:
- HTTP 201 Created (partial success)
- `status`: "PENDING_SETUP"
- `setup_errors`: Contains role assignment error
- Organization exists in AAA
- Local FPO reference created

**Recovery**:
```bash
POST /api/v1/identity/fpo/complete-setup
{
  "org_id": "{aaa_org_id}"
}
```

---

### TC-AAA-005: User Group Creation Fails

**Objective**: Test partial setup with group creation failure

**Setup**:
- User, org, and role succeed
- One or more group creations fail

**Expected Result**:
- HTTP 201 Created (partial success)
- `status`: "PENDING_SETUP"
- `setup_errors`: Contains group creation errors
- `user_groups`: Contains only successfully created groups

**Verification**:
```sql
SELECT setup_errors FROM fpo_refs WHERE aaa_org_id = '{aaa_org_id}';
-- Should contain error details for failed groups
```

## Database Failure Scenarios

### TC-DB-001: Local FPO Reference Save Fails

**Objective**: Test local database failure

**Setup**:
- AAA operations succeed
- Database connection fails or constraint violation

**Expected Result**:
- HTTP 500 Internal Server Error (or 201 with warning)
- Organization exists in AAA
- No local FPO reference

**Issue**: AAA organization created but no local tracking

**Recovery**: Manual data reconciliation needed

## Recovery Scenarios

### TC-REC-001: Complete Pending Setup

**Objective**: Verify CompleteFPOSetup endpoint

**Preconditions**:
- FPO in PENDING_SETUP status
- Some user groups or roles missing

**Request**:
```bash
POST /api/v1/identity/fpo/complete-setup
{
  "org_id": "org_pending_123"
}
```

**Expected Result**:
- HTTP 200 OK
- Retries failed operations
- If all succeed: Status updated to ACTIVE
- If still failing: Status remains PENDING_SETUP with updated errors

**Verification**:
```sql
SELECT status, setup_errors FROM fpo_refs WHERE aaa_org_id = 'org_pending_123';
```

## Performance Test Scenarios

### TC-PERF-001: Response Time

**Objective**: Verify acceptable response times

**Test**: Create 10 FPOs sequentially

**Expected Result**:
- P50 < 1000ms
- P95 < 2000ms
- P99 < 3000ms

**Metrics to Track**:
- AAA user lookup time
- AAA org creation time
- AAA role assignment time
- User group creation time (x4)
- Database write time

---

### TC-PERF-002: Concurrent Requests

**Objective**: Test concurrent FPO creation

**Test**: 10 concurrent requests with different phone numbers

**Expected Result**:
- All requests succeed (201)
- No duplicate organizations
- No race conditions in CEO role check

---

### TC-PERF-003: Large Business Config

**Objective**: Test with large JSONB payload

**Request**:
```json
{
  "name": "Large Config FPO",
  "registration_number": "FPO/TEST/2024/LARGE",
  "ceo_user": {...},
  "business_config": {
    // 100+ key-value pairs
    // Nested objects and arrays
    // Total size ~50KB
  }
}
```

**Expected Result**:
- HTTP 201 Created
- All data stored correctly
- Response time acceptable (<3s)

## Security Test Scenarios

### TC-SEC-001: SQL Injection in FPO Name

**Objective**: Verify SQL injection protection

**Request**:
```json
{
  "name": "Test'; DROP TABLE fpo_refs;--",
  "registration_number": "FPO/TEST/2024/SEC01",
  "ceo_user": {...}
}
```

**Expected Result**:
- HTTP 201 Created (or 400 if name validation rejects)
- Name stored as literal string
- No SQL execution
- Table not dropped

---

### TC-SEC-002: XSS in FPO Description

**Objective**: Verify XSS protection

**Request**:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/TEST/2024/SEC02",
  "description": "<script>alert('XSS')</script>",
  "ceo_user": {...}
}
```

**Expected Result**:
- HTTP 201 Created
- Script tags stored as text (not executed)
- Safe retrieval and display

---

### TC-SEC-003: Password in Logs

**Objective**: Verify passwords not logged

**Request**: Valid FPO creation with password

**Verification**:
- Check application logs
- Ensure password not in log output
- Only masked or omitted values

## Integration Test Scenarios

### TC-INT-001: End-to-End FPO Lifecycle

**Objective**: Complete FPO workflow test

**Steps**:
1. Create FPO (this endpoint)
2. Link farmers to FPO
3. Create farms under FPO
4. Start crop cycles
5. Verify permissions and access control

**Expected Result**:
- All operations succeed
- Data consistency across systems
- Proper access control enforcement

---

### TC-INT-002: Multiple FPOs Same State

**Objective**: Verify multiple FPOs can coexist

**Steps**:
1. Create FPO 1 in Maharashtra
2. Create FPO 2 in Maharashtra (different CEO)
3. Verify both active and independent

**Expected Result**:
- Both FPOs created successfully
- No conflicts or interference
- Each has separate org_id and user groups

## Regression Test Scenarios

### TC-REG-001: Field Name Compatibility

**Objective**: Ensure `registration_number` field works

**Request**: Use `registration_number` (correct field name)

**Expected**: HTTP 201 Created

---

### TC-REG-002: Backward Compatibility

**Objective**: Verify no breaking changes

**Test Cases**:
- Old request formats still work
- Optional fields remain optional
- Response structure unchanged

## Automated Test Template

```go
func TestCreateFPO_HappyPath(t *testing.T) {
    // Setup
    req := requests.CreateFPORequest{
        Name:           "Test FPO",
        RegistrationNo: "FPO/TEST/2024/001",
        CEOUser: requests.CEOUserData{
            FirstName:   "Test",
            LastName:    "CEO",
            PhoneNumber: "+919999999999",
            Password:    "TestPass@123",
        },
    }

    // Execute
    resp, err := fpoService.CreateFPO(ctx, &req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, resp)

    fpoData := resp.(*responses.CreateFPOData)
    assert.NotEmpty(t, fpoData.FPOID)
    assert.NotEmpty(t, fpoData.AAAOrgID)
    assert.Equal(t, "ACTIVE", fpoData.Status)
    assert.Len(t, fpoData.UserGroups, 4)
}
```

## Test Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: All happy paths + critical error paths
- **E2E Tests**: Complete workflows
- **Performance Tests**: Load and stress testing
- **Security Tests**: OWASP Top 10 scenarios

## Test Execution Checklist

- [ ] All happy path scenarios pass
- [ ] All error scenarios handled correctly
- [ ] AAA failure scenarios tested
- [ ] Database failure scenarios tested
- [ ] Recovery mechanisms verified
- [ ] Performance benchmarks met
- [ ] Security controls validated
- [ ] No sensitive data in logs
- [ ] No orphan data in failed operations
- [ ] Documentation matches actual behavior
