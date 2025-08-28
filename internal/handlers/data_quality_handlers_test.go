package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataQualityHandlers_ValidateGeometry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    requests.ValidateGeometryRequest
		mockResponse   *responses.ValidateGeometryResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "Successful validation",
			requestBody: requests.ValidateGeometryRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req123",
				},
				WKT:         "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				CheckBounds: false,
			},
			mockResponse: &responses.ValidateGeometryResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req123",
					Message:   "Geometry validation passed",
				},
				WKT:      "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
				IsValid:  true,
				Errors:   []string{},
				Warnings: []string{},
				SRID:     4326,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid geometry",
			requestBody: requests.ValidateGeometryRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req124",
				},
				WKT:         "INVALID_WKT",
				CheckBounds: false,
			},
			mockResponse: &responses.ValidateGeometryResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req124",
					Message:   "Geometry validation completed",
				},
				WKT:      "INVALID_WKT",
				IsValid:  false,
				Errors:   []string{"only POLYGON geometries are supported"},
				Warnings: []string{},
				SRID:     4326,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &services.MockDataQualityService{}
			mockService.On("ValidateGeometry", mock.Anything, mock.AnythingOfType("*requests.ValidateGeometryRequest")).
				Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewDataQualityHandlers(mockService)

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/data-quality/validate-geometry", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("aaa_subject", "user123")
			c.Set("aaa_org", "org123")
			c.Set("request_id", tt.requestBody.RequestID)

			// Call handler
			handler.ValidateGeometry(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response responses.ValidateGeometryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.IsValid, response.IsValid)
				assert.Equal(t, tt.mockResponse.WKT, response.WKT)
				assert.Equal(t, tt.mockResponse.RequestID, response.RequestID)
			}

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataQualityHandlers_ReconcileAAALinks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    requests.ReconcileAAALinksRequest
		mockResponse   *responses.ReconcileAAALinksResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "Successful reconciliation",
			requestBody: requests.ReconcileAAALinksRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req123",
				},
				DryRun: false,
			},
			mockResponse: &responses.ReconcileAAALinksResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req123",
					Message:   "Reconciliation completed: 5 links processed, 2 fixed, 1 broken",
				},
				ProcessedLinks: 5,
				FixedLinks:     2,
				BrokenLinks:    1,
				Errors:         []string{},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Dry run reconciliation",
			requestBody: requests.ReconcileAAALinksRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req124",
				},
				DryRun: true,
			},
			mockResponse: &responses.ReconcileAAALinksResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req124",
					Message:   "Dry run completed: 5 links processed, 2 would be fixed, 1 broken",
				},
				ProcessedLinks: 5,
				FixedLinks:     2,
				BrokenLinks:    1,
				Errors:         []string{},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &services.MockDataQualityService{}
			mockService.On("ReconcileAAALinks", mock.Anything, mock.AnythingOfType("*requests.ReconcileAAALinksRequest")).
				Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewDataQualityHandlers(mockService)

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/data-quality/reconcile-aaa-links", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("aaa_subject", "admin123")
			c.Set("aaa_org", "org123")
			c.Set("request_id", tt.requestBody.RequestID)

			// Call handler
			handler.ReconcileAAALinks(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response responses.ReconcileAAALinksResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ProcessedLinks, response.ProcessedLinks)
				assert.Equal(t, tt.mockResponse.FixedLinks, response.FixedLinks)
				assert.Equal(t, tt.mockResponse.BrokenLinks, response.BrokenLinks)
				assert.Equal(t, tt.mockResponse.RequestID, response.RequestID)
			}

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataQualityHandlers_RebuildSpatialIndexes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &services.MockDataQualityService{}
	mockResponse := &responses.RebuildSpatialIndexesResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: "req123",
			Message:   "Successfully rebuilt 2 spatial indexes",
		},
		RebuiltIndexes: []string{"idx_farms_geometry"},
		Errors:         []string{},
	}

	mockService.On("RebuildSpatialIndexes", mock.Anything, mock.AnythingOfType("*requests.RebuildSpatialIndexesRequest")).
		Return(mockResponse, nil)

	// Create handler
	handler := NewDataQualityHandlers(mockService)

	// Create request
	requestBody := requests.RebuildSpatialIndexesRequest{
		BaseRequest: requests.BaseRequest{
			RequestID: "req123",
		},
	}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/data-quality/rebuild-spatial-indexes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("aaa_subject", "admin123")
	c.Set("aaa_org", "org123")
	c.Set("request_id", "req123")

	// Call handler
	handler.RebuildSpatialIndexes(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response responses.RebuildSpatialIndexesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, mockResponse.RebuiltIndexes, response.RebuiltIndexes)
	assert.Equal(t, mockResponse.RequestID, response.RequestID)

	// Verify mock
	mockService.AssertExpectations(t)
}

