package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmerLinkRepository defines the interface for farmer link repository operations
type FarmerLinkRepository interface {
	Create(ctx context.Context, entity *farmerentity.FarmerLink) error
	Update(ctx context.Context, entity *farmerentity.FarmerLink) error
	Find(ctx context.Context, filter *base.Filter) ([]*farmerentity.FarmerLink, error)
}

// FarmerRepository defines the interface for farmer repository operations
type FarmerRepository interface {
	Create(ctx context.Context, entity *farmerentity.Farmer) error
	FindOne(ctx context.Context, filter *base.Filter) (*farmerentity.Farmer, error)
}

// FarmerLinkageServiceImpl implements FarmerLinkageService
type FarmerLinkageServiceImpl struct {
	farmerLinkageRepo FarmerLinkRepository
	farmerRepo        FarmerRepository
	aaaService        AAAService
}

// NewFarmerLinkageService creates a new farmer linkage service
func NewFarmerLinkageService(farmerLinkageRepo FarmerLinkRepository, farmerRepo FarmerRepository, aaaService AAAService) FarmerLinkageService {
	return &FarmerLinkageServiceImpl{
		farmerLinkageRepo: farmerLinkageRepo,
		farmerRepo:        farmerRepo,
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can link farmer
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farmer", "link", linkReq.AAAUserID, linkReq.AAAOrgID)
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

	// Verify farmer exists in local database by aaa_user_id only
	// A farmer is identified by their user ID and can be linked to multiple FPOs
	farmerFilter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, linkReq.AAAUserID).
		Build()

	existingFarmer, err := s.farmerRepo.FindOne(ctx, farmerFilter)
	if err != nil || existingFarmer == nil {
		return fmt.Errorf("farmer with aaa_user_id=%s must be created before linking to FPO", linkReq.AAAUserID)
	}

	// Check if linkage already exists
	existingLink, err := s.getFarmerLinkByUserAndOrg(ctx, linkReq.AAAUserID, linkReq.AAAOrgID)
	if err == nil && existingLink != nil {
		// Update existing link to ACTIVE if it was inactive
		if existingLink.Status != "ACTIVE" {
			existingLink.Status = "ACTIVE"
			return s.farmerLinkageRepo.Update(ctx, existingLink)
		}
		return nil // Already linked and active
	}

	// Create new farmer link with proper ID generation
	farmerLink := farmerentity.NewFarmerLink()
	farmerLink.AAAUserID = linkReq.AAAUserID
	farmerLink.AAAOrgID = linkReq.AAAOrgID
	farmerLink.Status = "ACTIVE"

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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can unlink farmer
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farmer", "unlink", unlinkReq.AAAUserID, unlinkReq.AAAOrgID)
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
	// Clear KisanSathi assignment when unlinking
	existingLink.KisanSathiUserID = nil

	return s.farmerLinkageRepo.Update(ctx, existingLink)
}

