package responses

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
)

// FPOConfigData represents FPO configuration data in responses
type FPOConfigData struct {
	ID              string         `json:"id"`
	AAAOrgID        string         `json:"aaa_org_id"`
	FPOName         string         `json:"fpo_name"`
	ERPBaseURL      string         `json:"erp_base_url"`
	ERPAPIVersion   string         `json:"erp_api_version"`
	Features        entities.JSONB `json:"features"`
	Contact         entities.JSONB `json:"contact"`
	BusinessHours   entities.JSONB `json:"business_hours"`
	Metadata        entities.JSONB `json:"metadata"`
	APIHealthStatus string         `json:"api_health_status"`
	LastSyncedAt    *time.Time     `json:"last_synced_at"`
	SyncInterval    int            `json:"sync_interval_minutes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// FPOConfigResponse represents a response for FPO configuration operations
type FPOConfigResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	Data      *FPOConfigData `json:"data,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

// FPOConfigListResponse represents a response for listing FPO configurations
type FPOConfigListResponse struct {
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
	Data       []*FPOConfigData `json:"data"`
	Pagination *PaginationInfo  `json:"pagination,omitempty"`
	RequestID  string           `json:"request_id,omitempty"`
}

// FPOHealthCheckData represents ERP health check data
type FPOHealthCheckData struct {
	AAAOrgID       string    `json:"aaa_org_id"`
	ERPBaseURL     string    `json:"erp_base_url"`
	Status         string    `json:"status"`
	ResponseTimeMs int64     `json:"response_time_ms,omitempty"`
	Error          string    `json:"error,omitempty"`
	LastChecked    time.Time `json:"last_checked"`
}

// FPOHealthCheckResponse represents a response for FPO health check
type FPOHealthCheckResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message,omitempty"`
	Data      *FPOHealthCheckData `json:"data"`
	RequestID string              `json:"request_id,omitempty"`
}

// NewFPOConfigResponse creates a new FPO config response
func NewFPOConfigResponse(data *FPOConfigData, message string) *FPOConfigResponse {
	return &FPOConfigResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewFPOConfigListResponse creates a new FPO config list response
func NewFPOConfigListResponse(data []*FPOConfigData, pagination *PaginationInfo, message string) *FPOConfigListResponse {
	return &FPOConfigListResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// NewFPOHealthCheckResponse creates a new FPO health check response
func NewFPOHealthCheckResponse(data *FPOHealthCheckData) *FPOHealthCheckResponse {
	return &FPOHealthCheckResponse{
		Success: true,
		Data:    data,
	}
}

// SetRequestID sets the request ID for the response
func (r *FPOConfigResponse) SetRequestID(requestID string) {
	r.RequestID = requestID
}

// SetRequestID sets the request ID for the list response
func (r *FPOConfigListResponse) SetRequestID(requestID string) {
	r.RequestID = requestID
}

// SetRequestID sets the request ID for the health check response
func (r *FPOHealthCheckResponse) SetRequestID(requestID string) {
	r.RequestID = requestID
}

// SwaggerFPOConfigResponse represents Swagger documentation for FPO config response
type SwaggerFPOConfigResponse struct {
	Success   bool           `json:"success" example:"true"`
	Message   string         `json:"message" example:"FPO configuration retrieved successfully"`
	Data      *FPOConfigData `json:"data"`
	RequestID string         `json:"request_id" example:"req-123"`
}

// SwaggerFPOConfigListResponse represents Swagger documentation for FPO config list response
type SwaggerFPOConfigListResponse struct {
	Success    bool             `json:"success" example:"true"`
	Message    string           `json:"message" example:"FPO configurations retrieved successfully"`
	Data       []*FPOConfigData `json:"data"`
	Pagination *PaginationInfo  `json:"pagination"`
	RequestID  string           `json:"request_id" example:"req-123"`
}

// SwaggerFPOHealthCheckResponse represents Swagger documentation for FPO health check response
type SwaggerFPOHealthCheckResponse struct {
	Success   bool                `json:"success" example:"true"`
	Data      *FPOHealthCheckData `json:"data"`
	RequestID string              `json:"request_id" example:"req-123"`
}
