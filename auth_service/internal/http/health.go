package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tird4d/go-microservices/auth_service/config"
)

func StartHealthServer() {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// چک کردن اتصال Redis
		_, err := config.RedisClient.Ping(ctx).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "unhealthy",
				"redis":   "unreachable",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"redis":  "ok",
		})
	})

	go func() {
		_ = router.Run(":8081")
	}()
}
