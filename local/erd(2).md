# Engineering Requirements Document (ERD) - Farmers Module Enhancements

## Introduction

This document outlines the engineering requirements for enhancing the Farmers Module based on specific feature requirements. After analyzing the current codebase implementation, this ERD categorizes features as "Already Implemented" vs "Needs Implementation" to provide accurate technical specifications.

### Context
- **Project**: KisanLink Farmers Module Enhancement
- **Current Version**: 1.0.0 with basic farm and crop cycle management
- **Enhancement Goal**: Add comprehensive crop master data system and enhanced crop cycle management
- **Technology Stack**: Go 1.24, PostgreSQL + PostGIS, GORM, Gin Framework

## Requirements Analysis

### Farm Addition Requirements

#### **Already Fully Implemented**

1. **Farm Location & Area**
   - PostGIS geometry support with auto-calculated area
   - Complete spatial validation and SRID checks
   - API endpoints: GET, POST, PUT, DELETE /api/v1/farms

2. **Farm Name (Optional)**
   - Field: `Name *string` in Farm entity
   - Full API support in CreateFarmRequest/UpdateFarmRequest

3. **Ownership Type**
   - Enum: `OwnershipOwn`, `OwnershipLease`, `OwnershipShared` (code uses "SHARED", requirements say "Others")
   - Field: `OwnershipType` in Farm entity
   - Validation in Farm.Validate() method
   - Note: Requirements list "Others" but code implements "SHARED" - functionally equivalent

4. **Soil Types**
   - Complete implementation: `soil_type.SoilType` entity
   - Predefined types: BLACK, RED, SANDY, LOAMY, ALLUVIAL, MIXED
   - Many-to-many relationship: `farm_soil_type.FarmSoilType`
   - Full API support through farm creation/update

5. **Irrigation Sources**
   - Complete implementation: `irrigation_source.IrrigationSource` entity
   - Predefined sources: BOREWELL, FLOOD_IRRIGATION, DRIP_IRRIGATION, CANAL, RAINFED, OTHER
   - Many-to-many relationship: `farm_irrigation_source.FarmIrrigationSource`
   - Full API support through farm creation/update

6. **Number of Bore Wells**
   - Field: `BoreWellCount int` in Farm entity
   - Full API support with validation

**Conclusion: ALL Farm Addition requirements are fully implemented. AutoMigrate will create all necessary tables. No farm-related changes needed.**

### Crop Addition Requirements

#### **Missing - Needs Implementation**

1. **Crop Master Data System**
   - No `crops` table exists
   - Missing fields: Crop Name, Category, Crop Duration, Units, Seasons
   - No structured crop categories (Cereal, Legume, Vegetable, Oil Seeds, Fruit, Spice)

2. **Crop Varieties System**
   - No `crop_varieties` table exists
   - Cannot select specific varieties within crops
   - No variety-specific attributes

3. **Growth Stages System**
   - No `crop_stages` table exists
   - Missing stages: Nursery, Vegetative Stage, Tillering Stage, Panicle initiation Stage, Flowering Stage, Grain Filling Stage, Harvesting stage

4. **Units Management**
   - No structured unit system (KG, Quintal, Tonnes, Pieces)
   - Not linked to crop types

5. **Image and Document Support**
   - No image field in crop entities
   - No document_id field for crop documentation

#### **Current Implementation (Basic)**

1. **Crop Handling**
   - Current: `planned_crops []string` (simple JSON array of crop names)
   - Required: Structured crop master data with categories, varieties, etc.

2. **Seasons**
   - Existing: RABI, KHARIF, ZAID, OTHER (in enum)
   - Replace: ZAID, OTHER should be replaced with SUMMER, PERENNIAL (requirements specify these specifically)

### Crop Cycle Requirements


#### **Missing - Needs Implementation**

1. **Enhanced Crop Selection**
   - No "Select Crop" dropdown (no crop master data)
   - No "Select Variety" dropdown (no variety system)
   - Current: `planned_crops []string` (JSON array of crop names)

2. **Perennial Crop Support**
   - No `number_of_trees` field in crop_cycles table/entity
   - No perennial-specific logic

3. **Additional Crop Cycle Fields**
   - No `acreage` field (cultivation area within farm)
   - No `harvest_date` field
   - Rename `start_date` to `sowing_transplanting_date` (requirement specifies this naming)
   - No `image_url` field
   - No `document_id` field
   - No `report_data` field

