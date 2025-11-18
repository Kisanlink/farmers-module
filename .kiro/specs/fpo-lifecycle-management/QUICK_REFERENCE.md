# FPO Lifecycle Management - Quick Reference

## Quick Links
- **Full Documentation**: [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)
- **Architecture**: [ARCHITECTURE_DESIGN.md](./ARCHITECTURE_DESIGN.md)
- **Implementation**: [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)

---

## Key Endpoints (Copy-Paste Ready)

### 1. Sync FPO from AAA (Most Important!)
```bash
POST /api/v1/identity/fpo/sync/:aaa_org_id
```

**Use When**: You get "FPO reference not found" error

**cURL Example**:
```bash
curl -X POST https://api.example.com/api/v1/identity/fpo/sync/org_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**JavaScript**:
```javascript
await fetch(`${API_URL}/identity/fpo/sync/${aaaOrgId}`, {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});
```

---

### 2. Get FPO (With Auto-Sync)
```bash
GET /api/v1/identity/fpo/by-org/:aaa_org_id
```

**JavaScript**:
```javascript
const response = await fetch(`${API_URL}/identity/fpo/by-org/${aaaOrgId}`, {
  headers: { 'Authorization': `Bearer ${token}` }
});
const fpo = await response.json();
```

---

### 3. Retry Failed Setup
```bash
POST /api/v1/identity/fpo/:id/retry-setup
```

**JavaScript**:
```javascript
await fetch(`${API_URL}/identity/fpo/${fpoId}/retry-setup`, {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});
```

---

### 4. Suspend FPO
```bash
PUT /api/v1/identity/fpo/:id/suspend
```

**JavaScript**:
```javascript
await fetch(`${API_URL}/identity/fpo/${fpoId}/suspend`, {
  method: 'PUT',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ reason: 'Compliance violation' })
});
```

---

### 5. Get Audit History
```bash
GET /api/v1/identity/fpo/:id/history
```

**JavaScript**:
```javascript
const response = await fetch(`${API_URL}/identity/fpo/${fpoId}/history`, {
  headers: { 'Authorization': `Bearer ${token}` }
});
const history = await response.json();
```

---

## FPO Status Values

```typescript
type FPOStatus =
  | 'DRAFT'                  // Initial creation
  | 'PENDING_VERIFICATION'   // Awaiting approval
  | 'VERIFIED'               // Approved, ready for setup
  | 'REJECTED'               // Verification failed
  | 'PENDING_SETUP'          // AAA setup in progress
  | 'SETUP_FAILED'           // Setup encountered errors
  | 'ACTIVE'                 // Fully operational ✓
  | 'SUSPENDED'              // Temporarily disabled
  | 'INACTIVE'               // Permanently disabled
  | 'ARCHIVED';              // Historical record
```

---

## Common Code Snippets

### Handle "Not Found" Error with Auto-Sync

```javascript
async function getFPO(aaaOrgId) {
  try {
    // Try regular endpoint
    const response = await fetch(`/api/v1/identity/fpo/reference/${aaaOrgId}`);
    if (!response.ok && response.status === 404) {
      // Auto-sync if not found
      return await syncFPO(aaaOrgId);
    }
    return await response.json();
  } catch (error) {
    console.error('Failed to get FPO:', error);
    throw error;
  }
}

async function syncFPO(aaaOrgId) {
  const response = await fetch(`/api/v1/identity/fpo/sync/${aaaOrgId}`, {
    method: 'POST'
  });
  return await response.json();
}
```

---

### Status Badge Helper

```javascript
function getStatusBadge(status) {
  const config = {
    'ACTIVE': { color: 'green', icon: '✓', text: 'Active' },
    'SETUP_FAILED': { color: 'red', icon: '✗', text: 'Setup Failed' },
    'PENDING_SETUP': { color: 'yellow', icon: '⏳', text: 'Setting Up' },
    'SUSPENDED': { color: 'orange', icon: '⏸', text: 'Suspended' },
    'DRAFT': { color: 'gray', icon: '📝', text: 'Draft' }
  };

  return config[status] || { color: 'gray', icon: '?', text: status };
}
```

---

### Check if Action is Allowed

```javascript
function canPerformAction(currentStatus, action) {
  const allowedActions = {
    'ACTIVE': ['suspend', 'deactivate'],
    'SUSPENDED': ['reactivate', 'deactivate'],
    'SETUP_FAILED': ['retry'],
    'INACTIVE': ['archive']
  };

  return allowedActions[currentStatus]?.includes(action) || false;
}
```

---

## State Transition Chart (Simple)

```
DRAFT → PENDING_VERIFICATION → VERIFIED → PENDING_SETUP → ACTIVE
                    ↓              ↓            ↓
                REJECTED       ARCHIVED   SETUP_FAILED
                                              ↓ (retry)
                                        PENDING_SETUP

