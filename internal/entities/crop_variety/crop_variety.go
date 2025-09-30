package crop_variety

import (
	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropVariety represents a variety of a crop with specific characteristics
type CropVariety struct {
	base.BaseModel
	CropID       string            `json:"crop_id" gorm:"type:uuid;not null;index"`
	Name         string            `json:"name" gorm:"type:varchar(255);not null"`
	Description  *string           `json:"description" gorm:"type:text"`
	DurationDays *int              `json:"duration_days" gorm:"type:integer"`
	YieldPerAcre *float64          `json:"yield_per_acre" gorm:"type:numeric(10,2)"`
	YieldPerTree *float64          `json:"yield_per_tree" gorm:"type:numeric(10,2)"`
	Properties   map[string]string `json:"properties" gorm:"type:jsonb;not null;default:'{}'"`
	IsActive     bool              `json:"is_active" gorm:"type:boolean;not null;default:true"`

	// Relationships
	Crop crop.Crop `json:"crop" gorm:"foreignKey:CropID;references:ID"`
}

// TableName returns the table name for the CropVariety model
func (cv *CropVariety) TableName() string {
	return "crop_varieties"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cv *CropVariety) GetTableIdentifier() string {
	return "crop_variety"
}

// GetTableSize returns the table size for ID generation
func (cv *CropVariety) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewCropVariety creates a new crop variety model with proper initialization
func NewCropVariety() *CropVariety {
	baseModel := base.NewBaseModel("crop_variety", hash.Medium)
	return &CropVariety{
		BaseModel:  *baseModel,
		Properties: make(map[string]string),
		IsActive:   true,
	}
}

// Validate validates the crop variety model
func (cv *CropVariety) Validate() error {
	if cv.CropID == "" {
		return common.ErrInvalidInput
	}
	if cv.Name == "" {
		return common.ErrInvalidInput
	}

	// Validate duration if provided
	if cv.DurationDays != nil && *cv.DurationDays <= 0 {
		return common.ErrInvalidInput
	}

	// Validate yield per acre if provided
	if cv.YieldPerAcre != nil && *cv.YieldPerAcre < 0 {
		return common.ErrInvalidInput
	}

	// Validate yield per tree if provided
	if cv.YieldPerTree != nil && *cv.YieldPerTree < 0 {
		return common.ErrInvalidInput
	}

	return nil
}

// GetDurationDays returns the duration days, falling back to crop default if not specified
func (cv *CropVariety) GetDurationDays() *int {
	if cv.DurationDays != nil {
		return cv.DurationDays
	}
	if cv.Crop.DurationDays != nil {
		return cv.Crop.DurationDays
	}
	return nil
}

// GetEffectiveDuration returns the effective duration for this variety
// considering both variety-specific and crop-level defaults
func (cv *CropVariety) GetEffectiveDuration() int {
	if cv.DurationDays != nil && *cv.DurationDays > 0 {
		return *cv.DurationDays
	}
	if cv.Crop.DurationDays != nil && *cv.Crop.DurationDays > 0 {
		return *cv.Crop.DurationDays
	}
	return 0 // Unknown duration
}