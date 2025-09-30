package services

import (
	"context"
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
	"go.uber.org/zap"
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
		result, err := s.processSingleFarmer(ctx, bulkOp, farmer, i, options)

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
				// Set actual farmer and AAA user IDs from processing result
				farmerID := result.FarmerID
				aaaUserID := result.AAAUserID
				if farmerID == "" {
					farmerID = "unknown"
				}
				if aaaUserID == "" {
					aaaUserID = "unknown"
				}
				detail.SetSuccess(farmerID, aaaUserID)
				_ = s.processingRepo.Update(ctx, detail)
			}
		}

		// Update progress periodically
		if processed%10 == 0 || processed == len(farmers) {
			_ = s.bulkOpRepo.UpdateProgress(ctx, bulkOp.ID, processed, successful, failed, skipped)
		}
	}

	// Final progress update
	_ = s.bulkOpRepo.UpdateProgress(ctx, bulkOp.ID, processed, successful, failed, skipped)

	// Update final status
	finalStatus := bulk.StatusCompleted
	if failed > 0 && successful == 0 {
		finalStatus = bulk.StatusFailed
	}
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
		_, err := s.processSingleFarmer(ctx, bulkOp, farmer, globalIndex, options)

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

			if !options.ContinueOnError {
				progressChan <- update
				return
			}
		} else {
			update.successful = 1
		}

		progressChan <- update
	}
}

// processSingleFarmer processes a single farmer using the pipeline
func (s *BulkFarmerServiceImpl) processSingleFarmer(ctx context.Context, bulkOp *bulk.BulkOperation, farmer *requests.FarmerBulkData, index int, options requests.BulkProcessingOptions) (*ProcessingResult, error) {
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

	// Extract results from processing context
	result := &ProcessingResult{}

	// Get farmer registration result
	if farmerRegResult, exists := procCtx.GetStageResult("farmer_registration"); exists {
		if regData, ok := farmerRegResult.(map[string]interface{}); ok {
			if farmerID, ok := regData["farmer_id"].(string); ok {
				result.FarmerID = farmerID
			}
			if aaaUserID, ok := regData["aaa_user_id"].(string); ok {
				result.AAAUserID = aaaUserID
			}
		}
	}

	return result, nil
}

// ProcessingResult contains the results of processing a single farmer
type ProcessingResult struct {
	FarmerID  string
	AAAUserID string
}

