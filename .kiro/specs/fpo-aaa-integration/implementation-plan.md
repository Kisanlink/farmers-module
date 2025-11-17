# FPO-AAA Integration Implementation Plan

**Version:** 1.0
**Date:** 2025-11-17
**Owner:** Backend Architecture Team

---

## Overview

This document provides detailed implementation guidance for integrating FPO creation with AAA service organization creation, based on [ADR-002](../adr-002-fpo-aaa-integration-architecture.md).

---

## Implementation Components

### 1. AAA Integration Service Layer

Create a new service layer that wraps the existing AAA client with enhanced capabilities:

**Location:** `/Users/kaushik/farmers-module/internal/services/aaa_integration/`

#### 1.1 Interface Definition

```go
// File: /Users/kaushik/farmers-module/internal/services/aaa_integration/interfaces.go

package aaa_integration

import (
    "context"
    "time"
)

// AAAIntegrationService provides enhanced AAA operations with resilience
type AAAIntegrationService interface {
    // Organization operations
    CreateOrganizationWithRetry(ctx context.Context, req *CreateOrgRequest) (*CreateOrgResponse, error)
    EnsureOrganizationExists(ctx context.Context, req *CreateOrgRequest) (string, error)

    // User operations
    EnsureUserExists(ctx context.Context, req *CreateUserRequest) (string, error)

    // Group operations
    EnsureUserGroupExists(ctx context.Context, req *CreateGroupRequest) (string, error)

    // Role operations
    EnsureRoleAssignment(ctx context.Context, userID, orgID, role string) error

    // Health operations
    CheckOrganizationHealth(ctx context.Context, orgID string) (*OrgHealth, error)
}

// Request/Response types
type CreateOrgRequest struct {
    Name           string            `json:"name"`
    Type           string            `json:"type"`
    Description    string            `json:"description"`
    CEOUserID      string            `json:"ceo_user_id"`
    Metadata       map[string]string `json:"metadata"`
    IdempotencyKey string            `json:"idempotency_key,omitempty"`
}

type CreateOrgResponse struct {
    OrgID     string    `json:"org_id"`
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
    PhoneNumber string            `json:"phone_number"`
    FirstName   string            `json:"first_name"`
    LastName    string            `json:"last_name"`
    Email       string            `json:"email,omitempty"`
    Password    string            `json:"password,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type CreateGroupRequest struct {
    Name        string   `json:"name"`
    OrgID       string   `json:"org_id"`
    Description string   `json:"description"`
    Permissions []string `json:"permissions"`
}

type OrgHealth struct {
    OrgID        string    `json:"org_id"`
    Status       string    `json:"status"`
    CEOAssigned  bool      `json:"ceo_assigned"`
    GroupsCount  int       `json:"groups_count"`
    UsersCount   int       `json:"users_count"`
    LastChecked  time.Time `json:"last_checked"`
}
```

#### 1.2 Implementation with Retry Logic

```go
// File: /Users/kaushik/farmers-module/internal/services/aaa_integration/aaa_integration_service.go

package aaa_integration

import (
    "context"
    "crypto/sha256"
    "fmt"
    "time"

    "github.com/Kisanlink/farmers-module/internal/clients/aaa"
    "github.com/Kisanlink/farmers-module/internal/interfaces"
    "github.com/avast/retry-go"
    "github.com/sony/gobreaker"
    "go.uber.org/zap"
)

type AAAIntegrationServiceImpl struct {
    aaaClient      *aaa.Client
    logger         interfaces.Logger
    cache          interfaces.Cache
    circuitBreaker *gobreaker.CircuitBreaker
    retryConfig    RetryConfig
}

type RetryConfig struct {
    MaxAttempts    uint
    InitialDelay   time.Duration
    MaxDelay       time.Duration
    BackoffFactor  float64
}

