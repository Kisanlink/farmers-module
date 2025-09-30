package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockLogger implements interfaces.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...interface{}) interfaces.Logger {
	args := m.Called(fields)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) GetZapLogger() *zap.Logger {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*zap.Logger)
	}
	return zap.NewNop()
}

// MockPipelineStage implements PipelineStage for testing
type MockPipelineStage struct {
	mock.Mock
	name      string
	timeout   time.Duration
	retryable bool
}

func NewMockPipelineStage(name string, timeout time.Duration, retryable bool) *MockPipelineStage {
	return &MockPipelineStage{
		name:      name,
		timeout:   timeout,
		retryable: retryable,
	}
}

func (m *MockPipelineStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockPipelineStage) GetName() string {
	return m.name
}

func (m *MockPipelineStage) CanRetry() bool {
	return m.retryable
}

func (m *MockPipelineStage) GetTimeout() time.Duration {
	return m.timeout
}

func TestNewPipeline(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	pipeline := NewPipeline(logger)
	assert.NotNil(t, pipeline)
	assert.Len(t, pipeline.GetStages(), 0)
}

func TestPipeline_AddStage(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	pipeline := NewPipeline(logger)
	stage := NewMockPipelineStage("test_stage", 10*time.Second, true)

	pipeline.AddStage(stage)
	stages := pipeline.GetStages()
	assert.Len(t, stages, 1)
	assert.Equal(t, "test_stage", stages[0].GetName())
}

func TestPipeline_Execute_Success(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	pipeline := NewPipeline(logger)

	// Create mock stages
	stage1 := NewMockPipelineStage("stage1", 10*time.Second, true)
	stage2 := NewMockPipelineStage("stage2", 10*time.Second, true)

	// Set up expectations
	stage1.On("Process", mock.Anything, "input").Return("stage1_output", nil)
	stage2.On("Process", mock.Anything, "stage1_output").Return("final_output", nil)

	pipeline.AddStage(stage1).AddStage(stage2)

	ctx := context.Background()
	result, err := pipeline.Execute(ctx, "input")

	require.NoError(t, err)
	assert.Equal(t, "final_output", result)

	stage1.AssertExpectations(t)
	stage2.AssertExpectations(t)
}

func TestPipeline_Execute_StageFailure(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	pipeline := NewPipeline(logger)

	// Create mock stages
	stage1 := NewMockPipelineStage("stage1", 10*time.Second, true)
	stage2 := NewMockPipelineStage("stage2", 10*time.Second, true)

	// Set up expectations - stage1 fails
	stage1.On("Process", mock.Anything, "input").Return(nil, errors.New("stage1 failed"))
	// stage2 should not be called

	pipeline.AddStage(stage1).AddStage(stage2)

	ctx := context.Background()
	result, err := pipeline.Execute(ctx, "input")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "stage1 failed")

	stage1.AssertExpectations(t)
	stage2.AssertNotCalled(t, "Process")
}

func TestPipeline_Reset(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	pipeline := NewPipeline(logger)
	stage := NewMockPipelineStage("test_stage", 10*time.Second, true)

	pipeline.AddStage(stage)
	assert.Len(t, pipeline.GetStages(), 1)

	pipeline.Reset()
	assert.Len(t, pipeline.GetStages(), 0)
}

func TestProcessingContext_Creation(t *testing.T) {
	ctx := NewProcessingContext("op123", "fpo456", "user789", 5, "farmer_data")

	assert.Equal(t, "op123", ctx.OperationID)
	assert.Equal(t, "fpo456", ctx.FPOOrgID)
	assert.Equal(t, "user789", ctx.UserID)
	assert.Equal(t, 5, ctx.RecordIndex)
	assert.Equal(t, "farmer_data", ctx.FarmerData)
	assert.NotNil(t, ctx.ProcessingResult)
	assert.NotNil(t, ctx.Metadata)
	assert.NotNil(t, ctx.StageResults)
	assert.False(t, ctx.StartTime.IsZero())
}

func TestProcessingContext_StageResults(t *testing.T) {
	ctx := NewProcessingContext("op123", "fpo456", "user789", 0, "data")

	// Test setting and getting stage results
	ctx.SetStageResult("validation", map[string]interface{}{"status": "success"})

	result, exists := ctx.GetStageResult("validation")
	assert.True(t, exists)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "success", resultMap["status"])

	// Test non-existent stage
	_, exists = ctx.GetStageResult("non_existent")
	assert.False(t, exists)
}

