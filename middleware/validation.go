package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

// ValidateCropCreation validates crop creation requests
func (vm *ValidationMiddleware) ValidateCropCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData map[string]interface{}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			log.Printf("DEBUG: JSON parsing error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		log.Printf("DEBUG: Parsed request data: %+v", requestData)

		// Validate required fields
		if err := vm.validateCropFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated data in context for handlers
		c.Set("validated_crop_data", requestData)
		c.Next()
	}
}

// ValidateCropCycleCreation validates crop cycle creation requests
func (vm *ValidationMiddleware) ValidateCropCycleCreation() gin.HandlerFunc {
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
		if err := vm.validateCropCycleFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated data in context for handlers
		c.Set("validated_crop_cycle_data", requestData)
		c.Next()
	}
}

// ValidateCropVarietyCreation validates crop variety creation requests
func (vm *ValidationMiddleware) ValidateCropVarietyCreation() gin.HandlerFunc {
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
		if err := vm.validateCropVarietyFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated data in context for handlers
		c.Set("validated_crop_variety_data", requestData)
		c.Next()
	}
}

// ValidateFarmActivityCreation validates farm activity creation requests
func (vm *ValidationMiddleware) ValidateFarmActivityCreation() gin.HandlerFunc {
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
		if err := vm.validateFarmActivityFields(requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Store validated data in context for handlers
		c.Set("validated_farm_activity_data", requestData)
		c.Next()
	}
}

// ValidateCropLookupData validates crop lookup data requests
func (vm *ValidationMiddleware) ValidateCropLookupData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For GET requests, validate query parameters
		category := c.Query("category")
		season := c.Query("season")

		// Validate category if provided
		if category != "" {
			validCategories := []string{"CEREALS", "PULSES", "VEGETABLES", "FRUITS", "OIL_SEEDS", "SPICES", "CASH_CROPS", "FODDER", "MEDICINAL", "OTHER"}
			valid := false
			for _, validCategory := range validCategories {
				if category == validCategory {
					valid = true
					break
				}
			}
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid category",
					"message": fmt.Sprintf("Category must be one of: %v", validCategories),
				})
				c.Abort()
				return
			}
		}

		// Validate season if provided
		if season != "" {
			validSeasons := []string{"RABI", "KHARIF", "ZAID"}
			valid := false
			for _, validSeason := range validSeasons {
				if season == validSeason {
					valid = true
					break
				}
			}
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid season",
					"message": fmt.Sprintf("Season must be one of: %v", validSeasons),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// validateCropFields validates crop-specific fields
func (vm *ValidationMiddleware) validateCropFields(data map[string]interface{}) error {
	requiredFields := []string{"name", "category", "seasons", "unit"}

	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}

		// Special handling for different field types
		switch field {
		case "seasons":
			// For seasons array, check if it's empty
			if seasons, ok := value.([]interface{}); !ok || len(seasons) == 0 {
				return fmt.Errorf("required field '%s' is missing or empty", field)
			}
		default:
			// For string fields, check if empty
			if strValue, ok := value.(string); !ok || strValue == "" {
				return fmt.Errorf("required field '%s' is missing or empty", field)
			}
		}
	}

	// Validate category
	if category, exists := data["category"]; exists {
		validCategories := []string{"CEREALS", "PULSES", "VEGETABLES", "FRUITS", "OIL_SEEDS", "SPICES", "CASH_CROPS", "FODDER", "MEDICINAL", "OTHER"}
		valid := false
		for _, validCategory := range validCategories {
			if category == validCategory {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid category: %v", category)
		}
	}

	// Validate seasons
	if seasons, exists := data["seasons"]; exists {
		if seasonsSlice, ok := seasons.([]interface{}); ok {
			validSeasons := []string{"RABI", "KHARIF", "ZAID"}
			for _, season := range seasonsSlice {
				if seasonStr, ok := season.(string); ok {
					valid := false
					for _, validSeason := range validSeasons {
						if seasonStr == validSeason {
							valid = true
							break
						}
					}
					if !valid {
						return fmt.Errorf("invalid season: %v", seasonStr)
					}
				}
			}
		}
	}

	// Validate duration if present
	if duration, exists := data["duration_days"]; exists {
		if durationFloat, ok := duration.(float64); ok {
			if durationFloat < 1 || durationFloat > 365 {
				return fmt.Errorf("duration_days must be between 1 and 365")
			}
		}
	}

	return nil
}

// validateCropCycleFields validates crop cycle-specific fields
func (vm *ValidationMiddleware) validateCropCycleFields(data map[string]interface{}) error {
	requiredFields := []string{"farm_id", "season", "start_date", "crop_id"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate season
	if season, exists := data["season"]; exists {
		validSeasons := []string{"RABI", "KHARIF", "ZAID"}
		valid := false
		for _, validSeason := range validSeasons {
			if season == validSeason {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid season: %v", season)
		}
	}

	// Validate start_date format
	if startDate, exists := data["start_date"]; exists {
		if startDateStr, ok := startDate.(string); ok {
			if _, err := time.Parse(time.RFC3339, startDateStr); err != nil {
				return fmt.Errorf("invalid start_date format: %v", startDateStr)
			}
		}
	}

	return nil
}

// validateCropVarietyFields validates crop variety-specific fields
func (vm *ValidationMiddleware) validateCropVarietyFields(data map[string]interface{}) error {
	requiredFields := []string{"crop_id", "name"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate duration if present
	if duration, exists := data["duration_days"]; exists {
		if durationFloat, ok := duration.(float64); ok {
			if durationFloat < 1 || durationFloat > 365 {
				return fmt.Errorf("duration_days must be between 1 and 365")
			}
		}
	}

	// Validate yield if present
	if yield, exists := data["yield_per_acre"]; exists {
		if yieldFloat, ok := yield.(float64); ok {
			if yieldFloat < 0 {
				return fmt.Errorf("yield_per_acre must be non-negative")
			}
		}
	}

	return nil
}

// validateFarmActivityFields validates farm activity-specific fields
func (vm *ValidationMiddleware) validateFarmActivityFields(data map[string]interface{}) error {
	requiredFields := []string{"crop_cycle_id", "activity_type"}

	for _, field := range requiredFields {
		if value, exists := data[field]; !exists || value == "" {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}

	// Validate activity_type
	if activityType, exists := data["activity_type"]; exists {
		validTypes := []string{"SOWING", "IRRIGATION", "FERTILIZATION", "PEST_CONTROL", "HARVESTING", "PLOUGHING", "WEEDING", "PRUNING", "OTHER"}
		valid := false
		for _, validType := range validTypes {
			if activityType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid activity_type: %v", activityType)
		}
	}

	// Validate planned_at format if present
	if plannedAt, exists := data["planned_at"]; exists {
		if plannedAtStr, ok := plannedAt.(string); ok {
			if _, err := time.Parse(time.RFC3339, plannedAtStr); err != nil {
				return fmt.Errorf("invalid planned_at format: %v", plannedAtStr)
			}
		}
	}

	return nil
}
