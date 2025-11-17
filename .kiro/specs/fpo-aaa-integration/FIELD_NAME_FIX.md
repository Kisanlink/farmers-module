# Field Name Validation Issue: registration_number vs registration_number

## Issue Description

**Problem**: API consumers receiving validation errors when using `registration_number` field name

**Root Cause**: Field name mismatch between expected API field and what clients are sending

- **Correct field name**: `registration_number`
- **Incorrect field name** (causing errors): `registration_number`

## Current Implementation

### Request Model
**File**: `/internal/entities/requests/fpo_ref.go`

```go
type CreateFPORequest struct {
    BaseRequest
    Name           string                 `json:"name" validate:"required"`
    RegistrationNo string                 `json:"registration_number" validate:"required"`  // Line 16
    Description    string                 `json:"description"`
    CEOUser        CEOUserData            `json:"ceo_user" validate:"required"`
    BusinessConfig map[string]interface{} `json:"business_config"`
    Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
```

### Database Model
**File**: `/internal/entities/fpo/fpo.go`

```go
type FPORef struct {
    base.BaseModel
    AAAOrgID       string         `json:"aaa_org_id" gorm:"type:varchar(255);unique;not null"`
    Name           string         `json:"name" gorm:"type:varchar(255);not null"`
    RegistrationNo string         `json:"registration_number" gorm:"type:varchar(255)"`  // Line 49
    Status         FPOStatus      `json:"status" gorm:"type:varchar(50);default:'ACTIVE'"`
    BusinessConfig entities.JSONB `json:"business_config" gorm:"type:jsonb;default:'{}';serializer:json"`
    SetupErrors    entities.JSONB `json:"setup_errors,omitempty" gorm:"type:jsonb;serializer:json"`
}
```

## Fix Options

### Option 1: Client-Side Fix (RECOMMENDED)

**Action**: Update API consumers to use correct field name

**Pros**:
- No server-side changes required
- Maintains consistency with Go naming conventions
- Matches existing database schema
- No risk of breaking other integrations

**Cons**:
- Requires coordination with all API consumers
- May need documentation updates

**Implementation**:
Update client code from:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/2024/001",  // Wrong
  "ceo_user": {...}
}
```

To:
```json
{
  "name": "Test FPO",
  "registration_number": "FPO/2024/001",  // Correct
  "ceo_user": {...}
}
```

**Communication Template**:
```
Subject: API Field Name Correction - FPO Creation Endpoint

The FPO creation endpoint requires the field name `registration_number` (not `registration_number`).

INCORRECT:
  "registration_number": "FPO/2024/001"

CORRECT:
  "registration_number": "FPO/2024/001"

Please update your integration code accordingly.

Updated API documentation: [link to API_GUIDE.md]
```

---

### Option 2: Add JSON Tag Alias (Alternative)

**Action**: Support both field names via JSON unmarshaling

**Pros**:
- Backward compatible with existing clients
- Forward compatible with corrected clients
- No client-side changes required

**Cons**:
- Perpetuates inconsistent naming
- More complex deserialization logic
- May confuse future developers

**Implementation**:

#### Step 1: Create custom unmarshaler

**File**: `/internal/entities/requests/fpo_ref.go`

```go
// UnmarshalJSON implements custom JSON unmarshaling to support both field names
func (r *CreateFPORequest) UnmarshalJSON(data []byte) error {
    // Create alias type to avoid recursion
    type Alias CreateFPORequest

    // Temporary struct with both field names
    temp := &struct {
        RegistrationNumber string `json:"registration_number,omitempty"`
        *Alias
    }{
        Alias: (*Alias)(r),
    }

    if err := json.Unmarshal(data, temp); err != nil {
        return err
    }

    // If registration_number was provided, use it
    if temp.RegistrationNumber != "" && r.RegistrationNo == "" {
        r.RegistrationNo = temp.RegistrationNumber
    }

    return nil
}
```

#### Step 2: Add validation test

**File**: `/internal/entities/requests/fpo_ref_test.go`

```go
func TestCreateFPORequest_UnmarshalJSON_BothFieldNames(t *testing.T) {
    tests := []struct {
        name     string
        jsonData string
        expected string
    }{
        {
            name:     "registration_number field",
            jsonData: `{"name":"Test","registration_number":"FPO/001","ceo_user":{...}}`,
            expected: "FPO/001",
        },
        {
            name:     "registration_number field (legacy)",
            jsonData: `{"name":"Test","registration_number":"FPO/002","ceo_user":{...}}`,
            expected: "FPO/002",
        },
        {
            name:     "both fields (registration_number takes precedence)",
            jsonData: `{"name":"Test","registration_number":"FPO/003","registration_number":"FPO/004","ceo_user":{...}}`,
            expected: "FPO/003",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var req CreateFPORequest
            err := json.Unmarshal([]byte(tt.jsonData), &req)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, req.RegistrationNo)
        })
    }
}
```

#### Step 3: Update Swagger documentation

**File**: `/internal/handlers/fpo_handlers.go`

Update Swagger comment to document both field names:

```go
// @Param request body requests.CreateFPORequest true "Create FPO Request (use 'registration_number' or 'registration_number')"
```

---

### Option 3: API Versioning (Future-Proof)

**Action**: Create v2 API with consistent naming

**Pros**:
- Clean break from legacy naming
- Opportunity for other improvements
- Clear migration path

**Cons**:
- Requires maintaining two API versions
- More complex deployment
- Significant effort

**Implementation**:
- Create `/api/v2/identity/fpo/create` with strict field names
- Deprecate `/api/v1/identity/fpo/create`
- Provide migration timeline

Not recommended for this specific issue (overkill).

## Recommendation

**Recommended Approach**: **Option 1 (Client-Side Fix)**

### Rationale

1. **Simplicity**: No server-side code changes required
2. **Consistency**: Maintains Go naming conventions throughout codebase
3. **Standards Compliance**: `snake_case` for JSON is standard in Go ecosystem
4. **Risk Mitigation**: No risk of breaking existing working clients
5. **Clear Communication**: One-time update to API consumers

### Implementation Plan

#### Phase 1: Documentation (Immediate)
- ✅ Update API_GUIDE.md with correct field name (already done)
- ✅ Add troubleshooting section for this specific error (already done)
- ✅ Update Swagger/OpenAPI specs to emphasize correct field name

#### Phase 2: Communication (Week 1)
- [ ] Identify all API consumers (frontend, mobile, partners)
- [ ] Send notification about correct field name
- [ ] Provide code examples for each platform
- [ ] Set deadline for updates (2-4 weeks)

#### Phase 3: Monitoring (Week 2-4)
- [ ] Add logging to track which clients are still using wrong field
- [ ] Monitor error logs for "registration number is required" errors
- [ ] Follow up with teams showing errors

#### Phase 4: Validation (Week 4+)
- [ ] Verify all clients updated
- [ ] Confirm error rate decreased to zero
- [ ] Close issue

### Fallback Plan

If Option 1 proves too difficult (too many clients to update), implement **Option 2** as a compromise:

1. Add custom UnmarshalJSON to support both field names
2. Log deprecation warning when `registration_number` is used
3. Set sunset date for `registration_number` support (e.g., 6 months)
4. Remove alias support after sunset date

## Communication Templates

### Email to API Consumers

```
Subject: Action Required - FPO Creation API Field Name Correction

