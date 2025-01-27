package routes

import (
	"github.com/Kisanlink/farmers-module/controllers"
	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

// Dependencies holds all the controllers required for route handling.
type Dependencies struct {
	FarmerController    *controllers.FarmerController
	FarmController      *controllers.FarmController
	OrderController     *controllers.OrderController
}

// Setup initializes the application routes and dependencies.
func Setup() *gin.Engine {
	// Step 1: Initialize the database connection
	database.InitializeDatabase()
	db := database.GetDatabase()

	// Step 2: Initialize repositories
	farmerRepo := repositories.NewFarmerRepository(db)
	farmRepo := repositories.NewFarmRepository(db)
	commodityRepo := repositories.NewCommodityPriceRepository(db) // Initialize the CommodityPriceRepository
	orderRepo := repositories.NewOrderRepository(db) // Initialize the OrderRepository

	// Step 3: Initialize controllers
	farmerController := controllers.NewFarmerController(farmerRepo)
	farmController := controllers.NewFarmController(farmRepo, commodityRepo)
	orderController := controllers.NewOrderController(orderRepo) // Initialize the OrderController

	// Step 4: Setup dependencies
	deps := &Dependencies{
		FarmerController: farmerController,
		FarmController:   farmController,
		OrderController:  orderController, // Add OrderController
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

	// Farm routes
	farms := v1.Group("/farms")
	{
		// Route to get all farms for a specific farmer by farmerID
		farms.GET("/farmer/:farmerId", deps.FarmController.GetFarmsByFarmerID)
		// Additional routes (POST, PUT, DELETE) for farms can be added here
	}

	// Order routes
	orders := v1.Group("/orders")
	{
		// Route to get all orders for a specific farmer by farmerID
		orders.GET("/farmer/:farmerId", deps.OrderController.GetOrdersByFarmerID)
		// Additional routes (POST, PUT, DELETE) for orders can be added here
	}
}

