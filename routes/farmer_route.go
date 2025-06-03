package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// RegisterFarmerRoutes registers routes related to farmers
func RegisterFarmerRoutes(router *gin.RouterGroup, farmerService services.FarmerServiceInterface) {
	farmerHandler := handlers.NewFarmerHandler(farmerService)

	router.POST("/farmers", farmerHandler.FarmerSignupHandler)
	router.GET("/farmers", farmerHandler.FetchFarmersHandler)

	router.POST("/farmers/:id/subscription", farmerHandler.SubscribeHandler)
}
