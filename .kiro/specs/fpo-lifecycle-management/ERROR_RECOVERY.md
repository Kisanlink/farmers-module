# FPO Lifecycle Error Recovery Mechanisms

## Overview

This document details comprehensive error recovery mechanisms for the FPO lifecycle management system, specifically addressing the "failed to get FPO reference: no matching records found" error and other potential failure scenarios.

## Error Categories

### 1. Missing Reference Errors
**Primary Error**: "failed to get FPO reference: no matching records found"

### 2. Partial Setup Failures
**Error**: Components of FPO setup fail (user groups, permissions, roles)

### 3. AAA Service Failures
**Error**: AAA service unavailable or returning errors

### 4. Data Inconsistency Errors
**Error**: Mismatch between AAA and local database

### 5. State Transition Errors
**Error**: Invalid state transitions or stuck states

## Recovery Mechanisms

### 1. Missing FPO Reference Recovery

#### Automatic Sync Recovery

```go
// GetOrSyncFPO - Primary recovery mechanism
func (s *FPOLifecycleService) GetOrSyncFPO(ctx context.Context, aaaOrgID string) (*FPORef, error) {
    // Step 1: Try local database
    fpoRef, err := s.repo.FindByAAAOrgID(ctx, aaaOrgID)
    if err == nil && fpoRef != nil {
        return fpoRef, nil
    }

    // Step 2: Log the miss for monitoring
    s.logger.Warn("FPO reference not found locally, attempting recovery",
        zap.String("aaa_org_id", aaaOrgID),
        zap.String("error", err.Error()))

    // Step 3: Attempt sync from AAA
    org, err := s.aaaService.GetOrganization(ctx, aaaOrgID)
    if err != nil {
        // Step 4: Check cache if AAA is down
        if cachedFPO := s.cache.Get(aaaOrgID); cachedFPO != nil {
            s.logger.Info("Using cached FPO reference",
                zap.String("aaa_org_id", aaaOrgID))
            return cachedFPO, nil
        }
        return nil, fmt.Errorf("failed to recover FPO reference: %w", err)
    }

    // Step 5: Create local reference
    fpoRef = s.createFPOFromAAA(org)
    if err := s.repo.Create(ctx, fpoRef); err != nil {
        // Handle duplicate key error
        if IsDuplicateKeyError(err) {
            return s.repo.FindByAAAOrgID(ctx, aaaOrgID)
        }
        return nil, err
    }

    // Step 6: Cache the reference
    s.cache.Set(aaaOrgID, fpoRef, 1*time.Hour)

    s.logger.Info("Successfully recovered FPO reference",
        zap.String("fpo_id", fpoRef.ID),
        zap.String("aaa_org_id", aaaOrgID))

    return fpoRef, nil
}
```

#### Manual Sync Endpoint

```go
// POST /api/v1/fpo/sync/{aaa_org_id}
func (h *FPOHandler) SyncFPOFromAAA(c *gin.Context) {
    aaaOrgID := c.Param("aaa_org_id")

    // Validate permissions
    if !hasPermission(c, "fpo:sync") {
        c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
        return
    }

    result, err := h.service.ForceSyncFromAAA(c.Request.Context(), aaaOrgID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Sync failed",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "FPO synchronized successfully",
        "data": result,
    })
}
```

#### Batch Recovery for Multiple FPOs

