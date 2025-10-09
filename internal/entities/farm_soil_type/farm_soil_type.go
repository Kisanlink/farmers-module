package farm_soil_type

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmSoilType represents the junction table linking farms to soil types
type FarmSoilType struct {
	base.BaseModel
	FarmID       string     `json:"farm_id" gorm:"type:varchar(255);not null;index"`
	SoilTypeID   string     `json:"soil_type_id" gorm:"type:varchar(255);not null;index"`
	Percentage   float64    `json:"percentage" gorm:"type:decimal(5,2);default:100.00"` // percentage of farm area
	SoilReportID *string    `json:"soil_report_id" gorm:"type:varchar(255)"`            // reference to soil report when available
	VerifiedAt   *time.Time `json:"verified_at"`                                        // when soil type was verified through report

	// Relationships - removed farm reference to avoid circular import
	SoilType soil_type.SoilType `json:"soil_type" gorm:"foreignKey:SoilTypeID;references:ID"`
}

// TableName returns the table name for the FarmSoilType model
func (f *FarmSoilType) TableName() string {
	return "farm_soil_types"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FarmSoilType) GetTableIdentifier() string {
	return "FSTP"
}

// GetTableSize returns the table size for ID generation
func (f *FarmSoilType) GetTableSize() hash.TableSize {
	return hash.Medium
}
