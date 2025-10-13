# Stages Management Feature - Design Document

## 1. Overview

The Stages Management feature provides comprehensive functionality for defining and managing crop growth stages within the farmers module. It enables agricultural planning by associating specific growth stages with crops, tracking durations, and maintaining stage-specific metadata.

## 2. Data Model

### 2.1 Entity Relationship Diagram

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────┐
│     stages      │         │   crop_stages    │         │    crops    │
├─────────────────┤         ├──────────────────┤         ├─────────────┤
│ id (PK)         │◄────────┤ stage_id (FK)    │         │ id (PK)     │
│ stage_name      │         │ crop_id (FK)     │────────►│ name        │
│ description     │         │ id (PK)          │         │ category    │
│ properties      │         │ stage_order      │         │ ...         │
│ is_active       │         │ duration_days    │         └─────────────┘
│ created_at      │         │ duration_unit    │
│ updated_at      │         │ properties       │
│ deleted_at      │         │ is_active        │
└─────────────────┘         │ created_at       │
                            │ updated_at       │
                            │ deleted_at       │
                            └──────────────────┘
```

### 2.2 Entity Details

#### Stage Entity (Master Data)
- **Purpose**: Defines reusable growth stages across the agricultural system
- **Key Fields**:
  - `id`: Unique identifier with STGE prefix (e.g., STGE-ABC123XYZ)
  - `stage_name`: Unique name for the stage (e.g., "Seeding", "Germination")
  - `description`: Optional detailed description
  - `properties`: JSONB field for flexible metadata
  - `is_active`: Boolean flag for soft enable/disable
  - Audit fields from BaseModel

#### CropStage Entity (Junction with Attributes)
- **Purpose**: Associates stages with specific crops and defines crop-specific attributes
- **Key Fields**:
  - `id`: Unique identifier with CSTG prefix
  - `crop_id`: Foreign key to crops table
  - `stage_id`: Foreign key to stages table
  - `stage_order`: Integer defining sequence (must be >= 1)
  - `duration_days`: Optional duration value
  - `duration_unit`: Enum (DAYS, WEEKS, MONTHS)
  - `properties`: JSONB for crop-specific stage configuration
  - `is_active`: Boolean flag
  - Audit fields from BaseModel

### 2.3 JSONB Property Schemas

#### Stage Properties Example
```json
{
  "category": "growth",
  "indicators": ["leaf_count", "height"],
  "critical": true,
  "notes": "Monitor water requirements closely"
}
```

#### CropStage Properties Example
```json
{
  "water_requirement": "high",
  "fertilizer_schedule": ["NPK-20-20-20"],
  "temperature_range": {"min": 20, "max": 30, "unit": "celsius"},
  "pest_risks": ["aphids", "whitefly"],
  "custom_duration_variance": 5
}
```

## 3. API Design

### 3.1 Master Stage Operations

#### Create Stage
- **Endpoint**: `POST /api/v1/stages`
- **Request Body**:
  ```json
  {
    "stage_name": "Flowering",
    "description": "Flower development stage",
    "properties": {
      "category": "reproductive",
      "critical": true
    }
  }
  ```
- **Response**: Stage object with generated ID
- **Permissions**: `stage:create`

#### List Stages
- **Endpoint**: `GET /api/v1/stages`
- **Query Parameters**:
  - `page` (default: 1)
  - `page_size` (default: 20, max: 100)
  - `search` (searches name and description)
  - `is_active` (filter by status)
- **Response**: Paginated list of stages
- **Permissions**: `stage:list`

#### Get Stage Details
- **Endpoint**: `GET /api/v1/stages/:id`
- **Response**: Complete stage object
- **Permissions**: `stage:read`

#### Update Stage
- **Endpoint**: `PUT /api/v1/stages/:id`
- **Request Body**: Partial update fields
- **Response**: Updated stage object
- **Permissions**: `stage:update`

#### Delete Stage
- **Endpoint**: `DELETE /api/v1/stages/:id`
- **Response**: Success confirmation
- **Note**: Soft delete only
- **Permissions**: `stage:delete`

#### Stage Lookup
- **Endpoint**: `GET /api/v1/stages/lookup`
- **Response**: Simplified list for dropdowns
- **Permissions**: `stage:list`

### 3.2 Crop-Stage Relationship Operations

#### Assign Stage to Crop
- **Endpoint**: `POST /api/v1/crops/:crop_id/stages`
- **Request Body**:
  ```json
  {
    "stage_id": "STGE-ABC123",
    "stage_order": 2,
    "duration_days": 30,
    "duration_unit": "DAYS",
    "properties": {
      "water_requirement": "moderate"
    }
  }
  ```
- **Validations**:
  - Stage must exist
  - Order must be unique for the crop
  - Stage-crop combination must not exist
- **Permissions**: `crop_stage:create`

#### Get Crop Stages
- **Endpoint**: `GET /api/v1/crops/:crop_id/stages`
- **Response**: Ordered list of stages with details
- **Permissions**: `crop_stage:read`

#### Update Crop Stage
- **Endpoint**: `PUT /api/v1/crops/:crop_id/stages/:stage_id`
- **Request Body**: Partial update (order, duration, properties)
- **Validations**: Order uniqueness if changed
- **Permissions**: `crop_stage:update`

#### Remove Stage from Crop
- **Endpoint**: `DELETE /api/v1/crops/:crop_id/stages/:stage_id`
- **Response**: Success confirmation
- **Note**: Soft delete only
- **Permissions**: `crop_stage:delete`

#### Reorder Crop Stages
- **Endpoint**: `POST /api/v1/crops/:crop_id/stages/reorder`
- **Request Body**:
  ```json
  {
    "stage_orders": {
      "STGE-ABC123": 1,
      "STGE-DEF456": 2,
      "STGE-GHI789": 3
    }
  }
  ```
- **Note**: Transactional update
- **Permissions**: `crop_stage:update`

## 4. Business Rules and Validations

### 4.1 Stage Management Rules
1. **Unique Stage Names**: Case-insensitive uniqueness
2. **Name Length**: Maximum 100 characters
3. **Soft Delete**: Stages are never hard deleted
4. **Active Filtering**: Inactive stages excluded from lookups

### 4.2 Crop-Stage Assignment Rules
1. **Stage Order**:
   - Must be >= 1
   - Unique per crop
   - Sequential recommended but not enforced

2. **Duration**:
   - Optional field
   - If provided, must be > 0
   - Unit must be valid (DAYS/WEEKS/MONTHS)

3. **Duplicate Prevention**:
   - Same stage cannot be assigned twice to a crop
   - Checked even for soft-deleted records

### 4.3 Data Integrity Rules
1. **Foreign Key Constraints**: Enforced at database level
2. **Cascade Behavior**: Soft delete cascades to relationships
3. **Orphan Prevention**: Cannot delete stage if actively used

## 5. Security Model

### 5.1 Permission Matrix

| Operation | Resource | Permission Required | Scope |
|-----------|----------|-------------------|-------|
| Create Stage | stages | stage:create | Organization |
| View Stage | stages | stage:read | Resource/Organization |
| Update Stage | stages | stage:update | Resource/Organization |
| Delete Stage | stages | stage:delete | Resource/Organization |
| List Stages | stages | stage:list | Organization |
| Assign to Crop | crop_stages | crop_stage:create | Organization |
| View Crop Stages | crop_stages | crop_stage:read | Organization |
| Update Crop Stage | crop_stages | crop_stage:update | Organization |
| Remove from Crop | crop_stages | crop_stage:delete | Organization |

### 5.2 Context Requirements
All operations require:
- Valid `user_id` from authentication
- Valid `org_id` for organization scope
- Optional `request_id` for tracing

## 6. Edge Cases and Error Handling

### 6.1 Handled Edge Cases

1. **Concurrent Updates**:
   - Optimistic locking via updated_at timestamp
   - Transaction isolation for reordering

2. **Missing References**:
   - Returns 404 for non-existent stage/crop
   - Validates FKs before insertion

3. **Order Conflicts**:
   - Detects duplicate orders before save
   - Returns 409 Conflict with details

4. **Soft Delete Complications**:
   - Queries filter deleted records
   - Unique constraints consider deletion

5. **JSONB Validation**:
   - Accepts any valid JSON structure
   - Application-level schema validation if needed

### 6.2 Error Responses

| Scenario | HTTP Status | Error Code | Message |
|----------|-------------|------------|---------|
| Invalid input | 400 | INVALID_INPUT | Validation details |
| Unauthorized | 401 | UNAUTHORIZED | Authentication required |
| Forbidden | 403 | FORBIDDEN | Insufficient permissions |
| Not found | 404 | NOT_FOUND | Resource not found |
| Conflict | 409 | ALREADY_EXISTS | Duplicate resource |
| Server error | 500 | INTERNAL_ERROR | Internal server error |

## 7. Integration Points

### 7.1 AAA Service Integration
- **Authentication**: Token validation via AAA service
- **Authorization**: Permission checks for all operations
- **Audit**: Operation logging through AAA audit trail

### 7.2 Database Integration
- **kisanlink-db**: Leverages BaseModel and BaseFilterableRepository
- **GORM**: ORM for database operations
- **PostgreSQL**: JSONB support and indexing

### 7.3 ID Generation
- **Hash Package**: Uses kisanlink-db hash generation
- **Counter Management**: Maintains distributed counters
- **Prefix System**: STGE for stages, CSTG for crop stages

## 8. Performance Considerations

### 8.1 Database Indexes
```sql
-- Stages table
CREATE UNIQUE INDEX idx_stages_name ON stages(LOWER(stage_name));
CREATE INDEX idx_stages_active ON stages(is_active) WHERE deleted_at IS NULL;

