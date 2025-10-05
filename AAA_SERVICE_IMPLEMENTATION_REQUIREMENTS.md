# AAA Service gRPC Implementation Requirements

## Executive Summary

The farmers-module requires several gRPC services from the AAA service that are currently not implemented. This document provides comprehensive architectural requirements for implementing these missing services in the AAA service repository to ensure full integration with the farmers-module.

## Context

The farmers-module has been built with stub implementations for AAA service integration. These stubs need to be replaced with actual gRPC service implementations in the AAA service. The farmers-module expects the following services to be available:

1. **OrganizationService** - Managing organizations (FPOs, cooperatives)
2. **GroupService** - Managing user groups within organizations
3. **RoleService** - Managing role assignments to users
4. **PermissionService** - Managing permission assignments to groups
5. **CatalogService** - Managing role and permission definitions

## Architecture Overview

### Service Dependencies

```
┌─────────────────────────────────────────────────────────────────┐
│                         Farmers Module                          │
├─────────────────────────────────────────────────────────────────┤
│  - Farmer Management     - Farm Management                      │
│  - FPO Management        - Crop Cycle Management                │
│  - KisanSathi Assignment - Reporting                           │
└───────────────────┬─────────────────────────────────────────────┘
                    │ gRPC
┌───────────────────▼─────────────────────────────────────────────┐
│                         AAA Service                             │
├─────────────────────────────────────────────────────────────────┤
│  Existing Services:                                             │
│  ✓ UserServiceV2                                                │
│  ✓ AuthorizationService                                         │
│                                                                  │
│  Required New Services:                                         │
│  • OrganizationService                                          │
│  • GroupService                                                 │
│  • RoleService                                                  │
│  • PermissionService                                            │
│  • CatalogService                                               │
└──────────────────────────────────────────────────────────────────┘
```

## 1. Protocol Buffer Definitions

### 1.1 OrganizationService Proto

```protobuf
syntax = "proto3";

package aaa.v1;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service OrganizationService {
  // Create a new organization
  rpc CreateOrganization(CreateOrganizationRequest) returns (CreateOrganizationResponse) {
    option (google.api.http) = {
      post: "/api/v1/organizations"
      body: "*"
    };
  }

  // Get organization by ID
  rpc GetOrganization(GetOrganizationRequest) returns (GetOrganizationResponse) {
    option (google.api.http) = {
      get: "/api/v1/organizations/{org_id}"
    };
  }

  // Update organization details
  rpc UpdateOrganization(UpdateOrganizationRequest) returns (UpdateOrganizationResponse) {
    option (google.api.http) = {
      put: "/api/v1/organizations/{org_id}"
      body: "*"
    };
  }

  // List organizations with filters
  rpc ListOrganizations(ListOrganizationsRequest) returns (ListOrganizationsResponse) {
    option (google.api.http) = {
      get: "/api/v1/organizations"
    };
  }

  // Delete organization (soft delete)
  rpc DeleteOrganization(DeleteOrganizationRequest) returns (DeleteOrganizationResponse) {
    option (google.api.http) = {
      delete: "/api/v1/organizations/{org_id}"
    };
  }
}

message Organization {
  string id = 1;
  string name = 2;
  string description = 3;
  string type = 4; // FPO, COOPERATIVE, COMPANY, NGO
  string status = 5; // ACTIVE, INACTIVE, SUSPENDED
  string ceo_user_id = 6;
  map<string, string> metadata = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message CreateOrganizationRequest {
  string name = 1;
  string description = 2;
  string type = 3;
  string ceo_user_id = 4;
  map<string, string> metadata = 5;
}

message CreateOrganizationResponse {
  string org_id = 1;
  string name = 2;
  string status = 3;
  google.protobuf.Timestamp created_at = 4;
}

message GetOrganizationRequest {
  string org_id = 1;
}

message GetOrganizationResponse {
  Organization organization = 1;
}

message UpdateOrganizationRequest {
  string org_id = 1;
  string name = 2;
  string description = 3;
  string ceo_user_id = 4;
  map<string, string> metadata = 5;
}

message UpdateOrganizationResponse {
  bool success = 1;
  string message = 2;
  Organization organization = 3;
}

message ListOrganizationsRequest {
  string type_filter = 1;
  string status_filter = 2;
  int32 page = 3;
  int32 page_size = 4;
  string search_query = 5;
}

message ListOrganizationsResponse {
  repeated Organization organizations = 1;
  int32 total_count = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message DeleteOrganizationRequest {
  string org_id = 1;
}

message DeleteOrganizationResponse {
  bool success = 1;
  string message = 2;
}
```

