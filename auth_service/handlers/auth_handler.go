// user_handlers.go
package handlers

import (
	"context"
	"log"
	"time"

	"github.com/tird4d/go-microservices/auth_service/proto"
	pb "github.com/tird4d/go-microservices/auth_service/proto"
	"github.com/tird4d/go-microservices/auth_service/services"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("📥 Login called for email: %s and pass is: %s", req.Email, req.Password)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	token, err := services.LoginUser(ctx, req.Email, req.Password)
	message := "Login successful"
	if err != nil {
		log.Printf("❌ Login failed: %v", err)
		message = err.Error()
	}

	return &pb.LoginResponse{
		Token:   token,
		Message: message,
	}, nil
}

func (s *AuthServer) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	log.Printf("🔐 Validate called with token: %s", req.Token)

	// فقط برای تست
	return &pb.ValidateResponse{
		UserId: "user123",
		Email:  "ali@example.com",
	}, nil
}
