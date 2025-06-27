package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterFPORoutes(r *gin.RouterGroup, svc services.FPOServiceInterface) {
	h := handlers.NewFPOHandler(svc)

	fpo := r.Group("/fpo")
	{
		fpo.POST("", h.Create)
		fpo.GET("", h.List)
		fpo.GET("/:reg_no", h.Get)
		fpo.PUT("/:reg_no", h.Update)
		fpo.DELETE("/:reg_no", h.Delete)
	}
}
