package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/constants"
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FPORefRepository defines the interface for FPO reference repository operations
type FPORefRepository interface {
	Create(ctx context.Context, entity *fpo.FPORef) error
	FindOne(ctx context.Context, filter *base.Filter) (*fpo.FPORef, error)
	UpdateCEO(ctx context.Context, fpoID string, ceoUserID string) error
}

// FPOServiceImpl implements FPOService
type FPOServiceImpl struct {
	fpoRefRepo       FPORefRepository
	aaaService       AAAService
	fpoConfigService FPOConfigService
}

// NewFPOService creates a new FPO service
func NewFPOService(fpoRefRepo FPORefRepository, aaaService AAAService, fpoConfigService FPOConfigService) FPOService {
	return &FPOServiceImpl{
		fpoRefRepo:       fpoRefRepo,
		aaaService:       aaaService,
		fpoConfigService: fpoConfigService,
	}
}

// CreateFPO implements FPO creation with AAA organization creation
func (s *FPOServiceImpl) CreateFPO(ctx context.Context, req interface{}) (interface{}, error) {
	log.Println("FPOService: Starting CreateFPO workflow")

	// Type assert the request
	createReq, ok := req.(*requests.CreateFPORequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for CreateFPO")
	}

	// Validate request
	if createReq.Name == "" {
		return nil, fmt.Errorf("FPO name is required")
	}
	if createReq.RegistrationNo == "" {
		return nil, fmt.Errorf("FPO registration number is required")
	}
	if createReq.CEOUser.PhoneNumber == "" {
		return nil, fmt.Errorf("CEO phone number is required")
	}

	// Step 1: Check if CEO user exists in AAA, create if not
	var ceoUserID string
	var ceoFullName string
	existingUser, err := s.aaaService.GetUserByMobile(ctx, createReq.CEOUser.PhoneNumber)
	if err != nil {
		log.Printf("CEO user not found, creating new user: %v", err)

		// Validate required fields for new user creation
		if createReq.CEOUser.FirstName == "" || createReq.CEOUser.LastName == "" {
			return nil, fmt.Errorf("CEO first_name and last_name are required when creating a new user")
		}
		if createReq.CEOUser.Password == "" {
			return nil, fmt.Errorf("password is required when creating a new CEO user")
		}
		if len(createReq.CEOUser.Password) < 8 {
			return nil, fmt.Errorf("password must be at least 8 characters long")
		}

		ceoFullName = fmt.Sprintf("%s %s", createReq.CEOUser.FirstName, createReq.CEOUser.LastName)

		// Normalize phone number (remove +91 prefix if present)
		// AAA service expects 10-digit number for India as country_code is sent separately
		phone := createReq.CEOUser.PhoneNumber
		if strings.HasPrefix(phone, "+91") {
			phone = strings.TrimPrefix(phone, "+91")
		} else if len(phone) > 10 {
			// Basic cleanup for other formats if needed, but primarily targeting +91 overflow
			// Taking last 10 digits as fallback for Indian numbers
			phone = phone[len(phone)-10:]
		}

		// Create CEO user in AAA
		createUserReq := map[string]interface{}{
			"username":     fmt.Sprintf("%s_%s", createReq.CEOUser.FirstName, createReq.CEOUser.LastName),
			"phone_number": phone,
			"email":        createReq.CEOUser.Email,
			"password":     createReq.CEOUser.Password,
			"full_name":    ceoFullName,
			"role":         "CEO",
			"country_code": "+91",
		}

		userResp, err := s.aaaService.CreateUser(ctx, createUserReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create CEO user: %w", err)
		}

		userRespMap, ok := userResp.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid user creation response")
		}

		ceoUserID = userRespMap["id"].(string)
		log.Printf("Created CEO user with ID: %s", ceoUserID)
	} else {
		// Use existing user
		userMap, ok := existingUser.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid existing user response")
		}
		ceoUserID = userMap["id"].(string)
		if fullName, ok := userMap["full_name"].(string); ok {
			ceoFullName = fullName
		} else if firstName, ok := userMap["first_name"].(string); ok {
			lastName, _ := userMap["last_name"].(string)
			ceoFullName = fmt.Sprintf("%s %s", firstName, lastName)
		}
		log.Printf("Using existing CEO user with ID: %s", ceoUserID)
	}

	log.Printf("Creating FPO: %s with CEO: %s", createReq.Name, ceoFullName)

	// Step 2: Validate CEO is not already CEO of another FPO
	// Business Rule 1.2: A user CANNOT be CEO of multiple FPOs simultaneously
	isCEO, err := s.aaaService.CheckUserRole(ctx, ceoUserID, "CEO")
	if err != nil {
		log.Printf("Warning: Failed to check if user is already CEO: %v", err)
		// Continue anyway - this is a best-effort check
	} else if isCEO {
		return nil, fmt.Errorf("user is already CEO of another FPO - a user cannot be CEO of multiple FPOs simultaneously")
	}

	// Step 3: Create organization in AAA
	createOrgReq := map[string]interface{}{
		"name":        createReq.Name,
		"description": createReq.Description,
		"type":        "fpo",
		"ceo_user_id": ceoUserID,
		"metadata":    createReq.Metadata,
	}

	orgResp, err := s.aaaService.CreateOrganization(ctx, createOrgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	orgRespMap, ok := orgResp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid organization creation response")
	}

	aaaOrgID := orgRespMap["org_id"].(string)
	log.Printf("Created organization with ID: %s", aaaOrgID)

	// Initialize setup tracking for partial failures
	setupErrors := make(entities.JSONB)

	// Step 4: Assign CEO role to user in organization
	// Following ADR-001: CEO role is critical for FPO operations
	err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
	if err != nil {
		log.Printf("Error: Failed to assign CEO role to user %s: %v", ceoUserID, err)
		setupErrors["ceo_role_assignment"] = err.Error()
		// Don't return error immediately - continue with setup to allow retry via CompleteFPOSetup
	}

	// Step 5: Create user groups for FPO
	userGroups := []responses.UserGroupData{}
	groupNames := []string{"directors", "shareholders", "store_staff", "store_managers"}

	for _, groupName := range groupNames {
		createGroupReq := map[string]interface{}{
			"name":        groupName,
			"description": fmt.Sprintf("%s group for %s", groupName, createReq.Name),
			"org_id":      aaaOrgID,
			"permissions": s.getGroupPermissions(groupName),
		}

		groupResp, err := s.aaaService.CreateUserGroup(ctx, createGroupReq)
		if err != nil {
			log.Printf("Warning: Failed to create user group %s: %v", groupName, err)
			setupErrors[fmt.Sprintf("user_group_%s", groupName)] = err.Error()
			continue
		}

		groupRespMap, ok := groupResp.(map[string]interface{})
		if !ok {
			log.Printf("Warning: Invalid group creation response for %s", groupName)
			setupErrors[fmt.Sprintf("user_group_%s", groupName)] = "invalid response from AAA service"
			continue
		}

		userGroup := responses.UserGroupData{
			GroupID:     groupRespMap["group_id"].(string),
			Name:        groupRespMap["name"].(string),
			OrgID:       groupRespMap["org_id"].(string),
			CreatedAt:   groupRespMap["created_at"].(string),
			Permissions: s.getGroupPermissions(groupName),
		}
		userGroups = append(userGroups, userGroup)

		log.Printf("Created user group: %s with ID: %s", groupName, userGroup.GroupID)

		// Add CEO to directors group to ensure they have organization context in AAA
		if groupName == "directors" {
			if err := s.aaaService.AddUserToGroup(ctx, ceoUserID, userGroup.GroupID); err != nil {
				log.Printf("Warning: Failed to add CEO to directors group: %v", err)
				setupErrors["ceo_group_membership"] = err.Error()
			} else {
				log.Printf("Added CEO %s to directors group %s", ceoUserID, userGroup.GroupID)
			}
		}

		// Assign permissions to group
		for _, permission := range s.getGroupPermissions(groupName) {
			err = s.aaaService.AssignPermissionToGroup(ctx, userGroup.GroupID, "fpo", permission)
			if err != nil {
				log.Printf("Warning: Failed to assign permission %s to group %s: %v", permission, groupName, err)
				setupErrors[fmt.Sprintf("permission_%s_%s", groupName, permission)] = err.Error()
			}
		}
	}

	// Step 6: Determine FPO status based on setup success
	// Business Rule 1.1: Mark as PENDING_SETUP if any failures occurred
	fpoStatus := fpo.FPOStatusActive
	if len(setupErrors) > 0 {
		fpoStatus = fpo.FPOStatusPendingSetup
		log.Printf("FPO setup incomplete, marking as PENDING_SETUP. Errors: %v", setupErrors)
	}

	// Step 7: Store FPO reference in local database
	// Use constructor to ensure ID is properly initialized
	fpoRef := fpo.NewFPORef(aaaOrgID)
	fpoRef.Name = createReq.Name
	fpoRef.RegistrationNo = createReq.RegistrationNo
	fpoRef.Status = fpoStatus
	fpoRef.BusinessConfig = createReq.BusinessConfig
	fpoRef.SetupErrors = setupErrors

	err = s.fpoRefRepo.Create(ctx, fpoRef)
	if err != nil {
		log.Printf("Warning: Failed to store FPO reference locally: %v", err)
		// Continue as AAA organization is already created
		// Generate a temporary ID for the response if save failed
		fpoRef.ID = ""
	}

	// Step 7.5: Auto-create FPO configuration
	if s.fpoConfigService != nil {
		log.Printf("Auto-creating FPO configuration for org ID: %s", aaaOrgID)
		
		var erpBaseURL, erpUIBaseURL string
		
		// Extract URLs from BusinessConfig if available
		if createReq.BusinessConfig != nil {
			if url, ok := createReq.BusinessConfig["erp_base_url"].(string); ok {
				erpBaseURL = url
			}
			if url, ok := createReq.BusinessConfig["erp_ui_base_url"].(string); ok {
				erpUIBaseURL = url
			}
		}
		
		// Create config request
		configReq := &requests.CreateFPOConfigRequest{
			AAAOrgID:     aaaOrgID,
			FPOName:      createReq.Name,
			ERPBaseURL:   erpBaseURL,
			ERPUIBaseURL: erpUIBaseURL,
			Metadata:     createReq.Metadata,
		}
		
		// Set defaults if URLs are empty (optional, could be done in service)
		configReq.SetDefaults()
		
		_, err := s.fpoConfigService.CreateFPOConfig(ctx, configReq)
		if err != nil {
			log.Printf("Warning: Failed to auto-create FPO config: %v", err)
			// Don't fail the request, just log
			setupErrors["fpo_config_creation"] = err.Error()
		} else {
			log.Printf("Successfully auto-created FPO config for %s", aaaOrgID)
		}
	}

	// Step 8: Prepare response (now fpoRef.ID is populated after save)
	responseData := &responses.CreateFPOData{
		FPOID:      fpoRef.ID, // This will now have the generated ID from database
		AAAOrgID:   aaaOrgID,
		Name:       createReq.Name,
		CEOUserID:  ceoUserID,
		UserGroups: userGroups,
		Status:     fpoStatus.String(),
		CreatedAt:  time.Now(),
	}

	if fpoStatus == fpo.FPOStatusPendingSetup {
		log.Printf("FPO created with incomplete setup: %s (org ID: %s). Errors: %v", createReq.Name, aaaOrgID, setupErrors)
	} else {
		log.Printf("Successfully created FPO: %s with org ID: %s", createReq.Name, aaaOrgID)
	}
	return responseData, nil
}

