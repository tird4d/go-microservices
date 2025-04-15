// user_handlers.go
package handlers

import (
	"context"
	"log"

	"github.com/tird4d/go-microservices/user_service/events"
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

	_ = events.PublishUserRegisteredEvent(events.UserRegisteredEvent{
		UserID: "123456",
		Email:  req.GetEmail(),
		Name:   req.Name,
	})

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

func (s *Server) GetUserCredential(ctx context.Context, req *userpb.GetUserCredentialRequest) (*userpb.UserCredentialResponse, error) {
	log.Printf("📥 Received GetCredential request: %v", req)

	repo := &repositories.MongoUserRepository{}
	user, err := services.GetUserCredential(ctx, repo, req.GetEmail())
	if err != nil {
		log.Printf("❌ Error getting user credential: %v", err)
		return nil, err
	}
	log.Printf("✅ User credential retrieved: %v", user)

	return &userpb.UserCredentialResponse{
		Id:       user.ID.Hex(),
		Email:    user.Email,
		Password: user.Password,
	}, nil
}
