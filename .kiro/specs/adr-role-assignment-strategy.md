# ADR-001: Role Assignment Strategy for Entity Creation Workflows

**Status:** Accepted
**Date:** 2025-10-16
**Deciders:** Backend Architecture Team, Security Team
**Context Owner:** @agent-sde3-backend-architect

---

## Context and Problem Statement

The farmers-module creates entities (farmers, FPOs, KisanSathis) that map to users and organizations in the AAA (Authentication, Authorization, Auditing) service. These users require specific roles to perform authorized operations:

- `farmer` role for all farmer users
- `CEO` / `fpo_ceo` role for FPO chief executive officers
- `kisansathi` role for field agents (KisanSathis)

**Core Question:** Should role assignment be part of the critical path (fatal error if assignment fails) or best-effort with eventual consistency (non-fatal, retry later)?

**Trade-offs:**
- **Immediate Consistency** (fatal): Ensures data integrity but reduces availability
- **Eventual Consistency** (non-fatal): Maintains availability but introduces temporary inconsistency

**Current State (Broken):**
- Farmer creation does NOT assign `farmer` role at all (security gap)
- FPO creation assigns `CEO` role but treats failure as non-fatal warning
- KisanSathi assignment correctly ensures role with verification

---

## Decision Drivers

1. **System Availability**: Farmers-module must remain operational even if AAA service experiences transient failures
2. **Data Integrity**: Users must have appropriate roles to prevent authorization failures
3. **Security**: Missing roles could allow privilege escalation or unauthorized access
4. **Operational Complexity**: Recovery mechanisms should be straightforward
5. **Developer Experience**: Error handling should be clear and consistent
6. **Audit Requirements**: All role assignments must be traceable

---

## Considered Options

### Option 1: Fatal Role Assignment (Immediate Consistency)

**Pattern:**
```go
func CreateFarmer(ctx context.Context, req) (*Response, error) {
    user, err := createAAAUser(ctx, req)
    if err != nil {
        return nil, err // Fatal: Cannot create farmer without AAA user
    }

    err = assignRole(ctx, user.ID, "farmer")
    if err != nil {
        return nil, err // Fatal: Cannot create farmer without role
    }

    farmer, err := createFarmerEntity(ctx, user.ID, req)
    if err != nil {
        return nil, err // Fatal: Database write failed
    }

    return farmer, nil
}
```

**Pros:**
- Strong consistency guarantee (user always has correct role)
- Simpler to reason about (no eventual consistency complexity)
- No retry or reconciliation logic needed
- Clear failure semantics (operation succeeded or failed completely)

**Cons:**
- Single point of failure (AAA role service outage blocks all farmer creation)
- Higher operational burden (requires AAA service to be highly available)
- Partial failures are harder to recover (user created in AAA but farmer creation failed)
- Cannot create entities during AAA maintenance windows

**Risk Assessment:**
- **Availability Risk**: HIGH - AAA outage completely blocks farmer operations
- **Complexity Risk**: LOW - Straightforward error handling
- **Recovery Risk**: MEDIUM - Manual intervention required for partial failures

---

### Option 2: Best-Effort Role Assignment (Eventual Consistency)

**Pattern:**
```go
func CreateFarmer(ctx context.Context, req) (*Response, error) {
    user, err := createAAAUser(ctx, req)
    if err != nil {
        return nil, err // Fatal: Cannot proceed without AAA user
    }

    farmer, err := createFarmerEntity(ctx, user.ID, req)
    if err != nil {
        return nil, err // Fatal: Primary operation failed
    }

    // Best-effort role assignment (non-fatal)
    roleErr := assignRoleWithRetry(ctx, user.ID, "farmer")
    if roleErr != nil {
        log.Warnf("Role assignment failed for farmer %s: %v", farmer.ID, roleErr)
        farmer.Metadata["role_assignment_pending"] = "true"
        farmer.Metadata["role_assignment_error"] = roleErr.Error()
        updateFarmerMetadata(ctx, farmer) // Store failure for later retry
    }

    return farmer, nil // Success even if role assignment failed
}
```

