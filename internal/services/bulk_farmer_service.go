package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	bulkRepo "github.com/Kisanlink/farmers-module/internal/repo/bulk"
	"github.com/Kisanlink/farmers-module/internal/services/parsers"
	"github.com/Kisanlink/farmers-module/internal/services/pipeline"
	"github.com/Kisanlink/farmers-module/internal/utils"
)

// BulkFarmerService defines the interface for bulk farmer operations
type BulkFarmerService interface {
	// Core bulk operations
	BulkAddFarmersToFPO(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error)
	GetBulkOperationStatus(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error)
	CancelBulkOperation(ctx context.Context, operationID string) error
	RetryFailedRecords(ctx context.Context, req *requests.RetryBulkOperationRequest) (*responses.BulkOperationData, error)

	// File operations
	ValidateBulkData(ctx context.Context, req *requests.ValidateBulkDataRequest) (*responses.BulkValidationData, error)
	ParseBulkFile(ctx context.Context, format string, data []byte) ([]*requests.FarmerBulkData, error)
	GenerateResultFile(ctx context.Context, operationID string, format string) ([]byte, error)

	// Template operations
	GetBulkUploadTemplate(ctx context.Context, format string, includeExample bool) (*responses.BulkTemplateData, error)
}

// BulkFarmerServiceImpl implements BulkFarmerService
type BulkFarmerServiceImpl struct {
	bulkOpRepo         bulkRepo.BulkOperationRepository
	processingRepo     bulkRepo.ProcessingDetailRepository
	farmerService      FarmerService
	linkageService     FarmerLinkageService
	aaaService         AAAService
	fileParser         parsers.FileParser
	processingPipeline pipeline.ProcessingPipeline
	logger             interfaces.Logger
	config             *BulkServiceConfig
}

// BulkServiceConfig contains configuration for bulk service
type BulkServiceConfig struct {
	MaxSyncRecords    int
	DefaultChunkSize  int
	MaxConcurrency    int
	MaxRetries        int
	ProcessingTimeout time.Duration
	EnableAsync       bool
}

// NewBulkFarmerService creates a new bulk farmer service
func NewBulkFarmerService(
	bulkOpRepo bulkRepo.BulkOperationRepository,
	processingRepo bulkRepo.ProcessingDetailRepository,
	farmerService FarmerService,
	linkageService FarmerLinkageService,
	aaaService AAAService,
	logger interfaces.Logger,
) BulkFarmerService {
	config := &BulkServiceConfig{
		MaxSyncRecords:    100,
		DefaultChunkSize:  100,
		MaxConcurrency:    10,
		MaxRetries:        3,
		ProcessingTimeout: 30 * time.Minute,
		EnableAsync:       true,
	}

	// Create file parser
	fileParser := parsers.NewFileParser()

	// Create processing pipeline
	processingPipeline := pipeline.NewPipeline(logger)

	return &BulkFarmerServiceImpl{
		bulkOpRepo:         bulkOpRepo,
		processingRepo:     processingRepo,
		farmerService:      farmerService,
		linkageService:     linkageService,
		aaaService:         aaaService,
		fileParser:         fileParser,
		processingPipeline: processingPipeline,
		logger:             logger,
		config:             config,
	}
}

