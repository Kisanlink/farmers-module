package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
)

func RegisterFarmRoutes(router *gin.RouterGroup, farmService services.FarmServiceInterface, userService services.UserServiceInterface) {
	// Initialize handler with required services only
	farmHandler := handlers.NewFarmHandler(farmService, userService)
	
	// Register farm endpoints
	router.POST("/farms", farmHandler.CreateFarmHandler)
}