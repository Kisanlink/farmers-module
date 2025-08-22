# Requirements Document

## Introduction

This specification defines the critical workflows for the farmers-module service, a Go-based microservice that manages farmer-to-FPO (Farmer Producer Organization) relationships, farm management, and agricultural activities. The service integrates with PostgreSQL/PostGIS for geospatial data storage and delegates authentication/authorization to an external AAA service via gRPC. The system enforces role-based access control across all operations and provides HTTP APIs for client applications.

## Requirements

### Requirement 1: Farmer Registration and User Management

**User Story:** As a system administrator, I want to register farmers with both AAA identity and farmer-specific details, so that farmers can access the system with proper authentication and profile information.

#### Acceptance Criteria

1. WHEN RegisterFarmer is called THEN the service SHALL validate farmer registration details including name, phone, email, and farmer-specific fields
2. WHEN farmer details are valid THEN the service SHALL call AAA service gRPC to check if user already exists by phone/email
3. IF user exists in AAA THEN the service SHALL retrieve existing user details and create local farmer profile
4. IF user does not exist THEN the service SHALL call AAA service to create new user with farmer role
5. WHEN farmer is created in AAA THEN the service SHALL store additional farmer-specific fields in local farmers table
6. WHEN KisanSathi assignment is requested THEN the service SHALL verify user ID exists in AAA and has KisanSathi role
7. IF KisanSathi role validation fails THEN the service SHALL return 400 Bad Request with role validation error
8. WHEN farmer registration succeeds THEN the service SHALL return 201 Created with farmer profile and AAA user ID
9. IF AAA service is unavailable THEN the service SHALL return 503 Service Unavailable with retry guidance

### Requirement 2: FPO Organization Creation and Management

**User Story:** As a system administrator, I want to create FPO organizations with proper role hierarchy, so that FPOs can operate with appropriate user groups and permissions.

#### Acceptance Criteria

1. WHEN CreateFPO is called THEN the service SHALL validate FPO details including name, registration number, and CEO user information
2. WHEN FPO creation is requested THEN the service SHALL call AAA service to create organization with provided details
3. WHEN organization is created THEN the service SHALL create CEO user in AAA if not exists and assign CEO role to organization
4. WHEN CEO is configured THEN the service SHALL create user groups for directors, shareholders, store_staff, and store_managers
5. WHEN user groups are created THEN the service SHALL configure appropriate permissions for each group within the organization
6. WHEN FPO setup is complete THEN the service SHALL store FPO reference in local fpo_refs table with AAA organization ID
7. IF CEO user creation fails THEN the service SHALL rollback organization creation and return appropriate error
8. IF user group creation fails THEN the service SHALL attempt cleanup and return 500 Internal Server Error with details
9. WHEN FPO operations succeed THEN the service SHALL emit fpo.created event with organization and user group details

### Requirement 3: AAA Service Extension Requirements

**User Story:** As a developer, I want to ensure all required AAA service functionality is available, so that the farmers-module can integrate seamlessly with authentication and authorization.

#### Acceptance Criteria

1. WHEN AAA integration is needed THEN the AAA service SHALL provide CreateUser gRPC method with user details and role assignment
2. WHEN user lookup is required THEN the AAA service SHALL provide GetUserByPhone and GetUserByEmail methods
3. WHEN organization management is needed THEN the AAA service SHALL provide CreateOrganization method with metadata support
4. WHEN user group management is required THEN the AAA service SHALL provide CreateUserGroup and AddUserToGroup methods
5. WHEN role validation is needed THEN the AAA service SHALL provide CheckUserRole method for role verification
6. WHEN permission management is required THEN the AAA service SHALL provide AssignPermissionToGroup method
7. IF any required AAA methods are missing THEN they SHALL be implemented as part of this specification scope
8. WHEN AAA methods are implemented THEN they SHALL follow consistent error handling and return structured responses

### Requirement 4: Database Foundation and Connectivity

**User Story:** As a system administrator, I want a robust database foundation with PostGIS support, so that the service can store and query geospatial farm data reliably.

#### Acceptance Criteria

