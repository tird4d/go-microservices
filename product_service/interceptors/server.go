package interceptors

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tird4d/go-microservices/product_service/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var tracer = otel.Tracer("product-service")

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract trace context from incoming RPC
	propagator := otel.GetTextMapPropagator()
	carrier := MetadataCarrierFromContext(ctx)
	ctx = propagator.Extract(ctx, carrier)

	// Create span for this RPC
	ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	span.SetAttributes(
attribute.String("rpc.service", info.FullMethod),
attribute.String("rpc.method", info.FullMethod),
)

	// ============ METRICS ============
	timer := prometheus.NewTimer(metrics.RequestDurationHistogram.WithLabelValues(info.FullMethod))
	defer timer.ObserveDuration()
	metrics.RequestCounter.WithLabelValues(info.FullMethod).Inc()

	// Call the handler
	resp, err := handler(ctx, req)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return resp, err
}

func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Create span for outgoing call
	ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
attribute.String("rpc.service", method),
attribute.String("rpc.method", method),
)

	// Inject trace context into outgoing RPC
	propagator := otel.GetTextMapPropagator()
	carrier := make(propagation.MapCarrier)
	propagator.Inject(ctx, carrier)

	ctx = InjectIntoMetadata(ctx, carrier)

	// Call the RPC
	err := invoker(ctx, method, req, reply, cc, opts...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}
