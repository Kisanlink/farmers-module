# FPO Configuration API Guide

**For Client/Frontend Teams**

This guide explains how to properly create, update, and manage FPO configurations for ERP integration.

---

## Quick Reference

| Operation | Method | Endpoint | Auth Required |
|-----------|--------|----------|---------------|
| Get Config | GET | `/api/v1/fpo/{aaa_org_id}/configuration` | Yes |
| Create Config | POST | `/api/v1/fpo-config` | Yes |
| Update Config | PUT | `/api/v1/fpo/{aaa_org_id}/configuration` | Yes |
| Delete Config | DELETE | `/api/v1/fpo/{aaa_org_id}/configuration` | Yes |
| List Configs | GET | `/api/v1/fpo-config` | Yes (Admin) |
| Check ERP Health | GET | `/api/v1/fpo/{aaa_org_id}/configuration/health` | Yes |

---

## Important Notes

### 1. Endpoint Patterns

⚠️ **WRONG:**
```
PUT /api/v1/fpo-config/ORGN00000004     ❌ 404 - Route doesn't exist
POST /api/v1/fpo-config/ORGN00000004    ❌ Wrong pattern
```

✅ **CORRECT:**
```
GET  /api/v1/fpo/ORGN00000004/configuration      ✅ Get config
PUT  /api/v1/fpo/ORGN00000004/configuration      ✅ Update config
POST /api/v1/fpo-config                          ✅ Create config
```

### 2. Create vs Update Flow

```mermaid
graph TD
    A[Want to configure FPO] --> B[GET /api/v1/fpo/:id/configuration]
    B --> C{Response?}
    C -->|200 with config_status: not_configured| D[POST /api/v1/fpo-config]
    C -->|200 with existing config| E[PUT /api/v1/fpo/:id/configuration]
    D --> F[Config Created]
    E --> G[Config Updated]
```

**Flow:**
1. **Always GET first** to check if config exists
2. If `config_status: "not_configured"` in metadata → **POST to create**
3. If config has data → **PUT to update**

---

## API Endpoints

### 1. Get FPO Configuration

Get the current configuration for an FPO (or default config if not set up yet).

**Endpoint:** `GET /api/v1/fpo/{aaa_org_id}/configuration`

**Headers:**
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/fpo/ORGN00000004/configuration" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response (Config Exists):**
```json
{
  "success": true,
  "message": "FPO configuration retrieved successfully",
  "data": {
    "id": "ORGN00000004",
    "aaa_org_id": "ORGN00000004",
    "fpo_name": "Green Valley FPO",
    "erp_base_url": "https://erp.greenvalley.com",
    "erp_api_version": "v1",
    "features": {
      "inventory_sync": true,
      "order_management": true
    },
    "contact": {
      "email": "admin@greenvalley.com",
      "phone": "+919876543210"
    },
    "business_hours": {
      "monday": "09:00-17:00",
      "tuesday": "09:00-17:00"
    },
    "metadata": {
      "last_synced_at": "2025-11-19T10:30:00Z",
      "sync_interval_minutes": 30,
      "api_health_status": "healthy"
    },
    "api_health_status": "healthy",
    "last_synced_at": "2025-11-19T10:30:00Z",
    "sync_interval_minutes": 30,
    "created_at": "2025-11-01T10:00:00Z",
    "updated_at": "2025-11-19T10:30:00Z"
  },
  "request_id": "req_123456"
}
```

**Response (Config Not Set Up Yet):**
```json
{
  "success": true,
  "message": "FPO configuration retrieved successfully",
  "data": {
    "id": "ORGN00000004",
    "aaa_org_id": "ORGN00000004",
    "fpo_name": "",
    "erp_base_url": "",
    "erp_api_version": "",
    "features": {},
    "contact": {},
    "business_hours": {},
    "metadata": {
      "config_status": "not_configured",
      "message": "FPO configuration has not been set up yet"
    },
    "api_health_status": "not_configured",
    "last_synced_at": null,
    "sync_interval_minutes": 0
  },
  "request_id": "req_123456"
}
```