```go
// ReconcileAllFPOReferences - Batch recovery mechanism
func (s *FPOLifecycleService) ReconcileAllFPOReferences(ctx context.Context) error {
    // Get all organizations from AAA
    orgs, err := s.aaaService.ListOrganizations(ctx, map[string]interface{}{
        "type": "FPO",
    })
    if err != nil {
        return fmt.Errorf("failed to list organizations: %w", err)
    }

    var syncErrors []error
    successCount := 0

    for _, org := range orgs {
        orgID := org["org_id"].(string)

        // Check if exists locally
        _, err := s.repo.FindByAAAOrgID(ctx, orgID)
        if err == nil {
            continue // Already exists
        }

        // Sync missing FPO
        if _, err := s.SyncFPOFromAAA(ctx, orgID); err != nil {
            syncErrors = append(syncErrors, fmt.Errorf("org %s: %w", orgID, err))
        } else {
            successCount++
        }

        // Rate limiting
        time.Sleep(100 * time.Millisecond)
    }

    s.logger.Info("FPO reconciliation completed",
        zap.Int("success_count", successCount),
        zap.Int("error_count", len(syncErrors)))

    if len(syncErrors) > 0 {
        return fmt.Errorf("reconciliation partially failed: %d errors", len(syncErrors))
    }

    return nil
}
```

### 2. Partial Setup Failure Recovery

#### Retry Mechanism with Exponential Backoff

```go
type SetupRetryConfig struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
}

func (s *FPOLifecycleService) RetryFailedSetup(ctx context.Context, fpoID string) error {
    config := SetupRetryConfig{
        MaxAttempts:   3,
        InitialDelay:  1 * time.Second,
        MaxDelay:      30 * time.Second,
        BackoffFactor: 2.0,
    }

    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        return err
    }

    // Check retry eligibility
    if fpoRef.SetupAttempts >= config.MaxAttempts {
        return fmt.Errorf("maximum retry attempts (%d) exceeded", config.MaxAttempts)
    }

    // Calculate backoff delay
    delay := config.InitialDelay * time.Duration(math.Pow(config.BackoffFactor, float64(fpoRef.SetupAttempts)))
    if delay > config.MaxDelay {
        delay = config.MaxDelay
    }

    s.logger.Info("Scheduling setup retry",
        zap.String("fpo_id", fpoID),
        zap.Int("attempt", fpoRef.SetupAttempts+1),
        zap.Duration("delay", delay))

    // Schedule retry with delay
    time.AfterFunc(delay, func() {
        if err := s.performSetupRetry(context.Background(), fpoID); err != nil {
            s.logger.Error("Setup retry failed",
                zap.String("fpo_id", fpoID),
                zap.Error(err))
        }
    })

    return nil
}

func (s *FPOLifecycleService) performSetupRetry(ctx context.Context, fpoID string) error {
    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        return err
    }

    // Increment attempt counter
    fpoRef.SetupAttempts++
    now := time.Now()
    fpoRef.LastSetupAt = &now

    // Clear previous errors
    setupErrors := make(map[string]interface{})
    setupProgress := make(map[string]interface{})

    // Retry failed components only
    if fpoRef.SetupErrors != nil {
        for component, _ := range fpoRef.SetupErrors {
            switch component {
            case "user_groups":
                if err := s.retryUserGroups(ctx, fpoRef, setupProgress); err != nil {
                    setupErrors[component] = err.Error()
                }
            case "permissions":
                if err := s.retryPermissions(ctx, fpoRef, setupProgress); err != nil {
                    setupErrors[component] = err.Error()
                }
            case "ceo_role":
                if err := s.retryCEORole(ctx, fpoRef, setupProgress); err != nil {
                    setupErrors[component] = err.Error()
                }
            }
        }
    }

    // Update FPO with results
    if len(setupErrors) == 0 {
        fpoRef.Status = FPOStatusActive
        fpoRef.SetupErrors = nil
        fpoRef.SetupProgress = setupProgress
    } else {
        fpoRef.Status = FPOStatusSetupFailed
        fpoRef.SetupErrors = setupErrors
        fpoRef.SetupProgress = setupProgress
    }

    return s.repo.Update(ctx, fpoRef)
}
```

#### Component-Specific Recovery

