// user_handlers.go
package handlers

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tird4d/go-microservices/user_service/events"
	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/metrics"
	userpb "github.com/tird4d/go-microservices/user_service/proto"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
}

func (s *Server) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {

	// Start the timer for request duration
	timer := prometheus.NewTimer(metrics.RequestDurationHistogram.WithLabelValues("RegisterUser"))
	defer timer.ObserveDuration()
	// Increment the request counter for the RegisterUser endpoint
	metrics.RequestCounter.WithLabelValues("RegisterUser").Inc()

	logger.Log.Info("Received Register request", "request", req)

	repo := &repositories.MongoUserRepository{}
	result, err := services.RegisterUser(ctx, repo, req.GetName(), req.GetEmail(), req.GetPassword(), req.GetRole())
	if err != nil {
		return nil, err
	}

	//Publish user registered event
	err = events.PublishUserRegisteredEvent(events.UserRegisteredEvent{
		UserID: result.InsertedID.(primitive.ObjectID).Hex(),
		Email:  req.GetEmail(),
		Name:   req.Name,
	})

	if err != nil {
		logger.Log.Error("Error publishing user registered event", "error", err)
		return nil, err
	}

	return &userpb.RegisterResponse{
		Id:      result.InsertedID.(primitive.ObjectID).Hex(),
		Message: "User registered successfully....",
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {

	logger.Log.Info("Received GetUser request", "request", req)

	repo := &repositories.MongoUserRepository{}
	// Convert string ID to ObjectID

	user, err := services.GetUserByID(ctx, repo, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	return &userpb.UserResponse{
		Id:    user.ID.Hex(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *Server) GetUserCredential(ctx context.Context, req *userpb.GetUserCredentialRequest) (*userpb.UserCredentialResponse, error) {
	logger.Log.Info("Received GetCredential request", "request", req)

	repo := &repositories.MongoUserRepository{}
	user, err := services.GetUserCredential(ctx, repo, req.GetEmail())

	if err != nil {
		logger.Log.Error("Error getting user credential", "error", err)
		return nil, err
	}
	logger.Log.Info("User credential retrieved successfully", "user", user)

	return &userpb.UserCredentialResponse{
		Id:       user.ID.Hex(),
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
	}, nil
}
