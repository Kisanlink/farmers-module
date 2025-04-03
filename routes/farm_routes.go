package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
)

func RegisterFarmRoutes(router *gin.RouterGroup, farmService services.FarmServiceInterface, userService services.UserServiceInterface) {
    farmHandler := handlers.NewFarmHandler(farmService, userService)
    
    router.POST("/farms", farmHandler.CreateFarmHandler)
    // Add GET endpoint
    router.GET("/farms", farmHandler.GetFarmsHandler)
    router.GET("/farms/:id", farmHandler.GetFarmByIDHandler)
}