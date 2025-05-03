package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Kisanlink/farmers-module/entities"
)

type FarmActivity struct {
	Base
	FarmId         string                `json:"farm_id" gorm:"type:varchar(10);not null"`
	CropCycleId    string                `json:"crop_cycle_id" gorm:"type:varchar(10);not null"`
	Activity       entities.ActivityType `json:"activity" gorm:"type:varchar(150);not null"`
	StartDate      *time.Time            `json:"start_date"`
	EndDate        *time.Time            `json:"end_date"`
	ActivityReport string                `json:"activity_report" gorm:"type:text"`

	CropCycle *CropCycle `json:"crop_cycle,omitempty" gorm:"foreignKey:CropCycleId"`
}

func (f *FarmActivity) BeforeCreate(tx *gorm.DB) (err error) {
	// Additional validation
	if !entities.ACTIVITY_TYPES.IsValid(string(f.Activity)) {
		return fmt.Errorf("invalid activity type: %s", f.Activity)
	}
	return
}
