package repo

import (
	"context"

	fpoConfigEntity "github.com/Kisanlink/farmers-module/internal/entities/fpo_config"
	"github.com/Kisanlink/farmers-module/internal/repo/bulk"
	"github.com/Kisanlink/farmers-module/internal/repo/crop"
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/repo/farm"
	"github.com/Kisanlink/farmers-module/internal/repo/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/repo/farmer"
	"github.com/Kisanlink/farmers-module/internal/repo/fpo"
	"github.com/Kisanlink/farmers-module/internal/repo/fpo_config"
	"github.com/Kisanlink/farmers-module/internal/repo/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/repo/soil_type"
	"github.com/Kisanlink/farmers-module/internal/repo/stage"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// RepositoryFactory provides access to all domain repositories
type RepositoryFactory struct {
	FarmerRepo           *farmer.FarmerRepository
	FarmerLinkageRepo    *farmer.FarmerLinkRepository
	FPORefRepo           *fpo.FPORepository
	FPOConfigRepo        *base.BaseFilterableRepository[*fpoConfigEntity.FPOConfig]
	FarmRepo             *farm.FarmRepository
	CropRepo             *crop.CropRepository
	CropVarietyRepo      *crop.CropVarietyRepository
	CropCycleRepo        *crop_cycle.CropCycleRepository
	FarmActivityRepo     *farm_activity.FarmActivityRepository
	BulkOperationRepo    bulk.BulkOperationRepository
	ProcessingDetailRepo bulk.ProcessingDetailRepository
	StageRepo            *stage.StageRepository
	CropStageRepo        *stage.CropStageRepository
	SoilTypeRepo         *soil_type.SoilTypeRepository
	IrrigationSourceRepo *irrigation_source.IrrigationSourceRepository
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
		FarmerLinkageRepo:    farmer.NewFarmerLinkRepository(dbManager),
		FPORefRepo:           fpo.NewFPORepository(dbManager),
		FPOConfigRepo:        fpo_config.NewFPOConfigRepository(dbManager),
		FarmRepo:             farm.NewFarmRepository(dbManager),
		CropRepo:             crop.NewCropRepository(dbManager),
		CropVarietyRepo:      crop.NewCropVarietyRepository(dbManager),
		CropCycleRepo:        crop_cycle.NewRepository(dbManager),
		FarmActivityRepo:     farm_activity.NewFarmActivityRepository(dbManager),
		BulkOperationRepo:    bulk.NewBulkOperationRepository(gormDB),
		ProcessingDetailRepo: bulk.NewProcessingDetailRepository(gormDB),
		StageRepo:            stage.NewStageRepository(dbManager),
		CropStageRepo:        stage.NewCropStageRepository(dbManager),
		SoilTypeRepo:         soil_type.NewSoilTypeRepository(dbManager),
		IrrigationSourceRepo: irrigation_source.NewIrrigationSourceRepository(dbManager),
	}
}
