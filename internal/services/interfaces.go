package services

import (
	"context"
)

// FarmerLinkageService handles farmer-to-FPO linkage workflows
type FarmerLinkageService interface {
	// W1: Link farmer to FPO
	LinkFarmerToFPO(ctx context.Context, req interface{}) error
	// W2: Unlink farmer from FPO
	UnlinkFarmerFromFPO(ctx context.Context, req interface{}) error
	// Get farmer linkage status
	GetFarmerLinkage(ctx context.Context, farmerID, orgID string) (interface{}, error)
}

// FPORefService handles FPO reference data workflows
type FPORefService interface {
	// W3: Register FPO reference
	RegisterFPORef(ctx context.Context, req interface{}) error
	// Get FPO reference
	GetFPORef(ctx context.Context, orgID string) (interface{}, error)
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
}

// AAAService handles AAA integration workflows
type AAAService interface {
	// W18: Seed roles and permissions
	SeedRolesAndPermissions(ctx context.Context) error
	// W19: Check permission (for RPC interceptor)
	CheckPermission(ctx context.Context, req interface{}) (bool, error)

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
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
	HealthCheck(ctx context.Context) error
}