```go
// Retry User Groups Creation
func (s *FPOLifecycleService) retryUserGroups(ctx context.Context, fpo *FPORef, progress map[string]interface{}) error {
    requiredGroups := []string{"directors", "shareholders", "store_staff", "store_managers"}
    createdGroups := []string{}

    // Check existing groups
    existingGroups, err := s.aaaService.ListUserGroups(ctx, fpo.AAAOrgID)
    if err != nil {
        return fmt.Errorf("failed to list existing groups: %w", err)
    }

    existingMap := make(map[string]bool)
    for _, group := range existingGroups {
        existingMap[group.Name] = true
    }

    // Create missing groups
    for _, groupName := range requiredGroups {
        if existingMap[groupName] {
            createdGroups = append(createdGroups, groupName)
            continue
        }

        req := map[string]interface{}{
            "name":        groupName,
            "description": fmt.Sprintf("%s group for %s", groupName, fpo.Name),
            "org_id":      fpo.AAAOrgID,
            "permissions": s.getGroupPermissions(groupName),
        }

        if _, err := s.aaaService.CreateUserGroup(ctx, req); err != nil {
            // Check if already exists error
            if !IsAlreadyExistsError(err) {
                return fmt.Errorf("failed to create group %s: %w", groupName, err)
            }
        }
        createdGroups = append(createdGroups, groupName)
    }

    progress["user_groups"] = createdGroups
    return nil
}

// Retry Permissions Assignment
func (s *FPOLifecycleService) retryPermissions(ctx context.Context, fpo *FPORef, progress map[string]interface{}) error {
    permissionErrors := make(map[string]string)

    groups, err := s.aaaService.ListUserGroups(ctx, fpo.AAAOrgID)
    if err != nil {
        return fmt.Errorf("failed to list groups: %w", err)
    }

    for _, group := range groups {
        permissions := s.getGroupPermissions(group.Name)
        for _, permission := range permissions {
            err := s.aaaService.AssignPermissionToGroup(ctx, group.ID, "fpo", permission)
            if err != nil && !IsAlreadyAssignedError(err) {
                permissionErrors[fmt.Sprintf("%s:%s", group.Name, permission)] = err.Error()
            }
        }
    }

    if len(permissionErrors) > 0 {
        return fmt.Errorf("permission assignment failed: %v", permissionErrors)
    }

    progress["permissions"] = "completed"
    return nil
}

// Retry CEO Role Assignment
func (s *FPOLifecycleService) retryCEORole(ctx context.Context, fpo *FPORef, progress map[string]interface{}) error {
    if fpo.CEOUserID == "" {
        return fmt.Errorf("CEO user ID not set")
    }

    // Check if already assigned
    hasRole, err := s.aaaService.CheckUserRole(ctx, fpo.CEOUserID, "CEO")
    if err != nil {
        return fmt.Errorf("failed to check CEO role: %w", err)
    }

    if hasRole {
        progress["ceo_role"] = "already_assigned"
        return nil
    }

    // Assign role
    if err := s.aaaService.AssignRole(ctx, fpo.CEOUserID, fpo.AAAOrgID, "CEO"); err != nil {
        return fmt.Errorf("failed to assign CEO role: %w", err)
    }

    progress["ceo_role"] = "assigned"
    return nil
}
```

### 3. AAA Service Failure Recovery

#### Circuit Breaker Implementation

```go
type CircuitBreaker struct {
    maxFailures      int
    resetTimeout     time.Duration
    failureCount     int
    lastFailureTime  time.Time
    state            CircuitState
    mutex            sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    // Check circuit state
    switch cb.state {
    case CircuitOpen:
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.state = CircuitHalfOpen
            cb.failureCount = 0
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }

    // Execute function
    err := fn()

    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()

        if cb.failureCount >= cb.maxFailures {
            cb.state = CircuitOpen
            return fmt.Errorf("circuit breaker opened: %w", err)
        }
        return err
    }

    // Success - reset state
    if cb.state == CircuitHalfOpen {
        cb.state = CircuitClosed
    }
    cb.failureCount = 0
    return nil
}

// Usage in AAA client
func (s *AAAServiceWithCircuitBreaker) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
    return s.circuitBreaker.Execute(func() error {
        return s.underlying.GetOrganization(ctx, orgID)
    })
}
```

