package farm

import (
	"github.com/Kisanlink/farmers-module/internal/entities/farm_irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_soil_type"
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

// Farm represents a farm with geographic boundaries
type Farm struct {
	base.BaseModel
	AAAFarmerUserID           string            `json:"aaa_farmer_user_id" gorm:"type:varchar(255);not null"`
	AAAOrgID                  string            `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
	Name                      *string           `json:"name" gorm:"type:varchar(255)"`
	OwnershipType             OwnershipType     `json:"ownership_type" gorm:"type:varchar(20);default:'OWN'"`
	Geometry                  string            `json:"geometry" gorm:"type:geometry(POLYGON,4326)"`
	AreaHa                    float64           `json:"area_ha" gorm:"type:numeric(12,4);->"`
	SoilTypeID                *string           `json:"soil_type_id" gorm:"type:varchar(255);index"`
	PrimaryIrrigationSourceID *string           `json:"primary_irrigation_source_id" gorm:"type:varchar(255);index"`
	BoreWellCount             int               `json:"bore_well_count" gorm:"default:0"`
	OtherIrrigationDetails    *string           `json:"other_irrigation_details" gorm:"type:text"`
	Metadata                  map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Relationships
	SoilType                *soil_type.SoilType                           `json:"soil_type" gorm:"foreignKey:SoilTypeID;references:ID"`
	PrimaryIrrigationSource *irrigation_source.IrrigationSource           `json:"primary_irrigation_source" gorm:"foreignKey:PrimaryIrrigationSourceID;references:ID"`
	IrrigationSources       []farm_irrigation_source.FarmIrrigationSource `json:"irrigation_sources" gorm:"foreignKey:FarmID;references:ID"`
	SoilTypes               []farm_soil_type.FarmSoilType                 `json:"soil_types" gorm:"foreignKey:FarmID;references:ID"`
}

// TableName returns the table name for the Farm model
func (f *Farm) TableName() string {
	return "farms"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *Farm) GetTableIdentifier() string {
	return "farm"
}

// GetTableSize returns the table size for ID generation
func (f *Farm) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (f *Farm) Validate() error {
	if f.AAAFarmerUserID == "" {
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
