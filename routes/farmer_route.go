package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
)

// RegisterFarmerRoutes registers routes related to farmers
func RegisterFarmerRoutes(router *gin.RouterGroup, farmerService services.FarmerServiceInterface) {
	farmerHandler := handlers.NewFarmerHandler(farmerService)

	router.POST("/farmers", farmerHandler.FarmerSignupHandler)
	router.GET("/farmers", farmerHandler.FetchFarmersHandler)
}
