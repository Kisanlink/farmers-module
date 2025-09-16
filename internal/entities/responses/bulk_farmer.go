package responses

import (
	"time"
)

// BulkOperationResponse represents the response for a bulk operation
type BulkOperationResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
	Data      *BulkOperationData `json:"data,omitempty"`
}

// BulkOperationData contains the bulk operation details
type BulkOperationData struct {
	OperationID         string     `json:"operation_id"`
	Status              string     `json:"status"`
	StatusURL           string     `json:"status_url"`
	ResultURL           string     `json:"result_url,omitempty"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
	Message             string     `json:"message"`
}

// BulkOperationStatusResponse represents the status response for a bulk operation
type BulkOperationStatusResponse struct {
	Success   bool                     `json:"success"`
	Message   string                   `json:"message"`
	RequestID string                   `json:"request_id,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
	Data      *BulkOperationStatusData `json:"data,omitempty"`
}

// BulkOperationStatusData contains detailed status information
type BulkOperationStatusData struct {
	OperationID         string                 `json:"operation_id"`
	FPOOrgID            string                 `json:"fpo_org_id"`
	Status              string                 `json:"status"`
	Progress            ProgressInfo           `json:"progress"`
	StartTime           *time.Time             `json:"start_time,omitempty"`
	EndTime             *time.Time             `json:"end_time,omitempty"`
	ProcessingTime      string                 `json:"processing_time,omitempty"`
	EstimatedCompletion *time.Time             `json:"estimated_completion,omitempty"`
	CurrentBatch        int                    `json:"current_batch,omitempty"`
	TotalBatches        int                    `json:"total_batches,omitempty"`
	ErrorSummary        map[string]int         `json:"error_summary,omitempty"`
	ResultFileURL       string                 `json:"result_file_url,omitempty"`
	CanRetry            bool                   `json:"can_retry"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressInfo contains progress information
type ProgressInfo struct {
	Total      int     `json:"total"`
	Processed  int     `json:"processed"`
	Successful int     `json:"successful"`
	Failed     int     `json:"failed"`
	Skipped    int     `json:"skipped"`
	Percentage float64 `json:"percentage"`
}

// BulkValidationResponse represents the validation response for bulk data
type BulkValidationResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id,omitempty"`
	Timestamp time.Time           `json:"timestamp"`
	Data      *BulkValidationData `json:"data,omitempty"`
}

// BulkValidationData contains validation results
type BulkValidationData struct {
	IsValid      bool                   `json:"is_valid"`
	TotalRecords int                    `json:"total_records"`
	ValidRecords int                    `json:"valid_records"`
	Errors       []ValidationError      `json:"errors,omitempty"`
	Warnings     []ValidationWarning    `json:"warnings,omitempty"`
	Summary      map[string]interface{} `json:"summary,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	RecordIndex int         `json:"record_index"`
	Field       string      `json:"field"`
	Value       interface{} `json:"value,omitempty"`
	Message     string      `json:"message"`
	Code        string      `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	RecordIndex int    `json:"record_index"`
	Field       string `json:"field,omitempty"`
	Message     string `json:"message"`
	Code        string `json:"code"`
}

// BulkProcessingDetailResponse represents individual record processing details
type BulkProcessingDetailResponse struct {
	Success   bool                      `json:"success"`
	Message   string                    `json:"message"`
	RequestID string                    `json:"request_id,omitempty"`
	Timestamp time.Time                 `json:"timestamp"`
	Data      *BulkProcessingDetailData `json:"data,omitempty"`
}

// BulkProcessingDetailData contains processing details for individual records
type BulkProcessingDetailData struct {
	Details []ProcessingDetail `json:"details"`
	Summary ProcessingSummary  `json:"summary"`
}

// ProcessingDetail represents the detail of a single record processing
type ProcessingDetail struct {
	RecordIndex    int                    `json:"record_index"`
	ExternalID     string                 `json:"external_id,omitempty"`
	Status         string                 `json:"status"`
	FarmerID       string                 `json:"farmer_id,omitempty"`
	AAAUserID      string                 `json:"aaa_user_id,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ErrorCode      string                 `json:"error_code,omitempty"`
	ProcessedAt    *time.Time             `json:"processed_at,omitempty"`
	ProcessingTime string                 `json:"processing_time,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	InputData      map[string]interface{} `json:"input_data,omitempty"`
}

// ProcessingSummary contains a summary of processing details
type ProcessingSummary struct {
	TotalProcessed int            `json:"total_processed"`
	Successful     int            `json:"successful"`
	Failed         int            `json:"failed"`
	Skipped        int            `json:"skipped"`
	ErrorTypes     map[string]int `json:"error_types,omitempty"`
}

// BulkTemplateResponse represents the response for bulk template request
type BulkTemplateResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Data      *BulkTemplateData `json:"data,omitempty"`
}

// BulkTemplateData contains template information
type BulkTemplateData struct {
	Format       string      `json:"format"`
	FileName     string      `json:"file_name"`
	Content      []byte      `json:"content,omitempty"`
	DownloadURL  string      `json:"download_url,omitempty"`
	Fields       []FieldInfo `json:"fields"`
	Instructions string      `json:"instructions"`
}

// FieldInfo contains information about a template field
type FieldInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Format      string `json:"format,omitempty"`
	Example     string `json:"example,omitempty"`
	Description string `json:"description,omitempty"`
}

// NewBulkOperationResponse creates a new bulk operation response
func NewBulkOperationResponse(data *BulkOperationData, message string) *BulkOperationResponse {
	return &BulkOperationResponse{
		Success:   true,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewBulkOperationStatusResponse creates a new bulk operation status response
func NewBulkOperationStatusResponse(data *BulkOperationStatusData) *BulkOperationStatusResponse {
	return &BulkOperationStatusResponse{
		Success:   true,
		Message:   "Operation status retrieved successfully",
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewBulkValidationResponse creates a new bulk validation response
func NewBulkValidationResponse(data *BulkValidationData) *BulkValidationResponse {
	message := "Validation completed successfully"
	if !data.IsValid {
		message = "Validation failed"
	}
	return &BulkValidationResponse{
		Success:   data.IsValid,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewBulkTemplateResponse creates a new bulk template response
func NewBulkTemplateResponse(data *BulkTemplateData) *BulkTemplateResponse {
	return &BulkTemplateResponse{
		Success:   true,
		Message:   "Template generated successfully",
		Timestamp: time.Now(),
		Data:      data,
	}
}
