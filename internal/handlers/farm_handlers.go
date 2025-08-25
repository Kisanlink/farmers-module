package handlers

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// CreateFarm handles W6: Create farm
// @Summary Create a new farm
// @Description Create a new farm with geographic boundaries and metadata
// @Tags farms
// @Accept json
// @Produce json
// @Param farm body requests.CreateFarmRequest true "Farm data"
// @Success 201 {object} responses.FarmResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Router /farms [post]
func CreateFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CreateFarmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set request ID for tracking
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		result, err := service.CreateFarm(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.FarmResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// UpdateFarm handles W7: Update farm
// @Summary Update an existing farm
// @Description Update farm details including name, geometry, and location
// @Tags farms
// @Accept json
// @Produce json
// @Param farm_id path string true "Farm ID"
// @Param farm body requests.UpdateFarmRequest true "Farm update data"
// @Success 200 {object} responses.FarmResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /farms/{farm_id} [put]
func UpdateFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")
		var req requests.UpdateFarmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set farm ID from path parameter
		req.ID = farmID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		result, err := service.UpdateFarm(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.FarmResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// DeleteFarm handles W8: Delete farm
// @Summary Delete a farm
// @Description Delete a farm by ID
// @Tags farms
// @Accept json
// @Produce json
// @Param farm_id path string true "Farm ID"
// @Success 204 "Farm deleted successfully"
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /farms/{farm_id} [delete]
func DeleteFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")

		req := requests.NewDeleteFarmRequest()
		req.ID = farmID
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		err := service.DeleteFarm(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// ListFarms handles W9: List farms
// @Summary List all farms
// @Description Retrieve a list of all farms with optional filtering
// @Tags farms
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param farmer_id query string false "Filter by farmer ID"
// @Param org_id query string false "Filter by organization ID"
// @Param min_area query number false "Minimum area in hectares"
// @Param max_area query number false "Maximum area in hectares"
// @Success 200 {object} responses.FarmListResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Router /farms [get]
func ListFarms(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := requests.NewListFarmsRequest()

		// Parse query parameters
		if page := c.Query("page"); page != "" {
			if p, err := strconv.Atoi(page); err == nil && p > 0 {
				req.Page = p
			}
		}
		if pageSize := c.Query("page_size"); pageSize != "" {
			if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
				req.PageSize = ps
			}
		}
		if farmerID := c.Query("farmer_id"); farmerID != "" {
			req.AAAFarmerUserID = farmerID
		}
		if orgID := c.Query("org_id"); orgID != "" {
			req.AAAOrgID = orgID
		}
		if minArea := c.Query("min_area"); minArea != "" {
			if ma, err := strconv.ParseFloat(minArea, 64); err == nil {
				req.MinArea = &ma
			}
		}
		if maxArea := c.Query("max_area"); maxArea != "" {
			if ma, err := strconv.ParseFloat(maxArea, 64); err == nil {
				req.MaxArea = &ma
			}
		}

		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call service
		result, err := service.ListFarms(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.FarmListResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetFarm handles getting farm by ID
// @Summary Get farm by ID
// @Description Retrieve a specific farm by its ID
// @Tags farms
// @Accept json
// @Produce json
// @Param farm_id path string true "Farm ID"
// @Success 200 {object} responses.FarmResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /farms/{farm_id} [get]
func GetFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")

		// Call service
		result, err := service.GetFarm(c.Request.Context(), farmID)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response, ok := result.(*responses.FarmResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response type"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// Helper functions
