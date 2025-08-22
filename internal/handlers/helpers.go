package handlers

import (
	"strings"

	"github.com/google/uuid"
)

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return uuid.New().String()
}

// isValidationError checks if the error is a validation error
func isValidationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "format")
}

// isPermissionError checks if the error is a permission error
func isPermissionError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "permission") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "access denied")
}

// isNotFoundError checks if the error is a not found error
func isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist")
}