#### Fallback Mechanisms

```go
type FallbackStrategy struct {
    cache        Cache
    staticData   map[string]interface{}
    degradedMode bool
}

func (s *FPOServiceWithFallback) GetFPORef(ctx context.Context, aaaOrgID string) (*FPORef, error) {
    // Try primary path
    fpo, err := s.primary.GetFPORef(ctx, aaaOrgID)
    if err == nil {
        return fpo, nil
    }

    s.logger.Warn("Primary FPO retrieval failed, trying fallback",
        zap.Error(err))

    // Fallback 1: Check cache
    if cached := s.cache.Get(fmt.Sprintf("fpo:%s", aaaOrgID)); cached != nil {
        s.logger.Info("Using cached FPO data")
        return cached.(*FPORef), nil
    }

    // Fallback 2: Check read replica
    if s.readReplica != nil {
        if fpo, err := s.readReplica.GetFPORef(ctx, aaaOrgID); err == nil {
            s.logger.Info("Retrieved FPO from read replica")
            return fpo, nil
        }
    }

    // Fallback 3: Degraded mode - return minimal data
    if s.degradedMode {
        return &FPORef{
            AAAOrgID: aaaOrgID,
            Status:   FPOStatusUnknown,
            Metadata: map[string]interface{}{
                "degraded_mode": true,
                "timestamp":     time.Now(),
            },
        }, nil
    }

    return nil, fmt.Errorf("all fallback mechanisms failed: %w", err)
}
```

### 4. Data Consistency Recovery

#### Consistency Checker

```go
type ConsistencyChecker struct {
    local      FPORepository
    aaaService AAAService
    logger     Logger
}

func (cc *ConsistencyChecker) CheckAndRepair(ctx context.Context) error {
    // Get all local FPOs
    localFPOs, err := cc.local.FindAll(ctx, nil)
    if err != nil {
        return err
    }

    inconsistencies := []Inconsistency{}

    for _, fpo := range localFPOs {
        // Check against AAA
        aaaOrg, err := cc.aaaService.GetOrganization(ctx, fpo.AAAOrgID)
        if err != nil {
            inconsistencies = append(inconsistencies, Inconsistency{
                Type:    "MISSING_IN_AAA",
                FPOID:   fpo.ID,
                Details: err.Error(),
            })
            continue
        }

        // Compare data
        if issues := cc.compareData(fpo, aaaOrg); len(issues) > 0 {
            inconsistencies = append(inconsistencies, Inconsistency{
                Type:    "DATA_MISMATCH",
                FPOID:   fpo.ID,
                Details: strings.Join(issues, ", "),
            })
        }
    }

    // Repair inconsistencies
    for _, inc := range inconsistencies {
        if err := cc.repair(ctx, inc); err != nil {
            cc.logger.Error("Failed to repair inconsistency",
                zap.String("fpo_id", inc.FPOID),
                zap.Error(err))
        }
    }

    return nil
}

func (cc *ConsistencyChecker) repair(ctx context.Context, inc Inconsistency) error {
    switch inc.Type {
    case "MISSING_IN_AAA":
        // Mark FPO as orphaned
        return cc.local.UpdateStatus(ctx, inc.FPOID, FPOStatusOrphaned,
            "AAA organization not found", "consistency_checker")

    case "DATA_MISMATCH":
        // Sync from AAA (AAA is source of truth)
        return cc.syncFromAAA(ctx, inc.FPOID)

    default:
        return fmt.Errorf("unknown inconsistency type: %s", inc.Type)
    }
}
```

### 5. State Transition Recovery

#### Stuck State Detection

