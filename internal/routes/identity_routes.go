package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterIdentityRoutes registers routes for Identity & Organization Linkage workflows
func RegisterIdentityRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize validation middleware
	validationMiddleware := middleware.NewValidationMiddleware(services.AAAClient, cfg)

	identity := router.Group("/identity")
	{
		// Farmer management endpoints
		farmers := identity.Group("/farmers")
		{
			// Create a new farmer with validation
			farmers.POST("", validationMiddleware.ValidateFarmerCreation(), handlers.CreateFarmer(services.FarmerService, logger))

			// List farmers with filtering and pagination
			farmers.GET("", handlers.ListFarmers(services.FarmerService, logger))

			// Get farmer by farmer ID (primary key)
			farmers.GET("/id/:farmer_id", handlers.GetFarmerByID(services.FarmerService, logger))

			// Get farmer by user ID only (no org required)
			farmers.GET("/user/:aaa_user_id", handlers.GetFarmerByUserID(services.FarmerService, logger))

			// Get farmer by user ID and org ID (legacy endpoint)
			farmers.GET("/:aaa_user_id/:aaa_org_id", handlers.GetFarmer(services.FarmerService, logger))

			// Update farmer by farmer ID (primary key)
			farmers.PUT("/id/:farmer_id", validationMiddleware.ValidateFarmerCreation(), handlers.UpdateFarmerByID(services.FarmerService, logger))

			// Update farmer by user ID only (no org required)
			farmers.PUT("/user/:aaa_user_id", validationMiddleware.ValidateFarmerCreation(), handlers.UpdateFarmerByUserID(services.FarmerService, logger))

			// Update farmer by user ID and org ID (legacy endpoint)
			farmers.PUT("/:aaa_user_id/:aaa_org_id", validationMiddleware.ValidateFarmerCreation(), handlers.UpdateFarmer(services.FarmerService, logger))

			// Delete farmer by farmer ID (primary key)
			farmers.DELETE("/id/:farmer_id", handlers.DeleteFarmerByID(services.FarmerService, logger))

			// Delete farmer by user ID only (no org required)
			farmers.DELETE("/user/:aaa_user_id", handlers.DeleteFarmerByUserID(services.FarmerService, logger))

			// Delete farmer by user ID and org ID (legacy endpoint)
			farmers.DELETE("/:aaa_user_id/:aaa_org_id", handlers.DeleteFarmer(services.FarmerService, logger))
		}

		// W1: Link farmer to FPO
		identity.POST("/farmer/link", handlers.LinkFarmerToFPO(services.FarmerLinkageService))

		// W2: Unlink farmer from FPO
		identity.DELETE("/farmer/unlink", handlers.UnlinkFarmerFromFPO(services.FarmerLinkageService))

		// Get farmer linkage status
		identity.GET("/farmer/linkage/:farmer_id/:org_id", handlers.GetFarmerLinkage(services.FarmerLinkageService))

		// KisanSathi management endpoints
		kisanSathi := identity.Group("/kisansathi")
		{
			// W4: Assign KisanSathi to farmer
			kisanSathi.POST("/assign", handlers.AssignKisanSathi(services.FarmerLinkageService, logger))

			// W5: Reassign or remove KisanSathi
			kisanSathi.PUT("/reassign", handlers.ReassignOrRemoveKisanSathi(services.FarmerLinkageService, logger))

			// Create new KisanSathi user with role assignment
			kisanSathi.POST("/create-user", handlers.CreateKisanSathiUser(services.FarmerLinkageService, logger))

			// Get KisanSathi assignment for farmer
			kisanSathi.GET("/assignment/:farmer_id/:org_id", handlers.GetKisanSathiAssignment(services.FarmerLinkageService, logger))
		}

		// FPO management endpoints
		fpo := identity.Group("/fpo")
		{
			// Create FPO handler
			fpoHandler := handlers.NewFPOHandler(services.FPOService, logger)

			// Create FPO organization with AAA integration
			fpo.POST("/create", fpoHandler.CreateFPO)

			// Register FPO reference with validation
			fpo.POST("/register", fpoHandler.RegisterFPORef)

			// Get FPO reference with organization access validation
			fpo.GET("/reference/:aaa_org_id", fpoHandler.GetFPORef)
		}
	}
}