// GetFarmerLinkage gets farmer linkage status
func (s *FarmerLinkageServiceImpl) GetFarmerLinkage(ctx context.Context, farmerID, orgID string) (interface{}, error) {
	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can read farmer linkage
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "farmer", "read", farmerID, orgID)
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can assign KisanSathi
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "kisansathi", "assign", assignReq.AAAUserID, assignReq.AAAOrgID)
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can reassign KisanSathi
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "kisansathi", "reassign", reassignReq.AAAUserID, reassignReq.AAAOrgID)
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

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can create user
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "user", "create", "", "")
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
func (s *FarmerLinkageServiceImpl) getFarmerLinkByUserAndOrg(ctx context.Context, userID, orgID string) (*farmerentity.FarmerLink, error) {
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

// ListKisanSathis lists all KisanSathis (users assigned to at least one farmer)
func (s *FarmerLinkageServiceImpl) ListKisanSathis(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListKisanSathisRequest)
	if !ok {
		listReq = &requests.ListKisanSathisRequest{
			Page:     1,
			PageSize: 50,
		}
	}

	// Set default pagination
	if listReq.Page < 1 {
		listReq.Page = 1
	}
	if listReq.PageSize < 1 {
		listReq.PageSize = 50
	}
	if listReq.PageSize > 100 {
		listReq.PageSize = 100
	}

	// Query all farmer links (we'll filter for non-null kisan_sathi_user_id in-memory)
	filter := base.NewFilterBuilder().Build()

	farmerLinks, err := s.farmerLinkageRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query farmer links: %w", err)
	}

	// Extract unique KisanSathi user IDs
	kisanSathiMap := make(map[string]bool)
	for _, link := range farmerLinks {
		if link.KisanSathiUserID != nil && *link.KisanSathiUserID != "" {
			kisanSathiMap[*link.KisanSathiUserID] = true
		}
	}

	// Convert map keys to slice
	uniqueKisanSathis := make([]string, 0, len(kisanSathiMap))
	for userID := range kisanSathiMap {
		uniqueKisanSathis = append(uniqueKisanSathis, userID)
	}

	// Apply pagination
	totalCount := int64(len(uniqueKisanSathis))
	start := (listReq.Page - 1) * listReq.PageSize
	end := start + listReq.PageSize

	if start >= len(uniqueKisanSathis) {
		// Return empty result if page is beyond available data
		return &responses.KisanSathiListResponse{
			Success:   true,
			Message:   "KisanSathis retrieved successfully",
			Data:      []*responses.KisanSathiData{},
			Page:      listReq.Page,
			PageSize:  listReq.PageSize,
			Total:     totalCount,
			RequestID: listReq.RequestID,
		}, nil
	}

	if end > len(uniqueKisanSathis) {
		end = len(uniqueKisanSathis)
	}

	paginatedKisanSathis := uniqueKisanSathis[start:end]

	// Fetch user details from AAA for each KisanSathi
	kisanSathiList := make([]*responses.KisanSathiData, 0, len(paginatedKisanSathis))
	for _, userID := range paginatedKisanSathis {
		// Get user details from AAA
		userData, err := s.aaaService.GetUser(ctx, userID)
		if err != nil {
			// If we can't get user details, skip this user but log the error
			fmt.Printf("Warning: failed to get user details for KisanSathi %s: %v\n", userID, err)
			continue
		}

		// Convert user data to map
		userMap, ok := userData.(map[string]interface{})
		if !ok {
			fmt.Printf("Warning: invalid user data format for KisanSathi %s\n", userID)
			continue
		}

		// Extract user fields
		username, _ := userMap["username"].(string)
		phoneNumber, _ := userMap["phone"].(string)
		email, _ := userMap["email"].(string)
		fullName, _ := userMap["full_name"].(string)
		status, _ := userMap["status"].(string)

		kisanSathiData := &responses.KisanSathiData{
			ID:          userID,
			Username:    username,
			PhoneNumber: phoneNumber,
			Email:       email,
			FullName:    fullName,
			Status:      status,
		}

		kisanSathiList = append(kisanSathiList, kisanSathiData)
	}

	response := &responses.KisanSathiListResponse{
		Success:   true,
		Message:   "KisanSathis retrieved successfully",
		Data:      kisanSathiList,
		Page:      listReq.Page,
		PageSize:  listReq.PageSize,
		Total:     totalCount,
		RequestID: listReq.RequestID,
	}

	return response, nil
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

