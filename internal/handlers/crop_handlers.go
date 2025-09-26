package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// CropCycleResponse represents a simple crop cycle response
type CropCycleResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// CropCycleListResponse represents a simple crop cycle list response
type CropCycleListResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Data      []interface{} `json:"data"`
	Page      int           `json:"page"`
	PageSize  int           `json:"page_size"`
	Total     int           `json:"total"`
}

// Crop Cycle Handlers (W10-W13)

// StartCycle handles W10: Start crop cycle
// @Summary Start a new crop cycle
// @Description Start a new crop cycle for a specific farm
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param cycle body requests.StartCycleRequest true "Crop cycle data"
// @Success 201 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/cycles [post]
func StartCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.StartCycleRequest

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
		req.SetRequestType("start_cycle")

		// Call service
		result, err := service.StartCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response := result.(responses.CropCycleResponse)
		response.SetRequestID(req.RequestID)
		c.JSON(http.StatusCreated, response)
	}
}

// UpdateCycle handles W11: Update crop cycle
// @Summary Update a crop cycle
// @Description Update an existing crop cycle details
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Crop Cycle ID"
// @Param cycle body requests.UpdateCycleRequest true "Crop cycle update data"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/cycles/{cycle_id} [put]
func UpdateCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")
		var req requests.UpdateCycleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set cycle ID from path parameter
		req.ID = cycleID

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set request metadata
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("update_cycle")

		// Call service
		result, err := service.UpdateCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response := result.(responses.CropCycleResponse)
		response.SetRequestID(req.RequestID)
		c.JSON(http.StatusOK, response)
	}
}

// EndCycle handles W12: End crop cycle
// @Summary End a crop cycle
// @Description End an active crop cycle with status and outcome
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Crop Cycle ID"
// @Param cycle body requests.EndCycleRequest true "Crop cycle end data"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/cycles/{cycle_id}/end [post]
func EndCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")
		var req requests.EndCycleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Invalid request data", err.Error()))
			return
		}

		// Set cycle ID from path parameter
		req.ID = cycleID

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set request metadata
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("end_cycle")

		// Call service
		result, err := service.EndCycle(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response := result.(responses.CropCycleResponse)
		response.SetRequestID(req.RequestID)
		c.JSON(http.StatusOK, response)
	}
}

// ListCycles handles W13: List crop cycles
// @Summary List crop cycles
// @Description Retrieve a list of all crop cycles with filtering
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param farm_id query string false "Filter by farm ID"
// @Param farmer_id query string false "Filter by farmer ID"
// @Param season query string false "Filter by season" Enums(RABI, KHARIF, ZAID)
// @Param status query string false "Filter by status" Enums(PLANNED, ACTIVE, COMPLETED, CANCELLED)
// @Success 200 {object} responses.SwaggerCropCycleListResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/cycles [get]
func ListCycles(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.ListCyclesRequest

		// Parse query parameters
		req.Page = parseIntQuery(c, "page", 1)
		req.PageSize = parseIntQuery(c, "page_size", 10)
		req.FarmID = c.Query("farm_id")
		req.FarmerID = c.Query("farmer_id")
		req.Season = c.Query("season")
		req.Status = c.Query("status")

		// Set user context from middleware
		if userID, exists := c.Get("aaa_subject"); exists {
			req.UserID = userID.(string)
		}
		if orgID, exists := c.Get("aaa_org"); exists {
			req.OrgID = orgID.(string)
		}

		// Set request metadata
		req.SetRequestID(c.GetString("request_id"))
		req.SetRequestType("list_cycles")

		// Call service
		result, err := service.ListCycles(c.Request.Context(), &req)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response := result.(responses.CropCycleListResponse)
		response.SetRequestID(req.RequestID)
		c.JSON(http.StatusOK, response)
	}
}

