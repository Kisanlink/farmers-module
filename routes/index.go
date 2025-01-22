package routes

import (
	"github.com/Kisanlink/farmers-module/controllers"
	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

// Dependencies holds all the controllers required for route handling.
type Dependencies struct {
	FarmerController *controllers.FarmerController
}

// Setup initializes the application routes and dependencies.
func Setup() *gin.Engine {
	// Step 1: Initialize the database connection
	database.InitializeDatabase()
	db := database.GetDatabase()

	// Step 2: Initialize repositories
	farmerRepo := repositories.NewFarmerRepository(db)

	// Step 3: Initialize controllers
	farmerController := controllers.NewFarmerController(farmerRepo)

	// Step 4: Setup dependencies
	deps := &Dependencies{
		FarmerController: farmerController,
	}

	// Step 5: Setup router and routes
	router := gin.Default()
	InitializeRoutes(router, deps)

	return router
}

// InitializeRoutes registers all the routes for the application.
func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	// Version 1 API group
	v1 := router.Group("/api/v1")

	// Farmer routes
	farmer := v1.Group("/farmers")
	{
		farmer.GET("/:id", deps.FarmerController.GetFarmerPersonalDetailsByID)
		// Additional routes (POST, PUT, DELETE) can be added here
	}

	// Product routes
	// v1.GET("/products/:id", deps.ProductController.GetProductByID)
	// v1.POST("/products", deps.ProductController.CreateProduct)
}
