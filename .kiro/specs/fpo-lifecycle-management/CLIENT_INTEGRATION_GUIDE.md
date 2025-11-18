# FPO Lifecycle Management - Client Integration Guide

**Version**: 1.0
**Date**: 2025-11-18
**Author**: Farmers Module Team

## Overview

This document provides comprehensive guidance for client applications to integrate with the new FPO lifecycle management system. The system solves the "failed to get FPO reference: no matching records found" error and provides complete lifecycle management for FPO organizations.

---

## Table of Contents

1. [What's New](#whats-new)
2. [Breaking Changes](#breaking-changes)
3. [Migration Guide](#migration-guide)
4. [API Endpoints](#api-endpoints)
5. [FPO Status Lifecycle](#fpo-status-lifecycle)
6. [Error Handling](#error-handling)
7. [Client Implementation Examples](#client-implementation-examples)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

---

## What's New

### Key Features

1. **Automatic FPO Sync from AAA** - Resolves missing FPO reference errors
2. **Complete Lifecycle Management** - 10 lifecycle states with validated transitions
3. **Audit Trail** - Full history of all FPO state changes
4. **Setup Retry Logic** - Automatic retry for failed setups (max 3 attempts)
5. **Suspend/Reactivate** - Administrative controls for FPO operations
6. **Enhanced Error Messages** - Helpful guidance when FPO not found

### New FPO Statuses

| Status | Description | Client Actions |
|--------|-------------|----------------|
| `DRAFT` | Initial FPO creation | Can edit all fields |
| `PENDING_VERIFICATION` | Awaiting document verification | Read-only, wait for approval |
| `VERIFIED` | Documents verified | Ready for setup |
| `REJECTED` | Verification failed | Fix issues and resubmit |
| `PENDING_SETUP` | AAA setup in progress | Monitor setup progress |
| `SETUP_FAILED` | Setup encountered errors | Retry setup operation |
| `ACTIVE` | Fully operational | All operations enabled |
| `SUSPENDED` | Temporarily suspended | Limited operations |
| `INACTIVE` | Permanently deactivated | Read-only access |
| `ARCHIVED` | Historical record | Historical queries only |

---

## Breaking Changes

### ⚠️ IMPORTANT: Field Name Changes

**Previous**: `registration_no`
**Now**: `registration_number` (both accepted for backward compatibility)

**Action Required**: Update client code to use `registration_number` in new integrations.

### Status Field Changes

**Previous Default**: `ACTIVE`
**New Default**: `DRAFT`

**Migration**: Existing FPOs remain `ACTIVE`. Only new FPOs start as `DRAFT`.

### New Required Fields

When creating FPOs, these fields are now tracked:
- `status` - Lifecycle status (auto-managed)
- `status_reason` - Reason for current status
- `status_changed_at` - Timestamp of last change
- `ceo_user_id` - CEO's AAA user ID

**Action Required**: None - these fields are auto-populated by the system.

---

## Migration Guide

### Step 1: Update Client Dependencies

No client dependency changes required. All changes are backward compatible.

### Step 2: Handle New Status Values

Update your client's status handling to recognize new statuses:

```javascript
// OLD CODE
const isActive = fpo.status === 'ACTIVE';

// NEW CODE - More robust
const isOperational = ['ACTIVE', 'SUSPENDED'].includes(fpo.status);
const needsAction = ['SETUP_FAILED', 'REJECTED'].includes(fpo.status);
const isPending = ['DRAFT', 'PENDING_VERIFICATION', 'PENDING_SETUP'].includes(fpo.status);
```

### Step 3: Implement Error Recovery

Replace error handling with sync fallback:

```javascript
// OLD CODE - Just show error
async function getFPO(orgId) {
  const response = await fetch(`/identity/fpo/reference/${orgId}`);
  if (!response.ok) {
    throw new Error('FPO not found');
  }
  return response.json();
}

// NEW CODE - Auto-sync on error
async function getFPO(orgId) {
  try {
    const response = await fetch(`/identity/fpo/reference/${orgId}`);
    if (!response.ok) {
      // Try sync endpoint if not found
      if (response.status === 404) {
        console.log('FPO not found locally, syncing from AAA...');
        return await syncFPOFromAAA(orgId);
      }
      throw new Error('FPO request failed');
    }
    return response.json();
  } catch (error) {
    console.error('Failed to get FPO:', error);
    throw error;
  }
}

async function syncFPOFromAAA(orgId) {
  const response = await fetch(`/identity/fpo/sync/${orgId}`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${getToken()}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    throw new Error('Failed to sync FPO from AAA');
  }

  return response.json();
}
```

### Step 4: Display Status Information

Add status badges/indicators in your UI:

```javascript
function getFPOStatusBadge(status) {
  const statusConfig = {
    'ACTIVE': { color: 'green', text: 'Active' },
    'DRAFT': { color: 'gray', text: 'Draft' },
    'PENDING_VERIFICATION': { color: 'yellow', text: 'Pending Verification' },
    'VERIFIED': { color: 'blue', text: 'Verified' },
    'REJECTED': { color: 'red', text: 'Rejected' },
    'PENDING_SETUP': { color: 'yellow', text: 'Setup in Progress' },
    'SETUP_FAILED': { color: 'orange', text: 'Setup Failed' },
    'SUSPENDED': { color: 'orange', text: 'Suspended' },
    'INACTIVE': { color: 'gray', text: 'Inactive' },
    'ARCHIVED': { color: 'gray', text: 'Archived' }
  };

  return statusConfig[status] || { color: 'gray', text: status };
}
```

---

## API Endpoints

### Base URL
```
https://your-api-domain.com/api/v1/identity/fpo
```

### Authentication
All endpoints require Bearer token authentication:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

### 1. Sync FPO from AAA (NEW - Key Endpoint)

**Resolves the "no matching records found" error**

```http
POST /identity/fpo/sync/:aaa_org_id
```

#### Purpose
Synchronizes FPO reference from AAA service to local database. Use this when you encounter "FPO reference not found" errors.

#### Request

```bash
curl -X POST https://api.example.com/api/v1/identity/fpo/sync/org_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "FPO synchronized successfully",
  "data": {
    "fpo_id": "FPOR_1234567890",
    "aaa_org_id": "org_abc123",
    "name": "Rampur Farmers Producer Company",
    "registration_number": "FPO/MP/2024/001234",
    "status": "ACTIVE",
    "created_at": "2025-11-18T10:00:00Z",
    "updated_at": "2025-11-18T10:00:00Z"
  }
}
```

#### Error Responses

```json
// 400 Bad Request - Missing org ID
{
  "success": false,
  "message": "AAA organization ID is required"
}

// 404 Not Found - Org doesn't exist in AAA
{
  "success": false,
  "message": "failed to get organization from AAA: organization not found"
}

// 500 Internal Server Error
{
  "success": false,
  "message": "failed to create FPO reference: database error"
}
```

#### When to Use
- After creating an FPO in AAA service directly
- When getting "FPO reference not found" errors
- During system migration/data reconciliation
- When local database was cleared/reset

---

### 2. Get FPO by AAA Org ID (NEW - With Auto-Sync)

```http
GET /identity/fpo/by-org/:aaa_org_id
```

#### Purpose
Retrieves FPO by AAA organization ID. Automatically syncs from AAA if not found locally.

#### Request

```bash
curl -X GET https://api.example.com/api/v1/identity/fpo/by-org/org_abc123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "fpo_id": "FPOR_1234567890",
    "aaa_org_id": "org_abc123",
    "name": "Rampur Farmers Producer Company",
    "registration_number": "FPO/MP/2024/001234",
    "status": "ACTIVE",
    "business_config": {
      "max_farmers": 1000,
      "procurement_enabled": true
    },
    "metadata": {
      "region": "Madhya Pradesh",
      "district": "Rampur"
    },
    "created_at": "2025-11-18T10:00:00Z",
    "updated_at": "2025-11-18T10:00:00Z"
  }
}
```

#### When to Use
- **Primary endpoint for getting FPO by org ID**
- Replaces `/identity/fpo/reference/:aaa_org_id` for better error recovery
- Use in all new client code

---

### 3. Retry Failed Setup (NEW)

```http
POST /identity/fpo/:id/retry-setup
```

#### Purpose
Retries setup operations for FPOs in `SETUP_FAILED` status.

#### Request

```bash
curl -X POST https://api.example.com/api/v1/identity/fpo/FPOR_1234567890/retry-setup \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "FPO setup retry initiated successfully"
}
```

#### Error Responses

```json
// 400 Bad Request - Not in SETUP_FAILED status
{
  "success": false,
  "message": "FPO is not in SETUP_FAILED status, current status: ACTIVE"
}

// 400 Bad Request - Max retries exceeded
{
  "success": false,
  "message": "maximum setup retries (3) exceeded"
}
```

#### When to Use
- When FPO creation partially fails
- After fixing AAA service connectivity issues
- When user groups or permissions failed to create

---

### 4. Suspend FPO (NEW)

```http
PUT /identity/fpo/:id/suspend
```

#### Purpose
Temporarily suspends an FPO (compliance issues, administrative holds).

#### Request

```bash
curl -X PUT https://api.example.com/api/v1/identity/fpo/FPOR_1234567890/suspend \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Compliance violation - pending investigation"
  }'
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "FPO suspended successfully"
}
```

#### Error Responses

```json
// 400 Bad Request - Invalid transition
{
  "success": false,
  "message": "cannot transition from INACTIVE to SUSPENDED"
}

// 400 Bad Request - Missing reason
{
  "success": false,
  "message": "Invalid request body"
}
```

#### When to Use
- Compliance violations
- Pending investigations
- Administrative holds
- Temporary operational issues

---

### 5. Reactivate FPO (NEW)

```http
PUT /identity/fpo/:id/reactivate
```

#### Purpose
Reactivates a suspended FPO.

#### Request

```bash
curl -X PUT https://api.example.com/api/v1/identity/fpo/FPOR_1234567890/reactivate \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "FPO reactivated successfully"
}
```

#### When to Use
- After resolving compliance issues
- Completing investigations
- Lifting administrative holds

---

### 6. Deactivate FPO (NEW)

```http
DELETE /identity/fpo/:id/deactivate
```

#### Purpose
Permanently deactivates an FPO.

#### Request

```bash
curl -X DELETE https://api.example.com/api/v1/identity/fpo/FPOR_1234567890/deactivate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Business closure - voluntary dissolution"
  }'
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "FPO deactivated successfully"
}
```

#### When to Use
- Business closure
- Voluntary dissolution
- Permanent shutdown

---

### 7. Get FPO History (NEW)

```http
GET /identity/fpo/:id/history
```

#### Purpose
Retrieves complete audit trail of FPO lifecycle changes.

#### Request

```bash
curl -X GET https://api.example.com/api/v1/identity/fpo/FPOR_1234567890/history \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": [
    {
      "action": "STATUS_CHANGE",
      "previous_state": "PENDING_SETUP",
      "new_state": "ACTIVE",
      "reason": "Setup completed successfully",
      "performed_by": "system",
      "performed_at": "2025-11-18T10:05:00Z",
      "details": {
        "user_groups_created": 4,
        "setup_attempts": 1
      }
    },
    {
      "action": "STATUS_CHANGE",
      "previous_state": "VERIFIED",
      "new_state": "PENDING_SETUP",
      "reason": "Starting AAA setup",
      "performed_by": "system",
      "performed_at": "2025-11-18T10:00:00Z",
      "details": null
    }
  ]
}
```

#### When to Use
- Compliance audits
- Debugging lifecycle issues
- Understanding FPO status history
- Generating reports

---

### 8. Existing Endpoints (No Changes)

These endpoints continue to work as before:

```http
# Create FPO (existing)
POST /identity/fpo/create

# Register FPO Reference (existing)
POST /identity/fpo/register

# Get FPO Reference (existing - but consider using /by-org instead)
GET /identity/fpo/reference/:aaa_org_id
```

---

## FPO Status Lifecycle

### State Transition Diagram

```
┌─────────┐
│  DRAFT  │
└────┬────┘
     │ submit
     ▼
┌──────────────────────┐
│ PENDING_VERIFICATION │
└──────┬───────────────┘
       │
       ├─── approve ──→ ┌──────────┐
       │                │ VERIFIED │
       │                └────┬─────┘
       │                     │ initialize setup
       │                     ▼
       │                ┌──────────────┐
       │                │PENDING_SETUP │
       │                └──────┬───────┘
       │                       │
       │                       ├─── success ──→ ┌────────┐
       │                       │                │ ACTIVE │
       │                       │                └───┬────┘
       │                       │                    │
       │                       │                    ├─ suspend ─→ ┌───────────┐
       │                       │                    │             │ SUSPENDED │
       │                       │                    │             └─────┬─────┘
       │                       │                    │                   │
       │                       │                    │                   └─ reactivate ─→ back to ACTIVE
       │                       │                    │
       │                       │                    └─ deactivate ─→ ┌──────────┐
       │                       │                                     │ INACTIVE │
       │                       │                                     └────┬─────┘
       │                       │                                          │
       │                       └─── failure ──→ ┌──────────────┐         │
       │                                        │ SETUP_FAILED │         │
       │                                        └──────┬───────┘         │
       │                                               │                 │
       │                                               └─ retry (max 3x) │
       │                                                                 │
       └─── reject ──→ ┌──────────┐                                    │
                       │ REJECTED │                                    │
                       └──────────┘                                    │
                                                                       │
                                                                       ▼
                                                                  ┌──────────┐
                                                                  │ ARCHIVED │
                                                                  └──────────┘
```

### Valid State Transitions

| From | To | Condition | API Endpoint |
|------|----|-----------| ------------|
| DRAFT | PENDING_VERIFICATION | All fields complete | TBD |
| PENDING_VERIFICATION | VERIFIED | Documents approved | TBD |
| PENDING_VERIFICATION | REJECTED | Documents rejected | TBD |
| VERIFIED | PENDING_SETUP | Setup initiated | TBD |
| PENDING_SETUP | ACTIVE | Setup successful | Automatic |
| PENDING_SETUP | SETUP_FAILED | Setup errors | Automatic |
| SETUP_FAILED | PENDING_SETUP | Retry (max 3) | POST /:id/retry-setup |
| ACTIVE | SUSPENDED | Admin action | PUT /:id/suspend |
| SUSPENDED | ACTIVE | Issue resolved | PUT /:id/reactivate |
| ACTIVE | INACTIVE | Deactivation | DELETE /:id/deactivate |
| INACTIVE | ARCHIVED | Archival | TBD |

---

## Error Handling

### Common Error Scenarios

#### 1. FPO Not Found (404)

**Old Behavior**:
```json
{
  "error": "failed to get FPO reference: no matching records found"
}
```

**New Behavior**:
```json
{
  "error": "FPO reference not found for organization ID: org_abc123. Consider using the FPO lifecycle sync endpoint: POST /identity/fpo/sync/org_abc123"
}
```

**Client Action**:
1. Parse error message to extract org ID
2. Call sync endpoint: `POST /identity/fpo/sync/org_abc123`
3. Retry original operation
4. If still fails, show error to user

#### 2. Invalid State Transition (400)

```json
{
  "error": "cannot transition from ARCHIVED to ACTIVE"
}
```

**Client Action**:
- Show error to user
- Display current state and valid next states
- Disable invalid action buttons in UI

#### 3. Setup Failed (SETUP_FAILED status)

**Client Action**:
1. Check FPO status
2. If `SETUP_FAILED`, show retry button
3. Call `POST /identity/fpo/:id/retry-setup`
4. Monitor status changes via polling or webhooks

---

## Client Implementation Examples

### React Example

```typescript
// services/fpoService.ts
import axios from 'axios';

const API_BASE = process.env.REACT_APP_API_URL;

interface FPO {
  fpo_id: string;
  aaa_org_id: string;
  name: string;
  registration_number: string;
  status: FPOStatus;
  created_at: string;
  updated_at: string;
}

type FPOStatus =
  | 'DRAFT'
  | 'PENDING_VERIFICATION'
  | 'VERIFIED'
  | 'REJECTED'
  | 'PENDING_SETUP'
  | 'SETUP_FAILED'
  | 'ACTIVE'
  | 'SUSPENDED'
  | 'INACTIVE'
  | 'ARCHIVED';

export class FPOService {
  private getHeaders() {
    const token = localStorage.getItem('auth_token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  }

  /**
   * Get FPO by org ID with automatic sync fallback
   */
  async getFPOByOrgId(aaaOrgId: string): Promise<FPO> {
    try {
      const response = await axios.get(
        `${API_BASE}/identity/fpo/by-org/${aaaOrgId}`,
        { headers: this.getHeaders() }
      );
      return response.data.data;
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.status === 404) {
        // Try explicit sync as fallback
        console.log('FPO not found, attempting sync...');
        return await this.syncFPOFromAAA(aaaOrgId);
      }
      throw error;
    }
  }

  /**
   * Manually sync FPO from AAA
   */
  async syncFPOFromAAA(aaaOrgId: string): Promise<FPO> {
    const response = await axios.post(
      `${API_BASE}/identity/fpo/sync/${aaaOrgId}`,
      {},
      { headers: this.getHeaders() }
    );
    return response.data.data;
  }

  /**
   * Retry failed setup
   */
  async retrySetup(fpoId: string): Promise<void> {
    await axios.post(
      `${API_BASE}/identity/fpo/${fpoId}/retry-setup`,
      {},
      { headers: this.getHeaders() }
    );
  }

  /**
   * Suspend FPO
   */
  async suspendFPO(fpoId: string, reason: string): Promise<void> {
    await axios.put(
      `${API_BASE}/identity/fpo/${fpoId}/suspend`,
      { reason },
      { headers: this.getHeaders() }
    );
  }

  /**
   * Reactivate FPO
   */
  async reactivateFPO(fpoId: string): Promise<void> {
    await axios.put(
      `${API_BASE}/identity/fpo/${fpoId}/reactivate`,
      {},
      { headers: this.getHeaders() }
    );
  }

  /**
   * Get FPO history
   */
  async getHistory(fpoId: string): Promise<AuditEntry[]> {
    const response = await axios.get(
      `${API_BASE}/identity/fpo/${fpoId}/history`,
      { headers: this.getHeaders() }
    );
    return response.data.data;
  }
}
```

### React Component Example

```typescript
// components/FPODetails.tsx
import React, { useState, useEffect } from 'react';
import { FPOService } from '../services/fpoService';

interface Props {
  aaaOrgId: string;
}

export const FPODetails: React.FC<Props> = ({ aaaOrgId }) => {
  const [fpo, setFpo] = useState<FPO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);

  const fpoService = new FPOService();

  useEffect(() => {
    loadFPO();
  }, [aaaOrgId]);

  const loadFPO = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fpoService.getFPOByOrgId(aaaOrgId);
      setFpo(data);
    } catch (err) {
      setError(err.message || 'Failed to load FPO');
    } finally {
      setLoading(false);
    }
  };

  const handleRetrySetup = async () => {
    if (!fpo) return;

    try {
      setSyncing(true);
      await fpoService.retrySetup(fpo.fpo_id);
      // Poll for status update
      setTimeout(loadFPO, 2000);
    } catch (err) {
      setError(err.message || 'Failed to retry setup');
    } finally {
      setSyncing(false);
    }
  };

  const handleSuspend = async () => {
    if (!fpo) return;

    const reason = prompt('Enter suspension reason:');
    if (!reason) return;

    try {
      await fpoService.suspendFPO(fpo.fpo_id, reason);
      loadFPO();
    } catch (err) {
      setError(err.message || 'Failed to suspend FPO');
    }
  };

  const getStatusBadge = (status: FPOStatus) => {
    const badges = {
      'ACTIVE': { bg: 'bg-green-100', text: 'text-green-800', label: 'Active' },
      'SETUP_FAILED': { bg: 'bg-red-100', text: 'text-red-800', label: 'Setup Failed' },
      'PENDING_SETUP': { bg: 'bg-yellow-100', text: 'text-yellow-800', label: 'Setup In Progress' },
      'SUSPENDED': { bg: 'bg-orange-100', text: 'text-orange-800', label: 'Suspended' },
    };

    const badge = badges[status] || { bg: 'bg-gray-100', text: 'text-gray-800', label: status };

    return (
      <span className={`px-2 py-1 rounded ${badge.bg} ${badge.text}`}>
        {badge.label}
      </span>
    );
  };

  if (loading) return <div>Loading FPO...</div>;
  if (error) return <div className="text-red-600">Error: {error}</div>;
  if (!fpo) return <div>FPO not found</div>;

  return (
    <div className="p-4 border rounded">
      <div className="flex justify-between items-start mb-4">
        <div>
          <h2 className="text-2xl font-bold">{fpo.name}</h2>
          <p className="text-gray-600">{fpo.registration_number}</p>
        </div>
        {getStatusBadge(fpo.status)}
      </div>

      <div className="mb-4">
        <p><strong>FPO ID:</strong> {fpo.fpo_id}</p>
        <p><strong>AAA Org ID:</strong> {fpo.aaa_org_id}</p>
        <p><strong>Created:</strong> {new Date(fpo.created_at).toLocaleString()}</p>
      </div>

      <div className="flex gap-2">
        {fpo.status === 'SETUP_FAILED' && (
          <button
            onClick={handleRetrySetup}
            disabled={syncing}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            {syncing ? 'Retrying...' : 'Retry Setup'}
          </button>
        )}

        {fpo.status === 'ACTIVE' && (
          <button
            onClick={handleSuspend}
            className="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600"
          >
            Suspend
          </button>
        )}
      </div>
    </div>
  );
};
```

### Angular Example

```typescript
// services/fpo.service.ts
import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { environment } from '../environments/environment';

export interface FPO {
  fpo_id: string;
  aaa_org_id: string;
  name: string;
  registration_number: string;
  status: FPOStatus;
  created_at: string;
  updated_at: string;
}

export type FPOStatus =
  | 'DRAFT'
  | 'PENDING_VERIFICATION'
  | 'VERIFIED'
  | 'REJECTED'
  | 'PENDING_SETUP'
  | 'SETUP_FAILED'
  | 'ACTIVE'
  | 'SUSPENDED'
  | 'INACTIVE'
  | 'ARCHIVED';

@Injectable({
  providedIn: 'root'
})
export class FPOService {
  private apiUrl = `${environment.apiUrl}/identity/fpo`;

  constructor(private http: HttpClient) {}

  private getHeaders(): HttpHeaders {
    const token = localStorage.getItem('auth_token');
    return new HttpHeaders({
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    });
  }

  getFPOByOrgId(aaaOrgId: string): Observable<FPO> {
    return this.http.get<any>(
      `${this.apiUrl}/by-org/${aaaOrgId}`,
      { headers: this.getHeaders() }
    ).pipe(
      map(response => response.data),
      catchError(error => {
        if (error.status === 404) {
          // Try sync as fallback
          return this.syncFPOFromAAA(aaaOrgId);
        }
        return throwError(() => error);
      })
    );
  }

  syncFPOFromAAA(aaaOrgId: string): Observable<FPO> {
    return this.http.post<any>(
      `${this.apiUrl}/sync/${aaaOrgId}`,
      {},
      { headers: this.getHeaders() }
    ).pipe(
      map(response => response.data)
    );
  }

  retrySetup(fpoId: string): Observable<void> {
    return this.http.post<any>(
      `${this.apiUrl}/${fpoId}/retry-setup`,
      {},
      { headers: this.getHeaders() }
    ).pipe(
      map(() => void 0)
    );
  }

  suspendFPO(fpoId: string, reason: string): Observable<void> {
    return this.http.put<any>(
      `${this.apiUrl}/${fpoId}/suspend`,
      { reason },
      { headers: this.getHeaders() }
    ).pipe(
      map(() => void 0)
    );
  }

  getHistory(fpoId: string): Observable<any[]> {
    return this.http.get<any>(
      `${this.apiUrl}/${fpoId}/history`,
      { headers: this.getHeaders() }
    ).pipe(
      map(response => response.data)
    );
  }
}
```

### Vue.js Example

```typescript
// composables/useFPO.ts
import { ref, Ref } from 'vue';
import axios from 'axios';

const API_BASE = import.meta.env.VITE_API_URL;

interface FPO {
  fpo_id: string;
  aaa_org_id: string;
  name: string;
  registration_number: string;
  status: FPOStatus;
  created_at: string;
  updated_at: string;
}

type FPOStatus =
  | 'DRAFT'
  | 'PENDING_VERIFICATION'
  | 'VERIFIED'
  | 'REJECTED'
  | 'PENDING_SETUP'
  | 'SETUP_FAILED'
  | 'ACTIVE'
  | 'SUSPENDED'
  | 'INACTIVE'
  | 'ARCHIVED';

export function useFPO() {
  const fpo: Ref<FPO | null> = ref(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const getHeaders = () => {
    const token = localStorage.getItem('auth_token');
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  };

  const getFPOByOrgId = async (aaaOrgId: string) => {
    try {
      loading.value = true;
      error.value = null;

      const response = await axios.get(
        `${API_BASE}/identity/fpo/by-org/${aaaOrgId}`,
        { headers: getHeaders() }
      );

      fpo.value = response.data.data;
    } catch (err: any) {
      if (err.response?.status === 404) {
        // Try sync
        await syncFPOFromAAA(aaaOrgId);
      } else {
        error.value = err.message || 'Failed to load FPO';
        throw err;
      }
    } finally {
      loading.value = false;
    }
  };

  const syncFPOFromAAA = async (aaaOrgId: string) => {
    try {
      loading.value = true;
      const response = await axios.post(
        `${API_BASE}/identity/fpo/sync/${aaaOrgId}`,
        {},
        { headers: getHeaders() }
      );

      fpo.value = response.data.data;
    } catch (err: any) {
      error.value = err.message || 'Failed to sync FPO';
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const retrySetup = async (fpoId: string) => {
    try {
      await axios.post(
        `${API_BASE}/identity/fpo/${fpoId}/retry-setup`,
        {},
        { headers: getHeaders() }
      );
      // Refresh FPO data after retry
      if (fpo.value) {
        await getFPOByOrgId(fpo.value.aaa_org_id);
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to retry setup';
      throw err;
    }
  };

  const suspendFPO = async (fpoId: string, reason: string) => {
    try {
      await axios.put(
        `${API_BASE}/identity/fpo/${fpoId}/suspend`,
        { reason },
        { headers: getHeaders() }
      );
      // Refresh FPO data
      if (fpo.value) {
        await getFPOByOrgId(fpo.value.aaa_org_id);
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to suspend FPO';
      throw err;
    }
  };

  return {
    fpo,
    loading,
    error,
    getFPOByOrgId,
    syncFPOFromAAA,
    retrySetup,
    suspendFPO
  };
}
```

---

## Best Practices

### 1. Always Use Auto-Sync Endpoint

**DO**: Use `/identity/fpo/by-org/:aaa_org_id` for getting FPO by org ID

```javascript
// Good
const fpo = await fpoService.getFPOByOrgId(orgId);
```

**DON'T**: Use legacy endpoint without fallback

```javascript
// Avoid - no auto-sync
const fpo = await fetch(`/identity/fpo/reference/${orgId}`);
```

### 2. Handle All Status States

```javascript
function canEditFPO(status: FPOStatus): boolean {
  return ['DRAFT', 'REJECTED'].includes(status);
}

function canOperateFPO(status: FPOStatus): boolean {
  return ['ACTIVE', 'SUSPENDED'].includes(status);
}

function needsAttention(status: FPOStatus): boolean {
  return ['SETUP_FAILED', 'REJECTED', 'SUSPENDED'].includes(status);
}
```

### 3. Implement Status Polling for Async Operations

```javascript
async function waitForSetupCompletion(fpoId: string, maxAttempts = 30) {
  for (let i = 0; i < maxAttempts; i++) {
    const fpo = await fpoService.getFPOById(fpoId);

    if (fpo.status === 'ACTIVE') {
      return fpo; // Success
    }

    if (fpo.status === 'SETUP_FAILED') {
      throw new Error('Setup failed');
    }

    // Wait 2 seconds before next poll
    await new Promise(resolve => setTimeout(resolve, 2000));
  }

  throw new Error('Setup timeout');
}
```

### 4. Show Helpful Error Messages

```javascript
function getFriendlyErrorMessage(error: any, aaaOrgId: string): string {
  if (error.response?.status === 404) {
    return `FPO not found. We're syncing the data from the server. Please try again in a moment.`;
  }

  if (error.response?.status === 400) {
    return error.response.data.message || 'Invalid request';
  }

  if (error.response?.status === 500) {
    return 'Server error. Please try again or contact support.';
  }

  return 'An unexpected error occurred';
}
```

### 5. Cache FPO Data Appropriately

```javascript
class FPOCache {
  private cache = new Map<string, { data: FPO; timestamp: number }>();
  private TTL = 5 * 60 * 1000; // 5 minutes

  get(aaaOrgId: string): FPO | null {
    const cached = this.cache.get(aaaOrgId);
    if (!cached) return null;

    if (Date.now() - cached.timestamp > this.TTL) {
      this.cache.delete(aaaOrgId);
      return null;
    }

    return cached.data;
  }

  set(aaaOrgId: string, data: FPO): void {
    this.cache.set(aaaOrgId, { data, timestamp: Date.now() });
  }

  invalidate(aaaOrgId: string): void {
    this.cache.delete(aaaOrgId);
  }
}
```

### 6. Monitor Setup Progress

```javascript
async function monitorFPOSetup(fpoId: string, onProgress: (status: string) => void) {
  const checkInterval = 3000; // 3 seconds
  const maxTime = 60000; // 1 minute
  const startTime = Date.now();

  while (Date.now() - startTime < maxTime) {
    const fpo = await fpoService.getFPOById(fpoId);
    onProgress(fpo.status);

    if (fpo.status === 'ACTIVE') {
      return { success: true, fpo };
    }

    if (fpo.status === 'SETUP_FAILED') {
      return { success: false, fpo };
    }

    await new Promise(resolve => setTimeout(resolve, checkInterval));
  }

  throw new Error('Setup monitoring timeout');
}
```

---

## Troubleshooting

### Problem: FPO Not Found After Creation

**Symptoms**:
- Created FPO in AAA service
- Client gets 404 when trying to fetch FPO

**Solution**:
1. Call sync endpoint: `POST /identity/fpo/sync/:aaa_org_id`
2. Or use auto-sync endpoint: `GET /identity/fpo/by-org/:aaa_org_id`

**Example**:
```javascript
// After creating FPO in AAA
const aaaOrgId = createdOrg.org_id;

// Give backend a moment to process
await new Promise(resolve => setTimeout(resolve, 1000));

// Sync to local DB
await fpoService.syncFPOFromAAA(aaaOrgId);

// Now fetch normally
const fpo = await fpoService.getFPOByOrgId(aaaOrgId);
```

---

### Problem: Setup Failed Status

**Symptoms**:
- FPO stuck in `SETUP_FAILED` status
- User groups or permissions not created

**Solution**:
1. Check setup errors in FPO details
2. Verify AAA service is accessible
3. Call retry endpoint: `POST /identity/fpo/:id/retry-setup`
4. Maximum 3 retry attempts allowed

**Example**:
```javascript
const fpo = await fpoService.getFPOById(fpoId);

if (fpo.status === 'SETUP_FAILED') {
  console.log('Setup errors:', fpo.setup_errors);

  // Check retry attempts
  if (fpo.setup_attempts < 3) {
    await fpoService.retrySetup(fpoId);

    // Poll for completion
    await waitForSetupCompletion(fpoId);
  } else {
    // Max retries exceeded - contact support
    notifySupport(fpoId, fpo.setup_errors);
  }
}
```

---

### Problem: Invalid State Transition

**Symptoms**:
- API returns 400 error
- Message: "cannot transition from X to Y"

**Solution**:
1. Check current FPO status
2. Verify transition is valid (see State Transition Diagram)
3. Update UI to only show valid actions

**Example**:
```javascript
function getAvailableActions(status: FPOStatus): string[] {
  const actions: Record<FPOStatus, string[]> = {
    'ACTIVE': ['suspend', 'deactivate'],
    'SUSPENDED': ['reactivate', 'deactivate'],
    'SETUP_FAILED': ['retry_setup'],
    'INACTIVE': ['archive'],
    // ... other states
  };

  return actions[status] || [];
}
```

---

### Problem: Audit History Not Available

**Symptoms**:
- History endpoint returns empty array
- Recent status changes not showing

**Solution**:
1. Verify FPO ID is correct
2. Check if FPO was created before lifecycle system deployment
3. Only changes after migration have audit logs

**Example**:
```javascript
const history = await fpoService.getHistory(fpoId);

if (history.length === 0) {
  console.log('No audit history available (FPO created before lifecycle system)');
  // Show creation date instead
  console.log('FPO created:', fpo.created_at);
}
```

---

## Summary Checklist

### For Frontend Developers

- [ ] Update API service to use new endpoints
- [ ] Add handling for all 10 FPO statuses
- [ ] Implement auto-sync fallback for FPO not found
- [ ] Add retry setup button for `SETUP_FAILED` status
- [ ] Add suspend/reactivate actions for admins
- [ ] Display status badges with appropriate colors
- [ ] Show audit history in FPO details
- [ ] Update error messages with helpful guidance
- [ ] Implement status polling for async operations
- [ ] Cache FPO data with appropriate TTL
- [ ] Test all state transitions
- [ ] Update TypeScript types/interfaces

### For Mobile Developers

- [ ] Update API calls to new endpoints
- [ ] Add status enum with all 10 values
- [ ] Implement sync retry logic
- [ ] Add UI for setup retry
- [ ] Add admin controls (suspend/reactivate)
- [ ] Show status indicators
- [ ] Display audit trail
- [ ] Handle all error scenarios
- [ ] Test offline sync behavior
- [ ] Update data models

### For Backend Teams Integrating

- [ ] Review API endpoint documentation
- [ ] Test sync endpoint with various org IDs
- [ ] Verify state transition validations
- [ ] Check audit log generation
- [ ] Test retry logic with failed setups
- [ ] Validate error responses
- [ ] Performance test with high volume
- [ ] Monitor setup success rates

---

## Support

For questions or issues:

1. **Documentation**: Review this guide and API specifications
2. **Logs**: Check FPO audit history: `GET /identity/fpo/:id/history`
3. **Status**: Verify FPO status and setup errors
4. **Sync**: Try manual sync: `POST /identity/fpo/sync/:aaa_org_id`
5. **Support**: Contact backend team with FPO ID and error details

---

## Appendix

### Complete TypeScript Type Definitions

```typescript
// types/fpo.ts

export type FPOStatus =
  | 'DRAFT'
  | 'PENDING_VERIFICATION'
  | 'VERIFIED'
  | 'REJECTED'
  | 'PENDING_SETUP'
  | 'SETUP_FAILED'
  | 'ACTIVE'
  | 'SUSPENDED'
  | 'INACTIVE'
  | 'ARCHIVED';

export interface FPO {
  fpo_id: string;
  aaa_org_id: string;
  name: string;
  registration_number: string;
  status: FPOStatus;
  status_reason?: string;
  status_changed_at?: string;
  status_changed_by?: string;
  previous_status?: FPOStatus;
  business_config: Record<string, any>;
  metadata?: Record<string, any>;
  setup_attempts?: number;
  setup_errors?: Record<string, string>;
  ceo_user_id?: string;
  created_at: string;
  updated_at: string;
}

export interface AuditEntry {
  action: string;
  previous_state?: FPOStatus;
  new_state?: FPOStatus;
  reason: string;
  performed_by: string;
  performed_at: string;
  details?: Record<string, any>;
}

export interface SyncFPORequest {
  aaa_org_id: string;
}

export interface SuspendFPORequest {
  reason: string;
}

export interface DeactivateFPORequest {
  reason: string;
}

export interface APIResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}
```

### State Transition Validation Matrix

```typescript
const STATE_TRANSITIONS: Record<FPOStatus, FPOStatus[]> = {
  'DRAFT': ['PENDING_VERIFICATION'],
  'PENDING_VERIFICATION': ['VERIFIED', 'REJECTED'],
  'VERIFIED': ['PENDING_SETUP'],
  'REJECTED': ['DRAFT', 'ARCHIVED'],
  'PENDING_SETUP': ['ACTIVE', 'SETUP_FAILED'],
  'SETUP_FAILED': ['PENDING_SETUP', 'INACTIVE'],
  'ACTIVE': ['SUSPENDED', 'INACTIVE'],
  'SUSPENDED': ['ACTIVE', 'INACTIVE'],
  'INACTIVE': ['ARCHIVED'],
  'ARCHIVED': []
};

export function canTransition(from: FPOStatus, to: FPOStatus): boolean {
  return STATE_TRANSITIONS[from]?.includes(to) ?? false;
}
```

---

**End of Client Integration Guide**

**Version**: 1.0
**Last Updated**: 2025-11-18
**Maintained By**: Farmers Module Team