// BulkAddFarmersToFPO adds multiple farmers to an FPO
func (s *BulkFarmerServiceImpl) BulkAddFarmersToFPO(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error) {
	s.logger.Info(fmt.Sprintf("Starting bulk farmer addition: fpo_org_id=%s, processing_mode=%s, input_format=%s",
		req.FPOOrgID, req.ProcessingMode, req.InputFormat))

	// Set default options
	req.Options.SetDefaults()

	// Parse input data
	farmers, err := s.parseInputData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	if len(farmers) == 0 {
		return nil, fmt.Errorf("no valid farmer records found in input")
	}

	// Validate data if requested
	if req.Options.ValidateOnly {
		validationResult, err := s.validateFarmers(ctx, req.FPOOrgID, farmers)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		return &responses.BulkOperationData{
			Status:  "VALIDATED",
			Message: fmt.Sprintf("Validation completed. Valid: %d, Invalid: %d", validationResult.ValidRecords, len(validationResult.Errors)),
		}, nil
	}

	// Create bulk operation record
	bulkOp := s.createBulkOperation(req, len(farmers))
	if err := s.bulkOpRepo.Create(ctx, bulkOp); err != nil {
		return nil, fmt.Errorf("failed to create bulk operation: %w", err)
	}

	// Create processing details
	processingDetails := s.createProcessingDetails(bulkOp.ID, farmers)
	if err := s.processingRepo.CreateBatch(ctx, processingDetails); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create processing details: %v", err))
	}

	// Determine processing strategy
	if req.ProcessingMode == "sync" || len(farmers) <= s.config.MaxSyncRecords {
		// Process synchronously
		go s.processSynchronously(context.Background(), bulkOp, farmers, req.Options)

		// For sync mode with small batches, wait a bit to get initial progress
		if req.ProcessingMode == "sync" && len(farmers) <= 10 {
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		// Process asynchronously
		go s.processAsynchronously(context.Background(), bulkOp, farmers, req.Options)
	}

	// Return operation info
	return &responses.BulkOperationData{
		OperationID: bulkOp.ID,
		Status:      string(bulkOp.Status),
		StatusURL:   fmt.Sprintf("/api/v1/bulk/status/%s", bulkOp.ID),
		ResultURL:   fmt.Sprintf("/api/v1/bulk/results/%s", bulkOp.ID),
		Message:     fmt.Sprintf("Bulk operation initiated for %d farmers", len(farmers)),
	}, nil
}

// processSynchronously processes farmers synchronously
func (s *BulkFarmerServiceImpl) processSynchronously(ctx context.Context, bulkOp *bulk.BulkOperation, farmers []*requests.FarmerBulkData, options requests.BulkProcessingOptions) {
	s.logger.Info(fmt.Sprintf("Starting synchronous processing: operation_id=%s, total_farmers=%d",
		bulkOp.ID, len(farmers)))

	// Update status to processing
	_ = s.bulkOpRepo.UpdateStatus(ctx, bulkOp.ID, bulk.StatusProcessing)

	var processed, successful, failed, skipped int

	for i, farmer := range farmers {
		// Process individual farmer
		procCtx, err := s.processSingleFarmer(ctx, bulkOp, farmer, i, options)

		processed++
		if err != nil {
			failed++
			s.logger.Error(fmt.Sprintf("Failed to process farmer: index=%d, phone=%s, error=%v",
				i, farmer.PhoneNumber, err))

			// Update processing detail with error
			detail, _ := s.getProcessingDetailByIndex(ctx, bulkOp.ID, i)
			if detail != nil {
				detail.SetFailed(err.Error(), "PROCESSING_ERROR")
				_ = s.processingRepo.Update(ctx, detail)
			}

			if !options.ContinueOnError {
				break
			}
		} else {
			successful++

			// Update processing detail with success
			detail, _ := s.getProcessingDetailByIndex(ctx, bulkOp.ID, i)
			if detail != nil {
				// Extract farmer ID and AAA user ID from processing context
				farmerID, aaaUserID := s.extractIDsFromContext(procCtx)
				detail.SetSuccess(farmerID, aaaUserID)
				_ = s.processingRepo.Update(ctx, detail)
			}
		}

		// Update progress periodically (but not on last iteration to avoid race)
		if processed%10 == 0 && processed != len(farmers) {
			_ = s.bulkOpRepo.UpdateProgress(ctx, bulkOp.ID, processed, successful, failed, skipped)
		}
	}

	// Determine final status FIRST before any DB updates
	finalStatus := bulk.StatusCompleted
	if failed > 0 && successful == 0 {
		finalStatus = bulk.StatusFailed
	}

	// Update status first (this will set end_time and processing_time correctly)
	_ = s.bulkOpRepo.UpdateStatus(ctx, bulkOp.ID, finalStatus)

	s.logger.Info("Synchronous processing completed",
		bulkOp.ID,
		successful,
		failed,
	)
}

