# Crop Cycles Area Allocation - Requirements Specification

## Executive Summary

This document outlines the comprehensive requirements for enhancing the crop cycles feature with area allocation capabilities and farm activity organization by crop stages in the farmers-module microservice.

## Business Context

### Problem Statement
Farmers manage multiple crops on the same farm simultaneously, allocating different portions of their land to different crops. Currently, the system lacks:
1. Ability to track area allocation for each crop cycle
2. Validation to prevent over-allocation of farm area
3. Organization of farm activities by crop growth stages

### Business Value
- **Precision Agriculture**: Enable accurate tracking of land utilization per crop
- **Resource Optimization**: Prevent over-allocation and optimize land usage
- **Activity Management**: Organize and track farm activities by crop growth stages
- **Decision Support**: Provide data for yield analysis per hectare per crop

## Functional Requirements

### FR1: Area Allocation for Crop Cycles

#### FR1.1 Area Assignment
- **Description**: Each crop cycle must have an allocated area in hectares
- **Acceptance Criteria**:
  - Area must be specified when creating a crop cycle
  - Area must be a positive decimal number with up to 4 decimal places
  - Area unit is always hectares (ha)
  - System must support partial hectare allocations (e.g., 0.5 ha, 2.75 ha)

#### FR1.2 Area Validation
- **Description**: System must validate that total allocated area doesn't exceed farm area
- **Acceptance Criteria**:
  - Sum of all active crop cycles' areas for a farm must not exceed total farm area
  - Validation must occur on create, update, and delete operations
  - System must provide clear error messages when validation fails
  - Must handle concurrent operations correctly (prevent race conditions)

#### FR1.3 Area Modification
- **Description**: Allow updating crop cycle area allocation
- **Acceptance Criteria**:
  - Area can be modified for PLANNED and ACTIVE cycles
  - Area cannot be modified for COMPLETED or CANCELLED cycles
  - Modification must trigger re-validation of total area constraint
  - Must maintain audit trail of area changes

### FR2: Farm Activity Organization by Crop Stages

#### FR2.1 Stage Association
- **Description**: Farm activities must be associated with crop stages
- **Acceptance Criteria**:
  - Each activity must reference a specific crop stage
  - Stage must belong to the crop of the associated crop cycle
  - Activities without stage association should be allowed (backward compatibility)

#### FR2.2 Stage-based Filtering
- **Description**: Provide ability to filter and group activities by stage
- **Acceptance Criteria**:
  - API must support filtering activities by stage_id
  - API must support filtering activities by stage_order
  - Response should include stage details (name, order, duration)
  - Support pagination when filtering by stage

#### FR2.3 Stage Progress Tracking
- **Description**: Track completion of activities per stage
- **Acceptance Criteria**:
  - Calculate percentage of completed activities per stage
  - Identify current stage based on activity completion
  - Provide stage timeline visualization data

### FR3: Business Rules and Constraints

#### FR3.1 Area Allocation Rules
- **Rule**: Total allocated area ≤ Total farm area
- **Formula**: `SUM(crop_cycle.area_ha WHERE status IN ('PLANNED', 'ACTIVE')) <= farm.area_ha`
- **Enforcement**: Database constraint + application validation

#### FR3.2 Crop Cycle Status Rules
- **PLANNED**: Can modify area freely (within farm limits)
- **ACTIVE**: Can modify area (with validation)
- **COMPLETED**: Cannot modify area (historical record)
- **CANCELLED**: Cannot modify area (terminated cycle)

#### FR3.3 Concurrent Cycles Rules
- Multiple cycles can be active on same farm
- Cycles can overlap in time if total area permits
- Same crop can have multiple cycles if area permits

## Non-Functional Requirements

### NFR1: Performance
- **Query Performance**: Area validation queries must complete within 100ms
- **Bulk Operations**: Support bulk creation of crop cycles with area validation
- **Caching**: Implement caching for frequently accessed farm area data
- **Database Indexing**: Proper indexes on area_ha, farm_id, status fields

### NFR2: Scalability
- **Concurrent Users**: Support 1000+ concurrent farmers managing crop cycles
- **Data Volume**: Handle farms with 50+ crop cycles per year
- **Geographic Distribution**: Optimize for distributed farm locations

### NFR3: Reliability
- **Data Consistency**: ACID compliance for area allocation transactions
- **Idempotency**: All modification operations must be idempotent
- **Rollback Capability**: Support transaction rollback on validation failure
- **Audit Trail**: Complete audit log of all area modifications

### NFR4: Security
- **Authorization**: AAA service integration for all operations
- **Data Validation**: Input sanitization for all area values
- **SQL Injection Prevention**: Parameterized queries only
- **Rate Limiting**: Prevent abuse of validation endpoints

### NFR5: Usability
- **Error Messages**: Clear, actionable error messages in local languages
- **API Documentation**: Comprehensive Swagger/OpenAPI documentation
- **Default Values**: Sensible defaults for optional parameters
- **Backward Compatibility**: Existing APIs continue to work

## Data Requirements

### DR1: Crop Cycle Data
```json
{
  "id": "CRCY_xxxxx",
  "farm_id": "FARM_xxxxx",
  "farmer_id": "FRMR_xxxxx",
  "crop_id": "CROP_xxxxx",
  "variety_id": "VARTY_xxxxx",
  "area_ha": 5.5,  // NEW FIELD
  "season": "KHARIF",
  "status": "ACTIVE",
  "start_date": "2024-06-01",
  "end_date": "2024-11-30",
  "outcome": {}
}
```

