package aaa

import (
	"context"
	"testing"

	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockUserServiceClient is a mock implementation of UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LoginResponse), args.Error(1)
}

func (m *MockUserServiceClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.RegisterResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetAllUsers(ctx context.Context, in *pb.GetAllUsersRequest, opts ...grpc.CallOption) (*pb.GetAllUsersResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.GetAllUsersResponse), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest, opts ...grpc.CallOption) (*pb.UpdateUserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.UpdateUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest, opts ...grpc.CallOption) (*pb.DeleteUserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.DeleteUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.RefreshTokenResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.RefreshTokenResponse), args.Error(1)
}

func (m *MockUserServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.LogoutResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LogoutResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByPhone(ctx context.Context, in *pb.GetUserByPhoneRequest, opts ...grpc.CallOption) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func (m *MockUserServiceClient) VerifyUserPassword(ctx context.Context, in *pb.VerifyPasswordRequest, opts ...grpc.CallOption) (*pb.VerifyPasswordResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.VerifyPasswordResponse), args.Error(1)
}

// MockAuthorizationServiceClient is a mock implementation of AuthorizationServiceClient
type MockAuthorizationServiceClient struct {
	mock.Mock
}

func (m *MockAuthorizationServiceClient) Check(ctx context.Context, in *pb.CheckRequest, opts ...grpc.CallOption) (*pb.CheckResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.CheckResponse), args.Error(1)
}

// MockOrganizationServiceClient is a mock implementation of OrganizationServiceClient
type MockOrganizationServiceClient struct {
	mock.Mock
}

func (m *MockOrganizationServiceClient) CreateOrganization(ctx context.Context, in *pb.CreateOrganizationRequest, opts ...grpc.CallOption) (*pb.CreateOrganizationResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateOrganizationResponse), args.Error(1)
}

func (m *MockOrganizationServiceClient) GetOrganization(ctx context.Context, in *pb.GetOrganizationRequest, opts ...grpc.CallOption) (*pb.GetOrganizationResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetOrganizationResponse), args.Error(1)
}

func (m *MockOrganizationServiceClient) ListOrganizations(ctx context.Context, in *pb.ListOrganizationsRequest, opts ...grpc.CallOption) (*pb.ListOrganizationsResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) UpdateOrganization(ctx context.Context, in *pb.UpdateOrganizationRequest, opts ...grpc.CallOption) (*pb.UpdateOrganizationResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) DeleteOrganization(ctx context.Context, in *pb.DeleteOrganizationRequest, opts ...grpc.CallOption) (*pb.DeleteOrganizationResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) AddUserToOrganization(ctx context.Context, in *pb.AddUserToOrganizationRequest, opts ...grpc.CallOption) (*pb.AddUserToOrganizationResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) RemoveUserFromOrganization(ctx context.Context, in *pb.RemoveUserFromOrganizationRequest, opts ...grpc.CallOption) (*pb.RemoveUserFromOrganizationResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) ValidateOrganizationAccess(ctx context.Context, in *pb.ValidateOrganizationAccessRequest, opts ...grpc.CallOption) (*pb.ValidateOrganizationAccessResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) CreateRole(ctx context.Context, in *pb.CreateRoleRequest, opts ...grpc.CallOption) (*pb.CreateRoleResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) ListRoles(ctx context.Context, in *pb.ListRolesRequest, opts ...grpc.CallOption) (*pb.ListRolesResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) UpdateRole(ctx context.Context, in *pb.UpdateRoleRequest, opts ...grpc.CallOption) (*pb.UpdateRoleResponse, error) {
	return nil, nil
}

func (m *MockOrganizationServiceClient) DeleteRole(ctx context.Context, in *pb.DeleteRoleRequest, opts ...grpc.CallOption) (*pb.DeleteRoleResponse, error) {
	return nil, nil
}

// MockGroupServiceClient is a mock implementation of GroupServiceClient
type MockGroupServiceClient struct {
	mock.Mock
}

func (m *MockGroupServiceClient) CreateGroup(ctx context.Context, in *pb.CreateGroupRequest, opts ...grpc.CallOption) (*pb.CreateGroupResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateGroupResponse), args.Error(1)
}

