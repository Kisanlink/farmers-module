package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmerLinkRepository defines the interface for farmer link repository operations
type FarmerLinkRepository interface {
	Create(ctx context.Context, entity *entities.FarmerLink) error
	Update(ctx context.Context, entity *entities.FarmerLink) error
	Find(ctx context.Context, filter *base.Filter) ([]*entities.FarmerLink, error)
}

// FarmerLinkageServiceImpl implements FarmerLinkageService
type FarmerLinkageServiceImpl struct {
	farmerLinkageRepo FarmerLinkRepository
	aaaService        AAAService
}

// NewFarmerLinkageService creates a new farmer linkage service
func NewFarmerLinkageService(farmerLinkageRepo FarmerLinkRepository, aaaService AAAService) FarmerLinkageService {
	return &FarmerLinkageServiceImpl{
		farmerLinkageRepo: farmerLinkageRepo,
		aaaService:        aaaService,
	}
}

// LinkFarmerToFPO implements W1: Link farmer to FPO with AAA validation
func (s *FarmerLinkageServiceImpl) LinkFarmerToFPO(ctx context.Context, req interface{}) error {
	linkReq, ok := req.(*requests.LinkFarmerRequest)
	if !ok {
		return fmt.Errorf("invalid request type for LinkFarmerToFPO")
	}

	// Validate input
	if linkReq.AAAUserID == "" || linkReq.AAAOrgID == "" {
		return fmt.Errorf("aaa_user_id and aaa_org_id are required")
	}

	// Check 'farmer.link' permission on the target organization
	hasPermission, err := s.aaaService.CheckPermission(ctx, linkReq.AAAUserID, "farmer", "link", linkReq.AAAUserID, linkReq.AAAOrgID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to link farmer to FPO")
	}

	// Verify both farmer and FPO exist in AAA service
	_, err = s.aaaService.GetUser(ctx, linkReq.AAAUserID)
	if err != nil {
		return fmt.Errorf("farmer not found in AAA service: %w", err)
	}

	_, err = s.aaaService.GetOrganization(ctx, linkReq.AAAOrgID)
	if err != nil {
		return fmt.Errorf("FPO not found in AAA service: %w", err)
	}

	// Check if linkage already exists
	existingLink, err := s.getFarmerLinkByUserAndOrg(ctx, linkReq.AAAUserID, linkReq.AAAOrgID)
	if err == nil && existingLink != nil {
		// Update existing link to ACTIVE if it was inactive
		if existingLink.Status != "ACTIVE" {
			existingLink.Status = "ACTIVE"
			now := time.Now()
			existingLink.LinkedAt = &now
			existingLink.UnlinkedAt = nil
			return s.farmerLinkageRepo.Update(ctx, existingLink)
		}
		return nil // Already linked and active
	}

	// Create new farmer link
	now := time.Now()
	farmerLink := &entities.FarmerLink{
		AAAUserID: linkReq.AAAUserID,
		AAAOrgID:  linkReq.AAAOrgID,
		Status:    "ACTIVE",
		LinkedAt:  &now,
	}

	return s.farmerLinkageRepo.Create(ctx, farmerLink)
}

// UnlinkFarmerFromFPO implements W2: Unlink farmer from FPO with soft delete
func (s *FarmerLinkageServiceImpl) UnlinkFarmerFromFPO(ctx context.Context, req interface{}) error {
	unlinkReq, ok := req.(*requests.UnlinkFarmerRequest)
	if !ok {
		return fmt.Errorf("invalid request type for UnlinkFarmerFromFPO")
	}

	// Validate input
	if unlinkReq.AAAUserID == "" || unlinkReq.AAAOrgID == "" {
		return fmt.Errorf("aaa_user_id and aaa_org_id are required")
	}

	// Check 'farmer.unlink' permission on the target organization
	hasPermission, err := s.aaaService.CheckPermission(ctx, unlinkReq.AAAUserID, "farmer", "unlink", unlinkReq.AAAUserID, unlinkReq.AAAOrgID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to unlink farmer from FPO")
	}

	// Find existing link
	existingLink, err := s.getFarmerLinkByUserAndOrg(ctx, unlinkReq.AAAUserID, unlinkReq.AAAOrgID)
	if err != nil {
		return fmt.Errorf("farmer link not found: %w", err)
	}

	// Check if already inactive
	if existingLink.Status == "INACTIVE" {
		return fmt.Errorf("farmer is already unlinked from FPO")
	}

	// Soft delete by setting status to INACTIVE
	existingLink.Status = "INACTIVE"
	now := time.Now()
	existingLink.UnlinkedAt = &now
	// Clear KisanSathi assignment when unlinking
	existingLink.KisanSathiUserID = nil

	return s.farmerLinkageRepo.Update(ctx, existingLink)
}

