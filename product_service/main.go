package main

import (
	"context"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/product_service/config"
	"github.com/tird4d/go-microservices/product_service/handlers"
	"github.com/tird4d/go-microservices/product_service/interceptors"
	"github.com/tird4d/go-microservices/product_service/logger"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
	"github.com/tird4d/go-microservices/product_service/tracing"

	"google.golang.org/grpc"
)

func main() {

	logger.InitLogger(true)

	err := godotenv.Load()
	if err != nil {
		logger.Log.Infow("⚠️ Error loading .env file", "error", err)
	}

	// Initialize Jaeger tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "jaeger:4317" // Default for docker-compose
	}
	tp, err := tracing.InitTracer("product-service", jaegerEndpoint)
	if err != nil {
		logger.Log.Errorw("❌ Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer tp.Shutdown(context.Background())

	_, err = config.ConnectDB()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logger.Log.Errorw("❌ Failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor))

	// ✅ Register UserService
	productpb.RegisterProductServiceServer(grpcServer, &handlers.Server{})

	// ✅ Health Check setup
	// healthServer := health.NewServer()
	// // Set the health status for the main service
	// healthServer.SetServingStatus("user.UserService", healthpb.HealthCheckResponse_SERVING)
	// // وضعیت کل سرور gRPC
	// healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	// // ثبت health server روی gRPC
	// healthpb.RegisterHealthServer(grpcServer, healthServer)

	logger.Log.Infow("َ✅ gRPC server is running on port", "port", 50053)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("❌ Failed to serve", "error", err)
	}

}
