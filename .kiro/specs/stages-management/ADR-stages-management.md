# Architecture Decision Record: Stages Management Feature

## Status
Accepted and Implemented

## Context

The farmers module requires a comprehensive system for managing crop growth stages to support agricultural planning and tracking. This feature enables:

1. **Master Stage Management**: Defining reusable growth stages (e.g., Seeding, Germination, Vegetative Growth, Flowering, Harvesting)
2. **Crop-Stage Relationships**: Assigning specific stages to different crops with custom ordering and duration
3. **Flexible Metadata**: Supporting custom properties for both stages and crop-specific configurations
4. **Agricultural Planning**: Providing timeline information for farm activities and crop cycles

### Business Requirements
- Support multiple crops with different growth patterns
- Allow customization of stage durations per crop variety
- Maintain stage ordering for sequential workflow
- Enable soft deletion to preserve historical data
- Integrate with existing AAA service for permissions

## Decision

We have implemented a two-entity architecture with the following design choices:

### 1. Entity Design

#### Stage Entity (Master Table)
```
stages {
  id: string (STGE-prefixed hash ID)
  stage_name: string (unique, indexed)
  description: text (optional)
  properties: jsonb (flexible metadata)
  is_active: boolean
  + BaseModel fields (created_at, updated_at, deleted_at, etc.)
}
```

#### CropStage Entity (Junction Table with Attributes)
```
crop_stages {
  id: string (CSTG-prefixed hash ID)
  crop_id: string (FK to crops)
  stage_id: string (FK to stages)
  stage_order: integer (>= 1, unique per crop)
  duration_days: integer (optional)
  duration_unit: enum (DAYS, WEEKS, MONTHS)
  properties: jsonb (stage-specific configuration)
  is_active: boolean
  + BaseModel fields
}
```

### 2. Architecture Patterns

#### Repository Pattern
- Inherits from `BaseFilterableRepository` (kisanlink-db)
- Custom methods for domain-specific queries
- GORM-based implementation with preloading support
- Transaction support for complex operations (reordering)

#### Service Layer
- Business logic encapsulation
- AAA permission checks at service level
- Input validation and conflict detection
- Error transformation to domain-specific errors

#### Handler Layer
- HTTP endpoint management
- Request/response DTO transformation
- Context extraction (user_id, org_id, request_id)
- Swagger documentation annotations

### 3. Key Design Decisions

#### JSONB for Flexible Properties
- **Decision**: Use PostgreSQL JSONB columns for both entities
- **Rationale**: Allows schema evolution without migrations
- **Trade-off**: Less type safety, requires validation in application layer

#### Soft Delete Support
- **Decision**: Leverage BaseModel's soft delete functionality
- **Rationale**: Preserve audit trail and enable data recovery
- **Implementation**: All queries filter by `deleted_at IS NULL`

#### Stage Ordering Strategy
- **Decision**: Integer-based ordering with uniqueness per crop
- **Rationale**: Simple, efficient sorting and reordering
- **Alternative Considered**: Float-based (rejected due to precision issues)

#### Duration Flexibility
- **Decision**: Optional duration with multiple units (days/weeks/months)
- **Rationale**: Different crops have varying growth patterns
- **Storage**: Separate fields for value and unit (not normalized to single unit)

#### ID Generation
- **Decision**: Use kisanlink-db hash-based IDs with prefixes
- **Rationale**: Consistent with module patterns, human-readable prefixes
- **Configuration**: STGE (Medium table), CSTG (Medium table)

## Consequences

### Positive Consequences

1. **Flexibility**: JSONB properties enable custom attributes without schema changes
2. **Reusability**: Master stages can be shared across multiple crops
3. **Scalability**: Efficient indexing and query patterns
4. **Maintainability**: Clean separation of concerns with repository/service/handler layers
5. **Consistency**: Follows established kisanlink-db patterns
6. **Auditability**: Soft delete preserves historical records
7. **Type Safety**: Strong typing in Go with proper validation
8. **Permission Control**: Integrated AAA service for fine-grained access control

### Negative Consequences

1. **JSONB Validation**: Properties validation must be handled in application code
2. **Migration Complexity**: Changes to JSONB structure require data migration scripts
3. **Query Performance**: Complex JSONB queries may be slower than normalized columns
4. **Testing Overhead**: Junction table logic requires comprehensive test coverage

### Mitigation Strategies

1. **JSONB Validation**: Implement property validators in entity layer
2. **Performance**: Add GIN indexes on JSONB columns if needed
3. **Testing**: Comprehensive unit and integration test suites
4. **Documentation**: Clear API documentation for property schemas

## Implementation Details

### API Endpoints

#### Master Stage Operations
- `POST /api/v1/stages` - Create stage
- `GET /api/v1/stages` - List stages (paginated)
- `GET /api/v1/stages/:id` - Get stage details
- `PUT /api/v1/stages/:id` - Update stage
- `DELETE /api/v1/stages/:id` - Soft delete stage
- `GET /api/v1/stages/lookup` - Simplified lookup data

#### Crop-Stage Relationship Operations
- `POST /api/v1/crops/:crop_id/stages` - Assign stage to crop
- `GET /api/v1/crops/:crop_id/stages` - Get crop stages (ordered)
- `PUT /api/v1/crops/:crop_id/stages/:stage_id` - Update crop stage
- `DELETE /api/v1/crops/:crop_id/stages/:stage_id` - Remove stage from crop
- `POST /api/v1/crops/:crop_id/stages/reorder` - Reorder stages

### Security Model

All operations require AAA service permissions:
- Stage operations: `stage:create`, `stage:read`, `stage:update`, `stage:delete`, `stage:list`
- Crop-stage operations: `crop_stage:create`, `crop_stage:read`, `crop_stage:update`, `crop_stage:delete`

### Database Indexes

```sql
-- Stages table
CREATE UNIQUE INDEX stages_stage_name_idx ON stages(stage_name);
CREATE INDEX stages_is_active_idx ON stages(is_active);

-- Crop stages table
CREATE INDEX crop_stages_crop_id_idx ON crop_stages(crop_id);
CREATE INDEX crop_stages_stage_id_idx ON crop_stages(stage_id);
CREATE INDEX crop_stages_stage_order_idx ON crop_stages(stage_order);
CREATE UNIQUE INDEX crop_stages_crop_stage_idx ON crop_stages(crop_id, stage_id) WHERE deleted_at IS NULL;
```

## Alternatives Considered

### Alternative 1: Single Table with Crop-Specific Stages
- **Rejected**: Would duplicate stage definitions across crops
- **Reason**: Violates DRY principle, harder to maintain consistency

### Alternative 2: Normalized Duration Storage
- **Rejected**: Store all durations in days
- **Reason**: Loss of user intent, conversion complexity

### Alternative 3: Graph-based Stage Transitions
- **Rejected**: Model stages as a directed graph with transitions
- **Reason**: Over-engineering for current requirements

## References

- kisanlink-db documentation for BaseModel and repository patterns
- PostgreSQL JSONB documentation
- GORM soft delete patterns
- AAA service integration guidelines

## Review and Approval

- **Author**: System Architect
- **Reviewers**: Backend Team Lead, Product Owner
- **Approval Date**: [Implementation Completed]
- **Implementation Status**: 100% Complete
