# Business Rules and Logic Decisions

## Overview

This document captures critical business logic decisions for the farmers-module service. These rules define system behavior for edge cases, failure scenarios, and data consistency requirements.

**Last Updated**: 2025-10-06
**Decision Authority**: Product Owner / System Architect

---

## 1. FPO Creation and Organization Management

### 1.1 Partial Failure Handling

**Rule**: When FPO creation experiences partial failure (e.g., organization created but user groups fail), the system SHALL mark the FPO with status `PENDING_SETUP` rather than rolling back the entire operation.

**Rationale**:
- Distributed transactions across AAA service are complex and error-prone
- Maintaining partial state allows for recovery operations
- Rollback would require coordinating state across multiple services

**Implementation**:
- FPO status enum includes: `ACTIVE`, `PENDING_SETUP`, `INACTIVE`, `SUSPENDED`
- CompleteFPOSetup endpoint allows retry of failed user group creation
- Audit logs track which groups succeeded/failed for troubleshooting

### 1.2 CEO User Management

**Rule**: A user CANNOT be CEO of multiple FPOs simultaneously. However, if a user already exists (as farmer or other role), the system SHALL add the CEO role to the existing user rather than creating a new account.

**Rationale**:
- Phone numbers are globally unique identifiers
- Single user identity across organizations reduces confusion
- Role accumulation allows users to operate in multiple contexts

**Validation**:
- Before assigning CEO role, query AAA to check if user already has CEO role for another organization
- Return 409 Conflict if user is already a CEO elsewhere
- Add CEO role to existing user if they only have farmer/other roles

**Exception Handling**:
- If CEO assignment fails, mark FPO as `PENDING_SETUP`
- Allow manual intervention to assign alternative CEO

---

## 2. Farmer Registration and Identity Management

### 2.1 Phone Number Uniqueness

**Rule**: Phone numbers are globally unique across the entire system and permanently bound to a single user identity. Multiple roles can be assigned to the same user across different organizations.

**Implications**:
- RegisterFarmer with existing phone number SHALL return existing farmer profile (idempotent)
- Same person can be farmer in FPO-A and CEO in FPO-B (different roles)
- Phone verification prevents duplicate accounts

**Implementation**:
- AAA service enforces phone uniqueness at user creation
- Local farmer registration checks AAA first before creating users
- Eventual consistency acceptable due to phone verification workflow

### 2.2 KisanSathi Assignment Validation

**Rule**: If KisanSathi validation fails AFTER farmer creation in AAA, the system SHALL set `kisan_sathi_user_id` to NULL and return a warning message. The AAA user SHALL NOT be deleted.

**Rationale**:
- Farmer's AAA identity is independent of KisanSathi assignment
- Deleting AAA user creates orphaned authentication state
- KisanSathi can be assigned later through separate workflow

**Implementation**:
```go
// Validate KisanSathi if provided
if req.KisanSathiUserID != nil {
    if err := ValidateKisanSathi(ctx, *req.KisanSathiUserID); err != nil {
        // Don't fail registration, just warn
        log.Warn("KisanSathi validation failed, proceeding without assignment",
                 zap.String("farmer_id", farmerID),
                 zap.Error(err))
        req.KisanSathiUserID = nil
        response.Warnings = append(response.Warnings,
            "KisanSathi validation failed: " + err.Error())
    }
}
```

**Recovery**:
- KisanSathi can be assigned later via AssignKisanSathi endpoint
- Farmers without KisanSathi can still perform basic operations

---

## 3. Farmer-FPO Relationship Management

### 3.1 Relationship Cardinality

**Rule**: A farmer can be ACTIVELY linked to only ONE FPO at a time. However, the same user can participate in multiple organizations with different roles (e.g., farmer in FPO-A, store_staff in FPO-B).

**Database Constraint**:
```sql
-- Unique constraint ensures 1:1 active relationship
CREATE UNIQUE INDEX idx_farmer_active_fpo ON farmer_links
    (aaa_user_id) WHERE status = 'ACTIVE';
```

**Behavioral Rules**:
- LinkFarmerToFPO on already-linked farmer SHALL return 409 Conflict
- Client must explicitly UnlinkFarmerFromFPO before relinking to different FPO
- Soft-deleted links (status='INACTIVE') don't prevent new linkages

**Multi-Organization Participation**:
- User ID `user123` can have:
  - farmer_link to FPO-A (status='ACTIVE')
  - AAA role 'store_staff' in FPO-B
  - AAA role 'shareholder' in FPO-C
- Role-based access control in AAA determines operation permissions

### 3.2 KisanSathi Reference Integrity

**Rule**: When a farmer link is unlinked (status set to INACTIVE), the KisanSathi assignment SHALL remain in the record for audit purposes but SHALL NOT affect active KisanSathi operations.

**Implementation**:
- UnlinkFarmerFromFPO sets `status='INACTIVE'` and preserves `kisan_sathi_user_id`
- Active KisanSathi queries filter by `status='ACTIVE'`
- Historical reporting includes inactive links with KisanSathi assignment

