package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
)

func AdminMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {

	return func(c *gin.Context) {

		role, ok := c.Get("role")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this resource"})
			c.Abort()
			return
		}

		c.Next()

	}
}
