# Technology Stack & Build System

## Core Technologies

- **Language**: Go 1.24+ with toolchain go1.24.4
- **Database**: PostgreSQL 16+ with PostGIS extension
- **ORM**: GORM with AutoMigrate for schema management
- **Web Framework**: Gin for HTTP REST APIs
- **gRPC**: Protocol Buffers with grpc-gateway for dual protocol support
- **Authentication**: Delegated to aaa-service via gRPC client
- **Logging**: Uber Zap structured logging
- **Configuration**: Environment variables with godotenv

## Key Dependencies

- `github.com/Kisanlink/kisanlink-db` - Shared database utilities
- `github.com/gin-gonic/gin` - HTTP web framework
- `google.golang.org/grpc` - gRPC client/server
- `gorm.io/gorm` - ORM with PostgreSQL driver
- `go.uber.org/zap` - Structured logging
- `github.com/swaggo/swag` - Swagger documentation generation

## Build System (Makefile)

```bash
# Development Commands
make build          # Build the application binary
make test           # Run all tests
make run            # Start development server
make dev            # Alias for run command

# Documentation
make docs           # Generate Swagger documentation

# Development Workflow
make dev-setup      # Setup development environment with docs
make clean          # Clean build artifacts

# Testing & Quality
make test           # Run unit tests
make lint           # Run code linting (via .golangci.yml)
```

## Database Management

- **Migrations**: Auto-migration via GORM with custom post-migration setup
- **Spatial**: PostGIS extension with SRID validation and GIST indexes
- **Enums**: Custom PostgreSQL enums (season, cycle_status, activity_status)
- **Code Generation**: SQLC for type-safe SQL queries

## Protocol Buffer Generation

```bash
# Generate protobuf code (when implemented)
make proto          # Generate Go code from .proto files
make sqlc           # Generate SQLC queries
```

## Development Server

- **HTTP Port**: 8000 (configurable via SERVICE_PORT)
- **API Documentation**: Available at `/docs` endpoint
- **Health Check**: Built-in health endpoints for monitoring
- **Hot Reload**: Use `make run` for development with auto-restart

Dont use generic objects for the request and response bodies in the swagger docs or annotation comments
