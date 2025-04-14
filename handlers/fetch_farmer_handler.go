package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/gin-gonic/gin"
	"github.com/kisanlink/protobuf/pb-aaa"
)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	// Extract query parameters
	userId := c.Query("user_id")
	farmerId := c.Query("farmer_id")
	kisansathiUserId := c.Query("kisansathi_user_id")

	var farmers []models.Farmer
	var userDetails *pb.GetUserByIdResponse
	var err error

	if userId != "" {
		// Fetch user details and farmers
		farmers, userDetails, err = h.farmerService.FetchFarmers(userId, farmerId, kisansathiUserId)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
			return
		}

		if userDetails != nil && userDetails.Data != nil {
			for i := range farmers {
				farmers[i].UserDetails = userDetails.Data // Unwrap and set the inner user object.
			}
		}

	} else {
		// Fetch all farmers without user details
		farmers, err = h.farmerService.FetchFarmersWithoutUserDetails(farmerId, kisansathiUserId)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
			return
		}
	}

	response := farmers

	/*
		// Return success response with farmers and user details (if available)
		response := gin.H{
			"farmers": farmers,
		}

		if userDetails != nil {
			response["user"] = userDetails
		}
	*/

	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", response)
}
