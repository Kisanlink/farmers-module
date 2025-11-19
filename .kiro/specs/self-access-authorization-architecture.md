# Self-Access Authorization Architecture

## Executive Summary

This document outlines the architecture for implementing self-access authorization in the farmers-module backend service. The design ensures users can always access their own data regardless of role permissions, while maintaining backward compatibility with the existing AAA permission system.

## Problem Statement

Current authorization middleware blocks users from accessing their own data when their role lacks specific permissions:
- A user with "farmer" role cannot view their own farmer profile (requires `farmer.read`)
- Farmers cannot list their own farms (requires `farm.list`)
- Similar restrictions apply to all entity types

This violates the principle that users should always have access to their own data.

## Architectural Decision

### Approach: Hybrid Middleware + Service Layer Pattern

After analyzing the alternatives, the recommended approach is a **hybrid solution** that combines:
1. **Middleware enhancement** for early detection and bypass
2. **Service layer enforcement** for data filtering
3. **Repository pattern integration** for automatic ownership filtering

This approach provides defense-in-depth while maintaining clean separation of concerns.

## Detailed Design

### 1. Core Components

#### 1.1 Self-Access Detection Middleware

```go
// internal/middleware/self_access.go
package middleware

import (
    "github.com/Kisanlink/farmers-module/internal/auth"
    "github.com/gin-gonic/gin"
)

type SelfAccessDetector struct {
    patterns []ResourcePattern
}

type ResourcePattern struct {
    Resource string
    PathParam string
    BodyField string
    QueryParam string
}

// DetectSelfAccess analyzes the request to determine if it's for own data
func (d *SelfAccessDetector) DetectSelfAccess(c *gin.Context) (*SelfAccessContext, error) {
    userContext, _ := c.Get("user_context")
    user := userContext.(*auth.UserContext)

    // Extract resource from route
    permission, exists := auth.GetPermissionForRoute(c.Request.Method, c.Request.URL.Path)
    if !exists {
        return nil, nil
    }

    // Check path parameters
    if targetUserID := c.Param("aaa_user_id"); targetUserID != "" {
        return &SelfAccessContext{
            IsSelfAccess: targetUserID == user.AAAUserID,
            ResourceType: permission.Resource,
            TargetUserID: targetUserID,
            RequestingUserID: user.AAAUserID,
        }, nil
    }

    // Check farmer_id parameter (requires lookup)
    if farmerID := c.Param("farmer_id"); farmerID != "" {
        return &SelfAccessContext{
            ResourceType: permission.Resource,
            ResourceID: farmerID,
            RequestingUserID: user.AAAUserID,
            RequiresOwnershipCheck: true,
        }, nil
    }

    // For list operations, mark as potential self-access
    if permission.Action == "list" {
        return &SelfAccessContext{
            IsSelfAccess: false, // Will be enforced at service layer
            ResourceType: permission.Resource,
            RequestingUserID: user.AAAUserID,
            RequiresFiltering: true,
        }, nil
    }

    return nil, nil
}
```

#### 1.2 Enhanced Authorization Middleware

```go
// internal/middleware/auth.go - Enhanced AuthorizationMiddleware
func AuthorizationMiddleware(aaaService services.AAAService, selfAccessDetector *SelfAccessDetector, logger interfaces.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip for public routes
        if auth.IsPublicRoute(c.Request.Method, c.Request.URL.Path) {
            c.Next()
            return
        }

        // Get user context
        userContextInterface, exists := c.Get("user_context")
        if !exists {
            // Handle error...
            return
        }
        userContext := userContextInterface.(*auth.UserContext)

        // Detect self-access
        selfAccessCtx, err := selfAccessDetector.DetectSelfAccess(c)
        if err != nil {
            logger.Error("Failed to detect self-access", zap.Error(err))
            // Continue with normal authorization
        }

        // Store self-access context for downstream use
        if selfAccessCtx != nil {
            c.Set("self_access_context", selfAccessCtx)

            // If confirmed self-access, bypass permission check
            if selfAccessCtx.IsSelfAccess && !selfAccessCtx.RequiresOwnershipCheck {
                logger.Debug("Self-access detected, bypassing permission check",
                    zap.String("user_id", userContext.AAAUserID),
                    zap.String("resource", selfAccessCtx.ResourceType))
                c.Next()
                return
            }
        }

        // Get required permission for route
        permission, exists := auth.GetPermissionForRoute(c.Request.Method, c.Request.URL.Path)
        if !exists {
            c.Next()
            return
        }

        // Check with AAA service
        hasPermission, err := aaaService.CheckPermission(
            ctx,
            userContext.AAAUserID,
            permission.Resource,
            permission.Action,
            "",
            orgID,
        )

        // If no permission but requires ownership check, allow through to service layer
        if !hasPermission && selfAccessCtx != nil && selfAccessCtx.RequiresOwnershipCheck {
            logger.Debug("No permission but ownership check required, deferring to service layer",
                zap.String("user_id", userContext.AAAUserID),
                zap.String("resource", permission.Resource))
            c.Set("requires_ownership_validation", true)
            c.Next()
            return
        }

        if !hasPermission {
            // Return 403 Forbidden
            c.JSON(http.StatusForbidden, common.ErrorResponse{
                Error: "forbidden",
                Message: "Insufficient permissions to access this resource",
                Code: "AUTH_PERMISSION_DENIED",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 2. Service Layer Enhancement

#### 2.1 Base Service Pattern

```go
// internal/services/base_service.go
package services

