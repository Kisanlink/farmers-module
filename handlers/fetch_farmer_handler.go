package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kisanlink/protobuf/pb-aaa"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	// Extract query parameters
	user_id := c.Query("user_id")
	farmer_id := c.Query("farmer_id")
	kisansathi_user_id := c.Query("kisansathi_user_id")
	include_user_details := c.Query("user_details") == "true"

	var farmers []models.Farmer
	var err error

	// Always fetch farmers first
	farmers, err = h.FarmerService.FetchFarmers(user_id, farmer_id, kisansathi_user_id)
	if err != nil {
		utils.Log.Error("Failed to fetch farmers", "error", err.Error()) // Replaced log with utils.logger
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}

	// If user details are requested and we have farmers with user_ids
	if include_user_details && len(farmers) > 0 {
		// Collect all unique user IDs from farmers
		user_ids := make([]string, 0, len(farmers))
		for _, farmer := range farmers {
			if farmer.UserId != "" {
				user_ids = append(user_ids, farmer.UserId)
			}
		}

		// Fetch user details for all user_ids
		user_details_map := make(map[string]*pb.User)
		for _, uid := range user_ids {
			user_details, err := services.GetUserByIdClient(context.Background(), uid)
			if err != nil {
				utils.Log.Error("Error fetching user details for user_id", "user_id", uid, "error", err.Error()) // Replaced log with utils.logger
				continue
			}
			if user_details != nil && user_details.Data != nil {
				user_details_map[uid] = user_details.Data
			}
		}

		// Assign user details to farmers
		for i := range farmers {
			if details, exists := user_details_map[farmers[i].UserId]; exists {
				farmers[i].UserDetails = details
			}
		}
	}

	utils.Log.Info("Farmers fetched successfully", "farmer_count", len(farmers)) // Replaced log with utils.logger
	utils.SendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}
