package models

import (
	"github.com/Kisanlink/farmers-module/entities"
)

type Crop struct {
	Base
	CropName      string                `json:"crop_name" gorm:"type:varchar(100);not null"`
	Variant       string                `json:"variant" gorm:"type:varchar(100)"`
	CycleDuration int                   `json:"cycle_duration"`
	Category      entities.CropCategory `json:"category" gorm:"type:varchar(100);not null"`
	Unit          entities.CropUnit     `json:"unit" gorm:"type:varchar(20);not null"`
	Image         string                `json:"image" gorm:"type:text"`
	DocumentID    string                `json:"document_id" gorm:"type:text"`
}