// GetFarmerLinkage gets farmer linkage status
func (s *FarmerLinkageServiceImpl) GetFarmerLinkage(ctx context.Context, farmerID, orgID string) (interface{}, error) {
	// Check 'farmer.read' permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, farmerID, "farmer", "read", farmerID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to read farmer linkage")
	}

	farmerLink, err := s.getFarmerLinkByUserAndOrg(ctx, farmerID, orgID)
	if err != nil {
		return nil, fmt.Errorf("farmer linkage not found: %w", err)
	}

	// Convert to response format
	linkageData := &responses.FarmerLinkageData{
		ID:               farmerLink.ID,
		AAAUserID:        farmerLink.AAAUserID,
		AAAOrgID:         farmerLink.AAAOrgID,
		KisanSathiUserID: farmerLink.KisanSathiUserID,
		Status:           farmerLink.Status,
		CreatedAt:        farmerLink.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        farmerLink.UpdatedAt.Format(time.RFC3339),
	}

	if farmerLink.LinkedAt != nil {
		linkageData.LinkedAt = farmerLink.LinkedAt.Format(time.RFC3339)
	}
	if farmerLink.UnlinkedAt != nil {
		linkageData.UnlinkedAt = farmerLink.UnlinkedAt.Format(time.RFC3339)
	}

	return linkageData, nil
}

// AssignKisanSathi implements W4: Assign KisanSathi to farmer with role validation
func (s *FarmerLinkageServiceImpl) AssignKisanSathi(ctx context.Context, req interface{}) (interface{}, error) {
	assignReq, ok := req.(*requests.AssignKisanSathiRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for AssignKisanSathi")
	}

	// Validate input
	if assignReq.AAAUserID == "" || assignReq.AAAOrgID == "" || assignReq.KisanSathiUserID == "" {
		return nil, fmt.Errorf("aaa_user_id, aaa_org_id, and kisan_sathi_user_id are required")
	}

	// Check 'kisansathi.assign' permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, assignReq.KisanSathiUserID, "kisansathi", "assign", assignReq.AAAUserID, assignReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to assign KisanSathi")
	}

	// Verify KisanSathi user exists in AAA service
	_, err = s.aaaService.GetUser(ctx, assignReq.KisanSathiUserID)
	if err != nil {
		return nil, fmt.Errorf("KisanSathi user not found in AAA service: %w", err)
	}

	// Ensure KisanSathi role exists and user has it
	err = s.ensureKisanSathiRole(ctx, assignReq.KisanSathiUserID, assignReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure KisanSathi role: %w", err)
	}

	// Find existing farmer link
	farmerLink, err := s.getFarmerLinkByUserAndOrg(ctx, assignReq.AAAUserID, assignReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("farmer link not found: %w", err)
	}

	// Check if farmer link is active
	if farmerLink.Status != "ACTIVE" {
		return nil, fmt.Errorf("cannot assign KisanSathi to inactive farmer link")
	}

	// Update KisanSathi assignment
	farmerLink.KisanSathiUserID = &assignReq.KisanSathiUserID

	err = s.farmerLinkageRepo.Update(ctx, farmerLink)
	if err != nil {
		return nil, fmt.Errorf("failed to assign KisanSathi: %w", err)
	}

	// Convert to response format
	assignmentData := &responses.KisanSathiAssignmentData{
		ID:               farmerLink.ID,
		AAAUserID:        farmerLink.AAAUserID,
		AAAOrgID:         farmerLink.AAAOrgID,
		KisanSathiUserID: farmerLink.KisanSathiUserID,
		Status:           farmerLink.Status,
		AssignedAt:       time.Now().Format(time.RFC3339),
		CreatedAt:        farmerLink.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        farmerLink.UpdatedAt.Format(time.RFC3339),
	}

	return assignmentData, nil
}

