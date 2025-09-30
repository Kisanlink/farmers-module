package crop_cycle

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropCycle represents an agricultural cycle within a farm
type CropCycle struct {
	base.BaseModel
	FarmID    string            `json:"farm_id" gorm:"type:varchar(255);not null"`
	FarmerID  string            `json:"farmer_id" gorm:"type:uuid"`
	Season    string            `json:"season" gorm:"type:season;not null"`
	Status    string            `json:"status" gorm:"type:cycle_status;not null;default:'PLANNED'"`
	StartDate *time.Time        `json:"start_date" gorm:"type:date"`
	EndDate   *time.Time        `json:"end_date" gorm:"type:date"`
	CropID    string            `json:"crop_id" gorm:"type:uuid;not null;index"`
	VarietyID *string           `json:"variety_id" gorm:"type:uuid;index"`
	Outcome   map[string]string `json:"outcome" gorm:"type:jsonb;default:'{}'"`

	// Relationships
	Crop    *crop.Crop                `json:"crop,omitempty" gorm:"foreignKey:CropID;references:ID"`
	Variety *crop_variety.CropVariety `json:"variety,omitempty" gorm:"foreignKey:VarietyID;references:ID"`
}

// TableName returns the table name for the CropCycle model
func (cc *CropCycle) TableName() string {
	return "crop_cycles"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cc *CropCycle) GetTableIdentifier() string {
	return "crop_cycle"
}

// GetTableSize returns the table size for ID generation
func (cc *CropCycle) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (cc *CropCycle) Validate() error {
	if cc.FarmID == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.FarmerID == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.Season == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.CropID == "" {
		return common.ErrInvalidCropCycleData
	}
	return nil
}

// GetCropName returns the crop name if crop relationship is loaded
func (cc *CropCycle) GetCropName() string {
	if cc.Crop != nil {
		return cc.Crop.Name
	}
	return ""
}

// GetVarietyName returns the variety name if variety relationship is loaded
func (cc *CropCycle) GetVarietyName() string {
	if cc.Variety != nil {
		return cc.Variety.Name
	}
	return ""
}
