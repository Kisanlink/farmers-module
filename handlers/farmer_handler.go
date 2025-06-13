package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/entities"
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

	// Validate req.Type if provided in JSON
	if req.Type != "" {
		// Optionally normalize case, e.g., uppercase
		tp := strings.ToUpper(req.Type)
		if !entities.FARMER_TYPES.IsValid(tp) {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"Invalid farmer type",
				fmt.Sprintf("`type` must be one of: %v", entities.FARMER_TYPES.StringValues()))
			return
		}
		req.Type = tp
	}
	// If req.Type is empty, service layer will default to OTHER

	// If UserId is provided, handle that flow
	if req.UserId != nil {
		// (existing KisansathiUserId verification...)
		if req.KisansathiUserId != nil {
			userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserId)
			if err != nil {
				h.sendErrorResponse(c, http.StatusInternalServerError,
					"Failed to verify Kisansathi user", err.Error())
				return
			}
			if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.Data == nil {
				h.sendErrorResponse(c, http.StatusUnauthorized,
					"Kisansathi user not found", "invalid user response")
				return
			}
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

		// Create Farmer Record with UserId, optional KisansathiUserId, and req.Type
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
		h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   userDetails,
		})
		return
	}

	// Mobile-based signup: same as before
	if len(req.MobileNumberString) != 10 {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "must be exactly 10 digits")
		return
	}
	if req.MobileNumberString[0] == '0' {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "should not start with 0")
		return
	}
	mobileUint, err := strconv.ParseUint(req.MobileNumberString, 10, 64)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"Invalid mobile number", "must contain only digits")
		return
	}
	req.MobileNumber = mobileUint

	if req.KisansathiUserId != nil {
		userResp, err := services.GetUserByIdClient(c.Request.Context(), *req.KisansathiUserId)
		if err != nil {
			h.sendErrorResponse(c, http.StatusInternalServerError,
				"Failed to verify Kisansathi user", err.Error())
			return
		}
		if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.Data == nil {
			h.sendErrorResponse(c, http.StatusUnauthorized,
				"Kisansathi user not found", "invalid user response")
			return
		}
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

	farmer, userDetails, err := h.farmerService.CreateFarmer(userId, req)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Failed to create farmer record", err.Error())
		return
	}
	if _, err := services.AssignRoleToUserClient(c.Request.Context(), userId, config.ROLE_FARMER); err != nil {
		log.Printf("Role assignment failed for user %s: %v", userId, err)
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"Role assignment failed", "invalid user Id format or system error")
		return
	}
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

// SetSubscription handles “POST /farmers/:farmer_id/subscription?is_subscribed={true|false}”
func (h *FarmerHandler) SetSubscription(c *gin.Context) {
	// 1. Extract path parameter
	farmerID := c.Param("farmer_id")
	if farmerID == "" {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"farmer_id is required in path", "missing path parameter")
		return
	}

	// 2. Extract query parameter
	isSubStr := c.Query("is_subscribed")
	if isSubStr == "" {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"is_subscribed query parameter is required", "missing query parameter")
		return
	}

	// 3. Parse “is_subscribed” into a bool
	isSubscribed, err := strconv.ParseBool(isSubStr)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest,
			"invalid value for is_subscribed; must be true or false", err.Error())
		return
	}

	// 4. Call service to toggle the subscription flag
	if err := h.farmerService.SetSubscriptionStatus(farmerID, isSubscribed); err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError,
			"failed to update subscription status", err.Error())
		return
	}

	// 5. Return a success response
	h.sendSuccessResponse(c, http.StatusOK, "subscription status updated", gin.H{
		"farmer_id":     farmerID,
		"is_subscribed": isSubscribed,
	})
}