import (
    "context"
    "fmt"
    "github.com/Kisanlink/farmers-module/internal/auth"
)

type OwnershipValidator interface {
    ValidateOwnership(ctx context.Context, resourceID string, userID string) (bool, error)
    GetResourceOwner(ctx context.Context, resourceID string) (string, error)
}

type SelfAccessEnforcer struct {
    ownershipValidators map[string]OwnershipValidator
}

func NewSelfAccessEnforcer() *SelfAccessEnforcer {
    return &SelfAccessEnforcer{
        ownershipValidators: make(map[string]OwnershipValidator),
    }
}

func (e *SelfAccessEnforcer) RegisterValidator(resourceType string, validator OwnershipValidator) {
    e.ownershipValidators[resourceType] = validator
}

func (e *SelfAccessEnforcer) EnforceAccess(ctx context.Context, resourceType, resourceID string) error {
    // Check if ownership validation is required
    if requiresValidation := ctx.Value("requires_ownership_validation"); requiresValidation == true {
        userContext, err := auth.GetUserFromContext(ctx)
        if err != nil {
            return fmt.Errorf("user context not found")
        }

        validator, exists := e.ownershipValidators[resourceType]
        if !exists {
            return fmt.Errorf("no ownership validator for resource type: %s", resourceType)
        }

        isOwner, err := validator.ValidateOwnership(ctx, resourceID, userContext.AAAUserID)
        if err != nil {
            return fmt.Errorf("ownership validation failed: %w", err)
        }

        if !isOwner {
            return fmt.Errorf("access denied: not the owner of this resource")
        }
    }

    return nil
}
```

#### 2.2 Enhanced Farmer Service

```go
// internal/services/farmer_service.go - Enhanced implementation
type FarmerServiceImpl struct {
    repository       interfaces.FarmerRepository
    selfAccessEnforcer *SelfAccessEnforcer
    // ... other fields
}

