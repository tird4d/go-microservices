package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connecting to gRPC server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ could not connect to gRPC server: %v", err)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)

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

	log.Println("🚀 API Gateway is running on http://localhost:8080")
	router.Run(":8080")
}
