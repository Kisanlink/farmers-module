# Business Logic Tests for Stages Management Feature

## Executive Summary

This document provides comprehensive test scenarios for the stages management feature, including:
- Business logic validation and edge cases
- Security and abuse path testing
- Concurrency and race condition scenarios
- Data integrity and invariant validation
- Permission and authorization testing
- Real-world agricultural scenarios

## Business Rules Identified

### Stage Entity Rules
1. **Stage Name Uniqueness**: Stage names must be unique (case-insensitive)
2. **Stage Name Length**: Must be 1-100 characters
3. **Soft Delete Support**: Deleted stages remain in DB with deleted_at timestamp
4. **Active Status**: Stages have is_active flag for enabling/disabling
5. **JSONB Properties**: Flexible metadata storage for stage-specific data

### CropStage Relationship Rules
1. **Unique Combination**: Each crop can have each stage only once
2. **Stage Order Validation**: Must be >= 1
3. **Unique Order per Crop**: Stage orders must be unique within a crop
4. **Duration Validation**: If specified, duration must be positive (> 0)
5. **Duration Unit Validation**: Must be DAYS, WEEKS, or MONTHS
6. **Foreign Key Constraints**: Must reference valid crop_id and stage_id

### Service Layer Rules
1. **Permission Checks**: All operations require AAA permission validation
2. **Cascade Soft Delete**: Soft deletes preserve referential integrity
3. **Transactional Reordering**: Stage reordering happens in transaction
4. **Lookup Filtering**: Only active stages shown in lookup endpoint

## Business Invariants

### Critical Invariants That Must Always Hold
1. **I1: Stage Name Uniqueness**
   - No two active stages can have the same name (case-insensitive)
   - Soft-deleted stages don't affect uniqueness constraint

2. **I2: Crop-Stage Uniqueness**
   - A crop cannot have the same stage assigned twice
   - Even if soft-deleted, relationship remains unique

3. **I3: Stage Order Continuity**
   - Stage orders for a crop must be positive integers starting from 1
   - No duplicate orders allowed for active stages

4. **I4: Referential Integrity**
   - All crop_stages must reference existing crops and stages
   - Soft-deleted stages can still be referenced by existing crop_stages

5. **I5: Duration Consistency**
   - If duration_days is specified, it must be positive
   - Duration_unit must be valid when duration is specified

## Test Scenarios

### 1. Stage Management Tests

#### 1.1 Stage Creation
```yaml
Test: Create Stage - Happy Path
Given: Valid stage data with unique name
When: CreateStage is called with proper permissions
Then: Stage is created with unique ID and defaults

Test: Create Stage - Duplicate Name (Case Insensitive)
Given: Stage "Germination" exists
When: CreateStage with "germination" or "GERMINATION"
Then: Returns ErrAlreadyExists

Test: Create Stage - Empty Name
Given: Stage request with empty name
When: CreateStage is called
Then: Returns ErrInvalidInput

Test: Create Stage - Name Too Long
Given: Stage name > 100 characters
When: CreateStage is called
Then: Returns ErrInvalidInput

Test: Create Stage - Invalid JSONB Properties
Given: Malformed JSON in properties field
When: CreateStage is called
Then: Returns parsing error

Test: Create Stage - No Permission
Given: User lacks stage:create permission
When: CreateStage is called
Then: Returns ErrForbidden
```

#### 1.2 Stage Updates
```yaml
Test: Update Stage - Name to Existing
Given: Stages "Flowering" and "Vegetative" exist
When: Update "Flowering" name to "vegetative"
Then: Returns ErrAlreadyExists

Test: Update Stage - Self Name Update
Given: Stage "Germination" exists
When: Update "Germination" to "Germination" (same name)
Then: Update succeeds (no-op)

Test: Update Stage - Inactive Stage References
Given: Stage is set to inactive with active crop assignments
When: UpdateStage with is_active=false
Then: Stage becomes inactive, crop_stages remain but may need validation

Test: Update Stage - Deleted Stage
Given: Stage is soft-deleted
When: UpdateStage is attempted
Then: Returns ErrNotFound
```