4. **Yield Management Requirements**
   - No yield tracking for regular crops ("Yield for other crops per Acre")
   - No yield management for perennial crops ("Age Range: ___ to ____, Yield Per tree ____")

5. **Quantity Management**
   - Remove Expected Quantity support (currently in outcome field, requirement says to remove)
   - Actual Quantity can be stored in existing outcome field

#### **Current Implementation (Basic)**

1. **Basic Crop Cycle Management**
   - Basic CropCycle entity: ID, FarmID, FarmerID, Season, Status, StartDate, EndDate, PlannedCrops, Outcome
   - Complete API endpoints: POST, PUT, GET /api/v1/crops/cycles
   - Start date, end date support (currently as start_date, needs rename to sowing_transplanting_date)
   - Status management: PLANNED, ACTIVE, COMPLETED, CANCELLED
   - planned_crops as JSON string array for current basic crop handling
   - Basic outcome field as JSON map for storing results

## Implementation Requirements

**Note: AutoMigrate will handle all database schema creation/updates automatically since there's no existing data.**

### Crop Master Data System

#### New Database Tables

```sql
-- Crops master table
CREATE TABLE crops (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL, -- 'CEREAL', 'LEGUME', 'VEGETABLE', 'OIL_SEEDS', 'FRUIT', 'SPICE'
    crop_duration_days INTEGER, -- default duration
    typical_units VARCHAR(20)[], -- ['KG', 'QUINTAL', 'TONNES', 'PIECES']
    seasons VARCHAR(20)[], -- ['KHARIF', 'RABI', 'SUMMER', 'PERENNIAL']
    image_url VARCHAR(500),
    document_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Crop varieties table
CREATE TABLE crop_varieties (
    id VARCHAR(255) PRIMARY KEY,
    crop_id VARCHAR(255) NOT NULL REFERENCES crops(id),
    variety_name VARCHAR(100) NOT NULL,
    duration_days INTEGER,
    characteristics TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Crop stages table
CREATE TABLE crop_stages (
    id VARCHAR(255) PRIMARY KEY,
    crop_id VARCHAR(255) NOT NULL REFERENCES crops(id),
    stage_name VARCHAR(100) NOT NULL,
    stage_order INTEGER NOT NULL,
    typical_duration_days INTEGER,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### New API Endpoints

```go
// Crop Master Data Management
POST   /api/v1/crops                    # Create crop
GET    /api/v1/crops                    # List crops
GET    /api/v1/crops/:id                # Get crop details
PUT    /api/v1/crops/:id                # Update crop
DELETE /api/v1/crops/:id                # Delete crop

POST   /api/v1/crops/:id/varieties      # Create variety
GET    /api/v1/crops/:id/varieties      # List varieties for crop
PUT    /api/v1/crops/varieties/:id      # Update variety
DELETE /api/v1/crops/varieties/:id      # Delete variety

POST   /api/v1/crops/:id/stages         # Create stage
GET    /api/v1/crops/:id/stages         # List stages for crop
PUT    /api/v1/crops/stages/:id         # Update stage
DELETE /api/v1/crops/stages/:id         # Delete stage

# Dropdown/Lookup APIs
GET    /api/v1/lookups/crop-categories  # Get crop categories
GET    /api/v1/lookups/units           # Get available units
GET    /api/v1/lookups/seasons         # Get seasons (including new ones)
```

### Enhanced Crop Cycles

#### Database Schema Changes

```sql
-- Extend existing crop_cycles table
ALTER TABLE crop_cycles ADD COLUMN crop_id VARCHAR(255) REFERENCES crops(id);
ALTER TABLE crop_cycles ADD COLUMN variety_id VARCHAR(255) REFERENCES crop_varieties(id);
ALTER TABLE crop_cycles ADD COLUMN number_of_trees INTEGER; -- only for perennial crops
ALTER TABLE crop_cycles ADD COLUMN acreage DECIMAL(12,4); -- cultivation area within farm
ALTER TABLE crop_cycles ADD COLUMN harvest_date DATE;
ALTER TABLE crop_cycles ADD COLUMN sowing_transplanting_date DATE; -- rename from start_date
ALTER TABLE crop_cycles ADD COLUMN yield_per_acre DECIMAL(12,4); -- for regular crops
ALTER TABLE crop_cycles ADD COLUMN tree_age_range_min INTEGER; -- for perennial crops
ALTER TABLE crop_cycles ADD COLUMN tree_age_range_max INTEGER; -- for perennial crops
ALTER TABLE crop_cycles ADD COLUMN yield_per_tree DECIMAL(12,4); -- for perennial crops
ALTER TABLE crop_cycles ADD COLUMN image_url VARCHAR(500);
ALTER TABLE crop_cycles ADD COLUMN document_id VARCHAR(255);
ALTER TABLE crop_cycles ADD COLUMN report_data JSONB;

