package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterStageRoutes(router *gin.RouterGroup, stageService services.StageServiceInterface) {
	handler := handlers.NewStageHandler(stageService)

	stageRoutes := router.Group("/stages")
	{
		stageRoutes.POST("", handler.CreateStage)
		stageRoutes.GET("", handler.GetAllStages)
		stageRoutes.GET("/:id", handler.GetStageByID)
		stageRoutes.PUT("/:id", handler.UpdateStage)
		stageRoutes.DELETE("/:id", handler.DeleteStage)
	}
}
