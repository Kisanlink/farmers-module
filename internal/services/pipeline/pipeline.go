package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"go.uber.org/zap"
)

// ProcessingPipeline defines the interface for processing pipelines
type ProcessingPipeline interface {
	AddStage(stage PipelineStage) ProcessingPipeline
	Execute(ctx context.Context, data interface{}) (interface{}, error)
	GetStages() []PipelineStage
	Reset() ProcessingPipeline
}

// PipelineStage defines the interface for pipeline stages
type PipelineStage interface {
	Process(ctx context.Context, data interface{}) (interface{}, error)
	GetName() string
	CanRetry() bool
	GetTimeout() time.Duration
}

// PipelineImpl implements ProcessingPipeline
type PipelineImpl struct {
	stages []PipelineStage
	logger interfaces.Logger
}

// NewPipeline creates a new processing pipeline
func NewPipeline(logger interfaces.Logger) ProcessingPipeline {
	return &PipelineImpl{
		stages: make([]PipelineStage, 0),
		logger: logger,
	}
}

// AddStage adds a stage to the pipeline
func (p *PipelineImpl) AddStage(stage PipelineStage) ProcessingPipeline {
	p.stages = append(p.stages, stage)
	return p
}

// Execute executes all stages in the pipeline
func (p *PipelineImpl) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	var result interface{} = data

	for i, stage := range p.stages {
		p.logger.Debug("Executing pipeline stage",
			zap.Int("stage_index", i),
			zap.String("stage_name", stage.GetName()),
		)

		startTime := time.Now()

		// Create context with timeout for the stage
		stageCtx := ctx
		if stage.GetTimeout() > 0 {
			var cancel context.CancelFunc
			stageCtx, cancel = context.WithTimeout(ctx, stage.GetTimeout())
			defer cancel()
		}

		// Execute stage
		var err error
		result, err = stage.Process(stageCtx, result)

		duration := time.Since(startTime)

		if err != nil {
			p.logger.Error("Pipeline stage failed",
				zap.Int("stage_index", i),
				zap.String("stage_name", stage.GetName()),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
			return nil, fmt.Errorf("stage %d (%s) failed: %w", i, stage.GetName(), err)
		}

		p.logger.Debug("Pipeline stage completed",
			zap.Int("stage_index", i),
			zap.String("stage_name", stage.GetName()),
			zap.Duration("duration", duration),
		)
	}

	return result, nil
}

// GetStages returns all stages in the pipeline
func (p *PipelineImpl) GetStages() []PipelineStage {
	return p.stages
}

// Reset clears all stages from the pipeline
func (p *PipelineImpl) Reset() ProcessingPipeline {
	p.stages = make([]PipelineStage, 0)
	return p
}

// BasePipelineStage provides a base implementation for pipeline stages
type BasePipelineStage struct {
	name      string
	timeout   time.Duration
	retryable bool
	logger    interfaces.Logger
}

// NewBasePipelineStage creates a new base pipeline stage
func NewBasePipelineStage(name string, timeout time.Duration, retryable bool, logger interfaces.Logger) *BasePipelineStage {
	return &BasePipelineStage{
		name:      name,
		timeout:   timeout,
		retryable: retryable,
		logger:    logger,
	}
}

// GetName returns the stage name
func (b *BasePipelineStage) GetName() string {
	return b.name
}

// CanRetry returns whether the stage can be retried
func (b *BasePipelineStage) CanRetry() bool {
	return b.retryable
}

// GetTimeout returns the stage timeout
func (b *BasePipelineStage) GetTimeout() time.Duration {
	return b.timeout
}

// ProcessingContext contains context data passed between pipeline stages
type ProcessingContext struct {
	OperationID       string                 `json:"operation_id"`
	FPOOrgID          string                 `json:"fpo_org_id"`
	UserID            string                 `json:"user_id"`
	RecordIndex       int                    `json:"record_index"`
	FarmerData        interface{}            `json:"farmer_data"`
	ProcessingResult  map[string]interface{} `json:"processing_result"`
	Metadata          map[string]interface{} `json:"metadata"`
	StartTime         time.Time              `json:"start_time"`
	StageResults      map[string]interface{} `json:"stage_results"`
	DeduplicationMode string                 `json:"deduplication_mode"` // skip, update, error
}

