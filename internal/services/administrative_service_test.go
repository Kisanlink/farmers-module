package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockAAAServiceForAdmin is a mock implementation of AAAService for administrative tests
type MockAAAServiceForAdmin struct {
	mock.Mock
}

func (m *MockAAAServiceForAdmin) SeedRolesAndPermissions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAServiceForAdmin) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) GetUser(ctx context.Context, userID string) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error) {
	args := m.Called(ctx, mobileNumber)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceForAdmin) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAServiceForAdmin) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAServiceForAdmin) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAServiceForAdmin) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAServiceForAdmin) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UserInfo), args.Error(1)
}

func (m *MockAAAServiceForAdmin) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAdministrativeService_SeedRolesAndPermissions(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockAAAServiceForAdmin)
		request        *requests.SeedRolesAndPermissionsRequest
		expectedError  bool
		expectedResult bool
	}{
		{
			name: "successful seeding",
			setupMocks: func(mockAAA *MockAAAServiceForAdmin) {
				mockAAA.On("SeedRolesAndPermissions", mock.Anything).Return(nil)
			},
			request: &requests.SeedRolesAndPermissionsRequest{
				Force:  false,
				DryRun: false,
			},
			expectedError:  false,
			expectedResult: true,
		},
		{
			name: "seeding with force flag",
			setupMocks: func(mockAAA *MockAAAServiceForAdmin) {
				mockAAA.On("SeedRolesAndPermissions", mock.Anything).Return(nil)
			},
			request: &requests.SeedRolesAndPermissionsRequest{
				Force:  true,
				DryRun: false,
			},
			expectedError:  false,
			expectedResult: true,
		},
		{
			name: "seeding failure",
			setupMocks: func(mockAAA *MockAAAServiceForAdmin) {
				mockAAA.On("SeedRolesAndPermissions", mock.Anything).Return(errors.New("AAA service error"))
			},
			request: &requests.SeedRolesAndPermissionsRequest{
				Force:  false,
				DryRun: false,
			},
			expectedError:  true,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := new(MockAAAServiceForAdmin)
			tt.setupMocks(mockAAA)

			// Create service
			service := &AdministrativeServiceImpl{
				postgresManager: nil, // Not needed for this test
				gormDB:          nil, // Not needed for this test
				aaaService:      mockAAA,
			}

			// Execute
			ctx := context.Background()
			result, err := service.SeedRolesAndPermissions(ctx, tt.request)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.Success)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, result.Success)
			}

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestAdministrativeService_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockAAAService)
		request        *requests.HealthCheckRequest
		expectedStatus string
		expectedError  bool
	}{
		{
			name: "all components healthy",
			setupMocks: func(mockAAA *MockAAAService) {
				mockAAA.On("HealthCheck", mock.Anything).Return(nil)
			},
			request: &requests.HealthCheckRequest{
				Components: []string{},
			},
			expectedStatus: "healthy",
			expectedError:  false,
		},
		{
			name: "AAA service unhealthy",
			setupMocks: func(mockAAA *MockAAAService) {
				mockAAA.On("HealthCheck", mock.Anything).Return(errors.New("AAA service down"))
			},
			request: &requests.HealthCheckRequest{
				Components: []string{},
			},
			expectedStatus: "unhealthy",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := new(MockAAAService)
			tt.setupMocks(mockAAA)

			// Create service with proper in-memory database
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			assert.NoError(t, err)

			service := &AdministrativeServiceImpl{
				postgresManager: nil,
				gormDB:          db,
				aaaService:      mockAAA,
			}

			// Execute
			ctx := context.Background()
			result, err := service.HealthCheck(ctx, tt.request)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.NotEmpty(t, result.Components)
				assert.True(t, result.Duration > 0)
				assert.False(t, result.Timestamp.IsZero())
			}

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestAdministrativeService_HealthCheck_ComponentDetails(t *testing.T) {
	// Setup mocks
	mockAAA := new(MockAAAService)
	mockAAA.On("HealthCheck", mock.Anything).Return(nil)

	// Create service with proper in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	service := &AdministrativeServiceImpl{
		postgresManager: nil,
		gormDB:          db,
		aaaService:      mockAAA,
	}

	// Execute
	ctx := context.Background()
	req := &requests.HealthCheckRequest{}
	result, err := service.HealthCheck(ctx, req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "healthy", result.Status)

	// Check component details
	assert.Contains(t, result.Components, "database")
	assert.Contains(t, result.Components, "aaa_service")

	dbHealth := result.Components["database"]
	assert.Equal(t, "PostgreSQL Database", dbHealth.Name)
	assert.NotEmpty(t, dbHealth.Status)
	assert.False(t, dbHealth.Timestamp.IsZero())

	aaaHealth := result.Components["aaa_service"]
	assert.Equal(t, "AAA Service", aaaHealth.Name)
	assert.Equal(t, "healthy", aaaHealth.Status)
	assert.False(t, aaaHealth.Timestamp.IsZero())

	// Verify mock expectations
	mockAAA.AssertExpectations(t)
}

