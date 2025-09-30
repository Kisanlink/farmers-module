package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/utils"
	"go.uber.org/zap"
)

// ValidationStage validates farmer data
type ValidationStage struct {
	*BasePipelineStage
}

// NewValidationStage creates a new validation stage
func NewValidationStage(logger interfaces.Logger) PipelineStage {
	return &ValidationStage{
		BasePipelineStage: NewBasePipelineStage("validation", 30*time.Second, true, logger),
	}
}

// Process validates the farmer data
func (vs *ValidationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for validation stage")
	}

	farmerData, ok := procCtx.FarmerData.(*requests.FarmerBulkData)
	if !ok {
		return nil, fmt.Errorf("invalid farmer data type")
	}

	vs.logger.Debug("Validating farmer data",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("phone", farmerData.PhoneNumber),
	)

	// Validate required fields
	if err := vs.validateRequiredFields(farmerData); err != nil {
		return nil, fmt.Errorf("required field validation failed: %w", err)
	}

	// Validate field formats
	if err := vs.validateFormats(farmerData); err != nil {
		return nil, fmt.Errorf("format validation failed: %w", err)
	}

	// Validate business rules
	if err := vs.validateBusinessRules(farmerData); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	procCtx.SetStageResult("validation", map[string]interface{}{
		"status":       "success",
		"validated_at": time.Now(),
	})

	return procCtx, nil
}

func (vs *ValidationStage) validateRequiredFields(farmer *requests.FarmerBulkData) error {
	var errors []string

	if farmer.FirstName == "" {
		errors = append(errors, "first_name is required")
	}

	if farmer.LastName == "" {
		errors = append(errors, "last_name is required")
	}

	if farmer.PhoneNumber == "" {
		errors = append(errors, "phone_number is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("missing required fields: %v", errors)
	}

	return nil
}

func (vs *ValidationStage) validateFormats(farmer *requests.FarmerBulkData) error {
	var errors []string

	// Validate phone number format
	if farmer.PhoneNumber != "" && !vs.isValidPhoneNumber(farmer.PhoneNumber) {
		errors = append(errors, "invalid phone number format")
	}

	// Validate email format
	if farmer.Email != "" && !vs.isValidEmail(farmer.Email) {
		errors = append(errors, "invalid email format")
	}

	// Validate gender
	if farmer.Gender != "" && !vs.isValidGender(farmer.Gender) {
		errors = append(errors, "invalid gender (must be male, female, or other)")
	}

	if len(errors) > 0 {
		return fmt.Errorf("format validation errors: %v", errors)
	}

	return nil
}

func (vs *ValidationStage) validateBusinessRules(farmer *requests.FarmerBulkData) error {
	// Add any business-specific validation rules here
	return nil
}

func (vs *ValidationStage) isValidPhoneNumber(phone string) bool {
	// Remove non-digit characters
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// Indian mobile numbers are 10 digits and start with 6-9
	if len(digits) == 10 {
		firstDigit := digits[0]
		return firstDigit >= '6' && firstDigit <= '9'
	}

	return false
}

func (vs *ValidationStage) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (vs *ValidationStage) isValidGender(gender string) bool {
	gender = strings.ToLower(gender)
	return gender == "male" || gender == "female" || gender == "other" || gender == "m" || gender == "f"
}

// FarmerServiceInterface defines the interface for farmer service used by pipeline
type FarmerServiceInterface interface {
	CreateFarmer(ctx context.Context, req *requests.CreateFarmerRequest) (*responses.FarmerResponse, error)
	// Add other methods as needed
}

// AAAServiceInterface defines the interface for AAA service used by pipeline
type AAAServiceInterface interface {
	GetUserByMobile(ctx context.Context, mobile string) (interface{}, error)
	CreateUser(ctx context.Context, req interface{}) (interface{}, error)
	GetOrganization(ctx context.Context, orgID string) (interface{}, error)
}

// FarmerLinkageServiceInterface defines the interface for linkage service used by pipeline
type FarmerLinkageServiceInterface interface {
	LinkFarmerToFPO(ctx context.Context, req interface{}) error
	AssignKisanSathi(ctx context.Context, req interface{}) (interface{}, error)
}

// DeduplicationStage checks for duplicate farmers
type DeduplicationStage struct {
	*BasePipelineStage
	farmerService FarmerServiceInterface
}

// NewDeduplicationStage creates a new deduplication stage
func NewDeduplicationStage(farmerService FarmerServiceInterface, logger interfaces.Logger) PipelineStage {
	return &DeduplicationStage{
		BasePipelineStage: NewBasePipelineStage("deduplication", 10*time.Second, true, logger),
		farmerService:     farmerService,
	}
}

// Process checks for duplicate farmers
func (ds *DeduplicationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for deduplication stage")
	}

	farmerData, ok := procCtx.FarmerData.(*requests.FarmerBulkData)
	if !ok {
		return nil, fmt.Errorf("invalid farmer data type")
	}

	ds.logger.Debug("Checking for duplicates",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("phone", farmerData.PhoneNumber),
	)

	// Check for existing farmer by phone number
	// TODO: Implement actual duplicate checking logic
	// This would involve querying the database for existing farmers with the same phone number

	isDuplicate := false
	duplicateReason := ""
	existingFarmerID := ""

	// Placeholder logic - randomly mark some as duplicates for testing
	if procCtx.RecordIndex%13 == 0 {
		isDuplicate = true
		duplicateReason = "Phone number already exists"
		existingFarmerID = "existing_farmer_123"
	}

	procCtx.SetStageResult("deduplication", map[string]interface{}{
		"is_duplicate":       isDuplicate,
		"duplicate_reason":   duplicateReason,
		"existing_farmer_id": existingFarmerID,
		"checked_at":         time.Now(),
	})

	if isDuplicate {
		return nil, fmt.Errorf("duplicate farmer found: %s", duplicateReason)
	}

	return procCtx, nil
}

