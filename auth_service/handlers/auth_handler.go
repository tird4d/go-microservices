// user_handlers.go
package handlers

import (
	"context"
	"log"

	pb "github.com/tird4d/go-microservices/auth_service/proto"
	"github.com/tird4d/go-microservices/auth_service/services"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	UserClient userpb.UserServiceClient
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("📥 Login called for email: %s and pass is: %s", req.Email, req.Password)

	token, err := services.LoginUser(ctx, s.UserClient, req.Email, req.Password)
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
