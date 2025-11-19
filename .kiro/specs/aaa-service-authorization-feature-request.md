# AAA Service: Service-Level Authorization for Catalog Operations

**Document Type:** Feature Request
**Priority:** High
**Requested By:** Farmers Module Team
**Date:** 2025-11-19
**Target AAA Service Version:** v2.1.6+

---

## Executive Summary

The farmers-module requires the ability to seed roles and permissions in the AAA service via the `CatalogService.SeedRolesAndPermissions` RPC. Currently, this operation is failing with the error:

```
service 'Farmers Module' cannot seed default farmers-module roles
```

We are requesting the implementation of a **configuration-based service authorization system** that allows authorized services (like farmers-module) to perform catalog management operations.

---

## Problem Statement

### Current Behavior

When the farmers-module attempts to seed roles and permissions at startup or via the `/admin/seed` endpoint, the AAA service rejects the request with:

```
Error: service 'Farmers Module' cannot seed default farmers-module roles
```

**Impact:**
- Farmers-module cannot initialize required roles (farmer, kisansathi, CEO, fpo_manager, etc.)
- Manual intervention required for every deployment
- Farmer users cannot access their own data due to missing permissions
- Production deployments are blocked

### Root Cause

The AAA service's `CatalogService` has authorization checks that prevent external services from seeding roles, but there is no mechanism to whitelist authorized services via configuration.

---

## Proposed Solution

Implement a **configuration-based service authorization system** that:

1. Reads authorized service permissions from a YAML configuration file
2. Validates incoming gRPC requests against the configured permissions
3. Uses the `ServiceId` field from the request to identify the calling service
4. Allows granular permission control per service

---

## Implementation Specification

### 1. Configuration File Format

**Location:** `config/service_permissions.yaml` or embedded in main `config.yaml`

```yaml
# Service Authorization Configuration
service_authorization:
  enabled: true  # Global toggle for service authorization

  # List of authorized services and their permissions
  services:
    farmers-module:
      service_id: "farmers-module"
      display_name: "Farmers Module Service"
      description: "Farmer management and agricultural operations service"
      api_key_required: true  # Requires x-api-key header
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"

    erp-module:
      service_id: "erp-module"
      display_name: "ERP Module Service"
      description: "Enterprise resource planning service"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"

    admin-service:
      service_id: "admin-service"
      display_name: "Admin Service"
      description: "System administration service"
      api_key_required: true
      permissions:
        - "catalog:*"  # Wildcard for all catalog permissions
        - "user:*"
        - "organization:*"

# Fallback behavior when authorization is disabled
default_behavior:
  when_disabled: "allow_all"  # Options: "allow_all", "deny_all"
  log_unauthorized_attempts: true
```

### 2. Permission Naming Convention

Format: `<resource>:<action>`

**Catalog Permissions:**
- `catalog:seed_roles` - Seed roles and permissions
- `catalog:seed_permissions` - Seed permissions only
- `catalog:register_resource` - Register new resources
- `catalog:register_action` - Register new actions
- `catalog:list_resources` - List all resources
- `catalog:list_actions` - List all actions
- `catalog:*` - All catalog operations

**Other Permissions (Future):**
- `user:create` - Create users programmatically
- `organization:create` - Create organizations
- `role:assign` - Assign roles to users

### 3. Request Validation Flow

```
1. gRPC Request arrives (e.g., SeedRolesAndPermissionsRequest)
   ↓
2. Extract ServiceId from request (e.g., "farmers-module")
   ↓
3. Extract x-api-key from gRPC metadata
   ↓
4. Load service configuration from service_permissions.yaml
   ↓
5. Validate:
   - Is service_id registered?
   - Does x-api-key match configured key? (if required)
   - Does service have required permission? (e.g., "catalog:seed_roles")
   ↓
6. If validation passes → Execute operation
   If validation fails → Return gRPC error with details
```

### 4. Code Implementation (Pseudocode)

```go
// internal/authorization/service_authorizer.go

type ServiceAuthorizer struct {
    config ServiceAuthorizationConfig
}

func (sa *ServiceAuthorizer) Authorize(ctx context.Context, serviceId string, permission string) error {
    // 1. Check if authorization is enabled
    if !sa.config.Enabled {
        if sa.config.DefaultBehavior == "allow_all" {
            return nil
        }
        return errors.New("service authorization is disabled and default behavior is deny_all")
    }

    // 2. Lookup service configuration
    serviceConfig, exists := sa.config.Services[serviceId]
    if !exists {
        sa.logUnauthorizedAttempt(serviceId, permission)
        return fmt.Errorf("service '%s' is not authorized", serviceId)
    }

    // 3. Validate API key if required
    if serviceConfig.APIKeyRequired {
        apiKey, err := extractAPIKeyFromContext(ctx)
        if err != nil {
            return fmt.Errorf("x-api-key required but not provided")
        }

        if !sa.validateAPIKey(serviceId, apiKey) {
            return fmt.Errorf("invalid x-api-key for service '%s'", serviceId)
        }
    }

    // 4. Check permissions
    if !sa.hasPermission(serviceConfig.Permissions, permission) {
        sa.logUnauthorizedAttempt(serviceId, permission)
        return fmt.Errorf("service '%s' lacks permission '%s'", serviceId, permission)
    }

    return nil
}

func (sa *ServiceAuthorizer) hasPermission(servicePerms []string, requiredPerm string) bool {
    for _, perm := range servicePerms {
        // Exact match
        if perm == requiredPerm {
            return true
        }

        // Wildcard match (e.g., "catalog:*" matches "catalog:seed_roles")
        if strings.HasSuffix(perm, ":*") {
            prefix := strings.TrimSuffix(perm, ":*")
            if strings.HasPrefix(requiredPerm, prefix+":") {
                return true
            }
        }
    }
    return false
}
```

