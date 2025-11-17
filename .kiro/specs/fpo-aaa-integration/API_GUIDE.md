# FPO Creation API Guide

## Overview

This guide provides complete documentation for the FPO (Farmer Producer Organization) creation endpoint, which integrates with the AAA service for organization management.

## Endpoint Details

### Create FPO Organization

```
POST /api/v1/identity/fpo/create
```

Creates a new FPO organization with automatic AAA service integration for identity, roles, and permissions.

**Authentication**: Required (Bearer token)

**Authorization**: Requires `create` permission on `fpo` resource

## Request Specification

### Headers

```
Content-Type: application/json
Authorization: Bearer <jwt_token>
```

### Request Body

```json
{
  "name": "string (required)",
  "registration_number": "string (required)",
  "description": "string (optional)",
  "ceo_user": {
    "first_name": "string (required)",
    "last_name": "string (required)",
    "phone_number": "string (required, E.164 format)",
    "email": "string (optional, valid email)",
    "password": "string (required, min 8 chars)"
  },
  "business_config": {
    "key": "value (optional)"
  },
  "metadata": {
    "key": "value (optional)"
  }
}
```

### Field Specifications

#### FPO Details

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `name` | string | Yes | Non-empty | Official name of the FPO |
| `registration_number` | string | Yes | Non-empty | Government registration number |
| `description` | string | No | - | Brief description of FPO operations |

**⚠️ IMPORTANT**: Use `registration_number`, NOT `registration_number`

#### CEO User Details

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `first_name` | string | Yes | Non-empty | CEO's first name |
| `last_name` | string | Yes | Non-empty | CEO's last name |
| `phone_number` | string | Yes | Valid phone | CEO's phone (E.164 format) |
| `email` | string | No | Valid email | CEO's email address |
| `password` | string | Yes | Min 8 chars | Initial password for CEO account |

#### Optional Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `business_config` | object | No | FPO-specific business configuration |
| `metadata` | object | No | Additional metadata for organization |

## Request Examples

### Minimal Request

```json
{
  "name": "Rampur Farmers Producer Company",
  "registration_number": "FPO/MP/2024/001234",
  "ceo_user": {
    "first_name": "Rajesh",
    "last_name": "Sharma",
    "phone_number": "+919876543210",
    "password": "SecurePass@123"
  }
}
```

### Complete Request

```json
{
  "name": "Rampur Farmers Producer Company",
  "registration_number": "FPO/MP/2024/001234",
  "description": "A farmer producer organization serving 500+ farmers in Rampur region, focused on organic farming and direct market access",
  "ceo_user": {
    "first_name": "Rajesh",
    "last_name": "Sharma",
    "phone_number": "+919876543210",
    "email": "rajesh.sharma@rampurfpo.com",
    "password": "SecurePass@123"
  },
  "business_config": {
    "max_farmers": 1000,
    "procurement_enabled": true,
    "credit_limit": 5000000,
    "storage_capacity_mt": 500,
    "focus_crops": ["wheat", "soybean", "pulses"]
  },
  "metadata": {
    "state": "Madhya Pradesh",
    "district": "Sagar",
    "established_year": 2024,
    "registration_authority": "Ministry of Corporate Affairs"
  }
}
```

## Response Specification

### Success Response (201 Created)

