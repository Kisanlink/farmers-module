package testutils

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/bulk"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

// MockAAAService provides a mock AAA service implementation for testing
type MockAAAService struct {
	CheckPermissionFunc         func(ctx context.Context, subject, resource, action, object, orgID string) (bool, error)
	SeedRolesAndPermissionsFunc func(ctx context.Context) error
	CreateUserFunc              func(ctx context.Context, req interface{}) (interface{}, error)
	GetUserFunc                 func(ctx context.Context, userID string) (interface{}, error)
	GetUserByMobileFunc         func(ctx context.Context, mobileNumber string) (interface{}, error)
	GetUserByEmailFunc          func(ctx context.Context, email string) (interface{}, error)
	CreateOrganizationFunc      func(ctx context.Context, req interface{}) (interface{}, error)
	GetOrganizationFunc         func(ctx context.Context, orgID string) (interface{}, error)
	CreateUserGroupFunc         func(ctx context.Context, req interface{}) (interface{}, error)
	AddUserToGroupFunc          func(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroupFunc     func(ctx context.Context, userID, groupID string) error
	AssignRoleFunc              func(ctx context.Context, userID, orgID, roleName string) error
	CheckUserRoleFunc           func(ctx context.Context, userID, roleName string) (bool, error)
	AssignPermissionToGroupFunc func(ctx context.Context, groupID, resource, action string) error
	ValidateTokenFunc           func(ctx context.Context, token string) (*interfaces.UserInfo, error)
	HealthCheckFunc             func(ctx context.Context) error
}

func (m *MockAAAService) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx, subject, resource, action, object, orgID)
	}
	return true, nil // Default: allow all permissions for tests
}

func (m *MockAAAService) SeedRolesAndPermissions(ctx context.Context) error {
	if m.SeedRolesAndPermissionsFunc != nil {
		return m.SeedRolesAndPermissionsFunc(ctx)
	}
	return nil
}

func (m *MockAAAService) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAAAService) GetUser(ctx context.Context, userID string) (interface{}, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAAAService) GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error) {
	if m.GetUserByMobileFunc != nil {
		return m.GetUserByMobileFunc(ctx, mobileNumber)
	}
	return nil, nil
}

func (m *MockAAAService) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockAAAService) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	if m.CreateOrganizationFunc != nil {
		return m.CreateOrganizationFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAAAService) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	if m.GetOrganizationFunc != nil {
		return m.GetOrganizationFunc(ctx, orgID)
	}
	return nil, nil
}

func (m *MockAAAService) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	if m.CreateUserGroupFunc != nil {
		return m.CreateUserGroupFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockAAAService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	if m.AddUserToGroupFunc != nil {
		return m.AddUserToGroupFunc(ctx, userID, groupID)
	}
	return nil
}

func (m *MockAAAService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	if m.RemoveUserFromGroupFunc != nil {
		return m.RemoveUserFromGroupFunc(ctx, userID, groupID)
	}
	return nil
}

func (m *MockAAAService) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	if m.AssignRoleFunc != nil {
		return m.AssignRoleFunc(ctx, userID, orgID, roleName)
	}
	return nil
}

func (m *MockAAAService) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	if m.CheckUserRoleFunc != nil {
		return m.CheckUserRoleFunc(ctx, userID, roleName)
	}
	return true, nil
}

func (m *MockAAAService) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	if m.AssignPermissionToGroupFunc != nil {
		return m.AssignPermissionToGroupFunc(ctx, groupID, resource, action)
	}
	return nil
}

func (m *MockAAAService) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return &interfaces.UserInfo{}, nil
}

func (m *MockAAAService) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

// MockLogger provides a mock logger implementation for testing
type MockLogger struct {
	DebugFunc        func(msg string, fields ...interface{})
	InfoFunc         func(msg string, fields ...interface{})
	WarnFunc         func(msg string, fields ...interface{})
	ErrorFunc        func(msg string, fields ...interface{})
	FatalFunc        func(msg string, fields ...interface{})
	GetZapLoggerFunc func() *zap.Logger
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

func (m *MockLogger) GetZapLogger() *zap.Logger {
	if m.GetZapLoggerFunc != nil {
		return m.GetZapLoggerFunc()
	}
	// Return a no-op logger for tests
	return zap.NewNop()
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