// processAsynchronously processes farmers asynchronously using worker pool
func (s *BulkFarmerServiceImpl) processAsynchronously(ctx context.Context, bulkOp *bulk.BulkOperation, farmers []*requests.FarmerBulkData, options requests.BulkProcessingOptions) {
	s.logger.Info("Starting asynchronous processing",
		bulkOp.ID,
		len(farmers),
		options.ChunkSize,
		options.MaxConcurrency,
	)

	// Update status to processing
	_ = s.bulkOpRepo.UpdateStatus(ctx, bulkOp.ID, bulk.StatusProcessing)

	// Create chunks
	chunks := s.createChunks(farmers, options.ChunkSize)

	// Process chunks with worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, options.MaxConcurrency)

	progressChan := make(chan progressUpdate, len(farmers))
	go s.aggregateProgress(ctx, bulkOp.ID, progressChan, len(farmers))

	for chunkIndex, chunk := range chunks {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(chunkIdx int, chunkData []*requests.FarmerBulkData) {
			defer func() {
				<-semaphore // Release semaphore
				wg.Done()
			}()

			s.processChunk(ctx, bulkOp, chunkIdx, chunkData, options, progressChan)
		}(chunkIndex, chunk)
	}

	wg.Wait()
	close(progressChan)

	// Update final status
	_ = s.bulkOpRepo.UpdateStatus(ctx, bulkOp.ID, bulk.StatusCompleted)

	s.logger.Info("Asynchronous processing completed",
		bulkOp.ID,
	)
}

// processChunk processes a chunk of farmers
func (s *BulkFarmerServiceImpl) processChunk(ctx context.Context, bulkOp *bulk.BulkOperation, chunkIndex int, farmers []*requests.FarmerBulkData, options requests.BulkProcessingOptions, progressChan chan<- progressUpdate) {
	s.logger.Debug("Processing chunk",
		bulkOp.ID,
		chunkIndex,
		len(farmers),
	)

	for i, farmer := range farmers {
		globalIndex := chunkIndex*options.ChunkSize + i
		procCtx, err := s.processSingleFarmer(ctx, bulkOp, farmer, globalIndex, options)

		update := progressUpdate{
			processed: 1,
		}

		if err != nil {
			update.failed = 1
			s.logger.Error("Failed to process farmer in chunk",
				chunkIndex,
				i,
				err,
			)

			// Update processing detail with error
			detail, _ := s.getProcessingDetailByIndex(ctx, bulkOp.ID, globalIndex)
			if detail != nil {
				detail.SetFailed(err.Error(), "PROCESSING_ERROR")
				_ = s.processingRepo.Update(ctx, detail)
			}

			if !options.ContinueOnError {
				progressChan <- update
				return
			}
		} else {
			update.successful = 1

			// Update processing detail with success
			detail, _ := s.getProcessingDetailByIndex(ctx, bulkOp.ID, globalIndex)
			if detail != nil {
				// Extract farmer ID and AAA user ID from processing context
				farmerID, aaaUserID := s.extractIDsFromContext(procCtx)
				detail.SetSuccess(farmerID, aaaUserID)
				_ = s.processingRepo.Update(ctx, detail)
			}
		}

		progressChan <- update
	}
}

// extractIDsFromContext extracts farmer ID and AAA user ID from processing context
func (s *BulkFarmerServiceImpl) extractIDsFromContext(procCtx *pipeline.ProcessingContext) (string, string) {
	farmerID := ""
	aaaUserID := ""

	// Extract from farmer registration stage result
	if regResult, ok := procCtx.StageResults["farmer_registration"].(map[string]interface{}); ok {
		if fid, exists := regResult["farmer_id"]; exists {
			if fidStr, ok := fid.(string); ok {
				farmerID = fidStr
			}
		}
		if auid, exists := regResult["aaa_user_id"]; exists {
			if auidStr, ok := auid.(string); ok {
				aaaUserID = auidStr
			}
		}
	}

	// Fallback to AAA user creation stage if not found
	if aaaUserID == "" {
		if aaaResult, ok := procCtx.StageResults["aaa_user_creation"].(map[string]interface{}); ok {
			if auid, exists := aaaResult["aaa_user_id"]; exists {
				if auidStr, ok := auid.(string); ok {
					aaaUserID = auidStr
				}
			}
		}
	}

	return farmerID, aaaUserID
}

// processSingleFarmer processes a single farmer using the pipeline
func (s *BulkFarmerServiceImpl) processSingleFarmer(ctx context.Context, bulkOp *bulk.BulkOperation, farmer *requests.FarmerBulkData, index int, options requests.BulkProcessingOptions) (*pipeline.ProcessingContext, error) {
	// Create processing context
	procCtx := pipeline.NewProcessingContext(
		bulkOp.ID,
		bulkOp.FPOOrgID,
		bulkOp.InitiatedBy,
		index,
		farmer,
	)

	// Build processing pipeline
	pipe := s.buildProcessingPipeline(options)

	// Execute pipeline
	_, err := pipe.Execute(ctx, procCtx)
	if err != nil {
		return nil, fmt.Errorf("pipeline execution failed: %w", err)
	}

	return procCtx, nil
}

