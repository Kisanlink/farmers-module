package routes

import (
	"log"

	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	FarmService   services.FarmServiceInterface
	UserService   services.UserServiceInterface
	FarmerService services.FarmerServiceInterface
}

func Setup() *gin.Engine {
	log.Println("Initializing database connection...")
	db := database.GetDatabase()

	// Initialize repositories
	farmRepo := repositories.NewFarmRepository(db)
	userRepo := repositories.NewUserRepository(db)
	farmerRepo := repositories.NewFarmerRepository(db)

	// Initialize services
	farmService := services.NewFarmService(farmRepo)
	userService := services.NewUserService(userRepo)
	farmerService := services.NewFarmerService(farmerRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmService:   farmService,
		UserService:   userService,
		FarmerService: farmerService,
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://farmers.kisanlink.in"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	InitializeRoutes(router, deps)
	return router
}

func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	v1 := router.Group("/api/v1")
	RegisterFarmRoutes(v1, deps.FarmService, deps.UserService)
	RegisterFarmerRoutes(v1, deps.FarmerService)
}