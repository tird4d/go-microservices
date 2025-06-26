package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/tird4d/go-microservices/user_service/logger"
	"github.com/tird4d/go-microservices/user_service/models"
	"github.com/tird4d/go-microservices/user_service/repositories"
	"github.com/tird4d/go-microservices/user_service/utils"
	customErrors "github.com/tird4d/go-microservices/user_service/utils/errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RegisterUser(ctx context.Context, repo repositories.UserRepository, name, email, password, role string) (*mongo.InsertOneResult, error) {

	// Check if the email is already registered
	existingUser, err := repo.FindUserByEmail(ctx, email)
	if err != nil && !customErrors.IsNotFound(err) {
		logger.Log.Errorw("Failed to check existing email", "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user info")

	}

	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "email already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logger.Log.Errorw("Failed to hash password", "error", err)
		return nil, status.Error(codes.Internal, "password hashing failed")

	}

	// Set default role if not provided
	if role == "" {
		role = "user"
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     role,
	}

	result, err := repo.InsertNewUser(&user)
	if err != nil {
		logger.Log.Errorw("Insert user failed", "error", err)
		return nil, fmt.Errorf("user insert failed: %w", err)
	}

	return result, nil

}

func GetUserCredential(ctx context.Context, repo repositories.UserRepository, email string) (*models.User, error) {
	user, err := repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "user not found")

		}

		logger.Log.Errorw("Failed to find user by email", "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user info")

	}
	return user, nil
}

func GetUserByID(ctx context.Context, repo repositories.UserRepository, id string) (*models.User, error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logger.Log.Errorw("Failed to convert string to ObjectID", "error", err)
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	user, err := repo.FindUserByID(ctx, oid)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logger.Log.Errorw("Failed to find user by ID", "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user info")
	}
	return user, nil
}

func GetAllUsers(ctx context.Context, repo repositories.UserRepository, page, pageSize int64) ([]*models.User, int64, error) {

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	users, err := repo.FindUsers(ctx, skip, pageSize)
	if err != nil {
		return nil, 0, err
	}

	totalCount, err := repo.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil

}

type UpdateUserFields struct {
	Name  *string
	Email *string
	Role  *string
}

func UpdateUser(ctx context.Context, repo repositories.UserRepository, oid primitive.ObjectID, updates map[string]any) (*mongo.UpdateResult, error) {

	if updates["email"] != nil {
		existingUser, err := repo.FindUserByEmail(ctx, updates["email"].(string))
		if err != nil && !customErrors.IsNotFound(err) {
			logger.Log.Errorw("Failed to check existing email", "error", err)
			return nil, status.Error(codes.Internal, "failed to retrieve user info")
		}

		if existingUser != nil && existingUser.ID != oid {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
	}

	result, err := repo.UpdateUser(ctx, oid, updates)
	if err != nil {

		logger.Log.Errorw("Failed to update user", "error", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	return result, nil
}

func DeleteUser(ctx context.Context, repo repositories.UserRepository, oid primitive.ObjectID) (*mongo.DeleteResult, error) {
	_, err := repo.FindUserByID(ctx, oid)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logger.Log.Errorw("Failed to find user by ID", "error", err)
		return nil, err
	}

	result, err := repo.DeleteUser(ctx, oid)
	if err != nil {
		logger.Log.Errorw("Failed to delete user", "error", err)
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	if result.DeletedCount == 0 {
		logger.Log.Errorw("No user deleted", "user_id", oid.Hex())
		return nil, status.Error(codes.Internal, "no user deleted")
	}

	return result, nil
}
