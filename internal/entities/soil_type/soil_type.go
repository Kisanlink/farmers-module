package soil_type

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// SoilType represents different soil types available
type SoilType struct {
	base.BaseModel
	Name        string         `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text"`
	Properties  entities.JSONB `json:"properties" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the SoilType model
func (s *SoilType) TableName() string {
	return "soil_types"
}

// GetTableIdentifier returns the table identifier for ID generation
func (s *SoilType) GetTableIdentifier() string {
	return "SOIL"
}

// GetTableSize returns the table size for ID generation
func (s *SoilType) GetTableSize() hash.TableSize {
	return hash.Small
}

// Predefined soil types
var PredefinedSoilTypes = []SoilType{
	{Name: "BLACK", Description: "Black soil - rich in clay content, good for cotton cultivation"},
	{Name: "RED", Description: "Red soil - well-drained, suitable for various crops"},
	{Name: "SANDY", Description: "Sandy soil - well-drained but low water retention"},
	{Name: "LOAMY", Description: "Loamy soil - ideal mixture of sand, silt, and clay"},
	{Name: "ALLUVIAL", Description: "Alluvial soil - fertile soil deposited by rivers"},
	{Name: "MIXED", Description: "Mixed soil types - combination of different soil types"},
}
