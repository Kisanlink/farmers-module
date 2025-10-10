package handlers

import (
	"net/http"
	"strconv"

	_ "github.com/Kisanlink/farmers-module/internal/docs" // Import for Swagger docs
	farmerReq "github.com/Kisanlink/farmers-module/internal/entities/requests"
	farmerResp "github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FarmerResponse represents a simple farmer response
type FarmerResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// FarmerListResponse represents a simple farmer list response
type FarmerListResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Data      []interface{} `json:"data"`
	Page      int           `json:"page"`
	PageSize  int           `json:"page_size"`
	Total     int           `json:"total"`
}

// SimpleResponse represents a simple success response
type SimpleResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// FarmerHandler handles HTTP requests for farmer operations
type FarmerHandler struct {
	farmerService services.FarmerService
	logger        interfaces.Logger
}

// NewFarmerHandler creates a new farmer handler
func NewFarmerHandler(farmerService services.FarmerService, logger interfaces.Logger) *FarmerHandler {
	return &FarmerHandler{
		farmerService: farmerService,
		logger:        logger,
	}
}

// CreateFarmer handles POST /api/v1/identity/farmers
// @Summary Create a new farmer
// @Description Create a new farmer profile
// @Tags identity
// @Accept json
// @Produce json
// @Param farmer body github_com_Kisanlink_farmers-module_internal_entities_requests.CreateFarmerRequest true "Farmer data"
// @Success 201 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 401 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 403 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 409 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /identity/farmers [post]
func (h *FarmerHandler) CreateFarmer(c *gin.Context) {
	var req farmerReq.CreateFarmerRequest
	_ = farmerResp.FarmerResponse{}

	h.logger.Info("Creating new farmer profile")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	response, err := h.farmerService.CreateFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer already exists" {
			status = http.StatusConflict
			apiError = base.NewConflictError("Farmer", err.Error())
		} else {
			apiError = base.NewInternalServerError("Failed to create farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to create farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	h.logger.Info("Farmer created successfully",
		zap.String("aaa_user_id", response.Data.AAAUserID),
		zap.String("aaa_org_id", response.Data.AAAOrgID))

	c.JSON(http.StatusCreated, response)
}

