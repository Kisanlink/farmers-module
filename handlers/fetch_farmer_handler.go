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

	// 1) base fetch
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

	// 2) filter on is_subscribed if requested
	if subQ := c.Query("is_subscribed"); subQ != "" {
		wantSub, perr := strconv.ParseBool(subQ)
		if perr != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid is_subscribed flag",
				Error:      perr.Error(),
				TimeStamp:  time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
		filtered := farmers[:0]
		for _, f := range farmers {
			if f.IsSubscribed == wantSub {
				filtered = append(filtered, f)
			}
		}
		farmers = filtered
	}

	// 3) enrich with user details if requested
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

	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Farmers fetched successfully",
		Data:       farmers,
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}

// -----------------------------------------
// 3) SUBSCRIBE / UNSUBSCRIBE
// -----------------------------------------
type subscribeRequest struct {
	IsSubscribed bool `json:"is_subscribed" binding:"required"`
}

func (h *FarmerHandler) SubscribeHandler(c *gin.Context) {
	farmerID := c.Param("id")
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			StatusCode: http.StatusBadRequest,
			Success:    false,
			Message:    "Invalid request body",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	if err := h.farmerService.SetSubscriptionStatus(farmerID, req.IsSubscribed); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Could not update subscription",
			Error:      err.Error(),
			TimeStamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	msg := "unsubscribed"
	if req.IsSubscribed {
		msg = "subscribed"
	}
	c.JSON(http.StatusOK, models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Farmer " + msg + " successfully",
		TimeStamp:  time.Now().UTC().Format(time.RFC3339),
	})
}