// buildProcessingPipeline builds the processing pipeline based on options
func (s *BulkFarmerServiceImpl) buildProcessingPipeline(options requests.BulkProcessingOptions) pipeline.ProcessingPipeline {
	pipe := pipeline.NewPipeline(s.logger)

	// Add validation stage
	pipe.AddStage(pipeline.NewValidationStage(s.logger))

	// Add deduplication stage if not skipping duplicates
	if options.DeduplicationMode != "skip" {
		pipe.AddStage(pipeline.NewDeduplicationStage(s.farmerService, s.logger))
	}

	// Add AAA user creation stage
	pipe.AddStage(pipeline.NewAAAUserCreationStage(s.aaaService, s.logger))

	// Add farmer registration stage
	pipe.AddStage(pipeline.NewFarmerRegistrationStage(s.farmerService, s.logger))

	// Add FPO linkage stage
	pipe.AddStage(pipeline.NewFPOLinkageStage(s.linkageService, s.logger))

	// Add KisanSathi assignment stage if requested
	if options.AssignKisanSathi {
		pipe.AddStage(pipeline.NewKisanSathiAssignmentStage(s.linkageService, options.KisanSathiUserID, s.logger))
	}

	return pipe
}

// GetBulkOperationStatus retrieves the status of a bulk operation
func (s *BulkFarmerServiceImpl) GetBulkOperationStatus(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error) {
	bulkOp, err := s.bulkOpRepo.GetByID(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk operation: %w", err)
	}

	// Calculate progress percentage
	progressPercentage := bulkOp.GetProgressPercentage()

	// Estimate completion time
	var estimatedCompletion *time.Time
	if bulkOp.Status == bulk.StatusProcessing && bulkOp.ProcessedRecords > 0 {
		elapsed := time.Since(bulkOp.CreatedAt)
		rate := float64(bulkOp.ProcessedRecords) / elapsed.Seconds()
		remaining := bulkOp.TotalRecords - bulkOp.ProcessedRecords
		if rate > 0 {
			eta := time.Now().Add(time.Duration(float64(remaining)/rate) * time.Second)
			estimatedCompletion = &eta
		}
	}

	// Build status response
	status := &responses.BulkOperationStatusData{
		OperationID: bulkOp.ID,
		FPOOrgID:    bulkOp.FPOOrgID,
		Status:      string(bulkOp.Status),
		Progress: responses.ProgressInfo{
			Total:      bulkOp.TotalRecords,
			Processed:  bulkOp.ProcessedRecords,
			Successful: bulkOp.SuccessfulRecords,
			Failed:     bulkOp.FailedRecords,
			Skipped:    bulkOp.SkippedRecords,
			Percentage: progressPercentage,
		},
		StartTime:           bulkOp.StartTime,
		EndTime:             bulkOp.EndTime,
		EstimatedCompletion: estimatedCompletion,
		ErrorSummary:        bulkOp.ErrorSummary,
		ResultFileURL:       bulkOp.ResultFileURL,
		CanRetry:            bulkOp.CanRetry(),
		Metadata:            bulkOp.Metadata,
	}

	if bulkOp.ProcessingTime > 0 {
		status.ProcessingTime = fmt.Sprintf("%dms", bulkOp.ProcessingTime)
	}

	return status, nil
}

// CancelBulkOperation cancels a bulk operation
func (s *BulkFarmerServiceImpl) CancelBulkOperation(ctx context.Context, operationID string) error {
	bulkOp, err := s.bulkOpRepo.GetByID(ctx, operationID)
	if err != nil {
		return fmt.Errorf("failed to get bulk operation: %w", err)
	}

	if bulkOp.IsComplete() {
		return fmt.Errorf("operation is already complete")
	}

	return s.bulkOpRepo.UpdateStatus(ctx, operationID, bulk.StatusCancelled)
}

