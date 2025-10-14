# Implementation Plan - Crop Cycles Area Allocation

## Overview

This document provides a detailed, step-by-step implementation plan for adding area allocation to crop cycles and organizing farm activities by crop stages. The implementation follows a phased approach to minimize risk and ensure backward compatibility.

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Database schema changes
- Domain model updates
- Basic validation logic

### Phase 2: Core Features (Week 2)
- Repository layer implementation
- Service layer with business logic
- Area validation and allocation

### Phase 3: API Layer (Week 3)
- REST API endpoints
- Request/response models
- Error handling

### Phase 4: Stage Integration (Week 4)
- Farm activity stage association
- Stage-based filtering
- Progress tracking

### Phase 5: Testing & Optimization (Week 5)
- Comprehensive testing
- Performance optimization
- Documentation

## Detailed Implementation Tasks

### Task 1: Database Schema Updates

#### Task 1.1: Add area_ha to crop_cycles
**Priority**: Critical
**Estimated Time**: 2 hours
**Dependencies**: None

```bash
# Create migration file
touch internal/db/migrations/20250114_add_area_to_crop_cycles.sql
```

**Implementation Steps:**
1. Create migration to add area_ha column as nullable
2. Add check constraint for positive values
3. Create performance indexes
4. Test migration on development database

**Validation Criteria:**
- Column added successfully
- Existing data not affected
- Indexes created and working

#### Task 1.2: Add crop_stage_id to farm_activities
**Priority**: High
**Estimated Time**: 2 hours
**Dependencies**: Task 1.1

```bash
# Create migration file
touch internal/db/migrations/20250114_add_stage_to_activities.sql
```

**Implementation Steps:**
1. Add crop_stage_id column as nullable
2. Add foreign key constraint to crop_stages
3. Create indexes for stage-based queries
4. Test foreign key relationships

#### Task 1.3: Create area allocation tracking table
**Priority**: Medium
**Estimated Time**: 3 hours
**Dependencies**: Task 1.1, 1.2

```bash
# Create migration file
touch internal/db/migrations/20250114_create_area_allocations.sql
```

**Implementation Steps:**
1. Create farm_area_allocations table
2. Add constraints and indexes
3. Create triggers for maintaining consistency
4. Test trigger functionality

### Task 2: Domain Model Enhancements

#### Task 2.1: Update CropCycle entity
**Priority**: Critical
**Estimated Time**: 2 hours
**Dependencies**: Task 1.1

```go
// File: internal/entities/crop_cycle/crop_cycle.go

// Add to CropCycle struct
AreaHa float64 `json:"area_ha" gorm:"type:decimal(12,4);check:area_ha > 0"`

// Add validation method
func (cc *CropCycle) ValidateArea() error {
    if cc.AreaHa <= 0 {
        return ErrInvalidArea
    }
    return nil
}
```

**Implementation Steps:**
1. Add AreaHa field to struct
2. Update Validate() method
3. Add helper methods for area calculations
4. Update unit tests

#### Task 2.2: Update FarmActivity entity
**Priority**: High
**Estimated Time**: 2 hours
**Dependencies**: Task 1.2

```go
// File: internal/entities/farm_activity/farm_activity.go

// Add to FarmActivity struct
CropStageID *string `json:"crop_stage_id" gorm:"type:varchar(20);index"`

// Add relationship
CropStage *stage.CropStage `json:"crop_stage,omitempty" gorm:"foreignKey:CropStageID"`
```

**Implementation Steps:**
1. Add CropStageID field
2. Define relationship with CropStage
3. Update validation logic
4. Test relationship loading

#### Task 2.3: Create AreaAllocation entity
**Priority**: Medium
**Estimated Time**: 3 hours
**Dependencies**: Task 2.1, 2.2

```go
// File: internal/entities/area_allocation/area_allocation.go

type FarmAreaAllocation struct {
    base.BaseModel
    FarmID            string  `json:"farm_id"`
    TotalAreaHa       float64 `json:"total_area_ha"`
    AllocatedAreaHa   float64 `json:"allocated_area_ha"`
    AvailableAreaHa   float64 `json:"available_area_ha"`
    Version           int     `json:"version"`
    LastValidatedAt   time.Time `json:"last_validated_at"`
}
```

### Task 3: Repository Layer Implementation

#### Task 3.1: Extend CropCycleRepository
**Priority**: Critical
**Estimated Time**: 4 hours
**Dependencies**: Task 2.1

