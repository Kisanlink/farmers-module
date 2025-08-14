package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterFarmRoutes(router *gin.RouterGroup, farmService services.FarmServiceInterface, userService services.UserServiceInterface) {
	farmHandler := handlers.NewFarmHandler(farmService, userService)

	router.POST("/farms", farmHandler.CreateFarmHandler)
	router.GET("/farms", farmHandler.GetFarmsHandler)
	router.GET("/farms/:farmId", farmHandler.GetFarmByFarmID)
	router.GET("/getFarmCentroids", farmHandler.GetFarmCentroidsHandler)
	router.GET("/getFarmHeatmap", farmHandler.GetFarmHeatmapHandler)
}
