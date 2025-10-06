package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBulkTestRouter(handler *handlers.BulkFarmerHandler) *gin.Engine {
	router := testutils.SetupTestRouter()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)
	return router
}

func TestBulkFarmerHandler_BulkAddFarmers_JSON(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	// Setup service expectations
	expectedResponse := &responses.BulkOperationData{
		OperationID: "bulk_op_123",
		Status:      "PENDING",
		StatusURL:   "/api/v1/bulk/status/bulk_op_123",
		Message:     "Bulk operation initiated",
	}

	mockService.BulkAddFarmersToFPOFunc = func(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error) {
		return expectedResponse, nil
	}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Create request
	reqBody := requests.BulkFarmerAdditionRequest{
		FPOOrgID:       "fpo_123",
		InputFormat:    "json",
		ProcessingMode: "async",
		Data:           []byte(`[{"first_name":"John","last_name":"Doe","phone_number":"9876543210"}]`),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/bulk/farmers/add", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response responses.BulkOperationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "bulk_op_123", response.Data.OperationID)

}

func TestBulkFarmerHandler_BulkAddFarmers_Multipart(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	// Setup service expectations
	expectedResponse := &responses.BulkOperationData{
		OperationID: "bulk_op_123",
		Status:      "PENDING",
		StatusURL:   "/api/v1/bulk/status/bulk_op_123",
		Message:     "Bulk operation initiated",
	}

	mockService.BulkAddFarmersToFPOFunc = func(ctx context.Context, req *requests.BulkFarmerAdditionRequest) (*responses.BulkOperationData, error) {
		return expectedResponse, nil
	}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Create multipart request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("fpo_org_id", "fpo_123")
	writer.WriteField("input_format", "csv")
	writer.WriteField("processing_mode", "sync")

	part, _ := writer.CreateFormFile("file", "farmers.csv")
	part.Write([]byte("first_name,last_name,phone_number\nJohn,Doe,9876543210"))

	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/bulk/farmers/add", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response responses.BulkOperationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "bulk_op_123", response.Data.OperationID)

}

func TestBulkFarmerHandler_BulkAddFarmers_ValidationError(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Create request with missing required fields
	reqBody := requests.BulkFarmerAdditionRequest{
		// Missing FPOOrgID, InputFormat, ProcessingMode
		Data: []byte(`[{"first_name":"John"}]`),
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/bulk/farmers/add", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "Missing required fields")
}

func TestBulkFarmerHandler_GetBulkOperationStatus(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Setup service expectations
	expectedStatus := &responses.BulkOperationStatusData{
		OperationID: "bulk_op_123",
		Status:      "PROCESSING",
		Progress: responses.ProgressInfo{
			Total:      100,
			Processed:  50,
			Successful: 45,
			Failed:     5,
			Percentage: 50.0,
		},
	}

	mockService.GetBulkOperationStatusFunc = func(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error) {
		if operationID == "bulk_op_123" {
			return expectedStatus, nil
		}
		return nil, errors.New("operation not found")
	}

	req := httptest.NewRequest("GET", "/api/v1/bulk/status/bulk_op_123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response responses.BulkOperationStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "bulk_op_123", response.Data.OperationID)
	assert.Equal(t, "PROCESSING", response.Data.Status)
	assert.Equal(t, 50.0, response.Data.Progress.Percentage)

}

func TestBulkFarmerHandler_GetBulkOperationStatus_NotFound(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockService.GetBulkOperationStatusFunc = func(ctx context.Context, operationID string) (*responses.BulkOperationStatusData, error) {
		return nil, errors.New("operation not found")
	}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	req := httptest.NewRequest("GET", "/api/v1/bulk/status/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

}

func TestBulkFarmerHandler_CancelBulkOperation(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Setup service expectations
	mockService.CancelBulkOperationFunc = func(ctx context.Context, operationID string) error {
		return nil
	}

	reqBody := requests.CancelBulkOperationRequest{
		Reason: "User requested cancellation",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/bulk/cancel/bulk_op_123", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response responses.BaseResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "cancelled successfully")

}

func TestBulkFarmerHandler_GetBulkUploadTemplate(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Setup service expectations
	expectedTemplate := &responses.BulkTemplateData{
		Format:   "csv",
		FileName: "farmer_upload_template.csv",
		Content:  []byte("first_name,last_name,phone_number\n"),
	}

	mockService.GetBulkUploadTemplateFunc = func(ctx context.Context, format string, includeExample bool) (*responses.BulkTemplateData, error) {
		return expectedTemplate, nil
	}

	req := httptest.NewRequest("GET", "/api/v1/bulk/template?format=csv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "farmer_upload_template.csv")
	assert.Equal(t, "first_name,last_name,phone_number\n", w.Body.String())

}

func TestBulkFarmerHandler_ValidateBulkData(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	// Setup service expectations
	expectedValidation := &responses.BulkValidationData{
		IsValid:      true,
		TotalRecords: 2,
		ValidRecords: 2,
		Errors:       []responses.ValidationError{},
	}

	mockService.ValidateBulkDataFunc = func(ctx context.Context, req *requests.ValidateBulkDataRequest) (*responses.BulkValidationData, error) {
		return expectedValidation, nil
	}

	reqBody := requests.ValidateBulkDataRequest{
		FPOOrgID:    "fpo_123",
		InputFormat: "json",
		Farmers: []requests.FarmerBulkData{
			{
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "9876543210",
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/bulk/validate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response responses.BulkValidationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.True(t, response.Data.IsValid)
	assert.Equal(t, 2, response.Data.TotalRecords)
	assert.Equal(t, 2, response.Data.ValidRecords)

}

func TestBulkFarmerHandler_UnsupportedContentType(t *testing.T) {
	mockService := &testutils.MockBulkFarmerService{}
	mockLogger := &testutils.MockLogger{}

	mockAAAService := &testutils.MockAAAService{}
	handler := handlers.NewBulkFarmerHandler(mockService, mockAAAService, mockLogger)
	router := setupBulkTestRouter(handler)

	req := httptest.NewRequest("POST", "/api/v1/bulk/farmers/add", strings.NewReader("unsupported content"))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "Unsupported content type")
}
