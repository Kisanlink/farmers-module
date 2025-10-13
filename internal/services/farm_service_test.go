package services

import (
	"context"
	"testing"

	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestFarmService_ValidateGeometry(t *testing.T) {
	tests := []struct {
		name        string
		wkt         string
		expectError bool
		description string
	}{
		{
			name:        "valid polygon",
			wkt:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			expectError: false,
			description: "Simple square polygon",
		},
		{
			name:        "empty geometry",
			wkt:         "",
			expectError: true,
			description: "Empty WKT string",
		},
		{
			name:        "non-polygon geometry",
			wkt:         "POINT(0 0)",
			expectError: true,
			description: "Point geometry instead of polygon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &FarmServiceImpl{
				db: nil, // No database for unit tests
			}

			// Call the method
			err := service.validateGeometry(context.Background(), tt.wkt)

			// Assertions
			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.description)
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.description)
			}
		})
	}
}

func TestFarmService_ValidateCreateFarmRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     *requests.CreateFarmRequest
		expectError bool
		description string
	}{
		{
			name: "valid request",
			request: &requests.CreateFarmRequest{
				AAAUserID: "farmer123",
				AAAOrgID:  "org123",
				Geometry: requests.GeometryData{
					WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				},
			},
			expectError: false,
			description: "Valid create farm request",
		},
		{
			name: "missing farmer ID",
			request: &requests.CreateFarmRequest{
				AAAOrgID: "org123",
				Geometry: requests.GeometryData{
					WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				},
			},
			expectError: true,
			description: "Missing farmer user ID",
		},
		{
			name: "missing org ID",
			request: &requests.CreateFarmRequest{
				AAAUserID: "farmer123",
				Geometry: requests.GeometryData{
					WKT: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				},
			},
			expectError: true,
			description: "Missing organization ID",
		},
		{
			name: "missing geometry",
			request: &requests.CreateFarmRequest{
				AAAUserID: "farmer123",
				AAAOrgID:  "org123",
			},
			expectError: true,
			description: "Missing geometry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &FarmServiceImpl{}

			// Call the validation method
			err := service.validateCreateFarmRequest(tt.request)

			// Assertions
			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.description)
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.description)
			}
		})
	}
}

func TestFarmService_FilterFarmsByArea(t *testing.T) {
	// This tests the area filtering logic without database dependencies
	service := &FarmServiceImpl{}

	// Create test farms
	farms := []*farmEntity.Farm{
		{Name: testutils.StringPtr("Small Farm"), AreaHa: 0.5},
		{Name: testutils.StringPtr("Medium Farm"), AreaHa: 2.0},
		{Name: testutils.StringPtr("Large Farm"), AreaHa: 10.0},
	}

	tests := []struct {
		name          string
		minArea       *float64
		maxArea       *float64
		expectedCount int
		description   string
	}{
		{
			name:          "no filters",
			minArea:       nil,
			maxArea:       nil,
			expectedCount: 3,
			description:   "All farms should be returned",
		},
		{
			name:          "min area filter",
			minArea:       func() *float64 { v := 1.0; return &v }(),
			maxArea:       nil,
			expectedCount: 2,
			description:   "Only medium and large farms",
		},
		{
			name:          "max area filter",
			minArea:       nil,
			maxArea:       func() *float64 { v := 5.0; return &v }(),
			expectedCount: 2,
			description:   "Only small and medium farms",
		},
		{
			name:          "both filters",
			minArea:       func() *float64 { v := 1.0; return &v }(),
			maxArea:       func() *float64 { v := 5.0; return &v }(),
			expectedCount: 1,
			description:   "Only medium farm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.filterFarmsByArea(farms, tt.minArea, tt.maxArea)
			assert.Equal(t, tt.expectedCount, len(result), tt.description)
		})
	}
}
