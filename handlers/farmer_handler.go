package handlers

import (
	"log"
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// FarmerSignupHandler handles farmer registration requests
func FarmerSignupHandler(farmerService services.FarmerServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.FarmerSignupRequest
		log.Println("Received request for farmer signup")

		// Parse Request Body
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Invalid request parameters", "error": err.Error()})
			return
		}
		log.Println("Request parsed successfully:", req)

		// Call the AAA service to create a user
		response, err := services.CreateUserClient(req, "")
		if err != nil {
			log.Println("Failed to create user in AAA service:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": 500, "message": "Failed to create user in AAA service", "error": err.Error()})
			return
		}

		// Register Farmer with the user_id from the AAA service response
		newFarmer, err := farmerService.CreateFarmer(response.User.Id, req)
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
				"kisansathi_user_id": newFarmer.KisansathiUserID,
				"is_active":          newFarmer.IsActive,
			},
		})
	}
}