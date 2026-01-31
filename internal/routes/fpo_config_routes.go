package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/middleware"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterFPOConfigRoutes registers FPO configuration routes
func RegisterFPOConfigRoutes(api *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Create handler
	handler := handlers.NewFPOConfigHandler(services.FPOConfigService, logger)

	// Create middleware
	authenticationMW := middleware.AuthenticationMiddleware(services.AAAService, logger)
	authorizationMW := middleware.AuthorizationMiddleware(services.AAAService, logger)

	// Self-access routes - require authentication only (no special permissions)
	// These endpoints use the org_id from JWT context, so any authenticated user
	// can access their own organization's data
	meGroup := api.Group("/me")
	meGroup.Use(authenticationMW) // Authentication only, no authorization
	{
		// GET /api/v1/me/organization/configuration - Get user's org FPO config
		meGroup.GET("/organization/configuration", handler.GetMyOrganizationConfig)
	}

	// FPO Config routes - nested under /fpo/:aaa_org_id/configuration
	fpoGroup := api.Group("/fpo")
	fpoGroup.Use(authenticationMW, authorizationMW) // Apply auth to all FPO routes
	{
		// GET /api/v1/fpo/:aaa_org_id/configuration - Get FPO config
		fpoGroup.GET("/:aaa_org_id/configuration", handler.GetFPOConfig)

		// GET /api/v1/fpo/:aaa_org_id/configuration/health - Check ERP health
		fpoGroup.GET("/:aaa_org_id/configuration/health", handler.CheckERPHealth)

		// PUT /api/v1/fpo/:aaa_org_id/configuration - Update FPO config
		fpoGroup.PUT("/:aaa_org_id/configuration", handler.UpdateFPOConfig)

		// DELETE /api/v1/fpo/:aaa_org_id/configuration - Delete FPO config (admin only)
		fpoGroup.DELETE("/:aaa_org_id/configuration", handler.DeleteFPOConfig)
	}

	// Admin routes for FPO config management
	fpoConfigAdminGroup := api.Group("/fpo-config")
	fpoConfigAdminGroup.Use(authenticationMW, authorizationMW)
	{
		// GET /api/v1/fpo-config - List all FPO configs (admin only)
		fpoConfigAdminGroup.GET("", handler.ListFPOConfigs)

		// POST /api/v1/fpo-config - Create FPO config (admin only)
		fpoConfigAdminGroup.POST("", handler.CreateFPOConfig)

		// Legacy routes for backward compatibility
		// GET /api/v1/fpo-config/:aaa_org_id - Get FPO config (legacy)
		fpoConfigAdminGroup.GET("/:aaa_org_id", handler.GetFPOConfig)

		// POST /api/v1/fpo-config/:aaa_org_id - Create FPO config with ID in path
		fpoConfigAdminGroup.POST("/:aaa_org_id", handler.CreateFPOConfigWithID)

		// GET /api/v1/fpo-config/:aaa_org_id/health - Check ERP health (legacy)
		fpoConfigAdminGroup.GET("/:aaa_org_id/health", handler.CheckERPHealth)

		// PUT /api/v1/fpo-config/:aaa_org_id - Update FPO config (legacy)
		fpoConfigAdminGroup.PUT("/:aaa_org_id", handler.UpdateFPOConfig)

		// DELETE /api/v1/fpo-config/:aaa_org_id - Delete FPO config (legacy)
		fpoConfigAdminGroup.DELETE("/:aaa_org_id", handler.DeleteFPOConfig)
	}
}
