# FPO Lifecycle Management System

## Overview

Complete lifecycle management system for Farmer Producer Organizations (FPOs) with automatic synchronization from AAA service, comprehensive audit trails, and robust error recovery.

## Quick Start

### 1. For Backend Developers
```bash
# Run database migration
psql -U username -d database < migrations/002_fpo_lifecycle_management.sql

# Build and deploy
go build ./...
```

### 2. For Frontend Developers
Start with: [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)

Quick Reference: [QUICK_REFERENCE.md](./QUICK_REFERENCE.md)

### 3. For Mobile Developers
See API endpoints and examples in: [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)

## Problem Solved

**Error**: `"failed to get FPO reference: no matching records found"`

**Solution**: Automatic sync from AAA service with new endpoints:
- `POST /identity/fpo/sync/:aaa_org_id` - Manual sync
- `GET /identity/fpo/by-org/:aaa_org_id` - Auto-sync on query

## Key Features

✅ **10 Lifecycle States** - Complete FPO lifecycle from draft to archive
✅ **Auto-Sync from AAA** - Never lose FPO references again
✅ **Setup Retry Logic** - Automatic retry for failed setups (max 3 attempts)
✅ **Complete Audit Trail** - Track all state changes with reasons
✅ **Admin Controls** - Suspend, reactivate, deactivate operations
✅ **Backward Compatible** - All existing endpoints still work

## Documentation

| Document | Description | Size |
|----------|-------------|------|
| [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md) | Complete integration guide with examples | 38KB |
| [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) | Copy-paste code snippets and quick tips | 7.7KB |
| [ARCHITECTURE_DESIGN.md](./ARCHITECTURE_DESIGN.md) | System architecture and design | 22KB |
| [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) | Implementation roadmap and tasks | 34KB |
| [STATE_DIAGRAM.md](./STATE_DIAGRAM.md) | State machine specification | 9.1KB |
| [ERROR_RECOVERY.md](./ERROR_RECOVERY.md) | Error handling and recovery | 23KB |
| [CHANGELOG.md](./CHANGELOG.md) | Complete changelog | This file |

## API Endpoints Summary

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/identity/fpo/sync/:aaa_org_id` | POST | Sync FPO from AAA |
| `/identity/fpo/by-org/:aaa_org_id` | GET | Get with auto-sync |
| `/identity/fpo/:id/retry-setup` | POST | Retry failed setup |
| `/identity/fpo/:id/suspend` | PUT | Suspend FPO |
| `/identity/fpo/:id/reactivate` | PUT | Reactivate FPO |
| `/identity/fpo/:id/deactivate` | DELETE | Deactivate FPO |
| `/identity/fpo/:id/history` | GET | Get audit trail |

## FPO Status Flow

```
DRAFT → PENDING_VERIFICATION → VERIFIED → PENDING_SETUP → ACTIVE
           ↓                      ↓            ↓
       REJECTED               ARCHIVED   SETUP_FAILED
                                            ↓ (retry)
                                       PENDING_SETUP

ACTIVE → SUSPENDED → ACTIVE
   ↓         ↓
INACTIVE ← ←
   ↓
ARCHIVED
```

## Quick Examples

### JavaScript/TypeScript
```typescript
// Get FPO with auto-sync
const fpo = await fetch(`/api/v1/identity/fpo/by-org/${orgId}`);

// Manual sync
await fetch(`/api/v1/identity/fpo/sync/${orgId}`, { method: 'POST' });

// Retry setup
await fetch(`/api/v1/identity/fpo/${fpoId}/retry-setup`, { method: 'POST' });
```

### cURL
```bash
# Sync FPO
curl -X POST https://api.example.com/api/v1/identity/fpo/sync/org_123 \
  -H "Authorization: Bearer TOKEN"

# Get audit history
curl https://api.example.com/api/v1/identity/fpo/FPOR_123/history \
  -H "Authorization: Bearer TOKEN"
```

## Migration Checklist

### Backend
- [x] Database schema updated
- [x] New endpoints implemented
- [x] Service layer complete
- [x] Repository pattern using kisanlink-db
- [x] Audit logging enabled
- [ ] Deploy to staging
- [ ] Deploy to production

### Frontend
- [ ] Update API service
- [ ] Add new status values
- [ ] Implement sync fallback
- [ ] Add retry UI
- [ ] Update status badges
- [ ] Test all scenarios

### Mobile
- [ ] Update API calls
- [ ] Add status enum
- [ ] Implement sync
- [ ] Add retry UI
- [ ] Update models
- [ ] Test offline behavior

## Testing

```bash
# Run tests
go test ./...

# Test specific package
go test ./internal/services -v

# Test with coverage
go test ./... -cover
```

## Support

- **Documentation**: See files in this directory
- **API Issues**: Check audit history and logs
- **Integration Help**: Review CLIENT_INTEGRATION_GUIDE.md
- **Emergency**: Use manual sync endpoint

## Version

**Current Version**: 1.0.0
**Release Date**: 2025-11-18
**Status**: ✅ Production Ready

## License

Internal - Kisanlink Platform

---

**Start Here**: [CLIENT_INTEGRATION_GUIDE.md](./CLIENT_INTEGRATION_GUIDE.md)
