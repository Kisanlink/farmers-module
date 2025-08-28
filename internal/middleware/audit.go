package middleware

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	Timestamp    time.Time         `json:"timestamp"`
	RequestID    string            `json:"request_id"`
	Subject      string            `json:"subject"`
	Username     string            `json:"username"`
	Organization string            `json:"organization"`
	Resource     string            `json:"resource"`
	Action       string            `json:"action"`
	Object       string            `json:"object,omitempty"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	StatusCode   int               `json:"status_code"`
	Duration     time.Duration     `json:"duration"`
	UserAgent    string            `json:"user_agent"`
	ClientIP     string            `json:"client_ip"`
	Success      bool              `json:"success"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AuditMiddleware logs all requests with structured audit information
func AuditMiddleware(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Extract user context
		var subject, username, organization string
		if userContextInterface, exists := c.Get("user_context"); exists {
			if userContext, ok := userContextInterface.(*auth.UserContext); ok && userContext != nil {
				subject = userContext.AAAUserID
				username = userContext.Username
			}
		}

		if orgContextInterface, exists := c.Get("org_context"); exists {
			if orgContext, ok := orgContextInterface.(*auth.OrgContext); ok && orgContext != nil {
				organization = orgContext.AAAOrgID
			}
		}

		// Get permission information
		var resource, action string
		if permission, exists := auth.GetPermissionForRoute(c.Request.Method, c.Request.URL.Path); exists {
			resource = permission.Resource
			action = permission.Action
		}

		// Extract object ID from path parameters if available
		var object string
		if id := c.Param("id"); id != "" {
			object = id
		}

		// Determine success based on status code
		success := c.Writer.Status() < 400

		// Extract error message if request failed
		var errorMessage string
		if !success {
			if errors := c.Errors; len(errors) > 0 {
				errorMessage = errors.Last().Error()
			}
		}

		// Create audit event
		auditEvent := AuditEvent{
			Timestamp:    startTime,
			RequestID:    getRequestIDFromGin(c),
			Subject:      subject,
			Username:     username,
			Organization: organization,
			Resource:     resource,
			Action:       action,
			Object:       object,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			StatusCode:   c.Writer.Status(),
			Duration:     duration,
			UserAgent:    c.Request.UserAgent(),
			ClientIP:     c.ClientIP(),
			Success:      success,
			ErrorMessage: errorMessage,
		}

		// Add query parameters as metadata for GET requests
		if c.Request.Method == "GET" && len(c.Request.URL.RawQuery) > 0 {
			auditEvent.Metadata = map[string]string{
				"query": c.Request.URL.RawQuery,
			}
		}

		// Log audit event
		logger.Info("Audit Event",
			zap.Time("timestamp", auditEvent.Timestamp),
			zap.String("request_id", auditEvent.RequestID),
			zap.String("subject", auditEvent.Subject),
			zap.String("username", auditEvent.Username),
			zap.String("organization", auditEvent.Organization),
			zap.String("resource", auditEvent.Resource),
			zap.String("action", auditEvent.Action),
			zap.String("object", auditEvent.Object),
			zap.String("method", auditEvent.Method),
			zap.String("path", auditEvent.Path),
			zap.Int("status_code", auditEvent.StatusCode),
			zap.Duration("duration", auditEvent.Duration),
			zap.String("user_agent", auditEvent.UserAgent),
			zap.String("client_ip", auditEvent.ClientIP),
			zap.Bool("success", auditEvent.Success),
			zap.String("error_message", auditEvent.ErrorMessage),
			zap.Any("metadata", auditEvent.Metadata),
		)

		// Emit audit event for downstream processing if needed
		// This could be extended to publish to an event bus or queue
		if eventEmitter := getEventEmitterFromContext(c); eventEmitter != nil {
			if err := eventEmitter.EmitAuditEvent(auditEvent); err != nil {
				logger.Error("Failed to emit audit event", "error", err)
			}
		}
	}
}

// getEventEmitterFromContext extracts event emitter from context if available
func getEventEmitterFromContext(c *gin.Context) interfaces.EventEmitter {
	if emitter, exists := c.Get("event_emitter"); exists {
		if eventEmitter, ok := emitter.(interfaces.EventEmitter); ok {
			return eventEmitter
		}
	}
	return nil
}
