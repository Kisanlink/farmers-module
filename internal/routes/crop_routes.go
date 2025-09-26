package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterCropRoutes registers routes for Crop Management workflows
func RegisterCropRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Create crop master handler
	cropMasterHandler := handlers.NewCropMasterHandler(services.CropService, logger)

	crops := router.Group("/crops")
	{
		// Crop Master Data (CRUD operations)
		{
			// Create crop
			crops.POST("/", cropMasterHandler.CreateCrop)

			// List crops with filtering
			crops.GET("/", cropMasterHandler.ListCrops)

			// Get crop by ID
			crops.GET("/:id", cropMasterHandler.GetCrop)

			// Update crop
			crops.PUT("/:id", cropMasterHandler.UpdateCrop)

			// Delete crop
			crops.DELETE("/:id", cropMasterHandler.DeleteCrop)
		}

		// Crop Varieties
		varieties := crops.Group("/varieties")
		{
			// Create variety
			varieties.POST("/", cropMasterHandler.CreateVariety)

			// Get variety by ID
			varieties.GET("/:id", cropMasterHandler.GetVariety)

			// Update variety
			varieties.PUT("/:id", cropMasterHandler.UpdateVariety)

			// Delete variety
			varieties.DELETE("/:id", cropMasterHandler.DeleteVariety)
		}

		// List varieties for a specific crop
		crops.GET("/:crop_id/varieties", cropMasterHandler.ListVarieties)

		// Crop Stages
		stages := crops.Group("/stages")
		{
			// Create stage
			stages.POST("/", cropMasterHandler.CreateStage)

			// Get stage by ID
			stages.GET("/:id", cropMasterHandler.GetStage)

			// Update stage
			stages.PUT("/:id", cropMasterHandler.UpdateStage)

			// Delete stage
			stages.DELETE("/:id", cropMasterHandler.DeleteStage)
		}

		// List stages for a specific crop
		crops.GET("/:crop_id/stages", cropMasterHandler.ListStages)

		// Crop Cycles (W10-W13)
		cycles := crops.Group("/cycles")
		{
			// W10: Start crop cycle
			cycles.POST("/", handlers.StartCycle(services.CropCycleService))

			// W11: Update crop cycle
			cycles.PUT("/:cycle_id", handlers.UpdateCycle(services.CropCycleService))

			// W12: End crop cycle
			cycles.PUT("/:cycle_id/end", handlers.EndCycle(services.CropCycleService))

			// W13: List crop cycles
			cycles.GET("/", handlers.ListCycles(services.CropCycleService))

			// Get crop cycle by ID
			cycles.GET("/:cycle_id", handlers.GetCropCycle(services.CropCycleService))

			// Record harvest data
			cycles.PUT("/:cycle_id/harvest", handlers.RecordHarvest(services.CropCycleService))

			// Upload report
			cycles.POST("/:cycle_id/report", handlers.UploadReport(services.CropCycleService))
		}

		// Farm Activities (W14-W17)
		activities := crops.Group("/activities")
		{
			// W14: Create farm activity
			activities.POST("/", handlers.CreateActivity(services.FarmActivityService))

			// W15: Complete farm activity
			activities.PUT("/:activity_id/complete", handlers.CompleteActivity(services.FarmActivityService))

			// W16: Update farm activity
			activities.PUT("/:activity_id", handlers.UpdateActivity(services.FarmActivityService))

			// W17: List farm activities
			activities.GET("/", handlers.ListActivities(services.FarmActivityService))

			// Get farm activity by ID
			activities.GET("/:activity_id", handlers.GetFarmActivity(services.FarmActivityService))
		}
	}

	// Lookup routes
	lookups := router.Group("/lookups")
	{
		// Get crop lookup data (categories, units, seasons)
		lookups.GET("/crop-data", cropMasterHandler.GetLookupData)
	}
}
