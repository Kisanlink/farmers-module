# Architecture Decision Record: Crop Cycles Area Allocation

**Title**: Enhanced Crop Cycle Management with Area Allocation and Stage-based Activity Organization
**Date**: 2025-01-14
**Status**: Proposed
**Authors**: SDE3 Backend Architect

## Context and Problem Statement

The farmers-module currently manages crop cycles without tracking area allocation, leading to potential over-allocation of farm land and inability to track resource utilization accurately. Additionally, farm activities lack organization by crop growth stages, making it difficult to track farming progress systematically.

### Current Limitations
1. No area tracking for crop cycles - cannot determine land utilization
2. No validation to prevent over-allocation of farm area
3. Farm activities not linked to crop stages - difficult to track stage-wise progress
4. No mechanism to handle concurrent area allocations
5. Limited visibility into farm resource optimization

### Business Drivers
- **Precision Agriculture**: Need for accurate land utilization tracking
- **Resource Optimization**: Prevent wastage through over-allocation
- **Operational Efficiency**: Organize activities by growth stages
- **Data-Driven Decisions**: Enable yield per hectare analytics

## Decision Drivers

1. **Data Integrity**: Must ensure sum of allocated areas never exceeds farm area
2. **Concurrency**: Handle simultaneous area allocation requests safely
3. **Performance**: Area validation must not impact user experience
4. **Backward Compatibility**: Existing systems must continue functioning
5. **Scalability**: Solution must handle farms with numerous crop cycles
6. **Auditability**: All area changes must be traceable

## Considered Options

### Option 1: Application-Level Validation Only
- **Pros**: Simple implementation, flexible business rules
- **Cons**: Race conditions possible, data integrity not guaranteed

### Option 2: Database Constraints Only
- **Pros**: Guaranteed data integrity, no race conditions
- **Cons**: Complex constraint logic, difficult error handling

### Option 3: Hybrid Approach (Selected)
- **Pros**: Best of both worlds, guaranteed integrity with good UX
- **Cons**: More complex implementation

## Decision

We will implement a **hybrid approach** combining application-level validation with database-level constraints and transactions:

### 1. Database Schema Enhancement

#### Crop Cycles Table Modification
```sql
ALTER TABLE crop_cycles
ADD COLUMN area_ha DECIMAL(12,4) CHECK (area_ha > 0);

-- Add index for performance
CREATE INDEX idx_crop_cycles_area ON crop_cycles(farm_id, status, area_ha);
```

#### Farm Activities Table Modification
```sql
ALTER TABLE farm_activities
ADD COLUMN crop_stage_id VARCHAR(20) REFERENCES crop_stages(id);

-- Add index for stage-based queries
CREATE INDEX idx_farm_activities_stage ON farm_activities(crop_stage_id, crop_cycle_id);
```

#### Area Allocation Tracking (New Table)
```sql
CREATE TABLE farm_area_allocations (
    id VARCHAR(20) PRIMARY KEY,
    farm_id VARCHAR(20) NOT NULL REFERENCES farms(id),
    total_area_ha DECIMAL(12,4) NOT NULL,
    allocated_area_ha DECIMAL(12,4) NOT NULL DEFAULT 0,
    available_area_ha DECIMAL(12,4) GENERATED ALWAYS AS (total_area_ha - allocated_area_ha) STORED,
    last_validated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_allocated_not_exceed CHECK (allocated_area_ha <= total_area_ha),
    CONSTRAINT chk_positive_areas CHECK (total_area_ha > 0 AND allocated_area_ha >= 0)
);

CREATE UNIQUE INDEX idx_farm_area_allocations_farm ON farm_area_allocations(farm_id);
```

### 2. Domain Model Changes

#### Enhanced CropCycle Entity
```go
type CropCycle struct {
    base.BaseModel
    FarmID    string         `json:"farm_id" gorm:"type:varchar(20);not null;index"`
    FarmerID  string         `json:"farmer_id" gorm:"type:varchar(20);not null;index"`
    AreaHa    float64        `json:"area_ha" gorm:"type:decimal(12,4);check:area_ha > 0"`  // NEW
    Season    string         `json:"season" gorm:"type:season;not null"`
    Status    string         `json:"status" gorm:"type:cycle_status;not null;default:'PLANNED'"`
    StartDate *time.Time     `json:"start_date" gorm:"type:date"`
    EndDate   *time.Time     `json:"end_date" gorm:"type:date"`
    CropID    string         `json:"crop_id" gorm:"type:varchar(20);not null;index"`
    VarietyID *string        `json:"variety_id" gorm:"type:varchar(20);index"`
    Outcome   entities.JSONB `json:"outcome" gorm:"type:jsonb;default:'{}';serializer:json"`

    // Relationships
    Farm    *farm.Farm                `json:"farm,omitempty" gorm:"foreignKey:FarmID"`
    Farmer  *farmer.Farmer            `json:"farmer,omitempty" gorm:"foreignKey:FarmerID"`
    Crop    *crop.Crop                `json:"crop,omitempty" gorm:"foreignKey:CropID"`
    Variety *crop_variety.CropVariety `json:"variety,omitempty" gorm:"foreignKey:VarietyID"`
}
```