**Frontend Logic:**
```javascript
async function getFPOConfig(aaaOrgId) {
  const response = await fetch(`/api/v1/fpo/${aaaOrgId}/configuration`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    }
  });

  const result = await response.json();

  if (result.data.metadata?.config_status === 'not_configured') {
    // Show "Setup Configuration" button
    return { needsSetup: true, config: result.data };
  } else {
    // Show configuration details
    return { needsSetup: false, config: result.data };
  }
}
```

---

### 2. Create FPO Configuration

Create a new FPO configuration (first-time setup).

**Endpoint:** `POST /api/v1/fpo-config`

**Headers:**
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "aaa_org_id": "ORGN00000004",
  "fpo_name": "Green Valley FPO",
  "erp_base_url": "https://erp.greenvalley.com",
  "erp_api_version": "v1",
  "features": {
    "inventory_sync": true,
    "order_management": true,
    "farmer_portal": false
  },
  "contact": {
    "email": "admin@greenvalley.com",
    "phone": "+919876543210",
    "contact_person": "John Doe"
  },
  "business_hours": {
    "monday": "09:00-17:00",
    "tuesday": "09:00-17:00",
    "wednesday": "09:00-17:00",
    "thursday": "09:00-17:00",
    "friday": "09:00-17:00",
    "saturday": "09:00-13:00",
    "sunday": "closed"
  },
  "sync_interval": 30
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `aaa_org_id` | string | ✅ Yes | AAA Organization ID (e.g., "ORGN00000004") |
| `fpo_name` | string | ✅ Yes | Display name of the FPO |
| `erp_base_url` | string | ✅ Yes | Base URL of the ERP system (e.g., "https://erp.example.com") |
| `erp_api_version` | string | ❌ No | API version (default: "v1") |
| `features` | object | ❌ No | Feature flags as key-value pairs (JSON) |
| `contact` | object | ❌ No | Contact information (JSON) |
| `business_hours` | object | ❌ No | Operating hours (JSON) |
| `sync_interval` | number | ❌ No | Sync interval in minutes (default: 30) |

**Example Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/fpo-config" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "aaa_org_id": "ORGN00000004",
    "fpo_name": "Green Valley FPO",
    "erp_base_url": "https://erp.greenvalley.com",
    "erp_api_version": "v1",
    "features": {
      "inventory_sync": true,
      "order_management": true
    },
    "contact": {
      "email": "admin@greenvalley.com",
      "phone": "+919876543210"
    },
    "sync_interval": 30
  }'
```

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "FPO configuration created successfully",
  "data": {
    "id": "ORGN00000004",
    "aaa_org_id": "ORGN00000004",
    "fpo_name": "Green Valley FPO",
    "erp_base_url": "https://erp.greenvalley.com",
    "erp_api_version": "v1",
    "features": {
      "inventory_sync": true,
      "order_management": true
    },
    "contact": {
      "email": "admin@greenvalley.com",
      "phone": "+919876543210"
    },
    "business_hours": {},
    "metadata": {},
    "api_health_status": "unknown",
    "last_synced_at": null,
    "sync_interval_minutes": 30,
    "created_at": "2025-11-19T15:30:00Z",
    "updated_at": "2025-11-19T15:30:00Z"
  },
  "request_id": "req_789012"
}
```

**Error Response (400 Bad Request - Missing Required Fields):**
```json
{
  "success": false,
  "message": "Invalid request body",
  "error": "ERR_INVALID_INPUT"
}
```

**Error Response (409 Conflict - Already Exists):**
```json
{
  "success": false,
  "message": "FPO config already exists for aaa_org_id: ORGN00000004",
  "error": "ERR_ALREADY_EXISTS"
}
```

**Frontend Logic:**
```javascript
async function createFPOConfig(formData) {
  const response = await fetch('/api/v1/fpo-config', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      aaa_org_id: formData.aaaOrgId,
      fpo_name: formData.fpoName,
      erp_base_url: formData.erpBaseUrl,
      erp_api_version: formData.erpApiVersion || 'v1',
      features: formData.features || {},
      contact: formData.contact || {},
      business_hours: formData.businessHours || {},
      sync_interval: parseInt(formData.syncInterval) || 30
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to create configuration');
  }

  return await response.json();
}
```

---

### 3. Update FPO Configuration

Update an existing FPO configuration.

