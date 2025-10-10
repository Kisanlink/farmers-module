package irrigation_source

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// IrrigationSource represents different irrigation sources available
type IrrigationSource struct {
	base.BaseModel
	Name          string            `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description   string            `json:"description" gorm:"type:text"`
	RequiresCount bool              `json:"requires_count" gorm:"default:false"` // indicates if this source needs count (like bore wells)
	Properties    map[string]string `json:"properties" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the IrrigationSource model
func (i *IrrigationSource) TableName() string {
	return "irrigation_sources"
}

// GetTableIdentifier returns the table identifier for ID generation
func (i *IrrigationSource) GetTableIdentifier() string {
	return "IRRG"
}

// GetTableSize returns the table size for ID generation
func (i *IrrigationSource) GetTableSize() hash.TableSize {
	return hash.Small
}

// Predefined irrigation sources
var PredefinedIrrigationSources = []IrrigationSource{
	{Name: "BOREWELL", Description: "Borewell irrigation system", RequiresCount: true},
	{Name: "FLOOD_IRRIGATION", Description: "Flood irrigation method", RequiresCount: false},
	{Name: "DRIP_IRRIGATION", Description: "Drip irrigation system", RequiresCount: false},
	{Name: "CANAL", Description: "Canal irrigation", RequiresCount: false},
	{Name: "RAINFED", Description: "Rain-fed agriculture", RequiresCount: false},
	{Name: "OTHER", Description: "Other irrigation sources", RequiresCount: false},
}
