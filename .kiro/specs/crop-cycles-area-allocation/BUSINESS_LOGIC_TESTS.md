# Business Logic Validation Report - Crop Cycles Area Allocation

## Executive Summary

**Date**: 2025-10-14
**Module**: Farmers Module - Crop Cycles Area Allocation Feature
**Status**: ⚠️ **CRITICAL - Feature Not Implemented**

### Key Findings

1. **Missing Implementation**: The crop cycles area allocation feature described in the invariants is NOT implemented
2. **Model Gaps**: Critical fields are missing from data models
3. **No Area Validation**: No business logic exists to validate farm area allocation
4. **No Stage Validation**: Farm activities don't reference crop stages as specified

## Critical Implementation Gaps

### 1. CropCycle Entity Missing Fields
**Severity**: 🔴 CRITICAL

The `CropCycle` entity (`/internal/entities/crop_cycle/crop_cycle.go`) is missing the `area_ha` field entirely.

**Current Structure**:
```go
type CropCycle struct {
    FarmID    string     // ✅ Present
    FarmerID  string     // ✅ Present
    Season    string     // ✅ Present
    Status    string     // ✅ Present
    CropID    string     // ✅ Present
    // ❌ MISSING: area_ha float64
}
```

**Impact**: Cannot implement any area allocation invariants (INVARIANTS 1-5)

### 2. FarmActivity Entity Missing Stage Reference
**Severity**: 🔴 CRITICAL

The `FarmActivity` entity (`/internal/entities/farm_activity/farm_activity.go`) lacks `crop_stage_id` field.

**Current Structure**:
```go
type FarmActivity struct {
    CropCycleID  string    // ✅ Present
    ActivityType string    // ✅ Present (but not linked to stages)
    // ❌ MISSING: CropStageID string
}
```

**Impact**: Cannot validate stage-related invariants (INVARIANTS 6-9)

## Invariant Validation Status

### Area Allocation Invariants (1-5)

| Invariant | Description | Status | Notes |
|-----------|-------------|---------|--------|
| INVARIANT 1 | Sum of crop cycles area ≤ farm.area_ha | ❌ **BLOCKED** | No area_ha field in CropCycle |
| INVARIANT 2 | Individual crop cycle area > 0 | ❌ **BLOCKED** | No area_ha field to validate |
| INVARIANT 3 | Update validation for area | ❌ **BLOCKED** | No area update logic exists |
| INVARIANT 4 | Soft-deleted cycles excluded | ❌ **BLOCKED** | No area tracking to exclude |
| INVARIANT 5 | Farm area update validation | ⚠️ **PARTIAL** | Farm has area_ha but no cycle validation |

### Farm Activity Stage Invariants (6-9)

| Invariant | Description | Status | Notes |
|-----------|-------------|---------|--------|
| INVARIANT 6 | Valid crop stage reference | ❌ **NOT IMPLEMENTED** | No crop_stage_id field |
| INVARIANT 7 | Stage must be active | ❌ **NOT IMPLEMENTED** | No stage validation |
| INVARIANT 8 | Filter by crop stage | ❌ **NOT IMPLEMENTED** | No stage field to filter |
| INVARIANT 9 | Stage transition validation | ❌ **NOT IMPLEMENTED** | No stage tracking |

## Existing Business Logic Analysis

### Tested Invariants in Current Implementation

#### ✅ Working Invariants

1. **Farmer-FPO Relationship**
   - A farmer can only be ACTIVE in one FPO at a time
   - Status: Working correctly in tests

2. **Farm Geometry Validation**
   - Farm boundaries must be valid PostGIS polygons
   - Area calculation using ST_Area(geography) is correct
   - Status: Properly implemented

3. **Crop Cycle Status Transitions**
   - PLANNED → ACTIVE → COMPLETED/CANCELLED
   - Cannot move backwards in status
   - Status: Validated in service layer

#### ⚠️ Partially Implemented

1. **One Active Cycle per Farm**
   - Logic exists in tests but NOT enforced in service
   - `CropCycleService.StartCycle()` doesn't check for existing active cycles

## Security & Abuse Path Analysis

### 1. Race Condition: Multiple Active Cycles
**Severity**: 🟡 MEDIUM

**Attack Vector**:
```go
// Concurrent requests to start crop cycle for same farm
// Both could succeed, violating "one active cycle" rule
POST /api/crop-cycles/start (Request 1)
POST /api/crop-cycles/start (Request 2)
```

**Current Protection**: None
**Recommendation**: Add database constraint or use SELECT FOR UPDATE

### 2. Missing Validation in UpdateCycle
**Severity**: 🟡 MEDIUM

The `UpdateCycle` method allows changing `CropID` and `VarietyID` after cycle starts:
```go
if updateReq.CropID != nil {
    cycle.CropID = *updateReq.CropID  // Dangerous: changing crop mid-cycle
}
```

**Impact**: Could lead to data inconsistency

### 3. No Multi-tenancy Validation in Service Layer
**Severity**: 🔴 HIGH

Services rely entirely on AAA for org isolation but don't verify data belongs to the org:
```go
// Missing: Verify cycle.FarmID belongs to req.OrgID
cycle, err := s.cropCycleRepo.GetByID(ctx, cycleID, cycle)
```

### 4. Missing Crop/Farm Existence Validation
**Severity**: 🔴 HIGH

The service doesn't validate:
- Farm exists and belongs to the requesting farmer
- Crop ID is valid and exists in the system
- Variety belongs to the specified crop

