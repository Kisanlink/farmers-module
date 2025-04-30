package routes

import (
	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	FarmService         services.FarmServiceInterface
	UserService         services.UserServiceInterface
	FarmerService       services.FarmerServiceInterface
	FarmActivityService services.FarmActivityServiceInterface
	CropCycleService    services.CropCycleServiceInterface
	CropService         services.CropServiceInterface
}

func Setup() *gin.Engine {
	utils.Log.Info("Initializing database connection...")
	db := database.GetDatabase()

	// Initialize repositories
	farmRepo := repositories.NewFarmRepository(db)
	userRepo := repositories.NewUserRepository(db)
	farmerRepo := repositories.NewFarmerRepository(db)
	farmActivityRepo := repositories.NewFarmActivityRepository(db)
	cropCycleRepo := repositories.NewCropCycleRepository(db)
	cropRepo := repositories.NewCropRepository(db)

	// Initialize services
	farmService := services.NewFarmService(farmRepo)
	userService := services.NewUserService(userRepo)
	farmerService := services.NewFarmerService(farmerRepo)
	farmActivityService := services.NewFarmActivityService(farmActivityRepo)
	cropCycleService := services.NewCropCycleService(cropCycleRepo, farmRepo)
	cropService := services.NewCropService(cropRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmService:         farmService,
		UserService:         userService,
		FarmerService:       farmerService,
		FarmActivityService: farmActivityService,
		CropCycleService:    cropCycleService,
		CropService:         cropService,
	}

	// Setup router and middleware
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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}

func InitializeRoutes(router *gin.Engine, deps *Dependencies) {
	api := router.Group("/api/v1")
	{
		RegisterFarmRoutes(api, deps.FarmService, deps.UserService)
		RegisterFarmerRoutes(api, deps.FarmerService)
		RegisterFarmActivityRoutes(api, deps.FarmActivityService)
		RegisterCropCycleRoutes(api, deps.CropCycleService)
		RegisterCropRoutes(api, deps.CropService)
	}
}