**Pros:**
- High availability (AAA role service outage doesn't block operations)
- Resilient to transient failures (automatic retry possible)
- Clear ownership (farmer entity is source of truth for pending assignments)
- Graceful degradation (core operations continue, authorization may temporarily fail)

**Cons:**
- Eventual consistency window (brief period where user lacks proper role)
- Requires additional infrastructure (retry queue, reconciliation jobs)
- More complex error handling and monitoring
- Potential authorization failures during inconsistency window

**Risk Assessment:**
- **Availability Risk**: LOW - System remains operational during AAA issues
- **Complexity Risk**: HIGH - Requires retry/reconciliation mechanisms
- **Recovery Risk**: LOW - Automatic retry handles most failures

---

### Option 3: Hybrid Approach (Context-Dependent)

**Pattern:**
```go
func CreateFarmer(ctx context.Context, req) (*Response, error) {
    // Critical roles: Fatal failure
    if req.IsCriticalRole() {
        return createWithFatalRoleAssignment(ctx, req)
    }

    // Standard roles: Eventual consistency
    return createWithBestEffortRoleAssignment(ctx, req)
}
```

**Example Distinctions:**
- **Fatal**: CEO role assignment for FPO (critical for organization leadership)
- **Best-Effort**: Farmer role assignment (can be reconciled later)

**Pros:**
- Flexible approach based on role criticality
- Balances consistency and availability needs
- Allows per-role configuration

**Cons:**
- Inconsistent patterns across codebase (harder to maintain)
- Requires clear criteria for "critical" vs "non-critical" roles
- More complex decision-making during design

**Risk Assessment:**
- **Availability Risk**: MEDIUM - Some operations blocked, others continue
- **Complexity Risk**: HIGH - Multiple patterns to maintain
- **Recovery Risk**: MEDIUM - Mixed recovery strategies

---

### Option 4: Synchronous Retry with Timeout

**Pattern:**
```go
func CreateFarmer(ctx context.Context, req) (*Response, error) {
    user, err := createAAAUser(ctx, req)
    if err != nil {
        return nil, err
    }

    farmer, err := createFarmerEntity(ctx, user.ID, req)
    if err != nil {
        return nil, err
    }

    // Retry role assignment with exponential backoff
    roleErr := retryAssignRole(ctx, user.ID, "farmer", maxRetries=3, timeout=10s)
    if roleErr != nil {
        // After retries exhausted, continue with eventual consistency
        log.Warnf("Role assignment failed after retries: %v", roleErr)
        farmer.Metadata["role_assignment_pending"] = "true"
    }

    return farmer, nil
}
```

**Pros:**
- Handles transient failures automatically
- Higher success rate than single-attempt
- Bounded latency (timeout prevents long delays)

**Cons:**
- Adds latency to farmer creation (retry delays)
- Timeout tuning is difficult (too short = unnecessary failures, too long = poor UX)
- Still requires eventual consistency fallback for extended outages

**Risk Assessment:**
- **Availability Risk**: MEDIUM - Short outages handled, long outages still problematic
- **Complexity Risk**: MEDIUM - Retry logic adds some complexity
- **Recovery Risk**: LOW - Most transient failures resolved automatically

---

## Decision Outcome

**Chosen Option:** **Option 2 (Best-Effort Role Assignment with Eventual Consistency)** with elements from Option 4 (synchronous retry).

### Rationale

1. **Availability is Critical**: Farmer onboarding must not be blocked by AAA service issues
2. **Security is Acceptable**: Brief inconsistency window is tolerable if:
   - Role assignment is retried automatically
   - Authorization checks fail-closed (deny by default)
   - Inconsistent state is observable and alertable
3. **Operational Reality**: AAA service is a shared dependency that may experience issues independent of farmers-module
4. **User Experience**: Farmers should be able to complete registration even if authorization setup is delayed
5. **Precedent**: KisanSathi assignment already uses a similar pattern successfully

### Implementation Strategy

```go
func CreateFarmer(ctx context.Context, req) (*Response, error) {
    // Step 1: Create or retrieve AAA user (FATAL)
    aaaUserID, err := s.getOrCreateAAAUser(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create AAA user: %w", err)
    }

    // Step 2: Create local farmer entity (FATAL - primary operation)
    farmer, err := s.createFarmerEntity(ctx, aaaUserID, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create farmer entity: %w", err)
    }

    // Step 3: Assign farmer role (BEST-EFFORT with single retry)
    roleErr := s.ensureFarmerRoleWithRetry(ctx, aaaUserID, req.AAAOrgID)
    if roleErr != nil {
        log.Warnf("Role assignment failed for farmer %s (user %s): %v",
            farmer.ID, aaaUserID, roleErr)

        // Store failure metadata for reconciliation
        if farmer.Metadata == nil {
            farmer.Metadata = make(entities.JSONB)
        }
        farmer.Metadata["role_assignment_error"] = roleErr.Error()
        farmer.Metadata["role_assignment_pending"] = "true"
        farmer.Metadata["role_assignment_attempted_at"] = time.Now().Format(time.RFC3339)

        // Best-effort metadata update (non-fatal if fails)
        if updateErr := s.repository.Update(ctx, farmer); updateErr != nil {
            log.Errorf("Failed to store role assignment failure metadata: %v", updateErr)
        }

        // Trigger async retry (future enhancement)
        // s.roleAssignmentQueue.Enqueue(farmer.ID, aaaUserID, "farmer")
    }

    return s.buildResponse(farmer, roleErr), nil
}

func (s *FarmerServiceImpl) ensureFarmerRoleWithRetry(ctx, userID, orgID string) error {
    // Implement "Check-Assign-Verify" pattern with one retry
    for attempt := 1; attempt <= 2; attempt++ {
        err := s.ensureFarmerRole(ctx, userID, orgID)
        if err == nil {
            return nil // Success
        }

        if attempt == 1 {
            log.Warnf("Role assignment attempt %d failed, retrying: %v", attempt, err)
            time.Sleep(500 * time.Millisecond) // Brief delay before retry
            continue
        }

        return fmt.Errorf("role assignment failed after %d attempts: %w", attempt, err)
    }
    return nil
}

func (s *FarmerServiceImpl) ensureFarmerRole(ctx, userID, orgID string) error {
    // 1. Check if role already exists (idempotency)
    hasRole, err := s.aaaService.CheckUserRole(ctx, userID, constants.RoleFarmer)
    if err != nil {
        // If check fails, try assignment anyway (AAA may be degraded but functional)
        log.Warnf("Failed to check farmer role for user %s: %v", userID, err)
    } else if hasRole {
        log.Debugf("User %s already has farmer role, skipping assignment", userID)
        return nil
    }

    // 2. Assign role
    err = s.aaaService.AssignRole(ctx, userID, orgID, constants.RoleFarmer)
    if err != nil {
        return fmt.Errorf("failed to assign farmer role: %w", err)
    }

    // 3. Verify assignment succeeded
    hasRole, err = s.aaaService.CheckUserRole(ctx, userID, constants.RoleFarmer)
    if err != nil {
        return fmt.Errorf("failed to verify farmer role assignment: %w", err)
    }
    if !hasRole {
        return fmt.Errorf("farmer role assignment verification failed (role not present)")
    }

    log.Infof("Successfully assigned and verified farmer role for user %s", userID)
    return nil
}
```

---

## Consequences

### Positive Consequences

1. **High Availability**: Farmer creation remains operational during AAA service issues
2. **Clear Failure Tracking**: Metadata field explicitly marks pending role assignments
3. **Automatic Recovery**: Single retry handles most transient failures
4. **Observable**: Can query entities with `role_assignment_pending = true` for monitoring
5. **Graceful Degradation**: Core entity creation succeeds even if authorization setup delayed
6. **Consistent Pattern**: Same approach can be used for FPO CEO and KisanSathi roles

### Negative Consequences

1. **Eventual Consistency Window**: Brief period (seconds to minutes) where user lacks role
2. **Authorization Failures**: Operations may be denied until role assignment completes
3. **Monitoring Overhead**: Requires alerts for high pending assignment rates
4. **Reconciliation Required**: Need background job to retry failed assignments
5. **Complex Error States**: Must handle entities in "partial" state
6. **Metadata Pollution**: Failure information stored in entity metadata (alternative: separate table)

### Mitigation Strategies

1. **Fail-Closed Authorization**: Ensure permission checks default to deny (AAA service must enforce)
2. **Proactive Monitoring**: Alert when pending assignments exceed threshold
3. **Reconciliation Job**: Implement daily job to retry pending assignments
4. **User Communication**: Show "Setup in progress" message if role assignment pending
5. **Admin Override**: Provide manual role assignment endpoint for urgent cases

---

## Compliance and Security Considerations

### OWASP ASVS Compliance

| Requirement | Status | Notes |
|-------------|--------|-------|
| V4.1.1 - Enforce access control on trusted service layer | ✅ COMPLIANT | AAA service enforces authorization |
| V4.1.3 - Principle of least privilege | ⚠️ PARTIAL | Brief window where user may lack role |
| V4.1.5 - Access controls fail securely | ✅ COMPLIANT | Permission checks default to deny |

### Security Threat Analysis

**Threat:** User exploits eventual consistency window to bypass authorization

**Likelihood:** LOW
- Window is brief (< 1 minute with retry)
- User must know role is pending
- Permission checks still enforced by AAA

**Impact:** MEDIUM
- User could attempt operations without proper role
- Operations should be denied by AAA (fail-closed)

**Mitigation:**
- Ensure AAA service defaults to deny when role is missing
- Monitor for permission-denied errors during onboarding
- Implement rate limiting on sensitive operations

**Threat:** Pending role assignments accumulate, indicating systemic AAA issue

**Likelihood:** MEDIUM
- AAA service may experience extended outage
- Network issues between services

**Impact:** HIGH
- Large number of users lack proper roles
- Degraded user experience (operations denied)

**Mitigation:**
- Alert when pending assignments > 100 entities
- Implement circuit breaker to stop entity creation during extended AAA outage
- Provide status page showing system health

---

## Implementation Checklist

### Phase 1: Core Implementation (Week 1)
- [ ] Create role constants file (`internal/constants/roles.go`)
- [ ] Implement `ensureFarmerRole` method in `farmer_service.go`
- [ ] Add role assignment call to `CreateFarmer` workflow
- [ ] Update FPO `CreateFPO` to use same pattern for CEO role
- [ ] Add unit tests for `ensureFarmerRole`
- [ ] Add integration test for role assignment failure

### Phase 2: Monitoring and Observability (Week 2)
- [ ] Add Prometheus metrics for role assignments
- [ ] Create Grafana dashboard for pending assignments
- [ ] Configure PagerDuty alerts for high pending rate
- [ ] Add structured logging for role assignment lifecycle

### Phase 3: Reconciliation and Recovery (Week 3)
- [ ] Implement background reconciliation job
- [ ] Add admin endpoint to manually trigger role assignment
- [ ] Create runbook for role assignment failures
- [ ] Test reconciliation with simulated AAA outage

### Phase 4: Data Migration (Week 4)
- [ ] Create migration script for existing farmers
- [ ] Run dry-run in staging environment
- [ ] Execute production migration (off-peak hours)
- [ ] Validate all entities have correct roles

---

## Links and References

- [OWASP ASVS v4.0 - Access Control Verification](https://owasp.org/www-project-application-security-verification-standard/)
- [Eventual Consistency Patterns - Martin Kleppmann](https://martin.kleppmann.com/2015/05/11/please-stop-calling-databases-cp-or-ap.html)
- [AAA Service Role Assignment Review](./.kiro/specs/aaa-role-assignment-review.md)
- [Farmers Module Architecture Review](./.kiro/specs/architectural-review-report.md)
- [KisanSathi Role Assignment Implementation](../internal/services/farmer_linkage_service.go#L582-L609)

---

## Approval and Sign-off

| Role | Name | Approval Date | Signature |
|------|------|---------------|-----------|
| Backend Architect | @agent-sde3-backend-architect | 2025-10-16 | ✓ |
| Security Engineer | TBD | Pending | - |
| Product Owner | TBD | Pending | - |
| DevOps Lead | TBD | Pending | - |

---

**ADR Status:** Accepted (Implementation in Progress)
**Next Review:** After Phase 1 completion
**Supersedes:** None
**Superseded By:** None
**Last Modified:** 2025-10-16
