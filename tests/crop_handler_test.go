package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/models"
)

type MockCropService struct {
	mock.Mock
}

func (m *MockCropService) CreateCrop(crop *models.Crop) error {
	args := m.Called(crop)
	return args.Error(0)
}

func (m *MockCropService) GetAllCrops(name string, page, size int) ([]*models.Crop, int64, error) {
	args := m.Called(name, page, size)
	return args.Get(0).([]*models.Crop), args.Get(1).(int64), args.Error(2)
}

func (m *MockCropService) GetCropById(id string) (*models.Crop, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Crop), args.Error(1)
}

func (m *MockCropService) UpdateCrop(crop *models.Crop) error {
	args := m.Called(crop)
	return args.Error(0)
}

func (m *MockCropService) DeleteCrop(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupCropRouter(handler *handlers.CropHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/crops", handler.CreateCrop)
	r.GET("/crops", handler.GetAllCrops)
	r.GET("/crops/:id", handler.GetCropById)
	r.PUT("/crops/:id", handler.UpdateCrop)
	r.DELETE("/crops/:id", handler.DeleteCrop)
	return r
}

func TestCreateCrop_Success(t *testing.T) {
	mockService := new(MockCropService)
	handler := handlers.NewCropHandler(mockService)

	crop := &models.Crop{CropName: "Wheat", Category: "Food", Unit: "Kg"}
	mockService.On("CreateCrop", crop).Return(nil)

	body, _ := json.Marshal(crop)
	req := httptest.NewRequest(http.MethodPost, "/crops", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := setupCropRouter(handler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestCreateCrop_InvalidJSON(t *testing.T) {
	mockService := new(MockCropService)
	handler := handlers.NewCropHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/crops", bytes.NewBuffer([]byte(`bad-json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := setupCropRouter(handler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCropById_NotFound(t *testing.T) {
	mockService := new(MockCropService)
	handler := handlers.NewCropHandler(mockService)

	mockService.On("GetCropById", "nonexistent").Return(&models.Crop{}, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/crops/nonexistent", nil)
	rec := httptest.NewRecorder()

	r := setupCropRouter(handler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateCrop_InvalidUnit(t *testing.T) {
	mockService := new(MockCropService)
	handler := handlers.NewCropHandler(mockService)

	crop := &models.Crop{CropName: "Maize", Category: "Food", Unit: "invalid-unit"}
	body, _ := json.Marshal(crop)
	req := httptest.NewRequest(http.MethodPut, "/crops/crop123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r := setupCropRouter(handler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteCrop_Success(t *testing.T) {
	mockService := new(MockCropService)
	handler := handlers.NewCropHandler(mockService)

	mockService.On("DeleteCrop", "crop123").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/crops/crop123", nil)
	rec := httptest.NewRecorder()

	r := setupCropRouter(handler)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}
