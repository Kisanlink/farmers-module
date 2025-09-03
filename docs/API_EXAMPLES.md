# Farmers Module API Examples

This document provides comprehensive examples for using the Farmers Module API endpoints.

## Base URL
```
http://localhost:8000
```

## Authentication
Most endpoints require authentication. Include the JWT token in the Authorization header:
```bash
Authorization: Bearer <your-jwt-token>
```

## Content Type
All requests should include:
```bash
Content-Type: application/json
```

---

## 1. Farmer Management

### Create Farmer
```bash
curl -X POST http://localhost:8000/farmers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "username": "john_doe",
    "phone_number": "+1234567890",
    "country_code": "+1",
    "email": "john@example.com",
    "password": "securepassword123",
    "full_name": "John Doe",
    "role": "farmer"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Farmer created successfully",
  "request_id": "req_123456",
  "data": {
    "id": "farmer_001",
    "username": "john_doe",
    "phone_number": "+1234567890",
    "email": "john@example.com",
    "full_name": "John Doe",
    "role": "farmer",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Get Farmer by ID
```bash
curl -X GET http://localhost:8000/farmers/farmer_001 \
  -H "Authorization: Bearer <token>"
```

### Update Farmer
```bash
curl -X PUT http://localhost:8000/farmers/farmer_001 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "full_name": "John Smith",
    "email": "johnsmith@example.com"
  }'
```

### List Farmers
```bash
curl -X GET "http://localhost:8000/farmers?page=1&page_size=10&role=farmer" \
  -H "Authorization: Bearer <token>"
```

---

## 2. Farm Management

### Create Farm
```bash
curl -X POST http://localhost:8000/farms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Green Valley Farm",
    "farmer_id": "farmer_001",
    "org_id": "org_001",
    "geometry": {
      "type": "Polygon",
      "coordinates": [[
        [77.123456, 28.654321],
        [77.123456, 28.655321],
        [77.124456, 28.655321],
        [77.124456, 28.654321],
        [77.123456, 28.654321]
      ]]
    },
    "location": {
      "village": "Green Valley",
      "block": "Block A",
      "district": "Sample District",
      "state": "Sample State",
      "country": "India",
      "pincode": "123456"
    },
    "area_hectares": 2.5,
    "soil_type": "loamy",
    "irrigation_type": "drip"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Farm created successfully",
  "request_id": "req_123457",
  "data": {
    "id": "farm_001",
    "name": "Green Valley Farm",
    "farmer_id": "farmer_001",
    "org_id": "org_001",
    "area_hectares": 2.5,
    "soil_type": "loamy",
    "irrigation_type": "drip",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Get Farm by ID
```bash
curl -X GET http://localhost:8000/farms/farm_001 \
  -H "Authorization: Bearer <token>"
```

### Update Farm
```bash
curl -X PUT http://localhost:8000/farms/farm_001 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Green Valley Organic Farm",
    "area_hectares": 3.0,
    "soil_type": "organic_loamy"
  }'
```

### List Farms
```bash
curl -X GET "http://localhost:8000/farms?page=1&page_size=10&farmer_id=farmer_001" \
  -H "Authorization: Bearer <token>"
```

### Delete Farm
```bash
curl -X DELETE http://localhost:8000/farms/farm_001 \
  -H "Authorization: Bearer <token>"
```

---

## 3. Crop Cycle Management

### Create Crop Cycle
```bash
curl -X POST http://localhost:8000/crop-cycles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "farm_id": "farm_001",
    "crop_name": "Wheat",
    "variety": "HD-2967",
    "planting_date": "2024-01-15",
    "expected_harvest_date": "2024-04-15",
    "area_hectares": 2.5,
    "seed_quantity_kg": 50,
    "expected_yield_kg": 1000
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Crop cycle created successfully",
  "request_id": "req_123458",
  "data": {
    "id": "cycle_001",
    "farm_id": "farm_001",
    "crop_name": "Wheat",
    "variety": "HD-2967",
    "planting_date": "2024-01-15",
    "expected_harvest_date": "2024-04-15",
    "area_hectares": 2.5,
    "seed_quantity_kg": 50,
    "expected_yield_kg": 1000,
    "status": "planted",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Get Crop Cycle by ID
```bash
curl -X GET http://localhost:8000/crop-cycles/cycle_001 \
  -H "Authorization: Bearer <token>"
```

### Update Crop Cycle
```bash
curl -X PUT http://localhost:8000/crop-cycles/cycle_001 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "status": "growing",
    "actual_harvest_date": "2024-04-20",
    "actual_yield_kg": 1200
  }'
```

### List Crop Cycles
```bash
curl -X GET "http://localhost:8000/crop-cycles?page=1&page_size=10&farm_id=farm_001" \
  -H "Authorization: Bearer <token>"
