package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
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
}
