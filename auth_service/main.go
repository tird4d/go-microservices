package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/joho/godotenv"
	"github.com/tird4d/go-microservices/auth_service/config"
	"github.com/tird4d/go-microservices/auth_service/handlers"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	userpb "github.com/tird4d/go-microservices/user_service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}

	// Set up a timeout for the connection
	// This is useful to avoid hanging indefinitely if the server is not reachable
	// or if there are network issues.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to Redis
	config.ConnectRedis()

	// Make gRPC connection to User Service
	// The connection is made to the User Service running on localhost:50051
	conn, err := grpc.DialContext(ctx, "user-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	if err != nil {
		log.Fatalf("cannot connect to user service  updated: %v", err)
	}

	// if conn != nil {
	// 	defer conn.Close()
	// }

	userClient := userpb.NewUserServiceClient(conn)

	// Start the gRPC server
	// The server listens on port 50052
	// and handles incoming requests for the AuthService
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	authServer := &handlers.AuthServer{
		UserClient: userClient,
	}

	// Register the AuthService with the gRPC server
	// This allows the server to handle requests for the AuthService
	// and route them to the appropriate handler methods
	// defined in the AuthServer struct
	authpb.RegisterAuthServiceServer(grpcServer, authServer)

	log.Println("✅ gRPC server is running on port 50052")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}

}
