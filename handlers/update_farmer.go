package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AssignKisansathiToFarmers handles assigning Kisansathi UserId to multiple farmers
func (h *FarmerHandler) AssignKisansathiToFarmers(c *gin.Context) {
	// Extract Kisansathi User ID from query parameter
	kisansathiUserId := c.DefaultQuery("kisansathi_user_id", "")
	if kisansathiUserId == "" {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Kisansathi user ID is required", "missing Kisansathi user ID")
		return
	}

	// Bind the list of farmer IDs from the request body
	var farmerIds []string
	if err := c.ShouldBindJSON(&farmerIds); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid request body", err.Error())
		return
	}

	// Call service to update the Kisansathi User ID for these farmers
	if err := h.farmerService.AssignKisansathiToFarmers(kisansathiUserId, farmerIds); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to assign Kisansathi User ID", err.Error())
		return
	}

	// Send success response
	h.sendSuccessResponse(c, http.StatusOK,
		"Kisansathi User ID assigned to farmers successfully", nil)
}