### 1.2 GroupService Proto

```protobuf
syntax = "proto3";

package aaa.v1;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service GroupService {
  // Create a new user group
  rpc CreateUserGroup(CreateUserGroupRequest) returns (CreateUserGroupResponse) {
    option (google.api.http) = {
      post: "/api/v1/groups"
      body: "*"
    };
  }

  // Add user to group
  rpc AddUserToGroup(AddUserToGroupRequest) returns (AddUserToGroupResponse) {
    option (google.api.http) = {
      post: "/api/v1/groups/{group_id}/members"
      body: "*"
    };
  }

  // Remove user from group
  rpc RemoveUserFromGroup(RemoveUserFromGroupRequest) returns (RemoveUserFromGroupResponse) {
    option (google.api.http) = {
      delete: "/api/v1/groups/{group_id}/members/{user_id}"
    };
  }

  // List group members
  rpc ListGroupMembers(ListGroupMembersRequest) returns (ListGroupMembersResponse) {
    option (google.api.http) = {
      get: "/api/v1/groups/{group_id}/members"
    };
  }

  // Get user's groups
  rpc GetUserGroups(GetUserGroupsRequest) returns (GetUserGroupsResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}/groups"
    };
  }

  // Update group details
  rpc UpdateGroup(UpdateGroupRequest) returns (UpdateGroupResponse) {
    option (google.api.http) = {
      put: "/api/v1/groups/{group_id}"
      body: "*"
    };
  }

  // Delete group
  rpc DeleteGroup(DeleteGroupRequest) returns (DeleteGroupResponse) {
    option (google.api.http) = {
      delete: "/api/v1/groups/{group_id}"
    };
  }
}

message UserGroup {
  string id = 1;
  string name = 2;
  string description = 3;
  string org_id = 4;
  repeated string permissions = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message CreateUserGroupRequest {
  string name = 1;
  string description = 2;
  string org_id = 3;
  repeated string permissions = 4;
}

message CreateUserGroupResponse {
  string group_id = 1;
  string name = 2;
  string org_id = 3;
  google.protobuf.Timestamp created_at = 4;
}

message AddUserToGroupRequest {
  string user_id = 1;
  string group_id = 2;
}

message AddUserToGroupResponse {
  bool success = 1;
  string message = 2;
}

message RemoveUserFromGroupRequest {
  string user_id = 1;
  string group_id = 2;
}

message RemoveUserFromGroupResponse {
  bool success = 1;
  string message = 2;
}

message ListGroupMembersRequest {
  string group_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message ListGroupMembersResponse {
  repeated UserSummary members = 1;
  int32 total_count = 2;
}

message GetUserGroupsRequest {
  string user_id = 1;
}

message GetUserGroupsResponse {
  repeated UserGroup groups = 1;
}

message UpdateGroupRequest {
  string group_id = 1;
  string name = 2;
  string description = 3;
  repeated string permissions = 4;
}

message UpdateGroupResponse {
  bool success = 1;
  string message = 2;
  UserGroup group = 3;
}

message DeleteGroupRequest {
  string group_id = 1;
}

message DeleteGroupResponse {
  bool success = 1;
  string message = 2;
}

message UserSummary {
  string id = 1;
  string username = 2;
  string email = 3;
  string full_name = 4;
}
```