-- Crop stages table
CREATE INDEX idx_crop_stages_crop ON crop_stages(crop_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_crop_stages_stage ON crop_stages(stage_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_crop_stages_order ON crop_stages(crop_id, stage_order) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_crop_stages_unique ON crop_stages(crop_id, stage_id) WHERE deleted_at IS NULL;
```

### 8.2 Query Optimization
- Preloading relationships where needed
- Pagination for list operations
- Filtered indexes for soft delete queries
- Limit page size to 100 records

### 8.3 Caching Strategy
- Stage lookup data cacheable (changes infrequently)
- Crop stages cacheable with crop-based invalidation
- Cache TTL: 5 minutes for lookups

## 9. Migration and Deployment

### 9.1 Database Migration
- Tables created via GORM AutoMigrate
- Indexes added in post-migration setup
- ID counters initialized from existing data

### 9.2 Rollback Strategy
- Soft delete allows data recovery
- Version tracking in properties field
- Audit trail for change history

### 9.3 Feature Flags
- Can be gated behind feature flag if needed
- Gradual rollout by organization
- API versioning for breaking changes

## 10. Testing Strategy

### 10.1 Unit Tests
- Entity validation logic
- Repository CRUD operations
- Service business rules
- Handler request/response

### 10.2 Integration Tests
- End-to-end API flows
- Database constraint validation
- AAA permission checks
- Transaction integrity

### 10.3 Performance Tests
- Bulk stage creation
- Large crop-stage assignments
- Concurrent reordering operations
- JSONB query performance

## 11. Monitoring and Observability

### 11.1 Metrics
- API endpoint latencies
- Database query performance
- Error rates by endpoint
- Stage/crop-stage creation rates

### 11.2 Logging
- Structured logging with Zap
- Request/response correlation
- Error stack traces
- Audit events

### 11.3 Alerts
- High error rates (> 1%)
- Slow queries (> 100ms)
- Failed permission checks
- Database connection issues

## 12. Future Enhancements

1. **Stage Templates**: Predefined stage sets for common crops
2. **Stage Transitions**: Rules for valid stage progressions
3. **Duration Predictions**: ML-based duration estimates
4. **Stage Dependencies**: Prerequisites and dependencies
5. **Bulk Operations**: Import/export stage configurations
6. **Version Control**: Track stage definition changes
7. **Localization**: Multi-language stage names/descriptions
8. **Notifications**: Alerts for stage transitions in crop cycles