ACTIVE → SUSPENDED → ACTIVE (reactivate)
   ↓         ↓
INACTIVE ← ←
   ↓
ARCHIVED
```

---

## Migration Checklist

### Backend Team
- [x] Run database migration: `002_fpo_lifecycle_management.sql`
- [x] Deploy updated farmers-module service
- [x] Verify new endpoints are accessible
- [x] Test sync endpoint with sample org ID
- [ ] Update API documentation/Swagger
- [ ] Monitor error logs for issues
- [ ] Set up alerts for SETUP_FAILED status

### Frontend Team
- [ ] Update API service with new endpoints
- [ ] Add all 10 status values to types
- [ ] Implement sync fallback logic
- [ ] Add retry button for SETUP_FAILED
- [ ] Add admin controls (suspend/reactivate)
- [ ] Update status badges/indicators
- [ ] Add audit history view
- [ ] Test all error scenarios
- [ ] Update documentation

### Mobile Team
- [ ] Update API calls to new endpoints
- [ ] Add status enum (10 values)
- [ ] Implement sync retry
- [ ] Add UI for setup retry
- [ ] Add admin controls
- [ ] Update status indicators
- [ ] Test offline behavior
- [ ] Update models

---

## Testing Commands

### Test Sync Endpoint
```bash
# Replace with your values
ORG_ID="org_123abc"
TOKEN="your_jwt_token"

curl -X POST "https://api.example.com/api/v1/identity/fpo/sync/${ORG_ID}" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json"
```

### Test Get with Auto-Sync
```bash
curl -X GET "https://api.example.com/api/v1/identity/fpo/by-org/${ORG_ID}" \
  -H "Authorization: Bearer ${TOKEN}"
```

### Test Retry Setup
```bash
FPO_ID="FPOR_1234567890"

curl -X POST "https://api.example.com/api/v1/identity/fpo/${FPO_ID}/retry-setup" \
  -H "Authorization: Bearer ${TOKEN}"
```

### Test Audit History
```bash
curl -X GET "https://api.example.com/api/v1/identity/fpo/${FPO_ID}/history" \
  -H "Authorization: Bearer ${TOKEN}"
```

---

## Error Responses

### FPO Not Found (404)
```json
{
  "error": "FPO reference not found for organization ID: org_123. Consider using the FPO lifecycle sync endpoint: POST /identity/fpo/sync/org_123"
}
```

**Solution**: Call sync endpoint

---

### Invalid State Transition (400)
```json
{
  "error": "cannot transition from ARCHIVED to ACTIVE"
}
```

**Solution**: Check valid transitions for current state

---

### Max Retries Exceeded (400)
```json
{
  "error": "maximum setup retries (3) exceeded"
}
```

**Solution**: Contact support for manual intervention

---

## Performance Tips

1. **Cache FPO Data**: Cache for 5 minutes to reduce API calls
2. **Use Auto-Sync Endpoint**: Prefer `/by-org/:id` over `/reference/:id`
3. **Poll Wisely**: Use 3-5 second intervals for status updates
4. **Batch Requests**: Get multiple FPOs in single request when possible
5. **Handle Errors Gracefully**: Always implement sync fallback

---

## Support Contacts

- **API Issues**: backend-team@example.com
- **Integration Help**: dev-support@example.com
- **Documentation**: See [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-11-18 | Initial release |

---

**Quick Links**:
- Full Guide: [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)
- Architecture: [ARCHITECTURE_DESIGN.md](./ARCHITECTURE_DESIGN.md)
- State Machine: [STATE_DIAGRAM.md](./STATE_DIAGRAM.md)
- Error Recovery: [ERROR_RECOVERY.md](./ERROR_RECOVERY.md)