1. WHEN the service starts THEN the system SHALL establish a GORM connection to PostgreSQL database
2. WHEN the database connection is established THEN the system SHALL enable PostGIS extension for geospatial operations
3. WHEN the service initializes THEN the system SHALL bootstrap required ENUM types (season, cycle_status, farmer_status, linkage_status)
4. WHEN the service starts THEN the system SHALL execute AutoMigrate for all domain models (Farmer, FPORef, FarmerLink, Farm, CropCycle, FarmActivity)
5. WHEN AutoMigrate completes THEN the system SHALL execute raw SQL to ensure SRID validation, ST_IsValid constraints, area_ha generated columns, and GIST spatial indexes
6. IF database connection fails THEN the system SHALL log structured error and fail to start
7. WHEN database operations are performed THEN the system SHALL maintain connection pooling and handle connection timeouts gracefully

### Requirement 5: Core Domain Models and Repository Layer

**User Story:** As a developer, I want well-defined GORM models and repository interfaces, so that I can perform consistent CRUD operations on all domain entities.

#### Acceptance Criteria

1. WHEN models are defined THEN each model SHALL inherit from a Base model with ID, CreatedAt, UpdatedAt, DeletedAt fields
2. WHEN Farmer model is created THEN it SHALL contain AAA user ID reference and farmer-specific fields (phone, address, land_ownership_type, etc.)
3. WHEN FPORef model is created THEN it SHALL contain AAA organization ID reference, name, registration details, and business configuration
4. WHEN FarmerLink model is created THEN it SHALL establish relationship between farmers and FPOs with status tracking and KisanSathi assignment
5. WHEN Farm model is created THEN it SHALL contain geospatial fields (geometry as PostGIS type), area_ha as generated column, and farmer association
6. WHEN CropCycle model is created THEN it SHALL track seasonal farming cycles with status, dates, and farm association
7. WHEN FarmActivity model is created THEN it SHALL log farming activities with timestamps, types, and crop cycle association
8. WHEN repositories are implemented THEN each SHALL provide Create, Read, Update, Delete, and List operations with AAA ID lookups
9. WHEN repository operations execute THEN they SHALL return structured errors for constraint violations and not found conditions

### Requirement 6: AAA Service Integration

**User Story:** As a security administrator, I want all operations to be authenticated and authorized through the AAA service, so that access control is consistently enforced across the system.

#### Acceptance Criteria

1. WHEN the service starts THEN it SHALL establish a gRPC client connection to the AAA service
2. WHEN authentication is required THEN the system SHALL call CheckPermission with resource and action parameters
3. WHEN user context is needed THEN the system SHALL call ResolveUser to get user details from token
4. WHEN organization context is needed THEN the system SHALL call ResolveOrg to get organization details
5. WHEN the service initializes THEN it SHALL call EnsureRolesAndPermissions to seed required resources, actions, and role bindings idempotently
6. IF AAA service is unavailable THEN the system SHALL return 503 Service Unavailable with retry-after header
7. WHEN AAA operations fail THEN the system SHALL log structured errors with correlation IDs

### Requirement 7: Authentication and Authorization Middleware

**User Story:** As an API consumer, I want my requests to be automatically authenticated and authorized, so that I can access only the resources I'm permitted to use.

#### Acceptance Criteria

1. WHEN a request is received THEN the authentication middleware SHALL extract the bearer token from Authorization header
2. WHEN token is extracted THEN the middleware SHALL resolve it to aaa_subject and aaa_org using AAA service
3. WHEN route is accessed THEN the authorization middleware SHALL map the route to appropriate (resource, action) tuple
4. WHEN authorization check is performed THEN the middleware SHALL call AAA CheckPermission with subject, org, resource, and action
5. IF authentication fails THEN the system SHALL return 401 Unauthorized with structured error response
6. IF authorization fails THEN the system SHALL return 403 Forbidden with structured error response
7. WHEN requests are processed THEN the audit middleware SHALL log {subject, org, resource, action, status} in structured JSON format
8. WHEN errors occur THEN the system SHALL return structured JSON errors with error codes, messages, and correlation IDs

### Requirement 8: Identity and Organization Linkage Workflows

**User Story:** As an FPO administrator, I want to manage farmer-FPO relationships and organizational references, so that I can maintain proper organizational structure and member management.

#### Acceptance Criteria