func TestProcessingContext_Metadata(t *testing.T) {
	ctx := NewProcessingContext("op123", "fpo456", "user789", 0, "data")

	// Test setting and getting metadata
	ctx.SetMetadata("test_key", "test_value")

	value, exists := ctx.GetMetadata("test_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test non-existent metadata
	_, exists = ctx.GetMetadata("non_existent")
	assert.False(t, exists)
}

func TestProcessingContext_ProcessingResults(t *testing.T) {
	ctx := NewProcessingContext("op123", "fpo456", "user789", 0, "data")

	// Test setting and getting processing results
	ctx.SetProcessingResult("farmer_id", "farmer_123")

	value, exists := ctx.GetProcessingResult("farmer_id")
	assert.True(t, exists)
	assert.Equal(t, "farmer_123", value)

	// Test non-existent result
	_, exists = ctx.GetProcessingResult("non_existent")
	assert.False(t, exists)
}

func TestProcessingContext_ElapsedTime(t *testing.T) {
	ctx := NewProcessingContext("op123", "fpo456", "user789", 0, "data")

	// Wait a bit and check elapsed time
	time.Sleep(10 * time.Millisecond)
	elapsed := ctx.GetElapsedTime()

	assert.Greater(t, elapsed, 10*time.Millisecond)
	assert.Less(t, elapsed, 100*time.Millisecond) // Should be reasonable
}

func TestPipelineError(t *testing.T) {
	innerErr := errors.New("inner error")
	pipelineErr := NewPipelineError(
		"test_stage",
		1,
		"TEST_ERROR",
		"Test error message",
		true,
		100*time.Millisecond,
		innerErr,
	)

	assert.Equal(t, "test_stage", pipelineErr.StageName)
	assert.Equal(t, 1, pipelineErr.StageIndex)
	assert.Equal(t, "TEST_ERROR", pipelineErr.ErrorCode)
	assert.Equal(t, "Test error message", pipelineErr.Message)
	assert.True(t, pipelineErr.Retryable)
	assert.Equal(t, 100*time.Millisecond, pipelineErr.Duration)
	assert.Equal(t, innerErr, pipelineErr.Unwrap())

	assert.Contains(t, pipelineErr.Error(), "pipeline stage test_stage failed")
	assert.Contains(t, pipelineErr.Error(), "Test error message")
}

func TestStageMetrics(t *testing.T) {
	metrics := NewStageMetrics()

	// Test initial state
	assert.Equal(t, int64(0), metrics.Executions)
	assert.Equal(t, int64(0), metrics.Successes)
	assert.Equal(t, int64(0), metrics.Failures)
	assert.Equal(t, time.Duration(0), metrics.AverageTime)
	assert.Equal(t, time.Duration(0), metrics.MaxTime)
	assert.Equal(t, time.Hour, metrics.MinTime) // Initialized to high value

	// Record successful execution
	metrics.RecordExecution(100*time.Millisecond, true, "")

	assert.Equal(t, int64(1), metrics.Executions)
	assert.Equal(t, int64(1), metrics.Successes)
	assert.Equal(t, int64(0), metrics.Failures)
	assert.Equal(t, 100*time.Millisecond, metrics.AverageTime)
	assert.Equal(t, 100*time.Millisecond, metrics.MaxTime)
	assert.Equal(t, 100*time.Millisecond, metrics.MinTime)

	// Record failed execution
	metrics.RecordExecution(200*time.Millisecond, false, "VALIDATION_ERROR")

	assert.Equal(t, int64(2), metrics.Executions)
	assert.Equal(t, int64(1), metrics.Successes)
	assert.Equal(t, int64(1), metrics.Failures)
	assert.Equal(t, 150*time.Millisecond, metrics.AverageTime)
	assert.Equal(t, 200*time.Millisecond, metrics.MaxTime)
	assert.Equal(t, 100*time.Millisecond, metrics.MinTime)
	assert.Equal(t, int64(1), metrics.ErrorCounts["VALIDATION_ERROR"])
}

func TestMetricsTracker(t *testing.T) {
	tracker := NewMetricsTracker()

	// Test initial state
	metrics := tracker.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalExecutions)
	assert.Equal(t, int64(0), metrics.SuccessfulRuns)
	assert.Equal(t, int64(0), metrics.FailedRuns)

	// Record pipeline execution
	tracker.RecordPipelineExecution(500*time.Millisecond, true)

	metrics = tracker.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalExecutions)
	assert.Equal(t, int64(1), metrics.SuccessfulRuns)
	assert.Equal(t, int64(0), metrics.FailedRuns)
	assert.Equal(t, 500*time.Millisecond, metrics.AverageExecutionTime)

	// Record stage execution
	tracker.RecordStageExecution("validation", 100*time.Millisecond, true, "")

	stageMetrics, exists := metrics.StageMetrics["validation"]
	assert.True(t, exists)
	assert.Equal(t, int64(1), stageMetrics.Executions)
	assert.Equal(t, int64(1), stageMetrics.Successes)
}