```go
type StuckStateDetector struct {
    repo      FPORepository
    thresholds map[FPOStatus]time.Duration
}

func NewStuckStateDetector() *StuckStateDetector {
    return &StuckStateDetector{
        thresholds: map[FPOStatus]time.Duration{
            FPOStatusPendingVerification: 48 * time.Hour,
            FPOStatusPendingSetup:        1 * time.Hour,
            FPOStatusSetupFailed:         24 * time.Hour,
        },
    }
}

func (ssd *StuckStateDetector) DetectAndRecover(ctx context.Context) error {
    for status, threshold := range ssd.thresholds {
        filter := base.NewFilterBuilder().
            Where("status", base.OpEqual, status).
            Where("status_changed_at", base.OpLessThan, time.Now().Add(-threshold)).
            Build()

        stuckFPOs, err := ssd.repo.FindAll(ctx, filter)
        if err != nil {
            return err
        }

        for _, fpo := range stuckFPOs {
            if err := ssd.recoverStuckFPO(ctx, fpo); err != nil {
                log.Printf("Failed to recover stuck FPO %s: %v", fpo.ID, err)
            }
        }
    }

    return nil
}

func (ssd *StuckStateDetector) recoverStuckFPO(ctx context.Context, fpo *FPORef) error {
    switch fpo.Status {
    case FPOStatusPendingVerification:
        // Notify verifiers
        return ssd.notifyVerifiers(ctx, fpo)

    case FPOStatusPendingSetup:
        // Check setup progress and retry if needed
        return ssd.checkAndRetrySetup(ctx, fpo)

    case FPOStatusSetupFailed:
        // Auto-retry if under limit
        if fpo.SetupAttempts < 3 {
            return ssd.triggerSetupRetry(ctx, fpo)
        }
        // Otherwise mark as requiring manual intervention
        return ssd.escalateToAdmin(ctx, fpo)

    default:
        return nil
    }
}
```

## Recovery Monitoring

### Metrics

```go
var (
    recoveryAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "fpo_recovery_attempts_total",
            Help: "Total number of recovery attempts",
        },
        []string{"type", "status"},
    )

    recoveryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "fpo_recovery_duration_seconds",
            Help:    "Duration of recovery operations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"type"},
    )

    inconsistenciesFound = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "fpo_inconsistencies_found_total",
            Help: "Total number of inconsistencies found",
        },
        []string{"type"},
    )
)
```

### Alerting Rules

```yaml
groups:
  - name: fpo_recovery
    rules:
      - alert: HighRecoveryFailureRate
        expr: rate(fpo_recovery_attempts_total{status="failure"}[5m]) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: High FPO recovery failure rate

      - alert: ManyStuckFPOs
        expr: fpo_state_count{state="PENDING_SETUP"} > 10
        for: 30m
        labels:
          severity: critical
        annotations:
          summary: Many FPOs stuck in PENDING_SETUP

      - alert: ConsistencyCheckFailed
        expr: increase(fpo_inconsistencies_found_total[1h]) > 5
        labels:
          severity: warning
        annotations:
          summary: Multiple FPO inconsistencies detected
```

## Recovery Procedures

### Manual Recovery Steps

#### 1. Single FPO Recovery
```bash
# Check FPO status
curl -X GET "https://api.example.com/api/v1/fpo/{fpo_id}"

# Force sync from AAA
curl -X POST "https://api.example.com/api/v1/fpo/sync/{aaa_org_id}" \
  -H "Authorization: Bearer $TOKEN"

# Retry failed setup
curl -X POST "https://api.example.com/api/v1/fpo/{fpo_id}/retry-setup" \
  -H "Authorization: Bearer $TOKEN"
```

