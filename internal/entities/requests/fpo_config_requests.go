package requests

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
)

// CreateFPOConfigRequest represents a request to create FPO configuration
// Minimal configuration with API endpoint, UI link, contact, business hours and metadata
type CreateFPOConfigRequest struct {
	BaseRequest
	AAAOrgID      string         `json:"aaa_org_id" binding:"required"`
	FPOName       string         `json:"fpo_name" binding:"required"`
	ERPBaseURL    string         `json:"erp_base_url" binding:"required"`
	ERPUIBaseURL  string         `json:"erp_ui_base_url"`
	Contact       entities.JSONB `json:"contact"`
	BusinessHours entities.JSONB `json:"business_hours"`
	Metadata      entities.JSONB `json:"metadata"`
}

// UpdateFPOConfigRequest represents a request to update FPO configuration
// All fields are optional - only send fields you want to update
type UpdateFPOConfigRequest struct {
	BaseRequest
	FPOName       *string        `json:"fpo_name"`
	ERPBaseURL    *string        `json:"erp_base_url"`
	ERPUIBaseURL  *string        `json:"erp_ui_base_url"`
	Contact       entities.JSONB `json:"contact"`
	BusinessHours entities.JSONB `json:"business_hours"`
	Metadata      entities.JSONB `json:"metadata"`
}

// ListFPOConfigsRequest represents a request to list FPO configurations
type ListFPOConfigsRequest struct {
	BaseRequest
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	Status   string `form:"status"`
}

// GetFPOConfigRequest represents a request to get FPO configuration
type GetFPOConfigRequest struct {
	BaseRequest
	AAAOrgID string `uri:"aaa_org_id" binding:"required"`
}

// SetDefaults sets default values for CreateFPOConfigRequest
func (r *CreateFPOConfigRequest) SetDefaults() {
	if r.Contact == nil {
		r.Contact = entities.JSONB{}
	}
	if r.BusinessHours == nil {
		r.BusinessHours = entities.JSONB{}
	}
	if r.Metadata == nil {
		r.Metadata = entities.JSONB{}
	}
}

// SetDefaults sets default values for UpdateFPOConfigRequest
func (r *UpdateFPOConfigRequest) SetDefaults() {
	// Nothing to set for now
}

// SetDefaults sets default values for ListFPOConfigsRequest
func (r *ListFPOConfigsRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
}

// SetDefaults sets default values for GetFPOConfigRequest
func (r *GetFPOConfigRequest) SetDefaults() {
	// Nothing to set for now
}

// Validate validates CreateFPOConfigRequest
func (r *CreateFPOConfigRequest) Validate() error {
	if r.AAAOrgID == "" {
		return fmt.Errorf("aaa_org_id is required")
	}
	if r.FPOName == "" {
		return fmt.Errorf("fpo_name is required")
	}
	if r.ERPBaseURL == "" {
		return fmt.Errorf("erp_base_url is required")
	}
	// Basic URL validation for ERPBaseURL
	if len(r.ERPBaseURL) < 10 || (r.ERPBaseURL[:7] != "http://" && r.ERPBaseURL[:8] != "https://") {
		return fmt.Errorf("erp_base_url must be a valid URL")
	}
	// Basic URL validation for ERPUIBaseURL if provided
	if r.ERPUIBaseURL != "" {
		if len(r.ERPUIBaseURL) < 10 || (r.ERPUIBaseURL[:7] != "http://" && r.ERPUIBaseURL[:8] != "https://") {
			return fmt.Errorf("erp_ui_base_url must be a valid URL")
		}
	}
	return nil
}

// Validate validates UpdateFPOConfigRequest
func (r *UpdateFPOConfigRequest) Validate() error {
	if r.ERPBaseURL != nil {
		if *r.ERPBaseURL == "" {
			return fmt.Errorf("erp_base_url cannot be empty")
		}
		// Basic URL validation
		if len(*r.ERPBaseURL) < 10 || ((*r.ERPBaseURL)[:7] != "http://" && (*r.ERPBaseURL)[:8] != "https://") {
			return fmt.Errorf("erp_base_url must be a valid URL")
		}
	}
	if r.ERPUIBaseURL != nil && *r.ERPUIBaseURL != "" {
		// Basic URL validation
		if len(*r.ERPUIBaseURL) < 10 || ((*r.ERPUIBaseURL)[:7] != "http://" && (*r.ERPUIBaseURL)[:8] != "https://") {
			return fmt.Errorf("erp_ui_base_url must be a valid URL")
		}
	}
	return nil
}

// Validate validates GetFPOConfigRequest
func (r *GetFPOConfigRequest) Validate() error {
	if r.AAAOrgID == "" {
		return fmt.Errorf("aaa_org_id is required")
	}
	return nil
}

// Validate validates ListFPOConfigsRequest
func (r *ListFPOConfigsRequest) Validate() error {
	if r.Page < 0 {
		return fmt.Errorf("page must be >= 0")
	}
	if r.PageSize < 0 || r.PageSize > 100 {
		return fmt.Errorf("page_size must be between 0 and 100")
	}
	return nil
}
