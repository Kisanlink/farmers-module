package fpo_config

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPOConfig represents FPO configuration for e-commerce integration
// Minimal configuration with API endpoint, UI link, contact, business hours and metadata
type FPOConfig struct {
	base.BaseModel
	AAAOrgID      string         `json:"aaa_org_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	FPOName       string         `json:"fpo_name" gorm:"type:varchar(255);not null"`
	ERPBaseURL    string         `json:"erp_base_url" gorm:"type:varchar(500);not null"`
	ERPUIBaseURL  string         `json:"erp_ui_base_url" gorm:"type:varchar(500)"`
	Contact       entities.JSONB `json:"contact" gorm:"type:jsonb;default:'{}';serializer:json"`
	BusinessHours entities.JSONB `json:"business_hours" gorm:"type:jsonb;default:'{}';serializer:json"`
	Metadata      entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`
}

// NewFPOConfig creates a new FPO configuration with proper initialization
// This ensures ID is set correctly from the start for consistency
func NewFPOConfig(aaaOrgID string) *FPOConfig {
	return &FPOConfig{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID: aaaOrgID, // Set ID to aaa_org_id for lookups
			},
		},
		AAAOrgID:      aaaOrgID,
		Contact:       make(map[string]interface{}), // Initialize empty map
		BusinessHours: make(map[string]interface{}), // Initialize empty map
		Metadata:      make(map[string]interface{}), // Initialize empty map
	}
}

// TableName returns the table name for the FPOConfig model
func (f *FPOConfig) TableName() string {
	return "fpo_configs"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FPOConfig) GetTableIdentifier() string {
	return "FPOC"
}

// GetTableSize returns the table size for ID generation
func (f *FPOConfig) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validate validates the FPOConfig model
func (f *FPOConfig) Validate() error {
	if f.AAAOrgID == "" {
		return fmt.Errorf("%w: aaa_org_id is required", common.ErrInvalidInput)
	}
	if f.FPOName == "" {
		return fmt.Errorf("%w: fpo_name is required", common.ErrInvalidInput)
	}
	if f.ERPBaseURL == "" {
		return fmt.Errorf("%w: erp_base_url is required", common.ErrInvalidInput)
	}
	// Basic URL validation for ERPBaseURL
	if len(f.ERPBaseURL) < 10 || (f.ERPBaseURL[:7] != "http://" && f.ERPBaseURL[:8] != "https://") {
		return fmt.Errorf("%w: erp_base_url must be a valid URL", common.ErrInvalidInput)
	}
	// Basic URL validation for ERPUIBaseURL if provided
	if f.ERPUIBaseURL != "" {
		if len(f.ERPUIBaseURL) < 10 || (f.ERPUIBaseURL[:7] != "http://" && f.ERPUIBaseURL[:8] != "https://") {
			return fmt.Errorf("%w: erp_ui_base_url must be a valid URL", common.ErrInvalidInput)
		}
	}
	return nil
}

// BeforeCreate is a GORM hook that syncs ID with AAAOrgID before creating
func (f *FPOConfig) BeforeCreate() error {
	// Set ID to AAAOrgID if not already set
	// This ensures the config can be looked up by the organization ID
	if f.AAAOrgID != "" && f.ID == "" {
		f.ID = f.AAAOrgID
	}
	return nil
}

// BeforeUpdate is a GORM hook that ensures ID stays synced with AAAOrgID
func (f *FPOConfig) BeforeUpdate() error {
	// Maintain ID-AAAOrgID sync on updates
	if f.AAAOrgID != "" && f.ID != f.AAAOrgID {
		f.ID = f.AAAOrgID
	}
	return nil
}

// ContactData represents the contact information
type ContactData struct {
	AdminName  string `json:"admin_name"`
	AdminPhone string `json:"admin_phone"`
	AdminEmail string `json:"admin_email"`
}

// BusinessHoursData represents the business hours configuration
type BusinessHoursData struct {
	Timezone    string   `json:"timezone"`
	OpenTime    string   `json:"open_time"`
	CloseTime   string   `json:"close_time"`
	WorkingDays []string `json:"working_days"`
}
