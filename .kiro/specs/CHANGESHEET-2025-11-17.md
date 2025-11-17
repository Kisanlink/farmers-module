# Changesheet - Farmers Module API Updates

**Date:** 2025-11-17
**Version:** 1.1.0
**Status:** Production Ready

---

## Overview

This changesheet documents breaking and non-breaking changes to the Farmers Module API for downstream service integration.

## Summary of Changes

| Change Type | Endpoint | Field/Param | Impact | Action Required |
|-------------|----------|-------------|---------|-----------------|
| **BREAKING** | `/api/v1/identity/fpo/create` | `registration_no` → `registration_number` | HIGH | Update all API calls |
| **FEATURE** | `/api/v1/identity/farmers` | Added `phone_number` query param | NONE | Optional enhancement |

---

## 1. BREAKING CHANGE: FPO Registration Field Rename

### Change Description
The FPO registration field has been renamed from `registration_no` to `registration_number` for clarity and consistency.

### Affected Endpoints

#### POST `/api/v1/identity/fpo/create`

**Before (DEPRECATED):**
```json
{
  "name": "Sree Rama FPO",
  "registration_no": "FPO00000001",
  "description": "...",
  "ceo_user": { ... }
}
```

**After (CURRENT):**
```json
{
  "name": "Sree Rama FPO",
  "registration_number": "FPO00000001",
  "description": "...",
  "ceo_user": { ... }
}
```

#### POST `/api/v1/identity/fpo/register`

**Before (DEPRECATED):**
```json
{
  "aaa_org_id": "org_123",
  "name": "Sree Rama FPO",
  "registration_no": "FPO00000001"
}
```

**After (CURRENT):**
```json
{
  "aaa_org_id": "org_123",
  "name": "Sree Rama FPO",
  "registration_number": "FPO00000001"
}
```

#### GET `/api/v1/identity/fpo/reference/:aaa_org_id`

**Response Change:**
```json
{
  "status": "success",
  "data": {
    "id": "FPOR00000001",
    "aaa_org_id": "org_123",
    "name": "Sree Rama FPO",
    "registration_number": "FPO00000001",  // ← Changed from registration_no
    "status": "ACTIVE",
    "created_at": "2025-11-17T10:00:00Z"
  }
}
```

### Error Handling

**Old requests using `registration_no` will fail with:**
```json
{
  "error": "FPO registration number is required"
}
```

**Status Code:** `400 Bad Request`

### Migration Guide

#### Step 1: Update Request Payloads
Replace all instances of `registration_no` with `registration_number` in:
- API clients
- Frontend forms
- Integration tests
- Documentation

#### Step 2: Update Response Parsers
Update code that reads FPO reference responses:

**Before:**
```javascript
const registrationNumber = fpoData.registration_no;
```

**After:**
```javascript
const registrationNumber = fpoData.registration_number;
```

#### Step 3: Test Thoroughly
- Test FPO creation flows
- Test FPO retrieval flows
- Verify error handling

### Timeline
- **Deployed:** 2025-11-17
- **Grace Period:** None (Breaking change)
- **Old Field Removed:** Immediately

---

## 2. NEW FEATURE: Farmer Search by Phone Number

### Change Description
Added `phone_number` query parameter to the List Farmers endpoint for fast farmer lookup.

### Affected Endpoints

#### GET `/api/v1/identity/farmers`

**New Query Parameter:**
```
GET /api/v1/identity/farmers?phone_number=9876543210
```

**Complete Parameter List:**
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `page` | integer | No | Page number (default: 1) | `1` |
| `page_size` | integer | No | Items per page (default: 10, max: 100) | `20` |
| `aaa_org_id` | string | No | Filter by organization | `ORGN00000001` |
| `kisan_sathi_user_id` | string | No | Filter by KisanSathi | `USER00000002` |
| `phone_number` | string | No | **NEW:** Filter by phone number | `9876543210` |

### Usage Examples

**1. Search by Phone Number Only:**
```bash
curl -X GET "https://api.example.com/api/v1/identity/farmers?phone_number=9876543210" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "status": "success",
  "message": "Farmers retrieved successfully",
  "data": [
    {
      "id": "FARM00000123",
      "aaa_user_id": "USER00000456",
      "aaa_org_id": "ORGN00000001",
      "first_name": "Ramesh",
      "last_name": "Kumar",
      "phone_number": "9876543210",
      "email": "ramesh@example.com",
      "created_at": "2025-10-15T08:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_count": 1,
    "total_pages": 1
  }
}
```

**2. Combined Filters (Phone + Organization):**
```bash
curl -X GET "https://api.example.com/api/v1/identity/farmers?phone_number=9876543210&aaa_org_id=ORGN00000001" \
  -H "Authorization: Bearer <token>"
```

