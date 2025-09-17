package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// FarmActivityHandler handles HTTP requests for farm activities.
type FarmActivityHandler struct {
	service services.FarmActivityServiceInterface
}

// NewFarmActivityHandler creates an instance of FarmActivityHandler.
func NewFarmActivityHandler(service services.FarmActivityServiceInterface) *FarmActivityHandler {
	return &FarmActivityHandler{service: service}
}

// CreateFarmActivityRequest represents the expected JSON body for creating a farm activity.
type CreateFarmActivityRequest struct {
	FarmID         string     `json:"farm_id" binding:"required"`
	CropCycleID    string     `json:"crop_cycle_id" binding:"required"`
	Activity       string     `json:"activity"` // Defaults to "sowing" via BeforeCreate if empty
	StartDate      time.Time  `json:"start_date" binding:"required"`
	EndDate        *time.Time `json:"end_date"`
	ActivityReport string     `json:"activity_report"`
}

// CreateActivity handles POST /farm-activities
func (h *FarmActivityHandler) CreateActivity(c *gin.Context) {
	var req CreateFarmActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	activity := &models.FarmActivity{
		FarmID:         req.FarmID,
		CropCycleID:    req.CropCycleID,
		Activity:       entities.ActivityType(req.Activity),
		StartDate:      &req.StartDate,
		EndDate:        req.EndDate,
		ActivityReport: req.ActivityReport,
	}

	if err := h.service.CreateActivity(activity); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Could not create farm activity: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		StatusCode: http.StatusCreated,
		Success:    true,
		Message:    "Farm activity created successfully",
		Data:       activity,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// GetActivities handles GET /farm-activities with various query parameters.
// Supported query parameters:
//   - id: returns a single activity by its ID.
//   - farm_id: returns activities for a given farm.
//   - cycle_id: returns activities for a given crop cycle.
//   - start_date and end_date (with farm_id): returns activities for a farm within the specified date range.
//
// If both farm_id and cycle_id are provided with a date range, the results are filtered by cycle.
func (h *FarmActivityHandler) GetActivities(c *gin.Context) {
	id := c.Query("id")
	farmID := c.Query("farm_id")
	cycleID := c.Query("crop_cycle_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var (
		data interface{}
		err  error
	)

	if id != "" {
		data, err = h.service.GetActivityByID(id)
	} else if (farmID != "" || cycleID != "") && startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse(time.RFC3339, startDateStr)
		endDate, err2 := time.Parse(time.RFC3339, endDateStr)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid date format. Please use RFC3339 format.",
				Data:       nil,
				Error:      "date parsing error",
				TimeStamp:  time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		var activities []*models.FarmActivity

		// Get by farm if provided
		if farmID != "" {
			activities, err = h.service.GetActivitiesByDateRange(farmID, startDate, endDate)
		}

		// Filter by cycleID if also provided
		if err == nil && cycleID != "" {
			var filtered []*models.FarmActivity
			for _, a := range activities {
				if a.CropCycleID == cycleID {
					filtered = append(filtered, a)
				}
			}
			data = filtered
		} else {
			data = activities
		}
	} else if farmID != "" && cycleID != "" {
		activities, err := h.service.GetActivitiesByFarmID(farmID)
		if err == nil {
			var filtered []*models.FarmActivity
			for _, a := range activities {
				if a.CropCycleID == cycleID {
					filtered = append(filtered, a)
				}
			}
			data = filtered
		}
	} else if cycleID != "" {
		data, err = h.service.GetActivitiesByCropCycle(cycleID)
	} else if farmID != "" {
		data, err = h.service.GetActivitiesByFarmID(farmID)
	} else {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Please provide at least one query parameter: id, farm_id, cycle_id, or date range (with farm_id or cycle_id).",
			Data:       nil,
			Error:      "Missing required query parameter",
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to retrieve activity(ies): " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Farm activity(ies) retrieved successfully",
		Data:       data,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// UpdateActivityRequest represents the JSON body for updating a farm activity.
type UpdateActivityRequest struct {
	Activity       string    `json:"activity"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	ActivityReport string    `json:"activity_report"`
}

// UpdateActivity handles PUT /farm-activities/:id
func (h *FarmActivityHandler) UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	var req UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid input: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Retrieve the current activity.
	activity, err := h.service.GetActivityByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Activity not found: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Update fields (only update if new values are provided).
	if req.Activity != "" {
		activity.Activity = entities.ActivityType(req.Activity)
	}
	activity.StartDate = &req.StartDate
	activity.EndDate = &req.EndDate
	activity.ActivityReport = req.ActivityReport

	if err := h.service.UpdateActivity(activity); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to update activity: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Activity updated successfully",
		Data:       activity,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// DeleteActivity handles DELETE /farm-activities/:id
func (h *FarmActivityHandler) DeleteActivity(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteActivity(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to delete activity: " + err.Error(),
			Data:       nil,
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Activity deleted successfully",
		Data:       nil,
		Error:      nil,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// BatchFarmActivitiesRequest represents the request structure for batch farm activities
type BatchFarmActivitiesRequest struct {
	models.BatchRequest
	Filters struct {
		StartDate *string `json:"start_date,omitempty"` // RFC3339 format
		EndDate   *string `json:"end_date,omitempty"`   // RFC3339 format
	} `json:"filters,omitempty"`
}

// GetBatchFarmActivities handles POST /api/v1/batch/farm-activities
func (h *FarmActivityHandler) GetBatchFarmActivities(c *gin.Context) {
	var req BatchFarmActivitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BatchResponse{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request body",
			Data:       make(map[string]interface{}),
			Errors:     map[string]string{"validation": err.Error()},
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	var data map[string]interface{}
	var errors map[string]string

	// Check if date range filters are provided
	if req.Filters.StartDate != nil && req.Filters.EndDate != nil {
		// Parse dates
		startDate, err1 := time.Parse(time.RFC3339, *req.Filters.StartDate)
		endDate, err2 := time.Parse(time.RFC3339, *req.Filters.EndDate)

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, models.BatchResponse{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid date format",
				Data:       make(map[string]interface{}),
				Errors:     map[string]string{"validation": "dates must be in RFC3339 format"},
				TimeStamp:  time.Now().Format(time.RFC3339),
			})
			return
		}

		// Get activities by date range
		data, errors = h.service.GetActivitiesByDateRangeBatch(req.FarmIDs, startDate, endDate)
	} else {
		// Get all activities for farms
		data, errors = h.service.GetActivitiesByFarmIDsBatch(req.FarmIDs)
	}

	success := len(errors) == 0
	message := "Batch farm activities retrieved successfully"
	if len(errors) > 0 && len(data) > 0 {
		message = "Batch farm activities retrieved with some errors"
	} else if len(errors) > 0 {
		message = "Failed to retrieve farm activities for all farms"
	}

	statusCode := http.StatusOK
	if len(errors) > 0 && len(data) > 0 {
		statusCode = http.StatusPartialContent
	} else if len(errors) > 0 && len(data) == 0 {
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, models.BatchResponse{
		StatusCode: statusCode,
		Success:    success,
		Message:    message,
		Data:       data,
		Errors:     errors,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}
