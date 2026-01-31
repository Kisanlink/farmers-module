# Crop Variety Yield Per Tree - Age Range Based Implementation

**Date:** December 15, 2025  
**Module:** Farmers Module - Crop Catalog  
**Change Type:** API Structure Update

## Overview

Yield per tree is now expressed **only via age ranges**. The single `yield_per_tree` value has been removed. Total yield per tree is calculated as the **sum of all yields** across all age ranges.

### Summary of Changes

| Aspect | Before | After |
|--------|--------|-------|
| **Request Field** | `yield_per_tree` (optional) + `yield_by_age` (optional) | `yield_by_age` (required) |
| **Response Field** | `yield_per_tree` (single value) | `total_yield_per_tree` (calculated sum) |
| **Calculation** | N/A | Sum of all `yield_per_tree` values from all age ranges |
| **Validation** | Either field could be provided | `yield_by_age` must have at least one range |

## Problem Statement

Previously both a single `yield_per_tree` and `yield_by_age` array were accepted, which caused inconsistency. We now standardize on age ranges to capture perennial crop yields accurately.

## Changes

- **Entities (`internal/entities/crop_variety/crop_variety.go`)**
  - Removed `YieldPerTree` field.
  - `YieldByAge` is required; validation enforces non-empty and non-overlapping ranges.
  - Added `GetTotalYieldPerTree()` to compute sum of all yields across ranges.
  - `GetYieldForAge()` now only uses age ranges.

- **Requests (`internal/entities/requests/crop.go`)**
  - Removed `yield_per_tree` from create/update requests.
  - `yield_by_age` required for create; optional but validated when provided on update.

- **Responses (`internal/entities/responses/crop_responses.go`)**
  - Removed `yield_per_tree` from `CropVarietyData`.
  - Added `total_yield_per_tree` (computed) and returned alongside `yield_by_age`.

- **Service (`internal/services/crop_service.go`)**
  - Create/Update map request `yield_by_age` to entity.
  - All variety responses include `yield_by_age` and computed `total_yield_per_tree`.
  - Helpers added to convert age-range payloads between request/entity/response.

## API Endpoints

### 1. Create Crop Variety

**Endpoint:** `POST /varieties`

#### Before (Old Request)
```json
{
  "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
  "name": "Mango - Alphonso",
  "description": "Premium mango variety",
  "duration_days": 365,
  "yield_per_acre": 25.5,
  "yield_per_tree": 50.5,
  "yield_by_age": [
    {
      "age_from": 0,
      "age_to": 5,
      "yield_per_tree": 10.0
    },
    {
      "age_from": 6,
      "age_to": 15,
      "yield_per_tree": 60.0
    }
  ],
  "properties": {}
}
```

#### After (New Request)
```json
{
  "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
  "name": "Mango - Alphonso",
  "description": "Premium mango variety",
  "duration_days": 365,
  "yield_per_acre": 25.5,
  "yield_by_age": [
    {
      "age_from": 0,
      "age_to": 5,
      "yield_per_tree": 10.0
    },
    {
      "age_from": 6,
      "age_to": 15,
      "yield_per_tree": 60.0
    }
  ],
  "properties": {}
}
```

**Note:** `yield_per_tree` field is **removed** from request. `yield_by_age` is now **required**.

#### Before (Old Response)
```json
{
  "success": true,
  "message": "Crop variety created successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "name": "Mango - Alphonso",
    "description": "Premium mango variety",
    "duration_days": 365,
    "yield_per_acre": 25.5,
    "yield_per_tree": 50.5,
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 10.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 60.0
      }
    ],
    "properties": {},
    "is_active": true,
    "created_at": "2025-12-15T10:00:00Z",
    "updated_at": "2025-12-15T10:00:00Z"
  }
}
```

#### After (New Response)
```json
{
  "success": true,
  "message": "Crop variety created successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "name": "Mango - Alphonso",
    "description": "Premium mango variety",
    "duration_days": 365,
    "yield_per_acre": 25.5,
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 10.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 60.0
      }
    ],
    "total_yield_per_tree": 70.0,
    "properties": {},
    "is_active": true,
    "created_at": "2025-12-15T10:00:00Z",
    "updated_at": "2025-12-15T10:00:00Z"
  }
}
```

**Calculation:**
- Range 0-5: 10.0
- Range 6-15: 60.0
- **Total: 10.0 + 60.0 = 70.0** (sum of all yields)

---

### 2. Update Crop Variety

**Endpoint:** `PUT /varieties/{id}`

#### Before (Old Request)
```json
{
  "name": "Mango - Alphonso (Improved)",
  "yield_per_tree": 55.0,
  "yield_by_age": [
    {
      "age_from": 0,
      "age_to": 5,
      "yield_per_tree": 12.0
    },
    {
      "age_from": 6,
      "age_to": 15,
      "yield_per_tree": 65.0
    }
  ]
}
```

