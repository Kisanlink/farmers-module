package farm_irrigation_source

import (
	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmIrrigationSource represents the junction table linking farms to irrigation sources
type FarmIrrigationSource struct {
	base.BaseModel
	FarmID             string `json:"farm_id" gorm:"type:varchar(255);not null;index"`
	IrrigationSourceID string `json:"irrigation_source_id" gorm:"type:varchar(255);not null;index"`
	Count              int    `json:"count" gorm:"default:0"`   // for sources that require count like bore wells
	Details            string `json:"details" gorm:"type:text"` // for additional details like "Others" description
	IsPrimary          bool   `json:"is_primary" gorm:"default:false"`

	// Relationships - removed farm reference to avoid circular import
	IrrigationSource irrigation_source.IrrigationSource `json:"irrigation_source" gorm:"foreignKey:IrrigationSourceID;references:ID"`
}

// TableName returns the table name for the FarmIrrigationSource model
func (f *FarmIrrigationSource) TableName() string {
	return "farm_irrigation_sources"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FarmIrrigationSource) GetTableIdentifier() string {
	return "FISC"
}

// GetTableSize returns the table size for ID generation
func (f *FarmIrrigationSource) GetTableSize() hash.TableSize {
	return hash.Medium
}