// AAAUserCreationStage creates users in the AAA service
type AAAUserCreationStage struct {
	*BasePipelineStage
	aaaService  AAAServiceInterface
	passwordGen *utils.PasswordGenerator
}

// NewAAAUserCreationStage creates a new AAA user creation stage
func NewAAAUserCreationStage(aaaService AAAServiceInterface, logger interfaces.Logger) PipelineStage {
	return &AAAUserCreationStage{
		BasePipelineStage: NewBasePipelineStage("aaa_user_creation", 30*time.Second, true, logger),
		aaaService:        aaaService,
		passwordGen:       utils.NewPasswordGenerator(),
	}
}

// Process creates a user in the AAA service
func (aus *AAAUserCreationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for AAA user creation stage")
	}

	farmerData, ok := procCtx.FarmerData.(*requests.FarmerBulkData)
	if !ok {
		return nil, fmt.Errorf("invalid farmer data type")
	}

	aus.logger.Debug("Creating AAA user",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("phone", farmerData.PhoneNumber),
	)

	// Check if user already exists
	existingUser, err := aus.aaaService.GetUserByMobile(ctx, farmerData.PhoneNumber)
	if err == nil && existingUser != nil {
		// User already exists, extract user ID
		if userMap, ok := existingUser.(map[string]interface{}); ok {
			if userID, exists := userMap["id"]; exists {
				procCtx.SetStageResult("aaa_user_creation", map[string]interface{}{
					"aaa_user_id":  fmt.Sprintf("%v", userID),
					"user_existed": true,
					"created_at":   time.Now(),
				})
				return procCtx, nil
			}
		}
	}

	// Create new user
	createUserReq := map[string]interface{}{
		"username":      fmt.Sprintf("farmer_%s", farmerData.PhoneNumber),
		"mobile_number": farmerData.PhoneNumber,
		"email":         farmerData.Email,
		"password":      farmerData.Password,
		"country_code":  "+91",
		"full_name":     fmt.Sprintf("%s %s", farmerData.FirstName, farmerData.LastName),
	}

	// Use provided password or generate one
	if farmerData.Password == "" {
		createUserReq["password"] = aus.generatePassword(farmerData)
	}

	userResponse, err := aus.aaaService.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create AAA user: %w", err)
	}

	// Extract user ID from response
	var aaaUserID string
	if userMap, ok := userResponse.(map[string]interface{}); ok {
		if id, exists := userMap["id"]; exists {
			aaaUserID = fmt.Sprintf("%v", id)
		}
	}

	if aaaUserID == "" {
		return nil, fmt.Errorf("failed to get AAA user ID from response")
	}

	procCtx.SetStageResult("aaa_user_creation", map[string]interface{}{
		"aaa_user_id":  aaaUserID,
		"user_existed": false,
		"created_at":   time.Now(),
		"username":     createUserReq["username"],
		"password":     createUserReq["password"],
	})

	return procCtx, nil
}

