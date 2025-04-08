// user_handlers.go
package handlers

import (
	"context"
	"log"

	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/services"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
}

func (s *Server) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	log.Printf("📥 Received Register request: %v", req)

	repo := &repositories.MongoUserRepository{}
	services.RegisterUser(ctx, repo, req.GetName(), req.GetEmail(), req.GetPassword())

	return &userpb.RegisterResponse{
		Id:      "12345",
		Message: "User registered successfully",
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	log.Printf("📥 Received GetUser request: %v", req)
	return &userpb.UserResponse{
		Id:    req.GetId(),
		Name:  "Ali",
		Email: "ali@example.com",
	}, nil
}