// Implement OwnershipValidator
func (s *FarmerServiceImpl) ValidateOwnership(ctx context.Context, farmerID string, userID string) (bool, error) {
    filter := base.NewFilterBuilder().
        Where("id", base.OpEqual, farmerID).
        Where("aaa_user_id", base.OpEqual, userID).
        Build()

    count, err := s.repository.Count(ctx, filter)
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

func (s *FarmerServiceImpl) GetResourceOwner(ctx context.Context, farmerID string) (string, error) {
    filter := base.NewFilterBuilder().
        Where("id", base.OpEqual, farmerID).
        Build()

    farmer, err := s.repository.FindOne(ctx, filter)
    if err != nil {
        return "", err
    }

    return farmer.AAAUserID, nil
}

// Enhanced GetFarmer method
func (s *FarmerServiceImpl) GetFarmer(ctx context.Context, req *requests.GetFarmerRequest) (*responses.FarmerProfileResponse, error) {
    // If farmer_id provided, check ownership
    if req.FarmerID != "" {
        if err := s.selfAccessEnforcer.EnforceAccess(ctx, "farmer", req.FarmerID); err != nil {
            return nil, fmt.Errorf("access denied: %w", err)
        }
    }

    // Check for self-access context
    selfAccessCtx := GetSelfAccessContext(ctx)
    if selfAccessCtx != nil && selfAccessCtx.IsSelfAccess {
        // Ensure we're only accessing own data
        if req.AAAUserID != "" && req.AAAUserID != selfAccessCtx.RequestingUserID {
            return nil, fmt.Errorf("access denied: cannot access other user's data")
        }
        // Force filter to own user ID
        req.AAAUserID = selfAccessCtx.RequestingUserID
    }

    // Continue with existing logic...
    var filterBuilder *base.FilterBuilder
    if req.FarmerID != "" {
        filterBuilder = base.NewFilterBuilder().
            Where("id", base.OpEqual, req.FarmerID)
    } else if req.AAAUserID != "" {
        filterBuilder = base.NewFilterBuilder().
            Where("aaa_user_id", base.OpEqual, req.AAAUserID)
        if req.AAAOrgID != "" {
            filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
        }
    }

    // Rest of the implementation...
}

// Enhanced ListFarmers method
func (s *FarmerServiceImpl) ListFarmers(ctx context.Context, req *requests.ListFarmersRequest) (*responses.FarmerListResponse, error) {
    filterBuilder := base.NewFilterBuilder()

    // Check for self-access context
    selfAccessCtx := GetSelfAccessContext(ctx)
    if selfAccessCtx != nil && selfAccessCtx.RequiresFiltering {
        // Check if user has permission to list all
        userContext, _ := auth.GetUserFromContext(ctx)
        hasListAllPermission := s.checkPermissionSilently(ctx, "farmer", "list")

        if !hasListAllPermission {
            // Force filter to own data only
            filterBuilder = filterBuilder.Where("aaa_user_id", base.OpEqual, userContext.AAAUserID)
        }
    }

    // Apply other filters from request...
    if req.AAAOrgID != "" {
        filterBuilder = filterBuilder.Where("aaa_org_id", base.OpEqual, req.AAAOrgID)
    }

    // Continue with existing logic...
}
```

#### 2.3 Enhanced Farm Service

```go
// internal/services/farm_service.go
type FarmServiceImpl struct {
    repository         interfaces.FarmRepository
    farmerRepository   interfaces.FarmerRepository
    selfAccessEnforcer *SelfAccessEnforcer
    // ... other fields
}

// Implement OwnershipValidator
func (s *FarmServiceImpl) ValidateOwnership(ctx context.Context, farmID string, userID string) (bool, error) {
    // Get farm with farmer details
    filter := base.NewFilterBuilder().
        Where("id", base.OpEqual, farmID).
        Preload("Farmer").
        Build()

    farm, err := s.repository.FindOne(ctx, filter)
    if err != nil {
        return false, err
    }

    // Check if the farmer's AAA user ID matches
    return farm.Farmer.AAAUserID == userID, nil
}

// Enhanced ListFarms method
func (s *FarmServiceImpl) ListFarms(ctx context.Context, req *requests.ListFarmsRequest) (*responses.FarmListResponse, error) {
    filterBuilder := base.NewFilterBuilder()

    // Check for self-access context
    selfAccessCtx := GetSelfAccessContext(ctx)
    userContext, _ := auth.GetUserFromContext(ctx)

    if selfAccessCtx != nil && selfAccessCtx.RequiresFiltering {
        hasListAllPermission := s.checkPermissionSilently(ctx, "farm", "list")

        if !hasListAllPermission {
            // Get farmer ID for the current user
            farmerFilter := base.NewFilterBuilder().
                Where("aaa_user_id", base.OpEqual, userContext.AAAUserID).
                Build()

            farmer, err := s.farmerRepository.FindOne(ctx, farmerFilter)
            if err != nil {
                // User is not a farmer, return empty list
                return &responses.FarmListResponse{
                    Farms: []*responses.FarmData{},
                    Total: 0,
                }, nil
            }

            // Filter farms to only those owned by this farmer
            filterBuilder = filterBuilder.Where("farmer_id", base.OpEqual, farmer.ID)
        }
    }

    // Apply other filters from request...
    if req.FarmerID != "" {
        filterBuilder = filterBuilder.Where("farmer_id", base.OpEqual, req.FarmerID)
    }

    // Continue with existing logic...
}
```

### 3. Context Utilities

```go
// internal/auth/self_access_context.go
package auth

import "context"

type SelfAccessContext struct {
    IsSelfAccess           bool
    ResourceType           string
    ResourceID             string
    TargetUserID           string
    RequestingUserID       string
    RequiresOwnershipCheck bool
    RequiresFiltering      bool
}

const selfAccessContextKey = contextKey("self_access")

func SetSelfAccessContext(ctx context.Context, selfAccess *SelfAccessContext) context.Context {
    return context.WithValue(ctx, selfAccessContextKey, selfAccess)
}

func GetSelfAccessContext(ctx context.Context) *SelfAccessContext {
    if val := ctx.Value(selfAccessContextKey); val != nil {
        return val.(*SelfAccessContext)
    }
    return nil
}
```

### 4. Configuration and Registration

```go
// internal/config/self_access_patterns.go
package config

var SelfAccessPatterns = []SelfAccessPattern{
    // Farmer patterns
    {
        RoutePattern: "GET /api/v1/identity/farmers/user/:aaa_user_id",
        ResourceType: "farmer",
        OwnerField: "aaa_user_id",
        PathParam: "aaa_user_id",
    },
    {
        RoutePattern: "GET /api/v1/identity/farmers/id/:farmer_id",
        ResourceType: "farmer",
        OwnerField: "aaa_user_id",
        PathParam: "farmer_id",
        RequiresLookup: true,
    },
    {
        RoutePattern: "PUT /api/v1/identity/farmers/user/:aaa_user_id",
        ResourceType: "farmer",
        OwnerField: "aaa_user_id",
        PathParam: "aaa_user_id",
    },

    // Farm patterns
    {
        RoutePattern: "GET /api/v1/farms",
        ResourceType: "farm",
        OwnerField: "farmer.aaa_user_id",
        RequiresFiltering: true,
    },
    {
        RoutePattern: "GET /api/v1/farms/:farm_id",
        ResourceType: "farm",
        OwnerField: "farmer.aaa_user_id",
        PathParam: "farm_id",
        RequiresLookup: true,
    },

    // Crop cycle patterns
    {
        RoutePattern: "GET /api/v1/crops/cycles",
        ResourceType: "cycle",
        OwnerField: "farm.farmer.aaa_user_id",
        RequiresFiltering: true,
    },
}
```

## Security Considerations

### 1. Attack Vectors and Mitigations

#### 1.1 Parameter Tampering
**Attack**: User modifies path/query parameters to access other users' data
**Mitigation**:
- Validate ownership at service layer
- Force filter to authenticated user's ID when self-access is detected
- Log all access attempts for audit

#### 1.2 Privilege Escalation
**Attack**: User with limited permissions tries to access admin functions
**Mitigation**:
- Self-access only applies to own data, not administrative functions
- Maintain strict permission checks for non-self operations
- Separate self-access patterns from role-based permissions

#### 1.3 Data Leakage
**Attack**: List operations revealing existence of other users' data
**Mitigation**:
- Automatic filtering in list operations when user lacks list permission
- No error differentiation between "not found" and "access denied"
- Consistent response times to prevent timing attacks

### 2. Audit and Monitoring

```go
// internal/middleware/audit.go
type SelfAccessAuditLogger struct {
    logger interfaces.Logger
}

func (a *SelfAccessAuditLogger) LogAccess(ctx context.Context, event SelfAccessEvent) {
    a.logger.Info("Self-access event",
        zap.String("user_id", event.UserID),
        zap.String("resource_type", event.ResourceType),
        zap.String("resource_id", event.ResourceID),
        zap.String("action", event.Action),
        zap.Bool("granted", event.Granted),
        zap.String("reason", event.Reason),
        zap.Time("timestamp", event.Timestamp),
    )
}
```

## Implementation Strategy

### Phase 1: Foundation (Week 1)
1. Implement SelfAccessContext and utilities
2. Create SelfAccessDetector middleware
3. Add audit logging infrastructure
4. Unit tests for detection logic

### Phase 2: Middleware Integration (Week 2)
1. Enhance AuthorizationMiddleware
2. Integrate SelfAccessDetector
3. Add configuration for self-access patterns
4. Integration tests for middleware

### Phase 3: Service Layer (Week 3-4)
1. Implement OwnershipValidator interface
2. Enhance FarmerService with ownership checks
3. Enhance FarmService with ownership checks
4. Update other services (CropCycle, Activity, etc.)
5. Service layer tests

### Phase 4: Testing and Rollout (Week 5)
1. End-to-end testing
2. Performance testing
3. Security audit
4. Gradual rollout with feature flags

## Migration Path

### 1. Feature Flag Control

```go
// internal/config/features.go
type FeatureFlags struct {
    SelfAccessEnabled bool `env:"FEATURE_SELF_ACCESS_ENABLED" default:"false"`
    SelfAccessAuditOnly bool `env:"FEATURE_SELF_ACCESS_AUDIT_ONLY" default:"true"`
}

// In middleware
if config.Features.SelfAccessEnabled {
    if config.Features.SelfAccessAuditOnly {
        // Log but don't enforce
        auditLogger.LogAccess(ctx, event)
    } else {
        // Full enforcement
        // ... apply self-access logic
    }
}
```

### 2. Backward Compatibility
- Existing permission system remains unchanged
- Self-access is additive, not replacing existing permissions
- Users with proper permissions bypass self-access checks
- No changes to AAA service required

### 3. Rollback Plan
1. Disable feature flag
2. Middleware reverts to original behavior
3. No database changes required
4. Audit logs retained for analysis

## Testing Strategy

### 1. Unit Tests

```go
// internal/middleware/self_access_test.go
func TestSelfAccessDetection(t *testing.T) {
    tests := []struct {
        name string
        route string
        userID string
        pathParam string
        expectedSelfAccess bool
    }{
        {
            name: "User accessing own farmer profile",
            route: "/api/v1/identity/farmers/user/USER123",
            userID: "USER123",
            pathParam: "USER123",
            expectedSelfAccess: true,
        },
        {
            name: "User accessing another's farmer profile",
            route: "/api/v1/identity/farmers/user/USER456",
            userID: "USER123",
            pathParam: "USER456",
            expectedSelfAccess: false,
        },
    }
    // ... test implementation
}
```

### 2. Integration Tests

```go
// internal/services/farmer_service_self_access_test.go
func TestFarmerServiceSelfAccess(t *testing.T) {
    // Test user can access own profile without farmer.read permission
    // Test user cannot access others' profiles without permission
    // Test user with permission can access any profile
}
```

### 3. E2E Tests

```go
// tests/e2e/self_access_test.go
func TestEndToEndSelfAccess(t *testing.T) {
    // Create user with farmer role
    // Verify can access own farmer profile
    // Verify can list own farms
    // Verify cannot access other farmers' data
    // Verify admin can access all data
}
```

## Performance Considerations

### 1. Caching Strategy
- Cache farmer-to-user mappings (TTL: 5 minutes)
- Cache ownership validation results (TTL: 1 minute)
- Invalidate on updates

### 2. Database Optimization
- Index on `aaa_user_id` in farmers table
- Composite index on `(farmer_id, aaa_user_id)` for ownership checks
- Query optimization for filtered list operations

### 3. Expected Performance Impact
- Minimal overhead for users with permissions (1 additional context check)
- ~10-20ms additional latency for ownership validation
- Negligible impact on list operations (already filtered)

## Monitoring and Observability

### 1. Metrics
```go
// Prometheus metrics
self_access_requests_total{resource, action, granted}
self_access_ownership_checks_total{resource, result}
self_access_latency_seconds{resource, action}
```

### 2. Dashboards
- Self-access usage by resource type
- Denial rate for ownership validation
- Performance impact analysis
- User access patterns

### 3. Alerts
- High denial rate (> 10%)
- Ownership validation failures
- Performance degradation
- Suspicious access patterns

## Edge Cases

### 1. Multi-Organization Users
- User belongs to multiple organizations
- Must specify org_id for disambiguation
- Default to first org if only one farmer profile exists

### 2. Delegated Access
- KisanSathi accessing farmer's data
- Maintain existing delegation logic
- Self-access applies to KisanSathi's own profile

### 3. Bulk Operations
- Filter bulk results to own data
- Maintain pagination consistency
- Clear messaging about filtered results

## Conclusion

This architecture provides a robust, secure, and performant solution for self-access authorization. The hybrid approach ensures:

1. **Security**: Multiple layers of validation prevent unauthorized access
2. **Performance**: Minimal overhead with intelligent caching
3. **Maintainability**: Clean separation of concerns
4. **Compatibility**: Full backward compatibility with existing systems
5. **Observability**: Comprehensive monitoring and audit trails

The phased implementation allows for gradual rollout with minimal risk, while feature flags enable quick rollback if issues arise. The solution scales horizontally and maintains the existing security posture while adding the critical self-access capability.
