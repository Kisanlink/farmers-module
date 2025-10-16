# AAA Service: CatalogService.SeedRolesAndPermissions Implementation Request

**Date:** 2025-10-16
**Requester:** Farmers Module Team
**Priority:** HIGH
**Type:** New Feature Implementation

---

## Context

The farmers-module needs to seed predefined roles and permissions in the AAA service at startup. While the protobuf definition for `CatalogService.SeedRolesAndPermissions` already exists in `pkg/proto/catalog.proto`, the actual gRPC service implementation is missing.

**Current Issue:**
```
Warning: Failed to seed AAA roles and permissions: failed to seed roles and permissions:
unknown service pb.CatalogService
```

The farmers-module has temporarily disabled this call (see commit 5480a97), but we need this functionality for proper role-based access control.

---

## Required Implementation

### 1. Implement CatalogService Server

The CatalogService is already defined in `pkg/proto/catalog.proto` (line 260-281):

```protobuf
service CatalogService {
    // Action management
    rpc RegisterAction(RegisterActionRequest) returns (RegisterActionResponse);
    rpc ListActions(ListActionsRequest) returns (ListActionsResponse);

    // Resource management
    rpc RegisterResource(RegisterResourceRequest) returns (RegisterResourceResponse);
    rpc SetResourceParent(SetResourceParentRequest) returns (SetResourceParentResponse);
    rpc ListResources(ListResourcesRequest) returns (ListResourcesResponse);

    // Role management
    rpc CreateRole(CreateRoleRequest) returns (CreateRoleResponse);
    rpc ListRoles(ListRolesRequest) returns (ListRolesResponse);

    // Permission management
    rpc CreatePermission(CreatePermissionRequest) returns (CreatePermissionResponse);
    rpc AttachPermissions(AttachPermissionsRequest) returns (AttachPermissionsResponse);
    rpc ListPermissions(ListPermissionsRequest) returns (ListPermissionsResponse);

    // Seeding
    rpc SeedRolesAndPermissions(SeedRolesAndPermissionsRequest) returns (SeedRolesAndPermissionsResponse);
}
```

### 2. Implement SeedRolesAndPermissions Method

**Request/Response Already Defined** (lines 242-257):

```protobuf
// Seed Roles and Permissions Request
message SeedRolesAndPermissionsRequest {
    bool force = 1;  // Force re-seed even if data exists
    string organization_id = 2;  // Optional - seed for specific org
}

// Seed Roles and Permissions Response
message SeedRolesAndPermissionsResponse {
    int32 status_code = 1;
    string message = 2;
    int32 roles_created = 3;
    int32 permissions_created = 4;
    int32 resources_created = 5;
    int32 actions_created = 6;
    repeated string created_roles = 7;
}
```

---

## Required Roles to Seed

The following roles must be created with `scope: "GLOBAL"`:

### 1. **farmer** (RoleFarmer)
- **Description:** Individual agricultural practitioner with permissions to manage own farms and crop cycles
- **Scope:** GLOBAL
- **Permissions Required:**
  - `farmer:create` - Create own farmer profile
  - `farmer:read` - Read own farmer data
  - `farmer:update` - Update own farmer profile
  - `farm:create` - Create farms under own profile
  - `farm:read` - Read own farms
  - `farm:update` - Update own farms
  - `farm:delete` - Delete own farms
  - `cycle:create` - Start crop cycles on own farms
  - `cycle:read` - Read own crop cycles
  - `cycle:update` - Update own crop cycles
  - `cycle:end` - End own crop cycles

### 2. **kisansathi** (RoleKisanSathi)
- **Description:** Field agent assigned to support farmers with data collection and advisory services
- **Scope:** GLOBAL
- **Permissions Required:**
  - All `farmer` permissions
  - `farmer:list` - List farmers assigned to them
  - `farm:list` - List farms of assigned farmers
  - `cycle:list` - List crop cycles of assigned farmers
  - `activity:create` - Create activities for assigned farmers
  - `activity:update` - Update activities
  - `activity:delete` - Delete activities

### 3. **CEO** (RoleFPOCEO)
- **Description:** Chief executive officer of a Farmer Producer Organization with full organizational management permissions
- **Scope:** GLOBAL
- **Permissions Required:**
  - All `kisansathi` permissions
  - `fpo:create` - Create FPO organization
  - `fpo:read` - Read FPO details
  - `fpo:update` - Update FPO details
  - `fpo:manage` - Full FPO management
  - `farmer:manage` - Manage all farmers in FPO
  - `kisansathi:assign` - Assign KisanSathis
  - `kisansathi:manage` - Manage KisanSathis

### 4. **fpo_manager** (RoleFPOManager)
- **Description:** Manager within an FPO with permissions to manage operations and access organizational data
- **Scope:** GLOBAL
- **Permissions Required:**
  - All `kisansathi` permissions
  - `fpo:read` - Read FPO details
  - `farmer:list` - List all farmers in FPO
  - `farm:list` - List all farms in FPO
  - `cycle:list` - List all crop cycles in FPO

