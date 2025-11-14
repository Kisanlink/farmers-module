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

	// FPO Config routes group
	fpoConfigGroup := api.Group("/fpo-config")
	fpoConfigGroup.Use(authenticationMW, authorizationMW) // Apply auth to all FPO config routes
	{
		// GET /api/v1/fpo-config/:fpo_id - Get FPO config
		fpoConfigGroup.GET("/:fpo_id", handler.GetFPOConfig)

		// GET /api/v1/fpo-config/:fpo_id/health - Check ERP health
		fpoConfigGroup.GET("/:fpo_id/health", handler.CheckERPHealth)

		// GET /api/v1/fpo-config - List all FPO configs (admin only)
		fpoConfigGroup.GET("", handler.ListFPOConfigs)

		// POST /api/v1/fpo-config - Create FPO config (admin only)
		fpoConfigGroup.POST("", handler.CreateFPOConfig)

		// PUT /api/v1/fpo-config/:fpo_id - Update FPO config (admin only)
		fpoConfigGroup.PUT("/:fpo_id", handler.UpdateFPOConfig)

		// DELETE /api/v1/fpo-config/:fpo_id - Delete FPO config (admin only)
		fpoConfigGroup.DELETE("/:fpo_id", handler.DeleteFPOConfig)
	}
}
