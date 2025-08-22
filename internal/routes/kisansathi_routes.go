package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterKisanSathiRoutes registers routes for KisanSathi Assignment workflows
func RegisterKisanSathiRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	kisansathi := router.Group("/kisansathi")
	{
		// W4: Assign KisanSathi to farmer
		kisansathi.POST("/assign", handlers.AssignKisanSathi(services.FarmerLinkageService, logger))

		// W5: Reassign or remove KisanSathi
		kisansathi.PUT("/reassign", handlers.ReassignOrRemoveKisanSathi(services.FarmerLinkageService, logger))

		// Get KisanSathi assignment
		kisansathi.GET("/assignment/:farmer_id/:org_id", handlers.GetKisanSathiAssignment(services.FarmerLinkageService, logger))

		// Create KisanSathi user
		kisansathi.POST("/create-user", handlers.CreateKisanSathiUser(services.FarmerLinkageService, logger))
	}
}
