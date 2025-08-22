# Implementation Plan

- [x] 1. Database Foundation and Core Models Setup

  - Set up enhanced database connection with PostGIS extension and custom ENUMs
  - Implement all core GORM models (Farmer, FPORef, FarmerLink, Farm, CropCycle, FarmActivity)
  - Create post-migration setup for computed columns, spatial indexes, and constraints
  - Write unit tests for model validation and database operations
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

- [ ] 2. AAA Service Client Enhancement and Integration

  - Extend existing AAA client with missing methods (CreateOrganization, CreateUserGroup, etc.)
  - Implement user management methods (CreateUser, GetUserByPhone, CheckUserRole)
  - Add organization and user group management capabilities
  - Create comprehensive AAA client tests with gRPC mocks
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

- [ ] 3. Repository Layer Implementation

  - Implement FarmerRepository with CRUD operations and AAA user ID lookups
  - Create FPORefRepository with organization reference management
  - Build FarmerLinkRepository for farmer-FPO relationship management
  - Implement FarmRepository with PostGIS spatial operations and geometry validation
  - Create CropCycleRepository and FarmActivityRepository with proper filtering
  - Write comprehensive repository tests with database integration
  - _Requirements: 5.8, 5.9_

- [ ] 4. Authentication and Authorization Middleware

  - Implement JWT token extraction and validation middleware
  - Create authorization middleware with route-to-permission mapping
  - Build audit logging middleware with structured JSON output
  - Implement error handling middleware with correlation IDs and structured responses
  - Write middleware tests with HTTP mocks and AAA service integration
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8_

- [ ] 5. Farmer Registration and Management Service

  - Implement RegisterFarmer service with AAA user creation and local profile storage
  - Create GetFarmer, UpdateFarmer, and ListFarmers service methods
  - Add phone/email validation and duplicate user checking through AAA
  - Implement farmer registration HTTP handlers with proper validation
  - Write comprehensive tests for farmer registration workflow
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9_

- [ ] 6. FPO Creation and Organization Management Service

  - Implement CreateFPO service with AAA organization creation
  - Add CEO user setup and user group creation (directors, shareholders, store_staff, store_managers)
  - Create RegisterFPORef service for local FPO reference management
  - Implement permission assignment for user groups
  - Build FPO management HTTP handlers and write comprehensive tests
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9_

- [ ] 7. Farmer-FPO Linkage and KisanSathi Assignment Services

  - Implement LinkFarmerToFPO service with AAA validation and farmer_links management
  - Create UnlinkFarmerFromFPO service with soft delete functionality
  - Build AssignKisanSathi service with role validation through AAA
  - Implement ReassignOrRemoveKisanSathi service for KisanSathi management
  - Create linkage management HTTP handlers and write comprehensive tests
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8_

- [ ] 8. Farm Management Service with Geospatial Operations

  - Implement CreateFarm service with WKT validation and PostGIS integration
  - Create UpdateFarm and DeleteFarm services with proper authorization
  - Build ListFarms service with spatial filtering and bounding box queries
  - Add geometry validation service with SRID enforcement and integrity checks
  - Implement farm management HTTP handlers and write geospatial tests
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7, 9.8, 9.9_

- [ ] 9. Crop Cycle Lifecycle Management Service

  - Implement StartCycle service with farm validation and cycle creation
  - Create UpdateCycle service for non-terminal cycle modifications
  - Build EndCycle service with completion/cancellation logic
  - Add ListCycles service with filtering by farm/status/season
  - Create crop cycle HTTP handlers and write lifecycle tests
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7_

- [ ] 10. Farm Activity Management Service

  - Implement CreateActivity service with cycle validation and activity creation
  - Create CompleteActivity service with output recording and status updates
  - Build UpdateActivity service for pre-completion activity modifications
  - Add ListActivities service with filtering by cycle/type/date/status
  - Create activity management HTTP handlers and write activity lifecycle tests
  - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7_

- [ ] 11. Data Quality and Validation Services

  - Implement ValidateGeometry service with PostGIS validation and SRID checks
  - Create ReconcileAAALinks service for healing broken AAA references
  - Build RebuildSpatialIndexes service for database maintenance
  - Add overlap detection service for farm boundary conflicts
  - Write data quality tests and validation scenarios
  - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6_

- [ ] 12. Reporting and Analytics Services

  - Implement ExportFarmerPortfolio service with data aggregation
  - Create OrgDashboardCounters service for organizational KPIs
  - Build reporting HTTP handlers with proper scope validation
  - Add report generation tests with various data scenarios
  - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_

- [ ] 13. Administrative and System Management Services

  - Implement SeedRolesAndPermissions service for AAA bootstrapping
  - Create HealthCheck service with database and AAA service validation
  - Build administrative HTTP handlers with proper authorization
  - Add system management tests and health check scenarios
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7_

- [ ] 14. Error Handling and Observability Implementation

  - Implement structured error response system with correlation IDs
  - Create GORM error mapping to HTTP status codes
  - Build comprehensive audit logging with event emission
  - Add observability features (metrics, tracing, structured logging)
  - Write error handling tests and observability validation
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7_

- [ ] 15. API Documentation and Integration Testing

  - Generate comprehensive API documentation with Swagger/OpenAPI
  - Create integration tests for all workflow scenarios
  - Build end-to-end tests with real database and AAA service integration
  - Add performance tests for concurrent operations and large datasets
  - Write API usage examples and cURL commands for all endpoints
  - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6, 16.7_

- [ ] 16. Service Configuration and Deployment Setup
  - Enhance configuration management with validation and environment-specific settings
  - Create Docker containerization with proper health checks
  - Build deployment scripts and environment setup documentation
  - Add monitoring and alerting configuration
  - Create production readiness checklist and operational runbooks
  - _Requirements: All requirements for production deployment_
