package middleware

import (
	"github.com/gin-gonic/gin"
)

// SetupMiddlewares applies all necessary middlewares to the router
func SetupMiddlewares(router *gin.Engine) {
	// Apply CORS middleware first (should be one of the first middlewares)
	router.Use(CORSMiddleware())

	// Apply logging middleware for request monitoring
	router.Use(RequestLoggerMiddleware())

	// Add other middlewares here as needed
	// router.Use(LoggerMiddleware())
	// router.Use(RecoveryMiddleware())
	// router.Use(AuthenticateGRPC())
}

// SetupMiddlewaresWithAuth applies middlewares including authentication
func SetupMiddlewaresWithAuth(router *gin.Engine) {
	// Apply CORS middleware first
	router.Use(CORSMiddleware())

	// Apply logging middleware
	router.Use(RequestLoggerMiddleware())

	// Apply authentication middleware
	router.Use(AuthenticateGRPC())

	// Add other middlewares here as needed
}

// SetupMiddlewaresWithCredentials applies middlewares with CORS credentials support
func SetupMiddlewaresWithCredentials(router *gin.Engine, allowedOrigins []string) {
	// Apply CORS middleware with credentials support
	router.Use(CORSMiddlewareWithCredentials(allowedOrigins))

	// Apply logging middleware
	router.Use(RequestLoggerMiddleware())

	// Add other middlewares here as needed
}
