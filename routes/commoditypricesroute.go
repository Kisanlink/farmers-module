package routes

import (
	"github.com/gin-gonic/gin"
)

func InitializeCommodityPriceRoutes(router *gin.RouterGroup, deps *Dependencies) {
 
	// Search routes (Using query parameters)
	search := router.Group("/search")
	{
		// Commodity routes
		search.GET("/commodity/prices", deps.CommodityPriceController.GetAllCommodityPrices) // Fetch all commodity prices
	}

	
	
		farmer:= router.Group("/farmers")
		{
			farmer.GET("/:id/ownedcommodityprices", deps.CommodityPriceController.GetOwnedCommodityPricesByFarmerID) // Get orders by farmer ID
		}
	
}
