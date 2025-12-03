package services

import (
	"context"
	"log"
	"time"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/repo"
	repofpo "github.com/Kisanlink/farmers-module/internal/repo/fpo"
	"github.com/Kisanlink/farmers-module/internal/services/audit"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"gorm.io/gorm"
)

// ServiceFactory provides access to all domain services
type ServiceFactory struct {
	// Identity & Organization Services
	FarmerService        FarmerService
	FarmerLinkageService FarmerLinkageService
	FPOService           FPOService
	FPOLifecycleService  *FPOLifecycleService
	FPOConfigService     FPOConfigService
	KisanSathiService    KisanSathiService

	// Farm Management Services
	FarmService FarmService

	// Crop Management Services
	CropService         CropService
	CropCycleService    CropCycleService
	FarmActivityService FarmActivityService

	// Data Quality Services
	DataQualityService DataQualityService

	// Lookup Services
	LookupService LookupService

	// Reporting Services
	ReportingService ReportingService

	// Administrative Services
	AdministrativeService AdministrativeService

	// Bulk Operations Services
	BulkFarmerService BulkFarmerService

	// AAA Integration Service
	AAAService AAAService

	// Audit Service
	AuditService *audit.AuditService

	// AAA Client for direct integration
	AAAClient *aaa.Client

	// Stage Management Services
	StageService StageService

	// Background Jobs
	ReconciliationJob *ReconciliationJob
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(repoFactory *repo.RepositoryFactory, postgresManager *db.PostgresManager, cfg *config.Config, logger interfaces.Logger) *ServiceFactory {
	var aaaClient *aaa.Client
	var err error

	// Initialize AAA client only if enabled
	if cfg.AAA.Enabled {
		aaaClient, err = aaa.NewClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to create AAA client: %v", err)
			log.Printf("Services will run in degraded mode without AAA integration")
			aaaClient = nil
		}
	} else {
		log.Printf("AAA service integration is disabled")
		aaaClient = nil
	}

	// Initialize AAA service
	aaaService := NewAAAService(cfg)

	// Initialize FPO config service first (needed by farmer service)
	fpoConfigService := NewFPOConfigService(repoFactory.FPOConfigRepo)

	// Initialize identity services
	// Use NewFarmerServiceWithFPOConfig to enable FPO config linking
	farmerService := NewFarmerServiceWithFPOConfig(repoFactory.FarmerRepo, aaaService, fpoConfigService, cfg.AAA.DefaultPassword)
	farmerLinkageService := NewFarmerLinkageService(repoFactory.FarmerLinkageRepo, repoFactory.FarmerRepo, aaaService)
	fpoService := NewFPOService(repoFactory.FPORefRepo, aaaService)

	// Initialize FPO lifecycle service with enhanced repository
	fpoRepo := repofpo.NewFPORepository(postgresManager)
	fpoLifecycleService := NewFPOLifecycleService(fpoRepo, aaaService)
	kisanSathiService := NewKisanSathiService(repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize farm management services
	// Get GORM DB for farm service
	var gormDB *gorm.DB
	if db, err := postgresManager.GetDB(context.Background(), false); err == nil {
		gormDB = db
	}
	farmService := NewFarmService(repoFactory.FarmRepo, repoFactory.FarmerRepo, aaaService, gormDB)

	// Initialize crop management services
	cropService := NewCropService(repoFactory.CropRepo, repoFactory.CropVarietyRepo, aaaService)
	cropCycleService := NewCropCycleService(repoFactory.CropCycleRepo, farmService, aaaService)
	farmActivityService := NewFarmActivityService(repoFactory.FarmActivityRepo, repoFactory.CropCycleRepo, repoFactory.CropStageRepo, repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize notification service
	notificationService := NewNotificationService(aaaService)

	// Initialize data quality service
	dataQualityService := NewDataQualityService(gormDB, repoFactory.FarmRepo, repoFactory.FarmerLinkageRepo, aaaService, notificationService)

	// Initialize lookup service
	lookupService := NewLookupService(gormDB)

	// Initialize reporting service
	reportingService := NewReportingService(repoFactory, gormDB, aaaService)

	// Initialize administrative service
	concreteAdminService := NewAdministrativeService(
		postgresManager,
		gormDB,
		aaaService,
		repoFactory.SoilTypeRepo,
		repoFactory.IrrigationSourceRepo,
	)
	administrativeService := NewAdministrativeServiceWrapper(concreteAdminService)

	// Initialize bulk farmer service
	bulkFarmerService := NewBulkFarmerService(
		repoFactory.BulkOperationRepo,
		repoFactory.ProcessingDetailRepo,
		farmerService,
		farmerLinkageService,
		aaaService,
		logger,
	)

	// Initialize audit service
	auditService := audit.NewAuditService(logger.GetZapLogger(), nil) // No remote client for now

	// Initialize stage service
	stageService := NewStageService(
		repoFactory.StageRepo,
		repoFactory.CropStageRepo,
		aaaService,
	)

	// Initialize reconciliation job (runs every 5 minutes)
	reconciliationJob := NewReconciliationJob(gormDB, aaaService, logger, 5*time.Minute)

	return &ServiceFactory{
		FarmerService:         farmerService,
		FarmerLinkageService:  farmerLinkageService,
		FPOService:            fpoService,
		FPOLifecycleService:   fpoLifecycleService,
		FPOConfigService:      fpoConfigService,
		KisanSathiService:     kisanSathiService,
		FarmService:           farmService,
		CropService:           cropService,
		CropCycleService:      cropCycleService,
		FarmActivityService:   farmActivityService,
		DataQualityService:    dataQualityService,
		LookupService:         lookupService,
		ReportingService:      reportingService,
		AdministrativeService: administrativeService,
		BulkFarmerService:     bulkFarmerService,
		AAAService:            aaaService,
		AuditService:          auditService,
		AAAClient:             aaaClient,
		StageService:          stageService,
		ReconciliationJob:     reconciliationJob,
	}
}