func TestDataQualityHandlers_DetectFarmOverlaps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    requests.DetectFarmOverlapsRequest
		mockResponse   *responses.DetectFarmOverlapsResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "Successful overlap detection",
			requestBody: requests.DetectFarmOverlapsRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req123",
				},
				MinOverlapAreaHa: nil,
				Limit:            nil,
			},
			mockResponse: &responses.DetectFarmOverlapsResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req123",
					Message:   "Detected 2 farm overlaps",
				},
				Overlaps: []responses.FarmOverlap{
					{
						Farm1ID:                "farm1",
						Farm1Name:              "Farm 1",
						Farm1FarmerID:          "farmer1",
						Farm2ID:                "farm2",
						Farm2Name:              "Farm 2",
						Farm2FarmerID:          "farmer2",
						OverlapAreaHa:          0.5,
						Farm1AreaHa:            2.0,
						Farm2AreaHa:            1.5,
						OverlapPercentageFarm1: 25.0,
						OverlapPercentageFarm2: 33.33,
					},
				},
				TotalOverlaps: 1,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "No overlaps detected",
			requestBody: requests.DetectFarmOverlapsRequest{
				BaseRequest: requests.BaseRequest{
					RequestID: "req124",
				},
				MinOverlapAreaHa: func() *float64 { v := 0.1; return &v }(),
				Limit:            func() *int { v := 10; return &v }(),
			},
			mockResponse: &responses.DetectFarmOverlapsResponse{
				BaseResponse: responses.BaseResponse{
					RequestID: "req124",
					Message:   "No farm overlaps detected",
				},
				Overlaps:      []responses.FarmOverlap{},
				TotalOverlaps: 0,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := &services.MockDataQualityService{}
			mockService.On("DetectFarmOverlaps", mock.Anything, mock.AnythingOfType("*requests.DetectFarmOverlapsRequest")).
				Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewDataQualityHandlers(mockService)

			// Create request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/data-quality/detect-farm-overlaps", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("aaa_subject", "user123")
			c.Set("aaa_org", "org123")
			c.Set("request_id", tt.requestBody.RequestID)

			// Call handler
			handler.DetectFarmOverlaps(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response responses.DetectFarmOverlapsResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.TotalOverlaps, response.TotalOverlaps)
				assert.Equal(t, tt.mockResponse.RequestID, response.RequestID)
				assert.Equal(t, len(tt.mockResponse.Overlaps), len(response.Overlaps))
			}

			// Verify mock
			mockService.AssertExpectations(t)
		})
	}
}

func TestDataQualityHandlers_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &services.MockDataQualityService{}
	handler := NewDataQualityHandlers(mockService)

	tests := []struct {
		name        string
		handlerFunc func(*gin.Context)
		endpoint    string
	}{
		{
			name:        "ValidateGeometry with invalid JSON",
			handlerFunc: handler.ValidateGeometry,
			endpoint:    "/api/v1/data-quality/validate-geometry",
		},
		{
			name:        "ReconcileAAALinks with invalid JSON",
			handlerFunc: handler.ReconcileAAALinks,
			endpoint:    "/api/v1/data-quality/reconcile-aaa-links",
		},
		{
			name:        "RebuildSpatialIndexes with invalid JSON",
			handlerFunc: handler.RebuildSpatialIndexes,
			endpoint:    "/api/v1/data-quality/rebuild-spatial-indexes",
		},
		{
			name:        "DetectFarmOverlaps with invalid JSON",
			handlerFunc: handler.DetectFarmOverlaps,
			endpoint:    "/api/v1/data-quality/detect-farm-overlaps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with invalid JSON
			req, _ := http.NewRequest("POST", tt.endpoint, bytes.NewBuffer([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			tt.handlerFunc(c)

			// Assertions
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
