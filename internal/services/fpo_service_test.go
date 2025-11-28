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
		BusinessConfig: map[string]interface{}{"type": "agricultural"},
		Metadata:       map[string]interface{}{"region": "north"},
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

	// Mock CEO role check (Business Rule 1.2: user not already CEO of another FPO)
	mockAAA.On("CheckUserRole", ctx, "user123", "CEO").Return(false, nil)

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
	ctx := context.Background()

	// Test cases for early validation errors (before AAA call)
	t.Run("Missing FPO name", func(t *testing.T) {
		mockRepo := &MockFPORefRepository{}
		mockAAA := &MockAAAService{}
		service := NewFPOService(mockRepo, mockAAA)

		result, err := service.CreateFPO(ctx, &requests.CreateFPORequest{
			RegistrationNo: "FPO123456",
			CEOUser: requests.CEOUserData{
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "+919876543210",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "FPO name is required")
	})

	t.Run("Missing registration number", func(t *testing.T) {
		mockRepo := &MockFPORefRepository{}
		mockAAA := &MockAAAService{}
		service := NewFPOService(mockRepo, mockAAA)

		result, err := service.CreateFPO(ctx, &requests.CreateFPORequest{
			Name: "Test FPO",
			CEOUser: requests.CEOUserData{
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "+919876543210",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "FPO registration number is required")
	})

	t.Run("Missing CEO phone number", func(t *testing.T) {
		mockRepo := &MockFPORefRepository{}
		mockAAA := &MockAAAService{}
		service := NewFPOService(mockRepo, mockAAA)

		result, err := service.CreateFPO(ctx, &requests.CreateFPORequest{
			Name:           "Test FPO",
			RegistrationNo: "FPO123456",
			CEOUser: requests.CEOUserData{
				FirstName: "John",
				LastName:  "Doe",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "CEO phone number is required")
	})

	// Test case for validation after AAA lookup (user not found)
	t.Run("Missing CEO name when creating new user", func(t *testing.T) {
		mockRepo := &MockFPORefRepository{}
		mockAAA := &MockAAAService{}
		service := NewFPOService(mockRepo, mockAAA)

		// Mock AAA returns user not found
		mockAAA.On("GetUserByMobile", ctx, "+919876543210").Return(nil, errors.New("user not found"))

		result, err := service.CreateFPO(ctx, &requests.CreateFPORequest{
			Name:           "Test FPO",
			RegistrationNo: "FPO123456",
			CEOUser: requests.CEOUserData{
				LastName:    "Doe",
				PhoneNumber: "+919876543210",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "CEO first_name and last_name are required when creating a new user")
	})

	// Test case: existing user found - no need for first_name/last_name
	t.Run("Existing CEO user - no name required", func(t *testing.T) {
		mockRepo := &MockFPORefRepository{}
		mockAAA := &MockAAAService{}
		service := NewFPOService(mockRepo, mockAAA)

		// Mock AAA returns existing user
		mockAAA.On("GetUserByMobile", ctx, "+919876543210").Return(
			map[string]interface{}{
				"id":        "user123",
				"full_name": "Existing User",
				"username":  "existing_user",
				"status":    "active",
			}, nil)

		// Mock CEO role check
		mockAAA.On("CheckUserRole", ctx, "user123", "CEO").Return(false, nil)

		// Mock org creation
		mockAAA.On("CreateOrganization", ctx, mock.AnythingOfType("map[string]interface {}")).Return(
			map[string]interface{}{
				"org_id": "org123",
				"name":   "Test FPO",
				"status": "active",
			}, nil)

		// Mock role assignment
		mockAAA.On("AssignRole", ctx, "user123", "org123", "CEO").Return(nil)

		// Mock user groups
		groupNames := []string{"directors", "shareholders", "store_staff", "store_managers"}
		for _, groupName := range groupNames {
			mockAAA.On("CreateUserGroup", ctx, mock.MatchedBy(func(req map[string]interface{}) bool {
				return req["name"].(string) == groupName
			})).Return(
				map[string]interface{}{
					"group_id":   "group_" + groupName,
					"name":       groupName,
					"org_id":     "org123",
					"created_at": "2025-01-01T00:00:00Z",
				}, nil)

			mockAAA.On("AssignPermissionToGroup", ctx, "group_"+groupName, "fpo", mock.AnythingOfType("string")).Return(nil)
		}

		// Mock repo create
		mockRepo.On("Create", ctx, mock.AnythingOfType("*fpo.FPORef")).Return(nil)

		result, err := service.CreateFPO(ctx, &requests.CreateFPORequest{
			Name:           "Test FPO",
			RegistrationNo: "FPO123456",
			CEOUser: requests.CEOUserData{
				PhoneNumber: "+919876543210",
				// No first_name/last_name - should work since user exists
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
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
		BusinessConfig: map[string]interface{}{"type": "agricultural"},
		Metadata:       map[string]interface{}{"region": "north"},
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
		BusinessConfig: map[string]interface{}{"type": "agricultural"},
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