1. WHEN LinkFarmerToFPO is called THEN the service SHALL validate 'farmer.link' permission on the target organization
2. WHEN farmer linkage is requested THEN the service SHALL verify both farmer and FPO exist in AAA service
3. WHEN linkage is valid THEN the service SHALL upsert farmer_links record with aaa_user_id and aaa_org_id
4. WHEN UnlinkFarmerFromFPO is called THEN the service SHALL soft delete the link by setting status to INACTIVE
5. WHEN RegisterFPORef is called THEN the service SHALL cache FPO business configuration in fpo_refs table
6. WHEN KisanSathi assignment is requested THEN the service SHALL update farmer_links.kisan_sathi_user_id field
7. WHEN operations succeed THEN the service SHALL emit appropriate events (farmer.linked, farmer.unlinked, etc.)
8. IF AAA validation fails THEN the service SHALL return appropriate 404 or 403 errors with correlation IDs

### Requirement 9: Farm Management and Geospatial Workflows

**User Story:** As a farmer, I want to manage my farm boundaries and geospatial data, so that I can track farming activities and participate in FPO programs with accurate location information.

#### Acceptance Criteria

1. WHEN CreateFarm is called THEN the service SHALL validate 'farm.create' permission with appropriate scope (self/assigned/org)
2. WHEN farm geometry is provided THEN the service SHALL enforce SRID 4326 and validate polygon integrity using PostGIS ST_IsValid
3. WHEN farm is created THEN the database SHALL automatically compute area_ha using ST_Area function and store in generated column
4. WHEN UpdateFarm is called THEN the service SHALL validate changes and update geometry/metadata with 'farm.update' permission
5. WHEN DeleteFarm is called THEN the service SHALL cascade delete to related cycles and activities with 'farm.delete' permission
6. WHEN ListFarms is called THEN the service SHALL support filtering by farmer/org/bbox with spatial intersects queries
7. WHEN ValidateGeometry is performed THEN the service SHALL check for self-intersections and optional bbox constraints (India)
8. WHEN farm operations succeed THEN the service SHALL emit events (farm.created, farm.updated, farm.deleted)
9. IF geometry validation fails THEN the service SHALL return 400 Bad Request with specific PostGIS validation errors

### Requirement 10: Administrative and Health Check Workflows

**User Story:** As a system administrator, I want administrative endpoints for system management and health monitoring, so that I can maintain service reliability and troubleshoot issues.

#### Acceptance Criteria

1. WHEN SeedRolesAndPermissions endpoint is called THEN the service SHALL trigger a complete reseed of AAA resources, actions, and role bindings
2. WHEN seeding is performed THEN the operation SHALL be idempotent and not create duplicate entries
3. WHEN HealthCheck endpoint is called THEN the service SHALL verify database connectivity and AAA service availability
4. WHEN health check passes THEN the service SHALL return 200 OK with component status details
5. IF database is unavailable THEN health check SHALL return 503 Service Unavailable with database status
6. IF AAA service is unavailable THEN health check SHALL return 503 Service Unavailable with AAA status
7. WHEN administrative operations are performed THEN they SHALL require 'admin' role and be fully audited

### Requirement 11: Error Handling and Observability

**User Story:** As a developer and operator, I want comprehensive error handling and structured logging, so that I can debug issues and monitor system behavior effectively.

#### Acceptance Criteria

1. WHEN errors occur THEN the system SHALL return structured JSON responses with error codes, messages, and correlation IDs
2. WHEN database constraints are violated THEN the system SHALL map GORM errors to appropriate HTTP status codes
3. WHEN validation fails THEN the system SHALL return 400 Bad Request with field-specific validation errors
4. WHEN resources are not found THEN the system SHALL return 404 Not Found with resource-specific messages
5. WHEN authorization fails THEN the system SHALL return 403 Forbidden without exposing sensitive information
6. WHEN system errors occur THEN the system SHALL return 500 Internal Server Error with correlation ID for tracking
7. WHEN operations are performed THEN the system SHALL log structured JSON with timestamps, correlation IDs, user context, and operation details

### Requirement 12: Crop Cycle Lifecycle Management

**User Story:** As a farmer, I want to manage crop cycles on my farms, so that I can track seasonal farming activities and plan agricultural operations effectively.

#### Acceptance Criteria

