package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterKisanSathiRoutes registers routes for KisanSathi Assignment workflows
func RegisterKisanSathiRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config) {
	kisansathi := router.Group("/kisansathi")
	{
		// W4: Assign KisanSathi to farmer
		kisansathi.POST("/assign", handlers.AssignKisanSathi(services.KisanSathiService))

		// W5: Reassign or remove KisanSathi
		kisansathi.PUT("/reassign", handlers.ReassignOrRemoveKisanSathi(services.KisanSathiService))

		// Get KisanSathi assignment
		kisansathi.GET("/assignment/:farmer_id", handlers.GetKisanSathiAssignment(services.KisanSathiService))
	}
}
