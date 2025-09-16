package bulk

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// OperationStatus represents the status of a bulk operation
type OperationStatus string

const (
	StatusPending    OperationStatus = "PENDING"
	StatusProcessing OperationStatus = "PROCESSING"
	StatusCompleted  OperationStatus = "COMPLETED"
	StatusFailed     OperationStatus = "FAILED"
	StatusCancelled  OperationStatus = "CANCELLED"
)

// ProcessingMode represents the processing mode for bulk operations
type ProcessingMode string

const (
	ModeSync  ProcessingMode = "SYNC"
	ModeAsync ProcessingMode = "ASYNC"
	ModeBatch ProcessingMode = "BATCH"
)

// InputFormat represents the input file format
type InputFormat string

const (
	FormatCSV   InputFormat = "CSV"
	FormatExcel InputFormat = "EXCEL"
	FormatJSON  InputFormat = "JSON"
)

// BulkOperation represents a bulk farmer addition operation
type BulkOperation struct {
	base.BaseModel
	FPOOrgID          string                 `json:"fpo_org_id" gorm:"type:varchar(255);not null;index"`
	InitiatedBy       string                 `json:"initiated_by" gorm:"type:varchar(255);not null"`
	Status            OperationStatus        `json:"status" gorm:"type:varchar(50);not null;default:'PENDING';index"`
	InputFormat       InputFormat            `json:"input_format" gorm:"type:varchar(20);not null"`
	ProcessingMode    ProcessingMode         `json:"processing_mode" gorm:"type:varchar(20);not null"`
	TotalRecords      int                    `json:"total_records" gorm:"type:integer;not null;default:0"`
	ProcessedRecords  int                    `json:"processed_records" gorm:"type:integer;not null;default:0"`
	SuccessfulRecords int                    `json:"successful_records" gorm:"type:integer;not null;default:0"`
	FailedRecords     int                    `json:"failed_records" gorm:"type:integer;not null;default:0"`
	SkippedRecords    int                    `json:"skipped_records" gorm:"type:integer;not null;default:0"`
	StartTime         *time.Time             `json:"start_time"`
	EndTime           *time.Time             `json:"end_time"`
	ProcessingTime    int64                  `json:"processing_time" gorm:"type:bigint"` // in milliseconds
	ResultFileURL     string                 `json:"result_file_url" gorm:"type:text"`
	ErrorSummary      map[string]int         `json:"error_summary" gorm:"type:jsonb;default:'{}'"`
	Options           map[string]interface{} `json:"options" gorm:"type:jsonb;default:'{}'"`
	Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for BulkOperation
func (b *BulkOperation) TableName() string {
	return "bulk_operations"
}

// GetTableIdentifier returns the table identifier for ID generation
func (b *BulkOperation) GetTableIdentifier() string {
	return "bulk_op"
}

// GetTableSize returns the table size for ID generation
func (b *BulkOperation) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewBulkOperation creates a new bulk operation with proper initialization
func NewBulkOperation() *BulkOperation {
	baseModel := base.NewBaseModel("bulk_op", hash.Medium)
	return &BulkOperation{
		BaseModel:    *baseModel,
		Status:       StatusPending,
		ErrorSummary: make(map[string]int),
		Options:      make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
	}
}

// UpdateProgress updates the progress of the bulk operation
func (b *BulkOperation) UpdateProgress(processed, successful, failed, skipped int) {
	b.ProcessedRecords = processed
	b.SuccessfulRecords = successful
	b.FailedRecords = failed
	b.SkippedRecords = skipped

	if b.ProcessedRecords >= b.TotalRecords {
		b.Status = StatusCompleted
		now := time.Now()
		b.EndTime = &now
		if b.StartTime != nil {
			b.ProcessingTime = now.Sub(*b.StartTime).Milliseconds()
		}
	}
}

// GetProgressPercentage returns the progress percentage
func (b *BulkOperation) GetProgressPercentage() float64 {
	if b.TotalRecords == 0 {
		return 0
	}
	return float64(b.ProcessedRecords) / float64(b.TotalRecords) * 100
}

// IsComplete returns true if the operation is complete
func (b *BulkOperation) IsComplete() bool {
	return b.Status == StatusCompleted || b.Status == StatusFailed || b.Status == StatusCancelled
}

// CanRetry returns true if the operation can be retried
func (b *BulkOperation) CanRetry() bool {
	return b.Status == StatusFailed || b.Status == StatusCompleted && b.FailedRecords > 0
}
