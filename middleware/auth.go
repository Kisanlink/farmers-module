package middleware

import(
"net/http"
"github.com/gin-gonic/gin"
)


func AuthenticateGRPC() gin.HandlerFunc {
    return func(c *gin.Context) {
        grpcToken := c.GetHeader("aaa-auth-token")
        if grpcToken == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC token is missing"})
            c.Abort()
            return
        }

        c.Set("aaa-auth-token", grpcToken)
        c.Next()
    }
	}