**Orphan Handling**:
- If KisanSathi user is deleted in AAA, reconciliation job marks affected farmer_links
- Farmers can continue operations; KisanSathi field becomes advisory
- Alert generated for FPO admin to reassign KisanSathi

---

## 4. Token Validation and Authentication

### 4.1 Single Source of Truth

**Rule**: ALL token validation SHALL be performed through AAA service's `ValidateToken` gRPC method. Local token validation logic SHALL NOT be used.

**Deprecation**:
- `internal/auth/token_validator.go` is deprecated and should be removed
- Middleware SHALL call `AAAClient.ValidateToken()` for all auth checks
- No local JWT parsing or validation allowed

**Performance Considerations**:
- Consider caching ValidateToken responses (TTL-based)
- Implement circuit breaker for AAA service calls
- If AAA unavailable, return 503 Service Unavailable (no fallback authentication)

**Token Refresh**:
- Client-side responsibility to refresh tokens before expiration
- Farmers-module does NOT implement token refresh endpoint
- AAA service handles all token lifecycle operations

### 4.2 Fallback Strategy

**Rule**: If AAA service is unavailable, authentication SHALL fail with 503 Service Unavailable. NO local fallback authentication is permitted.

**Rationale**:
- Security requires consistent authentication source
- Stale tokens could grant unauthorized access
- Better to fail closed than risk security breach

**Exception**: Health check endpoint (/health) does NOT require authentication

---

## 5. Idempotency and Eventual Consistency

### 5.1 RegisterFarmer Idempotency

**Rule**: Calling RegisterFarmer with an existing phone number SHALL return the existing farmer profile with status 200 OK rather than creating a duplicate or returning an error.

**Implementation**:
```go
// Check if user exists by phone
existingUser, err := aaaClient.GetUserByPhone(ctx, req.PhoneNumber)
if err == nil && existingUser != nil {
    // User exists - retrieve local farmer profile
    farmer, err := farmerRepo.GetByAAAUserID(ctx, existingUser.ID)
    if err == nil {
        return &RegisterFarmerResponse{
            FarmerID: farmer.ID,
            AAAUserID: farmer.AAAUserID,
            Status: "existing",
            Message: "Farmer already registered",
        }, nil
    }
}
// Otherwise, proceed with registration
```

**Duplicate Detection**:
- AAA service enforces phone uniqueness
- Concurrent registrations may create race condition
- Eventual consistency: phone verification prevents activation of duplicates

### 5.2 Concurrent Operation Handling

**Rule**: The system SHALL rely on eventual consistency for concurrent operations. Distributed locks are NOT required for farmer registration.

**Rationale**:
- Phone verification workflow prevents duplicate activations
- Complexity of distributed locks outweighs benefits
- AAA service provides ultimate uniqueness guarantee

**Race Condition Example**:
1. Request A and B register same phone simultaneously
2. Both check AAA (no existing user found)
3. Both create AAA user (one succeeds, one fails with conflict)
4. Winner creates local farmer profile
5. Loser returns existing farmer profile

---

## 6. Data Consistency and Reconciliation

### 6.1 AAA-Local State Invariants

**Invariants** (must ALWAYS be true):
1. Every `farmer_profiles.aaa_user_id` MUST reference a valid user in AAA
2. Every `fpo_refs.aaa_org_id` MUST reference a valid organization in AAA
3. Every `farmer_links` with status='ACTIVE' MUST reference valid AAA user and org
4. A farmer can have only ONE active farmer_link at a time

**Reconciliation Strategy**:
- Scheduled job (daily): Query AAA for all referenced IDs, mark orphans
- On-demand: ReconcileAAALinks endpoint for manual trigger
- Healing: Attempt to restore valid references or mark as ORPHANED status

### 6.2 Drift Detection and Recovery

**Rule**: If AAA data is modified externally (e.g., user deleted in AAA admin panel), the system SHALL detect drift during reconciliation and mark affected records as ORPHANED.

**Detection Methods**:
- Reconciliation job queries AAA for existence checks
- Real-time: API calls that fail with "user not found" trigger drift alerts
- Audit logs track drift detection events

**Recovery Actions**:
- Mark farmer_links as status='ORPHANED'
- Notify FPO admin of data inconsistency
- Provide manual resolution workflow (delete local record or restore in AAA)

---

## 7. Farm Management and Geospatial Data

### 7.1 Geometry Validation Business Rules

**Technical Validation** (via PostGIS):
- SRID must be 4326 (WGS84)
- Geometry must be valid (ST_IsValid)
- No self-intersections allowed

**Business Validation**:
- Maximum farm size: 100 hectares per farm (configurable)
- Minimum farm size: 0.01 hectares (100 sq meters)
- Farm boundaries should not overlap within same organization (warning, not error)

**Implementation**:
```go
const (
    MaxFarmSizeHa = 100.0
    MinFarmSizeHa = 0.01
)

func ValidateFarmGeometry(geometry string) error {
    // PostGIS validation first
    if err := postgis.Validate(geometry); err != nil {
        return err
    }

    // Business rule validation
    area := postgis.CalculateArea(geometry)
    if area > MaxFarmSizeHa {
        return fmt.Errorf("farm size %.2f ha exceeds maximum %d ha",
                          area, MaxFarmSizeHa)
    }
    if area < MinFarmSizeHa {
        return fmt.Errorf("farm size %.2f ha below minimum %.2f ha",
                          area, MinFarmSizeHa)
    }

    return nil
}
```