// GetCropCycle handles getting crop cycle by ID
// @Summary Get crop cycle by ID
// @Description Retrieve a specific crop cycle by its ID
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param cycle_id path string true "Crop Cycle ID"
// @Success 200 {object} responses.SwaggerCropCycleResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /crops/cycles/{cycle_id} [get]
func GetCropCycle(service services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")

		if cycleID == "" {
			c.JSON(http.StatusBadRequest, responses.NewValidationError("Cycle ID is required", "cycle_id parameter is missing"))
			return
		}

		// Call service
		result, err := service.GetCropCycle(c.Request.Context(), cycleID)
		if err != nil {
			handleServiceError(c, err)
			return
		}

		response := result.(responses.CropCycleResponse)
		response.SetRequestID(c.GetString("request_id"))
		c.JSON(http.StatusOK, response)
	}
}

// Farm Activity Handlers (W14-W17)

// CreateActivityRequest represents a request to create a farm activity
type CreateActivityRequest struct {
	CropCycleID  string    `json:"crop_cycle_id" binding:"required"`
	ActivityType string    `json:"activity_type" binding:"required"`
	PlannedAt    time.Time `json:"planned_at" binding:"required"`
	Description  string    `json:"description"`
	Metadata     string    `json:"metadata"`
}

// CreateActivityResponse represents a response for creating a farm activity
type CreateActivityResponse struct {
	Message string             `json:"message"`
	Data    CreateActivityData `json:"data"`
}

// CreateActivityData represents the data returned when creating a farm activity
type CreateActivityData struct {
	CropCycleID  string    `json:"crop_cycle_id"`
	ActivityType string    `json:"activity_type"`
	PlannedAt    time.Time `json:"planned_at"`
}

// CreateActivity handles W14: Create farm activity
// @Summary Create a new farm activity
// @Description Create a new farm activity for a crop cycle
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity body CreateActivityRequest true "Farm activity data"
// @Success 200 {object} CreateActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /crops/activities [post]
func CreateActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateActivityRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse{
				Error:         "Invalid request format",
				Message:       err.Error(),
				Code:          "INVALID_REQUEST",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		// Set request metadata
		// Note: BaseRequest fields should be set automatically

		// Call service
		result, err := service.CreateActivity(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusCreated, response)
		} else {
			c.JSON(http.StatusCreated, CreateActivityResponse{
				Message: "Farm activity created successfully",
				Data: CreateActivityData{
					CropCycleID:  req.CropCycleID,
					ActivityType: req.ActivityType,
					PlannedAt:    req.PlannedAt,
				},
			})
		}
	}
}

// CompleteActivityRequest represents a request to complete a farm activity
type CompleteActivityRequest struct {
	CompletedAt time.Time `json:"completed_at" binding:"required"`
	Notes       *string   `json:"notes,omitempty"`
	Outcome     *string   `json:"outcome,omitempty"`
}

// CompleteActivityResponse represents a response for completing a farm activity
type CompleteActivityResponse struct {
	Message string               `json:"message"`
	Data    CompleteActivityData `json:"data"`
}

