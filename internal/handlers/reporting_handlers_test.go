package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReportingHandlers_ExportFarmerPortfolio(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockReportingService)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful farmer portfolio export",
			requestBody: requests.ExportFarmerPortfolioRequest{
				FarmerID: "farmer123",
				Season:   "RABI",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockResponse := &responses.ExportFarmerPortfolioResponse{
					BaseResponse: responses.BaseResponse{
						Success: true,
						Message: "Farmer portfolio exported successfully",
					},
					Data: responses.FarmerPortfolioData{
						FarmerID:   "farmer123",
						FarmerName: "John Doe",
						OrgID:      "org123",
						Summary: responses.PortfolioSummary{
							TotalFarms:  1,
							TotalAreaHa: 2.5,
						},
					},
				}
				mockService.On("ExportFarmerPortfolio", mock.Anything, mock.Anything).Return(mockResponse, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
				c.Set("request_id", "req123")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for invalid JSON
			},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing farmer_id",
			requestBody: requests.ExportFarmerPortfolioRequest{
				Season: "RABI",
			},
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for validation error
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "farmer_id is required",
		},
		{
			name: "service error",
			requestBody: requests.ExportFarmerPortfolioRequest{
				FarmerID: "farmer123",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockService.On("ExportFarmerPortfolio", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			handler := NewReportingHandlers(mockService)

			// Create request
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/v1/reports/farmer-portfolio", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Setup context
			tt.setupContext(c)

			// Call handler
			handler.ExportFarmerPortfolio(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response responses.ExportFarmerPortfolioResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "farmer123", response.Data.FarmerID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReportingHandlers_OrgDashboardCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockReportingService)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful org dashboard counters",
			requestBody: requests.OrgDashboardCountersRequest{
				Season: "RABI",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockResponse := &responses.OrgDashboardCountersResponse{
					BaseResponse: responses.BaseResponse{
						Success: true,
						Message: "Organization dashboard counters retrieved successfully",
					},
					Data: responses.OrgDashboardData{
						OrgID: "org123",
						Counters: responses.OrgCounters{
							TotalFarmers: 10,
							TotalFarms:   15,
						},
						GeneratedAt: time.Now(),
					},
				}
				mockService.On("OrgDashboardCounters", mock.Anything, mock.Anything).Return(mockResponse, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
				c.Set("request_id", "req123")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for invalid JSON
			},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "service error",
			requestBody: requests.OrgDashboardCountersRequest{
				Season: "RABI",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockService.On("OrgDashboardCounters", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			handler := NewReportingHandlers(mockService)

			// Create request
			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/v1/reports/org-dashboard", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Setup context
			tt.setupContext(c)

			// Call handler
			handler.OrgDashboardCounters(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response responses.OrgDashboardCountersResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "org123", response.Data.OrgID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReportingHandlers_ExportFarmerPortfolioByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		farmerID       string
		queryParams    map[string]string
		setupMocks     func(*MockReportingService)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful farmer portfolio export by ID",
			farmerID: "farmer123",
			queryParams: map[string]string{
				"season": "RABI",
				"format": "json",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockResponse := &responses.ExportFarmerPortfolioResponse{
					BaseResponse: responses.BaseResponse{
						Success: true,
						Message: "Farmer portfolio exported successfully",
					},
					Data: responses.FarmerPortfolioData{
						FarmerID: "farmer123",
						OrgID:    "org123",
					},
				}
				mockService.On("ExportFarmerPortfolio", mock.Anything, mock.MatchedBy(func(req *requests.ExportFarmerPortfolioRequest) bool {
					return req.FarmerID == "farmer123" && req.Season == "RABI"
				})).Return(mockResponse, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Set("aaa_org", "org123")
				c.Params = gin.Params{{Key: "farmer_id", Value: "farmer123"}}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing farmer_id",
			farmerID: "",
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for validation error
			},
			setupContext: func(c *gin.Context) {
				c.Params = gin.Params{{Key: "farmer_id", Value: ""}}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "farmer_id is required",
		},
		{
			name:     "invalid date format",
			farmerID: "farmer123",
			queryParams: map[string]string{
				"start_date": "invalid-date",
			},
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for validation error
			},
			setupContext: func(c *gin.Context) {
				c.Params = gin.Params{{Key: "farmer_id", Value: "farmer123"}}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid start_date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			handler := NewReportingHandlers(mockService)

			// Create request with query parameters
			req, err := http.NewRequest("GET", "/api/v1/reports/farmer-portfolio/"+tt.farmerID, nil)
			assert.NoError(t, err)

			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Setup context
			tt.setupContext(c)

			// Call handler
			handler.ExportFarmerPortfolioByID(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response responses.ExportFarmerPortfolioResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReportingHandlers_OrgDashboardCountersByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orgID          string
		queryParams    map[string]string
		setupMocks     func(*MockReportingService)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "successful org dashboard counters by ID",
			orgID: "org123",
			queryParams: map[string]string{
				"season": "RABI",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockResponse := &responses.OrgDashboardCountersResponse{
					BaseResponse: responses.BaseResponse{
						Success: true,
						Message: "Organization dashboard counters retrieved successfully",
					},
					Data: responses.OrgDashboardData{
						OrgID: "org123",
						Counters: responses.OrgCounters{
							TotalFarmers: 10,
						},
						GeneratedAt: time.Now(),
					},
				}
				mockService.On("OrgDashboardCounters", mock.Anything, mock.MatchedBy(func(req *requests.OrgDashboardCountersRequest) bool {
					return req.OrgID == "org123" && req.Season == "RABI"
				})).Return(mockResponse, nil)
			},
			setupContext: func(c *gin.Context) {
				c.Set("aaa_subject", "user123")
				c.Params = gin.Params{{Key: "org_id", Value: "org123"}}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "missing org_id",
			orgID: "",
			setupMocks: func(mockService *MockReportingService) {
				// No mock setup needed for validation error
			},
			setupContext: func(c *gin.Context) {
				c.Params = gin.Params{{Key: "org_id", Value: ""}}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "org_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			handler := NewReportingHandlers(mockService)

			// Create request with query parameters
			req, err := http.NewRequest("GET", "/api/v1/reports/org-dashboard/"+tt.orgID, nil)
			assert.NoError(t, err)

			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Create response recorder
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Setup context
			tt.setupContext(c)

			// Call handler
			handler.OrgDashboardCountersByID(c)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response responses.OrgDashboardCountersResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// MockReportingService is a mock implementation of ReportingService for testing
type MockReportingService struct {
	mock.Mock
}

func (m *MockReportingService) ExportFarmerPortfolio(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockReportingService) OrgDashboardCounters(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}
