package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FPOHandler handles FPO-related HTTP requests
type FPOHandler struct {
	fpoService services.FPOService
	logger     interfaces.Logger
}

// NewFPOHandler creates a new FPO handler
func NewFPOHandler(fpoService services.FPOService, logger interfaces.Logger) *FPOHandler {
	return &FPOHandler{
		fpoService: fpoService,
		logger:     logger,
	}
}

// CreateFPO creates a new FPO organization with AAA integration
// @Summary Create FPO Organization
// @Description Creates a new FPO organization with CEO user setup and user groups
// @Tags FPO Management
// @Accept json
// @Produce json
// @Param request body requests.CreateFPORequest true "Create FPO Request"
// @Success 201 {object} responses.SwaggerCreateFPOResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /identity/fpo/create [post]
func (h *FPOHandler) CreateFPO(c *gin.Context) {
	h.logger.Info("Creating FPO organization")

	var req requests.CreateFPORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for CreateFPO", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set request metadata
	req.RequestID = c.GetString("request_id")
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	if timestampStr := c.GetString("timestamp"); timestampStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			req.Timestamp = parsedTime
		}
	}

	h.logger.Info("Processing CreateFPO request",
		zap.String("request_id", req.RequestID),
		zap.String("fpo_name", req.Name),
		zap.String("registration_no", req.RegistrationNo),
		zap.String("ceo_phone", req.CEOUser.PhoneNumber),
	)

	// Call service
	result, err := h.fpoService.CreateFPO(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create FPO",
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)

		handleServiceError(c, err)
		return
	}

	// Type assert result
	fpoData, ok := result.(*responses.CreateFPOData)
	if !ok {
		h.logger.Error("Invalid service response type for CreateFPO",
			zap.String("request_id", req.RequestID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create response
	response := responses.NewCreateFPOResponse(fpoData, "FPO created successfully")
	response.SetRequestID(req.RequestID)

	h.logger.Info("FPO created successfully",
		zap.String("request_id", req.RequestID),
		zap.String("fpo_id", fpoData.FPOID),
		zap.String("aaa_org_id", fpoData.AAAOrgID),
		zap.String("ceo_user_id", fpoData.CEOUserID),
		zap.Int("user_groups_count", len(fpoData.UserGroups)),
	)

	c.JSON(http.StatusCreated, response)
}

// RegisterFPORef registers an FPO reference for local management
// @Summary Register FPO Reference
// @Description Registers an existing FPO organization for local reference management
// @Tags FPO Management
// @Accept json
// @Produce json
// @Param request body requests.RegisterFPORefRequest true "Register FPO Reference Request"
// @Success 201 {object} responses.SwaggerFPORefResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 409 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /identity/fpo/register [post]
func (h *FPOHandler) RegisterFPORef(c *gin.Context) {
	h.logger.Info("Registering FPO reference")

	var req requests.RegisterFPORefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for RegisterFPORef", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set request metadata
	req.RequestID = c.GetString("request_id")
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	if timestampStr := c.GetString("timestamp"); timestampStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			req.Timestamp = parsedTime
		}
	}

	h.logger.Info("Processing RegisterFPORef request",
		zap.String("request_id", req.RequestID),
		zap.String("aaa_org_id", req.AAAOrgID),
		zap.String("fpo_name", req.Name),
	)

	// Call service
	result, err := h.fpoService.RegisterFPORef(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to register FPO reference",
			zap.String("request_id", req.RequestID),
			zap.Error(err),
		)

		handleServiceError(c, err)
		return
	}

	// Type assert result
	fpoRefData, ok := result.(*responses.FPORefData)
	if !ok {
		h.logger.Error("Invalid service response type for RegisterFPORef",
			zap.String("request_id", req.RequestID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create response
	response := responses.NewFPORefResponse(fpoRefData, "FPO reference registered successfully")
	response.SetRequestID(req.RequestID)

	h.logger.Info("FPO reference registered successfully",
		zap.String("request_id", req.RequestID),
		zap.String("fpo_ref_id", fpoRefData.ID),
		zap.String("aaa_org_id", fpoRefData.AAAOrgID),
	)

	c.JSON(http.StatusCreated, response)
}

// GetFPORef retrieves an FPO reference by organization ID
// @Summary Get FPO Reference
// @Description Retrieves FPO reference information by AAA organization ID
// @Tags FPO Management
// @Accept json
// @Produce json
// @Param aaa_org_id path string true "AAA Organization ID"
// @Success 200 {object} responses.SwaggerFPORefResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /identity/fpo/reference/{aaa_org_id} [get]
func (h *FPOHandler) GetFPORef(c *gin.Context) {
	aaaOrgID := c.Param("aaa_org_id")
	requestID := c.GetString("request_id")

	h.logger.Info("Getting FPO reference",
		zap.String("request_id", requestID),
		zap.String("aaa_org_id", aaaOrgID),
	)

	if aaaOrgID == "" {
		h.logger.Error("Missing AAA organization ID parameter",
			zap.String("request_id", requestID),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "AAA organization ID is required"})
		return
	}

	// Call service
	result, err := h.fpoService.GetFPORef(c.Request.Context(), aaaOrgID)
	if err != nil {
		h.logger.Error("Failed to get FPO reference",
			zap.String("request_id", requestID),
			zap.String("aaa_org_id", aaaOrgID),
			zap.Error(err),
		)

		handleServiceError(c, err)
		return
	}

	// Type assert result
	fpoRefData, ok := result.(*responses.FPORefData)
	if !ok {
		h.logger.Error("Invalid service response type for GetFPORef",
			zap.String("request_id", requestID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Create response
	response := responses.NewFPORefResponse(fpoRefData, "FPO reference retrieved successfully")
	response.SetRequestID(requestID)

	h.logger.Info("FPO reference retrieved successfully",
		zap.String("request_id", requestID),
		zap.String("fpo_ref_id", fpoRefData.ID),
		zap.String("aaa_org_id", fpoRefData.AAAOrgID),
	)

	c.JSON(http.StatusOK, response)
}