### 1.3 RoleService Proto

```protobuf
syntax = "proto3";

package aaa.v1;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service RoleService {
  // Assign role to user
  rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse) {
    option (google.api.http) = {
      post: "/api/v1/roles/assign"
      body: "*"
    };
  }

  // Check if user has role
  rpc CheckUserRole(CheckUserRoleRequest) returns (CheckUserRoleResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}/roles/{role_name}/check"
    };
  }

  // Remove role from user
  rpc RemoveRole(RemoveRoleRequest) returns (RemoveRoleResponse) {
    option (google.api.http) = {
      delete: "/api/v1/users/{user_id}/roles/{role_name}"
    };
  }

  // Get user's roles
  rpc GetUserRoles(GetUserRolesRequest) returns (GetUserRolesResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}/roles"
    };
  }

  // List users with specific role
  rpc ListUsersWithRole(ListUsersWithRoleRequest) returns (ListUsersWithRoleResponse) {
    option (google.api.http) = {
      get: "/api/v1/roles/{role_name}/users"
    };
  }
}

message Role {
  string id = 1;
  string name = 2;
  string description = 3;
  repeated string permissions = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
}

message AssignRoleRequest {
  string user_id = 1;
  string org_id = 2;
  string role_name = 3;
}

message AssignRoleResponse {
  bool success = 1;
  string message = 2;
}

message CheckUserRoleRequest {
  string user_id = 1;
  string role_name = 2;
  string org_id = 3; // optional, check role in specific org
}

message CheckUserRoleResponse {
  bool has_role = 1;
  string org_id = 2;
}

message RemoveRoleRequest {
  string user_id = 1;
  string org_id = 2;
  string role_name = 3;
}

message RemoveRoleResponse {
  bool success = 1;
  string message = 2;
}

message GetUserRolesRequest {
  string user_id = 1;
  string org_id = 2; // optional, filter by org
}

message GetUserRolesResponse {
  repeated UserRole roles = 1;
}

message UserRole {
  string role_name = 1;
  string org_id = 2;
  string org_name = 3;
  google.protobuf.Timestamp assigned_at = 4;
}

message ListUsersWithRoleRequest {
  string role_name = 1;
  string org_id = 2; // optional, filter by org
  int32 page = 3;
  int32 page_size = 4;
}

message ListUsersWithRoleResponse {
  repeated UserSummary users = 1;
  int32 total_count = 2;
}
```

### 1.4 PermissionService Proto

```protobuf
syntax = "proto3";

package aaa.v1;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service PermissionService {
  // Assign permission to group
  rpc AssignPermissionToGroup(AssignPermissionToGroupRequest) returns (AssignPermissionToGroupResponse) {
    option (google.api.http) = {
      post: "/api/v1/groups/{group_id}/permissions"
      body: "*"
    };
  }

  // Check if group has permission
  rpc CheckGroupPermission(CheckGroupPermissionRequest) returns (CheckGroupPermissionResponse) {
    option (google.api.http) = {
      get: "/api/v1/groups/{group_id}/permissions/check"
    };
  }

  // List group permissions
  rpc ListGroupPermissions(ListGroupPermissionsRequest) returns (ListGroupPermissionsResponse) {
    option (google.api.http) = {
      get: "/api/v1/groups/{group_id}/permissions"
    };
  }

  // Remove permission from group
  rpc RemovePermissionFromGroup(RemovePermissionFromGroupRequest) returns (RemovePermissionFromGroupResponse) {
    option (google.api.http) = {
      delete: "/api/v1/groups/{group_id}/permissions"
    };
  }

  // Check user's effective permissions
  rpc GetUserEffectivePermissions(GetUserEffectivePermissionsRequest) returns (GetUserEffectivePermissionsResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}/permissions/effective"
    };
  }
}

message Permission {
  string id = 1;
  string resource = 2;
  string action = 3;
  string description = 4;
  google.protobuf.Timestamp created_at = 5;
}

message AssignPermissionToGroupRequest {
  string group_id = 1;
  string resource = 2;
  string action = 3;
}

message AssignPermissionToGroupResponse {
  bool success = 1;
  string message = 2;
}

message CheckGroupPermissionRequest {
  string group_id = 1;
  string resource = 2;
  string action = 3;
}

message CheckGroupPermissionResponse {
  bool has_permission = 1;
}

message ListGroupPermissionsRequest {
  string group_id = 1;
}

message ListGroupPermissionsResponse {
  repeated Permission permissions = 1;
}

message RemovePermissionFromGroupRequest {
  string group_id = 1;
  string resource = 2;
  string action = 3;
}

message RemovePermissionFromGroupResponse {
  bool success = 1;
  string message = 2;
}

message GetUserEffectivePermissionsRequest {
  string user_id = 1;
  string org_id = 2; // optional, filter by org
}

message GetUserEffectivePermissionsResponse {
  repeated Permission permissions = 1;
  repeated string roles = 2;
  repeated string groups = 3;
}
```

