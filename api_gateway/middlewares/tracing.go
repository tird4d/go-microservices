package middlewares

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// sanitizeRequestBody removes sensitive fields from request body
func sanitizeRequestBody(body string) string {
	// List of sensitive fields to mask
	sensitivePatterns := []string{
		`"password"\s*:\s*"[^"]*"`,
		`"refresh_token"\s*:\s*"[^"]*"`,
		`"access_token"\s*:\s*"[^"]*"`,
		`"token"\s*:\s*"[^"]*"`,
		`"secret"\s*:\s*"[^"]*"`,
		`"api_key"\s*:\s*"[^"]*"`,
		`"authorization"\s*:\s*"[^"]*"`,
	}

	sanitized := body
	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, `"***"`)
	}

	return sanitized
}

// truncateString limits string length for attributes
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "...[truncated]"
	}
	return s
}

// TracingMiddleware creates a trace span for each HTTP request with full request/response data
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
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// Capture request body for POST, PUT, PATCH requests
		if strings.ToUpper(c.Request.Method) != "GET" && strings.ToUpper(c.Request.Method) != "DELETE" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Reset body so handler can read it
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Sanitize and add to span
				bodyStr := string(bodyBytes)
				sanitized := sanitizeRequestBody(bodyStr)
				truncated := truncateString(sanitized, 500)

				span.SetAttributes(
					attribute.String("http.request_body", truncated),
					attribute.Int("http.request_body_size", len(bodyBytes)),
				)
			}
		}

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
