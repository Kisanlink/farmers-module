package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Kisanlink/farmers-module/entities"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
)

// FarmActivityHandler handles HTTP requests for farm activities.
type FarmActivityHandler struct {
	Service services.FarmActivityServiceInterface
}

// NewFarmActivityHandler creates an instance of FarmActivityHandler.
func NewFarmActivityHandler(service services.FarmActivityServiceInterface) *FarmActivityHandler {
	return &FarmActivityHandler{Service: service}
}

// CreateFarmActivityRequest represents the expected JSON body for creating a farm activity.
type CreateFarmActivityRequest struct {
	FarmId         string     `json:"farm_id" binding:"required"`
	CropCycleId    string     `json:"crop_cycle_id" binding:"required"`
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
		FarmId:         req.FarmId,
		CropCycleId:    req.CropCycleId,
		Activity:       entities.ActivityType(req.Activity),
		StartDate:      &req.StartDate,
		EndDate:        req.EndDate,
		ActivityReport: req.ActivityReport,
	}

	if err := h.Service.CreateActivity(activity); err != nil {
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
	farm_id := c.Query("farm_id")
	cycle_id := c.Query("crop_cycle_id")
	start_date_str := c.Query("start_date")
	end_date_str := c.Query("end_date")

	var (
		data interface{}
		err  error
	)

	if id != "" {
		data, err = h.Service.GetActivityById(id)
	} else if (farm_id != "" || cycle_id != "") && start_date_str != "" && end_date_str != "" {
		start_date, err1 := time.Parse(time.RFC3339, start_date_str)
		end_date, err2 := time.Parse(time.RFC3339, end_date_str)
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
		if farm_id != "" {
			activities, err = h.Service.GetActivitiesByDateRange(farm_id, start_date, end_date)
		}

		// Filter by cycle_id if also provided
		if err == nil && cycle_id != "" {
			var filtered []*models.FarmActivity
			for _, a := range activities {
				if a.CropCycleId == cycle_id {
					filtered = append(filtered, a)
				}
			}
			data = filtered
		} else {
			data = activities
		}
	} else if farm_id != "" && cycle_id != "" {
		activities, err := h.Service.GetActivitiesByFarmId(farm_id)
		if err == nil {
			var filtered []*models.FarmActivity
			for _, a := range activities {
				if a.CropCycleId == cycle_id {
					filtered = append(filtered, a)
				}
			}
			data = filtered
		}
	} else if cycle_id != "" {
		data, err = h.Service.GetActivitiesByCropCycle(cycle_id)
	} else if farm_id != "" {
		data, err = h.Service.GetActivitiesByFarmId(farm_id)
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
	activity, err := h.Service.GetActivityById(id)
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

	if err := h.Service.UpdateActivity(activity); err != nil {
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

	if err := h.Service.DeleteActivity(id); err != nil {
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