```go
// File: internal/repo/crop_cycle/area_methods.go

func (r *CropCycleRepository) GetTotalAllocatedArea(
    ctx context.Context,
    farmID string,
    excludeCycleID string,
) (float64, error) {
    var total float64
    query := r.db.Model(&CropCycle{}).
        Where("farm_id = ? AND status IN (?)", farmID, []string{"PLANNED", "ACTIVE"})

    if excludeCycleID != "" {
        query = query.Where("id != ?", excludeCycleID)
    }

    err := query.Select("COALESCE(SUM(area_ha), 0)").Scan(&total).Error
    return total, err
}

func (r *CropCycleRepository) ValidateAreaAllocation(
    ctx context.Context,
    farmID string,
    cycleID string,
    requestedArea float64,
) error {
    // Implementation with transaction and locking
}
```

**Implementation Steps:**
1. Create new file for area-related methods
2. Implement GetTotalAllocatedArea method
3. Implement ValidateAreaAllocation with locking
4. Add GetAreaAllocationSummary method
5. Write comprehensive unit tests

#### Task 3.2: Extend FarmActivityRepository
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 2.2

```go
// File: internal/repo/farm_activity/stage_methods.go

func (r *FarmActivityRepository) GetActivitiesByStage(
    ctx context.Context,
    cycleID string,
    stageID string,
    pagination *Pagination,
) ([]*FarmActivity, int64, error) {
    // Implementation
}

func (r *FarmActivityRepository) GetStageCompletionStats(
    ctx context.Context,
    cycleID string,
) ([]*StageCompletionStat, error) {
    // Implementation
}
```

**Implementation Steps:**
1. Create stage-related methods file
2. Implement filtering by stage
3. Add stage completion statistics
4. Test with various scenarios

### Task 4: Service Layer Implementation

#### Task 4.1: Create AreaAllocationService
**Priority**: Critical
**Estimated Time**: 6 hours
**Dependencies**: Task 3.1

```go
// File: internal/services/area_allocation_service.go

type AreaAllocationService struct {
    db               *gorm.DB
    cropCycleRepo    *CropCycleRepository
    farmRepo         *FarmRepository
    distributedLock  DistributedLock
    cache            AreaAllocationCache
    logger           *zap.Logger
}

func (s *AreaAllocationService) AllocateArea(
    ctx context.Context,
    req *AllocateAreaRequest,
) (*AllocateAreaResponse, error) {
    // Acquire distributed lock
    // Validate permissions
    // Check area availability
    // Perform allocation
    // Update cache
    // Return response
}
```

**Implementation Steps:**
1. Create service structure
2. Implement AllocateArea with locking
3. Add ReleaseArea method
4. Implement UpdateAllocation method
5. Add validation helpers
6. Write integration tests

#### Task 4.2: Extend CropCycleService
**Priority**: Critical
**Estimated Time**: 4 hours
**Dependencies**: Task 4.1

```go
// File: internal/services/crop_cycle_service.go

func (s *CropCycleService) CreateCropCycleWithArea(
    ctx context.Context,
    req *CreateCropCycleRequest,
) (*CropCycleResponse, error) {
    // Validate area availability
    // Create cycle with area
    // Update area tracking
    // Audit log
}

func (s *CropCycleService) UpdateCropCycleArea(
    ctx context.Context,
    cycleID string,
    req *UpdateAreaRequest,
) (*CropCycleResponse, error) {
    // Check status allows update
    // Validate new area
    // Update with validation
    // Refresh cache
}
```

**Implementation Steps:**
1. Modify CreateCropCycle to include area
2. Add UpdateCropCycleArea method
3. Integrate with AreaAllocationService
4. Add proper error handling
5. Test concurrent operations

#### Task 4.3: Extend FarmActivityService
**Priority**: High
**Estimated Time**: 4 hours
**Dependencies**: Task 3.2

```go
// File: internal/services/farm_activity_service.go

func (s *FarmActivityService) CreateActivityWithStage(
    ctx context.Context,
    req *CreateActivityRequest,
) (*ActivityResponse, error) {
    // Validate stage belongs to crop
    // Create activity with stage
    // Return enriched response
}

func (s *FarmActivityService) GetStageProgress(
    ctx context.Context,
    cycleID string,
) (*StageProgressResponse, error) {
    // Get all stages for crop
    // Calculate completion per stage
    // Determine current stage
    // Return progress summary
}
```

### Task 5: API Handler Implementation

#### Task 5.1: Create AreaAllocationHandler
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 4.1

