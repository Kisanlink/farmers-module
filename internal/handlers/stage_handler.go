package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StageHandler handles HTTP requests for stage operations
type StageHandler struct {
	stageService services.StageService
	logger       interfaces.Logger
}

// NewStageHandler creates a new stage handler
func NewStageHandler(stageService services.StageService, logger interfaces.Logger) *StageHandler {
	return &StageHandler{
		stageService: stageService,
		logger:       logger,
	}
}

// CreateStage handles POST /api/v1/stages
// @Summary Create a new stage
// @Description Create a new growth stage for crops
// @Tags Stages
// @Accept json
// @Produce json
// @Param stage body requests.CreateStageRequest true "Stage details"
// @Success 201 {object} responses.StageResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages [post]
func (h *StageHandler) CreateStage(c *gin.Context) {
	var req requests.CreateStageRequest

	h.logger.Info("Creating new stage")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Add user context from middleware
	req.UserID = c.GetString("user_id")
	req.OrgID = c.GetString("org_id")
	req.RequestID = c.GetString("request_id")

	response, err := h.stageService.CreateStage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create stage", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Stage created successfully", zap.String("stage_name", req.StageName))
	c.JSON(http.StatusCreated, response)
}

// GetStage handles GET /api/v1/stages/:id
// @Summary Get a stage by ID
// @Description Get a growth stage by its ID
// @Tags Stages
// @Accept json
// @Produce json
// @Param id path string true "Stage ID"
// @Success 200 {object} responses.StageResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages/{id} [get]
func (h *StageHandler) GetStage(c *gin.Context) {
	stageID := c.Param("id")

	h.logger.Info("Getting stage", zap.String("stage_id", stageID))

	req := &requests.GetStageRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
		ID: stageID,
	}

	response, err := h.stageService.GetStage(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get stage", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateStage handles PUT /api/v1/stages/:id
// @Summary Update a stage
// @Description Update an existing growth stage
// @Tags Stages
// @Accept json
// @Produce json
// @Param id path string true "Stage ID"
// @Param stage body requests.UpdateStageRequest true "Stage update details"
// @Success 200 {object} responses.StageResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages/{id} [put]
func (h *StageHandler) UpdateStage(c *gin.Context) {
	stageID := c.Param("id")
	var req requests.UpdateStageRequest

	h.logger.Info("Updating stage", zap.String("stage_id", stageID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set ID from path parameter
	req.ID = stageID
	req.UserID = c.GetString("user_id")
	req.OrgID = c.GetString("org_id")
	req.RequestID = c.GetString("request_id")

	response, err := h.stageService.UpdateStage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update stage", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Stage updated successfully", zap.String("stage_id", stageID))
	c.JSON(http.StatusOK, response)
}

// DeleteStage handles DELETE /api/v1/stages/:id
// @Summary Delete a stage
// @Description Delete a growth stage (soft delete)
// @Tags Stages
// @Accept json
// @Produce json
// @Param id path string true "Stage ID"
// @Success 200 {object} responses.BaseResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages/{id} [delete]
func (h *StageHandler) DeleteStage(c *gin.Context) {
	stageID := c.Param("id")

	h.logger.Info("Deleting stage", zap.String("stage_id", stageID))

	req := &requests.DeleteStageRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
		ID: stageID,
	}

	response, err := h.stageService.DeleteStage(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete stage", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Stage deleted successfully", zap.String("stage_id", stageID))
	c.JSON(http.StatusOK, response)
}

// ListStages handles GET /api/v1/stages
// @Summary List all stages
// @Description List all growth stages with pagination and filtering
// @Tags Stages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search by stage name or description"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} responses.StageListResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages [get]
func (h *StageHandler) ListStages(c *gin.Context) {
	h.logger.Info("Listing stages",
		zap.String("page", c.Query("page")),
		zap.String("page_size", c.Query("page_size")),
		zap.String("search", c.Query("search")))

	req := &requests.ListStagesRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
	}

	// Parse pagination
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

	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Parse filters
	if search := c.Query("search"); search != "" {
		req.Search = search
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			req.IsActive = &isActive
		}
	}

	response, err := h.stageService.ListStages(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list stages", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStageLookup handles GET /api/v1/stages/lookup
// @Summary Get stage lookup data
// @Description Get simplified stage data for dropdowns/lookups
// @Tags Stages
// @Accept json
// @Produce json
// @Success 200 {object} responses.StageLookupResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/stages/lookup [get]
func (h *StageHandler) GetStageLookup(c *gin.Context) {
	h.logger.Info("Getting stage lookup data")

	req := &requests.GetStageLookupRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
	}

	response, err := h.stageService.GetStageLookup(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get stage lookup data", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// AssignStageToCrop handles POST /api/v1/crops/:crop_id/stages
// @Summary Assign a stage to a crop
// @Description Assign a growth stage to a crop with order and duration
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param crop_stage body requests.AssignStageToCropRequest true "Crop stage details"
// @Success 201 {object} responses.CropStageResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/crops/{id}/stages [post]
func (h *StageHandler) AssignStageToCrop(c *gin.Context) {
	cropID := c.Param("id")
	var req requests.AssignStageToCropRequest

	h.logger.Info("Assigning stage to crop", zap.String("crop_id", cropID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set crop ID from path parameter
	req.CropID = cropID
	req.UserID = c.GetString("user_id")
	req.OrgID = c.GetString("org_id")
	req.RequestID = c.GetString("request_id")

	response, err := h.stageService.AssignStageToCrop(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to assign stage to crop", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Stage assigned to crop successfully",
		zap.String("crop_id", cropID),
		zap.String("stage_id", req.StageID))
	c.JSON(http.StatusCreated, response)
}

// GetCropStages handles GET /api/v1/crops/:crop_id/stages
// @Summary Get all stages for a crop
// @Description Get all growth stages assigned to a crop in order
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Success 200 {object} responses.CropStagesResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/crops/{id}/stages [get]
func (h *StageHandler) GetCropStages(c *gin.Context) {
	cropID := c.Param("id")

	h.logger.Info("Getting crop stages", zap.String("crop_id", cropID))

	req := &requests.GetCropStagesRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
		CropID: cropID,
	}

	response, err := h.stageService.GetCropStages(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get crop stages", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCropStage handles PUT /api/v1/crops/:crop_id/stages/:stage_id
// @Summary Update a crop stage
// @Description Update a growth stage assigned to a crop
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param stage_id path string true "Stage ID"
// @Param crop_stage body requests.UpdateCropStageRequest true "Crop stage update details"
// @Success 200 {object} responses.CropStageResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/crops/{id}/stages/{stage_id} [put]
func (h *StageHandler) UpdateCropStage(c *gin.Context) {
	cropID := c.Param("id")
	stageID := c.Param("stage_id")
	var req requests.UpdateCropStageRequest

	h.logger.Info("Updating crop stage",
		zap.String("crop_id", cropID),
		zap.String("stage_id", stageID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set IDs from path parameters
	req.CropID = cropID
	req.StageID = stageID
	req.UserID = c.GetString("user_id")
	req.OrgID = c.GetString("org_id")
	req.RequestID = c.GetString("request_id")

	response, err := h.stageService.UpdateCropStage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update crop stage", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Crop stage updated successfully",
		zap.String("crop_id", cropID),
		zap.String("stage_id", stageID))
	c.JSON(http.StatusOK, response)
}

// RemoveStageFromCrop handles DELETE /api/v1/crops/:crop_id/stages/:stage_id
// @Summary Remove a stage from a crop
// @Description Remove a growth stage from a crop (soft delete)
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param stage_id path string true "Stage ID"
// @Success 200 {object} responses.BaseResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/crops/{id}/stages/{stage_id} [delete]
func (h *StageHandler) RemoveStageFromCrop(c *gin.Context) {
	cropID := c.Param("id")
	stageID := c.Param("stage_id")

	h.logger.Info("Removing stage from crop",
		zap.String("crop_id", cropID),
		zap.String("stage_id", stageID))

	req := &requests.RemoveStageFromCropRequest{
		BaseRequest: requests.BaseRequest{
			UserID:    c.GetString("user_id"),
			OrgID:     c.GetString("org_id"),
			RequestID: c.GetString("request_id"),
		},
		CropID:  cropID,
		StageID: stageID,
	}

	response, err := h.stageService.RemoveStageFromCrop(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to remove stage from crop", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Stage removed from crop successfully",
		zap.String("crop_id", cropID),
		zap.String("stage_id", stageID))
	c.JSON(http.StatusOK, response)
}

// ReorderCropStages handles POST /api/v1/crops/:crop_id/stages/reorder
// @Summary Reorder crop stages
// @Description Reorder the growth stages for a crop
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param reorder body requests.ReorderCropStagesRequest true "Stage reorder details"
// @Success 200 {object} responses.BaseResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/crops/{id}/stages/reorder [post]
func (h *StageHandler) ReorderCropStages(c *gin.Context) {
	cropID := c.Param("id")
	var req requests.ReorderCropStagesRequest

	h.logger.Info("Reordering crop stages", zap.String("crop_id", cropID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		errorResp := base.NewErrorResponse("Invalid request format", base.NewValidationError("Invalid request format", err.Error()))
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	// Set crop ID from path parameter
	req.CropID = cropID
	req.UserID = c.GetString("user_id")
	req.OrgID = c.GetString("org_id")
	req.RequestID = c.GetString("request_id")

	response, err := h.stageService.ReorderCropStages(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to reorder crop stages", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	h.logger.Info("Crop stages reordered successfully", zap.String("crop_id", cropID))
	c.JSON(http.StatusOK, response)
}

// handleServiceError converts service errors to HTTP responses
func (h *StageHandler) handleServiceError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	var apiError base.ErrorInterface

	switch err {
	case common.ErrInvalidInput:
		status = http.StatusBadRequest
		apiError = base.NewValidationError("Invalid input", err.Error())
	case common.ErrNotFound:
		status = http.StatusNotFound
		apiError = base.NewNotFoundError("Resource", "")
	case common.ErrAlreadyExists:
		status = http.StatusConflict
		apiError = base.NewConflictError("Resource", err.Error())
	case common.ErrForbidden:
		status = http.StatusForbidden
		apiError = base.NewForbiddenError("Insufficient permissions")
	default:
		apiError = base.NewInternalServerError("Internal server error", err.Error())
	}

	errorResp := base.NewErrorResponse("Operation failed", apiError)
	c.JSON(status, errorResp)
}
