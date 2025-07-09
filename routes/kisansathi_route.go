package routes

import (
	"github.com/Kisanlink/farmers-module/handlers"
	"github.com/Kisanlink/farmers-module/services"
	"github.com/gin-gonic/gin"
)

func RegisterKisansathiRoutes(r *gin.RouterGroup, svc services.KisansathiServiceInterface) {
	h := handlers.NewKisansathiHandler(svc)
	r.GET("/kisansathis", h.List) // optional ?fpo_reg_no=...
}
