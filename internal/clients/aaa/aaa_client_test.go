package aaa

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockUserServiceV2Client is a mock implementation of UserServiceV2Client
type MockUserServiceV2Client struct {
	mock.Mock
}

func (m *MockUserServiceV2Client) Login(ctx context.Context, in *proto.LoginRequestV2, opts ...grpc.CallOption) (*proto.LoginResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.LoginResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) Register(ctx context.Context, in *proto.RegisterRequestV2, opts ...grpc.CallOption) (*proto.RegisterResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RegisterResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) GetUser(ctx context.Context, in *proto.GetUserRequestV2, opts ...grpc.CallOption) (*proto.GetUserResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.GetUserResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) GetAllUsers(ctx context.Context, in *proto.GetAllUsersRequestV2, opts ...grpc.CallOption) (*proto.GetAllUsersResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.GetAllUsersResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) UpdateUser(ctx context.Context, in *proto.UpdateUserRequestV2, opts ...grpc.CallOption) (*proto.UpdateUserResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.UpdateUserResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) DeleteUser(ctx context.Context, in *proto.DeleteUserRequestV2, opts ...grpc.CallOption) (*proto.DeleteUserResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.DeleteUserResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) RefreshToken(ctx context.Context, in *proto.RefreshTokenRequestV2, opts ...grpc.CallOption) (*proto.RefreshTokenResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.RefreshTokenResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) Logout(ctx context.Context, in *proto.LogoutRequestV2, opts ...grpc.CallOption) (*proto.LogoutResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.LogoutResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) GetUserByPhone(ctx context.Context, in *proto.GetUserByPhoneRequestV2, opts ...grpc.CallOption) (*proto.GetUserResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.GetUserResponseV2), args.Error(1)
}

func (m *MockUserServiceV2Client) VerifyUserPassword(ctx context.Context, in *proto.VerifyPasswordRequestV2, opts ...grpc.CallOption) (*proto.VerifyPasswordResponseV2, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.VerifyPasswordResponseV2), args.Error(1)
}

// MockAuthorizationServiceClient is a mock implementation of AuthorizationServiceClient
type MockAuthorizationServiceClient struct {
	mock.Mock
}

func (m *MockAuthorizationServiceClient) Check(ctx context.Context, in *proto.CheckRequest, opts ...grpc.CallOption) (*proto.CheckResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.CheckResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) BatchCheck(ctx context.Context, in *proto.BatchCheckRequest, opts ...grpc.CallOption) (*proto.BatchCheckResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.BatchCheckResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) LookupResources(ctx context.Context, in *proto.LookupResourcesRequest, opts ...grpc.CallOption) (*proto.LookupResourcesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.LookupResourcesResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) CheckColumns(ctx context.Context, in *proto.CheckColumnsRequest, opts ...grpc.CallOption) (*proto.CheckColumnsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.CheckColumnsResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) ListAllowedColumns(ctx context.Context, in *proto.ListAllowedColumnsRequest, opts ...grpc.CallOption) (*proto.ListAllowedColumnsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ListAllowedColumnsResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) Explain(ctx context.Context, in *proto.ExplainRequest, opts ...grpc.CallOption) (*proto.ExplainResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*proto.ExplainResponse), args.Error(1)
}

// Helper function to create a test client with mocks
func createTestClient() (*Client, *MockUserServiceV2Client, *MockAuthorizationServiceClient) {
	mockUserClient := &MockUserServiceV2Client{}
	mockAuthzClient := &MockAuthorizationServiceClient{}

	client := &Client{
		conn:        nil, // Not needed for unit tests
		config:      &config.Config{},
		userClient:  mockUserClient,
		authzClient: mockAuthzClient,
	}

	return client, mockUserClient, mockAuthzClient
}

