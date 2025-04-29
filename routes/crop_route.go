package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
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
