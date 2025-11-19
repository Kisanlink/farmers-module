package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	repofpo "github.com/Kisanlink/farmers-module/internal/repo/fpo"
)

const MaxSetupRetries = 3

// FPOLifecycleService handles FPO lifecycle operations
type FPOLifecycleService struct {
	repo         *repofpo.FPORepository
	stateMachine *FPOStateMachine
	aaaService   AAAService
}

// NewFPOLifecycleService creates a new FPO lifecycle service instance
func NewFPOLifecycleService(repo *repofpo.FPORepository, aaaService AAAService) *FPOLifecycleService {
	return &FPOLifecycleService{
		repo:         repo,
		stateMachine: NewFPOStateMachine(repo),
		aaaService:   aaaService,
	}
}

// SyncFPOFromAAA synchronizes FPO reference from AAA service
// This is the key method that solves the "no matching records found" error
func (s *FPOLifecycleService) SyncFPOFromAAA(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
	log.Printf("FPOLifecycleService: Syncing FPO from AAA for org ID: %s", aaaOrgID)

	// Check if already exists
	existing, _ := s.repo.FindByAAAOrgID(ctx, aaaOrgID)
	if existing != nil {
		log.Printf("FPO already exists for org %s, returning existing reference", aaaOrgID)
		return existing, nil
	}

	// Get organization from AAA
	org, err := s.aaaService.GetOrganization(ctx, aaaOrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization from AAA: %w", err)
	}

	orgMap, ok := org.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid organization response from AAA")
	}

	// Extract organization details
	name := ""
	if n, ok := orgMap["name"].(string); ok {
		name = n
	}

	metadata := make(entities.JSONB)
	if m, ok := orgMap["metadata"].(map[string]interface{}); ok {
		metadata = m
	}

	// Create local reference using constructor
	// Constructor ensures ID is properly initialized
	fpoRef := fpo.NewFPORef(aaaOrgID)
	fpoRef.Name = name
	fpoRef.Status = fpo.FPOStatusActive // Assume active if exists in AAA
	fpoRef.Metadata = metadata

	// Try to extract registration number if available
	if regNo, ok := orgMap["registration_number"].(string); ok {
		fpoRef.RegistrationNo = regNo
	}

	if err := s.repo.Create(ctx, fpoRef); err != nil {
		return nil, fmt.Errorf("failed to create FPO reference: %w", err)
	}

	log.Printf("Successfully synchronized FPO %s from AAA (org_id: %s)", fpoRef.ID, aaaOrgID)
	return fpoRef, nil
}

// GetOrSyncFPO attempts to get FPO from local DB, syncs from AAA if not found
// This method provides automatic recovery from missing FPO references
func (s *FPOLifecycleService) GetOrSyncFPO(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
	// Try local database first
	fpoRef, err := s.repo.FindByAAAOrgID(ctx, aaaOrgID)
	if err == nil && fpoRef != nil {
		return fpoRef, nil
	}

	// If not found locally, attempt sync from AAA
	log.Printf("FPO not found locally for org %s, attempting sync from AAA", aaaOrgID)
	return s.SyncFPOFromAAA(ctx, aaaOrgID)
}

// RetryFailedSetup retries failed setup operations
func (s *FPOLifecycleService) RetryFailedSetup(ctx context.Context, fpoID string) error {
	fpoRef, err := s.repo.FindByID(ctx, fpoID)
	if err != nil {
		return fmt.Errorf("failed to find FPO: %w", err)
	}

	if fpoRef.Status != fpo.FPOStatusSetupFailed {
		return fmt.Errorf("FPO is not in SETUP_FAILED status, current status: %s", fpoRef.Status)
	}

	// Check retry limit
	if fpoRef.SetupAttempts >= MaxSetupRetries {
		return fmt.Errorf("maximum setup retries (%d) exceeded", MaxSetupRetries)
	}

	// Increment retry attempt
	if err := s.repo.IncrementSetupAttempt(ctx, fpoID); err != nil {
		return fmt.Errorf("failed to increment setup attempts: %w", err)
	}

	// Transition back to pending setup
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}

	if err := s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusPendingSetup,
		fmt.Sprintf("Retry attempt %d", fpoRef.SetupAttempts+1), userID); err != nil {
		return fmt.Errorf("failed to transition to PENDING_SETUP: %w", err)
	}

	// Update retry tracking
	now := time.Now()
	fpoRef.LastSetupAt = &now
	if err := s.repo.Update(ctx, fpoRef); err != nil {
		return fmt.Errorf("failed to update FPO: %w", err)
	}

	// Trigger setup asynchronously
	go s.performAAASetup(context.Background(), fpoID)

	return nil
}