// NewProcessingContext creates a new processing context
func NewProcessingContext(operationID, fpoOrgID, userID string, recordIndex int, farmerData interface{}) *ProcessingContext {
	return &ProcessingContext{
		OperationID:       operationID,
		FPOOrgID:          fpoOrgID,
		UserID:            userID,
		RecordIndex:       recordIndex,
		FarmerData:        farmerData,
		ProcessingResult:  make(map[string]interface{}),
		Metadata:          make(map[string]interface{}),
		StartTime:         time.Now(),
		StageResults:      make(map[string]interface{}),
		DeduplicationMode: "skip", // default
	}
}

// NewProcessingContextWithOptions creates a new processing context with deduplication mode
func NewProcessingContextWithOptions(operationID, fpoOrgID, userID string, recordIndex int, farmerData interface{}, deduplicationMode string) *ProcessingContext {
	ctx := NewProcessingContext(operationID, fpoOrgID, userID, recordIndex, farmerData)
	if deduplicationMode != "" {
		ctx.DeduplicationMode = deduplicationMode
	}
	return ctx
}

// SetStageResult sets the result for a specific stage
func (pc *ProcessingContext) SetStageResult(stageName string, result interface{}) {
	pc.StageResults[stageName] = result
}

// GetStageResult gets the result from a specific stage
func (pc *ProcessingContext) GetStageResult(stageName string) (interface{}, bool) {
	result, exists := pc.StageResults[stageName]
	return result, exists
}

// SetMetadata sets metadata value
func (pc *ProcessingContext) SetMetadata(key string, value interface{}) {
	pc.Metadata[key] = value
}

// GetMetadata gets metadata value
func (pc *ProcessingContext) GetMetadata(key string) (interface{}, bool) {
	value, exists := pc.Metadata[key]
	return value, exists
}

// SetProcessingResult sets a processing result
func (pc *ProcessingContext) SetProcessingResult(key string, value interface{}) {
	pc.ProcessingResult[key] = value
}

// GetProcessingResult gets a processing result
func (pc *ProcessingContext) GetProcessingResult(key string) (interface{}, bool) {
	value, exists := pc.ProcessingResult[key]
	return value, exists
}

// GetElapsedTime returns the elapsed time since processing started
func (pc *ProcessingContext) GetElapsedTime() time.Duration {
	return time.Since(pc.StartTime)
}

// PipelineBuilder provides a fluent interface for building pipelines
type PipelineBuilder struct {
	pipeline ProcessingPipeline
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder(logger interfaces.Logger) *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: NewPipeline(logger),
	}
}

// AddValidationStage adds a validation stage
func (pb *PipelineBuilder) AddValidationStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// AddDeduplicationStage adds a deduplication stage
func (pb *PipelineBuilder) AddDeduplicationStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// AddAAAUserCreationStage adds AAA user creation stage
func (pb *PipelineBuilder) AddAAAUserCreationStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// AddFarmerRegistrationStage adds farmer registration stage
func (pb *PipelineBuilder) AddFarmerRegistrationStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// AddFPOLinkageStage adds FPO linkage stage
func (pb *PipelineBuilder) AddFPOLinkageStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// AddKisanSathiAssignmentStage adds KisanSathi assignment stage
func (pb *PipelineBuilder) AddKisanSathiAssignmentStage() *PipelineBuilder {
	// Will be implemented in stages file
	return pb
}

// Build returns the constructed pipeline
func (pb *PipelineBuilder) Build() ProcessingPipeline {
	return pb.pipeline
}

// PipelineError represents an error that occurred during pipeline execution
type PipelineError struct {
	StageName  string        `json:"stage_name"`
	StageIndex int           `json:"stage_index"`
	ErrorCode  string        `json:"error_code"`
	Message    string        `json:"message"`
	Retryable  bool          `json:"retryable"`
	Context    interface{}   `json:"context,omitempty"`
	Duration   time.Duration `json:"duration"`
	InnerError error         `json:"-"`
}