```

---

## 4. Farm Activity Management

### Create Farm Activity
```bash
curl -X POST http://localhost:8000/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "crop_cycle_id": "cycle_001",
    "activity_type": "fertilization",
    "activity_name": "NPK Fertilizer Application",
    "scheduled_date": "2024-02-01",
    "description": "Application of NPK fertilizer for wheat crop",
    "input_requirements": {
      "fertilizer_type": "NPK 20-20-20",
      "quantity_kg": 25,
      "application_method": "broadcast"
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Farm activity created successfully",
  "request_id": "req_123459",
  "data": {
    "id": "activity_001",
    "crop_cycle_id": "cycle_001",
    "activity_type": "fertilization",
    "activity_name": "NPK Fertilizer Application",
    "scheduled_date": "2024-02-01",
    "status": "scheduled",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### Complete Farm Activity
```bash
curl -X PUT http://localhost:8000/activities/activity_001/complete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "actual_date": "2024-02-01",
    "output_data": {
      "fertilizer_used_kg": 25,
      "area_covered_hectares": 2.5,
      "weather_conditions": "clear",
      "notes": "Applied evenly across the field"
    }
  }'
```

### Get Activity by ID
```bash
curl -X GET http://localhost:8000/activities/activity_001 \
  -H "Authorization: Bearer <token>"
```

### List Activities
```bash
curl -X GET "http://localhost:8000/activities?page=1&page_size=10&crop_cycle_id=cycle_001" \
  -H "Authorization: Bearer <token>"
```

---

## 5. Administrative Operations

### Health Check
```bash
curl -X GET http://localhost:8000/admin/health
```

**Response:**
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "version": "1.0.0",
    "components": {
      "database": "healthy",
      "aaa_service": "healthy"
    }
  }
}
```

### Seed Roles and Permissions
```bash
curl -X POST http://localhost:8000/admin/seed \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{
    "force": false,
    "dry_run": false
  }'
```

### Check Permission
```bash
curl -X POST http://localhost:8000/admin/permissions/check \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "subject": "user_123",
    "resource": "farm",
    "action": "create",
    "object": "farm_001",
    "org_id": "org_001"
  }'
```

### Get Audit Trail
```bash
curl -X GET "http://localhost:8000/admin/audit?start_date=2024-01-01&end_date=2024-01-31&user_id=user_123" \
  -H "Authorization: Bearer <admin-token>"
```

---

## 6. Reporting Operations

### Export Farmer Portfolio
```bash
curl -X GET "http://localhost:8000/reports/farmer-portfolio?farmer_id=farmer_001&format=json" \
  -H "Authorization: Bearer <token>"
```

### Get Organization Dashboard Counters
```bash
curl -X GET "http://localhost:8000/reports/org-dashboard?org_id=org_001" \
  -H "Authorization: Bearer <token>"
```

---

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid request data",
  "error": "validation failed: phone_number is required",
  "request_id": "req_123460"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Authentication required",
  "error": "invalid or missing token",
  "request_id": "req_123461"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "message": "Access denied",
  "error": "insufficient permissions",
  "request_id": "req_123462"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Resource not found",
  "error": "farmer with id 'farmer_999' not found",
  "request_id": "req_123463"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Internal server error",
  "error": "database connection failed",
  "request_id": "req_123464"
}
```

---

## Pagination

Most list endpoints support pagination:

- `page`: Page number (default: 1)
- `page_size`: Number of items per page (default: 10, max: 100)

**Example:**
```bash
curl -X GET "http://localhost:8000/farmers?page=2&page_size=20" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "success": true,
  "message": "Farmers retrieved successfully",
  "data": [...],
  "page": 2,
  "page_size": 20,
  "total": 150
}
```

---

## Filtering

Many endpoints support filtering:

### Farmers
- `role`: Filter by role (farmer, admin, etc.)
- `status`: Filter by status (active, inactive)
- `org_id`: Filter by organization

### Farms
- `farmer_id`: Filter by farmer
- `org_id`: Filter by organization
- `min_area`: Minimum area in hectares
- `max_area`: Maximum area in hectares

### Crop Cycles
- `farm_id`: Filter by farm
- `crop_name`: Filter by crop name
- `status`: Filter by status (planted, growing, harvested)

### Activities
- `crop_cycle_id`: Filter by crop cycle
- `activity_type`: Filter by activity type
- `status`: Filter by status (scheduled, in_progress, completed)

---

## Rate Limiting

The API implements rate limiting to ensure fair usage:
- 100 requests per minute per IP address
- 1000 requests per hour per authenticated user

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

---

## WebSocket Support

Real-time updates are available via WebSocket connections:

```javascript
const ws = new WebSocket('ws://localhost:8000/ws/farm-updates');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Farm update:', data);
};
```

---

## SDK Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'http://localhost:8000',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  }
});

// Create a farmer
const createFarmer = async (farmerData) => {
  try {
    const response = await api.post('/farmers', farmerData);
    return response.data;
  } catch (error) {
    console.error('Error creating farmer:', error.response.data);
    throw error;
  }
};
```

### Python
```python
import requests

class FarmersAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }

    def create_farmer(self, farmer_data):
        response = requests.post(
            f'{self.base_url}/farmers',
            json=farmer_data,
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

# Usage
api = FarmersAPI('http://localhost:8000', 'your-token')
farmer = api.create_farmer({
    'username': 'john_doe',
    'phone_number': '+1234567890',
    'email': 'john@example.com',
    'full_name': 'John Doe'
})
```

---

## Testing

### Health Check
```bash
# Basic health check
curl http://localhost:8000/admin/health

# Detailed health check
curl "http://localhost:8000/admin/health?components=database,aaa_service"
```

### Load Testing
```bash
# Using Apache Bench
ab -n 1000 -c 10 -H "Authorization: Bearer <token>" http://localhost:8000/farmers

# Using curl with timing
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8000/farmers
```

---

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if the service is running on the correct port
   - Verify firewall settings

2. **Authentication Errors**
   - Ensure the JWT token is valid and not expired
   - Check token format: `Bearer <token>`

3. **Validation Errors**
   - Review request body format
   - Check required fields

4. **Rate Limiting**
   - Implement exponential backoff
   - Use connection pooling

### Debug Mode
Enable debug logging by setting:
```bash
export LOG_LEVEL=debug
```

### Logs
Check application logs for detailed error information:
```bash
tail -f /var/log/farmers-module/app.log
```
