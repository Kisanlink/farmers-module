package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockAAAClient is a mock implementation of the AAA client
type MockAAAClient struct {
	mock.Mock
}

func (m *MockAAAClient) CreateUser(ctx context.Context, req *aaa.CreateUserRequest) (*aaa.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.CreateUserResponse), args.Error(1)
}

func (m *MockAAAClient) GetUser(ctx context.Context, userID string) (*aaa.UserData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.UserData), args.Error(1)
}

func (m *MockAAAClient) GetUserByPhone(ctx context.Context, phone string) (*aaa.UserData, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.UserData), args.Error(1)
}

func (m *MockAAAClient) GetUserByEmail(ctx context.Context, email string) (*aaa.UserData, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.UserData), args.Error(1)
}

func (m *MockAAAClient) GetUserByMobile(ctx context.Context, mobile string) (map[string]interface{}, error) {
	args := m.Called(ctx, mobile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAAAClient) CreateOrganization(ctx context.Context, req *aaa.CreateOrganizationRequest) (*aaa.CreateOrganizationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.CreateOrganizationResponse), args.Error(1)
}

func (m *MockAAAClient) GetOrganization(ctx context.Context, orgID string) (*aaa.OrganizationData, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.OrganizationData), args.Error(1)
}

func (m *MockAAAClient) CreateUserGroup(ctx context.Context, req *aaa.CreateUserGroupRequest) (*aaa.CreateUserGroupResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aaa.CreateUserGroupResponse), args.Error(1)
}

func (m *MockAAAClient) GetOrCreateFarmersGroup(ctx context.Context, orgID string) (string, error) {
	args := m.Called(ctx, orgID)
	return args.String(0), args.Error(1)
}

func (m *MockAAAClient) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAClient) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAClient) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAClient) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

func (m *MockAAAClient) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	args := m.Called(ctx, subject, resource, action, object, orgID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAClient) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAAAClient) SeedRolesAndPermissions(ctx context.Context, force bool) error {
	args := m.Called(ctx, force)
	return args.Error(0)
}

func (m *MockAAAClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Helper function to create a test service with mock client
func createTestAAAService(mockClient AAAClientInterface) *AAAServiceImpl {
	return &AAAServiceImpl{
		config: &config.Config{},
		client: mockClient,
	}
}

func TestAAAService_CreateUser_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	req := map[string]interface{}{
		"username":     "testuser",
		"phone_number": "+919876543210",
		"email":        "test@example.com",
		"password":     "password123",
		"full_name":    "Test User",
		"role":         "farmer",
	}

	expectedResponse := &aaa.CreateUserResponse{
		UserID:   "user123",
		Username: "testuser",
		Status:   "active",
	}

	mockClient.On("CreateUser", ctx, mock.AnythingOfType("*aaa.CreateUserRequest")).Return(expectedResponse, nil)

	result, err := service.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "user123", resultMap["id"])
	assert.Equal(t, "testuser", resultMap["username"])
	assert.Equal(t, "active", resultMap["status"])

	mockClient.AssertExpectations(t)
}

func TestAAAService_CreateUser_InvalidRequest(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	// Invalid request format (not a map)
	req := "invalid request"

	result, err := service.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid request format")
}

func TestAAAService_CreateUser_ClientUnavailable(t *testing.T) {
	service := &AAAServiceImpl{
		config: &config.Config{},
		client: nil, // No client available
	}
	ctx := context.Background()

	req := map[string]interface{}{
		"username": "testuser",
	}

	result, err := service.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "AAA client not available")
}

func TestAAAService_GetUser_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()
	userID := "user123"

	expectedUserData := &aaa.UserData{
		ID:       "user123",
		Username: "testuser",
		Status:   "active",
	}

	mockClient.On("GetUser", ctx, userID).Return(expectedUserData, nil)

	result, err := service.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUserData, result)
	mockClient.AssertExpectations(t)
}

func TestAAAService_GetUser_NotFound(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()
	userID := "nonexistent"

	mockClient.On("GetUser", ctx, userID).Return(nil, status.Error(codes.NotFound, "user not found"))

	result, err := service.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "resource not found")
	mockClient.AssertExpectations(t)
}

