package requests

import (
	"mime/multipart"
)

// BulkFarmerAdditionRequest represents a request to add multiple farmers to an FPO
type BulkFarmerAdditionRequest struct {
	BaseRequest
	FPOOrgID       string                `json:"fpo_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	InputFormat    string                `json:"input_format" validate:"required,oneof=csv excel json" example:"csv"`
	Data           []byte                `json:"data,omitempty"`                                                                         // For direct data upload
	FileURL        string                `json:"file_url,omitempty" example:"https://storage.example.com/uploads/farmers_batch_001.csv"` // For file URL reference
	ProcessingMode string                `json:"processing_mode" validate:"required,oneof=sync async batch" example:"async"`
	Options        BulkProcessingOptions `json:"options"`
}

// BulkProcessingOptions represents options for bulk processing
type BulkProcessingOptions struct {
	ValidateOnly        bool                   `json:"validate_only" example:"false"`                                              // Dry run mode
	ContinueOnError     bool                   `json:"continue_on_error" example:"true"`                                           // Continue processing on individual failures
	ChunkSize           int                    `json:"chunk_size" example:"100"`                                                   // Size of processing chunks (default: 100)
	MaxConcurrency      int                    `json:"max_concurrency" example:"10"`                                               // Max parallel workers (default: 10)
	DeduplicationMode   string                 `json:"deduplication_mode" example:"skip"`                                          // How to handle duplicates: skip, update, error
	NotificationWebhook string                 `json:"notification_webhook" example:"https://webhook.example.com/bulk-completion"` // Webhook for completion notification
	AssignKisanSathi    bool                   `json:"assign_kisan_sathi" example:"true"`                                          // Auto-assign KisanSathi if available
	KisanSathiUserID    *string                `json:"kisan_sathi_user_id" example:"ks_123e4567-e89b-12d3-a456-426614174001"`      // Specific KisanSathi to assign
	SendCredentials     bool                   `json:"send_credentials" example:"true"`                                            // Send login credentials to farmers
	CredentialMethod    string                 `json:"credential_method" example:"sms"`                                            // sms, email, both
	Metadata            map[string]interface{} `json:"metadata"`                                                                   // Additional metadata
}

// FarmerBulkData represents individual farmer data for bulk processing
type FarmerBulkData struct {
	FirstName         string                 `json:"first_name" validate:"required,min=2,max=50" example:"Suresh"`
	LastName          string                 `json:"last_name" validate:"required,min=2,max=50" example:"Patel"`
	PhoneNumber       string                 `json:"phone_number" validate:"required,phone" example:"+91-9876543220"`
	Email             string                 `json:"email" validate:"omitempty,email" example:"suresh.patel@example.com"`
	DateOfBirth       string                 `json:"date_of_birth" validate:"omitempty,datetime=2006-01-02" example:"1975-08-20"`
	Gender            string                 `json:"gender" validate:"omitempty,oneof=male female other" example:"male"`
	StreetAddress     string                 `json:"street_address,omitempty" example:"Village Khandwa, Post Ratlam"`
	City              string                 `json:"city,omitempty" example:"Ratlam"`
	State             string                 `json:"state,omitempty" example:"Madhya Pradesh"`
	PostalCode        string                 `json:"postal_code,omitempty" example:"457001"`
	Country           string                 `json:"country,omitempty" example:"India"`
	LandOwnershipType string                 `json:"land_ownership_type,omitempty" example:"OWN"`
	CustomFields      map[string]interface{} `json:"custom_fields,omitempty" example:"education:high_school,family_size:5"`
	ExternalID        string                 `json:"external_id,omitempty" example:"EXT_FARMER_001"` // For tracking and idempotency
	Password          string                 `json:"password,omitempty" example:"Farmer@123"`        // Optional password, will be generated if not provided
}

// BulkFarmerFileUploadRequest represents a file upload request
type BulkFarmerFileUploadRequest struct {
	BaseRequest
	FPOOrgID       string                `form:"fpo_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	InputFormat    string                `form:"input_format" validate:"required,oneof=csv excel json" example:"csv"`
	File           *multipart.FileHeader `form:"file" validate:"required"`
	ProcessingMode string                `form:"processing_mode" validate:"required,oneof=sync async batch" example:"async"`
	Options        string                `form:"options" example:"{\"continue_on_error\":true,\"chunk_size\":100}"` // JSON string of BulkProcessingOptions
}

// GetBulkOperationStatusRequest represents a request to get bulk operation status
type GetBulkOperationStatusRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required" example:"op_123e4567-e89b-12d3-a456-426614174000"`
}

// CancelBulkOperationRequest represents a request to cancel a bulk operation
type CancelBulkOperationRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required" example:"op_123e4567-e89b-12d3-a456-426614174000"`
	Reason      string `json:"reason,omitempty" example:"Duplicate upload detected"`
}

// RetryBulkOperationRequest represents a request to retry failed records
type RetryBulkOperationRequest struct {
	BaseRequest
	OperationID   string                `json:"operation_id" validate:"required" example:"op_123e4567-e89b-12d3-a456-426614174000"`
	RetryAll      bool                  `json:"retry_all" example:"false"`
	RecordIndices []int                 `json:"record_indices,omitempty" example:"5,12,25,48"`
	Options       BulkProcessingOptions `json:"options,omitempty"`
}

// ValidateBulkDataRequest represents a request to validate bulk data
type ValidateBulkDataRequest struct {
	BaseRequest
	FPOOrgID    string           `json:"fpo_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	InputFormat string           `json:"input_format" validate:"required,oneof=csv excel json" example:"json"`
	Data        []byte           `json:"data,omitempty"`
	Farmers     []FarmerBulkData `json:"farmers,omitempty"`
}

// GetBulkTemplateRequest represents a request to get a bulk upload template
type GetBulkTemplateRequest struct {
	BaseRequest
	Format        string   `json:"format" validate:"required,oneof=csv excel" example:"csv"`
	IncludeSample bool     `json:"include_sample" example:"true"`
	CustomFields  []string `json:"custom_fields,omitempty" example:"farmer_id_card,bank_account"`
}

// DownloadBulkResultsRequest represents a request to download bulk operation results
type DownloadBulkResultsRequest struct {
	BaseRequest
	OperationID string `json:"operation_id" validate:"required" example:"op_123e4567-e89b-12d3-a456-426614174000"`
	Format      string `json:"format" validate:"required,oneof=csv excel json" example:"csv"`
	IncludeAll  bool   `json:"include_all" example:"false"` // Include all records or just failures
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
		o.Metadata = make(map[string]interface{})
	}
}