#### 2. Batch Recovery
```bash
# Reconcile all FPOs
curl -X POST "https://api.example.com/api/v1/admin/fpo/reconcile" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Check stuck FPOs
curl -X GET "https://api.example.com/api/v1/admin/fpo/stuck" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Trigger consistency check
curl -X POST "https://api.example.com/api/v1/admin/fpo/consistency-check" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Database Recovery Queries

```sql
-- Find FPOs missing AAA org ID
SELECT id, name, status, created_at
FROM fpo_refs
WHERE aaa_org_id IS NULL OR aaa_org_id = ''
  AND deleted_at IS NULL;

-- Find stuck FPOs
SELECT id, name, status, status_changed_at,
       EXTRACT(EPOCH FROM (NOW() - status_changed_at))/3600 as hours_in_state
FROM fpo_refs
WHERE status IN ('PENDING_SETUP', 'PENDING_VERIFICATION')
  AND status_changed_at < NOW() - INTERVAL '24 hours'
  AND deleted_at IS NULL;

-- Fix orphaned FPOs
UPDATE fpo_refs
SET status = 'INACTIVE',
    status_reason = 'Orphaned - AAA organization not found',
    status_changed_at = NOW()
WHERE aaa_org_id IN (
    SELECT aaa_org_id FROM fpo_refs
    WHERE aaa_org_id NOT IN (
        -- List of valid AAA org IDs from AAA service
    )
);

-- Reset failed setup attempts
UPDATE fpo_refs
SET setup_attempts = 0,
    status = 'VERIFIED',
    setup_errors = NULL
WHERE status = 'SETUP_FAILED'
  AND setup_attempts >= 3
  AND deleted_at IS NULL;
```

## Testing Recovery Mechanisms

### Unit Tests

```go
func TestFPORecovery(t *testing.T) {
    t.Run("Missing FPO Reference Recovery", func(t *testing.T) {
        // Setup
        mockAAA := NewMockAAAService()
        mockAAA.On("GetOrganization", "org_123").Return(&Organization{
            ID:   "org_123",
            Name: "Test FPO",
        }, nil)

        service := NewFPOLifecycleService(repo, mockAAA)

        // Test recovery
        fpo, err := service.GetOrSyncFPO(ctx, "org_123")

        assert.NoError(t, err)
        assert.NotNil(t, fpo)
        assert.Equal(t, "org_123", fpo.AAAOrgID)
    })

    t.Run("Setup Retry with Backoff", func(t *testing.T) {
        // Test exponential backoff calculation
        delays := calculateBackoffDelays(3, 1*time.Second, 2.0)
        assert.Equal(t, []time.Duration{
            1 * time.Second,
            2 * time.Second,
            4 * time.Second,
        }, delays)
    })

    t.Run("Circuit Breaker", func(t *testing.T) {
        cb := NewCircuitBreaker(3, 1*time.Minute)

        // Simulate failures
        for i := 0; i < 3; i++ {
            err := cb.Execute(func() error {
                return fmt.Errorf("service error")
            })
            assert.Error(t, err)
        }

        // Circuit should be open
        err := cb.Execute(func() error {
            return nil
        })
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "circuit breaker is open")
    })
}
```

## Recovery Runbook

### Level 1: Automated Recovery
1. Automatic sync from AAA on missing reference
2. Retry with exponential backoff for transient failures
3. Circuit breaker for AAA service protection
4. Cache fallback for read operations

### Level 2: Semi-Automated Recovery
1. Admin-triggered batch reconciliation
2. Stuck state detection and alerting
3. Consistency checks with repair suggestions
4. Manual retry triggers via API

### Level 3: Manual Intervention
1. Direct database updates for critical issues
2. AAA service manual reconciliation
3. Data export/import for disaster recovery
4. Complete FPO recreation if necessary

## Success Metrics

1. **Recovery Success Rate**: > 95% automated recovery
2. **Mean Time to Recovery**: < 5 minutes for automated, < 30 minutes for manual
3. **False Positive Rate**: < 1% for inconsistency detection
4. **Circuit Breaker Effectiveness**: < 0.1% cascading failures
5. **Cache Hit Rate**: > 80% during degraded mode
