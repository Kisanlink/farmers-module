package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FPOConfigHandler handles FPO configuration HTTP requests
type FPOConfigHandler struct {
	service services.FPOConfigService
	logger  interfaces.Logger
}

// NewFPOConfigHandler creates a new FPO configuration handler
func NewFPOConfigHandler(service services.FPOConfigService, logger interfaces.Logger) *FPOConfigHandler {
	return &FPOConfigHandler{
		service: service,
		logger:  logger,
	}
}

// GetFPOConfig retrieves FPO configuration by FPO ID
// @Summary Get FPO Configuration
// @Description Retrieves FPO configuration for e-commerce integration
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param fpo_id path string true "FPO ID"
// @Success 200 {object} responses.SwaggerFPOConfigResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config/{fpo_id} [get]
func (h *FPOConfigHandler) GetFPOConfig(c *gin.Context) {
	fpoID := c.Param("fpo_id")
	requestID := c.GetString("request_id")

	h.logger.Info("Getting FPO configuration",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	if fpoID == "" {
		h.logger.Error("Missing FPO ID parameter", zap.String("request_id", requestID))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "FPO ID is required",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Call service
	data, err := h.service.GetFPOConfig(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to get FPO configuration",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewFPOConfigResponse(data, "FPO configuration retrieved successfully")
	response.SetRequestID(requestID)

	h.logger.Info("FPO configuration retrieved successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, response)
}

// CreateFPOConfig creates a new FPO configuration
// @Summary Create FPO Configuration
// @Description Creates a new FPO configuration for e-commerce integration (Admin only)
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param request body requests.CreateFPOConfigRequest true "Create FPO Config Request"
// @Success 201 {object} responses.SwaggerFPOConfigResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 409 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config [post]
func (h *FPOConfigHandler) CreateFPOConfig(c *gin.Context) {
	requestID := c.GetString("request_id")
	h.logger.Info("Creating FPO configuration", zap.String("request_id", requestID))

	var req requests.CreateFPOConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Set request metadata
	req.RequestID = requestID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")

	h.logger.Info("Processing CreateFPOConfig request",
		zap.String("request_id", requestID),
		zap.String("fpo_id", req.FPOID),
		zap.String("fpo_name", req.FPOName),
	)

	// Call service
	data, err := h.service.CreateFPOConfig(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create FPO configuration",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewFPOConfigResponse(data, "FPO configuration created successfully")
	response.SetRequestID(requestID)

	h.logger.Info("FPO configuration created successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", data.FPOID),
	)

	c.JSON(http.StatusCreated, response)
}

// UpdateFPOConfig updates an existing FPO configuration
// @Summary Update FPO Configuration
// @Description Updates an existing FPO configuration (Admin only)
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param fpo_id path string true "FPO ID"
// @Param request body requests.UpdateFPOConfigRequest true "Update FPO Config Request"
// @Success 200 {object} responses.SwaggerFPOConfigResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config/{fpo_id} [put]
func (h *FPOConfigHandler) UpdateFPOConfig(c *gin.Context) {
	fpoID := c.Param("fpo_id")
	requestID := c.GetString("request_id")

	h.logger.Info("Updating FPO configuration",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	if fpoID == "" {
		h.logger.Error("Missing FPO ID parameter", zap.String("request_id", requestID))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "FPO ID is required",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	var req requests.UpdateFPOConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Set request metadata
	req.RequestID = requestID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")

	// Call service
	data, err := h.service.UpdateFPOConfig(c.Request.Context(), fpoID, &req)
	if err != nil {
		h.logger.Error("Failed to update FPO configuration",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewFPOConfigResponse(data, "FPO configuration updated successfully")
	response.SetRequestID(requestID)

	h.logger.Info("FPO configuration updated successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, response)
}

// DeleteFPOConfig deletes an FPO configuration
// @Summary Delete FPO Configuration
// @Description Deletes an FPO configuration (soft delete, Admin only)
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param fpo_id path string true "FPO ID"
// @Success 200 {object} responses.SwaggerBaseResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config/{fpo_id} [delete]
func (h *FPOConfigHandler) DeleteFPOConfig(c *gin.Context) {
	fpoID := c.Param("fpo_id")
	requestID := c.GetString("request_id")

	h.logger.Info("Deleting FPO configuration",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	if fpoID == "" {
		h.logger.Error("Missing FPO ID parameter", zap.String("request_id", requestID))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "FPO ID is required",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Call service
	err := h.service.DeleteFPOConfig(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to delete FPO configuration",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	h.logger.Info("FPO configuration deleted successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "FPO configuration deleted successfully",
		"request_id": requestID,
	})
}

// ListFPOConfigs lists all FPO configurations
// @Summary List FPO Configurations
// @Description Lists all FPO configurations with pagination (Admin only)
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search by FPO ID or name"
// @Param status query string false "Filter by health status"
// @Success 200 {object} responses.SwaggerFPOConfigListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config [get]
func (h *FPOConfigHandler) ListFPOConfigs(c *gin.Context) {
	requestID := c.GetString("request_id")
	h.logger.Info("Listing FPO configurations", zap.String("request_id", requestID))

	var req requests.ListFPOConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid query parameters", zap.String("request_id", requestID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid query parameters",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Set request metadata
	req.RequestID = requestID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")

	// Call service
	data, pagination, err := h.service.ListFPOConfigs(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list FPO configurations",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewFPOConfigListResponse(data, pagination, "FPO configurations retrieved successfully")
	response.SetRequestID(requestID)

	h.logger.Info("FPO configurations listed successfully",
		zap.String("request_id", requestID),
		zap.Int("count", len(data)),
	)

	c.JSON(http.StatusOK, response)
}

// CheckERPHealth checks the health of FPO's ERP service
// @Summary Check ERP Health
// @Description Checks if the FPO's ERP service is reachable
// @Tags FPO Config
// @Accept json
// @Produce json
// @Param fpo_id path string true "FPO ID"
// @Success 200 {object} responses.SwaggerFPOHealthCheckResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /fpo-config/{fpo_id}/health [get]
func (h *FPOConfigHandler) CheckERPHealth(c *gin.Context) {
	fpoID := c.Param("fpo_id")
	requestID := c.GetString("request_id")

	h.logger.Info("Checking ERP health",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
	)

	if fpoID == "" {
		h.logger.Error("Missing FPO ID parameter", zap.String("request_id", requestID))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "FPO ID is required",
			"error":   "ERR_INVALID_INPUT",
		})
		return
	}

	// Call service
	data, err := h.service.CheckERPHealth(c.Request.Context(), fpoID)
	if err != nil {
		h.logger.Error("Failed to check ERP health",
			zap.String("request_id", requestID),
			zap.String("fpo_id", fpoID),
			zap.Error(err),
		)
		handleServiceError(c, err)
		return
	}

	// Create response
	response := responses.NewFPOHealthCheckResponse(data)
	response.SetRequestID(requestID)

	h.logger.Info("ERP health checked successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_id", fpoID),
		zap.String("status", data.Status),
	)

	c.JSON(http.StatusOK, response)
}
