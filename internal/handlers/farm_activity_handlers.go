package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// FarmActivityResponse represents a simple farm activity response
type FarmActivityResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// FarmActivityListResponse represents a simple farm activity list response
type FarmActivityListResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Data      []interface{} `json:"data"`
	Page      int           `json:"page"`
	PageSize  int           `json:"page_size"`
	Total     int           `json:"total"`
}

// CreateFarmActivity handles W14: Create farm activity
// @Summary Create a new farm activity
// @Description Create a new farm activity within a crop cycle
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity body requests.CreateActivityRequest true "Farm activity data"
// @Success 201 {object} responses.SwaggerFarmActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /activities [post]
func CreateFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CreateActivityRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set request metadata
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("create_activity")

		// Call service
		result, err := service.CreateActivity(c.Request.Context(), &req)
		if err != nil {
			if isValidationError(err) {
				c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			} else if isPermissionError(err) {
				c.JSON(http.StatusForbidden, responses.NewForbiddenError("Permission denied"))
			} else if isNotFoundError(err) {
				c.JSON(http.StatusNotFound, responses.NewNotFoundError("Resource not found", err.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, responses.NewInternalServerError("Internal server error", err.Error()))
			}
			return
		}

		// Set response metadata
		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.SetRequestID(c.GetString("request_id"))
		}

		c.JSON(http.StatusCreated, result)
	}
}

// CompleteFarmActivity handles W15: Complete farm activity
// @Summary Complete a farm activity
// @Description Mark a farm activity as completed with output data
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param activity body requests.CompleteActivityRequest true "Complete activity data"
// @Success 200 {object} responses.SwaggerFarmActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /activities/{id}/complete [put]
func CompleteFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CompleteActivityRequest

		// Get activity ID from path
		activityID := c.Param("id")
		if activityID == "" {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Activity ID is required", "Missing activity ID in path"))
			return
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set activity ID and request metadata
		req.ID = activityID
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("complete_activity")

		// Call service
		result, err := service.CompleteActivity(c.Request.Context(), &req)
		if err != nil {
			if isValidationError(err) {
				c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			} else if isPermissionError(err) {
				c.JSON(http.StatusForbidden, responses.NewForbiddenError("Permission denied"))
			} else if isNotFoundError(err) {
				c.JSON(http.StatusNotFound, responses.NewNotFoundError("Resource not found", err.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, responses.NewInternalServerError("Internal server error", err.Error()))
			}
			return
		}

		// Set response metadata
		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.SetRequestID(c.GetString("request_id"))
		}

		c.JSON(http.StatusOK, result)
	}
}

// UpdateFarmActivity handles W16: Update farm activity
// @Summary Update a farm activity
// @Description Update farm activity details (only for non-completed activities)
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Param activity body requests.UpdateActivityRequest true "Update activity data"
// @Success 200 {object} responses.SwaggerFarmActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /activities/{id} [put]
func UpdateFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateActivityRequest

		// Get activity ID from path
		activityID := c.Param("id")
		if activityID == "" {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Activity ID is required", "Missing activity ID in path"))
			return
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set activity ID and request metadata
		req.ID = activityID
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("update_activity")

		// Call service
		result, err := service.UpdateActivity(c.Request.Context(), &req)
		if err != nil {
			if isValidationError(err) {
				c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			} else if isPermissionError(err) {
				c.JSON(http.StatusForbidden, responses.NewForbiddenError("Permission denied"))
			} else if isNotFoundError(err) {
				c.JSON(http.StatusNotFound, responses.NewNotFoundError("Resource not found", err.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, responses.NewInternalServerError("Internal server error", err.Error()))
			}
			return
		}

		// Set response metadata
		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.SetRequestID(c.GetString("request_id"))
		}

		c.JSON(http.StatusOK, result)
	}
}

// ListFarmActivities handles W17: List farm activities
// @Summary List farm activities
// @Description Get a paginated list of farm activities with optional filtering
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param crop_cycle_id query string false "Filter by crop cycle ID"
// @Param activity_type query string false "Filter by activity type"
// @Param status query string false "Filter by status"
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} FarmActivityListResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /activities [get]
func ListFarmActivities(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		req := requests.ListActivitiesRequest{
			CropCycleID:  c.Query("crop_cycle_id"),
			ActivityType: c.Query("activity_type"),
			Status:       c.Query("status"),
			DateFrom:     c.Query("date_from"),
			DateTo:       c.Query("date_to"),
			Page:         parseIntQuery(c, "page", 1),
			PageSize:     parseIntQuery(c, "page_size", 10),
		}

		// Ensure page size is within limits
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set request metadata
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("list_activities")

		// Call service
		result, err := service.ListActivities(c.Request.Context(), &req)
		if err != nil {
			if isValidationError(err) {
				c.JSON(http.StatusBadRequest, responses.NewValidationError("Validation failed", err.Error()))
			} else if isPermissionError(err) {
				c.JSON(http.StatusForbidden, responses.NewForbiddenError("Permission denied"))
			} else {
				c.JSON(http.StatusInternalServerError, responses.NewInternalServerError("Internal server error", err.Error()))
			}
			return
		}

		// Set response metadata
		if response, ok := result.(*responses.FarmActivityListResponse); ok {
			response.SetRequestID(c.GetString("request_id"))
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetFarmActivity handles getting a farm activity by ID
// @Summary Get a farm activity by ID
// @Description Retrieve a specific farm activity by its ID
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID"
// @Success 200 {object} responses.SwaggerFarmActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /activities/{id} [get]
func GetFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get activity ID from path
		activityID := c.Param("id")
		if activityID == "" {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Activity ID is required", "Missing activity ID in path"))
			return
		}

		// Call service
		result, err := service.GetFarmActivity(c.Request.Context(), activityID)
		if err != nil {
			if isNotFoundError(err) {
				c.JSON(http.StatusNotFound, responses.NewNotFoundError("Resource not found", err.Error()))
			} else {
				c.JSON(http.StatusInternalServerError, responses.NewInternalServerError("Internal server error", err.Error()))
			}
			return
		}

		// Set response metadata
		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.SetRequestID(c.GetString("request_id"))
		}

		c.JSON(http.StatusOK, result)
	}
}
