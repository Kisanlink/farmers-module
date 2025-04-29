package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthenticateGRPC() gin.HandlerFunc {
	return func(c *gin.Context) {
		grpc_token := c.GetHeader("aaa-auth-token")
		if grpc_token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC token is missing"})
			c.Abort()
			return
		}

		c.Set("aaa-auth-token", grpc_token)
		c.Next()
	}
}
