# Project Structure & Architecture

## Directory Organization

```
farmers-module/
├── cmd/farmers-service/        # Application entry point
├── internal/                   # Private application code
│   ├── api/grpc/              # gRPC server implementation
│   ├── auth/                  # AAA integration middleware
│   ├── clients/aaa/           # AAA gRPC client
│   ├── config/                # Configuration management
│   ├── db/                    # Database connection & migrations
│   ├── entities/              # Domain models & request/response types
│   │   ├── crop_cycle/        # Crop cycle domain models
│   │   ├── farm/              # Farm domain models
│   │   ├── farm_activity/     # Farm activity domain models
│   │   ├── farmer/            # Farmer domain models
│   │   ├── fpo/               # FPO reference models
│   │   ├── requests/          # API request models
│   │   └── responses/         # API response models
│   ├── handlers/              # HTTP request handlers
│   ├── interfaces/            # Service interfaces
│   ├── middleware/            # HTTP middleware
│   ├── repo/                  # Repository layer (data access)
│   ├── routes/                # Route definitions
│   ├── services/              # Business logic layer
│   ├── transport/             # Transport layer (HTTP/gRPC)
│   └── utils/                 # Utility functions
├── middleware/                # Shared middleware (legacy)
├── pkg/                       # Public packages
│   ├── common/                # Common utilities & errors
│   └── proto/                 # Generated protobuf code
└── proto/                     # Protocol buffer definitions
```

## Architecture Patterns

### Clean Architecture Layers

1. **Transport Layer**: HTTP handlers, gRPC servers, middleware
2. **Service Layer**: Business logic, workflow orchestration
3. **Repository Layer**: Data access, GORM models
4. **Domain Layer**: Core entities, business rules

### Workflow-Based Design

- 19 defined workflows (W1-W19) grouped by domain:
  - W1-W3: Identity & Organization Linkage
  - W4-W5: KisanSathi Assignment
  - W6-W9: Farm Management
  - W10-W17: Crop Management
  - W18-W19: Access Control

### Request/Response Pattern

- Standardized request/response models in `internal/entities/`
- Base models with common fields (RequestID, Timestamp, UserID, OrgID)
- Pagination and filtering support built-in
- Validation using struct tags

## Naming Conventions

### Files & Directories

- Snake case for directories: `crop_cycle/`, `farm_activity/`
- Descriptive file names: `farmer_repository.go`, `farm_handlers.go`
- Test files: `*_test.go` alongside source files

### Go Code

- PascalCase for exported types: `FarmerService`, `CreateFarmRequest`
- camelCase for unexported: `farmerRepo`, `validateInput`
- Interface suffix: `FarmerService`, `FarmRepository`
- Implementation suffix: `FarmerServiceImpl`, `FarmRepositoryImpl`

### Database

- Snake case tables: `farmer_links`, `crop_cycles`, `farm_activities`
- Consistent ID fields: `id` (primary), `aaa_user_id`, `aaa_org_id`
- Timestamp fields: `created_at`, `updated_at`, `deleted_at`

## Configuration Management

- Environment-based configuration in `internal/config/`
- `.env` file support with `godotenv`
- Validation and defaults for all config values
- Structured config types: `DatabaseConfig`, `ServerConfig`, `AAAConfig`

## Error Handling

- Structured errors using `pkg/common/errors.go`
- HTTP status code mapping in handlers
- AAA service error propagation
- Audit logging for all operations