### 1.5 CatalogService Proto

```protobuf
syntax = "proto3";

package aaa.v1;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service CatalogService {
  // Seed default roles and permissions
  rpc SeedRolesAndPermissions(SeedRolesAndPermissionsRequest) returns (SeedRolesAndPermissionsResponse) {
    option (google.api.http) = {
      post: "/api/v1/catalog/seed"
      body: "*"
    };
  }

  // Create new role
  rpc CreateRole(CreateRoleRequest) returns (CreateRoleResponse) {
    option (google.api.http) = {
      post: "/api/v1/catalog/roles"
      body: "*"
    };
  }

  // Create new permission
  rpc CreatePermission(CreatePermissionRequest) returns (CreatePermissionResponse) {
    option (google.api.http) = {
      post: "/api/v1/catalog/permissions"
      body: "*"
    };
  }

  // List available roles
  rpc ListRoles(ListRolesRequest) returns (ListRolesResponse) {
    option (google.api.http) = {
      get: "/api/v1/catalog/roles"
    };
  }

  // List available permissions
  rpc ListPermissions(ListPermissionsRequest) returns (ListPermissionsResponse) {
    option (google.api.http) = {
      get: "/api/v1/catalog/permissions"
    };
  }

  // Get role details
  rpc GetRole(GetRoleRequest) returns (GetRoleResponse) {
    option (google.api.http) = {
      get: "/api/v1/catalog/roles/{role_name}"
    };
  }

  // Update role
  rpc UpdateRole(UpdateRoleRequest) returns (UpdateRoleResponse) {
    option (google.api.http) = {
      put: "/api/v1/catalog/roles/{role_name}"
      body: "*"
    };
  }

  // Delete role
  rpc DeleteRole(DeleteRoleRequest) returns (DeleteRoleResponse) {
    option (google.api.http) = {
      delete: "/api/v1/catalog/roles/{role_name}"
    };
  }
}

message SeedRolesAndPermissionsRequest {
  bool force = 1; // Force reseed even if data exists
}

message SeedRolesAndPermissionsResponse {
  bool success = 1;
  int32 roles_created = 2;
  int32 permissions_created = 3;
  string message = 4;
}

message CreateRoleRequest {
  string name = 1;
  string description = 2;
  repeated string permissions = 3;
}

message CreateRoleResponse {
  string role_id = 1;
  string name = 2;
  google.protobuf.Timestamp created_at = 3;
}

message CreatePermissionRequest {
  string resource = 1;
  string action = 2;
  string description = 3;
}

message CreatePermissionResponse {
  string permission_id = 1;
  string resource = 2;
  string action = 3;
  google.protobuf.Timestamp created_at = 4;
}

message ListRolesRequest {
  int32 page = 1;
  int32 page_size = 2;
}

message ListRolesResponse {
  repeated Role roles = 1;
  int32 total_count = 2;
}

message ListPermissionsRequest {
  string resource_filter = 1;
  string action_filter = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message ListPermissionsResponse {
  repeated Permission permissions = 1;
  int32 total_count = 2;
}

message GetRoleRequest {
  string role_name = 1;
}

message GetRoleResponse {
  Role role = 1;
}

message UpdateRoleRequest {
  string role_name = 1;
  string description = 2;
  repeated string permissions = 3;
}

message UpdateRoleResponse {
  bool success = 1;
  string message = 2;
  Role role = 3;
}

message DeleteRoleRequest {
  string role_name = 1;
}

message DeleteRoleResponse {
  bool success = 1;
  string message = 2;
}
```

