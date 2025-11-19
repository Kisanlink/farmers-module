# AAA Service Authorization - Quick Start Guide

> **Status:** ✅ AAA Service Authorization is LIVE!
>
> The farmers-module can now seed roles and permissions securely.

---

## 🚀 5-Minute Setup

### 1. Generate API Key

```bash
openssl rand -base64 32
```

Copy the output (e.g., `8X3mP9kL2nQ5vR7wY1zA4bC6dE0fG8hI9jK2lM3nO5p=`)

### 2. Add to .env File

```bash
# Add this line to your .env file
AAA_API_KEY=8X3mP9kL2nQ5vR7wY1zA4bC6dE0fG8hI9jK2lM3nO5p=
```

### 3. Restart Farmers Module

```bash
# Stop the service (Ctrl+C if running)
# Start it again
go run cmd/farmers-service/main.go
```

### 4. Verify Success

Look for this in the startup logs:

```
✅ Successfully seeded AAA roles and permissions
```

**That's it!** The farmers-module can now seed roles automatically.

---

## 🔧 Configuration Reference

### Required Environment Variable

```bash
# Service-to-service authentication key
AAA_API_KEY=your-generated-api-key-here
```

### Optional Environment Variables

```bash
# AAA service endpoint (default: localhost:50051)
AAA_GRPC_ADDR=localhost:50051

# Enable/disable AAA integration (default: true)
AAA_ENABLED=true

# Request timeout (default: 5s)
AAA_REQUEST_TIMEOUT=5s

# Retry attempts (default: 3)
AAA_RETRY_ATTEMPTS=3
```

---

## 🐛 Troubleshooting

### Problem: "service 'farmers-module' is not authorized"

**Solution:**
- Check that AAA service has farmers-module in its authorized services list
- Contact AAA service team to add authorization

### Problem: "invalid x-api-key for service 'farmers-module'"

**Solution:**
- API key mismatch between farmers-module and AAA service
- Regenerate key and update both services:
  ```bash
  openssl rand -base64 32
  # Update .env in farmers-module
  # Update AAA service config
  ```

### Problem: "x-api-key header is required"

**Solution:**
- `AAA_API_KEY` environment variable not set
- Add it to your `.env` file:
  ```bash
  echo "AAA_API_KEY=your-key-here" >> .env
  ```

### Problem: Startup warning but service runs

```
Warning: Failed to seed AAA roles and permissions: ...
Application will continue, but role assignments may fail if roles don't exist
```

**Solution:**
- This is non-fatal; roles may already exist
- Manually trigger seeding:
  ```bash
  curl -X POST http://localhost:8080/api/v1/admin/seed \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"force": true}'
  ```

---

## 📖 Detailed Documentation

For complete documentation, see:

1. **Setup Guide:** `.kiro/specs/aaa-service-authorization-setup-guide.md`
2. **Feature Request:** `.kiro/specs/aaa-service-authorization-feature-request.md`
3. **Self-Access Architecture:** `.kiro/specs/self-access-authorization-architecture.md`

---

## 🔑 Manual Seeding (If Needed)

If automatic seeding at startup fails, trigger it manually:

```bash
# Get an admin JWT token first
TOKEN="your-admin-jwt-token"

# Seed roles with force=true (overwrites existing)
curl -X POST http://localhost:8080/api/v1/admin/seed \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"force": true}'
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Roles and permissions re-seeded successfully (forced)",
  "duration": "1.2s"
}
```

---

## 🔐 Security Reminders

- ✅ **DO:** Store API keys in secrets manager (AWS, Vault)
- ✅ **DO:** Use different keys per environment (dev/staging/prod)
- ✅ **DO:** Rotate keys every 90 days
- ❌ **DON'T:** Commit API keys to git
- ❌ **DON'T:** Share keys between environments
- ❌ **DON'T:** Log the full key value

---

## 💡 What This Fixes

### Before:
```
❌ Farmers with "farmer" role couldn't view their own data
❌ Permission denied errors for legitimate users
❌ Manual role seeding required after every deployment
```

### After:
```
✅ Automatic role seeding at startup
✅ Farmers can view their own data
✅ Proper permissions assigned to all roles
✅ Secure service-to-service authentication
```

---

## 📞 Support

**Issues?**
- Check logs: `logs/farmers-module.log`
- Review this guide
- See detailed docs in `.kiro/specs/`

**Still stuck?**
- GitHub Issues: https://github.com/Kisanlink/farmers-module/issues
- Team Slack: #farmers-module-support

---

**Last Updated:** 2025-11-19