#### Enhanced FarmActivity Entity
```go
type FarmActivity struct {
    base.BaseModel
    CropCycleID  string         `json:"crop_cycle_id" gorm:"type:varchar(20);not null;index"`
    CropStageID  *string        `json:"crop_stage_id" gorm:"type:varchar(20);index"`  // NEW
    FarmerID     string         `json:"farmer_id" gorm:"type:varchar(20);not null;index"`
    ActivityType string         `json:"activity_type" gorm:"type:varchar(255);not null"`
    PlannedAt    *time.Time     `json:"planned_at" gorm:"type:timestamptz"`
    CompletedAt  *time.Time     `json:"completed_at" gorm:"type:timestamptz"`
    CreatedBy    string         `json:"created_by" gorm:"type:varchar(255);not null"`
    Status       string         `json:"status" gorm:"type:activity_status;not null;default:'PLANNED'"`
    Output       entities.JSONB `json:"output" gorm:"type:jsonb;default:'{}';serializer:json"`
    Metadata     entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`

    // Relationships
    CropCycle *crop_cycle.CropCycle `json:"crop_cycle,omitempty" gorm:"foreignKey:CropCycleID"`
    CropStage *stage.CropStage      `json:"crop_stage,omitempty" gorm:"foreignKey:CropStageID"`  // NEW
    Farmer    *farmer.Farmer        `json:"farmer,omitempty" gorm:"foreignKey:FarmerID"`
}
```

### 3. Repository Layer Enhancements

#### CropCycleRepository Extensions
```go
type CropCycleRepository interface {
    // Existing methods...

    // New area-related methods
    GetTotalAllocatedArea(ctx context.Context, farmID string, excludeCycleID string) (float64, error)
    GetAreaAllocationSummary(ctx context.Context, farmID string) (*AreaAllocationSummary, error)
    ValidateAreaAllocation(ctx context.Context, farmID string, cycleID string, requestedArea float64) error
    UpdateAreaAllocation(ctx context.Context, tx *gorm.DB, farmID string, delta float64) error
}

// Implementation with pessimistic locking
func (r *CropCycleRepositoryImpl) ValidateAreaAllocation(ctx context.Context, farmID string, cycleID string, requestedArea float64) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Lock farm record for update
        var farm farm.Farm
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            Where("id = ?", farmID).
            First(&farm).Error; err != nil {
            return err
        }

        // Calculate current allocation
        var totalAllocated float64
        if err := tx.Model(&CropCycle{}).
            Where("farm_id = ? AND status IN (?) AND id != ?",
                farmID, []string{"PLANNED", "ACTIVE"}, cycleID).
            Select("COALESCE(SUM(area_ha), 0)").
            Scan(&totalAllocated).Error; err != nil {
            return err
        }

        // Validate
        if totalAllocated + requestedArea > farm.AreaHa {
            return ErrExceedsFarmArea{
                FarmID: farmID,
                FarmArea: farm.AreaHa,
                Requested: requestedArea,
                Available: farm.AreaHa - totalAllocated,
            }
        }

        return nil
    }, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
}
```

#### FarmActivityRepository Extensions
```go
type FarmActivityRepository interface {
    // Existing methods...

    // New stage-related methods
    GetActivitiesByStage(ctx context.Context, cycleID string, stageID string, pagination *Pagination) ([]*FarmActivity, int64, error)
    GetStageCompletionStats(ctx context.Context, cycleID string) ([]*StageCompletionStat, error)
    GetCurrentStage(ctx context.Context, cycleID string) (*stage.CropStage, error)
}
```

### 4. Service Layer Implementation

#### CropCycleService Enhancements
```go
type CropCycleService interface {
    // Existing methods...

    // New area-related methods
    CreateCropCycleWithArea(ctx context.Context, req *CreateCropCycleRequest) (*CropCycleResponse, error)
    UpdateCropCycleArea(ctx context.Context, cycleID string, req *UpdateAreaRequest) (*CropCycleResponse, error)
    GetAreaAllocationSummary(ctx context.Context, farmID string) (*AreaAllocationSummaryResponse, error)
    ValidateAreaAvailability(ctx context.Context, farmID string, requestedArea float64) (*AreaAvailabilityResponse, error)
}

