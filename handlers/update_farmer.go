package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/gin-gonic/gin"
)

// AssignKisansathiToFarmers handles assigning Kisansathi UserId to multiple farmers
func (h *FarmerHandler) AssignKisansathiToFarmers(c *gin.Context) {
	// Extract Kisansathi User ID from path parameter
	kisansathiUserId := c.Param("kisansathi_user_id")
	if kisansathiUserId == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Kisansathi user ID is required",
			Error:      "Missing Kisansathi user ID in path",
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Bind the list of farmer IDs from the request body
	var req struct {
		FarmerIDs []string `json:"farmer_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request body",
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Call service to update the Kisansathi User ID for these farmers
	if err := h.farmerService.AssignKisansathiToFarmers(kisansathiUserId, req.FarmerIDs); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to assign Kisansathi User ID",
			Error:      err.Error(),
			TimeStamp:  time.Now().Format(time.RFC3339),
		})
		return
	}

	// Send success response
	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Kisansathi User ID assigned to farmers successfully",
		Data:       nil,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}

// UpdateFarmerHandler handles the API request to update a farmer's details.
func (h *FarmerHandler) UpdateFarmerHandler(c *gin.Context) {
	farmerId := c.Param("farmer_id")

	var req models.FarmerUpdateRequest
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

	updatedFarmer, err := h.farmerService.UpdateFarmer(farmerId, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Could not update farmer details"

		// Customize error response based on service-layer errors
		if err.Error() == "no update fields provided" {
			statusCode = http.StatusBadRequest
			message = err.Error()
		} else if err.Error() == fmt.Sprintf("farmer with ID %s not found", farmerId) {
			statusCode = http.StatusNotFound
			message = "Farmer not found"
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
		Message:    "Farmer details updated successfully",
		Data:       updatedFarmer,
		TimeStamp:  time.Now().Format(time.RFC3339),
	})
}
