package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminSeedResponse represents a seed roles and permissions response
type AdminSeedResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
}

// AdminHealthResponse represents a health check response
type AdminHealthResponse struct {
	Status     string                     `json:"status"`
	Message    string                     `json:"message,omitempty"`
	Components map[string]ComponentHealth `json:"components"`
	Duration   string                     `json:"duration"`
	Timestamp  string                     `json:"timestamp"`
}

// ComponentHealth represents the health status of a system component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// SeedRolesAndPermissions handles W18: Seed roles and permissions
// @Summary Seed roles and permissions
// @Description Initialize the system with default roles and permissions
// @Tags admin
// @Accept json
// @Produce json
// @Param request body requests.SeedRolesAndPermissionsRequest false "Seed request parameters"
// @Success 200 {object} responses.SwaggerAdminSeedResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /admin/seed [post]
func SeedRolesAndPermissions(service services.AdministrativeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.SeedRolesAndPermissionsRequest

		// Bind JSON request if provided, otherwise use defaults
		if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse{
				Error:         "Invalid request format",
				Message:       err.Error(),
				Code:          "INVALID_REQUEST",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		// Set base request fields
		req.RequestID = c.GetString("correlation_id")
		req.Timestamp = time.Now()
		if userID := c.GetString("user_id"); userID != "" {
			req.UserID = userID
		}
		if orgID := c.GetString("org_id"); orgID != "" {
			req.OrgID = orgID
		}

		result, err := service.SeedRolesAndPermissions(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
				Error:         "Failed to seed roles and permissions",
				Message:       err.Error(),
				Code:          "SEED_FAILED",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		if response, ok := result.(*responses.SeedRolesAndPermissionsResponse); ok {
			if response.Success {
				c.JSON(http.StatusOK, response)
			} else {
				c.JSON(http.StatusInternalServerError, response)
			}
		} else {
			c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
				Error:         "Invalid response format",
				Message:       "Service returned unexpected response type",
				Code:          "INTERNAL_ERROR",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
		}
	}
}

// CheckPermissionRequest represents a permission check request
type CheckPermissionRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
	Object   string `json:"object,omitempty"`
	OrgID    string `json:"org_id,omitempty"`
}

// CheckPermissionResponse represents a permission check response
type CheckPermissionResponse struct {
	Message       string              `json:"message"`
	Data          CheckPermissionData `json:"data"`
	CorrelationID string              `json:"correlation_id"`
	Timestamp     time.Time           `json:"timestamp"`
}

// CheckPermissionData represents the permission check result data
type CheckPermissionData struct {
	Subject  string `json:"subject"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Object   string `json:"object"`
	OrgID    string `json:"org_id"`
	Allowed  bool   `json:"allowed"`
}

// CheckPermission handles W19: Check permission
// @Summary Check user permission
// @Description Check if a user has permission to perform a specific action
// @Tags admin
// @Accept json
// @Produce json
// @Param permission body CheckPermissionRequest true "Permission check data"
// @Success 200 {object} responses.SwaggerCheckPermissionResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /admin/permissions/check [post]
func CheckPermission(service services.AAAService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CheckPermissionRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse{
				Error:         "Invalid request format",
				Message:       err.Error(),
				Code:          "INVALID_REQUEST",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		// Call the AAA service directly with the correct parameters
		allowed, err := service.CheckPermission(c.Request.Context(), req.Subject, req.Resource, req.Action, req.Object, req.OrgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
				Error:         "Failed to check permission",
				Message:       err.Error(),
				Code:          "PERMISSION_CHECK_FAILED",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		c.JSON(http.StatusOK, CheckPermissionResponse{
			Message: "Permission check completed",
			Data: CheckPermissionData{
				Subject:  req.Subject,
				Resource: req.Resource,
				Action:   req.Action,
				Object:   req.Object,
				OrgID:    req.OrgID,
				Allowed:  allowed,
			},
			CorrelationID: c.GetString("correlation_id"),
			Timestamp:     time.Now(),
		})
	}
}

// HealthCheck handles comprehensive health check
// @Summary Health check
// @Description Check the health status of the service and its dependencies
// @Tags admin
// @Accept json
// @Produce json
// @Param components query string false "Comma-separated list of components to check"
// @Success 200 {object} responses.SwaggerAdminHealthResponse
// @Failure 503 {object} responses.SwaggerAdminHealthResponse
// @Router /admin/health [get]
func HealthCheck(service services.AdministrativeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.HealthCheckRequest

		// Parse query parameters
		if components := c.Query("components"); components != "" {
			// Split comma-separated components
			req.Components = []string{components} // Simplified for now
		}

		// Set base request fields
		req.RequestID = c.GetString("correlation_id")
		req.Timestamp = time.Now()
		if userID := c.GetString("user_id"); userID != "" {
			req.UserID = userID
		}
		if orgID := c.GetString("org_id"); orgID != "" {
			req.OrgID = orgID
		}

		result, err := service.HealthCheck(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, responses.ErrorResponse{
				Error:         "Health check failed",
				Message:       err.Error(),
				Code:          "HEALTH_CHECK_FAILED",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
			return
		}

		if response, ok := result.(*responses.HealthCheckResponse); ok {
			statusCode := http.StatusOK
			if response.Status != "healthy" {
				statusCode = http.StatusServiceUnavailable
			}
			c.JSON(statusCode, response)
		} else {
			c.JSON(http.StatusInternalServerError, responses.ErrorResponse{
				Error:         "Invalid response format",
				Message:       "Service returned unexpected response type",
				Code:          "INTERNAL_ERROR",
				CorrelationID: c.GetString("correlation_id"),
				Timestamp:     time.Now(),
			})
		}
	}
}

// AuditTrailResponse represents an audit trail response
type AuditTrailResponse struct {
	Message       string         `json:"message"`
	Data          AuditTrailData `json:"data"`
	CorrelationID string         `json:"correlation_id"`
	Timestamp     time.Time      `json:"timestamp"`
}

// AuditTrailData represents the audit trail data
type AuditTrailData struct {
	AuditLogs  []interface{}     `json:"audit_logs"`
	Filters    AuditTrailFilters `json:"filters"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// AuditTrailFilters represents the filters applied to audit trail
type AuditTrailFilters struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
}

// GetAuditTrail handles getting audit trail
// @Summary Get audit trail
// @Description Retrieve the audit trail for system activities
// @Tags admin
// @Accept json
// @Produce json
// @Param start_date query string false "Start date for audit logs (YYYY-MM-DD)"
// @Param end_date query string false "End date for audit logs (YYYY-MM-DD)"
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Success 200 {object} responses.SwaggerAuditTrailResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
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
		// TODO: Implement proper audit trail functionality

		// For now return empty result with proper structure
		c.JSON(http.StatusOK, AuditTrailResponse{
			Message: "Audit trail retrieved successfully",
			Data: AuditTrailData{
				AuditLogs: []interface{}{},
				Filters: AuditTrailFilters{
					StartDate: startDate,
					EndDate:   endDate,
					UserID:    userID,
					Action:    action,
				},
				TotalCount: 0,
				Page:       1,
				PageSize:   50,
			},
			CorrelationID: c.GetString("correlation_id"),
			Timestamp:     time.Now(),
		})
	}
}
