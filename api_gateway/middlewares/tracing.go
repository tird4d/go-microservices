package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// TracingMiddleware creates a trace span for each HTTP request
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := otel.Tracer("api-gateway")
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
		)

		// Update context in request
		c.Request = c.Request.WithContext(ctx)

		// Call next handler
		c.Next()

		// Record response status
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)
	}
}
