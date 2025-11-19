# AAA Service Authorization - Farmers Module Setup Guide

**Status:** ✅ AAA Service Implementation Complete
**Date:** 2025-11-19
**Version:** 1.0

---

## Overview

The AAA service now implements **service-level authorization** for catalog operations. This allows the farmers-module to seed roles and permissions securely using API key authentication.

### What Changed

Previously:
```
❌ farmers-module → AAA service → "service 'Farmers Module' cannot seed roles"
```

Now:
```
✅ farmers-module → x-api-key header → AAA service → Authorization OK → Seed roles
```

---

## Quick Setup (5 Minutes)

### Step 1: Generate API Key

```bash
# Generate a secure API key for farmers-module
export AAA_API_KEY=$(openssl rand -base64 32)

# Display the key (save it securely!)
echo "AAA_API_KEY=${AAA_API_KEY}"
```

**Example output:**
```
AAA_API_KEY=8X3mP9kL2nQ5vR7wY1zA4bC6dE0fG8hI9jK2lM3nO5p=
```

### Step 2: Configure Farmers Module

Add to your `.env` file or environment:

```bash
# AAA Service API Key (for service-to-service authentication)
AAA_API_KEY=8X3mP9kL2nQ5vR7wY1zA4bC6dE0fG8hI9jK2lM3nO5p=
```

**Location:** `/Users/kaushik/farmers-module/.env`

### Step 3: Configure AAA Service

The AAA service should already have this in `config/service_permissions.yaml`:

```yaml
service_authorization:
  enabled: true

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
```

**Note:** The AAA service must have the same API key configured in its secrets/configuration.

### Step 4: Restart Services

```bash
# 1. Restart AAA service (if you made config changes)
# (Follow AAA service restart procedure)

# 2. Restart farmers-module
cd /Users/kaushik/farmers-module
go run cmd/farmers-service/main.go
```

### Step 5: Verify

Check the farmers-module startup logs:

```
✅ Success:
2025/11/19 13:15:00 Seeding AAA roles and permissions...
2025/11/19 13:15:01 AAA Client: Added x-api-key to request for method: /pb.CatalogService/SeedRolesAndPermissions
2025/11/19 13:15:02 Successfully seeded AAA roles and permissions

❌ Failure (missing/wrong API key):
2025/11/19 13:15:00 Seeding AAA roles and permissions...
2025/11/19 13:15:01 Warning: Failed to seed AAA roles and permissions: invalid x-api-key for service 'farmers-module'
```

---

## Environment-Specific Configuration

### Development Environment

**File:** `.env.development`

```bash
# Development - Authorization disabled in AAA
AAA_API_KEY=dev-key-not-validated
AAA_GRPC_ADDR=localhost:50051
AAA_ENABLED=true
```

**AAA Config:** `service_permissions.dev.yaml` with `enabled: false`

### Staging Environment

**File:** `.env.staging`

```bash
# Staging - Real API key required
AAA_API_KEY=${AAA_SERVICE_API_KEY_STAGING}  # From secrets manager
AAA_GRPC_ADDR=aaa-service-staging.internal:50051
AAA_ENABLED=true
```

### Production Environment

**File:** `.env.production`

```bash
# Production - Real API key from AWS Secrets Manager
AAA_API_KEY=${AAA_SERVICE_API_KEY_PRODUCTION}
AAA_GRPC_ADDR=aaa-service.internal:50051
AAA_ENABLED=true
```

**Best Practice:** Use AWS Secrets Manager or Vault:

```bash
# Fetch from AWS Secrets Manager
export AAA_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id farmers-module/aaa-api-key \
  --query SecretString \
  --output text)
```

---