func NewAAAIntegrationService(
    aaaClient *aaa.Client,
    logger interfaces.Logger,
    cache interfaces.Cache,
) AAAIntegrationService {
    // Circuit breaker configuration
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "AAA-Service",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    })

    return &AAAIntegrationServiceImpl{
        aaaClient:      aaaClient,
        logger:         logger,
        cache:          cache,
        circuitBreaker: cb,
        retryConfig: RetryConfig{
            MaxAttempts:   3,
            InitialDelay:  1 * time.Second,
            MaxDelay:      30 * time.Second,
            BackoffFactor: 2.0,
        },
    }
}

func (s *AAAIntegrationServiceImpl) CreateOrganizationWithRetry(
    ctx context.Context,
    req *CreateOrgRequest,
) (*CreateOrgResponse, error) {
    var resp *CreateOrgResponse

    operation := func() error {
        // Use circuit breaker
        result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
            // Call AAA service
            aaaReq := &aaa.CreateOrganizationRequest{
                Name:        req.Name,
                Type:        req.Type,
                Description: req.Description,
                CEOUserID:   req.CEOUserID,
                Metadata:    req.Metadata,
            }

            aaaResp, err := s.aaaClient.CreateOrganization(ctx, aaaReq)
            if err != nil {
                return nil, err
            }

            return &CreateOrgResponse{
                OrgID:     aaaResp.OrgID,
                Name:      aaaResp.Name,
                Status:    aaaResp.Status,
                CreatedAt: aaaResp.CreatedAt,
            }, nil
        })

        if err != nil {
            return err
        }

        resp = result.(*CreateOrgResponse)
        return nil
    }

    // Retry with exponential backoff
    err := retry.Do(
        operation,
        retry.Attempts(s.retryConfig.MaxAttempts),
        retry.Delay(s.retryConfig.InitialDelay),
        retry.MaxDelay(s.retryConfig.MaxDelay),
        retry.DelayType(retry.BackOffDelay),
        retry.OnRetry(func(n uint, err error) {
            s.logger.Warn("Retrying AAA organization creation",
                zap.Uint("attempt", n),
                zap.String("org_name", req.Name),
                zap.Error(err))
        }),
        retry.RetryIf(func(err error) bool {
            // Don't retry on business errors
            return !isBusinessError(err)
        }),
    )

    if err != nil {
        s.logger.Error("Failed to create organization after retries",
            zap.String("org_name", req.Name),
            zap.Error(err))
        return nil, err
    }

    return resp, nil
}

func (s *AAAIntegrationServiceImpl) EnsureUserExists(
    ctx context.Context,
    req *CreateUserRequest,
) (string, error) {
    // Generate cache key
    cacheKey := fmt.Sprintf("user:phone:%s", req.PhoneNumber)

    // Check cache
    if userID, err := s.cache.Get(ctx, cacheKey); err == nil {
        return userID.(string), nil
    }

    // Try to get existing user
    existingUser, err := s.aaaClient.GetUserByPhone(ctx, req.PhoneNumber)
    if err == nil && existingUser != nil {
        userID := existingUser.ID
        // Cache for 5 minutes
        s.cache.Set(ctx, cacheKey, userID, 5*time.Minute)
        return userID, nil
    }

    // Create new user
    createReq := &aaa.CreateUserRequest{
        Username:    fmt.Sprintf("%s_%s", req.FirstName, req.LastName),
        PhoneNumber: req.PhoneNumber,
        CountryCode: "+91",
        Email:       req.Email,
        Password:    req.Password,
        FullName:    fmt.Sprintf("%s %s", req.FirstName, req.LastName),
        Metadata:    req.Metadata,
    }

    resp, err := s.aaaClient.CreateUser(ctx, createReq)
    if err != nil {
        // Check if it's a duplicate error
        if isDuplicateUserError(err) {
            // Try to fetch again
            existingUser, err = s.aaaClient.GetUserByPhone(ctx, req.PhoneNumber)
            if err == nil && existingUser != nil {
                userID := existingUser.ID
                s.cache.Set(ctx, cacheKey, userID, 5*time.Minute)
                return userID, nil
            }
        }
        return "", fmt.Errorf("failed to ensure user exists: %w", err)
    }

    // Cache the user ID
    s.cache.Set(ctx, cacheKey, resp.UserID, 5*time.Minute)
    return resp.UserID, nil
}

