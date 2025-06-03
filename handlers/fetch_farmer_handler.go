package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
	"github.com/kisanlink/protobuf/pb-aaa"
)

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	userID := c.Query("user_id")
	farmerID := c.Query("farmer_id")
	kisanID := c.Query("kisansathi_user_id")
	includeUserDetails := c.Query("user_details") == "true"

	// Parse 'subscribed' query parameter
	subscribedParam := c.Query("subscribed")
	var filterBySubscribed bool
	var subscribedValue bool
	if subscribedParam != "" {
		val, err := strconv.ParseBool(subscribedParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid 'subscribed' value. Use true or false.",
				Error:      err.Error(),
				TimeStamp:  time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
		filterBySubscribed = true
		subscribedValue = val
	}

	// Step 1: Fetch all farmers
	farmers, err := h.farmerService.FetchFarmers(userID, farmerID, kisanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to fetch farmers",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Step 2: Apply 'subscribed' filter
	if filterBySubscribed {
		filtered := farmers[:0]
		for _, f := range farmers {
			if f.IsSubscribed == subscribedValue {
				filtered = append(filtered, f)
			}
		}
		farmers = filtered
	}

	// Step 3: Optionally enrich with user details
	if includeUserDetails && len(farmers) > 0 {
		userMap := make(map[string]*pb.User, len(farmers))
		for _, f := range farmers {
			if f.UserId == "" {
				continue
			}
			resp, err := services.GetUserByIdClient(context.Background(), f.UserId)
			if err != nil {
				log.Printf("Error fetching user %s: %v", f.UserId, err)
				continue
			}
			if resp != nil && resp.Data != nil {
				userMap[f.UserId] = resp.Data
			}
		}
		for i := range farmers {
			if ud := userMap[farmers[i].UserId]; ud != nil {
				farmers[i].UserDetails = ud
			}
		}
	}

	// Step 4: Respond
	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Farmers fetched successfully",
		Data:       farmers,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}
