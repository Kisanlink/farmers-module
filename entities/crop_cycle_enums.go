package entities

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ActivityType string

type activityTypes struct {
	SOWING      ActivityType
	IRRIGATION  ActivityType
	FERTILIZING ActivityType
	HARVESTING  ActivityType
	WEEDING     ActivityType
}

var ACTIVITY_TYPES = activityTypes{
	SOWING:      "SOWING",
	IRRIGATION:  "IRRIGATION",
	FERTILIZING: "FERTILIZING",
	HARVESTING:  "HARVESTING",
	WEEDING:     "WEEDING",
}

func (a *ActivityType) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for ActivityType: %v", value)
	}

	activity, err := ACTIVITY_TYPES.Parse(str)
	if err != nil {
		return err
	}
	*a = activity
	return nil
}

func (a ActivityType) Value() (driver.Value, error) {
	return string(a), nil
}

func (at activityTypes) All() []ActivityType {
	v := reflect.ValueOf(at)
	values := make([]ActivityType, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i).Interface().(ActivityType)
		values = append(values, val)
	}
	return values
}

func (at activityTypes) IsValid(activity string) bool {
	for _, act := range at.All() {
		if string(act) == activity {
			return true
		}
	}
	return false
}

func (at activityTypes) Parse(activity string) (ActivityType, error) {
	if at.IsValid(activity) {
		return ActivityType(activity), nil
	}
	return "", errors.New("invalid activity type. Valid values are: " + strings.Join(at.StringValues(), ", "))
}

func (at activityTypes) StringValues() []string {
	values := make([]string, 0, len(at.All()))
	for _, v := range at.All() {
		values = append(values, string(v))
	}
	return values
}

// -------------------
// Crop Category Enum
// -------------------
type CropCategory string

type cropCategories struct {
	CEREAL    CropCategory
	LEGUME    CropCategory
	VEGETABLE CropCategory
	FRUIT     CropCategory
	SPICE     CropCategory
}

var CROP_CATEGORIES = cropCategories{
	CEREAL:    "CEREAL",
	LEGUME:    "LEGUME",
	VEGETABLE: "VEGETABLE",
	FRUIT:     "FRUIT",
	SPICE:     "SPICE",
}

func (c *CropCategory) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for CropCategory: %v", value)
	}

	category, err := CROP_CATEGORIES.Parse(str)
	if err != nil {
		return err
	}
	*c = category
	return nil
}

func (c CropCategory) Value() (driver.Value, error) {
	return string(c), nil
}

func (cc cropCategories) All() []CropCategory {
	v := reflect.ValueOf(cc)
	values := make([]CropCategory, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i).Interface().(CropCategory)
		values = append(values, val)
	}
	return values
}

func (cc cropCategories) IsValid(category string) bool {
	for _, cat := range cc.All() {
		if string(cat) == category {
			return true
		}
	}
	return false
}

func (cc cropCategories) Parse(category string) (CropCategory, error) {
	if cc.IsValid(category) {
		return CropCategory(category), nil
	}
	return "", errors.New("invalid crop category: " + category)
}

// -------------------
// Crop Unit Enum
// -------------------
type CropUnit string

type cropUnits struct {
	KG    CropUnit
	TON   CropUnit
	LITRE CropUnit
	BAGS  CropUnit
}

var CROP_UNITS = cropUnits{
	KG:    "KILOGRAMS",
	TON:   "TONNES",
	LITRE: "LITRES",
	BAGS:  "BAGS",
}

func (u *CropUnit) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for CropUnit: %v", value)
	}

	unit, err := CROP_UNITS.Parse(str)
	if err != nil {
		return err
	}
	*u = unit
	return nil
}

func (u CropUnit) Value() (driver.Value, error) {
	return string(u), nil
}

func (cu cropUnits) All() []CropUnit {
	v := reflect.ValueOf(cu)
	values := make([]CropUnit, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		val := v.Field(i).Interface().(CropUnit)
		values = append(values, val)
	}
	return values
}

func (cu cropUnits) IsValid(unit string) bool {
	for _, u := range cu.All() {
		if string(u) == unit {
			return true
		}
	}
	return false
}

func (cu cropUnits) Parse(unit string) (CropUnit, error) {
	if cu.IsValid(unit) {
		return CropUnit(unit), nil
	}
	return "", errors.New("invalid crop unit: " + unit)
}

type CropCycleStatus string

type cropCycleStatuses struct {
	ONGOING   CropCycleStatus
	COMPLETED CropCycleStatus
}

var CROP_CYCLE_STATUSES = cropCycleStatuses{
	ONGOING:   "ONGOING",
	COMPLETED: "COMPLETED",
}

func (s *CropCycleStatus) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid data for CropCycleStatus: %v", value)
	}
	status, err := CROP_CYCLE_STATUSES.Parse(str)
	if err != nil {
		return err
	}
	*s = status
	return nil
}

func (s CropCycleStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (css cropCycleStatuses) All() []CropCycleStatus {
	v := reflect.ValueOf(css)
	values := make([]CropCycleStatus, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		values = append(values, v.Field(i).Interface().(CropCycleStatus))
	}
	return values
}

func (css cropCycleStatuses) IsValid(status string) bool {
	for _, st := range css.All() {
		if string(st) == status {
			return true
		}
	}
	return false
}

func (css cropCycleStatuses) Parse(status string) (CropCycleStatus, error) {
	if css.IsValid(status) {
		return CropCycleStatus(status), nil
	}
	return "", fmt.Errorf(
		"invalid crop cycle status: %s; valid values are: %s",
		status,
		strings.Join(css.StringValues(), ", "),
	)
}

func (css cropCycleStatuses) StringValues() []string {
	vals := make([]string, len(css.All()))
	for i, st := range css.All() {
		vals[i] = string(st)
	}
	return vals
}
