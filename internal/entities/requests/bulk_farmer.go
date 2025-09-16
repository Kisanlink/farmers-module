package requests

import (
	"mime/multipart"
)

// BulkFarmerAdditionRequest represents a request to add multiple farmers to an FPO
type BulkFarmerAdditionRequest struct {
	BaseRequest
	FPOOrgID       string                `json:"fpo_org_id" validate:"required"`
	InputFormat    string                `json:"input_format" validate:"required,oneof=csv excel json"`
	Data           []byte                `json:"data,omitempty"`     // For direct data upload
	FileURL        string                `json:"file_url,omitempty"` // For file URL reference
	ProcessingMode string                `json:"processing_mode" validate:"required,oneof=sync async batch"`
	Options        BulkProcessingOptions `json:"options"`
}

// BulkProcessingOptions represents options for bulk processing
type BulkProcessingOptions struct {
	ValidateOnly        bool              `json:"validate_only"`        // Dry run mode
	ContinueOnError     bool              `json:"continue_on_error"`    // Continue processing on individual failures
	ChunkSize           int               `json:"chunk_size"`           // Size of processing chunks (default: 100)
	MaxConcurrency      int               `json:"max_concurrency"`      // Max parallel workers (default: 10)
	DeduplicationMode   string            `json:"deduplication_mode"`   // How to handle duplicates: skip, update, error
	NotificationWebhook string            `json:"notification_webhook"` // Webhook for completion notification
	AssignKisanSathi    bool              `json:"assign_kisan_sathi"`   // Auto-assign KisanSathi if available
	KisanSathiUserID    *string           `json:"kisan_sathi_user_id"`  // Specific KisanSathi to assign
	SendCredentials     bool              `json:"send_credentials"`     // Send login credentials to farmers
	CredentialMethod    string            `json:"credential_method"`    // sms, email, both
	Metadata            map[string]string `json:"metadata"`             // Additional metadata
}

// FarmerBulkData represents individual farmer data for bulk processing
type FarmerBulkData struct {
	FirstName         string            `json:"first_name" validate:"required,min=2,max=50"`
	LastName          string            `json:"last_name" validate:"required,min=2,max=50"`
	PhoneNumber       string            `json:"phone_number" validate:"required,phone"`
	Email             string            `json:"email" validate:"omitempty,email"`
	DateOfBirth       string            `json:"date_of_birth" validate:"omitempty,datetime=2006-01-02"`
	Gender            string            `json:"gender" validate:"omitempty,oneof=male female other"`
	StreetAddress     string            `json:"street_address,omitempty"`
	City              string            `json:"city,omitempty"`
	State             string            `json:"state,omitempty"`
	PostalCode        string            `json:"postal_code,omitempty"`
	Country           string            `json:"country,omitempty"`
	LandOwnershipType string            `json:"land_ownership_type,omitempty"`
	CustomFields      map[string]string `json:"custom_fields,omitempty"`
	ExternalID        string            `json:"external_id,omitempty"` // For tracking and idempotency
	Password          string            `json:"password,omitempty"`    // Optional password, will be generated if not provided
}

// BulkFarmerFileUploadRequest represents a file upload request
type BulkFarmerFileUploadRequest struct {
	BaseRequest
	FPOOrgID       string                `form:"fpo_org_id" validate:"required"`
	InputFormat    string                `form:"input_format" validate:"required,oneof=csv excel json"`
	File           *multipart.FileHeader `form:"file" validate:"required"`
	ProcessingMode string                `form:"processing_mode" validate:"required,oneof=sync async batch"`
	Options        string                `form:"options"` // JSON string of BulkProcessingOptions
}

// GetBulkOperationStatusRequest represents a request to get bulk operation status
type GetBulkOperationStatusRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required"`
}

// CancelBulkOperationRequest represents a request to cancel a bulk operation
type CancelBulkOperationRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required"`
	Reason      string `json:"reason,omitempty"`
}

// RetryBulkOperationRequest represents a request to retry failed records
type RetryBulkOperationRequest struct {
	BaseRequest
	OperationID   string                `json:"operation_id" validate:"required"`
	RetryAll      bool                  `json:"retry_all"`
	RecordIndices []int                 `json:"record_indices,omitempty"`
	Options       BulkProcessingOptions `json:"options,omitempty"`
}

// ValidateBulkDataRequest represents a request to validate bulk data
type ValidateBulkDataRequest struct {
	BaseRequest
	FPOOrgID    string           `json:"fpo_org_id" validate:"required"`
	InputFormat string           `json:"input_format" validate:"required,oneof=csv excel json"`
	Data        []byte           `json:"data,omitempty"`
	Farmers     []FarmerBulkData `json:"farmers,omitempty"`
}

// GetBulkTemplateRequest represents a request to get a bulk upload template
type GetBulkTemplateRequest struct {
	BaseRequest
	Format        string   `json:"format" validate:"required,oneof=csv excel"`
	IncludeSample bool     `json:"include_sample"`
	CustomFields  []string `json:"custom_fields,omitempty"`
}

// DownloadBulkResultsRequest represents a request to download bulk operation results
type DownloadBulkResultsRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required"`
	Format      string `json:"format" validate:"required,oneof=csv excel json"`
	IncludeAll  bool   `json:"include_all"` // Include all records or just failures
}

// NewBulkFarmerAdditionRequest creates a new bulk farmer addition request
func NewBulkFarmerAdditionRequest() BulkFarmerAdditionRequest {
	return BulkFarmerAdditionRequest{
		BaseRequest: NewBaseRequest(),
		Options: BulkProcessingOptions{
			ChunkSize:         100,
			MaxConcurrency:    10,
			ContinueOnError:   true,
			DeduplicationMode: "skip",
			CredentialMethod:  "sms",
		},
	}
}

// SetDefaults sets default values for BulkProcessingOptions
func (o *BulkProcessingOptions) SetDefaults() {
	if o.ChunkSize == 0 {
		o.ChunkSize = 100
	}
	if o.MaxConcurrency == 0 {
		o.MaxConcurrency = 10
	}
	if o.DeduplicationMode == "" {
		o.DeduplicationMode = "skip"
	}
	if o.CredentialMethod == "" {
		o.CredentialMethod = "sms"
	}
	if o.Metadata == nil {
		o.Metadata = make(map[string]string)
	}
}