## Configuration Reference

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AAA_API_KEY` | Yes* | `""` | API key for service authentication |
| `AAA_GRPC_ADDR` | Yes | `localhost:50051` | AAA service gRPC endpoint |
| `AAA_ENABLED` | No | `true` | Enable/disable AAA integration |
| `AAA_REQUEST_TIMEOUT` | No | `5s` | Timeout for AAA requests |
| `AAA_RETRY_ATTEMPTS` | No | `3` | Retry attempts for failed requests |

*Required in production; optional in development if AAA authorization is disabled.

### Code Configuration

The API key is automatically sent in gRPC metadata by the `apiKeyInterceptor`:

**File:** `internal/clients/aaa/aaa_client.go:158-182`

```go
func apiKeyInterceptor(apiKey string) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        if apiKey != "" {
            md, ok := metadata.FromOutgoingContext(ctx)
            if !ok {
                md = metadata.New(nil)
            } else {
                md = md.Copy()
            }

            // Add x-api-key header
            md.Set("x-api-key", apiKey)
            ctx = metadata.NewOutgoingContext(ctx, md)

            log.Printf("AAA Client: Added x-api-key to request for method: %s", method)
        }

        return invoker(ctx, method, req, reply, cc, opts...)
    }
}
```

**ServiceId is set in seed request:**

**File:** `internal/clients/aaa/aaa_client.go:960-962`

```go
grpcReq := &pb.SeedRolesAndPermissionsRequest{
    Force:     force,
    ServiceId: "farmers-module", // Must match AAA config
}
```

---

## Manual Role Seeding

If automatic seeding fails at startup, you can manually trigger it:

### Via HTTP API

```bash
# Get admin JWT token
TOKEN="your-admin-jwt-token"

# Seed roles (force=true to overwrite existing)
curl -X POST http://localhost:8080/api/v1/admin/seed \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "force": true
  }'
```

**Expected Response:**

```json
{
  "success": true,
  "message": "Roles and permissions re-seeded successfully (forced)",
  "duration": "1.245s",
  "timestamp": "2025-11-19T13:15:00Z"
}
```

### Via gRPC Direct

```bash
# Direct gRPC call (requires grpcurl)
grpcurl -d '{
  "force": true,
  "service_id": "farmers-module"
}' \
  -H "x-api-key: ${AAA_API_KEY}" \
  localhost:50051 \
  pb.CatalogService/SeedRolesAndPermissions
```

---

## Troubleshooting

### Error: "service 'farmers-module' is not authorized"

**Cause:** AAA service doesn't have farmers-module in authorized services list.

**Solution:**
1. Check AAA's `config/service_permissions.yaml`
2. Ensure `farmers-module` entry exists
3. Restart AAA service

### Error: "invalid x-api-key for service 'farmers-module'"

**Cause:** API key mismatch between farmers-module and AAA service.

**Solution:**
1. Generate new API key: `openssl rand -base64 32`
2. Update farmers-module `.env`: `AAA_API_KEY=<new-key>`
3. Update AAA service configuration with same key
4. Restart both services

### Error: "x-api-key header is required"

**Cause:** `AAA_API_KEY` environment variable not set in farmers-module.

**Solution:**
```bash
# Set the variable
export AAA_API_KEY="your-api-key-here"

# Or add to .env file
echo "AAA_API_KEY=your-api-key-here" >> .env
```

### Warning: "Failed to seed AAA roles and permissions"

**Non-fatal - Service continues:**
- Roles may already exist (normal on restart)
- Use `/admin/seed` endpoint to manually reseed

**Check logs for specific error:**
```bash
# Search for detailed error
grep "SeedRolesAndPermissions" logs/farmers-module.log
```

---

## Security Best Practices

### 1. API Key Management

✅ **DO:**
- Use secrets manager (AWS Secrets Manager, HashiCorp Vault)
- Rotate keys every 90 days
- Use different keys per environment
- Never commit keys to git

❌ **DON'T:**
- Hardcode keys in code
- Share keys between environments
- Log the full key value
- Store keys in plain text files

### 2. Key Rotation Process

```bash
# 1. Generate new key
NEW_KEY=$(openssl rand -base64 32)

# 2. Update AAA service config (add new key alongside old)
# Allow both keys temporarily

# 3. Update farmers-module environment
export AAA_API_KEY="${NEW_KEY}"

# 4. Restart farmers-module (rolling restart in production)

# 5. Remove old key from AAA config after 24 hours

# 6. Store new key in secrets manager
aws secretsmanager update-secret \
  --secret-id farmers-module/aaa-api-key \
  --secret-string "${NEW_KEY}"
```

### 3. Monitoring

Set up alerts for:
- Failed authentication attempts (possible key compromise)
- Unauthorized access attempts
- Repeated 403 errors from farmers-module

**CloudWatch/Grafana Alert Example:**
```
Alert: AAA Authorization Failures
Condition: sum(aaa_auth_failures{service="farmers-module"}) > 5
Duration: 5 minutes
Action: Page on-call engineer
```

---

## Testing

### Local Development

```bash
# 1. Start AAA service (development mode, auth disabled)
cd ../aaa-service
export AAA_ENV=development
make run

