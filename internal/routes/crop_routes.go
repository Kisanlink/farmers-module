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
	crops := router.Group("/crops")
	{
		// Crop Master Data (CRUD operations)
		crops.POST("/", handlers.CreateCrop(services.CropService))
		crops.GET("/", handlers.ListCrops(services.CropService))
		crops.GET("/:id", handlers.GetCrop(services.CropService))
		crops.PUT("/:id", handlers.UpdateCrop(services.CropService))
		crops.DELETE("/:id", handlers.DeleteCrop(services.CropService))

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
		}

		// Farm Activities (W14-W17)
		activities := crops.Group("/activities")
		{
			// W14: Create farm activity
			activities.POST("/", handlers.CreateFarmActivity(services.FarmActivityService))

			// W15: Complete farm activity
			activities.PUT("/:activity_id/complete", handlers.CompleteFarmActivity(services.FarmActivityService))

			// W16: Update farm activity
			activities.PUT("/:activity_id", handlers.UpdateFarmActivity(services.FarmActivityService))

			// W17: List farm activities
			activities.GET("/", handlers.ListFarmActivities(services.FarmActivityService))

			// Get farm activity by ID
			activities.GET("/:activity_id", handlers.GetFarmActivity(services.FarmActivityService))
		}
	}

	// Crop Varieties
	varieties := router.Group("/varieties")
	{
		varieties.POST("/", handlers.CreateCropVariety(services.CropService))
		varieties.GET("/", handlers.ListCropVarieties(services.CropService))
		varieties.GET("/:id", handlers.GetCropVariety(services.CropService))
		varieties.PUT("/:id", handlers.UpdateCropVariety(services.CropService))
		varieties.DELETE("/:id", handlers.DeleteCropVariety(services.CropService))
	}

	// Get varieties for a specific crop
	crops.GET("/:crop_id/varieties", handlers.ListCropVarieties(services.CropService))

	// Lookup/Dropdown data
	lookups := router.Group("/lookups")
	{
		lookups.GET("/crops", handlers.GetCropLookupData(services.CropService))
		lookups.GET("/varieties/:crop_id", handlers.GetVarietyLookupData(services.CropService))
		lookups.GET("/crop-categories", handlers.GetCropCategories(services.CropService))
		lookups.GET("/crop-seasons", handlers.GetCropSeasons(services.CropService))
	}
}
