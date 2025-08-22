package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockFarmerLinkageService is a mock implementation of FarmerLinkageService
type MockFarmerLinkageService struct {
	mock.Mock
}

func (m *MockFarmerLinkageService) LinkFarmerToFPO(ctx context.Context, req interface{}) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockFarmerLinkageService) UnlinkFarmerFromFPO(ctx context.Context, req interface{}) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockFarmerLinkageService) GetFarmerLinkage(ctx context.Context, farmerID, orgID string) (interface{}, error) {
	args := m.Called(ctx, farmerID, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmerLinkageService) AssignKisanSathi(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmerLinkageService) ReassignOrRemoveKisanSathi(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockFarmerLinkageService) CreateKisanSathiUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// SimpleLogger is a simple logger implementation for testing
type SimpleLogger struct{}

func (s *SimpleLogger) Debug(msg string, fields ...zap.Field)      {}
func (s *SimpleLogger) Info(msg string, fields ...zap.Field)       {}
func (s *SimpleLogger) Warn(msg string, fields ...zap.Field)       {}
func (s *SimpleLogger) Error(msg string, fields ...zap.Field)      {}
func (s *SimpleLogger) Fatal(msg string, fields ...zap.Field)      {}
func (s *SimpleLogger) With(fields ...zap.Field) interfaces.Logger { return s }
func (s *SimpleLogger) Named(name string) interfaces.Logger        { return s }
func (s *SimpleLogger) Sync() error                                { return nil }

func TestAssignKisanSathi_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockFarmerLinkageService)
		expectedStatus int
	}{
		{
			name: "successful assignment",
			requestBody: requests.AssignKisanSathiRequest{
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			setupMock: func(m *MockFarmerLinkageService) {
				assignmentData := &responses.KisanSathiAssignmentData{
					ID:               "link123",
					AAAUserID:        "user123",
					AAAOrgID:         "org456",
					KisanSathiUserID: stringPtr("ks789"),
					Status:           "ACTIVE",
				}
				m.On("AssignKisanSathi", mock.Anything, mock.Anything).Return(assignmentData, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			requestBody: requests.AssignKisanSathiRequest{
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			setupMock: func(m *MockFarmerLinkageService) {
				m.On("AssignKisanSathi", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFarmerLinkageService{}
			// Create a simple logger that doesn't panic
			logger := &SimpleLogger{}
			tt.setupMock(mockService)

			router := gin.New()
			router.POST("/assign", AssignKisanSathi(mockService, logger))

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/assign", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestCreateKisanSathiUser_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockFarmerLinkageService)
		expectedStatus int
	}{
		{
			name: "successful user creation",
			requestBody: requests.CreateKisanSathiUserRequest{
				Username:    "kisansathi1",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			setupMock: func(m *MockFarmerLinkageService) {
				userData := &responses.KisanSathiUserData{
					ID:          "ks123",
					Username:    "kisansathi1",
					PhoneNumber: "+919876543210",
					Email:       "ks1@example.com",
					FullName:    "KisanSathi One",
					Role:        "KisanSathi",
					Status:      "ACTIVE",
				}
				m.On("CreateKisanSathiUser", mock.Anything, mock.Anything).Return(userData, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "service error",
			requestBody: requests.CreateKisanSathiUserRequest{
				Username:    "kisansathi1",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			setupMock: func(m *MockFarmerLinkageService) {
				m.On("CreateKisanSathiUser", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockFarmerLinkageService{}
			// Create a simple logger that doesn't panic
			logger := &SimpleLogger{}
			tt.setupMock(mockService)

			router := gin.New()
			router.POST("/create-user", CreateKisanSathiUser(mockService, logger))

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
