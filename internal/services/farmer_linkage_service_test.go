package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFarmerLinkageServiceImpl_LinkFarmerToFPO(t *testing.T) {
	tests := []struct {
		name          string
		request       *requests.LinkFarmerRequest
		setupMocks    func(*MockFarmerLinkageRepoShared, *MockFarmerRepository, *MockAAAServiceShared)
		expectedError string
		shouldSucceed bool
	}{
		{
			name: "successful link farmer to FPO",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User and org exist
				aaa.On("GetUser", mock.Anything, "user123").Return(map[string]interface{}{"id": "user123"}, nil)
				aaa.On("GetOrganization", mock.Anything, "org456").Return(map[string]interface{}{"id": "org456"}, nil)
				// Farmer exists in local database
				farmerRepo.On("FindOne", mock.Anything, mock.Anything).Return(&farmerentity.Farmer{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
				}, nil)
				// No existing link
				repo.On("Find", mock.Anything, mock.Anything).Return([]*farmerentity.FarmerLink{}, nil)
				// Create succeeds
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "invalid request - missing user ID",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// No mocks needed as validation should fail early
			},
			expectedError: "aaa_user_id and aaa_org_id are required",
		},
		{
			name: "permission denied",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// Permission check fails
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
			},
			expectedError: "insufficient permissions to link farmer to FPO",
		},
		{
			name: "user not found in AAA",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User not found
				aaa.On("GetUser", mock.Anything, "user123").Return(nil, errors.New("user not found"))
			},
			expectedError: "farmer not found in AAA service",
		},
		{
			name: "farmer not found in local database",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User and org exist in AAA
				aaa.On("GetUser", mock.Anything, "user123").Return(map[string]interface{}{"id": "user123"}, nil)
				aaa.On("GetOrganization", mock.Anything, "org456").Return(map[string]interface{}{"id": "org456"}, nil)
				// Farmer not found in local database
				farmerRepo.On("FindOne", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedError: "farmer with aaa_user_id=user123 and aaa_org_id=org456 must be created before linking to FPO",
		},
		{
			name: "reactivate existing inactive link",
			request: &requests.LinkFarmerRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, farmerRepo *MockFarmerRepository, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User and org exist
				aaa.On("GetUser", mock.Anything, "user123").Return(map[string]interface{}{"id": "user123"}, nil)
				aaa.On("GetOrganization", mock.Anything, "org456").Return(map[string]interface{}{"id": "org456"}, nil)
				// Farmer exists in local database
				farmerRepo.On("FindOne", mock.Anything, mock.Anything).Return(&farmerentity.Farmer{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
				}, nil)
				// Existing inactive link
				existingLink := &farmerentity.FarmerLink{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
					Status:    "INACTIVE",
				}
				repo.On("Find", mock.Anything, mock.Anything).Return([]*farmerentity.FarmerLink{existingLink}, nil)
				// Update succeeds
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFarmerLinkageRepoShared{}
			mockFarmerRepo := &MockFarmerRepository{}
			mockAAA := &MockAAAServiceShared{}

			tt.setupMocks(mockRepo, mockFarmerRepo, mockAAA)

			service := &FarmerLinkageServiceImpl{
				farmerLinkageRepo: mockRepo,
				farmerRepo:        mockFarmerRepo,
				aaaService:        mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			userCtx := &auth.UserContext{
				AAAUserID: tt.request.UserID,
				Username:  "testadmin",
				Roles:     []string{"admin"},
			}
			ctx = auth.SetUserInContext(ctx, userCtx)

			err := service.LinkFarmerToFPO(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
			mockFarmerRepo.AssertExpectations(t)
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestFarmerLinkageServiceImpl_AssignKisanSathi(t *testing.T) {
	tests := []struct {
		name          string
		request       *requests.AssignKisanSathiRequest
		setupMocks    func(*MockFarmerLinkageRepoShared, *MockAAAServiceShared)
		expectedError string
		shouldSucceed bool
	}{
		{
			name: "successful assign KisanSathi",
			request: &requests.AssignKisanSathiRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// KisanSathi user exists
				aaa.On("GetUser", mock.Anything, "ks789").Return(map[string]interface{}{"id": "ks789"}, nil)
				// Role check passes
				aaa.On("CheckUserRole", mock.Anything, "ks789", "KisanSathi").Return(true, nil)
				// Existing active farmer link
				existingLink := &farmerentity.FarmerLink{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
					Status:    "ACTIVE",
				}
				repo.On("Find", mock.Anything, mock.Anything).Return([]*farmerentity.FarmerLink{existingLink}, nil)
				// Update succeeds
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "invalid request - missing KisanSathi user ID",
			request: &requests.AssignKisanSathiRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// No mocks needed as validation should fail early
			},
			expectedError: "aaa_user_id, aaa_org_id, and kisan_sathi_user_id are required",
		},
		{
			name: "KisanSathi user needs role assignment",
			request: &requests.AssignKisanSathiRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// KisanSathi user exists
				aaa.On("GetUser", mock.Anything, "ks789").Return(map[string]interface{}{"id": "ks789"}, nil)
				// Role check fails initially
				aaa.On("CheckUserRole", mock.Anything, "ks789", "KisanSathi").Return(false, nil).Once()
				// Role assignment succeeds
				aaa.On("AssignRole", mock.Anything, "ks789", "org456", "KisanSathi").Return(nil)
				// Role check passes after assignment
				aaa.On("CheckUserRole", mock.Anything, "ks789", "KisanSathi").Return(true, nil).Once()
				// Existing active farmer link
				existingLink := &farmerentity.FarmerLink{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
					Status:    "ACTIVE",
				}
				repo.On("Find", mock.Anything, mock.Anything).Return([]*farmerentity.FarmerLink{existingLink}, nil)
				// Update succeeds
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "cannot assign to inactive farmer link",
			request: &requests.AssignKisanSathiRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// KisanSathi user exists
				aaa.On("GetUser", mock.Anything, "ks789").Return(map[string]interface{}{"id": "ks789"}, nil)
				// Role check passes
				aaa.On("CheckUserRole", mock.Anything, "ks789", "KisanSathi").Return(true, nil)
				// Existing inactive farmer link
				existingLink := &farmerentity.FarmerLink{
					AAAUserID: "user123",
					AAAOrgID:  "org456",
					Status:    "INACTIVE",
				}
				repo.On("Find", mock.Anything, mock.Anything).Return([]*farmerentity.FarmerLink{existingLink}, nil)
			},
			expectedError: "cannot assign KisanSathi to inactive farmer link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFarmerLinkageRepoShared{}
			mockAAA := &MockAAAServiceShared{}

			tt.setupMocks(mockRepo, mockAAA)

			service := &FarmerLinkageServiceImpl{
				farmerLinkageRepo: mockRepo,
				aaaService:        mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			userCtx := &auth.UserContext{
				AAAUserID: tt.request.UserID,
				Username:  "testadmin",
				Roles:     []string{"admin"},
			}
			ctx = auth.SetUserInContext(ctx, userCtx)

			result, err := service.AssignKisanSathi(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify response structure
				assignmentData, ok := result.(*responses.KisanSathiAssignmentData)
				assert.True(t, ok)
				assert.Equal(t, tt.request.AAAUserID, assignmentData.AAAUserID)
				assert.Equal(t, tt.request.AAAOrgID, assignmentData.AAAOrgID)
				assert.Equal(t, &tt.request.KisanSathiUserID, assignmentData.KisanSathiUserID)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestFarmerLinkageServiceImpl_CreateKisanSathiUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *requests.CreateKisanSathiUserRequest
		setupMocks    func(*MockFarmerLinkageRepoShared, *MockAAAServiceShared)
		expectedError string
		shouldSucceed bool
	}{
		{
			name: "successful create new KisanSathi user",
			request: &requests.CreateKisanSathiUserRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				Username:    "kisansathi1",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User doesn't exist by phone
				aaa.On("GetUserByMobile", mock.Anything, "+919876543210").Return(nil, errors.New("user not found"))
				// User doesn't exist by email
				aaa.On("GetUserByEmail", mock.Anything, "ks1@example.com").Return(nil, errors.New("user not found"))
				// User creation succeeds
				userResponse := map[string]interface{}{
					"id":         "new_user_id",
					"status":     "ACTIVE",
					"created_at": time.Now().Format(time.RFC3339),
				}
				aaa.On("CreateUser", mock.Anything, mock.Anything).Return(userResponse, nil)
				// Role assignment succeeds
				aaa.On("CheckUserRole", mock.Anything, "new_user_id", "KisanSathi").Return(false, nil).Once()
				aaa.On("AssignRole", mock.Anything, "new_user_id", "", "KisanSathi").Return(nil)
				aaa.On("CheckUserRole", mock.Anything, "new_user_id", "KisanSathi").Return(true, nil).Once()
			},
			shouldSucceed: true,
		},
		{
			name: "invalid request - missing required fields",
			request: &requests.CreateKisanSathiUserRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				Username:    "",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// No mocks needed as validation should fail early
			},
			expectedError: "username, phone_number, password, and full_name are required",
		},
		{
			name: "user already exists by phone - assign role",
			request: &requests.CreateKisanSathiUserRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "admin123",
					OrgID:  "org456",
				},
				Username:    "kisansathi1",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			setupMocks: func(repo *MockFarmerLinkageRepoShared, aaa *MockAAAServiceShared) {
				// Permission check passes
				aaa.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				// User exists by phone
				existingUser := map[string]interface{}{
					"id":         "existing_user_id",
					"username":   "existing_user",
					"full_name":  "Existing User",
					"status":     "ACTIVE",
					"created_at": time.Now().Format(time.RFC3339),
				}
				aaa.On("GetUserByMobile", mock.Anything, "+919876543210").Return(existingUser, nil)
				// Role assignment succeeds
				aaa.On("CheckUserRole", mock.Anything, "existing_user_id", "KisanSathi").Return(false, nil).Once()
				aaa.On("AssignRole", mock.Anything, "existing_user_id", "", "KisanSathi").Return(nil)
				aaa.On("CheckUserRole", mock.Anything, "existing_user_id", "KisanSathi").Return(true, nil).Once()
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockFarmerLinkageRepoShared{}
			mockAAA := &MockAAAServiceShared{}

			tt.setupMocks(mockRepo, mockAAA)

			service := &FarmerLinkageServiceImpl{
				farmerLinkageRepo: mockRepo,
				aaaService:        mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			userCtx := &auth.UserContext{
				AAAUserID: tt.request.UserID,
				Username:  "testadmin",
				Roles:     []string{"admin"},
			}
			ctx = auth.SetUserInContext(ctx, userCtx)

			result, err := service.CreateKisanSathiUser(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify response structure
				userData, ok := result.(*responses.KisanSathiUserData)
				assert.True(t, ok)
				assert.Equal(t, "KisanSathi", userData.Role)
				assert.NotEmpty(t, userData.ID)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestFarmerLinkageServiceImpl_RequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request interface{}
		valid   bool
	}{
		{
			name: "valid link farmer request",
			request: &requests.LinkFarmerRequest{
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			valid: true,
		},
		{
			name: "valid unlink farmer request",
			request: &requests.UnlinkFarmerRequest{
				AAAUserID: "user123",
				AAAOrgID:  "org456",
			},
			valid: true,
		},
		{
			name: "valid assign kisansathi request",
			request: &requests.AssignKisanSathiRequest{
				AAAUserID:        "user123",
				AAAOrgID:         "org456",
				KisanSathiUserID: "ks789",
			},
			valid: true,
		},
		{
			name: "valid reassign kisansathi request",
			request: &requests.ReassignKisanSathiRequest{
				AAAUserID:           "user123",
				AAAOrgID:            "org456",
				NewKisanSathiUserID: stringPtr("ks999"),
			},
			valid: true,
		},
		{
			name: "valid remove kisansathi request",
			request: &requests.ReassignKisanSathiRequest{
				AAAUserID:           "user123",
				AAAOrgID:            "org456",
				NewKisanSathiUserID: nil,
			},
			valid: true,
		},
		{
			name: "valid create kisansathi user request",
			request: &requests.CreateKisanSathiUserRequest{
				Username:    "kisansathi1",
				PhoneNumber: "+919876543210",
				Email:       "ks1@example.com",
				Password:    "password123",
				FullName:    "KisanSathi One",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check that request structures are properly formed
			assert.NotNil(t, tt.request)

			// Type assertions to verify request types
			switch req := tt.request.(type) {
			case *requests.LinkFarmerRequest:
				assert.NotEmpty(t, req.AAAUserID)
				assert.NotEmpty(t, req.AAAOrgID)
			case *requests.UnlinkFarmerRequest:
				assert.NotEmpty(t, req.AAAUserID)
				assert.NotEmpty(t, req.AAAOrgID)
			case *requests.AssignKisanSathiRequest:
				assert.NotEmpty(t, req.AAAUserID)
				assert.NotEmpty(t, req.AAAOrgID)
				assert.NotEmpty(t, req.KisanSathiUserID)
			case *requests.ReassignKisanSathiRequest:
				assert.NotEmpty(t, req.AAAUserID)
				assert.NotEmpty(t, req.AAAOrgID)
				// NewKisanSathiUserID can be nil for removal
			case *requests.CreateKisanSathiUserRequest:
				assert.NotEmpty(t, req.Username)
				assert.NotEmpty(t, req.PhoneNumber)
				assert.NotEmpty(t, req.Password)
				assert.NotEmpty(t, req.FullName)
			default:
				t.Errorf("Unknown request type: %T", req)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