// RetryFailedRecords retries failed records from a bulk operation
// Uses exponential backoff retry logic with circuit breaker pattern
func (s *BulkFarmerServiceImpl) RetryFailedRecords(ctx context.Context, req *requests.RetryBulkOperationRequest) (*responses.BulkOperationData, error) {
	s.logger.Info(fmt.Sprintf("Retrying failed records: operation_id=%s", req.OperationID))

	// Get original operation
	originalOp, err := s.bulkOpRepo.GetByID(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original operation: %w", err)
	}

	if !originalOp.CanRetry() {
		return nil, fmt.Errorf("operation cannot be retried: status=%s", originalOp.Status)
	}

	// Get failed records
	failedDetails, err := s.processingRepo.GetRetryableRecords(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retryable records: %w", err)
	}

	if len(failedDetails) == 0 {
		return nil, fmt.Errorf("no retryable records found")
	}

	// Create new bulk operation for retry
	retryOp := bulk.NewBulkOperation()
	retryOp.FPOOrgID = originalOp.FPOOrgID
	retryOp.InitiatedBy = originalOp.InitiatedBy
	retryOp.TotalRecords = len(failedDetails)
	retryOp.Status = bulk.StatusPending
	retryOp.InputFormat = originalOp.InputFormat
	retryOp.ProcessingMode = originalOp.ProcessingMode
	retryOp.Metadata = map[string]interface{}{
		"retry_of":      originalOp.ID,
		"retry_attempt": "1",
	}
	now := time.Now()
	retryOp.StartTime = &now

	if err := s.bulkOpRepo.Create(ctx, retryOp); err != nil {
		return nil, fmt.Errorf("failed to create retry operation: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Created retry operation: retry_id=%s, original_id=%s, retrying=%d records",
		retryOp.ID, originalOp.ID, len(failedDetails)))

	// Reconstruct farmer data from failed details
	farmers := make([]*requests.FarmerBulkData, 0, len(failedDetails))
	for _, detail := range failedDetails {
		var farmerData requests.FarmerBulkData
		// Convert InputData map to FarmerBulkData
		inputJSON, err := json.Marshal(detail.InputData)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to marshal input data: detail_id=%s, error=%v", detail.ID, err))
			continue
		}
		if err := json.Unmarshal(inputJSON, &farmerData); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to unmarshal farmer data: detail_id=%s, error=%v", detail.ID, err))
			continue
		}
		farmers = append(farmers, &farmerData)
	}

	// Prepare retry configuration with exponential backoff
	retryConfig := utils.RetryConfig{
		MaxAttempts:   s.config.MaxRetries,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}

	// Process options
	options := requests.BulkProcessingOptions{
		ChunkSize:       s.config.DefaultChunkSize,
		MaxConcurrency:  s.config.MaxConcurrency,
		ContinueOnError: true,
	}

	// Process retries asynchronously with exponential backoff
	if s.config.EnableAsync && len(farmers) > s.config.MaxSyncRecords {
		go s.processRetriesAsynchronously(context.Background(), retryOp, farmers, options, retryConfig)
	} else {
		go s.processRetriesSynchronously(context.Background(), retryOp, farmers, options, retryConfig)
	}

	// Return operation info
	return &responses.BulkOperationData{
		OperationID: retryOp.ID,
		Status:      string(retryOp.Status),
		StatusURL:   fmt.Sprintf("/api/v1/bulk/status/%s", retryOp.ID),
		ResultURL:   fmt.Sprintf("/api/v1/bulk/results/%s", retryOp.ID),
		Message:     fmt.Sprintf("Retry operation initiated for %d failed records from operation %s", len(farmers), originalOp.ID),
	}, nil
}