#### 1.3 Stage Deletion
```yaml
Test: Delete Stage - With Active Crop Assignments
Given: Stage assigned to multiple crops
When: DeleteStage is called
Then: Stage is soft-deleted, assignments remain but orphaned

Test: Delete Stage - Already Deleted
Given: Stage is already soft-deleted
When: DeleteStage is called again
Then: Returns ErrNotFound

Test: Delete Stage - Cascade Impact
Given: Stage with crop assignments and properties
When: DeleteStage is called
Then: Verify crop_stages still exist but reference deleted stage
```

### 2. Crop-Stage Relationship Tests

#### 2.1 Stage Assignment
```yaml
Test: Assign Stage - Happy Path
Given: Valid crop and stage IDs
When: AssignStageToCrop with order=1
Then: CropStage created with proper relationships

Test: Assign Stage - Duplicate Assignment
Given: Crop already has Stage X assigned
When: AssignStageToCrop with same Stage X
Then: Returns ErrAlreadyExists

Test: Assign Stage - Duplicate Order
Given: Crop has stage at order=2
When: AssignStageToCrop with different stage at order=2
Then: Returns ErrAlreadyExists with order conflict message

Test: Assign Stage - Invalid Stage ID
Given: Non-existent stage ID
When: AssignStageToCrop is called
Then: Returns ErrNotFound for stage

Test: Assign Stage - Invalid Crop ID
Given: Non-existent crop ID
When: AssignStageToCrop is called
Then: No validation, fails on FK constraint (MISSING VALIDATION)

Test: Assign Stage - Deleted Stage
Given: Soft-deleted stage
When: AssignStageToCrop is called
Then: Currently succeeds (POTENTIAL BUG - should validate is_active)

Test: Assign Stage - Zero Order
Given: Stage order = 0
When: AssignStageToCrop is called
Then: Returns ErrInvalidInput

Test: Assign Stage - Negative Order
Given: Stage order = -1
When: AssignStageToCrop is called
Then: Returns ErrInvalidInput

Test: Assign Stage - Order Gap
Given: Crop has stages at order 1,2,5
When: AssignStageToCrop with order=3
Then: Succeeds, creating gap (may need validation)

Test: Assign Stage - Invalid Duration
Given: duration_days = 0 or negative
When: AssignStageToCrop is called
Then: Returns ErrInvalidInput

Test: Assign Stage - Invalid Duration Unit
Given: duration_unit = "HOURS" or invalid
When: AssignStageToCrop is called
Then: Returns ErrInvalidInput

Test: Assign Stage - Missing Duration Unit
Given: duration_days specified but no unit
When: AssignStageToCrop is called
Then: Defaults to DAYS
```

#### 2.2 Crop Stage Updates
```yaml
Test: Update CropStage - Order Conflict
Given: Crop has stages at order 1,2,3
When: UpdateCropStage order from 3 to 2
Then: Returns ErrAlreadyExists for order conflict

Test: Update CropStage - Self Order Update
Given: Stage at order=2
When: UpdateCropStage to same order=2
Then: Succeeds (no-op)

Test: Update CropStage - Duration to Zero
Given: Existing duration_days=14
When: UpdateCropStage with duration_days=0
Then: Returns ErrInvalidInput

Test: Update CropStage - Change Duration Unit
Given: duration=14 DAYS
When: UpdateCropStage to WEEKS
Then: Succeeds, duration semantic changes

Test: Update CropStage - Deactivate
Given: Active crop stage
When: UpdateCropStage with is_active=false
Then: Stage remains but inactive
```

#### 2.3 Stage Removal
```yaml
Test: Remove Stage - Middle of Sequence
Given: Crop has stages at order 1,2,3,4
When: RemoveStageFromCrop for order=2
Then: Stage removed, orders 3,4 remain (creating gap)

Test: Remove Stage - Non-existent Assignment
Given: Stage not assigned to crop
When: RemoveStageFromCrop is called
Then: Returns ErrNotFound

Test: Remove Stage - Already Removed
Given: Soft-deleted crop_stage
When: RemoveStageFromCrop is called again
Then: Returns ErrNotFound
```

### 3. Reordering Tests

