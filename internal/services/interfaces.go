package services

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
)

// FarmerLinkageService handles farmer-to-FPO linkage workflows
type FarmerLinkageService interface {
	// W1: Link farmer to FPO
	LinkFarmerToFPO(ctx context.Context, req interface{}) error
	// W2: Unlink farmer from FPO
	UnlinkFarmerFromFPO(ctx context.Context, req interface{}) error
	// Get farmer linkage status
	GetFarmerLinkage(ctx context.Context, farmerID, orgID string) (interface{}, error)
	// W4: Assign KisanSathi to farmer
	AssignKisanSathi(ctx context.Context, req interface{}) (interface{}, error)
	// W5: Reassign or remove KisanSathi
	ReassignOrRemoveKisanSathi(ctx context.Context, req interface{}) (interface{}, error)
	// Create KisanSathi user with role assignment
	CreateKisanSathiUser(ctx context.Context, req interface{}) (interface{}, error)
	// List KisanSathis (users assigned to farmers)
	ListKisanSathis(ctx context.Context, req interface{}) (interface{}, error)
}

// FPOService handles FPO creation and management workflows
type FPOService interface {
	// CreateFPO creates an FPO organization with AAA integration
	CreateFPO(ctx context.Context, req interface{}) (interface{}, error)
	// RegisterFPORef registers FPO reference for local management
	RegisterFPORef(ctx context.Context, req interface{}) (interface{}, error)
	// GetFPORef gets FPO reference
	GetFPORef(ctx context.Context, orgID string) (interface{}, error)
	// CompleteFPOSetup retries failed setup operations for PENDING_SETUP FPOs
	CompleteFPOSetup(ctx context.Context, orgID string) (interface{}, error)
}

// KisanSathiService handles KisanSathi assignment workflows
type KisanSathiService interface {
	// W4: Assign KisanSathi to farmer
	AssignKisanSathi(ctx context.Context, req interface{}) error
	// W5: Reassign or remove KisanSathi
	ReassignOrRemoveKisanSathi(ctx context.Context, req interface{}) error
	// Get KisanSathi assignment
	GetKisanSathiAssignment(ctx context.Context, farmerID string) (interface{}, error)
}

// FarmService handles farm management workflows
type FarmService interface {
	// W6: Create farm
	CreateFarm(ctx context.Context, req interface{}) (interface{}, error)
	// W7: Update farm
	UpdateFarm(ctx context.Context, req interface{}) (interface{}, error)
	// W8: Delete farm
	DeleteFarm(ctx context.Context, req interface{}) error
	// W9: List farms
	ListFarms(ctx context.Context, req interface{}) (interface{}, error)
	// Get farm by ID
	GetFarm(ctx context.Context, farmID string) (interface{}, error)
}

// CropCycleService handles crop cycle workflows
type CropCycleService interface {
	// W10: Start crop cycle
	StartCycle(ctx context.Context, req interface{}) (interface{}, error)
	// W11: Update crop cycle
	UpdateCycle(ctx context.Context, req interface{}) (interface{}, error)
	// W12: End crop cycle
	EndCycle(ctx context.Context, req interface{}) (interface{}, error)
	// W13: List crop cycles
	ListCycles(ctx context.Context, req interface{}) (interface{}, error)
	// Get crop cycle by ID
	GetCropCycle(ctx context.Context, cycleID string) (interface{}, error)
	// Get area allocation summary for a farm
	GetAreaAllocationSummary(ctx context.Context, farmID string) (interface{}, error)
}

// FarmActivityService handles farm activity workflows
type FarmActivityService interface {
	// W14: Create farm activity
	CreateActivity(ctx context.Context, req interface{}) (interface{}, error)
	// W15: Complete farm activity
	CompleteActivity(ctx context.Context, req interface{}) (interface{}, error)
	// W16: Update farm activity
	UpdateActivity(ctx context.Context, req interface{}) (interface{}, error)
	// W17: List farm activities
	ListActivities(ctx context.Context, req interface{}) (interface{}, error)
	// Get farm activity by ID
	GetFarmActivity(ctx context.Context, activityID string) (interface{}, error)
	// Get stage-wise progress for a crop cycle
	GetStageProgress(ctx context.Context, cropCycleID string) (interface{}, error)
}

