package main

import (
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/auth_service/handlers"
	pb "github.com/tird4d/go-microservices/auth_service/proto"

	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(grpcServer, &handlers.AuthServer{})

	log.Println("✅ gRPC server is running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}

}
