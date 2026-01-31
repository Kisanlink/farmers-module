package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCORSMiddleware returns a CORS middleware configuration with specified origins and credentials
// This is the unified CORS middleware function that should be used across the application.
//
// Parameters:
//   - allowedOrigins: List of allowed origins (e.g., ["http://localhost:3000", "https://example.com"])
//   - allowCredentials: Whether to allow credentials (cookies, authorization headers, etc.)
//
// Hardcoded sensible defaults:
//   - Methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
//   - Headers: Origin, Content-Type, Accept, Authorization, X-Requested-With, aaa-auth-token, X-Request-ID
//   - Exposed Headers: Content-Length, Content-Type, Authorization
//   - Max Age: 12 hours (43200 seconds)
func NewCORSMiddleware(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	// Default to allowing all origins if none provided (for development)
	if len(allowedOrigins) == 0 {
		return cors.New(cors.Config{
			AllowAllOrigins: true,
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"HEAD",
				"OPTIONS",
			},
			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Accept",
				"Authorization",
				"X-Requested-With",
				"aaa-auth-token",
				"X-Request-ID",
				"X-Organization-ID",
			},
			ExposeHeaders: []string{
				"Content-Length",
				"Content-Type",
				"Authorization",
			},
			AllowCredentials: allowCredentials,
			MaxAge:           12 * 60 * 60, // 12 hours
		})
	}

	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"aaa-auth-token",
			"X-Request-ID",
			"X-Organization-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * 60 * 60, // 12 hours in seconds
	})
}
