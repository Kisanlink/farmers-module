package farm_activity

import (
	"time"

	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmActivity represents an individual activity within a crop cycle
type FarmActivity struct {
	base.BaseModel
	CycleID      string            `json:"cycle_id" gorm:"type:uuid;not null"`
	ActivityType string            `json:"activity_type" gorm:"type:varchar(255);not null"`
	PlannedAt    *time.Time        `json:"planned_at" gorm:"type:timestamptz"`
	CompletedAt  *time.Time        `json:"completed_at" gorm:"type:timestamptz"`
	Metadata     map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	Output       map[string]string `json:"output" gorm:"type:jsonb"`
}

// TableName returns the table name for the FarmActivity model
func (fa *FarmActivity) TableName() string {
	return "farm_activities"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fa *FarmActivity) GetTableIdentifier() string {
	return "farm_activity"
}

// GetTableSize returns the table size for ID generation
func (fa *FarmActivity) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (fa *FarmActivity) Validate() error {
	if fa.CycleID == "" {
		return common.ErrInvalidFarmActivityData
	}
	if fa.ActivityType == "" {
		return common.ErrInvalidFarmActivityData
	}
	return nil
}