**Endpoint:** `PUT /api/v1/fpo/{aaa_org_id}/configuration`

**Headers:**
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body (Partial Update - All Fields Optional):**
```json
{
  "fpo_name": "Green Valley Farmers Producer Organization",
  "erp_base_url": "https://new-erp.greenvalley.com",
  "features": {
    "inventory_sync": true,
    "order_management": true,
    "farmer_portal": true
  },
  "sync_interval": 15
}
```

**Field Descriptions:**

All fields are **optional** - only send fields you want to update.

| Field | Type | Description |
|-------|------|-------------|
| `fpo_name` | string | Updated display name |
| `erp_base_url` | string | Updated ERP base URL |
| `erp_api_version` | string | Updated API version |
| `features` | object | Updated feature flags (replaces entire object) |
| `contact` | object | Updated contact info (replaces entire object) |
| `business_hours` | object | Updated business hours (replaces entire object) |
| `sync_interval` | number | Updated sync interval in minutes |

**Example Request:**
```bash
curl -X PUT "http://localhost:8080/api/v1/fpo/ORGN00000004/configuration" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "fpo_name": "Green Valley Farmers Producer Organization",
    "sync_interval": 15
  }'
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "FPO configuration updated successfully",
  "data": {
    "id": "ORGN00000004",
    "aaa_org_id": "ORGN00000004",
    "fpo_name": "Green Valley Farmers Producer Organization",
    "erp_base_url": "https://erp.greenvalley.com",
    "erp_api_version": "v1",
    "features": {
      "inventory_sync": true,
      "order_management": true
    },
    "contact": {
      "email": "admin@greenvalley.com",
      "phone": "+919876543210"
    },
    "business_hours": {},
    "metadata": {},
    "api_health_status": "healthy",
    "last_synced_at": "2025-11-19T10:30:00Z",
    "sync_interval_minutes": 15,
    "created_at": "2025-11-01T10:00:00Z",
    "updated_at": "2025-11-19T15:45:00Z"
  },
  "request_id": "req_345678"
}
```

**Error Response (404 Not Found):**
```json
{
  "success": false,
  "message": "FPO configuration not found",
  "error": "ERR_NOT_FOUND"
}
```

**Frontend Logic:**
```javascript
async function updateFPOConfig(aaaOrgId, updates) {
  const response = await fetch(`/api/v1/fpo/${aaaOrgId}/configuration`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(updates) // Only send changed fields
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to update configuration');
  }

  return await response.json();
}
```

---

### 4. Delete FPO Configuration

Soft-delete an FPO configuration.

**Endpoint:** `DELETE /api/v1/fpo/{aaa_org_id}/configuration`

**Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Example Request:**
```bash
curl -X DELETE "http://localhost:8080/api/v1/fpo/ORGN00000004/configuration" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "FPO configuration deleted successfully",
  "request_id": "req_901234"
}
```

---

### 5. Check ERP Health

Check if the FPO's ERP system is reachable and healthy.

**Endpoint:** `GET /api/v1/fpo/{aaa_org_id}/configuration/health`

**Headers:**
```http
Authorization: Bearer <jwt_token>
```

**Example Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/fpo/ORGN00000004/configuration/health" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK - Healthy):**
```json
{
  "success": true,
  "data": {
    "aaa_org_id": "ORGN00000004",
    "erp_base_url": "https://erp.greenvalley.com",
    "status": "healthy",
    "last_checked": "2025-11-19T15:50:00Z",
    "response_time_ms": 145,
    "error": null
  },
  "request_id": "req_567890"
}
```

**Response (ERP Unreachable):**
```json
{
  "success": true,
  "data": {
    "aaa_org_id": "ORGN00000004",
    "erp_base_url": "https://erp.greenvalley.com",
    "status": "unhealthy",
    "last_checked": "2025-11-19T15:50:00Z",
    "response_time_ms": 5000,
    "error": "dial tcp: lookup erp.greenvalley.com: no such host"
  },
  "request_id": "req_567891"
}
```

---

## Complete Frontend Example

### React Component