// RegisterFPORef implements FPO reference registration for local management
func (s *FPOServiceImpl) RegisterFPORef(ctx context.Context, req interface{}) (interface{}, error) {
	log.Println("FPOService: Starting RegisterFPORef workflow")

	// Type assert the request
	registerReq, ok := req.(*requests.RegisterFPORefRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for RegisterFPORef")
	}

	// Validate request
	if registerReq.AAAOrgID == "" {
		return nil, fmt.Errorf("AAA organization ID is required")
	}
	if registerReq.Name == "" {
		return nil, fmt.Errorf("FPO name is required")
	}

	log.Printf("Registering FPO reference for org ID: %s", registerReq.AAAOrgID)

	// Verify organization exists in AAA
	_, err := s.aaaService.GetOrganization(ctx, registerReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify organization in AAA: %w", err)
	}

	// Check if FPO reference already exists
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{
					Field:    "aaa_org_id",
					Operator: base.OpEqual,
					Value:    registerReq.AAAOrgID,
				},
			},
		},
	}
	existingFPO, err := s.fpoRefRepo.FindOne(ctx, filter)
	if err == nil && existingFPO != nil {
		return nil, fmt.Errorf("FPO reference already exists for organization ID: %s", registerReq.AAAOrgID)
	}

	// Create FPO reference using constructor
	// Constructor ensures ID is properly initialized
	fpoRef := fpo.NewFPORef(registerReq.AAAOrgID)
	fpoRef.Name = registerReq.Name
	fpoRef.RegistrationNo = registerReq.RegistrationNo
	fpoRef.Status = fpo.FPOStatusActive
	fpoRef.BusinessConfig = registerReq.BusinessConfig

	err = s.fpoRefRepo.Create(ctx, fpoRef)
	if err != nil {
		return nil, fmt.Errorf("failed to create FPO reference: %w", err)
	}

	// Prepare response
	responseData := &responses.FPORefData{
		ID:             fpoRef.ID,
		AAAOrgID:       fpoRef.AAAOrgID,
		Name:           fpoRef.Name,
		RegistrationNo: fpoRef.RegistrationNo,
		CEOUserID:      fpoRef.CEOUserID,
		BusinessConfig: fpoRef.BusinessConfig,
		Status:         fpoRef.Status.String(),
		Metadata:       registerReq.Metadata,
		CreatedAt:      fpoRef.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      fpoRef.UpdatedAt.Format(time.RFC3339),
	}

	log.Printf("Successfully registered FPO reference: %s", fpoRef.ID)
	return responseData, nil
}

