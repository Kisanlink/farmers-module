package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	fpoRefRepo "github.com/Kisanlink/farmers-module/internal/repo/fpo_ref"
)

// FPORefServiceImpl implements FPORefService
type FPORefServiceImpl struct {
	fpoRefRepo fpoRefRepo.FPORefRepository
	aaaClient  *aaa.Client
}

// NewFPORefService creates a new FPO ref service
func NewFPORefService(fpoRefRepo fpoRefRepo.FPORefRepository, aaaClient *aaa.Client) FPORefService {
	return &FPORefServiceImpl{
		fpoRefRepo: fpoRefRepo,
		aaaClient:  aaaClient,
	}
}

// RegisterFPORef implements W3: Register FPO reference with AAA service integration
func (s *FPORefServiceImpl) RegisterFPORef(ctx context.Context, req interface{}) error {
	// Type assert the request
	registerReq, ok := req.(requests.RegisterFPORefRequest)
	if !ok {
		return fmt.Errorf("invalid request type for FPO registration")
	}

	log.Printf("Registering FPO reference: %s", registerReq.AAAOrgID)

	// Step 1: Create organization in AAA service
	orgData, err := s.aaaClient.CreateOrganization(
		ctx,
		registerReq.AAAOrgID, // Using AAAOrgID as name for now
		registerReq.BusinessConfig,
		"fpo", // Default type for FPO
		map[string]string{
			"business_config": registerReq.BusinessConfig,
			"source":          "farmers-module",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create organization in AAA service: %w", err)
	}

	log.Printf("Organization created in AAA service: %s", orgData.ID)

	// Step 2: Save FPO reference in local database
	fpoRefData := &fpo.FPORef{
		AAAOrgID: orgData.ID, // Use the AAA service generated ID
		BusinessConfig: map[string]string{
			"business_config": registerReq.BusinessConfig,
			"source":          "farmers-module",
		},
	}

	// Save to repository
	if err := s.fpoRefRepo.Create(ctx, fpoRefData); err != nil {
		// TODO: Consider rolling back the AAA service organization creation
		return fmt.Errorf("failed to save FPO reference: %w", err)
	}

	log.Printf("FPO reference saved successfully: %s", fpoRefData.ID)
	return nil
}

// GetFPORef gets FPO reference with AAA service validation
func (s *FPORefServiceImpl) GetFPORef(ctx context.Context, orgID string) (interface{}, error) {
	log.Printf("Getting FPO reference: %s", orgID)

	// Step 1: Verify organization exists in AAA service
	orgData, err := s.aaaClient.VerifyOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify organization in AAA service: %w", err)
	}

	// Step 2: Get local FPO reference data
	fpoRef, err := s.fpoRefRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		// If not found locally, return AAA service data
		log.Printf("FPO reference not found locally, returning AAA service data")
		return &responses.FPORefData{
			ID:             orgData.ID,
			AAAOrgID:       orgData.ID,
			BusinessConfig: orgData.Description,
			Status:         orgData.Status,
		}, nil
	}

	return fpoRef, nil
}

// UpdateFPORef updates FPO reference with AAA service synchronization
func (s *FPORefServiceImpl) UpdateFPORef(ctx context.Context, orgID string, req interface{}) error {
	// Type assert the request
	updateReq, ok := req.(requests.RegisterFPORefRequest)
	if !ok {
		return fmt.Errorf("invalid request type for FPO update")
	}

	log.Printf("Updating FPO reference: %s", orgID)

	// Step 1: Verify organization exists in AAA service
	_, err := s.aaaClient.VerifyOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to verify organization in AAA service: %w", err)
	}

	// Step 2: Update local FPO reference
	fpoRefData := &fpo.FPORef{
		AAAOrgID: orgID,
		BusinessConfig: map[string]string{
			"business_config": updateReq.BusinessConfig,
			"source":          "farmers-module",
		},
	}

	if err := s.fpoRefRepo.Update(ctx, fpoRefData); err != nil {
		return fmt.Errorf("failed to update FPO reference: %w", err)
	}

	log.Printf("FPO reference updated successfully: %s", orgID)
	return nil
}

// DeleteFPORef deletes FPO reference with AAA service cleanup
func (s *FPORefServiceImpl) DeleteFPORef(ctx context.Context, orgID string) error {
	log.Printf("Deleting FPO reference: %s", orgID)

	// Step 1: Verify organization exists in AAA service
	_, err := s.aaaClient.VerifyOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to verify organization in AAA service: %w", err)
	}

	// Step 2: Delete local FPO reference
	if err := s.fpoRefRepo.Delete(ctx, orgID); err != nil {
		return fmt.Errorf("failed to delete FPO reference: %w", err)
	}

	// TODO: Consider deactivating the organization in AAA service instead of deletion
	log.Printf("FPO reference deleted successfully: %s", orgID)
	return nil
}
