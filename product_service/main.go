package main

import (
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/product_service/config"
	"github.com/tird4d/go-microservices/product_service/handlers"
	"github.com/tird4d/go-microservices/product_service/logger"
	productpb "github.com/tird4d/go-microservices/product_service/proto"

	"google.golang.org/grpc"
)

func main() {

	logger.InitLogger(true)

	err := godotenv.Load()
	if err != nil {
		logger.Log.Infow("⚠️ Error loading .env file", "error", err)
	}

	_, err = config.ConnectDB()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logger.Log.Errorw("❌ Failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

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
