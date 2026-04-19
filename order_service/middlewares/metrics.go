package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tird4d/go-microservices/order_service/metrics"
)

// MetricsMiddleware records request count and latency per route.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		metrics.RequestCounter.WithLabelValues(endpoint).Inc()
		metrics.RequestDurationHistogram.WithLabelValues(endpoint).Observe(time.Since(start).Seconds())
	}
}
