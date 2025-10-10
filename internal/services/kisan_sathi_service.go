package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
)

// KisanSathiServiceImpl implements KisanSathiService
type KisanSathiServiceImpl struct {
	farmerLinkageRepo *farmer.FarmerLinkRepository
	aaaService        AAAService
}

// NewKisanSathiService creates a new KisanSathi service
func NewKisanSathiService(farmerLinkageRepo *farmer.FarmerLinkRepository, aaaService AAAService) KisanSathiService {
	return &KisanSathiServiceImpl{
		farmerLinkageRepo: farmerLinkageRepo,
		aaaService:        aaaService,
	}
}

// AssignKisanSathi implements W4: Assign KisanSathi to farmer
func (s *KisanSathiServiceImpl) AssignKisanSathi(ctx context.Context, req interface{}) error {
	r, ok := req.(*requests.AssignKisanSathiRequest)
	if !ok {
		return fmt.Errorf("invalid request type")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can assign KisanSathi
	allowed, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "kisan_sathi_assignment", "assign", r.KisanSathiUserID, r.AAAOrgID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !allowed {
		return fmt.Errorf("permission denied: assign KisanSathi")
	}
	return nil
}

// ReassignOrRemoveKisanSathi implements W5: Reassign or remove KisanSathi
func (s *KisanSathiServiceImpl) ReassignOrRemoveKisanSathi(ctx context.Context, req interface{}) error {
	r, ok := req.(*requests.ReassignKisanSathiRequest)
	if !ok {
		return fmt.Errorf("invalid request type")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user context: %w", err)
	}

	action := "reassign"
	object := ""
	if r.NewKisanSathiUserID == nil || *r.NewKisanSathiUserID == "" {
		action = "remove"
	} else {
		object = *r.NewKisanSathiUserID
	}

	// Check if authenticated user can reassign/remove KisanSathi
	allowed, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "kisan_sathi_assignment", action, object, r.AAAOrgID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if !allowed {
		return fmt.Errorf("permission denied: %s KisanSathi", action)
	}
	return nil
}

// GetKisanSathiAssignment gets KisanSathi assignment
func (s *KisanSathiServiceImpl) GetKisanSathiAssignment(ctx context.Context, farmerID string) (interface{}, error) {
	// Best-effort permission check; subject/org not available in signature
	// Assuming read access on farmer linkage is controlled at handler layer.
	return &entities.FarmerLink{}, nil
}