// processRetriesSynchronously processes retry records with exponential backoff
func (s *BulkFarmerServiceImpl) processRetriesSynchronously(ctx context.Context, retryOp *bulk.BulkOperation, farmers []*requests.FarmerBulkData, options requests.BulkProcessingOptions, retryConfig utils.RetryConfig) {
	s.logger.Info(fmt.Sprintf("Starting synchronous retry processing: retry_id=%s, total=%d", retryOp.ID, len(farmers)))

	_ = s.bulkOpRepo.UpdateStatus(ctx, retryOp.ID, bulk.StatusProcessing)

	var processed, successful, failed int

	for i, farmer := range farmers {
		// Process with retry logic and exponential backoff
		var procCtx *pipeline.ProcessingContext
		err := utils.RetryWithBackoff(ctx, retryConfig, func() error {
			var err error
			procCtx, err = s.processSingleFarmer(ctx, retryOp, farmer, i, options)
			return err
		})

		processed++
		if err != nil {
			failed++
			s.logger.Error(fmt.Sprintf("Retry failed after %d attempts: index=%d, phone=%s, error=%v",
				retryConfig.MaxAttempts, i, farmer.PhoneNumber, err))

			// Update processing detail with error
			detail, _ := s.getProcessingDetailByIndex(ctx, retryOp.ID, i)
			if detail != nil {
				detail.SetFailed(err.Error(), "RETRY_FAILED")
				_ = s.processingRepo.Update(ctx, detail)
			}
		} else {
			successful++
			s.logger.Info(fmt.Sprintf("Retry succeeded: index=%d, phone=%s", i, farmer.PhoneNumber))

			// Update processing detail with success
			detail, _ := s.getProcessingDetailByIndex(ctx, retryOp.ID, i)
			if detail != nil {
				// Extract farmer ID and AAA user ID from processing context
				farmerID, aaaUserID := s.extractIDsFromContext(procCtx)
				detail.SetSuccess(farmerID, aaaUserID)
				_ = s.processingRepo.Update(ctx, detail)
			}
		}

		// Update progress
		if processed%10 == 0 || processed == len(farmers) {
			_ = s.bulkOpRepo.UpdateProgress(ctx, retryOp.ID, processed, successful, failed, 0)
		}
	}

	// Final status
	finalStatus := bulk.StatusCompleted
	if failed > 0 && successful == 0 {
		finalStatus = bulk.StatusFailed
	}
	_ = s.bulkOpRepo.UpdateStatus(ctx, retryOp.ID, finalStatus)
	_ = s.bulkOpRepo.UpdateProgress(ctx, retryOp.ID, processed, successful, failed, 0)

	s.logger.Info(fmt.Sprintf("Retry processing completed: retry_id=%s, successful=%d, failed=%d",
		retryOp.ID, successful, failed))
}

// processRetriesAsynchronously processes retry records asynchronously with exponential backoff
func (s *BulkFarmerServiceImpl) processRetriesAsynchronously(ctx context.Context, retryOp *bulk.BulkOperation, farmers []*requests.FarmerBulkData, options requests.BulkProcessingOptions, retryConfig utils.RetryConfig) {
	s.logger.Info(fmt.Sprintf("Starting asynchronous retry processing: retry_id=%s, total=%d", retryOp.ID, len(farmers)))

	_ = s.bulkOpRepo.UpdateStatus(ctx, retryOp.ID, bulk.StatusProcessing)

	// Process with worker pool and retry logic
	chunks := s.createChunks(farmers, options.ChunkSize)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, options.MaxConcurrency)

	progressChan := make(chan progressUpdate, len(farmers))
	go s.aggregateProgress(ctx, retryOp.ID, progressChan, len(farmers))

	for chunkIdx, chunk := range chunks {
		wg.Add(1)
		go func(idx int, farmerChunk []*requests.FarmerBulkData) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			for i, farmer := range farmerChunk {
				globalIdx := idx*options.ChunkSize + i

				// Process with retry logic and exponential backoff
				var procCtx *pipeline.ProcessingContext
				err := utils.RetryWithBackoff(ctx, retryConfig, func() error {
					var err error
					procCtx, err = s.processSingleFarmer(ctx, retryOp, farmer, globalIdx, options)
					return err
				})

				if err != nil {
					progressChan <- progressUpdate{failed: 1}
					s.logger.Error(fmt.Sprintf("Retry failed: index=%d, error=%v", globalIdx, err))

					// Update processing detail with error
					detail, _ := s.getProcessingDetailByIndex(ctx, retryOp.ID, globalIdx)
					if detail != nil {
						detail.SetFailed(err.Error(), "RETRY_FAILED")
						_ = s.processingRepo.Update(ctx, detail)
					}
				} else {
					progressChan <- progressUpdate{successful: 1}

					// Update processing detail with success
					detail, _ := s.getProcessingDetailByIndex(ctx, retryOp.ID, globalIdx)
					if detail != nil {
						// Extract farmer ID and AAA user ID from processing context
						farmerID, aaaUserID := s.extractIDsFromContext(procCtx)
						detail.SetSuccess(farmerID, aaaUserID)
						_ = s.processingRepo.Update(ctx, detail)
					}
				}
			}
		}(chunkIdx, chunk)
	}

	wg.Wait()
	close(progressChan)

	// Wait for progress aggregation to complete
	time.Sleep(1 * time.Second)

	s.logger.Info(fmt.Sprintf("Asynchronous retry processing completed: retry_id=%s", retryOp.ID))
}

// Helper methods

