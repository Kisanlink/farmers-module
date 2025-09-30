package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/middleware"
	"github.com/Kisanlink/farmers-module/internal/services"
	validationMiddleware "github.com/Kisanlink/farmers-module/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterFarmRoutes registers routes for Farm Management workflows
func RegisterFarmRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize validation middleware
	validation := validationMiddleware.NewValidationMiddleware(services.AAAClient, cfg)

	// Initialize authentication and authorization middleware
	authenticationMW := middleware.AuthenticationMiddleware(services.AAAService, logger)
	authorizationMW := middleware.AuthorizationMiddleware(services.AAAService, logger)

	farms := router.Group("/farms")
	farms.Use(authenticationMW, authorizationMW) // Apply auth middleware to all farm routes
	{
		// W6: Create farm with validation
		farms.POST("/", validation.ValidateFarmCreation(), handlers.CreateFarm(services.FarmService))

		// W7: Update farm with validation
		farms.PUT("/:farm_id", validation.ValidateFarmCreation(), handlers.UpdateFarm(services.FarmService))

		// W8: Delete farm
		farms.DELETE("/:farm_id", handlers.DeleteFarm(services.FarmService))

		// W9: List farms
		farms.GET("/", handlers.ListFarms(services.FarmService))

		// Get farm by ID
		farms.GET("/:farm_id", handlers.GetFarm(services.FarmService))
	}
}
