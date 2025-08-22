package farmer

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFarmerRepository creates a new farmer repository using BaseFilterableRepository
func NewFarmerRepository(dbManager interface{}) *base.BaseFilterableRepository[*entities.FarmerProfile] {
	repo := base.NewBaseFilterableRepository[*entities.FarmerProfile]()
	repo.SetDBManager(dbManager)
	return repo
}

// NewFarmerLinkRepository creates a new farmer link repository using BaseFilterableRepository
func NewFarmerLinkRepository(dbManager interface{}) *base.BaseFilterableRepository[*entities.FarmerLink] {
	repo := base.NewBaseFilterableRepository[*entities.FarmerLink]()
	repo.SetDBManager(dbManager)
	return repo
}
