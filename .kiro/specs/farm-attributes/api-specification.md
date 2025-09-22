# Farm Attributes API Specification

## Enhanced Farm Creation Request

### POST /api/v1/farms

```json
{
  "aaa_farmer_user_id": "user_123",
  "aaa_org_id": "org_456",
  "name": "My Farm", // Optional
  "ownership_type": "OWN", // Optional: OWN|LEASE|SHARED
  "area_ha": 2.5,
  "geometry": {
    "wkt": "POLYGON((...))"
  },
  "soil_type_id": "soil_type_id_123", // Optional
  "primary_irrigation_source_id": "irrigation_source_id_456", // Optional
  "bore_well_count": 2, // Optional, used when irrigation source is BOREWELL
  "other_irrigation_details": "Details for other irrigation", // Optional
  "irrigation_sources": [ // Optional array for multiple sources
    {
      "irrigation_source_id": "irrigation_source_id_456",
      "count": 2, // For sources that require count
      "details": "Additional details", // Optional
      "is_primary": true // Optional
    }
  ],
  "metadata": {} // Optional
}
```

### PUT /api/v1/farms/{id}

Same structure as creation but all fields are optional except `id`.

## Lookup Endpoints

### GET /api/v1/lookup/soil-types

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "soil_type_id_1",
      "name": "BLACK",
      "description": "Black soil - rich in clay content, good for cotton cultivation",
      "properties": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "soil_type_id_2",
      "name": "RED",
      "description": "Red soil - well-drained, suitable for various crops",
      "properties": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### GET /api/v1/lookup/irrigation-sources

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "irrigation_source_id_1",
      "name": "BOREWELL",
      "description": "Borewell irrigation system",
      "requires_count": true,
      "properties": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "irrigation_source_id_2",
      "name": "DRIP_IRRIGATION",
      "description": "Drip irrigation system",
      "requires_count": false,
      "properties": {},
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## Enhanced Farm Response

### GET /api/v1/farms/{id}

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "farm_id_123",
    "aaa_farmer_user_id": "user_123",
    "aaa_org_id": "org_456",
    "name": "My Farm",
    "ownership_type": "OWN",
    "geometry": "POLYGON(...)",
    "area_ha": 2.5,
    "soil_type_id": "soil_type_id_123",
    "primary_irrigation_source_id": "irrigation_source_id_456",
    "bore_well_count": 2,
    "other_irrigation_details": null,
    "metadata": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "soil_type": {
      "id": "soil_type_id_123",
      "name": "BLACK",
      "description": "Black soil - rich in clay content, good for cotton cultivation"
    },
    "primary_irrigation_source": {
      "id": "irrigation_source_id_456",
      "name": "BOREWELL",
      "description": "Borewell irrigation system",
      "requires_count": true
    },
    "irrigation_sources": [
      {
        "id": "farm_irrigation_source_id_1",
        "farm_id": "farm_id_123",
        "irrigation_source_id": "irrigation_source_id_456",
        "count": 2,
        "details": null,
        "is_primary": true,
        "irrigation_source": {
          "id": "irrigation_source_id_456",
          "name": "BOREWELL",
          "description": "Borewell irrigation system"
        }
      }
    ],
    "soil_types": [
      {
        "id": "farm_soil_type_id_1",
        "farm_id": "farm_id_123",
        "soil_type_id": "soil_type_id_123",
        "percentage": 100.00,
        "soil_report_id": null,
        "verified_at": null,
        "soil_type": {
          "id": "soil_type_id_123",
          "name": "BLACK",
          "description": "Black soil - rich in clay content, good for cotton cultivation"
        }
      }
    ]
  }
}
```

## Validation Rules

### Ownership Type
- Must be one of: `OWN`, `LEASE`, `SHARED`
- Defaults to `OWN` if not provided

### Bore Well Count
- Must be >= 0
- Only relevant when irrigation source is `BOREWELL`

### Irrigation Sources Array
- Each item must have valid `irrigation_source_id`
- `count` must be >= 0
- Only one source can have `is_primary: true`
- If no primary is specified, the first source becomes primary

### Soil Type
- Must reference valid soil type ID from lookup table
- Optional during creation but can be updated later based on soil reports
