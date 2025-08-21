package crop_cycle

import (
	"time"

	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropCycle represents an agricultural cycle within a farm
type CropCycle struct {
	base.BaseModel
	FarmID       string            `json:"farm_id" gorm:"type:uuid;not null"`
	Season       string            `json:"season" gorm:"type:season;not null"`
	Status       string            `json:"status" gorm:"type:cycle_status;not null;default:'PLANNED'"`
	StartDate    time.Time         `json:"start_date" gorm:"type:date;not null"`
	EndDate      *time.Time        `json:"end_date" gorm:"type:date"`
	PlannedCrops []string          `json:"planned_crops" gorm:"type:jsonb;default:'[]'"`
	Outcome      map[string]string `json:"outcome" gorm:"type:jsonb"`
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
	if cc.Season == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.StartDate.IsZero() {
		return common.ErrInvalidCropCycleData
	}
	return nil
}
