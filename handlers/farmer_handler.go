package handlers

import (
	"log"
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// FarmerSignupHandler handles farmer registration requests
func FarmerSignupHandler(farmerService services.FarmerServiceInterface, aaaService services.AAAServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.FarmerSignupRequest
		log.Println("Received request for farmer signup")

		// Parse Request Body
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Invalid request parameters", "error": err.Error()})
			return
		}
		log.Println("Request parsed successfully:", req)

		// Define Farmer Role ID (you should retrieve this from config or DB)
		farmerRoleID := "FARMER_ROLE_ID" // Change this to actual role ID

		// Call gRPC to create user in AAA service
		userID, err := aaaService.CreateUser(req.Email, req.Name, []string{farmerRoleID})
		if err != nil {
			log.Println("Failed to create user in AAA service:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "message": "Failed to create user", "error": err.Error()})
			return
		}
		log.Println("User created successfully, userID:", userID)

		// Register Farmer
		newFarmer, err := farmerService.CreateFarmer(userID, req)
		if err != nil {
			log.Println("Failed to register farmer:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Failed to register farmer", "error": err.Error()})
			return
		}
		log.Println("Farmer registered successfully:", newFarmer)

		// Return Success Response
		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Farmer registered successfully",
			"error":   false,
			"data": gin.H{
				"id":                 newFarmer.ID,
				"user_id":            newFarmer.UserID,
				"kisansathi_user_id": newFarmer.KisanSathiUserID,
				"is_active":          newFarmer.IsActive,
			},
		})
	}
}

