package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *gin.Context, status int, userMessage string, errorDetail string) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   userMessage,
		"error":     errorDetail,
		"timestamp": time.Now().UTC(),
		"data":      nil,
		"success":   false,
	})
}

// SendSuccessResponse sends a standardized success response
func SendSuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, gin.H{
		"status":    status,
		"message":   message,
		"error":     nil,
		"timestamp": time.Now().UTC(),
		"data":      data,
		"success":   true,
	})
}
