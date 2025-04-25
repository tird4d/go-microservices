// user_handlers.go
package handlers

import (
	"context"
	"log"

	"github.com/tird4d/go-microservices/user_service/events"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
}

func (s *Server) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	log.Printf("📥 RReceived Register request: %v", req)

	repo := &repositories.MongoUserRepository{}
	result, err := services.RegisterUser(ctx, repo, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		log.Printf("❌ Error registering user: %v", err)
		return nil, err
	}
	log.Printf("✅ User registered successfully: %v", result)

	err = events.PublishUserRegisteredEvent(events.UserRegisteredEvent{
		UserID: result.InsertedID.(primitive.ObjectID).Hex(),
		Email:  req.GetEmail(),
		Name:   req.Name,
	})

	if err != nil {
		log.Printf("❌ Error publishing user registered event: %v", err)
		return nil, err
	}

	return &userpb.RegisterResponse{
		Id:      result.InsertedID.(primitive.ObjectID).Hex(),
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
