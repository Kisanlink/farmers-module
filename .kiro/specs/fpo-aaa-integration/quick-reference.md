# FPO-AAA Integration Quick Reference

## Key Files to Modify

### 1. Core Service Files
- **Current FPO Service:** `/Users/kaushik/farmers-module/internal/services/fpo_ref_service.go`
- **AAA Client:** `/Users/kaushik/farmers-module/internal/clients/aaa/aaa_client.go`
- **FPO Entity:** `/Users/kaushik/farmers-module/internal/entities/fpo/fpo.go`
- **FPO Handler:** `/Users/kaushik/farmers-module/internal/handlers/fpo_handlers.go`

### 2. New Files to Create
```bash
# AAA Integration Service
/Users/kaushik/farmers-module/internal/services/aaa_integration/
├── interfaces.go
├── aaa_integration_service.go
├── retry_config.go
└── aaa_integration_service_test.go

# Reconciliation Service
/Users/kaushik/farmers-module/internal/services/reconciliation/
├── reconciliation_service.go
├── reconciliation_jobs.go
└── reconciliation_service_test.go

# Event Definitions
/Users/kaushik/farmers-module/internal/events/
├── fpo_events.go
└── reconciliation_events.go
```

## Critical Code Changes

### 1. Update FPORef Entity
```go
// In: /Users/kaushik/farmers-module/internal/entities/fpo/fpo.go

type FPORef struct {
    base.BaseModel
    AAAOrgID       string         `gorm:"unique;not null;index"`
    Name           string         `gorm:"not null"`
    RegistrationNo string         `gorm:"index"`
    Status         FPOStatus      `gorm:"type:varchar(50);default:'ACTIVE'"`

    // NEW FIELDS
    SetupErrors    entities.JSONB `gorm:"type:jsonb"`
    LastSyncAt     *time.Time     `gorm:"index"`
    SyncVersion    int            `gorm:"default:0"`

    BusinessConfig entities.JSONB `gorm:"type:jsonb;default:'{}'"`
}
```

### 2. Modify CreateFPO Method
```go
// In: /Users/kaushik/farmers-module/internal/services/fpo_ref_service.go

func (s *FPOServiceImpl) CreateFPO(ctx context.Context, req interface{}) (interface{}, error) {
    // CHANGE: Use AAAIntegrationService instead of direct AAAService

    // OLD:
    // orgResp, err := s.aaaService.CreateOrganization(ctx, createOrgReq)

    // NEW:
    orgResp, err := s.aaaIntegration.CreateOrganizationWithRetry(ctx, &aaa_integration.CreateOrgRequest{
        Name:        req.Name,
        Type:        "FPO",
        Description: req.Description,
        CEOUserID:   ceoUserID,
        Metadata:    req.Metadata,
    })
}
```

### 3. Add Retry Configuration
```go
// In: /Users/kaushik/farmers-module/internal/config/config.go

type AAAConfig struct {
    GRPCEndpoint   string        `env:"AAA_GRPC_ENDPOINT"`
    APIKey         string        `env:"AAA_API_KEY"`
    RequestTimeout string        `env:"AAA_REQUEST_TIMEOUT"`

    // NEW FIELDS
    Retry struct {
        MaxAttempts   int           `env:"AAA_RETRY_MAX_ATTEMPTS" default:"3"`
        InitialDelay  time.Duration `env:"AAA_RETRY_INITIAL_DELAY" default:"1s"`
        MaxDelay      time.Duration `env:"AAA_RETRY_MAX_DELAY" default:"30s"`
        BackoffFactor float64       `env:"AAA_RETRY_BACKOFF_FACTOR" default:"2.0"`
    }

    CircuitBreaker struct {
        FailureThreshold int           `env:"AAA_CB_FAILURE_THRESHOLD" default:"5"`
        SuccessThreshold int           `env:"AAA_CB_SUCCESS_THRESHOLD" default:"2"`
        Timeout          time.Duration `env:"AAA_CB_TIMEOUT" default:"60s"`
    }
}
```

## Environment Variables

Add to `.env` file:
```bash
# AAA Retry Configuration
AAA_RETRY_MAX_ATTEMPTS=3
AAA_RETRY_INITIAL_DELAY=1s
AAA_RETRY_MAX_DELAY=30s
AAA_RETRY_BACKOFF_FACTOR=2.0

# Circuit Breaker Configuration
AAA_CB_FAILURE_THRESHOLD=5
AAA_CB_SUCCESS_THRESHOLD=2
AAA_CB_TIMEOUT=60s

# Reconciliation Configuration
RECONCILIATION_ENABLED=true
RECONCILIATION_INTERVAL=15m
RECONCILIATION_BATCH_SIZE=10

# Redis Cache (for idempotency)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Dependencies to Add

Update `go.mod`:
```go
require (
    github.com/avast/retry-go v3.0.0
    github.com/sony/gobreaker v0.5.0
    github.com/go-redis/redis/v8 v8.11.5
    github.com/prometheus/client_golang v1.14.0
)
```

Install:
```bash
go get github.com/avast/retry-go
go get github.com/sony/gobreaker
go get github.com/go-redis/redis/v8
go get github.com/prometheus/client_golang
```

## Database Migration

Run migration:
```sql
-- Add new fields to fpo_refs table
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS setup_errors JSONB DEFAULT '{}';
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMP;
ALTER TABLE fpo_refs ADD COLUMN IF NOT EXISTS sync_version INTEGER DEFAULT 0;

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_fpo_refs_status_sync ON fpo_refs(status, last_sync_at) WHERE status = 'PENDING_SETUP';
CREATE INDEX IF NOT EXISTS idx_fpo_refs_aaa_org_id ON fpo_refs(aaa_org_id);
```

## API Endpoints

### Existing (Modified Behavior)
```
POST /api/v1/identity/fpo/create
- Now creates AAA organization with retry
- Returns PENDING_SETUP status if partial failure
- Supports idempotent retries
```

### New Endpoints
```
POST /api/v1/identity/fpo/{org_id}/complete-setup
- Completes pending setup for FPOs in PENDING_SETUP status
- Retries failed operations

