# API Specification - Crop Cycles Area Allocation

## Overview

This document defines the REST API endpoints for crop cycles area allocation and stage-based farm activity management.

## Base URL

```
https://api.kisanlink.com/farmers-module/api/v1
```

## Authentication

All endpoints require Bearer token authentication:

```http
Authorization: Bearer <aaa-service-token>
X-Organization-ID: <org-id>
X-Request-ID: <unique-request-id>
```

## API Endpoints

### 1. Crop Cycles Management

#### 1.1 Create Crop Cycle with Area Allocation

**Endpoint:** `POST /crop-cycles`

**Description:** Creates a new crop cycle with area allocation

**Request Body:**
```json
{
  "farm_id": "FARM_12345",
  "farmer_id": "FRMR_67890",
  "crop_id": "CROP_11111",
  "variety_id": "VARTY_22222",
  "area_ha": 5.5,
  "season": "KHARIF",
  "start_date": "2024-06-01",
  "end_date": "2024-11-30",
  "metadata": {
    "planting_method": "DIRECT_SEEDING",
    "expected_yield_kg": 2750
  }
}
```

**Validation Rules:**
- `farm_id`: Required, must exist and belong to farmer
- `farmer_id`: Required, must match authenticated user's farmer
- `crop_id`: Required, must be active crop
- `variety_id`: Optional, must belong to specified crop if provided
- `area_ha`: Required, positive decimal up to 4 decimal places
- `season`: Required, enum: KHARIF, RABI, ZAID
- `start_date`: Optional, format: YYYY-MM-DD
- `end_date`: Optional, must be after start_date if provided

**Success Response (201):**
```json
{
  "success": true,
  "message": "Crop cycle created successfully",
  "data": {
    "id": "CRCY_99999",
    "farm_id": "FARM_12345",
    "farmer_id": "FRMR_67890",
    "crop_id": "CROP_11111",
    "crop_name": "Moringa",
    "variety_id": "VARTY_22222",
    "variety_name": "PKM-1",
    "area_ha": 5.5,
    "season": "KHARIF",
    "status": "PLANNED",
    "start_date": "2024-06-01",
    "end_date": "2024-11-30",
    "metadata": {
      "planting_method": "DIRECT_SEEDING",
      "expected_yield_kg": 2750
    },
    "created_at": "2024-01-14T10:30:00Z",
    "updated_at": "2024-01-14T10:30:00Z"
  }
}
```

**Error Responses:**

*400 Bad Request:*
```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Area must be a positive number",
    "field": "area_ha"
  }
}
```

*409 Conflict:*
```json
{
  "success": false,
  "error": {
    "code": "AREA_EXCEEDED",
    "message": "Requested area 6.0 ha exceeds available area 5.5 ha for farm FARM_12345",
    "details": {
      "farm_id": "FARM_12345",
      "total_area": 10.0,
      "allocated_area": 4.5,
      "available_area": 5.5,
      "requested_area": 6.0
    }
  }
}
```

#### 1.2 Update Crop Cycle Area

**Endpoint:** `PATCH /crop-cycles/{id}/area`

**Description:** Updates the allocated area for a crop cycle

**Path Parameters:**
- `id`: Crop cycle ID

**Request Body:**
```json
{
  "area_ha": 6.0
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Crop cycle area updated successfully",
  "data": {
    "id": "CRCY_99999",
    "area_ha": 6.0,
    "previous_area_ha": 5.5,
    "updated_at": "2024-01-14T11:00:00Z"
  }
}
```

**Error Responses:**

*403 Forbidden:*
```json
{
  "success": false,
  "error": {
    "code": "STATUS_INVALID",
    "message": "Cannot modify area for COMPLETED crop cycle"
  }
}
```

#### 1.3 Get Crop Cycle Details

**Endpoint:** `GET /crop-cycles/{id}`

**Description:** Retrieves detailed information about a crop cycle including area allocation

**Path Parameters:**
- `id`: Crop cycle ID

