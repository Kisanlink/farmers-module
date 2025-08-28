package services

import (
	"context"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataQualityService_ValidateGeometry(t *testing.T) {
	// Create mocks
	mockAAAService := &MockAAAService{}

	// Setup AAA service mock
	mockAAAService.On("CheckPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	// Create service with nil repositories for testing
	service := NewDataQualityService(nil, nil, nil, mockAAAService)

	req := &requests.ValidateGeometryRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    "user123",
			OrgID:     "org123",
			RequestID: "req123",
		},
		WKT:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
		CheckBounds: false,
	}

	// Call the method
	response, err := service.ValidateGeometry(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Verify mocks
	mockAAAService.AssertExpectations(t)
}
