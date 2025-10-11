package fpo

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPOStatus represents the status of an FPO organization
type FPOStatus string

const (
	// FPOStatusActive represents an active FPO with complete setup
	FPOStatusActive FPOStatus = "ACTIVE"

	// FPOStatusPendingSetup represents an FPO with incomplete setup (partial failure during creation)
	FPOStatusPendingSetup FPOStatus = "PENDING_SETUP"

	// FPOStatusInactive represents a deactivated FPO
	FPOStatusInactive FPOStatus = "INACTIVE"

	// FPOStatusSuspended represents a temporarily suspended FPO
	FPOStatusSuspended FPOStatus = "SUSPENDED"
)

// IsValid checks if the FPO status is valid
func (s FPOStatus) IsValid() bool {
	switch s {
	case FPOStatusActive, FPOStatusPendingSetup, FPOStatusInactive, FPOStatusSuspended:
		return true
	default:
		return false
	}
}

// String returns the string representation of FPO status
func (s FPOStatus) String() string {
	return string(s)
}

// FPORef represents a reference to an FPO organization
type FPORef struct {
	base.BaseModel
	AAAOrgID       string         `json:"aaa_org_id" gorm:"type:varchar(255);unique;not null"`
	Name           string         `json:"name" gorm:"type:varchar(255);not null"`
	RegistrationNo string         `json:"registration_no" gorm:"type:varchar(255)"`
	Status         FPOStatus      `json:"status" gorm:"type:varchar(50);default:'ACTIVE'"`
	BusinessConfig entities.JSONB `json:"business_config" gorm:"type:jsonb;default:'{}'"`
	SetupErrors    entities.JSONB `json:"setup_errors,omitempty" gorm:"type:jsonb"` // Track partial setup failures
}

// TableName returns the table name for the FPORef model
func (f *FPORef) TableName() string {
	return "fpo_refs"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FPORef) GetTableIdentifier() string {
	return "FPOR"
}

// GetTableSize returns the table size for ID generation
func (f *FPORef) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validate validates the FPORef model
func (f *FPORef) Validate() error {
	if f.AAAOrgID == "" {
		return common.ErrInvalidInput
	}
	if f.Name == "" {
		return common.ErrInvalidInput
	}
	if f.Status != "" && !f.Status.IsValid() {
		return fmt.Errorf("%w: invalid FPO status '%s'", common.ErrInvalidInput, f.Status)
	}
	return nil
}

// IsPendingSetup checks if the FPO is in PENDING_SETUP status
func (f *FPORef) IsPendingSetup() bool {
	return f.Status == FPOStatusPendingSetup
}

// IsActive checks if the FPO is in ACTIVE status
func (f *FPORef) IsActive() bool {
	return f.Status == FPOStatusActive
}

// CanRetrySetup checks if the FPO can retry setup operations
func (f *FPORef) CanRetrySetup() bool {
	return f.Status == FPOStatusPendingSetup
}
