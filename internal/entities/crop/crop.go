package crop

import (
	"encoding/json"

	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropCategory represents the category of a crop
type CropCategory string

const (
	CropCategoryCereal    CropCategory = "CEREAL"
	CropCategoryLegume    CropCategory = "LEGUME"
	CropCategoryVegetable CropCategory = "VEGETABLE"
	CropCategoryOilSeeds  CropCategory = "OIL_SEEDS"
	CropCategoryFruit     CropCategory = "FRUIT"
	CropCategorySpice     CropCategory = "SPICE"
)

// CropUnit represents the unit of measurement for crops
type CropUnit string

const (
	CropUnitKG      CropUnit = "KG"
	CropUnitQuintal CropUnit = "QUINTAL"
	CropUnitTonnes  CropUnit = "TONNES"
	CropUnitPieces  CropUnit = "PIECES"
)

// CropSeason represents the season when crops are grown
type CropSeason string

const (
	CropSeasonKharif    CropSeason = "KHARIF"
	CropSeasonRabi      CropSeason = "RABI"
	CropSeasonSummer    CropSeason = "SUMMER"
	CropSeasonPerennial CropSeason = "PERENNIAL"
)

// Crop represents a crop master data entity
type Crop struct {
	base.BaseModel
	Name             string                 `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Category         CropCategory           `json:"category" gorm:"type:varchar(50);not null"`
	CropDurationDays *int                   `json:"crop_duration_days" gorm:"type:integer"`
	TypicalUnits     []CropUnit             `json:"typical_units" gorm:"type:jsonb;default:'[]'"`
	Seasons          []CropSeason           `json:"seasons" gorm:"type:jsonb;default:'[]'"`
	ImageURL         *string                `json:"image_url" gorm:"type:varchar(500)"`
	DocumentID       *string                `json:"document_id" gorm:"type:varchar(255)"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the Crop model
func (c *Crop) TableName() string {
	return "crops"
}

// GetTableIdentifier returns the table identifier for ID generation
func (c *Crop) GetTableIdentifier() string {
	return "crop"
}

// GetTableSize returns the table size for ID generation
func (c *Crop) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validate validates the crop data
func (c *Crop) Validate() error {
	if c.Name == "" {
		return common.ErrInvalidCropData
	}
	if c.Category == "" {
		return common.ErrInvalidCropData
	}
	if !c.IsValidCategory() {
		return common.ErrInvalidCropData
	}
	if c.CropDurationDays != nil && *c.CropDurationDays <= 0 {
		return common.ErrInvalidCropData
	}
	return nil
}

// IsValidCategory checks if the category is valid
func (c *Crop) IsValidCategory() bool {
	return c.Category == CropCategoryCereal ||
		c.Category == CropCategoryLegume ||
		c.Category == CropCategoryVegetable ||
		c.Category == CropCategoryOilSeeds ||
		c.Category == CropCategoryFruit ||
		c.Category == CropCategorySpice
}

// IsValidUnit checks if the unit is valid
func (c *Crop) IsValidUnit(unit CropUnit) bool {
	return unit == CropUnitKG ||
		unit == CropUnitQuintal ||
		unit == CropUnitTonnes ||
		unit == CropUnitPieces
}

// IsValidSeason checks if the season is valid
func (c *Crop) IsValidSeason(season CropSeason) bool {
	return season == CropSeasonKharif ||
		season == CropSeasonRabi ||
		season == CropSeasonSummer ||
		season == CropSeasonPerennial
}

// GetValidCategories returns all valid crop categories
func GetValidCategories() []CropCategory {
	return []CropCategory{
		CropCategoryCereal,
		CropCategoryLegume,
		CropCategoryVegetable,
		CropCategoryOilSeeds,
		CropCategoryFruit,
		CropCategorySpice,
	}
}

// GetValidUnits returns all valid crop units
func GetValidUnits() []CropUnit {
	return []CropUnit{
		CropUnitKG,
		CropUnitQuintal,
		CropUnitTonnes,
		CropUnitPieces,
	}
}

// GetValidSeasons returns all valid crop seasons
func GetValidSeasons() []CropSeason {
	return []CropSeason{
		CropSeasonKharif,
		CropSeasonRabi,
		CropSeasonSummer,
		CropSeasonPerennial,
	}
}

// MarshalJSON custom JSON marshaling for slices
func (c *Crop) MarshalJSON() ([]byte, error) {
	type Alias Crop
	return json.Marshal(&struct {
		*Alias
		TypicalUnits []string `json:"typical_units"`
		Seasons      []string `json:"seasons"`
	}{
		Alias:        (*Alias)(c),
		TypicalUnits: convertUnitsToStrings(c.TypicalUnits),
		Seasons:      convertSeasonsToStrings(c.Seasons),
	})
}

// UnmarshalJSON custom JSON unmarshaling for slices
func (c *Crop) UnmarshalJSON(data []byte) error {
	type Alias Crop
	aux := &struct {
		*Alias
		TypicalUnits []string `json:"typical_units"`
		Seasons      []string `json:"seasons"`
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.TypicalUnits = convertStringsToUnits(aux.TypicalUnits)
	c.Seasons = convertStringsToSeasons(aux.Seasons)
	return nil
}

// Helper functions for JSON conversion
func convertUnitsToStrings(units []CropUnit) []string {
	result := make([]string, len(units))
	for i, unit := range units {
		result[i] = string(unit)
	}
	return result
}

func convertSeasonsToStrings(seasons []CropSeason) []string {
	result := make([]string, len(seasons))
	for i, season := range seasons {
		result[i] = string(season)
	}
	return result
}

func convertStringsToUnits(strs []string) []CropUnit {
	result := make([]CropUnit, len(strs))
	for i, str := range strs {
		result[i] = CropUnit(str)
	}
	return result
}

func convertStringsToSeasons(strs []string) []CropSeason {
	result := make([]CropSeason, len(strs))
	for i, str := range strs {
		result[i] = CropSeason(str)
	}
	return result
}

// NewCrop creates a new crop with proper initialization
func NewCrop() *Crop {
	baseModel := base.NewBaseModel("crop", hash.Medium)
	return &Crop{
		BaseModel:    *baseModel,
		TypicalUnits: []CropUnit{},
		Seasons:      []CropSeason{},
		Metadata:     make(map[string]interface{}),
	}
}
