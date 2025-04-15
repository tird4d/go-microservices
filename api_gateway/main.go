package main

import (
	"context"
	"log"
	"net/http"
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

	router.POST("/register", func(c *gin.Context) {
		var body struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := userClient.Register(ctx, &userpb.RegisterRequest{
			Name:     body.Name,
			Email:    body.Email,
			Password: body.Password,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id": res.Id,
			"message": res.Message,
		})
	})

	router.POST("/login", func(c *gin.Context) {
		var body struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := authClient.Login(ctx, &authpb.LoginRequest{
			Email:    body.Email,
			Password: body.Password,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":   res.Token,
			"message": res.Message,
		})
	})

	auth := router.Group("/")
	auth.Use(middlewares.JWTAuthMiddleware(authClient))
	auth.GET("/me", handlers.MeHandler)

	log.Println("🚀 API Gateway is running on http://localhost:8080")
	router.Run(":8080")
}
