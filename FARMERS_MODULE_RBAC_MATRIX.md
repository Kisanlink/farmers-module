# Farmers Module RBAC Matrix

## Overview

This document provides a comprehensive role vs permissions vs resources vs actions matrix for the farmers-module service, showing the complete RBAC (Role-Based Access Control) structure and permission mappings.

## Core Components

### 1. Roles (Who)

The farmers-module defines the following roles:

- **FARMER**: Individual farmer users
- **KISAN_SATHI**: Field agents who support farmers
- **FPO_CEO**: Chief Executive Officer of Farmer Producer Organization
- **FPO_DIRECTOR**: Director of Farmer Producer Organization
- **FPO_SHAREHOLDER**: Shareholder in Farmer Producer Organization

### 2. Resources (What)

The farmers-module manages the following resources:

- **farmer**: Farmer profiles and management
- **farm**: Farm properties and operations
- **crop_cycle**: Agricultural crop cycles
- **farm_activity**: Farm activities and tasks
- **fpo_ref**: FPO reference data
- **fpo**: FPO organization management
- **cycle**: Crop cycle management (alias for crop_cycle)
- **activity**: Farm activity management (alias for farm_activity)
- **admin**: Administrative operations
- **report**: Reporting and analytics
- **system**: System health and status

### 3. Actions (How)

The farmers-module supports the following actions:

- **create**: Create new resources
- **read**: Read/view resource data
- **update**: Modify existing resources
- **delete**: Remove resources
- **list**: List multiple resources
- **assign**: Assign resources to users
- **start**: Start processes or cycles
- **end**: End processes or cycles
- **complete**: Mark activities as complete
- **link**: Link farmers to FPOs
- **unlink**: Unlink farmers from FPOs
- **assign_kisan_sathi**: Assign KisanSathi to farmers
- **audit**: Audit and validation operations
- **maintain**: System maintenance operations
- **health**: Health check operations

## Complete RBAC Matrix

### Role: FARMER

| Resource | Actions | Description |
|----------|---------|-------------|
| farmer | read, update | View and update own profile |
| farm | create, read, update, delete, list | Manage own farms |
| crop_cycle | start, read, update, end, list | Manage own crop cycles |
| farm_activity | create, read, update, complete, list | Manage own farm activities |
| fpo_ref | read | View FPO reference information |

**Total Permissions**: 20

### Role: KISAN_SATHI

| Resource | Actions | Description |
|----------|---------|-------------|
| farmer | read, update, assign | View and update farmer profiles, assign to FPOs |
| farm | read, list | View farms under supervision |
| crop_cycle | read, list | View crop cycles under supervision |
| farm_activity | read, list | View farm activities under supervision |
| fpo_ref | read | View FPO reference information |

**Total Permissions**: 15

### Role: FPO_CEO

| Resource | Actions | Description |
|----------|---------|-------------|
| farmer | create, read, update, delete, list, assign | Full farmer management |
| farm | create, read, update, delete, list | Full farm management |
| crop_cycle | create, read, update, delete, list | Full crop cycle management |
| farm_activity | create, read, update, delete, list | Full farm activity management |
| fpo_ref | create, read, update, delete, list | Full FPO reference management |
| fpo | create, read, update, delete, list | Full FPO management |
| report | read | Access to organizational reports |

**Total Permissions**: 35

### Role: FPO_DIRECTOR

| Resource | Actions | Description |
|----------|---------|-------------|
| farmer | read, update, list | View and update farmer information |
| farm | read, update, list | View and update farm information |
| crop_cycle | read, update, list | View and update crop cycle information |
| farm_activity | read, update, list | View and update farm activity information |
| fpo_ref | read, update, list | View and update FPO reference information |
| fpo | read, update, list | View and update FPO information |
| report | read | Access to organizational reports |

**Total Permissions**: 28

### Role: FPO_SHAREHOLDER

| Resource | Actions | Description |
|----------|---------|-------------|
| farmer | read, list | View farmer information |
| farm | read, list | View farm information |
| crop_cycle | read, list | View crop cycle information |
| farm_activity | read, list | View farm activity information |
| fpo_ref | read, list | View FPO reference information |
| fpo | read, list | View FPO information |
| report | read | Access to organizational reports |

**Total Permissions**: 21

## Route Permission Mapping

### Farmer Management Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/farmers` | farmer | create | farmer.create |
| GET | `/api/v1/farmers/:id` | farmer | read | farmer.read |
| PUT | `/api/v1/farmers/:id` | farmer | update | farmer.update |
| DELETE | `/api/v1/farmers/:id` | farmer | delete | farmer.delete |
| GET | `/api/v1/farmers` | farmer | list | farmer.list |