// Error implements the error interface
func (pe *PipelineError) Error() string {
	return fmt.Sprintf("pipeline stage %s failed: %s", pe.StageName, pe.Message)
}

// Unwrap returns the inner error
func (pe *PipelineError) Unwrap() error {
	return pe.InnerError
}

// NewPipelineError creates a new pipeline error
func NewPipelineError(stageName string, stageIndex int, errorCode, message string, retryable bool, duration time.Duration, innerError error) *PipelineError {
	return &PipelineError{
		StageName:  stageName,
		StageIndex: stageIndex,
		ErrorCode:  errorCode,
		Message:    message,
		Retryable:  retryable,
		Duration:   duration,
		InnerError: innerError,
	}
}

// PipelineMetrics tracks metrics for pipeline execution
type PipelineMetrics struct {
	TotalExecutions      int64                    `json:"total_executions"`
	SuccessfulRuns       int64                    `json:"successful_runs"`
	FailedRuns           int64                    `json:"failed_runs"`
	AverageExecutionTime time.Duration            `json:"average_execution_time"`
	StageMetrics         map[string]*StageMetrics `json:"stage_metrics"`
}

// StageMetrics tracks metrics for individual stages
type StageMetrics struct {
	Executions  int64            `json:"executions"`
	Successes   int64            `json:"successes"`
	Failures    int64            `json:"failures"`
	AverageTime time.Duration    `json:"average_time"`
	MaxTime     time.Duration    `json:"max_time"`
	MinTime     time.Duration    `json:"min_time"`
	ErrorCounts map[string]int64 `json:"error_counts"`
}

// NewStageMetrics creates new stage metrics
func NewStageMetrics() *StageMetrics {
	return &StageMetrics{
		ErrorCounts: make(map[string]int64),
		MinTime:     time.Hour, // Initialize to high value
	}
}

// RecordExecution records a stage execution
func (sm *StageMetrics) RecordExecution(duration time.Duration, success bool, errorCode string) {
	sm.Executions++

	if duration > sm.MaxTime {
		sm.MaxTime = duration
	}
	if duration < sm.MinTime {
		sm.MinTime = duration
	}

	// Update average time
	if sm.Executions == 1 {
		sm.AverageTime = duration
	} else {
		sm.AverageTime = time.Duration((int64(sm.AverageTime)*(sm.Executions-1) + int64(duration)) / sm.Executions)
	}

	if success {
		sm.Successes++
	} else {
		sm.Failures++
		if errorCode != "" {
			sm.ErrorCounts[errorCode]++
		}
	}
}

// MetricsTracker tracks pipeline metrics
type MetricsTracker struct {
	metrics *PipelineMetrics
}

// NewMetricsTracker creates a new metrics tracker
func NewMetricsTracker() *MetricsTracker {
	return &MetricsTracker{
		metrics: &PipelineMetrics{
			StageMetrics: make(map[string]*StageMetrics),
		},
	}
}

// GetMetrics returns current metrics
func (mt *MetricsTracker) GetMetrics() *PipelineMetrics {
	return mt.metrics
}

// RecordPipelineExecution records a pipeline execution
func (mt *MetricsTracker) RecordPipelineExecution(duration time.Duration, success bool) {
	mt.metrics.TotalExecutions++

	if success {
		mt.metrics.SuccessfulRuns++
	} else {
		mt.metrics.FailedRuns++
	}

	// Update average execution time
	if mt.metrics.TotalExecutions == 1 {
		mt.metrics.AverageExecutionTime = duration
	} else {
		mt.metrics.AverageExecutionTime = time.Duration(
			(int64(mt.metrics.AverageExecutionTime)*(mt.metrics.TotalExecutions-1) + int64(duration)) / mt.metrics.TotalExecutions,
		)
	}
}

// RecordStageExecution records a stage execution
func (mt *MetricsTracker) RecordStageExecution(stageName string, duration time.Duration, success bool, errorCode string) {
	if _, exists := mt.metrics.StageMetrics[stageName]; !exists {
		mt.metrics.StageMetrics[stageName] = NewStageMetrics()
	}

	mt.metrics.StageMetrics[stageName].RecordExecution(duration, success, errorCode)
}
