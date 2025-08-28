package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
)

// AdministrativeServiceWrapper wraps the concrete administrative service to match the interface
type AdministrativeServiceWrapper struct {
	service ConcreteAdministrativeService
}

// NewAdministrativeServiceWrapper creates a new wrapper
func NewAdministrativeServiceWrapper(service ConcreteAdministrativeService) AdministrativeService {
	return &AdministrativeServiceWrapper{
		service: service,
	}
}

// SeedRolesAndPermissions implements the interface method
func (w *AdministrativeServiceWrapper) SeedRolesAndPermissions(ctx context.Context, req interface{}) (interface{}, error) {
	// Convert interface{} to proper request type
	var seedReq *requests.SeedRolesAndPermissionsRequest

	if req == nil {
		seedReq = &requests.SeedRolesAndPermissionsRequest{}
	} else if typedReq, ok := req.(*requests.SeedRolesAndPermissionsRequest); ok {
		seedReq = typedReq
	} else if mapReq, ok := req.(map[string]interface{}); ok {
		seedReq = &requests.SeedRolesAndPermissionsRequest{}
		if force, exists := mapReq["force"]; exists {
			if forceBool, ok := force.(bool); ok {
				seedReq.Force = forceBool
			}
		}
		if dryRun, exists := mapReq["dry_run"]; exists {
			if dryRunBool, ok := dryRun.(bool); ok {
				seedReq.DryRun = dryRunBool
			}
		}
	} else {
		return nil, fmt.Errorf("invalid request type for SeedRolesAndPermissions")
	}

	return w.service.SeedRolesAndPermissions(ctx, seedReq)
}

// HealthCheck implements the interface method
func (w *AdministrativeServiceWrapper) HealthCheck(ctx context.Context, req interface{}) (interface{}, error) {
	// Convert interface{} to proper request type
	var healthReq *requests.HealthCheckRequest

	if req == nil {
		healthReq = &requests.HealthCheckRequest{}
	} else if typedReq, ok := req.(*requests.HealthCheckRequest); ok {
		healthReq = typedReq
	} else if mapReq, ok := req.(map[string]interface{}); ok {
		healthReq = &requests.HealthCheckRequest{}
		if components, exists := mapReq["components"]; exists {
			if compSlice, ok := components.([]interface{}); ok {
				healthReq.Components = make([]string, len(compSlice))
				for i, comp := range compSlice {
					if compStr, ok := comp.(string); ok {
						healthReq.Components[i] = compStr
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("invalid request type for HealthCheck")
	}

	return w.service.HealthCheck(ctx, healthReq)
}
