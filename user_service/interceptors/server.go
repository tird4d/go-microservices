package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor handles tracing, metrics, and logging for all incoming RPC calls
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// ============ TRACE CONTEXT EXTRACTION ============
	// Extract trace context from incoming gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		carrier := make(propagation.MapCarrier)
		for k, v := range md {
			if len(v) > 0 {
				carrier[k] = v[0]
			}
		}
		// Extract using global propagator (W3C Trace Context)
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	// ============ TRACING ============
	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, info.FullMethod)
	defer span.End()

	// Add RPC method as attribute
	span.SetAttributes(
		attribute.String("rpc.method", info.FullMethod),
		attribute.String("rpc.service", "user-service"),
	)

	// ============ METRICS ============
	// Start timer for request duration
	timer := prometheus.NewTimer(metrics.RequestDurationHistogram.WithLabelValues(info.FullMethod))
	defer timer.ObserveDuration()

	// Increment request counter
	metrics.RequestCounter.WithLabelValues(info.FullMethod).Inc()

	// ============ LOGGING ============
	logger.Log.Infow("RPC called", "method", info.FullMethod)

	// ============ EXECUTE HANDLER ============
	resp, err := handler(ctx, req)

	// ============ ERROR HANDLING ============
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Log.Errorw("RPC error", "method", info.FullMethod, "error", err)
	}

	// ============ DURATION LOGGING ============
	duration := time.Since(start).Milliseconds()
	logger.Log.Infow("RPC completed", "method", info.FullMethod, "duration_ms", duration, "error", err != nil)

	return resp, err
}