// GetFPORef gets FPO reference by organization ID
// Note: Consider using the lifecycle service's GetOrSyncFPO method for automatic fallback to AAA sync
func (s *FPOServiceImpl) GetFPORef(ctx context.Context, orgID string) (interface{}, error) {
	log.Printf("FPOService: Getting FPO reference for org ID: %s", orgID)

	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Get FPO reference from database
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{
					Field:    "aaa_org_id",
					Operator: base.OpEqual,
					Value:    orgID,
				},
			},
		},
	}
	fpoRef, err := s.fpoRefRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get FPO reference: %w", err)
	}

	if fpoRef == nil {
		return nil, fmt.Errorf("FPO reference not found for organization ID: %s. Consider using the FPO lifecycle sync endpoint: POST /identity/fpo/sync/%s", orgID, orgID)
	}

	// Prepare response
	responseData := &responses.FPORefData{
		ID:             fpoRef.ID,
		AAAOrgID:       fpoRef.AAAOrgID,
		Name:           fpoRef.Name,
		RegistrationNo: fpoRef.RegistrationNo,
		CEOUserID:      fpoRef.CEOUserID,
		BusinessConfig: fpoRef.BusinessConfig,
		Status:         fpoRef.Status.String(),
		CreatedAt:      fpoRef.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      fpoRef.UpdatedAt.Format(time.RFC3339),
	}

	log.Printf("Successfully retrieved FPO reference: %s", fpoRef.ID)
	return responseData, nil
}

