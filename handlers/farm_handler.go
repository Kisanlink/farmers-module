package handlers

import (
	"log"
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// CreateFarmHandler handles farm creation based on user permissions.
func FarmHandler(farmService services.FarmServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("CreateFarmHandler triggered")

		var req models.FarmRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Invalid request parameters", "error": err.Error()})
			return
		}

		log.Println("Received farm creation request:", req)

		// 1️⃣ **Ensure FarmerID is present in the request**
		if req.FarmerID == "" {
			log.Println("Missing FarmerID in request")
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Farmer ID is required"})
			return
		}

		
	// 1️⃣ Check if KisansathiUserID exists and validate it
		if req.KisansathiUserID != nil {
			// Fetch user details from AAA service
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserID)
			if err != nil || userResp.User == nil {
				log.Println("Failed to fetch user or user not found:", err)
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  401,
					"message": "User not found",
					"error":   "Unauthorized",
				})
				return
			}
		}

		// 2️⃣ Check if Kisansathi has permission to create a farm
		if req.KisansathiUserID != nil {
permResp, err := services.CheckPermissionClient(c.Request.Context(), *req.KisansathiUserID,  req.Actions,"")		
	if err != nil {
				log.Println("User does not have permission to create farm:", *req.KisansathiUserID)
				c.JSON(http.StatusForbidden, gin.H{
					"status":  403,
					"message": "User does not have permission to create farm",
					"error":   "Forbidden",
				})
				return
			}
			log.Println("User has permission to create farm:", permResp)
		} else {
			log.Println("KisansathiUserID is nil, skipping permission check")
		}


		// 5️⃣ **Proceed with Farm Creation**
		log.Println("Creating farm for FarmerID:", req.FarmerID)
		newFarm, err := farmService.CreateFarm(req)

		if err != nil {
			log.Println("Failed to create farm:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Failed to create farm", "error": err.Error()})
			return
		}

		log.Println("Farm created successfully:", newFarm)
		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Farm created successfully",
			"data": gin.H{
				"id":        newFarm.ID,
				"user_id":   newFarm.FarmerID,
				"verified":  newFarm.Verified,
				"location":  newFarm.Location,
				"area":      newFarm.Area,
				"locality":  newFarm.Locality,
			},
		})
	}
}
