package farmer

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// FarmerRepository defines the interface for farmer operations
type FarmerRepository interface {
	// Core farmer operations
	CreateFarmer(ctx context.Context, farmer *entities.FarmerProfile) error
	GetFarmerByID(ctx context.Context, aaaUserID, aaaOrgID string) (*entities.FarmerProfile, error)
	UpdateFarmer(ctx context.Context, farmer *entities.FarmerProfile) error
	DeleteFarmer(ctx context.Context, aaaUserID, aaaOrgID string) error

	// Farmer linkage operations
	CreateFarmerLink(ctx context.Context, link *entities.FarmerLink) error
	GetFarmerLink(ctx context.Context, aaaUserID, aaaOrgID string) (*entities.FarmerLink, error)
	UpdateFarmerLink(ctx context.Context, link *entities.FarmerLink) error
	DeleteFarmerLink(ctx context.Context, aaaUserID, aaaOrgID string) error

	// Query operations
	GetFarmersByOrg(ctx context.Context, aaaOrgID string) ([]*entities.FarmerProfile, error)
	GetFarmerLinksByUser(ctx context.Context, aaaUserID string) ([]*entities.FarmerLink, error)
	GetFarmerLinksByOrg(ctx context.Context, aaaOrgID string) ([]*entities.FarmerLink, error)
}

// farmerRepo implements FarmerRepository
type farmerRepo struct {
	postgresManager *db.PostgresManager
}

// NewFarmerRepository creates a new farmer repository
func NewFarmerRepository(postgresManager *db.PostgresManager) FarmerRepository {
	return &farmerRepo{postgresManager: postgresManager}
}

// CreateFarmer creates a new farmer profile
func (r *farmerRepo) CreateFarmer(ctx context.Context, farmer *entities.FarmerProfile) error {
	// Implementation will be added when we have the farmer table
	return fmt.Errorf("farmer table not yet implemented")
}

// GetFarmerByID retrieves a farmer profile by AAA user ID and org ID
func (r *farmerRepo) GetFarmerByID(ctx context.Context, aaaUserID, aaaOrgID string) (*entities.FarmerProfile, error) {
	// Implementation will be added when we have the farmer table
	return nil, fmt.Errorf("farmer table not yet implemented")
}

// UpdateFarmer updates an existing farmer profile
func (r *farmerRepo) UpdateFarmer(ctx context.Context, farmer *entities.FarmerProfile) error {
	// Implementation will be added when we have the farmer table
	return fmt.Errorf("farmer table not yet implemented")
}

// DeleteFarmer deletes a farmer profile
func (r *farmerRepo) DeleteFarmer(ctx context.Context, aaaUserID, aaaOrgID string) error {
	// Implementation will be added when we have the farmer table
	return fmt.Errorf("farmer table not yet implemented")
}

// CreateFarmerLink creates a new farmer link
func (r *farmerRepo) CreateFarmerLink(ctx context.Context, link *entities.FarmerLink) error {
	return r.postgresManager.Create(ctx, link)
}

// GetFarmerLink retrieves a farmer link by AAA user ID and org ID
func (r *farmerRepo) GetFarmerLink(ctx context.Context, aaaUserID, aaaOrgID string) (*entities.FarmerLink, error) {
	var link entities.FarmerLink

	// Use the List method with filters to find the specific link
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_user_id", Operator: base.OpEqual, Value: aaaUserID},
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: aaaOrgID},
			},
			Logic: base.LogicAnd,
		},
		Limit: 1,
	}

	err := r.postgresManager.List(ctx, filter, &link)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

// UpdateFarmerLink updates an existing farmer link
func (r *farmerRepo) UpdateFarmerLink(ctx context.Context, link *entities.FarmerLink) error {
	return r.postgresManager.Update(ctx, link)
}

// DeleteFarmerLink soft deletes a farmer link
func (r *farmerRepo) DeleteFarmerLink(ctx context.Context, aaaUserID, aaaOrgID string) error {
	// First get the link to soft delete it
	var link entities.FarmerLink
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_user_id", Operator: base.OpEqual, Value: aaaUserID},
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: aaaOrgID},
			},
			Logic: base.LogicAnd,
		},
		Limit: 1,
	}

	err := r.postgresManager.List(ctx, filter, &link)
	if err != nil {
		return err
	}

	return r.postgresManager.SoftDelete(ctx, aaaUserID, &link, "")
}

// GetFarmerLinksByUser retrieves all farmer links for a specific user
func (r *farmerRepo) GetFarmerLinksByUser(ctx context.Context, aaaUserID string) ([]*entities.FarmerLink, error) {
	var links []*entities.FarmerLink

	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_user_id", Operator: base.OpEqual, Value: aaaUserID},
			},
			Logic: base.LogicAnd,
		},
		Sort: []base.SortField{
			{Field: "created_at", Direction: "desc"},
		},
	}

	err := r.postgresManager.List(ctx, filter, &links)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// GetFarmerLinksByOrg retrieves all farmer links for a specific organization
func (r *farmerRepo) GetFarmerLinksByOrg(ctx context.Context, aaaOrgID string) ([]*entities.FarmerLink, error) {
	var links []*entities.FarmerLink

	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: aaaOrgID},
			},
			Logic: base.LogicAnd,
		},
		Sort: []base.SortField{
			{Field: "created_at", Direction: "desc"},
		},
	}

	err := r.postgresManager.List(ctx, filter, &links)
	if err != nil {
		return nil, err
	}

	return links, nil
}

// GetFarmersByOrg retrieves all farmers for a specific organization
func (r *farmerRepo) GetFarmersByOrg(ctx context.Context, aaaOrgID string) ([]*entities.FarmerProfile, error) {
	// Implementation will be added when we have the farmer table
	return nil, fmt.Errorf("farmer table not yet implemented")
}
