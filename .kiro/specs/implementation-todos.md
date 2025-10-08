# Implementation TODO List - Farmers Module

**Generated**: 2025-10-06
**Status**: Production Readiness Assessment
**Priority Levels**: Critical | High | Medium | Low

---

## Executive Summary

This document identifies all stub implementations, missing features, and production readiness gaps in the farmers-module codebase. Items are prioritized by business impact and security/reliability concerns.

---

## 🔴 CRITICAL - Security & Authorization Issues

### 1. Missing Permission Checks
**Location**: `internal/handlers/bulk_farmer_handler.go:88`
```go
// TODO: Add proper permission check using AAA service
```
**Impact**: Unauthorized access to bulk operations could lead to data breaches
**Action**: Implement AAA permission verification before processing bulk operations
**Effort**: 2-4 hours

---

## 🟠 HIGH - Business Logic Implementation Gaps

### 1. FPO Admin Notification System
**Location**: `internal/services/data_quality_service.go:253`
```go
// TODO: Notify FPO admin of data inconsistency (Business Rule 6.2)
```
**Impact**: Data quality issues won't be communicated to administrators
**Action**: Implement notification service integration
**Effort**: 1 day

### 2. Farmer/AAA User ID Assignment
**Location**: `internal/services/bulk_farmer_service.go:200`
```go
// TODO: Set actual farmer and AAA user IDs
```
**Impact**: Bulk imported farmers won't have proper identity mapping
**Action**: Integrate with AAA service for user creation during bulk import
**Effort**: 2 days

### 3. Retry Logic for Bulk Operations
**Location**: `internal/services/bulk_farmer_service.go:451`
```go
// TODO: Implement retry logic
```
**Impact**: Failed bulk operations cannot be recovered automatically
**Action**: Implement exponential backoff retry mechanism
**Effort**: 1 day

### 4. File Download and Parsing for URLs
**Location**: `internal/services/bulk_farmer_service.go:469`
```go
// TODO: Download file from URL and parse
```
**Impact**: Cannot process files from external URLs
**Action**: Implement secure file download with validation
**Effort**: 1 day

---

## 🟡 MEDIUM - Service Implementation Stubs

### 1. Crop Cycle Service Implementations
**Locations**:
- `internal/handlers/crop_handlers.go:327` - Start cycle service call
- `internal/handlers/crop_handlers.go:385` - Update cycle service call
- `internal/handlers/crop_handlers.go:441` - End cycle service call
- `internal/handlers/crop_handlers.go:468` - Get cycle service call
- `internal/handlers/crop_handlers.go:501` - List cycles service call

**Impact**: Core crop cycle management features non-functional
**Action**: Complete crop cycle service implementation
**Effort**: 3-5 days total

### 2. Identity Handler - GetFPORef Implementation
**Location**: `internal/handlers/identity_handlers.go:210-221`
```go
// TODO: Implement the actual service call
// Currently returns placeholder data
```
**Impact**: FPO reference retrieval returns mock data
**Action**: Implement actual FPO reference service integration
**Effort**: 4 hours

### 3. Bulk Operation Validations
**Locations**:
- `internal/services/bulk_farmer_service.go:578` - Validation logic stub
- `internal/services/bulk_farmer_service.go:615` - Result file generation stub
- `internal/services/bulk_farmer_service.go:697` - Additional validation stub

**Impact**: Bulk operations lack proper validation and reporting
**Action**: Implement comprehensive validation rules
**Effort**: 2 days

### 4. Duplicate Checking Logic
**Location**: `internal/services/pipeline/stages.go:200-207`
```go
// TODO: Implement actual duplicate checking logic
// Placeholder logic - randomly mark some as duplicates for testing
```
**Impact**: No real duplicate detection in pipeline
**Action**: Implement phone/Aadhaar based duplicate detection
**Effort**: 1 day