---

## 8. Error Handling and Audit Requirements

### 8.1 Soft Delete vs Archive

**Rule**: All entity deletions SHALL use soft delete (setting `deleted_at` timestamp). Hard deletes are NEVER performed except for GDPR compliance requests.

**Soft Delete Behavior**:
- Queries exclude soft-deleted records by default (GORM handles via WHERE deleted_at IS NULL)
- Restoration possible within 90-day retention window
- After 90 days, archive job moves to cold storage

**GDPR "Right to be Forgotten"**:
- Explicit deletion request triggers hard delete workflow
- Requires manual approval from compliance officer
- Cascades to all related data (farms, cycles, activities)
- Audit log retains deletion event but not personal data

### 8.2 Audit Log Requirements

**What to Log**:
- All state-changing operations (CREATE, UPDATE, DELETE)
- Authorization decisions (ALLOWED, DENIED)
- Authentication events (LOGIN, TOKEN_VALIDATION_FAILED)
- Reconciliation events (DRIFT_DETECTED, ORPHAN_MARKED)

**Audit Log Format**:
```json
{
  "timestamp": "2025-10-06T10:30:00Z",
  "correlation_id": "req-abc-123",
  "event_type": "FARMER_REGISTERED",
  "actor": {
    "user_id": "admin-123",
    "org_id": "fpo-456",
    "role": "system_admin"
  },
  "resource": {
    "type": "farmer",
    "id": "farmer-789",
    "aaa_user_id": "user-xyz"
  },
  "action": "create",
  "status": "success",
  "metadata": {
    "kisan_sathi_assigned": "false",
    "warnings": ["KisanSathi validation failed"]
  }
}
```

**Retention**:
- Hot storage: 30 days (queryable via API)
- Warm storage: 1 year (archived, queryable via admin interface)
- Cold storage: 7 years (compliance requirement)

---

## 9. Security and Access Control

### 9.1 Cross-Organization Operations

**Rule**: System administrators (role='system_admin') MAY perform cross-organization queries for analytics purposes. All other users are restricted to their organization scope.

**Permission Checks**:
```go
// Example: List all farmers across organizations
if !hasRole(ctx, "system_admin") && req.CrossOrg {
    return nil, errors.New("cross-org queries require system_admin role")
}

// Example: FPO admin viewing farmers in their org
if !hasPermission(ctx, "farmer.list", orgID) {
    return nil, errors.New("insufficient permissions")
}
```

**Data Masking**:
- Cross-org analytics return aggregated data only (counts, averages)
- PII (phone numbers, addresses) masked unless viewing own organization

### 9.2 KisanSathi Permission Scope

**Rule**: KisanSathi users can create/update farm activities ONLY for farmers they are assigned to.

**Implementation**:
```go
func (s *FarmActivityService) CreateActivity(ctx context.Context, req *CreateActivityRequest) error {
    // Get user context from token
    userCtx := auth.GetUserContext(ctx)

    // If user is KisanSathi, verify they're assigned to this farmer
    if userCtx.Role == "kisansathi" {
        link, err := s.farmerLinkRepo.GetByUserID(ctx, req.FarmerID)
        if err != nil || link.KisanSathiUserID == nil ||
           *link.KisanSathiUserID != userCtx.UserID {
            return errors.New("KisanSathi can only create activities for assigned farmers")
        }
    }

    // Proceed with activity creation
}
```

---

## 10. Future Considerations

### 10.1 Multi-FPO Farmer Membership

**Current Limitation**: Farmer can be actively linked to only ONE FPO.

**Future Enhancement**: If business requirements change to allow multi-FPO membership:
1. Remove unique constraint on `idx_farmer_active_fpo`
2. Update LinkFarmerToFPO to allow multiple active links
3. Add `primary_fpo_id` field to indicate main FPO affiliation
4. Update reporting to aggregate across multiple FPO memberships

**Migration Path**:
- Phased rollout with feature flag
- Pilot with specific FPOs before general release
- Update AAA permission model for multi-org farmer operations

### 10.2 Blockchain Integration for Land Records

**Consideration**: Future integration with government land record blockchain

**Requirements**:
- Farm geometry must reference official survey numbers
- Immutable audit trail for land ownership changes
- Smart contract for farm transfer operations

**Preparation**:
- Add `survey_number` field to Farm model
- Implement geometry hash for tamper detection
- Design event sourcing architecture for land records

---

## Revision History

| Date       | Version | Author | Changes                                      |
|------------|---------|--------|----------------------------------------------|
| 2025-10-06 | 1.0     | System | Initial business rules documentation         |
| TBD        | 1.1     | TBD    | Add rules for bulk farmer import workflow    |
| TBD        | 1.2     | TBD    | Add rules for FPO federation/merger workflow |

---

## Approval

**Approved By**: [Product Owner Name]
**Date**: [Approval Date]
**Status**: DRAFT (Pending formal approval)