func (m *MockGroupServiceClient) GetGroup(ctx context.Context, in *pb.GetGroupRequest, opts ...grpc.CallOption) (*pb.GetGroupResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) ListGroups(ctx context.Context, in *pb.ListGroupsRequest, opts ...grpc.CallOption) (*pb.ListGroupsResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) UpdateGroup(ctx context.Context, in *pb.UpdateGroupRequest, opts ...grpc.CallOption) (*pb.UpdateGroupResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) DeleteGroup(ctx context.Context, in *pb.DeleteGroupRequest, opts ...grpc.CallOption) (*pb.DeleteGroupResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) AddGroupMember(ctx context.Context, in *pb.AddGroupMemberRequest, opts ...grpc.CallOption) (*pb.AddGroupMemberResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AddGroupMemberResponse), args.Error(1)
}

func (m *MockGroupServiceClient) RemoveGroupMember(ctx context.Context, in *pb.RemoveGroupMemberRequest, opts ...grpc.CallOption) (*pb.RemoveGroupMemberResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RemoveGroupMemberResponse), args.Error(1)
}

func (m *MockGroupServiceClient) ListGroupMembers(ctx context.Context, in *pb.ListGroupMembersRequest, opts ...grpc.CallOption) (*pb.ListGroupMembersResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) LinkGroups(ctx context.Context, in *pb.LinkGroupsRequest, opts ...grpc.CallOption) (*pb.LinkGroupsResponse, error) {
	return nil, nil
}

func (m *MockGroupServiceClient) UnlinkGroups(ctx context.Context, in *pb.UnlinkGroupsRequest, opts ...grpc.CallOption) (*pb.UnlinkGroupsResponse, error) {
	return nil, nil
}

// MockRoleServiceClient is a mock implementation of RoleServiceClient
type MockRoleServiceClient struct {
	mock.Mock
}

func (m *MockRoleServiceClient) AssignRole(ctx context.Context, in *pb.AssignRoleRequest, opts ...grpc.CallOption) (*pb.AssignRoleResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AssignRoleResponse), args.Error(1)
}

func (m *MockRoleServiceClient) CheckUserRole(ctx context.Context, in *pb.CheckUserRoleRequest, opts ...grpc.CallOption) (*pb.CheckUserRoleResponse, error) {
	return nil, nil
}

func (m *MockRoleServiceClient) RemoveRole(ctx context.Context, in *pb.RemoveRoleRequest, opts ...grpc.CallOption) (*pb.RemoveRoleResponse, error) {
	return nil, nil
}

func (m *MockRoleServiceClient) GetUserRoles(ctx context.Context, in *pb.GetUserRolesRequest, opts ...grpc.CallOption) (*pb.GetUserRolesResponse, error) {
	return nil, nil
}

func (m *MockRoleServiceClient) ListUsersWithRole(ctx context.Context, in *pb.ListUsersWithRoleRequest, opts ...grpc.CallOption) (*pb.ListUsersWithRoleResponse, error) {
	return nil, nil
}

// MockPermissionServiceClient is a mock implementation of PermissionServiceClient
type MockPermissionServiceClient struct {
	mock.Mock
}

func (m *MockPermissionServiceClient) AssignPermissionToGroup(ctx context.Context, in *pb.AssignPermissionToGroupRequest, opts ...grpc.CallOption) (*pb.AssignPermissionToGroupResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AssignPermissionToGroupResponse), args.Error(1)
}

func (m *MockPermissionServiceClient) CheckGroupPermission(ctx context.Context, in *pb.CheckGroupPermissionRequest, opts ...grpc.CallOption) (*pb.CheckGroupPermissionResponse, error) {
	return nil, nil
}

func (m *MockPermissionServiceClient) ListGroupPermissions(ctx context.Context, in *pb.ListGroupPermissionsRequest, opts ...grpc.CallOption) (*pb.ListGroupPermissionsResponse, error) {
	return nil, nil
}

func (m *MockPermissionServiceClient) RemovePermissionFromGroup(ctx context.Context, in *pb.RemovePermissionFromGroupRequest, opts ...grpc.CallOption) (*pb.RemovePermissionFromGroupResponse, error) {
	return nil, nil
}

func (m *MockPermissionServiceClient) GetUserEffectivePermissions(ctx context.Context, in *pb.GetUserEffectivePermissionsRequest, opts ...grpc.CallOption) (*pb.GetUserEffectivePermissionsResponse, error) {
	return nil, nil
}

// MockCatalogServiceClient is a mock implementation of CatalogServiceClient
type MockCatalogServiceClient struct {
	mock.Mock
}