**Query Parameters:**
- `include`: Comma-separated list of relationships to include (farm,farmer,crop,variety,activities)

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "CRCY_99999",
    "farm_id": "FARM_12345",
    "farmer_id": "FRMR_67890",
    "crop_id": "CROP_11111",
    "variety_id": "VARTY_22222",
    "area_ha": 5.5,
    "season": "KHARIF",
    "status": "ACTIVE",
    "start_date": "2024-06-01",
    "end_date": "2024-11-30",
    "outcome": {
      "yield_kg": 2650,
      "quality_grade": "A"
    },
    "farm": {
      "id": "FARM_12345",
      "name": "North Field",
      "total_area_ha": 10.0
    },
    "crop": {
      "id": "CROP_11111",
      "name": "Moringa",
      "scientific_name": "Moringa oleifera"
    },
    "variety": {
      "id": "VARTY_22222",
      "name": "PKM-1",
      "duration_days": 180
    },
    "created_at": "2024-01-14T10:30:00Z",
    "updated_at": "2024-01-14T11:00:00Z"
  }
}
```

#### 1.4 List Crop Cycles

**Endpoint:** `GET /crop-cycles`

**Description:** Lists crop cycles with filtering and pagination

**Query Parameters:**
- `farm_id`: Filter by farm ID
- `farmer_id`: Filter by farmer ID
- `crop_id`: Filter by crop ID
- `season`: Filter by season (KHARIF, RABI, ZAID)
- `status`: Filter by status (PLANNED, ACTIVE, COMPLETED, CANCELLED)
- `min_area`: Minimum area in hectares
- `max_area`: Maximum area in hectares
- `start_date_from`: Start date range beginning (YYYY-MM-DD)
- `start_date_to`: Start date range ending (YYYY-MM-DD)
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `sort_by`: Sort field (area_ha, start_date, created_at)
- `sort_order`: Sort order (asc, desc)

**Success Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "CRCY_99999",
      "farm_id": "FARM_12345",
      "farm_name": "North Field",
      "farmer_id": "FRMR_67890",
      "farmer_name": "John Farmer",
      "crop_id": "CROP_11111",
      "crop_name": "Moringa",
      "variety_name": "PKM-1",
      "area_ha": 5.5,
      "season": "KHARIF",
      "status": "ACTIVE",
      "start_date": "2024-06-01",
      "created_at": "2024-01-14T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 45,
    "total_pages": 3,
    "has_next": true,
    "has_previous": false
  }
}
```

### 2. Area Allocation Management

#### 2.1 Get Farm Area Allocation Summary

**Endpoint:** `GET /farms/{farm_id}/area-allocation`

**Description:** Retrieves area allocation summary for a farm

**Path Parameters:**
- `farm_id`: Farm ID

