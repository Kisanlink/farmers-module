package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterFarmRoutes registers routes for Farm Management workflows
func RegisterFarmRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize validation middleware
	validationMiddleware := middleware.NewValidationMiddleware(services.AAAClient, cfg)

	farms := router.Group("/farms")
	{
		// W6: Create farm with validation
		farms.POST("/", validationMiddleware.ValidateFarmCreation(), handlers.CreateFarm(services.FarmService))

		// W7: Update farm with validation
		farms.PUT("/:farm_id", validationMiddleware.ValidateFarmCreation(), handlers.UpdateFarm(services.FarmService))

		// W8: Delete farm
		farms.DELETE("/:farm_id", handlers.DeleteFarm(services.FarmService))

		// W9: List farms
		farms.GET("/", handlers.ListFarms(services.FarmService))

		// Get farm by ID
		farms.GET("/:farm_id", handlers.GetFarm(services.FarmService))
	}
}
