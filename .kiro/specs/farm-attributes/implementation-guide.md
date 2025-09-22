# Farm Attributes Implementation Guide

## Database Schema Implementation

### 1. GORM Models Created

#### Core Models
- `internal/entities/soil_type/soil_type.go`: Soil type lookup table
- `internal/entities/irrigation_source/irrigation_source.go`: Irrigation source lookup table
- `internal/entities/farm_irrigation_source/farm_irrigation_source.go`: Junction table for farm-irrigation relationships
- `internal/entities/farm_soil_type/farm_soil_type.go`: Junction table for farm-soil relationships

#### Updated Models
- `internal/entities/farm/farm.go`: Enhanced with new attributes and relationships

### 2. Migration Service
- `internal/db/migration_service.go`: Handles AutoMigrate for all models
- Ensures proper migration order to handle foreign key dependencies

### 3. Lookup Service
- `internal/services/lookup_service.go`: Manages predefined lookup data
- `InitializeLookupData()`: Seeds database with predefined soil types and irrigation sources
- `GetSoilTypes()` and `GetIrrigationSources()`: Retrieval methods

### 4. API Handlers
- `internal/handlers/lookup_handlers.go`: HTTP handlers for lookup endpoints

## Usage Instructions

### 1. Database Migration

```go
// Initialize migration service
migrationService := db.NewMigrationService(database)

// Run auto migration
err := migrationService.AutoMigrate()
if err != nil {
    log.Fatal("Failed to migrate database:", err)
}

// Initialize lookup data
lookupService := services.NewLookupService(database)
err = lookupService.InitializeLookupData(context.Background())
if err != nil {
    log.Fatal("Failed to initialize lookup data:", err)
}
```

### 2. Farm Creation with Attributes

```go
// Create farm request with all attributes
createReq := &requests.CreateFarmRequest{
    AAAFarmerUserID: "user_123",
    AAAOrgID: "org_456",
    Name: &farmName, // Optional
    OwnershipType: "OWN",
    AreaHa: 2.5,
    Geometry: requests.GeometryData{
        WKT: "POLYGON(...)",
    },
    SoilTypeID: &soilTypeID, // Optional
    PrimaryIrrigationSourceID: &irrigationSourceID, // Optional
    BoreWellCount: 2,
    IrrigationSources: []requests.IrrigationSourceRequest{
        {
            IrrigationSourceID: irrigationSourceID,
            Count: 2,
            IsPrimary: true,
        },
    },
}

// Call farm service
farmService := services.NewFarmService(farmRepo, aaaService, db)
response, err := farmService.CreateFarm(ctx, createReq)
```

### 3. Lookup Data Retrieval

```go
// Get all soil types
soilTypes, err := lookupService.GetSoilTypes(ctx)

// Get all irrigation sources
irrigationSources, err := lookupService.GetIrrigationSources(ctx)
```

## API Integration

### 1. Register Lookup Routes

```go
// Register lookup routes
lookupHandlers := handlers.NewLookupHandlers(lookupService)
v1.GET("/lookup/soil-types", lookupHandlers.GetSoilTypes)
v1.GET("/lookup/irrigation-sources", lookupHandlers.GetIrrigationSources)
```

### 2. Client Integration

```javascript
// Fetch lookup data
const soilTypes = await fetch('/api/v1/lookup/soil-types').then(r => r.json());
const irrigationSources = await fetch('/api/v1/lookup/irrigation-sources').then(r => r.json());

// Create farm with attributes
const farmData = {
  aaa_farmer_user_id: "user_123",
  aaa_org_id: "org_456",
  name: "My Farm",
  ownership_type: "OWN",
  area_ha: 2.5,
  geometry: { wkt: "POLYGON(...)" },
  soil_type_id: "selected_soil_type_id",
  primary_irrigation_source_id: "selected_irrigation_source_id",
  bore_well_count: 2,
  irrigation_sources: [
    {
      irrigation_source_id: "selected_irrigation_source_id",
      count: 2,
      is_primary: true
    }
  ]
};

const response = await fetch('/api/v1/farms', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(farmData)
});
```

## Future Enhancements

### 1. Soil Report Integration
- Update soil types based on actual soil test reports
- Support multiple soil types per farm with percentages
- Track verification status and timestamps

### 2. Dynamic Irrigation Sources
- Allow custom irrigation sources beyond predefined list
- Support complex irrigation combinations
- Track seasonal variations in irrigation methods

### 3. Validation Enhancements
- Cross-validate irrigation sources with regional availability
- Validate soil types against geographic location
- Implement business rules for ownership type restrictions

## Backward Compatibility

- All new fields are optional to maintain backward compatibility
- Existing farm records will have default values for new attributes
- API responses include new fields only when they have values
- Legacy clients can continue using existing endpoints without modification