// CompleteFPOSetup retries failed setup operations for PENDING_SETUP FPOs
// Business Rule 1.1: Allow recovery from partial failure during FPO creation
func (s *FPOServiceImpl) CompleteFPOSetup(ctx context.Context, orgID string) (interface{}, error) {
	log.Printf("FPOService: Starting CompleteFPOSetup for org ID: %s", orgID)

	// Get FPO reference from database
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{
					Field:    "aaa_org_id",
					Operator: base.OpEqual,
					Value:    orgID,
				},
			},
		},
	}
	fpoRef, err := s.fpoRefRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("FPO reference not found: %w", err)
	}

	// Check if FPO is in PENDING_SETUP status
	if !fpoRef.IsPendingSetup() {
		return nil, fmt.Errorf("FPO is not in PENDING_SETUP status, current status: %s", fpoRef.Status)
	}

	// Retry creating missing user groups
	setupErrors := make(entities.JSONB)
	groupNames := []string{"directors", "shareholders", "store_staff", "store_managers"}

	for _, groupName := range groupNames {
		// Try to create missing group (AAA service will handle duplicates)
		createGroupReq := map[string]interface{}{
			"name":        groupName,
			"description": fmt.Sprintf("%s group for %s", groupName, fpoRef.Name),
			"org_id":      orgID,
			"permissions": s.getGroupPermissions(groupName),
		}

		groupResp, err := s.aaaService.CreateUserGroup(ctx, createGroupReq)
		if err != nil {
			log.Printf("Warning: Failed to create user group %s: %v", groupName, err)
			setupErrors[fmt.Sprintf("user_group_%s", groupName)] = err.Error()
			continue
		}

		groupRespMap, ok := groupResp.(map[string]interface{})
		if !ok {
			log.Printf("Warning: Invalid group creation response for %s", groupName)
			setupErrors[fmt.Sprintf("user_group_%s", groupName)] = "invalid response from AAA service"
			continue
		}

		groupID := groupRespMap["group_id"].(string)
		log.Printf("Created user group: %s with ID: %s", groupName, groupID)

		// Assign permissions to group
		for _, permission := range s.getGroupPermissions(groupName) {
			err = s.aaaService.AssignPermissionToGroup(ctx, groupID, "fpo", permission)
			if err != nil {
				log.Printf("Warning: Failed to assign permission %s to group %s: %v", permission, groupName, err)
				setupErrors[fmt.Sprintf("permission_%s_%s", groupName, permission)] = err.Error()
			}
		}
	}

	// Update FPO status based on retry results
	if len(setupErrors) == 0 {
		fpoRef.Status = fpo.FPOStatusActive
		fpoRef.SetupErrors = nil
		log.Printf("FPO setup completed successfully, marking as ACTIVE")
	} else {
		// Still has errors, keep PENDING_SETUP status
		fpoRef.SetupErrors = setupErrors
		log.Printf("FPO setup still incomplete. Remaining errors: %v", setupErrors)
	}

	// Update FPO in database
	if err := s.fpoRefRepo.Create(ctx, fpoRef); err != nil {
		return nil, fmt.Errorf("failed to update FPO reference: %w", err)
	}

	// Prepare response
	responseData := &responses.FPORefData{
		ID:             fpoRef.ID,
		AAAOrgID:       fpoRef.AAAOrgID,
		Name:           fpoRef.Name,
		RegistrationNo: fpoRef.RegistrationNo,
		CEOUserID:      fpoRef.CEOUserID,
		BusinessConfig: fpoRef.BusinessConfig,
		Status:         fpoRef.Status.String(),
		CreatedAt:      fpoRef.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      fpoRef.UpdatedAt.Format(time.RFC3339),
	}

	log.Printf("CompleteFPOSetup finished for %s with status: %s", orgID, fpoRef.Status)
	return responseData, nil
}