1. WHEN StartCycle is called THEN the service SHALL create crop_cycles record with status PLANNED or ACTIVE and validate 'cycle.start' permission
2. WHEN cycle is started THEN the service SHALL record start_date, season, and planned_crops information
3. WHEN UpdateCycle is called THEN the service SHALL allow changes to season/dates/planned crops for non-terminal cycles with 'cycle.update' permission
4. WHEN EndCycle is called THEN the service SHALL set status to COMPLETED or CANCELLED with end_date and outcome
5. WHEN ListCycles is called THEN the service SHALL support filtering by farm/status/season with 'cycle.list' permission
6. WHEN cycle operations succeed THEN the service SHALL emit events (cycle.started, cycle.updated, cycle.ended)
7. IF cycle is in terminal state THEN update operations SHALL be restricted to allowed fields only

### Requirement 13: Farm Activity Lifecycle Management

**User Story:** As a farmer and KisanSathi, I want to create and track farm activities within crop cycles, so that I can monitor agricultural operations and maintain activity records.

#### Acceptance Criteria

1. WHEN CreateActivity is called THEN the service SHALL insert farm_activities record with 'activity.create' permission and proper scope validation
2. WHEN activity is created THEN the service SHALL record activity_type, planned_at, metadata, and created_by information
3. WHEN CompleteActivity is called THEN the service SHALL set completed_at, output, and update metadata with 'activity.complete' permission
4. WHEN UpdateActivity is called THEN the service SHALL allow editing of planned details for non-completed activities with 'activity.update' permission
5. WHEN ListActivities is called THEN the service SHALL support filtering by cycle/type/date/status with 'activity.list' permission
6. WHEN activity operations succeed THEN the service SHALL emit events (activity.created, activity.completed, activity.updated)
7. IF activity belongs to different farmer scope THEN operations SHALL be denied with 403 Forbidden

### Requirement 14: Data Quality and Validation Workflows

**User Story:** As a system administrator, I want automated data quality checks and validation, so that the system maintains data integrity and prevents conflicts.

#### Acceptance Criteria

1. WHEN geometry is submitted THEN the service SHALL perform synchronous validation including SRID enforcement and self-intersection checks
2. WHEN farm boundaries overlap within organization THEN the service SHALL detect spatial intersections and report or block with 'farm.audit' permission
3. WHEN validation fails THEN the service SHALL emit geo.validation_failed events for audit purposes
4. WHEN ReconcileAAALinks is executed THEN the service SHALL probe AAA service and heal broken references in farmer_links
5. WHEN RebuildSpatialIndexes is called THEN the service SHALL reindex GIST indexes for maintenance with 'admin.maintain' permission
6. IF drift is detected during reconciliation THEN the service SHALL mark farmer_links status appropriately or fix references

### Requirement 15: Reporting and Analytics Workflows

**User Story:** As a farmer and FPO administrator, I want access to portfolio summaries and organizational dashboards, so that I can monitor performance and make informed decisions.

#### Acceptance Criteria

1. WHEN ExportFarmerPortfolio is called THEN the service SHALL aggregate farms, cycles, and activities data with 'report.read' permission
2. WHEN OrgDashboardCounters is requested THEN the service SHALL provide org-level KPIs including counts and areas by season/status
3. WHEN reports are generated THEN the service SHALL respect user scope and organizational boundaries
4. WHEN reporting operations are performed THEN the service SHALL return transient data without persistent storage
5. IF user lacks proper scope THEN reporting operations SHALL return 403 Forbidden with appropriate error messages

### Requirement 16: Comprehensive Testing and Observability

**User Story:** As a developer and operator, I want comprehensive testing capabilities and observability features, so that I can ensure system reliability and troubleshoot issues effectively.

#### Acceptance Criteria

1. WHEN HealthCheck endpoint is called THEN the service SHALL perform DB ping and AAA ping for liveness/readiness checks
2. WHEN any mutation or authorization denial occurs THEN the service SHALL emit structured audit logs with subject/org/resource/action/object details
3. WHEN tests are executed THEN each workflow SHALL have corresponding unit tests, integration tests, and API tests
4. WHEN test scenarios run THEN they SHALL cover success paths, error conditions, permission denials, and edge cases
5. WHEN audit events are emitted THEN they SHALL include correlation IDs, timestamps, and structured JSON format
6. WHEN system errors occur THEN they SHALL be logged with sufficient context for debugging and monitoring
7. WHEN performance testing is conducted THEN the system SHALL handle concurrent requests and maintain response time SLAs
