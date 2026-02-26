package interceptors

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

// Helper to extract trace context from incoming RPC
func MetadataCarrierFromContext(ctx context.Context) propagation.MapCarrier {
	carrier := make(propagation.MapCarrier)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			if len(v) > 0 {
				carrier.Set(k, v[0])
			}
		}
	}
	return carrier
}

// Helper to inject trace context into outgoing metadata
func InjectIntoMetadata(ctx context.Context, carrier propagation.MapCarrier) context.Context {
	md := metadata.MD{}
	for k, v := range carrier {
		md.Set(k, v)
	}

	if existingMd, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existingMd, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
