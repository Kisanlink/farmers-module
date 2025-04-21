package models

import (
	"time"

	"github.com/Kisanlink/farmers-module/entities"
)

// CropCycle represents a planting cycle for a particular crop on a farm.
// DB constraints ensure data integrity: positive acreage/quantities, valid status, and date consistency.
type CropCycle struct {
	Base
	FarmID           string     `json:"farm_id" gorm:"type:varchar(10);not null"`
	CropID           string     `json:"crop_id" gorm:"type:varchar(10);not null"`
	StartDate        time.Time  `json:"start_date" gorm:"not null"`
	EndDate          *time.Time `json:"end_date" gorm:"column:end_date;check:end_date IS NULL OR end_date >= start_date"`
	Acreage          float64    `json:"acreage" gorm:"type:numeric(10,2);not null;check:acreage > 0"`
	ExpectedQuantity *float64   `json:"expected_quantity" gorm:"type:numeric(10,2);check:expected_quantity >= 0"`
	Quantity         float64    `json:"quantity" gorm:"type:numeric(10,2);not null;check:quantity >= 0"`
	Report           string     `json:"report" gorm:"type:text"`
	Status           string     `json:"status" gorm:"type:varchar(20);not null;default:'ONGOING';check:status IN ('ONGOING','COMPLETED')"`

	Crop Crop `json:"crop" gorm:"foreignKey:CropID;references:Id"`
}

var (
	CycleStatusOngoing   = string(entities.CROP_CYCLE_STATUSES.ONGOING)
	CycleStatusCompleted = string(entities.CROP_CYCLE_STATUSES.COMPLETED)
)
