package routes

import (
	"log"

	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// RegisterFarmerRoutes registers routes related to farmers
func RegisterFarmerRoutes(router *gin.RouterGroup, farmerService services.FarmerServiceInterface) {
	log.Println("Inside RegisterFarmerRoutes") // ✅ Log when function is called

	router.POST("/farmers", handlers.FarmerSignupHandler(farmerService))
}