**Query Parameters:**
- `include_cycles`: Include detailed cycle information (true/false, default: false)

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "farm_id": "FARM_12345",
    "farm_name": "North Field",
    "total_area_ha": 10.0,
    "allocated_area_ha": 8.5,
    "available_area_ha": 1.5,
    "utilization_percentage": 85.0,
    "active_cycles_count": 2,
    "planned_cycles_count": 1,
    "allocations": [
      {
        "crop_cycle_id": "CRCY_001",
        "crop_name": "Moringa",
        "variety_name": "PKM-1",
        "area_ha": 5.0,
        "status": "ACTIVE",
        "season": "KHARIF",
        "start_date": "2024-06-01"
      },
      {
        "crop_cycle_id": "CRCY_002",
        "crop_name": "Rice",
        "variety_name": "Basmati",
        "area_ha": 3.5,
        "status": "ACTIVE",
        "season": "KHARIF",
        "start_date": "2024-06-15"
      }
    ],
    "last_updated": "2024-01-14T12:00:00Z"
  }
}
```

#### 2.2 Check Area Availability

**Endpoint:** `POST /farms/{farm_id}/check-area-availability`

**Description:** Validates if requested area is available for allocation

**Path Parameters:**
- `farm_id`: Farm ID

**Request Body:**
```json
{
  "requested_area_ha": 3.0,
  "exclude_cycle_id": "CRCY_99999"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "is_available": true,
    "farm_id": "FARM_12345",
    "total_area_ha": 10.0,
    "currently_allocated_ha": 5.5,
    "requested_area_ha": 3.0,
    "available_area_ha": 4.5,
    "can_allocate": true
  }
}
```

**Error Response (200 - Area Not Available):**
```json
{
  "success": true,
  "data": {
    "is_available": false,
    "farm_id": "FARM_12345",
    "total_area_ha": 10.0,
    "currently_allocated_ha": 8.5,
    "requested_area_ha": 3.0,
    "available_area_ha": 1.5,
    "can_allocate": false,
    "shortage_ha": 1.5
  }
}
```

#### 2.3 Get Area Allocation Analytics

**Endpoint:** `GET /analytics/area-allocation`

**Description:** Provides analytics on area allocation across farms

**Query Parameters:**
- `farmer_id`: Filter by farmer
- `org_id`: Filter by organization
- `date_from`: Start date for analysis
- `date_to`: End date for analysis
- `group_by`: Grouping (farmer, crop, season)

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_farms": 150,
      "total_area_ha": 1500.0,
      "allocated_area_ha": 1200.0,
      "available_area_ha": 300.0,
      "average_utilization": 80.0
    },
    "by_crop": [
      {
        "crop_name": "Moringa",
        "total_area_ha": 600.0,
        "cycle_count": 120,
        "average_area_per_cycle": 5.0
      },
      {
        "crop_name": "Rice",
        "total_area_ha": 400.0,
        "cycle_count": 100,
        "average_area_per_cycle": 4.0
      }
    ],
    "by_season": [
      {
        "season": "KHARIF",
        "allocated_area_ha": 800.0,
        "cycle_count": 180
      },
      {
        "season": "RABI",
        "allocated_area_ha": 400.0,
        "cycle_count": 90
      }
    ],
    "utilization_trend": [
      {
        "month": "2024-01",
        "utilization_percentage": 75.0
      },
      {
        "month": "2024-02",
        "utilization_percentage": 80.0
      }
    ]
  }
}
```

### 3. Farm Activities with Stage Management

#### 3.1 Create Farm Activity with Stage

**Endpoint:** `POST /farm-activities`

**Description:** Creates a farm activity linked to a crop stage

**Request Body:**
```json
{
  "crop_cycle_id": "CRCY_99999",
  "crop_stage_id": "CSTG_11111",
  "farmer_id": "FRMR_67890",
  "activity_type": "IRRIGATION",
  "planned_at": "2024-07-15T10:00:00Z",
  "metadata": {
    "water_amount_liters": 500,
    "irrigation_method": "DRIP"
  }
}
```

**Validation Rules:**
- `crop_cycle_id`: Required, must exist and be active
- `crop_stage_id`: Optional, must belong to the crop of the cycle if provided
- `farmer_id`: Required, must match cycle's farmer
- `activity_type`: Required, string
- `planned_at`: Optional, ISO 8601 timestamp

**Success Response (201):**
```json
{
  "success": true,
  "message": "Farm activity created successfully",
  "data": {
    "id": "FACT_88888",
    "crop_cycle_id": "CRCY_99999",
    "crop_stage_id": "CSTG_11111",
    "stage_name": "Vegetative Growth",
    "farmer_id": "FRMR_67890",
    "activity_type": "IRRIGATION",
    "status": "PLANNED",
    "planned_at": "2024-07-15T10:00:00Z",
    "metadata": {
      "water_amount_liters": 500,
      "irrigation_method": "DRIP"
    },
    "created_by": "USER_12345",
    "created_at": "2024-01-14T13:00:00Z"
  }
}
```

