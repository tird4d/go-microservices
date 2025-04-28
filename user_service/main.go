package main

import (
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/user_service/config"
	"github.com/tird4d/go-microservices/user_service/handlers"
	"github.com/tird4d/go-microservices/user_service/logger"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatal("⚠️ Error loading .env file", "error", err)
	}

	logger.InitLogger(true)

	config.ConnectDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Log.Errorw("❌ Failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

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
