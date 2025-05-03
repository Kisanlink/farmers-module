package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
)

func RegisterCropRoutes(router *gin.RouterGroup, cropService services.CropServiceInterface) {
	handler := handlers.NewCropHandler(cropService)

	cropRoutes := router.Group("/crops")
	{
		cropRoutes.GET("", handler.GetAllCrops)
		cropRoutes.POST("", handler.CreateCrop)
		cropRoutes.GET("/:id", handler.GetCropById)
		cropRoutes.PUT("/:id", handler.UpdateCrop)
		cropRoutes.DELETE("/:id", handler.DeleteCrop)
	}
}
