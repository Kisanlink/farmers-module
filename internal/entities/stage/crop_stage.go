package stage

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// DurationUnit represents the unit of duration
type DurationUnit string

const (
	DurationUnitDays   DurationUnit = "DAYS"
	DurationUnitWeeks  DurationUnit = "WEEKS"
	DurationUnitMonths DurationUnit = "MONTHS"
)

// CropStage represents the relationship between crop and stage
type CropStage struct {
	base.BaseModel
	CropID       string         `json:"crop_id" gorm:"type:varchar(20);not null;index"`
	StageID      string         `json:"stage_id" gorm:"type:varchar(20);not null;index"`
	StageOrder   int            `json:"stage_order" gorm:"type:integer;not null;index"`
	DurationDays *int           `json:"duration_days" gorm:"type:integer"`
	DurationUnit DurationUnit   `json:"duration_unit" gorm:"type:varchar(20);not null;default:'DAYS'"`
	Properties   entities.JSONB `json:"properties" gorm:"type:jsonb;not null;default:'{}';serializer:json"`
	IsActive     bool           `json:"is_active" gorm:"type:boolean;not null;default:true"`

	// Relationships
	Crop  *cropEntity.Crop `json:"crop,omitempty" gorm:"foreignKey:CropID"`
	Stage *Stage           `json:"stage,omitempty" gorm:"foreignKey:StageID"`
}

// TableName returns the table name for the CropStage model
func (cs *CropStage) TableName() string {
	return "crop_stages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cs *CropStage) GetTableIdentifier() string {
	return "CSTG"
}

// GetTableSize returns the table size for ID generation
func (cs *CropStage) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewCropStage creates a new crop stage model with proper initialization
func NewCropStage() *CropStage {
	baseModel := base.NewBaseModel("CSTG", hash.Medium)
	return &CropStage{
		BaseModel:    *baseModel,
		Properties:   make(entities.JSONB),
		DurationUnit: DurationUnitDays,
		IsActive:     true,
	}
}

// Validate validates the crop stage model
func (cs *CropStage) Validate() error {
	if cs.CropID == "" || cs.StageID == "" {
		return common.ErrInvalidInput
	}
	if cs.StageOrder < 1 {
		return common.ErrInvalidInput
	}
	if cs.DurationDays != nil && *cs.DurationDays <= 0 {
		return common.ErrInvalidInput
	}

	// Validate duration unit
	validUnits := map[DurationUnit]bool{
		DurationUnitDays:   true,
		DurationUnitWeeks:  true,
		DurationUnitMonths: true,
	}
	if !validUnits[cs.DurationUnit] {
		return common.ErrInvalidInput
	}

	return nil
}

// GetValidDurationUnits returns all valid duration units
func GetValidDurationUnits() []DurationUnit {
	return []DurationUnit{
		DurationUnitDays,
		DurationUnitWeeks,
		DurationUnitMonths,
	}
}
