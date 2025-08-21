package services

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/repo"
)

// ServiceFactory provides access to all domain services
type ServiceFactory struct {
	// Identity & Organization Services
	FarmerLinkageService FarmerLinkageService
	FPORefService        FPORefService
	KisanSathiService    KisanSathiService

	// Farm Management Services
	FarmService FarmService

	// Crop Management Services
	CropCycleService    CropCycleService
	FarmActivityService FarmActivityService

	// AAA Integration Service
	AAAService AAAService
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(repoFactory *repo.RepositoryFactory, cfg *config.Config) *ServiceFactory {
	// Initialize AAA service first as it's used by other services
	aaaService := NewAAAService(cfg)

	// Initialize identity services
	farmerLinkageService := NewFarmerLinkageService(repoFactory.FarmerLinkageRepo, aaaService)
	fpoRefService := NewFPORefService(repoFactory.FPORefRepo, aaaService)
	kisanSathiService := NewKisanSathiService(repoFactory.FarmerLinkageRepo, aaaService)

	// Initialize farm management services
	farmService := NewFarmService(repoFactory.FarmRepo, aaaService)

	// Initialize crop management services
	cropCycleService := NewCropCycleService(repoFactory.CropCycleRepo, aaaService)
	farmActivityService := NewFarmActivityService(repoFactory.FarmActivityRepo, aaaService)

	return &ServiceFactory{
		FarmerLinkageService: farmerLinkageService,
		FPORefService:        fpoRefService,
		KisanSathiService:    kisanSathiService,
		FarmService:          farmService,
		CropCycleService:     cropCycleService,
		FarmActivityService:  farmActivityService,
		AAAService:           aaaService,
	}
}
