package handlers

import (
	"net/http"
	"time"

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

	
	// Handle Kisansathi User ID if present
if req.KisansathiUserID != nil {
    // Verify Kisansathi user exists - using the existing GetUserByIdClient
    userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserID)
    if err != nil {
        h.sendErrorResponse(c, http.StatusInternalServerError, 
            "Failed to verify Kisansathi user", err.Error())
        return
    }
    
    // Check if user exists and response is valid
    if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.User == nil {
        h.sendErrorResponse(c, http.StatusUnauthorized, 
            "Kisansathi user not found", "invalid user response")
        return
    }
    
    // Check permissions and actions
    if userResp.User.UsageRight == nil {
        h.sendErrorResponse(c, http.StatusForbidden,
            "Permission denied", "user has no usage rights defined")
        return
    }
    
    hasPermission := false
    for _, perm := range userResp.User.UsageRight.Permissions {
        if perm == "manage_farmers" {
            hasPermission = true
            break
        }
    }
    
    hasAction := false
    for _, action := range userResp.User.UsageRight.Actions {
        if action == "create" {
            hasAction = true
            break
        }
    }
    
    if !hasPermission || !hasAction {
        h.sendErrorResponse(c, http.StatusForbidden,
            "Permission denied", "missing required permissions or actions")
        return
    }
}

	// Handle User Creation
	var userID string
	if req.UserID != nil {
		userID = *req.UserID
	} else {
		// Create new user via AAA service
		createUserResp, err := services.CreateUserClient(req, "")
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError, 
				"Failed to create user in AAA service", err.Error())
			return
		}
		if createUserResp == nil || createUserResp.User == nil || createUserResp.User.Id == "" {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Invalid response from AAA service", 
				"empty user ID in response")
			return
		}
		userID = createUserResp.User.Id
	}

	// Create Farmer Record
	farmer, err := h.farmerService.CreateFarmer(userID, req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}

	// Assign Farmer Role
	if _, err := services.AssignRoleToUserClient(
		c.Request.Context(), 
		userID, 
		req.Roles,
	); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Farmer created but role assignment failed", err.Error())
		return
	}

	// Return Success Response
	h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", farmer)
}

func (h *FarmerHandler) FetchFarmersHandler(c *gin.Context) {
	farmers, err := h.farmerService.FetchFarmers()
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch farmers", err.Error())
		return
	}
	h.sendSuccessResponse(c, http.StatusOK, "Farmers fetched successfully", farmers)
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