### 5. **admin** (RoleAdmin)
- **Description:** System administrator with full permissions across all organizations and resources
- **Scope:** GLOBAL
- **Permissions Required:**
  - ALL permissions (wildcard `*:*`)

### 6. **readonly** (RoleReadOnly)
- **Description:** User with read-only access to resources within their organization
- **Scope:** GLOBAL
- **Permissions Required:**
  - `*:read` - Read access to all resources
  - `*:list` - List access to all resources

---

## Resources to Register

Register these resources for the farmers-module:

1. **farmer** - Farmer profile resource
2. **farm** - Farm resource
3. **cycle** - Crop cycle resource
4. **activity** - Activity resource
5. **fpo** - FPO organization resource
6. **kisansathi** - KisanSathi assignment resource
7. **stage** - Crop stage resource
8. **variety** - Crop variety resource

---

## Actions to Register

Register these actions:

1. **create** - Create a resource
2. **read** - Read a resource
3. **update** - Update a resource
4. **delete** - Delete a resource
5. **list** - List resources
6. **manage** - Full management access
7. **start** - Start a process (e.g., crop cycle)
8. **end** - End a process
9. **assign** - Assign a resource to another entity

---

## Implementation Guidelines

### Idempotency
```go
func (s *CatalogService) SeedRolesAndPermissions(ctx context.Context, req *pb.SeedRolesAndPermissionsRequest) (*pb.SeedRolesAndPermissionsResponse, error) {
    // 1. Check if data already exists
    if !req.Force {
        existingRoles, err := s.listExistingRoles(ctx)
        if err != nil {
            return nil, err
        }
        if len(existingRoles) > 0 {
            return &pb.SeedRolesAndPermissionsResponse{
                StatusCode: 200,
                Message: "Roles already seeded, use force=true to reseed",
                RolesCreated: 0,
            }, nil
        }
    }

    // 2. Create actions (create, read, update, delete, etc.)
    actions := []string{"create", "read", "update", "delete", "list", "manage", "start", "end", "assign"}
    actionsCreated := 0
    for _, action := range actions {
        if err := s.createActionIdempotent(ctx, action); err != nil {
            log.Printf("Warning: Failed to create action %s: %v", action, err)
        } else {
            actionsCreated++
        }
    }

    // 3. Create resources (farmer, farm, cycle, etc.)
    resources := []string{"farmer", "farm", "cycle", "activity", "fpo", "kisansathi", "stage", "variety"}
    resourcesCreated := 0
    for _, resource := range resources {
        if err := s.createResourceIdempotent(ctx, resource); err != nil {
            log.Printf("Warning: Failed to create resource %s: %v", resource, err)
        } else {
            resourcesCreated++
        }
    }

    // 4. Create permissions (resource + action combinations)
    permissionsCreated := 0
    // ... create permissions for each resource-action pair

    // 5. Create roles with permissions
    rolesCreated := 0
    createdRoleNames := []string{}

    // Create farmer role
    if err := s.createRoleWithPermissions(ctx, "farmer", farmerPermissions); err != nil {
        log.Printf("Warning: Failed to create farmer role: %v", err)
    } else {
        rolesCreated++
        createdRoleNames = append(createdRoleNames, "farmer")
    }

    // ... create other roles

    return &pb.SeedRolesAndPermissionsResponse{
        StatusCode: 200,
        Message: "Roles and permissions seeded successfully",
        RolesCreated: int32(rolesCreated),
        PermissionsCreated: int32(permissionsCreated),
        ResourcesCreated: int32(resourcesCreated),
        ActionsCreated: int32(actionsCreated),
        CreatedRoles: createdRoleNames,
    }, nil
}
```