func (aus *AAAUserCreationStage) generatePassword(farmer *requests.FarmerBulkData) string {
	// Generate a cryptographically secure password
	password, err := aus.passwordGen.GenerateSecurePassword()
	if err != nil {
		// Fallback to a more secure pattern if generation fails
		aus.logger.Error("Failed to generate secure password, using fallback",
			zap.Error(err),
			zap.String("phone", farmer.PhoneNumber))

		// Generate a more secure fallback than the original
		return fmt.Sprintf("%s%s!%d",
			strings.ToTitle(farmer.FirstName[:min(3, len(farmer.FirstName))]),
			farmer.PhoneNumber[len(farmer.PhoneNumber)-4:],
			time.Now().Unix()%10000)
	}

	aus.logger.Debug("Generated secure password for farmer",
		zap.String("phone", farmer.PhoneNumber))

	return password
}

// FarmerRegistrationStage registers the farmer in the local database
type FarmerRegistrationStage struct {
	*BasePipelineStage
	farmerService FarmerServiceInterface
}

// NewFarmerRegistrationStage creates a new farmer registration stage
func NewFarmerRegistrationStage(farmerService FarmerServiceInterface, logger interfaces.Logger) PipelineStage {
	return &FarmerRegistrationStage{
		BasePipelineStage: NewBasePipelineStage("farmer_registration", 20*time.Second, true, logger),
		farmerService:     farmerService,
	}
}

// Process registers the farmer in the local database
func (frs *FarmerRegistrationStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for farmer registration stage")
	}

	farmerData, ok := procCtx.FarmerData.(*requests.FarmerBulkData)
	if !ok {
		return nil, fmt.Errorf("invalid farmer data type")
	}

	// Get AAA user ID from previous stage
	aaaResult, exists := procCtx.GetStageResult("aaa_user_creation")
	if !exists {
		return nil, fmt.Errorf("AAA user creation result not found")
	}

	aaaResultMap, ok := aaaResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid AAA user creation result format")
	}

	aaaUserID, ok := aaaResultMap["aaa_user_id"].(string)
	if !ok || aaaUserID == "" {
		return nil, fmt.Errorf("AAA user ID not found in result")
	}

	frs.logger.Debug("Registering farmer",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("aaa_user_id", aaaUserID),
		zap.String("phone", farmerData.PhoneNumber),
	)

	// Create farmer registration request
	createFarmerReq := &requests.CreateFarmerRequest{
		AAAUserID: aaaUserID,
		AAAOrgID:  procCtx.FPOOrgID,
		Profile: requests.FarmerProfileData{
			FirstName:   farmerData.FirstName,
			LastName:    farmerData.LastName,
			PhoneNumber: farmerData.PhoneNumber,
			Email:       farmerData.Email,
			DateOfBirth: farmerData.DateOfBirth,
			Gender:      farmerData.Gender,
			Address: requests.AddressData{
				StreetAddress: farmerData.StreetAddress,
				City:          farmerData.City,
				State:         farmerData.State,
				PostalCode:    farmerData.PostalCode,
				Country:       farmerData.Country,
			},
			Preferences: farmerData.CustomFields,
		},
	}

	// Register farmer
	farmerResponse, err := frs.farmerService.CreateFarmer(ctx, createFarmerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register farmer: %w", err)
	}

	// Extract farmer ID from response
	farmerID := "" // TODO: Extract from farmerResponse
	if farmerResponse != nil {
		// Assuming the response has a farmer ID field
		farmerID = "farmer_" + aaaUserID // Placeholder
	}

	procCtx.SetStageResult("farmer_registration", map[string]interface{}{
		"farmer_id":     farmerID,
		"aaa_user_id":   aaaUserID,
		"registered_at": time.Now(),
	})

	return procCtx, nil
}

// FPOLinkageStage links the farmer to the FPO
type FPOLinkageStage struct {
	*BasePipelineStage
	linkageService FarmerLinkageServiceInterface
}

// NewFPOLinkageStage creates a new FPO linkage stage
func NewFPOLinkageStage(linkageService FarmerLinkageServiceInterface, logger interfaces.Logger) PipelineStage {
	return &FPOLinkageStage{
		BasePipelineStage: NewBasePipelineStage("fpo_linkage", 15*time.Second, true, logger),
		linkageService:    linkageService,
	}
}

