package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterDataQualityRoutes registers data quality and validation routes
func RegisterDataQualityRoutes(api *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Create data quality handlers
	dataQualityHandlers := handlers.NewDataQualityHandlers(services.DataQualityService)

	// Data quality routes group
	dataQuality := api.Group("/data-quality")
	{
		// Geometry validation
		dataQuality.POST("/validate-geometry", dataQualityHandlers.ValidateGeometry)

		// AAA links reconciliation
		dataQuality.POST("/reconcile-aaa-links", dataQualityHandlers.ReconcileAAALinks)

		// Spatial indexes rebuild
		dataQuality.POST("/rebuild-spatial-indexes", dataQualityHandlers.RebuildSpatialIndexes)

		// Farm overlaps detection
		dataQuality.POST("/detect-farm-overlaps", dataQualityHandlers.DetectFarmOverlaps)
	}
}