**3. Phone Number Search with Pagination:**
```bash
curl -X GET "https://api.example.com/api/v1/identity/farmers?phone_number=9876543210&page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

### Performance

- **Database Index:** Phone number field is indexed for fast lookups
- **Expected Response Time:** < 50ms for single phone number search
- **Exact Match Only:** Phone number must match exactly (no partial matching)

### Important Notes

1. **Exact Match:** The phone number filter uses exact matching
   ```
   ✅ Matches: phone_number=9876543210
   ❌ Won't Match: phone_number=987654
   ```

2. **Format Consistency:** Ensure phone numbers are stored consistently
   - No spaces: `9876543210` ✅
   - No hyphens: `987-654-3210` ❌
   - No country codes in filter (unless stored with them)

3. **Empty Results:** Returns empty array if no match found
   ```json
   {
     "status": "success",
     "data": [],
     "pagination": {
       "total_count": 0,
       "total_pages": 0
     }
   }
   ```

### Use Cases

- **Customer Support:** Quick farmer lookup by phone during support calls
- **Mobile Apps:** "Find My Account" functionality
- **Integration:** Link external systems using phone as identifier
- **Verification:** Check if farmer exists before onboarding

### Backward Compatibility

✅ **Fully Backward Compatible** - Existing queries without `phone_number` parameter continue to work normally.

---

## 3. Technical Implementation Details

### Database Schema

**farmers table:**
```sql
CREATE TABLE farmers (
    id VARCHAR(255) PRIMARY KEY,
    aaa_user_id VARCHAR(255) NOT NULL,
    aaa_org_id VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),  -- ← Searchable field
    email VARCHAR(255),
    -- ... other fields
);

-- Index for fast phone number lookups
CREATE INDEX farmers_phone_idx ON farmers (phone_number);
```

**fpo_refs table:**
```sql
CREATE TABLE fpo_refs (
    id VARCHAR(255) PRIMARY KEY,
    aaa_org_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(255),  -- ← Renamed from registration_no
    status VARCHAR(50) DEFAULT 'ACTIVE',
    -- ... other fields
);
```

### API Contract Changes

**OpenAPI/Swagger Specification Updated:**
- FPO endpoints: `registration_no` → `registration_number`
- Farmers endpoint: Added `phone_number` query parameter

**Updated Swagger URL:**
```
https://<your-domain>/docs
```

---

## 4. Testing Checklist

### For Downstream Services

- [ ] **FPO Creation Tests**
  - [ ] Update test payloads to use `registration_number`
  - [ ] Verify successful FPO creation
  - [ ] Test error handling for missing registration_number

- [ ] **FPO Retrieval Tests**
  - [ ] Update response parsers to read `registration_number`
  - [ ] Verify existing FPO data displays correctly

- [ ] **Farmer Search Tests**
  - [ ] Test phone number search with valid number
  - [ ] Test phone number search with invalid number
  - [ ] Test combined filters (phone + org)
  - [ ] Test pagination with phone number filter
  - [ ] Test backward compatibility (queries without phone filter)

### Sample Test Cases

**Test 1: FPO Creation with New Field**
```javascript
// Test: Create FPO with registration_number
const response = await createFPO({
  name: "Test FPO",
  registration_number: "FPO2025001",  // ← New field name
  ceo_user: { /* ... */ }
});

assert.equal(response.status, 201);
assert.equal(response.data.registration_number, "FPO2025001");
```

**Test 2: Farmer Phone Search**
```javascript
// Test: Search farmer by phone number
const response = await listFarmers({
  phone_number: "9876543210"
});

assert.equal(response.status, 200);
assert.isArray(response.data);
if (response.data.length > 0) {
  assert.equal(response.data[0].phone_number, "9876543210");
}
```

---

## 5. Rollout Plan

### Phase 1: Immediate (2025-11-17)
- ✅ Deploy farmers-module with changes
- ✅ Update Swagger documentation
- ✅ Publish changesheet

### Phase 2: Downstream Updates (Next 24-48 hours)
- [ ] Frontend team updates FPO forms
- [ ] Mobile apps update API calls
- [ ] Integration services update payloads
- [ ] Update automated tests

### Phase 3: Monitoring (Ongoing)
- [ ] Monitor error rates for 400 errors on FPO endpoints
- [ ] Track usage of phone_number filter
- [ ] Collect performance metrics

---

## 6. Support & Questions

### Contact Points

- **Backend Team:** @sde-backend-engineer
- **Architecture:** @sde3-backend-architect
- **On-Call:** Check PagerDuty schedule

### Resources

- **Swagger Docs:** `https://<your-domain>/docs`
- **API Guide:** `.kiro/specs/fpo-aaa-integration/API_GUIDE.md`
- **Test Scenarios:** `.kiro/specs/fpo-aaa-integration/TEST_SCENARIOS.md`