// MockTokenServiceClient is a mock implementation of TokenServiceClient
type MockTokenServiceClient struct {
	mock.Mock
}

func (m *MockCatalogServiceClient) SeedRolesAndPermissions(ctx context.Context, in *pb.SeedRolesAndPermissionsRequest, opts ...grpc.CallOption) (*pb.SeedRolesAndPermissionsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SeedRolesAndPermissionsResponse), args.Error(1)
}

func (m *MockCatalogServiceClient) RegisterAction(ctx context.Context, in *pb.RegisterActionRequest, opts ...grpc.CallOption) (*pb.RegisterActionResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) ListActions(ctx context.Context, in *pb.ListActionsRequest, opts ...grpc.CallOption) (*pb.ListActionsResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) RegisterResource(ctx context.Context, in *pb.RegisterResourceRequest, opts ...grpc.CallOption) (*pb.RegisterResourceResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) SetResourceParent(ctx context.Context, in *pb.SetResourceParentRequest, opts ...grpc.CallOption) (*pb.SetResourceParentResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) ListResources(ctx context.Context, in *pb.ListResourcesRequest, opts ...grpc.CallOption) (*pb.ListResourcesResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) CreateRole(ctx context.Context, in *pb.CreateRoleRequest, opts ...grpc.CallOption) (*pb.CreateRoleResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) ListRoles(ctx context.Context, in *pb.ListRolesRequest, opts ...grpc.CallOption) (*pb.ListRolesResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) CreatePermission(ctx context.Context, in *pb.CreatePermissionRequest, opts ...grpc.CallOption) (*pb.CreatePermissionResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) AttachPermissions(ctx context.Context, in *pb.AttachPermissionsRequest, opts ...grpc.CallOption) (*pb.AttachPermissionsResponse, error) {
	return nil, nil
}

func (m *MockCatalogServiceClient) ListPermissions(ctx context.Context, in *pb.ListPermissionsRequest, opts ...grpc.CallOption) (*pb.ListPermissionsResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) ValidateToken(ctx context.Context, in *pb.ValidateTokenRequest, opts ...grpc.CallOption) (*pb.ValidateTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ValidateTokenResponse), args.Error(1)
}

func (m *MockTokenServiceClient) RefreshAccessToken(ctx context.Context, in *pb.RefreshAccessTokenRequest, opts ...grpc.CallOption) (*pb.RefreshAccessTokenResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) RevokeToken(ctx context.Context, in *pb.RevokeTokenRequest, opts ...grpc.CallOption) (*pb.RevokeTokenResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) IntrospectToken(ctx context.Context, in *pb.IntrospectTokenRequest, opts ...grpc.CallOption) (*pb.IntrospectTokenResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) CreateToken(ctx context.Context, in *pb.CreateTokenRequest, opts ...grpc.CallOption) (*pb.CreateTokenResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) ListActiveTokens(ctx context.Context, in *pb.ListActiveTokensRequest, opts ...grpc.CallOption) (*pb.ListActiveTokensResponse, error) {
	return nil, nil
}

func (m *MockTokenServiceClient) BlacklistToken(ctx context.Context, in *pb.BlacklistTokenRequest, opts ...grpc.CallOption) (*pb.BlacklistTokenResponse, error) {
	return nil, nil
}

func (m *MockAuthorizationServiceClient) BatchCheck(ctx context.Context, in *pb.BatchCheckRequest, opts ...grpc.CallOption) (*pb.BatchCheckResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.BatchCheckResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) LookupResources(ctx context.Context, in *pb.LookupResourcesRequest, opts ...grpc.CallOption) (*pb.LookupResourcesResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.LookupResourcesResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) CheckColumns(ctx context.Context, in *pb.CheckColumnsRequest, opts ...grpc.CallOption) (*pb.CheckColumnsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.CheckColumnsResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) ListAllowedColumns(ctx context.Context, in *pb.ListAllowedColumnsRequest, opts ...grpc.CallOption) (*pb.ListAllowedColumnsResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ListAllowedColumnsResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) BulkEvaluatePermissions(ctx context.Context, in *pb.BulkEvaluatePermissionsRequest, opts ...grpc.CallOption) (*pb.BulkEvaluatePermissionsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.BulkEvaluatePermissionsResponse), args.Error(1)
}

