package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
)

// ✅ Accept FarmServiceInterface instead of FarmService
func RegisterFarmRoutes(router *gin.RouterGroup, farmService services.FarmServiceInterface) {
	router.POST("/farms", handlers.FarmHandler(farmService)) // ✅ Now it accepts the interface
}

