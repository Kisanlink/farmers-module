package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/permission"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/gin-gonic/gin"
)

type FarmerHandler struct {
	FarmerService services.FarmerServiceInterface
}

func NewFarmerHandler(farmer_service services.FarmerServiceInterface) *FarmerHandler {
	return &FarmerHandler{
		FarmerService: farmer_service,
	}
}

func (h *FarmerHandler) FarmerSignupHandler(c *gin.Context) {
	var req models.FarmerSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Log.Warnf("Invalid signup request: %v", err)
		h.sendErrorResponse(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	if req.UserId != nil {
		if req.KisansathiUserId != nil {
			requiredPermission := config.PERMISSION_KISANSATHI
			hasPerm, statusCode, userMsg, errDetail := permission.CheckUserPermission(c.Request.Context(), *req.KisansathiUserId, requiredPermission)
			if !hasPerm {
				utils.SendErrorResponse(c, statusCode, userMsg, errDetail)
				return
			}

		}

		farmer, user_details, err := h.FarmerService.CreateFarmer(*req.UserId, req)
		if err != nil {
			utils.Log.Errorf("Failed to create farmer for user %s: %v", *req.UserId, err)
			h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to create farmer record", err.Error())
			return
		}

		if _, err := services.AssignRoleToUserClient(c.Request.Context(), *req.UserId, config.ROLE_FARMER); err != nil {
			utils.Log.Errorf("Role assignment failed for user %s: %v", *req.UserId, err)
			h.sendErrorResponse(c, http.StatusInternalServerError, "Farmer created but role assignment failed", err.Error())
			return
		}

		utils.Log.Infof("Farmer registered successfully with existing user ID: %s", *req.UserId)
		h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   user_details,
		})
		return
	}

	if req.MobileNumber == 0 {
		utils.Log.Warn("Signup failed: Mobile number missing")
		h.sendErrorResponse(c, http.StatusBadRequest, "Mobile number is required", "mobile_number field is missing or invalid")
		return
	}

	if req.KisansathiUserId != nil {
		requiredPermission := config.PERMISSION_KISANSATHI
		hasPerm, statusCode, userMsg, errDetail := permission.CheckUserPermission(c.Request.Context(), *req.KisansathiUserId, requiredPermission)
		if !hasPerm {
			utils.SendErrorResponse(c, statusCode, userMsg, errDetail)
			return
		}

	}

	if req.UserName == nil || req.AadhaarNumber == nil {
		utils.Log.Warn("Signup failed: Missing user name or Aadhaar number")
		h.sendErrorResponse(c, http.StatusBadRequest, "Name and Aadhaar number are required", "missing required fields")
		return
	}

	create_user_resp, err := services.CreateUserClient(req, "")
	if err != nil {
		utils.Log.Errorf("Failed to create user in AAA service: %v", err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to create user in AAA service", err.Error())
		return
	}
	if create_user_resp == nil || create_user_resp.Data == nil || create_user_resp.Data.Id == "" {
		utils.Log.Error("AAA service returned empty or invalid user ID")
		h.sendErrorResponse(c, http.StatusInternalServerError, "Invalid response from AAA service", "empty user Id in response")
		return
	}
	user_id := create_user_resp.Data.Id

	farmer, user_details, err := h.FarmerService.CreateFarmer(user_id, req)
	if err != nil {
		utils.Log.Errorf("Failed to create farmer record for user %s: %v", user_id, err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "Failed to create farmer record", err.Error())
		return
	}

	if _, err := services.AssignRoleToUserClient(c.Request.Context(), user_id, "FARMER"); err != nil {
		utils.Log.Errorf("Role assignment failed for new user %s: %v", user_id, err)
		h.sendErrorResponse(c, http.StatusInternalServerError, "Role assignment failed", "invalid user Id format or system error")
		return
	}

	utils.Log.Infof("Farmer registered successfully with newly created user ID: %s", user_id)
	h.sendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
		"farmer": farmer,
		"user":   user_details,
	})
}

// Helper methods as receiver functions
func (h *FarmerHandler) sendErrorResponse(c *gin.Context, status int, user_message string, error_detail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   user_message,
		"error":     error_detail,
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