// performAAASetup performs AAA setup operations
func (s *FPOLifecycleService) performAAASetup(ctx context.Context, fpoID string) {
	fpoRef, err := s.repo.FindByID(ctx, fpoID)
	if err != nil {
		log.Printf("Failed to find FPO for setup: %v", err)
		return
	}

	setupErrors := make(entities.JSONB)

	// Create organization in AAA if not exists
	if fpoRef.AAAOrgID == "" {
		orgResp, err := s.aaaService.CreateOrganization(ctx, map[string]interface{}{
			"name":     fpoRef.Name,
			"type":     "FPO",
			"metadata": fpoRef.Metadata,
		})

		if err != nil {
			setupErrors["organization"] = err.Error()
		} else if orgMap, ok := orgResp.(map[string]interface{}); ok {
			if orgID, ok := orgMap["org_id"].(string); ok {
				fpoRef.AAAOrgID = orgID
			}
		}
	}

	// Create user groups if organization exists
	if fpoRef.AAAOrgID != "" {
		groupErrors := s.createUserGroups(ctx, fpoRef.AAAOrgID, fpoRef.Name)
		for k, v := range groupErrors {
			setupErrors[k] = v
		}
	}

	// Update status based on results
	var newStatus fpo.FPOStatus
	var reason string

	if len(setupErrors) == 0 {
		newStatus = fpo.FPOStatusActive
		reason = "Setup completed successfully"
		fpoRef.SetupErrors = nil
	} else {
		newStatus = fpo.FPOStatusSetupFailed
		reason = fmt.Sprintf("Setup failed with %d errors", len(setupErrors))
		fpoRef.SetupErrors = setupErrors
	}

	// Update FPO
	now := time.Now()
	fpoRef.LastSetupAt = &now
	if err := s.repo.Update(ctx, fpoRef); err != nil {
		log.Printf("Failed to update FPO after setup: %v", err)
		return
	}

	// Transition state
	if err := s.stateMachine.Transition(ctx, fpoID, newStatus, reason, "system"); err != nil {
		log.Printf("Failed to transition FPO state after setup: %v", err)
	}
}

// createUserGroups creates user groups for FPO
func (s *FPOLifecycleService) createUserGroups(ctx context.Context, aaaOrgID string, fpoName string) map[string]interface{} {
	errors := make(map[string]interface{})
	groupNames := []string{"directors", "shareholders", "store_staff", "store_managers"}

	for _, groupName := range groupNames {
		createGroupReq := map[string]interface{}{
			"name":        groupName,
			"description": fmt.Sprintf("%s group for %s", groupName, fpoName),
			"org_id":      aaaOrgID,
		}

		_, err := s.aaaService.CreateUserGroup(ctx, createGroupReq)
		if err != nil {
			log.Printf("Warning: Failed to create user group %s: %v", groupName, err)
			errors[fmt.Sprintf("user_group_%s", groupName)] = err.Error()
		}
	}

	return errors
}

// GetFPOHistory retrieves audit history for an FPO
func (s *FPOLifecycleService) GetFPOHistory(ctx context.Context, fpoID string) ([]*fpo.FPOAuditLog, error) {
	return s.repo.GetAuditHistory(ctx, fpoID)
}

// SuspendFPO suspends an FPO
func (s *FPOLifecycleService) SuspendFPO(ctx context.Context, fpoID string, reason string) error {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	return s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusSuspended, reason, userID)
}

// ReactivateFPO reactivates a suspended FPO
func (s *FPOLifecycleService) ReactivateFPO(ctx context.Context, fpoID string) error {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	return s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusActive, "Reactivated", userID)
}

// DeactivateFPO deactivates an FPO
func (s *FPOLifecycleService) DeactivateFPO(ctx context.Context, fpoID string, reason string) error {
	userID := GetUserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	return s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusInactive, reason, userID)
}

// GetUserIDFromContext extracts user ID from context (placeholder implementation)
func GetUserIDFromContext(ctx context.Context) string {
	// TODO: Implement actual context extraction based on your auth system
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}
