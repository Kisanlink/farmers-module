package permission

import (
	"context"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
	"net/http"
)

// CheckUserPermission returns true if permission granted, else returns statusCode, userMessage, errorDetail
func CheckUserPermission(ctx context.Context, userId string, requiredPermission string) (bool, int, string, string) {
	userResp, err := services.GetUserByIdClient(ctx, userId)
	if err != nil {
		utils.Log.Errorf("Error fetching user %s: %v", userId, err)
		return false, http.StatusInternalServerError, "Failed to verify user", err.Error()
	}

	if userResp == nil || userResp.StatusCode != http.StatusOK || userResp.Data == nil {
		utils.Log.Warnf("User %s not found or invalid response: %+v", userId, userResp)
		return false, http.StatusUnauthorized, "User not found", "invalid user response"
	}

	if len(userResp.Data.RolePermissions) == 0 {
		utils.Log.Warnf("User %s has no role permissions", userId)
		return false, http.StatusForbidden, "Permission denied", "user has no role permissions defined"
	}

	for _, rolePerm := range userResp.Data.RolePermissions {
		for _, perm := range rolePerm.Permissions {
			if perm.Name == requiredPermission {
				return true, 0, "", "" // Permission found ✅
			}
		}
	}

	utils.Log.Warnf("User %s lacks permission: %s", userId, requiredPermission)
	return false, http.StatusForbidden, "Permission denied", "missing required permissions or actions"
}