```go
// File: internal/handlers/area_allocation_handler.go

type AreaAllocationHandler struct {
    areaService      *AreaAllocationService
    cropCycleService *CropCycleService
    logger           *zap.Logger
}

// @Summary Get farm area allocation summary
// @Router /farms/{farm_id}/area-allocation [get]
func (h *AreaAllocationHandler) GetAreaAllocationSummary(c *gin.Context) {
    // Extract farm_id
    // Call service
    // Return response
}

// @Summary Check area availability
// @Router /farms/{farm_id}/check-area-availability [post]
func (h *AreaAllocationHandler) CheckAreaAvailability(c *gin.Context) {
    // Parse request
    // Validate input
    // Check availability
    // Return result
}
```

#### Task 5.2: Update CropCycleHandler
**Priority**: Critical
**Estimated Time**: 2 hours
**Dependencies**: Task 5.1

```go
// File: internal/handlers/crop_cycle_handler.go

// Modify existing CreateCropCycle
func (h *CropCycleHandler) CreateCropCycle(c *gin.Context) {
    // Include area_ha in request
    // Call updated service method
}

// Add new endpoint
func (h *CropCycleHandler) UpdateCropCycleArea(c *gin.Context) {
    // Parse cycle ID and area
    // Call service
    // Return response
}
```

#### Task 5.3: Update FarmActivityHandler
**Priority**: High
**Estimated Time**: 2 hours
**Dependencies**: Task 4.3

```go
// File: internal/handlers/farm_activity_handler.go

// Modify CreateActivity to include stage
func (h *FarmActivityHandler) CreateActivity(c *gin.Context) {
    // Include crop_stage_id in request
    // Validate stage if provided
}

// Add new endpoint
func (h *FarmActivityHandler) GetActivitiesByStage(c *gin.Context) {
    // Parse query parameters
    // Call service with filters
    // Return grouped response
}

// Add stage progress endpoint
func (h *FarmActivityHandler) GetStageProgress(c *gin.Context) {
    // Get cycle ID
    // Call service
    // Return progress data
}
```

### Task 6: Request/Response Models

#### Task 6.1: Create area allocation DTOs
**Priority**: High
**Estimated Time**: 2 hours
**Dependencies**: None

```go
// File: internal/entities/requests/area_allocation.go

type AllocateAreaRequest struct {
    FarmID        string  `json:"farm_id" binding:"required"`
    CropCycleID   string  `json:"crop_cycle_id" binding:"required"`
    AreaHa        float64 `json:"area_ha" binding:"required,gt=0"`
}

type CheckAvailabilityRequest struct {
    RequestedAreaHa float64 `json:"requested_area_ha" binding:"required,gt=0"`
    ExcludeCycleID  string  `json:"exclude_cycle_id,omitempty"`
}

// File: internal/entities/responses/area_allocation.go

type AreaAllocationSummaryResponse struct {
    FarmID               string                `json:"farm_id"`
    TotalAreaHa          float64              `json:"total_area_ha"`
    AllocatedAreaHa      float64              `json:"allocated_area_ha"`
    AvailableAreaHa      float64              `json:"available_area_ha"`
    UtilizationPercent   float64              `json:"utilization_percentage"`
    Allocations          []AllocationDetail   `json:"allocations"`
}
```

### Task 7: Caching Implementation

#### Task 7.1: Implement Redis cache for area allocations
**Priority**: Medium
**Estimated Time**: 3 hours
**Dependencies**: Task 4.1

```go
// File: internal/cache/area_allocation_cache.go

type RedisAreaCache struct {
    client *redis.Client
    ttl    time.Duration
}

func (c *RedisAreaCache) Get(farmID string) (*AreaAllocationSummary, bool) {
    key := fmt.Sprintf("area:farm:%s", farmID)
    // Implementation
}

func (c *RedisAreaCache) Set(farmID string, summary *AreaAllocationSummary) {
    // Implementation with TTL
}

func (c *RedisAreaCache) Invalidate(farmID string) {
    // Implementation
}
```

### Task 8: Error Handling

#### Task 8.1: Create custom error types
**Priority**: High
**Estimated Time**: 1 hour
**Dependencies**: None

```go
// File: pkg/errors/area_errors.go

type ErrExceedsFarmArea struct {
    FarmID    string
    FarmArea  float64
    Requested float64
    Available float64
}

func (e ErrExceedsFarmArea) Error() string {
    return fmt.Sprintf("requested area %.2f ha exceeds available %.2f ha for farm %s",
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
```

### Task 9: Testing Implementation

#### Task 9.1: Unit tests for domain models
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 2 completed