### FPO Management Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/fpos` | fpo | create | fpo.create |
| GET | `/api/v1/fpos/:id` | fpo | read | fpo.read |
| PUT | `/api/v1/fpos/:id` | fpo | update | fpo.update |
| DELETE | `/api/v1/fpos/:id` | fpo | delete | fpo.delete |
| GET | `/api/v1/fpos` | fpo | list | fpo.list |

### Farmer Linkage Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/farmer-links` | farmer | link | farmer.link |
| DELETE | `/api/v1/farmer-links` | farmer | unlink | farmer.unlink |
| PUT | `/api/v1/farmer-links/kisan-sathi` | farmer | assign_kisan_sathi | farmer.assign_kisan_sathi |

### Farm Management Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/farms` | farm | create | farm.create |
| GET | `/api/v1/farms/:id` | farm | read | farm.read |
| PUT | `/api/v1/farms/:id` | farm | update | farm.update |
| DELETE | `/api/v1/farms/:id` | farm | delete | farm.delete |
| GET | `/api/v1/farms` | farm | list | farm.list |

### Crop Cycle Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/crop-cycles` | cycle | start | cycle.start |
| GET | `/api/v1/crop-cycles/:id` | cycle | read | cycle.read |
| PUT | `/api/v1/crop-cycles/:id` | cycle | update | cycle.update |
| DELETE | `/api/v1/crop-cycles/:id` | cycle | end | cycle.end |
| GET | `/api/v1/crop-cycles` | cycle | list | cycle.list |

### Farm Activity Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/farm-activities` | activity | create | activity.create |
| GET | `/api/v1/farm-activities/:id` | activity | read | activity.read |
| PUT | `/api/v1/farm-activities/:id` | activity | update | activity.update |
| PATCH | `/api/v1/farm-activities/:id/complete` | activity | complete | activity.complete |
| GET | `/api/v1/farm-activities` | activity | list | activity.list |

### Data Quality Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/data-quality/validate-geometry` | farm | audit | farm.audit |
| POST | `/api/v1/data-quality/reconcile-aaa-links` | admin | maintain | admin.maintain |
| POST | `/api/v1/data-quality/rebuild-spatial-indexes` | admin | maintain | admin.maintain |
| POST | `/api/v1/data-quality/detect-farm-overlaps` | farm | audit | farm.audit |

### Reporting Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| GET | `/api/v1/reports/farmer-portfolio` | report | read | report.read |
| GET | `/api/v1/reports/org-dashboard` | report | read | report.read |

### Administrative Routes

| HTTP Method | Route | Resource | Action | Required Permission |
|-------------|-------|----------|--------|-------------------|
| POST | `/api/v1/admin/seed-roles` | admin | maintain | admin.maintain |
| GET | `/api/v1/health` | system | health | system.health |

## Permission Inheritance and Hierarchy

### Role Hierarchy

```
FPO_CEO (Highest Authority)
├── FPO_DIRECTOR
├── FPO_SHAREHOLDER
├── KISAN_SATHI
└── FARMER (Base Level)
```

### Permission Inheritance Rules

1. **FPO_CEO**: Has all permissions across all resources
2. **FPO_DIRECTOR**: Inherits read and update permissions from FPO_CEO
3. **FPO_SHAREHOLDER**: Inherits read permissions from FPO_DIRECTOR
4. **KISAN_SATHI**: Has limited management permissions for assigned farmers
5. **FARMER**: Has permissions only for own resources

### Resource Access Patterns

- **Own Resources**: Users can always access resources they own
- **Organization Resources**: FPO roles can access resources within their organization
- **Supervised Resources**: KisanSathi can access resources under their supervision
- **Public Resources**: Some resources (like FPO references) are readable by all authenticated users

## Implementation Details

### AAA Service Integration

The farmers-module integrates with an external AAA service for:
- User authentication and token validation
- Role-based permission checking
- Organization-scoped access control
- Audit logging and compliance

### Permission Checking Flow

1. **Authentication**: JWT token validation via AAA service
2. **Authorization**: Route-to-permission mapping
3. **Permission Check**: AAA service validates user permissions
4. **Access Control**: Grant or deny access based on permissions
5. **Audit Logging**: Log all access attempts and decisions

### Security Considerations

- All API endpoints require authentication (except health checks)
- Permissions are checked at the resource-action level
- Organization-scoped access prevents cross-organization data access
- Audit trails track all permission checks and access attempts
- Soft deletes maintain data integrity while allowing recovery

## Summary Statistics

- **Total Roles**: 5
- **Total Resources**: 11
- **Total Actions**: 16
- **Total Route Permissions**: 45
- **Total Role-Permission Combinations**: 119

This RBAC matrix provides a comprehensive foundation for secure access control in the farmers-module service, ensuring that users can only access resources and perform actions appropriate to their role and organizational context.