GET /api/v1/identity/fpo/{org_id}/health
- Returns health status of FPO integration
- Shows setup completion status

POST /api/v1/admin/fpo/reconcile
- Triggers manual reconciliation
- Admin endpoint (requires special permission)
```

## Testing Commands

### Unit Tests
```bash
# Test new AAA Integration Service
go test ./internal/services/aaa_integration/...

# Test enhanced FPO Service
go test ./internal/services/... -run TestFPO

# Test reconciliation
go test ./internal/services/reconciliation/...
```

### Integration Tests
```bash
# Run with test AAA service
AAA_GRPC_ENDPOINT=localhost:50051 go test ./test/integration/fpo_test.go
```

### Manual Testing
```bash
# Create FPO with retry behavior
curl -X POST http://localhost:8000/api/v1/identity/fpo/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test FPO",
    "registration_number": "REG123",
    "ceo_user": {
      "first_name": "John",
      "last_name": "Doe",
      "phone_number": "9876543210",
      "email": "john@example.com",
      "password": "secure123"
    }
  }'

# Complete pending setup
curl -X POST http://localhost:8000/api/v1/identity/fpo/{org_id}/complete-setup \
  -H "Authorization: Bearer $TOKEN"

# Check FPO health
curl http://localhost:8000/api/v1/identity/fpo/{org_id}/health \
  -H "Authorization: Bearer $TOKEN"
```

## Monitoring Queries

### Prometheus Metrics
```
# FPO creation success rate
rate(fpo_creation_success_total[5m]) / rate(fpo_creation_total[5m])

# AAA retry rate
rate(aaa_retry_total[5m])

# Circuit breaker status
aaa_circuit_breaker_open

# Pending FPOs count
fpo_pending_setup_count
```

### PostgreSQL Queries
```sql
-- FPOs in PENDING_SETUP status
SELECT id, aaa_org_id, name, setup_errors, created_at
FROM fpo_refs
WHERE status = 'PENDING_SETUP'
ORDER BY created_at DESC;

-- FPOs needing reconciliation
SELECT id, aaa_org_id, name, last_sync_at
FROM fpo_refs
WHERE status = 'PENDING_SETUP'
  AND (last_sync_at IS NULL OR last_sync_at < NOW() - INTERVAL '1 hour')
ORDER BY last_sync_at ASC NULLS FIRST;

-- Setup error statistics
SELECT
  jsonb_object_keys(setup_errors) as error_type,
  COUNT(*) as count
FROM fpo_refs
WHERE status = 'PENDING_SETUP'
  AND setup_errors != '{}'
GROUP BY error_type
ORDER BY count DESC;
```

## Common Issues and Solutions

### Issue 1: AAA Service Timeout
**Symptom:** `context deadline exceeded` errors
**Solution:**
- Check AAA service health
- Increase `AAA_REQUEST_TIMEOUT`
- Verify network connectivity

### Issue 2: Too Many Retries
**Symptom:** High retry count in metrics
**Solution:**
- Check circuit breaker status
- Review error logs for root cause
- Adjust retry configuration

### Issue 3: FPOs Stuck in PENDING_SETUP
**Symptom:** FPOs remain in PENDING_SETUP status
**Solution:**
- Run manual reconciliation
- Check `setup_errors` field
- Use `complete-setup` endpoint

### Issue 4: Duplicate Organizations
**Symptom:** Multiple orgs with same name in AAA
**Solution:**
- Implement idempotency checks
- Use deterministic org IDs
- Add unique constraints

## Rollback Plan

If issues arise after deployment:

1. **Revert Code:**
```bash
git revert HEAD
git push origin main
```

2. **Database Rollback:**
```sql
-- Remove new columns (if needed)
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS setup_errors;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS last_sync_at;
ALTER TABLE fpo_refs DROP COLUMN IF EXISTS sync_version;
```

3. **Disable Reconciliation:**
```bash
RECONCILIATION_ENABLED=false
```

4. **Switch to Old Service:**
- Use feature flag to route to old implementation
- Or deploy previous Docker image

## Support Contacts

- **Architecture Team:** @agent-sde3-backend-architect
- **Implementation Team:** @agent-sde-backend-engineer
- **Testing Team:** @agent-business-logic-tester
- **On-Call:** Check PagerDuty schedule

## References

- [ADR-002: FPO-AAA Integration Architecture](../adr-002-fpo-aaa-integration-architecture.md)
- [Implementation Plan](./implementation-plan.md)
- [AAA Service Documentation](https://docs.kisanlink.com/aaa-service)
- [Farmers Module API Spec](../../docs/api/openapi.yaml)
