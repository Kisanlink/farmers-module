package routes

import (
	"github.com/Kisanlink/farmers-module/controllers"
	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	FarmerController         *controllers.FarmerController
	FarmController           *controllers.FarmController
	OrderController          *controllers.OrderController
	CommodityPriceController *controllers.CommodityPriceController
	SoilTestReportController *controllers.SoilTestReportController
}

func Setup() *gin.Engine {
	database.InitializeDatabase()
	db := database.GetDatabase()

	// Initialize repositories
	farmerRepo := repositories.NewFarmerRepository(db)
	farmRepo := repositories.NewFarmRepository(db)
	commodityRepo := repositories.NewCommodityPriceRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	soilTestRepo := repositories.NewSoilTestReportRepository(db)

	// Initialize controllers
	farmerController := controllers.NewFarmerController(farmerRepo)
	farmController := controllers.NewFarmController(farmRepo) // No price or soil test info in farm
	orderController := controllers.NewOrderController(orderRepo)
	commodityPriceController := controllers.NewCommodityPriceController(commodityRepo)
	soilTestReportController := controllers.NewSoilTestReportController(soilTestRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmerController:         farmerController,
		FarmController:           farmController,
		OrderController:          orderController,
		CommodityPriceController: commodityPriceController,
		SoilTestReportController: soilTestReportController,
	}

	// Setup router and routes
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204) // No content response for preflight requests
	})
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

	// Farm routes (With query parameters for fields)
	farms := v1.Group("/farms")
	{
		farms.GET("", deps.FarmController.GetFarmsByFarmerID) // Modified to use query params
		farms.GET("soil-test", deps.SoilTestReportController.GetSoilTestReports)
	}

	// Order routes (Optimized: Using query parameters)
	orders := v1.Group("/orders")
	{
		// Fetch orders using query params like ?farmerId=123&status=Delivered
		orders.GET("", deps.OrderController.GetOrdersByFarmerID)
	}

	commodity := v1.Group("/commodity")
	{
		commodity.GET("/prices/farmer/:farmerId", deps.CommodityPriceController.GetCommodityPricesByFarmerID)
	}

	// here commodity
	// Credit Order route (Updated to use query parameters)
	credits := v1.Group("/credits")
	{
		credits.GET("", deps.OrderController.GetCreditOrdersByFarmerID) // Use query parameters
	}
}
