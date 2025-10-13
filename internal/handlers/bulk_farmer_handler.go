package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BulkFarmerHandler handles bulk farmer-related HTTP requests
type BulkFarmerHandler struct {
	bulkService services.BulkFarmerService
	aaaService  services.AAAService
	logger      interfaces.Logger
}

// NewBulkFarmerHandler creates a new bulk farmer handler
func NewBulkFarmerHandler(bulkService services.BulkFarmerService, aaaService services.AAAService, logger interfaces.Logger) *BulkFarmerHandler {
	return &BulkFarmerHandler{
		bulkService: bulkService,
		aaaService:  aaaService,
		logger:      logger,
	}
}

// BulkAddFarmers handles bulk farmer addition to FPO
// @Summary Bulk add farmers to FPO
// @Description Add multiple farmers to an FPO in a single operation
// @Tags Bulk Operations
// @Accept multipart/form-data
// @Produce json
// @Param fpo_org_id formData string true "FPO Organization ID"
// @Param input_format formData string true "Input format (csv, excel, json)"
// @Param processing_mode formData string true "Processing mode (sync, async, batch)"
// @Param file formData file true "File containing farmer data"
// @Param options formData string false "Processing options as JSON string"
// @Success 202 {object} responses.BulkOperationResponse
// @Success 200 {object} responses.BulkOperationResponse "For synchronous operations"
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 413 {object} responses.SwaggerErrorResponse "File too large"
// @Failure 415 {object} responses.SwaggerErrorResponse "Unsupported media type"
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /bulk/farmers/add [post]
func (h *BulkFarmerHandler) BulkAddFarmers(c *gin.Context) {
	h.logger.Info("Processing bulk farmer addition request")

	// Check content type
	contentType := c.ContentType()

	var req *requests.BulkFarmerAdditionRequest
	var err error

	if strings.Contains(contentType, "multipart/form-data") {
		req, err = h.parseMultipartRequest(c)
	} else if strings.Contains(contentType, "application/json") {
		req = &requests.BulkFarmerAdditionRequest{}
		err = c.ShouldBindJSON(req)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type"})
		return
	}

	if err != nil {
		h.logger.Error("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Set request metadata
	req.RequestID = c.GetString("request_id")
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.Timestamp = time.Now()

	// Validate required fields
	if req.FPOOrgID == "" || req.InputFormat == "" || req.ProcessingMode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Check if user has permission to perform bulk operations for this FPO
	// Resource: farmer, Action: bulk_create
	hasPermission, err := h.aaaService.CheckPermission(
		c.Request.Context(),
		req.UserID,
		"farmer",
		"bulk_create",
		req.FPOOrgID,
		req.OrgID,
	)
	if err != nil {
		h.logger.Error("Failed to check permission for bulk operation",
			zap.String("request_id", req.RequestID),
			zap.String("user_id", req.UserID),
			zap.String("fpo_org_id", req.FPOOrgID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to verify permissions",
			"request_id": req.RequestID,
		})
		return
	}

	if !hasPermission {
		h.logger.Warn("Permission denied for bulk farmer operation",
			zap.String("request_id", req.RequestID),
			zap.String("user_id", req.UserID),
			zap.String("fpo_org_id", req.FPOOrgID),
		)
		c.JSON(http.StatusForbidden, gin.H{
			"error":      "Insufficient permissions to perform bulk farmer operations",
			"request_id": req.RequestID,
		})
		return
	}

	h.logger.Info("Starting bulk farmer addition",
		zap.String("request_id", req.RequestID),
		zap.String("fpo_org_id", req.FPOOrgID),
		zap.String("input_format", req.InputFormat),
		zap.String("processing_mode", req.ProcessingMode),
		zap.Int("data_size", len(req.Data)),
	)

	// Call service
	result, err := h.bulkService.BulkAddFarmersToFPO(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to initiate bulk farmer addition",
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewBulkOperationResponse(result, "Bulk operation initiated successfully")
	response.RequestID = req.RequestID

	// Return appropriate status code based on processing mode
	statusCode := http.StatusAccepted // 202 for async operations
	if req.ProcessingMode == "sync" && result.Status == "COMPLETED" {
		statusCode = http.StatusOK
	}

	h.logger.Info("Bulk operation initiated successfully",
		zap.String("request_id", req.RequestID),
		zap.String("operation_id", result.OperationID),
		zap.String("status", result.Status),
	)

	c.JSON(statusCode, response)
}

// GetBulkOperationStatus retrieves the status of a bulk operation
// @Summary Get bulk operation status
// @Description Get the current status and progress of a bulk operation
// @Tags Bulk Operations
// @Produce json
// @Param operation_id path string true "Operation ID"
// @Success 200 {object} responses.BulkOperationStatusResponse
// @Failure 400 {object} responses.SwaggerErrorResponse "Invalid operation ID"
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse "Access denied to this operation"
// @Failure 404 {object} responses.SwaggerErrorResponse "Operation not found"
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /bulk/status/{operation_id} [get]
func (h *BulkFarmerHandler) GetBulkOperationStatus(c *gin.Context) {
	operationID := c.Param("operation_id")
	if operationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID is required"})
		return
	}

	h.logger.Info("Getting bulk operation status",
		zap.String("operation_id", operationID),
	)

	// Get status from service
	status, err := h.bulkService.GetBulkOperationStatus(c.Request.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to get operation status",
			zap.String("operation_id", operationID),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}

		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewBulkOperationStatusResponse(status)
	response.RequestID = c.GetString("request_id")

	c.JSON(http.StatusOK, response)
}

// CancelBulkOperation cancels a bulk operation
// @Summary Cancel bulk operation
// @Description Cancel an in-progress bulk operation
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param operation_id path string true "Operation ID"
// @Success 200 {object} responses.SwaggerBaseResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /bulk/cancel/{operation_id} [post]
func (h *BulkFarmerHandler) CancelBulkOperation(c *gin.Context) {
	operationID := c.Param("operation_id")
	if operationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID is required"})
		return
	}

	var req requests.CancelBulkOperationRequest
	req.OperationID = operationID

	// Optional: get cancellation reason from body
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore error as body is optional
		req.Reason = "User requested cancellation"
	}

	h.logger.Info("Cancelling bulk operation",
		zap.String("operation_id", operationID),
		zap.String("reason", req.Reason),
		zap.String("user_id", c.GetString("aaa_subject")),
	)

	// Cancel operation
	err := h.bulkService.CancelBulkOperation(c.Request.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to cancel operation",
			zap.String("operation_id", operationID),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}

		if strings.Contains(err.Error(), "already complete") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Operation is already complete"})
			return
		}

		handleServiceError(c, err)
		return
	}

	response := &responses.BaseResponse{
		Success:   true,
		Message:   "Operation cancelled successfully",
		RequestID: c.GetString("request_id"),
	}

	c.JSON(http.StatusOK, response)
}

