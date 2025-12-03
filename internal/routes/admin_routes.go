package routes

import (
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/handlers"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers routes for Admin & Access Control workflows
func RegisterAdminRoutes(router *gin.RouterGroup, services *services.ServiceFactory, cfg *config.Config, logger interfaces.Logger) {
	admin := router.Group("/admin")
	{
		// W18: Seed roles and permissions
		admin.POST("/seed", handlers.SeedRolesAndPermissions(services.AdministrativeService))

		// Seed lookup data (soil types, irrigation sources)
		admin.POST("/seed/lookups", handlers.SeedLookupData(services.AdministrativeService))

		// W19: Check permission (for testing)
		admin.POST("/check-permission", handlers.CheckPermission(services.AAAService))

		// Health check
		admin.GET("/health", handlers.HealthCheck(services.AdministrativeService))

		// Audit trail
		admin.GET("/audit", handlers.GetAuditTrail(services.AuditService))

		// Reconciliation endpoints
		admin.POST("/reconcile", handlers.TriggerReconciliation(services.ReconciliationJob))
		admin.GET("/reconcile/status", handlers.GetReconciliationStatus(services.ReconciliationJob))

		// Permanent delete endpoints (super admin only)
		admin.POST("/permanent-delete", handlers.PermanentDelete(services.PermanentDeleteService))
		admin.POST("/permanent-delete/org", handlers.PermanentDeleteByOrg(services.PermanentDeleteService))
		admin.POST("/cleanup-orphaned", handlers.CleanupOrphanedRecords(services.PermanentDeleteService))
	}
}
