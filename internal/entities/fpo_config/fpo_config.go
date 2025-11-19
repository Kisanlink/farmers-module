package fpo_config

import (
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FPOConfig represents FPO configuration for e-commerce integration
type FPOConfig struct {
	base.BaseModel
	AAAOrgID        string         `json:"aaa_org_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	FPOName         string         `json:"fpo_name" gorm:"type:varchar(255);not null"`
	ERPBaseURL      string         `json:"erp_base_url" gorm:"type:varchar(500);not null"`
	ERPAPIVersion   string         `json:"erp_api_version" gorm:"type:varchar(10);default:'v1'"`
	Features        entities.JSONB `json:"features" gorm:"type:jsonb;default:'{}';serializer:json"`
	Contact         entities.JSONB `json:"contact" gorm:"type:jsonb;default:'{}';serializer:json"`
	BusinessHours   entities.JSONB `json:"business_hours" gorm:"type:jsonb;default:'{}';serializer:json"`
	Metadata        entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`
	APIHealthStatus string         `json:"api_health_status" gorm:"type:varchar(50);default:'unknown'"`
	LastSyncedAt    *time.Time     `json:"last_synced_at" gorm:"type:timestamp"`
	SyncInterval    int            `json:"sync_interval_minutes" gorm:"type:integer;default:5"`
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
		AAAOrgID:        aaaOrgID,
		ERPAPIVersion:   "v1",                         // Default version
		Features:        make(map[string]interface{}), // Initialize empty map
		Contact:         make(map[string]interface{}), // Initialize empty map
		BusinessHours:   make(map[string]interface{}), // Initialize empty map
		Metadata:        make(map[string]interface{}), // Initialize empty map
		APIHealthStatus: "unknown",                    // Default status
		SyncInterval:    30,                           // Default 30 minutes
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
	// Basic URL validation
	if len(f.ERPBaseURL) < 10 || (f.ERPBaseURL[:7] != "http://" && f.ERPBaseURL[:8] != "https://") {
		return fmt.Errorf("%w: erp_base_url must be a valid URL", common.ErrInvalidInput)
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

// IsHealthy checks if the FPO's ERP service is healthy
func (f *FPOConfig) IsHealthy() bool {
	return f.APIHealthStatus == "healthy"
}

// FeaturesData represents the features configuration
type FeaturesData struct {
	InventoryRealTime bool `json:"inventory_real_time"`
	CreditLimitCheck  bool `json:"credit_limit_check"`
	BatchOperations   bool `json:"batch_operations"`
	MinOrderValue     int  `json:"min_order_value"`
	MaxOrderValue     int  `json:"max_order_value"`
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

// MetadataData represents additional metadata
type MetadataData struct {
	LastSyncedAt        time.Time `json:"last_synced_at"`
	SyncIntervalMinutes int       `json:"sync_interval_minutes"`
	APIHealthStatus     string    `json:"api_health_status"`
}