func (m *MockAuthorizationServiceClient) EvaluatePermission(ctx context.Context, in *pb.EvaluatePermissionRequest, opts ...grpc.CallOption) (*pb.EvaluatePermissionResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.EvaluatePermissionResponse), args.Error(1)
}

// Helper function to create a test client with mocks
func createTestClient() (*Client, *MockUserServiceClient, *MockAuthorizationServiceClient, *MockOrganizationServiceClient, *MockGroupServiceClient, *MockRoleServiceClient, *MockPermissionServiceClient, *MockCatalogServiceClient, *MockTokenServiceClient) {
	mockUserClient := &MockUserServiceClient{}
	mockAuthzClient := &MockAuthorizationServiceClient{}
	mockOrgClient := &MockOrganizationServiceClient{}
	mockGroupClient := &MockGroupServiceClient{}
	mockRoleClient := &MockRoleServiceClient{}
	mockPermClient := &MockPermissionServiceClient{}
	mockCatalogClient := &MockCatalogServiceClient{}
	mockTokenClient := &MockTokenServiceClient{}

	client := &Client{
		conn:          nil, // Not needed for unit tests
		config:        &config.Config{},
		userClient:    mockUserClient,
		authzClient:   mockAuthzClient,
		orgClient:     mockOrgClient,
		groupClient:   mockGroupClient,
		roleClient:    mockRoleClient,
		permClient:    mockPermClient,
		catalogClient: mockCatalogClient,
		tokenClient:   mockTokenClient,
	}

	return client, mockUserClient, mockAuthzClient, mockOrgClient, mockGroupClient, mockRoleClient, mockPermClient, mockCatalogClient, mockTokenClient
}

func TestCreateUser_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserRequest{
		Username:    "testuser",
		PhoneNumber: "+919876543210",
		Email:       "test@example.com",
		Password:    "password123",
		FullName:    "Test User",
	}

	expectedResponse := &pb.RegisterResponse{
		StatusCode: 201,
		Message:    "User created successfully",
		User: &pb.User{
			Id:       "user123",
			Username: "testuser",
			Status:   "active",
		},
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*pb.RegisterRequest")).Return(expectedResponse, nil)

	response, err := client.CreateUser(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "user123", response.UserID)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "active", response.Status)
	mockUserClient.AssertExpectations(t)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserRequest{
		Username:    "existinguser",
		PhoneNumber: "+919876543210",
		Email:       "existing@example.com",
		Password:    "password123",
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*pb.RegisterRequest")).Return(
		(*pb.RegisterResponse)(nil), status.Error(codes.AlreadyExists, "user already exists"))

	response, err := client.CreateUser(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "user already exists")
	mockUserClient.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()
	userID := "user123"

	expectedResponse := &pb.GetUserResponse{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &pb.User{
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

	mockUserClient.On("GetUser", ctx, &pb.GetUserRequest{Id: userID}).Return(expectedResponse, nil)

	userData, err := client.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, userData)
	assert.Equal(t, "user123", userData.ID)
	assert.Equal(t, "testuser", userData.Username)
	assert.Equal(t, "+919876543210", userData.PhoneNumber)
	mockUserClient.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()
	userID := "nonexistent"

	mockUserClient.On("GetUser", ctx, &pb.GetUserRequest{Id: userID}).Return(
		(*pb.GetUserResponse)(nil), status.Error(codes.NotFound, "user not found"))

	userData, err := client.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, userData)
	assert.Contains(t, err.Error(), "user not found")
	mockUserClient.AssertExpectations(t)
}

func TestGetUserByPhone_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()
	phoneNumber := "+919876543210"

	expectedResponse := &pb.GetUserResponse{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &pb.User{
			Id:          "user123",
			Username:    "testuser",
			PhoneNumber: phoneNumber,
			Email:       "test@example.com",
			Status:      "active",
		},
	}

	expectedRequest := &pb.GetUserByPhoneRequest{
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
	client, _, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	userData, err := client.GetUserByPhone(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, userData)
	assert.Contains(t, err.Error(), "phone number is required")
}

