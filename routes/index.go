package routes

import (
	"github.com/Kisanlink/farmers-module/controllers"
	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	FarmerController *controllers.FarmerController
	FarmController   *controllers.FarmController
	OrderController  *controllers.OrderController
}

func Setup() *gin.Engine {
	database.InitializeDatabase()
	db := database.GetDatabase()

	// Initialize repositories
	farmerRepo := repositories.NewFarmerRepository(db)
	farmRepo := repositories.NewFarmRepository(db)
	commodityRepo := repositories.NewCommodityPriceRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	soilTestRepo := repositories.NewSoilTestReportRepository(db) // Still needed but used inside FarmController

	// Initialize controllers
	farmerController := controllers.NewFarmerController(farmerRepo)
	farmController := controllers.NewFarmController(farmRepo, commodityRepo, soilTestRepo) // Inject Soil Test Repo
	orderController := controllers.NewOrderController(orderRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmerController: farmerController,
		FarmController:   farmController,
		OrderController:  orderController,
	}

	// Setup router and routes
	router := gin.Default()
	InitializeRoutes(router, deps)

	return router
}

func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	v1 := router.Group("/api/v1")

	// Farmer routes
	farmer := v1.Group("/farmers")
	{
		farmer.GET("/:id", deps.FarmerController.GetFarmerPersonalDetailsByID)
	}

	// Farm routes (Includes Soil Test Reports)
	farms := v1.Group("/farms")
	{
		farms.GET("/farmer/:farmerId", deps.FarmController.GetFarmsByFarmerID) // Fetches farms WITH soil reports
	}

	// Order routes
	orders := v1.Group("/orders")
	{
		orders.GET("/farmer/:farmerId", deps.OrderController.GetOrdersByFarmerID)
	}
}
