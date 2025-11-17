# Farmers Module API - Quick Reference

## Base Information

- **Base URL**: `http://localhost:8000/api/v1`
- **Authentication**: Bearer JWT token required
- **Content-Type**: `application/json`

## Quick Endpoint Reference

### FPO Management

| Method | Endpoint                  | Description                     |
| ------ | ------------------------- | ------------------------------- |
| POST   | `/fpo/create`             | Create new FPO organization     |
| POST   | `/fpo/register-ref`       | Register existing FPO reference |
| GET    | `/fpo/reference/{org_id}` | Get FPO reference details       |

### Farmer-FPO Linkage

| Method | Endpoint                                 | Description            |
| ------ | ---------------------------------------- | ---------------------- |
| POST   | `/identity/link-farmer`                  | Link farmer to FPO     |
| POST   | `/identity/unlink-farmer`                | Unlink farmer from FPO |
| GET    | `/identity/linkage/{farmer_id}/{org_id}` | Get linkage status     |

### KisanSathi Management

| Method | Endpoint                                      | Description                   |
| ------ | --------------------------------------------- | ----------------------------- |
| POST   | `/kisansathi/assign`                          | Assign KisanSathi to farmer   |
| PUT    | `/kisansathi/reassign`                        | Reassign or remove KisanSathi |
| POST   | `/kisansathi/create-user`                     | Create new KisanSathi user    |
| GET    | `/kisansathi/assignment/{farmer_id}/{org_id}` | Get assignment details        |

### Admin & Health

| Method | Endpoint        | Description          |
| ------ | --------------- | -------------------- |
| GET    | `/admin/health` | Service health check |
| GET    | `/admin/audit`  | System audit logs    |

## Common Request Headers

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
X-Request-ID: <unique_request_id>
```

## Standard Response Format

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "request_id": "req_123456789",
  "data": {
    /* response data */
  }
}
```

## Error Response Format

```json
{
  "error": "Error category",
  "message": "Detailed error message",
  "request_id": "req_123456789",
  "details": {
    /* error details */
  }
}
```

## HTTP Status Codes

- **200**: Success (GET, PUT)
- **201**: Created (POST)
- **400**: Bad Request (validation errors)
- **401**: Unauthorized (authentication required)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found (resource not found)
- **409**: Conflict (resource already exists)
- **500**: Internal Server Error (system error)

## Interactive Documentation

- **Swagger UI**: http://localhost:8000/swagger/index.html
- **OpenAPI JSON**: http://localhost:8000/swagger/doc.json

## Example Requests

### Create FPO

```bash
curl -X POST "http://localhost:8000/api/v1/fpo/create" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Green Valley FPO",
    "registration_number": "FPO2024001",
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

### Link Farmer

```bash
curl -X POST "http://localhost:8000/api/v1/identity/link-farmer" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "aaa_user_id": "user_farmer456",
    "aaa_org_id": "org_fpo789"
  }'
```

### Assign KisanSathi

```bash
curl -X POST "http://localhost:8000/api/v1/kisansathi/assign" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "aaa_user_id": "user_farmer456",
    "aaa_org_id": "org_fpo789",
    "kisan_sathi_user_id": "user_ks123"
  }'
```
