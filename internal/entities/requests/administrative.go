package requests

import "time"

// SeedRolesAndPermissionsRequest represents a request to seed roles and permissions
type SeedRolesAndPermissionsRequest struct {
	BaseRequest
	// Force indicates whether to force re-seeding even if already seeded
	Force bool `json:"force,omitempty" example:"false"`
	// DryRun indicates whether to perform a dry run without making changes
	DryRun bool `json:"dry_run,omitempty" example:"false"`
}

// HealthCheckRequest represents a request for health check
type HealthCheckRequest struct {
	BaseRequest
	// Components specifies which components to check (empty means all)
	Components []string `json:"components,omitempty" example:"database,redis,aaa_service"`
	// Timeout specifies the timeout for health checks
	Timeout time.Duration `json:"timeout,omitempty" example:"30s"`
}
