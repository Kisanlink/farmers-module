package repo

import (
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/internal/repo/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer_linkage"
	"github.com/Kisanlink/farmers-module/internal/repo/fpo_ref"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RepositoryFactory provides access to all domain repositories
type RepositoryFactory struct {
	PostgresManager   *db.PostgresManager
	FarmerRepo        farmer.FarmerRepository
	FarmerLinkageRepo farmer_linkage.FarmerLinkageRepository
	FPORefRepo        fpo_ref.FPORefRepository
	FarmRepo          farm.FarmRepository
	CropCycleRepo     crop_cycle.CropCycleRepository
	FarmActivityRepo  farm_activity.FarmActivityRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(postgresManager *db.PostgresManager) *RepositoryFactory {
	return &RepositoryFactory{
		FarmerRepo:        farmer.NewFarmerRepository(postgresManager),
		FarmerLinkageRepo: farmer_linkage.NewFarmerLinkageRepository(postgresManager),
		FPORefRepo:        fpo_ref.NewFPORefRepository(postgresManager),
		FarmRepo:          farm.NewFarmRepository(postgresManager),
		CropCycleRepo:     crop_cycle.NewCropCycleRepository(postgresManager),
		FarmActivityRepo:  farm_activity.NewFarmActivityRepository(postgresManager),
	}
}
