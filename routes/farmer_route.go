
package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

// RegisterFarmerRoutes registers routes related to farmers
func RegisterFarmerRoutes(router *gin.RouterGroup, farmerService services.FarmerServiceInterface) {
	router.POST("/farmers", handlers.FarmerSignupHandler(farmerService))
	router.GET("/farmers", handlers.FetchFarmersHandler(farmerService)) // New route for fetching farmersc
}