### Reporting Issues

If you encounter issues:
1. Check this changesheet for correct field names
2. Verify request payload format
3. Check Swagger docs for current API contract
4. Contact backend team with:
   - Request payload
   - Response received
   - Expected behavior

---

## 7. Quick Reference

### FPO Endpoints - Field Changes

| Endpoint | Method | Old Field | New Field |
|----------|--------|-----------|-----------|
| `/api/v1/identity/fpo/create` | POST | `registration_no` | `registration_number` |
| `/api/v1/identity/fpo/register` | POST | `registration_no` | `registration_number` |
| `/api/v1/identity/fpo/reference/:id` | GET | `registration_no` | `registration_number` |

### Farmers Endpoint - New Parameters

| Endpoint | Method | New Parameter | Type | Example |
|----------|--------|---------------|------|---------|
| `/api/v1/identity/farmers` | GET | `phone_number` | string | `9876543210` |

---

## 8. Code Migration Examples

### JavaScript/TypeScript

**Before:**
```typescript
// Old FPO creation
const fpo = await api.post('/api/v1/identity/fpo/create', {
  name: "My FPO",
  registration_no: "FPO123",  // ❌ Old field
  ceo_user: { ... }
});

// Old response parsing
console.log(fpo.data.registration_no);  // ❌ Old field
```

**After:**
```typescript
// New FPO creation
const fpo = await api.post('/api/v1/identity/fpo/create', {
  name: "My FPO",
  registration_number: "FPO123",  // ✅ New field
  ceo_user: { ... }
});

// New response parsing
console.log(fpo.data.registration_number);  // ✅ New field

// New farmer phone search
const farmers = await api.get('/api/v1/identity/farmers', {
  params: { phone_number: "9876543210" }  // ✅ New feature
});
```

### Python

**Before:**
```python
# Old FPO creation
response = requests.post(
    f"{API_BASE}/api/v1/identity/fpo/create",
    json={
        "name": "My FPO",
        "registration_no": "FPO123",  # ❌ Old field
        "ceo_user": { ... }
    }
)
registration = response.json()["data"]["registration_no"]  # ❌ Old field
```

**After:**
```python
# New FPO creation
response = requests.post(
    f"{API_BASE}/api/v1/identity/fpo/create",
    json={
        "name": "My FPO",
        "registration_number": "FPO123",  # ✅ New field
        "ceo_user": { ... }
    }
)
registration = response.json()["data"]["registration_number"]  # ✅ New field

# New farmer phone search
response = requests.get(
    f"{API_BASE}/api/v1/identity/farmers",
    params={"phone_number": "9876543210"}  # ✅ New feature
)
```

### Go

**Before:**
```go
// Old FPO creation
type CreateFPORequest struct {
    Name           string     `json:"name"`
    RegistrationNo string     `json:"registration_no"`  // ❌ Old field
    CEOUser        CEOUser    `json:"ceo_user"`
}
```

**After:**
```go
// New FPO creation
type CreateFPORequest struct {
    Name               string     `json:"name"`
    RegistrationNumber string     `json:"registration_number"`  // ✅ New field
    CEOUser            CEOUser    `json:"ceo_user"`
}

// New farmer search with phone
params := url.Values{}
params.Add("phone_number", "9876543210")  // ✅ New feature
resp, err := http.Get(apiBase + "/api/v1/identity/farmers?" + params.Encode())
```

---

## Appendix A: Error Messages

### FPO Creation Errors

| Error Message | Status Code | Cause | Solution |
|--------------|-------------|-------|----------|
| "FPO registration number is required" | 400 | Using old field `registration_no` or missing field | Use `registration_number` field |
| "Invalid request body" | 400 | Malformed JSON | Validate JSON syntax |
| "FPO name is required" | 400 | Missing name field | Add `name` field |

### Farmer Search Errors

| Error Message | Status Code | Cause | Solution |
|--------------|-------------|-------|----------|
| "Invalid page number" | 400 | Page < 1 | Use page >= 1 |
| "Invalid page size" | 400 | Page size > 100 or < 1 | Use 1 <= page_size <= 100 |

---

## Appendix B: Postman Collection Updates

**Update your Postman collection:**

1. **FPO Creation Request:**
   - Find request: "Create FPO"
   - Update body: Change `registration_no` → `registration_number`
   - Save

2. **Add Farmer Phone Search:**
   - Create new request: "Search Farmer by Phone"
   - Method: GET
   - URL: `{{base_url}}/api/v1/identity/farmers`
   - Params: `phone_number = 9876543210`
   - Save

---

**END OF CHANGESHEET**

*For latest updates, check: `.kiro/specs/CHANGESHEET-2025-11-17.md`*