// Implementation with distributed locking for critical sections
func (s *CropCycleServiceImpl) CreateCropCycleWithArea(ctx context.Context, req *CreateCropCycleRequest) (*CropCycleResponse, error) {
    // Acquire distributed lock for farm
    lockKey := fmt.Sprintf("farm_area_lock:%s", req.FarmID)
    lock, err := s.distributedLock.Acquire(ctx, lockKey, 10*time.Second)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer lock.Release()

    // Validate permissions
    if !s.aaaService.HasPermission(ctx, req.UserID, "crop_cycle", "create", req.FarmID) {
        return nil, ErrPermissionDenied
    }

    // Start transaction
    tx := s.db.Begin()
    defer tx.Rollback()

    // Validate area allocation
    if err := s.repo.ValidateAreaAllocation(ctx, req.FarmID, "", req.AreaHa); err != nil {
        return nil, err
    }

    // Create crop cycle
    cycle := &CropCycle{
        FarmID:   req.FarmID,
        FarmerID: req.FarmerID,
        AreaHa:   req.AreaHa,
        Season:   req.Season,
        Status:   "PLANNED",
        CropID:   req.CropID,
    }

    if err := tx.Create(cycle).Error; err != nil {
        return nil, err
    }

    // Update area allocation tracking
    if err := s.updateAreaTracking(tx, req.FarmID, req.AreaHa); err != nil {
        return nil, err
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, err
    }

    // Audit log
    s.auditLogger.Log(ctx, AuditEntry{
        UserID:     req.UserID,
        Action:     "CREATE_CROP_CYCLE_WITH_AREA",
        ResourceID: cycle.ID,
        Details:    map[string]interface{}{"area_ha": req.AreaHa},
    })

    return s.toCropCycleResponse(cycle), nil
}
```

### 5. API Design

#### New/Modified Endpoints

```yaml
# Create crop cycle with area
POST /api/v1/crop-cycles
Request:
{
  "farm_id": "FARM_xxxxx",
  "farmer_id": "FRMR_xxxxx",
  "crop_id": "CROP_xxxxx",
  "variety_id": "VARTY_xxxxx",
  "area_ha": 5.5,  # REQUIRED
  "season": "KHARIF",
  "start_date": "2024-06-01",
  "metadata": {}
}

# Update crop cycle area
PATCH /api/v1/crop-cycles/{id}/area
Request:
{
  "area_ha": 6.0
}

# Get area allocation summary
GET /api/v1/farms/{farm_id}/area-allocation

Response:
{
  "farm_id": "FARM_xxxxx",
  "total_area_ha": 10.0,
  "allocated_area_ha": 8.5,
  "available_area_ha": 1.5,
  "allocations": [...],
  "last_updated": "2024-01-14T10:00:00Z"
}

# Create farm activity with stage
POST /api/v1/farm-activities
Request:
{
  "crop_cycle_id": "CRCY_xxxxx",
  "crop_stage_id": "CSTG_xxxxx",  # OPTIONAL
  "activity_type": "IRRIGATION",
  "planned_at": "2024-07-15T10:00:00Z",
  "metadata": {}
}

# Get activities by stage
GET /api/v1/crop-cycles/{id}/activities?stage_id=CSTG_xxxxx

# Get stage completion stats
GET /api/v1/crop-cycles/{id}/stage-progress
```

### 6. Error Handling Strategy

```go
// Custom error types
type ErrExceedsFarmArea struct {
    FarmID    string
    FarmArea  float64
    Requested float64
    Available float64
}

func (e ErrExceedsFarmArea) Error() string {
    return fmt.Sprintf("requested area %.2f ha exceeds available area %.2f ha for farm %s",
        e.Requested, e.Available, e.FarmID)
}

type ErrInvalidAreaValue struct {
    Value  float64
    Reason string
}

type ErrConcurrentModification struct {
    ResourceID string
    Operation  string
}

