package interceptors

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// truncateString limits string length
func truncateStringForSpan(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "...[truncated]"
	}
	return s
}

// ComprehensiveUnaryServerInterceptor captures EVERYTHING automatically:
// - Request parameters (serialized from proto)
// - Response data (serialized from proto)
// - Execution time
// - Errors
// - All without touching handler code
func ComprehensiveUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Extract trace context from metadata
	propagator := otel.GetTextMapPropagator()
	carrier := make(propagation.MapCarrier)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			if len(v) > 0 {
				carrier[k] = v[0]
			}
		}
	}
	ctx = propagator.Extract(ctx, carrier)

	// Create span for this RPC
	tracer := otel.Tracer("api-gateway")
	ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// ===== AUTOMATICALLY CAPTURE REQUEST DETAILS =====
	span.SetAttributes(
		attribute.String("rpc.method", info.FullMethod),
		attribute.String("rpc.request_type", fmt.Sprintf("%T", req)),
	)

	// If request is a protobuf message, capture its fields
	if msg, ok := req.(proto.Message); ok {
		// Convert proto to JSON for better readability
		jsonBytes, err := protojson.Marshal(msg)
		if err == nil {
			reqStr := string(jsonBytes)
			span.SetAttributes(
				attribute.String("rpc.request_data", truncateStringForSpan(reqStr, 500)),
			)
		}
	}

	// Call the actual handler
	resp, err := handler(ctx, req)

	// ===== AUTOMATICALLY CAPTURE RESPONSE DETAILS =====
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
	)

	// Capture response
	if resp != nil {
		span.SetAttributes(
			attribute.String("rpc.response_type", fmt.Sprintf("%T", resp)),
		)
		if msg, ok := resp.(proto.Message); ok {
			jsonBytes, err := protojson.Marshal(msg)
			if err == nil {
				respStr := string(jsonBytes)
				span.SetAttributes(
					attribute.String("rpc.response_data", truncateStringForSpan(respStr, 500)),
				)
			}
		}
	}

	// ===== AUTOMATICALLY CAPTURE ERRORS =====
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("error.type", "rpc_error"),
			attribute.Bool("error.occurred", true),
		)
	} else {
		span.SetStatus(codes.Ok, "")
		span.SetAttributes(
			attribute.Bool("error.occurred", false),
		)
	}

	return resp, err
}
