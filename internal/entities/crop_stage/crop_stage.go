package crop_stage

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// CropStage represents a growth stage of a specific crop
type CropStage struct {
	base.BaseModel
	CropID              string                 `json:"crop_id" gorm:"type:varchar(255);not null;index"`
	StageName           string                 `json:"stage_name" gorm:"type:varchar(100);not null"`
	StageOrder          int                    `json:"stage_order" gorm:"type:integer;not null"`
	TypicalDurationDays *int                   `json:"typical_duration_days" gorm:"type:integer"`
	Description         *string                `json:"description" gorm:"type:text"`
	Metadata            map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the CropStage model
func (cs *CropStage) TableName() string {
	return "crop_stages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cs *CropStage) GetTableIdentifier() string {
	return "crop_stage"
}

// GetTableSize returns the table size for ID generation
func (cs *CropStage) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validate validates the crop stage data
func (cs *CropStage) Validate() error {
	if cs.CropID == "" {
		return common.ErrInvalidCropStageData
	}
	if cs.StageName == "" {
		return common.ErrInvalidCropStageData
	}
	if cs.StageOrder < 0 {
		return common.ErrInvalidCropStageData
	}
	if cs.TypicalDurationDays != nil && *cs.TypicalDurationDays <= 0 {
		return common.ErrInvalidCropStageData
	}
	return nil
}

// NewCropStage creates a new crop stage with proper initialization
func NewCropStage() *CropStage {
	baseModel := base.NewBaseModel("crop_stage", hash.Medium)
	return &CropStage{
		BaseModel: *baseModel,
		Metadata:  make(map[string]interface{}),
	}
}

// GetCommonStages returns common crop growth stages
func GetCommonStages() []string {
	return []string{
		"Nursery",
		"Vegetative Stage",
		"Tillering Stage",
		"Panicle Initiation Stage",
		"Flowering Stage",
		"Grain Filling Stage",
		"Harvesting Stage",
	}
}