#### 3.1 Basic Reordering
```yaml
Test: Reorder - Valid Sequence
Given: Stages at order 1,2,3
When: ReorderCropStages to 2,3,1
Then: Orders updated atomically

Test: Reorder - Missing Stages
Given: Stages A,B,C assigned
When: ReorderCropStages with only A,B
Then: Returns error for missing stage C

Test: Reorder - Extra Stages
Given: Stages A,B assigned
When: ReorderCropStages with A,B,C (C not assigned)
Then: Returns ErrNotFound for stage C

Test: Reorder - Duplicate Orders
Given: Any stages
When: ReorderCropStages with duplicate order values
Then: Last one wins in map (POTENTIAL BUG)

Test: Reorder - Empty Map
Given: Crop with stages
When: ReorderCropStages with empty map
Then: Returns ErrInvalidInput

Test: Reorder - Concurrent Operations
Given: Two users reordering same crop
When: Both call ReorderCropStages simultaneously
Then: Transaction ensures consistency
```

### 4. Edge Cases and Boundary Tests

#### 4.1 Data Validation Edge Cases
```yaml
Test: Stage Name - Unicode Characters
Given: Stage name with emojis, RTL text, special chars
When: CreateStage is called
Then: Succeeds if within 100 chars

Test: Stage Name - SQL Injection
Given: Stage name = "'; DROP TABLE stages; --"
When: CreateStage is called
Then: Properly escaped, no injection

Test: JSONB Properties - Large Objects
Given: Properties with 1MB of JSON data
When: Create/UpdateStage is called
Then: Check performance and storage limits

Test: JSONB Properties - Nested Depth
Given: Deeply nested JSON (100+ levels)
When: Create/UpdateStage is called
Then: Check parsing limits and performance

Test: Pagination - Boundary Cases
Given: page=0, pageSize=0, pageSize>100
When: ListStages is called
Then: Defaults applied correctly

Test: Search - Special Characters
Given: Search term with %, _, wildcards
When: ListStages with search
Then: Properly escaped in LIKE clause
```

#### 4.2 Concurrency Tests
```yaml
Test: Concurrent Stage Creation - Same Name
Given: Two requests to create "Flowering"
When: Both execute simultaneously
Then: One succeeds, one gets ErrAlreadyExists

Test: Concurrent Order Assignment
Given: Two stages being assigned order=3
When: Both AssignStageToCrop simultaneously
Then: One succeeds, one gets order conflict

Test: Concurrent Reordering
Given: Multiple reorder operations
When: Execute in parallel
Then: Transaction ensures consistency

Test: Read-Write Conflicts
Given: Reordering while reading stages
When: GetCropStages during ReorderCropStages
Then: Read sees consistent state
```

### 5. Security and Abuse Path Tests

#### 5.1 Permission Bypass Attempts
```yaml
Test: Direct DB Manipulation
Risk: Bypassing service layer validations
Mitigation: DB constraints enforce invariants

Test: Permission Elevation
Risk: User modifying own permissions
Mitigation: AAA service validates all ops

Test: Cross-Organization Access
Risk: Accessing stages from other orgs
Mitigation: org_id filtering in all queries
```

#### 5.2 Resource Exhaustion
```yaml
Test: Too Many Stages per Crop
Given: Attempt to assign 1000+ stages
When: Repeated AssignStageToCrop calls
Then: Check performance degradation

Test: Large JSONB Properties
Given: 10MB JSON in properties
When: Create/Update operations
Then: Check memory usage and limits

Test: Pagination Abuse
Given: pageSize=10000 requests
When: ListStages is called
Then: Capped at 100 max
```

#### 5.3 Data Integrity Attacks
```yaml
Test: Foreign Key Manipulation
Given: Fake crop_id or stage_id
When: AssignStageToCrop
Then: FK constraint prevents invalid refs

Test: Soft Delete Exploitation
Given: Deleted stage references
When: Attempting to use deleted stages
Then: Service layer should validate

Test: Order Manipulation
Given: Order = MAX_INT or MIN_INT
When: AssignStageToCrop
Then: Check integer overflow handling
```

### 6. Real-World Agricultural Scenarios

#### 6.1 Common Crop Lifecycles
```yaml
Test: Rice Cultivation Stages
Stages: Seed Selection → Nursery → Land Prep → Transplanting →
        Vegetative → Reproductive → Ripening → Harvesting
Duration: 120-150 days total
Validation: Ensure order and duration tracking

Test: Wheat Growth Phases
Stages: Germination → Tillering → Stem Extension → Booting →
        Heading → Flowering → Grain Filling → Maturity
Duration: 90-120 days
Validation: Stage transitions and properties

Test: Tomato Production
Stages: Seeding → Transplanting → Vegetative → Flowering →
        Fruit Setting → Ripening → Harvesting
Duration: 60-85 days
Validation: Multiple harvest cycles support
```