// buildProcessingPipeline builds the processing pipeline based on options
func (s *BulkFarmerServiceImpl) buildProcessingPipeline(options requests.BulkProcessingOptions) pipeline.ProcessingPipeline {
	pipe := pipeline.NewPipeline(s.logger)

	// Add validation stage
	pipe.AddStage(pipeline.NewValidationStage(s.logger))

	// TODO: Add deduplication stage if not skipping duplicates
	// TODO: Add AAA user creation stage
	// TODO: Add farmer registration stage
	// TODO: Add FPO linkage stage
	// TODO: Add KisanSathi assignment stage if requested

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
func (s *BulkFarmerServiceImpl) RetryFailedRecords(ctx context.Context, req *requests.RetryBulkOperationRequest) (*responses.BulkOperationData, error) {
	// Get original operation
	originalOp, err := s.bulkOpRepo.GetByID(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original operation: %w", err)
	}

	if !originalOp.CanRetry() {
		return nil, fmt.Errorf("operation cannot be retried")
	}

	// Get failed records
	failedDetails, err := s.processingRepo.GetRetryableRecords(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retryable records: %w", err)
	}

	if len(failedDetails) == 0 {
		return nil, fmt.Errorf("no retryable records found")
	}

	// Implement retry logic - create a new bulk operation for the failed records

	// Create new bulk operation for retry
	retryOp := &bulk.BulkOperation{
		FPOOrgID:     originalOp.FPOOrgID,
		InitiatedBy:  originalOp.InitiatedBy,
		Status:       bulk.StatusPending,
		InputFormat:  originalOp.InputFormat,
		TotalRecords: len(failedDetails),
		Metadata: map[string]interface{}{
			"retry_of": originalOp.GetID(),
			"reason":   "Retry of failed records",
		},
	}

	// Save the retry operation
	if err := s.bulkOpRepo.Create(ctx, retryOp); err != nil {
		return nil, fmt.Errorf("failed to create retry operation: %w", err)
	}

	// Create processing details for retry records
	for _, detail := range failedDetails {
		// Extract original farmer data from the failed detail
		if detail.InputData != nil {
			retryDetail := &bulk.ProcessingDetail{
				BulkOperationID: retryOp.GetID(),
				RecordIndex:     detail.RecordIndex,
				Status:          bulk.RecordStatusPending,
				InputData:       detail.InputData,
				Metadata: map[string]interface{}{
					"retry_of":       detail.BulkOperationID,
					"original_index": fmt.Sprintf("%d", detail.RecordIndex),
				},
			}

			if err := s.processingRepo.Create(ctx, retryDetail); err != nil {
				s.logger.Error("Failed to create retry processing detail",
					zap.String("operation_id", retryOp.GetID()),
					zap.Int("record_index", detail.RecordIndex),
					zap.Error(err))
			}
		}
	}

	// Start processing the retry operation in background
	go func() {
		ctx := context.Background()

		// Parse chunk size from metadata
		chunkSize := 10 // default chunk size
		if chunkSizeStr, exists := originalOp.Metadata["chunk_size"]; exists {
			if chunkSizeStrStr, ok := chunkSizeStr.(string); ok {
				if parsed, err := fmt.Sscanf(chunkSizeStrStr, "%d", &chunkSize); err != nil || parsed != 1 {
					chunkSize = 10 // fallback to default
				}
			}
		}

		// Create a retry request
		retryReq := &requests.BulkFarmerAdditionRequest{
			FPOOrgID: retryOp.FPOOrgID,
			Options: requests.BulkProcessingOptions{
				ChunkSize:       chunkSize,
				ContinueOnError: true,
				ValidateOnly:    false,
			},
		}

		_, err := s.BulkAddFarmersToFPO(ctx, retryReq)
		if err != nil {
			s.logger.Error("Retry operation failed",
				zap.String("retry_operation_id", retryOp.GetID()),
				zap.Error(err))
		}
	}()

	return &responses.BulkOperationData{
		OperationID: retryOp.GetID(),
		Status:      "INITIATED",
		Message:     fmt.Sprintf("Retry initiated for %d failed records", len(failedDetails)),
	}, nil
}

// Helper methods

// downloadFileFromURL downloads a file from the given URL
func (s *BulkFarmerServiceImpl) downloadFileFromURL(ctx context.Context, url string) ([]byte, error) {
	// TODO: Implement file download from URL
	// Should handle HTTP client, timeout, file size limits, etc.
	return nil, fmt.Errorf("file download from URL not implemented")
}

// validateFarmerRecord validates a single farmer record
func (s *BulkFarmerServiceImpl) validateFarmerRecord(farmer *requests.FarmerBulkData, rowNumber int) []responses.ValidationError {
	// TODO: Implement comprehensive farmer record validation
	// Should validate required fields, formats, business rules, etc.
	return []responses.ValidationError{} // Placeholder - no validation
}

// generateCSVResult generates a CSV file with processing results
func (s *BulkFarmerServiceImpl) generateCSVResult(results []responses.ProcessingDetail, bulkOp *bulk.BulkOperation) ([]byte, error) {
	// TODO: Implement CSV result file generation
	// Should format processing results as CSV
	return nil, fmt.Errorf("CSV result generation not implemented")
}

// generateJSONResult generates a JSON file with processing results
func (s *BulkFarmerServiceImpl) generateJSONResult(results []responses.ProcessingDetail, bulkOp *bulk.BulkOperation) ([]byte, error) {
	// TODO: Implement JSON result file generation
	// Should format processing results as JSON
	return nil, fmt.Errorf("JSON result generation not implemented")
}

// generateExcelResult generates an Excel file with processing results
func (s *BulkFarmerServiceImpl) generateExcelResult(results []responses.ProcessingDetail, bulkOp *bulk.BulkOperation) ([]byte, error) {
	// TODO: Implement Excel result file generation
	// Should use a library like github.com/xuri/excelize for proper Excel generation
	return nil, fmt.Errorf("Excel result generation not implemented")
}

func (s *BulkFarmerServiceImpl) parseInputData(ctx context.Context, req *requests.BulkFarmerAdditionRequest) ([]*requests.FarmerBulkData, error) {
	// TODO: Implement input data parsing
	// Should handle both direct data and file URL downloads
	return nil, fmt.Errorf("input data parsing not implemented")
}

func (s *BulkFarmerServiceImpl) createBulkOperation(req *requests.BulkFarmerAdditionRequest, totalRecords int) *bulk.BulkOperation {
	// TODO: Implement bulk operation creation
	// Should create proper bulk operation with all required fields
	return &bulk.BulkOperation{} // Placeholder
}

func (s *BulkFarmerServiceImpl) createProcessingDetails(bulkOperationID string, farmers []*requests.FarmerBulkData) []*bulk.ProcessingDetail {
	// TODO: Implement processing details creation
	// Should create processing detail records for each farmer
	return []*bulk.ProcessingDetail{} // Placeholder
}

func (s *BulkFarmerServiceImpl) createChunks(farmers []*requests.FarmerBulkData, chunkSize int) [][]*requests.FarmerBulkData {
	// TODO: Implement farmer data chunking
	// Should split farmers into chunks for batch processing
	return [][]*requests.FarmerBulkData{} // Placeholder
}

func (s *BulkFarmerServiceImpl) getProcessingDetailByIndex(ctx context.Context, bulkOperationID string, index int) (*bulk.ProcessingDetail, error) {
	// TODO: Implement processing detail lookup by index
	// Should find specific processing detail by record index
	return nil, fmt.Errorf("processing detail lookup not implemented")
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
	s.logger.Debug("Validating bulk data",
		zap.String("format", req.InputFormat),
		zap.Int("data_size", len(req.Data)))

	// Parse the data first
	farmers, err := s.ParseBulkFile(ctx, req.InputFormat, req.Data)
	if err != nil {
		return &responses.BulkValidationData{
			IsValid: false,
			Errors: []responses.ValidationError{
				{
					RecordIndex: -1,
					Field:       "file",
					Message:     fmt.Sprintf("Failed to parse file: %v", err),
					Code:        "PARSE_ERROR",
				},
			},
			TotalRecords: 0,
			ValidRecords: 0,
		}, nil
	}

	// Validate each farmer record
	var validationErrors []responses.ValidationError
	validCount := 0

	for i, farmer := range farmers {
		errors := s.validateFarmerRecord(farmer, i+1)
		if len(errors) > 0 {
			validationErrors = append(validationErrors, errors...)
		} else {
			validCount++
		}
	}

	invalidCount := len(farmers) - validCount
	isValid := len(validationErrors) == 0

	s.logger.Debug("Validation completed",
		zap.Int("total_records", len(farmers)),
		zap.Int("valid_records", validCount),
		zap.Int("invalid_records", invalidCount),
		zap.Int("error_count", len(validationErrors)))

	return &responses.BulkValidationData{
		IsValid:      isValid,
		Errors:       validationErrors,
		TotalRecords: len(farmers),
		ValidRecords: validCount,
	}, nil
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
	s.logger.Debug("Generating result file",
		zap.String("operation_id", operationID),
		zap.String("format", format))

	// Get bulk operation details
	bulkOp, err := s.bulkOpRepo.GetByID(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find bulk operation: %w", err)
	}

	// Get all processing details for this operation
	details, err := s.processingRepo.GetByOperationID(ctx, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing details: %w", err)
	}

	// Convert processing details to result data
	var results []responses.ProcessingDetail
	for _, detail := range details {
		processedAt := detail.UpdatedAt

		// Handle pointer fields safely
		farmerID := ""
		if detail.FarmerID != nil {
			farmerID = *detail.FarmerID
		}

		aaaUserID := ""
		if detail.AAAUserID != nil {
			aaaUserID = *detail.AAAUserID
		}

		errorMsg := ""
		if detail.Error != nil {
			errorMsg = *detail.Error
		}

		result := responses.ProcessingDetail{
			RecordIndex:    detail.RecordIndex,
			Status:         string(detail.Status),
			FarmerID:       farmerID,
			AAAUserID:      aaaUserID,
			Error:          errorMsg,
			ProcessedAt:    &processedAt,
			ProcessingTime: fmt.Sprintf("%dms", detail.ProcessingTime),
			RetryCount:     detail.RetryCount,
		}
		results = append(results, result)
	}

	// Generate file based on format
	switch strings.ToLower(format) {
	case "csv":
		return s.generateCSVResult(results, bulkOp)
	case "json":
		return s.generateJSONResult(results, bulkOp)
	case "excel", "xlsx":
		return s.generateExcelResult(results, bulkOp)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
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
	s.logger.Debug("Validating farmers for FPO",
		zap.String("fpo_org_id", fpoOrgID),
		zap.Int("farmer_count", len(farmers)))

	// Validate each farmer record
	var validationErrors []responses.ValidationError
	validCount := 0

	for i, farmer := range farmers {
		errors := s.validateFarmerRecord(farmer, i+1)
		if len(errors) > 0 {
			validationErrors = append(validationErrors, errors...)
		} else {
			validCount++
		}
	}

	invalidCount := len(farmers) - validCount
	isValid := len(validationErrors) == 0

	s.logger.Debug("Farmer validation completed",
		zap.String("fpo_org_id", fpoOrgID),
		zap.Int("total_records", len(farmers)),
		zap.Int("valid_records", validCount),
		zap.Int("invalid_records", invalidCount),
		zap.Int("error_count", len(validationErrors)))

	return &responses.BulkValidationData{
		IsValid:      isValid,
		Errors:       validationErrors,
		TotalRecords: len(farmers),
		ValidRecords: validCount,
	}, nil
}