```json
{
  "success": true,
  "message": "FPO created successfully",
  "request_id": "req_abc123xyz",
  "data": {
    "fpo_id": "FPOR_1234567890abcdef",
    "aaa_org_id": "org_987654321fedcba",
    "name": "Rampur Farmers Producer Company",
    "ceo_user_id": "user_abc123def456",
    "user_groups": [
      {
        "group_id": "grp_directors_001",
        "name": "directors",
        "org_id": "org_987654321fedcba",
        "permissions": ["manage", "read", "write", "approve"],
        "created_at": "2025-11-17T09:00:00Z"
      },
      {
        "group_id": "grp_shareholders_001",
        "name": "shareholders",
        "org_id": "org_987654321fedcba",
        "permissions": ["read", "vote"],
        "created_at": "2025-11-17T09:00:00Z"
      },
      {
        "group_id": "grp_store_staff_001",
        "name": "store_staff",
        "org_id": "org_987654321fedcba",
        "permissions": ["read", "write", "inventory"],
        "created_at": "2025-11-17T09:00:00Z"
      },
      {
        "group_id": "grp_store_managers_001",
        "name": "store_managers",
        "org_id": "org_987654321fedcba",
        "permissions": ["read", "write", "manage", "inventory", "reports"],
        "created_at": "2025-11-17T09:00:00Z"
      }
    ],
    "status": "ACTIVE",
    "created_at": "2025-11-17T09:00:00Z"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Operation success indicator |
| `message` | string | Human-readable success message |
| `request_id` | string | Unique request identifier for tracking |
| `data.fpo_id` | string | Local FPO reference ID |
| `data.aaa_org_id` | string | AAA service organization ID |
| `data.name` | string | FPO organization name |
| `data.ceo_user_id` | string | AAA service user ID for CEO |
| `data.user_groups` | array | Created user groups with permissions |
| `data.status` | string | FPO status (`ACTIVE` or `PENDING_SETUP`) |
| `data.created_at` | timestamp | Creation timestamp (ISO 8601) |

### Status Values

- **ACTIVE**: FPO fully set up, all AAA operations succeeded
- **PENDING_SETUP**: Partial setup, some user groups or roles failed (retry with `/identity/fpo/complete-setup`)

## Error Responses

### 400 Bad Request - Invalid Input

```json
{
  "error": "Bad Request",
  "message": "Invalid request body",
  "request_id": "req_abc123xyz",
  "details": {
    "field": "registration_number",
    "issue": "field is required"
  }
}
```

**Common Causes**:
- Missing required fields
- Invalid field format
- Wrong field name (e.g., `registration_number` instead of `registration_number`)

### 400 Bad Request - Business Rule Violation

```json
{
  "error": "Bad Request",
  "message": "user is already CEO of another FPO - a user cannot be CEO of multiple FPOs simultaneously",
  "request_id": "req_abc123xyz"
}
```

**Cause**: CEO user already has CEO role in another organization

### 401 Unauthorized

```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token",
  "request_id": "req_abc123xyz"
}
```

### 403 Forbidden

```json
{
  "error": "Forbidden",
  "message": "Insufficient permissions to create FPO",
  "request_id": "req_abc123xyz"
}
```

### 500 Internal Server Error

```json
{
  "error": "Internal Server Error",
  "message": "Failed to create organization in AAA service",
  "request_id": "req_abc123xyz"
}
```

**Common Causes**:
- AAA service unavailable
- Database connection failure
- Network timeout

## Validation Rules

### FPO Name
- **Required**: Yes
- **Min Length**: 3 characters
- **Max Length**: 255 characters
- **Pattern**: Alphanumeric with spaces, hyphens, and periods

### Registration Number
- **Required**: Yes
- **Format**: Alphanumeric with slashes and hyphens
- **Example**: `FPO/MP/2024/001234`

### CEO Phone Number
- **Required**: Yes
- **Format**: E.164 international format
- **Pattern**: `+[country_code][number]`
- **Examples**:
  - ✅ `+919876543210`
  - ✅ `+91-9876543210`
  - ❌ `9876543210` (missing country code)

### CEO Email
- **Required**: No
- **Format**: Valid email address
- **Pattern**: RFC 5322 compliant

### CEO Password
- **Required**: Yes
- **Min Length**: 8 characters
- **Recommended**: Include uppercase, lowercase, numbers, and special characters

## Business Logic

### CEO User Handling

1. **Existing User**: If CEO phone number exists in AAA:
   - User is linked to new FPO
   - Password field is ignored
   - User must not be CEO of another FPO

2. **New User**: If CEO phone number not found:
   - New user account created in AAA
   - CEO role assigned to new organization
   - User can log in with provided password

### Automatic Setup

On successful FPO creation, the following are automatically configured:

1. **AAA Organization**: Created with type "FPO"
2. **CEO Role**: Assigned to CEO user
3. **User Groups**: 4 groups with predefined permissions
   - Directors (manage, read, write, approve)
   - Shareholders (read, vote)
   - Store Staff (read, write, inventory)
   - Store Managers (read, write, manage, inventory, reports)
4. **Local Reference**: Stored in farmers-module database

### Idempotency

⚠️ **Not Currently Implemented**

Multiple requests with the same data will create multiple FPO organizations. Use unique `request_id` or implement client-side deduplication.

## cURL Examples

### Basic Request

```bash
curl -X POST https://api.example.com/api/v1/identity/fpo/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{
    "name": "Rampur Farmers Producer Company",
    "registration_number": "FPO/MP/2024/001234",
    "ceo_user": {
      "first_name": "Rajesh",
      "last_name": "Sharma",
      "phone_number": "+919876543210",
      "password": "SecurePass@123"
    }
  }'
