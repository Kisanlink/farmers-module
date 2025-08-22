package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware provides comprehensive validation and AAA service integration
type ValidationMiddleware struct {
	aaaClient *aaa.Client
	config    *config.Config
	validate  *validator.Validate
}

// NewValidationMiddleware creates a new validation middleware instance
func NewValidationMiddleware(aaaClient *aaa.Client, config *config.Config) *ValidationMiddleware {
	validate := validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("phone", validatePhone); err != nil {
		log.Printf("Warning: Failed to register phone validator: %v", err)
	}
	if err := validate.RegisterValidation("email", validateEmail); err != nil {
		log.Printf("Warning: Failed to register email validator: %v", err)
	}
	if err := validate.RegisterValidation("org_type", validateOrgType); err != nil {
		log.Printf("Warning: Failed to register org_type validator: %v", err)
	}

	return &ValidationMiddleware{
		aaaClient: aaaClient,
		config:    config,
		validate:  validate,
	}
}

// ValidateFPOCreation validates FPO creation requests and ensures AAA service consistency
func (vm *ValidationMiddleware) ValidateFPOCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract request data for validation
		var requestData map[string]interface{}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate required fields
		if err := vm.validateFPOFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated data in context for handlers
		c.Set("validated_fpo_data", requestData)
		c.Next()
	}
}

// ValidateFarmerCreation validates farmer creation requests and ensures AAA service consistency
func (vm *ValidationMiddleware) ValidateFarmerCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract request data for validation
		var requestData map[string]interface{}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate required fields
		if err := vm.validateFarmerFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verify organization exists in AAA service
		if orgID, exists := requestData["aaa_org_id"]; exists {
			if err := vm.verifyOrganization(c.Request.Context(), fmt.Sprintf("%v", orgID)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Organization validation failed",
					"message": err.Error(),
				})
				c.Abort()
				return
			}
		}

		// Store validated data in context for handlers
		c.Set("validated_farmer_data", requestData)
		c.Next()
	}
}

// ValidateOrganizationAccess validates if the requesting user has access to the specified organization
func (vm *ValidationMiddleware) ValidateOrganizationAccess(orgIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param(orgIDParam)
		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Organization ID required",
				"message": "Organization ID parameter is missing",
			})
			c.Abort()
			return
		}

		// Extract user ID from context (this would come from authentication middleware)
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// Check if user has permission to access this organization
		hasPermission, err := vm.aaaClient.CheckPermission(
			c.Request.Context(),
			userID,
			"organization",
			"read",
			orgID,
			orgID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Permission check failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "User does not have permission to access this organization",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateFarmCreation validates farm creation requests
func (vm *ValidationMiddleware) ValidateFarmCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData map[string]interface{}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate required fields
		if err := vm.validateFarmFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verify organization exists in AAA service
		if orgID, exists := requestData["aaa_org_id"]; exists {
			if err := vm.verifyOrganization(c.Request.Context(), fmt.Sprintf("%v", orgID)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Organization validation failed",
					"message": err.Error(),
				})
				c.Abort()
				return
			}
		}

		// Store validated data in context for handlers
		c.Set("validated_farm_data", requestData)
		c.Next()
	}
}

// validateFPOFields validates FPO-specific fields
func (vm *ValidationMiddleware) validateFPOFields(data map[string]interface{}) error {
	requiredFields := []string{"name", "type"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate organization type
	if orgType, exists := data["type"]; exists {
		if err := vm.validate.Var(orgType, "org_type"); err != nil {
			return fmt.Errorf("invalid organization type: %v", orgType)
		}
	}

	return nil
}

// validateFarmerFields validates farmer-specific fields
func (vm *ValidationMiddleware) validateFarmerFields(data map[string]interface{}) error {
	requiredFields := []string{"aaa_user_id", "aaa_org_id"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate profile fields if present
	if profile, exists := data["profile"]; exists {
		if profileMap, ok := profile.(map[string]interface{}); ok {
			if err := vm.validateFarmerProfile(profileMap); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFarmFields validates farm-specific fields
func (vm *ValidationMiddleware) validateFarmFields(data map[string]interface{}) error {
	requiredFields := []string{"aaa_user_id", "aaa_org_id", "name", "location"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate area if present
	if area, exists := data["area"]; exists {
		if areaFloat, ok := area.(float64); ok {
			if areaFloat <= 0 {
				return fmt.Errorf("farm area must be greater than 0")
			}
		}
	}

	return nil
}

// validateFarmerProfile validates farmer profile data
func (vm *ValidationMiddleware) validateFarmerProfile(profile map[string]interface{}) error {
	// Validate email if present
	if email, exists := profile["email"]; exists && email != "" {
		if err := vm.validate.Var(email, "email"); err != nil {
			return fmt.Errorf("invalid email format: %v", email)
		}
	}

	// Validate phone if present
	if phone, exists := profile["phone_number"]; exists && phone != "" {
		if err := vm.validate.Var(phone, "phone"); err != nil {
			return fmt.Errorf("invalid phone format: %v", phone)
		}
	}

	return nil
}

// verifyOrganization verifies if an organization exists in AAA service
func (vm *ValidationMiddleware) verifyOrganization(ctx context.Context, orgID string) error {
	org, err := vm.aaaClient.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to verify organization %s: %w", orgID, err)
	}

	if org.Status != "ACTIVE" {
		return fmt.Errorf("organization %s is not active (status: %s)", orgID, org.Status)
	}

	return nil
}

// Custom validators
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Basic phone validation - can be enhanced
	return len(phone) >= 10 && strings.Contains(phone, "+")
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	// Basic email validation - can be enhanced
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func validateOrgType(fl validator.FieldLevel) bool {
	orgType := fl.Field().String()
	validTypes := []string{"fpo", "cooperative", "association", "company"}
	for _, validType := range validTypes {
		if orgType == validType {
			return true
		}
	}
	return false
}
