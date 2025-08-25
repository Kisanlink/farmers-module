package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// Crop Cycle Handlers (W10-W13)

// StartCycle handles W10: Start crop cycle
// @Summary Start a new crop cycle
// @Description Start a new crop cycle for a specific farm
// @Tags crop-cycles
// @Accept json
// @Produce json
// @Param cycle body requests.StartCycleRequest true "Crop cycle data"
// @Success 201 {object} responses.CropCycleResponse
// @Failure 400 {object} responses.BaseError
// @Failure 401 {object} responses.BaseError
// @Failure 403 {object} responses.BaseError
// @Failure 500 {object} responses.BaseError
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
// @Success 200 {object} responses.CropCycleResponse
// @Failure 400 {object} responses.BaseError
// @Failure 401 {object} responses.BaseError
// @Failure 403 {object} responses.BaseError
// @Failure 404 {object} responses.BaseError
// @Failure 500 {object} responses.BaseError
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
// @Success 200 {object} responses.CropCycleResponse
// @Failure 400 {object} responses.BaseError
// @Failure 401 {object} responses.BaseError
// @Failure 403 {object} responses.BaseError
// @Failure 404 {object} responses.BaseError
// @Failure 500 {object} responses.BaseError
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
// @Success 200 {object} responses.CropCycleListResponse
// @Failure 400 {object} responses.BaseError
// @Failure 401 {object} responses.BaseError
// @Failure 403 {object} responses.BaseError
// @Failure 500 {object} responses.BaseError
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
// @Success 200 {object} responses.CropCycleResponse
// @Failure 400 {object} responses.BaseError
// @Failure 401 {object} responses.BaseError
// @Failure 403 {object} responses.BaseError
// @Failure 404 {object} responses.BaseError
// @Failure 500 {object} responses.BaseError
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

// CreateActivity handles W14: Create farm activity
// @Summary Create a new farm activity
// @Description Create a new farm activity for a crop cycle
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity body object true "Farm activity data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /crops/activities [post]
func CreateActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CropCycleID  string    `json:"crop_cycle_id" binding:"required"`
			ActivityType string    `json:"activity_type" binding:"required"`
			PlannedAt    time.Time `json:"planned_at" binding:"required"`
			Description  string    `json:"description"`
			Metadata     string    `json:"metadata"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm activity created successfully",
			"data": gin.H{
				"crop_cycle_id": req.CropCycleID,
				"activity_type": req.ActivityType,
				"planned_at":    req.PlannedAt,
			},
		})
	}
}

// CompleteActivity handles W15: Complete farm activity
// @Summary Complete a farm activity
// @Description Mark a farm activity as completed
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Param activity body object true "Activity completion data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /crops/activities/{activity_id}/complete [post]
func CompleteActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")
		var req struct {
			CompletedAt time.Time `json:"completed_at" binding:"required"`
			Notes       *string   `json:"notes,omitempty"`
			Outcome     *string   `json:"outcome,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm activity completed successfully",
			"data": gin.H{
				"activity_id":  activityID,
				"completed_at": req.CompletedAt,
			},
		})
	}
}

// UpdateActivity handles W16: Update farm activity
// @Summary Update a farm activity
// @Description Update an existing farm activity details
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Param activity body object true "Activity update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /crops/activities/{activity_id} [put]
func UpdateActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")
		var req struct {
			ActivityType *string    `json:"activity_type,omitempty"`
			PlannedAt    *time.Time `json:"planned_at,omitempty"`
			Metadata     *string    `json:"metadata,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm activity updated successfully",
			"data": gin.H{
				"activity_id": activityID,
			},
		})
	}
}

// ListActivities handles W17: List farm activities
// @Summary List farm activities
// @Description Retrieve a list of all farm activities
// @Tags farm-activities
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /crops/activities [get]
func ListActivities(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm activities retrieved successfully",
			"data":    []interface{}{},
		})
	}
}

// GetFarmActivity handles getting farm activity by ID
// @Summary Get farm activity by ID
// @Description Retrieve a specific farm activity by its ID
// @Tags farm-activities
// @Accept json
// @Produce json
// @Param activity_id path string true "Activity ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /crops/activities/{activity_id} [get]
func GetCropFarmActivity(service services.FarmActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID := c.Param("activity_id")

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm activity retrieved successfully",
			"data": gin.H{
				"activity_id": activityID,
			},
		})
	}
}