**Error Response (400):**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_STAGE",
    "message": "Stage CSTG_11111 does not belong to the crop of cycle CRCY_99999",
    "details": {
      "cycle_crop_id": "CROP_11111",
      "stage_crop_id": "CROP_22222"
    }
  }
}
```

#### 3.2 Get Activities by Stage

**Endpoint:** `GET /crop-cycles/{cycle_id}/activities`

**Description:** Retrieves farm activities filtered and grouped by stage

**Path Parameters:**
- `cycle_id`: Crop cycle ID

**Query Parameters:**
- `stage_id`: Filter by specific stage ID
- `stage_order`: Filter by stage order number
- `status`: Filter by activity status (PLANNED, IN_PROGRESS, COMPLETED, CANCELLED)
- `activity_type`: Filter by activity type
- `date_from`: Activities planned/completed from date
- `date_to`: Activities planned/completed to date
- `group_by_stage`: Group results by stage (true/false, default: false)
- `include_stage_details`: Include stage information (true/false, default: true)
- `page`: Page number
- `page_size`: Items per page

**Success Response (200) - Grouped by Stage:**
```json
{
  "success": true,
  "data": {
    "crop_cycle_id": "CRCY_99999",
    "stages": [
      {
        "stage_id": "CSTG_10001",
        "stage_name": "Land Preparation",
        "stage_order": 1,
        "duration_days": 7,
        "total_activities": 5,
        "completed_activities": 5,
        "completion_percentage": 100.0,
        "activities": [
          {
            "id": "FACT_00001",
            "activity_type": "PLOWING",
            "status": "COMPLETED",
            "planned_at": "2024-06-01T08:00:00Z",
            "completed_at": "2024-06-01T12:00:00Z"
          },
          {
            "id": "FACT_00002",
            "activity_type": "LEVELING",
            "status": "COMPLETED",
            "planned_at": "2024-06-02T08:00:00Z",
            "completed_at": "2024-06-02T14:00:00Z"
          }
        ]
      },
      {
        "stage_id": "CSTG_10002",
        "stage_name": "Sowing",
        "stage_order": 2,
        "duration_days": 3,
        "total_activities": 3,
        "completed_activities": 2,
        "completion_percentage": 66.67,
        "activities": [
          {
            "id": "FACT_00003",
            "activity_type": "SEED_TREATMENT",
            "status": "COMPLETED",
            "planned_at": "2024-06-08T07:00:00Z",
            "completed_at": "2024-06-08T09:00:00Z"
          },
          {
            "id": "FACT_00004",
            "activity_type": "SOWING",
            "status": "IN_PROGRESS",
            "planned_at": "2024-06-09T06:00:00Z",
            "started_at": "2024-06-09T06:30:00Z"
          }
        ]
      }
    ],
    "summary": {
      "total_stages": 9,
      "completed_stages": 1,
      "current_stage": "Sowing",
      "total_activities": 45,
      "completed_activities": 7,
      "overall_completion": 15.56
    }
  }
}
```

#### 3.3 Update Activity Stage

**Endpoint:** `PATCH /farm-activities/{id}/stage`

**Description:** Updates or assigns a stage to an existing activity

**Path Parameters:**
- `id`: Farm activity ID

**Request Body:**
```json
{
  "crop_stage_id": "CSTG_10003"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Activity stage updated successfully",
  "data": {
    "id": "FACT_88888",
    "crop_stage_id": "CSTG_10003",
    "stage_name": "Vegetative Growth",
    "updated_at": "2024-01-14T14:00:00Z"
  }
}
```

#### 3.4 Get Stage Progress

**Endpoint:** `GET /crop-cycles/{cycle_id}/stage-progress`

**Description:** Retrieves stage-wise progress for a crop cycle

**Path Parameters:**
- `cycle_id`: Crop cycle ID

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "crop_cycle_id": "CRCY_99999",
    "crop_name": "Moringa",
    "current_stage": {
      "id": "CSTG_10003",
      "name": "Vegetative Growth",
      "order": 3,
      "started_on": "2024-06-15",
      "expected_completion": "2024-07-15"
    },
    "stages": [
      {
        "stage_id": "CSTG_10001",
        "stage_name": "Land Preparation",
        "stage_order": 1,
        "status": "COMPLETED",
        "duration_days": 7,
        "actual_days": 6,
        "total_activities": 5,
        "completed_activities": 5,
        "completion_percentage": 100.0,
        "started_date": "2024-06-01",
        "completed_date": "2024-06-07"
      },
      {
        "stage_id": "CSTG_10002",
        "stage_name": "Sowing",
        "stage_order": 2,
        "status": "COMPLETED",
        "duration_days": 3,
        "actual_days": 3,
        "total_activities": 3,
        "completed_activities": 3,
        "completion_percentage": 100.0,
        "started_date": "2024-06-08",
        "completed_date": "2024-06-11"
      },
      {
        "stage_id": "CSTG_10003",
        "stage_name": "Vegetative Growth",
        "stage_order": 3,
        "status": "IN_PROGRESS",
        "duration_days": 30,
        "elapsed_days": 4,
        "total_activities": 12,
        "completed_activities": 3,
        "completion_percentage": 25.0,
        "started_date": "2024-06-15",
        "expected_completion": "2024-07-15"
      },
      {
        "stage_id": "CSTG_10004",
        "stage_name": "Flowering",
        "stage_order": 4,
        "status": "PENDING",
        "duration_days": 15,
        "total_activities": 0,
        "completed_activities": 0,
        "completion_percentage": 0.0,
        "expected_start": "2024-07-16"
      }
    ],
    "overall_progress": {
      "total_stages": 9,
      "completed_stages": 2,
      "in_progress_stages": 1,
      "pending_stages": 6,
      "overall_completion_percentage": 22.22,
      "days_elapsed": 18,
      "expected_total_days": 180,
      "on_track": true
    }
  }
}
```