func TestCreateUser_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserRequest{
		Username:    "testuser",
		PhoneNumber: "+919876543210",
		Email:       "test@example.com",
		Password:    "password123",
		FullName:    "Test User",
	}

	expectedResponse := &proto.RegisterResponseV2{
		StatusCode: 201,
		Message:    "User created successfully",
		User: &proto.UserV2{
			Id:       "user123",
			Username: "testuser",
			Status:   "active",
		},
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*proto.RegisterRequestV2")).Return(expectedResponse, nil)

	response, err := client.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "user123", response.UserID)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "active", response.Status)
	mockUserClient.AssertExpectations(t)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserRequest{
		Username:    "existinguser",
		PhoneNumber: "+919876543210",
		Email:       "existing@example.com",
		Password:    "password123",
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*proto.RegisterRequestV2")).Return(
		(*proto.RegisterResponseV2)(nil), status.Error(codes.AlreadyExists, "user already exists"))

	response, err := client.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "user already exists")
	mockUserClient.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()
	userID := "user123"

	expectedResponse := &proto.GetUserResponseV2{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &proto.UserV2{
			Id:          "user123",
			Username:    "testuser",
			PhoneNumber: "+919876543210",
			Email:       "test@example.com",
			FullName:    "Test User",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	mockUserClient.On("GetUser", ctx, &proto.GetUserRequestV2{Id: userID}).Return(expectedResponse, nil)

	userData, err := client.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, userData)
	assert.Equal(t, "user123", userData.ID)
	assert.Equal(t, "testuser", userData.Username)
	assert.Equal(t, "+919876543210", userData.PhoneNumber)
	mockUserClient.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()
	userID := "nonexistent"

	mockUserClient.On("GetUser", ctx, &proto.GetUserRequestV2{Id: userID}).Return(
		(*proto.GetUserResponseV2)(nil), status.Error(codes.NotFound, "user not found"))

	userData, err := client.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, userData)
	assert.Contains(t, err.Error(), "user not found")
	mockUserClient.AssertExpectations(t)
}

func TestGetUserByPhone_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()
	phoneNumber := "+919876543210"

	expectedResponse := &proto.GetUserResponseV2{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &proto.UserV2{
			Id:          "user123",
			Username:    "testuser",
			PhoneNumber: phoneNumber,
			Email:       "test@example.com",
			Status:      "active",
		},
	}

	expectedRequest := &proto.GetUserByPhoneRequestV2{
		PhoneNumber:        phoneNumber,
		CountryCode:        "+91",
		IncludeRoles:       false,
		IncludePermissions: false,
	}

	mockUserClient.On("GetUserByPhone", ctx, expectedRequest).Return(expectedResponse, nil)

	userData, err := client.GetUserByPhone(ctx, phoneNumber)

	assert.NoError(t, err)
	assert.NotNil(t, userData)
	assert.Equal(t, "user123", userData.ID)
	assert.Equal(t, phoneNumber, userData.PhoneNumber)
	mockUserClient.AssertExpectations(t)
}

func TestGetUserByPhone_EmptyPhone(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	userData, err := client.GetUserByPhone(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, userData)
	assert.Contains(t, err.Error(), "phone number is required")
}

func TestCheckPermission_Success(t *testing.T) {
	client, _, mockAuthzClient := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "create"
	object := "farm456"
	orgID := "org789"

	expectedRequest := &proto.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   object,
		Action:       action,
	}

	expectedResponse := &proto.CheckResponse{
		Allowed: true,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestCheckPermission_Denied(t *testing.T) {
	client, _, mockAuthzClient := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "delete"
	object := "farm456"
	orgID := "org789"

	expectedRequest := &proto.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   object,
		Action:       action,
	}

	expectedResponse := &proto.CheckResponse{
		Allowed: false,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.False(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestCheckPermission_MissingParameters(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	// Test missing subject
	allowed, err := client.CheckPermission(ctx, "", "farm", "create", "farm456", "org789")
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "missing permission parameters")

	// Test missing resource
	allowed, err = client.CheckPermission(ctx, "user123", "", "create", "farm456", "org789")
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "missing permission parameters")

	// Test missing action
	allowed, err = client.CheckPermission(ctx, "user123", "farm", "", "farm456", "org789")
	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "missing permission parameters")
}

func TestCheckPermission_WildcardObject(t *testing.T) {
	client, _, mockAuthzClient := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "list"
	object := "" // Empty object should become wildcard
	orgID := "org789"

	expectedRequest := &proto.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   "*", // Should be converted to wildcard
		Action:       action,
	}

	expectedResponse := &proto.CheckResponse{
		Allowed: true,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestHealthCheck_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &proto.GetUserRequestV2{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return NotFound error, which indicates service is healthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*proto.GetUserResponseV2)(nil),
		status.Error(codes.NotFound, "user not found"),
	)

	err := client.HealthCheck(ctx)

	assert.NoError(t, err)
	mockUserClient.AssertExpectations(t)
}