#### After (New Request)
```json
{
  "name": "Mango - Alphonso (Improved)",
  "yield_by_age": [
    {
      "age_from": 0,
      "age_to": 5,
      "yield_per_tree": 12.0
    },
    {
      "age_from": 6,
      "age_to": 15,
      "yield_per_tree": 65.0
    }
  ]
}
```

**Note:** `yield_per_tree` field is **removed** from update request.

#### Before (Old Response)
```json
{
  "success": true,
  "message": "Crop variety updated successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "name": "Mango - Alphonso (Improved)",
    "yield_per_tree": 55.0,
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 12.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 65.0
      }
    ],
    "is_active": true,
    "updated_at": "2025-12-15T11:00:00Z"
  }
}
```

#### After (New Response)
```json
{
  "success": true,
  "message": "Crop variety updated successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "name": "Mango - Alphonso (Improved)",
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 12.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 65.0
      }
    ],
    "total_yield_per_tree": 77.0,
    "is_active": true,
    "updated_at": "2025-12-15T11:00:00Z"
  }
}
```

---

### 3. Get Crop Variety

**Endpoint:** `GET /varieties/{id}`

#### Before (Old Response)
```json
{
  "success": true,
  "message": "Crop variety retrieved successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "crop_name": "Mango",
    "name": "Mango - Alphonso",
    "description": "Premium mango variety",
    "duration_days": 365,
    "yield_per_acre": 25.5,
    "yield_per_tree": 50.5,
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 10.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 60.0
      }
    ],
    "properties": {},
    "is_active": true,
    "created_at": "2025-12-15T10:00:00Z",
    "updated_at": "2025-12-15T10:00:00Z"
  }
}
```

#### After (New Response)
```json
{
  "success": true,
  "message": "Crop variety retrieved successfully",
  "request_id": "req_123456789",
  "data": {
    "id": "variety_123e4567-e89b-12d3-a456-426614174000",
    "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
    "crop_name": "Mango",
    "name": "Mango - Alphonso",
    "description": "Premium mango variety",
    "duration_days": 365,
    "yield_per_acre": 25.5,
    "yield_by_age": [
      {
        "age_from": 0,
        "age_to": 5,
        "yield_per_tree": 10.0
      },
      {
        "age_from": 6,
        "age_to": 15,
        "yield_per_tree": 60.0
      }
    ],
    "total_yield_per_tree": 70.0,
    "properties": {},
    "is_active": true,
    "created_at": "2025-12-15T10:00:00Z",
    "updated_at": "2025-12-15T10:00:00Z"
  }
}
```

---

### 4. List Crop Varieties

**Endpoint:** `GET /varieties?crop_id={crop_id}&page=1&page_size=20`

#### Before (Old Response)
```json
{
  "success": true,
  "message": "Crop varieties retrieved successfully",
  "request_id": "req_123456789",
  "data": [
    {
      "id": "variety_123e4567-e89b-12d3-a456-426614174000",
      "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
      "crop_name": "Mango",
      "name": "Mango - Alphonso",
      "yield_per_tree": 50.5,
      "yield_by_age": [
        {
          "age_from": 0,
          "age_to": 5,
          "yield_per_tree": 10.0
        },
        {
          "age_from": 6,
          "age_to": 15,
          "yield_per_tree": 60.0
        }
      ],
      "is_active": true
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

#### After (New Response)
```json
{
  "success": true,
  "message": "Crop varieties retrieved successfully",
  "request_id": "req_123456789",
  "data": [
    {
      "id": "variety_123e4567-e89b-12d3-a456-426614174000",
      "crop_id": "crop_123e4567-e89b-12d3-a456-426614174000",
      "crop_name": "Mango",
      "name": "Mango - Alphonso",
      "yield_by_age": [
        {
          "age_from": 0,
          "age_to": 5,
          "yield_per_tree": 10.0
        },
        {
          "age_from": 6,
          "age_to": 15,
          "yield_per_tree": 60.0
        }
      ],
      "total_yield_per_tree": 70.0,
      "is_active": true
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

## Frontend Actions

- **Forms:** remove single "Yield per tree" input; require "Yield by age" rows for perennial crops; show computed total (read-only).
- **Display:** show age-range table and computed total yield per tree.
- **Validation:** enforce at least one age range, non-overlapping, ascending ranges.

## Testing Checklist

- Create variety with age ranges only.
- Update variety age ranges and verify validation (required, non-overlapping).
- Confirm computed `total_yield_per_tree` matches sum of all age range yields.
- List/Get variety responses include age ranges and total.