Hi Team,

We've identified a field name discrepancy in the FPO Creation API that may be causing validation errors.

ISSUE:
When creating FPOs, the field for registration number must be named "registration_number" (not "registration_number").

REQUIRED CHANGES:
Please update your API integration code to use the correct field name.

Before:
{
  "name": "Test FPO",
  "registration_number": "FPO/2024/001",  ❌ INCORRECT
  ...
}

After:
{
  "name": "Test FPO",
  "registration_number": "FPO/2024/001",  ✅ CORRECT
  ...
}

RESOURCES:
- Full API Guide: .kiro/specs/fpo-aaa-integration/API_GUIDE.md
- Swagger Docs: https://api.example.com/docs
- Support: backend-team@example.com

DEADLINE: [Date - 2 weeks from now]

Thank you for your cooperation!

Backend Team
```

### Slack Message

```
⚠️ FPO Creation API - Field Name Correction

If you're integrating with the FPO creation endpoint, please note:

The correct field name is `registration_number` (NOT `registration_number`)

✅ Correct:   "registration_number": "FPO/2024/001"
❌ Incorrect: "registration_number": "FPO/2024/001"

📖 Full docs: [link to API_GUIDE.md]
🙋 Questions? Ask in #backend-support

Deadline for updates: [Date]
```

## Monitoring & Validation

### Log Pattern to Monitor

```bash
# Check for registration number validation errors
grep "FPO registration number is required" /var/log/farmers-module/*.log | wc -l

# Should trend to zero after client updates
```

### Metric to Track

```
fpo_creation_validation_errors_total{field="registration_number"}
```

## Testing Updates

### Add Test Case for Common Error

**File**: `/internal/handlers/fpo_handlers_test.go`

```go
func TestCreateFPO_WrongFieldName_ReturnsValidationError(t *testing.T) {
    // Simulate client using wrong field name
    reqBody := `{
        "name": "Test FPO",
        "registration_number": "FPO/2024/001",
        "ceo_user": {
            "first_name": "Test",
            "last_name": "CEO",
            "phone_number": "+919999999999",
            "password": "Pass@123"
        }
    }`

    resp := performRequest(http.MethodPost, "/api/v1/identity/fpo/create", reqBody)

    assert.Equal(t, http.StatusBadRequest, resp.Code)
    assert.Contains(t, resp.Body.String(), "registration number is required")
}
```

## Documentation Updates Required

1. ✅ **API_GUIDE.md**: Add prominent note about correct field name
2. ✅ **API_GUIDE.md**: Add troubleshooting section for this error
3. ✅ **TEST_SCENARIOS.md**: Add test case TC-ERR-004
4. [ ] **Swagger annotations**: Update with example emphasizing correct field
5. [ ] **Postman collection**: Update with correct field name
6. [ ] **Client SDKs**: Update generated code (if any)

## Conclusion

The `registration_number` vs `registration_number` issue is a client-side field name mismatch. The **recommended fix is client-side updates** to use the correct field name `registration_number`, which aligns with:

- Go JSON naming conventions
- Existing database schema
- Internal consistency across the codebase

This approach minimizes risk, maintains standards compliance, and requires no server-side code changes.

**Action Items**:
1. Notify all API consumers of correct field name
2. Provide migration deadline (2-4 weeks)
3. Monitor error logs to track adoption
4. Consider Option 2 (alias support) only if client updates prove infeasible

**Status**: Ready for client communication
