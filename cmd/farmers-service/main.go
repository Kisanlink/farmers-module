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
	postgresManager, ok := dbManager.(*kisanlinkDB.PostgresManager)
	if !ok {
		log.Fatal("Failed to cast dbManager to PostgresManager")
	}
	serviceFactory := services.NewServiceFactory(repoFactory, postgresManager, cfg, logger)

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
