package farmer_linkage

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFarmerLinkageRepository creates a new farmer linkage repository using BaseFilterableRepository
func NewFarmerLinkageRepository(dbManager interface{}) *base.BaseFilterableRepository[*entities.FarmerLink] {
	repo := base.NewBaseFilterableRepository[*entities.FarmerLink]()
	repo.SetDBManager(dbManager)
	return repo
}
