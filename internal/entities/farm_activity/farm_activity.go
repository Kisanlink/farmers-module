package farm_activity

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/stage"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmActivity represents an individual activity within a crop cycle
type FarmActivity struct {
	base.BaseModel
	CropCycleID  string         `json:"crop_cycle_id" gorm:"type:varchar(255);not null;index"`
	CropStageID  *string        `json:"crop_stage_id" gorm:"type:varchar(20);index:idx_farm_activities_cycle_stage"`
	FarmerID     string         `json:"farmer_id" gorm:"type:varchar(255);not null;index"`
	ActivityType string         `json:"activity_type" gorm:"type:varchar(255);not null"`
	PlannedAt    *time.Time     `json:"planned_at" gorm:"type:timestamptz"`
	CompletedAt  *time.Time     `json:"completed_at" gorm:"type:timestamptz"`
	CreatedBy    string         `json:"created_by" gorm:"type:varchar(255);not null"`
	Status       string         `json:"status" gorm:"type:activity_status;not null;default:'PLANNED'"`
	Output       entities.JSONB `json:"output" gorm:"type:jsonb;default:'{}';serializer:json"`
	Metadata     entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`

	// Relationships
	Farmer    *farmer.Farmer   `json:"farmer,omitempty" gorm:"foreignKey:FarmerID;references:ID;constraint:OnDelete:CASCADE"`
	CropStage *stage.CropStage `json:"crop_stage,omitempty" gorm:"foreignKey:CropStageID;references:ID"`
}

// TableName returns the table name for the FarmActivity model
func (fa *FarmActivity) TableName() string {
	return "farm_activities"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fa *FarmActivity) GetTableIdentifier() string {
	return "FACT"
}

// GetTableSize returns the table size for ID generation
func (fa *FarmActivity) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (fa *FarmActivity) Validate() error {
	if fa.CropCycleID == "" {
		return common.ErrInvalidFarmActivityData
	}
	if fa.FarmerID == "" {
		return common.ErrInvalidFarmActivityData
	}
	if fa.ActivityType == "" {
		return common.ErrInvalidFarmActivityData
	}
	if fa.CreatedBy == "" {
		return common.ErrInvalidFarmActivityData
	}
	return nil
}
