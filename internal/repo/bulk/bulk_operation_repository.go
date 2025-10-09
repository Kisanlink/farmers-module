package bulk

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BulkOperationRepository defines the interface for bulk operation repository
type BulkOperationRepository interface {
	Create(ctx context.Context, operation *bulk.BulkOperation) error
	GetByID(ctx context.Context, id string) (*bulk.BulkOperation, error)
	Update(ctx context.Context, operation *bulk.BulkOperation) error
	UpdateStatus(ctx context.Context, id string, status bulk.OperationStatus) error
	UpdateProgress(ctx context.Context, id string, processed, successful, failed, skipped int) error
	List(ctx context.Context, filter *base.Filter) ([]*bulk.BulkOperation, error)
	ListByFPO(ctx context.Context, fpoOrgID string, filter *base.Filter) ([]*bulk.BulkOperation, error)
	Delete(ctx context.Context, id string) error
	GetActiveOperations(ctx context.Context, fpoOrgID string) ([]*bulk.BulkOperation, error)
}

// BulkOperationRepositoryImpl implements BulkOperationRepository
type BulkOperationRepositoryImpl struct {
	db *gorm.DB
}

// NewBulkOperationRepository creates a new bulk operation repository
func NewBulkOperationRepository(db *gorm.DB) BulkOperationRepository {
	return &BulkOperationRepositoryImpl{
		db: db,
	}
}

// Create creates a new bulk operation
func (r *BulkOperationRepositoryImpl) Create(ctx context.Context, operation *bulk.BulkOperation) error {
	if operation.ID == "" {
		baseModel := base.NewBaseModel("BLKO", hash.Medium)
		operation.ID = baseModel.ID
		operation.CreatedAt = baseModel.CreatedAt
		operation.UpdatedAt = baseModel.UpdatedAt
	}

	if err := r.db.WithContext(ctx).Create(operation).Error; err != nil {
		return fmt.Errorf("failed to create bulk operation: %w", err)
	}
	return nil
}

// GetByID retrieves a bulk operation by ID
func (r *BulkOperationRepositoryImpl) GetByID(ctx context.Context, id string) (*bulk.BulkOperation, error) {
	var operation bulk.BulkOperation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&operation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bulk operation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get bulk operation: %w", err)
	}
	return &operation, nil
}

// Update updates a bulk operation
func (r *BulkOperationRepositoryImpl) Update(ctx context.Context, operation *bulk.BulkOperation) error {
	if err := r.db.WithContext(ctx).Save(operation).Error; err != nil {
		return fmt.Errorf("failed to update bulk operation: %w", err)
	}
	return nil
}

// UpdateStatus updates the status of a bulk operation
func (r *BulkOperationRepositoryImpl) UpdateStatus(ctx context.Context, id string, status bulk.OperationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == bulk.StatusProcessing {
		updates["start_time"] = gorm.Expr("COALESCE(start_time, NOW())")
	} else if status == bulk.StatusCompleted || status == bulk.StatusFailed || status == bulk.StatusCancelled {
		updates["end_time"] = gorm.Expr("NOW()")
		updates["processing_time"] = gorm.Expr("EXTRACT(EPOCH FROM (NOW() - COALESCE(start_time, created_at))) * 1000")
	}

	if err := r.db.WithContext(ctx).Model(&bulk.BulkOperation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update bulk operation status: %w", err)
	}
	return nil
}

// UpdateProgress updates the progress of a bulk operation
func (r *BulkOperationRepositoryImpl) UpdateProgress(ctx context.Context, id string, processed, successful, failed, skipped int) error {
	updates := map[string]interface{}{
		"processed_records":  processed,
		"successful_records": successful,
		"failed_records":     failed,
		"skipped_records":    skipped,
	}

	// Check if operation is complete
	var operation bulk.BulkOperation
	if err := r.db.WithContext(ctx).Select("total_records").Where("id = ?", id).First(&operation).Error; err == nil {
		if processed >= operation.TotalRecords {
			updates["status"] = bulk.StatusCompleted
			updates["end_time"] = gorm.Expr("NOW()")
			updates["processing_time"] = gorm.Expr("EXTRACT(EPOCH FROM (NOW() - COALESCE(start_time, created_at))) * 1000")
		}
	}

	if err := r.db.WithContext(ctx).Model(&bulk.BulkOperation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update bulk operation progress: %w", err)
	}
	return nil
}

