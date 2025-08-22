package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// SeedRolesAndPermissions handles W18: Seed roles and permissions
// @Summary Seed roles and permissions
// @Description Initialize the system with default roles and permissions
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/seed [post]
func SeedRolesAndPermissions(service services.AAAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := service.SeedRolesAndPermissions(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to seed roles and permissions",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Roles and permissions seeded successfully",
		})
	}
}

// CheckPermission handles W19: Check permission
// @Summary Check user permission
// @Description Check if a user has permission to perform a specific action
// @Tags admin
// @Accept json
// @Produce json
// @Param permission body object true "Permission check data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/permissions/check [post]
func CheckPermission(service services.AAAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Subject  string `json:"subject" binding:"required"`
			Resource string `json:"resource" binding:"required"`
			Action   string `json:"action" binding:"required"`
			Object   string `json:"object,omitempty"`
			OrgID    string `json:"org_id,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert to map for service call
		permissionReq := map[string]interface{}{
			"subject":  req.Subject,
			"resource": req.Resource,
			"action":   req.Action,
			"object":   req.Object,
			"org_id":   req.OrgID,
		}

		allowed, err := service.CheckPermission(c.Request.Context(), permissionReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to check permission",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Permission check completed",
			"data": gin.H{
				"subject":  req.Subject,
				"resource": req.Resource,
				"action":   req.Action,
				"object":   req.Object,
				"org_id":   req.OrgID,
				"allowed":  allowed,
			},
		})
	}
}

// HealthCheck handles health check
// @Summary Health check
// @Description Check the health status of the service
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/health [get]
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Service is healthy",
		})
	}
}

// GetAuditTrail handles getting audit trail
// @Summary Get audit trail
// @Description Retrieve the audit trail for system activities
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/audit [get]
func GetAuditTrail() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		userID := c.Query("user_id")
		action := c.Query("action")

		// TODO: Add validation for date formats

		// TODO: Call audit service to retrieve filtered audit logs

		// For now return empty result
		c.JSON(http.StatusOK, gin.H{
			"message": "Audit trail retrieved successfully",
			"data": gin.H{
				"audit_logs": []interface{}{},
				"filters": gin.H{
					"start_date": startDate,
					"end_date":   endDate,
					"user_id":    userID,
					"action":     action,
				},
			},
		})
	}
}
