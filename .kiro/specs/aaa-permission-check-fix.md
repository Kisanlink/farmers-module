# AAA Permission Check Fix - Design Specification

## Problem Summary

When AAA service is enabled (`AAA_ENABLED=true`), multiple API endpoints fail with:

```json
{
  "error": "failed to check permission: missing permission parameters"
}
```

**Root Cause**: Services use optional request parameters (from query params or request body) as the subject for permission checks, instead of extracting the authenticated user ID from the JWT token context set by the auth middleware.

## Root Cause Analysis

### The Design Flaw

The code conflates two separate concerns:

| Concept | What It Means | Should Come From | Example Value |
|---------|---------------|------------------|---------------|
| **Authentication Subject** | WHO is making the request | JWT token (auth middleware) | `user_abc123` |
| **Filter Parameter** | WHOSE resources to show | Query parameter (optional) | `?farmer_id=farmer_xyz` |

### Current Broken Flow

1. **Client Request**
   ```
   GET /api/v1/farms?org_id=fpo_123
   Authorization: Bearer <jwt_token>
   ```

2. **Auth Middleware** (auth.go:95-174)
   - ✅ Validates JWT token
   - ✅ Extracts user info: `user_id = "user_abc123"`
   - ✅ Stores in Gin context: `c.Set("user_context", userContext)`
   - ❌ Does NOT store in request context for services

3. **Handler** (farm_handlers.go:185-189)
   - Parses OPTIONAL query param: `?farmer_id=xxx`
   - If NOT provided → `req.AAAFarmerUserID = ""` (empty)

4. **Service** (farm_service.go:242)
   - ❌ Calls: `CheckPermission(ctx, req.AAAFarmerUserID, ...)`
   - ❌ Uses empty string as subject (when no query param)

5. **AAA Client** (aaa_client.go:714-717)
   - ❌ Validates: `if subject == "" → return error`
   - ❌ Returns: `"missing permission parameters"`

### Code Evidence

**AAA Client Validation** (aaa_client.go:714-717):
```go
// Validate input parameters
if subject == "" || resource == "" || action == "" {
    log.Printf("Warning: Missing permission parameters")
    return false, fmt.Errorf("missing permission parameters")
}
```

**FarmService Permission Check** (farm_service.go:242):
```go
// Check AAA permission for farm.list
hasPermission, err := s.aaaService.CheckPermission(ctx, listReq.AAAFarmerUserID, "farm", "list", "", listReq.AAAOrgID)
```
Problem: `listReq.AAAFarmerUserID` comes from optional query parameter

**Handler Query Parameter Parsing** (farm_handlers.go:185-189):
```go
// Parse query parameters
if farmerID := c.Query("farmer_id"); farmerID != "" {
    req.AAAFarmerUserID = farmerID
}
```
Problem: `farmer_id` is optional filter parameter, not authentication subject

## Affected Files

### Services with Broken Permission Checks

| File | Methods | Lines |
|------|---------|-------|
| `farm_service.go` | CreateFarm, UpdateFarm, DeleteFarm, ListFarms, GetFarmByID, ListFarmsByBoundingBox | 46, 126, 218, 242, 313, 331 |
| `farm_activity_service.go` | CreateActivity, CompleteActivity, UpdateActivity, ListActivities | 44, 114, 173, 230 |
| `crop_cycle_service.go` | StartCycle, UpdateCycle, EndCycle, ListCycles | 36, 103, 181, 247 |
| `data_quality_service.go` | ValidateGeometry, ReconcileAAALinks, RebuildSpatialIndexes, DetectFarmOverlaps | 48, 182, 286, 386 |
| `reporting_service.go` | ExportFarmerPortfolio, OrgDashboardCounters | 39, 125 |

### Infrastructure Files Requiring Changes

| File | Purpose |
|------|---------|
| `internal/middleware/auth.go` | Update to store user context in request context |
| `internal/auth/context.go` | Add helper function to extract authenticated user ID |

## Solution Design

### Principle

**ALWAYS use authenticated user (from JWT) as the permission check subject**
**NEVER use request parameters (query params or body fields) as the subject**

### Implementation Pattern

#### Before (WRONG)
```go
// ❌ Using request parameter as subject
hasPermission, err := s.aaaService.CheckPermission(ctx, req.AAAFarmerUserID, "farm", "list", "", req.AAAOrgID)
```

#### After (CORRECT)
```go
// ✅ Extract authenticated user from context
userCtx, err := auth.GetUserFromContext(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get user context: %w", err)
}

// ✅ Use authenticated user as subject
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "list", "", req.AAAOrgID)
if err != nil {
    return nil, fmt.Errorf("failed to check permission: %w", err)
}
if !hasPermission {
    return nil, common.ErrForbidden
}

// ✅ Use request parameters for filtering only
filters := make(map[string]interface{})
if req.AAAFarmerUserID != "" {
    filters["aaa_farmer_user_id"] = req.AAAFarmerUserID
}
```

### Detailed Changes

#### 1. Update Auth Middleware (auth.go)

**Current Code** (lines 152-159):
```go
// Store contexts in Gin context
c.Set("user_context", userContext)
c.Set("org_context", orgContext)
c.Set("token", token)

// Store token in Request context for downstream services (e.g., gRPC calls)
ctx = auth.SetTokenInContext(ctx, token)
c.Request = c.Request.WithContext(ctx)
```

**New Code**:
```go
// Store contexts in Gin context
c.Set("user_context", userContext)
c.Set("org_context", orgContext)
c.Set("token", token)

// Store user context and token in Request context for downstream services
ctx = auth.SetUserInContext(ctx, userContext)
if orgContext != nil {
    ctx = auth.SetOrgInContext(ctx, orgContext)
}
ctx = auth.SetTokenInContext(ctx, token)
c.Request = c.Request.WithContext(ctx)
```

