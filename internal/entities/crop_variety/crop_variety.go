package crop_variety

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropVariety represents a variety of a specific crop
type CropVariety struct {
	base.BaseModel
	CropID          string                 `json:"crop_id" gorm:"type:varchar(255);not null;index"`
	VarietyName     string                 `json:"variety_name" gorm:"type:varchar(100);not null"`
	DurationDays    *int                   `json:"duration_days" gorm:"type:integer"`
	Characteristics *string                `json:"characteristics" gorm:"type:text"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
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

// Validate validates the crop variety data
func (cv *CropVariety) Validate() error {
	if cv.CropID == "" {
		return common.ErrInvalidCropVarietyData
	}
	if cv.VarietyName == "" {
		return common.ErrInvalidCropVarietyData
	}
	if cv.DurationDays != nil && *cv.DurationDays <= 0 {
		return common.ErrInvalidCropVarietyData
	}
	return nil
}

// NewCropVariety creates a new crop variety with proper initialization
func NewCropVariety() *CropVariety {
	baseModel := base.NewBaseModel("crop_variety", hash.Medium)
	return &CropVariety{
		BaseModel: *baseModel,
		Metadata:  make(map[string]interface{}),
	}
}
