package crop_cycle

import (
	"time"

	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropType represents the type of crop (annual or perennial)
type CropType string

const (
	CropTypeAnnual    CropType = "ANNUAL"
	CropTypePerennial CropType = "PERENNIAL"
)

// CropCycle represents an agricultural cycle within a farm
type CropCycle struct {
	base.BaseModel
	FarmID   string `json:"farm_id" gorm:"type:varchar(255);not null"`
	FarmerID string `json:"farmer_id" gorm:"type:varchar(255);not null"`
	Season   string `json:"season" gorm:"type:season;not null"`
	Status   string `json:"status" gorm:"type:cycle_status;not null;default:'PLANNED'"`

	// Enhanced crop details
	CropID      *string  `json:"crop_id" gorm:"type:varchar(255)"`
	VarietyID   *string  `json:"variety_id" gorm:"type:varchar(255)"`
	CropType    CropType `json:"crop_type" gorm:"type:varchar(20);not null;default:'ANNUAL'"`
	CropName    *string  `json:"crop_name" gorm:"type:varchar(100)"`
	VarietyName *string  `json:"variety_name" gorm:"type:varchar(100)"`

	// Area and planting details
	Acreage       *float64 `json:"acreage" gorm:"type:decimal(12,4)"`
	NumberOfTrees *int     `json:"number_of_trees" gorm:"type:integer"`

	// Date fields
	SowingTransplantingDate *time.Time `json:"sowing_transplanting_date" gorm:"type:date"`
	HarvestDate             *time.Time `json:"harvest_date" gorm:"type:date"`
	StartDate               *time.Time `json:"start_date" gorm:"type:date"`
	EndDate                 *time.Time `json:"end_date" gorm:"type:date"`

	// Yield information
	YieldPerAcre *float64 `json:"yield_per_acre" gorm:"type:decimal(12,4)"`
	YieldPerTree *float64 `json:"yield_per_tree" gorm:"type:decimal(12,4)"`
	TotalYield   *float64 `json:"total_yield" gorm:"type:decimal(12,4)"`
	YieldUnit    *string  `json:"yield_unit" gorm:"type:varchar(20)"`

	// Tree age information (for perennial crops)
	TreeAgeRangeMin *int `json:"tree_age_range_min" gorm:"type:integer"`
	TreeAgeRangeMax *int `json:"tree_age_range_max" gorm:"type:integer"`

	// Media and documentation
	ImageURL   *string `json:"image_url" gorm:"type:varchar(500)"`
	DocumentID *string `json:"document_id" gorm:"type:varchar(255)"`

	// Additional data
	PlannedCrops []string               `json:"planned_crops" gorm:"type:jsonb;default:'[]'"`
	Outcome      map[string]string      `json:"outcome" gorm:"type:jsonb;default:'{}'"`
	ReportData   map[string]interface{} `json:"report_data" gorm:"type:jsonb;default:'{}'"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the CropCycle model
func (cc *CropCycle) TableName() string {
	return "crop_cycles"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cc *CropCycle) GetTableIdentifier() string {
	return "crop_cycle"
}

// GetTableSize returns the table size for ID generation
func (cc *CropCycle) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (cc *CropCycle) Validate() error {
	if cc.FarmID == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.FarmerID == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.Season == "" {
		return common.ErrInvalidCropCycleData
	}
	if cc.CropType != CropTypeAnnual && cc.CropType != CropTypePerennial {
		return common.ErrInvalidCropCycleData
	}
	if cc.Acreage != nil && *cc.Acreage <= 0 {
		return common.ErrInvalidCropCycleData
	}
	if cc.NumberOfTrees != nil && *cc.NumberOfTrees <= 0 {
		return common.ErrInvalidCropCycleData
	}
	if cc.YieldPerAcre != nil && *cc.YieldPerAcre < 0 {
		return common.ErrInvalidCropCycleData
	}
	if cc.YieldPerTree != nil && *cc.YieldPerTree < 0 {
		return common.ErrInvalidCropCycleData
	}
	if cc.TotalYield != nil && *cc.TotalYield < 0 {
		return common.ErrInvalidCropCycleData
	}
	if cc.TreeAgeRangeMin != nil && cc.TreeAgeRangeMax != nil && *cc.TreeAgeRangeMin > *cc.TreeAgeRangeMax {
		return common.ErrInvalidCropCycleData
	}
	return nil
}

// IsValidCropType checks if the crop type is valid
func (cc *CropCycle) IsValidCropType() bool {
	return cc.CropType == CropTypeAnnual || cc.CropType == CropTypePerennial
}

// GetValidCropTypes returns all valid crop types
func GetValidCropTypes() []CropType {
	return []CropType{
		CropTypeAnnual,
		CropTypePerennial,
	}
}

// NewCropCycle creates a new crop cycle with proper initialization
func NewCropCycle() *CropCycle {
	baseModel := base.NewBaseModel("crop_cycle", hash.Medium)
	return &CropCycle{
		BaseModel:    *baseModel,
		CropType:     CropTypeAnnual, // default to annual
		PlannedCrops: []string{},
		Outcome:      make(map[string]string),
		ReportData:   make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}
}
