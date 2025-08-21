package routes

import (
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterIdentityRoutes registers routes for Identity & Organization Linkage workflows
func RegisterIdentityRoutes(router *gin.RouterGroup, services *services.ServiceFactory) {
	identity := router.Group("/identity")
	{
		// W1: Link farmer to FPO
		identity.POST("/farmer/link", handlers.LinkFarmerToFPO(services.FarmerLinkageService))

		// W2: Unlink farmer from FPO
		identity.DELETE("/farmer/unlink", handlers.UnlinkFarmerFromFPO(services.FarmerLinkageService))

		// Get farmer linkage status
		identity.GET("/farmer/linkage/:farmer_id/:org_id", handlers.GetFarmerLinkage(services.FarmerLinkageService))

		// W3: Register FPO reference
		identity.POST("/fpo/register", handlers.RegisterFPORef(services.FPORefService))

		// Get FPO reference
		identity.GET("/fpo/:org_id", handlers.GetFPORef(services.FPORefService))
	}
}
