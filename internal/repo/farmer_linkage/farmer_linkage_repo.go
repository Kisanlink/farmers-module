package farmer_linkage

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// FarmerLinkageRepository defines the interface for farmer linkage operations
type FarmerLinkageRepository interface {
	Create(ctx context.Context, link *entities.FarmerLink) error
	GetByUserAndOrg(ctx context.Context, userID, orgID string) (*entities.FarmerLink, error)
	Update(ctx context.Context, link *entities.FarmerLink) error
	SoftDelete(ctx context.Context, userID, orgID string) error
	GetByUserID(ctx context.Context, userID string) ([]*entities.FarmerLink, error)
	GetByOrgID(ctx context.Context, orgID string) ([]*entities.FarmerLink, error)
}

// farmerLinkageRepo implements FarmerLinkageRepository
type farmerLinkageRepo struct {
	postgresManager *db.PostgresManager
}

// NewFarmerLinkageRepository creates a new farmer linkage repository
func NewFarmerLinkageRepository(postgresManager *db.PostgresManager) FarmerLinkageRepository {
	return &farmerLinkageRepo{postgresManager: postgresManager}
}

// Create creates a new farmer linkage
func (r *farmerLinkageRepo) Create(ctx context.Context, link *entities.FarmerLink) error {
	return r.postgresManager.Create(ctx, link)
}

// GetByUserAndOrg retrieves a farmer linkage by user ID and organization ID
func (r *farmerLinkageRepo) GetByUserAndOrg(ctx context.Context, userID, orgID string) (*entities.FarmerLink, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_user_id", Operator: base.OpEqual, Value: userID},
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
		Limit: 1,
	}

	var links []*entities.FarmerLink
	err := r.postgresManager.List(ctx, filter, &links)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, fmt.Errorf("farmer linkage not found")
	}

	return links[0], nil
}

// Update updates an existing farmer linkage
func (r *farmerLinkageRepo) Update(ctx context.Context, link *entities.FarmerLink) error {
	return r.postgresManager.Update(ctx, link)
}

// SoftDelete soft deletes a farmer linkage
func (r *farmerLinkageRepo) SoftDelete(ctx context.Context, userID, orgID string) error {
	link, err := r.GetByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return err
	}

	link.Status = "INACTIVE"
	return r.postgresManager.Update(ctx, link)
}

// GetByUserID retrieves all linkages for a specific user
func (r *farmerLinkageRepo) GetByUserID(ctx context.Context, userID string) ([]*entities.FarmerLink, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_user_id", Operator: base.OpEqual, Value: userID},
			},
			Logic: base.LogicAnd,
		},
	}

	var links []*entities.FarmerLink
	err := r.postgresManager.List(ctx, filter, &links)
	return links, err
}

// GetByOrgID retrieves all linkages for a specific organization
func (r *farmerLinkageRepo) GetByOrgID(ctx context.Context, orgID string) ([]*entities.FarmerLink, error) {
	filter := &base.Filter{
		Group: base.FilterGroup{
			Conditions: []base.FilterCondition{
				{Field: "aaa_org_id", Operator: base.OpEqual, Value: orgID},
			},
			Logic: base.LogicAnd,
		},
	}

	var links []*entities.FarmerLink
	err := r.postgresManager.List(ctx, filter, &links)
	return links, err
}
