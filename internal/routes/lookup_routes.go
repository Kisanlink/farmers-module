package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterLookupRoutes registers routes for Lookup Data workflows
func RegisterLookupRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize lookup handlers
	lookupHandlers := handlers.NewLookupHandlers(services.LookupService)

	// Lookup routes - no auth required as these are master data
	lookups := router.Group("/lookups")
	{
		// Get all soil types
		lookups.GET("/soil-types", lookupHandlers.GetSoilTypes)

		// Get all irrigation sources
		lookups.GET("/irrigation-sources", lookupHandlers.GetIrrigationSources)
	}
}
