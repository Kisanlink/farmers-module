package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdministrativeService for testing handlers
type MockAdministrativeService struct {
	mock.Mock
}

func (m *MockAdministrativeService) SeedRolesAndPermissions(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAdministrativeService) HealthCheck(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// MockAAAService for testing permission check handler
type MockAAAServiceForHandlers struct {
	mock.Mock
}

func (m *MockAAAServiceForHandlers) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) SeedRolesAndPermissions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAServiceForHandlers) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) GetUser(ctx context.Context, userID string) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error) {
	args := m.Called(ctx, mobileNumber)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceForHandlers) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceForHandlers) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAServiceForHandlers) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceForHandlers) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAServiceForHandlers) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UserInfo), args.Error(1)
}

func (m *MockAAAServiceForHandlers) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestSeedRolesAndPermissions_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAdministrativeService)
		expectedStatus int
		expectedFields []string
	}{
		{
			name:        "successful seeding with empty request",
			requestBody: nil,
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.SeedRolesAndPermissionsResponse{
					Success:   true,
					Message:   "Roles and permissions seeded successfully",
					Duration:  time.Millisecond * 100,
					Timestamp: time.Now(),
				}
				mockService.On("SeedRolesAndPermissions", mock.Anything, mock.AnythingOfType("*requests.SeedRolesAndPermissionsRequest")).
					Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "message", "duration", "timestamp"},
		},
		{
			name: "successful seeding with request body",
			requestBody: map[string]interface{}{
				"force":   true,
				"dry_run": false,
			},
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.SeedRolesAndPermissionsResponse{
					Success:   true,
					Message:   "Roles and permissions seeded successfully",
					Duration:  time.Millisecond * 150,
					Timestamp: time.Now(),
				}
				mockService.On("SeedRolesAndPermissions", mock.Anything, mock.AnythingOfType("*requests.SeedRolesAndPermissionsRequest")).
					Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"success", "message", "duration", "timestamp"},
		},
		{
			name: "seeding failure",
			requestBody: map[string]interface{}{
				"force": false,
			},
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.SeedRolesAndPermissionsResponse{
					Success: false,
					Message: "Failed to seed roles and permissions",
					Error:   "AAA service unavailable",
				}
				mockService.On("SeedRolesAndPermissions", mock.Anything, mock.AnythingOfType("*requests.SeedRolesAndPermissionsRequest")).
					Return(response, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedFields: []string{"error", "message", "code", "correlation_id", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAdministrativeService)
			tt.setupMocks(mockService)

			// Create router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("correlation_id", "test-correlation-id")
				c.Set("user_id", "test-user-id")
				c.Set("org_id", "test-org-id")
				c.Next()
			})
			router.POST("/admin/seed", SeedRolesAndPermissions(mockService))

			// Create request
			var reqBody []byte
			if tt.requestBody != nil {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest("POST", "/admin/seed", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHealthCheck_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func(*MockAdministrativeService)
		expectedStatus int
		expectedFields []string
	}{
		{
			name:        "healthy system",
			queryParams: "",
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.HealthCheckResponse{
					Status:    "healthy",
					Message:   "All systems operational",
					Duration:  time.Millisecond * 50,
					Timestamp: time.Now(),
					Components: map[string]responses.ComponentHealth{
						"database": {
							Name:      "PostgreSQL Database",
							Status:    "healthy",
							Message:   "Database is healthy and responsive",
							Timestamp: time.Now(),
						},
						"aaa_service": {
							Name:      "AAA Service",
							Status:    "healthy",
							Message:   "AAA service is healthy and responsive",
							Timestamp: time.Now(),
						},
					},
				}
				mockService.On("HealthCheck", mock.Anything, mock.AnythingOfType("*requests.HealthCheckRequest")).
					Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "components", "duration", "timestamp"},
		},
		{
			name:        "unhealthy system",
			queryParams: "",
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.HealthCheckResponse{
					Status:    "unhealthy",
					Message:   "Some components are unhealthy",
					Duration:  time.Millisecond * 75,
					Timestamp: time.Now(),
					Components: map[string]responses.ComponentHealth{
						"database": {
							Name:      "PostgreSQL Database",
							Status:    "healthy",
							Message:   "Database is healthy and responsive",
							Timestamp: time.Now(),
						},
						"aaa_service": {
							Name:      "AAA Service",
							Status:    "unhealthy",
							Error:     "AAA service is not responding",
							Timestamp: time.Now(),
						},
					},
				}
				mockService.On("HealthCheck", mock.Anything, mock.AnythingOfType("*requests.HealthCheckRequest")).
					Return(response, nil)
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedFields: []string{"status", "components", "duration", "timestamp"},
		},
		{
			name:        "health check with components filter",
			queryParams: "?components=database",
			setupMocks: func(mockService *MockAdministrativeService) {
				response := &responses.HealthCheckResponse{
					Status:    "healthy",
					Duration:  time.Millisecond * 25,
					Timestamp: time.Now(),
					Components: map[string]responses.ComponentHealth{
						"database": {
							Name:      "PostgreSQL Database",
							Status:    "healthy",
							Message:   "Database is healthy and responsive",
							Timestamp: time.Now(),
						},
					},
				}
				mockService.On("HealthCheck", mock.Anything, mock.AnythingOfType("*requests.HealthCheckRequest")).
					Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "components", "duration", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAdministrativeService)
			tt.setupMocks(mockService)

			// Create router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("correlation_id", "test-correlation-id")
				c.Set("user_id", "test-user-id")
				c.Set("org_id", "test-org-id")
				c.Next()
			})
			router.GET("/admin/health", HealthCheck(mockService))

			// Create request
			req, _ := http.NewRequest("GET", "/admin/health"+tt.queryParams, nil)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Verify component structure if present
			if components, exists := response["components"]; exists {
				componentsMap := components.(map[string]interface{})
				for _, comp := range componentsMap {
					compMap := comp.(map[string]interface{})
					assert.Contains(t, compMap, "name")
					assert.Contains(t, compMap, "status")
					assert.Contains(t, compMap, "timestamp")
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestCheckPermission_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*MockAAAServiceForHandlers)
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "permission allowed",
			requestBody: map[string]interface{}{
				"subject":  "user123",
				"resource": "farm",
				"action":   "create",
				"object":   "farm456",
				"org_id":   "org789",
			},
			setupMocks: func(mockService *MockAAAServiceForHandlers) {
				mockService.On("CheckPermission", mock.Anything, "user123", "farm", "create", "farm456", "org789").
					Return(true, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"message", "data", "correlation_id", "timestamp"},
		},
		{
			name: "permission denied",
			requestBody: map[string]interface{}{
				"subject":  "user123",
				"resource": "farm",
				"action":   "delete",
				"object":   "farm456",
				"org_id":   "org789",
			},
			setupMocks: func(mockService *MockAAAServiceForHandlers) {
				mockService.On("CheckPermission", mock.Anything, "user123", "farm", "delete", "farm456", "org789").
					Return(false, nil)
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"message", "data", "correlation_id", "timestamp"},
		},
		{
			name: "invalid request - missing required field",
			requestBody: map[string]interface{}{
				"subject":  "user123",
				"resource": "farm",
				// missing action
			},
			setupMocks: func(mockService *MockAAAServiceForHandlers) {
				// No mock setup needed as validation should fail first
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "message", "code", "correlation_id", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAAAServiceForHandlers)
			tt.setupMocks(mockService)

			// Create router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("correlation_id", "test-correlation-id")
				c.Set("user_id", "test-user-id")
				c.Set("org_id", "test-org-id")
				c.Next()
			})
			router.POST("/admin/check-permission", CheckPermission(mockService))

			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/admin/check-permission", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Verify permission check result if successful
			if tt.expectedStatus == http.StatusOK {
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "allowed")
				assert.Equal(t, tt.requestBody["subject"], data["subject"])
				assert.Equal(t, tt.requestBody["resource"], data["resource"])
				assert.Equal(t, tt.requestBody["action"], data["action"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetAuditTrail_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "get audit trail without filters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"message", "data", "correlation_id", "timestamp"},
		},
		{
			name:           "get audit trail with filters",
			queryParams:    "?start_date=2023-01-01&end_date=2023-12-31&user_id=user123&action=create",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"message", "data", "correlation_id", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("correlation_id", "test-correlation-id")
				c.Set("user_id", "test-user-id")
				c.Set("org_id", "test-org-id")
				c.Next()
			})
			router.GET("/admin/audit", GetAuditTrail())

			// Create request
			req, _ := http.NewRequest("GET", "/admin/audit"+tt.queryParams, nil)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Verify data structure
			data := response["data"].(map[string]interface{})
			assert.Contains(t, data, "audit_logs")
			assert.Contains(t, data, "filters")
			assert.Contains(t, data, "total_count")
			assert.Contains(t, data, "page")
			assert.Contains(t, data, "page_size")

			// Verify filters are properly parsed
			filters := data["filters"].(map[string]interface{})
			if tt.queryParams != "" {
				assert.NotEmpty(t, filters["start_date"])
				assert.NotEmpty(t, filters["end_date"])
				assert.NotEmpty(t, filters["user_id"])
				assert.NotEmpty(t, filters["action"])
			}
		})
	}
}
