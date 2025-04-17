package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
	"github.com/kisanlink/protobuf/pb-aaa"
)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	// Extract query parameters
	userId := c.Query("user_id")
	farmerId := c.Query("farmer_id")
	kisansathiUserId := c.Query("kisansathi_user_id")
	includeUserDetails := c.Query("user_details") == "true"

	var farmers []models.Farmer
	var err error

	// Always fetch farmers first
	farmers, err = h.farmerService.FetchFarmers(userId, farmerId, kisansathiUserId)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}

	// If user details are requested and we have farmers with user_ids
	if includeUserDetails && len(farmers) > 0 {
		// Collect all unique user IDs from farmers
		userIds := make([]string, 0, len(farmers))
		for _, farmer := range farmers {
			if farmer.UserId != "" {
				userIds = append(userIds, farmer.UserId)
			}
		}

		// Fetch user details for all user_ids
		userDetailsMap := make(map[string]*pb.User)
		for _, uid := range userIds {
			userDetails, err := services.GetUserByIdClient(context.Background(), uid)
			if err != nil {
				log.Printf("Error fetching user details for %s: %v", uid, err)
				continue
			}
			if userDetails != nil && userDetails.Data != nil {
				userDetailsMap[uid] = userDetails.Data
			}
		}

		// Assign user details to farmers
		for i := range farmers {
			if details, exists := userDetailsMap[farmers[i].UserId]; exists {
				farmers[i].UserDetails = details
			}
		}
	}

	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}
