# Farmer Table Normalization Design

## Problem Statement

Currently, the farmers-module has TWO separate farmer tables:
1. `farmers` - with embedded address fields (denormalized)
2. `farmer_profiles` - with FK to `addresses` table (normalized)

This creates:
- Data duplication
- Inconsistent data models
- Confusion about which table to use
- Service using wrong entity

## Proposed Solution: Consolidated Normalized Design

### Database Schema

#### 1. `farmers` table (PRIMARY)
```sql
CREATE TABLE farmers (
    id VARCHAR(255) PRIMARY KEY,  -- FMRR prefix

    -- AAA Integration (External IDs)
    aaa_user_id VARCHAR(255) NOT NULL,
    aaa_org_id VARCHAR(255) NOT NULL,
    kisan_sathi_user_id VARCHAR(255),  -- Optional field agent

    -- Personal Information
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    email VARCHAR(255),
    date_of_birth DATE,
    gender VARCHAR(50),

    -- Address (Normalized via FK)
    address_id VARCHAR(255),  -- FK to addresses table

    -- Additional Fields
    land_ownership_type VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',

    -- Flexible Data
    preferences JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    -- Constraints
    CONSTRAINT fk_farmers_address FOREIGN KEY (address_id) REFERENCES addresses(id),
    CONSTRAINT uq_farmer_aaa_user_org UNIQUE (aaa_user_id, aaa_org_id)
);

CREATE INDEX idx_farmers_aaa_user_id ON farmers(aaa_user_id);
CREATE INDEX idx_farmers_aaa_org_id ON farmers(aaa_org_id);
CREATE INDEX idx_farmers_kisan_sathi ON farmers(kisan_sathi_user_id);
CREATE INDEX idx_farmers_status ON farmers(status);
```

#### 2. `addresses` table (SHARED)
```sql
CREATE TABLE addresses (
    id VARCHAR(255) PRIMARY KEY,  -- ADDR prefix
    street_address TEXT,
    city VARCHAR(255),
    state VARCHAR(255),
    postal_code VARCHAR(50),
    country VARCHAR(255) DEFAULT 'India',
    coordinates GEOMETRY(Point, 4326),  -- PostGIS for spatial queries

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255)
);

CREATE INDEX idx_addresses_coordinates ON addresses USING GIST(coordinates);
CREATE INDEX idx_addresses_city ON addresses(city);
CREATE INDEX idx_addresses_state ON addresses(state);
```

#### 3. `farmer_links` table (RELATIONSHIPS)
```sql
CREATE TABLE farmer_links (
    id VARCHAR(255) PRIMARY KEY,  -- FMLK prefix

    -- AAA Integration
    aaa_user_id VARCHAR(255) NOT NULL,
    aaa_org_id VARCHAR(255) NOT NULL,
    kisan_sathi_user_id VARCHAR(255),

    -- Link Status
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    linked_at TIMESTAMP,
    unlinked_at TIMESTAMP,

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP,
    deleted_by VARCHAR(255),

    -- Constraints
    CONSTRAINT uq_farmer_link_aaa UNIQUE (aaa_user_id, aaa_org_id)
);

CREATE INDEX idx_farmer_links_aaa_user_id ON farmer_links(aaa_user_id);
CREATE INDEX idx_farmer_links_aaa_org_id ON farmer_links(aaa_org_id);
CREATE INDEX idx_farmer_links_status ON farmer_links(status);
```

### Entity Design (Go)

