package services

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/repo"
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
	KisanSathiService    KisanSathiService

	// Farm Management Services
	FarmService FarmService

	// Crop Management Services
	CropCycleService    CropCycleService
	FarmActivityService FarmActivityService

	// Data Quality Services
	DataQualityService DataQualityService

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

	// Initialize identity services
	farmerService := NewFarmerService(repoFactory.FarmerRepo, aaaService)
	farmerLinkageService := NewFarmerLinkageService(repoFactory.FarmerLinkageRepo, aaaService)
	fpoService := NewFPOService(repoFactory.FPORefRepo, aaaService)
	kisanSathiService := NewKisanSathiService(repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize farm management services
	// Get GORM DB for farm service
	var gormDB *gorm.DB
	if db, err := postgresManager.GetDB(context.Background(), false); err == nil {
		gormDB = db
	}
	farmService := NewFarmService(repoFactory.FarmRepo, aaaService, gormDB)

	// Initialize crop management services
	cropCycleService := NewCropCycleService(repoFactory.CropCycleRepo, aaaService)
	farmActivityService := NewFarmActivityService(repoFactory.FarmActivityRepo, repoFactory.CropCycleRepo, repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize data quality service
	dataQualityService := NewDataQualityService(gormDB, repoFactory.FarmRepo, repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize reporting service
	reportingService := NewReportingService(repoFactory, gormDB, aaaService)

	// Initialize administrative service
	concreteAdminService := NewAdministrativeService(postgresManager, gormDB, aaaService)
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

	return &ServiceFactory{
		FarmerService:         farmerService,
		FarmerLinkageService:  farmerLinkageService,
		FPOService:            fpoService,
		KisanSathiService:     kisanSathiService,
		FarmService:           farmService,
		CropCycleService:      cropCycleService,
		FarmActivityService:   farmActivityService,
		DataQualityService:    dataQualityService,
		ReportingService:      reportingService,
		AdministrativeService: administrativeService,
		BulkFarmerService:     bulkFarmerService,
		AAAService:            aaaService,
		AuditService:          auditService,
		AAAClient:             aaaClient,
	}
}