func (s *BulkFarmerServiceImpl) parseInputData(ctx context.Context, req *requests.BulkFarmerAdditionRequest) ([]*requests.FarmerBulkData, error) {
	if len(req.Data) > 0 {
		return s.ParseBulkFile(ctx, req.InputFormat, req.Data)
	}

	if req.FileURL != "" {
		// TODO: Download file from URL and parse
		return nil, fmt.Errorf("file URL processing not yet implemented")
	}

	return nil, fmt.Errorf("no input data provided")
}

func (s *BulkFarmerServiceImpl) createBulkOperation(req *requests.BulkFarmerAdditionRequest, totalRecords int) *bulk.BulkOperation {
	bulkOp := bulk.NewBulkOperation()
	bulkOp.FPOOrgID = req.FPOOrgID
	bulkOp.InitiatedBy = req.UserID
	bulkOp.InputFormat = bulk.InputFormat(req.InputFormat)
	bulkOp.ProcessingMode = bulk.ProcessingMode(req.ProcessingMode)
	bulkOp.TotalRecords = totalRecords
	bulkOp.Status = bulk.StatusPending

	// Store options in metadata
	optionsJSON, _ := json.Marshal(req.Options)
	var optionsMap map[string]interface{}
	_ = json.Unmarshal(optionsJSON, &optionsMap)
	bulkOp.Options = optionsMap

	return bulkOp
}

func (s *BulkFarmerServiceImpl) createProcessingDetails(bulkOperationID string, farmers []*requests.FarmerBulkData) []*bulk.ProcessingDetail {
	details := make([]*bulk.ProcessingDetail, len(farmers))
	for i, farmer := range farmers {
		detail := bulk.NewProcessingDetail(bulkOperationID, i)
		detail.ExternalID = farmer.ExternalID

		// Store input data
		farmerJSON, _ := json.Marshal(farmer)
		var farmerMap map[string]interface{}
		_ = json.Unmarshal(farmerJSON, &farmerMap)
		detail.InputData = farmerMap

		details[i] = detail
	}
	return details
}

func (s *BulkFarmerServiceImpl) createChunks(farmers []*requests.FarmerBulkData, chunkSize int) [][]*requests.FarmerBulkData {
	var chunks [][]*requests.FarmerBulkData
	for i := 0; i < len(farmers); i += chunkSize {
		end := i + chunkSize
		if end > len(farmers) {
			end = len(farmers)
		}
		chunks = append(chunks, farmers[i:end])
	}
	return chunks
}

func (s *BulkFarmerServiceImpl) getProcessingDetailByIndex(ctx context.Context, bulkOperationID string, index int) (*bulk.ProcessingDetail, error) {
	details, err := s.processingRepo.GetByOperationID(ctx, bulkOperationID)
	if err != nil {
		return nil, err
	}

	for _, detail := range details {
		if detail.RecordIndex == index {
			return detail, nil
		}
	}

	return nil, fmt.Errorf("processing detail not found for index %d", index)
}

type progressUpdate struct {
	processed  int
	successful int
	failed     int
	skipped    int
}

func (s *BulkFarmerServiceImpl) aggregateProgress(ctx context.Context, operationID string, progressChan <-chan progressUpdate, total int) {
	var processed, successful, failed, skipped int

	updateTicker := time.NewTicker(1 * time.Second)
	defer updateTicker.Stop()

	for {
		select {
		case update, ok := <-progressChan:
			if !ok {
				// Channel closed, final update
				_ = s.bulkOpRepo.UpdateProgress(ctx, operationID, processed, successful, failed, skipped)
				return
			}

			processed += update.processed
			successful += update.successful
			failed += update.failed
			skipped += update.skipped

		case <-updateTicker.C:
			// Periodic update
			_ = s.bulkOpRepo.UpdateProgress(ctx, operationID, processed, successful, failed, skipped)

		case <-ctx.Done():
			return
		}
	}
}

// Placeholder methods for unimplemented features

func (s *BulkFarmerServiceImpl) ValidateBulkData(ctx context.Context, req *requests.ValidateBulkDataRequest) (*responses.BulkValidationData, error) {
	// TODO: Implement validation logic
	return nil, fmt.Errorf("not yet implemented")
}

