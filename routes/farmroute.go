package routes

import (
	"github.com/gin-gonic/gin"
)

func InitializeFarmRoutes(router *gin.RouterGroup, deps *Dependencies) {
	farm := router.Group("/farms")
	{
		farm.GET("/:id", deps.FarmController.GetFarmByFarmID)                        // Get farm by farmID
		farm.GET("/:id/soil-test", deps.SoilTestReportController.GetSoilTestReports) // Get soil test report for farm by ID
	}
}
