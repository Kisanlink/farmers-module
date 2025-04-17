package models

import (
	"time"
)

type CropCycle struct {
	Base
	FarmID           string     `json:"farm_id" gorm:"type:varchar(10);not null"`
	CropID           string     `json:"crop_id" gorm:"type:varchar(10);not null"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Acreage          float64    `json:"acreage" gorm:"type:numeric(10,2)"`
	ExpectedQuantity float64    `json:"expected_quantity" gorm:"type:numeric(10,2)"`
	Quantity         float64    `json:"quantity" gorm:"type:numeric(10,2)"`
	Report           string     `json:"report" gorm:"type:text"`

	// Fixed relationship configuration
	Crop Crop `json:"crop" gorm:"foreignKey:CropID;references:Id"`
}