func TestHealthCheck_PermissionDeniedIsHealthy(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &proto.GetUserRequestV2{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return PermissionDenied error, which indicates service is healthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*proto.GetUserResponseV2)(nil),
		status.Error(codes.PermissionDenied, "permission denied"),
	)

	err := client.HealthCheck(ctx)

	assert.NoError(t, err) // Permission denied means service is healthy
	mockUserClient.AssertExpectations(t)
}

func TestHealthCheck_ServiceUnavailable(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &proto.GetUserRequestV2{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return Unavailable error, which indicates service is unhealthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*proto.GetUserResponseV2)(nil),
		status.Error(codes.Unavailable, "service unavailable"),
	)

	err := client.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AAA service health check failed")
	mockUserClient.AssertExpectations(t)
}

func TestAddRequestMetadata(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	requestID := "req123"
	userID := "user456"

	newCtx := client.AddRequestMetadata(ctx, requestID, userID)

	assert.NotEqual(t, ctx, newCtx)
	// Note: In a real test, you would extract and verify the metadata
	// This requires access to grpc metadata package functionality
}

func TestValidateToken_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	tokenData, err := client.ValidateToken(ctx, "some-token")

	assert.Error(t, err)
	assert.Nil(t, tokenData)
	assert.Contains(t, err.Error(), "ValidateToken not implemented")
}

func TestValidateToken_EmptyToken(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	tokenData, err := client.ValidateToken(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, tokenData)
	assert.Contains(t, err.Error(), "token is required")
}

func TestCreateOrganization_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateOrganizationRequest{
		Name:        "Test FPO",
		Description: "Test FPO Description",
		Type:        "FPO",
		CEOUserID:   "user123",
	}

	response, err := client.CreateOrganization(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "CreateOrganization not implemented")
}

func TestCreateUserGroup_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserGroupRequest{
		Name:        "Directors",
		Description: "FPO Directors Group",
		OrgID:       "org123",
		Permissions: []string{"farm.create", "farm.update"},
	}

	response, err := client.CreateUserGroup(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "CreateUserGroup not implemented")
}

func TestAssignRole_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	err := client.AssignRole(ctx, "user123", "org456", "farmer")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AssignRole not implemented")
}

func TestCheckUserRole_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	hasRole, err := client.CheckUserRole(ctx, "user123", "farmer")

	assert.Error(t, err)
	assert.False(t, hasRole)
	assert.Contains(t, err.Error(), "CheckUserRole not implemented")
}

func TestSeedRolesAndPermissions_NotImplemented(t *testing.T) {
	client, _, _ := createTestClient()
	ctx := context.Background()

	err := client.SeedRolesAndPermissions(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SeedRolesAndPermissions not implemented")
}

// Test backward compatibility methods
func TestCreateUserLegacy_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()

	username := "testuser"
	mobileNumber := "+919876543210"
	password := "password123"
	countryCode := "+91"
	aadhaarNumber := "123456789012"

	expectedResponse := &proto.RegisterResponseV2{
		StatusCode: 201,
		Message:    "User created successfully",
		User: &proto.UserV2{
			Id:       "user123",
			Username: username,
			Status:   "active",
		},
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*proto.RegisterRequestV2")).Return(expectedResponse, nil)

	userID, err := client.CreateUserLegacy(ctx, username, mobileNumber, password, countryCode, &aadhaarNumber)

	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	mockUserClient.AssertExpectations(t)
}

func TestGetUserByMobile_Success(t *testing.T) {
	client, mockUserClient, _ := createTestClient()
	ctx := context.Background()
	mobileNumber := "+919876543210"

	expectedResponse := &proto.GetUserResponseV2{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &proto.UserV2{
			Id:          "user123",
			Username:    "testuser",
			PhoneNumber: mobileNumber,
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	expectedRequest := &proto.GetUserByPhoneRequestV2{
		PhoneNumber:        mobileNumber,
		CountryCode:        "+91",
		IncludeRoles:       false,
		IncludePermissions: false,
	}

	mockUserClient.On("GetUserByPhone", ctx, expectedRequest).Return(expectedResponse, nil)

	userData, err := client.GetUserByMobile(ctx, mobileNumber)

	assert.NoError(t, err)
	assert.NotNil(t, userData)

	// Check that it returns a map (backward compatibility)
	assert.Equal(t, "user123", userData["id"])
	assert.Equal(t, mobileNumber, userData["mobile_number"])

	mockUserClient.AssertExpectations(t)
}
