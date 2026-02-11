package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCORSMiddleware returns a CORS middleware configuration with specified origins and credentials.
// Supports exact origins (e.g., "https://localhost:3000") and wildcard subdomain
// patterns (e.g., "https://*.kisanlink.in").
func NewCORSMiddleware(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	headers := []string{
		"Origin", "Content-Type", "Accept", "Authorization",
		"X-Requested-With", "aaa-auth-token", "X-Request-ID", "X-Organization-ID",
	}
	exposed := []string{"Content-Length", "Content-Type", "Authorization"}

	// Default to allowing all origins if none provided (for development)
	if len(allowedOrigins) == 0 {
		return cors.New(cors.Config{
			AllowAllOrigins:  true,
			AllowMethods:     methods,
			AllowHeaders:     headers,
			ExposeHeaders:    exposed,
			AllowCredentials: allowCredentials,
			MaxAge:           12 * 60 * 60,
		})
	}

	// Check if any origin contains a wildcard pattern
	hasWildcard := false
	for _, o := range allowedOrigins {
		if strings.Contains(o, "*.") {
			hasWildcard = true
			break
		}
	}

	// If no wildcards, use the simple AllowOrigins list
	if !hasWildcard {
		return cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowMethods:     methods,
			AllowHeaders:     headers,
			ExposeHeaders:    exposed,
			AllowCredentials: allowCredentials,
			MaxAge:           12 * 60 * 60,
		})
	}

	// Separate exact origins from wildcard patterns for efficient matching
	exactOrigins := make(map[string]bool)
	var wildcardPatterns []string
	for _, o := range allowedOrigins {
		if strings.Contains(o, "*.") {
			wildcardPatterns = append(wildcardPatterns, o)
		} else {
			exactOrigins[o] = true
		}
	}

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Exact match
			if exactOrigins[origin] {
				return true
			}
			// Wildcard subdomain match
			for _, pattern := range wildcardPatterns {
				if isOriginAllowed(origin, pattern) {
					return true
				}
			}
			return false
		},
		AllowMethods:     methods,
		AllowHeaders:     headers,
		ExposeHeaders:    exposed,
		AllowCredentials: allowCredentials,
		MaxAge:           12 * 60 * 60,
	})
}

// isOriginAllowed checks if an origin matches a wildcard subdomain pattern.
// Pattern format: "https://*.kisanlink.in" matches "https://admin.kisanlink.in"
// but not "https://a.b.kisanlink.in" (single subdomain level only).
func isOriginAllowed(origin, pattern string) bool {
	wildcardIdx := strings.Index(pattern, "*.")
	if wildcardIdx < 0 {
		return origin == pattern
	}

	scheme := pattern[:wildcardIdx]   // e.g., "https://"
	suffix := pattern[wildcardIdx+1:] // e.g., ".kisanlink.in"

	if !strings.HasPrefix(origin, scheme) {
		return false
	}
	if !strings.HasSuffix(origin, suffix) {
		return false
	}

	subdomain := origin[len(scheme) : len(origin)-len(suffix)]
	return len(subdomain) > 0 && !strings.Contains(subdomain, ".") && !strings.Contains(subdomain, "/")
}