```

### With Business Config

```bash
curl -X POST https://api.example.com/api/v1/identity/fpo/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -d '{
    "name": "Rampur Farmers Producer Company",
    "registration_number": "FPO/MP/2024/001234",
    "description": "Organic farming focused FPO",
    "ceo_user": {
      "first_name": "Rajesh",
      "last_name": "Sharma",
      "phone_number": "+919876543210",
      "email": "rajesh@rampurfpo.com",
      "password": "SecurePass@123"
    },
    "business_config": {
      "max_farmers": 1000,
      "procurement_enabled": true
    }
  }'
```

## Postman Collection

### Request Setup

1. **Method**: POST
2. **URL**: `{{base_url}}/api/v1/identity/fpo/create`
3. **Headers**:
   - `Content-Type`: `application/json`
   - `Authorization`: `Bearer {{auth_token}}`
4. **Body** (raw JSON):
   ```json
   {
     "name": "{{fpo_name}}",
     "registration_number": "{{registration_number}}",
     "ceo_user": {
       "first_name": "{{ceo_first_name}}",
       "last_name": "{{ceo_last_name}}",
       "phone_number": "{{ceo_phone}}",
       "email": "{{ceo_email}}",
       "password": "{{ceo_password}}"
     }
   }
   ```

### Environment Variables

```json
{
  "base_url": "https://api.example.com",
  "auth_token": "your_jwt_token_here",
  "fpo_name": "Test FPO",
  "registration_number": "FPO/TEST/2024/001",
  "ceo_first_name": "John",
  "ceo_last_name": "Doe",
  "ceo_phone": "+919999999999",
  "ceo_email": "john.doe@example.com",
  "ceo_password": "TestPass@123"
}
```

## Common Integration Patterns

### Frontend Integration (React/Vue)

```javascript
async function createFPO(fpoData) {
  const response = await fetch('/api/v1/identity/fpo/create', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getAuthToken()}`
    },
    body: JSON.stringify({
      name: fpoData.name,
      registration_number: fpoData.registrationNo,  // Note: camelCase to snake_case conversion
      description: fpoData.description,
      ceo_user: {
        first_name: fpoData.ceo.firstName,
        last_name: fpoData.ceo.lastName,
        phone_number: fpoData.ceo.phoneNumber,
        email: fpoData.ceo.email,
        password: fpoData.ceo.password
      },
      business_config: fpoData.businessConfig
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message);
  }

  return response.json();
}
```

### Backend Integration (Go)

```go
type CreateFPOAPIRequest struct {
    Name          string                 `json:"name"`
    RegistrationNo string                `json:"registration_number"`
    Description   string                 `json:"description,omitempty"`
    CEOUser       CEOUserAPIData         `json:"ceo_user"`
    BusinessConfig map[string]interface{} `json:"business_config,omitempty"`
}

func callCreateFPOAPI(ctx context.Context, req CreateFPOAPIRequest) (*CreateFPOResponse, error) {
    body, _ := json.Marshal(req)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.example.com/api/v1/identity/fpo/create",
        bytes.NewBuffer(body))

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer " + token)

    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CreateFPOResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

## Troubleshooting

### Issue: "Field registration_number not recognized"

**Problem**: Using wrong field name
**Solution**: Use `registration_number` instead of `registration_number`

```diff
{
  "name": "Test FPO",
- "registration_number": "FPO/2024/001",
+ "registration_number": "FPO/2024/001",
  ...
}
```

### Issue: "User is already CEO of another FPO"

**Problem**: CEO phone number already assigned as CEO
**Solutions**:
1. Use different phone number for CEO
2. Use existing FPO if it's the same organization
3. Remove CEO role from previous FPO (contact admin)

### Issue: Status is "PENDING_SETUP"

**Problem**: Some user groups or roles failed to create
**Solution**: Call completion endpoint

```bash
POST /api/v1/identity/fpo/complete-setup
{
  "org_id": "org_987654321fedcba"
}
```

### Issue: "AAA service unavailable"

**Problem**: AAA service is down or unreachable
**Solutions**:
1. Check AAA service health
2. Verify network connectivity
3. Retry request after service recovery

## Rate Limiting

**Recommended**: 10 requests per minute per API key

This operation is resource-intensive (multiple AAA calls), so rate limiting is advisable.

## Related Endpoints

- **Register FPO Reference**: `POST /api/v1/identity/fpo/register`
- **Get FPO Reference**: `GET /api/v1/identity/fpo/reference/{aaa_org_id}`
- **Complete FPO Setup**: `POST /api/v1/identity/fpo/complete-setup`

## Support

For API support or issues:
- Check logs with `request_id` for debugging
- Review AAA service status
- Contact backend team with request details
