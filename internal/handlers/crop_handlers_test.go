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
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCropCycleService is a mock implementation of CropCycleService
type MockCropCycleService struct {
	mock.Mock
}

func (m *MockCropCycleService) StartCycle(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockCropCycleService) UpdateCycle(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockCropCycleService) EndCycle(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockCropCycleService) ListCycles(ctx context.Context, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockCropCycleService) GetCropCycle(ctx context.Context, cycleID string) (interface{}, error) {
	args := m.Called(ctx, cycleID)
	return args.Get(0), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user context
	router.Use(func(c *gin.Context) {
		c.Set("aaa_subject", "test-user-123")
		c.Set("aaa_org", "test-org-123")
		c.Set("request_id", "test-request-123")
		c.Next()
	})

	return router
}

func TestStartCycle_Handler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockCropCycleService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful cycle start",
			requestBody: requests.StartCycleRequest{
				FarmID:       "farm123",
				Season:       "RABI",
				StartDate:    time.Now(),
				PlannedCrops: []string{"wheat", "barley"},
			},
			setupMock: func(mockService *MockCropCycleService) {
				cycleData := &responses.CropCycleData{
					ID:           "cycle123",
					FarmID:       "farm123",
					FarmerID:     "test-user-123",
					Season:       "RABI",
					Status:       "PLANNED",
					PlannedCrops: []string{"wheat", "barley"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockService.On("StartCycle", mock.Anything, mock.AnythingOfType("*requests.StartCycleRequest")).Return(
					responses.NewCropCycleResponse(cycleData, "Crop cycle started successfully"), nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "invalid request body",
			requestBody: requests.StartCycleRequest{
				FarmID: "", // Invalid empty farm_id
				Season: "RABI",
			},
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return validation error
				mockService.On("StartCycle", mock.Anything, mock.AnythingOfType("*requests.StartCycleRequest")).Return(
					nil, common.ErrInvalidCropCycleData)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "invalid season",
			requestBody: requests.StartCycleRequest{
				FarmID:       "farm123",
				Season:       "INVALID_SEASON",
				StartDate:    time.Now(),
				PlannedCrops: []string{"wheat"},
			},
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return validation error
				mockService.On("StartCycle", mock.Anything, mock.AnythingOfType("*requests.StartCycleRequest")).Return(
					nil, common.ErrInvalidCropCycleData)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCropCycleService{}
			tt.setupMock(mockService)

			router := setupTestRouter()
			router.POST("/crops/cycles", StartCycle(mockService))

			// Prepare request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/crops/cycles", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response responses.CropCycleResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, "farm123", response.Data.FarmID)
				assert.Equal(t, "RABI", response.Data.Season)
				assert.Equal(t, "PLANNED", response.Data.Status)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateCycle_Handler(t *testing.T) {
	tests := []struct {
		name           string
		cycleID        string
		requestBody    interface{}
		setupMock      func(*MockCropCycleService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "successful cycle update",
			cycleID: "cycle123",
			requestBody: requests.UpdateCycleRequest{
				Season:       stringPtr("KHARIF"),
				PlannedCrops: []string{"rice", "sugarcane"},
			},
			setupMock: func(mockService *MockCropCycleService) {
				cycleData := &responses.CropCycleData{
					ID:           "cycle123",
					FarmID:       "farm123",
					FarmerID:     "test-user-123",
					Season:       "KHARIF",
					Status:       "PLANNED",
					PlannedCrops: []string{"rice", "sugarcane"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockService.On("UpdateCycle", mock.Anything, mock.AnythingOfType("*requests.UpdateCycleRequest")).Return(
					responses.NewCropCycleResponse(cycleData, "Crop cycle updated successfully"), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "invalid season in update",
			cycleID: "cycle123",
			requestBody: requests.UpdateCycleRequest{
				Season: stringPtr("INVALID_SEASON"),
			},
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return validation error
				mockService.On("UpdateCycle", mock.Anything, mock.AnythingOfType("*requests.UpdateCycleRequest")).Return(
					nil, common.ErrInvalidCropCycleData)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCropCycleService{}
			tt.setupMock(mockService)

			router := setupTestRouter()
			router.PUT("/crops/cycles/:cycle_id", UpdateCycle(mockService))

			// Prepare request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/crops/cycles/"+tt.cycleID, bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response responses.CropCycleResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.cycleID, response.Data.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestEndCycle_Handler(t *testing.T) {
	tests := []struct {
		name           string
		cycleID        string
		requestBody    interface{}
		setupMock      func(*MockCropCycleService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "successful cycle end",
			cycleID: "cycle123",
			requestBody: requests.EndCycleRequest{
				Status:  "COMPLETED",
				EndDate: time.Now(),
				Outcome: map[string]string{"yield": "good", "quality": "high"},
			},
			setupMock: func(mockService *MockCropCycleService) {
				cycleData := &responses.CropCycleData{
					ID:        "cycle123",
					FarmID:    "farm123",
					FarmerID:  "test-user-123",
					Season:    "RABI",
					Status:    "COMPLETED",
					EndDate:   timePtrCrop(time.Now()),
					Outcome:   map[string]string{"yield": "good", "quality": "high"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockService.On("EndCycle", mock.Anything, mock.AnythingOfType("*requests.EndCycleRequest")).Return(
					responses.NewCropCycleResponse(cycleData, "Crop cycle ended successfully"), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "invalid status",
			cycleID: "cycle123",
			requestBody: requests.EndCycleRequest{
				Status:  "INVALID_STATUS",
				EndDate: time.Now(),
			},
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return validation error
				mockService.On("EndCycle", mock.Anything, mock.AnythingOfType("*requests.EndCycleRequest")).Return(
					nil, common.ErrInvalidCropCycleData)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:    "missing required fields",
			cycleID: "cycle123",
			requestBody: requests.EndCycleRequest{
				Status: "COMPLETED",
				// Missing end_date - will be zero value
			},
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return validation error
				mockService.On("EndCycle", mock.Anything, mock.AnythingOfType("*requests.EndCycleRequest")).Return(
					nil, common.ErrInvalidCropCycleData)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCropCycleService{}
			tt.setupMock(mockService)

			router := setupTestRouter()
			router.POST("/crops/cycles/:cycle_id/end", EndCycle(mockService))

			// Prepare request
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/crops/cycles/"+tt.cycleID+"/end", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response responses.CropCycleResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.cycleID, response.Data.ID)
				assert.Equal(t, "COMPLETED", response.Data.Status)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListCycles_Handler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*MockCropCycleService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "successful list cycles",
			queryParams: "?page=1&page_size=10&farm_id=farm123&season=RABI",
			setupMock: func(mockService *MockCropCycleService) {
				cycles := []*responses.CropCycleData{
					{
						ID:        "cycle1",
						FarmID:    "farm123",
						FarmerID:  "test-user-123",
						Season:    "RABI",
						Status:    "ACTIVE",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					{
						ID:        "cycle2",
						FarmID:    "farm123",
						FarmerID:  "test-user-123",
						Season:    "RABI",
						Status:    "PLANNED",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				mockService.On("ListCycles", mock.Anything, mock.AnythingOfType("*requests.ListCyclesRequest")).Return(
					responses.NewCropCycleListResponse(cycles, 1, 10, 2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "list cycles with no filters",
			queryParams: "",
			setupMock: func(mockService *MockCropCycleService) {
				mockService.On("ListCycles", mock.Anything, mock.AnythingOfType("*requests.ListCyclesRequest")).Return(
					responses.NewCropCycleListResponse([]*responses.CropCycleData{}, 1, 10, 0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCropCycleService{}
			tt.setupMock(mockService)

			router := setupTestRouter()
			router.GET("/crops/cycles", ListCycles(mockService))

			// Prepare request
			req, _ := http.NewRequest("GET", "/crops/cycles"+tt.queryParams, nil)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response responses.CropCycleListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetCropCycle_Handler(t *testing.T) {
	tests := []struct {
		name           string
		cycleID        string
		setupMock      func(*MockCropCycleService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "successful get cycle",
			cycleID: "cycle123",
			setupMock: func(mockService *MockCropCycleService) {
				cycleData := &responses.CropCycleData{
					ID:        "cycle123",
					FarmID:    "farm123",
					FarmerID:  "test-user-123",
					Season:    "RABI",
					Status:    "ACTIVE",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mockService.On("GetCropCycle", mock.Anything, "cycle123").Return(
					responses.NewCropCycleResponse(cycleData, "Crop cycle retrieved successfully"), nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "cycle not found",
			cycleID: "nonexistent",
			setupMock: func(mockService *MockCropCycleService) {
				// Mock will be called but should return not found error
				mockService.On("GetCropCycle", mock.Anything, "nonexistent").Return(
					nil, common.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCropCycleService{}
			tt.setupMock(mockService)

			router := setupTestRouter()
			router.GET("/crops/cycles/:cycle_id", GetCropCycle(mockService))

			// Prepare request
			url := "/crops/cycles/" + tt.cycleID
			req, _ := http.NewRequest("GET", url, nil)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response responses.CropCycleResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Data)
				assert.Equal(t, tt.cycleID, response.Data.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Helper functions
func timePtrCrop(t time.Time) *time.Time {
	return &t
}
