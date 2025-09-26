package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CropMasterHandler handles crop master data HTTP requests
type CropMasterHandler struct {
	cropService services.CropService
	logger      interfaces.Logger
}

// NewCropMasterHandler creates a new crop master handler
func NewCropMasterHandler(cropService services.CropService, logger interfaces.Logger) *CropMasterHandler {
	return &CropMasterHandler{
		cropService: cropService,
		logger:      logger,
	}
}

// CreateCrop handles crop creation
// @Summary Create a new crop
// @Description Create a new crop with master data
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param request body requests.CreateCropRequest true "Create crop request"
// @Success 201 {object} responses.CropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops [post]
func (h *CropMasterHandler) CreateCrop(c *gin.Context) {
	var req requests.CreateCropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create crop request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Creating crop", zap.String("name", req.Name), zap.String("category", string(req.Category)))

	// Call service
	result, err := h.cropService.CreateCrop(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create crop", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusCreated, result)
}

// ListCrops handles crop listing
// @Summary List crops
// @Description List crops with optional filtering
// @Tags Crop Master Data
// @Produce json
// @Param category query string false "Filter by category"
// @Param season query string false "Filter by season"
// @Param search query string false "Search by name"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} responses.CropsListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops [get]
func (h *CropMasterHandler) ListCrops(c *gin.Context) {
	req := &requests.ListCropsRequest{}

	// Parse query parameters
	if category := c.Query("category"); category != "" {
		categoryType := crop.CropCategory(category)
		req.Category = &categoryType
	}
	if season := c.Query("season"); season != "" {
		seasonType := crop.CropSeason(season)
		req.Season = &seasonType
	}
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	h.logger.Info("Listing crops", zap.String("user_id", req.UserID))

	// Call service
	result, err := h.cropService.ListCrops(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list crops", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusOK, result)
}

// GetCrop handles getting a specific crop
// @Summary Get crop by ID
// @Description Get a specific crop by its ID
// @Tags Crop Master Data
// @Produce json
// @Param id path string true "Crop ID"
// @Success 200 {object} responses.CropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [get]
func (h *CropMasterHandler) GetCrop(c *gin.Context) {
	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	h.logger.Info("Getting crop", zap.String("crop_id", cropID))

	// Call service
	result, err := h.cropService.GetCrop(c.Request.Context(), cropID)
	if err != nil {
		h.logger.Error("Failed to get crop", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}

// UpdateCrop handles crop updates
// @Summary Update crop
// @Description Update an existing crop
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param request body requests.UpdateCropRequest true "Update crop request"
// @Success 200 {object} responses.CropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [put]
func (h *CropMasterHandler) UpdateCrop(c *gin.Context) {
	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	var req requests.UpdateCropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update crop request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.CropID = cropID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Updating crop", zap.String("crop_id", cropID))

	// Call service
	result, err := h.cropService.UpdateCrop(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update crop", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusOK, result)
}

// DeleteCrop handles crop deletion
// @Summary Delete crop
// @Description Delete a crop
// @Tags Crop Master Data
// @Param id path string true "Crop ID"
// @Success 204
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [delete]
func (h *CropMasterHandler) DeleteCrop(c *gin.Context) {
	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	h.logger.Info("Deleting crop", zap.String("crop_id", cropID))

	// Call service
	err := h.cropService.DeleteCrop(c.Request.Context(), cropID)
	if err != nil {
		h.logger.Error("Failed to delete crop", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateVariety handles variety creation
// @Summary Create a new crop variety
// @Description Create a new variety for a specific crop
// @Tags Crop Varieties
// @Accept json
// @Produce json
// @Param request body requests.CreateVarietyRequest true "Create variety request"
// @Success 201 {object} responses.CropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/varieties [post]
func (h *CropMasterHandler) CreateVariety(c *gin.Context) {
	var req requests.CreateVarietyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create variety request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Creating variety", zap.String("crop_id", req.CropID), zap.String("variety_name", req.VarietyName))

	// Call service
	result, err := h.cropService.CreateVariety(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create variety", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusCreated, result)
}

// ListVarieties handles variety listing for a crop
// @Summary List varieties for a crop
// @Description List all varieties for a specific crop
// @Tags Crop Varieties
// @Produce json
// @Param crop_id path string true "Crop ID"
// @Success 200 {object} responses.CropVarietiesListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/{crop_id}/varieties [get]
func (h *CropMasterHandler) ListVarieties(c *gin.Context) {
	cropID := c.Param("crop_id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	h.logger.Info("Listing varieties", zap.String("crop_id", cropID))

	// Call service
	result, err := h.cropService.ListVarietiesByCrop(c.Request.Context(), cropID)
	if err != nil {
		h.logger.Error("Failed to list varieties", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}

// GetVariety handles getting a specific variety
// @Summary Get variety by ID
// @Description Get a specific variety by its ID
// @Tags Crop Varieties
// @Produce json
// @Param id path string true "Variety ID"
// @Success 200 {object} responses.CropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/varieties/{id} [get]
func (h *CropMasterHandler) GetVariety(c *gin.Context) {
	varietyID := c.Param("id")
	if varietyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Variety ID is required"})
		return
	}

	h.logger.Info("Getting variety", zap.String("variety_id", varietyID))

	// Call service
	result, err := h.cropService.GetVariety(c.Request.Context(), varietyID)
	if err != nil {
		h.logger.Error("Failed to get variety", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}

// UpdateVariety handles variety updates
// @Summary Update variety
// @Description Update an existing variety
// @Tags Crop Varieties
// @Accept json
// @Produce json
// @Param id path string true "Variety ID"
// @Param request body requests.UpdateVarietyRequest true "Update variety request"
// @Success 200 {object} responses.CropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/varieties/{id} [put]
func (h *CropMasterHandler) UpdateVariety(c *gin.Context) {
	varietyID := c.Param("id")
	if varietyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Variety ID is required"})
		return
	}

	var req requests.UpdateVarietyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update variety request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.VarietyID = varietyID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Updating variety", zap.String("variety_id", varietyID))

	// Call service
	result, err := h.cropService.UpdateVariety(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update variety", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusOK, result)
}

// DeleteVariety handles variety deletion
// @Summary Delete variety
// @Description Delete a variety
// @Tags Crop Varieties
// @Param id path string true "Variety ID"
// @Success 204
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/varieties/{id} [delete]
func (h *CropMasterHandler) DeleteVariety(c *gin.Context) {
	varietyID := c.Param("id")
	if varietyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Variety ID is required"})
		return
	}

	h.logger.Info("Deleting variety", zap.String("variety_id", varietyID))

	// Call service
	err := h.cropService.DeleteVariety(c.Request.Context(), varietyID)
	if err != nil {
		h.logger.Error("Failed to delete variety", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateStage handles stage creation
// @Summary Create a new crop stage
// @Description Create a new growth stage for a specific crop
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param request body requests.CreateStageRequest true "Create stage request"
// @Success 201 {object} responses.CropStageResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/stages [post]
func (h *CropMasterHandler) CreateStage(c *gin.Context) {
	var req requests.CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create stage request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Creating stage", zap.String("crop_id", req.CropID), zap.String("stage_name", req.StageName))

	// Call service
	result, err := h.cropService.CreateStage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create stage", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusCreated, result)
}

// ListStages handles stage listing for a crop
// @Summary List stages for a crop
// @Description List all growth stages for a specific crop
// @Tags Crop Stages
// @Produce json
// @Param crop_id path string true "Crop ID"
// @Success 200 {object} responses.CropStagesListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/{crop_id}/stages [get]
func (h *CropMasterHandler) ListStages(c *gin.Context) {
	cropID := c.Param("crop_id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	h.logger.Info("Listing stages", zap.String("crop_id", cropID))

	// Call service
	result, err := h.cropService.ListStagesByCrop(c.Request.Context(), cropID)
	if err != nil {
		h.logger.Error("Failed to list stages", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}

// GetStage handles getting a specific stage
// @Summary Get stage by ID
// @Description Get a specific stage by its ID
// @Tags Crop Stages
// @Produce json
// @Param id path string true "Stage ID"
// @Success 200 {object} responses.CropStageResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/stages/{id} [get]
func (h *CropMasterHandler) GetStage(c *gin.Context) {
	stageID := c.Param("id")
	if stageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stage ID is required"})
		return
	}

	h.logger.Info("Getting stage", zap.String("stage_id", stageID))

	// Call service
	result, err := h.cropService.GetStage(c.Request.Context(), stageID)
	if err != nil {
		h.logger.Error("Failed to get stage", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}

// UpdateStage handles stage updates
// @Summary Update stage
// @Description Update an existing stage
// @Tags Crop Stages
// @Accept json
// @Produce json
// @Param id path string true "Stage ID"
// @Param request body requests.UpdateStageRequest true "Update stage request"
// @Success 200 {object} responses.CropStageResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/stages/{id} [put]
func (h *CropMasterHandler) UpdateStage(c *gin.Context) {
	stageID := c.Param("id")
	if stageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stage ID is required"})
		return
	}

	var req requests.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update stage request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set request metadata
	req.StageID = stageID
	req.UserID = c.GetString("aaa_subject")
	req.OrgID = c.GetString("aaa_org")
	req.RequestID = c.GetString("request_id")

	h.logger.Info("Updating stage", zap.String("stage_id", stageID))

	// Call service
	result, err := h.cropService.UpdateStage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update stage", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = req.RequestID
	c.JSON(http.StatusOK, result)
}

// DeleteStage handles stage deletion
// @Summary Delete stage
// @Description Delete a stage
// @Tags Crop Stages
// @Param id path string true "Stage ID"
// @Success 204
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/stages/{id} [delete]
func (h *CropMasterHandler) DeleteStage(c *gin.Context) {
	stageID := c.Param("id")
	if stageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stage ID is required"})
		return
	}

	h.logger.Info("Deleting stage", zap.String("stage_id", stageID))

	// Call service
	err := h.cropService.DeleteStage(c.Request.Context(), stageID)
	if err != nil {
		h.logger.Error("Failed to delete stage", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetLookupData handles lookup data requests
// @Summary Get lookup data
// @Description Get lookup data for dropdowns (categories, units, seasons)
// @Tags Lookups
// @Produce json
// @Param type query string true "Lookup type (categories, units, seasons)"
// @Success 200 {object} responses.LookupResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /lookups/crop-data [get]
func (h *CropMasterHandler) GetLookupData(c *gin.Context) {
	lookupType := c.Query("type")
	if lookupType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lookup type is required"})
		return
	}

	h.logger.Info("Getting lookup data", zap.String("type", lookupType))

	// Call service
	result, err := h.cropService.GetLookupData(c.Request.Context(), lookupType)
	if err != nil {
		h.logger.Error("Failed to get lookup data", zap.Error(err))
		handleServiceError(c, err)
		return
	}

	result.RequestID = c.GetString("request_id")
	c.JSON(http.StatusOK, result)
}