func TestCheckPermission_Success(t *testing.T) {
	client, _, mockAuthzClient, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "create"
	object := "farm456"
	orgID := "org789"

	expectedRequest := &pb.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   object,
		Action:       action,
	}

	expectedResponse := &pb.CheckResponse{
		Allowed: true,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestCheckPermission_Denied(t *testing.T) {
	client, _, mockAuthzClient, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "delete"
	object := "farm456"
	orgID := "org789"

	expectedRequest := &pb.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   object,
		Action:       action,
	}

	expectedResponse := &pb.CheckResponse{
		Allowed: false,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.False(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestCheckPermission_MissingParameters(t *testing.T) {
	client, _, _, _, _, _, _, _, _ := createTestClient()
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
	client, _, mockAuthzClient, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	subject := "user123"
	resource := "farm"
	action := "list"
	object := "" // Empty object should become wildcard
	orgID := "org789"

	expectedRequest := &pb.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   "*", // Should be converted to wildcard
		Action:       action,
	}

	expectedResponse := &pb.CheckResponse{
		Allowed: true,
	}

	mockAuthzClient.On("Check", mock.Anything, expectedRequest).Return(expectedResponse, nil)

	allowed, err := client.CheckPermission(ctx, subject, resource, action, object, orgID)

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockAuthzClient.AssertExpectations(t)
}

func TestHealthCheck_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &pb.GetUserRequest{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return NotFound error, which indicates service is healthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*pb.GetUserResponse)(nil),
		status.Error(codes.NotFound, "user not found"),
	)

	err := client.HealthCheck(ctx)

	assert.NoError(t, err)
	mockUserClient.AssertExpectations(t)
}

func TestHealthCheck_PermissionDeniedIsHealthy(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &pb.GetUserRequest{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return PermissionDenied error, which indicates service is healthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*pb.GetUserResponse)(nil),
		status.Error(codes.PermissionDenied, "permission denied"),
	)

	err := client.HealthCheck(ctx)

	assert.NoError(t, err) // Permission denied means service is healthy
	mockUserClient.AssertExpectations(t)
}

func TestHealthCheck_ServiceUnavailable(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	expectedRequest := &pb.GetUserRequest{
		Id: "health-check-user-id",
	}

	// Mock GetUser to return Unavailable error, which indicates service is unhealthy
	mockUserClient.On("GetUser", mock.Anything, expectedRequest).Return(
		(*pb.GetUserResponse)(nil),
		status.Error(codes.Unavailable, "service unavailable"),
	)

	err := client.HealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AAA service health check failed")
	mockUserClient.AssertExpectations(t)
}

func TestAddRequestMetadata(t *testing.T) {
	client, _, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	requestID := "req123"
	userID := "user456"

	newCtx := client.AddRequestMetadata(ctx, requestID, userID)

	assert.NotEqual(t, ctx, newCtx)
	// Note: In a real test, you would extract and verify the metadata
	// This requires access to grpc metadata package functionality
}

func TestValidateToken_InvalidToken(t *testing.T) {
	client, _, _, _, _, _, _, _, mockTokenClient := createTestClient()
	ctx := context.Background()

	// Mock token validation failure
	mockTokenClient.On("ValidateToken", mock.Anything, mock.AnythingOfType("*pb.ValidateTokenRequest")).Return(
		(*pb.ValidateTokenResponse)(nil),
		status.Error(codes.InvalidArgument, "failed to parse token"),
	)

	// Test with invalid token
	tokenData, err := client.ValidateToken(ctx, "invalid-token")

	assert.Error(t, err)
	assert.Nil(t, tokenData)
	assert.Contains(t, err.Error(), "failed to parse token")
	mockTokenClient.AssertExpectations(t)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	client, _, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	tokenData, err := client.ValidateToken(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, tokenData)
	assert.Contains(t, err.Error(), "token is required")
}