// Error handler middleware
func HandleAreaAllocationErrors(err error) (int, interface{}) {
    switch e := err.(type) {
    case ErrExceedsFarmArea:
        return 409, ErrorResponse{
            Code:    "AREA_EXCEEDED",
            Message: e.Error(),
            Details: map[string]interface{}{
                "available_area": e.Available,
                "requested_area": e.Requested,
            },
        }
    case ErrInvalidAreaValue:
        return 400, ErrorResponse{
            Code:    "INVALID_AREA",
            Message: e.Error(),
        }
    case ErrConcurrentModification:
        return 409, ErrorResponse{
            Code:    "CONCURRENT_MODIFICATION",
            Message: "Resource was modified by another request",
        }
    default:
        return 500, ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "An unexpected error occurred",
        }
    }
}
```

### 7. Concurrency Control Strategy

#### Pessimistic Locking
- Used for critical area allocation operations
- Database-level row locks with SELECT FOR UPDATE
- Distributed locks for farm-level operations

#### Optimistic Locking
- Version field in farm_area_allocations table
- Check version on update, retry on conflict

#### Transaction Isolation
- SERIALIZABLE isolation for area allocation operations
- READ COMMITTED for read-only operations

### 8. Performance Optimization

#### Caching Strategy
```go
// Cache area allocation summaries
type AreaAllocationCache interface {
    Get(farmID string) (*AreaAllocationSummary, bool)
    Set(farmID string, summary *AreaAllocationSummary, ttl time.Duration)
    Invalidate(farmID string)
}

// Redis implementation
type RedisAreaCache struct {
    client *redis.Client
    ttl    time.Duration
}

func (c *RedisAreaCache) Get(farmID string) (*AreaAllocationSummary, bool) {
    key := fmt.Sprintf("area_allocation:%s", farmID)
    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, false
    }

    var summary AreaAllocationSummary
    if err := json.Unmarshal(data, &summary); err != nil {
        return nil, false
    }

    return &summary, true
}
```

#### Query Optimization
```sql
-- Materialized view for area summaries (PostgreSQL)
CREATE MATERIALIZED VIEW farm_area_summary AS
SELECT
    f.id as farm_id,
    f.area_ha_computed as total_area,
    COALESCE(SUM(cc.area_ha) FILTER (WHERE cc.status IN ('PLANNED', 'ACTIVE')), 0) as allocated_area,
    f.area_ha_computed - COALESCE(SUM(cc.area_ha) FILTER (WHERE cc.status IN ('PLANNED', 'ACTIVE')), 0) as available_area,
    COUNT(cc.id) FILTER (WHERE cc.status IN ('PLANNED', 'ACTIVE')) as active_cycles_count
FROM farms f
LEFT JOIN crop_cycles cc ON f.id = cc.farm_id AND cc.deleted_at IS NULL
WHERE f.deleted_at IS NULL
GROUP BY f.id, f.area_ha_computed;

CREATE UNIQUE INDEX ON farm_area_summary(farm_id);

-- Refresh strategy
REFRESH MATERIALIZED VIEW CONCURRENTLY farm_area_summary;
```

### 9. Migration Strategy

#### Phase 1: Schema Migration
```sql
-- Add area_ha with nullable initially
ALTER TABLE crop_cycles ADD COLUMN area_ha DECIMAL(12,4);

-- Add crop_stage_id with nullable
ALTER TABLE farm_activities ADD COLUMN crop_stage_id VARCHAR(20);

