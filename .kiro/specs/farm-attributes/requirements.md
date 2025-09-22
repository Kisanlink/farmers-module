# Farm Attributes Enhancement

## Overview
Enhance the farm creation workflow to include comprehensive farm attributes including farm name, ownership type, soil type, and irrigation sources with proper normalization.

## Requirements

### 1. Farm Basic Information
- **Farm Name**: Optional field for farm identification
- **Ownership Type**: Required enum with values:
  - `OWN`: Farmer owns the land
  - `LEASE`: Farmer leases the land
  - `SHARED`: Shared ownership arrangement

### 2. Soil Type Management
- **Predefined Soil Types**:
  - `BLACK`: Black soil - rich in clay content, good for cotton cultivation
  - `RED`: Red soil - well-drained, suitable for various crops
  - `SANDY`: Sandy soil - well-drained but low water retention
  - `LOAMY`: Loamy soil - ideal mixture of sand, silt, and clay
  - `ALLUVIAL`: Alluvial soil - fertile soil deposited by rivers
  - `MIXED`: Mixed soil types - combination of different soil types

- **Future Enhancement**: Support for multiple soil types per farm with percentages based on soil reports

### 3. Irrigation Source Management
- **Predefined Irrigation Sources**:
  - `BOREWELL`: Borewell irrigation system (requires count)
  - `FLOOD_IRRIGATION`: Flood irrigation method
  - `DRIP_IRRIGATION`: Drip irrigation system
  - `CANAL`: Canal irrigation
  - `RAINFED`: Rain-fed agriculture
  - `OTHER`: Other irrigation sources (requires description)

- **Special Handling**:
  - Borewell requires count field
  - Other sources may require additional details
  - Support for multiple irrigation sources per farm
  - Primary irrigation source designation

### 4. Data Normalization
- Separate lookup tables for soil types and irrigation sources
- Junction tables for farm-soil type and farm-irrigation source relationships
- Flexibility for future updates based on soil reports or field surveys

## API Endpoints

### Lookup Endpoints
- `GET /api/v1/lookup/soil-types`: Retrieve all available soil types
- `GET /api/v1/lookup/irrigation-sources`: Retrieve all available irrigation sources

### Enhanced Farm Endpoints
- Existing farm CRUD operations updated to support new attributes
- Backward compatibility maintained for existing API consumers

## Database Design

### New Tables
1. **soil_types**: Lookup table for soil types
2. **irrigation_sources**: Lookup table for irrigation sources
3. **farm_irrigation_sources**: Junction table linking farms to irrigation sources
4. **farm_soil_types**: Junction table linking farms to soil types

### Updated Tables
1. **farms**: Enhanced with new attributes and foreign key relationships

## Migration Strategy
- Use GORM AutoMigrate for schema updates
- Populate lookup tables with predefined values
- Maintain backward compatibility with existing farm records
