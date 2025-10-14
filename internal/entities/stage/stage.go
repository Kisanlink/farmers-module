package stage

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Stage represents a master growth stage for crops
type Stage struct {
	base.BaseModel
	StageName   string         `json:"stage_name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description *string        `json:"description" gorm:"type:text"`
	Properties  entities.JSONB `json:"properties" gorm:"type:jsonb;not null;default:'{}';serializer:json"`
	IsActive    bool           `json:"is_active" gorm:"type:boolean;not null;default:true"`
}

// TableName returns the table name for the Stage model
func (s *Stage) TableName() string {
	return "stages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (s *Stage) GetTableIdentifier() string {
	return "STGE"
}

// GetTableSize returns the table size for ID generation
func (s *Stage) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewStage creates a new stage model with proper initialization
func NewStage() *Stage {
	baseModel := base.NewBaseModel("STGE", hash.Medium)
	return &Stage{
		BaseModel:  *baseModel,
		Properties: make(entities.JSONB),
		IsActive:   true,
	}
}

// Validate validates the stage model
func (s *Stage) Validate() error {
	if s.StageName == "" {
		return common.ErrInvalidInput
	}
	if len(s.StageName) > 100 {
		return common.ErrInvalidInput
	}
	return nil
}
