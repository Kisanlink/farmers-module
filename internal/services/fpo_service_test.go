package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFPORefRepository is a mock implementation of the FPO reference repository
type MockFPORefRepository struct {
	mock.Mock
}

func (m *MockFPORefRepository) Create(ctx context.Context, entity *fpo.FPORef) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockFPORefRepository) FindOne(ctx context.Context, filter *base.Filter) (*fpo.FPORef, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fpo.FPORef), args.Error(1)
}

// MockAAAService is a mock implementation of the AAA service
type MockAAAService struct {
	mock.Mock
}

func (m *MockAAAService) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByMobile(ctx context.Context, mobile string) (interface{}, error) {
	args := m.Called(ctx, mobile)
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

func (m *MockAAAService) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	args := m.Called(ctx, userID, orgID, roleName)
	return args.Error(0)
}

func (m *MockAAAService) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	args := m.Called(ctx, groupID, resource, action)
	return args.Error(0)
}

// Implement other required methods for AAAService interface
func (m *MockAAAService) SeedRolesAndPermissions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAAAService) CheckPermission(ctx context.Context, req interface{}) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) GetUser(ctx context.Context, userID string) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	args := m.Called(ctx, email)
	return args.Get(0), args.Error(1)
}

func (m *MockAAAService) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	args := m.Called(ctx, userID, groupID)
	return args.Error(0)
}

func (m *MockAAAService) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	args := m.Called(ctx, userID, roleName)
	return args.Bool(0), args.Error(1)
}

