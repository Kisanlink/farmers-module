package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// StartCycle handles starting a new crop cycle
// @Summary Start a new crop cycle
// @Description Start a new crop cycle with farm and crop details
// @Tags Crop Cycles
// @Accept json
// @Produce json
// @Param request body requests.StartCycleRequest true "Start cycle request"
// @Success 201 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /crops/cycles [post]
func StartCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.StartCycleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Validate request
		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			return
		}

		// Call service
		result, err := service.StartCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

// UpdateCycle handles updating an existing crop cycle
// @Summary Update a crop cycle
// @Description Update details of an existing crop cycle
// @Tags Crop Cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Cycle ID"
// @Param request body requests.UpdateCycleRequest true "Update cycle request"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /crops/cycles/{cycle_id} [put]
func UpdateCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateCycleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Get cycle ID from path
		req.ID = c.Param("cycle_id")

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Validate request
		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			return
		}

		// Call service
		result, err := service.UpdateCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// EndCycle handles ending a crop cycle
// @Summary End a crop cycle
// @Description End a crop cycle and mark it as completed
// @Tags Crop Cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Cycle ID"
// @Param request body requests.EndCycleRequest true "End cycle request"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /crops/cycles/{cycle_id}/end [put]
func EndCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.EndCycleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Get cycle ID from path
		req.ID = c.Param("cycle_id")

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Call service (EndCycleRequest doesn't have Validate method)
		result, err := service.EndCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// ListCycles handles listing crop cycles with filtering
// @Summary List crop cycles
// @Description Get a paginated list of crop cycles with optional filtering
// @Tags Crop Cycles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param farm_id query string false "Filter by farm ID"
// @Param status query string false "Filter by status"
// @Param season query string false "Filter by season"
// @Success 200 {object} responses.SwaggerCropCycleListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /crops/cycles [get]
func ListCycles(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := requests.NewListCyclesRequest()

		// Extract query parameters
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid query parameters", err.Error()))
			return
		}

		// Set context information
		req.RequestID = c.GetString("request_id")
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}
		req.UserID = c.GetString("user_id")
		req.OrgID = c.GetString("org_id")

		// Set defaults
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 || req.PageSize > 100 {
			req.PageSize = 10
		}

		// Call service
		result, err := service.ListCycles(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetCropCycle handles getting a specific crop cycle by ID
// @Summary Get crop cycle by ID
// @Description Get detailed information about a specific crop cycle
// @Tags Crop Cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Cycle ID"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Security BearerAuth
// @Router /crops/cycles/{cycle_id} [get]
func GetCropCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")
		if cycleID == "" {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Missing cycle ID", "cycle_id is required"))
			return
		}

		// Call service
		result, err := service.GetCropCycle(c.Request.Context(), cycleID)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		// Set request ID if response supports it
		if respWithReqID, ok := result.(interface{ SetRequestID(string) }); ok {
			requestID := c.GetString("request_id")
			if requestID == "" {
				requestID = generateRequestID()
			}
			respWithReqID.SetRequestID(requestID)
		}

		c.JSON(http.StatusOK, result)
	}
}