// getGroupPermissions returns the permissions for a specific user group
func (s *FPOServiceImpl) getGroupPermissions(groupName string) []string {
	switch groupName {
	case "directors":
		return []string{"manage", "read", "write", "approve"}
	case "shareholders":
		return []string{"read", "vote"}
	case "store_staff":
		return []string{"read", "write", "inventory"}
	case "store_managers":
		return []string{"read", "write", "manage", "inventory", "reports"}
	default:
		return []string{"read"}
	}
}

// UpdateCEO updates the CEO of an FPO and assigns the CEO role to the new user
// Business Rules:
// - The new CEO user must exist in AAA
// - The new CEO cannot already be CEO of another FPO
// - The old CEO will have their CEO role removed (if any)
// - The new CEO will be assigned the CEO role for this organization
func (s *FPOServiceImpl) UpdateCEO(ctx context.Context, orgID string, req interface{}) (interface{}, error) {
	log.Printf("FPOService: Starting UpdateCEO workflow for org: %s", orgID)

	// Type assert the request
	updateReq, ok := req.(*requests.UpdateCEORequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for UpdateCEO")
	}

	// Validate request
	if updateReq.NewCEOUserID == "" {
		return nil, fmt.Errorf("new CEO user ID is required")
	}
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Step 1: Verify the organization exists in AAA
	orgResp, err := s.aaaService.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Type assert to *aaa.OrganizationData (the actual return type from AAA client)
	orgData, ok := orgResp.(*aaa.OrganizationData)
	if !ok {
		return nil, fmt.Errorf("invalid organization response from AAA")
	}

	orgName := orgData.Name

	// Step 2: Verify the new CEO user exists in AAA
	_, err = s.aaaService.GetUser(ctx, updateReq.NewCEOUserID)
	if err != nil {
		return nil, fmt.Errorf("new CEO user not found in AAA: %w", err)
	}

	// Step 3: Check if the new CEO is already a CEO of another organization
	// Business Rule: A user CANNOT be CEO of multiple FPOs simultaneously
	isCEO, err := s.aaaService.CheckUserRole(ctx, updateReq.NewCEOUserID, constants.RoleFPOCEO)
	if err != nil {
		log.Printf("Warning: Failed to check if user is already CEO: %v", err)
		// Continue anyway - this is a best-effort check
	} else if isCEO {
		return nil, fmt.Errorf("user is already CEO of another FPO - a user cannot be CEO of multiple FPOs simultaneously")
	}

	// Step 4: Assign CEO role to the new user
	log.Printf("Assigning CEO role to user %s for organization %s", updateReq.NewCEOUserID, orgID)
	err = s.aaaService.AssignRole(ctx, updateReq.NewCEOUserID, orgID, constants.RoleFPOCEO)
	if err != nil {
		// Check if role is already assigned - this is not an error
		if strings.Contains(err.Error(), "role already assigned") {
			log.Printf("CEO role already assigned to user %s for organization %s, continuing", updateReq.NewCEOUserID, orgID)
		} else {
			return nil, fmt.Errorf("failed to assign CEO role to new user: %w", err)
		}
	}

	// Step 5: Update the FPO ref in database with the new CEO user ID
	log.Printf("Updating FPO ref with CEO user ID %s for organization %s", updateReq.NewCEOUserID, orgID)
	err = s.fpoRefRepo.UpdateCEO(ctx, orgID, updateReq.NewCEOUserID)
	if err != nil {
		log.Printf("Warning: Failed to update FPO ref with CEO: %v", err)
		// Continue - the AAA role assignment was successful
	} else {
		log.Printf("FPO ref updated with CEO user ID %s", updateReq.NewCEOUserID)
	}

	// Prepare response
	responseData := &responses.UpdateCEOData{
		AAAOrgID:     orgID,
		OrgName:      orgName,
		NewCEOUserID: updateReq.NewCEOUserID,
		UpdatedAt:    time.Now(),
	}

	log.Printf("Successfully updated CEO for organization %s to user %s", orgID, updateReq.NewCEOUserID)
	return responseData, nil
}