## 2. Database Schema Requirements

### 2.1 Organizations Table

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('FPO', 'COOPERATIVE', 'COMPANY', 'NGO')),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    ceo_user_id UUID REFERENCES users(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_type ON organizations(type);
CREATE INDEX idx_organizations_status ON organizations(status);
CREATE INDEX idx_organizations_ceo ON organizations(ceo_user_id);
```

### 2.2 User Groups Table

```sql
CREATE TABLE user_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    org_id UUID NOT NULL REFERENCES organizations(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(name, org_id)
);

CREATE INDEX idx_user_groups_org ON user_groups(org_id);
```

### 2.3 User Group Memberships Table

```sql
CREATE TABLE user_group_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    group_id UUID NOT NULL REFERENCES user_groups(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    UNIQUE(user_id, group_id)
);

CREATE INDEX idx_group_memberships_user ON user_group_memberships(user_id);
CREATE INDEX idx_group_memberships_group ON user_group_memberships(group_id);
```

### 2.4 Roles Catalog Table

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_roles_name ON roles(name);
```

### 2.5 Permissions Catalog Table

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
```

### 2.6 Role Permissions Table

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
```

### 2.7 User Role Assignments Table

```sql
CREATE TABLE user_role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    org_id UUID NOT NULL REFERENCES organizations(id),
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    UNIQUE(user_id, role_id, org_id)
);

CREATE INDEX idx_user_roles_user ON user_role_assignments(user_id);
CREATE INDEX idx_user_roles_role ON user_role_assignments(role_id);
CREATE INDEX idx_user_roles_org ON user_role_assignments(org_id);
```

### 2.8 Group Permissions Table

```sql
CREATE TABLE group_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES user_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    UNIQUE(group_id, permission_id)
);

CREATE INDEX idx_group_permissions_group ON group_permissions(group_id);
CREATE INDEX idx_group_permissions_permission ON group_permissions(permission_id);
```

## 3. Security Considerations

### 3.1 Authentication & Authorization

1. **Service-to-Service Authentication**
   - Use mutual TLS for gRPC connections
   - Implement service account tokens for inter-service calls
   - Validate JWT tokens from farmers-module

2. **Permission Model**
   - Implement RBAC (Role-Based Access Control)
   - Support resource-level permissions (e.g., farmers:create, farms:read)
   - Implement self-service permissions with :self suffix (e.g., farmers:update:self)
   - Cache permission checks for performance

3. **Data Isolation**
   - Enforce organization-level data isolation
   - Implement row-level security where applicable
   - Audit all permission changes

### 3.2 Input Validation

```go
// Example validation for CreateOrganization
func ValidateCreateOrganizationRequest(req *pb.CreateOrganizationRequest) error {
    if req.Name == "" {
        return status.Error(codes.InvalidArgument, "organization name is required")
    }
    if len(req.Name) > 255 {
        return status.Error(codes.InvalidArgument, "organization name too long")
    }
    if !isValidOrgType(req.Type) {
        return status.Error(codes.InvalidArgument, "invalid organization type")
    }
    // Validate CEO user exists
    if req.CeoUserId != "" {
        if !userExists(req.CeoUserId) {
            return status.Error(codes.NotFound, "CEO user not found")
        }
    }
    return nil
}
```

### 3.3 Audit Logging

```go
// Audit log structure
type AuditLog struct {
    ID          string    `json:"id"`
    Timestamp   time.Time `json:"timestamp"`
    Service     string    `json:"service"`
    Method      string    `json:"method"`
    UserID      string    `json:"user_id"`
    OrgID       string    `json:"org_id"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Result      string    `json:"result"`
    Details     map[string]interface{} `json:"details"`
}
```

## 4. Implementation Guidelines

### 4.1 Service Implementation Pattern

```go
package services

