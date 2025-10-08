package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// Crop Handlers

// CreateCrop handles crop creation
// @Summary Create a new crop
// @Description Create a new crop with master data
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param crop body requests.CreateCropRequest true "Crop data"
// @Success 201 {object} responses.SwaggerCropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Router /crops [post]
func CreateCrop(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get validated data from context (set by validation middleware)
		validatedData, exists := c.Get("validated_crop_data")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validated data not found in context"})
			return
		}

		// Convert validated data to request struct
		var req requests.CreateCropRequest
		if err := convertValidatedData(validatedData, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.CreateCrop(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// ListCrops handles crop listing with filtering
// @Summary List crops
// @Description List crops with optional filtering by category, season, etc.
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param season query string false "Filter by season"
// @Param search query string false "Search in name or scientific name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} responses.SwaggerCropListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Router /crops [get]
func ListCrops(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := requests.NewListCropsRequest()

		// Parse query parameters
		if category := c.Query("category"); category != "" {
			req.Category = category
		}
		if season := c.Query("season"); season != "" {
			req.Season = season
		}
		if search := c.Query("search"); search != "" {
			req.Search = search
		}
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
				req.Page = page
			}
		}
		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
				req.PageSize = pageSize
			}
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.ListCrops(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropListResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCrop handles getting a crop by ID
// @Summary Get crop by ID
// @Description Get detailed information about a specific crop
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Success 200 {object} responses.SwaggerCropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [get]
func GetCrop(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cropID := c.Param("id")
		if cropID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crop ID is required"})
			return
		}

		req := requests.NewGetCropRequest()
		req.ID = cropID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.GetCrop(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateCrop handles crop updates
// @Summary Update crop
// @Description Update crop master data
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Param crop body requests.UpdateCropRequest true "Crop update data"
// @Success 200 {object} responses.SwaggerCropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [put]
func UpdateCrop(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cropID := c.Param("id")
		if cropID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crop ID is required"})
			return
		}

		// Get validated data from context (set by validation middleware)
		validatedData, exists := c.Get("validated_crop_data")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validated data not found in context"})
			return
		}

		// Convert validated data to request struct
		var req requests.UpdateCropRequest
		if err := convertValidatedData(validatedData, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set ID from path parameter
		req.ID = cropID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.UpdateCrop(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// DeleteCrop handles crop deletion
// @Summary Delete crop
// @Description Soft delete a crop (marks as inactive)
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop ID"
// @Success 200 {object} responses.SwaggerCropResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /crops/{id} [delete]
func DeleteCrop(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cropID := c.Param("id")
		if cropID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crop ID is required"})
			return
		}

		req := requests.NewDeleteCropRequest()
		req.ID = cropID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.DeleteCrop(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// Crop Variety Handlers

// CreateCropVariety handles crop variety creation
// @Summary Create crop variety
// @Description Create a new variety for a specific crop
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param variety body requests.CreateCropVarietyRequest true "Crop variety data"
// @Success 201 {object} responses.SwaggerCropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Router /varieties [post]
func CreateCropVariety(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get validated data from context (set by validation middleware)
		validatedData, exists := c.Get("validated_crop_variety_data")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validated data not found in context"})
			return
		}

		// Convert validated data to request struct
		var req requests.CreateCropVarietyRequest
		if err := convertValidatedData(validatedData, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.CreateCropVariety(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// ListCropVarieties handles crop variety listing
// @Summary List crop varieties
// @Description List crop varieties with optional filtering
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param crop_id query string false "Filter by crop ID"
// @Param search query string false "Search in name or description"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} responses.CropVarietyListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Router /varieties [get]
func ListCropVarieties(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := requests.NewListCropVarietiesRequest()

		// Parse query parameters
		if cropID := c.Query("crop_id"); cropID != "" {
			req.CropID = cropID
		}
		if search := c.Query("search"); search != "" {
			req.Search = search
		}
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
				req.Page = page
			}
		}
		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
				req.PageSize = pageSize
			}
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.ListCropVarieties(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyListResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCropVariety handles getting a crop variety by ID
// @Summary Get crop variety by ID
// @Description Get detailed information about a specific crop variety
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop variety ID"
// @Success 200 {object} responses.SwaggerCropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /varieties/{id} [get]
func GetCropVariety(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		varietyID := c.Param("id")
		if varietyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "variety ID is required"})
			return
		}

		req := requests.NewGetCropVarietyRequest()
		req.ID = varietyID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.GetCropVariety(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateCropVariety handles crop variety updates
// @Summary Update crop variety
// @Description Update crop variety data
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop variety ID"
// @Param variety body requests.UpdateCropVarietyRequest true "Crop variety update data"
// @Success 200 {object} responses.SwaggerCropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /varieties/{id} [put]
func UpdateCropVariety(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		varietyID := c.Param("id")
		if varietyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "variety ID is required"})
			return
		}

		// Get validated data from context (set by validation middleware)
		validatedData, exists := c.Get("validated_crop_variety_data")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validated data not found in context"})
			return
		}

		// Convert validated data to request struct
		var req requests.UpdateCropVarietyRequest
		if err := convertValidatedData(validatedData, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set ID from path parameter
		req.ID = varietyID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.UpdateCropVariety(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// DeleteCropVariety handles crop variety deletion
// @Summary Delete crop variety
// @Description Soft delete a crop variety (marks as inactive)
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param id path string true "Crop variety ID"
// @Success 200 {object} responses.SwaggerCropVarietyResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Router /varieties/{id} [delete]
func DeleteCropVariety(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		varietyID := c.Param("id")
		if varietyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "variety ID is required"})
			return
		}

		req := requests.NewDeleteCropVarietyRequest()
		req.ID = varietyID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service
		result, err := service.DeleteCropVariety(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// Lookup/Dropdown Handlers

// GetCropLookupData handles crop lookup data for dropdowns
// @Summary Get crop lookup data
// @Description Get simplified crop data for dropdown/lookup purposes
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param season query string false "Filter by season"
// @Success 200 {object} responses.SwaggerCropLookupResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /lookups/crops [get]
func GetCropLookupData(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := requests.NewGetCropLookupRequest()

		// Parse query parameters
		if category := c.Query("category"); category != "" {
			req.Category = category
		}
		if season := c.Query("season"); season != "" {
			req.Season = season
		}

		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		result, err := service.GetCropLookupData(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropLookupResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetVarietyLookupData handles crop variety lookup data
// @Summary Get variety lookup data
// @Description Get simplified variety data for dropdown/lookup purposes
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Param crop_id path string true "Crop ID"
// @Success 200 {object} responses.CropVarietyLookupResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /lookups/varieties/{crop_id} [get]
func GetVarietyLookupData(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cropID := c.Param("crop_id")
		if cropID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "crop ID is required"})
			return
		}

		req := requests.NewGetVarietyLookupRequest()
		req.CropID = cropID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		result, err := service.GetVarietyLookupData(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropVarietyLookupResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCropCategories handles getting available crop categories
// @Summary Get crop categories
// @Description Get list of available crop categories
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Success 200 {object} responses.CropCategoriesResponse
// @Router /lookups/crop-categories [get]
func GetCropCategories(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call service
		result, err := service.GetCropCategories(c.Request.Context())
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropCategoriesResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCropSeasons handles getting available crop seasons
// @Summary Get crop seasons
// @Description Get list of available crop seasons
// @Tags Crop Master Data
// @Accept json
// @Produce json
// @Success 200 {object} responses.CropSeasonsResponse
// @Router /lookups/crop-seasons [get]
func GetCropSeasons(service services.CropService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Call service
		result, err := service.GetCropSeasons(c.Request.Context())
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.CropSeasonsResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
