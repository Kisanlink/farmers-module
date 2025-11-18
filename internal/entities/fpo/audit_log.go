package fpo

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPOAuditLog represents an audit log entry for FPO lifecycle changes
type FPOAuditLog struct {
	base.BaseModel

	FPOID         string         `json:"fpo_id" gorm:"type:varchar(255);not null;index"`
	Action        string         `json:"action" gorm:"type:varchar(100);not null"`
	PreviousState FPOStatus      `json:"previous_state" gorm:"type:varchar(50)"`
	NewState      FPOStatus      `json:"new_state" gorm:"type:varchar(50)"`
	Reason        string         `json:"reason" gorm:"type:text"`
	PerformedBy   string         `json:"performed_by" gorm:"type:varchar(255);not null"`
	PerformedAt   time.Time      `json:"performed_at" gorm:"not null"`
	Details       entities.JSONB `json:"details,omitempty" gorm:"type:jsonb;serializer:json"`
	RequestID     string         `json:"request_id" gorm:"type:varchar(255)"`
}

// TableName returns the table name for the FPOAuditLog model
func (a *FPOAuditLog) TableName() string {
	return "fpo_audit_logs"
}

// GetTableIdentifier returns the table identifier for ID generation
func (a *FPOAuditLog) GetTableIdentifier() string {
	return "FPOA"
}

// GetTableSize returns the table size for ID generation
func (a *FPOAuditLog) GetTableSize() hash.TableSize {
	return hash.Medium
}
