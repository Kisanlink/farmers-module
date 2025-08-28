package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthenticationMiddleware handles JWT token validation and user context setup
func AuthenticationMiddleware(aaaService services.AAAService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public routes
		if auth.IsPublicRoute(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:         "unauthorized",
				Message:       "Authorization header is required",
				Code:          "AUTH_MISSING_TOKEN",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		// Extract bearer token
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			logger.Warn("Invalid Authorization header format",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:         "unauthorized",
				Message:       "Invalid Authorization header format. Expected: Bearer <token>",
				Code:          "AUTH_INVALID_FORMAT",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			logger.Warn("Empty bearer token",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:         "unauthorized",
				Message:       "Bearer token cannot be empty",
				Code:          "AUTH_EMPTY_TOKEN",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		// Validate token with AAA service
		ctx := context.Background()
		userInfo, err := aaaService.ValidateToken(ctx, token)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
				zap.Error(err),
			)
			c.JSON(http.StatusUnauthorized, common.ErrorResponse{
				Error:         "unauthorized",
				Message:       "Invalid or expired token",
				Code:          "AUTH_INVALID_TOKEN",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		// Set user context
		userContext := &auth.UserContext{
			AAAUserID: userInfo.UserID,
			Username:  userInfo.Username,
			Email:     userInfo.Email,
			Phone:     userInfo.Phone,
			Roles:     userInfo.Roles,
		}

		// Set organization context if available
		var orgContext *auth.OrgContext
		if userInfo.OrgID != "" {
			orgContext = &auth.OrgContext{
				AAAOrgID: userInfo.OrgID,
				Name:     userInfo.OrgName,
				Type:     userInfo.OrgType,
			}
		}

		// Store contexts in Gin context
		c.Set("user_context", userContext)
		c.Set("org_context", orgContext)
		c.Set("token", token)

		logger.Debug("Authentication successful",
			zap.String("user_id", userContext.AAAUserID),
			zap.String("username", userContext.Username),
			zap.String("org_id", userInfo.OrgID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("request_id", getRequestIDFromGin(c)),
		)

		c.Next()
	}
}

// AuthorizationMiddleware handles permission checking for routes
func AuthorizationMiddleware(aaaService services.AAAService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authorization for public routes
		if auth.IsPublicRoute(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get user context from previous middleware
		userContextInterface, exists := c.Get("user_context")
		if !exists {
			logger.Error("User context not found in authorization middleware",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:         "internal_server_error",
				Message:       "Authentication context not found",
				Code:          "AUTH_CONTEXT_MISSING",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		userContext, ok := userContextInterface.(*auth.UserContext)
		if !ok {
			logger.Error("Invalid user context type",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:         "internal_server_error",
				Message:       "Invalid authentication context",
				Code:          "AUTH_CONTEXT_INVALID",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		// Get organization context
		var orgID string
		orgContextInterface, exists := c.Get("org_context")
		if exists {
			if orgContext, ok := orgContextInterface.(*auth.OrgContext); ok && orgContext != nil {
				orgID = orgContext.AAAOrgID
			}
		}

		// Get required permission for this route
		permission, exists := auth.GetPermissionForRoute(c.Request.Method, c.Request.URL.Path)
		if !exists {
			logger.Warn("No permission mapping found for route",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			// Allow access if no specific permission is required
			c.Next()
			return
		}

		// Check permission with AAA service
		ctx := context.Background()
		hasPermission, err := aaaService.CheckPermission(ctx, userContext.AAAUserID, permission.Resource, permission.Action, "", orgID)
		if err != nil {
			logger.Error("Permission check failed",
				zap.String("user_id", userContext.AAAUserID),
				zap.String("resource", permission.Resource),
				zap.String("action", permission.Action),
				zap.String("org_id", orgID),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{
				Error:         "internal_server_error",
				Message:       "Permission check failed",
				Code:          "AUTH_PERMISSION_CHECK_FAILED",
				CorrelationID: getRequestIDFromGin(c),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			logger.Warn("Permission denied",
				zap.String("user_id", userContext.AAAUserID),
				zap.String("username", userContext.Username),
				zap.String("resource", permission.Resource),
				zap.String("action", permission.Action),
				zap.String("org_id", orgID),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("request_id", getRequestIDFromGin(c)),
			)
			c.JSON(http.StatusForbidden, common.ErrorResponse{
				Error:         "forbidden",
				Message:       "Insufficient permissions to access this resource",
				Code:          "AUTH_PERMISSION_DENIED",
				CorrelationID: getRequestIDFromGin(c),
				Details: map[string]string{
					"required_resource": permission.Resource,
					"required_action":   permission.Action,
				},
			})
			c.Abort()
			return
		}

		logger.Debug("Authorization successful",
			zap.String("user_id", userContext.AAAUserID),
			zap.String("username", userContext.Username),
			zap.String("resource", permission.Resource),
			zap.String("action", permission.Action),
			zap.String("org_id", orgID),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("request_id", getRequestIDFromGin(c)),
		)

		c.Next()
	}
}

// getRequestIDFromGin extracts request ID from gin context
func getRequestIDFromGin(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}
