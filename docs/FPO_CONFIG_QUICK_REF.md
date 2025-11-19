# FPO Configuration API - Quick Reference

**Send this to your client team immediately**

---

## ⚠️ CRITICAL: Endpoint Corrections

### ❌ WRONG (Causing 404s)
```
PUT  /api/v1/fpo-config/ORGN00000004          # 404 - doesn't exist
POST /api/v1/fpo-config/ORGN00000004          # 404 - doesn't exist
```

### ✅ CORRECT
```
GET    /api/v1/fpo/{id}/configuration         # Get config
POST   /api/v1/fpo-config                     # Create new config
PUT    /api/v1/fpo/{id}/configuration         # Update existing config
DELETE /api/v1/fpo/{id}/configuration         # Delete config
GET    /api/v1/fpo/{id}/configuration/health  # Check ERP health
```

---

## Flow: How to Create Configuration

```
1. GET  /api/v1/fpo/ORGN00000004/configuration
   └─> Check metadata.config_status

2a. If "not_configured":
    POST /api/v1/fpo-config
    {
      "aaa_org_id": "ORGN00000004",
      "fpo_name": "Green Valley FPO",
      "erp_base_url": "https://erp.example.com"
    }

2b. If config exists:
    PUT /api/v1/fpo/ORGN00000004/configuration
    {
      "fpo_name": "Updated Name"
    }
```

---

## Minimum Required JSON (POST)

```json
{
  "aaa_org_id": "ORGN00000004",
  "fpo_name": "Green Valley FPO",
  "erp_base_url": "https://erp.greenvalley.com"
}
```

**Optional fields:**
- `erp_api_version` (default: "v1")
- `features` (JSON object)
- `contact` (JSON object)
- `business_hours` (JSON object)
- `sync_interval` (number, default: 30)

---

## Full Example (POST)

```bash
curl -X POST "http://localhost:8080/api/v1/fpo-config" \
  -H "Authorization: Bearer YOUR_TOKEN" \
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

---

## Full Example (PUT)

```bash
curl -X PUT "http://localhost:8080/api/v1/fpo/ORGN00000004/configuration" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "fpo_name": "Updated FPO Name",
    "sync_interval": 15
  }'
```

---

## JavaScript/TypeScript Example

```typescript
// 1. Check if config exists
const checkResponse = await fetch(
  `/api/v1/fpo/${aaaOrgId}/configuration`,
  {
    headers: { 'Authorization': `Bearer ${token}` }
  }
);
const checkResult = await checkResponse.json();

// 2. Create or Update based on status
if (checkResult.data.metadata?.config_status === 'not_configured') {
  // CREATE
  await fetch('/api/v1/fpo-config', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      aaa_org_id: aaaOrgId,
      fpo_name: 'Green Valley FPO',
      erp_base_url: 'https://erp.example.com',
      erp_api_version: 'v1',
      sync_interval: 30
    })
  });
} else {
  // UPDATE
  await fetch(`/api/v1/fpo/${aaaOrgId}/configuration`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      fpo_name: 'Updated Name',
      sync_interval: 15
    })
  });
}
```

---

## Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| 404 on PUT | Wrong endpoint | Use `/api/v1/fpo/{id}/configuration` not `/api/v1/fpo-config/{id}` |
| 400 Invalid body | Missing required fields | Include `aaa_org_id`, `fpo_name`, `erp_base_url` |
| 409 Conflict | Config already exists | Use PUT to update instead of POST |
| 404 on UPDATE | Config doesn't exist | Use POST to create first |

---

## Need More Details?

See full documentation: `docs/FPO_CONFIG_API_GUIDE.md`

**Last Updated:** 2025-11-19
