package routes

import (
	"github.com/gin-gonic/gin"
)

func InitializeFarmerRoutes(router *gin.RouterGroup, deps *Dependencies) {
	farmer := router.Group("/farmers")
	{
		farmer.GET("", deps.FarmerController.SearchFarmers)                                    // Search farmers for later use
		farmer.GET("/:id", deps.FarmerController.GetFarmerPersonalDetailsByID)                 // Get particular farmer data
		farmer.PUT("/:id", deps.FarmerController.UpdateFarmer)                                 // Update farmer details for later use
		farmer.POST("/", deps.FarmerController.CreateFarmer)                                   // Create a new farmer for later use
		farmer.GET("/:id/farms", deps.FarmController.GetFarmsByFarmerID)                       // Get farms by farmer ID
		farmer.GET("/:id/orders", deps.OrderController.GetOrdersByFarmerID)                    // Get orders by farmer ID
		farmer.GET("/:id/orders/filters", deps.OrderController.GetOrdersByFarmerIDWithFilters) // Get orders by farmer ID with filters
	}
}