// RetryFailedRecords retries failed records from a bulk operation
// @Summary Retry failed records
// @Description Retry processing of failed records from a bulk operation
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param operation_id path string true "Original operation ID"
// @Param request body requests.RetryBulkOperationRequest true "Retry request"
// @Success 202 {object} responses.BulkOperationResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /bulk/retry/{operation_id} [post]
func (h *BulkFarmerHandler) RetryFailedRecords(c *gin.Context) {
	operationID := c.Param("operation_id")
	if operationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID is required"})
		return
	}

	var req requests.RetryBulkOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid retry request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.OperationID = operationID
	req.RequestID = c.GetString("request_id")
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.Timestamp = time.Now()

	h.logger.Info("Retrying failed records",
		zap.String("operation_id", operationID),
		zap.Bool("retry_all", req.RetryAll),
		zap.Int("specific_records", len(req.RecordIndices)),
	)

	// Retry failed records
	result, err := h.bulkService.RetryFailedRecords(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to retry records",
			zap.String("operation_id", operationID),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}

		if strings.Contains(err.Error(), "cannot be retried") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		handleServiceError(c, err)
		return
	}

	response := responses.NewBulkOperationResponse(result, "Retry initiated successfully")
	response.RequestID = req.RequestID

	c.JSON(http.StatusAccepted, response)
}