func (s *AAAIntegrationServiceImpl) EnsureRoleAssignment(
    ctx context.Context,
    userID, orgID, role string,
) error {
    // Check if role is already assigned
    hasRole, err := s.aaaClient.CheckUserRole(ctx, userID, role)
    if err == nil && hasRole {
        s.logger.Info("User already has role",
            zap.String("user_id", userID),
            zap.String("role", role))
        return nil
    }

    // Assign role
    err = retry.Do(
        func() error {
            return s.aaaClient.AssignRole(ctx, userID, orgID, role)
        },
        retry.Attempts(s.retryConfig.MaxAttempts),
        retry.Delay(s.retryConfig.InitialDelay),
        retry.OnRetry(func(n uint, err error) {
            s.logger.Warn("Retrying role assignment",
                zap.Uint("attempt", n),
                zap.String("user_id", userID),
                zap.String("role", role),
                zap.Error(err))
        }),
    )

    return err
}

// Helper functions
func isBusinessError(err error) bool {
    // Add business error detection logic
    return false
}

func isDuplicateUserError(err error) bool {
    // Add duplicate user error detection
    return false
}
```

### 2. Enhanced FPO Service

Update the existing FPO service to use the new integration layer:

```go
// File: /Users/kaushik/farmers-module/internal/services/fpo_service_enhanced.go

package services

import (
    "context"
    "fmt"
    "time"

    "github.com/Kisanlink/farmers-module/internal/entities"
    "github.com/Kisanlink/farmers-module/internal/entities/fpo"
    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/entities/responses"
    "github.com/Kisanlink/farmers-module/internal/services/aaa_integration"
    "go.uber.org/zap"
)

type EnhancedFPOService struct {
    fpoRepo         FPORefRepository
    aaaIntegration  aaa_integration.AAAIntegrationService
    eventBus        interfaces.EventEmitter
    logger          interfaces.Logger
}

func NewEnhancedFPOService(
    fpoRepo FPORefRepository,
    aaaIntegration aaa_integration.AAAIntegrationService,
    eventBus interfaces.EventEmitter,
    logger interfaces.Logger,
) FPOService {
    return &EnhancedFPOService{
        fpoRepo:        fpoRepo,
        aaaIntegration: aaaIntegration,
        eventBus:       eventBus,
        logger:         logger,
    }
}

// Implementation follows the pattern in ADR-002
```

### 3. Reconciliation Service

Create a reconciliation service for handling inconsistencies:

```go
// File: /Users/kaushik/farmers-module/internal/services/reconciliation/reconciliation_service.go

package reconciliation

import (
    "context"
    "errors"
    "time"

    "github.com/Kisanlink/farmers-module/internal/clients/aaa"
    "github.com/Kisanlink/farmers-module/internal/entities/fpo"
    "github.com/Kisanlink/farmers-module/internal/interfaces"
    "go.uber.org/zap"
)

type ReconciliationService struct {
    fpoRepo   FPORepository
    aaaClient *aaa.Client
    logger    interfaces.Logger
    metrics   MetricsCollector
}

func NewReconciliationService(
    fpoRepo FPORepository,
    aaaClient *aaa.Client,
    logger interfaces.Logger,
    metrics MetricsCollector,
) *ReconciliationService {
    return &ReconciliationService{
        fpoRepo:   fpoRepo,
        aaaClient: aaaClient,
        logger:    logger,
        metrics:   metrics,
    }
}

func (r *ReconciliationService) ReconcileFPOStatus(
    ctx context.Context,
    orgID string,
) error {
    r.logger.Info("Starting FPO reconciliation",
        zap.String("org_id", orgID))

    // Implementation as defined in ADR-002
    return nil
}

