package repo

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	cropStageEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_stage"
	cropVarietyEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
	farmActivityEntity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	fpoEntity "github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/farmers-module/internal/repo/bulk"
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/internal/repo/farm_activity"
	farmerRepo "github.com/Kisanlink/farmers-module/internal/repo/farmer"
	farmerLinkageRepo "github.com/Kisanlink/farmers-module/internal/repo/farmer_linkage"
	fpoRefRepo "github.com/Kisanlink/farmers-module/internal/repo/fpo_ref"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// RepositoryFactory provides access to all domain repositories
type RepositoryFactory struct {
	FarmerRepo           *base.BaseFilterableRepository[*farmer.Farmer]
	FarmerLinkageRepo    *base.BaseFilterableRepository[*entities.FarmerLink]
	FPORefRepo           *base.BaseFilterableRepository[*fpoEntity.FPORef]
	FarmRepo             *farm.FarmRepository
	CropRepo             *base.BaseFilterableRepository[*crop.Crop]
	CropVarietyRepo      *base.BaseFilterableRepository[*cropVarietyEntity.CropVariety]
	CropStageRepo        *base.BaseFilterableRepository[*cropStageEntity.CropStage]
	CropCycleRepo        *base.BaseFilterableRepository[*cropCycleEntity.CropCycle]
	FarmActivityRepo     *base.BaseFilterableRepository[*farmActivityEntity.FarmActivity]
	BulkOperationRepo    bulk.BulkOperationRepository
	ProcessingDetailRepo bulk.ProcessingDetailRepository
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(dbManager db.DBManager) *RepositoryFactory {
	return &RepositoryFactory{
		FarmerRepo:           farmerRepo.NewFarmerRepository(dbManager),
		FarmerLinkageRepo:    farmerLinkageRepo.NewFarmerLinkageRepository(dbManager),
		FPORefRepo:           fpoRefRepo.NewFPORefRepository(dbManager),
		FarmRepo:             farm.NewFarmRepository(dbManager),
		CropRepo:             base.NewBaseFilterableRepository[*crop.Crop](),
		CropVarietyRepo:      base.NewBaseFilterableRepository[*cropVarietyEntity.CropVariety](),
		CropStageRepo:        base.NewBaseFilterableRepository[*cropStageEntity.CropStage](),
		CropCycleRepo:        crop_cycle.NewRepository(dbManager),
		FarmActivityRepo:     farm_activity.NewFarmActivityRepository(dbManager),
		BulkOperationRepo:    bulk.NewBulkOperationRepository(getGormDB(dbManager)),
		ProcessingDetailRepo: bulk.NewProcessingDetailRepository(getGormDB(dbManager)),
	}
}

// getGormDB extracts the gorm.DB from the dbManager
func getGormDB(dbManager db.DBManager) *gorm.DB {
	// Try to cast to PostgresManager to access GetDB method
	if postgresManager, ok := dbManager.(*db.PostgresManager); ok {
		if db, err := postgresManager.GetDB(context.Background(), false); err == nil {
			return db
		}
	}
	return nil
}
