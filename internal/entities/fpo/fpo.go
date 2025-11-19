package fpo

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPOStatus represents the status of an FPO organization
type FPOStatus string

const (
	// FPOStatusDraft represents an FPO in draft state
	FPOStatusDraft FPOStatus = "DRAFT"

	// FPOStatusPendingVerification represents an FPO awaiting document verification
	FPOStatusPendingVerification FPOStatus = "PENDING_VERIFICATION"

	// FPOStatusVerified represents an FPO with verified documents
	FPOStatusVerified FPOStatus = "VERIFIED"

	// FPOStatusRejected represents an FPO with rejected verification
	FPOStatusRejected FPOStatus = "REJECTED"

	// FPOStatusPendingSetup represents an FPO with incomplete setup (partial failure during creation)
	FPOStatusPendingSetup FPOStatus = "PENDING_SETUP"

	// FPOStatusSetupFailed represents an FPO with failed setup
	FPOStatusSetupFailed FPOStatus = "SETUP_FAILED"

	// FPOStatusActive represents an active FPO with complete setup
	FPOStatusActive FPOStatus = "ACTIVE"

	// FPOStatusInactive represents a deactivated FPO
	FPOStatusInactive FPOStatus = "INACTIVE"

	// FPOStatusSuspended represents a temporarily suspended FPO
	FPOStatusSuspended FPOStatus = "SUSPENDED"

	// FPOStatusArchived represents an archived FPO
	FPOStatusArchived FPOStatus = "ARCHIVED"
)

// IsValid checks if the FPO status is valid
func (s FPOStatus) IsValid() bool {
	switch s {
	case FPOStatusDraft, FPOStatusPendingVerification, FPOStatusVerified,
		FPOStatusRejected, FPOStatusPendingSetup, FPOStatusSetupFailed,
		FPOStatusActive, FPOStatusInactive, FPOStatusSuspended, FPOStatusArchived:
		return true
	default:
		return false
	}
}

// String returns the string representation of FPO status
func (s FPOStatus) String() string {
	return string(s)
}

// CanTransitionTo checks if a status can transition to another status
func (s FPOStatus) CanTransitionTo(target FPOStatus) bool {
	transitions := map[FPOStatus][]FPOStatus{
		FPOStatusDraft:               {FPOStatusPendingVerification},
		FPOStatusPendingVerification: {FPOStatusVerified, FPOStatusRejected},
		FPOStatusVerified:            {FPOStatusPendingSetup},
		FPOStatusRejected:            {FPOStatusDraft, FPOStatusArchived},
		FPOStatusPendingSetup:        {FPOStatusActive, FPOStatusSetupFailed},
		FPOStatusSetupFailed:         {FPOStatusPendingSetup, FPOStatusInactive},
		FPOStatusActive:              {FPOStatusSuspended, FPOStatusInactive},
		FPOStatusSuspended:           {FPOStatusActive, FPOStatusInactive},
		FPOStatusInactive:            {FPOStatusArchived},
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// FPORef represents a reference to an FPO organization
type FPORef struct {
	base.BaseModel

	// Core Fields
	AAAOrgID       string `json:"aaa_org_id" gorm:"type:varchar(255);unique"`
	Name           string `json:"name" gorm:"type:varchar(255);not null"`
	RegistrationNo string `json:"registration_number" gorm:"type:varchar(255)"`

	// Lifecycle Fields
	Status          FPOStatus  `json:"status" gorm:"type:varchar(50);not null;default:'DRAFT'"`
	PreviousStatus  FPOStatus  `json:"previous_status" gorm:"type:varchar(50)"`
	StatusReason    string     `json:"status_reason" gorm:"type:text"`
	StatusChangedAt *time.Time `json:"status_changed_at"`
	StatusChangedBy string     `json:"status_changed_by"`

	// Verification Fields
	VerificationStatus string     `json:"verification_status" gorm:"type:varchar(50)"`
	VerifiedAt         *time.Time `json:"verified_at"`
	VerifiedBy         string     `json:"verified_by"`
	VerificationNotes  string     `json:"verification_notes" gorm:"type:text"`

	// Setup Tracking
	SetupAttempts int            `json:"setup_attempts" gorm:"default:0"`
	LastSetupAt   *time.Time     `json:"last_setup_at"`
	SetupErrors   entities.JSONB `json:"setup_errors,omitempty" gorm:"type:jsonb;serializer:json"` // Track partial setup failures
	SetupProgress entities.JSONB `json:"setup_progress,omitempty" gorm:"type:jsonb;serializer:json"`

	// Business Configuration
	BusinessConfig entities.JSONB `json:"business_config" gorm:"type:jsonb;default:'{}';serializer:json"`
	Metadata       entities.JSONB `json:"metadata,omitempty" gorm:"type:jsonb;serializer:json"`

	// Relationships
	CEOUserID   string  `json:"ceo_user_id" gorm:"type:varchar(255)"`
	ParentFPOID *string `json:"parent_fpo_id" gorm:"type:varchar(255)"`
}

// NewFPORef creates a new FPO reference with proper initialization
// This ensures ID is set correctly from the start for consistency
func NewFPORef(aaaOrgID string) *FPORef {
	return &FPORef{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID: aaaOrgID, // Set ID to aaa_org_id for lookups
			},
		},
		AAAOrgID:       aaaOrgID,
		Status:         FPOStatusDraft,               // Default status
		BusinessConfig: make(map[string]interface{}), // Initialize empty map
		Metadata:       make(map[string]interface{}), // Initialize empty map
		SetupAttempts:  0,                            // Initialize to 0
	}
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

// BeforeCreate is a GORM hook that syncs ID with AAAOrgID before creating
func (f *FPORef) BeforeCreate() error {
	// Set ID to AAAOrgID if not already set
	// This ensures the FPO ref can be looked up by the organization ID
	if f.AAAOrgID != "" && f.ID == "" {
		f.ID = f.AAAOrgID
	}
	return nil
}

// BeforeUpdate is a GORM hook that ensures ID stays synced with AAAOrgID
func (f *FPORef) BeforeUpdate() error {
	// Maintain ID-AAAOrgID sync on updates
	if f.AAAOrgID != "" && f.ID != f.AAAOrgID {
		f.ID = f.AAAOrgID
	}
	return nil
}
