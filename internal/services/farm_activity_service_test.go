package services

import (
	"context"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestFarmActivityService_CreateActivity_BusinessLogic tests the business logic for creating activities
func TestFarmActivityService_CreateActivity_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAService)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAService) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.CreateActivityRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				CropCycleID:  "cycle123",
				ActivityType: "planting",
			},
			setupMocks: func(aaa *MockAAAService) {
				aaa.On("CheckPermission", mock.Anything, "user123", "activity", "create", "cycle123", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAAService := &MockAAAService{}
			tt.setupMocks(mockAAAService)

			// Create a minimal service for business logic testing
			service := &FarmActivityServiceImpl{
				aaaService: mockAAAService,
			}

			result, err := service.CreateActivity(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, result)

			mockAAAService.AssertExpectations(t)
		})
	}
}

// TestFarmActivityService_CompleteActivity_BusinessLogic tests the business logic for completing activities
func TestFarmActivityService_CompleteActivity_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAService)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAService) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.CompleteActivityRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				ID:          "activity123",
				CompletedAt: time.Now(),
			},
			setupMocks: func(aaa *MockAAAService) {
				aaa.On("CheckPermission", mock.Anything, "user123", "activity", "complete", "activity123", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAAService := &MockAAAService{}
			tt.setupMocks(mockAAAService)

			service := &FarmActivityServiceImpl{
				aaaService: mockAAAService,
			}

			result, err := service.CompleteActivity(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, result)

			mockAAAService.AssertExpectations(t)
		})
	}
}

// TestFarmActivityService_UpdateActivity_BusinessLogic tests the business logic for updating activities
func TestFarmActivityService_UpdateActivity_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAService)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAService) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.UpdateActivityRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				ID: "activity123",
			},
			setupMocks: func(aaa *MockAAAService) {
				aaa.On("CheckPermission", mock.Anything, "user123", "activity", "update", "activity123", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAAService := &MockAAAService{}
			tt.setupMocks(mockAAAService)

			service := &FarmActivityServiceImpl{
				aaaService: mockAAAService,
			}

			result, err := service.UpdateActivity(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, result)

			mockAAAService.AssertExpectations(t)
		})
	}
}

// TestFarmActivityService_ListActivities_BusinessLogic tests the business logic for listing activities
func TestFarmActivityService_ListActivities_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		setupMocks    func(*MockAAAService)
		expectedError error
	}{
		{
			name:    "invalid request type",
			request: "invalid request",
			setupMocks: func(aaa *MockAAAService) {
				// No setup needed
			},
			expectedError: common.ErrInvalidInput,
		},
		{
			name: "permission denied",
			request: &requests.ListActivitiesRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				Page:     1,
				PageSize: 10,
			},
			setupMocks: func(aaa *MockAAAService) {
				aaa.On("CheckPermission", mock.Anything, "user123", "activity", "list", "", "org123").Return(false, nil)
			},
			expectedError: common.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAAAService := &MockAAAService{}
			tt.setupMocks(mockAAAService)

			service := &FarmActivityServiceImpl{
				aaaService: mockAAAService,
			}

			result, err := service.ListActivities(context.Background(), tt.request)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, result)

			mockAAAService.AssertExpectations(t)
		})
	}
}

// TestFarmActivityService_DateFiltering tests date filtering logic
func TestFarmActivityService_DateFiltering(t *testing.T) {
	mockAAAService := &MockAAAService{}
	mockAAAService.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	service := &FarmActivityServiceImpl{
		aaaService: mockAAAService,
	}

	t.Run("invalid date format", func(t *testing.T) {
		req := &requests.ListActivitiesRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "user123",
				OrgID:  "org123",
			},
			DateFrom: "invalid-date",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListActivities(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date_from format")
		assert.Nil(t, result)
	})

	t.Run("invalid date_to format", func(t *testing.T) {
		req := &requests.ListActivitiesRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "user123",
				OrgID:  "org123",
			},
			DateFrom: "2024-01-01",
			DateTo:   "invalid-date",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListActivities(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date_to format")
		assert.Nil(t, result)
	})

	mockAAAService.AssertExpectations(t)
}