-- Update season enum to replace old values with new ones
-- Note: ZAID and OTHER will be replaced with SUMMER and PERENNIAL
-- Implementation will handle migration of existing data if any exists
DROP TYPE IF EXISTS season CASCADE;
CREATE TYPE season AS ENUM ('RABI', 'KHARIF', 'SUMMER', 'PERENNIAL');

-- Create indexes for performance
CREATE INDEX idx_crop_cycles_crop_id ON crop_cycles (crop_id);
CREATE INDEX idx_crop_cycles_variety_id ON crop_cycles (variety_id);
```

#### Enhanced API Endpoints

```go
// Enhanced crop cycle endpoints
POST   /api/v1/crops/cycles            # Create with crop/variety selection
PUT    /api/v1/crops/cycles/:id        # Update with new fields
GET    /api/v1/crops/cycles/:id        # Get with crop/variety details
GET    /api/v1/crops/cycles            # List with enhanced filtering

# New specialized endpoints
PUT    /api/v1/crops/cycles/:id/harvest # Record harvest with actual quantity
POST   /api/v1/crops/cycles/:id/report  # Generate/upload report
GET    /api/v1/crops/cycles/:id/stages  # Get stage progression (future scope)
```

### Service Layer Implementation

#### New Services Required

```go
// internal/services/crop_service.go
type CropService interface {
    CreateCrop(ctx context.Context, req *requests.CreateCropRequest) (*responses.CropResponse, error)
    ListCrops(ctx context.Context, req *requests.ListCropsRequest) (*responses.CropsListResponse, error)
    GetCrop(ctx context.Context, cropID string) (*responses.CropResponse, error)
    UpdateCrop(ctx context.Context, req *requests.UpdateCropRequest) (*responses.CropResponse, error)
    DeleteCrop(ctx context.Context, cropID string) error
}

// internal/services/crop_variety_service.go
type CropVarietyService interface {
    CreateVariety(ctx context.Context, req *requests.CreateVarietyRequest) (*responses.VarietyResponse, error)
    ListVarietiesByCrop(ctx context.Context, cropID string) (*responses.VarietyListResponse, error)
    GetVariety(ctx context.Context, varietyID string) (*responses.VarietyResponse, error)
    UpdateVariety(ctx context.Context, req *requests.UpdateVarietyRequest) (*responses.VarietyResponse, error)
    DeleteVariety(ctx context.Context, varietyID string) error
}

// Enhanced internal/services/crop_cycle_service.go
type CropCycleService interface {
    // Enhanced existing methods
    StartCycle(ctx context.Context, req *requests.StartEnhancedCycleRequest) (*responses.CropCycleResponse, error)
    UpdateCycle(ctx context.Context, req *requests.UpdateEnhancedCycleRequest) (*responses.CropCycleResponse, error)

    // New methods
    RecordHarvest(ctx context.Context, req *requests.RecordHarvestRequest) (*responses.CropCycleResponse, error)
    UploadReport(ctx context.Context, req *requests.UploadReportRequest) (*responses.ReportResponse, error)
}
```

## Success Criteria

1. **Complete Crop Master Data System**
   - All crop categories, varieties, and stages properly stored and accessible
   - Dropdown APIs providing structured data for UI
   - Full CRUD operations with proper validation

2. **Enhanced Crop Cycle Management**
   - Crop/variety selection integrated into cycle creation
   - Perennial crop support with tree counting
   - Complete field coverage as per requirements
   - Harvest tracking and reporting functionality

3. **Backward Compatibility**
   - Existing crop cycles continue to work
   - Legacy planned_crops field maintained
   - No breaking changes to current APIs

4. **Performance & Scalability**
   - Efficient queries with proper indexing
   - Support for thousands of crops and varieties
   - Fast dropdown/lookup responses

## Technical Considerations

### Database Design
- Use proper foreign key constraints for data integrity
- Implement cascading deletes where appropriate
- Add indexes for frequently queried fields
- Use JSONB for flexible metadata storage

### API Design
- Follow existing RESTful patterns
- Implement proper error handling and validation
- Add request/response DTOs for type safety
- Include comprehensive Swagger documentation