```go
// File: internal/entities/crop_cycle/crop_cycle_test.go

func TestCropCycleAreaValidation(t *testing.T) {
    tests := []struct {
        name    string
        area    float64
        wantErr bool
    }{
        {"Valid area", 5.5, false},
        {"Zero area", 0, true},
        {"Negative area", -1, true},
    }
    // Implementation
}
```

#### Task 9.2: Integration tests for area allocation
**Priority**: Critical
**Estimated Time**: 4 hours
**Dependencies**: Task 4 completed

```go
// File: internal/services/area_allocation_service_test.go

func TestConcurrentAreaAllocation(t *testing.T) {
    // Setup test database
    // Create farm with 10 ha
    // Attempt 3 concurrent allocations of 4 ha each
    // Verify only 2 succeed
}
```

#### Task 9.3: API endpoint tests
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 5 completed

```go
// File: internal/handlers/area_allocation_handler_test.go

func TestGetAreaAllocationSummary(t *testing.T) {
    // Setup test router
    // Create test data
    // Call endpoint
    // Verify response
}
```

### Task 10: Database Migration

#### Task 10.1: Create migration scripts
**Priority**: Critical
**Estimated Time**: 2 hours
**Dependencies**: Task 1 completed

```bash
# Create migration files
mkdir -p internal/db/migrations
touch internal/db/migrations/20250114_area_allocation_phase1.sql
touch internal/db/migrations/20250115_area_allocation_phase2.sql
touch internal/db/migrations/20250116_area_allocation_phase3.sql
```

#### Task 10.2: Test migrations
**Priority**: Critical
**Estimated Time**: 2 hours
**Dependencies**: Task 10.1

**Testing Steps:**
1. Backup development database
2. Run migrations
3. Verify schema changes
4. Test rollback
5. Document any issues

#### Task 10.3: Create data migration script
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 10.2

```go
// File: cmd/migrate/area_allocation.go

func MigrateExistingCropCycles(db *gorm.DB) error {
    // Get farms with active cycles
    // Calculate proportional area allocation
    // Update cycles with calculated area
    // Verify totals
}
```

### Task 11: Performance Optimization

#### Task 11.1: Create materialized views
**Priority**: Medium
**Estimated Time**: 2 hours
**Dependencies**: Task 10 completed

```sql
-- File: internal/db/views/area_summary.sql

CREATE MATERIALIZED VIEW mv_farm_area_summary AS
-- Implementation

-- Create refresh job
CREATE OR REPLACE FUNCTION refresh_area_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_farm_area_summary;
END;
$$ LANGUAGE plpgsql;
```

#### Task 11.2: Implement query optimization
**Priority**: Medium
**Estimated Time**: 3 hours
**Dependencies**: Task 11.1

**Optimization Tasks:**
1. Analyze slow queries with EXPLAIN
2. Add appropriate indexes
3. Optimize N+1 queries
4. Implement query result caching

### Task 12: Documentation

#### Task 12.1: API documentation
**Priority**: High
**Estimated Time**: 3 hours
**Dependencies**: Task 5 completed

```go
// Add Swagger annotations to all new endpoints
// Update Swagger definitions
// Generate updated documentation
make docs
```

#### Task 12.2: Technical documentation
**Priority**: Medium
**Estimated Time**: 2 hours
**Dependencies**: All tasks completed

**Documentation Tasks:**
1. Update README with new features
2. Create migration guide
3. Document configuration options
4. Add troubleshooting guide

### Task 13: Monitoring and Observability

#### Task 13.1: Add metrics
**Priority**: Medium
**Estimated Time**: 2 hours
**Dependencies**: Task 4 completed

```go
// File: internal/metrics/area_metrics.go

var (
    areaValidationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "crop_cycle_area_validation_duration_seconds",
            Help: "Duration of area validation operations",
        },
        []string{"farm_id", "status"},
    )
)
```

#### Task 13.2: Add structured logging
**Priority**: Medium
**Estimated Time**: 1 hour
**Dependencies**: Task 13.1

```go
// Add logging to critical operations
logger.Info("area allocation attempted",
    zap.String("farm_id", farmID),
    zap.Float64("requested_area", area),
    zap.String("trace_id", traceID),
)
```

## Testing Strategy

### Unit Test Coverage Goals
- Domain models: 95%
- Repository layer: 90%
- Service layer: 85%
- Handlers: 80%

### Integration Test Scenarios
1. Single cycle allocation
2. Multiple cycles on same farm
3. Concurrent allocation attempts
4. Area modification scenarios
5. Edge cases (exact allocation, zero availability)