// List retrieves bulk operations based on filter
func (r *BulkOperationRepositoryImpl) List(ctx context.Context, filter *base.Filter) ([]*bulk.BulkOperation, error) {
	var operations []*bulk.BulkOperation

	query := r.db.WithContext(ctx)

	if filter != nil {
		query = applyFilter(query, filter)
	}

	if err := query.Find(&operations).Error; err != nil {
		return nil, fmt.Errorf("failed to list bulk operations: %w", err)
	}

	return operations, nil
}

// ListByFPO retrieves bulk operations for a specific FPO
func (r *BulkOperationRepositoryImpl) ListByFPO(ctx context.Context, fpoOrgID string, filter *base.Filter) ([]*bulk.BulkOperation, error) {
	var operations []*bulk.BulkOperation

	query := r.db.WithContext(ctx).Where("fpo_org_id = ?", fpoOrgID)

	if filter != nil {
		query = applyFilter(query, filter)
	}

	if err := query.Find(&operations).Error; err != nil {
		return nil, fmt.Errorf("failed to list bulk operations by FPO: %w", err)
	}

	return operations, nil
}

// Delete deletes a bulk operation
func (r *BulkOperationRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&bulk.BulkOperation{}).Error; err != nil {
		return fmt.Errorf("failed to delete bulk operation: %w", err)
	}
	return nil
}

// GetActiveOperations retrieves active bulk operations for an FPO
func (r *BulkOperationRepositoryImpl) GetActiveOperations(ctx context.Context, fpoOrgID string) ([]*bulk.BulkOperation, error) {
	var operations []*bulk.BulkOperation

	if err := r.db.WithContext(ctx).
		Where("fpo_org_id = ? AND status IN ?", fpoOrgID, []string{
			string(bulk.StatusPending),
			string(bulk.StatusProcessing),
		}).
		Order("created_at DESC").
		Find(&operations).Error; err != nil {
		return nil, fmt.Errorf("failed to get active operations: %w", err)
	}

	return operations, nil
}

// ProcessingDetailRepository defines the interface for processing detail repository
type ProcessingDetailRepository interface {
	Create(ctx context.Context, detail *bulk.ProcessingDetail) error
	CreateBatch(ctx context.Context, details []*bulk.ProcessingDetail) error
	GetByID(ctx context.Context, id string) (*bulk.ProcessingDetail, error)
	GetByOperationID(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error)
	GetByStatus(ctx context.Context, operationID string, status bulk.RecordStatus) ([]*bulk.ProcessingDetail, error)
	Update(ctx context.Context, detail *bulk.ProcessingDetail) error
	UpdateBatch(ctx context.Context, details []*bulk.ProcessingDetail) error
	GetFailedRecords(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error)
	GetRetryableRecords(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error)
}

// ProcessingDetailRepositoryImpl implements ProcessingDetailRepository
type ProcessingDetailRepositoryImpl struct {
	db *gorm.DB
}

// NewProcessingDetailRepository creates a new processing detail repository
func NewProcessingDetailRepository(db *gorm.DB) ProcessingDetailRepository {
	return &ProcessingDetailRepositoryImpl{
		db: db,
	}
}

// Create creates a new processing detail
func (r *ProcessingDetailRepositoryImpl) Create(ctx context.Context, detail *bulk.ProcessingDetail) error {
	if detail.ID == "" {
		baseModel := base.NewBaseModel("BKDT", hash.Large)
		detail.ID = baseModel.ID
		detail.CreatedAt = baseModel.CreatedAt
		detail.UpdatedAt = baseModel.UpdatedAt
	}

	if err := r.db.WithContext(ctx).Create(detail).Error; err != nil {
		return fmt.Errorf("failed to create processing detail: %w", err)
	}
	return nil
}

