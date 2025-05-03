package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/permission"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ----- Mock Interfaces -----

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) VerifyUserAndType(userID string) (bool, bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Bool(1), args.Error(2)
}

type MockFarmService struct {
	mock.Mock
}

func (m *MockFarmService) CreateFarm(
	farmerID string,
	location models.GeoJSONPolygon,
	area float64,
	locality string,
	pincode int,
	ownerID string,
) (*models.Farm, error) {
	args := m.Called(farmerID, location, area, locality, pincode, ownerID)
	return args.Get(0).(*models.Farm), args.Error(1)
}

func (m *MockFarmService) GetAllFarms(farmerID, pincode, date, id string) ([]*models.Farm, error) {
	args := m.Called(farmerID, pincode, date, id)
	return args.Get(0).([]*models.Farm), args.Error(1)
}

func (m *MockFarmService) GetFarmsWithFilters(farmerID, pincode string) ([]*models.Farm, error) {
	args := m.Called(farmerID, pincode)
	return args.Get(0).([]*models.Farm), args.Error(1)
}

func (m *MockFarmService) GetFarmByID(farmID string) (*models.Farm, error) {
	args := m.Called(farmID)
	return args.Get(0).(*models.Farm), args.Error(1)
}

// ----- Test Function -----

func TestCreateFarmHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"
	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	// Mock UserService.VerifyUserAndType
	mockUserService.On("VerifyUserAndType", actorID).Return(true, false, nil)

	// Patch permission check
	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId, requiredPerm string) (bool, int, string, string) {
		return true, http.StatusOK, "", ""
	})
	defer patchPerm.Unpatch()

	// Mock FarmService.CreateFarm
	mockFarm := &models.Farm{
		Base: models.Base{
			Id: "farm-id",
		},
	}

	mockFarmService.
		On("CreateFarm", reqBody.FarmerId, mock.Anything, reqBody.Area, reqBody.Locality, reqBody.Pincode, reqBody.OwnerId).
		Return(mockFarm, nil)

	// Patch CreateFarmData async call
	patchAsync := monkey.Patch(handlers.CreateFarmData, func(farmId string) {})
	defer patchAsync.Unpatch()

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockFarmService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestCreateFarmHandler_MissingUserIDHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Intentionally NOT setting "user-id" header

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Please include your user ID in headers")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_UserVerificationFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"

	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Simulate error from UserService
	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(false, false, errors.New("db unavailable"))

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Something went wrong on our end")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"

	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Simulate user not found
	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(false, false, nil)

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Your account isn't registered")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"

	invalidJSON := `{"farmer_id": "farmer-id", "location": "not-a-valid-format"}`

	// Mock user exists
	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(true, false, nil)

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid farm details provided")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_InvalidPolygonCoordinates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"

	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}}}, // only 3 points
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(true, false, nil)

	// Patch permission check
	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId, requiredPerm string) (bool, int, string, string) {
		return true, http.StatusOK, "", ""
	})
	defer patchPerm.Unpatch()

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "A polygon requires at least 4 points")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_PermissionDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"

	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(true, false, nil)

	// Patch permission check to deny access
	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId, requiredPerm string) (bool, int, string, string) {
		return false, http.StatusForbidden, "permission_denied", "User does not have permission"
	})
	defer patchPerm.Unpatch()

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "User does not have permission")

	mockUserService.AssertExpectations(t)
	mockFarmService.AssertExpectations(t)
}

func TestCreateFarmHandler_CreateFarmFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"
	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {0, 1}, {1, 1}, {0, 0}}},
		Area:     10.5,
		Locality: "Test Locality",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	mockUserService.
		On("VerifyUserAndType", actorID).
		Return(true, false, nil)

	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId, requiredPerm string) (bool, int, string, string) {
		return true, http.StatusOK, "", ""
	})
	defer patchPerm.Unpatch()

	mockFarmService.
		On("CreateFarm", reqBody.FarmerId, mock.Anything, reqBody.Area, reqBody.Locality, reqBody.Pincode, reqBody.OwnerId).
		Return((*models.Farm)(nil), errors.New("failed to create farm: overlaps with existing"))

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code) // 409 Conflict

	// Check for a substring in the actual message returned to the client, not the internal error log
	assert.Contains(t, rec.Body.String(), "overlaps with an existing farm")

	mockFarmService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestCreateFarmHandler_InsufficientPolygonPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()

	mockUserService := new(MockUserService)
	mockFarmService := new(MockFarmService)
	handler := handlers.NewFarmHandler(mockFarmService, mockUserService)

	actorID := "test-user-id"
	reqBody := handlers.FarmRequest{
		FarmerId: "farmer-id",
		Location: [][][]float64{{{0, 0}, {1, 1}, {2, 2}}}, // Only 3 points
		Area:     10.0,
		Locality: "Invalid Polygon",
		Pincode:  123456,
		OwnerId:  "owner-id",
	}

	bodyBytes, _ := json.Marshal(reqBody)

	mockUserService.On("VerifyUserAndType", actorID).Return(true, false, nil)

	patchPerm := monkey.Patch(permission.CheckUserPermission, func(ctx context.Context, userId, requiredPerm string) (bool, int, string, string) {
		return true, http.StatusOK, "", ""
	})
	defer patchPerm.Unpatch()

	req := httptest.NewRequest(http.MethodPost, "/farm", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", actorID)

	rec := httptest.NewRecorder()
	r := gin.Default()
	r.POST("/farm", handler.CreateFarmHandler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "requires at least 4 points")
	mockUserService.AssertExpectations(t)
}
