package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/auth_service/handlers"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	userpb "github.com/tird4d/go-microservices/user_service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	if err != nil {
		status.Errorf(codes.Unavailable, "cannot connect to user service: %v", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	authServer := &handlers.AuthServer{
		UserClient: userClient,
	}

	authpb.RegisterAuthServiceServer(grpcServer, authServer)

	log.Println("✅ gRPC server is running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}

}
