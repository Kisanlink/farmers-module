package models

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

type Crop struct {
	Base
	CropName      string                `json:"crop_name" gorm:"type:varchar(100);not null"`
	Variant       string                `json:"variant" gorm:"type:varchar(100)"`
	CycleDuration int                   `json:"cycle_duration"`
	Season        entities.CropSeason   `json:"season,omitempty" gorm:"column:season;type:varchar(20)"`
	Category      entities.CropCategory `json:"category" gorm:"type:varchar(100);not null"`
	Unit          entities.CropUnit     `json:"unit" gorm:"type:varchar(20);not null"`
	Image         string                `json:"image" gorm:"type:text"`
	DocumentID    string                `json:"document_id" gorm:"type:text"`

	Stages []CropStage `json:"stages,omitempty" gorm:"foreignKey:CropID"`
}

// Stage represents a master, reusable growth stage for any crop.
type Stage struct {
	Base
	StageName   string `json:"stage_name" gorm:"type:varchar(100);not null;unique"`
	Description string `json:"description,omitempty" gorm:"type:text"`
}

// TableName sets the table name for the Stage model.
func (Stage) TableName() string {
	return "stages"
}

// BeforeCreate hook for Stage to generate ID.
func (s *Stage) BeforeCreate(tx *gorm.DB) (err error) {
	s.Id = "STG" + utils.Generate7DigitId() // Prefix for clarity
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return
}

// CropStage is the join table between Crop and Stage, defining the order and duration.
type CropStage struct {
	CropID       string    `json:"-" gorm:"primaryKey;type:varchar(10);not null"`
	StageID      string    `json:"id" gorm:"primaryKey;type:varchar(10);not null"`
	Order        int       `json:"order" gorm:"column:\"order\";not null"` // Quoting "order" is crucial
	Duration     int       `json:"duration"`
	DurationUnit string    `json:"duration_unit" gorm:"type:varchar(20);default:'DAYS'"`
	CreatedAt    time.Time `json:"-"`
	Stage        Stage     `json:"-" gorm:"foreignKey:StageID"` // Used for preloading stage details
}

// TableName sets the table name for the CropStage model.
func (CropStage) TableName() string {
	return "crop_stages"
}

func (c *Crop) BeforeCreate(tx *gorm.DB) (err error) {

	if c.Id == "" {
		c.Id = "CRP" + utils.Generate7DigitId() // Prefix for crop IDs
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

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
