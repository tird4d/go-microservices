// user_handlers.go
package handlers

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	// err = events.PublishUserRegisteredEvent(events.UserRegisteredEvent{
	// 	UserID: result.InsertedID.(primitive.ObjectID).Hex(),
	// 	Email:  req.GetEmail(),
	// 	Name:   req.Name,
	// })

	// if err != nil {
	// 	logger.Log.Error("Error publishing user registered event", "error", err)
	// 	return nil, err
	// }

	return &userpb.RegisterResponse{
		Id:      result.InsertedID.(primitive.ObjectID).Hex(),
		Message: "User registered successfully....",
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()

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

func (s *Server) GetAllUsers(ctx context.Context, req *userpb.GetAllUsersRequest) (*userpb.GetAllUsersResponse, error) {

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 10
	}

	repo := &repositories.MongoUserRepository{}

	users, totalCount, err := services.GetAllUsers(ctx, repo, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve users")
	}

	// protoUsers := make([]*userpb.UserResponse, 0, len(users))
	protoUsers := []*userpb.UserResponse{}
	for _, user := range users {
		protoUsers = append(protoUsers, &userpb.UserResponse{
			Id:    user.ID.Hex(),
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		})
	}
	return &userpb.GetAllUsersResponse{
		Users:       protoUsers,
		Total:       totalCount,
		CurrentPage: req.GetPage(),
		TotalPages:  int64(math.Ceil(float64(totalCount) / float64(pageSize))),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()

	repo := &repositories.MongoUserRepository{}

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		logger.Log.Error("Invalid user ID format", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format")
	}

	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = req.Name.GetValue()
	}
	if req.Email != nil {
		updates["email"] = req.Email.GetValue()
	}
	if req.Role != nil {
		updates["role"] = req.Role.GetValue()
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	result, err := services.UpdateUser(ctx, repo, oid, updates)

	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil, status.Errorf(codes.AlreadyExists, "Email already exists")
		}
		return nil, status.Errorf(codes.Internal, "Failed to update user")
	}

	if result.MatchedCount == 0 {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	return &userpb.UpdateUserResponse{
		Message: fmt.Sprintf("User updated (%d modified)", result.ModifiedCount),
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		logger.Log.Error("Invalid user ID format", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format")
	}

	repo := &repositories.MongoUserRepository{}

	result, err := services.DeleteUser(ctx, repo, oid)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to delete user")
	}

	return &userpb.DeleteUserResponse{
		Id:           req.Id,
		DeletedCount: result.DeletedCount,
		Message:      "User deleted successfully",
	}, nil
}