### 4. Bulk Operations

#### 4.1 Bulk Create Crop Cycles

**Endpoint:** `POST /crop-cycles/bulk`

**Description:** Creates multiple crop cycles with area validation

**Request Body:**
```json
{
  "validate_area": true,
  "stop_on_error": false,
  "cycles": [
    {
      "farm_id": "FARM_12345",
      "farmer_id": "FRMR_67890",
      "crop_id": "CROP_11111",
      "area_ha": 3.0,
      "season": "KHARIF"
    },
    {
      "farm_id": "FARM_12345",
      "farmer_id": "FRMR_67890",
      "crop_id": "CROP_22222",
      "area_ha": 2.5,
      "season": "KHARIF"
    }
  ]
}
```

**Success Response (207 Multi-Status):**
```json
{
  "success": true,
  "message": "Bulk operation completed",
  "data": {
    "total": 2,
    "successful": 2,
    "failed": 0,
    "results": [
      {
        "index": 0,
        "status": "success",
        "cycle_id": "CRCY_00001",
        "message": "Crop cycle created successfully"
      },
      {
        "index": 1,
        "status": "success",
        "cycle_id": "CRCY_00002",
        "message": "Crop cycle created successfully"
      }
    ]
  }
}
```

### 5. Reports and Exports

#### 5.1 Export Area Allocation Report

**Endpoint:** `GET /reports/area-allocation/export`

**Description:** Exports area allocation data in various formats

**Query Parameters:**
- `format`: Export format (csv, xlsx, pdf)
- `farmer_id`: Filter by farmer
- `date_from`: Start date
- `date_to`: End date
- `include_details`: Include cycle details (true/false)

**Success Response (200):**
```http
Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
Content-Disposition: attachment; filename="area_allocation_report_2024-01-14.xlsx"

[Binary data]
```

## Error Codes Reference

| Code | HTTP Status | Description |
|------|------------|-------------|
| INVALID_INPUT | 400 | Invalid request parameters |
| INVALID_AREA | 400 | Area value is invalid (negative or zero) |
| INVALID_STAGE | 400 | Stage doesn't belong to crop |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| AREA_EXCEEDED | 409 | Area allocation exceeds farm capacity |
| CONCURRENT_MODIFICATION | 409 | Resource modified by another request |
| STATUS_INVALID | 409 | Operation not allowed in current status |
| INTERNAL_ERROR | 500 | Server error |

## Rate Limiting

API endpoints are rate-limited per organization:

- Standard endpoints: 1000 requests per minute
- Bulk operations: 100 requests per minute
- Report exports: 10 requests per minute

Rate limit headers:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1704369600
```

## Pagination

All list endpoints support pagination:

```json
{
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false,
    "next_page": 2,
    "previous_page": null
  }
}
```

## Webhooks

Webhook events for area allocation changes:

### Event: crop_cycle.area.updated
```json
{
  "event": "crop_cycle.area.updated",
  "timestamp": "2024-01-14T15:00:00Z",
  "data": {
    "cycle_id": "CRCY_99999",
    "farm_id": "FARM_12345",
    "previous_area": 5.5,
    "new_area": 6.0,
    "user_id": "USER_12345"
  }
}
```

### Event: farm.area.exceeded
```json
{
  "event": "farm.area.exceeded",
  "timestamp": "2024-01-14T15:00:00Z",
  "data": {
    "farm_id": "FARM_12345",
    "attempted_allocation": 12.0,
    "farm_capacity": 10.0,
    "user_id": "USER_12345"
  }
}
```

## SDK Examples

### JavaScript/TypeScript

```typescript
import { FarmersModuleClient } from '@kisanlink/farmers-module-sdk';