func TestAAAService_GetUserByMobile_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()
	mobileNumber := "+919876543210"

	expectedUserData := map[string]interface{}{
		"id":            "user123",
		"username":      "testuser",
		"mobile_number": mobileNumber,
		"status":        "active",
	}

	mockClient.On("GetUserByMobile", ctx, mobileNumber).Return(expectedUserData, nil)

	result, err := service.GetUserByMobile(ctx, mobileNumber)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUserData, result)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckPermission_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	mockClient.On("CheckPermission", ctx, "user123", "farm", "create", "farm456", "org789").Return(true, nil)

	allowed, err := service.CheckPermission(ctx, "user123", "farm", "create", "farm456", "org789")

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckPermission_InvalidRequest(t *testing.T) {
	mockClient := &MockAAAClient{}

	// Setup mock to return error for invalid request
	mockClient.On("CheckPermission", mock.Anything, "", "farm", "create", "farm456", "org789").Return(false, fmt.Errorf("missing permission parameters"))

	service := createTestAAAService(mockClient)
	ctx := context.Background()

	// Test with empty subject
	allowed, err := service.CheckPermission(ctx, "", "farm", "create", "farm456", "org789")

	assert.Error(t, err)
	assert.False(t, allowed)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckPermission_MissingSubject(t *testing.T) {
	mockClient := &MockAAAClient{}

	// Setup mock to return error for missing subject
	mockClient.On("CheckPermission", mock.Anything, "", "farm", "create", "farm456", "org789").Return(false, fmt.Errorf("missing permission parameters"))

	service := createTestAAAService(mockClient)
	ctx := context.Background()

	// Test with empty subject
	allowed, err := service.CheckPermission(ctx, "", "farm", "create", "farm456", "org789")

	assert.Error(t, err)
	assert.False(t, allowed)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckPermission_ClientUnavailable(t *testing.T) {
	service := &AAAServiceImpl{
		config: &config.Config{},
		client: nil, // No client available
	}
	ctx := context.Background()

	allowed, err := service.CheckPermission(ctx, "user123", "farm", "create", "farm456", "org789")

	assert.NoError(t, err) // Should allow when client unavailable
	assert.True(t, allowed)
}

func TestAAAService_SeedRolesAndPermissions_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	mockClient.On("SeedRolesAndPermissions", ctx, false).Return(nil)

	err := service.SeedRolesAndPermissions(ctx, false)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAAAService_SeedRolesAndPermissions_ClientUnavailable(t *testing.T) {
	service := &AAAServiceImpl{
		config: &config.Config{},
		client: nil, // No client available
	}
	ctx := context.Background()

	err := service.SeedRolesAndPermissions(ctx, false)

	assert.NoError(t, err) // Should not error when client unavailable
}

func TestAAAService_CreateOrganization_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	req := map[string]interface{}{
		"name":        "Test FPO",
		"description": "Test FPO Description",
		"type":        "FPO",
		"ceo_user_id": "user123",
		"metadata":    map[string]string{"key": "value"},
	}

	expectedResponse := &aaa.CreateOrganizationResponse{
		OrgID:  "org123",
		Name:   "Test FPO",
		Status: "active",
	}

	mockClient.On("CreateOrganization", ctx, mock.AnythingOfType("*aaa.CreateOrganizationRequest")).Return(expectedResponse, nil)

	result, err := service.CreateOrganization(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "org123", resultMap["org_id"])
	assert.Equal(t, "Test FPO", resultMap["name"])
	assert.Equal(t, "active", resultMap["status"])

	mockClient.AssertExpectations(t)
}

func TestAAAService_AssignRole_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	userID := "user123"
	orgID := "org456"
	roleName := "farmer"

	mockClient.On("AssignRole", ctx, userID, orgID, roleName).Return(nil)

	err := service.AssignRole(ctx, userID, orgID, roleName)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckUserRole_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	userID := "user123"
	roleName := "farmer"

	mockClient.On("CheckUserRole", ctx, userID, roleName).Return(true, nil)

	hasRole, err := service.CheckUserRole(ctx, userID, roleName)

	assert.NoError(t, err)
	assert.True(t, hasRole)
	mockClient.AssertExpectations(t)
}

func TestAAAService_CheckUserRole_ClientUnavailable(t *testing.T) {
	service := &AAAServiceImpl{
		config: &config.Config{},
		client: nil, // No client available
	}
	ctx := context.Background()

	hasRole, err := service.CheckUserRole(ctx, "user123", "farmer")

	assert.NoError(t, err)
	assert.False(t, hasRole) // Should return false when client unavailable
}

func TestAAAService_HealthCheck_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	mockClient.On("HealthCheck", ctx).Return(nil)

	err := service.HealthCheck(ctx)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestAAAService_HealthCheck_ClientUnavailable(t *testing.T) {
	service := &AAAServiceImpl{
		config: &config.Config{},
		client: nil, // No client available
	}
	ctx := context.Background()

	err := service.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AAA client not available")
}

func TestAAAService_ValidateToken_Success(t *testing.T) {
	mockClient := &MockAAAClient{}
	service := createTestAAAService(mockClient)
	ctx := context.Background()

	token := "valid-jwt-token"
	expectedTokenData := map[string]interface{}{
		"user_id": "user123",
		"org_id":  "org456",
		"exp":     1234567890,
	}

	mockClient.On("ValidateToken", ctx, token).Return(expectedTokenData, nil)

	tokenData, err := service.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, tokenData)
	assert.Equal(t, "user123", tokenData.UserID)
	assert.Equal(t, "org456", tokenData.OrgID)
	mockClient.AssertExpectations(t)
}

func TestAAAService_MapGRPCError_NotFound(t *testing.T) {
	service := &AAAServiceImpl{}

	grpcErr := status.Error(codes.NotFound, "resource not found")
	mappedErr := service.mapGRPCError(grpcErr, "test operation")

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "test operation failed: resource not found")
}

func TestAAAService_MapGRPCError_AlreadyExists(t *testing.T) {
	service := &AAAServiceImpl{}

	grpcErr := status.Error(codes.AlreadyExists, "resource already exists")
	mappedErr := service.mapGRPCError(grpcErr, "create user")

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "create user failed: resource already exists")
}

func TestAAAService_MapGRPCError_PermissionDenied(t *testing.T) {
	service := &AAAServiceImpl{}

	grpcErr := status.Error(codes.PermissionDenied, "access denied")
	mappedErr := service.mapGRPCError(grpcErr, "check permission")

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "check permission failed: permission denied")
}

func TestAAAService_MapGRPCError_Unavailable(t *testing.T) {
	service := &AAAServiceImpl{}

	grpcErr := status.Error(codes.Unavailable, "service unavailable")
	mappedErr := service.mapGRPCError(grpcErr, "health check")

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "health check failed: AAA service unavailable")
}

func TestAAAService_MapGRPCError_NonGRPCError(t *testing.T) {
	service := &AAAServiceImpl{}

	regularErr := assert.AnError
	mappedErr := service.mapGRPCError(regularErr, "some operation")

	assert.Error(t, mappedErr)
	assert.Contains(t, mappedErr.Error(), "some operation failed:")
}
