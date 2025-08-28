package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReportingService_ExportFarmerPortfolio(t *testing.T) {
	tests := []struct {
		name           string
		request        *requests.ExportFarmerPortfolioRequest
		setupMocks     func(*MockReportingService)
		expectedError  string
		expectedResult bool
	}{
		{
			name: "successful farmer portfolio export",
			request: &requests.ExportFarmerPortfolioRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
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
						Farms: []responses.FarmSummary{
							{
								FarmID: "farm123",
								Name:   "Test Farm",
								AreaHa: 2.5,
							},
						},
						Cycles: []responses.CycleSummary{
							{
								CycleID: "cycle123",
								FarmID:  "farm123",
								Season:  "RABI",
								Status:  "ACTIVE",
							},
						},
						Activities: []responses.ActivitySummary{
							{
								ActivityID:   "activity123",
								CycleID:      "cycle123",
								ActivityType: "PLANTING",
								Status:       "COMPLETED",
							},
						},
						Summary: responses.PortfolioSummary{
							TotalFarms:          1,
							TotalAreaHa:         2.5,
							TotalCycles:         1,
							ActiveCycles:        1,
							CompletedCycles:     0,
							TotalActivities:     1,
							CompletedActivities: 1,
						},
					},
				}
				mockService.On("ExportFarmerPortfolio", mock.Anything, mock.Anything).Return(mockResponse, nil)
			},
			expectedResult: true,
		},
		{
			name: "invalid request type",
			request: &requests.ExportFarmerPortfolioRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				FarmerID: "farmer123",
			},
			setupMocks: func(mockService *MockReportingService) {
				// Pass invalid request type to trigger error
				mockService.On("ExportFarmerPortfolio", mock.Anything, "invalid").Return(nil, fmt.Errorf("invalid request type"))
			},
			expectedError: "invalid request type",
		},
		{
			name: "permission denied",
			request: &requests.ExportFarmerPortfolioRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
				FarmerID: "farmer123",
			},
			setupMocks: func(mockService *MockReportingService) {
				mockService.On("ExportFarmerPortfolio", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("insufficient permissions to read farmer portfolio"))
			},
			expectedError: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			var result interface{}
			var err error

			if tt.name == "invalid request type" {
				result, err = mockService.ExportFarmerPortfolio(context.Background(), "invalid")
			} else {
				result, err = mockService.ExportFarmerPortfolio(context.Background(), tt.request)
			}

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectedResult {
					response, ok := result.(*responses.ExportFarmerPortfolioResponse)
					assert.True(t, ok)
					assert.True(t, response.Success)
					assert.Equal(t, "farmer123", response.Data.FarmerID)
					assert.Equal(t, 1, response.Data.Summary.TotalFarms)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReportingService_OrgDashboardCounters(t *testing.T) {
	tests := []struct {
		name           string
		request        *requests.OrgDashboardCountersRequest
		setupMocks     func(*MockReportingService)
		expectedError  string
		expectedResult bool
	}{
		{
			name: "successful org dashboard counters",
			request: &requests.OrgDashboardCountersRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
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
							TotalFarmers:        10,
							ActiveFarmers:       8,
							TotalFarms:          15,
							TotalAreaHa:         50.5,
							TotalCycles:         20,
							ActiveCycles:        12,
							CompletedCycles:     8,
							TotalActivities:     100,
							CompletedActivities: 75,
						},
						SeasonalBreakdown: []responses.SeasonalCounters{
							{
								Season:     "RABI",
								Cycles:     12,
								AreaHa:     30.0,
								Activities: 60,
							},
							{
								Season:     "KHARIF",
								Cycles:     8,
								AreaHa:     20.5,
								Activities: 40,
							},
						},
						CycleStatusBreakdown: []responses.StatusCounters{
							{Status: "ACTIVE", Count: 12},
							{Status: "COMPLETED", Count: 8},
						},
						ActivityStatusBreakdown: []responses.StatusCounters{
							{Status: "COMPLETED", Count: 75},
							{Status: "PLANNED", Count: 25},
						},
						GeneratedAt: time.Now(),
					},
				}
				mockService.On("OrgDashboardCounters", mock.Anything, mock.Anything).Return(mockResponse, nil)
			},
			expectedResult: true,
		},
		{
			name: "invalid request type",
			request: &requests.OrgDashboardCountersRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
			},
			setupMocks: func(mockService *MockReportingService) {
				mockService.On("OrgDashboardCounters", mock.Anything, "invalid").Return(nil, fmt.Errorf("invalid request type"))
			},
			expectedError: "invalid request type",
		},
		{
			name: "permission denied",
			request: &requests.OrgDashboardCountersRequest{
				BaseRequest: requests.BaseRequest{
					UserID: "user123",
					OrgID:  "org123",
				},
			},
			setupMocks: func(mockService *MockReportingService) {
				mockService.On("OrgDashboardCounters", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("insufficient permissions to read organization dashboard"))
			},
			expectedError: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockReportingService{}
			tt.setupMocks(mockService)

			var result interface{}
			var err error

			if tt.name == "invalid request type" {
				result, err = mockService.OrgDashboardCounters(context.Background(), "invalid")
			} else {
				result, err = mockService.OrgDashboardCounters(context.Background(), tt.request)
			}

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectedResult {
					response, ok := result.(*responses.OrgDashboardCountersResponse)
					assert.True(t, ok)
					assert.True(t, response.Success)
					assert.Equal(t, "org123", response.Data.OrgID)
					assert.Equal(t, 10, response.Data.Counters.TotalFarmers)
					assert.Equal(t, 15, response.Data.Counters.TotalFarms)
					assert.Equal(t, 50.5, response.Data.Counters.TotalAreaHa)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestReportingService_DateFiltering(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	request := &requests.ExportFarmerPortfolioRequest{
		BaseRequest: requests.BaseRequest{
			UserID: "user123",
			OrgID:  "org123",
		},
		FarmerID:  "farmer123",
		StartDate: &startDate,
		EndDate:   &endDate,
		Season:    "RABI",
	}

	mockService := &MockReportingService{}
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

	mockService.On("ExportFarmerPortfolio", mock.Anything, mock.MatchedBy(func(req *requests.ExportFarmerPortfolioRequest) bool {
		return req.StartDate != nil && req.EndDate != nil && req.Season == "RABI"
	})).Return(mockResponse, nil)

	result, err := mockService.ExportFarmerPortfolio(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*responses.ExportFarmerPortfolioResponse)
	assert.True(t, ok)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestReportingService_EmptyResults(t *testing.T) {
	request := &requests.ExportFarmerPortfolioRequest{
		BaseRequest: requests.BaseRequest{
			UserID: "user123",
			OrgID:  "org123",
		},
		FarmerID: "farmer123",
	}

	mockService := &MockReportingService{}
	mockResponse := &responses.ExportFarmerPortfolioResponse{
		BaseResponse: responses.BaseResponse{
			Success: true,
			Message: "Farmer portfolio exported successfully",
		},
		Data: responses.FarmerPortfolioData{
			FarmerID:   "farmer123",
			FarmerName: "John Doe",
			OrgID:      "org123",
			Farms:      []responses.FarmSummary{},
			Cycles:     []responses.CycleSummary{},
			Activities: []responses.ActivitySummary{},
			Summary: responses.PortfolioSummary{
				TotalFarms:          0,
				TotalAreaHa:         0,
				TotalCycles:         0,
				ActiveCycles:        0,
				CompletedCycles:     0,
				TotalActivities:     0,
				CompletedActivities: 0,
			},
		},
	}

	mockService.On("ExportFarmerPortfolio", mock.Anything, mock.Anything).Return(mockResponse, nil)

	result, err := mockService.ExportFarmerPortfolio(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(*responses.ExportFarmerPortfolioResponse)
	assert.True(t, ok)
	assert.True(t, response.Success)
	assert.Equal(t, 0, response.Data.Summary.TotalFarms)
	assert.Equal(t, 0, len(response.Data.Farms))

	mockService.AssertExpectations(t)
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
