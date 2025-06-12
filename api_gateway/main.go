package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tird4d/go-microservices/api_gateway/handlers"
	"github.com/tird4d/go-microservices/api_gateway/middlewares"
	authpb "github.com/tird4d/go-microservices/auth_service/proto"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connecting to gRPC server for user service
	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ could not connect to gRPC server: %v", err)
	}
	// defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

	// Connecting to gRPC server for auth service
	authConn, err := grpc.DialContext(ctx, "localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ could not connect to auth gRPC server: %v", err)
	}
	// defer authConn.Close()

	authClient := authpb.NewAuthServiceClient(authConn)

	// ایجاد روت‌ها
	router := gin.Default()

	userHandler := handlers.UserHandler{
		UserClient: userClient,
	}

	authHandler := handlers.GatewayHandler{
		AuthClient: authClient,
	}

	adminHandler := handlers.AdminHandler{
		UserClient: userClient,
	}

	router.POST("/api/v1/register", userHandler.RegisterHandler)
	router.POST("/api/v1/refresh-token", authHandler.RefreshTokenHandler)
	router.POST("/api/v1/login", authHandler.LoginHandler)

	auth := router.Group("/api/v1/")
	auth.Use(middlewares.JWTAuthMiddleware(authClient))
	auth.GET("/me", userHandler.MeHandler)
	auth.POST("/logout", authHandler.LogoutHandler)

	admin := router.Group("/api/v1/admin")
	admin.Use(middlewares.JWTAuthMiddleware(authClient))
	admin.Use(middlewares.AdminMiddleware(authClient))
	admin.GET("/users", adminHandler.UsersHandler)
	admin.PUT("/users/{id}", adminHandler.UpdateUserHandler)

	log.Println("🚀 API Gateway is running on http://localhost:8080")
	router.Run(":8080")
}
