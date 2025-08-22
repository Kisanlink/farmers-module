package repo

import (
	entities "github.com/Kisanlink/farmers-module/internal/entities"
	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	farmActivityEntity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	fpoEntity "github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/internal/repo/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer_linkage"
	"github.com/Kisanlink/farmers-module/internal/repo/fpo_ref"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// RepositoryFactory provides access to all domain repositories
type RepositoryFactory struct {
	FarmerRepo        *base.BaseFilterableRepository[*entities.FarmerProfile]
	FarmerLinkageRepo *base.BaseFilterableRepository[*entities.FarmerLink]
	FPORefRepo        *base.BaseFilterableRepository[*fpoEntity.FPORef]
	FarmRepo          *base.BaseFilterableRepository[*farmEntity.Farm]
	CropCycleRepo     *base.BaseFilterableRepository[*cropCycleEntity.CropCycle]
	FarmActivityRepo  *base.BaseFilterableRepository[*farmActivityEntity.FarmActivity]
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(dbManager interface{}) *RepositoryFactory {
	return &RepositoryFactory{
		FarmerRepo:        farmer.NewFarmerRepository(dbManager),
		FarmerLinkageRepo: farmer_linkage.NewFarmerLinkageRepository(dbManager),
		FPORefRepo:        fpo_ref.NewFPORefRepository(dbManager),
		FarmRepo:          farm.NewFarmRepository(dbManager),
		CropCycleRepo:     crop_cycle.NewRepository(dbManager),
		FarmActivityRepo:  farm_activity.NewFarmActivityRepository(dbManager),
	}
}
