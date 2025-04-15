package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
)

func JWTAuthMiddleware(authClient authpb.AuthServiceClient) gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missed"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := authClient.Validate(c.Request.Context(), &authpb.ValidateRequest{
			Token: token,
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserId)
		c.Set("email", claims.Email)
		// c.Set("role", claims["role"])
		// c.Set("auth_at", claims["auth_at"])
		c.Next()

	}
}