// CompleteActivityData represents the data returned when completing a farm activity
type CompleteActivityData struct {
	ActivityID  string    `json:"activity_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// CompleteActivity handles W15: Complete farm activity
// @Summary Complete a farm activity
// @Description Mark a farm activity as completed
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Param activity body CompleteActivityRequest true "Activity completion data"
// @Success 200 {object} CompleteActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /crops/activities/{activity_id}/complete [post]
func CompleteActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")
		var req CompleteActivityRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse{
				Error:         "Invalid request format",
				Message:       err.Error(),
				Code:          "INVALID_REQUEST",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		// Set request metadata
		// Note: ID should be set from URL parameter
		// Note: BaseRequest fields should be set automatically

		// Call service
		result, err := service.CompleteActivity(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, CompleteActivityResponse{
				Message: "Farm activity completed successfully",
				Data: CompleteActivityData{
					ActivityID:  activityID,
					CompletedAt: req.CompletedAt,
				},
			})
		}
	}
}

// UpdateActivityRequest represents a request to update a farm activity
type UpdateActivityRequest struct {
	ActivityType *string    `json:"activity_type,omitempty"`
	PlannedAt    *time.Time `json:"planned_at,omitempty"`
	Metadata     *string    `json:"metadata,omitempty"`
}

// UpdateActivityResponse represents a response for updating a farm activity
type UpdateActivityResponse struct {
	Message string             `json:"message"`
	Data    UpdateActivityData `json:"data"`
}

// UpdateActivityData represents the data returned when updating a farm activity
type UpdateActivityData struct {
	ActivityID string `json:"activity_id"`
}

// UpdateActivity handles W16: Update farm activity
// @Summary Update a farm activity
// @Description Update an existing farm activity details
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Param activity body UpdateActivityRequest true "Activity update data"
// @Success 200 {object} UpdateActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /crops/activities/{activity_id} [put]
func UpdateActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")
		var req UpdateActivityRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse{
				Error:         "Invalid request format",
				Message:       err.Error(),
				Code:          "INVALID_REQUEST",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		// Set request metadata
		// Note: ID should be set from URL parameter
		// Note: BaseRequest fields should be set automatically

		// Call service
		result, err := service.UpdateActivity(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, UpdateActivityResponse{
				Message: "Farm activity updated successfully",
				Data: UpdateActivityData{
					ActivityID: activityID,
				},
			})
		}
	}
}

// ListActivitiesResponse represents a response for listing farm activities
type ListActivitiesResponse struct {
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

// ListActivities handles W17: List farm activities
// @Summary List farm activities
// @Description Retrieve a list of all farm activities
// @Tags farm-activities
// @Accept json
// @Produce json
// @Success 200 {object} ListActivitiesResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /crops/activities [get]
func ListActivities(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create request with query parameters
		req := &requests.ListActivitiesRequest{}

		// Parse query parameters
		if cropCycleID := c.Query("crop_cycle_id"); cropCycleID != "" {
			req.CropCycleID = cropCycleID
		}
		if activityType := c.Query("activity_type"); activityType != "" {
			req.ActivityType = activityType
		}
		if status := c.Query("status"); status != "" {
			req.Status = status
		}

		// Call service
		result, err := service.ListActivities(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.FarmActivityListResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, ListActivitiesResponse{
				Message: "Farm activities retrieved successfully",
				Data:    []interface{}{},
			})
		}
	}
}

// GetFarmActivityResponse represents a response for getting a farm activity
type GetFarmActivityResponse struct {
	Message string              `json:"message"`
	Data    GetFarmActivityData `json:"data"`
}

// GetFarmActivityData represents the data returned when getting a farm activity
type GetFarmActivityData struct {
	ActivityID string `json:"activity_id"`
}

// GetFarmActivity handles getting farm activity by ID
// @Summary Get farm activity by ID
// @Description Retrieve a specific farm activity by its ID
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Success 200 {object} GetFarmActivityResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Router /crops/activities/{activity_id} [get]
func GetCropFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")

		// Call service
		result, err := service.GetFarmActivity(c.Request.Context(), activityID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.FarmActivityResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, GetFarmActivityResponse{
				Message: "Farm activity retrieved successfully",
				Data: GetFarmActivityData{
					ActivityID: activityID,
				},
			})
		}
	}
}

// RecordHarvest handles harvest recording for crop cycles
func RecordHarvest(cropCycleService services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")
		if cycleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cycle ID is required"})
			return
		}

		var req requests.RecordHarvestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Set request metadata
		req.CycleID = cycleID
		// Note: BaseRequest fields should be set automatically

		// Call service
		result, err := cropCycleService.RecordHarvest(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.CropCycleResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, CropCycleResponse{
				Success:   true,
				Message:   "Harvest recorded successfully",
				RequestID: req.RequestID,
				Data:      result,
			})
		}
	}
}

// UploadReport handles report upload for crop cycles
func UploadReport(cropCycleService services.CropCycleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cycleID := c.Param("cycle_id")
		if cycleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cycle ID is required"})
			return
		}

		var req requests.UploadReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Set request metadata
		req.CycleID = cycleID
		// Note: BaseRequest fields should be set automatically

		// Call service
		result, err := cropCycleService.UploadReport(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if response, ok := result.(*responses.CropCycleResponse); ok {
			response.RequestID = c.GetString("request_id")
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusOK, CropCycleResponse{
				Success:   true,
				Message:   "Report uploaded successfully",
				RequestID: req.RequestID,
				Data:      result,
			})
		}
	}
}
