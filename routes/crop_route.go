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
		cropRoutes.POST("/create", handler.CreateCrop)
		cropRoutes.GET("/fetch/:id", handler.GetCropByID)
		cropRoutes.PUT("/update/:id", handler.UpdateCrop)
		cropRoutes.DELETE("/delete/:id", handler.DeleteCrop)
	}
}
