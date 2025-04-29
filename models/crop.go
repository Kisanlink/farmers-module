package models

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/entities"
	"gorm.io/gorm"
)

type Crop struct {
	Base
	CropName      string                `json:"crop_name" gorm:"type:varchar(100);not null"`
	Variant       string                `json:"variant" gorm:"type:varchar(100)"`
	CycleDuration int                   `json:"cycle_duration"`
	Category      entities.CropCategory `json:"category" gorm:"type:varchar(100);not null"`
	Unit          entities.CropUnit     `json:"unit" gorm:"type:varchar(20);not null"`
	Image         string                `json:"image" gorm:"type:text"`
	DocumentId    string                `json:"document_id" gorm:"type:text"`
}

func (c *Crop) BeforeCreate(tx *gorm.DB) (err error) {

	// Validate Category
	if !entities.CROP_CATEGORIES.IsValid(string(c.Category)) {
		return fmt.Errorf("invalid crop category: %s. Valid values are: %v",
			c.Category, entities.CROP_CATEGORIES)
	}

	// Validate Unit
	if !entities.CROP_UNITS.IsValid(string(c.Unit)) {
		return fmt.Errorf("invalid crop unit: %s. Valid values are: %v",
			c.Unit, entities.CROP_UNITS)
	}

	return nil
}
