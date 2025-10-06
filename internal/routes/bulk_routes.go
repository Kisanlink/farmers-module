package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterBulkOperationsRoutes registers bulk operations routes
// Provides endpoints for bulk farmer operations and file processing
func RegisterBulkOperationsRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize bulk farmer handler with AAA service for permission checks
	bulkFarmerHandler := handlers.NewBulkFarmerHandler(services.BulkFarmerService, services.AAAService, logger)

	// Register bulk operation routes
	bulkFarmerHandler.RegisterRoutes(router)
}
