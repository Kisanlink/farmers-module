package routes

import (
	"github.com/Kisanlink/farmers-module/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterFarmerRoutes(router *gin.RouterGroup, farmerController *controllers.FarmerController) {
	farmer := router.Group("/farmer")
	{
		farmer.GET("/:id", farmerController.GetFarmerPersonalDetailsByID)
		// Additional routes (POST, PUT, DELETE) can be added here
	}
}
