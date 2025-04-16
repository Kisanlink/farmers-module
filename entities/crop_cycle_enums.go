// this is not used in the code
package entities

import (
	"database/sql/driver"
	"fmt"
)

// -------------------
// Farm Activity Enum
// -------------------
type ActivityType string

const (
	ActivitySowing      ActivityType = "sowing"
	ActivityIrrigation  ActivityType = "irrigation"
	ActivityFertilizing ActivityType = "fertilizing"
	ActivityHarvesting  ActivityType = "harvesting"
	ActivityWeeding     ActivityType = "weeding"
)

func (a *ActivityType) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for ActivityType: %v", value)
	}
	*a = ActivityType(str)
	return nil
}

func (a ActivityType) Value() (driver.Value, error) {
	return string(a), nil
}

// -------------------
// Crop Category Enum
// -------------------
type CropCategory string

const (
	CategoryCereal    CropCategory = "cereal"
	CategoryLegume    CropCategory = "legume"
	CategoryVegetable CropCategory = "vegetable"
	CategoryFruit     CropCategory = "fruit"
	CategorySpice     CropCategory = "spice"
)

func (c *CropCategory) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for CropCategory: %v", value)
	}
	*c = CropCategory(str)
	return nil
}

func (c CropCategory) Value() (driver.Value, error) {
	return string(c), nil
}

// -------------------
// Crop Unit Enum
// -------------------
type CropUnit string

const (
	UnitKg    CropUnit = "kg"
	UnitTon   CropUnit = "tons"
	UnitLitre CropUnit = "liters"
	UnitBags  CropUnit = "bags"
)

func (u *CropUnit) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for CropUnit: %v", value)
	}
	*u = CropUnit(str)
	return nil
}

func (u CropUnit) Value() (driver.Value, error) {
	return string(u), nil
}