// CreateBatch creates multiple processing details in batch
func (r *ProcessingDetailRepositoryImpl) CreateBatch(ctx context.Context, details []*bulk.ProcessingDetail) error {
	if len(details) == 0 {
		return nil
	}

	// Assign IDs to details without IDs
	for _, detail := range details {
		if detail.ID == "" {
			baseModel := base.NewBaseModel("BKDT", hash.Large)
			detail.ID = baseModel.ID
			detail.CreatedAt = baseModel.CreatedAt
			detail.UpdatedAt = baseModel.UpdatedAt
		}
	}

	// Create in batches of 100 for better performance
	batchSize := 100
	for i := 0; i < len(details); i += batchSize {
		end := i + batchSize
		if end > len(details) {
			end = len(details)
		}

		batch := details[i:end]
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(batch, len(batch)).Error; err != nil {
			return fmt.Errorf("failed to create processing details batch: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a processing detail by ID
func (r *ProcessingDetailRepositoryImpl) GetByID(ctx context.Context, id string) (*bulk.ProcessingDetail, error) {
	var detail bulk.ProcessingDetail
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&detail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("processing detail not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get processing detail: %w", err)
	}
	return &detail, nil
}

// GetByOperationID retrieves all processing details for an operation
func (r *ProcessingDetailRepositoryImpl) GetByOperationID(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error) {
	var details []*bulk.ProcessingDetail
	if err := r.db.WithContext(ctx).
		Where("bulk_operation_id = ?", operationID).
		Order("record_index ASC").
		Find(&details).Error; err != nil {
		return nil, fmt.Errorf("failed to get processing details: %w", err)
	}
	return details, nil
}

// GetByStatus retrieves processing details by status
func (r *ProcessingDetailRepositoryImpl) GetByStatus(ctx context.Context, operationID string, status bulk.RecordStatus) ([]*bulk.ProcessingDetail, error) {
	var details []*bulk.ProcessingDetail
	if err := r.db.WithContext(ctx).
		Where("bulk_operation_id = ? AND status = ?", operationID, status).
		Order("record_index ASC").
		Find(&details).Error; err != nil {
		return nil, fmt.Errorf("failed to get processing details by status: %w", err)
	}
	return details, nil
}

// Update updates a processing detail
func (r *ProcessingDetailRepositoryImpl) Update(ctx context.Context, detail *bulk.ProcessingDetail) error {
	if err := r.db.WithContext(ctx).Save(detail).Error; err != nil {
		return fmt.Errorf("failed to update processing detail: %w", err)
	}
	return nil
}

// UpdateBatch updates multiple processing details
func (r *ProcessingDetailRepositoryImpl) UpdateBatch(ctx context.Context, details []*bulk.ProcessingDetail) error {
	if len(details) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, detail := range details {
		if err := tx.Save(detail).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update processing detail batch: %w", err)
		}
	}

	return tx.Commit().Error
}

// GetFailedRecords retrieves failed processing records
func (r *ProcessingDetailRepositoryImpl) GetFailedRecords(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error) {
	return r.GetByStatus(ctx, operationID, bulk.RecordStatusFailed)
}

// GetRetryableRecords retrieves retryable processing records
func (r *ProcessingDetailRepositoryImpl) GetRetryableRecords(ctx context.Context, operationID string) ([]*bulk.ProcessingDetail, error) {
	var details []*bulk.ProcessingDetail
	if err := r.db.WithContext(ctx).
		Where("bulk_operation_id = ? AND status = ? AND retry_count < ?",
			operationID, bulk.RecordStatusFailed, 3).
		Order("record_index ASC").
		Find(&details).Error; err != nil {
		return nil, fmt.Errorf("failed to get retryable records: %w", err)
	}
	return details, nil
}

// applyFilter applies filter conditions to the query
func applyFilter(query *gorm.DB, filter *base.Filter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Apply sorting
	for _, sortField := range filter.Sort {
		order := sortField.Field
		if sortField.Direction == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Note: Group filtering would need to be handled based on the actual filter implementation
	// This is a simplified version

	return query
}
