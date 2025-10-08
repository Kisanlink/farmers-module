package handlers

import (
	"encoding/json"
	"fmt"

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

// convertValidatedData converts map[string]interface{} from context to a specific request struct
func convertValidatedData(data interface{}, target interface{}) error {
	// Convert map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal validated data: %w", err)
	}

	// Convert JSON to target struct
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target struct: %w", err)
	}

	return nil
}