func TestAdministrativeService_HealthCheck_NilAAAService(t *testing.T) {
	// Create service with nil AAA service but proper database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	service := &AdministrativeServiceImpl{
		postgresManager: nil,
		gormDB:          db,
		aaaService:      nil,
	}

	// Execute
	ctx := context.Background()
	req := &requests.HealthCheckRequest{}
	result, err := service.HealthCheck(ctx, req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "unhealthy", result.Status)

	// Check AAA service component
	aaaHealth := result.Components["aaa_service"]
	assert.Equal(t, "unhealthy", aaaHealth.Status)
	assert.Contains(t, aaaHealth.Error, "AAA service not initialized")
}

func TestAdministrativeServiceWrapper_SeedRolesAndPermissions(t *testing.T) {
	// Create mock service
	mockService := &MockAdministrativeService{}
	wrapper := NewAdministrativeServiceWrapper(mockService)

	tests := []struct {
		name        string
		request     interface{}
		setupMocks  func(*MockAdministrativeService)
		expectError bool
	}{
		{
			name:    "nil request",
			request: nil,
			setupMocks: func(mock *MockAdministrativeService) {
				mock.On("SeedRolesAndPermissions", context.Background(), &requests.SeedRolesAndPermissionsRequest{}).
					Return(&responses.SeedRolesAndPermissionsResponse{Success: true}, nil)
			},
			expectError: false,
		},
		{
			name: "map request",
			request: map[string]interface{}{
				"force":   true,
				"dry_run": false,
			},
			setupMocks: func(mock *MockAdministrativeService) {
				expectedReq := &requests.SeedRolesAndPermissionsRequest{
					Force:  true,
					DryRun: false,
				}
				mock.On("SeedRolesAndPermissions", context.Background(), expectedReq).
					Return(&responses.SeedRolesAndPermissionsResponse{Success: true}, nil)
			},
			expectError: false,
		},
		{
			name:    "invalid request type",
			request: "invalid",
			setupMocks: func(mock *MockAdministrativeService) {
				// No mock setup needed as it should fail before calling the service
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockService)

			result, err := wrapper.SeedRolesAndPermissions(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// MockAdministrativeService for testing the wrapper
type MockAdministrativeService struct {
	mock.Mock
}

func (m *MockAdministrativeService) SeedRolesAndPermissions(ctx context.Context, req *requests.SeedRolesAndPermissionsRequest) (*responses.SeedRolesAndPermissionsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*responses.SeedRolesAndPermissionsResponse), args.Error(1)
}

func (m *MockAdministrativeService) HealthCheck(ctx context.Context, req *requests.HealthCheckRequest) (*responses.HealthCheckResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*responses.HealthCheckResponse), args.Error(1)
}
