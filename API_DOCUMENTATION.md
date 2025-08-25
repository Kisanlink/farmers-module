# Farmers Module API Documentation

## Overview

The Farmers Module is a comprehensive microservice for managing agricultural operations within the KisanLink ecosystem. It provides RESTful APIs for farmer management, FPO (Farmer Producer Organization) operations, KisanSathi assignments, and farm management workflows.

**Base URL**: `http://localhost:8000/api/v1`
**Version**: 1.0.0
**OpenAPI Specification**: Available at `/swagger/doc.json`
**Interactive Documentation**: Available at `/swagger/index.html`

## Authentication & Authorization

All API endpoints require authentication through the AAA (Authentication, Authorization, and Accounting) service. The service validates:

- **User Authentication**: Valid JWT tokens
- **Organization Context**: User's organization membership
- **Permission Validation**: Role-based access control
- **Audit Logging**: All operations are logged for compliance

### Required Headers

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
X-Request-ID: <unique_request_id>
```

## API Endpoints

### 🏢 FPO Management

#### Create FPO Organization

**POST** `/fpo/create`

Creates a new FPO organization with CEO user setup and organizational structure.

**Request Body:**

```json
{
  "request_id": "req_123456789",
  "name": "Green Valley FPO",
  "registration_no": "FPO2024001",
  "registration_date": "2024-01-15",
  "address": "Village Green Valley, District ABC",
  "ceo_user_data": {
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "+91-9876543210",
    "email": "john.doe@greenvalley.com",
    "username": "john_ceo",
    "password": "SecurePass123!"
  }
}
```

**Response (201 Created):**

```json
{
  "success": true,
  "message": "FPO created successfully",
  "request_id": "req_123456789",
  "data": {
    "fpo_id": "fpo_abc123",
    "aaa_org_id": "org_xyz789",
    "name": "Green Valley FPO",
    "ceo_user_id": "user_ceo456",
    "user_groups": [
      {
        "group_id": "grp_farmers",
        "name": "Farmers",
        "description": "Farmer members of the FPO",
        "org_id": "org_xyz789",
        "permissions": ["farmer.read", "farmer.update"],
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Register FPO Reference

**POST** `/fpo/register-ref`

Registers an existing AAA organization as an FPO for local management.

**Request Body:**

```json
{
  "request_id": "req_987654321",
  "aaa_org_id": "org_existing123",
  "name": "Existing FPO Name",
  "business_config": {
    "crop_types": "wheat,rice,cotton",
    "season_preference": "kharif,rabi"
  }
}
```

#### Get FPO Reference

**GET** `/fpo/reference/{aaa_org_id}`

Retrieves FPO reference information by organization ID.

### 👨‍🌾 Farmer-FPO Linkage Management

#### Link Farmer to FPO

**POST** `/identity/link-farmer`

Links a farmer to an FPO with comprehensive validation.

**Request Body:**

```json
{
  "request_id": "req_link123",
  "aaa_user_id": "user_farmer456",
  "aaa_org_id": "org_fpo789"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Farmer linked successfully",
  "request_id": "req_link123",
  "data": {
    "linkage_id": "link_abc123",
    "farmer_id": "user_farmer456",
    "fpo_id": "org_fpo789",
    "status": "active",
    "linked_at": "2024-01-15T11:00:00Z",
    "linked_by": "user_admin123"
  }
}
```

#### Unlink Farmer from FPO

**POST** `/identity/unlink-farmer`

Unlinks a farmer from an FPO with soft delete functionality.

**Request Body:**

```json
{
  "request_id": "req_unlink456",
  "aaa_user_id": "user_farmer456",
  "aaa_org_id": "org_fpo789"
}
```

#### Get Farmer Linkage Status

**GET** `/identity/linkage/{farmer_id}/{org_id}`

Retrieves current linkage status and history for a farmer-FPO relationship.

### 🤝 KisanSathi Management

#### Assign KisanSathi to Farmer

**POST** `/kisansathi/assign`

Assigns a KisanSathi (agricultural advisor) to a specific farmer.

**Request Body:**

```json
{
  "request_id": "req_assign789",
  "aaa_user_id": "user_farmer456",
  "aaa_org_id": "org_fpo789",
  "kisan_sathi_user_id": "user_ks123"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "KisanSathi assigned successfully",
  "request_id": "req_assign789",
  "data": {
    "assignment_id": "assign_xyz456",
    "farmer_id": "user_farmer456",
    "kisan_sathi_id": "user_ks123",
    "fpo_id": "org_fpo789",
    "status": "active",
    "assigned_at": "2024-01-15T12:00:00Z",
    "assigned_by": "user_admin123"
  }
}
```

#### Reassign or Remove KisanSathi

**PUT** `/kisansathi/reassign`

Reassigns a KisanSathi to a different farmer or removes the assignment.

**Request Body:**

```json
{
  "request_id": "req_reassign101",
  "aaa_user_id": "user_farmer456",
  "aaa_org_id": "org_fpo789",
  "new_kisan_sathi_user_id": "user_ks789"
}
```

#### Create KisanSathi User

**POST** `/kisansathi/create-user`

Creates a new KisanSathi user with automatic role assignment.

**Request Body:**

```json
{
  "request_id": "req_create_ks202",
  "username": "kisansathi_raj",
  "phone_number": "+91-8765432109",
  "email": "raj.ks@example.com",
  "password": "SecureKS123!",
  "full_name": "Raj Kumar Singh",
  "country_code": "+91",
  "metadata": {
    "specialization": "crop_management",
    "experience_years": "5",
    "languages": "hindi,english"
  }
}
```

#### Get KisanSathi Assignment

**GET** `/kisansathi/assignment/{farmer_id}/{org_id}`

Retrieves current KisanSathi assignment details for a farmer.

### 🏥 Health & Admin Endpoints

#### Health Check

**GET** `/admin/health`

Returns service health status and system information.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T13:00:00Z",
  "version": "1.0.0",
  "database": "connected",
  "aaa_service": "connected"
}
```

#### Audit Trail

**GET** `/admin/audit`

Retrieves system audit logs (admin access required).

## Request/Response Patterns

### Standard Request Structure

All requests include common fields:

```json
{
  "request_id": "unique_identifier",
  "correlation_id": "optional_trace_id",
  "timestamp": "2024-01-15T10:00:00Z"
}
```

### Standard Response Structure

All successful responses follow this pattern:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:00:00Z",
  "data": {
    // Response-specific data
  }
}
```

### Error Response Structure

All error responses follow this pattern:

```json
{
  "error": "Error category",
  "message": "Detailed error message",
  "request_id": "req_123456789",
  "correlation_id": "trace_abc123",
  "details": {
    "field": "validation_error_details"
  },
  "timestamp": "2024-01-15T10:00:00Z"
}
```

## HTTP Status Codes

| Code | Description           | Usage                                 |
| ---- | --------------------- | ------------------------------------- |
| 200  | OK                    | Successful GET, PUT operations        |
| 201  | Created               | Successful POST operations            |
| 400  | Bad Request           | Validation errors, malformed requests |
| 401  | Unauthorized          | Missing or invalid authentication     |
| 403  | Forbidden             | Insufficient permissions              |
| 404  | Not Found             | Resource not found                    |
| 409  | Conflict              | Resource already exists               |
| 500  | Internal Server Error | System errors                         |

## Error Categories

### Validation Errors (400)

- Missing required fields
- Invalid data formats
- Business rule violations

### Authentication Errors (401)

- Missing JWT token
- Expired or invalid token
- Token signature verification failed

### Authorization Errors (403)

- Insufficient permissions
- Organization access denied
- Role-based access violations

### Not Found Errors (404)

- User not found
- Organization not found
- Resource not found

### Conflict Errors (409)

- User already exists
- Linkage already exists
- Duplicate resource creation

### System Errors (500)

- Database connection issues
- External service failures
- Unexpected system errors

## Rate Limiting

API endpoints implement rate limiting to ensure fair usage:

- **Standard endpoints**: 100 requests per minute per user
- **Bulk operations**: 10 requests per minute per user
- **Admin endpoints**: 50 requests per minute per user

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

## Data Models

### Farmer Linkage Data

```json
{
  "linkage_id": "string",
  "farmer_id": "string",
  "fpo_id": "string",
  "status": "active|inactive|suspended",
  "linked_at": "timestamp",
  "linked_by": "string",
  "metadata": {}
}
```

### KisanSathi Assignment Data

```json
{
  "assignment_id": "string",
  "farmer_id": "string",
  "kisan_sathi_id": "string",
  "fpo_id": "string",
  "status": "active|inactive|reassigned",
  "assigned_at": "timestamp",
  "assigned_by": "string"
}
```

### FPO Reference Data

```json
{
  "id": "string",
  "aaa_org_id": "string",
  "name": "string",
  "registration_no": "string",
  "business_config": {},
  "status": "active|inactive",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## Security Considerations

### Authentication

- JWT tokens with RS256 signature
- Token expiration validation
- Refresh token mechanism

### Authorization

- Role-based access control (RBAC)
- Organization-scoped permissions
- Resource-level access validation

### Data Protection

- Input validation and sanitization
- SQL injection prevention
- XSS protection
- CORS configuration

### Audit & Compliance

- Complete audit trail logging
- Request/response tracking
- User action monitoring
- Compliance reporting

## SDK & Integration

### cURL Examples

**Create FPO:**

```bash
curl -X POST "http://localhost:8000/api/v1/fpo/create" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_123456789" \
  -d '{
    "name": "Green Valley FPO",
    "registration_no": "FPO2024001",
    "ceo_user_data": {
      "first_name": "John",
      "last_name": "Doe",
      "phone_number": "+91-9876543210",
      "email": "john@example.com",
      "username": "john_ceo",
      "password": "SecurePass123!"
    }
  }'
```

**Link Farmer:**

```bash
curl -X POST "http://localhost:8000/api/v1/identity/link-farmer" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: req_link123" \
  -d '{
    "aaa_user_id": "user_farmer456",
    "aaa_org_id": "org_fpo789"
  }'
```

## Testing

### Test Coverage

- Unit tests for all service layers
- Integration tests for API endpoints
- Mock implementations for external dependencies
- End-to-end workflow testing

### Test Data

- Comprehensive test fixtures
- Mock AAA service responses
- Database seeding for integration tests

## Monitoring & Observability

### Metrics

- Request/response times
- Error rates by endpoint
- Authentication success/failure rates
- Database query performance

### Logging

- Structured JSON logging
- Request correlation IDs
- User action tracking
- Error stack traces

### Health Checks

- Database connectivity
- External service availability
- System resource utilization

## Support & Documentation

### Interactive Documentation

- **Swagger UI**: `http://localhost:8000/swagger/index.html`
- **OpenAPI Spec**: `http://localhost:8000/swagger/doc.json`
- **YAML Spec**: Available in `docs/swagger.yaml`

### Additional Resources

- Implementation guide: `TASK_7_IMPLEMENTATION.md`
- Workflow specifications: `FARMERS_MODULE_WORKFLOW_SPECIFICATIONS.md`
- Architecture overview: `WORKFLOW_ARCHITECTURE.md`

### Getting Help

- Check the comprehensive test suite for usage examples
- Review the OpenAPI specification for detailed parameter information
- Refer to the workflow documentation for business logic understanding

---

_This documentation is automatically generated from the OpenAPI specification and is kept up-to-date with the latest API changes._
