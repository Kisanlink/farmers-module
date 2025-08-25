package handlers

import (
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
