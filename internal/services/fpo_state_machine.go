package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	repofpo "github.com/Kisanlink/farmers-module/internal/repo/fpo"
)

// FPOStateMachine handles FPO state transitions
type FPOStateMachine struct {
	repo *repofpo.FPORepository
}

// NewFPOStateMachine creates a new FPO state machine instance
func NewFPOStateMachine(repo *repofpo.FPORepository) *FPOStateMachine {
	return &FPOStateMachine{repo: repo}
}

// Transition performs a state transition with validation and side effects
func (sm *FPOStateMachine) Transition(ctx context.Context, fpoID string, targetState fpo.FPOStatus, reason string, performedBy string) error {
	// Get current FPO
	fpoRef, err := sm.repo.FindByID(ctx, fpoID)
	if err != nil {
		return fmt.Errorf("failed to find FPO: %w", err)
	}

	// Validate transition
	if !fpoRef.Status.CanTransitionTo(targetState) {
		return fmt.Errorf("cannot transition from %s to %s", fpoRef.Status, targetState)
	}

	// Perform transition-specific actions
	switch targetState {
	case fpo.FPOStatusVerified:
		if err := sm.onVerified(ctx, fpoRef); err != nil {
			return err
		}
	case fpo.FPOStatusActive:
		if err := sm.onActivated(ctx, fpoRef); err != nil {
			return err
		}
	case fpo.FPOStatusSuspended:
		if err := sm.onSuspended(ctx, fpoRef, reason); err != nil {
			return err
		}
	}

	// Update status with audit
	return sm.repo.UpdateStatus(ctx, fpoID, targetState, reason, performedBy)
}

// onVerified handles actions when FPO is verified
func (sm *FPOStateMachine) onVerified(ctx context.Context, fpoRef *fpo.FPORef) error {
	now := time.Now()
	fpoRef.VerifiedAt = &now
	return sm.repo.Update(ctx, fpoRef)
}

// onActivated handles actions when FPO is activated
func (sm *FPOStateMachine) onActivated(ctx context.Context, fpoRef *fpo.FPORef) error {
	// Clear setup errors
	fpoRef.SetupErrors = nil
	return sm.repo.Update(ctx, fpoRef)
}

// onSuspended handles actions when FPO is suspended
func (sm *FPOStateMachine) onSuspended(ctx context.Context, fpoRef *fpo.FPORef, reason string) error {
	// Log suspension (audit is handled by UpdateStatus)
	return nil
}