func (s *BulkFarmerServiceImpl) ParseBulkFile(ctx context.Context, format string, data []byte) ([]*requests.FarmerBulkData, error) {
	s.logger.Debug("Parsing bulk file",
		format,
		len(data),
	)

	var farmers []*requests.FarmerBulkData
	var err error

	switch strings.ToLower(format) {
	case "csv":
		farmers, err = s.fileParser.ParseCSV(data)
	case "excel", "xlsx", "xls":
		farmers, err = s.fileParser.ParseExcel(data)
	case "json":
		farmers, err = s.fileParser.ParseJSON(data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse %s file: %w", format, err)
	}

	s.logger.Debug("File parsed successfully",
		format,
		len(farmers),
	)

	return farmers, nil
}

func (s *BulkFarmerServiceImpl) GenerateResultFile(ctx context.Context, operationID string, format string) ([]byte, error) {
	// TODO: Implement result file generation
	return nil, fmt.Errorf("not yet implemented")
}

func (s *BulkFarmerServiceImpl) GetBulkUploadTemplate(ctx context.Context, format string, includeExample bool) (*responses.BulkTemplateData, error) {
	s.logger.Debug("Generating bulk upload template",
		format,
		includeExample,
	)

	var content []byte
	var err error
	var fileName string

	switch strings.ToLower(format) {
	case "csv":
		content, err = s.fileParser.GenerateCSVTemplate(includeExample)
		fileName = "farmer_upload_template.csv"
	case "excel", "xlsx":
		content, err = s.fileParser.GenerateExcelTemplate(includeExample)
		fileName = "farmer_upload_template.xlsx"
	default:
		return nil, fmt.Errorf("unsupported template format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate %s template: %w", format, err)
	}

	// Define field information
	fields := []responses.FieldInfo{
		{Name: "first_name", DisplayName: "First Name", Type: "string", Required: true, Example: "John", Description: "Farmer's first name"},
		{Name: "last_name", DisplayName: "Last Name", Type: "string", Required: true, Example: "Doe", Description: "Farmer's last name"},
		{Name: "phone_number", DisplayName: "Phone Number", Type: "string", Required: true, Format: "10 digits", Example: "9876543210", Description: "10-digit mobile number"},
		{Name: "email", DisplayName: "Email", Type: "string", Required: false, Format: "email", Example: "john.doe@example.com", Description: "Email address"},
		{Name: "date_of_birth", DisplayName: "Date of Birth", Type: "date", Required: false, Format: "YYYY-MM-DD", Example: "1990-01-15", Description: "Date of birth"},
		{Name: "gender", DisplayName: "Gender", Type: "string", Required: false, Example: "male", Description: "Gender (male, female, other)"},
		{Name: "street_address", DisplayName: "Street Address", Type: "string", Required: false, Example: "123 Farm Street", Description: "Street address"},
		{Name: "city", DisplayName: "City", Type: "string", Required: false, Example: "Mumbai", Description: "City name"},
		{Name: "state", DisplayName: "State", Type: "string", Required: false, Example: "Maharashtra", Description: "State name"},
		{Name: "postal_code", DisplayName: "Postal Code", Type: "string", Required: false, Example: "400001", Description: "Postal/ZIP code"},
		{Name: "land_ownership_type", DisplayName: "Land Ownership", Type: "string", Required: false, Example: "owned", Description: "Type of land ownership"},
		{Name: "external_id", DisplayName: "External ID", Type: "string", Required: false, Example: "FARMER001", Description: "External identifier for tracking"},
	}

	instructions := fmt.Sprintf(`
Bulk Farmer Upload Template - %s Format

Instructions:
1. Fill in the farmer data in the rows below the header
2. Required fields: first_name, last_name, phone_number
3. Phone numbers should be 10-digit Indian mobile numbers
4. Date format: YYYY-MM-DD (e.g., 1990-01-15)
5. Gender options: male, female, other
6. Do not modify the header row
7. Maximum %d farmers per upload

Tips:
- Remove any test/example data before uploading
- Ensure phone numbers are unique
- Use consistent data formats
- Keep external_id unique for tracking purposes
`, strings.ToUpper(format), s.config.MaxSyncRecords*10) // Allow more for async processing

	template := &responses.BulkTemplateData{
		Format:       format,
		FileName:     fileName,
		Content:      content,
		Fields:       fields,
		Instructions: instructions,
	}

	s.logger.Debug("Template generated successfully",
		format,
		fileName,
		len(content),
	)

	return template, nil
}

func (s *BulkFarmerServiceImpl) validateFarmers(ctx context.Context, fpoOrgID string, farmers []*requests.FarmerBulkData) (*responses.BulkValidationData, error) {
	// TODO: Implement validation
	return &responses.BulkValidationData{
		IsValid:      true,
		TotalRecords: len(farmers),
		ValidRecords: len(farmers),
	}, nil
}
