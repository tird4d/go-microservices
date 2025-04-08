package main

import (
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/user_service/handlers"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"github.com/tird4d/user-api/config"
	"google.golang.org/grpc"
)

type userServer struct {
	userpb.UnimplementedUserServiceServer
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}

	config.ConnectDB()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// ثبت سرور با gRPC
	userpb.RegisterUserServiceServer(grpcServer, &handlers.Server{})

	log.Println("✅ gRPC server is running on port 50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}

}