-- Add foreign key after data migration
ALTER TABLE farm_activities
ADD CONSTRAINT fk_farm_activities_crop_stage
FOREIGN KEY (crop_stage_id) REFERENCES crop_stages(id);
```

#### Phase 2: Data Migration
```go
// Batch migration for existing crop cycles
func MigrateExistingCropCycles(db *gorm.DB) error {
    // Set default area for existing cycles (proportional distribution)
    query := `
        WITH farm_stats AS (
            SELECT
                farm_id,
                area_ha_computed,
                COUNT(*) as cycle_count
            FROM farms f
            JOIN crop_cycles cc ON f.id = cc.farm_id
            WHERE cc.area_ha IS NULL
                AND cc.status IN ('PLANNED', 'ACTIVE')
            GROUP BY farm_id, area_ha_computed
        )
        UPDATE crop_cycles cc
        SET area_ha = fs.area_ha_computed / fs.cycle_count
        FROM farm_stats fs
        WHERE cc.farm_id = fs.farm_id
            AND cc.area_ha IS NULL
            AND cc.status IN ('PLANNED', 'ACTIVE')
    `
    return db.Exec(query).Error
}
```

#### Phase 3: Constraint Application
```sql
-- After migration, make area_ha required for new cycles
ALTER TABLE crop_cycles
ALTER COLUMN area_ha SET NOT NULL,
ADD CONSTRAINT chk_positive_area CHECK (area_ha > 0);
```

### 10. Testing Strategy

#### Unit Tests
```go
func TestAreaAllocationValidation(t *testing.T) {
    tests := []struct {
        name          string
        farmArea      float64
        existingAlloc float64
        requestedArea float64
        expectError   bool
    }{
        {"Valid allocation", 10.0, 5.0, 3.0, false},
        {"Exact allocation", 10.0, 5.0, 5.0, false},
        {"Over allocation", 10.0, 8.0, 3.0, true},
        {"Zero area", 10.0, 5.0, 0.0, true},
        {"Negative area", 10.0, 5.0, -1.0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Tests
```go
func TestConcurrentAreaAllocation(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    farm := createTestFarm(db, 10.0)

    // Concurrent allocation attempts
    var wg sync.WaitGroup
    errors := make(chan error, 3)

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(area float64) {
            defer wg.Done()
            err := allocateCropCycle(db, farm.ID, area)
            if err != nil {
                errors <- err
            }
        }(4.0) // Each tries to allocate 4 ha
    }

    wg.Wait()
    close(errors)

    // Verify only 2 succeeded (8 ha total)
    successCount := 3 - len(errors)
    assert.Equal(t, 2, successCount)
}
```

#### Load Tests
```go
func BenchmarkAreaValidation(b *testing.B) {
    db := setupBenchDB(b)
    farm := createLargeFarm(db, 100.0, 50) // 100 ha with 50 cycles

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = validateAreaAllocation(db, farm.ID, 1.0)
    }
}
```

## Consequences

### Positive
1. **Data Integrity**: Guaranteed prevention of area over-allocation
2. **Concurrency Safety**: Proper handling of simultaneous requests
3. **Performance**: Optimized queries with caching and indexes
4. **Auditability**: Complete audit trail of all changes
5. **Flexibility**: Support for complex farming scenarios
6. **Analytics**: Enable data-driven insights on land utilization

### Negative
1. **Complexity**: More complex than simple CRUD operations
2. **Migration Effort**: Requires careful data migration
3. **Lock Contention**: Potential for lock wait under high load
4. **Cache Invalidation**: Complexity in maintaining cache consistency

### Risks and Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Race conditions | High | Medium | Pessimistic locking + distributed locks |
| Performance degradation | Medium | Low | Caching + query optimization |
| Migration failure | High | Low | Phased rollout with rollback plan |
| Cache inconsistency | Medium | Medium | TTL + event-based invalidation |

## Monitoring and Observability

### Key Metrics
```go
// Prometheus metrics
var (
    areaValidationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "crop_cycle_area_validation_duration_seconds",
            Help: "Duration of area validation operations",
        },
        []string{"farm_id", "status"},
    )

    areaAllocationErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "crop_cycle_area_allocation_errors_total",
            Help: "Total number of area allocation errors",
        },
        []string{"error_type"},
    )

    concurrentAllocationConflicts = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "crop_cycle_concurrent_allocation_conflicts_total",
            Help: "Total number of concurrent allocation conflicts",
        },
    )
)
```

### Logging Strategy
```go
// Structured logging for area operations
logger.Info("area allocation attempted",
    zap.String("farm_id", farmID),
    zap.Float64("requested_area", area),
    zap.Float64("available_area", available),
    zap.String("user_id", userID),
    zap.String("trace_id", traceID),
)
```

### Alerting Rules
```yaml
- alert: HighAreaValidationLatency
  expr: histogram_quantile(0.95, rate(crop_cycle_area_validation_duration_seconds_bucket[5m])) > 0.5
  for: 5m
  annotations:
    summary: "Area validation taking too long"

- alert: FrequentAllocationConflicts
  expr: rate(crop_cycle_concurrent_allocation_conflicts_total[5m]) > 1
  for: 10m
  annotations:
    summary: "High rate of concurrent allocation conflicts"
```

## References

1. [PostgreSQL Locking Documentation](https://www.postgresql.org/docs/current/explicit-locking.html)
2. [GORM Transaction Guide](https://gorm.io/docs/transactions.html)
3. [Distributed Locking Patterns](https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html)
4. [Cache-Aside Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
5. [Event Sourcing for Audit Logs](https://martinfowler.com/eaaDev/EventSourcing.html)
