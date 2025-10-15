package bulk

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// RecordStatus represents the status of an individual record processing
type RecordStatus string

const (
	RecordStatusSuccess RecordStatus = "SUCCESS"
	RecordStatusFailed  RecordStatus = "FAILED"
	RecordStatusSkipped RecordStatus = "SKIPPED"
	RecordStatusPending RecordStatus = "PENDING"
)

// ProcessingDetail represents the processing detail of each record in a bulk operation
type ProcessingDetail struct {
	base.BaseModel
	BulkOperationID string                 `json:"bulk_operation_id" gorm:"type:varchar(255);not null;index:idx_processing_details_op_idx,priority:1;index:idx_processing_details_op_status,priority:1;index:idx_processing_details_op_retry,priority:1"`
	RecordIndex     int                    `json:"record_index" gorm:"type:integer;not null;index:idx_processing_details_op_idx,priority:2"`
	ExternalID      string                 `json:"external_id" gorm:"type:varchar(255);index:idx_processing_details_external"`
	Status          RecordStatus           `json:"status" gorm:"type:varchar(50);not null;default:'PENDING';index:idx_processing_details_op_status,priority:2;index:idx_processing_details_op_retry,priority:2,where:status = 'FAILED'"`
	FarmerID        *string                `json:"farmer_id" gorm:"type:varchar(255);index:idx_processing_details_farmer"`
	AAAUserID       *string                `json:"aaa_user_id" gorm:"type:varchar(255);index:idx_processing_details_aaa"`
	InputData       map[string]interface{} `json:"input_data" gorm:"type:jsonb;serializer:json"`
	Error           *string                `json:"error" gorm:"type:text"`
	ErrorCode       *string                `json:"error_code" gorm:"type:varchar(100)"`
	ProcessedAt     *time.Time             `json:"processed_at"`
	ProcessingTime  int64                  `json:"processing_time" gorm:"type:bigint"` // in milliseconds
	RetryCount      int                    `json:"retry_count" gorm:"type:integer;not null;default:0;index:idx_processing_details_op_retry,priority:3,where:status = 'FAILED'"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}';serializer:json"`
}

// TableName returns the table name for ProcessingDetail
func (p *ProcessingDetail) TableName() string {
	return "bulk_processing_details"
}

// GetTableIdentifier returns the table identifier for ID generation
func (p *ProcessingDetail) GetTableIdentifier() string {
	return "BLKD"
}

// GetTableSize returns the table size for ID generation
func (p *ProcessingDetail) GetTableSize() hash.TableSize {
	return hash.Large
}

// NewProcessingDetail creates a new processing detail with proper initialization
func NewProcessingDetail(bulkOperationID string, recordIndex int) *ProcessingDetail {
	baseModel := base.NewBaseModel("BLKD", hash.Large)
	return &ProcessingDetail{
		BaseModel:       *baseModel,
		BulkOperationID: bulkOperationID,
		RecordIndex:     recordIndex,
		Status:          RecordStatusPending,
		InputData:       make(map[string]interface{}),
		Metadata:        make(map[string]interface{}),
	}
}

// SetSuccess marks the processing detail as successful
func (p *ProcessingDetail) SetSuccess(farmerID, aaaUserID string) {
	p.Status = RecordStatusSuccess
	p.FarmerID = &farmerID
	p.AAAUserID = &aaaUserID
	now := time.Now()
	p.ProcessedAt = &now
}

// SetFailed marks the processing detail as failed
func (p *ProcessingDetail) SetFailed(error string, errorCode string) {
	p.Status = RecordStatusFailed
	p.Error = &error
	p.ErrorCode = &errorCode
	now := time.Now()
	p.ProcessedAt = &now
}

// SetSkipped marks the processing detail as skipped
func (p *ProcessingDetail) SetSkipped(reason string) {
	p.Status = RecordStatusSkipped
	p.Error = &reason
	now := time.Now()
	p.ProcessedAt = &now
}

// CanRetry returns true if the record can be retried
func (p *ProcessingDetail) CanRetry() bool {
	return p.Status == RecordStatusFailed && p.RetryCount < 3
}

// IncrementRetry increments the retry count
func (p *ProcessingDetail) IncrementRetry() {
	p.RetryCount++
}
