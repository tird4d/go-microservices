package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tird4d/go-microservices/product_service/config"
	"github.com/tird4d/go-microservices/product_service/handlers"
	"github.com/tird4d/go-microservices/product_service/interceptors"
	"github.com/tird4d/go-microservices/product_service/logger"
	"github.com/tird4d/go-microservices/product_service/metrics"
	productpb "github.com/tird4d/go-microservices/product_service/proto"
	"github.com/tird4d/go-microservices/product_service/tracing"

	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {

	logger.InitLogger(true)

	// ============ PROMETHEUS METRICS ============
	metrics.InitMetrics()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Log.Fatalw("❌ Failed to start metrics HTTP server", "error", err)
		}
	}()

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
	healthServer := health.NewServer()
	// // Set the health status for the main service
	healthServer.SetServingStatus("product.ProductService", healthpb.HealthCheckResponse_SERVING)
	// // وضعیت کل سرور gRPC
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	// // ثبت health server روی gRPC
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	logger.Log.Infow("َ✅ gRPC server is running on port", "port", 50053)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("❌ Failed to serve", "error", err)
	}

}
