package repo

import (
	"context"

	entities "github.com/Kisanlink/farmers-module/internal/entities"
	cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	cropStageEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_stage"
	cropVarietyEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	farmActivityEntity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	fpoEntity "github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/repo/bulk"
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/internal/repo/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer_linkage"
	"github.com/Kisanlink/farmers-module/internal/repo/fpo_ref"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RepositoryFactory provides access to all domain repositories
type RepositoryFactory struct {
	FarmerRepo           *base.BaseFilterableRepository[*entities.FarmerProfile]
	FarmerLinkageRepo    *base.BaseFilterableRepository[*entities.FarmerLink]
	FPORefRepo           *base.BaseFilterableRepository[*fpoEntity.FPORef]
	FarmRepo             *farm.FarmRepository
	CropRepo             *base.BaseFilterableRepository[*cropEntity.Crop]
	CropVarietyRepo      *base.BaseFilterableRepository[*cropVarietyEntity.CropVariety]
	CropStageRepo        *base.BaseFilterableRepository[*cropStageEntity.CropStage]
	CropCycleRepo        *base.BaseFilterableRepository[*cropCycleEntity.CropCycle]
	FarmActivityRepo     *base.BaseFilterableRepository[*farmActivityEntity.FarmActivity]
	BulkOperationRepo    bulk.BulkOperationRepository
	ProcessingDetailRepo bulk.ProcessingDetailRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(dbManager *db.PostgresManager) *RepositoryFactory {
	// Get GORM DB instance for bulk repositories
	gormDB, err := dbManager.GetDB(context.Background(), false)
	if err != nil {
		panic("Failed to get GORM DB: " + err.Error())
	}

	return &RepositoryFactory{
		FarmerRepo:           farmer.NewFarmerRepository(dbManager),
		FarmerLinkageRepo:    farmer_linkage.NewFarmerLinkageRepository(dbManager),
		FPORefRepo:           fpo_ref.NewFPORefRepository(dbManager),
		FarmRepo:             farm.NewFarmRepository(dbManager),
		CropRepo:             base.NewBaseFilterableRepository[*cropEntity.Crop](),
		CropVarietyRepo:      base.NewBaseFilterableRepository[*cropVarietyEntity.CropVariety](),
		CropStageRepo:        base.NewBaseFilterableRepository[*cropStageEntity.CropStage](),
		CropCycleRepo:        crop_cycle.NewRepository(dbManager),
		FarmActivityRepo:     farm_activity.NewFarmActivityRepository(dbManager),
		BulkOperationRepo:    bulk.NewBulkOperationRepository(gormDB),
		ProcessingDetailRepo: bulk.NewProcessingDetailRepository(gormDB),
	}
}
