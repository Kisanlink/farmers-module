# Farmers Module

A Go microservice for managing farmers, farms, crop cycles, and farm activities. This service integrates with the AAA service for authentication and authorization, and uses PostgreSQL with PostGIS for spatial data storage.

## Architecture

The farmers module follows a clean architecture pattern with the following layers:

- **Domain**: Business entities and logic
- **Repository**: Data access layer using kisanlink-db
- **Service**: Business logic and orchestration
- **Transport**: gRPC server and HTTP gateway
- **Auth**: AAA service integration for permissions

## Features

- **Farmer Management**: Link farmers to FPOs, assign Kisan Sathis
- **Farm Management**: Create and manage farms with PostGIS geometry
- **Crop Cycles**: Track agricultural cycles and seasons
- **Farm Activities**: Monitor and complete farm activities
- **Spatial Operations**: PostGIS integration for geographic queries
- **AAA Integration**: Delegated authentication and authorization
- **gRPC + HTTP**: Dual protocol support via grpc-gateway

## Prerequisites

- Go 1.23+
- PostgreSQL 16+ with PostGIS extension
- AAA service running and accessible
- Docker and Docker Compose (for development)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd farmers-module
```

### 2. Install Dependencies

```bash
make deps
make tidy
```

### 3. Environment Configuration

```bash
cp env.example .env
# Edit .env with your configuration
```

### 4. Database Setup

```bash
# Start PostgreSQL with PostGIS
docker-compose up -d postgres

# Run migrations
make migrate-up
```

### 5. Generate Code

```bash
# Generate protobuf code
make proto

# Generate SQL code
make sqlc
```

### 6. Build and Run

```bash
# Build the service
make build

# Run the service
make run
```

## Development

### Project Structure

```
.
├── cmd/farmers-service/     # Main application entry point
├── internal/
│   ├── app/                # Application wiring and startup
│   ├── auth/               # AAA client and interceptors
│   ├── config/             # Configuration management
│   ├── domain/             # Business entities and logic
│   │   ├── farmer/         # Farmer-related entities
│   │   ├── farm/           # Farm-related entities
│   │   ├── crop_cycle/     # Crop cycle entities
│   │   ├── farm_activity/  # Farm activity entities
│   │   └── fpo/            # FPO reference entities
│   ├── repo/               # Data access layer
│   ├── service/            # Business services
│   └── transport/          # gRPC and HTTP transport
├── pkg/                    # Shared packages
│   └── common/             # Common utilities and errors
├── proto/                  # Protocol buffer definitions
├── migrations/             # Database migrations
└── Makefile               # Build and development tasks
```

### Key Commands

```bash
# Development setup
make dev-setup

# Run tests
make test

# Lint code
make lint

# Full build
make full-build

# Clean build artifacts
make clean
```

### Database Schema

The service uses the following main tables:

- **farms**: Farm boundaries with PostGIS geometry
- **crop_cycles**: Agricultural cycles within farms
- **farm_activities**: Individual activities within cycles
- **farmer_links**: Links between farmers and FPOs
- **fpo_refs**: FPO organization references

### AAA Integration

The service integrates with AAA for:

- User authentication and organization membership
- Permission checks on all operations
- Role-based access control
- Automatic seeding of required roles and permissions

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | gRPC service port | 8080 |
| `GATEWAY_PORT` | HTTP gateway port | 8081 |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `AAA_GRPC_ADDR` | AAA service address | localhost:9090 |
| `POSTGIS_SRID` | Spatial reference system | 4326 |

### Database Configuration

The service supports PostgreSQL with PostGIS for spatial operations. Ensure the PostGIS extension is enabled:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
```

## API Reference

### gRPC Services

- **FarmerService**: Farmer management operations
- **FarmService**: Farm CRUD and spatial operations
- **CropCycleService**: Crop cycle management
- **FarmActivityService**: Activity tracking
- **FPOService**: FPO reference management
- **AdminService**: Administrative operations

### HTTP Gateway

The service exposes HTTP endpoints via grpc-gateway for all gRPC operations.

## Testing

```bash
# Run unit tests
make test

# Run integration tests (requires test database)
make test-integration

# Run tests with coverage
go test -cover ./...
```

## Deployment

### Docker

```bash
# Build image
make docker-build

# Run container
docker run -p 8080:8080 -p 8081:8081 farmers-module
```

### Kubernetes

See `deploy/` directory for Kubernetes manifests.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run linting and tests
6. Submit a pull request

## License

[License information]