import (
    "context"
    "github.com/your-org/aaa-service/pkg/proto"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "gorm.io/gorm"
)

type OrganizationServiceImpl struct {
    proto.UnimplementedOrganizationServiceServer
    db *gorm.DB
    cache CacheInterface
    logger LoggerInterface
}

func NewOrganizationService(db *gorm.DB, cache CacheInterface, logger LoggerInterface) *OrganizationServiceImpl {
    return &OrganizationServiceImpl{
        db: db,
        cache: cache,
        logger: logger,
    }
}

func (s *OrganizationServiceImpl) CreateOrganization(ctx context.Context, req *proto.CreateOrganizationRequest) (*proto.CreateOrganizationResponse, error) {
    // 1. Validate request
    if err := ValidateCreateOrganizationRequest(req); err != nil {
        return nil, err
    }

    // 2. Check permissions
    if !hasPermission(ctx, "organizations", "create") {
        return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
    }

    // 3. Begin transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 4. Create organization
    org := &models.Organization{
        Name:        req.Name,
        Description: req.Description,
        Type:        req.Type,
        CEOUserID:   req.CeoUserId,
        Metadata:    req.Metadata,
        Status:      "ACTIVE",
    }

    if err := tx.Create(org).Error; err != nil {
        tx.Rollback()
        return nil, status.Error(codes.Internal, "failed to create organization")
    }

    // 5. Create audit log
    s.auditLog(ctx, "CreateOrganization", org.ID, "SUCCESS")

    // 6. Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, status.Error(codes.Internal, "failed to commit transaction")
    }

    // 7. Invalidate cache
    s.cache.Delete("org:" + org.ID)

    // 8. Return response
    return &proto.CreateOrganizationResponse{
        OrgId:     org.ID,
        Name:      org.Name,
        Status:    org.Status,
        CreatedAt: timestamppb.New(org.CreatedAt),
    }, nil
}
```

### 4.2 Default Roles and Permissions

```go
var DefaultRoles = []Role{
    {
        Name:        "admin",
        Description: "Administrator with full access",
        Permissions: []string{
            "farmers:create", "farmers:read", "farmers:update", "farmers:delete", "farmers:list",
            "farms:create", "farms:read", "farms:update", "farms:delete", "farms:list",
            "fpos:create", "fpos:read", "fpos:update", "fpos:delete", "fpos:list",
            "users:create", "users:read", "users:update", "users:delete", "users:list",
            "reports:read", "reports:generate",
            "admin:maintain",
        },
    },
    {
        Name:        "fpo_manager",
        Description: "FPO Manager with organization management access",
        Permissions: []string{
            "farmers:create", "farmers:read", "farmers:update", "farmers:list",
            "farms:create", "farms:read", "farms:update", "farms:list",
            "fpos:read", "fpos:update",
            "reports:read",
        },
    },
    {
        Name:        "kisansathi",
        Description: "Field agent with farmer management access",
        Permissions: []string{
            "farmers:create", "farmers:read", "farmers:update",
            "farms:create", "farms:read", "farms:update",
        },
    },
    {
        Name:        "farmer",
        Description: "Farmer with self-service access",
        Permissions: []string{
            "farmers:read:self", "farmers:update:self",
            "farms:read:self", "farms:update:self", "farms:create:self",
        },
    },
    {
        Name:        "readonly",
        Description: "Read-only access to all resources",
        Permissions: []string{
            "farmers:read", "farmers:list",
            "farms:read", "farms:list",
            "fpos:read", "fpos:list",
            "reports:read",
        },
    },
}
```

### 4.3 Error Handling

```go
func MapDatabaseError(err error) error {
    if err == nil {
        return nil
    }

    if errors.Is(err, gorm.ErrRecordNotFound) {
        return status.Error(codes.NotFound, "resource not found")
    }

    if strings.Contains(err.Error(), "duplicate key") {
        return status.Error(codes.AlreadyExists, "resource already exists")
    }

    if strings.Contains(err.Error(), "foreign key") {
        return status.Error(codes.FailedPrecondition, "related resource constraint violation")
    }

    return status.Error(codes.Internal, "database operation failed")
}
```

## 5. Testing Requirements

### 5.1 Unit Tests

```go
func TestOrganizationService_CreateOrganization(t *testing.T) {
    tests := []struct {
        name    string
        request *pb.CreateOrganizationRequest
        want    *pb.CreateOrganizationResponse
        wantErr bool
    }{
        {
            name: "Valid FPO creation",
            request: &pb.CreateOrganizationRequest{
                Name:        "Test FPO",
                Description: "Test FPO Description",
                Type:        "FPO",
                CeoUserId:   "user123",
            },
            wantErr: false,
        },
        {
            name: "Missing name",
            request: &pb.CreateOrganizationRequest{
                Description: "Test Description",
                Type:        "FPO",
            },
            wantErr: true,
        },
        {
            name: "Invalid type",
            request: &pb.CreateOrganizationRequest{
                Name: "Test Org",
                Type: "INVALID",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 5.2 Integration Tests

```go
func TestIntegration_RolePermissionFlow(t *testing.T) {
    // 1. Create organization
    org := createTestOrganization(t)

    // 2. Create user group
    group := createTestGroup(t, org.ID)

    // 3. Add user to group
    addUserToGroup(t, testUserID, group.ID)

    // 4. Assign permission to group
    assignPermissionToGroup(t, group.ID, "farmers", "create")

    // 5. Verify user has permission through group
    hasPermission := checkUserPermission(t, testUserID, "farmers", "create")
    assert.True(t, hasPermission)

    // 6. Remove user from group
    removeUserFromGroup(t, testUserID, group.ID)

    // 7. Verify permission revoked
    hasPermission = checkUserPermission(t, testUserID, "farmers", "create")
    assert.False(t, hasPermission)
}
```

### 5.3 Performance Tests

```go
func BenchmarkPermissionCheck(b *testing.B) {
    // Setup
    ctx := context.Background()
    service := setupService()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.CheckPermission(ctx, "user123", "farmers", "read", "farm456", "org789")
    }
}
```

## 6. Performance Considerations

### 6.1 Caching Strategy

```go
type CacheKey struct {
    Type   string // "permission", "role", "group", "org"
    UserID string
    OrgID  string
    Extra  string // resource:action for permissions
}

// Cache TTLs
const (
    PermissionCacheTTL = 5 * time.Minute
    RoleCacheTTL       = 10 * time.Minute
    GroupCacheTTL      = 10 * time.Minute
    OrgCacheTTL        = 30 * time.Minute
)
```

### 6.2 Database Optimization

1. **Indexes**: All foreign keys and frequently queried columns
2. **Connection Pooling**: Configure appropriate pool size
3. **Query Optimization**: Use batch operations where possible
4. **Pagination**: Implement cursor-based pagination for large datasets

### 6.3 gRPC Optimization

```go
// Connection pooling configuration
type GRPCPoolConfig struct {
    MaxConnections     int
    MaxIdleConnections int
    ConnectionTimeout  time.Duration
    KeepAliveInterval  time.Duration
}

// Recommended settings
var DefaultPoolConfig = GRPCPoolConfig{
    MaxConnections:     100,
    MaxIdleConnections: 10,
    ConnectionTimeout:  10 * time.Second,
    KeepAliveInterval:  30 * time.Second,
}
```

## 7. Monitoring & Observability

### 7.1 Metrics

```go
// Prometheus metrics
var (
    grpcRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "aaa_grpc_request_duration_seconds",
            Help: "gRPC request duration in seconds",
        },
        []string{"service", "method", "status"},
    )

    permissionCheckCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aaa_permission_checks_total",
            Help: "Total number of permission checks",
        },
        []string{"resource", "action", "result"},
    )

    cacheHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "aaa_cache_hit_rate",
            Help: "Cache hit rate percentage",
        },
        []string{"cache_type"},
    )
)
```

### 7.2 Logging

```go
// Structured logging example
logger.Info("Organization created",
    zap.String("org_id", org.ID),
    zap.String("org_name", org.Name),
    zap.String("org_type", org.Type),
    zap.String("created_by", userID),
    zap.Duration("duration", time.Since(start)),
)
```

### 7.3 Tracing

```go
// OpenTelemetry tracing
func (s *Service) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
    ctx, span := tracer.Start(ctx, "OrganizationService.CreateOrganization")
    defer span.End()

    span.SetAttributes(
        attribute.String("org.name", req.Name),
        attribute.String("org.type", req.Type),
    )

    // Implementation...
}
```

## 8. Migration Plan

### Phase 1: Service Implementation (Week 1-2)
1. Implement proto definitions
2. Create database migrations
3. Implement service logic
4. Write unit tests

### Phase 2: Integration (Week 3)
1. Deploy services to staging
2. Test with farmers-module
3. Performance testing
4. Security audit

### Phase 3: Production Rollout (Week 4)
1. Deploy to production with feature flags
2. Monitor metrics and logs
3. Gradual rollout
4. Full activation

## 9. Compatibility Requirements

### 9.1 Backward Compatibility
- Maintain existing UserServiceV2 and AuthorizationService interfaces
- Version new services appropriately (v1)
- Support graceful degradation

### 9.2 Forward Compatibility
- Design extensible proto messages
- Use semantic versioning
- Implement feature flags for new capabilities

## 10. Documentation Requirements

### 10.1 API Documentation
- Generate OpenAPI specs from proto definitions
- Provide example requests and responses
- Document error codes and handling

### 10.2 Integration Guide
- Step-by-step integration instructions
- Code examples in multiple languages
- Troubleshooting guide

### 10.3 Operations Manual
- Deployment procedures
- Configuration options
- Monitoring and alerting setup

## Appendix A: Example Integration Code

### Farmers Module Integration

```go
// Example of how farmers-module will use the new services
func (s *FarmerService) CreateFarmerWithRole(ctx context.Context, farmer *models.Farmer) error {
    // 1. Create user in AAA
    userResp, err := s.aaaClient.CreateUser(ctx, &aaa.CreateUserRequest{
        Username:    farmer.Username,
        PhoneNumber: farmer.PhoneNumber,
        Email:       farmer.Email,
        FullName:    farmer.FullName,
    })
    if err != nil {
        return err
    }

    // 2. Assign farmer role
    err = s.aaaClient.AssignRole(ctx, userResp.UserID, farmer.OrgID, "farmer")
    if err != nil {
        return err
    }

    // 3. Create farmer record
    farmer.UserID = userResp.UserID
    return s.db.Create(farmer).Error
}
```

## Appendix B: Testing Checklist

- [ ] Unit tests for all service methods
- [ ] Integration tests for cross-service flows
- [ ] Performance benchmarks
- [ ] Security penetration testing
- [ ] Load testing (1000+ concurrent users)
- [ ] Disaster recovery testing
- [ ] API compatibility testing
- [ ] Client SDK testing

## Conclusion

This comprehensive specification provides all necessary details for implementing the missing gRPC services in the AAA service. The implementation should follow the patterns and guidelines outlined here to ensure compatibility with the farmers-module and maintain high standards for security, performance, and reliability.

For questions or clarifications, please contact the farmers-module team or raise issues in the project repository.
