package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAAAService for testing
type MockAAAService struct {
	mock.Mock
}

func (m *MockAAAService) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UserInfo), args.Error(1)
}

func (m *MockAAAService) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) SeedRolesAndPermissions(ctx context.Context, force bool) error {
	args := m.Called(ctx, force)
	return args.Error(0)
}

func (m *MockAAAService) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUser(ctx context.Context, userID string) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error) {
	args := m.Called(ctx, mobileNumber)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetOrCreateFarmersGroup(ctx context.Context, orgID string) (string, error) {
	args := m.Called(ctx, orgID)
	return args.String(0), args.Error(1)
}

func (m *MockAAAService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAService) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLogger for testing
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

func (m *MockLogger) GetZapLogger() *zap.Logger {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*zap.Logger)
	}
	return zap.NewNop()
}

func TestAuthenticationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		method         string
		authHeader     string
		setupMocks     func(*MockAAAService, *MockLogger)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Public route - no auth required",
			path:       "/api/v1/health",
			method:     "GET",
			authHeader: "",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Missing Authorization header",
			path:       "/api/v1/farmers",
			method:     "GET",
			authHeader: "",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid Authorization header format",
			path:       "/api/v1/farmers",
			method:     "GET",
			authHeader: "InvalidFormat token123",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Empty bearer token",
			path:       "/api/v1/farmers",
			method:     "GET",
			authHeader: "Bearer ",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Valid token - successful authentication",
			path:       "/api/v1/farmers",
			method:     "GET",
			authHeader: "Bearer valid_token_123",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				userInfo := &interfaces.UserInfo{
					UserID:   "user123",
					Username: "testuser",
					Email:    "test@example.com",
					Phone:    "+1234567890",
					Roles:    []string{"farmer"},
					OrgID:    "org123",
					OrgName:  "Test FPO",
					OrgType:  "FPO",
				}
				aaa.On("ValidateToken", mock.Anything, "valid_token_123").Return(userInfo, nil)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Invalid token - authentication failed",
			path:       "/api/v1/farmers",
			method:     "GET",
			authHeader: "Bearer invalid_token",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				aaa.On("ValidateToken", mock.Anything, "invalid_token").Return(nil, assert.AnError)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAService{}
			mockLogger := &MockLogger{}
			tt.setupMocks(mockAAA, mockLogger)

			// Setup Gin router
			router := gin.New()
			router.Use(RequestID())
			router.Use(AuthenticationMiddleware(mockAAA, mockLogger))

			// Add test route
			router.GET("/api/v1/farmers", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.GET("/api/v1/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mocks
			mockAAA.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestAuthorizationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		method         string
		userContext    *auth.UserContext
		orgContext     *auth.OrgContext
		setupMocks     func(*MockAAAService, *MockLogger)
		expectedStatus int
	}{
		{
			name:           "Public route - no authorization required",
			path:           "/api/v1/health",
			method:         "GET",
			setupMocks:     func(aaa *MockAAAService, logger *MockLogger) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Valid permission - authorization successful",
			path:   "/api/v1/farmers",
			method: "GET",
			userContext: &auth.UserContext{
				AAAUserID: "user123",
				Username:  "testuser",
			},
			orgContext: &auth.OrgContext{
				AAAOrgID: "org123",
			},
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				aaa.On("CheckPermission", mock.Anything, "user123", "farmer", "list", "", "org123").Return(true, nil)
				logger.On("Debug", mock.AnythingOfType("string"), mock.Anything).Return()
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return() // Handle token not found warning
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Permission denied - authorization failed",
			path:   "/api/v1/farmers",
			method: "GET",
			userContext: &auth.UserContext{
				AAAUserID: "user123",
				Username:  "testuser",
			},
			orgContext: &auth.OrgContext{
				AAAOrgID: "org123",
			},
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				aaa.On("CheckPermission", mock.Anything, "user123", "farmer", "list", "", "org123").Return(false, nil)
				logger.On("Warn", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Missing user context",
			path:   "/api/v1/farmers",
			method: "GET",
			setupMocks: func(aaa *MockAAAService, logger *MockLogger) {
				logger.On("Error", mock.AnythingOfType("string"), mock.Anything).Return()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAService{}
			mockLogger := &MockLogger{}
			tt.setupMocks(mockAAA, mockLogger)

			// Setup Gin router
			router := gin.New()
			router.Use(RequestID())

			// Add middleware that sets context
			router.Use(func(c *gin.Context) {
				if tt.userContext != nil {
					c.Set("user_context", tt.userContext)
				}
				if tt.orgContext != nil {
					c.Set("org_context", tt.orgContext)
				}
				c.Next()
			})

			router.Use(AuthorizationMiddleware(mockAAA, mockLogger))

			// Add test routes
			router.GET("/api/v1/farmers", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.GET("/api/v1/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "healthy"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify mocks
			mockAAA.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetPermissionForRoute(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedPerm   auth.Permission
		expectedExists bool
	}{
		{
			name:           "Farmer list route",
			method:         "GET",
			path:           "/api/v1/farmers",
			expectedPerm:   auth.Permission{Resource: "farmer", Action: "list"},
			expectedExists: true,
		},
		{
			name:           "Farmer detail route with ID",
			method:         "GET",
			path:           "/api/v1/farmers/123e4567-e89b-12d3-a456-426614174000",
			expectedPerm:   auth.Permission{Resource: "farmer", Action: "read"},
			expectedExists: true,
		},
		{
			name:           "Farm create route",
			method:         "POST",
			path:           "/api/v1/farms",
			expectedPerm:   auth.Permission{Resource: "farm", Action: "create"},
			expectedExists: true,
		},
		{
			name:           "Unknown route",
			method:         "GET",
			path:           "/api/v1/unknown",
			expectedPerm:   auth.Permission{},
			expectedExists: false,
		},
		{
			name:           "Health check route",
			method:         "GET",
			path:           "/api/v1/health",
			expectedPerm:   auth.Permission{Resource: "system", Action: "health"},
			expectedExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, exists := auth.GetPermissionForRoute(tt.method, tt.path)
			assert.Equal(t, tt.expectedExists, exists)
			if exists {
				assert.Equal(t, tt.expectedPerm.Resource, perm.Resource)
				assert.Equal(t, tt.expectedPerm.Action, perm.Action)
			}
		})
	}
}

func TestIsPublicRoute(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{
			name:     "Health check is public",
			method:   "GET",
			path:     "/api/v1/health",
			expected: true,
		},
		{
			name:     "Docs path is public",
			method:   "GET",
			path:     "/docs/swagger.json",
			expected: true,
		},
		{
			name:     "Swagger path is public",
			method:   "GET",
			path:     "/swagger/index.html",
			expected: true,
		},
		{
			name:     "Farmer route is not public",
			method:   "GET",
			path:     "/api/v1/farmers",
			expected: false,
		},
		{
			name:     "Farm route is not public",
			method:   "POST",
			path:     "/api/v1/farms",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsPublicRoute(tt.method, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
