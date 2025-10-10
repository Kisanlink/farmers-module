package common

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "format")
}

// IsPermissionError checks if the error is a permission error
func IsPermissionError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "permission") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "access denied")
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist")
}

// ParseIntQuery parses an integer query parameter with a default value
func ParseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// HandleServiceError converts service errors to appropriate HTTP responses
func HandleServiceError(c *gin.Context, err error) {
	// Log the error for debugging
	log.Printf("[ERROR] Service error: %v", err)

	switch err {
	case ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	case ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "Access forbidden"})
	case ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	case ErrInvalidInput, ErrInvalidFarmData, ErrInvalidFarmGeometry,
		ErrInvalidCropCycleData, ErrInvalidFarmActivityData:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		// Handle specific error messages
		errMsg := err.Error()

		// Not found errors
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
			// Validation errors
		} else if errMsg == "FPO name is required" ||
			errMsg == "FPO registration number is required" ||
			errMsg == "CEO user details are required" ||
			errMsg == "CEO phone number is required" ||
			errMsg == "AAA organization ID is required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			// Conflict errors
		} else if errMsg == "failed to create CEO user: user already exists" ||
			strings.Contains(errMsg, "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
			// Permission/Authorization errors
		} else if strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "forbidden") ||
			strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
			// Generic service errors that should be returned as-is
		} else if strings.Contains(errMsg, "Failed to") || strings.Contains(errMsg, "failed to") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}
}