// Process links the farmer to the FPO
func (fls *FPOLinkageStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for FPO linkage stage")
	}

	// Get AAA user ID from farmer registration stage
	registrationResult, exists := procCtx.GetStageResult("farmer_registration")
	if !exists {
		return nil, fmt.Errorf("farmer registration result not found")
	}

	registrationMap, ok := registrationResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid farmer registration result format")
	}

	aaaUserID, ok := registrationMap["aaa_user_id"].(string)
	if !ok || aaaUserID == "" {
		return nil, fmt.Errorf("AAA user ID not found in registration result")
	}

	fls.logger.Debug("Linking farmer to FPO",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("aaa_user_id", aaaUserID),
		zap.String("fpo_org_id", procCtx.FPOOrgID),
	)

	// Create linkage request
	linkageReq := &requests.LinkFarmerRequest{
		AAAUserID: aaaUserID,
		AAAOrgID:  procCtx.FPOOrgID,
	}

	// Link farmer to FPO
	err := fls.linkageService.LinkFarmerToFPO(ctx, linkageReq)
	if err != nil {
		return nil, fmt.Errorf("failed to link farmer to FPO: %w", err)
	}

	procCtx.SetStageResult("fpo_linkage", map[string]interface{}{
		"linked_at":   time.Now(),
		"aaa_user_id": aaaUserID,
		"fpo_org_id":  procCtx.FPOOrgID,
		"link_status": "ACTIVE",
	})

	return procCtx, nil
}

// KisanSathiAssignmentStage assigns a KisanSathi to the farmer
type KisanSathiAssignmentStage struct {
	*BasePipelineStage
	linkageService   FarmerLinkageServiceInterface
	kisanSathiUserID *string
}

// NewKisanSathiAssignmentStage creates a new KisanSathi assignment stage
func NewKisanSathiAssignmentStage(linkageService FarmerLinkageServiceInterface, kisanSathiUserID *string, logger interfaces.Logger) PipelineStage {
	return &KisanSathiAssignmentStage{
		BasePipelineStage: NewBasePipelineStage("kisan_sathi_assignment", 10*time.Second, true, logger),
		linkageService:    linkageService,
		kisanSathiUserID:  kisanSathiUserID,
	}
}

// Process assigns a KisanSathi to the farmer
func (ksas *KisanSathiAssignmentStage) Process(ctx context.Context, data interface{}) (interface{}, error) {
	procCtx, ok := data.(*ProcessingContext)
	if !ok {
		return nil, fmt.Errorf("invalid data type for KisanSathi assignment stage")
	}

	// Skip if no KisanSathi to assign
	if ksas.kisanSathiUserID == nil || *ksas.kisanSathiUserID == "" {
		procCtx.SetStageResult("kisan_sathi_assignment", map[string]interface{}{
			"assigned":   false,
			"reason":     "No KisanSathi specified",
			"skipped_at": time.Now(),
		})
		return procCtx, nil
	}

	// Get farmer linkage info from previous stage
	linkageResult, exists := procCtx.GetStageResult("fpo_linkage")
	if !exists {
		return nil, fmt.Errorf("FPO linkage result not found")
	}

	linkageMap, ok := linkageResult.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid FPO linkage result format")
	}

	aaaUserID, ok := linkageMap["aaa_user_id"].(string)
	if !ok || aaaUserID == "" {
		return nil, fmt.Errorf("AAA user ID not found in linkage result")
	}

	ksas.logger.Debug("Assigning KisanSathi",
		zap.String("operation_id", procCtx.OperationID),
		zap.Int("record_index", procCtx.RecordIndex),
		zap.String("aaa_user_id", aaaUserID),
		zap.String("kisan_sathi_user_id", *ksas.kisanSathiUserID),
	)

	// Create assignment request
	assignmentReq := &requests.AssignKisanSathiRequest{
		AAAUserID:        aaaUserID,
		AAAOrgID:         procCtx.FPOOrgID,
		KisanSathiUserID: *ksas.kisanSathiUserID,
	}

	// Assign KisanSathi
	_, err := ksas.linkageService.AssignKisanSathi(ctx, assignmentReq)
	if err != nil {
		// Log warning but don't fail the entire pipeline
		ksas.logger.Warn("Failed to assign KisanSathi",
			zap.String("operation_id", procCtx.OperationID),
			zap.Int("record_index", procCtx.RecordIndex),
			zap.Error(err),
		)

		procCtx.SetStageResult("kisan_sathi_assignment", map[string]interface{}{
			"assigned":     false,
			"error":        err.Error(),
			"attempted_at": time.Now(),
		})
		return procCtx, nil
	}

	procCtx.SetStageResult("kisan_sathi_assignment", map[string]interface{}{
		"assigned":            true,
		"kisan_sathi_user_id": *ksas.kisanSathiUserID,
		"assigned_at":         time.Now(),
	})

	return procCtx, nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