### 5. Farmer ID Extraction from Response
**Location**: `internal/services/pipeline/stages.go:421-424`
```go
farmerID := "" // TODO: Extract from farmerResponse
// Using placeholder logic
```
**Impact**: Farmer IDs not properly tracked in pipeline
**Action**: Parse and extract actual farmer IDs from AAA response
**Effort**: 2 hours

---

## 🟢 LOW - Configuration & Hardcoded Values

### 1. Configurable Country Code
**Location**: `internal/services/farmer_service.go:80`
```go
"country_code": "+91", // TODO: Make configurable
```
**Impact**: Service limited to Indian phone numbers
**Action**: Move to configuration file
**Effort**: 30 minutes

### 2. Missing Aadhaar Number in Request
**Location**: `internal/services/farmer_service.go:81`
```go
"aadhaar_number": "", // TODO: Add to request if available
```
**Impact**: Aadhaar not captured during farmer registration
**Action**: Add optional Aadhaar field to request structure
**Effort**: 2 hours

### 3. Farm Data Loading
**Locations**:
- `internal/services/farmer_service.go:339`
- `internal/services/farmer_service.go:447`
- `internal/services/farmer_service.go:561`
```go
Farms: []*responses.FarmData{}, // TODO: Load actual farms
```
**Impact**: Farmer responses don't include associated farm data
**Action**: Implement farm repository integration
**Effort**: 1 day

---

## Test Mock Implementations (Not Production Code)

### Test File Stubs
**Locations**: Multiple occurrences in `internal/clients/aaa/aaa_client_test.go`
- Lines 103, 107, 111, 115, 119, 123, 127, 131, 135, 139, 156, 160, 164, 168, etc.
```go
return nil, nil
```
**Note**: These are in test files and don't impact production functionality

---

## Handler Annotation Issues

### Missing/Incomplete Swagger Annotations

1. **Bulk Operations Handler** (`internal/handlers/bulk_farmer_handler.go`)
   - Missing @Security annotations for authenticated endpoints
   - Incomplete error response documentation

2. **Crop Handlers** (`internal/handlers/crop_handlers.go`)
   - Need @Security annotations for all endpoints
   - Missing 401, 403 error response documentation

3. **Identity Handlers** (`internal/handlers/identity_handlers.go`)
   - GetFPORef handler missing complete annotations
   - Placeholder response structure needs documentation

---

## Implementation Priority Matrix

| Priority | Count | Estimated Effort | Business Impact |
|----------|-------|-----------------|-----------------|
| CRITICAL | 1 | 2-4 hours | Security breach risk |
| HIGH | 5 | 5-7 days | Core features broken |
| MEDIUM | 15+ | 10-15 days | Feature completeness |
| LOW | 6 | 2-3 days | Polish & configuration |

**Total Estimated Effort**: 17-25 days

---

## Recommended Action Plan

### Phase 1: Critical Security (Week 1)
1. Implement AAA permission checks in bulk handler
2. Add authentication middleware to all protected endpoints
3. Review and fix all authorization gaps

### Phase 2: Core Business Logic (Week 2-3)
1. Complete crop cycle service implementation
2. Implement bulk operation retry logic
3. Add FPO admin notification system
4. Fix farmer/AAA ID mapping in bulk imports

### Phase 3: Feature Completeness (Week 4)
1. Implement duplicate detection
2. Complete validation logic for bulk operations
3. Add farm data loading to farmer responses
4. Implement file download from URLs

### Phase 4: Polish & Configuration (Week 5)
1. Move hardcoded values to configuration
2. Update all handler annotations
3. Complete result file generation
4. Add comprehensive error handling

---

## Monitoring & Validation

After implementation, validate:
1. All endpoints have proper authentication/authorization
2. Bulk operations handle failures gracefully
3. Data quality notifications are sent
4. No placeholder/stub responses in production
5. All TODOs are resolved or documented as tech debt

---

## Notes

- Some `return nil, nil` patterns in non-test code need investigation
- Consider implementing circuit breaker for external service calls
- Add comprehensive logging for all stub replacements
- Ensure backward compatibility when replacing stubs