### 5. Invalid Date Sequences Accepted
**Severity**: 🟡 MEDIUM

No validation for:
- End date before start date
- Dates outside reasonable agricultural bounds
- Overlapping seasons (KHARIF cycle extending into RABI period)

## Test Cases for Missing Implementation

### Area Allocation Tests (When Implemented)

```go
func TestCropCycleAreaAllocation(t *testing.T) {
    t.Run("total area cannot exceed farm area", func(t *testing.T) {
        // Given: Farm with 10 hectares
        farm := &Farm{AreaHa: 10.0}

        // Given: Existing cycle with 6 hectares
        existingCycle := &CropCycle{FarmID: farm.ID, AreaHa: 6.0, Status: "ACTIVE"}

        // When: Trying to create new cycle with 5 hectares
        newCycle := &CropCycle{FarmID: farm.ID, AreaHa: 5.0}

        // Then: Should fail (6 + 5 = 11 > 10)
        err := service.ValidateAreaAllocation(ctx, newCycle)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "exceeds available area")
    })

    t.Run("race condition: concurrent area allocation", func(t *testing.T) {
        // Test concurrent creation of crop cycles
        // Both trying to allocate 6 hectares on 10-hectare farm
        // Only one should succeed
    })

    t.Run("soft-deleted cycles release area", func(t *testing.T) {
        // Soft-delete a cycle and verify area is available again
    })
}
```

### Stage Validation Tests (When Implemented)

```go
func TestFarmActivityStageValidation(t *testing.T) {
    t.Run("activity must reference valid crop stage", func(t *testing.T) {
        // Given: Crop cycle for Wheat (crop_id: wheat123)
        // Given: Rice stage (belongs to different crop)
        // When: Creating activity with rice stage for wheat cycle
        // Then: Should fail validation
    })

    t.Run("cannot use deleted stage", func(t *testing.T) {
        // Soft-deleted stages should not be assignable
    })
}
```

## Recommended Implementation Plan

### Phase 1: Data Model Updates
1. Add `area_ha NUMERIC(10,2)` to crop_cycles table
2. Add `crop_stage_id VARCHAR(255)` to farm_activities table
3. Add foreign key constraints

### Phase 2: Business Logic Implementation
1. Implement `ValidateAreaAllocation()` method
2. Add area checking to `StartCycle()` and `UpdateCycle()`
3. Implement stage validation in farm activities

### Phase 3: Database Constraints
```sql
-- Prevent negative areas
ALTER TABLE crop_cycles ADD CONSTRAINT check_positive_area
    CHECK (area_ha > 0);

-- Create function to validate total area
CREATE FUNCTION validate_crop_cycle_area() RETURNS TRIGGER AS $$
BEGIN
    -- Check sum doesn't exceed farm area
    IF (SELECT SUM(area_ha) FROM crop_cycles
        WHERE farm_id = NEW.farm_id
        AND deleted_at IS NULL
        AND status IN ('PLANNED', 'ACTIVE')) >
       (SELECT area_ha_computed FROM farms WHERE id = NEW.farm_id) THEN
        RAISE EXCEPTION 'Total allocated area exceeds farm area';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### Phase 4: Concurrency Protection
1. Use database transactions with proper isolation
2. Implement optimistic locking with version fields
3. Add retry logic for concurrent operations

## Performance Concerns

1. **Area Calculation Query**: Summing areas for validation could be slow
   - Solution: Add indexed computed column for total_allocated_area on farms

2. **Stage Lookups**: Validating stages for each activity
   - Solution: Cache crop-stage relationships

## Monitoring & Alerts

### Recommended Production Monitors

1. **Area Allocation Violations**
   ```sql
   -- Alert if sum exceeds farm area (data corruption indicator)
   SELECT farm_id, SUM(area_ha) as total,
          (SELECT area_ha_computed FROM farms f WHERE f.id = farm_id) as farm_area
   FROM crop_cycles
   WHERE deleted_at IS NULL AND status IN ('ACTIVE', 'PLANNED')
   GROUP BY farm_id
   HAVING SUM(area_ha) > (SELECT area_ha_computed FROM farms f WHERE f.id = farm_id);
   ```

2. **Multiple Active Cycles**
   ```sql
   -- Detect farms with multiple active cycles
   SELECT farm_id, COUNT(*) as active_count
   FROM crop_cycles
   WHERE status = 'ACTIVE' AND deleted_at IS NULL
   GROUP BY farm_id
   HAVING COUNT(*) > 1;
   ```

## Conclusion

The crop cycles area allocation feature is **NOT IMPLEMENTED** and requires significant development work. The current implementation has no area tracking or validation, making it impossible to enforce the specified business invariants.

### Risk Assessment
- **Business Risk**: 🔴 HIGH - Farmers could over-allocate land leading to planning failures
- **Data Integrity Risk**: 🔴 HIGH - No constraints to prevent invalid data
- **Security Risk**: 🟡 MEDIUM - Race conditions and missing validations

### Recommendation
**DO NOT DEPLOY** the crop cycles feature to production until:
1. Area allocation is fully implemented
2. Stage validation is added
3. Proper database constraints are in place
4. Concurrency issues are addressed

## Test Execution Log

```
Tests Run: 0 (Feature not implemented)
Tests Passed: N/A
Tests Failed: N/A
Blocked Tests: 9 (All invariants blocked by missing implementation)
```

---
*Generated by Business Logic Tester*
*Analysis Date: 2025-10-14*
