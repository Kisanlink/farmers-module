package routes

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/middleware"
	"github.com/Kisanlink/farmers-module/internal/services"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-gonic/gin"
)

// RegisterAllRoutes registers all workflow-based routes
func RegisterAllRoutes(router *gin.Engine, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// API v1 group
	api := router.Group("/api/v1")
	{
		// Identity & Organization Linkage (W1-W3)
		RegisterIdentityRoutes(api, services, cfg, logger)

		// KisanSathi Assignment (W4-W5)
		RegisterKisanSathiRoutes(api, services, cfg, logger)

		// Farm Management (W6-W9)
		RegisterFarmRoutes(api, services, cfg, logger)

		// Crop Management (W10-W17)
		RegisterCropRoutes(api, services, cfg, logger)

		// Data Quality and Validation
		RegisterDataQualityRoutes(api, services, cfg, logger)

		// Lookup Data (Master Data)
		RegisterLookupRoutes(api, services, cfg, logger)

		// Bulk Operations
		RegisterBulkOperationsRoutes(api, services, cfg, logger)

		// Admin & Access Control (W18-W19)
		RegisterAdminRoutes(api, services, cfg, logger)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "farmers-module"})
	})
}

// SetupRoutes sets up all routes with proper handlers and middleware
func SetupRoutes(router *gin.Engine, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	// Add core middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add audit middleware for all routes
	if services.AuditService != nil {
		auditMW := middleware.AuditMiddleware(logger)
		router.Use(auditMW)
	}

	// Add request ID middleware
	router.Use(middleware.RequestIDMiddleware())

	// Register all routes
	RegisterAllRoutes(router, services, cfg, logger)

	// Add Scalar-powered Swagger documentation route
	router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:  specURL,
			DarkMode: false,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Farmers Module API Reference",
			},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to render API docs: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// Add Swagger JSON specification route
	router.GET("/docs/swagger.json", func(c *gin.Context) {
		c.File("docs/swagger.json")
	})

	// Add root route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Farmers Module Server - Workflow-Based Architecture",
			"version": "1.0.0",
			"docs":    "/docs",
		})
	})
}
