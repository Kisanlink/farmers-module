package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterIdentityRoutes registers routes for Identity & Organization Linkage workflows
func RegisterIdentityRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config) {
	// Initialize validation middleware
	validationMiddleware := middleware.NewValidationMiddleware(services.AAAClient, cfg)

	identity := router.Group("/identity")
	{
		// Farmer management endpoints
		farmers := identity.Group("/farmers")
		{
			// Create a new farmer with validation
			farmers.POST("", validationMiddleware.ValidateFarmerCreation(), handlers.CreateFarmer(services.FarmerService))

			// List farmers with filtering and pagination
			farmers.GET("", handlers.ListFarmers(services.FarmerService))

			// Get farmer by ID
			farmers.GET("/:aaa_user_id/:aaa_org_id", handlers.GetFarmer(services.FarmerService))

			// Update farmer with validation
			farmers.PUT("/:aaa_user_id/:aaa_org_id", validationMiddleware.ValidateFarmerCreation(), handlers.UpdateFarmer(services.FarmerService))

			// Delete farmer
			farmers.DELETE("/:aaa_user_id/:aaa_org_id", handlers.DeleteFarmer(services.FarmerService))
		}

		// W1: Link farmer to FPO
		identity.POST("/farmer/link", handlers.LinkFarmerToFPO(services.FarmerLinkageService))

		// W2: Unlink farmer from FPO
		identity.DELETE("/farmer/unlink", handlers.UnlinkFarmerFromFPO(services.FarmerLinkageService))

		// Get farmer linkage status
		identity.GET("/farmer/linkage/:farmer_id/:org_id", handlers.GetFarmerLinkage(services.FarmerLinkageService))

		// W3: Register FPO reference with validation
		identity.POST("/fpo/register", validationMiddleware.ValidateFPOCreation(), handlers.RegisterFPORef(services.FPORefService))

		// Get FPO reference with organization access validation
		identity.GET("/fpo/:org_id", validationMiddleware.ValidateOrganizationAccess("org_id"), handlers.GetFPORef(services.FPORefService))
	}
}