func (m *MockAAAService) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAAAService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestFPOService_CreateFPO_Success(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	// Test data
	ctx := context.Background()
	req := &requests.CreateFPORequest{
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

	// Mock AAA service calls
	mockAAA.On("GetUserByMobile", ctx, "+919876543210").Return(nil, errors.New("user not found"))

	mockAAA.On("CreateUser", ctx, mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]interface{}{
			"id":         "user123",
			"username":   "John_Doe",
			"status":     "active",
			"created_at": time.Now(),
		}, nil)

	mockAAA.On("CreateOrganization", ctx, mock.AnythingOfType("map[string]interface {}")).Return(
		map[string]interface{}{
			"org_id":     "org123",
			"name":       "Test FPO",
			"status":     "active",
			"created_at": time.Now(),
		}, nil)

	mockAAA.On("AssignRole", ctx, "user123", "org123", "CEO").Return(nil)

	// Mock user group creation
	groupNames := []string{"directors", "shareholders", "store_staff", "store_managers"}
	for _, groupName := range groupNames {
		mockAAA.On("CreateUserGroup", ctx, mock.MatchedBy(func(req map[string]interface{}) bool {
			return req["name"].(string) == groupName
		})).Return(
			map[string]interface{}{
				"group_id":   "group_" + groupName,
				"name":       groupName,
				"org_id":     "org123",
				"created_at": time.Now().Format(time.RFC3339),
			}, nil)

		// Mock permission assignments based on actual group permissions
		var permissions []string
		switch groupName {
		case "directors":
			permissions = []string{"manage", "read", "write", "approve"}
		case "shareholders":
			permissions = []string{"read", "vote"}
		case "store_staff":
			permissions = []string{"read", "write", "inventory"}
		case "store_managers":
			permissions = []string{"read", "write", "manage", "inventory", "reports"}
		default:
			permissions = []string{"read"}
		}
		for _, permission := range permissions {
			mockAAA.On("AssignPermissionToGroup", ctx, "group_"+groupName, "fpo", permission).Return(nil)
		}
	}

	// Mock repository call
	mockRepo.On("Create", ctx, mock.AnythingOfType("*fpo.FPORef")).Return(nil)

	// Execute
	result, err := service.CreateFPO(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	fpoData, ok := result.(*responses.CreateFPOData)
	assert.True(t, ok)
	assert.Equal(t, "Test FPO", fpoData.Name)
	assert.Equal(t, "org123", fpoData.AAAOrgID)
	assert.Equal(t, "user123", fpoData.CEOUserID)
	assert.Equal(t, "ACTIVE", fpoData.Status)
	assert.Len(t, fpoData.UserGroups, 4)

	// Verify mocks
	mockAAA.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestFPOService_CreateFPO_ValidationErrors(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	ctx := context.Background()

	// Test cases for validation errors
	testCases := []struct {
		name        string
		request     *requests.CreateFPORequest
		expectedErr string
	}{
		{
			name: "Missing FPO name",
			request: &requests.CreateFPORequest{
				RegistrationNo: "FPO123456",
				CEOUser: requests.CEOUserData{
					FirstName:   "John",
					LastName:    "Doe",
					PhoneNumber: "+919876543210",
				},
			},
			expectedErr: "FPO name is required",
		},
		{
			name: "Missing registration number",
			request: &requests.CreateFPORequest{
				Name: "Test FPO",
				CEOUser: requests.CEOUserData{
					FirstName:   "John",
					LastName:    "Doe",
					PhoneNumber: "+919876543210",
				},
			},
			expectedErr: "FPO registration number is required",
		},
		{
			name: "Missing CEO first name",
			request: &requests.CreateFPORequest{
				Name:           "Test FPO",
				RegistrationNo: "FPO123456",
				CEOUser: requests.CEOUserData{
					LastName:    "Doe",
					PhoneNumber: "+919876543210",
				},
			},
			expectedErr: "CEO user details are required",
		},
		{
			name: "Missing CEO phone number",
			request: &requests.CreateFPORequest{
				Name:           "Test FPO",
				RegistrationNo: "FPO123456",
				CEOUser: requests.CEOUserData{
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			expectedErr: "CEO phone number is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CreateFPO(ctx, tc.request)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestFPOService_RegisterFPORef_Success(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	// Test data
	ctx := context.Background()
	req := &requests.RegisterFPORefRequest{
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		BusinessConfig: map[string]string{"type": "agricultural"},
		Metadata:       map[string]string{"region": "north"},
	}

	// Mock AAA service calls
	mockAAA.On("GetOrganization", ctx, "org123").Return(
		map[string]interface{}{
			"id":     "org123",
			"name":   "Test FPO",
			"status": "active",
		}, nil)

	// Mock repository calls
	mockRepo.On("FindOne", ctx, mock.AnythingOfType("*base.Filter")).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*fpo.FPORef")).Return(nil)

	// Execute
	result, err := service.RegisterFPORef(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	fpoRefData, ok := result.(*responses.FPORefData)
	assert.True(t, ok)
	assert.Equal(t, "org123", fpoRefData.AAAOrgID)
	assert.Equal(t, "Test FPO", fpoRefData.Name)
	assert.Equal(t, "FPO123456", fpoRefData.RegistrationNo)
	assert.Equal(t, "ACTIVE", fpoRefData.Status)

	// Verify mocks
	mockAAA.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestFPOService_RegisterFPORef_AlreadyExists(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	// Test data
	ctx := context.Background()
	req := &requests.RegisterFPORefRequest{
		AAAOrgID: "org123",
		Name:     "Test FPO",
	}

	// Mock AAA service calls
	mockAAA.On("GetOrganization", ctx, "org123").Return(
		map[string]interface{}{
			"id":     "org123",
			"name":   "Test FPO",
			"status": "active",
		}, nil)

	// Mock existing FPO reference
	existingFPO := &fpo.FPORef{
		AAAOrgID: "org123",
		Name:     "Existing FPO",
		Status:   "ACTIVE",
	}
	mockRepo.On("FindOne", ctx, mock.AnythingOfType("*base.Filter")).Return(existingFPO, nil)

	// Execute
	result, err := service.RegisterFPORef(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "FPO reference already exists for organization ID: org123")

	// Verify mocks
	mockAAA.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestFPOService_GetFPORef_Success(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	// Test data
	ctx := context.Background()
	orgID := "org123"

	fpoRef := &fpo.FPORef{
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "FPO123456",
		Status:         "ACTIVE",
		BusinessConfig: map[string]string{"type": "agricultural"},
	}
	fpoRef.ID = "fpo_ref_123"
	fpoRef.CreatedAt = time.Now()
	fpoRef.UpdatedAt = time.Now()

	// Mock repository call
	mockRepo.On("FindOne", ctx, mock.AnythingOfType("*base.Filter")).Return(fpoRef, nil)

	// Execute
	result, err := service.GetFPORef(ctx, orgID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	fpoRefData, ok := result.(*responses.FPORefData)
	assert.True(t, ok)
	assert.Equal(t, "fpo_ref_123", fpoRefData.ID)
	assert.Equal(t, "org123", fpoRefData.AAAOrgID)
	assert.Equal(t, "Test FPO", fpoRefData.Name)
	assert.Equal(t, "FPO123456", fpoRefData.RegistrationNo)
	assert.Equal(t, "ACTIVE", fpoRefData.Status)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestFPOService_GetFPORef_NotFound(t *testing.T) {
	// Setup mocks
	mockRepo := &MockFPORefRepository{}
	mockAAA := &MockAAAService{}

	// Create service
	service := NewFPOService(mockRepo, mockAAA)

	// Test data
	ctx := context.Background()
	orgID := "org123"

	// Mock repository call
	mockRepo.On("FindOne", ctx, mock.AnythingOfType("*base.Filter")).Return(nil, errors.New("not found"))

	// Execute
	result, err := service.GetFPORef(ctx, orgID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get FPO reference")

	// Verify mocks
	mockRepo.AssertExpectations(t)
}
