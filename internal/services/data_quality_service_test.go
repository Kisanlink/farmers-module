package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataQualityService_ValidateGeometry(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}
	mockNotificationService := &MockNotificationService{}

	// Setup AAA service mock
	mockAAAService.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	// Create service with nil repositories for testing
	service := NewDataQualityService(nil, nil, nil, mockAAAService, mockNotificationService)

	req := &requests.ValidateGeometryRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    "user123",
			OrgID:     "org123",
			RequestID: "req123",
		},
		WKT:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		CheckBounds: false,
	}

	// Setup context with user information
	ctx := context.Background()
	userCtx := &auth.UserContext{
		AAAUserID: req.UserID,
		Username:  "testuser",
		Roles:     []string{"admin"},
	}
	ctx = auth.SetUserInContext(ctx, userCtx)

	// Call the method
	response, err := service.ValidateGeometry(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}
