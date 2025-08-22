package services

import (
	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/repo"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
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

	// AAA Integration Service
	AAAService AAAService

	// AAA Client for direct integration
	AAAClient *aaa.Client
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(repoFactory *repo.RepositoryFactory, postgresManager *db.PostgresManager, cfg *config.Config, logger interfaces.Logger) *ServiceFactory {
	// Initialize AAA client first as it's used by other services
	aaaClient, err := aaa.NewClient(cfg)
	if err != nil {
		// Log warning but continue - services will handle nil client gracefully
		// log.Printf("Warning: Failed to create AAA client: %v", err)
	}

	// Initialize AAA service
	aaaService := NewAAAService(cfg)

	// Initialize identity services
	farmerService := NewFarmerService(postgresManager, aaaService)
	farmerLinkageService := NewFarmerLinkageService(repoFactory.FarmerLinkageRepo, aaaService)
	fpoService := NewFPOService(repoFactory.FPORefRepo, aaaService)
	kisanSathiService := NewKisanSathiService(repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize farm management services
	farmService := NewFarmService(repoFactory.FarmRepo, aaaService)

	// Initialize crop management services
	cropCycleService := NewCropCycleService(repoFactory.CropCycleRepo, aaaService)
	farmActivityService := NewFarmActivityService(repoFactory.FarmActivityRepo, aaaService)

	return &ServiceFactory{
		FarmerService:        farmerService,
		FarmerLinkageService: farmerLinkageService,
		FPOService:           fpoService,
		KisanSathiService:    kisanSathiService,
		FarmService:          farmService,
		CropCycleService:     cropCycleService,
		FarmActivityService:  farmActivityService,
		AAAService:           aaaService,
		AAAClient:            aaaClient,
	}
}