// GetFarmer handles GET /api/v1/identity/farmers/:aaa_user_id/:aaa_org_id
// @Summary Get farmer by user ID and org ID
// @Description Retrieve a farmer profile by AAA user ID and org ID
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Param aaa_org_id path string true "AAA Org ID"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 401 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 403 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /identity/farmers/{aaa_user_id}/{aaa_org_id} [get]
func (h *FarmerHandler) GetFarmer(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")
	aaaOrgID := c.Param("aaa_org_id")

	h.logger.Info("Getting farmer profile",
		zap.String("aaa_user_id", aaaUserID),
		zap.String("aaa_org_id", aaaOrgID))

	req := farmerReq.GetFarmerRequest{
		AAAUserID: aaaUserID,
		AAAOrgID:  aaaOrgID,
	}

	response, err := h.farmerService.GetFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to get farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to get farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// GetFarmerByUserID handles GET /api/v1/identity/farmers/user/:aaa_user_id
// @Summary Get farmer by user ID only
// @Description Retrieve a farmer profile by AAA user ID (no org required)
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/user/{aaa_user_id} [get]
func (h *FarmerHandler) GetFarmerByUserID(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")

	h.logger.Info("Getting farmer profile by user ID",
		zap.String("aaa_user_id", aaaUserID))

	req := farmerReq.GetFarmerRequest{
		AAAUserID: aaaUserID,
	}

	response, err := h.farmerService.GetFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to get farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to get farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// GetFarmerByID handles GET /api/v1/identity/farmers/id/:farmer_id
// @Summary Get farmer by farmer ID
// @Description Retrieve a farmer profile by farmer ID (primary key)
// @Tags identity
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/id/{farmer_id} [get]
func (h *FarmerHandler) GetFarmerByID(c *gin.Context) {
	farmerID := c.Param("farmer_id")

	h.logger.Info("Getting farmer profile by farmer ID",
		zap.String("farmer_id", farmerID))

	req := farmerReq.GetFarmerRequest{
		FarmerID: farmerID,
	}

	response, err := h.farmerService.GetFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", farmerID)
		} else {
			apiError = base.NewInternalServerError("Failed to get farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to get farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// ListFarmers handles GET /api/v1/identity/farmers
// @Summary List farmers
// @Description List farmers with filtering and pagination
// @Tags identity
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param aaa_org_id query string false "AAA Org ID filter"
// @Param kisan_sathi_user_id query string false "KisanSathi User ID filter"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerListResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers [get]
func (h *FarmerHandler) ListFarmers(c *gin.Context) {
	h.logger.Info("Listing farmers with filters",
		zap.String("page", c.Query("page")),
		zap.String("page_size", c.Query("page_size")),
		zap.String("aaa_org_id", c.Query("aaa_org_id")),
		zap.String("kisan_sathi_user_id", c.Query("kisan_sathi_user_id")))

	req := farmerReq.NewListFarmersRequest()

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			req.Page = pageNum
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size > 0 && size <= 100 {
			req.PageSize = size
		}
	}

	// Ensure pagination values are always valid to prevent division by zero
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if orgID := c.Query("aaa_org_id"); orgID != "" {
		req.AAAOrgID = orgID
	}
	if kisanSathiID := c.Query("kisan_sathi_user_id"); kisanSathiID != "" {
		req.KisanSathiUserID = kisanSathiID
	}

	response, err := h.farmerService.ListFarmers(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list farmers", zap.Error(err))
		apiError := base.NewInternalServerError("Failed to list farmers", err.Error())
		errorResp := base.NewErrorResponse("Failed to list farmers", apiError)
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFarmer handles PUT /api/v1/identity/farmers/:aaa_user_id/:aaa_org_id
// @Summary Update farmer by user ID and org ID
// @Description Update an existing farmer profile by user ID and org ID
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Param aaa_org_id path string true "AAA Org ID"
// @Param farmer body github_com_Kisanlink_farmers-module_internal_entities_requests.UpdateFarmerRequest true "Farmer update data"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/{aaa_user_id}/{aaa_org_id} [put]
func (h *FarmerHandler) UpdateFarmer(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")
	aaaOrgID := c.Param("aaa_org_id")

	h.logger.Info("Updating farmer profile",
		zap.String("aaa_user_id", aaaUserID),
		zap.String("aaa_org_id", aaaOrgID))

	var req farmerReq.UpdateFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set the IDs from path parameters
	req.AAAUserID = aaaUserID
	req.AAAOrgID = aaaOrgID

	response, err := h.farmerService.UpdateFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to update farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to update farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFarmerByUserID handles PUT /api/v1/identity/farmers/user/:aaa_user_id
// @Summary Update farmer by user ID only
// @Description Update an existing farmer profile by user ID (no org required)
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Param farmer body github_com_Kisanlink_farmers-module_internal_entities_requests.UpdateFarmerRequest true "Farmer update data"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/user/{aaa_user_id} [put]
func (h *FarmerHandler) UpdateFarmerByUserID(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")

	h.logger.Info("Updating farmer profile by user ID",
		zap.String("aaa_user_id", aaaUserID))

	var req farmerReq.UpdateFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set the user ID from path parameter
	req.AAAUserID = aaaUserID

	response, err := h.farmerService.UpdateFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to update farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to update farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFarmerByID handles PUT /api/v1/identity/farmers/id/:farmer_id
// @Summary Update farmer by farmer ID
// @Description Update an existing farmer profile by farmer ID (primary key)
// @Tags identity
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Param farmer body github_com_Kisanlink_farmers-module_internal_entities_requests.UpdateFarmerRequest true "Farmer update data"
// @Success 200 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerFarmerResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/id/{farmer_id} [put]
func (h *FarmerHandler) UpdateFarmerByID(c *gin.Context) {
	farmerID := c.Param("farmer_id")

	h.logger.Info("Updating farmer profile by farmer ID",
		zap.String("farmer_id", farmerID))

	var req farmerReq.UpdateFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set the farmer ID from path parameter
	req.FarmerID = farmerID

	response, err := h.farmerService.UpdateFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", farmerID)
		} else {
			apiError = base.NewInternalServerError("Failed to update farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to update farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFarmer handles DELETE /api/v1/identity/farmers/:aaa_user_id/:aaa_org_id
// @Summary Delete farmer by user ID and org ID
// @Description Delete a farmer profile by user ID and org ID
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Param aaa_org_id path string true "AAA Org ID"
// @Success 200 {object} SimpleResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/{aaa_user_id}/{aaa_org_id} [delete]
func (h *FarmerHandler) DeleteFarmer(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")
	aaaOrgID := c.Param("aaa_org_id")

	h.logger.Info("Deleting farmer profile",
		zap.String("aaa_user_id", aaaUserID),
		zap.String("aaa_org_id", aaaOrgID))

	req := farmerReq.DeleteFarmerRequest{
		AAAUserID: aaaUserID,
		AAAOrgID:  aaaOrgID,
	}

	err := h.farmerService.DeleteFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to delete farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to delete farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to delete farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	successResp := base.NewSuccessResponse("Farmer deleted successfully", nil)
	if req.RequestID != "" {
		successResp.RequestID = req.RequestID
	}

	h.logger.Info("Farmer deleted successfully",
		zap.String("aaa_user_id", aaaUserID),
		zap.String("aaa_org_id", aaaOrgID))

	c.JSON(http.StatusOK, successResp)
}

// DeleteFarmerByUserID handles DELETE /api/v1/identity/farmers/user/:aaa_user_id
// @Summary Delete farmer by user ID only
// @Description Delete a farmer profile by user ID (no org required)
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Success 200 {object} SimpleResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/user/{aaa_user_id} [delete]
func (h *FarmerHandler) DeleteFarmerByUserID(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")

	h.logger.Info("Deleting farmer profile by user ID",
		zap.String("aaa_user_id", aaaUserID))

	req := farmerReq.DeleteFarmerRequest{
		AAAUserID: aaaUserID,
	}

	err := h.farmerService.DeleteFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to delete farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", aaaUserID)
		} else {
			apiError = base.NewInternalServerError("Failed to delete farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to delete farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	successResp := base.NewSuccessResponse("Farmer deleted successfully", nil)
	if req.RequestID != "" {
		successResp.RequestID = req.RequestID
	}

	h.logger.Info("Farmer deleted successfully",
		zap.String("aaa_user_id", aaaUserID))

	c.JSON(http.StatusOK, successResp)
}

// DeleteFarmerByID handles DELETE /api/v1/identity/farmers/id/:farmer_id
// @Summary Delete farmer by farmer ID
// @Description Delete a farmer profile by farmer ID (primary key)
// @Tags identity
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Success 200 {object} SimpleResponse
// @Failure 400 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Failure 404 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.ErrorResponse
// @Failure 500 {object} github_com_Kisanlink_farmers-module_internal_entities_responses.SwaggerErrorResponse
// @Router /identity/farmers/id/{farmer_id} [delete]
func (h *FarmerHandler) DeleteFarmerByID(c *gin.Context) {
	farmerID := c.Param("farmer_id")

	h.logger.Info("Deleting farmer profile by farmer ID",
		zap.String("farmer_id", farmerID))

	req := farmerReq.DeleteFarmerRequest{
		FarmerID: farmerID,
	}

	err := h.farmerService.DeleteFarmer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to delete farmer", zap.Error(err))
		status := http.StatusInternalServerError
		var apiError base.ErrorInterface

		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
			apiError = base.NewNotFoundError("Farmer", farmerID)
		} else {
			apiError = base.NewInternalServerError("Failed to delete farmer", err.Error())
		}

		errorResp := base.NewErrorResponse("Failed to delete farmer", apiError)
		c.JSON(status, errorResp)
		return
	}

	successResp := base.NewSuccessResponse("Farmer deleted successfully", nil)
	if req.RequestID != "" {
		successResp.RequestID = req.RequestID
	}

	h.logger.Info("Farmer deleted successfully",
		zap.String("farmer_id", farmerID))

	c.JSON(http.StatusOK, successResp)
}

// Wrapper functions for use in routes
// These functions create a handler instance and return the method as a gin.HandlerFunc

// CreateFarmer creates a handler function for creating farmers
func CreateFarmer(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.CreateFarmer
}

// GetFarmer creates a handler function for getting farmers
func GetFarmer(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.GetFarmer
}

// GetFarmerByUserID creates a handler function for getting farmers by user ID
func GetFarmerByUserID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.GetFarmerByUserID
}

// GetFarmerByID creates a handler function for getting farmers by farmer ID
func GetFarmerByID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.GetFarmerByID
}

// ListFarmers creates a handler function for listing farmers
func ListFarmers(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.ListFarmers
}

// UpdateFarmer creates a handler function for updating farmers
func UpdateFarmer(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.UpdateFarmer
}

// UpdateFarmerByUserID creates a handler function for updating farmers by user ID
func UpdateFarmerByUserID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.UpdateFarmerByUserID
}

// UpdateFarmerByID creates a handler function for updating farmers by farmer ID
func UpdateFarmerByID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.UpdateFarmerByID
}

// DeleteFarmer creates a handler function for deleting farmers
func DeleteFarmer(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.DeleteFarmer
}

// DeleteFarmerByUserID creates a handler function for deleting farmers by user ID
func DeleteFarmerByUserID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.DeleteFarmerByUserID
}

// DeleteFarmerByID creates a handler function for deleting farmers by farmer ID
func DeleteFarmerByID(farmerService services.FarmerService, logger interfaces.Logger) gin.HandlerFunc {
	handler := NewFarmerHandler(farmerService, logger)
	return handler.DeleteFarmerByID
}