### Performance Test Scenarios
1. Load test with 1000 concurrent requests
2. Stress test area validation
3. Cache performance verification
4. Database query optimization validation

## Rollout Plan

### Development Environment
1. Deploy schema changes
2. Deploy application code
3. Run smoke tests
4. Monitor for 24 hours

### Staging Environment
1. Deploy with feature flag disabled
2. Run migration scripts
3. Enable feature for test accounts
4. Full regression testing
5. Performance testing
6. Enable for all staging

### Production Rollout
1. Deploy schema changes during maintenance window
2. Deploy application with feature flag
3. Enable for 10% of users
4. Monitor metrics and errors
5. Gradual rollout to 100%
6. Remove feature flag after stability confirmed

## Risk Mitigation

### Risk 1: Data Migration Failure
**Mitigation:**
- Backup before migration
- Test on staging first
- Have rollback scripts ready
- Run in transaction where possible

### Risk 2: Performance Degradation
**Mitigation:**
- Load test before deployment
- Monitor query performance
- Have caching in place
- Database indexes optimized

### Risk 3: Concurrent Modification Issues
**Mitigation:**
- Implement proper locking
- Use transactions appropriately
- Add retry logic
- Monitor for conflicts

## Success Criteria

### Functional Success
- [ ] Area allocation prevents over-allocation
- [ ] Activities can be filtered by stage
- [ ] Stage progress calculation accurate
- [ ] Backward compatibility maintained

### Performance Success
- [ ] Area validation < 100ms p95
- [ ] No degradation in existing APIs
- [ ] Cache hit rate > 80%
- [ ] Database CPU < 70% under load

### Quality Success
- [ ] Test coverage meets targets
- [ ] No critical bugs in production
- [ ] Error rate < 0.1%
- [ ] All documentation updated

## Timeline

### Week 1 (Foundation)
- Day 1-2: Database schema and migrations
- Day 3-4: Domain model updates
- Day 5: Initial testing and validation

### Week 2 (Core Features)
- Day 1-2: Repository layer
- Day 3-4: Service layer
- Day 5: Integration testing

### Week 3 (API Layer)
- Day 1-2: Handlers and routes
- Day 3: Request/response models
- Day 4-5: API testing

### Week 4 (Stage Integration)
- Day 1-2: Activity stage features
- Day 3-4: Progress tracking
- Day 5: End-to-end testing

### Week 5 (Polish)
- Day 1-2: Performance optimization
- Day 3: Documentation
- Day 4: Staging deployment
- Day 5: Production preparation

## Dependencies

### External Dependencies
- AAA service for authorization
- Redis for caching
- PostgreSQL with PostGIS
- Distributed lock service

### Internal Dependencies
- Stages management feature completed
- Farm and Farmer modules stable
- Crop and Variety data available

## Team Responsibilities

### Backend Engineer
- Implementation of all code changes
- Unit and integration testing
- Performance optimization
- Technical documentation

### Database Administrator
- Review and approve schema changes
- Assist with migration scripts
- Performance tuning
- Backup and recovery planning

### DevOps Engineer
- Setup caching infrastructure
- Configure monitoring
- Deployment automation
- Rollback procedures

### QA Engineer
- Test scenario creation
- Manual testing
- Performance testing
- Bug verification

### Product Manager
- Requirement clarification
- Acceptance criteria validation
- Stakeholder communication
- Rollout coordination

## Appendix

### A. Configuration Parameters

```yaml
# Application configuration
area_allocation:
  enable_caching: true
  cache_ttl: 300s
  lock_timeout: 10s
  max_retry_attempts: 3
  validation_mode: strict

# Feature flags
features:
  area_allocation: true
  stage_activities: true
  progress_tracking: true
```

### B. Monitoring Dashboards

1. Area Allocation Dashboard
   - Total allocations per hour
   - Validation success/failure rate
   - Average validation time
   - Cache hit rate

2. Stage Progress Dashboard
   - Activities per stage
   - Completion percentages
   - Stage duration analysis
   - Delayed activities

### C. Troubleshooting Guide

**Problem**: Area allocation fails with "concurrent modification"
**Solution**: Retry the operation or increase lock timeout

**Problem**: Cache inconsistency
**Solution**: Clear cache for affected farm, check TTL settings

**Problem**: Slow area validation
**Solution**: Check indexes, analyze query plan, consider materialized view

## Conclusion

This implementation plan provides a comprehensive roadmap for adding area allocation to crop cycles and organizing activities by stages. Following this plan will ensure a robust, scalable, and maintainable solution that meets all business requirements while maintaining system integrity and performance.
