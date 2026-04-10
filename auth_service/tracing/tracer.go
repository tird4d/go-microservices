package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer initializes the OpenTelemetry tracer with OTLP exporter.
// Called once in main.go at startup. Takes two inputs:
//   - serviceName: e.g. "auth-service" — tags every span so Jaeger knows who sent it
//   - jaegerEndpoint: e.g. "jaeger:4317" — where to ship the spans
func InitTracer(serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// ============ STEP 1: EXPORTER (the "sender") ============
	// Creates a connection to Jaeger using OTLP protocol over gRPC.
	// Like a database connection — nothing is sent yet, just the pipe is set up.
	// WithInsecure() = no TLS needed (services talk on internal Docker network).
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(jaegerEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, err
	}

	// ============ STEP 2: RESOURCE (the "who am I" tag) ============
	// Every span sent to Jaeger will carry service.name="auth-service".
	// Without this, all spans from all services would look the same in Jaeger UI.
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// ============ STEP 3: TRACER PROVIDER (the "engine") ============
	// WithBatcher: collects spans in memory, sends in batches every few seconds
	//              (more efficient than sending each span one by one).
	// WithResource: attaches the service.name tag to every span automatically.
	// WithSampler AlwaysSample: records 100% of requests.
	//   NOTE: in production use TraceIDRatioBased(0.1) to sample only 10% — saves cost.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// ============ STEP 4: SET GLOBAL PROVIDER ============
	// Makes this TracerProvider available everywhere via otel.Tracer("auth-service").
	// This is how the interceptor picks it up without being passed anything explicitly.
	otel.SetTracerProvider(tp)

	// ============ STEP 5: SET PROPAGATOR (how trace IDs travel between services) ============
	// When auth-service calls user-service, the trace ID must be injected into
	// the outgoing gRPC headers, and extracted on the other side.
	// TraceContext{} = W3C standard format: "traceparent" header.
	// This is the universal format understood by Jaeger, Zipkin, and all modern tools.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tp, nil
}

// Shutdown gracefully flushes and closes the tracer provider.
// Called with defer in main.go so any spans still buffered in memory
// are sent to Jaeger before the process exits — without this they would be lost.
func Shutdown(tp *sdktrace.TracerProvider) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return tp.Shutdown(ctx)
}
