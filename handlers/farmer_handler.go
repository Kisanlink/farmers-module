package handlers

import (
	"net/http"
	"time"
	"log"
	
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type FarmerHandler struct {
	farmerService services.FarmerServiceInterface
}

func NewFarmerHandler(farmerService services.FarmerServiceInterface) *FarmerHandler {
	return &FarmerHandler{
		farmerService: farmerService,
	}
}

func (h *FarmerHandler) FarmerSignupHandler(c *gin.Context) {
	var req models.FarmerSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// If UserID is provided, we only need to check KisansathiUserID
	if req.UserID != nil {
		// Still need to validate KisansathiUserID if it's provided
		if req.KisansathiUserID != nil {
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserID)
			if err != nil {
				h.sendErrorResponse(c, http.StatusInternalServerError, 
					"Failed to verify Kisansathi user", err.Error())
				return
			}
			
			// Check if user exists and response is valid
			if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.Data == nil {
				h.sendErrorResponse(c, http.StatusUnauthorized, 
					"Kisansathi user not found", "invalid user response")
				return
			}
			
			// Check permissions
			if userResp.Data.UsageRight == nil {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "user has no usage rights defined")
				return
			}

			hasPermission := false
			for _, perm := range userResp.Data.UsageRight.Permissions {
				if perm.Name == "manage_farmers" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "missing required permissions or actions")
				return
			}
		}

		// Create Farmer Record with just the UserID and optional KisansathiUserID
		farmer, err := h.farmerService.CreateFarmer(*req.UserID, req)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to create farmer record", err.Error())
			return
		}

		// Assign Farmer Role
		if _, err := services.AssignRoleToUserClient(
			c.Request.Context(), 
			*req.UserID, 
			"FARMER",
		); err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Farmer created but role assignment failed", err.Error())
			return
		}

		// Return Success Response
		h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", farmer)
		return
	}

	// If UserID is not provided, we need all personal information
	// Check if phone number is present
	if req.MobileNumber == 0 {
		h.sendErrorResponse(c, http.StatusBadRequest, "Mobile number is required", "mobile_number field is missing or invalid")
		return
	}

	// Handle Kisansathi User ID if present (same as above)
	if req.KisansathiUserID != nil {
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserID)
			if err != nil {
				h.sendErrorResponse(c, http.StatusInternalServerError, 
					"Failed to verify Kisansathi user", err.Error())
				return
			}
			
			// Check if user exists and response is valid
			if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.Data == nil {
				h.sendErrorResponse(c, http.StatusUnauthorized, 
					"Kisansathi user not found", "invalid user response")
				return
			}
			
			// Check permissions
			if userResp.Data.UsageRight == nil {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "user has no usage rights defined")
				return
			}

			hasPermission := false
			for _, perm := range userResp.Data.UsageRight.Permissions {
				if perm.Name == "manage_farmers" {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "missing required permissions or actions")
				return
			}
		}

	// Create new user via AAA service since UserID wasn't provided
	if req.Name == nil || req.AadhaarNumber == nil {
		h.sendErrorResponse(c, http.StatusBadRequest, "Name and Aadhaar number are required", "missing required fields")
		return
	}

	createUserResp, err := services.CreateUserClient(req, "")
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, 
			"Failed to create user in AAA service", err.Error())
		return
	}
	if createUserResp == nil || createUserResp.Data == nil || createUserResp.Data.Id == "" {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Invalid response from AAA service", 
			"empty user ID in response")
		return
	}
	userID := createUserResp.Data.Id

	// Create Farmer Record
	farmer, err := h.farmerService.CreateFarmer(userID, req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}

if _, err := services.AssignRoleToUserClient(c.Request.Context(), userID, "FARMER"); err != nil {
    log.Printf("Role assignment failed for user %s: %v", userID, err)
    h.sendErrorResponse(c, http.StatusInternalServerError,
        "Role assignment failed", "invalid user ID format or system error")
    return
}

	// Return Success Response
	h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", farmer)
}

// Helper methods as receiver functions
func (h *FarmerHandler) sendErrorResponse(c *gin.Context, status int, userMessage string, errorDetail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   userMessage,
		"error":     errorDetail,
		"timestamp": time.Now().UTC(),
		"data":      nil,
		"success":   false,
	})
}

func (h *FarmerHandler) sendSuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   message,
		"error":     nil,
		"timestamp": time.Now().UTC(),
		"data":      data,
		"success":   true,
	})
}

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	farmers, err := h.farmerService.FetchFarmers()
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}
	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
}