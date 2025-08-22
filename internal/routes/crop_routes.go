package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterCropRoutes registers routes for Crop Management workflows
func RegisterCropRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config) {
	crops := router.Group("/crops")
	{
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
}