### DR2: Farm Activity Data
```json
{
  "id": "FACT_xxxxx",
  "crop_cycle_id": "CRCY_xxxxx",
  "crop_stage_id": "CSTG_xxxxx",  // NEW FIELD
  "farmer_id": "FRMR_xxxxx",
  "activity_type": "IRRIGATION",
  "planned_at": "2024-07-15T10:00:00Z",
  "completed_at": "2024-07-15T14:30:00Z",
  "status": "COMPLETED",
  "output": {},
  "metadata": {}
}
```

### DR3: Area Allocation Summary
```json
{
  "farm_id": "FARM_xxxxx",
  "total_area_ha": 10.0,
  "allocated_area_ha": 8.5,
  "available_area_ha": 1.5,
  "allocations": [
    {
      "crop_cycle_id": "CRCY_001",
      "crop_name": "Moringa",
      "area_ha": 5.0,
      "status": "ACTIVE"
    },
    {
      "crop_cycle_id": "CRCY_002",
      "crop_name": "Rice",
      "area_ha": 3.5,
      "status": "ACTIVE"
    }
  ]
}
```

## Integration Requirements

### IR1: AAA Service Integration
- All operations require authentication
- Permission checks for create/update/delete operations
- Organization-level data isolation
- Audit logging of all modifications

### IR2: Existing System Compatibility
- Maintain backward compatibility with existing APIs
- Support migration of existing crop cycles (null area_ha)
- Gradual migration strategy for farm activities

### IR3: Reporting Integration
- Provide data for area-based yield reports
- Support analytics on land utilization
- Enable stage-wise activity analysis

## User Stories

### US1: Farmer allocates land to multiple crops
**As a** farmer
**I want to** allocate specific areas to different crop cycles
**So that** I can grow multiple crops simultaneously on my farm

**Acceptance Criteria**:
- I can specify area when creating a crop cycle
- I can see remaining available area before allocation
- I receive an error if I try to allocate more than available area

### US2: Agricultural officer tracks activity progress by stage
**As an** agricultural officer
**I want to** view farm activities organized by crop stages
**So that** I can track farming progress and provide timely advice

**Acceptance Criteria**:
- I can filter activities by crop stage
- I can see completion percentage per stage
- I can identify delayed or missed activities per stage

### US3: System prevents over-allocation
**As a** system administrator
**I want** the system to prevent area over-allocation
**So that** data integrity is maintained

**Acceptance Criteria**:
- System blocks creation of crop cycles exceeding farm area
- System handles concurrent allocation requests correctly
- System provides clear error messages on validation failure

## Acceptance Criteria

### AC1: Area Allocation
✓ Crop cycles can be created with area_ha field
✓ Area validation prevents over-allocation
✓ Concurrent operations are handled correctly
✓ Area modifications are properly validated

### AC2: Activity Organization
✓ Activities can be linked to crop stages
✓ Activities can be filtered by stage
✓ Stage progress can be calculated
✓ Backward compatibility is maintained

### AC3: Performance
✓ Area validation completes within 100ms
✓ Bulk operations are supported
✓ No degradation in existing API performance

### AC4: Security
✓ All operations are properly authorized
✓ Input validation prevents injection attacks
✓ Audit trail is complete

## Out of Scope

1. **Crop rotation planning** - Future enhancement
2. **Yield prediction based on area** - Separate feature
3. **Geographic area overlap detection** - Requires GIS enhancement
4. **Multi-season planning** - Future roadmap
5. **Area-based cost calculation** - Finance module responsibility

## Dependencies

1. **Stages Management Feature** - Must be implemented and working
2. **AAA Service** - Must support required permission checks
3. **PostGIS** - Required for potential geographic validations
4. **Database Migration Tools** - For schema updates

## Risks and Mitigation

### Risk 1: Race Conditions in Area Allocation
- **Impact**: High - Could lead to over-allocation
- **Probability**: Medium
- **Mitigation**: Use database transactions with appropriate isolation level

### Risk 2: Performance Impact on Large Farms
- **Impact**: Medium - Slow validation on farms with many cycles
- **Probability**: Low
- **Mitigation**: Implement caching and optimized queries

### Risk 3: Breaking Existing Integrations
- **Impact**: High - Could disrupt current operations
- **Probability**: Low
- **Mitigation**: Maintain backward compatibility, gradual rollout

## Success Metrics

1. **Adoption Rate**: 80% of farmers using area allocation within 3 months
2. **Data Accuracy**: 99.9% accuracy in area validation
3. **Performance**: 95% of validation queries complete within 100ms
4. **User Satisfaction**: 4+ star rating from farmers on the feature
5. **Error Rate**: Less than 0.1% validation errors in production

## Glossary

- **Area Allocation**: Assignment of specific hectares to a crop cycle
- **Crop Cycle**: A complete growing season for a specific crop on a farm
- **Crop Stage**: A phase in the crop growth lifecycle
- **Farm Activity**: A specific task performed during farming operations
- **Over-allocation**: Attempting to allocate more area than available on farm
- **Active Cycle**: A crop cycle currently in progress (status = ACTIVE)
