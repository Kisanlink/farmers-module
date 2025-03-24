package handlers

import (
	"log"
	"net/http"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)
func FarmerSignupHandler(farmerService services.FarmerServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("FarmerSignupHandler triggered") // Debug log
		var req models.FarmerSignupRequest
		log.Println("Received request for farmer signup")

		// Parse Request Body
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 400, "message": "Invalid request parameters", "error": err.Error()})
			return
		}
		log.Println("Request parsed successfully:", req)

		// 1️⃣ Check if KisansathiUserID exists and validate it
		if req.KisansathiUserID != nil {
			// Fetch user details from AAA service
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserID )
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

		// 2️⃣ Check if Kisansathi has permission to create a farmer
		if req.KisansathiUserID != nil {
			permResp, err := services.CheckPermissionClient(c.Request.Context(), *req.KisansathiUserID,  req.Actions,"")
			if err != nil {
				log.Println("User does not have permission to create farmer:", *req.KisansathiUserID)
				c.JSON(http.StatusForbidden, gin.H{
					"status":  403,
					"message": "User does not have permission to create farmer",
					"error":   "Forbidden",
				})
				return
			}
			log.Println("User has permission to create farmer:", permResp)
		} else {
			log.Println("KisansathiUserID is nil, skipping permission check")
		}

		var userID string
		// **Step 1: Check if User ID is present in the request**
		if req.UserID != nil {
			userID = *req.UserID
		} else {
			// **Step 2: Create user via AAA Service**
			response, err := services.CreateUserClient(req, "")
			if err != nil {
				log.Println("Failed to create user in AAA service:", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  500,
					"message": "Failed to create user in AAA service",
					"error":   err.Error(),
				})
				return
			}

			// Ensure AAA service response is valid
			if response == nil || response.User == nil || response.User.Id == "" {
				log.Println("AAA service returned an invalid response")
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  500,
					"message": "AAA service returned an invalid response",
					"error":   "User ID not found in response",
				})
				return
			}

			// Use user ID from AAA response
			userID = response.User.Id
		}

		// **Step 3: Register Farmer with userID**
		newFarmer, err := farmerService.CreateFarmer(userID, req)
		if err != nil {
			log.Println("Failed to register farmer:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to register farmer",
				"error":   err.Error(),
			})
			return
		}

		log.Println("Farmer registered successfully:", newFarmer)
		roleResp, err := services.AssignRoleToUserClient(c.Request.Context(), userID, req.Roles)
		if err != nil {
			log.Println("Failed to assign role to user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500,
				"message": "Failed to assign role to user",
				"error":   err.Error(),
			})
			return
		}

		log.Println("Role assigned successfully:", roleResp)
		// **Step 4: Return Success Response**
		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Farmer registered successfully",
			"error":   false,
			"data": gin.H{
				"id":                 newFarmer.ID,
				"user_id":            newFarmer.UserID,
				"kisansathi_user_id": newFarmer.KisansathiUserID,
				"is_active":          newFarmer.IsActive,
				"role_assignment":    roleResp.Message,
			},
		})
	}
}

// FetchFarmersHandler handles fetching farmers
func FetchFarmersHandler(farmerService services.FarmerServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fetch farmers from the service
		farmers, err := farmerService.FetchFarmers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to fetch farmers",
				"error":   err.Error(),
			})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "Farmers fetched successfully",
			"data":    farmers,
		})
	}
}