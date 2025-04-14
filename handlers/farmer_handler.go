package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/config"
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

	// If UserId is provided, we only need to check KisansathiUserId
	if req.UserId != nil {
		// Still need to validate KisansathiUserId if it's provided
		if req.KisansathiUserId != nil {
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserId)
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

			/* // Check permissions
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
			*/

			// Check permissions using RolePermissions
			if len(userResp.Data.RolePermissions) == 0 {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "user has no role permissions defined")
				return
			}

			hasPermission := false
			for _, rolePerm := range userResp.Data.RolePermissions {
				for _, perm := range rolePerm.Permissions {
					if perm.Name == config.PERMISSION_KISANSATHI {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			if !hasPermission {
				h.sendErrorResponse(c, http.StatusForbidden,
					"Permission denied", "missing required permissions or actions")
				return
			}

		}

		// Create Farmer Record with just the UserId and optional KisansathiUserId
		farmer, userDetails, err := h.farmerService.CreateFarmer(*req.UserId, req)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to create farmer record", err.Error())
			return
		}

		// Assign Farmer Role
		if _, err := services.AssignRoleToUserClient(
			c.Request.Context(),
			*req.UserId,
			config.ROLE_FARMER,
		); err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Farmer created but role assignment failed", err.Error())
			return
		}

		// Return Success Response
		h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   userDetails,
		})
		return
	}

	// If UserId is not provided, we need all personal information
	// Check if phone number is present
	if req.MobileNumber == 0 {
		h.sendErrorResponse(c, http.StatusBadRequest, "Mobile number is required", "mobile_number field is missing or invalid")
		return
	}

	// Handle Kisansathi User Id if present (same as above)
	if req.KisansathiUserId != nil {
		userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserId)
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

		/*// Check permissions
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
		*/

		// Check permissions using RolePermissions
		if len(userResp.Data.RolePermissions) == 0 {
			h.sendErrorResponse(c, http.StatusForbidden,
				"Permission denied", "user has no role permissions defined")
			return
		}

		hasPermission := false
		for _, rolePerm := range userResp.Data.RolePermissions {
			for _, perm := range rolePerm.Permissions {
				if perm.Name == config.PERMISSION_KISANSATHI {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			h.sendErrorResponse(c, http.StatusForbidden,
				"Permission denied", "missing required permissions or actions")
			return
		}

	}

	// Create new user via AAA service since UserId wasn't provided
	if req.UserName == nil || req.AadhaarNumber == nil {
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
			"empty user Id in response")
		return
	}
	userId := createUserResp.Data.Id

	// Create Farmer Record
	farmer, userDetails, err := h.farmerService.CreateFarmer(userId, req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}

	if _, err := services.AssignRoleToUserClient(c.Request.Context(), userId, "FARMER"); err != nil {
		log.Printf("Role assignment failed for user %s: %v", userId, err)
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Role assignment failed", "invalid user Id format or system error")
		return
	}

	// Return Success Response
	h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
		"farmer": farmer,
		"user":   userDetails,
	})
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
