package routes

import (
	"log"

	"github.com/Kisanlink/farmers-module/database"
	"github.com/Kisanlink/farmers-module/middleware"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	FarmService         services.FarmServiceInterface
	UserService         services.UserServiceInterface
	FarmerService       services.FarmerServiceInterface
	FarmActivityService services.FarmActivityServiceInterface
	CropCycleService    services.CropCycleServiceInterface
	CropService         services.CropServiceInterface
	FPOService          services.FPOServiceInterface
	KisansathiService   services.KisansathiServiceInterface
}

func Setup() *gin.Engine {
	log.Println("Initializing database connection...")
	db := database.GetDatabase()

	// Initialize repositories
	farmRepo := repositories.NewFarmRepository(db)
	userRepo := repositories.NewUserRepository(db)
	farmerRepo := repositories.NewFarmerRepository(db)
	farmActivityRepo := repositories.NewFarmActivityRepository(db)
	cropCycleRepo := repositories.NewCropCycleRepository(db)
	cropRepo := repositories.NewCropRepository(db)
	fpoRepo := repositories.NewFPORepository(db)

	// Initialize services
	FPOService := services.NewFPOService(fpoRepo)

	farmService := services.NewFarmService(farmRepo)
	userService := services.NewUserService(userRepo)
	farmerService := services.NewFarmerService(farmerRepo, FPOService)
	farmActivityService := services.NewFarmActivityService(farmActivityRepo)
	cropCycleService := services.NewCropCycleService(cropCycleRepo, farmRepo)
	cropService := services.NewCropService(cropRepo)
	kisansathiService := services.NewKisansathiService(farmerRepo)

	// Setup dependencies
	deps := &Dependencies{
		FarmService:         farmService,
		UserService:         userService,
		FarmerService:       farmerService,
		FarmActivityService: farmActivityService,
		CropCycleService:    cropCycleService,
		CropService:         cropService,
		FPOService:          FPOService,
		KisansathiService:   kisansathiService,
	}

	// Setup router and middleware
	router := gin.Default()

	// Apply all middlewares including CORS
	middleware.SetupMiddlewares(router)

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
		RegisterFPORoutes(api, deps.FPOService)
		RegisterKisansathiRoutes(api, deps.KisansathiService)
	}
}
