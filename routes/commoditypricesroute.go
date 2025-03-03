package routes

import (
	"github.com/gin-gonic/gin"
)

func InitializeCommodityPriceRoutes(router *gin.RouterGroup, deps *Dependencies) {
	crops := router.Group("/crops")
	{
		crops.GET("", deps.CommodityPriceController.GetAllCommodityPrices) // Fetch all commodity prices

		crops.GET("/:id", deps.CommodityPriceController.GetCommodityPriceByID) // Fetch commodity price by crop ID
	}
}
