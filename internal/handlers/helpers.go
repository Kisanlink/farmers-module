package handlers

import (
	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
)

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return common.GenerateRequestID()
}

// isValidationError checks if the error is a validation error
func isValidationError(err error) bool {
	return common.IsValidationError(err)
}

// isPermissionError checks if the error is a permission error
func isPermissionError(err error) bool {
	return common.IsPermissionError(err)
}

// isNotFoundError checks if the error is a not found error
func isNotFoundError(err error) bool {
	return common.IsNotFoundError(err)
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	return common.ParseIntQuery(c, key, defaultValue)
}

// handleServiceError converts service errors to appropriate HTTP responses
func handleServiceError(c *gin.Context, err error) {
	common.HandleServiceError(c, err)
}

// getUserContext extracts user and org context from gin context
func getUserContext(c *gin.Context) (userID, orgID string) {
	// Extract user context
	if userCtx, exists := c.Get("user_context"); exists {
		if uc, ok := userCtx.(*auth.UserContext); ok && uc != nil {
			userID = uc.AAAUserID
		}
	}

	// Extract org context
	if orgCtx, exists := c.Get("org_context"); exists {
		if oc, ok := orgCtx.(*auth.OrgContext); ok && oc != nil {
			orgID = oc.AAAOrgID
		}
	}

	return userID, orgID
}