```go
// Farmer - Primary entity for farmer data
type Farmer struct {
    base.BaseModel

    // AAA Integration
    AAAUserID        string  `json:"aaa_user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
    AAAOrgID         string  `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
    KisanSathiUserID *string `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`

    // Personal Information
    FirstName   string  `json:"first_name" gorm:"type:varchar(255);not null"`
    LastName    string  `json:"last_name" gorm:"type:varchar(255);not null"`
    PhoneNumber string  `json:"phone_number" gorm:"type:varchar(50)"`
    Email       string  `json:"email" gorm:"type:varchar(255)"`
    DateOfBirth *string `json:"date_of_birth" gorm:"type:date"`
    Gender      string  `json:"gender" gorm:"type:varchar(50)"`

    // Address (Normalized)
    AddressID *string  `json:"address_id" gorm:"type:varchar(255)"`
    Address   *Address `json:"address,omitempty" gorm:"foreignKey:AddressID"`

    // Additional Fields
    LandOwnershipType string            `json:"land_ownership_type" gorm:"type:varchar(100)"`
    Status            string            `json:"status" gorm:"type:varchar(50);not null;default:'ACTIVE'"`
    Preferences       map[string]string `json:"preferences" gorm:"type:jsonb;default:'{}'"`
    Metadata          map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

func (f *Farmer) TableName() string {
    return "farmers"
}

func (f *Farmer) GetTableIdentifier() string {
    return "FMRR"
}

func (f *Farmer) GetTableSize() hash.TableSize {
    return hash.Large
}

// Address - Reusable address entity
type Address struct {
    base.BaseModel
    StreetAddress string `json:"street_address" gorm:"type:text"`
    City          string `json:"city" gorm:"type:varchar(255)"`
    State         string `json:"state" gorm:"type:varchar(255)"`
    PostalCode    string `json:"postal_code" gorm:"type:varchar(50)"`
    Country       string `json:"country" gorm:"type:varchar(255);default:'India'"`
    Coordinates   string `json:"coordinates" gorm:"type:geometry(Point,4326)"`
}

func (a *Address) TableName() string {
    return "addresses"
}

func (a *Address) GetTableIdentifier() string {
    return "ADDR"
}
```

## Migration Strategy

### Phase 1: Add New Normalized `farmers` Table
1. Create `farmers` table with normalized structure
2. Create migration to copy data from `farmer_profiles` to `farmers`
3. Keep `farmer_profiles` temporarily for backward compatibility

### Phase 2: Update Application Code
1. Update `FarmerService` to use `Farmer` entity instead of `FarmerProfile`
2. Update repository layer
3. Update all handlers and DTOs
4. Add tests

### Phase 3: Deprecate `farmer_profiles`
1. Add deprecation warnings
2. Update documentation
3. Schedule table drop for future release

### Phase 4: Clean Up
1. Drop `farmer_profiles` table
2. Remove legacy code
3. Update all documentation

## Benefits

1. **Single Source of Truth**: One `farmers` table
2. **Normalized (3NF)**: Addresses in separate table, reusable
3. **Scalability**: Can add multiple addresses per farmer if needed
4. **Data Integrity**: Foreign key constraints prevent orphaned records
5. **PostGIS Support**: Spatial queries on coordinates
6. **Clear Relationships**: Explicit FK relationships
7. **Audit Trail**: Full created/updated/deleted tracking

## Normalization Forms Achieved

- **1NF**: All columns contain atomic values
- **2NF**: No partial dependencies (all non-key attributes depend on entire primary key)
- **3NF**: No transitive dependencies (address fields moved to separate table)

## Example Queries

### Create Farmer with Address
```go
address := &Address{
    BaseModel:     *base.NewBaseModel("ADDR", hash.Medium),
    StreetAddress: "123 Farm Road",
    City:          "Village Name",
    State:         "Karnataka",
    PostalCode:    "560001",
    Country:       "India",
}

farmer := &Farmer{
    BaseModel:    *base.NewBaseModel("FMRR", hash.Large),
    AAAUserID:    "USER00000001",
    AAAOrgID:     "ORGN00000003",
    FirstName:    "John",
    LastName:     "Farmer",
    PhoneNumber:  "+919876543210",
    Address:      address,
    Status:       "ACTIVE",
}
```

### Query Farmers with Address
```go
var farmers []Farmer
db.Preload("Address").Where("aaa_org_id = ?", orgID).Find(&farmers)
```

### Spatial Query (PostGIS)
```sql
-- Find farmers within 10km radius
SELECT f.*
FROM farmers f
JOIN addresses a ON f.address_id = a.id
WHERE ST_DWithin(
    a.coordinates,
    ST_SetSRID(ST_MakePoint(77.5946, 12.9716), 4326),
    10000  -- 10km in meters
);
```

## Implementation Checklist

- [ ] Update `Farmer` entity with `AddressID` FK
- [ ] Remove embedded address fields from `Farmer`
- [ ] Create database migration script
- [ ] Update `FarmerService` to use `Farmer` entity
- [ ] Update `FarmerRepository` for new structure
- [ ] Update all DTOs and responses
- [ ] Add address management endpoints if needed
- [ ] Update tests
- [ ] Update API documentation
- [ ] Deploy migration
- [ ] Deprecate `farmer_profiles`