// ReassignOrRemoveKisanSathi implements W5: Reassign or remove KisanSathi
func (s *FarmerLinkageServiceImpl) ReassignOrRemoveKisanSathi(ctx context.Context, req interface{}) (interface{}, error) {
	reassignReq, ok := req.(*requests.ReassignKisanSathiRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for ReassignOrRemoveKisanSathi")
	}

	// Validate input
	if reassignReq.AAAUserID == "" || reassignReq.AAAOrgID == "" {
		return nil, fmt.Errorf("aaa_user_id and aaa_org_id are required")
	}

	// Check 'kisansathi.reassign' permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, reassignReq.AAAUserID, "kisansathi", "reassign", reassignReq.AAAUserID, reassignReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to reassign KisanSathi")
	}

	// If new KisanSathi is provided, verify they exist and ensure they have the role
	if reassignReq.NewKisanSathiUserID != nil {
		// Verify user exists in AAA service
		_, err = s.aaaService.GetUser(ctx, *reassignReq.NewKisanSathiUserID)
		if err != nil {
			return nil, fmt.Errorf("new KisanSathi user not found in AAA service: %w", err)
		}

		// Ensure KisanSathi role exists and user has it
		err = s.ensureKisanSathiRole(ctx, *reassignReq.NewKisanSathiUserID, reassignReq.AAAOrgID)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure KisanSathi role for new user: %w", err)
		}
	}

	// Find existing farmer link
	farmerLink, err := s.getFarmerLinkByUserAndOrg(ctx, reassignReq.AAAUserID, reassignReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("farmer link not found: %w", err)
	}

	// Check if farmer link is active
	if farmerLink.Status != "ACTIVE" {
		return nil, fmt.Errorf("cannot reassign KisanSathi for inactive farmer link")
	}

	// Update KisanSathi assignment (reassign or remove)
	farmerLink.KisanSathiUserID = reassignReq.NewKisanSathiUserID

	err = s.farmerLinkageRepo.Update(ctx, farmerLink)
	if err != nil {
		return nil, fmt.Errorf("failed to reassign KisanSathi: %w", err)
	}

	// Convert to response format
	assignmentData := &responses.KisanSathiAssignmentData{
		ID:               farmerLink.ID,
		AAAUserID:        farmerLink.AAAUserID,
		AAAOrgID:         farmerLink.AAAOrgID,
		KisanSathiUserID: farmerLink.KisanSathiUserID,
		Status:           farmerLink.Status,
		CreatedAt:        farmerLink.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        farmerLink.UpdatedAt.Format(time.RFC3339),
	}

	if reassignReq.NewKisanSathiUserID != nil {
		assignmentData.AssignedAt = time.Now().Format(time.RFC3339)
	} else {
		assignmentData.UnassignedAt = time.Now().Format(time.RFC3339)
	}

	return assignmentData, nil
}

