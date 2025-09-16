package testutils

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
)

// MockBulkFarmerService provides a mock implementation for testing
type MockBulkFarmerService struct {
	BulkAddFarmersToFPOFunc    func(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error)
	GetBulkOperationStatusFunc func(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error)
	CancelBulkOperationFunc    func(ctx context.Context, operationID string) error
	RetryFailedRecordsFunc     func(ctx context.Context, req *requests.RetryBulkOperationRequest) (*responses.BulkOperationData, error)
	ValidateBulkDataFunc       func(ctx context.Context, req *requests.ValidateBulkDataRequest) (*responses.BulkValidationData, error)
	ParseBulkFileFunc          func(ctx context.Context, format string, data []byte) ([]*requests.FarmerBulkData, error)
	GenerateResultFileFunc     func(ctx context.Context, operationID string, format string) ([]byte, error)
	GetBulkUploadTemplateFunc  func(ctx context.Context, format string, includeExample bool) (*responses.BulkTemplateData, error)
}

func (m *MockBulkFarmerService) BulkAddFarmersToFPO(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error) {
	if m.BulkAddFarmersToFPOFunc != nil {
		return m.BulkAddFarmersToFPOFunc(ctx, req)
	}
	return &responses.BulkOperationData{}, nil
}

func (m *MockBulkFarmerService) GetBulkOperationStatus(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error) {
	if m.GetBulkOperationStatusFunc != nil {
		return m.GetBulkOperationStatusFunc(ctx, operationID)
	}
	return &responses.BulkOperationStatusData{}, nil
}

func (m *MockBulkFarmerService) CancelBulkOperation(ctx context.Context, operationID string) error {
	if m.CancelBulkOperationFunc != nil {
		return m.CancelBulkOperationFunc(ctx, operationID)
	}
	return nil
}

func (m *MockBulkFarmerService) RetryFailedRecords(ctx context.Context, req *requests.RetryBulkOperationRequest) (*responses.BulkOperationData, error) {
	if m.RetryFailedRecordsFunc != nil {
		return m.RetryFailedRecordsFunc(ctx, req)
	}
	return &responses.BulkOperationData{}, nil
}

func (m *MockBulkFarmerService) ValidateBulkData(ctx context.Context, req *requests.ValidateBulkDataRequest) (*responses.BulkValidationData, error) {
	if m.ValidateBulkDataFunc != nil {
		return m.ValidateBulkDataFunc(ctx, req)
	}
	return &responses.BulkValidationData{}, nil
}

func (m *MockBulkFarmerService) ParseBulkFile(ctx context.Context, format string, data []byte) ([]*requests.FarmerBulkData, error) {
	if m.ParseBulkFileFunc != nil {
		return m.ParseBulkFileFunc(ctx, format, data)
	}
	return []*requests.FarmerBulkData{}, nil
}

func (m *MockBulkFarmerService) GenerateResultFile(ctx context.Context, operationID string, format string) ([]byte, error) {
	if m.GenerateResultFileFunc != nil {
		return m.GenerateResultFileFunc(ctx, operationID, format)
	}
	return []byte{}, nil
}

func (m *MockBulkFarmerService) GetBulkUploadTemplate(ctx context.Context, format string, includeExample bool) (*responses.BulkTemplateData, error) {
	if m.GetBulkUploadTemplateFunc != nil {
		return m.GetBulkUploadTemplateFunc(ctx, format, includeExample)
	}
	return &responses.BulkTemplateData{}, nil
}

// MockBulkOperationRepository provides a mock implementation for testing
type MockBulkOperationRepository struct {
	CreateFunc              func(ctx context.Context, operation *bulk.BulkOperation) error
	GetByIDFunc             func(ctx context.Context, id string) (*bulk.BulkOperation, error)
	UpdateFunc              func(ctx context.Context, operation *bulk.BulkOperation) error
	UpdateStatusFunc        func(ctx context.Context, id string, status bulk.OperationStatus) error
	UpdateProgressFunc      func(ctx context.Context, id string, processed, successful, failed, skipped int) error
	ListFunc                func(ctx context.Context, filter *base.Filter) ([]*bulk.BulkOperation, error)
	ListByFPOFunc           func(ctx context.Context, fpoOrgID string, filter *base.Filter) ([]*bulk.BulkOperation, error)
	DeleteFunc              func(ctx context.Context, id string) error
	GetActiveOperationsFunc func(ctx context.Context, fpoOrgID string) ([]*bulk.BulkOperation, error)
}

func (m *MockBulkOperationRepository) Create(ctx context.Context, operation *bulk.BulkOperation) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, operation)
	}
	return nil
}

func (m *MockBulkOperationRepository) GetByID(ctx context.Context, id string) (*bulk.BulkOperation, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &bulk.BulkOperation{}, nil
}

func (m *MockBulkOperationRepository) Update(ctx context.Context, operation *bulk.BulkOperation) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, operation)
	}
	return nil
}

func (m *MockBulkOperationRepository) UpdateStatus(ctx context.Context, id string, status bulk.OperationStatus) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockBulkOperationRepository) UpdateProgress(ctx context.Context, id string, processed, successful, failed, skipped int) error {
	if m.UpdateProgressFunc != nil {
		return m.UpdateProgressFunc(ctx, id, processed, successful, failed, skipped)
	}
	return nil
}

func (m *MockBulkOperationRepository) List(ctx context.Context, filter *base.Filter) ([]*bulk.BulkOperation, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter)
	}
	return []*bulk.BulkOperation{}, nil
}

func (m *MockBulkOperationRepository) ListByFPO(ctx context.Context, fpoOrgID string, filter *base.Filter) ([]*bulk.BulkOperation, error) {
	if m.ListByFPOFunc != nil {
		return m.ListByFPOFunc(ctx, fpoOrgID, filter)
	}
	return []*bulk.BulkOperation{}, nil
}

func (m *MockBulkOperationRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockBulkOperationRepository) GetActiveOperations(ctx context.Context, fpoOrgID string) ([]*bulk.BulkOperation, error) {
	if m.GetActiveOperationsFunc != nil {
		return m.GetActiveOperationsFunc(ctx, fpoOrgID)
	}
	return []*bulk.BulkOperation{}, nil
}

// MockFarmerService provides a mock implementation for testing
type MockFarmerService struct {
	CreateFarmerFunc func(ctx context.Context, farmer *farmer.Farmer) (*farmer.Farmer, error)
}

func (m *MockFarmerService) CreateFarmer(ctx context.Context, farmer *farmer.Farmer) (*farmer.Farmer, error) {
	if m.CreateFarmerFunc != nil {
		return m.CreateFarmerFunc(ctx, farmer)
	}
	return farmer, nil
}

// MockLogger provides a mock logger implementation for testing
type MockLogger struct {
	DebugFunc func(msg string, fields ...interface{})
	InfoFunc  func(msg string, fields ...interface{})
	WarnFunc  func(msg string, fields ...interface{})
	ErrorFunc func(msg string, fields ...interface{})
	FatalFunc func(msg string, fields ...interface{})
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	if m.DebugFunc != nil {
		m.DebugFunc(msg, fields...)
	}
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	if m.InfoFunc != nil {
		m.InfoFunc(msg, fields...)
	}
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	if m.WarnFunc != nil {
		m.WarnFunc(msg, fields...)
	}
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(msg, fields...)
	}
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	if m.FatalFunc != nil {
		m.FatalFunc(msg, fields...)
	}
}

// SetupTestRouter creates a basic gin router for testing
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set context values
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Set("aaa_subject", "test-user-id")
		c.Set("aaa_org", "test-org-id")
		c.Next()
	})

	return router
}
