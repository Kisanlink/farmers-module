package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterCropCycleRoutes(router *gin.RouterGroup, cropCycleService services.CropCycleServiceInterface) {
	handler := handlers.NewCropCycleHandler(cropCycleService)

	cropCycleRoutes := router.Group("/crop-cycles")
	{
		cropCycleRoutes.POST("", handler.CreateCropCycle)
		cropCycleRoutes.GET("", handler.GetCropCycles)
	}
}