// CreateKisanSathiUser creates a new user and assigns KisanSathi role
func (s *FarmerLinkageServiceImpl) CreateKisanSathiUser(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateKisanSathiUserRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for CreateKisanSathiUser")
	}

	// Validate input
	if createReq.Username == "" || createReq.PhoneNumber == "" || createReq.Password == "" || createReq.FullName == "" {
		return nil, fmt.Errorf("username, phone_number, password, and full_name are required")
	}

	// Check 'user.create' permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, createReq.Username, "user", "create", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to create user")
	}

	// Check if user already exists by phone or email
	if createReq.PhoneNumber != "" {
		existingUser, err := s.aaaService.GetUserByMobile(ctx, createReq.PhoneNumber)
		if err == nil && existingUser != nil {
			// If user exists, ensure they have KisanSathi role
			userMap, ok := existingUser.(map[string]interface{})
			if ok {
				userID, ok := userMap["id"].(string)
				if ok {
					err = s.ensureKisanSathiRole(ctx, userID, "")
					if err != nil {
						return nil, fmt.Errorf("failed to assign KisanSathi role to existing user: %w", err)
					}

					userData := &responses.KisanSathiUserData{
						ID:          userID,
						Username:    userMap["username"].(string),
						PhoneNumber: createReq.PhoneNumber,
						Email:       createReq.Email,
						FullName:    userMap["full_name"].(string),
						Role:        "KisanSathi",
						Status:      userMap["status"].(string),
						Metadata:    createReq.Metadata,
						CreatedAt:   userMap["created_at"].(string),
					}
					return userData, nil
				}
			}
			return nil, fmt.Errorf("user with phone number already exists")
		}
	}

	if createReq.Email != "" {
		existingUser, err := s.aaaService.GetUserByEmail(ctx, createReq.Email)
		if err == nil && existingUser != nil {
			return nil, fmt.Errorf("user with email already exists")
		}
	}

	// Create user with KisanSathi role
	userCreateReq := map[string]interface{}{
		"username":     createReq.Username,
		"phone_number": createReq.PhoneNumber,
		"email":        createReq.Email,
		"password":     createReq.Password,
		"full_name":    createReq.FullName,
		"country_code": createReq.CountryCode,
		"role":         "KisanSathi",
	}

	userResponse, err := s.aaaService.CreateUser(ctx, userCreateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create KisanSathi user: %w", err)
	}

	// Extract user data from response
	userMap, ok := userResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user creation response format")
	}

	userID, ok := userMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in response")
	}

	// Ensure KisanSathi role is properly assigned
	err = s.ensureKisanSathiRole(ctx, userID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to ensure KisanSathi role: %w", err)
	}

	userData := &responses.KisanSathiUserData{
		ID:          userID,
		Username:    createReq.Username,
		PhoneNumber: createReq.PhoneNumber,
		Email:       createReq.Email,
		FullName:    createReq.FullName,
		Role:        "KisanSathi",
		Status:      userMap["status"].(string),
		Metadata:    createReq.Metadata,
		CreatedAt:   userMap["created_at"].(string),
	}

	return userData, nil
}

// getFarmerLinkByUserAndOrg is a helper method to find farmer link by user and org
func (s *FarmerLinkageServiceImpl) getFarmerLinkByUserAndOrg(ctx context.Context, userID, orgID string) (*entities.FarmerLink, error) {
	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, userID).
		Where("aaa_org_id", base.OpEqual, orgID).
		Build()

	results, err := s.farmerLinkageRepo.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("farmer link not found")
	}

	return results[0], nil
}

// ensureKisanSathiRole ensures that the KisanSathi role exists in AAA service and assigns it to the user
func (s *FarmerLinkageServiceImpl) ensureKisanSathiRole(ctx context.Context, userID, orgID string) error {
	// First, check if user already has the KisanSathi role
	hasRole, err := s.aaaService.CheckUserRole(ctx, userID, "KisanSathi")
	if err != nil {
		// If role check fails, try to assign the role anyway
		// This handles the case where the role might not exist yet
	} else if hasRole {
		// User already has the role, nothing to do
		return nil
	}

	// Assign KisanSathi role to the user
	err = s.aaaService.AssignRole(ctx, userID, orgID, "KisanSathi")
	if err != nil {
		return fmt.Errorf("failed to assign KisanSathi role: %w", err)
	}

	// Verify the role was assigned successfully
	hasRole, err = s.aaaService.CheckUserRole(ctx, userID, "KisanSathi")
	if err != nil {
		return fmt.Errorf("failed to verify KisanSathi role assignment: %w", err)
	}
	if !hasRole {
		return fmt.Errorf("KisanSathi role assignment verification failed")
	}

	return nil
}
