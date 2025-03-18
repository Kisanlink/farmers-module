package routes

import (
	"log"

	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Dependencies struct to hold service dependencies
type Dependencies struct {
	FarmerService services.FarmerServiceInterface
	AAAService    *services.AAAService
}

// Setup initializes the database, services, handlers, and routes
func Setup() *gin.Engine {
  	// Get database instance
	db := database.GetDatabase()

	// Initialize repository
	farmerRepo := repositories.NewFarmerRepository(db)

	// Initialize service
	farmerService := services.NewFarmerService(farmerRepo)

// Initialize AAA service (Replace "AAA_SERVICE_ADDRESS" with actual address)
aaaService := services.NewAAAService("AAA_SERVICE_ADDRESS")


	// Setup dependencies
	deps := &Dependencies{
		FarmerService: farmerService,
		AAAService:    aaaService,
	}
	// Setup router with CORS middleware
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://farmers.kisanlink.in", "https://api.farmers.kisanlink.in"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	// Handle CORS preflight requests
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	// Initialize API routes
	InitializeRoutes(router, deps)

	return router
}

// InitializeRoutes sets up the API routes with handlers
func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	log.Println("Inside InitializeRoutes") // ✅ Log when function is called
	v1 := router.Group("/api/v1")

	RegisterFarmerRoutes(v1, deps.FarmerService, deps.AAAService)
}
