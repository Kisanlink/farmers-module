package crop

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropCategory represents the category of crop
type CropCategory string

const (
	CropCategoryCereals    CropCategory = "CEREALS"
	CropCategoryPulses     CropCategory = "PULSES"
	CropCategoryVegetables CropCategory = "VEGETABLES"
	CropCategoryFruits     CropCategory = "FRUITS"
	CropCategoryOilSeeds   CropCategory = "OIL_SEEDS"
	CropCategorySpices     CropCategory = "SPICES"
	CropCategoryCashCrops  CropCategory = "CASH_CROPS"
	CropCategoryFodder     CropCategory = "FODDER"
	CropCategoryMedicinal  CropCategory = "MEDICINAL"
	CropCategoryOther      CropCategory = "OTHER"
)

// Season represents the growing season
type Season string

const (
	SeasonRabi   Season = "RABI"
	SeasonKharif Season = "KHARIF"
	SeasonZaid   Season = "ZAID"
)

// Crop represents a crop master data entity
type Crop struct {
	base.BaseModel
	Name           string         `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	ScientificName *string        `json:"scientific_name" gorm:"type:varchar(255)"`
	Category       CropCategory   `json:"category" gorm:"type:crop_category;not null"`
	DurationDays   *int           `json:"duration_days" gorm:"type:integer"`
	Seasons        []string       `json:"seasons" gorm:"type:jsonb;not null;default:'[]'"`
	Unit           string         `json:"unit" gorm:"type:varchar(50);not null;default:'kg'"`
	Properties     entities.JSONB `json:"properties" gorm:"type:jsonb;not null;default:'{}'"`
	IsActive       bool           `json:"is_active" gorm:"type:boolean;not null;default:true"`
}

// TableName returns the table name for the Crop model
func (c *Crop) TableName() string {
	return "crops"
}

// GetTableIdentifier returns the table identifier for ID generation
func (c *Crop) GetTableIdentifier() string {
	return "CROP"
}

// GetTableSize returns the table size for ID generation
func (c *Crop) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewCrop creates a new crop model with proper initialization
func NewCrop() *Crop {
	baseModel := base.NewBaseModel("CROP", hash.Medium)
	return &Crop{
		BaseModel:  *baseModel,
		Properties: make(entities.JSONB),
		Seasons:    make([]string, 0),
		Unit:       "kg",
		IsActive:   true,
	}
}

// Validate validates the crop model
func (c *Crop) Validate() error {
	if c.Name == "" {
		return common.ErrInvalidInput
	}
	if c.Category == "" {
		return common.ErrInvalidInput
	}
	if len(c.Seasons) == 0 {
		return common.ErrInvalidInput
	}
	if c.Unit == "" {
		return common.ErrInvalidInput
	}

	// Validate category
	validCategories := map[CropCategory]bool{
		CropCategoryCereals:    true,
		CropCategoryPulses:     true,
		CropCategoryVegetables: true,
		CropCategoryFruits:     true,
		CropCategoryOilSeeds:   true,
		CropCategorySpices:     true,
		CropCategoryCashCrops:  true,
		CropCategoryFodder:     true,
		CropCategoryMedicinal:  true,
		CropCategoryOther:      true,
	}
	if !validCategories[c.Category] {
		return common.ErrInvalidInput
	}

	// Validate seasons
	validSeasons := map[string]bool{
		string(SeasonRabi):   true,
		string(SeasonKharif): true,
		string(SeasonZaid):   true,
	}
	for _, season := range c.Seasons {
		if !validSeasons[season] {
			return common.ErrInvalidInput
		}
	}

	// Validate duration if provided
	if c.DurationDays != nil && *c.DurationDays <= 0 {
		return common.ErrInvalidInput
	}

	return nil
}

// GetValidCategories returns all valid crop categories
func GetValidCategories() []CropCategory {
	return []CropCategory{
		CropCategoryCereals,
		CropCategoryPulses,
		CropCategoryVegetables,
		CropCategoryFruits,
		CropCategoryOilSeeds,
		CropCategorySpices,
		CropCategoryCashCrops,
		CropCategoryFodder,
		CropCategoryMedicinal,
		CropCategoryOther,
	}
}

// GetValidSeasons returns all valid seasons
func GetValidSeasons() []Season {
	return []Season{
		SeasonRabi,
		SeasonKharif,
		SeasonZaid,
	}
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	validCategories := map[string]bool{
		string(CropCategoryCereals):    true,
		string(CropCategoryPulses):     true,
		string(CropCategoryVegetables): true,
		string(CropCategoryFruits):     true,
		string(CropCategoryOilSeeds):   true,
		string(CropCategorySpices):     true,
		string(CropCategoryCashCrops):  true,
		string(CropCategoryFodder):     true,
		string(CropCategoryMedicinal):  true,
		string(CropCategoryOther):      true,
	}
	return validCategories[category]
}

// IsValidSeason checks if a season is valid
func IsValidSeason(season string) bool {
	validSeasons := map[string]bool{
		string(SeasonRabi):   true,
		string(SeasonKharif): true,
		string(SeasonZaid):   true,
	}
	return validSeasons[season]
}