#### 2. Add Helper Function (auth/context.go)

Add convenience function to extract authenticated user ID:

```go
// GetAuthenticatedUserID extracts the authenticated user ID from context
// Returns empty string if user context is not found
func GetAuthenticatedUserID(ctx context.Context) (string, error) {
    user, err := GetUserFromContext(ctx)
    if err != nil {
        return "", err
    }
    return user.AAAUserID, nil
}

// GetAuthenticatedOrgID extracts the authenticated user's organization ID from context
// Returns empty string if org context is not found
func GetAuthenticatedOrgID(ctx context.Context) string {
    org, err := GetOrgFromContext(ctx)
    if err != nil {
        return ""
    }
    return org.AAAOrgID
}
```

#### 3. Update Service Methods

For each service method with permission checks, apply this pattern:

```go
// Extract authenticated user from context
userCtx, err := auth.GetUserFromContext(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get user context: %w", err)
}

// Use authenticated user for permission check
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, resource, action, objectID, orgID)
if err != nil {
    return nil, fmt.Errorf("failed to check permission: %w", err)
}
if !hasPermission {
    return nil, common.ErrForbidden
}
```

### Special Cases

#### Case 1: Create Operations (farm_service.go:46)

**Current**:
```go
hasPermission, err := s.aaaService.CheckPermission(ctx, createReq.AAAFarmerUserID, "farm", "create", createReq.AAAOrgID, createReq.AAAOrgID)
```

**Issue**: `createReq.AAAFarmerUserID` is the farmer for whom the farm is being created (from request body), not the authenticated user

**Fix**:
```go
// Extract authenticated user
userCtx, err := auth.GetUserFromContext(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get user context: %w", err)
}

// Check if authenticated user can create farms
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "create", "", createReq.AAAOrgID)
```

#### Case 2: Update/Delete Operations (farm_service.go:126, 218)

**Current**:
```go
hasPermission, err := s.aaaService.CheckPermission(ctx, existingFarm.AAAFarmerUserID, "farm", "update", existingFarm.ID, existingFarm.AAAOrgID)
```

**Issue**: `existingFarm.AAAFarmerUserID` is the farmer who owns the farm (from database), not the authenticated user

**Fix**:
```go
// Extract authenticated user
userCtx, err := auth.GetUserFromContext(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get user context: %w", err)
}

// Check if authenticated user can update this farm
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "update", existingFarm.ID, existingFarm.AAAOrgID)
```

#### Case 3: List Operations (farm_service.go:242, 331)

**Current**:
```go
hasPermission, err := s.aaaService.CheckPermission(ctx, listReq.AAAFarmerUserID, "farm", "list", "", listReq.AAAOrgID)
```

**Issue**: `listReq.AAAFarmerUserID` is an optional filter parameter (from query param), not the authenticated user

**Fix**:
```go
// Extract authenticated user
userCtx, err := auth.GetUserFromContext(ctx)
if err != nil {
    return nil, fmt.Errorf("failed to get user context: %w", err)
}

// Check if authenticated user can list farms
hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farm", "list", "", listReq.AAAOrgID)
if err != nil {
    return nil, fmt.Errorf("failed to check permission: %w", err)
}
if !hasPermission {
    return nil, common.ErrForbidden
}

// Use AAAFarmerUserID for filtering only
filters := make(map[string]interface{})
if listReq.AAAFarmerUserID != "" {
    filters["aaa_farmer_user_id"] = listReq.AAAFarmerUserID
}
```

## Testing Strategy

### Manual Testing Scenarios

1. **CEO Lists All Farms**
   ```bash
   # Request
   GET /api/v1/farms?org_id=fpo_123
   Authorization: Bearer <ceo_token>

   # Expected: Success - CEO can see all farms in organization
   ```

2. **Farmer Lists Own Farms**
   ```bash
   # Request
   GET /api/v1/farms?farmer_id=farmer_456
   Authorization: Bearer <farmer_token>

   # Expected: Success - Farmer sees only their farms
   ```

3. **Farmer Without Filter**
   ```bash
   # Request
   GET /api/v1/farms
   Authorization: Bearer <farmer_token>

   # Expected: Success - Should work without query param
   ```

### Unit Test Updates

Update existing service tests to:
1. Mock `auth.GetUserFromContext()` to return test user context
2. Verify permission checks use authenticated user ID, not request parameters
3. Test error cases when user context is missing

## Migration Notes

### Breaking Changes

None - This is a bug fix that makes the system work as originally intended

### Deployment Notes

1. This fix is backward compatible
2. No database migrations required
3. No API contract changes
4. Deploy during normal maintenance window

## Success Criteria

- [ ] All permission checks use authenticated user from JWT token context
- [ ] No permission checks use request parameters as subject
- [ ] All affected endpoints return 200 OK instead of permission errors
- [ ] Request parameters (farmer_id, org_id) still work as filters
- [ ] All existing tests pass
- [ ] Manual testing scenarios pass

## Files Changed Summary

1. `internal/middleware/auth.go` - Store user context in request context
2. `internal/auth/context.go` - Add helper functions
3. `internal/services/farm_service.go` - Fix 6 methods
4. `internal/services/farm_activity_service.go` - Fix 4 methods
5. `internal/services/crop_cycle_service.go` - Fix 4 methods
6. `internal/services/data_quality_service.go` - Fix 4 methods
7. `internal/services/reporting_service.go` - Fix 2 methods
8. Any other services with similar patterns

## References

- Business Rules: `.kiro/steering/business-rules.md`
- Tech Stack: `.kiro/steering/tech.md`
- AAA Integration: Check existing AAA service documentation
