package main

import (
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tird4d/go-microservices/user_service/config"
	"github.com/tird4d/go-microservices/user_service/handlers"
	"github.com/tird4d/go-microservices/user_service/interceptors"
	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/metrics"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"github.com/tird4d/go-microservices/user_service/tracing"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger.InitLogger(true)

	err := godotenv.Load()
	if err != nil {
		logger.Log.Infow("⚠️ Error loading .env file", "error", err)
	}

	// Initialize tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "jaeger:4317" // Default for docker-compose
	}
	
	tp, err := tracing.InitTracer("user-service", jaegerEndpoint)
	if err != nil {
		logger.Log.Errorw("❌ Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := tracing.Shutdown(tp); err != nil {
			logger.Log.Errorw("❌ Failed to shutdown tracer", "error", err)
		}
	}()
	logger.Log.Infow("✅ Tracing initialized", "endpoint", jaegerEndpoint)



	_, err = config.ConnectDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Log.Errorw("❌ Failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		metrics.InitMetrics()

		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Log.Fatal("❌ Failed to start metrics HTTP server", "error", err)
		}
		logger.Log.Infow("✅ Metrics server is running on port", "port", 2112)
	}()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor),
	)

	// ✅ Register UserService
	userpb.RegisterUserServiceServer(grpcServer, &handlers.Server{})

	// ✅ Health Check setup
	healthServer := health.NewServer()
	// Set the health status for the main service
	healthServer.SetServingStatus("user.UserService", healthpb.HealthCheckResponse_SERVING)
	// وضعیت کل سرور gRPC
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	// ثبت health server روی gRPC
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	logger.Log.Infow("َ✅ gRPC server is running on port", "port", 50051)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("❌ Failed to serve", "error", err)
	}

}