### 5. Integration with CatalogService

```go
// internal/services/catalog_service.go

func (s *CatalogService) SeedRolesAndPermissions(ctx context.Context, req *pb.SeedRolesAndPermissionsRequest) (*pb.SeedRolesAndPermissionsResponse, error) {
    // Extract service ID from request
    serviceId := req.ServiceId
    if serviceId == "" {
        serviceId = "unknown"
    }

    // Authorize the service
    if err := s.authorizer.Authorize(ctx, serviceId, "catalog:seed_roles"); err != nil {
        log.Printf("Authorization failed for service '%s': %v", serviceId, err)
        return &pb.SeedRolesAndPermissionsResponse{
            StatusCode: 403,
            Message:    fmt.Sprintf("Authorization failed: %v", err),
        }, nil
    }

    // Proceed with seeding...
    log.Printf("Service '%s' authorized to seed roles and permissions", serviceId)

    // ... existing seeding logic ...
}
```

---

## Configuration Management

### Development Environment

**File:** `config/service_permissions.dev.yaml`

```yaml
service_authorization:
  enabled: false  # Disabled for local development

default_behavior:
  when_disabled: "allow_all"
  log_unauthorized_attempts: true
```

### Staging Environment

**File:** `config/service_permissions.staging.yaml`

```yaml
service_authorization:
  enabled: true

  services:
    farmers-module:
      service_id: "farmers-module"
      display_name: "Farmers Module Service (Staging)"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"
```

### Production Environment

**File:** `config/service_permissions.prod.yaml`

```yaml
service_authorization:
  enabled: true  # Always enabled in production

  services:
    farmers-module:
      service_id: "farmers-module"
      display_name: "Farmers Module Service"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"
        - "catalog:register_resource"
        - "catalog:register_action"

    erp-module:
      service_id: "erp-module"
      display_name: "ERP Module Service"
      api_key_required: true
      permissions:
        - "catalog:seed_roles"
        - "catalog:seed_permissions"

default_behavior:
  when_disabled: "deny_all"  # Fail secure
  log_unauthorized_attempts: true
```

---

## Testing Plan

### Unit Tests

```go
func TestServiceAuthorizer_Authorize(t *testing.T) {
    tests := []struct {
        name           string
        serviceId      string
        permission     string
        expectedError  bool
    }{
        {
            name:          "farmers-module can seed roles",
            serviceId:     "farmers-module",
            permission:    "catalog:seed_roles",
            expectedError: false,
        },
        {
            name:          "farmers-module cannot create users",
            serviceId:     "farmers-module",
            permission:    "user:create",
            expectedError: true,
        },
        {
            name:          "unknown service is rejected",
            serviceId:     "malicious-service",
            permission:    "catalog:seed_roles",
            expectedError: true,
        },
        {
            name:          "admin-service has wildcard access",
            serviceId:     "admin-service",
            permission:    "catalog:anything",
            expectedError: false,
        },
    }

    // ... test implementation ...
}
```

### Integration Tests

1. **Test farmers-module seeding:**
   ```bash
   # Should succeed
   grpcurl -d '{"force": true, "service_id": "farmers-module"}' \
     -H "x-api-key: ${FARMERS_MODULE_API_KEY}" \
     localhost:50051 \
     pb.CatalogService/SeedRolesAndPermissions
   ```

2. **Test unauthorized service:**
   ```bash
   # Should fail with 403
   grpcurl -d '{"force": true, "service_id": "evil-service"}' \
     localhost:50051 \
     pb.CatalogService/SeedRolesAndPermissions
   ```

3. **Test missing API key:**
   ```bash
   # Should fail with "x-api-key required"
   grpcurl -d '{"force": true, "service_id": "farmers-module"}' \
     localhost:50051 \
     pb.CatalogService/SeedRolesAndPermissions
   ```

---

## Security Considerations

### 1. API Key Management

- **Storage:** Store API keys in secrets manager (AWS Secrets Manager, HashiCorp Vault)
- **Rotation:** Implement key rotation policy (every 90 days)
- **Hashing:** Store hashed keys in configuration, compare hashes

