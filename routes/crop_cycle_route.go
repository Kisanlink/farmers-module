package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// RegisterCropCycleRoutes sets up the routes for crop cycle operations.
func RegisterCropCycleRoutes(router *gin.RouterGroup, svc services.CropCycleServiceInterface) {
	// Initialize the handler with the given service
	h := handlers.NewCropCycleHandler(svc)

	// Route for creating a new crop cycle
	router.POST("/farms/:farmId/crop-cycles", h.CreateCropCycle)

	// Group of routes for farm-based crop cycles
	farmGroup := router.Group("/farms/:farmId")
	{
		// Route for fetching crop cycles by farm
		farmGroup.GET("/crop-cycles", h.GetCropCycles)

		// Route for updating a specific crop cycle by farm ID and cycle ID
		farmGroup.PUT("/crop-cycles/:cycleId", h.UpdateCropCycle)
	}
}
