package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type CropCycleHandler struct {
	service services.CropCycleServiceInterface
}

func NewCropCycleHandler(service services.CropCycleServiceInterface) *CropCycleHandler {
	return &CropCycleHandler{service: service}
}

func (h *CropCycleHandler) CreateCropCycle(c *gin.Context) {
	farmID := c.Param("farmId") // Extract farm ID from the path

	var req struct {
		CropID           string    `json:"crop_id" binding:"required"`
		StartDate        time.Time `json:"start_date" binding:"required"`
		Acreage          float64   `json:"acreage" binding:"required"`
		ExpectedQuantity float64   `json:"expected_quantity" binding:"required"`
		Report           string    `json:"report"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request payload",
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	cycle, err := h.service.CreateCropCycle(
		farmID, req.CropID,
		req.StartDate, nil,
		req.Acreage, req.ExpectedQuantity,
		nil,
		req.Report,
	)

	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Could not create crop cycle"

		if err.Error() == "farm not found" {
			statusCode = http.StatusNotFound
			message = "Farm not found"
		} else if err.Error() == "acreage exceeds available area on farm" {
			statusCode = http.StatusBadRequest
			message = "Acreage exceeds available area on farm"
		}

		c.JSON(statusCode, models.Response{
			StatusCode: statusCode,
			Success:    false,
			Message:    message,
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		StatusCode: http.StatusCreated,
		Success:    true,
		Message:    "Crop cycle created successfully",
		Data:       cycle,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}

// GetCropCycles handles GET /api/v1/farms/{farmId}/crop-cycles
func (h *CropCycleHandler) GetCropCycles(c *gin.Context) {
	farmID := c.Param("farmId")
	cropID := c.Query("cropID")
	status := c.Query("status")

	var cropIDPtr, statusPtr *string
	if cropID != "" {
		cropIDPtr = &cropID
	}
	if status != "" {
		statusPtr = &status
	}

	cycles, err := h.service.GetCropCycles(farmID, cropIDPtr, statusPtr)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Could not fetch crop cycles"

		// Check for specific error types
		if err.Error() == "farm ID is required" {
			statusCode = http.StatusBadRequest
			message = "Farm ID is required"
		} else if err.Error() == "invalid status: must be either ONGOING or COMPLETED" {
			statusCode = http.StatusBadRequest
			message = "Invalid status parameter"
		}

		c.JSON(statusCode, models.Response{
			StatusCode: statusCode,
			Success:    false,
			Message:    message,
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop cycle(s) retrieved successfully",
		Data:       cycles,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}

// UpdateCropCycle handles PUT /api/v1/farms/{farmId}/crop-cycles/{cycleId}
func (h *CropCycleHandler) UpdateCropCycle(c *gin.Context) {
	farmID := c.Param("farmId")
	cycleID := c.Param("cycleId")

	var req struct {
		EndDate  *time.Time `json:"end_date"`
		Quantity *float64   `json:"quantity"`
		Report   *string    `json:"report"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request payload",
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Validate that cycle belongs to the specified farm
	_, err := h.service.ValidateCropCycleBelongsToFarm(cycleID, farmID)
	if err != nil {
		statusCode := http.StatusNotFound
		message := "Crop cycle not found"

		if err.Error() == "crop cycle does not belong to the specified farm" {
			statusCode = http.StatusForbidden
			message = "Crop cycle does not belong to the specified farm"
		}

		c.JSON(statusCode, models.Response{
			StatusCode: statusCode,
			Success:    false,
			Message:    message,
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Step 2: Validate required fields for completion
	if req.EndDate == nil || req.Quantity == nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "End date and quantity are required for completing a crop cycle",
			Error:      "Missing required fields",
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Get the report string if provided
	if req.Report != nil {
		_ = *req.Report // Use the report value if needed
	}

	// Update the crop cycle
	updatedCycle, err := h.service.UpdateCropCycleByID(cycleID, req.EndDate, req.Quantity, req.Report)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Could not update crop cycle"

		if err.Error() == "end date must be after start date" {
			statusCode = http.StatusBadRequest
			message = "End date must be after start date"
		}

		c.JSON(statusCode, models.Response{
			StatusCode: statusCode,
			Success:    false,
			Message:    message,
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Crop cycle updated successfully",
		Data:       updatedCycle,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}