const client = new FarmersModuleClient({
  apiKey: process.env.API_KEY,
  baseUrl: 'https://api.kisanlink.com/farmers-module'
});

// Create crop cycle with area
async function createCropCycle() {
  try {
    const response = await client.cropCycles.create({
      farmId: 'FARM_12345',
      farmerId: 'FRMR_67890',
      cropId: 'CROP_11111',
      areaHa: 5.5,
      season: 'KHARIF',
      startDate: '2024-06-01'
    });

    console.log('Crop cycle created:', response.data);
  } catch (error) {
    if (error.code === 'AREA_EXCEEDED') {
      console.error('Not enough area available:', error.details);
    }
  }
}

// Check area availability
async function checkAvailability() {
  const availability = await client.farms.checkAreaAvailability('FARM_12345', {
    requestedAreaHa: 3.0
  });

  if (availability.canAllocate) {
    console.log(`Can allocate ${availability.requestedAreaHa} ha`);
  } else {
    console.log(`Short by ${availability.shortageHa} ha`);
  }
}
```

### Python

```python
from kisanlink import FarmersModuleClient
from kisanlink.exceptions import AreaExceededException

client = FarmersModuleClient(
    api_key=os.environ['API_KEY'],
    base_url='https://api.kisanlink.com/farmers-module'
)

# Create crop cycle with area
try:
    cycle = client.crop_cycles.create(
        farm_id='FARM_12345',
        farmer_id='FRMR_67890',
        crop_id='CROP_11111',
        area_ha=5.5,
        season='KHARIF',
        start_date='2024-06-01'
    )
    print(f"Created cycle: {cycle['id']}")
except AreaExceededException as e:
    print(f"Not enough area: {e.available_area} ha available")

# Get stage progress
progress = client.crop_cycles.get_stage_progress('CRCY_99999')
for stage in progress['stages']:
    print(f"{stage['stage_name']}: {stage['completion_percentage']}%")
```

### Go

```go
package main

import (
    "fmt"
    "github.com/kisanlink/farmers-module-go"
)

func main() {
    client := farmersmodule.NewClient(
        farmersmodule.WithAPIKey(os.Getenv("API_KEY")),
        farmersmodule.WithBaseURL("https://api.kisanlink.com/farmers-module"),
    )

    // Create crop cycle with area
    cycle, err := client.CropCycles.Create(context.Background(), &farmersmodule.CreateCropCycleRequest{
        FarmID:    "FARM_12345",
        FarmerID:  "FRMR_67890",
        CropID:    "CROP_11111",
        AreaHa:    5.5,
        Season:    "KHARIF",
        StartDate: "2024-06-01",
    })

    if err != nil {
        if areaErr, ok := err.(*farmersmodule.AreaExceededError); ok {
            fmt.Printf("Available area: %.2f ha\n", areaErr.AvailableArea)
        }
        return
    }

    fmt.Printf("Created cycle: %s\n", cycle.ID)
}
```

## Testing

### Test Environment

```
Base URL: https://staging-api.kisanlink.com/farmers-module/api/v1
```

### Test Credentials

```json
{
  "api_key": "test_pk_xxxxxxxxxxxxx",
  "org_id": "ORG_TEST_001",
  "farmer_id": "FRMR_TEST_001",
  "farm_id": "FARM_TEST_001"
}
```

### Postman Collection

Import the Postman collection for testing:
[Download Collection](https://api.kisanlink.com/docs/farmers-module/postman-collection.json)

## API Versioning

The API uses URL versioning. Current version: v1

Future versions will be available at:
- `/api/v2/...`
- `/api/v3/...`

Deprecation notice will be provided 6 months before sunset.