// CropService handles crop master data operations
type CropService interface {
	// CRUD operations for crops
	CreateCrop(ctx context.Context, req interface{}) (interface{}, error)
	GetCrop(ctx context.Context, req interface{}) (interface{}, error)
	UpdateCrop(ctx context.Context, req interface{}) (interface{}, error)
	DeleteCrop(ctx context.Context, req interface{}) (interface{}, error)
	ListCrops(ctx context.Context, req interface{}) (interface{}, error)

	// CRUD operations for crop varieties
	CreateCropVariety(ctx context.Context, req interface{}) (interface{}, error)
	GetCropVariety(ctx context.Context, req interface{}) (interface{}, error)
	UpdateCropVariety(ctx context.Context, req interface{}) (interface{}, error)
	DeleteCropVariety(ctx context.Context, req interface{}) (interface{}, error)
	ListCropVarieties(ctx context.Context, req interface{}) (interface{}, error)

	// Lookup/dropdown data
	GetCropLookupData(ctx context.Context, req interface{}) (interface{}, error)
	GetVarietyLookupData(ctx context.Context, req interface{}) (interface{}, error)
	GetCropCategories(ctx context.Context) (interface{}, error)
	GetCropSeasons(ctx context.Context) (interface{}, error)

	// Seed operations
	SeedInitialCropData(ctx context.Context) error
}

// AAAService handles AAA integration workflows
type AAAService interface {
	// W18: Seed roles and permissions
	SeedRolesAndPermissions(ctx context.Context, force bool) error
	// W19: Check permission (for RPC interceptor)
	CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error)

	// User Management
	CreateUser(ctx context.Context, req interface{}) (interface{}, error)
	GetUser(ctx context.Context, userID string) (interface{}, error)
	GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error)
	GetUserByEmail(ctx context.Context, email string) (interface{}, error)

	// Organization Management
	CreateOrganization(ctx context.Context, req interface{}) (interface{}, error)
	GetOrganization(ctx context.Context, orgID string) (interface{}, error)

	// User Group Management
	CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error)
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error

	// Role and Permission Management
	AssignRole(ctx context.Context, userID, orgID, roleName string) error
	CheckUserRole(ctx context.Context, userID, roleName string) (bool, error)
	AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error

	// Token and Health Management
	ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error)
	HealthCheck(ctx context.Context) error
}

// DataQualityService handles data quality and validation workflows
type DataQualityService interface {
	// ValidateGeometry validates WKT geometry with PostGIS validation and SRID checks
	ValidateGeometry(ctx context.Context, req interface{}) (interface{}, error)

	// ReconcileAAALinks heals broken AAA references in farmer_links
	ReconcileAAALinks(ctx context.Context, req interface{}) (interface{}, error)

	// RebuildSpatialIndexes rebuilds GIST indexes for database maintenance
	RebuildSpatialIndexes(ctx context.Context, req interface{}) (interface{}, error)

	// DetectFarmOverlaps detects spatial intersections between farm boundaries
	DetectFarmOverlaps(ctx context.Context, req interface{}) (interface{}, error)
}

// ReportingService handles reporting and analytics workflows
type ReportingService interface {
	// ExportFarmerPortfolio aggregates farms, cycles, and activities data for a farmer
	ExportFarmerPortfolio(ctx context.Context, req interface{}) (interface{}, error)

	// OrgDashboardCounters provides org-level KPIs including counts and areas by season/status
	OrgDashboardCounters(ctx context.Context, req interface{}) (interface{}, error)
}

// AdministrativeService handles administrative and system management workflows
type AdministrativeService interface {
	// SeedRolesAndPermissions triggers a complete reseed of AAA resources, actions, and role bindings
	SeedRolesAndPermissions(ctx context.Context, req interface{}) (interface{}, error)

	// SeedLookupData seeds master lookup data (soil types, irrigation sources, etc.)
	SeedLookupData(ctx context.Context, req interface{}) (interface{}, error)

	// HealthCheck verifies database connectivity and AAA service availability
	HealthCheck(ctx context.Context, req interface{}) (interface{}, error)
}

// StageService handles stage-related operations
type StageService interface {
	// Stage CRUD
	CreateStage(ctx context.Context, req interface{}) (interface{}, error)
	GetStage(ctx context.Context, req interface{}) (interface{}, error)
	UpdateStage(ctx context.Context, req interface{}) (interface{}, error)
	DeleteStage(ctx context.Context, req interface{}) (interface{}, error)
	ListStages(ctx context.Context, req interface{}) (interface{}, error)

	// CropStage operations
	AssignStageToCrop(ctx context.Context, req interface{}) (interface{}, error)
	RemoveStageFromCrop(ctx context.Context, req interface{}) (interface{}, error)
	UpdateCropStage(ctx context.Context, req interface{}) (interface{}, error)
	GetCropStages(ctx context.Context, req interface{}) (interface{}, error)
	ReorderCropStages(ctx context.Context, req interface{}) (interface{}, error)

	// Lookup operations
	GetStageLookup(ctx context.Context, req interface{}) (interface{}, error)
}
