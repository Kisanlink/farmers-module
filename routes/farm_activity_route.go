package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterFarmActivityRoutes(router *gin.RouterGroup, activityService services.FarmActivityServiceInterface) {
	handler := handlers.NewFarmActivityHandler(activityService)

	activityRoutes := router.Group("/farm-activities")
	{
		activityRoutes.POST("", handler.CreateActivity)
		activityRoutes.GET("", handler.GetActivities)
		activityRoutes.PUT("/:id", handler.UpdateActivity)
		activityRoutes.DELETE("/:id", handler.DeleteActivity)
	}

	// Batch routes
	batchGroup := router.Group("/batch")
	{
		// Route for getting farm activities for multiple farms
		batchGroup.POST("/farm-activities", handler.GetBatchFarmActivities)
	}
}
