# Farmers Module - Workflow-Based Architecture

This document describes the reorganized architecture of the farmers-module based on the comprehensive workflow catalog.

## Architecture Overview

The module has been reorganized to follow a clean, maintainable structure that groups functionality by business workflows rather than technical layers.

## Directory Structure

```
farmers-module/
├── cmd/
│   └── farmers-service/
│       └── main.go                 # New main entry point
├── internal/
│   ├── app/
│   │   └── app.go                  # Application lifecycle management
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── routes/                     # Workflow-grouped routes
│   │   ├── index.go                # Main route registration
│   │   ├── identity_routes.go      # W1-W3: Identity & Org Linkage
│   │   ├── kisansathi_routes.go    # W4-W5: KisanSathi Assignment
│   │   ├── farm_routes.go          # W6-W9: Farm Management
│   │   ├── crop_routes.go          # W10-W17: Crop Management
│   │   └── admin_routes.go         # W18-W19: Access Control
│   ├── handlers/                   # Workflow-specific handlers
│   │   ├── identity_handlers.go    # W1-W3 handlers
│   │   ├── kisansathi_handlers.go  # W4-W5 handlers
│   │   └── ...                     # Other workflow handlers
│   ├── services/                   # Business logic services
│   │   ├── service_factory.go      # Service dependency injection
│   │   ├── interfaces.go           # Service contracts
│   │   └── ...                     # Implementation services
│   └── repo/                       # Data access layer
│       └── repository_factory.go   # Repository dependency injection
├── models/                         # Existing domain models
├── repositories/                   # Existing data access
└── services/                       # Existing business logic
```

## Workflow Groups

### 1. Identity & Organization Linkage (W1-W3)
- **Route Group**: `/api/v1/identity`
- **Workflows**:
  - W1: Link farmer to FPO
  - W2: Unlink farmer from FPO
  - W3: Register FPO reference

### 2. KisanSathi Assignment (W4-W5)
- **Route Group**: `/api/v1/kisansathi`
- **Workflows**:
  - W4: Assign KisanSathi to farmer
  - W5: Reassign or remove KisanSathi

### 3. Farm Management (W6-W9)
- **Route Group**: `/api/v1/farms`
- **Workflows**:
  - W6: Create farm
  - W7: Update farm
  - W8: Delete farm
  - W9: List farms

### 4. Crop Management (W10-W17)
- **Route Group**: `/api/v1/crops`
- **Subgroups**:
  - **Cycles** (W10-W13): `/api/v1/crops/cycles`
  - **Activities** (W14-W17): `/api/v1/crops/activities`

### 5. Admin & Access Control (W18-W19)
- **Route Group**: `/api/v1/admin`
- **Workflows**:
  - W18: Seed roles and permissions
  - W19: Check permission

## Key Benefits

1. **Clear Separation of Concerns**: Each workflow group has its own routes, handlers, and services
2. **Easy to Maintain**: Developers can work on specific workflows without affecting others
3. **Scalable**: New workflows can be added by creating new route groups
4. **Human-Friendly**: Code organization follows business logic rather than technical patterns
5. **Consistent Structure**: All workflow groups follow the same pattern

## Implementation Status

- ✅ **Routes**: All workflow-based routes defined
- ✅ **Handlers**: Basic handler structure created (with TODO placeholders)
- ✅ **Services**: Service interfaces and factory defined
- ✅ **Configuration**: Basic config structure created
- ✅ **Application**: Main app structure created
- 🔄 **Services**: Implementation services need to be created
- 🔄 **Handlers**: Service integration needs to be implemented
- 🔄 **Testing**: Unit and integration tests need to be created

## Next Steps

1. **Implement Service Layer**: Create concrete implementations of all service interfaces
2. **Complete Handlers**: Integrate handlers with actual service calls
3. **Add Validation**: Implement request validation and error handling
4. **Add Middleware**: Implement AAA integration, logging, and monitoring
5. **Create Tests**: Add comprehensive test coverage for all workflows
6. **Documentation**: Add API documentation and usage examples

## Running the Application

```bash
# From the farmers-module directory
go run cmd/farmers-service/main.go
```

The server will start on port 8080 (or the PORT environment variable) and display available workflow groups.

## API Endpoints

### Identity & Organization
- `POST /api/v1/identity/farmer/link` - Link farmer to FPO
- `DELETE /api/v1/identity/farmer/unlink` - Unlink farmer from FPO
- `GET /api/v1/identity/farmer/linkage/:farmer_id/:org_id` - Get linkage status
- `POST /api/v1/identity/fpo/register` - Register FPO reference
- `GET /api/v1/identity/fpo/:org_id` - Get FPO reference

### KisanSathi Assignment
- `POST /api/v1/kisansathi/assign` - Assign KisanSathi to farmer
- `PUT /api/v1/kisansathi/reassign` - Reassign or remove KisanSathi
- `GET /api/v1/kisansathi/assignment/:farmer_id` - Get assignment

### Farm Management
- `POST /api/v1/farms/` - Create farm
- `PUT /api/v1/farms/:farm_id` - Update farm
- `DELETE /api/v1/farms/:farm_id` - Delete farm
- `GET /api/v1/farms/` - List farms
- `GET /api/v1/farms/:farm_id` - Get farm

### Crop Management
- **Cycles**:
  - `POST /api/v1/crops/cycles/` - Start cycle
  - `PUT /api/v1/crops/cycles/:cycle_id` - Update cycle
  - `PUT /api/v1/crops/cycles/:cycle_id/end` - End cycle
  - `GET /api/v1/crops/cycles/` - List cycles
  - `GET /api/v1/crops/cycles/:cycle_id` - Get cycle
- **Activities**:
  - `POST /api/v1/crops/activities/` - Create activity
  - `PUT /api/v1/crops/activities/:activity_id/complete` - Complete activity
  - `PUT /api/v1/crops/activities/:activity_id` - Update activity
  - `GET /api/v1/crops/activities/` - List activities
  - `GET /api/v1/crops/activities/:activity_id` - Get activity

### Admin & Access Control
- `POST /api/v1/admin/seed` - Seed roles and permissions
- `POST /api/v1/admin/check-permission` - Check permission
- `GET /api/v1/admin/health` - Health check
- `GET /api/v1/admin/audit` - Get audit trail
