package models

import (
	"gorm.io/gorm"
)

type CropCategory string
type CropUnit string

const (
	CategoryCereal    CropCategory = "CEREAL"
	CategoryLegume    CropCategory = "LEGUME"
	CategoryVegetable CropCategory = "VEGETABLE"
	CategoryFruit     CropCategory = "FRUIT"
	CategorySpice     CropCategory = "SPICE"

	UnitKg    CropUnit = "KG"
	UnitTon   CropUnit = "TONS"
	UnitLitre CropUnit = "LITERS"
	UnitBags  CropUnit = "BAGS"
)

type Crop struct {
	Base
	CropName      string       `json:"crop_name" gorm:"type:varchar(100);not null"`
	Variant       string       `json:"variant" gorm:"type:varchar(100)"`
	CycleDuration int          `json:"cycle_duration"`
	Category      CropCategory `json:"category" gorm:"type:varchar(100);default:'CEREAL'"`
	Unit          CropUnit     `json:"unit" gorm:"type:varchar(20);default:'KG'"`
	Image         string       `json:"image" gorm:"type:text"`
	DocumentID    string       `json:"document_id" gorm:"type:text"`
}

func (c *Crop) BeforeCreate(tx *gorm.DB) (err error) {
	if c.Category == "" {
		c.Category = CategoryCereal
	}
	if c.Unit == "" {
		c.Unit = UnitKg
	}
	return
}