**Example:**
```yaml
services:
  farmers-module:
    service_id: "farmers-module"
    api_key_hash: "sha256:a1b2c3d4..."  # Hash of actual API key
    api_key_source: "aws-secrets-manager://farmers-module/api-key"
```

### 2. Audit Logging

Log all authorization attempts:

```json
{
  "timestamp": "2025-11-19T13:15:00Z",
  "event_type": "service_authorization",
  "service_id": "farmers-module",
  "permission": "catalog:seed_roles",
  "result": "allowed",
  "source_ip": "10.0.1.50",
  "request_id": "req-abc123"
}
```

### 3. Rate Limiting

Implement rate limiting per service:

```yaml
services:
  farmers-module:
    rate_limit:
      requests_per_minute: 10
      burst: 20
```

### 4. Monitoring Alerts

Set up alerts for:
- Unauthorized access attempts
- Repeated failures from known services (possible compromise)
- New services attempting to connect

---

## Migration Plan

### Phase 1: Implementation (Week 1)
- [ ] Implement `ServiceAuthorizer` component
- [ ] Add configuration file loading
- [ ] Write unit tests
- [ ] Update CatalogService to use authorizer

### Phase 2: Testing (Week 2)
- [ ] Deploy to development environment
- [ ] Run integration tests
- [ ] Test with farmers-module locally

### Phase 3: Staging Deployment (Week 3)
- [ ] Deploy to staging with `enabled: false`
- [ ] Monitor logs for authorization patterns
- [ ] Enable authorization in staging
- [ ] Validate all services work correctly

### Phase 4: Production Rollout (Week 4)
- [ ] Deploy to production with `enabled: false`
- [ ] Run in audit-only mode for 48 hours
- [ ] Enable authorization in production
- [ ] Monitor for 1 week

---

## Rollback Plan

If issues occur after enabling service authorization:

1. **Immediate:** Set `enabled: false` in configuration
2. **Quick:** Set `default_behavior: allow_all`
3. **Restart:** Restart AAA service to reload config

No code changes required for rollback.

---

## Impact Analysis

### Affected Services

| Service | Impact | Action Required |
|---------|--------|----------------|
| farmers-module | HIGH | Add to whitelist immediately |
| erp-module | MEDIUM | Add to whitelist before ERP deployment |
| admin-service | LOW | Add to whitelist (optional) |
| legacy-services | NONE | Continue working (when disabled) |

### Performance Impact

- **Authorization check overhead:** ~1-5ms per request
- **Config loading:** Once at startup
- **Memory footprint:** ~50KB for config data

### Breaking Changes

**None** - Feature is backward compatible when `enabled: false`

---

## Required Farmers-Module Changes

The farmers-module has already been updated to send the `ServiceId` field:

**File:** `internal/clients/aaa/aaa_client.go:960-962`
```go
grpcReq := &pb.SeedRolesAndPermissionsRequest{
    Force:     force,
    ServiceId: "farmers-module", // Explicitly set
}
```

No additional changes required on farmers-module side.

---

## Success Criteria

- [ ] Farmers-module can successfully seed roles at startup
- [ ] Unauthorized services are blocked with clear error messages
- [ ] Configuration can be updated without code changes
- [ ] All integration tests pass
- [ ] Zero false positives in production
- [ ] Audit logs capture all authorization attempts

---

## Questions & Answers

### Q: What if a service doesn't send ServiceId?
**A:** Default to `"unknown"` and deny access unless explicitly configured.

### Q: Can we use database instead of YAML?
**A:** Yes, but YAML is simpler for initial implementation. Database can be added later.

### Q: What about service-to-service authentication?
**A:** This feature focuses on authorization (what can you do). Authentication (who are you) is already handled by x-api-key.

### Q: How do we rotate API keys?
**A:** Update the key in secrets manager, reload AAA config (SIGHUP or API endpoint), update client services gradually.

---

## Contact Information

**Feature Requested By:**
Farmers Module Team

**Technical Contact:**
[Your Name/Email]

**AAA Service Maintainers:**
[AAA Team Contact]

**Timeline:**
**Requested:** 2025-11-19
**Needed By:** 2025-11-26 (1 week)
**Priority:** High (blocking production deployment)

---

## Appendix: Example Error Messages

### Unauthorized Service
```
Error: service 'evil-service' is not authorized to perform catalog operations
Request ID: req-abc123
Timestamp: 2025-11-19T13:15:00Z
```

### Missing Permission
```
Error: service 'farmers-module' lacks permission 'catalog:delete_roles'
Allowed permissions: catalog:seed_roles, catalog:seed_permissions
Request ID: req-def456
```

### Missing API Key
```
Error: x-api-key header is required for service 'farmers-module'
Request ID: req-ghi789
```

---

## References

- AAA Service Protobuf: `pkg/proto/catalog.proto`
- Farmers Module Implementation: `internal/clients/aaa/aaa_client.go`
- gRPC Metadata: https://grpc.io/docs/guides/metadata/
- YAML Configuration: https://yaml.org/spec/1.2/spec.html