// GetBulkUploadTemplate returns a template for bulk upload
// @Summary Get bulk upload template
// @Description Download a template file for bulk farmer upload
// @Tags Bulk Operations
// @Produce octet-stream
// @Param format query string true "Template format (csv, excel)"
// @Param include_sample query bool false "Include sample data"
// @Success 200 {file} file
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /bulk/template [get]
func (h *BulkFarmerHandler) GetBulkUploadTemplate(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	includeSample := c.Query("include_sample") == "true"

	h.logger.Info("Getting bulk upload template",
		zap.String("format", format),
		zap.Bool("include_sample", includeSample),
	)

	// Get template from service
	template, err := h.bulkService.GetBulkUploadTemplate(c.Request.Context(), format, includeSample)
	if err != nil {
		h.logger.Error("Failed to get template",
			zap.String("format", format),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Set appropriate content type and headers
	contentType := "text/csv"
	filename := "farmer_upload_template.csv"

	if format == "excel" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "farmer_upload_template.xlsx"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if template.Content != nil {
		c.Data(http.StatusOK, contentType, template.Content)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate template"})
	}
}

// DownloadBulkResults downloads the results of a bulk operation
// @Summary Download bulk operation results
// @Description Download a file containing the results of a bulk operation
// @Tags Bulk Operations
// @Produce octet-stream
// @Param operation_id path string true "Operation ID"
// @Param format query string false "Output format (csv, excel, json)"
// @Param include_all query bool false "Include all records or just failures"
// @Success 200 {file} file
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /bulk/results/{operation_id} [get]
func (h *BulkFarmerHandler) DownloadBulkResults(c *gin.Context) {
	operationID := c.Param("operation_id")
	if operationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID is required"})
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	h.logger.Info("Downloading bulk operation results",
		zap.String("operation_id", operationID),
		zap.String("format", format),
	)

	// Generate result file
	resultData, err := h.bulkService.GenerateResultFile(c.Request.Context(), operationID, format)
	if err != nil {
		h.logger.Error("Failed to generate result file",
			zap.String("operation_id", operationID),
			zap.Error(err),
		)

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Operation not found"})
			return
		}

		handleServiceError(c, err)
		return
	}

	// Set appropriate content type and headers
	contentType := "text/csv"
	filename := fmt.Sprintf("bulk_operation_%s_results.csv", operationID)

	switch format {
	case "excel":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = fmt.Sprintf("bulk_operation_%s_results.xlsx", operationID)
	case "json":
		contentType = "application/json"
		filename = fmt.Sprintf("bulk_operation_%s_results.json", operationID)
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	c.Data(http.StatusOK, contentType, resultData)
}

// ValidateBulkData validates bulk farmer data without processing
// @Summary Validate bulk data
// @Description Validate farmer data without actually processing it
// @Tags Bulk Operations
// @Accept json
// @Produce json
// @Param request body requests.ValidateBulkDataRequest true "Validation request"
// @Success 200 {object} responses.BulkValidationResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /bulk/validate [post]
func (h *BulkFarmerHandler) ValidateBulkData(c *gin.Context) {
	var req requests.ValidateBulkDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid validation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.RequestID = c.GetString("request_id")
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.Timestamp = time.Now()

	h.logger.Info("Validating bulk data",
		zap.String("request_id", req.RequestID),
		zap.String("fpo_org_id", req.FPOOrgID),
		zap.String("input_format", req.InputFormat),
		zap.Int("farmer_count", len(req.Farmers)),
	)

	// Validate data
	result, err := h.bulkService.ValidateBulkData(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to validate data",
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	response := responses.NewBulkValidationResponse(result)
	response.RequestID = req.RequestID

	c.JSON(http.StatusOK, response)
}

// Helper methods

func (h *BulkFarmerHandler) parseMultipartRequest(c *gin.Context) (*requests.BulkFarmerAdditionRequest, error) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(50 << 20); // 50 MB max
	err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	req := requests.NewBulkFarmerAdditionRequest()

	// Get form values
	req.FPOOrgID = c.PostForm("fpo_org_id")
	req.InputFormat = c.PostForm("input_format")
	req.ProcessingMode = c.PostForm("processing_mode")

	// Parse options if provided
	optionsStr := c.PostForm("options")
	if optionsStr != "" {
		if err := json.Unmarshal([]byte(optionsStr), &req.Options); err != nil {
			h.logger.Warn("Failed to parse options JSON", zap.Error(err))
			// Use default options
		}
	}

	// Get file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	req.Data = data

	h.logger.Debug("Parsed multipart request",
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
		zap.String("content_type", header.Header.Get("Content-Type")),
	)

	return &req, nil
}

// RegisterRoutes registers all bulk operation routes
func (h *BulkFarmerHandler) RegisterRoutes(router *gin.RouterGroup) {
	bulk := router.Group("/bulk")
	{
		// Farmer operations
		bulk.POST("/farmers/add", h.BulkAddFarmers)

		// Operation management
		bulk.GET("/status/:operation_id", h.GetBulkOperationStatus)
		bulk.POST("/cancel/:operation_id", h.CancelBulkOperation)
		bulk.POST("/retry/:operation_id", h.RetryFailedRecords)

		// Results and templates
		bulk.GET("/results/:operation_id", h.DownloadBulkResults)
		bulk.GET("/template", h.GetBulkUploadTemplate)

		// Validation
		bulk.POST("/validate", h.ValidateBulkData)
	}
}
