package services

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCropCycleService_StartCycle_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAServiceShared)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAServiceShared) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.StartCycleRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				FarmID:    "farm123",
				Season:    "RABI",
				StartDate: time.Now(),
				CropID:    "crop123",
			},
			setupMocks: func(aaa *MockAAAServiceShared) {
				aaa.On("CheckPermission", mock.Anything, "user123", "cycle", "start", "", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAServiceShared{}
			tt.setupMocks(mockAAA)

			// Create service with nil repository (we're only testing business logic)
			service := &CropCycleServiceImpl{
				cropCycleRepo: nil,
				farmService:   nil,
				aaaService:    mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			if req, ok := tt.request.(*requests.StartCycleRequest); ok {
				userCtx := &auth.UserContext{
					AAAUserID: req.UserID,
					Username:  "testuser",
					Roles:     []string{"admin"},
				}
				ctx = auth.SetUserInContext(ctx, userCtx)
			}

			// Call the service method
			result, err := service.StartCycle(ctx, tt.request)

			// Assertions
			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, result)

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestCropCycleService_UpdateCycle_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAServiceShared)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAServiceShared) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.UpdateCycleRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				ID:     "cycle123",
				Season: stringPtrCrop("KHARIF"),
			},
			setupMocks: func(aaa *MockAAAServiceShared) {
				aaa.On("CheckPermission", mock.Anything, "user123", "cycle", "update", "cycle123", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAServiceShared{}
			tt.setupMocks(mockAAA)

			// Create service with nil repository (we're only testing business logic)
			service := &CropCycleServiceImpl{
				cropCycleRepo: nil,
				farmService:   nil,
				aaaService:    mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			if req, ok := tt.request.(*requests.UpdateCycleRequest); ok {
				userCtx := &auth.UserContext{
					AAAUserID: req.UserID,
					Username:  "testuser",
					Roles:     []string{"admin"},
				}
				ctx = auth.SetUserInContext(ctx, userCtx)
			}

			// Call the service method - this will fail at repository access but we can test the early validation
			result, err := service.UpdateCycle(ctx, tt.request)

			// Assertions
			assert.Error(t, err)
			if tt.expectedError == common.ErrInvalidInput || tt.expectedError == common.ErrForbidden {
				assert.Equal(t, tt.expectedError, err)
			}
			assert.Nil(t, result)

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestCropCycleService_EndCycle_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAServiceShared)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAServiceShared) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.EndCycleRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				ID:      "cycle123",
				Status:  "COMPLETED",
				EndDate: time.Now(),
			},
			setupMocks: func(aaa *MockAAAServiceShared) {
				aaa.On("CheckPermission", mock.Anything, "user123", "cycle", "end", "cycle123", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAServiceShared{}
			tt.setupMocks(mockAAA)

			// Create service with nil repository (we're only testing business logic)
			service := &CropCycleServiceImpl{
				cropCycleRepo: nil,
				farmService:   nil,
				aaaService:    mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			if req, ok := tt.request.(*requests.EndCycleRequest); ok {
				userCtx := &auth.UserContext{
					AAAUserID: req.UserID,
					Username:  "testuser",
					Roles:     []string{"admin"},
				}
				ctx = auth.SetUserInContext(ctx, userCtx)
			}

			// Call the service method - this will fail at repository access but we can test the early validation
			result, err := service.EndCycle(ctx, tt.request)

			// Assertions
			assert.Error(t, err)
			if tt.expectedError == common.ErrInvalidInput || tt.expectedError == common.ErrForbidden {
				assert.Equal(t, tt.expectedError, err)
			}
			assert.Nil(t, result)

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

func TestCropCycleService_ListCycles_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAServiceShared)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAServiceShared) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.ListCyclesRequest{
				FilterRequest: requests.FilterRequest{
					PaginationRequest: requests.PaginationRequest{
						BaseRequest: requests.BaseRequest{
							UserID: "user123",
							OrgID:  "org123",
						},
						Page:     1,
						PageSize: 10,
					},
				},
			},
			setupMocks: func(aaa *MockAAAServiceShared) {
				aaa.On("CheckPermission", mock.Anything, "user123", "cycle", "list", "", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockAAA := &MockAAAServiceShared{}
			tt.setupMocks(mockAAA)

			// Create service with nil repository (we're only testing business logic)
			service := &CropCycleServiceImpl{
				cropCycleRepo: nil,
				farmService:   nil,
				aaaService:    mockAAA,
			}

			// Setup context with user information
			ctx := context.Background()
			if req, ok := tt.request.(*requests.ListCyclesRequest); ok {
				userCtx := &auth.UserContext{
					AAAUserID: req.UserID,
					Username:  "testuser",
					Roles:     []string{"admin"},
				}
				ctx = auth.SetUserInContext(ctx, userCtx)
			}

			// Call the service method - this will fail at repository access but we can test the early validation
			result, err := service.ListCycles(ctx, tt.request)

			// Assertions
			assert.Error(t, err)
			if tt.expectedError == common.ErrInvalidInput || tt.expectedError == common.ErrForbidden {
				assert.Equal(t, tt.expectedError, err)
			}
			assert.Nil(t, result)

			// Verify mock expectations
			mockAAA.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer for crop cycle tests
func stringPtrCrop(s string) *string {
	return &s
}
