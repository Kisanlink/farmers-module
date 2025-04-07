package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	// Extract query parameters
	userID := c.Query("user_id")
	farmerID := c.Query("farmer_id")
	kisansathiUserID := c.Query("kisansathi_user_id")

	// Fetch farmers with filters
	farmers, err := h.farmerService.FetchFarmers(userID, farmerID, kisansathiUserID)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}

	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}