# 2. Start farmers-module
cd /Users/kaushik/farmers-module
export AAA_API_KEY=dev-key-any-value
go run cmd/farmers-service/main.go

# 3. Check logs for successful seeding
# Expected: "Successfully seeded AAA roles and permissions"
```

### Integration Test

Create test script: `scripts/test-aaa-auth.sh`

```bash
#!/bin/bash
set -e

echo "Testing AAA service authorization..."

# 1. Test with correct API key
echo "Test 1: Valid API key"
grpcurl -d '{"force": true, "service_id": "farmers-module"}' \
  -H "x-api-key: ${AAA_API_KEY}" \
  localhost:50051 \
  pb.CatalogService/SeedRolesAndPermissions

if [ $? -eq 0 ]; then
  echo "✅ Test 1 PASSED: Valid API key accepted"
else
  echo "❌ Test 1 FAILED: Valid API key rejected"
  exit 1
fi

# 2. Test with invalid API key
echo "Test 2: Invalid API key"
grpcurl -d '{"force": true, "service_id": "farmers-module"}' \
  -H "x-api-key: invalid-key" \
  localhost:50051 \
  pb.CatalogService/SeedRolesAndPermissions

if [ $? -ne 0 ]; then
  echo "✅ Test 2 PASSED: Invalid API key rejected"
else
  echo "❌ Test 2 FAILED: Invalid API key accepted"
  exit 1
fi

# 3. Test with missing API key
echo "Test 3: Missing API key"
grpcurl -d '{"force": true, "service_id": "farmers-module"}' \
  localhost:50051 \
  pb.CatalogService/SeedRolesAndPermissions

if [ $? -ne 0 ]; then
  echo "✅ Test 3 PASSED: Missing API key rejected"
else
  echo "❌ Test 3 FAILED: Missing API key accepted"
  exit 1
fi

echo ""
echo "✅ All tests passed!"
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] Generate production API key
- [ ] Store key in secrets manager (AWS/Vault)
- [ ] Update AAA service configuration
- [ ] Update farmers-module environment config
- [ ] Test in staging environment
- [ ] Document key location for ops team

### Deployment

- [ ] Deploy AAA service with new config
- [ ] Verify AAA service health
- [ ] Deploy farmers-module with API key
- [ ] Verify farmers-module startup logs
- [ ] Test manual seeding via `/admin/seed`
- [ ] Verify farmer role permissions work
- [ ] Monitor logs for 1 hour

### Post-Deployment

- [ ] Verify no authorization errors in logs
- [ ] Test farmer user can access own data
- [ ] Test farmer user cannot access others' data
- [ ] Set up monitoring alerts
- [ ] Schedule key rotation (90 days)
- [ ] Document incident response plan

---

## Incident Response

### If API Key is Compromised

1. **Immediate (< 5 minutes):**
   ```bash
   # Generate new key
   NEW_KEY=$(openssl rand -base64 32)

   # Update AAA config (remove old key)
   # Restart AAA service

   # Update farmers-module
   export AAA_API_KEY="${NEW_KEY}"
   # Rolling restart farmers-module pods
   ```

2. **Investigation (< 30 minutes):**
   - Check AAA audit logs for unauthorized access
   - Identify compromised operations
   - Assess data exposure

3. **Remediation (< 2 hours):**
   - Revoke compromised tokens
   - Review and update security policies
   - Notify stakeholders if needed

---

## References

- **AAA Service Authorization Docs:** `.kiro/specs/service-authorization/`
- **Farmers Module Config:** `internal/config/config.go`
- **AAA Client Implementation:** `internal/clients/aaa/aaa_client.go`
- **gRPC Metadata Spec:** https://grpc.io/docs/guides/metadata/

---

## Support

**For farmers-module issues:**
- Check logs: `logs/farmers-module.log`
- GitHub Issues: https://github.com/Kisanlink/farmers-module/issues

**For AAA service issues:**
- Check AAA service documentation
- Contact AAA service team

**Emergency Contact:**
- On-call: [Your on-call rotation]
- Slack: #farmers-module-support

---

**Last Updated:** 2025-11-19
**Next Review:** 2025-12-19 (Monthly)