// BulkLinkFarmersToFPO links multiple farmers to an FPO
func (s *FarmerLinkageServiceImpl) BulkLinkFarmersToFPO(ctx context.Context, req interface{}) (interface{}, error) {
	bulkReq, ok := req.(*requests.BulkLinkFarmersRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for BulkLinkFarmersToFPO")
	}

	// Validate input
	if bulkReq.AAAOrgID == "" {
		return nil, fmt.Errorf("aaa_org_id is required")
	}
	if len(bulkReq.AAAUserIDs) == 0 {
		return nil, fmt.Errorf("aaa_user_ids is required and must not be empty")
	}
	if len(bulkReq.AAAUserIDs) > 1000 {
		return nil, fmt.Errorf("aaa_user_ids cannot exceed 1000 entries")
	}

	// Verify FPO exists in AAA service
	_, err := s.aaaService.GetOrganization(ctx, bulkReq.AAAOrgID)
	if err != nil {
		return nil, fmt.Errorf("FPO not found in AAA service: %w", err)
	}

	results := make([]responses.BulkLinkResult, 0, len(bulkReq.AAAUserIDs))
	successCount := 0
	failureCount := 0
	skippedCount := 0

	for _, userID := range bulkReq.AAAUserIDs {
		result := responses.BulkLinkResult{
			AAAUserID: userID,
		}

		// Verify farmer exists in local database
		farmerFilter := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, userID).
			Build()

		existingFarmer, err := s.farmerRepo.FindOne(ctx, farmerFilter)
		if err != nil || existingFarmer == nil {
			result.Success = false
			result.Error = fmt.Sprintf("farmer with aaa_user_id=%s not found in database", userID)
			result.Status = "FAILED"
			failureCount++
			results = append(results, result)
			if !bulkReq.ContinueOnError {
				break
			}
			continue
		}

		// Check if linkage already exists
		existingLink, err := s.getFarmerLinkByUserAndOrg(ctx, userID, bulkReq.AAAOrgID)
		if err == nil && existingLink != nil {
			// Update existing link to ACTIVE if it was inactive
			if existingLink.Status != "ACTIVE" {
				existingLink.Status = "ACTIVE"
				if err := s.farmerLinkageRepo.Update(ctx, existingLink); err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("failed to reactivate link: %v", err)
					result.Status = "FAILED"
					failureCount++
				} else {
					result.Success = true
					result.Status = "LINKED"
					successCount++
				}
			} else {
				result.Success = true
				result.Status = "ALREADY_LINKED"
				skippedCount++
			}
			results = append(results, result)
			continue
		}

		// Create new farmer link
		farmerLink := farmerentity.NewFarmerLink()
		farmerLink.AAAUserID = userID
		farmerLink.AAAOrgID = bulkReq.AAAOrgID
		farmerLink.Status = "ACTIVE"

		if err := s.farmerLinkageRepo.Create(ctx, farmerLink); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to create link: %v", err)
			result.Status = "FAILED"
			failureCount++
			if !bulkReq.ContinueOnError {
				results = append(results, result)
				break
			}
		} else {
			result.Success = true
			result.Status = "LINKED"
			successCount++
		}
		results = append(results, result)
	}

	responseData := &responses.BulkLinkFarmersData{
		AAAOrgID:     bulkReq.AAAOrgID,
		TotalCount:   len(bulkReq.AAAUserIDs),
		SuccessCount: successCount,
		FailureCount: failureCount,
		SkippedCount: skippedCount,
		Results:      results,
	}

	return responseData, nil
}

// BulkUnlinkFarmersFromFPO unlinks multiple farmers from an FPO
func (s *FarmerLinkageServiceImpl) BulkUnlinkFarmersFromFPO(ctx context.Context, req interface{}) (interface{}, error) {
	bulkReq, ok := req.(*requests.BulkUnlinkFarmersRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for BulkUnlinkFarmersFromFPO")
	}

	// Validate input
	if bulkReq.AAAOrgID == "" {
		return nil, fmt.Errorf("aaa_org_id is required")
	}
	if len(bulkReq.AAAUserIDs) == 0 {
		return nil, fmt.Errorf("aaa_user_ids is required and must not be empty")
	}
	if len(bulkReq.AAAUserIDs) > 1000 {
		return nil, fmt.Errorf("aaa_user_ids cannot exceed 1000 entries")
	}

	results := make([]responses.BulkLinkResult, 0, len(bulkReq.AAAUserIDs))
	successCount := 0
	failureCount := 0
	skippedCount := 0

	for _, userID := range bulkReq.AAAUserIDs {
		result := responses.BulkLinkResult{
			AAAUserID: userID,
		}

		// Find existing link
		existingLink, err := s.getFarmerLinkByUserAndOrg(ctx, userID, bulkReq.AAAOrgID)
		if err != nil {
			result.Success = false
			result.Error = "farmer link not found"
			result.Status = "FAILED"
			failureCount++
			results = append(results, result)
			if !bulkReq.ContinueOnError {
				break
			}
			continue
		}

		// Check if already inactive
		if existingLink.Status == "INACTIVE" {
			result.Success = true
			result.Status = "ALREADY_UNLINKED"
			skippedCount++
			results = append(results, result)
			continue
		}

		// Soft delete by setting status to INACTIVE
		existingLink.Status = "INACTIVE"
		existingLink.KisanSathiUserID = nil

		if err := s.farmerLinkageRepo.Update(ctx, existingLink); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to unlink: %v", err)
			result.Status = "FAILED"
			failureCount++
			if !bulkReq.ContinueOnError {
				results = append(results, result)
				break
			}
		} else {
			result.Success = true
			result.Status = "UNLINKED"
			successCount++
		}
		results = append(results, result)
	}

	responseData := &responses.BulkLinkFarmersData{
		AAAOrgID:     bulkReq.AAAOrgID,
		TotalCount:   len(bulkReq.AAAUserIDs),
		SuccessCount: successCount,
		FailureCount: failureCount,
		SkippedCount: skippedCount,
		Results:      results,
	}

	return responseData, nil
}