### Error Handling
- Use idempotent operations (skip if already exists, don't fail)
- Log warnings for non-critical failures
- Return partial success with counts
- Support `force` parameter to delete and recreate

### Transaction Safety
- Use database transactions for atomic operations
- Rollback on critical failures
- Allow partial success for non-critical operations

---

## How Farmers-Module Will Call This

**From farmers-module startup** (`cmd/farmers-service/main.go`):

```go
// Seed AAA roles and permissions on startup (non-fatal)
log.Println("Seeding AAA roles and permissions...")
seedCtx, seedCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer seedCancel()

if err := aaaClient.SeedRolesAndPermissions(seedCtx); err != nil {
    log.Printf("Warning: Failed to seed AAA roles and permissions: %v", err)
    log.Println("Application will continue, but role assignments may fail if roles don't exist")
} else {
    log.Println("Successfully seeded AAA roles and permissions")
}
```

**From AAA client** (`internal/clients/aaa/aaa_client.go:901`):

```go
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
    grpcReq := &pb.SeedRolesAndPermissionsRequest{
        Force: false, // Don't force reseed if data exists
    }

    response, err := c.catalogClient.SeedRolesAndPermissions(ctx, grpcReq)
    if err != nil {
        return fmt.Errorf("failed to seed roles and permissions: %w", err)
    }

    if response.StatusCode != 200 && response.StatusCode != 201 {
        return fmt.Errorf("failed to seed roles and permissions: %s", response.Message)
    }

    log.Printf("Roles and permissions seeded successfully: %d roles, %d permissions created",
        response.RolesCreated, response.PermissionsCreated)
    return nil
}
```

---

## Testing Requirements

### Unit Tests
1. Test idempotent creation (calling twice shouldn't fail)
2. Test force flag (should delete and recreate)
3. Test partial failure handling
4. Test organization-specific seeding

### Integration Tests
1. Test full seeding from scratch
2. Test role assignment after seeding
3. Test permission checks with seeded roles
4. Test re-seeding with existing data

### Manual Testing
1. Start AAA service from scratch
2. Call SeedRolesAndPermissions
3. Verify roles exist in database
4. Verify permissions are attached to roles
5. Test role assignment to users
6. Test permission checks

---

## Server Registration

Don't forget to register the CatalogService server in your main gRPC server:

```go
// In cmd/aaa-service/main.go or similar
catalogServer := catalog.NewCatalogServer(db, logger)
pb.RegisterCatalogServiceServer(grpcServer, catalogServer)
```

---

## Success Criteria

- [ ] CatalogService is registered and accessible via gRPC
- [ ] SeedRolesAndPermissions creates all 6 roles
- [ ] All resources and actions are registered
- [ ] Permissions are properly attached to roles
- [ ] Idempotent operation (safe to call multiple times)
- [ ] Farmers-module can successfully seed roles on startup
- [ ] Role assignment works after seeding
- [ ] Permission checks work with seeded roles

---

## Timeline

**Estimated Effort:** 2-3 days
**Priority:** HIGH (blocking farmers-module role-based access control)

**Suggested Breakdown:**
- Day 1: Implement CatalogService server and SeedRolesAndPermissions method
- Day 2: Add unit tests and integration tests
- Day 3: Manual testing and deployment

---

## References

- **Protobuf Definition:** `aaa-service/pkg/proto/catalog.proto`
- **Farmers-Module Role Constants:** `farmers-module/internal/constants/roles.go`
- **Farmers-Module AAA Client:** `farmers-module/internal/clients/aaa/aaa_client.go`
- **ADR-001:** `farmers-module/.kiro/specs/adr-role-assignment-strategy.md`
- **RBAC Matrix:** `farmers-module/FARMERS_MODULE_RBAC_MATRIX.md`

---

## Questions?

Contact the farmers-module team for clarification on role permissions or resource definitions.

**Status:** Awaiting Implementation
**Blocked By:** AAA Service Team

---

## Example Implementation Pseudocode

```go
package catalog

import (
    pb "github.com/Kisanlink/aaa-service/pkg/proto"
    "context"
)

type CatalogServer struct {
    pb.UnimplementedCatalogServiceServer
    db     *gorm.DB
    logger *zap.Logger
}

func NewCatalogServer(db *gorm.DB, logger *zap.Logger) *CatalogServer {
    return &CatalogServer{db: db, logger: logger}
}

func (s *CatalogServer) SeedRolesAndPermissions(
    ctx context.Context,
    req *pb.SeedRolesAndPermissionsRequest,
) (*pb.SeedRolesAndPermissionsResponse, error) {

    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Check if roles already exist (unless force=true)
    if !req.Force {
        var count int64
        tx.Model(&Role{}).Where("scope = ?", "GLOBAL").Count(&count)
        if count > 0 {
            return &pb.SeedRolesAndPermissionsResponse{
                StatusCode: 200,
                Message: "Roles already exist. Use force=true to reseed.",
            }, nil
        }
    } else {
        // Delete existing GLOBAL roles
        tx.Where("scope = ?", "GLOBAL").Delete(&Role{})
    }

    // Create actions
    actions := createActions(tx)

    // Create resources
    resources := createResources(tx)

    // Create permissions (resource + action pairs)
    permissions := createPermissions(tx, resources, actions)

    // Create roles with permissions
    roles := createRolesWithPermissions(tx, permissions)

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, fmt.Errorf("failed to commit: %w", err)
    }

    return &pb.SeedRolesAndPermissionsResponse{
        StatusCode: 200,
        Message: "Successfully seeded roles and permissions",
        RolesCreated: int32(len(roles)),
        PermissionsCreated: int32(len(permissions)),
        ResourcesCreated: int32(len(resources)),
        ActionsCreated: int32(len(actions)),
        CreatedRoles: getRoleNames(roles),
    }, nil
}

// Helper functions
func createActions(tx *gorm.DB) []*Action { /* ... */ }
func createResources(tx *gorm.DB) []*Resource { /* ... */ }
func createPermissions(tx *gorm.DB, resources []*Resource, actions []*Action) []*Permission { /* ... */ }
func createRolesWithPermissions(tx *gorm.DB, permissions []*Permission) []*Role { /* ... */ }
func getRoleNames(roles []*Role) []string { /* ... */ }
```

---

**END OF REQUEST**
