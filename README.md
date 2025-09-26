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
- **Crop Master Data**: Complete crop, variety, and stage management system
- **Enhanced Crop Cycles**: Track agricultural cycles with detailed yield and harvest data
- **Farm Activities**: Monitor and complete farm activities
- **Bulk Operations**: Bulk farmer upload with validation and processing
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
- **crops**: Crop master data with categories, units, and seasons
- **crop_varieties**: Crop varieties with characteristics and duration
- **crop_stages**: Growth stages for crops with order and duration
- **crop_cycles**: Enhanced agricultural cycles with yield and harvest data
- **farm_activities**: Individual activities within cycles
- **farmer_links**: Links between farmers and FPOs
- **fpo_refs**: FPO organization references
- **bulk_operations**: Bulk upload operations tracking
- **processing_details**: Individual record processing details

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

### HTTP Endpoints

The service exposes RESTful HTTP endpoints for all operations:

#### Crop Master Data
- `POST /api/v1/crops` - Create crop
- `GET /api/v1/crops` - List crops with filtering
- `GET /api/v1/crops/{id}` - Get crop by ID
- `PUT /api/v1/crops/{id}` - Update crop
- `DELETE /api/v1/crops/{id}` - Delete crop

#### Crop Varieties
- `POST /api/v1/crops/varieties` - Create variety
- `GET /api/v1/crops/{crop_id}/varieties` - List varieties for crop
- `GET /api/v1/crops/varieties/{id}` - Get variety by ID
- `PUT /api/v1/crops/varieties/{id}` - Update variety
- `DELETE /api/v1/crops/varieties/{id}` - Delete variety

#### Crop Stages
- `POST /api/v1/crops/stages` - Create stage
- `GET /api/v1/crops/{crop_id}/stages` - List stages for crop
- `GET /api/v1/crops/stages/{id}` - Get stage by ID
- `PUT /api/v1/crops/stages/{id}` - Update stage
- `DELETE /api/v1/crops/stages/{id}` - Delete stage

#### Enhanced Crop Cycles
- `POST /api/v1/crops/cycles` - Start crop cycle
- `PUT /api/v1/crops/cycles/{id}` - Update crop cycle
- `PUT /api/v1/crops/cycles/{id}/end` - End crop cycle
- `PUT /api/v1/crops/cycles/{id}/harvest` - Record harvest
- `POST /api/v1/crops/cycles/{id}/report` - Upload report
- `GET /api/v1/crops/cycles` - List crop cycles
- `GET /api/v1/crops/cycles/{id}` - Get crop cycle by ID

#### Farm Activities
- `POST /api/v1/crops/activities` - Create activity
- `PUT /api/v1/crops/activities/{id}` - Update activity
- `PUT /api/v1/crops/activities/{id}/complete` - Complete activity
- `GET /api/v1/crops/activities` - List activities
- `GET /api/v1/crops/activities/{id}` - Get activity by ID

#### Bulk Operations
- `POST /api/v1/bulk/farmers/upload` - Upload bulk farmer data
- `GET /api/v1/bulk/farmers/status/{id}` - Get bulk operation status
- `GET /api/v1/bulk/farmers/template` - Download bulk upload template

#### Lookups
- `GET /api/v1/lookups/crop-data?type=categories` - Get crop categories
- `GET /api/v1/lookups/crop-data?type=units` - Get crop units
- `GET /api/v1/lookups/crop-data?type=seasons` - Get crop seasons

### gRPC Services

- **FarmerService**: Farmer management operations
- **FarmService**: Farm CRUD and spatial operations
- **CropService**: Crop master data management
- **CropCycleService**: Enhanced crop cycle management
- **FarmActivityService**: Activity tracking
- **BulkFarmerService**: Bulk operations
- **FPOService**: FPO reference management
- **AdminService**: Administrative operations

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