func TestCreateOrganization_NotImplemented(t *testing.T) {
	client, _, _, mockOrgClient, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateOrganizationRequest{
		Name:        "Test FPO",
		Description: "Test FPO Description",
		Type:        "FPO",
		CEOUserID:   "user123",
	}

	expectedResponse := &pb.CreateOrganizationResponse{
		StatusCode: 201,
		Message:    "Organization created successfully",
		Organization: &pb.Organization{
			Id:     "org_123",
			Name:   "Test FPO",
			Type:   "FPO",
			Status: "active",
		},
	}

	mockOrgClient.On("CreateOrganization", ctx, mock.AnythingOfType("*pb.CreateOrganizationRequest")).Return(expectedResponse, nil)

	response, err := client.CreateOrganization(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Test FPO", response.Name)
	assert.Equal(t, "active", response.Status)
	assert.Equal(t, "org_123", response.OrgID)
	mockOrgClient.AssertExpectations(t)
}

func TestCreateUserGroup_NotImplemented(t *testing.T) {
	client, _, _, _, mockGroupClient, _, _, _, _ := createTestClient()
	ctx := context.Background()

	req := &CreateUserGroupRequest{
		Name:        "Directors",
		Description: "FPO Directors Group",
		OrgID:       "org123",
		Permissions: []string{"farm.create", "farm.update"},
	}

	expectedResponse := &pb.CreateGroupResponse{
		StatusCode: 201,
		Message:    "Group created successfully",
		Group: &pb.Group{
			Id:             "grp_123",
			Name:           "Directors",
			Description:    "FPO Directors Group",
			OrganizationId: "org123",
		},
	}

	mockGroupClient.On("CreateGroup", ctx, mock.AnythingOfType("*pb.CreateGroupRequest")).Return(expectedResponse, nil)

	response, err := client.CreateUserGroup(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "Directors", response.Name)
	assert.Equal(t, "org123", response.OrgID)
	assert.Equal(t, "grp_123", response.GroupID)
	mockGroupClient.AssertExpectations(t)
}

func TestAssignRole_NotImplemented(t *testing.T) {
	client, _, _, _, _, mockRoleClient, _, _, _ := createTestClient()
	ctx := context.Background()

	expectedResponse := &pb.AssignRoleResponse{
		StatusCode: 200,
		Message:    "Role assigned successfully",
	}

	mockRoleClient.On("AssignRole", ctx, mock.AnythingOfType("*pb.AssignRoleRequest")).Return(expectedResponse, nil)

	err := client.AssignRole(ctx, "user123", "org456", "farmer")

	assert.NoError(t, err)
	mockRoleClient.AssertExpectations(t)
}

func TestCheckUserRole_NotImplemented(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	// Mock GetUser to return a user without the farmer role
	mockUserClient.On("GetUser", mock.Anything, mock.Anything).Return(
		&pb.GetUserResponse{
			User: &pb.User{
				Id:        "user123",
				Username:  "testuser",
				UserRoles: []*pb.UserRole{}, // No roles
			},
		},
		nil,
	)

	hasRole, err := client.CheckUserRole(ctx, "user123", "farmer")

	// User doesn't have the role
	assert.NoError(t, err)
	assert.False(t, hasRole)
	mockUserClient.AssertExpectations(t)
}

func TestSeedRolesAndPermissions_NotImplemented(t *testing.T) {
	client, _, _, _, _, _, _, mockCatalogClient, _ := createTestClient()
	ctx := context.Background()

	expectedResponse := &pb.SeedRolesAndPermissionsResponse{
		StatusCode:         200,
		Message:            "Roles and permissions seeded successfully",
		RolesCreated:       5,
		PermissionsCreated: 20,
	}

	mockCatalogClient.On("SeedRolesAndPermissions", ctx, mock.AnythingOfType("*pb.SeedRolesAndPermissionsRequest")).Return(expectedResponse, nil)

	err := client.SeedRolesAndPermissions(ctx)

	assert.NoError(t, err)
	mockCatalogClient.AssertExpectations(t)
}

// Test backward compatibility methods
func TestCreateUserLegacy_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()

	username := "testuser"
	mobileNumber := "+919876543210"
	password := "password123"
	countryCode := "+91"
	aadhaarNumber := "123456789012"

	expectedResponse := &pb.RegisterResponse{
		StatusCode: 201,
		Message:    "User created successfully",
		User: &pb.User{
			Id:       "user123",
			Username: username,
			Status:   "active",
		},
	}

	mockUserClient.On("Register", ctx, mock.AnythingOfType("*pb.RegisterRequest")).Return(expectedResponse, nil)

	userID, err := client.CreateUserLegacy(ctx, username, mobileNumber, password, countryCode, &aadhaarNumber)

	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	mockUserClient.AssertExpectations(t)
}

func TestGetUserByMobile_Success(t *testing.T) {
	client, mockUserClient, _, _, _, _, _, _, _ := createTestClient()
	ctx := context.Background()
	mobileNumber := "+919876543210"

	expectedResponse := &pb.GetUserResponse{
		StatusCode: 200,
		Message:    "User retrieved successfully",
		User: &pb.User{
			Id:          "user123",
			Username:    "testuser",
			PhoneNumber: mobileNumber,
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	expectedRequest := &pb.GetUserByPhoneRequest{
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
