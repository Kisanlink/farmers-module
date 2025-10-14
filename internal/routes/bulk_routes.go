package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/middleware"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterBulkOperationsRoutes registers bulk operations routes
// Provides endpoints for bulk farmer operations and file processing
func RegisterBulkOperationsRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Initialize authentication and authorization middleware
	authenticationMW := middleware.AuthenticationMiddleware(services.AAAService, logger)
	authorizationMW := middleware.AuthorizationMiddleware(services.AAAService, logger)

	// Initialize bulk farmer handler with AAA service for permission checks
	bulkFarmerHandler := handlers.NewBulkFarmerHandler(services.BulkFarmerService, services.AAAService, logger)

	// Create bulk routes group with authentication and authorization
	bulk := router.Group("/bulk")
	bulk.Use(authenticationMW, authorizationMW) // Apply auth middleware to all bulk routes
	{
		// Farmer operations
		bulk.POST("/farmers/add", bulkFarmerHandler.BulkAddFarmers)

		// Operation management
		bulk.GET("/status/:operation_id", bulkFarmerHandler.GetBulkOperationStatus)
		bulk.POST("/cancel/:operation_id", bulkFarmerHandler.CancelBulkOperation)
		bulk.POST("/retry/:operation_id", bulkFarmerHandler.RetryFailedRecords)

		// Results and templates
		bulk.GET("/results/:operation_id", bulkFarmerHandler.DownloadBulkResults)
		bulk.GET("/template", bulkFarmerHandler.GetBulkUploadTemplate)

		// Validation
		bulk.POST("/validate", bulkFarmerHandler.ValidateBulkData)
	}
}
