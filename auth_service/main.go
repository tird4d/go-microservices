package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tird4d/go-microservices/auth_service/config"
	"github.com/tird4d/go-microservices/auth_service/handlers"
	"github.com/tird4d/go-microservices/auth_service/interceptors"
	"github.com/tird4d/go-microservices/auth_service/logger"
	"github.com/tird4d/go-microservices/auth_service/metrics"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	"github.com/tird4d/go-microservices/auth_service/tracing"
	userpb "github.com/tird4d/go-microservices/user_service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger.InitLogger(true)

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default values")
	}

	// Initialize tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "jaeger:4317" // Default for docker-compose
	}


	// Prometheus initialization
	metrics.InitMetrics()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			logger.Log.Fatalw("❌ Failed to start metrics HTTP server", "error", err)
		}
	}()


	tp, err := tracing.InitTracer("auth-service", jaegerEndpoint)
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



	// Set up a timeout for the connection
	// This is useful to avoid hanging indefinitely if the server is not reachable
	// or if there are network issues.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to Redis
	config.ConnectRedis()

	// Make gRPC connection to User Service
	// The connection is made to the User Service running on localhost:50051
	conn, err := grpc.DialContext(ctx, os.Getenv("USER_SERVICE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(interceptors.UnaryClientInterceptor))
	if err != nil {
		log.Fatalf("❌ could not connect to gRPC server: %v", err)
	}

	userClient := userpb.NewUserServiceClient(conn)

	// Start the gRPC server
	// The server listens on port 50052
	// and handles incoming requests for the AuthService
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}


	//Start gRPC Server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor),
	)

	authServer := &handlers.AuthServer{
		UserClient: userClient,
	}

	// Register the AuthService with the gRPC server
	// This allows the server to handle requests for the AuthService
	// and route them to the appropriate handler methods
	// defined in the AuthServer struct
	authpb.RegisterAuthServiceServer(grpcServer, authServer)



	// ✅ Health Check setup
	healthServer := health.NewServer()
	// Set the health status for the main service
	healthServer.SetServingStatus("user.UserService", healthpb.HealthCheckResponse_SERVING)
	// وضعیت کل سرور gRPC
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	// ثبت health server روی gRPC
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	
	log.Println("✅ gRPC server is running on port 50052")

	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}

}
