package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrorHandlerMiddleware handles panics and converts errors to structured responses
func ErrorHandlerMiddleware(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestIDFromGin(c)

				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("request_id", requestID),
				)

				c.JSON(http.StatusInternalServerError, common.ErrorResponse{
					Error:         "internal_server_error",
					Message:       "An unexpected error occurred",
					Code:          "INTERNAL_SERVER_ERROR",
					CorrelationID: requestID,
				})
				c.Abort()
			}
		}()

		c.Next()

		// Handle errors that were added to the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handleError(c, err.Err, logger)
		}
	}
}

// handleError converts different error types to appropriate HTTP responses
func handleError(c *gin.Context, err error, logger interfaces.Logger) {
	requestID := getRequestIDFromGin(c)

	// Log the error
	logger.Error("Request error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("request_id", requestID),
	)

	// Handle different error types
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, common.ErrorResponse{
			Error:         "not_found",
			Message:       "Resource not found",
			Code:          "RESOURCE_NOT_FOUND",
			CorrelationID: requestID,
		})

	case isAuthenticationError(err):
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			Error:         "unauthorized",
			Message:       err.Error(),
			Code:          "AUTHENTICATION_FAILED",
			CorrelationID: requestID,
		})

	case isAuthorizationError(err):
		c.JSON(http.StatusForbidden, common.ErrorResponse{
			Error:         "forbidden",
			Message:       err.Error(),
			Code:          "AUTHORIZATION_FAILED",
			CorrelationID: requestID,
		})

	case isValidationError(err):
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:         "validation_error",
			Message:       err.Error(),
			Code:          "VALIDATION_FAILED",
			CorrelationID: requestID,
		})

	case isDuplicateKeyError(err):
		c.JSON(http.StatusConflict, common.ErrorResponse{
			Error:         "conflict",
			Message:       "Resource already exists",
			Code:          "DUPLICATE_RESOURCE",
			CorrelationID: requestID,
		})

	case isForeignKeyError(err):
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:         "bad_request",
			Message:       "Invalid reference to related resource",
			Code:          "INVALID_REFERENCE",
			CorrelationID: requestID,
		})

	case isConstraintError(err):
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			Error:         "bad_request",
			Message:       "Data constraint violation",
			Code:          "CONSTRAINT_VIOLATION",
			CorrelationID: requestID,
		})

	case isServiceUnavailableError(err):
		c.JSON(http.StatusServiceUnavailable, common.ErrorResponse{
			Error:         "service_unavailable",
			Message:       err.Error(),
			Code:          "SERVICE_UNAVAILABLE",
			CorrelationID: requestID,
		})

	default:
		// Generic internal server error
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{
			Error:         "internal_server_error",
			Message:       "An unexpected error occurred",
			Code:          "INTERNAL_SERVER_ERROR",
			CorrelationID: requestID,
		})
	}
}

// Error type checking functions

func isValidationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "validation") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "format")
}

func isDuplicateKeyError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate") ||
		strings.Contains(errMsg, "unique") ||
		strings.Contains(errMsg, "already exists")
}

func isForeignKeyError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "foreign key") ||
		strings.Contains(errMsg, "violates foreign key constraint") ||
		strings.Contains(errMsg, "fk_")
}

func isConstraintError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "constraint") ||
		strings.Contains(errMsg, "check constraint") ||
		strings.Contains(errMsg, "violates check")
}

func isAuthenticationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "authentication") ||
		strings.Contains(errMsg, "invalid token") ||
		strings.Contains(errMsg, "token expired") ||
		strings.Contains(errMsg, "unauthorized")
}

func isAuthorizationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "authorization") ||
		strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "access denied")
}

func isServiceUnavailableError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "aaa service")
}