func (r *ReconciliationService) RunPeriodicReconciliation(ctx context.Context) {
    ticker := time.NewTicker(15 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            r.reconcileAll(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (r *ReconciliationService) reconcileAll(ctx context.Context) {
    // Get all FPOs with PENDING_SETUP status
    pendingFPOs, err := r.fpoRepo.FindByStatus(ctx, fpo.FPOStatusPendingSetup)
    if err != nil {
        r.logger.Error("Failed to fetch pending FPOs", zap.Error(err))
        return
    }

    for _, fpoRef := range pendingFPOs {
        if err := r.ReconcileFPOStatus(ctx, fpoRef.AAAOrgID); err != nil {
            r.logger.Error("Failed to reconcile FPO",
                zap.String("org_id", fpoRef.AAAOrgID),
                zap.Error(err))
        }
    }
}
```

### 4. Database Migration

Add new fields to support the enhanced FPO tracking:

```sql
-- File: /Users/kaushik/farmers-module/migrations/20251117_add_fpo_reconciliation_fields.sql

ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS setup_errors JSONB DEFAULT '{}';
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMP;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS sync_version INTEGER DEFAULT 0;

-- Add index for reconciliation queries
CREATE INDEX IF NOT EXISTS idx_fpo_refs_status_sync
ON fpo_refs(status, last_sync_at)
WHERE status = 'PENDING_SETUP';

-- Add index for AAA org ID lookups
CREATE INDEX IF NOT EXISTS idx_fpo_refs_aaa_org_id
ON fpo_refs(aaa_org_id);
```

### 5. Configuration Updates

Add configuration for the new services:

```yaml
# File: /Users/kaushik/farmers-module/config/config.yaml

aaa:
  grpc_endpoint: "${AAA_GRPC_ENDPOINT}"
  api_key: "${AAA_API_KEY}"
  request_timeout: "${AAA_REQUEST_TIMEOUT:10s}"

  # Retry configuration
  retry:
    max_attempts: 3
    initial_delay: 1s
    max_delay: 30s
    backoff_factor: 2.0

  # Circuit breaker configuration
  circuit_breaker:
    failure_threshold: 5
    success_threshold: 2
    timeout: 60s
    half_open_requests: 3

# Reconciliation configuration
reconciliation:
  enabled: true
  interval: 15m
  batch_size: 10
  max_retries: 3

# Cache configuration
cache:
  type: "redis"
  redis:
    host: "${REDIS_HOST:localhost}"
    port: "${REDIS_PORT:6379}"
    password: "${REDIS_PASSWORD:}"
    db: 0
    ttl: 5m
```

---

## Implementation Tasks

### Phase 1: Core Integration (Priority: HIGH)

#### Task 1.1: Create AAA Integration Service
- **File:** `/Users/kaushik/farmers-module/internal/services/aaa_integration/aaa_integration_service.go`
- **Dependencies:** AAA client, retry library, circuit breaker
- **Estimate:** 2 days
- **Testing:** Unit tests with mocked AAA client

#### Task 1.2: Update FPO Entity Model
- **File:** `/Users/kaushik/farmers-module/internal/entities/fpo/fpo.go`
- **Changes:** Add SetupErrors, LastSyncAt, SyncVersion fields
- **Estimate:** 0.5 days
- **Testing:** Model validation tests

#### Task 1.3: Enhance FPO Service
- **File:** `/Users/kaushik/farmers-module/internal/services/fpo_service_enhanced.go`
- **Changes:** Integrate with AAA Integration Service
- **Estimate:** 2 days
- **Testing:** Integration tests

#### Task 1.4: Database Migration
- **File:** `/Users/kaushik/farmers-module/migrations/20251117_add_fpo_reconciliation_fields.sql`
- **Estimate:** 0.5 days
- **Testing:** Migration rollback tests

### Phase 2: Resilience (Priority: HIGH)

#### Task 2.1: Implement Circuit Breaker
- **Package:** `github.com/sony/gobreaker`
- **Integration:** AAA Integration Service
- **Estimate:** 1 day
- **Testing:** Failure scenario tests

#### Task 2.2: Add Retry Logic
- **Package:** `github.com/avast/retry-go`
- **Integration:** All AAA operations
- **Estimate:** 1 day
- **Testing:** Retry behavior tests

#### Task 2.3: Implement Caching
- **Type:** Redis cache for idempotency
- **Estimate:** 1 day
- **Testing:** Cache hit/miss tests

### Phase 3: Recovery (Priority: MEDIUM)

#### Task 3.1: Create Reconciliation Service
- **File:** `/Users/kaushik/farmers-module/internal/services/reconciliation/reconciliation_service.go`
- **Estimate:** 2 days
- **Testing:** Reconciliation scenario tests

#### Task 3.2: Add CompleteFPOSetup Endpoint
- **Endpoint:** `POST /api/v1/identity/fpo/{org_id}/complete-setup`
- **Estimate:** 1 day
- **Testing:** API integration tests

#### Task 3.3: Implement Scheduled Jobs
- **Type:** Cron-based reconciliation
- **Estimate:** 1 day
- **Testing:** Job execution tests

### Phase 4: Monitoring (Priority: MEDIUM)

#### Task 4.1: Add Metrics Collection
- **Metrics:** Creation success/failure, retry counts, circuit breaker status
- **Estimate:** 1 day
- **Testing:** Metrics accuracy tests

#### Task 4.2: Implement Health Checks
- **Endpoint:** `/health/fpo-integration`
- **Estimate:** 0.5 days
- **Testing:** Health check scenarios

#### Task 4.3: Setup Alerting
- **Alerts:** Failure rates, pending FPOs, AAA unreachable
- **Estimate:** 0.5 days
- **Testing:** Alert triggering tests

---

## Testing Strategy

### Unit Tests

```go
// File: /Users/kaushik/farmers-module/internal/services/fpo_service_test.go

func TestCreateFPO_Success(t *testing.T) {
    // Test successful FPO creation
}

func TestCreateFPO_PartialFailure_PendingSetup(t *testing.T) {
    // Test PENDING_SETUP status on partial failure
}

func TestCreateFPO_AAAFailure_Retry(t *testing.T) {
    // Test retry behavior on AAA failure
}

func TestCreateFPO_Idempotency(t *testing.T) {
    // Test idempotent behavior
}
```

### Integration Tests

```go
// File: /Users/kaushik/farmers-module/test/integration/fpo_integration_test.go

func TestFPOCreation_EndToEnd(t *testing.T) {
    // Test complete flow with real AAA service
}

func TestReconciliation_HealsInconsistencies(t *testing.T) {
    // Test reconciliation service
}
```

### Load Tests

```go
// File: /Users/kaushik/farmers-module/test/load/fpo_load_test.go

func TestFPOCreation_UnderLoad(t *testing.T) {
    // Test system behavior under load
}
```

---

## Rollout Plan

### Stage 1: Development Environment
- Deploy enhanced FPO service
- Run integration tests
- Monitor for 24 hours

### Stage 2: Staging Environment
- Deploy with feature flag
- Test with production-like data
- Run chaos tests
- Monitor for 48 hours

### Stage 3: Production (Canary)
- Deploy to 10% of traffic
- Monitor metrics closely
- Gradual rollout over 1 week

### Stage 4: Full Production
- Enable for all traffic
- Enable reconciliation jobs
- Full monitoring and alerting

---

## Success Criteria

1. **Functional:**
   - FPO creation success rate > 99%
   - Reconciliation success rate > 95%
   - No orphaned records after 24 hours

2. **Performance:**
   - P95 latency < 2 seconds
   - P99 latency < 5 seconds
   - Circuit breaker trips < 1/day

3. **Reliability:**
   - Zero data loss
   - Recovery from AAA outage < 5 minutes
   - Automatic healing of inconsistencies

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| AAA service overload | Rate limiting, circuit breaker |
| Network partitions | Retry with backoff, reconciliation |
| Data inconsistency | Reconciliation service, monitoring |
| Performance degradation | Caching, async operations |

---

## Documentation Requirements

1. **API Documentation:** Update OpenAPI spec
2. **Runbooks:** Create operational procedures
3. **Architecture Diagrams:** Update system diagrams
4. **Monitoring Dashboard:** Create Grafana dashboard
5. **Team Training:** Conduct knowledge transfer session