#### 6.2 Stage Property Scenarios
```yaml
Test: Weather-Dependent Properties
Properties: {
  "optimal_temperature": "25-30°C",
  "water_requirement": "high",
  "fertilizer_schedule": "weekly"
}
Validation: JSONB flexibility for metadata

Test: Region-Specific Variations
Properties: {
  "region": "North India",
  "season": "Rabi",
  "soil_type": "loamy"
}
Validation: Custom properties per region
```

## Recommended Test Implementation

### Unit Tests Priority
1. **Critical**: Entity validation (Stage, CropStage)
2. **Critical**: Repository unique constraints
3. **Critical**: Service business logic
4. **High**: Permission checks
5. **High**: Edge cases and boundaries
6. **Medium**: JSONB property handling
7. **Medium**: Pagination and filtering
8. **Low**: Performance benchmarks

### Integration Test Workflows
1. Complete stage lifecycle (create → update → assign → reorder → delete)
2. Multi-crop stage management
3. Concurrent operations simulation
4. Permission-based access patterns
5. Error recovery scenarios

### E2E Test Scenarios
1. Farmer creating custom crop lifecycle
2. Admin managing master stage library
3. Bulk import of stages from template
4. Stage migration between seasons
5. Historical stage tracking for analytics

## Vulnerabilities and Recommendations

### Critical Issues Found

1. **Missing Crop Existence Validation**
   - Issue: No validation that crop_id exists when assigning stages
   - Risk: FK constraint error instead of proper validation
   - Fix: Add crop existence check in AssignStageToCrop

2. **Inactive Stage Assignment**
   - Issue: Can assign soft-deleted or inactive stages to crops
   - Risk: Data inconsistency, orphaned references
   - Fix: Validate stage is_active=true and deleted_at IS NULL

3. **Order Gap Handling**
   - Issue: No validation for order continuity (allows 1,2,5)
   - Risk: UI/UX issues, confusion in stage progression
   - Fix: Validate or auto-adjust order sequences

### Medium Priority Issues

1. **Reorder Map Validation**
   - Issue: Map allows duplicate handling uncertainty
   - Risk: Unpredictable behavior with duplicate orders
   - Fix: Validate all stages present and orders unique

2. **JSONB Size Limits**
   - Issue: No validation on properties size
   - Risk: Performance degradation, storage issues
   - Fix: Add size limits (e.g., 64KB max)

3. **Cascade Delete Impact**
   - Issue: Soft-deleted stages remain in crop assignments
   - Risk: Orphaned references, data inconsistency
   - Fix: Add cleanup or validation logic

### Low Priority Enhancements

1. **Stage Templates**
   - Add predefined stage templates for common crops
   - Allow copying stages between crops
   - Support stage groups or categories

2. **Duration Calculations**
   - Add total crop duration calculation
   - Support duration ranges (min-max days)
   - Add season-based duration adjustments

3. **Stage Transitions**
   - Track actual vs planned durations
   - Support stage completion tracking
   - Add transition conditions/rules

## Monitoring and Observability

### Key Metrics to Track
1. Stage creation rate per organization
2. Average stages per crop
3. Reordering frequency
4. Failed validation attempts
5. Permission denial rate
6. API response times (P50, P95, P99)

### Alerts to Configure
1. Unusual spike in stage creation (potential abuse)
2. High rate of validation failures
3. Database constraint violations
4. Transaction rollback frequency
5. JSONB parsing errors

### Audit Events
1. Stage lifecycle changes (create, update, delete)
2. Crop-stage assignments and removals
3. Reordering operations
4. Permission denials
5. Validation failures with context

## Conclusion

The stages management feature has solid foundational implementation but requires additional validation layers and edge case handling. Key areas for improvement:

1. **Add missing validations** for crop existence and stage activity
2. **Implement order continuity** checks or auto-adjustment
3. **Add size limits** for JSONB properties
4. **Enhance error messages** with more context
5. **Add concurrent operation** test coverage
6. **Implement stage templates** for common crops
7. **Add monitoring** for abuse detection

The feature follows good patterns with soft deletes, JSONB flexibility, and transactional consistency, but needs hardening for production use in agricultural scenarios.
