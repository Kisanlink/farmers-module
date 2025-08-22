package fpo

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPORef represents a reference to an FPO organization
type FPORef struct {
	base.BaseModel
	AAAOrgID       string            `json:"aaa_org_id" gorm:"type:varchar(255);unique;not null"`
	Name           string            `json:"name" gorm:"type:varchar(255);not null"`
	RegistrationNo string            `json:"registration_no" gorm:"type:varchar(255)"`
	Status         string            `json:"status" gorm:"type:varchar(50);default:'ACTIVE'"`
	BusinessConfig map[string]string `json:"business_config" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the FPORef model
func (f *FPORef) TableName() string {
	return "fpo_refs"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FPORef) GetTableIdentifier() string {
	return "fpo_ref"
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
	return nil
}
