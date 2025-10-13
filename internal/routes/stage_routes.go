package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/middleware"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterStageRoutes registers routes for Stage Management
func RegisterStageRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize authentication and authorization middleware
	authenticationMW := middleware.AuthenticationMiddleware(services.AAAService, logger)
	authorizationMW := middleware.AuthorizationMiddleware(services.AAAService, logger)

	// Initialize handler
	stageHandler := handlers.NewStageHandler(services.StageService, logger)

	// Stage Master Data (CRUD operations)
	stages := router.Group("/stages")
	stages.Use(authenticationMW, authorizationMW) // Apply auth middleware to all stage routes
	{
		stages.POST("", stageHandler.CreateStage)
		stages.GET("", stageHandler.ListStages)
		stages.GET("/lookup", stageHandler.GetStageLookup) // Lookup endpoint before :id to avoid conflicts
		stages.GET("/:id", stageHandler.GetStage)
		stages.PUT("/:id", stageHandler.UpdateStage)
		stages.DELETE("/:id", stageHandler.DeleteStage)
	}

	// Crop-Stage relationship endpoints
	// Note: These are registered under /crops/:id/stages
	crops := router.Group("/crops")
	crops.Use(authenticationMW, authorizationMW)
	{
		cropStages := crops.Group("/:id/stages")
		{
			cropStages.POST("", stageHandler.AssignStageToCrop)
			cropStages.GET("", stageHandler.GetCropStages)
			cropStages.POST("/reorder", stageHandler.ReorderCropStages) // Reorder endpoint before :stage_id to avoid conflicts
			cropStages.PUT("/:stage_id", stageHandler.UpdateCropStage)
			cropStages.DELETE("/:stage_id", stageHandler.RemoveStageFromCrop)
		}
	}
}
