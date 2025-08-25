package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockFPOService is a mock implementation of FPOService
type MockFPOService struct {
	mock.Mock
}

func (m *MockFPOService) CreateFPO(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFPOService) RegisterFPORef(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFPOService) GetFPORef(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Named(name string) interfaces.Logger {
	args := m.Called(name)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) With(fields ...zap.Field) interfaces.Logger {
	args := m.Called(fields)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func setupFPOHandler() (*FPOHandler, *MockFPOService, *MockLogger) {
	mockService := &MockFPOService{}
	mockLogger := &MockLogger{}
	handler := NewFPOHandler(mockService, mockLogger)
	return handler, mockService, mockLogger
}

func setupGinContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	// Set common context values
	c.Set("request_id", "test-request-123")
	c.Set("aaa_subject", "user123")
	c.Set("aaa_org", "org123")
	c.Set("timestamp", time.Now().Format(time.RFC3339))

	return c, w
}

func TestFPOHandler_CreateFPO_Success(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data
	requestBody := requests.CreateFPORequest{
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		Description:    "Test FPO Description",
		CEOUser: requests.CEOUserData{
			FirstName:   "John",
			LastName:    "Doe",
			PhoneNumber: "+919876543210",
			Email:       "john.doe@example.com",
			Password:    "password123",
		},
		BusinessConfig: map[string]string{"type": "agricultural"},
		Metadata:       map[string]string{"region": "north"},
	}

	expectedResponse := &responses.CreateFPOData{
		FPOID:     "fpo123",
		AAAOrgID:  "org456",
		Name:      "Test FPO",
		CEOUserID: "user789",
		UserGroups: []responses.UserGroupData{
			{
				GroupID:   "group1",
				Name:      "directors",
				OrgID:     "org456",
				CreatedAt: time.Now().Format(time.RFC3339),
			},
		},
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("CreateFPO", mock.Anything, mock.AnythingOfType("*requests.CreateFPORequest")).Return(expectedResponse, nil)

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/create", requestBody)

	// Execute
	handler.CreateFPO(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response responses.CreateFPOResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO created successfully", response.Message)
	assert.Equal(t, "fpo123", response.Data.FPOID)
	assert.Equal(t, "org456", response.Data.AAAOrgID)
	assert.Equal(t, "user789", response.Data.CEOUserID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_CreateFPO_ValidationError(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data with missing required field
	requestBody := requests.CreateFPORequest{
		RegistrationNo: "FPO123456",
		CEOUser: requests.CEOUserData{
			FirstName:   "John",
			LastName:    "Doe",
			PhoneNumber: "+919876543210",
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("CreateFPO", mock.Anything, mock.AnythingOfType("*requests.CreateFPORequest")).Return(nil, errors.New("FPO name is required"))

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/create", requestBody)

	// Execute
	handler.CreateFPO(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO name is required", response["error"])

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_CreateFPO_InvalidJSON(t *testing.T) {
	handler, _, mockLogger := setupFPOHandler()

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

	// Setup Gin context with invalid JSON
	c, w := setupGinContext("POST", "/fpo/create", "invalid json")

	// Execute
	handler.CreateFPO(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestFPOHandler_RegisterFPORef_Success(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data
	requestBody := requests.RegisterFPORefRequest{
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		BusinessConfig: map[string]string{"type": "agricultural"},
		Metadata:       map[string]string{"region": "north"},
	}

	expectedResponse := &responses.FPORefData{
		ID:             "fpo_ref_123",
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		BusinessConfig: map[string]string{"type": "agricultural"},
		Status:         "ACTIVE",
		CreatedAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("RegisterFPORef", mock.Anything, mock.AnythingOfType("*requests.RegisterFPORefRequest")).Return(expectedResponse, nil)

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/register", requestBody)

	// Execute
	handler.RegisterFPORef(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response responses.FPORefResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO reference registered successfully", response.Message)
	assert.Equal(t, "fpo_ref_123", response.Data.ID)
	assert.Equal(t, "org123", response.Data.AAAOrgID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_RegisterFPORef_AlreadyExists(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data
	requestBody := requests.RegisterFPORefRequest{
		AAAOrgID: "org123",
		Name:     "Test FPO",
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("RegisterFPORef", mock.Anything, mock.AnythingOfType("*requests.RegisterFPORefRequest")).Return(nil, errors.New("FPO reference already exists for organization ID: org123"))

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/register", requestBody)

	// Execute
	handler.RegisterFPORef(c)

	// Assertions
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO reference already exists for organization ID: org123", response["error"])

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_GetFPORef_Success(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	expectedResponse := &responses.FPORefData{
		ID:             "fpo_ref_123",
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		Status:         "ACTIVE",
		CreatedAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("GetFPORef", mock.Anything, "org123").Return(expectedResponse, nil)

	// Setup Gin context
	c, w := setupGinContext("GET", "/fpo/reference/org123", nil)
	c.Params = []gin.Param{{Key: "aaa_org_id", Value: "org123"}}

	// Execute
	handler.GetFPORef(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response responses.FPORefResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO reference retrieved successfully", response.Message)
	assert.Equal(t, "fpo_ref_123", response.Data.ID)
	assert.Equal(t, "org123", response.Data.AAAOrgID)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_GetFPORef_NotFound(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("GetFPORef", mock.Anything, "org123").Return(nil, errors.New("FPO reference not found for organization ID: org123"))

	// Setup Gin context
	c, w := setupGinContext("GET", "/fpo/reference/org123", nil)
	c.Params = []gin.Param{{Key: "aaa_org_id", Value: "org123"}}

	// Execute
	handler.GetFPORef(c)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FPO reference not found for organization ID: org123", response["error"])

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_GetFPORef_MissingParameter(t *testing.T) {
	handler, _, mockLogger := setupFPOHandler()

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()

	// Setup Gin context without parameter
	c, w := setupGinContext("GET", "/fpo/reference/", nil)
	c.Params = []gin.Param{{Key: "aaa_org_id", Value: ""}}

	// Execute
	handler.GetFPORef(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "AAA organization ID is required", response["error"])
}

func TestFPOHandler_CreateFPO_ServiceError(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data
	requestBody := requests.CreateFPORequest{
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		CEOUser: requests.CEOUserData{
			FirstName:   "John",
			LastName:    "Doe",
			PhoneNumber: "+919876543210",
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("CreateFPO", mock.Anything, mock.AnythingOfType("*requests.CreateFPORequest")).Return(nil, errors.New("failed to create organization: AAA service unavailable"))

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/create", requestBody)

	// Execute
	handler.CreateFPO(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "failed to create organization: AAA service unavailable", response["error"])

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestFPOHandler_CreateFPO_UserExistsError(t *testing.T) {
	handler, mockService, mockLogger := setupFPOHandler()

	// Test data
	requestBody := requests.CreateFPORequest{
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		CEOUser: requests.CEOUserData{
			FirstName:   "John",
			LastName:    "Doe",
			PhoneNumber: "+919876543210",
		},
	}

	// Setup mocks
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
	mockService.On("CreateFPO", mock.Anything, mock.AnythingOfType("*requests.CreateFPORequest")).Return(nil, errors.New("failed to create CEO user: user already exists"))

	// Setup Gin context
	c, w := setupGinContext("POST", "/fpo/create", requestBody)

	// Execute
	handler.CreateFPO(c)

	// Assertions
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "failed to create CEO user: user already exists", response["error"])

	// Verify mocks
	mockService.AssertExpectations(t)
}
