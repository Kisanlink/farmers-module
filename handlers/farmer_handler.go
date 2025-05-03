package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/permission"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
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
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
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
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create farmer record", err.Error())
			return
		}

		if _, err := services.AssignRoleToUserClient(c.Request.Context(), *req.UserId, config.ROLE_FARMER); err != nil {
			utils.Log.Errorf("Role assignment failed for user %s: %v", *req.UserId, err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Farmer created but role assignment failed", err.Error())
			return
		}

		utils.Log.Infof("Farmer registered successfully with existing user ID: %s", *req.UserId)
		utils.SendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
			"farmer": farmer,
			"user":   user_details,
		})
		return
	}

	if req.MobileNumber == 0 {
		utils.Log.Warn("Signup failed: Mobile number missing")
		utils.SendErrorResponse(c, http.StatusBadRequest, "Mobile number is required", "mobile_number field is missing or invalid")
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
		utils.SendErrorResponse(c, http.StatusBadRequest, "Name and Aadhaar number are required", "missing required fields")
		return
	}

	create_user_resp, err := services.CreateUserClient(req, "")
	if err != nil {
		utils.Log.Errorf("Failed to create user in AAA service: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create user in AAA service", err.Error())
		return
	}
	if create_user_resp == nil || create_user_resp.Data == nil || create_user_resp.Data.Id == "" {
		utils.Log.Error("AAA service returned empty or invalid user ID")
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid response from AAA service", "empty user Id in response")
		return
	}
	user_id := create_user_resp.Data.Id

	farmer, user_details, err := h.FarmerService.CreateFarmer(user_id, req)
	if err != nil {
		utils.Log.Errorf("Failed to create farmer record for user %s: %v", user_id, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create farmer record", err.Error())
		return
	}

	if _, err := services.AssignRoleToUserClient(c.Request.Context(), user_id, "FARMER"); err != nil {
		utils.Log.Errorf("Role assignment failed for new user %s: %v", user_id, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Role assignment failed", "invalid user Id format or system error")
		return
	}

	utils.Log.Infof("Farmer registered successfully with newly created user ID: %s", user_id)
	utils.SendSuccessResponse(c, http.StatusCreated, "Farmer registered successfully", gin.H{
		"farmer": farmer,
		"user":   user_details,
	})
}
