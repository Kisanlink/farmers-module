package routes

import (
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterAllRoutes registers all workflow-based routes
func RegisterAllRoutes(router *gin.Engine, services *services.ServiceFactory) {
	// API v1 group
	api := router.Group("/api/v1")
	{
		// Identity & Organization Linkage (W1-W3)
		RegisterIdentityRoutes(api, services)

		// KisanSathi Assignment (W4-W5)
		RegisterKisanSathiRoutes(api, services)

		// Farm Management (W6-W9)
		RegisterFarmRoutes(api, services)

		// Crop Management (W10-W17)
		RegisterCropRoutes(api, services)

		// Admin & Access Control (W18-W19)
		RegisterAdminRoutes(api, services)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
