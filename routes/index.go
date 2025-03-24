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
	FarmService   services.FarmServiceInterface // ✅ Use interface
}


// Setup initializes the database, services, handlers, and routes
func Setup() *gin.Engine {
	log.Println("Initializing database connection...") // ✅ Log DB initialization

	db := database.GetDatabase() // Get database instance

	// Initialize repositories
	farmerRepo := repositories.NewFarmerRepository(db)
	farmRepo := repositories.NewFarmRepository(db)

	// Initialize services
	farmerService := services.NewFarmerService(farmerRepo)
	farmService := services.NewFarmService(farmRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmerService: farmerService,
		FarmService:   farmService,
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

	InitializeRoutes(router, deps) // Initialize API routes

	return router
}

// InitializeRoutes sets up the API routes with handlers
func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	v1 := router.Group("/api/v1")

	RegisterFarmerRoutes(v1, deps.FarmerService)
	RegisterFarmRoutes(v1, deps.FarmService) // ✅ Register farm routes
}
