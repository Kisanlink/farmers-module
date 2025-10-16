package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Kisanlink/farmers-module/docs" // Import Swagger docs
	"github.com/Kisanlink/farmers-module/internal/config"
	farmersDB "github.com/Kisanlink/farmers-module/internal/db"
	"github.com/Kisanlink/farmers-module/internal/repo"
	"github.com/Kisanlink/farmers-module/internal/routes"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/internal/utils"
	kisanlinkDB "github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title Farmers Module API
// @version 1.0.0
// @description Farmers Module Service with Workflow-Based Architecture for Farm Management
// @host localhost:8000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	dbConfig := &kisanlinkDB.Config{
		PostgresHost:     cfg.Database.Host,
		PostgresPort:     cfg.Database.Port,
		PostgresUser:     cfg.Database.User,
		PostgresPassword: cfg.Database.Password,
		PostgresDBName:   cfg.Database.Name,
		PostgresSSLMode:  cfg.Database.SSLMode,
	}
	dbManager := farmersDB.Connect(dbConfig)

	// Setup database (run migrations and create tables)
	if err := farmersDB.SetupDatabase(dbManager); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Initialize repository factory
	repoFactory := repo.NewRepositoryFactory(dbManager)

	// Initialize structured logger
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := zapLogger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Create logger adapter
	logger := utils.NewLoggerAdapter(zapLogger)

	// Initialize service factory
	serviceFactory := services.NewServiceFactory(repoFactory, dbManager, cfg, logger)

	// Initialize router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup all routes with handlers and middleware
	routes.SetupRoutes(router, serviceFactory, cfg, logger)

	// TODO: Re-enable role seeding when AAA service implements CatalogService
	// Currently disabled because CatalogService.SeedRolesAndPermissions is not implemented in AAA service
	// Ensure roles (farmer, kisansathi, CEO, fpo_manager, admin, readonly) exist in AAA service manually
	// Following ADR-001: Role seeding should happen at startup but not block application start
	//
	// log.Println("Seeding AAA roles and permissions...")
	// seedCtx, seedCancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer seedCancel()
	//
	// if err := serviceFactory.AAAService.SeedRolesAndPermissions(seedCtx); err != nil {
	// 	log.Printf("Warning: Failed to seed AAA roles and permissions: %v", err)
	// 	log.Println("Application will continue, but role assignments may fail if roles don't exist")
	// 	log.Println("Use the /admin/seed-roles endpoint to manually trigger role seeding")
	// } else {
	// 	log.Println("Successfully seeded AAA roles and permissions")
	// }
	log.Println("Note: AAA role seeding disabled - ensure roles exist in AAA service (farmer, kisansathi, CEO, fpo_manager, admin, readonly)")

	// Get port from configuration
	port := cfg.Server.Port

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Farmers Module server on :%s", port)
		log.Println("Available workflow groups:")
		log.Println("  - /api/v1/identity     (W1-W3: Identity & Org Linkage)")
		log.Println("  - /api/v1/kisansathi   (W4-W5: KisanSathi Assignment)")
		log.Println("  - /api/v1/farms        (W6-W9: Farm Management)")
		log.Println("  - /api/v1/crops        (W10-W17: Crop Management)")
		log.Println("  - /api/v1/admin        (W18-W19: Access Control)")
		log.Println("  - /docs                (API Documentation)")

		if err := router.Run(":" + port); err != nil {
			log.Fatal("Error starting HTTP server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Close database connection before exit
	if err := dbManager.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server exited")
}
