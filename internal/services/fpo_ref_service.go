package services

import (
	"context"
	"fmt"
	"log"
	"time"

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
}

// FPOServiceImpl implements FPOService
type FPOServiceImpl struct {
	fpoRefRepo FPORefRepository
	aaaService AAAService
}

// NewFPOService creates a new FPO service
func NewFPOService(fpoRefRepo FPORefRepository, aaaService AAAService) FPOService {
	return &FPOServiceImpl{
		fpoRefRepo: fpoRefRepo,
		aaaService: aaaService,
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
	if createReq.CEOUser.FirstName == "" || createReq.CEOUser.LastName == "" {
		return nil, fmt.Errorf("CEO user details are required")
	}
	if createReq.CEOUser.PhoneNumber == "" {
		return nil, fmt.Errorf("CEO phone number is required")
	}

	log.Printf("Creating FPO: %s with CEO: %s %s", createReq.Name, createReq.CEOUser.FirstName, createReq.CEOUser.LastName)

	// Step 1: Check if CEO user exists in AAA, create if not
	var ceoUserID string
	existingUser, err := s.aaaService.GetUserByMobile(ctx, createReq.CEOUser.PhoneNumber)
	if err != nil {
		log.Printf("CEO user not found, creating new user: %v", err)

		// Create CEO user in AAA
		createUserReq := map[string]interface{}{
			"username":     fmt.Sprintf("%s_%s", createReq.CEOUser.FirstName, createReq.CEOUser.LastName),
			"phone_number": createReq.CEOUser.PhoneNumber,
			"email":        createReq.CEOUser.Email,
			"password":     createReq.CEOUser.Password,
			"full_name":    fmt.Sprintf("%s %s", createReq.CEOUser.FirstName, createReq.CEOUser.LastName),
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
		log.Printf("Using existing CEO user with ID: %s", ceoUserID)
	}

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
		"type":        "FPO",
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

	// Step 4: Assign CEO role to user in organization
	err = s.aaaService.AssignRole(ctx, ceoUserID, aaaOrgID, "CEO")
	if err != nil {
		log.Printf("Warning: Failed to assign CEO role: %v", err)
		// Continue as this might not be critical
	}

	// Step 5: Create user groups for FPO
	userGroups := []responses.UserGroupData{}
	setupErrors := make(entities.JSONB)
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
	fpoRef := &fpo.FPORef{
		AAAOrgID:       aaaOrgID,
		Name:           createReq.Name,
		RegistrationNo: createReq.RegistrationNo,
		Status:         fpoStatus,
		BusinessConfig: createReq.BusinessConfig,
		SetupErrors:    setupErrors,
	}

	err = s.fpoRefRepo.Create(ctx, fpoRef)
	if err != nil {
		log.Printf("Warning: Failed to store FPO reference locally: %v", err)
		// Continue as AAA organization is already created
	}

	// Step 8: Prepare response
	responseData := &responses.CreateFPOData{
		FPOID:      fpoRef.ID,
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

	// Create FPO reference
	fpoRef := &fpo.FPORef{
		AAAOrgID:       registerReq.AAAOrgID,
		Name:           registerReq.Name,
		RegistrationNo: registerReq.RegistrationNo,
		Status:         "ACTIVE",
		BusinessConfig: registerReq.BusinessConfig,
	}

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
		return nil, fmt.Errorf("FPO reference not found for organization ID: %s", orgID)
	}

	// Prepare response
	responseData := &responses.FPORefData{
		ID:             fpoRef.ID,
		AAAOrgID:       fpoRef.AAAOrgID,
		Name:           fpoRef.Name,
		RegistrationNo: fpoRef.RegistrationNo,
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
