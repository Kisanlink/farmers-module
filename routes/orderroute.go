package routes

import (
	"github.com/gin-gonic/gin"
)

func InitializeOrderRoutes(router *gin.RouterGroup, deps *Dependencies) {
	order := router.Group("/orders")
	{
		order.GET("", deps.OrderController.GetOrdersByFarmerID) // Fetch orders using query params like ?farmerId=123&paymentmode=online&orderstatus=delivered
	}

}
