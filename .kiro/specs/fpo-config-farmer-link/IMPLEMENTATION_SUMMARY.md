# FPO Configuration Integration in Farmer Creation

## Overview

This implementation adds optional FPO configuration linking during farmer creation. When enabled, the farmer profile is linked to the FPO's configuration metadata for integration purposes.

## Implementation Date
2025-11-25

## Changes Made

### 1. Request DTO Enhancement
**File**: `/Users/kaushik/farmers-module/internal/entities/requests/farmer.go`

Added `LinkFPOConfig` flag to `CreateFarmerRequest`:

```go
type CreateFarmerRequest struct {
    BaseRequest
    AAAUserID        string            `json:"aaa_user_id,omitempty"`
    AAAOrgID         string            `json:"aaa_org_id" validate:"required"`
    KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
    LinkFPOConfig    bool              `json:"link_fpo_config,omitempty"` // NEW FIELD
    Profile          FarmerProfileData `json:"profile" validate:"required"`
}
```

### 2. Service Layer Updates
**File**: `/Users/kaushik/farmers-module/internal/services/farmer_service.go`

#### Added FPO Config Service Dependency
- Added `fpoConfigService FPOConfigService` field to `FarmerServiceImpl`
- Created new constructor `NewFarmerServiceWithFPOConfig` that accepts FPO config service

#### Implemented FPO Config Linking Logic
- Added `linkFPOConfigToFarmer` helper method (lines 781-817)
- Integrated linking in `CreateFarmer` method (lines 292-309)

**Linking Flow**:
1. Farmer is created in database
2. If `link_fpo_config=true` and FPO config service is available:
   - Fetch FPO config for the organization
   - Verify config exists and is configured
   - Add FPO metadata to farmer profile:
     - `fpo_config_linked`: true
     - `fpo_config_id`: FPO config ID
     - `fpo_name`: FPO name
     - `fpo_config_linked_at`: timestamp
3. Update farmer with metadata
4. If linking fails, add error metadata (non-fatal)

### 3. Service Factory Update
**File**: `/Users/kaushik/farmers-module/internal/services/service_factory.go`

Updated service initialization to use new constructor:

```go
// Initialize FPO config service first (needed by farmer service)
fpoConfigService := NewFPOConfigService(repoFactory.FPOConfigRepo)

// Use NewFarmerServiceWithFPOConfig to enable FPO config linking
farmerService := NewFarmerServiceWithFPOConfig(
    repoFactory.FarmerRepo,
    aaaService,
    fpoConfigService,  // Injected
    cfg.AAA.DefaultPassword
)
```

### 4. Handler Documentation
**File**: `/Users/kaushik/farmers-module/internal/handlers/farmer_handler.go`

Updated Swagger documentation to include FPO config linking information:

```go
// @Description **FPO Configuration Linking**: Set link_fpo_config=true to link
// @Description the farmer to the FPO's configuration. This adds FPO metadata
// @Description to the farmer profile for integration purposes.
```

## Usage

### API Request Example

**Without FPO Config Linking** (default behavior):
```json
{
  "aaa_org_id": "ORGN00000001",
  "profile": {
    "first_name": "Ramesh",
    "last_name": "Kumar",
    "phone_number": "9876543210",
    "country_code": "+91"
  }
}
```

**With FPO Config Linking**:
```json
{
  "aaa_org_id": "ORGN00000001",
  "link_fpo_config": true,
  "profile": {
    "first_name": "Ramesh",
    "last_name": "Kumar",
    "phone_number": "9876543210",
    "country_code": "+91"
  }
}
```

### Response Metadata

When FPO config is successfully linked, the farmer's metadata will include:

```json
{
  "metadata": {
    "fpo_config_linked": true,
    "fpo_config_id": "ORGN00000001",
    "fpo_name": "Sample FPO",
    "fpo_config_linked_at": "2025-11-25T10:30:00Z"
  }
}
```

If linking fails (non-fatal), metadata includes error details:

```json
{
  "metadata": {
    "fpo_config_link_error": "FPO config not configured for org ORGN00000001",
    "fpo_config_link_pending": "true",
    "fpo_config_link_attempted_at": "2025-11-25T10:30:00Z"
  }
}
```

## Behavior

### When `link_fpo_config=true`:
1. **FPO Config Exists & Configured**: Farmer is linked with FPO config metadata
2. **FPO Config Exists but Not Configured**: Linking skipped, warning logged, error metadata added
3. **FPO Config Does Not Exist**: Linking fails, warning logged, error metadata added
4. **FPO Config Service Unavailable**: Linking skipped, warning logged

### When `link_fpo_config=false` or omitted:
- No FPO config linking is attempted (existing behavior)
- Farmer creation proceeds normally

## Error Handling

FPO config linking failures are **non-fatal** - they do not prevent farmer creation:
- Errors are logged for monitoring
- Error details are stored in farmer metadata
- Farmer creation succeeds regardless of linking outcome
- This ensures backward compatibility and resilient operation

## Database Impact

No database schema changes required. FPO config metadata is stored in the existing `farmers.metadata` JSONB column.

## Backward Compatibility

✅ Fully backward compatible:
- Default value of `link_fpo_config` is `false`
- Existing API calls work without changes
- Existing farmer service constructor still available for tests

## Testing Recommendations

1. **Test Case 1**: Create farmer with `link_fpo_config=true` and valid FPO config
   - Verify metadata contains FPO config details

2. **Test Case 2**: Create farmer with `link_fpo_config=true` and missing FPO config
   - Verify farmer is created
   - Verify error metadata is present

3. **Test Case 3**: Create farmer with `link_fpo_config=false` (or omitted)
   - Verify existing behavior unchanged
   - Verify no FPO config metadata present

4. **Test Case 4**: Create farmer with `link_fpo_config=true` and unconfigured FPO
   - Verify farmer is created
   - Verify error metadata indicates "not_configured"

## Files Modified

1. `/Users/kaushik/farmers-module/internal/entities/requests/farmer.go` - Added flag to request DTO
2. `/Users/kaushik/farmers-module/internal/services/farmer_service.go` - Added linking logic
3. `/Users/kaushik/farmers-module/internal/services/service_factory.go` - Updated initialization
4. `/Users/kaushik/farmers-module/internal/handlers/farmer_handler.go` - Updated API documentation

## Design Decisions

### Why Metadata Instead of Foreign Key?
- **Flexibility**: FPO config is optional and may not exist for all organizations
- **Decoupling**: Avoids hard dependency between farmers and FPO configs
- **Non-blocking**: Farmer creation can proceed even if FPO config is unavailable
- **Extensibility**: Metadata can hold additional FPO-related information as needed

### Why Non-Fatal Errors?
- **Resilience**: Farmer registration should not fail due to FPO config issues
- **User Experience**: Users can complete registration even if FPO is not fully configured
- **Observability**: Error metadata enables monitoring and reconciliation

### Why Optional Flag?
- **Backward Compatibility**: Existing clients unaffected
- **Performance**: Avoids unnecessary lookups when FPO config not needed
- **Explicit Intent**: Clear indication when FPO linking is desired

## Next Steps

1. Generate updated Swagger documentation: `make docs`
2. Run integration tests with FPO config scenarios
3. Update client team documentation with new flag
4. Monitor farmer creation logs for FPO config linking errors
5. Consider batch reconciliation job for failed linkages (if needed)
