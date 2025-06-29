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
	// 1) Extract query parameters
	userId := c.Query("user_id")
	farmerId := c.Query("farmer_id")
	kisansathiUserId := c.Query("kisansathi_user_id")
	fpoRegNo := c.Query("fpo_reg_no")
	includeUserDetails := c.Query("user_details") == "true"
	subscribed := c.Query("subscribed") == "true"

	var farmers []models.Farmer
	var err error

	// 2) If ?subscribed=true, call FetchSubscribedFarmers. Otherwise, call the normal FetchFarmers.
	if subscribed {
		// Note: FetchSubscribedFarmers only accepts userId and kisansathiUserId,
		// so we ignore farmerId here.
		farmers, err = h.farmerService.FetchSubscribedFarmers(userId, kisansathiUserId)
	} else {
		farmers, err = h.farmerService.FetchFarmers(userId, farmerId, kisansathiUserId, fpoRegNo)
	}
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}

	// 3) (Unchanged) If user_details=true, enrich each farmer with gRPC user info
	if includeUserDetails && len(farmers) > 0 {
		userIds := make([]string, 0, len(farmers))
		for _, farmer := range farmers {
			if farmer.UserId != "" {
				userIds = append(userIds, farmer.UserId)
			}
		}

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

		for i := range farmers {
			if details, exists := userDetailsMap[farmers[i].UserId]; exists {
				farmers[i].UserDetails = details
			}
		}
	}

	// 4) Return the final list
	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}
