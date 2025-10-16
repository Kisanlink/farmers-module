package farm

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/Kisanlink/farmers-module/internal/entities/farm_irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_soil_type"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// OwnershipType represents the type of farm ownership
type OwnershipType string

const (
	OwnershipOwn    OwnershipType = "OWN"
	OwnershipLease  OwnershipType = "LEASE"
	OwnershipShared OwnershipType = "SHARED"
)

// Metadata is a custom type for JSONB metadata fields
type Metadata map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB deserialization
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*m = result
	return nil
}

// Value implements the driver.Valuer interface for JSONB serialization
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Farm represents a farm with geographic boundaries
type Farm struct {
	base.BaseModel
	FarmerID                  string        `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
	AAAUserID                 string        `json:"aaa_user_id" gorm:"type:varchar(255);not null"`
	AAAOrgID                  string        `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
	Name                      *string       `json:"name" gorm:"type:varchar(255)"`
	OwnershipType             OwnershipType `json:"ownership_type" gorm:"type:varchar(20);default:'OWN'"`
	Geometry                  string        `json:"geometry" gorm:"type:geometry(POLYGON,4326)"`
	AreaHa                    float64       `json:"area_ha" gorm:"column:area_ha;type:numeric(12,4)"`
	AreaHaComputed            float64       `json:"area_ha_computed" gorm:"column:area_ha_computed;type:numeric(12,4);->"`
	SoilTypeID                *string       `json:"soil_type_id" gorm:"type:varchar(255);index"`
	PrimaryIrrigationSourceID *string       `json:"primary_irrigation_source_id" gorm:"type:varchar(255);index"`
	BoreWellCount             int           `json:"bore_well_count" gorm:"default:0"`
	OtherIrrigationDetails    *string       `json:"other_irrigation_details" gorm:"type:text"`
	Metadata                  Metadata      `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`

	// Relationships
	Farmer                  *farmer.Farmer                                `json:"farmer,omitempty" gorm:"foreignKey:FarmerID;references:ID"`
	SoilType                *soil_type.SoilType                           `json:"soil_type,omitempty" gorm:"foreignKey:SoilTypeID;references:ID"`
	PrimaryIrrigationSource *irrigation_source.IrrigationSource           `json:"primary_irrigation_source,omitempty" gorm:"foreignKey:PrimaryIrrigationSourceID;references:ID"`
	IrrigationSources       []farm_irrigation_source.FarmIrrigationSource `json:"irrigation_sources,omitempty" gorm:"foreignKey:FarmID;references:ID"`
	SoilTypes               []farm_soil_type.FarmSoilType                 `json:"soil_types,omitempty" gorm:"foreignKey:FarmID;references:ID"`
}

// TableName returns the table name for the Farm model
func (f *Farm) TableName() string {
	return "farms"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *Farm) GetTableIdentifier() string {
	return "FARM"
}

// GetTableSize returns the table size for ID generation
func (f *Farm) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewFarm creates a new farm with initialized BaseModel
func NewFarm() *Farm {
	baseModel := base.NewBaseModel("FARM", hash.Medium)
	return &Farm{
		BaseModel:     *baseModel,
		Metadata:      make(Metadata),
		BoreWellCount: 0,
		OwnershipType: OwnershipOwn, // Default ownership type
	}
}

// Validation methods
func (f *Farm) Validate() error {
	if f.FarmerID == "" {
		return common.ErrInvalidFarmData
	}
	if f.AAAUserID == "" {
		return common.ErrInvalidFarmData
	}
	if f.AAAOrgID == "" {
		return common.ErrInvalidFarmData
	}
	if f.Geometry == "" {
		return common.ErrInvalidFarmGeometry
	}

	// Validate ownership type
	if f.OwnershipType != "" {
		switch f.OwnershipType {
		case OwnershipOwn, OwnershipLease, OwnershipShared:
			// Valid ownership type
		default:
			return common.ErrInvalidFarmData
		}
	}

	return nil
}