```typescript
import { useState, useEffect } from 'react';

interface FPOConfig {
  id: string;
  aaa_org_id: string;
  fpo_name: string;
  erp_base_url: string;
  erp_api_version: string;
  features: Record<string, boolean>;
  contact: Record<string, string>;
  business_hours: Record<string, string>;
  metadata: Record<string, any>;
  api_health_status: string;
  sync_interval_minutes: number;
}

export function FPOConfigurationPage({ aaaOrgId }: { aaaOrgId: string }) {
  const [config, setConfig] = useState<FPOConfig | null>(null);
  const [needsSetup, setNeedsSetup] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadConfig();
  }, [aaaOrgId]);

  async function loadConfig() {
    try {
      const response = await fetch(`/api/v1/fpo/${aaaOrgId}/configuration`, {
        headers: {
          'Authorization': `Bearer ${getToken()}`,
        }
      });

      const result = await response.json();

      if (result.data.metadata?.config_status === 'not_configured') {
        setNeedsSetup(true);
        setConfig(null);
      } else {
        setNeedsSetup(false);
        setConfig(result.data);
      }
    } catch (error) {
      console.error('Failed to load config:', error);
    } finally {
      setLoading(false);
    }
  }

  async function createConfig(formData: any) {
    try {
      const response = await fetch('/api/v1/fpo-config', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getToken()}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          aaa_org_id: aaaOrgId,
          fpo_name: formData.fpoName,
          erp_base_url: formData.erpBaseUrl,
          erp_api_version: formData.erpApiVersion || 'v1',
          features: formData.features || {},
          contact: formData.contact || {},
          business_hours: formData.businessHours || {},
          sync_interval: parseInt(formData.syncInterval) || 30
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message);
      }

      await loadConfig(); // Reload after creation
      alert('Configuration created successfully!');
    } catch (error) {
      alert(`Error: ${error.message}`);
    }
  }

  async function updateConfig(updates: Partial<FPOConfig>) {
    try {
      const response = await fetch(`/api/v1/fpo/${aaaOrgId}/configuration`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${getToken()}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message);
      }

      await loadConfig(); // Reload after update
      alert('Configuration updated successfully!');
    } catch (error) {
      alert(`Error: ${error.message}`);
    }
  }

  if (loading) return <div>Loading...</div>;

  if (needsSetup) {
    return <SetupConfigForm onSubmit={createConfig} />;
  }

  return <ConfigurationDetails config={config} onUpdate={updateConfig} />;
}
```

---

## Common Errors and Solutions

### Error: 404 on PUT /api/v1/fpo-config/{id}

**Problem:** Using wrong endpoint pattern

**Solution:** Use `PUT /api/v1/fpo/{id}/configuration` instead

### Error: 400 Invalid request body on POST

**Problem:** Missing required fields or wrong JSON structure

**Solution:** Ensure you send at minimum:
```json
{
  "aaa_org_id": "ORGN00000004",
  "fpo_name": "Your FPO Name",
  "erp_base_url": "https://your-erp-url.com"
}
```

### Error: 409 Already exists

**Problem:** Trying to POST when config already exists

**Solution:** Use PUT to update instead, or DELETE first then POST

### Error: 404 on UPDATE

**Problem:** Config doesn't exist yet

**Solution:** Use POST to create first

---

## Validation Rules

### Required Fields (POST)
- `aaa_org_id`: Non-empty string, must match AAA organization ID pattern
- `fpo_name`: Non-empty string
- `erp_base_url`: Valid URL format (http:// or https://)

### Optional Fields
- `erp_api_version`: String (default: "v1")
- `features`: Valid JSON object
- `contact`: Valid JSON object
- `business_hours`: Valid JSON object
- `sync_interval`: Positive integer (minutes)

### URL Validation
```regex
^https?:\/\/.+
```

---

## Testing with Postman/cURL

### Collection Variables
```json
{
  "base_url": "http://localhost:8080",
  "token": "your-jwt-token-here",
  "aaa_org_id": "ORGN00000004"
}
```

### Test Flow
1. GET config → Check if needs setup
2. If not configured → POST to create
3. If configured → PUT to update
4. GET health → Verify ERP connectivity
5. DELETE → Remove configuration

---

## Support

**Issues?**
- Check this guide
- Review server logs at `logs/farmers-module.log`
- Contact backend team with request_id from error responses

**Last Updated:** 2025-11-19
